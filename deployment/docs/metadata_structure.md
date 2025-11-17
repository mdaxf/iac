# Package Metadata Structure

The `iacpackages.metadata` field stores comprehensive information about packaged entities and selection criteria in JSON format.

## Database Package Metadata

```json
{
  "packaged_entities": [
    {
      "name": "users",
      "type": "table",
      "row_count": 100,
      "column_count": 8,
      "pk_columns": ["id"],
      "fk_count": 0,
      "pk_strategy": "auto_increment",
      "where_clause": "created_at > '2024-01-01'",
      "excluded_columns": ["password_hash", "ssn"]
    },
    {
      "name": "sessions",
      "type": "table",
      "row_count": 50,
      "column_count": 4,
      "pk_columns": ["session_id"],
      "fk_count": 1,
      "pk_strategy": "preserve",
      "where_clause": "expires_at > NOW()"
    },
    {
      "name": "orders",
      "type": "table",
      "row_count": 200,
      "column_count": 12,
      "pk_columns": ["id"],
      "fk_count": 2,
      "pk_strategy": "auto_increment",
      "where_clause": "status = 'active'"
    }
  ],
  "entity_count": 3,
  "total_records": 350,
  "total_relationships": 2,
  "selection_criteria": {
    "tables": ["users", "sessions", "orders"],
    "include_related": true,
    "max_depth": 2,
    "where_clauses": {
      "users": "created_at > '2024-01-01'",
      "sessions": "expires_at > NOW()",
      "orders": "status = 'active'"
    },
    "excluded_columns": {
      "users": ["password_hash", "ssn"],
      "customers": ["credit_card"]
    }
  },
  "include_parent_data": true,
  "has_dependencies": false
}
```

## Document Package Metadata

```json
{
  "packaged_entities": [
    {
      "name": "TranCode",
      "type": "collection",
      "document_count": 25,
      "index_count": 3,
      "id_field": "_id",
      "id_strategy": "regenerate",
      "query_filter": "{\"status\": \"active\"}",
      "excluded_fields": ["_internal_cache"]
    },
    {
      "name": "Workflow",
      "type": "collection",
      "document_count": 10,
      "index_count": 2,
      "id_field": "_id",
      "id_strategy": "regenerate",
      "query_filter": "{\"version\": {\"$gte\": 2}}"
    },
    {
      "name": "UI_Page",
      "type": "collection",
      "document_count": 15,
      "index_count": 2,
      "id_field": "_id",
      "id_strategy": "regenerate"
    }
  ],
  "entity_count": 3,
  "total_documents": 50,
  "total_references": 3,
  "selection_criteria": {
    "collections": ["TranCode", "Workflow", "UI_Page"],
    "query_filters": {
      "TranCode": "{\"status\": \"active\"}",
      "Workflow": "{\"version\": {\"$gte\": 2}}"
    },
    "excluded_fields": {
      "TranCode": ["_internal_cache"],
      "Workflow": ["temp_data"]
    }
  },
  "include_parent_data": false,
  "has_dependencies": true,
  "dependency_count": 1
}
```

## Metadata Fields

### Per-Entity Fields (packaged_entities array)

#### Database Tables:
- **name**: Table name
- **type**: Always "table"
- **row_count**: Number of rows packaged
- **column_count**: Number of columns in table
- **pk_columns**: Array of primary key column names
- **fk_count**: Number of foreign key constraints
- **pk_strategy**: "auto_increment", "sequence", or "preserve"
- **where_clause**: SQL WHERE condition used (if any)
- **excluded_columns**: Columns excluded from package (if any)

#### Document Collections:
- **name**: Collection name
- **type**: Always "collection"
- **document_count**: Number of documents packaged
- **index_count**: Number of indexes on collection
- **id_field**: ID field name (usually "_id")
- **id_strategy**: "regenerate", "skip", or "preserve"
- **query_filter**: MongoDB query filter used (if any)
- **excluded_fields**: Fields excluded from documents (if any)

### Aggregate Fields

- **entity_count**: Total number of tables/collections in package
- **total_records** (database): Total rows across all tables
- **total_documents** (document): Total documents across all collections
- **total_relationships** (database): Number of FK relationships
- **total_references** (document): Number of document references

