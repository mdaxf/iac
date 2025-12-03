-- Initialize vec_iac database schema
-- Run this script on the vec_iac database to create the schema and tables

-- Create schema
CREATE SCHEMA IF NOT EXISTS vec_iac;

-- Set search path
SET search_path TO vec_iac;

-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Note: The actual table creation will be handled by the application
-- via the vectordb_schema.sql embedded in the binary.
-- This script just ensures the schema exists and pgvector is enabled.

-- Verify schema creation
SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'vec_iac';

-- Grant necessary permissions (adjust username as needed)
-- GRANT ALL PRIVILEGES ON SCHEMA vec_iac TO postgres;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA vec_iac TO postgres;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA vec_iac TO postgres;

COMMENT ON SCHEMA vec_iac IS 'Vector embeddings database schema for IAC system';
