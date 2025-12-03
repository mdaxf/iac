-- Enable pgvector extension in vec_iac database
-- Run this script to fix the "type 'vector' does not exist" error

-- Connect to vec_iac database
\c vec_iac

-- Enable pgvector extension (requires superuser privileges)
CREATE EXTENSION IF NOT EXISTS vector;

-- Verify the extension is installed
SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';

-- Test vector operations
SELECT '[1,2,3]'::vector <=> '[1,1,1]'::vector AS test_distance;

-- Show success message
\echo '✅ pgvector extension enabled successfully in vec_iac database!'
