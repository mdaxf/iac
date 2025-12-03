-- ===================================================================
-- AI Vector Database Schema Migration (PostgreSQL with pgvector)
-- ===================================================================
-- Description: Creates tables in the VECTOR database for AI services
-- Features:
-- 1. Database schema metadata with embeddings
-- 2. Business entities with embeddings
-- 3. Query templates with embeddings
-- 4. All AI-related data stored in vector database
-- 5. Main database only stores business data and reports
-- ===================================================================

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create enum types
DO $$ BEGIN
  CREATE TYPE metadata_type AS ENUM ('table', 'column');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE business_entity_type AS ENUM ('entity', 'metric', 'dimension');
EXCEPTION
  WHEN duplicate_object THEN null;
END $$;

-- ===================================================================
-- 1. Database Schema Metadata Table
-- ===================================================================
CREATE TABLE IF NOT EXISTS databaseschemametadata (
  id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  databasealias VARCHAR(100) NOT NULL,
  schemaname VARCHAR(100),
  tablename VARCHAR(100) NOT NULL,
  columnname VARCHAR(100),
  datatype VARCHAR(50),
  isnullable BOOLEAN,
  is_primary_key BOOLEAN,
  is_foreign_key BOOLEAN,
  columncomment TEXT,
  samplevalues JSONB,
  metadatatype metadata_type NOT NULL,
  description TEXT,
  business_name VARCHAR(255),
  businessterms JSONB,

  -- Vector embedding fields for semantic search
  embedding vector(1536),  -- OpenAI text-embedding-ada-002
  embedding_model VARCHAR(100),
  embedding_generated_at TIMESTAMP,

  -- Standard IAC audit fields
  active BOOLEAN DEFAULT TRUE,
  referenceid VARCHAR(36),
  createdby VARCHAR(45),
  createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  modifiedby VARCHAR(45),
  modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  rowversionstamp INT DEFAULT 1
);

-- Indexes for query performance
CREATE INDEX IF NOT EXISTS idx_databaseschemametadata_databasealias ON databaseschemametadata(databasealias);
CREATE INDEX IF NOT EXISTS idx_databaseschemametadata_tablename ON databaseschemametadata(tablename);
CREATE INDEX IF NOT EXISTS idx_databaseschemametadata_metadatatype ON databaseschemametadata(metadatatype);
CREATE INDEX IF NOT EXISTS idx_databaseschemametadata_database_table ON databaseschemametadata(databasealias, tablename);

-- Vector similarity search index (HNSW for fast approximate nearest neighbor)
CREATE INDEX IF NOT EXISTS idx_databaseschemametadata_embedding
ON databaseschemametadata
USING hnsw (embedding vector_cosine_ops);

COMMENT ON TABLE databaseschemametadata IS 'Database schema metadata with embeddings for AI semantic search';

-- ===================================================================
-- 2. Business Entities Table
-- ===================================================================
CREATE TABLE IF NOT EXISTS businessentities (
  id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  databasealias VARCHAR(100) NOT NULL,
  entityname VARCHAR(255) NOT NULL,
  entitytype business_entity_type NOT NULL,
  description TEXT,
  tablemappings JSONB,
  columnmappings JSONB,
  calculationformula TEXT,
  synonyms JSONB,
  examples JSONB,

  -- Vector embedding fields for semantic search
  embedding vector(1536),  -- OpenAI text-embedding-ada-002
  embedding_model VARCHAR(100),
  embedding_generated_at TIMESTAMP,

  -- Standard IAC audit fields
  active BOOLEAN DEFAULT TRUE,
  referenceid VARCHAR(36),
  createdby VARCHAR(45),
  createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  modifiedby VARCHAR(45),
  modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  rowversionstamp INT DEFAULT 1
);

-- Indexes for query performance
CREATE INDEX IF NOT EXISTS idx_businessentities_databasealias ON businessentities(databasealias);
CREATE INDEX IF NOT EXISTS idx_businessentities_entityname ON businessentities(entityname);
CREATE INDEX IF NOT EXISTS idx_businessentities_entitytype ON businessentities(entitytype);

-- Vector similarity search index
CREATE INDEX IF NOT EXISTS idx_businessentities_embedding
ON businessentities
USING hnsw (embedding vector_cosine_ops);

COMMENT ON TABLE businessentities IS 'Business entities with embeddings for AI semantic understanding';

-- ===================================================================
-- 3. Query Templates Table
-- ===================================================================
CREATE TABLE IF NOT EXISTS querytemplates (
  id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
  databasealias VARCHAR(100) NOT NULL,
  templatename VARCHAR(255) NOT NULL,
  description TEXT,
  naturallanguagepattern TEXT,
  sqltemplate TEXT NOT NULL,
  examplequestions JSONB,
  parameters JSONB,
  category VARCHAR(100),
  usagecount INT DEFAULT 0,
  successrate DECIMAL(3,2),
  lastused TIMESTAMP,

  -- Vector embedding fields for semantic search
  embedding vector(1536),  -- OpenAI text-embedding-ada-002
  embedding_model VARCHAR(100),
  embedding_generated_at TIMESTAMP,

  -- Standard IAC audit fields
  active BOOLEAN DEFAULT TRUE,
  referenceid VARCHAR(36),
  createdby VARCHAR(45),
  createdon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  modifiedby VARCHAR(45),
  modifiedon TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  rowversionstamp INT DEFAULT 1
);

-- Indexes for query performance
CREATE INDEX IF NOT EXISTS idx_querytemplates_databasealias ON querytemplates(databasealias);
CREATE INDEX IF NOT EXISTS idx_querytemplates_templatename ON querytemplates(templatename);
CREATE INDEX IF NOT EXISTS idx_querytemplates_category ON querytemplates(category);
CREATE INDEX IF NOT EXISTS idx_querytemplates_usagecount ON querytemplates(usagecount DESC);

-- Vector similarity search index
CREATE INDEX IF NOT EXISTS idx_querytemplates_embedding
ON querytemplates
USING hnsw (embedding vector_cosine_ops);

COMMENT ON TABLE querytemplates IS 'Query templates with embeddings for AI intent matching';

-- ===================================================================
-- Triggers for automatic modifiedon update
-- ===================================================================
CREATE OR REPLACE FUNCTION update_modifiedon_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.modifiedon = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_databaseschemametadata_modifiedon
BEFORE UPDATE ON databaseschemametadata
FOR EACH ROW
EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_businessentities_modifiedon
BEFORE UPDATE ON businessentities
FOR EACH ROW
EXECUTE FUNCTION update_modifiedon_column();

CREATE TRIGGER update_querytemplates_modifiedon
BEFORE UPDATE ON querytemplates
FOR EACH ROW
EXECUTE FUNCTION update_modifiedon_column();

-- ===================================================================
-- Notes:
-- 1. Requires PostgreSQL 11+ with pgvector extension
-- 2. Vector dimension = 1536 (OpenAI text-embedding-ada-002)
-- 3. HNSW index for fast approximate nearest neighbor search
-- 4. All AI-related data stored in vector database
-- 5. Main database only stores business data and reports
-- 6. Schema metadata gets actual table/column info from target databases
-- 7. Embeddings enable semantic search for AI chat
-- ===================================================================
