// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

// PostgresJSONBAdapter implements the DocumentDB interface using PostgreSQL JSONB
type PostgresJSONBAdapter struct {
	config      *documents.DocDBConfig
	db          *sql.DB
	connected   bool
	connectedAt time.Time
	mu          sync.RWMutex
	iLog        logger.Log
}

// NewPostgresJSONBAdapter creates a new PostgreSQL JSONB adapter
func NewPostgresJSONBAdapter(config *documents.DocDBConfig) *PostgresJSONBAdapter {
	return &PostgresJSONBAdapter{
		config: config,
		iLog:   logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "PostgresJSONBAdapter"},
	}
}

// Connect establishes connection to PostgreSQL
func (p *PostgresJSONBAdapter) Connect(config *documents.DocDBConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid PostgreSQL configuration: %w", err)
	}

	p.config = config

	// Build connection string
	connStr := p.buildConnectionString()

	// Open database connection
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	if config.MaxPoolSize > 0 {
		db.SetMaxOpenConns(config.MaxPoolSize)
	}
	if config.MinPoolSize > 0 {
		db.SetMaxIdleConns(config.MinPoolSize)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.ConnTimeout)*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.db = db
	p.connected = true
	p.connectedAt = time.Now()

	p.iLog.Info(fmt.Sprintf("Connected to PostgreSQL JSONB document store at %s:%d/%s",
		config.Host, config.Port, config.Database))

	// Create metadata table if it doesn't exist
	if err := p.createMetadataTable(context.Background()); err != nil {
		return fmt.Errorf("failed to create metadata table: %w", err)
	}

	return nil
}

// Close closes the PostgreSQL connection
func (p *PostgresJSONBAdapter) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.db != nil {
		if err := p.db.Close(); err != nil {
			return fmt.Errorf("failed to close PostgreSQL connection: %w", err)
		}

		p.connected = false
		p.iLog.Info("Disconnected from PostgreSQL JSONB document store")
	}

	return nil
}

// Ping checks the connection to PostgreSQL
func (p *PostgresJSONBAdapter) Ping(ctx context.Context) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.db == nil {
		return documents.ErrConnectionFailed
	}

	return p.db.PingContext(ctx)
}

// CreateCollection creates a new collection (table) for documents
func (p *PostgresJSONBAdapter) CreateCollection(ctx context.Context, name string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return documents.ErrConnectionFailed
	}

	// Create table with JSONB column
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			_id SERIAL PRIMARY KEY,
			_uuid UUID DEFAULT gen_random_uuid(),
			data JSONB NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, p.quoteIdentifier(name))

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Create GIN index for JSONB queries
	indexQuery := fmt.Sprintf(`
		CREATE INDEX IF NOT EXISTS %s ON %s USING GIN (data)
	`, p.quoteIdentifier(name+"_gin_idx"), p.quoteIdentifier(name))

	_, err = p.db.ExecContext(ctx, indexQuery)
	if err != nil {
		return fmt.Errorf("failed to create GIN index: %w", err)
	}

	// Register collection in metadata
	if err := p.registerCollection(ctx, name); err != nil {
		return fmt.Errorf("failed to register collection: %w", err)
	}

	p.iLog.Info(fmt.Sprintf("Created collection '%s'", name))

	return nil
}

// DropCollection drops a collection (table)
func (p *PostgresJSONBAdapter) DropCollection(ctx context.Context, name string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return documents.ErrConnectionFailed
	}

	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", p.quoteIdentifier(name))

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop collection: %w", err)
	}

	// Unregister collection from metadata
	if err := p.unregisterCollection(ctx, name); err != nil {
		return fmt.Errorf("failed to unregister collection: %w", err)
	}

	p.iLog.Info(fmt.Sprintf("Dropped collection '%s'", name))

	return nil
}

// ListCollections lists all collections
func (p *PostgresJSONBAdapter) ListCollections(ctx context.Context) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, documents.ErrConnectionFailed
	}

	query := "SELECT collection_name FROM _doc_metadata ORDER BY collection_name"

	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}
	defer rows.Close()

	var collections []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan collection name: %w", err)
		}
		collections = append(collections, name)
	}

	return collections, nil
}

