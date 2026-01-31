-- =====================================================
-- Hierarchical Integration Hub Migration
-- PostgreSQL Version
-- =====================================================
-- Description: Creates tables for hierarchical integration hub
--              with 4-level structure: Instance > Direction > Protocol Group > Endpoint
-- Features: Configuration inheritance, instance-based routing
-- Date: 2025-12-17
-- =====================================================

-- =====================================================
-- Table: hub_instances
-- Description: Hub instance assignments (links to system instances)
-- =====================================================
CREATE TABLE IF NOT EXISTS hub_instances (
    id VARCHAR(255) PRIMARY KEY,
    instance_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,
    UNIQUE(instance_id)
);

CREATE INDEX idx_hub_instances_instance_id ON hub_instances(instance_id);
CREATE INDEX idx_hub_instances_enabled ON hub_instances(enabled);
CREATE INDEX idx_hub_instances_active ON hub_instances(active);
CREATE INDEX idx_hub_instances_name ON hub_instances(name);

COMMENT ON TABLE hub_instances IS 'Hub instance assignments for managing integration configurations';
COMMENT ON COLUMN hub_instances.instance_id IS 'Links to system InstanceID (com.InstanceID)';
COMMENT ON COLUMN hub_instances.metadata IS 'Additional metadata in JSON format';

-- =====================================================
-- Table: hub_protocol_groups
-- Description: Protocol-level configuration groups with inheritance base
-- =====================================================
CREATE TABLE IF NOT EXISTS hub_protocol_groups (
    id VARCHAR(255) PRIMARY KEY,
    instance_id VARCHAR(255) NOT NULL,
    direction VARCHAR(20) NOT NULL CHECK (direction IN ('inbound', 'outbound')),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    protocol VARCHAR(50) NOT NULL,

    -- Base configuration that endpoints inherit
    base_config JSONB NOT NULL DEFAULT '{}'::JSONB,

    -- Protocol-specific defaults
    message_type VARCHAR(50),
    timeout INT DEFAULT 30000,
    retry_attempts INT DEFAULT 3,
    retry_interval INT DEFAULT 1000,

    -- Authentication defaults
    auth_type VARCHAR(50),
    auth_config JSONB,

    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (instance_id) REFERENCES hub_instances(id) ON DELETE CASCADE,
    UNIQUE(instance_id, direction, name)
);

CREATE INDEX idx_hub_protocol_groups_instance ON hub_protocol_groups(instance_id);
CREATE INDEX idx_hub_protocol_groups_direction ON hub_protocol_groups(direction);
CREATE INDEX idx_hub_protocol_groups_protocol ON hub_protocol_groups(protocol);
CREATE INDEX idx_hub_protocol_groups_enabled ON hub_protocol_groups(enabled);
CREATE INDEX idx_hub_protocol_groups_active ON hub_protocol_groups(active);

COMMENT ON TABLE hub_protocol_groups IS 'Protocol-level configuration groups (REST, MQTT, Kafka, etc.)';
COMMENT ON COLUMN hub_protocol_groups.direction IS 'Traffic direction: inbound or outbound';
COMMENT ON COLUMN hub_protocol_groups.protocol IS 'Protocol type: REST, SOAP, MQTT, Kafka, TCP, GraphQL, etc.';
COMMENT ON COLUMN hub_protocol_groups.base_config IS 'Base configuration inherited by child endpoints';

-- =====================================================
-- Table: hub_endpoints
-- Description: Individual endpoint configurations with override capability
-- =====================================================
CREATE TABLE IF NOT EXISTS hub_endpoints (
    id VARCHAR(255) PRIMARY KEY,
    protocol_group_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Endpoint-specific configuration
    endpoint_url VARCHAR(500),
    port INT,
    path VARCHAR(500),
    method VARCHAR(10),

    -- Override settings (NULL means inherit from protocol_group)
    override_config JSONB,
    message_type VARCHAR(50),
    timeout INT,
    retry_attempts INT,
    retry_interval INT,

    -- Authentication overrides
    auth_type VARCHAR(50),
    auth_config JSONB,

    -- Protocol-specific configs
    queue_config JSONB,
    file_config JSONB,

    -- Schema validation
    validate_schema BOOLEAN DEFAULT FALSE,
    schema_definition TEXT,

    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (protocol_group_id) REFERENCES hub_protocol_groups(id) ON DELETE CASCADE,
    UNIQUE(protocol_group_id, name)
);

