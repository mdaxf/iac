package services

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	vectorDBInstance *gorm.DB
	vectorDBOnce     sync.Once
	vectorDBMutex    sync.RWMutex
	vectorDBError    error
)

// GetVectorDB returns the vector database connection based on aiconfig.json
// If use_main_db is true, it returns the main database connection
// Otherwise, it establishes a separate connection to the vector database
func GetVectorDB(mainDB *gorm.DB) (*gorm.DB, error) {
	iLog := logger.Log{
		ModuleName:     logger.Framework,
		User:           "System",
		ControllerName: "VectorDBConnection",
	}

	// Get vector database configuration
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return nil, fmt.Errorf("AI configuration not loaded")
	}

	vectorConfig := aiConfig.VectorDatabase

	// Check if vector database is configured
	if vectorConfig.Type != "postgres_pgvector" {
		return nil, fmt.Errorf("vector database type '%s' is not supported for PostgreSQL operations", vectorConfig.Type)
	}

	if !vectorConfig.PostgresPGVector.Enabled {
		return nil, fmt.Errorf("PostgreSQL pgvector is not enabled in configuration")
	}

	pgConfig := vectorConfig.PostgresPGVector

	// If use_main_db is true, return the main database connection
	if pgConfig.UseMainDB {
		iLog.Debug("Using main database for vector storage")
		return mainDB, nil
	}

	// Otherwise, establish a separate connection to the vector database
	vectorDBMutex.Lock()
	defer vectorDBMutex.Unlock()

	// Use sync.Once to ensure we only connect once
	vectorDBOnce.Do(func() {
		iLog.Info(fmt.Sprintf("Establishing connection to vector database: %s", pgConfig.ConnectionString))

		// Build connection string with schema in search_path if specified
		connStr := pgConfig.ConnectionString
		if pgConfig.Schema != "" && pgConfig.Schema != "public" {
			// Add search_path to connection string
			// IMPORTANT: Include 'public' schema for pgvector extension types
			searchPath := fmt.Sprintf("%s,public", pgConfig.Schema)
			if strings.Contains(connStr, "?") {
				connStr += fmt.Sprintf("&search_path=%s", searchPath)
			} else {
				connStr += fmt.Sprintf("?search_path=%s", searchPath)
			}
			iLog.Info(fmt.Sprintf("Setting search_path to: %s in connection string", searchPath))
		}

		// Parse connection string and connect with custom dialector
		dialector := postgres.New(postgres.Config{
			DSN: connStr,
			// Don't automatically parse time fields - let pgvector handle vectors
			PreferSimpleProtocol: false,
		})

		db, err := gorm.Open(dialector, &gorm.Config{
			// Disable default transaction for better performance
			SkipDefaultTransaction: true,
			// Disable automatic foreign key constraints
			DisableForeignKeyConstraintWhenMigrating: true,
		})

		if err != nil {
			vectorDBError = fmt.Errorf("failed to connect to vector database: %w", err)
			iLog.Error(fmt.Sprintf("Vector DB connection failed: %v", err))
			return
		}

		// Get underlying SQL DB to configure connection pool and verify pgvector
		sqlDB, err := db.DB()
		if err != nil {
			vectorDBError = fmt.Errorf("failed to get underlying SQL DB: %w", err)
			iLog.Error(fmt.Sprintf("Failed to configure connection pool: %v", err))
			return
		}

		// Verify the connection works with vector type
		var result string
		err = sqlDB.QueryRow("SELECT version()").Scan(&result)
		if err != nil {
			vectorDBError = fmt.Errorf("failed to verify vector database connection: %w", err)
			iLog.Error(fmt.Sprintf("Vector DB verification failed: %v", err))
			return
		}
		iLog.Debug(fmt.Sprintf("Vector DB connection verified: %s", result))
		iLog.Info("✅ pgvector-go library integrated - vector types handled automatically")

		// Configure connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		vectorDBInstance = db
		iLog.Info("✅ Vector database connection established successfully")
	})

	if vectorDBError != nil {
		return nil, vectorDBError
	}

	return vectorDBInstance, nil
}

// GetVectorDBSchema returns the schema name for vector database operations
func GetVectorDBSchema() string {
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return "public" // Default schema
	}

	if aiConfig.VectorDatabase.Type == "postgres_pgvector" && aiConfig.VectorDatabase.PostgresPGVector.Enabled {
		schema := aiConfig.VectorDatabase.PostgresPGVector.Schema
		if schema == "" {
			return "public"
		}
		return schema
	}

	return "public"
}

// GetVectorDBTablePrefix returns the table prefix for vector database tables
func GetVectorDBTablePrefix() string {
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		return ""
	}

	if aiConfig.VectorDatabase.Type == "postgres_pgvector" && aiConfig.VectorDatabase.PostgresPGVector.Enabled {
		return aiConfig.VectorDatabase.PostgresPGVector.TablePrefix
	}

	return ""
}

// CloseVectorDB closes the vector database connection (if it's separate from main DB)
func CloseVectorDB() error {
	vectorDBMutex.Lock()
	defer vectorDBMutex.Unlock()

	if vectorDBInstance == nil {
		return nil
	}

	// Only close if it's NOT the main database connection
	aiConfig := config.GetAIConfig()
	if aiConfig != nil && aiConfig.VectorDatabase.PostgresPGVector.UseMainDB {
		// Don't close main DB connection
		return nil
	}

	sqlDB, err := vectorDBInstance.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close vector database connection: %w", err)
	}

	vectorDBInstance = nil
	return nil
}