// CollectionExists checks if a collection exists
func (p *PostgresJSONBAdapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return false, documents.ErrConnectionFailed
	}

	query := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`

	var exists bool
	err := p.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check collection existence: %w", err)
	}

	return exists, nil
}

// InsertOne inserts a single document
func (p *PostgresJSONBAdapter) InsertOne(ctx context.Context, collection string, document interface{}) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return "", documents.ErrConnectionFailed
	}

	// Convert document to JSON
	jsonData, err := json.Marshal(document)
	if err != nil {
		return "", fmt.Errorf("failed to marshal document: %w", err)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (data) VALUES ($1) RETURNING _uuid
	`, p.quoteIdentifier(collection))

	var uuid string
	err = p.db.QueryRowContext(ctx, query, jsonData).Scan(&uuid)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	return uuid, nil
}

// InsertMany inserts multiple documents
func (p *PostgresJSONBAdapter) InsertMany(ctx context.Context, collection string, documents []interface{}) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, documents.ErrConnectionFailed
	}

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := fmt.Sprintf(`
		INSERT INTO %s (data) VALUES ($1) RETURNING _uuid
	`, p.quoteIdentifier(collection))

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	ids := make([]string, 0, len(documents))
	for _, doc := range documents {
		jsonData, err := json.Marshal(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document: %w", err)
		}

		var uuid string
		if err := stmt.QueryRowContext(ctx, jsonData).Scan(&uuid); err != nil {
			return nil, fmt.Errorf("failed to insert document: %w", err)
		}
		ids = append(ids, uuid)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

// FindOne finds a single document
func (p *PostgresJSONBAdapter) FindOne(ctx context.Context, collection string, filter interface{}) (map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, documents.ErrConnectionFailed
	}

	whereClause, args := p.buildWhereClause(filter)

	query := fmt.Sprintf(`
		SELECT data FROM %s WHERE %s LIMIT 1
	`, p.quoteIdentifier(collection), whereClause)

	var jsonData []byte
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&jsonData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, documents.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	return result, nil
}

// FindMany finds multiple documents
func (p *PostgresJSONBAdapter) FindMany(ctx context.Context, collection string, filter interface{}, opts *documents.FindOptions) ([]map[string]interface{}, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, documents.ErrConnectionFailed
	}

	whereClause, args := p.buildWhereClause(filter)

	query := fmt.Sprintf("SELECT data FROM %s WHERE %s", p.quoteIdentifier(collection), whereClause)

	// Add sorting
	if opts != nil && opts.Sort != nil {
		sortClauses := make([]string, 0)
		for field, order := range opts.Sort {
			direction := "ASC"
			if order == -1 {
				direction = "DESC"
			}
			sortClauses = append(sortClauses, fmt.Sprintf("data->>'%s' %s", field, direction))
		}
		if len(sortClauses) > 0 {
			query += " ORDER BY " + strings.Join(sortClauses, ", ")
		}
	}

	// Add limit and offset
	if opts != nil {
		if opts.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		}
		if opts.Skip > 0 {
			query += fmt.Sprintf(" OFFSET %d", opts.Skip)
		}
	}

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var jsonData []byte
		if err := rows.Scan(&jsonData); err != nil {
			return nil, fmt.Errorf("failed to scan document: %w", err)
		}

		var doc map[string]interface{}
		if err := json.Unmarshal(jsonData, &doc); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}

		results = append(results, doc)
	}

	return results, nil
}

// UpdateOne updates a single document
func (p *PostgresJSONBAdapter) UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return documents.ErrConnectionFailed
	}

	whereClause, whereArgs := p.buildWhereClause(filter)
	updateClause, updateArgs := p.buildUpdateClause(update)

	args := append(updateArgs, whereArgs...)

	query := fmt.Sprintf(`
		UPDATE %s SET %s, updated_at = CURRENT_TIMESTAMP WHERE %s
	`, p.quoteIdentifier(collection), updateClause, whereClause)

	_, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// UpdateMany updates multiple documents
