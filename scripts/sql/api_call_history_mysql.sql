-- API Call History Tables for MySQL
-- Run this script to create the necessary tables for the API Call History feature
-- Requires MySQL 5.7+ (for JSON support) or MySQL 8.0+ (recommended)

-- =====================================================
-- Table: iac_api_call_history
-- Description: Stores API call history records
-- =====================================================
CREATE TABLE IF NOT EXISTS iac_api_call_history (
    id CHAR(36) PRIMARY KEY,

    -- Request Information
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    full_path VARCHAR(2000),
    request_headers JSON,
    request_body JSON,
    query_params JSON,

    -- Response Information
    status_code INT NOT NULL,
    response_body JSON,
    response_headers JSON,

    -- Source Information
    source_ip VARCHAR(45) NOT NULL,
    source_machine VARCHAR(255),
    user_agent VARCHAR(1000),

    -- User Information
    user_id VARCHAR(100),
    user_name VARCHAR(255),
    client_id VARCHAR(100),
    auth_type VARCHAR(50),

    -- Timing Information
    start_time DATETIME(6) NOT NULL,
    end_time DATETIME(6) NOT NULL,
    duration_ms BIGINT NOT NULL,

    -- Instance Information
    instance_id VARCHAR(100),
    instance_name VARCHAR(255),

    -- Error Information
    error_message TEXT,

    -- Metadata
    tags JSON,
    metadata JSON,

    -- Audit
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6),

    -- Indexes
    INDEX idx_start_time (start_time DESC),
    INDEX idx_endpoint (endpoint(255)),
    INDEX idx_method (method),
    INDEX idx_status_code (status_code),
    INDEX idx_user_id (user_id),
    INDEX idx_source_ip (source_ip),
    INDEX idx_instance_id (instance_id),
    INDEX idx_endpoint_time (endpoint(255), start_time DESC),
    INDEX idx_user_time (user_id, start_time DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Table: iac_api_call_history_config
-- Description: Stores API call history tracking configuration
-- =====================================================
CREATE TABLE IF NOT EXISTS iac_api_call_history_config (
    id CHAR(36) PRIMARY KEY,

    -- Master Enable/Disable
    enabled BOOLEAN NOT NULL DEFAULT FALSE,

    -- Endpoint Filters (stored as JSON arrays)
    include_endpoints JSON,
    exclude_endpoints JSON,
    include_methods JSON,
    exclude_methods JSON,

    -- Source Filters
    include_source_ips JSON,
    exclude_source_ips JSON,
    include_users JSON,
    exclude_users JSON,

    -- Status Filters
    include_status_codes JSON,
    exclude_status_codes JSON,
    only_errors BOOLEAN NOT NULL DEFAULT FALSE,

    -- Data Capture Options
    capture_request_body BOOLEAN NOT NULL DEFAULT TRUE,
    capture_response_body BOOLEAN NOT NULL DEFAULT TRUE,
    capture_headers BOOLEAN NOT NULL DEFAULT FALSE,
    max_body_size INT NOT NULL DEFAULT 10240,

    -- Sensitive Data Handling
    mask_sensitive_fields JSON,

    -- Retention
    retention_days INT NOT NULL DEFAULT 30,

    -- Sampling
    sampling_rate DECIMAL(3,2) NOT NULL DEFAULT 1.00,

    -- Audit Information
    updated_by VARCHAR(100),
    updated_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    created_at DATETIME(6) DEFAULT CURRENT_TIMESTAMP(6)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- =====================================================
-- Insert default configuration
-- =====================================================
INSERT IGNORE INTO iac_api_call_history_config (
    id,
    enabled,
    include_endpoints,
    exclude_endpoints,
    include_methods,
    exclude_methods,
    include_source_ips,
    exclude_source_ips,
    include_users,
    exclude_users,
    include_status_codes,
    exclude_status_codes,
    only_errors,
    capture_request_body,
    capture_response_body,
    capture_headers,
    max_body_size,
    mask_sensitive_fields,
    retention_days,
    sampling_rate,
    updated_by
) VALUES (
    UUID(),
    FALSE,
    JSON_ARRAY(),
    JSON_ARRAY('/health', '/api/apicallhistory'),
    JSON_ARRAY(),
    JSON_ARRAY(),
    JSON_ARRAY(),
    JSON_ARRAY(),
    JSON_ARRAY(),
    JSON_ARRAY(),
    JSON_ARRAY(),
    JSON_ARRAY(),
    FALSE,
    TRUE,
    TRUE,
    FALSE,
    10240,
    JSON_ARRAY('password', 'token', 'secret', 'apikey', 'authorization'),
    30,
    1.00,
    'system'
);

-- =====================================================
-- Stored Procedure to cleanup old records based on retention
-- =====================================================
DELIMITER //

CREATE PROCEDURE IF NOT EXISTS cleanup_api_call_history()
BEGIN
    DECLARE retention_days_val INT DEFAULT 30;
    DECLARE deleted_count INT DEFAULT 0;

    -- Get retention days from config
    SELECT retention_days INTO retention_days_val
    FROM iac_api_call_history_config
    LIMIT 1;

    IF retention_days_val IS NULL THEN
        SET retention_days_val = 30;
    END IF;

    -- Delete old records
    DELETE FROM iac_api_call_history
    WHERE start_time < DATE_SUB(NOW(), INTERVAL retention_days_val DAY);

    SELECT ROW_COUNT() INTO deleted_count;

    SELECT deleted_count AS records_deleted;
END //

DELIMITER ;

-- =====================================================
-- Event to auto-cleanup (requires event_scheduler = ON)
-- Run: SET GLOBAL event_scheduler = ON;
-- =====================================================
CREATE EVENT IF NOT EXISTS evt_cleanup_api_call_history
ON SCHEDULE EVERY 1 DAY
STARTS CURRENT_DATE + INTERVAL 2 HOUR
DO
    CALL cleanup_api_call_history();

-- =====================================================
-- Helper view for error records
-- =====================================================
CREATE OR REPLACE VIEW v_api_call_history_errors AS
SELECT
    id,
    method,
    endpoint,
    status_code,
    error_message,
    user_id,
    user_name,
    source_ip,
    start_time,
    duration_ms,
    instance_name
FROM iac_api_call_history
WHERE status_code >= 400
ORDER BY start_time DESC;

-- =====================================================
-- Helper view for statistics
-- =====================================================
CREATE OR REPLACE VIEW v_api_call_history_stats AS
SELECT
    DATE(start_time) AS call_date,
    endpoint,
    method,
    COUNT(*) AS total_calls,
    AVG(duration_ms) AS avg_duration_ms,
    MAX(duration_ms) AS max_duration_ms,
    MIN(duration_ms) AS min_duration_ms,
    SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) AS error_count,
    SUM(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 ELSE 0 END) AS success_count
FROM iac_api_call_history
GROUP BY DATE(start_time), endpoint, method;
