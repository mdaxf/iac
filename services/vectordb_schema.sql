-- AI Vector Embeddings Schema for PostgreSQL (IAC Standard)
-- Stores vector embeddings for schema metadata, business entities, and query templates
-- Follows IAC standard with: id, uuid, referenceid, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp

-- Enable pgvector extension for vector operations
CREATE EXTENSION IF NOT EXISTS vector;

-- Create sequence for id generation
CREATE SEQUENCE IF NOT EXISTS ai_embedding_config_id_seq;
CREATE SEQUENCE IF NOT EXISTS db_schema_embedding_id_seq;
CREATE SEQUENCE IF NOT EXISTS business_entity_id_seq;
CREATE SEQUENCE IF NOT EXISTS query_template_id_seq;
CREATE SEQUENCE IF NOT EXISTS embedding_job_id_seq;
CREATE SEQUENCE IF NOT EXISTS embedding_search_log_id_seq;

-- Table: ai_embedding_configurations
-- Stores AI configuration metadata and embedding model information
CREATE TABLE IF NOT EXISTS ai_embedding_configurations (
    id INTEGER PRIMARY KEY DEFAULT nextval('ai_embedding_config_id_seq'),
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    referenceid VARCHAR(255) UNIQUE,
    config_name VARCHAR(255) NOT NULL UNIQUE,
    embedding_model VARCHAR(255) NOT NULL,
    embedding_dimensions INTEGER NOT NULL,
    vector_database_type VARCHAR(50) NOT NULL DEFAULT 'postgresql',
    vector_database_config JSONB,
    active BOOLEAN DEFAULT true,
    createdby VARCHAR(255) NOT NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP,
    rowversionstamp INTEGER DEFAULT 1,
    CONSTRAINT chk_embedding_dimensions CHECK (embedding_dimensions > 0)
);

CREATE INDEX idx_ai_embedding_config_active ON ai_embedding_configurations(active) WHERE active = true;
CREATE INDEX idx_ai_embedding_config_uuid ON ai_embedding_configurations(uuid);

-- Table: database_schema_embeddings
-- Stores vector embeddings for database tables and columns metadata
CREATE TABLE IF NOT EXISTS database_schema_embeddings (
    id INTEGER PRIMARY KEY DEFAULT nextval('db_schema_embedding_id_seq'),
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    referenceid VARCHAR(255) UNIQUE,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    database_alias VARCHAR(255) NOT NULL,
    schema_name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    column_name VARCHAR(255),
    description TEXT,
    metadata JSONB,
    embedding vector,
    embedding_hash VARCHAR(64),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT true,
    createdby VARCHAR(255) NOT NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP,
    rowversionstamp INTEGER DEFAULT 1,
    CONSTRAINT uq_schema_embedding UNIQUE(config_id, database_alias, schema_name, table_name, column_name)
);

CREATE INDEX idx_db_schema_emb_config ON database_schema_embeddings(config_id);
CREATE INDEX idx_db_schema_emb_db_alias ON database_schema_embeddings(database_alias);
CREATE INDEX idx_db_schema_emb_table ON database_schema_embeddings(schema_name, table_name);
CREATE INDEX idx_db_schema_emb_generated ON database_schema_embeddings(generated_at DESC);
CREATE INDEX idx_db_schema_emb_active ON database_schema_embeddings(active) WHERE active = true;
CREATE INDEX idx_db_schema_emb_uuid ON database_schema_embeddings(uuid);

-- Table: business_entities
-- Stores business entity definitions with vector embeddings for natural language search
CREATE TABLE IF NOT EXISTS business_entities (
    id INTEGER PRIMARY KEY DEFAULT nextval('business_entity_id_seq'),
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    referenceid VARCHAR(255) UNIQUE,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    entity_name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100),
    description TEXT NOT NULL,
    database_alias VARCHAR(255),
    schema_name VARCHAR(255),
    table_name VARCHAR(255),
    field_mappings JSONB,
    relationships JSONB,
    business_rules JSONB,
    metadata JSONB,
    embedding vector,
    embedding_hash VARCHAR(64),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT true,
    createdby VARCHAR(255) NOT NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP,
    rowversionstamp INTEGER DEFAULT 1,
    CONSTRAINT uq_business_entity UNIQUE(config_id, entity_name)
);

