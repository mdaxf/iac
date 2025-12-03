# Vector Database Setup Guide

## Error: "type 'vector' does not exist"

This error occurs when the PostgreSQL `pgvector` extension is not installed or not enabled in the database.

## Prerequisites

1. **PostgreSQL 11+** installed
2. **pgvector extension** installed on the PostgreSQL server

## Step-by-Step Setup

### 1. Install pgvector Extension (One-Time Setup)

#### On Windows:
```bash
# Download pgvector from: https://github.com/pgvector/pgvector/releases
# Or install via pgAdmin Extension Manager
# Or use pre-built binaries
```

#### On Linux/Ubuntu:
```bash
sudo apt install postgresql-15-pgvector
# Or for other versions: postgresql-<version>-pgvector
```

#### On macOS:
```bash
brew install pgvector
```

### 2. Create and Configure vec_iac Database

Connect to PostgreSQL as superuser:
```bash
psql -U postgres -d postgres
```

Run the setup script:
```sql
-- Create the database
CREATE DATABASE vec_iac OWNER postgres;

-- Connect to vec_iac
\c vec_iac

-- Enable pgvector extension (requires superuser)
CREATE EXTENSION IF NOT EXISTS vector;

-- Verify installation
SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';
```

Or run the provided SQL file:
```bash
psql -U postgres -d postgres -f setup_vec_iac_database.sql
```

### 3. Verify Connection String in aiconfig.json

Ensure your `aiconfig.json` has the correct configuration:

```json
{
  "vector_database": {
    "type": "postgres_pgvector",
    "postgres_pgvector": {
      "enabled": true,
      "use_main_db": false,
      "connection_string": "postgresql://postgres:PASSWORD@localhost:5432/vec_iac?sslmode=disable",
      "schema": "vec_iac",
      "table_prefix": "",
      "dimension": 3072
    }
  }
}
```

**Important**: The connection string should include `vec_iac` database name.

### 4. Initialize Application Tables

When the application starts, it will automatically create the necessary tables in the `vec_iac` schema:
- `database_schema_embeddings`
- `business_entities`
- `query_templates`
- `ai_embedding_configurations`
- `embedding_generation_jobs`
- `embedding_search_logs`

You can verify tables were created:
```sql
\c vec_iac
SET search_path TO vec_iac;
\dt
```

## Troubleshooting

### Error: "extension 'vector' does not exist"

**Solution**: pgvector is not installed on your PostgreSQL server. Follow Step 1 above.

### Error: "permission denied to create extension"

**Solution**: You need superuser privileges to create extensions. Connect as `postgres` user.

### Error: "relation does not exist"

**Solution**: Tables haven't been created. Ensure:
1. The application started successfully
2. Check logs for "Vector database schema initialized successfully"
3. Manually run `vectordb_schema.sql` if needed

### Error: "could not connect to database"

**Solution**: Check connection string in `aiconfig.json`:
- Correct host and port
- Database `vec_iac` exists
- Credentials are correct
- PostgreSQL service is running

## Manual Schema Creation

If automatic initialization fails, manually run:

```bash
# Connect to vec_iac database
psql -U postgres -d vec_iac

# Set search path
SET search_path TO vec_iac;

# Run the embedded schema
# Copy content from services/vectordb_schema.sql and execute
```

## Verification Checklist

✅ PostgreSQL is running
✅ pgvector extension is installed on server
✅ Database `vec_iac` exists
✅ Extension `vector` is enabled in `vec_iac` database
✅ Schema `vec_iac` exists
✅ Application can connect to `vec_iac` database
✅ Tables are created in `vec_iac` schema

## Check Current Status

Run these commands in PostgreSQL:

```sql
-- Check if pgvector is installed
SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';

-- Check if schema exists
SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'vec_iac';

-- Check tables in vec_iac schema
SELECT table_name FROM information_schema.tables
WHERE table_schema = 'vec_iac'
ORDER BY table_name;

-- Test vector type
CREATE TEMP TABLE test_vector (id int, embedding vector(3));
INSERT INTO test_vector VALUES (1, '[1,2,3]');
SELECT * FROM test_vector;
DROP TABLE test_vector;
```

If all commands succeed, your vector database is properly configured.

## Connection String Format

```
postgresql://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=disable&search_path=SCHEMA
```

Example:
```
postgresql://postgres:iacf12345678@localhost:5432/vec_iac?sslmode=disable&search_path=vec_iac
```

## Common Issues

### Issue: Tables created in public schema instead of vec_iac

**Cause**: `search_path` not set in connection string
**Solution**: Add `&search_path=vec_iac` to connection string

### Issue: "insufficient privileges for database vec_iac"

**Cause**: User doesn't have access to the database
**Solution**: Grant privileges:
```sql
GRANT ALL PRIVILEGES ON DATABASE vec_iac TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA vec_iac TO postgres;
```

### Issue: Old data in wrong database

**Cause**: Previously used main database for vectors
**Solution**:
1. Delete old embeddings: `DELETE FROM databaseschemametadata WHERE embedding IS NOT NULL;`
2. Regenerate in vector database using the application

## Support

If issues persist:
1. Check application logs for detailed error messages
2. Verify PostgreSQL version supports pgvector
3. Ensure firewall allows database connection
4. Test connection string manually with `psql`
