# Schema Embedding Service Refactoring Plan

## Overview

The `SchemaEmbeddingService` currently tries to store embeddings directly in the `databaseschemametadata` table, but we need to refactor it to use the separate `database_schema_embeddings` table instead.

## Current Problem

```go
// Line 316 - This query fails because embedding column doesn't exist
err := s.DB.Where("databasealias = ? AND metadatatype = ? AND (embedding IS NULL OR embedding_generated_at IS NULL)",
    databaseAlias, models.MetadataTypeTable).Find(&tableMetadata).Error
```

**Error**: `pq: column "embedding" does not exist`

## Required Changes

### 1. Service Structure

Add a `ConfigID` field to the service (needed for foreign key in vector tables):

```go
type SchemaEmbeddingService struct {
    DB        *gorm.DB
    OpenAIKey string
    ModelName string
    ConfigID  int    // NEW: Default config ID for embeddings
    iLog      logger.Log
}
```

### 2. Constructor

Update `NewSchemaEmbeddingService` to get or create a default config:

```go
func NewSchemaEmbeddingService(db *gorm.DB, openAIKey string) *SchemaEmbeddingService {
    service := &SchemaEmbeddingService{
        DB:        db,
        OpenAIKey: openAIKey,
        ModelName: "text-embedding-ada-002",
        iLog: logger.Log{
            ModuleName:     logger.Framework,
            User:           "System",
            ControllerName: "SchemaEmbeddingService",
        },
    }

    // Get or create default config
    service.ConfigID = service.getOrCreateDefaultConfig()

    return service
}

func (s *SchemaEmbeddingService) getOrCreateDefaultConfig() int {
    var config models.AIEmbeddingConfiguration

    // Try to find existing config
    err := s.DB.Where("config_name = ?", "default").First(&config).Error
    if err == nil {
        return config.ID
    }

    // Create new default config
    config = models.AIEmbeddingConfiguration{
        ConfigName:          "default",
        EmbeddingModel:      s.ModelName,
        EmbeddingDimensions: 1536,
        VectorDatabaseType:  "postgresql",
        Active:              true,
        CreatedBy:           "System",
    }

    if err := s.DB.Create(&config).Error; err != nil {
        s.iLog.Error(fmt.Sprintf("Failed to create default config: %v", err))
        return 1 // Fallback to ID 1
    }

    return config.ID
}
```

### 3. GenerateEmbeddingsForDatabase Method

Replace the entire method with this logic:

```go
func (s *SchemaEmbeddingService) GenerateEmbeddingsForDatabase(ctx context.Context, databaseAlias string) error {
    s.iLog.Info(fmt.Sprintf("Starting batch embedding generation for database: %s", databaseAlias))

    // STEP 1: Get all tables from databaseschemametadata
    var tableMetadata []models.DatabaseSchemaMetadata
    err := s.DB.Where("databasealias = ? AND metadatatype = ?",
        databaseAlias, models.MetadataTypeTable).Find(&tableMetadata).Error
    if err != nil {
        return fmt.Errorf("failed to fetch table metadata: %w", err)
    }

    s.iLog.Info(fmt.Sprintf("Found %d tables", len(tableMetadata)))

    // STEP 2: Check which tables already have embeddings
    for i, meta := range tableMetadata {
        // Check if embedding already exists
        var existingEmbedding models.DatabaseSchemaEmbedding
        err := s.DB.Where("config_id = ? AND database_alias = ? AND table_name = ? AND column_name IS NULL",
            s.ConfigID, databaseAlias, meta.Table).First(&existingEmbedding).Error

        if err == gorm.ErrRecordNotFound {
            // Generate new embedding
            s.iLog.Debug(fmt.Sprintf("Generating embedding for table %d/%d: %s", i+1, len(tableMetadata), meta.Table))
            if err := s.GenerateAndStoreTableEmbedding(ctx, databaseAlias, &meta); err != nil {
                s.iLog.Error(fmt.Sprintf("Failed to generate embedding for table %s: %v", meta.Table, err))
                continue
            }
            time.Sleep(100 * time.Millisecond) // Rate limiting
        } else if err != nil {
            s.iLog.Error(fmt.Sprintf("Error checking existing embedding: %v", err))
            continue
        } else {
            s.iLog.Debug(fmt.Sprintf("Table %s already has embedding, skipping", meta.Table))
        }
    }

    // STEP 3: Same for columns
    var columnMetadata []models.DatabaseSchemaMetadata
    err = s.DB.Where("databasealias = ? AND metadatatype = ?",
        databaseAlias, models.MetadataTypeColumn).Find(&columnMetadata).Error
    if err != nil {
        return fmt.Errorf("failed to fetch column metadata: %w", err)
    }

    s.iLog.Info(fmt.Sprintf("Found %d columns", len(columnMetadata)))

    for i, meta := range columnMetadata {
        var existingEmbedding models.DatabaseSchemaEmbedding
        err := s.DB.Where("config_id = ? AND database_alias = ? AND table_name = ? AND column_name = ?",
            s.ConfigID, databaseAlias, meta.Table, meta.Column).First(&existingEmbedding).Error

        if err == gorm.ErrRecordNotFound {
            s.iLog.Debug(fmt.Sprintf("Generating embedding for column %d/%d: %s.%s", i+1, len(columnMetadata), meta.Table, meta.Column))
            if err := s.GenerateAndStoreColumnEmbedding(ctx, databaseAlias, &meta); err != nil {
                s.iLog.Error(fmt.Sprintf("Failed to generate embedding for column %s.%s: %v", meta.Table, meta.Column, err))
                continue
            }
            time.Sleep(100 * time.Millisecond)
        } else if err != nil {
            s.iLog.Error(fmt.Sprintf("Error checking existing embedding: %v", err))
            continue
        }
    }

    s.iLog.Info(fmt.Sprintf("Batch embedding generation completed for database: %s", databaseAlias))
    return nil
}
```