func (p *PostgresJSONBAdapter) UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return 0, documents.ErrConnectionFailed
	}

	whereClause, whereArgs := p.buildWhereClause(filter)
	updateClause, updateArgs := p.buildUpdateClause(update)

	args := append(updateArgs, whereArgs...)

	query := fmt.Sprintf(`
		UPDATE %s SET %s, updated_at = CURRENT_TIMESTAMP WHERE %s
	`, p.quoteIdentifier(collection), updateClause, whereClause)

	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to update documents: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// DeleteOne deletes a single document
func (p *PostgresJSONBAdapter) DeleteOne(ctx context.Context, collection string, filter interface{}) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return documents.ErrConnectionFailed
	}

	whereClause, args := p.buildWhereClause(filter)

	query := fmt.Sprintf(`
		DELETE FROM %s WHERE _id IN (
			SELECT _id FROM %s WHERE %s LIMIT 1
		)
	`, p.quoteIdentifier(collection), p.quoteIdentifier(collection), whereClause)

	_, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteMany deletes multiple documents
func (p *PostgresJSONBAdapter) DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return 0, documents.ErrConnectionFailed
	}

	whereClause, args := p.buildWhereClause(filter)

	query := fmt.Sprintf(`
		DELETE FROM %s WHERE %s
	`, p.quoteIdentifier(collection), whereClause)

	result, err := p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

