-- API Call History Tables for PostgreSQL
-- Run this script to create the necessary tables for the API Call History feature

-- =====================================================
-- Table: iac_api_call_history
-- Description: Stores API call history records
-- =====================================================
CREATE TABLE IF NOT EXISTS iac_api_call_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Request Information
    method VARCHAR(10) NOT NULL,
    endpoint VARCHAR(500) NOT NULL,
    full_path VARCHAR(2000),
    request_headers JSONB,
    request_body JSONB,
    query_params JSONB,

    -- Response Information
    status_code INTEGER NOT NULL,
    response_body JSONB,
    response_headers JSONB,

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
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_ms BIGINT NOT NULL,

    -- Instance Information
    instance_id VARCHAR(100),
    instance_name VARCHAR(255),

    -- Error Information
    error_message TEXT,

    -- Metadata
    tags TEXT[],
    metadata JSONB,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_api_call_history_start_time ON iac_api_call_history(start_time DESC);
CREATE INDEX IF NOT EXISTS idx_api_call_history_endpoint ON iac_api_call_history(endpoint);
CREATE INDEX IF NOT EXISTS idx_api_call_history_method ON iac_api_call_history(method);
CREATE INDEX IF NOT EXISTS idx_api_call_history_status_code ON iac_api_call_history(status_code);
CREATE INDEX IF NOT EXISTS idx_api_call_history_user_id ON iac_api_call_history(user_id);
CREATE INDEX IF NOT EXISTS idx_api_call_history_source_ip ON iac_api_call_history(source_ip);
CREATE INDEX IF NOT EXISTS idx_api_call_history_instance_id ON iac_api_call_history(instance_id);

-- Composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_api_call_history_endpoint_time ON iac_api_call_history(endpoint, start_time DESC);
CREATE INDEX IF NOT EXISTS idx_api_call_history_user_time ON iac_api_call_history(user_id, start_time DESC);

-- Index for error queries
CREATE INDEX IF NOT EXISTS idx_api_call_history_errors ON iac_api_call_history(status_code) WHERE status_code >= 400;

-- =====================================================
-- Table: iac_api_call_history_config
-- Description: Stores API call history tracking configuration
-- =====================================================
CREATE TABLE IF NOT EXISTS iac_api_call_history_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Master Enable/Disable
    enabled BOOLEAN NOT NULL DEFAULT false,

    -- Endpoint Filters
    include_endpoints TEXT[],
    exclude_endpoints TEXT[],
    include_methods TEXT[],
    exclude_methods TEXT[],

    -- Source Filters
    include_source_ips TEXT[],
    exclude_source_ips TEXT[],
    include_users TEXT[],
    exclude_users TEXT[],

    -- Status Filters
    include_status_codes INTEGER[],
    exclude_status_codes INTEGER[],
    only_errors BOOLEAN NOT NULL DEFAULT false,

    -- Data Capture Options
    capture_request_body BOOLEAN NOT NULL DEFAULT true,
    capture_response_body BOOLEAN NOT NULL DEFAULT true,
    capture_headers BOOLEAN NOT NULL DEFAULT false,
    max_body_size INTEGER NOT NULL DEFAULT 10240,

    -- Sensitive Data Handling
    mask_sensitive_fields TEXT[],

    -- Retention
    retention_days INTEGER NOT NULL DEFAULT 30,

    -- Sampling
    sampling_rate DECIMAL(3,2) NOT NULL DEFAULT 1.00,

    -- Audit Information
    updated_by VARCHAR(100),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =====================================================
-- Insert default configuration
-- =====================================================
INSERT INTO iac_api_call_history_config (
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
    gen_random_uuid(),
    false,
    ARRAY[]::TEXT[],
    ARRAY['/health', '/api/apicallhistory']::TEXT[],
    ARRAY[]::TEXT[],
    ARRAY[]::TEXT[],
    ARRAY[]::TEXT[],
    ARRAY[]::TEXT[],
    ARRAY[]::TEXT[],
    ARRAY[]::TEXT[],
    ARRAY[]::INTEGER[],
    ARRAY[]::INTEGER[],
    false,
    true,
    true,
    false,
    10240,
    ARRAY['password', 'token', 'secret', 'apikey', 'authorization']::TEXT[],
    30,
    1.00,
    'system'
) ON CONFLICT DO NOTHING;

-- =====================================================
-- Function to auto-delete old records based on retention
-- =====================================================
CREATE OR REPLACE FUNCTION cleanup_api_call_history()
RETURNS INTEGER AS $$
DECLARE
    retention_days_val INTEGER;
    deleted_count INTEGER;
BEGIN
    -- Get retention days from config
    SELECT retention_days INTO retention_days_val
    FROM iac_api_call_history_config
    LIMIT 1;

    IF retention_days_val IS NULL THEN
        retention_days_val := 30;
    END IF;

    -- Delete old records
    DELETE FROM iac_api_call_history
    WHERE start_time < CURRENT_TIMESTAMP - (retention_days_val || ' days')::INTERVAL;

    GET DIAGNOSTICS deleted_count = ROW_COUNT;

    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- Optional: Create a scheduled job for cleanup (requires pg_cron extension)
-- Uncomment if pg_cron is available
-- =====================================================
-- SELECT cron.schedule('api_call_history_cleanup', '0 2 * * *', 'SELECT cleanup_api_call_history()');

COMMENT ON TABLE iac_api_call_history IS 'Stores API call history records for auditing and monitoring';
COMMENT ON TABLE iac_api_call_history_config IS 'Configuration for API call history tracking';
