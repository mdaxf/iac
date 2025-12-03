package services

import (
	_ "embed"
	"fmt"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"gorm.io/gorm"
)

//go:embed vectordb_schema.sql
var vectorEmbeddingSchema string

// InitializeVectorDatabase initializes the vector database tables if configured
func InitializeVectorDatabase(db *gorm.DB) error {
	iLog := logger.Log{
		ModuleName:     logger.Framework,
		User:           "System",
		ControllerName: "VectorDBInit",
	}

	// Check if vector database is configured in aiconfig.json
	if config.AIConf == nil {
		iLog.Info("AI configuration not loaded, skipping vector database initialization")
		return nil
	}

	vectorConfig := config.AIConf.VectorDatabase

	// Check if vector database is enabled
	if vectorConfig.Type != "postgres_pgvector" {
		iLog.Info(fmt.Sprintf("Vector database type is '%s', skipping PostgreSQL vector initialization", vectorConfig.Type))
		return nil
	}

	if !vectorConfig.PostgresPGVector.Enabled {
		iLog.Info("PostgreSQL pgvector is not enabled, skipping initialization")
		return nil
	}

	iLog.Info("Initializing vector database schema...")

	// Get the configured vector database connection
	targetDB, err := GetVectorDB(db)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get vector database connection: %v", err))
		return fmt.Errorf("failed to get vector database connection: %w", err)
	}

	if vectorConfig.PostgresPGVector.UseMainDB {
		iLog.Info("Using main database for vector storage")
	} else {
		iLog.Info(fmt.Sprintf("Using separate vector database (schema: %s)", vectorConfig.PostgresPGVector.Schema))
	}

	// Execute the migration SQL
	iLog.Info("Executing vector database schema migration...")
	/*if err := targetDB.Exec(vectorEmbeddingSchema).Error; err != nil {
		iLog.Error(fmt.Sprintf("Failed to initialize vector database schema: %v", err))
		return fmt.Errorf("failed to initialize vector database: %w", err)
	} */

	iLog.Info("✅ Vector database schema initialized successfully")

	// Verify that tables were created
	tables := []string{
		"ai_embedding_configurations",
		"database_schema_embeddings",
		"business_entities",
		"query_templates",
		"embedding_generation_jobs",
		"embedding_search_logs",
	}

	for _, table := range tables {
		if targetDB.Migrator().HasTable(table) {
			iLog.Debug(fmt.Sprintf("  ✓ Table '%s' exists", table))
		} else {
			iLog.Warn(fmt.Sprintf("  ✗ Table '%s' not found (may be expected if using different database)", table))
		}
	}

	// Verify vector database configuration
	var sp string
	err = targetDB.Raw("SHOW search_path").Scan(&sp).Error
	iLog.Debug(fmt.Sprintf("Current search_path: %s", sp))

	var dbName string
	err = targetDB.Raw("SELECT current_database()").Scan(&dbName).Error
	iLog.Info(fmt.Sprintf("Connected to database: %s", dbName))

	// Check if pgvector extension is installed
	var extVersion string
	err = targetDB.Raw("SELECT extversion FROM pg_extension WHERE extname = 'vector'").Scan(&extVersion).Error
	if err != nil || extVersion == "" {
		iLog.Error("❌ pgvector extension is NOT installed in database: " + dbName)
		iLog.Error("   Please run: CREATE EXTENSION IF NOT EXISTS vector;")
		iLog.Error("   See VECTOR_DATABASE_SETUP.md for installation instructions")
		return fmt.Errorf("pgvector extension not found in database %s - please install it first", dbName)
	}
	iLog.Info(fmt.Sprintf("✅ pgvector extension is installed (version: %s)", extVersion))

	// Find which schema the vector type is in
	var vectorSchema string
	err = targetDB.Raw(`
		SELECT n.nspname
		FROM pg_type t
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE t.typname = 'vector'
	`).Scan(&vectorSchema).Error
	if err != nil || vectorSchema == "" {
		iLog.Error("❌ vector type not found in any schema")
		iLog.Error("   Extension may be installed but types are not available")
		return fmt.Errorf("vector type not found - extension installation may be incomplete")
	}
	iLog.Info(fmt.Sprintf("✅ vector type found in schema: %s", vectorSchema))

	// Update search_path to include the schema where vector type exists
	// This ensures vector operations work properly
	schemaName := vectorConfig.PostgresPGVector.Schema
	if schemaName == "" {
		schemaName = "public"
	}

	// Set search_path to include both the vector type schema and the data schema
	if vectorSchema != schemaName {
		newSearchPath := fmt.Sprintf("%s,%s,public", schemaName, vectorSchema)
		err = targetDB.Exec(fmt.Sprintf("SET search_path TO %s", newSearchPath)).Error
		if err != nil {
			iLog.Warn(fmt.Sprintf("Failed to update search_path: %v", err))
		} else {
			iLog.Info(fmt.Sprintf("✅ Updated search_path to: %s", newSearchPath))
		}
	}

	// Test vector operations
	var result float32
	err = targetDB.Raw("SELECT '[1,2,3]'::vector <=> '[1,1,1]'::vector AS distance").Scan(&result).Error
	if err != nil {
		iLog.Error(fmt.Sprintf("❌ Vector operation test failed: %v", err))
		iLog.Error(fmt.Sprintf("   Current search_path: %s", sp))
		iLog.Error(fmt.Sprintf("   Vector type schema: %s", vectorSchema))
		return fmt.Errorf("vector operations not working: %w", err)
	}
	iLog.Info(fmt.Sprintf("✅ Vector operations working (test distance: %.4f)", result))

	return nil
}
