-- ===================================================================
-- Vector Embeddings Schema Migration for AI-powered Semantic Search
-- ===================================================================
-- Description: Adds vector embedding support to schema metadata and
-- business entities tables for improved AI chat report generation
--
-- Features:
-- 1. Vector columns for semantic search (MySQL 8.0.36+ / 8.4 LTS)
-- 2. Vector indexes for fast similarity search
-- 3. Support for OpenAI text-embedding-ada-002 (1536 dimensions)
-- ===================================================================

-- Add vector embedding column to databaseschemametadata table
ALTER TABLE databaseschemametadata
  ADD COLUMN embedding VECTOR(1536) NULL COMMENT 'Vector embedding for semantic search';

-- Create vector index for cosine similarity search on schema metadata
ALTER TABLE databaseschemametadata
  ADD VECTOR INDEX idx_schemametadata_embedding (embedding)
  WITH DISTANCE_METRIC = COSINE;

-- Add vector embedding column to businessentities table
ALTER TABLE businessentities
  ADD COLUMN embedding VECTOR(1536) NULL COMMENT 'Vector embedding for semantic search';

-- Create vector index for cosine similarity search on business entities
ALTER TABLE businessentities
  ADD VECTOR INDEX idx_businessentities_embedding (embedding)
  WITH DISTANCE_METRIC = COSINE;

-- Add embedding metadata columns to track generation
ALTER TABLE databaseschemametadata
  ADD COLUMN embedding_model VARCHAR(100) NULL COMMENT 'Model used for embedding generation',
  ADD COLUMN embedding_generated_at TIMESTAMP NULL COMMENT 'When embedding was generated';

ALTER TABLE businessentities
  ADD COLUMN embedding_model VARCHAR(100) NULL COMMENT 'Model used for embedding generation',
  ADD COLUMN embedding_generated_at TIMESTAMP NULL COMMENT 'When embedding was generated';

-- ===================================================================
-- Notes:
-- 1. Requires MySQL 8.0.36+ or 8.4 LTS for VECTOR type support
-- 2. Vector dimension = 1536 (OpenAI text-embedding-ada-002)
-- 3. Using COSINE distance metric for similarity search
-- 4. Existing rows will have NULL embeddings until populated
-- 5. Use batch processing to generate embeddings for existing data
-- ===================================================================

-- Example query for vector similarity search (after embeddings are populated):
-- SELECT id, tablename, columnname, description,
--        embedding <-> CAST('[query_vector]' AS JSON) AS distance
-- FROM databaseschemametadata
-- WHERE databasealias = 'your_alias'
-- AND embedding IS NOT NULL
-- ORDER BY distance ASC
-- LIMIT 10;