CREATE INDEX idx_business_entity_config ON business_entities(config_id);
CREATE INDEX idx_business_entity_type ON business_entities(entity_type);
CREATE INDEX idx_business_entity_active ON business_entities(active) WHERE active = true;
CREATE INDEX idx_business_entity_db ON business_entities(database_alias, schema_name, table_name);
CREATE INDEX idx_business_entity_generated ON business_entities(generated_at DESC);
CREATE INDEX idx_business_entity_uuid ON business_entities(uuid);

-- Table: query_templates
-- Stores query templates with vector embeddings for natural language to SQL conversion
CREATE TABLE IF NOT EXISTS query_templates (
    id INTEGER PRIMARY KEY DEFAULT nextval('query_template_id_seq'),
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    referenceid VARCHAR(255) UNIQUE,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    template_name VARCHAR(255) NOT NULL,
    template_category VARCHAR(100),
    natural_language_query TEXT NOT NULL,
    sql_template TEXT NOT NULL,
    parameters JSONB,
    database_alias VARCHAR(255),
    entities_used JSONB,
    example_queries JSONB,
    expected_results_schema JSONB,
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    embedding vector,
    embedding_hash VARCHAR(64),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN DEFAULT true,
    createdby VARCHAR(255) NOT NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP,
    rowversionstamp INTEGER DEFAULT 1,
    CONSTRAINT uq_query_template UNIQUE(config_id, template_name)
);

CREATE INDEX idx_query_template_config ON query_templates(config_id);
CREATE INDEX idx_query_template_category ON query_templates(template_category);
CREATE INDEX idx_query_template_active ON query_templates(active) WHERE active = true;
CREATE INDEX idx_query_template_db ON query_templates(database_alias);
CREATE INDEX idx_query_template_usage ON query_templates(usage_count DESC);
CREATE INDEX idx_query_template_generated ON query_templates(generated_at DESC);
CREATE INDEX idx_query_template_uuid ON query_templates(uuid);

-- Table: embedding_generation_jobs
-- Tracks batch embedding generation jobs
CREATE TABLE IF NOT EXISTS embedding_generation_jobs (
    id INTEGER PRIMARY KEY DEFAULT nextval('embedding_job_id_seq'),
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    referenceid VARCHAR(255) UNIQUE,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    job_type VARCHAR(50) NOT NULL,
    database_alias VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_items INTEGER,
    processed_items INTEGER DEFAULT 0,
    failed_items INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    active BOOLEAN DEFAULT true,
    createdby VARCHAR(255) NOT NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP,
    rowversionstamp INTEGER DEFAULT 1
);

CREATE INDEX idx_emb_job_config ON embedding_generation_jobs(config_id);
CREATE INDEX idx_emb_job_status ON embedding_generation_jobs(status);
CREATE INDEX idx_emb_job_created ON embedding_generation_jobs(createdon DESC);
CREATE INDEX idx_emb_job_uuid ON embedding_generation_jobs(uuid);

-- Table: embedding_search_logs
-- Logs vector similarity searches for analytics and improvement
CREATE TABLE IF NOT EXISTS embedding_search_logs (
    id INTEGER PRIMARY KEY DEFAULT nextval('embedding_search_log_id_seq'),
    uuid UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    referenceid VARCHAR(255) UNIQUE,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    search_type VARCHAR(50) NOT NULL,
    search_query TEXT NOT NULL,
    search_vector vector,
    results_count INTEGER,
    top_results JSONB,
    search_duration_ms INTEGER,
    user_feedback VARCHAR(50),
    active BOOLEAN DEFAULT true,
    createdby VARCHAR(255) NOT NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP,
    rowversionstamp INTEGER DEFAULT 1
);

CREATE INDEX idx_search_log_config ON embedding_search_logs(config_id);
CREATE INDEX idx_search_log_type ON embedding_search_logs(search_type);
CREATE INDEX idx_search_log_created ON embedding_search_logs(createdon DESC);
CREATE INDEX idx_search_log_uuid ON embedding_search_logs(uuid);