CREATE INDEX idx_hub_endpoints_protocol_group ON hub_endpoints(protocol_group_id);
CREATE INDEX idx_hub_endpoints_enabled ON hub_endpoints(enabled);
CREATE INDEX idx_hub_endpoints_active ON hub_endpoints(active);
CREATE INDEX idx_hub_endpoints_endpoint_url ON hub_endpoints(endpoint_url);
CREATE INDEX idx_hub_endpoints_name ON hub_endpoints(name);

COMMENT ON TABLE hub_endpoints IS 'Individual endpoint configurations with inheritance from protocol groups';
COMMENT ON COLUMN hub_endpoints.override_config IS 'Override specific base_config values from protocol_group';
COMMENT ON COLUMN hub_endpoints.timeout IS 'Timeout override (NULL = inherit from protocol_group)';
COMMENT ON COLUMN hub_endpoints.queue_config IS 'Queue configuration for MQTT, AMQP, Kafka, RabbitMQ';
COMMENT ON COLUMN hub_endpoints.file_config IS 'File configuration for FTP, SFTP, File protocols';

-- =====================================================
-- Table: hub_routes
-- Description: Message routing rules between endpoints
-- =====================================================
CREATE TABLE IF NOT EXISTS hub_routes (
    id VARCHAR(255) PRIMARY KEY,
    source_endpoint_id VARCHAR(255) NOT NULL,
    destination_endpoint_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Routing logic
    source_filter TEXT,
    conditions JSONB,

    -- Transformation
    transformation_type VARCHAR(50),
    transformation TEXT,
    field_mappings JSONB,

    -- Execution
    priority INT NOT NULL DEFAULT 0,
    async_mode BOOLEAN DEFAULT FALSE,

    -- Error handling
    on_error VARCHAR(50),
    dead_letter_queue VARCHAR(255),
    max_retries INT DEFAULT 3,

    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (source_endpoint_id) REFERENCES hub_endpoints(id) ON DELETE CASCADE,
    FOREIGN KEY (destination_endpoint_id) REFERENCES hub_endpoints(id) ON DELETE CASCADE
);

CREATE INDEX idx_hub_routes_source ON hub_routes(source_endpoint_id);
CREATE INDEX idx_hub_routes_destination ON hub_routes(destination_endpoint_id);
CREATE INDEX idx_hub_routes_priority ON hub_routes(priority DESC);
CREATE INDEX idx_hub_routes_enabled ON hub_routes(enabled);
CREATE INDEX idx_hub_routes_active ON hub_routes(active);
CREATE INDEX idx_hub_routes_name ON hub_routes(name);

COMMENT ON TABLE hub_routes IS 'Message routing rules between source and destination endpoints';
COMMENT ON COLUMN hub_routes.source_filter IS 'JSONPath or XPath filter for source messages';
COMMENT ON COLUMN hub_routes.conditions IS 'Array of routing conditions (field, operator, value)';
COMMENT ON COLUMN hub_routes.transformation_type IS 'None, XSLT, JavaScript, JSONata, Custom';
COMMENT ON COLUMN hub_routes.on_error IS 'Error handling: Retry, Skip, SendToDeadLetter';

