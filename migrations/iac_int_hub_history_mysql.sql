-- =====================================================
-- Integration Hub History Tables Migration
-- MySQL Version
-- =====================================================
-- Description: Creates tables for tracking integration hub transaction history
--              including main history records and action-level tracking
-- Features: Transaction tracking, action timeline, statistics support
-- Date: 2025-01-30
-- =====================================================

-- =====================================================
-- Table: iac_int_hub_history
-- Description: Main transaction history records for integration hub
-- =====================================================
CREATE TABLE IF NOT EXISTS iac_int_hub_history (
    id VARCHAR(255) PRIMARY KEY,
    hub_id VARCHAR(255) NOT NULL,
    hub_name VARCHAR(255),
    direction VARCHAR(20) NOT NULL,
    protocol VARCHAR(50) NOT NULL,
    protocol_group_id VARCHAR(255),
    protocol_group_name VARCHAR(255),
    endpoint_id VARCHAR(255),
    endpoint_name VARCHAR(255),
    topic VARCHAR(255),
    source VARCHAR(500),
    payload LONGTEXT,
    payload_size INT,
    message_type VARCHAR(50),
    action VARCHAR(100),
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    mapped_data LONGTEXT,
    response LONGTEXT,
    response_status INT,
    start_time DATETIME(3) NOT NULL,
    end_time DATETIME(3),
    duration_ms BIGINT,
    instance_id VARCHAR(255),
    instance_name VARCHAR(255),
    user_id VARCHAR(255),
    client_id VARCHAR(255),
    metadata JSON,

    -- IAC Standard Fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    -- Constraints
    CONSTRAINT chk_iac_int_hub_history_direction CHECK (direction IN ('inbound', 'outbound')),
    CONSTRAINT chk_iac_int_hub_history_status CHECK (status IN ('pending', 'processing', 'completed', 'failed'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Indexes for common queries
CREATE INDEX idx_iac_int_hub_history_hub_id ON iac_int_hub_history(hub_id);
CREATE INDEX idx_iac_int_hub_history_direction ON iac_int_hub_history(direction);
CREATE INDEX idx_iac_int_hub_history_protocol ON iac_int_hub_history(protocol);
CREATE INDEX idx_iac_int_hub_history_status ON iac_int_hub_history(status);
CREATE INDEX idx_iac_int_hub_history_start_time ON iac_int_hub_history(start_time DESC);
CREATE INDEX idx_iac_int_hub_history_endpoint_id ON iac_int_hub_history(endpoint_id);
CREATE INDEX idx_iac_int_hub_history_topic ON iac_int_hub_history(topic);
CREATE INDEX idx_iac_int_hub_history_instance_id ON iac_int_hub_history(instance_id);
CREATE INDEX idx_iac_int_hub_history_active ON iac_int_hub_history(active);

-- Composite indexes for filtered queries
CREATE INDEX idx_iac_int_hub_history_hub_status ON iac_int_hub_history(hub_id, status);
CREATE INDEX idx_iac_int_hub_history_hub_direction_time ON iac_int_hub_history(hub_id, direction, start_time DESC);

-- =====================================================
-- Table: iac_int_hub_action_history
-- Description: Individual action records within a transaction
-- =====================================================
CREATE TABLE IF NOT EXISTS iac_int_hub_action_history (
    id VARCHAR(255) PRIMARY KEY,
    int_hub_his_id VARCHAR(255) NOT NULL,
    action_sequence INT NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    action_name VARCHAR(255),
    action_config LONGTEXT,
    input_data LONGTEXT,
    output_data LONGTEXT,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    start_time DATETIME(3) NOT NULL,
    end_time DATETIME(3),
    duration_ms BIGINT,
    metadata JSON,

    -- IAC Standard Fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    -- Constraints
    CONSTRAINT chk_iac_int_hub_action_history_status CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'skipped')),

    -- Foreign Key
    CONSTRAINT fk_iac_int_hub_action_history_his_id
        FOREIGN KEY (int_hub_his_id) REFERENCES iac_int_hub_history(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Indexes for action history
CREATE INDEX idx_iac_int_hub_action_history_his_id ON iac_int_hub_action_history(int_hub_his_id);
CREATE INDEX idx_iac_int_hub_action_history_sequence ON iac_int_hub_action_history(int_hub_his_id, action_sequence);
CREATE INDEX idx_iac_int_hub_action_history_type ON iac_int_hub_action_history(action_type);
CREATE INDEX idx_iac_int_hub_action_history_status ON iac_int_hub_action_history(status);
CREATE INDEX idx_iac_int_hub_action_history_active ON iac_int_hub_action_history(active);

-- =====================================================
-- End of Migration Script
-- =====================================================
