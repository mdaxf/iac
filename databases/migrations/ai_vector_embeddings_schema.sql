-- AI Vector Embeddings Schema for PostgreSQL
-- Stores vector embeddings for schema metadata, business entities, and query templates
-- Supports multiple AI configurations and embedding models

-- Enable pgvector extension for vector operations
CREATE EXTENSION IF NOT EXISTS vector;

-- Table: ai_embedding_configurations
-- Stores AI configuration metadata and embedding model information
CREATE TABLE IF NOT EXISTS ai_embedding_configurations (
    id INTEGER PRIMARY KEY,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    referenceid VARCHAR(255) UNIQUE,
    config_name VARCHAR(255) NOT NULL UNIQUE,
    embedding_model VARCHAR(255) NOT NULL,
    embedding_dimensions INTEGER NOT NULL,
    vector_database_type VARCHAR(50) NOT NULL DEFAULT 'postgresql', -- 'postgresql', 'pinecone', 'chromadb', etc.
    vector_database_config JSONB, -- Connection details for external vector DBs
    active BOOLEAN DEFAULT true,
    createdby VARCHAR(255) NOT NULL,
    createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modifiedby VARCHAR(255),
    modifiedon TIMESTAMP,
    rowversionstamp INTEGER DEFAULT 1,
    CONSTRAINT chk_embedding_dimensions CHECK (embedding_dimensions > 0)
);

-- Index for active configurations
CREATE INDEX idx_ai_embedding_config_active ON ai_embedding_configurations(is_active) WHERE is_active = true;

-- Table: database_schema_embeddings
-- Stores vector embeddings for database tables and columns metadata
CREATE TABLE IF NOT EXISTS database_schema_embeddings (
    id SERIAL PRIMARY KEY,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    database_alias VARCHAR(255) NOT NULL, -- Reference to database connection alias
    schema_name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    column_name VARCHAR(255), -- NULL if embedding is for entire table
    description TEXT,
    metadata JSONB, -- Additional metadata (data_type, constraints, etc.)
    embedding vector, -- Vector embedding (dimensions based on config)
    embedding_hash VARCHAR(64), -- SHA256 hash of content for change detection
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    CONSTRAINT uq_schema_embedding UNIQUE(config_id, database_alias, schema_name, table_name, column_name)
);

-- Indexes for database schema embeddings
CREATE INDEX idx_db_schema_emb_config ON database_schema_embeddings(config_id);
CREATE INDEX idx_db_schema_emb_db_alias ON database_schema_embeddings(database_alias);
CREATE INDEX idx_db_schema_emb_table ON database_schema_embeddings(schema_name, table_name);
CREATE INDEX idx_db_schema_emb_generated ON database_schema_embeddings(generated_at DESC);

-- HNSW index for vector similarity search (requires pgvector)
-- Note: Dimension will vary by model, this is an example for 1536-dim embeddings (OpenAI)
-- CREATE INDEX idx_db_schema_emb_vector ON database_schema_embeddings USING hnsw (embedding vector_cosine_ops);

-- Table: business_entities
-- Stores business entity definitions with vector embeddings for natural language search
CREATE TABLE IF NOT EXISTS business_entities (
    id SERIAL PRIMARY KEY,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    entity_name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100), -- 'table', 'view', 'virtual', 'composite'
    description TEXT NOT NULL,
    database_alias VARCHAR(255), -- Reference to database if entity maps to DB
    schema_name VARCHAR(255),
    table_name VARCHAR(255),
    field_mappings JSONB, -- Maps logical field names to DB columns
    relationships JSONB, -- Defines relationships with other entities
    business_rules JSONB, -- Business logic and validation rules
    metadata JSONB, -- Additional metadata
    embedding vector, -- Vector embedding for natural language search
    embedding_hash VARCHAR(64),
    is_active BOOLEAN DEFAULT true,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    CONSTRAINT uq_business_entity UNIQUE(config_id, entity_name)
);

-- Indexes for business entities
CREATE INDEX idx_business_entity_config ON business_entities(config_id);
CREATE INDEX idx_business_entity_type ON business_entities(entity_type);
CREATE INDEX idx_business_entity_active ON business_entities(is_active) WHERE is_active = true;
CREATE INDEX idx_business_entity_db ON business_entities(database_alias, schema_name, table_name);
CREATE INDEX idx_business_entity_generated ON business_entities(generated_at DESC);

-- HNSW index for vector similarity search
-- CREATE INDEX idx_business_entity_vector ON business_entities USING hnsw (embedding vector_cosine_ops);

-- Table: query_templates
-- Stores query templates with vector embeddings for natural language to SQL conversion
CREATE TABLE IF NOT EXISTS query_templates (
    id SERIAL PRIMARY KEY,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    template_name VARCHAR(255) NOT NULL,
    template_category VARCHAR(100), -- 'select', 'insert', 'update', 'delete', 'report', 'analytics'
    natural_language_query TEXT NOT NULL, -- User-friendly description
    sql_template TEXT NOT NULL, -- SQL template with placeholders
    parameters JSONB, -- Parameter definitions and types
    database_alias VARCHAR(255), -- Target database
    entities_used JSONB, -- List of business entities used in query
    example_queries JSONB, -- Example natural language queries
    expected_results_schema JSONB, -- Schema of expected results
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP,
    embedding vector, -- Vector embedding for NL query matching
    embedding_hash VARCHAR(64),
    is_active BOOLEAN DEFAULT true,
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    updated_by VARCHAR(255),
    CONSTRAINT uq_query_template UNIQUE(config_id, template_name)
);