-- =====================================================
-- Table: hub_endpoint_jobs
-- Description: Job configurations linked to specific endpoints
-- =====================================================
CREATE TABLE IF NOT EXISTS hub_endpoint_jobs (
    id VARCHAR(255) PRIMARY KEY,
    endpoint_id VARCHAR(255) NOT NULL,
    job_id VARCHAR(255) NOT NULL,
    job_type VARCHAR(50) NOT NULL CHECK (job_type IN ('Manual', 'Scheduled', 'Triggered')),

    -- Schedule configuration
    cron_expression VARCHAR(100),
    interval_seconds INT,

    -- Trigger configuration
    trigger_event_type VARCHAR(100),
    trigger_event_source VARCHAR(255),
    trigger_event_filter TEXT,

    -- Execution
    timeout INT DEFAULT 300,
    max_concurrent INT DEFAULT 1,
    retry_on_failure BOOLEAN DEFAULT TRUE,
    max_retries INT DEFAULT 3,

    parameters JSONB,

    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    metadata JSONB,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    referenceid VARCHAR(255),
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    rowversionstamp INT NOT NULL DEFAULT 1,

    FOREIGN KEY (endpoint_id) REFERENCES hub_endpoints(id) ON DELETE CASCADE
);

CREATE INDEX idx_hub_endpoint_jobs_endpoint ON hub_endpoint_jobs(endpoint_id);
CREATE INDEX idx_hub_endpoint_jobs_job ON hub_endpoint_jobs(job_id);
CREATE INDEX idx_hub_endpoint_jobs_job_type ON hub_endpoint_jobs(job_type);
CREATE INDEX idx_hub_endpoint_jobs_enabled ON hub_endpoint_jobs(enabled);
CREATE INDEX idx_hub_endpoint_jobs_active ON hub_endpoint_jobs(active);

COMMENT ON TABLE hub_endpoint_jobs IS 'Job configurations linked to specific endpoints';
COMMENT ON COLUMN hub_endpoint_jobs.job_type IS 'Manual, Scheduled, or Triggered';
COMMENT ON COLUMN hub_endpoint_jobs.cron_expression IS 'Cron expression for scheduled jobs';

-- =====================================================
-- Table: hub_migration_log
-- Description: Tracks migration from old flat structure to hierarchical
-- =====================================================
CREATE TABLE IF NOT EXISTS hub_migration_log (
    id VARCHAR(255) PRIMARY KEY,
    old_hub_id VARCHAR(255),
    new_instance_id VARCHAR(255),
    migration_status VARCHAR(50),
    migration_details JSONB,
    error_message TEXT,
    migrated_at TIMESTAMP,
    createdby VARCHAR(255),
    createdon TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_hub_migration_log_old_hub_id ON hub_migration_log(old_hub_id);
CREATE INDEX idx_hub_migration_log_new_instance_id ON hub_migration_log(new_instance_id);
CREATE INDEX idx_hub_migration_log_status ON hub_migration_log(migration_status);

COMMENT ON TABLE hub_migration_log IS 'Migration log from old Integration_Hub collection to new hierarchical structure';
COMMENT ON COLUMN hub_migration_log.migration_status IS 'pending, completed, failed';

-- =====================================================
-- Sample Data (Optional - for development/testing)
-- =====================================================

-- Uncomment to insert sample data:
/*
-- Sample Instance
INSERT INTO hub_instances (id, instance_id, name, description, enabled, active, createdby)
VALUES ('hub_inst_001', 'iac-instance-uuid-001', 'Primary IAC Instance', 'Main integration hub instance', TRUE, TRUE, 'system');

-- Sample Protocol Group (Inbound REST)
INSERT INTO hub_protocol_groups (id, instance_id, direction, name, description, protocol, base_config, timeout, retry_attempts, auth_type, enabled, active, createdby)
VALUES (
    'pg_001',
    'hub_inst_001',
    'inbound',
    'REST APIs',
    'Inbound REST API endpoints',
    'REST',
    '{"ssl_verify": true, "compression": "gzip"}'::JSONB,
    30000,
    3,
    'Bearer',
    TRUE,
    TRUE,
    'system'
);

-- Sample Endpoint
INSERT INTO hub_endpoints (id, protocol_group_id, name, description, endpoint_url, port, path, method, enabled, active, createdby)
VALUES (
    'ep_001',
    'pg_001',
    'Customer API',
    'Customer data API endpoint',
    'https://api.example.com',
    443,
    '/api/customers',
    'POST',
    TRUE,
    TRUE,
    'system'
);
*/

-- =====================================================
-- End of Migration Script
-- =====================================================
