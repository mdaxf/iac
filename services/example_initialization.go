package services

// This file provides examples of how to initialize and use the multi-database service layer
// It is for documentation purposes and should not be imported in production code

import (
	"context"
	"fmt"
	"log"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/dbinitializer"
	"gorm.io/gorm"
)

// Example1_BasicInitialization shows basic service initialization
func Example1_BasicInitialization() {
	// Step 1: Initialize database layer from environment variables
	dbInit := dbinitializer.NewDatabaseInitializer()
	if err := dbInit.InitializeFromEnvironment(); err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}

	// Step 2: Get the pool manager
	poolManager := dbInit.PoolManager

	// Note: You would need to create/inject your own GORM DB for app tables
	// This is just an example - in production, initialize GORM separately
	var appDB *gorm.DB // Initialize your GORM DB here

	// Step 3: Create service factory
	serviceFactory, err := NewServiceFactory(poolManager, appDB)
	if err != nil {
		log.Fatalf("Failed to create service factory: %v", err)
	}

	// Step 4: Use services
	ctx := context.Background()

	// Get business entity service (uses app database)
	businessEntitySvc := serviceFactory.GetBusinessEntityService()
	entities, err := businessEntitySvc.ListEntities(ctx, "")
	if err != nil {
		log.Printf("Error listing entities: %v", err)
	}
	fmt.Printf("Found %d business entities\n", len(entities))

	// Get schema metadata service (supports all database types)
	schemaMetadataSvc := serviceFactory.GetSchemaMetadataServiceMultiDB()

	// Discover schema for a MySQL database
	err = schemaMetadataSvc.DiscoverSchema(ctx, "mysql_db", "myschema")
	if err != nil {
		log.Printf("Error discovering MySQL schema: %v", err)
	}

	// Discover schema for a PostgreSQL database
	err = schemaMetadataSvc.DiscoverSchema(ctx, "postgres_db", "public")
	if err != nil {
		log.Printf("Error discovering PostgreSQL schema: %v", err)
	}
}

// Example2_CustomDatabaseQuery shows how to execute custom queries against user databases
func Example2_CustomDatabaseQuery(serviceFactory *ServiceFactory) {
	ctx := context.Background()

	// Get database helper
	dbHelper := serviceFactory.GetDatabaseHelper()

	// Get database connection for a specific alias
	db, err := dbHelper.GetUserDB(ctx, "customer_db")
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	defer db.Close()

	// Get the database type to determine dialect
	dbType := string(db.GetType())
	fmt.Printf("Database type: %s\n", dbType)

	// Execute a dialect-aware query
	var query string
	switch dbType {
	case "mysql":
		query = "SELECT * FROM customers LIMIT 10"
	case "postgres":
		query = "SELECT * FROM customers LIMIT 10"
	case "mssql":
		query = "SELECT TOP 10 * FROM customers"
	case "oracle":
		query = "SELECT * FROM customers FETCH FIRST 10 ROWS ONLY"
	}

	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// Process rows...
	for rows.Next() {
		// Scan row data...
	}
}

// Example3_DialectAwareQueries shows how to use dialect helper functions
func Example3_DialectAwareQueries(serviceFactory *ServiceFactory) {
	ctx := context.Background()
	dbHelper := serviceFactory.GetDatabaseHelper()

	// Get database
	db, err := dbHelper.GetUserDB(ctx, "customer_db")
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	defer db.Close()

	dbType := string(db.GetType())

	// Build a query with pagination
	baseQuery := "SELECT * FROM customers ORDER BY id"
	paginationClause := LimitOffsetClause(dbType, 10, 20) // limit 10, offset 20
	fullQuery := fmt.Sprintf("%s %s", baseQuery, paginationClause)

	rows, err := db.Query(ctx, fullQuery)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// Build a query with case-insensitive search
	likeOp := LikeOperator(dbType, false) // case-insensitive
	searchQuery := fmt.Sprintf("SELECT * FROM customers WHERE name %s ?", likeOp)

	rows2, err := db.Query(ctx, searchQuery, "%john%")
	if err != nil {
		log.Fatalf("Search query failed: %v", err)
	}
	defer rows2.Close()

	// Build a query with JSON extraction
	jsonExtract := JSONExtractExpr(dbType, "metadata", "email")
	jsonQuery := fmt.Sprintf("SELECT id, %s as email FROM customers", jsonExtract)

	rows3, err := db.Query(ctx, jsonQuery)
	if err != nil {
		log.Fatalf("JSON query failed: %v", err)
	}
	defer rows3.Close()
}