// CountDocuments counts documents matching the filter
func (p *PostgresJSONBAdapter) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return 0, documents.ErrConnectionFailed
	}

	whereClause, args := p.buildWhereClause(filter)

	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s WHERE %s
	`, p.quoteIdentifier(collection), whereClause)

	var count int64
	err := p.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// CreateIndex creates an index on a collection
func (p *PostgresJSONBAdapter) CreateIndex(ctx context.Context, collection string, keys map[string]int, opts *documents.IndexOptions) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return documents.ErrConnectionFailed
	}

	indexName := fmt.Sprintf("%s_idx", collection)
	if opts != nil && opts.Name != "" {
		indexName = opts.Name
	}

	// Build index fields
	fields := make([]string, 0)
	for field, order := range keys {
		direction := "ASC"
		if order == -1 {
			direction = "DESC"
		}
		fields = append(fields, fmt.Sprintf("(data->>'%s') %s", field, direction))
	}

	unique := ""
	if opts != nil && opts.Unique {
		unique = "UNIQUE"
	}

	query := fmt.Sprintf(`
		CREATE %s INDEX IF NOT EXISTS %s ON %s (%s)
	`, unique, p.quoteIdentifier(indexName), p.quoteIdentifier(collection), strings.Join(fields, ", "))

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// DropIndex drops an index from a collection
func (p *PostgresJSONBAdapter) DropIndex(ctx context.Context, collection string, name string) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return documents.ErrConnectionFailed
	}

	query := fmt.Sprintf("DROP INDEX IF EXISTS %s", p.quoteIdentifier(name))

	_, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	return nil
}

// ListIndexes lists all indexes on a collection
func (p *PostgresJSONBAdapter) ListIndexes(ctx context.Context, collection string) ([]documents.IndexInfo, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, documents.ErrConnectionFailed
	}

	query := `
		SELECT indexname, indexdef
		FROM pg_indexes
		WHERE tablename = $1
	`

	rows, err := p.db.QueryContext(ctx, query, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer rows.Close()

	var indexes []documents.IndexInfo
	for rows.Next() {
		var name, def string
		if err := rows.Scan(&name, &def); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		indexInfo := documents.IndexInfo{
			Name:   name,
			Keys:   make(map[string]int),
			Unique: strings.Contains(def, "UNIQUE"),
		}

		indexes = append(indexes, indexInfo)
	}

	return indexes, nil
}

// Aggregate performs an aggregation operation (simplified for PostgreSQL)
func (p *PostgresJSONBAdapter) Aggregate(ctx context.Context, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error) {
	// This is a simplified implementation
	// Full MongoDB-style aggregation pipeline translation is complex
	return nil, fmt.Errorf("aggregation not fully implemented for PostgreSQL JSONB adapter")
}

// GetType returns the database type
func (p *PostgresJSONBAdapter) GetType() documents.DocDBType {
	return documents.DocDBTypePostgres
}

// IsConnected checks if the adapter is connected
func (p *PostgresJSONBAdapter) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// Stats returns database statistics
func (p *PostgresJSONBAdapter) Stats(ctx context.Context) (*documents.DBStats, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.connected {
		return nil, documents.ErrConnectionFailed
	}

	collections, err := p.ListCollections(ctx)
	if err != nil {
		return nil, err
	}

	stats := &documents.DBStats{
		Collections: int64(len(collections)),
	}

	// Get total document count and sizes
	for _, coll := range collections {
		count, _ := p.CountDocuments(ctx, coll, map[string]interface{}{})
		stats.Documents += count
	}

	// Get database size
	var dbSize int64
	query := "SELECT pg_database_size(current_database())"
	if err := p.db.QueryRowContext(ctx, query).Scan(&dbSize); err == nil {
		stats.DataSize = dbSize
	}

	return stats, nil
}

// Helper methods

// buildConnectionString builds PostgreSQL connection string
func (p *PostgresJSONBAdapter) buildConnectionString() string {
	sslMode := p.config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.config.Host, p.config.Port, p.config.Username, p.config.Password,
		p.config.Database, sslMode)
}

// quoteIdentifier quotes an identifier for PostgreSQL
func (p *PostgresJSONBAdapter) quoteIdentifier(name string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

// buildWhereClause builds a WHERE clause from a filter
func (p *PostgresJSONBAdapter) buildWhereClause(filter interface{}) (string, []interface{}) {
	filterMap, ok := filter.(map[string]interface{})
	if !ok {
		return "TRUE", nil
	}

	if len(filterMap) == 0 {
		return "TRUE", nil
	}

	clauses := make([]string, 0)
	args := make([]interface{}, 0)
	argNum := 1

	for key, value := range filterMap {
		clauses = append(clauses, fmt.Sprintf("data->>'%s' = $%d", key, argNum))
		args = append(args, fmt.Sprintf("%v", value))
		argNum++
	}

	return strings.Join(clauses, " AND "), args
}

// buildUpdateClause builds an UPDATE clause from an update specification
func (p *PostgresJSONBAdapter) buildUpdateClause(update interface{}) (string, []interface{}) {
	updateMap, ok := update.(map[string]interface{})
	if !ok {
		return "", nil
	}

	// Check for MongoDB-style update operators
	if set, ok := updateMap["$set"].(map[string]interface{}); ok {
		updateMap = set
	}

	if len(updateMap) == 0 {
		return "", nil
	}

	setClauses := make([]string, 0)
	args := make([]interface{}, 0)
	argNum := 1

	for key, value := range updateMap {
		jsonValue, _ := json.Marshal(value)
		setClauses = append(setClauses, fmt.Sprintf("data = jsonb_set(data, '{%s}', $%d)", key, argNum))
		args = append(args, string(jsonValue))
		argNum++
	}

	return strings.Join(setClauses, ", "), args
}

// createMetadataTable creates a metadata table to track collections
func (p *PostgresJSONBAdapter) createMetadataTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS _doc_metadata (
			collection_name TEXT PRIMARY KEY,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := p.db.ExecContext(ctx, query)
	return err
}

// registerCollection registers a collection in metadata
func (p *PostgresJSONBAdapter) registerCollection(ctx context.Context, name string) error {
	query := `
		INSERT INTO _doc_metadata (collection_name)
		VALUES ($1)
		ON CONFLICT (collection_name) DO NOTHING
	`

	_, err := p.db.ExecContext(ctx, query, name)
	return err
}

// unregisterCollection unregisters a collection from metadata
func (p *PostgresJSONBAdapter) unregisterCollection(ctx context.Context, name string) error {
	query := "DELETE FROM _doc_metadata WHERE collection_name = $1"

	_, err := p.db.ExecContext(ctx, query, name)
	return err
}
