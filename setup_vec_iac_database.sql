-- Setup vec_iac database for vector embeddings
-- Run this script as a PostgreSQL superuser (postgres) to set up the vector database

-- Step 1: Create the database (if it doesn't exist)
-- Note: This command must be run separately or by connecting to postgres database first
-- CREATE DATABASE vec_iac OWNER postgres;

-- Step 2: Connect to vec_iac database and enable pgvector extension
-- Run: \c vec_iac

-- Enable pgvector extension at database level (requires superuser)
CREATE EXTENSION IF NOT EXISTS vector;

-- Create the schema
CREATE SCHEMA IF NOT EXISTS vec_iac;

-- Grant permissions to the application user (adjust if using different user)
GRANT ALL PRIVILEGES ON SCHEMA vec_iac TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA vec_iac TO postgres;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA vec_iac TO postgres;
GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA vec_iac TO postgres;

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA vec_iac GRANT ALL PRIVILEGES ON TABLES TO postgres;
ALTER DEFAULT PRIVILEGES IN SCHEMA vec_iac GRANT ALL PRIVILEGES ON SEQUENCES TO postgres;
ALTER DEFAULT PRIVILEGES IN SCHEMA vec_iac GRANT ALL PRIVILEGES ON FUNCTIONS TO postgres;

-- Verify extension is installed
SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';

-- Verify schema exists
SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'vec_iac';

-- Set search path for current session
SET search_path TO vec_iac, public;

COMMENT ON DATABASE vec_iac IS 'Vector embeddings database for IAC system using pgvector';
COMMENT ON SCHEMA vec_iac IS 'Schema for storing vector embeddings and related tables';