### Selection Criteria

#### Database Packages:
```json
{
  "tables": ["table1", "table2"],
  "include_related": true,
  "max_depth": 2,
  "where_clauses": {
    "table1": "condition1",
    "table2": "condition2"
  },
  "excluded_columns": {
    "table1": ["col1", "col2"]
  }
}
```

#### Document Packages:
```json
{
  "collections": ["collection1", "collection2"],
  "query_filters": {
    "collection1": "{\"field\": \"value\"}",
    "collection2": "{\"status\": \"active\"}"
  },
  "excluded_fields": {
    "collection1": ["field1", "field2"]
  }
}
```

### Common Fields

- **include_parent_data**: Whether related parent data was auto-included
- **has_dependencies**: Whether package has dependencies
- **dependency_count**: Number of dependent packages

## Querying Metadata

### Find packages with specific tables:
```sql
SELECT * FROM iacpackages
WHERE JSON_EXTRACT(metadata, '$.selection_criteria.tables') LIKE '%users%';
```

### Find packages by entity count:
```sql
SELECT name, version,
       JSON_EXTRACT(metadata, '$.entity_count') as entity_count,
       JSON_EXTRACT(metadata, '$.total_records') as total_records
FROM iacpackages
WHERE JSON_EXTRACT(metadata, '$.entity_count') > 5;
```

### Find packages with WHERE clauses:
```sql
SELECT * FROM iacpackages
WHERE JSON_EXTRACT(metadata, '$.selection_criteria.where_clauses') IS NOT NULL;
```

### PostgreSQL JSONB queries:
```sql
-- Find packages containing specific table
SELECT * FROM iacpackages
WHERE metadata @> '{"selection_criteria": {"tables": ["users"]}}';

-- Find packages with more than 100 total records
SELECT * FROM iacpackages
WHERE (metadata->>'total_records')::int > 100;

-- Get all packaged entities
SELECT
    name,
    version,
    jsonb_array_elements(metadata->'packaged_entities')->>'name' as entity_name,
    jsonb_array_elements(metadata->'packaged_entities')->>'row_count' as row_count
FROM iacpackages
WHERE package_type = 'database';
```

## Use Cases

### 1. Audit Trail
Know exactly what was packaged:
```sql
SELECT
    p.name,
    p.version,
    p.created_by,
    p.created_at,
    JSON_EXTRACT(metadata, '$.packaged_entities[*].name') as tables_packaged
FROM iacpackages p
WHERE created_at > '2024-01-01';
```

### 2. Package Discovery
Find packages containing specific data:
```sql
-- Find all packages with user data
SELECT * FROM iacpackages
WHERE metadata LIKE '%users%';
```

### 3. Impact Analysis
Understand package scope:
```sql
SELECT
    name,
    version,
    JSON_EXTRACT(metadata, '$.entity_count') as tables,
    JSON_EXTRACT(metadata, '$.total_records') as records,
    JSON_EXTRACT(metadata, '$.total_relationships') as relationships
FROM iacpackages
ORDER BY JSON_EXTRACT(metadata, '$.total_records') DESC;
```

### 4. Compliance
Track excluded sensitive data:
```sql
SELECT
    name,
    version,
    JSON_EXTRACT(metadata, '$.selection_criteria.excluded_columns') as excluded_columns
FROM iacpackages
WHERE JSON_EXTRACT(metadata, '$.selection_criteria.excluded_columns') IS NOT NULL;
```

## Benefits

1. **Self-Documenting**: Packages include complete information about their contents
2. **Searchable**: Query packages by contents, filters, and criteria
3. **Auditable**: Track what data was packaged and how
4. **Reproducible**: Selection criteria allows recreating packages
5. **Discoverable**: Find relevant packages for deployment scenarios
6. **Compliant**: Track excluded sensitive data for compliance

## Example API Response

```json
{
  "package_id": "pkg-12345",
  "name": "UserManagement",
  "version": "1.0.0",
  "checksum": "abc123...",
  "file_size": 245862,
  "tables": 3,
  "records": 350,
  "metadata": {
    "packaged_entities": [...],
    "selection_criteria": {...},
    "total_records": 350,
    "entity_count": 3
  }
}
```