-- Indexes for query templates
CREATE INDEX idx_query_template_config ON query_templates(config_id);
CREATE INDEX idx_query_template_category ON query_templates(template_category);
CREATE INDEX idx_query_template_active ON query_templates(is_active) WHERE is_active = true;
CREATE INDEX idx_query_template_db ON query_templates(database_alias);
CREATE INDEX idx_query_template_usage ON query_templates(usage_count DESC);
CREATE INDEX idx_query_template_generated ON query_templates(generated_at DESC);

-- HNSW index for vector similarity search
-- CREATE INDEX idx_query_template_vector ON query_templates USING hnsw (embedding vector_cosine_ops);

-- Table: embedding_generation_jobs
-- Tracks batch embedding generation jobs
CREATE TABLE IF NOT EXISTS embedding_generation_jobs (
    id SERIAL PRIMARY KEY,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    job_type VARCHAR(50) NOT NULL, -- 'schema_metadata', 'business_entities', 'query_templates'
    database_alias VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'running', 'completed', 'failed'
    total_items INTEGER,
    processed_items INTEGER DEFAULT 0,
    failed_items INTEGER DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- Indexes for embedding generation jobs
CREATE INDEX idx_emb_job_config ON embedding_generation_jobs(config_id);
CREATE INDEX idx_emb_job_status ON embedding_generation_jobs(status);
CREATE INDEX idx_emb_job_created ON embedding_generation_jobs(created_at DESC);

-- Table: embedding_search_logs
-- Logs vector similarity searches for analytics and improvement
CREATE TABLE IF NOT EXISTS embedding_search_logs (
    id SERIAL PRIMARY KEY,
    config_id INTEGER NOT NULL REFERENCES ai_embedding_configurations(id) ON DELETE CASCADE,
    search_type VARCHAR(50) NOT NULL, -- 'schema', 'entity', 'query'
    search_query TEXT NOT NULL,
    search_vector vector,
    results_count INTEGER,
    top_results JSONB, -- Top N results with similarity scores
    search_duration_ms INTEGER,
    user_feedback VARCHAR(50), -- 'helpful', 'not_helpful'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255)
);

-- Indexes for search logs
CREATE INDEX idx_search_log_config ON embedding_search_logs(config_id);
CREATE INDEX idx_search_log_type ON embedding_search_logs(search_type);
CREATE INDEX idx_search_log_created ON embedding_search_logs(created_at DESC);

-- Function: Update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for updated_at
CREATE TRIGGER update_ai_embedding_config_updated_at
    BEFORE UPDATE ON ai_embedding_configurations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_business_entities_updated_at
    BEFORE UPDATE ON business_entities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_query_templates_updated_at
    BEFORE UPDATE ON query_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Sample data for testing
INSERT INTO ai_embedding_configurations (config_name, embedding_model, embedding_dimensions, vector_database_type, created_by)
VALUES
    ('OpenAI-Default', 'text-embedding-ada-002', 1536, 'postgresql', 'system'),
    ('OpenAI-Small', 'text-embedding-3-small', 1536, 'postgresql', 'system'),
    ('OpenAI-Large', 'text-embedding-3-large', 3072, 'postgresql', 'system')
ON CONFLICT (config_name) DO NOTHING;

-- Views for easy querying

-- View: Active embedding configurations with stats
CREATE OR REPLACE VIEW v_embedding_configurations_stats AS
SELECT
    c.id,
    c.config_name,
    c.embedding_model,
    c.embedding_dimensions,
    c.vector_database_type,
    c.is_active,
    c.created_at,
    COUNT(DISTINCT s.database_alias) as databases_with_embeddings,
    COUNT(DISTINCT s.table_name) as tables_with_embeddings,
    COUNT(s.id) as total_schema_embeddings,
    COUNT(DISTINCT b.id) as business_entities_count,
    COUNT(DISTINCT q.id) as query_templates_count,
    MAX(s.generated_at) as last_schema_embedding_generated,
    MAX(b.generated_at) as last_entity_generated,
    MAX(q.generated_at) as last_template_generated
FROM ai_embedding_configurations c
LEFT JOIN database_schema_embeddings s ON c.id = s.config_id
LEFT JOIN business_entities b ON c.id = b.config_id
LEFT JOIN query_templates q ON c.id = q.config_id
GROUP BY c.id, c.config_name, c.embedding_model, c.embedding_dimensions, c.vector_database_type, c.is_active, c.created_at;

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
GROUP BY database_alias, schema_name, table_name;

COMMENT ON TABLE ai_embedding_configurations IS 'Stores AI configuration metadata and embedding model information';
COMMENT ON TABLE database_schema_embeddings IS 'Stores vector embeddings for database schema metadata (tables and columns)';
COMMENT ON TABLE business_entities IS 'Stores business entity definitions with vector embeddings';
COMMENT ON TABLE query_templates IS 'Stores SQL query templates with vector embeddings for NL search';
COMMENT ON TABLE embedding_generation_jobs IS 'Tracks batch embedding generation jobs';
COMMENT ON TABLE embedding_search_logs IS 'Logs vector similarity searches for analytics';