// Example4_SchemaDiscovery shows comprehensive schema discovery
func Example4_SchemaDiscovery(serviceFactory *ServiceFactory) {
	ctx := context.Background()
	svc := serviceFactory.GetSchemaMetadataServiceMultiDB()

	// Discover schema for different database types
	databases := map[string]string{
		"mysql_prod":    "production",
		"postgres_dev":  "public",
		"mssql_staging": "dbo",
		"oracle_prod":   "PRODUSER",
	}

	for dbAlias, schemaName := range databases {
		fmt.Printf("Discovering schema for %s (schema: %s)...\n", dbAlias, schemaName)

		err := svc.DiscoverSchema(ctx, dbAlias, schemaName)
		if err != nil {
			log.Printf("Error discovering %s: %v", dbAlias, err)
			continue
		}

		// Get discovered tables
		tables, err := svc.GetTableMetadata(ctx, dbAlias)
		if err != nil {
			log.Printf("Error getting tables for %s: %v", dbAlias, err)
			continue
		}

		fmt.Printf("Found %d tables in %s\n", len(tables), dbAlias)

		// Get columns for each table
		for _, table := range tables {
			columns, err := svc.GetColumnMetadata(ctx, dbAlias, table.Table)
			if err != nil {
				log.Printf("Error getting columns for %s.%s: %v", dbAlias, table.Table, err)
				continue
			}

			fmt.Printf("  Table: %s (%d columns)\n", table.Table, len(columns))

			// Print first few columns
			for i, col := range columns {
				if i >= 3 {
					fmt.Printf("    ... and %d more columns\n", len(columns)-3)
					break
				}
				nullable := "NOT NULL"
				if col.IsNullable != nil && *col.IsNullable {
					nullable = "NULL"
				}
				fmt.Printf("    - %s (%s) %s\n", col.Column, col.DataType, nullable)
			}

			// Discover indexes for this table
			indexes, err := svc.DiscoverIndexes(ctx, dbAlias, schemaName, table.Table)
			if err != nil {
				log.Printf("Error discovering indexes for %s.%s: %v", dbAlias, table.Table, err)
				continue
			}

			if len(indexes) > 0 {
				fmt.Printf("  Indexes: %d\n", len(indexes))
				indexMap := make(map[string][]string)
				for _, idx := range indexes {
					indexMap[idx.IndexName] = append(indexMap[idx.IndexName], idx.ColumnName)
				}
				for idxName, cols := range indexMap {
					fmt.Printf("    - %s: %v\n", idxName, cols)
				}
			}
		}

		fmt.Println()
	}
}

// Example5_ServiceWithCustomLogic shows how to create a custom service using DatabaseHelper
func Example5_ServiceWithCustomLogic(serviceFactory *ServiceFactory) {
	// Custom service that needs to query user databases
	type CustomerService struct {
		dbHelper *DatabaseHelper
		appDB    *gorm.DB
	}

	// Create custom service
	customerSvc := &CustomerService{
		dbHelper: serviceFactory.GetDatabaseHelper(),
		appDB:    serviceFactory.GetAppDB(),
	}

	// Use the service
	ctx := context.Background()

	// Method 1: Query user database
	db, err := customerSvc.dbHelper.GetUserDB(ctx, "customer_db")
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	defer db.Close()

	dbType := string(db.GetType())

	// Build dialect-aware query
	query := fmt.Sprintf(
		"SELECT * FROM customers WHERE status = ? %s",
		LimitOffsetClause(dbType, 100, 0),
	)

	rows, err := db.Query(ctx, query, "active")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	// Method 2: Use app database for metadata
	var count int64
	customerSvc.appDB.Model(&struct {
		ID           string
		DatabaseName string
	}{}).Where("database_name = ?", "customer_db").Count(&count)

	fmt.Printf("Found %d records in app database\n", count)
}

// Example6_TransactionSupport shows how to use transactions with multi-database support
func Example6_TransactionSupport(serviceFactory *ServiceFactory) {
	ctx := context.Background()
	dbHelper := serviceFactory.GetDatabaseHelper()

	// Get database for write operations
	db, err := dbHelper.GetUserDBForWrite(ctx, "customer_db")
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	defer db.Close()

	// Check if database supports transactions (it should)
	if !db.SupportsFeature("transactions") {
		log.Fatal("Database does not support transactions")
	}

	// Begin transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Execute queries in transaction
	_, err = tx.Exec("INSERT INTO customers (name, email) VALUES (?, ?)", "John Doe", "john@example.com")
	if err != nil {
		tx.Rollback()
		log.Fatalf("Insert failed: %v", err)
	}

	_, err = tx.Exec("UPDATE customer_stats SET total_customers = total_customers + 1")
	if err != nil {
		tx.Rollback()
		log.Fatalf("Update failed: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Commit failed: %v", err)
	}

	fmt.Println("Transaction completed successfully")
}

// Example7_HealthChecksAndMonitoring shows how to monitor database health
func Example7_HealthChecksAndMonitoring(serviceFactory *ServiceFactory) {
	ctx := context.Background()
	dbHelper := serviceFactory.GetDatabaseHelper()

	// Check multiple database aliases
	aliases := []string{"mysql_db", "postgres_db", "oracle_db"}

	for _, alias := range aliases {
		db, err := dbHelper.GetUserDB(ctx, alias)
		if err != nil {
			log.Printf("Failed to connect to %s: %v", alias, err)
			continue
		}

		// Ping database
		if err := db.Ping(); err != nil {
			log.Printf("Database %s is not healthy: %v", alias, err)
			db.Close()
			continue
		}

		// Get database type
		dbType := string(db.GetType())

		// Check features
		features := []dbconn.Feature{
			dbconn.FeatureCTE,
			dbconn.FeatureJSON,
			dbconn.FeatureWindowFunctions,
			dbconn.FeatureFullTextSearch,
		}
		supported := make([]string, 0)
		for _, feature := range features {
			if db.SupportsFeature(feature) {
				supported = append(supported, string(feature))
			}
		}

		fmt.Printf("Database: %s, Type: %s, Healthy: ✓, Features: %v\n", alias, dbType, supported)
		db.Close()
	}
}

// These examples demonstrate the complete usage of the multi-database service layer
// In production, you would initialize the ServiceFactory once at application startup
// and pass it to your handlers/controllers