-- Function: Update modifiedon and rowversionstamp
CREATE OR REPLACE FUNCTION update_modified_columns()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modifiedon = CURRENT_TIMESTAMP;
    NEW.rowversionstamp = OLD.rowversionstamp + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for modified columns
CREATE TRIGGER update_ai_embedding_config_modified
    BEFORE UPDATE ON ai_embedding_configurations
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_columns();

CREATE TRIGGER update_db_schema_emb_modified
    BEFORE UPDATE ON database_schema_embeddings
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_columns();

CREATE TRIGGER update_business_entities_modified
    BEFORE UPDATE ON business_entities
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_columns();

CREATE TRIGGER update_query_templates_modified
    BEFORE UPDATE ON query_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_columns();

CREATE TRIGGER update_emb_jobs_modified
    BEFORE UPDATE ON embedding_generation_jobs
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_columns();

CREATE TRIGGER update_search_logs_modified
    BEFORE UPDATE ON embedding_search_logs
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_columns();

-- Sample data for testing
INSERT INTO ai_embedding_configurations (id, config_name, embedding_model, embedding_dimensions, vector_database_type, createdby)
VALUES
    (1, 'OpenAI-Default', 'text-embedding-ada-002', 1536, 'postgresql', 'system'),
    (2, 'OpenAI-Small', 'text-embedding-3-small', 1536, 'postgresql', 'system'),
    (3, 'OpenAI-Large', 'text-embedding-3-large', 3072, 'postgresql', 'system')
ON CONFLICT (id) DO NOTHING;

-- Update sequences
SELECT setval('ai_embedding_config_id_seq', (SELECT MAX(id) FROM ai_embedding_configurations), true);

-- Views for easy querying

-- View: Active embedding configurations with stats
CREATE OR REPLACE VIEW v_embedding_configurations_stats AS
SELECT
    c.id,
    c.uuid,
    c.config_name,
    c.embedding_model,
    c.embedding_dimensions,
    c.vector_database_type,
    c.active,
    c.createdon,
    COUNT(DISTINCT s.database_alias) as databases_with_embeddings,
    COUNT(DISTINCT s.table_name) as tables_with_embeddings,
    COUNT(s.id) as total_schema_embeddings,
    COUNT(DISTINCT b.id) as business_entities_count,
    COUNT(DISTINCT q.id) as query_templates_count,
    MAX(s.generated_at) as last_schema_embedding_generated,
    MAX(b.generated_at) as last_entity_generated,
    MAX(q.generated_at) as last_template_generated
FROM ai_embedding_configurations c
LEFT JOIN database_schema_embeddings s ON c.id = s.config_id AND s.active = true
LEFT JOIN business_entities b ON c.id = b.config_id AND b.active = true
LEFT JOIN query_templates q ON c.id = q.config_id AND q.active = true
GROUP BY c.id, c.uuid, c.config_name, c.embedding_model, c.embedding_dimensions, c.vector_database_type, c.active, c.createdon;

-- View: Database schema metadata summary
CREATE OR REPLACE VIEW v_database_schema_metadata AS
SELECT
    database_alias,
    schema_name,
    table_name,
    COUNT(CASE WHEN column_name IS NULL THEN 1 END) as has_table_description,
    COUNT(CASE WHEN column_name IS NOT NULL THEN 1 END) as columns_with_descriptions,
    MAX(generated_at) as last_updated
FROM database_schema_embeddings
WHERE active = true
GROUP BY database_alias, schema_name, table_name;

COMMENT ON TABLE ai_embedding_configurations IS 'Stores AI configuration metadata and embedding model information (IAC Standard)';
COMMENT ON TABLE database_schema_embeddings IS 'Stores vector embeddings for database schema metadata (IAC Standard)';
COMMENT ON TABLE business_entities IS 'Stores business entity definitions with vector embeddings (IAC Standard)';
COMMENT ON TABLE query_templates IS 'Stores SQL query templates with vector embeddings (IAC Standard)';
COMMENT ON TABLE embedding_generation_jobs IS 'Tracks batch embedding generation jobs (IAC Standard)';
COMMENT ON TABLE embedding_search_logs IS 'Logs vector similarity searches for analytics (IAC Standard)';
