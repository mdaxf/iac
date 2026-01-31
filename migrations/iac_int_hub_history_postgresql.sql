-- =====================================================
-- Integration Hub History Tables Migration
-- PostgreSQL Version
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
    direction VARCHAR(20) NOT NULL CHECK (direction IN ('inbound', 'outbound')),
    protocol VARCHAR(50) NOT NULL,
    protocol_group_id VARCHAR(255),
    protocol_group_name VARCHAR(255),
    endpoint_id VARCHAR(255),
    endpoint_name VARCHAR(255),
    topic VARCHAR(255),
    source VARCHAR(500),
    payload TEXT,
    payload_size INT,
    message_type VARCHAR(50),
    action VARCHAR(100),
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    mapped_data TEXT,
    response TEXT,
    response_status INT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_ms BIGINT,
    instance_id VARCHAR(255),
    instance_name VARCHAR(255),
    user_id VARCHAR(255),
    client_id VARCHAR(255),
    metadata JSONB,

    -- IAC Standard Fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1
);

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

-- Comments
COMMENT ON TABLE iac_int_hub_history IS 'Integration hub transaction history records';
COMMENT ON COLUMN iac_int_hub_history.hub_id IS 'Reference to integration hub configuration';
COMMENT ON COLUMN iac_int_hub_history.direction IS 'Traffic direction: inbound or outbound';
COMMENT ON COLUMN iac_int_hub_history.protocol IS 'Protocol type: REST, MQTT, Kafka, etc.';
COMMENT ON COLUMN iac_int_hub_history.topic IS 'Topic name for message broker protocols';
COMMENT ON COLUMN iac_int_hub_history.payload IS 'Request payload (JSON/XML/etc.)';
COMMENT ON COLUMN iac_int_hub_history.mapped_data IS 'Data after mapping/transformation';
COMMENT ON COLUMN iac_int_hub_history.status IS 'Transaction status: pending, processing, completed, failed';
COMMENT ON COLUMN iac_int_hub_history.duration_ms IS 'Duration in milliseconds';
COMMENT ON COLUMN iac_int_hub_history.metadata IS 'Additional metadata in JSON format';

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
    action_config TEXT,
    input_data TEXT,
    output_data TEXT,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'skipped')),
    error_message TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration_ms BIGINT,
    metadata JSONB,

    -- IAC Standard Fields
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    -- Foreign Key
    FOREIGN KEY (int_hub_his_id) REFERENCES iac_int_hub_history(id) ON DELETE CASCADE
);

-- Indexes for action history
CREATE INDEX idx_iac_int_hub_action_history_his_id ON iac_int_hub_action_history(int_hub_his_id);
CREATE INDEX idx_iac_int_hub_action_history_sequence ON iac_int_hub_action_history(int_hub_his_id, action_sequence);
CREATE INDEX idx_iac_int_hub_action_history_type ON iac_int_hub_action_history(action_type);
CREATE INDEX idx_iac_int_hub_action_history_status ON iac_int_hub_action_history(status);
CREATE INDEX idx_iac_int_hub_action_history_active ON iac_int_hub_action_history(active);

-- Comments
COMMENT ON TABLE iac_int_hub_action_history IS 'Individual action records within integration hub transactions';
COMMENT ON COLUMN iac_int_hub_action_history.int_hub_his_id IS 'Reference to parent transaction history record';
COMMENT ON COLUMN iac_int_hub_action_history.action_sequence IS 'Order of action within the transaction';
COMMENT ON COLUMN iac_int_hub_action_history.action_type IS 'Action type: route, transform, validate, store, webhook, script';
COMMENT ON COLUMN iac_int_hub_action_history.action_config IS 'Configuration used for this action';
COMMENT ON COLUMN iac_int_hub_action_history.input_data IS 'Input data to this action';
COMMENT ON COLUMN iac_int_hub_action_history.output_data IS 'Output data from this action';
COMMENT ON COLUMN iac_int_hub_action_history.status IS 'Action status: pending, processing, completed, failed, skipped';

-- =====================================================
-- End of Migration Script
-- =====================================================