### 4. New Helper Methods

Create new methods that store embeddings in the vector tables:

```go
func (s *SchemaEmbeddingService) GenerateAndStoreTableEmbedding(ctx context.Context, databaseAlias string, meta *models.DatabaseSchemaMetadata) error {
    // Build description text for embedding
    text := fmt.Sprintf("Table: %s", meta.Table)
    if meta.Description != "" {
        text += fmt.Sprintf("\nDescription: %s", meta.Description)
    }
    if meta.BusinessName != "" {
        text += fmt.Sprintf("\nBusiness Name: %s", meta.BusinessName)
    }

    // Generate embedding
    embedding, err := s.GenerateEmbedding(ctx, text)
    if err != nil {
        return err
    }

    // Convert to Vector type
    vectorData := make(models.Vector, len(embedding))
    for i, v := range embedding {
        vectorData[i] = float32(v)
    }

    // Create embedding record
    embeddingRecord := models.DatabaseSchemaEmbedding{
        ConfigID:        s.ConfigID,
        DatabaseAlias:   databaseAlias,
        SchemaName:      meta.SchemaName,
        MappedTableName: meta.Table,
        ColumnName:      nil,
        Description:     meta.Description,
        Embedding:       vectorData,
        GeneratedAt:     time.Now(),
        Active:          true,
        CreatedBy:       "System",
    }

    return s.DB.Create(&embeddingRecord).Error
}

func (s *SchemaEmbeddingService) GenerateAndStoreColumnEmbedding(ctx context.Context, databaseAlias string, meta *models.DatabaseSchemaMetadata) error {
    // Build description text for embedding
    text := fmt.Sprintf("Table: %s, Column: %s, Type: %s", meta.Table, meta.Column, meta.DataType)
    if meta.Description != "" {
        text += fmt.Sprintf("\nDescription: %s", meta.Description)
    }
    if meta.ColumnComment != "" {
        text += fmt.Sprintf("\nComment: %s", meta.ColumnComment)
    }

    // Generate embedding
    embedding, err := s.GenerateEmbedding(ctx, text)
    if err != nil {
        return err
    }

    // Convert to Vector type
    vectorData := make(models.Vector, len(embedding))
    for i, v := range embedding {
        vectorData[i] = float32(v)
    }

    columnName := meta.Column
    embeddingRecord := models.DatabaseSchemaEmbedding{
        ConfigID:        s.ConfigID,
        DatabaseAlias:   databaseAlias,
        SchemaName:      meta.SchemaName,
        MappedTableName: meta.Table,
        ColumnName:      &columnName,
        Description:     meta.Description,
        Embedding:       vectorData,
        GeneratedAt:     time.Now(),
        Active:          true,
        CreatedBy:       "System",
    }

    return s.DB.Create(&embeddingRecord).Error
}
```

### 5. Search Methods

Update search methods to query `database_schema_embeddings` and join with `databaseschemametadata`:

```go
func (s *SchemaEmbeddingService) SearchSimilarTables(ctx context.Context, databaseAlias, query string, limit int) ([]models.DatabaseSchemaMetadata, error) {
    // Generate query embedding
    queryEmbedding, err := s.GenerateEmbedding(ctx, query)
    if err != nil {
        return nil, err
    }

    // Convert to Vector type
    vectorData := make(models.Vector, len(queryEmbedding))
    for i, v := range queryEmbedding {
        vectorData[i] = float32(v)
    }

    // Search using cosine similarity in database_schema_embeddings
    var embeddings []models.DatabaseSchemaEmbedding
    err = s.DB.Where("database_alias = ? AND column_name IS NULL AND active = true", databaseAlias).
        Order(gorm.Expr("embedding <-> ?::vector", vectorData)).
        Limit(limit).
        Find(&embeddings).Error

    if err != nil {
        return nil, err
    }

    // Get the actual metadata for these tables
    var results []models.DatabaseSchemaMetadata
    for _, emb := range embeddings {
        var meta models.DatabaseSchemaMetadata
        err := s.DB.Where("databasealias = ? AND tablename = ? AND metadatatype = ?",
            emb.DatabaseAlias, emb.MappedTableName, models.MetadataTypeTable).
            First(&meta).Error

        if err == nil {
            results = append(results, meta)
        }
    }

    return results, nil
}
```

## Migration Steps

1. ✅ Remove embedding fields from `DatabaseSchemaMetadata` model
2. ✅ Ensure `database_schema_embeddings` table exists (run migration SQL)
3. ⏳ Update `SchemaEmbeddingService` with new methods
4. ⏳ Test embedding generation
5. ⏳ Test embedding search

## Files to Modify

- ✅ `models/chat.go` - Remove embedding fields from DatabaseSchemaMetadata
- ⏳ `services/schemaembeddingservice.go` - Refactor all methods
- ⏳ Test the endpoints after changes

## Testing

```bash
# 1. Generate embeddings
curl -X POST http://localhost:8080/api/ai/embeddings/generate \
  -H "Content-Type: application/json" \
  -d '{"databasealias": "default"}'

# 2. Check status
curl http://localhost:8080/api/ai/embeddings/status/default

# 3. Search tables
curl -X POST http://localhost:8080/api/ai/embeddings/search/tables \
  -H "Content-Type: application/json" \
  -d '{"databasealias": "default", "query": "user information", "limit": 5}'
```

## Notes

- The vector database tables use `config_id` as foreign key
- Need to ensure a default config exists before generating embeddings
- Rate limiting (100ms between API calls) is important for OpenAI API
- Use pgvector's `<->` operator for cosine distance in PostgreSQL
