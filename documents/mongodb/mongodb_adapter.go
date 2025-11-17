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

package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
)

// MongoDBAdapter implements the DocumentDB interface for MongoDB
type MongoDBAdapter struct {
	config      *documents.DocDBConfig
	client      *mongo.Client
	database    *mongo.Database
	connected   bool
	connectedAt time.Time
	mu          sync.RWMutex
	iLog        logger.Log
}

// NewMongoDBAdapter creates a new MongoDB adapter
func NewMongoDBAdapter(config *documents.DocDBConfig) *MongoDBAdapter {
	return &MongoDBAdapter{
		config: config,
		iLog:   logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MongoDBAdapter"},
	}
}

// Connect establishes connection to MongoDB
func (m *MongoDBAdapter) Connect(config *documents.DocDBConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid MongoDB configuration: %w", err)
	}

	m.config = config

	// Build connection string
	connStr := m.buildConnectionString()

	// Set client options
	clientOptions := options.Client().ApplyURI(connStr)

	// Set pool size
	if config.MaxPoolSize > 0 {
		maxPoolSize := uint64(config.MaxPoolSize)
		clientOptions.SetMaxPoolSize(maxPoolSize)
	}
	if config.MinPoolSize > 0 {
		minPoolSize := uint64(config.MinPoolSize)
		clientOptions.SetMinPoolSize(minPoolSize)
	}

	// Set authentication if provided
	if config.Username != "" && config.Password != "" {
		credential := options.Credential{
			Username:   config.Username,
			Password:   config.Password,
			AuthSource: config.AuthSource,
		}
		if config.AuthSource == "" {
			credential.AuthSource = config.Database
		}
		clientOptions.SetAuth(credential)
	}

	// Set replica set if provided
	if config.ReplicaSet != "" {
		clientOptions.SetReplicaSet(config.ReplicaSet)
	}

	// Create client
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.ConnTimeout)*time.Second)
	defer cancel()

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.database = client.Database(config.Database)
	m.connected = true
	m.connectedAt = time.Now()

	m.iLog.Info(fmt.Sprintf("Connected to MongoDB at %s:%d/%s",
		config.Host, config.Port, config.Database))

	return nil
}

// Close closes the MongoDB connection
func (m *MongoDBAdapter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := m.client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
		}

		m.connected = false
		m.iLog.Info("Disconnected from MongoDB")
	}

	return nil
}

// Ping checks the connection to MongoDB
func (m *MongoDBAdapter) Ping(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.client == nil {
		return documents.ErrConnectionFailed
	}

	return m.client.Ping(ctx, nil)
}

// CreateCollection creates a new collection
func (m *MongoDBAdapter) CreateCollection(ctx context.Context, name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return documents.ErrConnectionFailed
	}

	return m.database.CreateCollection(ctx, name)
}

// DropCollection drops a collection
func (m *MongoDBAdapter) DropCollection(ctx context.Context, name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return documents.ErrConnectionFailed
	}

	collection := m.database.Collection(name)
	return collection.Drop(ctx)
}

// ListCollections lists all collections in the database
func (m *MongoDBAdapter) ListCollections(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, documents.ErrConnectionFailed
	}

	names, err := m.database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return names, nil
}

// CollectionExists checks if a collection exists
func (m *MongoDBAdapter) CollectionExists(ctx context.Context, name string) (bool, error) {
	collections, err := m.ListCollections(ctx)
	if err != nil {
		return false, err
	}

	for _, col := range collections {
		if col == name {
			return true, nil
		}
	}

	return false, nil
}

// InsertOne inserts a single document
func (m *MongoDBAdapter) InsertOne(ctx context.Context, collection string, document interface{}) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return "", documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	result, err := coll.InsertOne(ctx, document)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	// Extract the inserted ID
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		return oid.Hex(), nil
	}

	return fmt.Sprintf("%v", result.InsertedID), nil
}

// InsertMany inserts multiple documents
func (m *MongoDBAdapter) InsertMany(ctx context.Context, collection string, docs []interface{}) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	result, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("failed to insert documents: %w", err)
	}

	// Extract inserted IDs
	ids := make([]string, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		if oid, ok := id.(primitive.ObjectID); ok {
			ids[i] = oid.Hex()
		} else {
			ids[i] = fmt.Sprintf("%v", id)
		}
	}

	return ids, nil
}

// FindOne finds a single document
func (m *MongoDBAdapter) FindOne(ctx context.Context, collection string, filter interface{}) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)

	var result bson.M
	err := coll.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, documents.ErrDocumentNotFound
		}
		return nil, fmt.Errorf("failed to find document: %w", err)
	}

	return m.convertBsonMToMap(result), nil
}

// FindMany finds multiple documents
func (m *MongoDBAdapter) FindMany(ctx context.Context, collection string, filter interface{}, opts *documents.FindOptions) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)

	// Build find options
	findOptions := options.Find()

	if opts != nil {
		if opts.Limit > 0 {
			findOptions.SetLimit(opts.Limit)
		}
		if opts.Skip > 0 {
			findOptions.SetSkip(opts.Skip)
		}
		if opts.Sort != nil {
			findOptions.SetSort(opts.Sort)
		}
		if opts.Projection != nil {
			findOptions.SetProjection(opts.Projection)
		}
	}

	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode document: %w", err)
		}
		results = append(results, m.convertBsonMToMap(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return results, nil
}

// UpdateOne updates a single document
func (m *MongoDBAdapter) UpdateOne(ctx context.Context, collection string, filter interface{}, update interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// UpdateMany updates multiple documents
func (m *MongoDBAdapter) UpdateMany(ctx context.Context, collection string, filter interface{}, update interface{}) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return 0, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	result, err := coll.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, fmt.Errorf("failed to update documents: %w", err)
	}

	return result.ModifiedCount, nil
}

// DeleteOne deletes a single document
func (m *MongoDBAdapter) DeleteOne(ctx context.Context, collection string, filter interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	_, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteMany deletes multiple documents
func (m *MongoDBAdapter) DeleteMany(ctx context.Context, collection string, filter interface{}) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return 0, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	result, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents: %w", err)
	}

	return result.DeletedCount, nil
}

// CountDocuments counts documents matching the filter
func (m *MongoDBAdapter) CountDocuments(ctx context.Context, collection string, filter interface{}) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return 0, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// CreateIndex creates an index on a collection
func (m *MongoDBAdapter) CreateIndex(ctx context.Context, collection string, keys map[string]int, opts *documents.IndexOptions) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)

	// Build index model
	indexModel := mongo.IndexModel{
		Keys: keys,
	}

	if opts != nil {
		indexOptions := options.Index()

		if opts.Name != "" {
			indexOptions.SetName(opts.Name)
		}
		if opts.Unique {
			indexOptions.SetUnique(true)
		}
		if opts.Background {
			indexOptions.SetBackground(true)
		}
		if opts.Sparse {
			indexOptions.SetSparse(true)
		}

		indexModel.Options = indexOptions
	}

	_, err := coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// DropIndex drops an index from a collection
func (m *MongoDBAdapter) DropIndex(ctx context.Context, collection string, name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	_, err := coll.Indexes().DropOne(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to drop index: %w", err)
	}

	return nil
}

// ListIndexes lists all indexes on a collection
func (m *MongoDBAdapter) ListIndexes(ctx context.Context, collection string) ([]documents.IndexInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)
	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list indexes: %w", err)
	}
	defer cursor.Close(ctx)

	var indexes []documents.IndexInfo
	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			return nil, fmt.Errorf("failed to decode index: %w", err)
		}

		indexInfo := documents.IndexInfo{
			Name: idx["name"].(string),
		}

		// Extract keys
		if keys, ok := idx["key"].(bson.M); ok {
			indexInfo.Keys = make(map[string]int)
			for k, v := range keys {
				if val, ok := v.(int32); ok {
					indexInfo.Keys[k] = int(val)
				}
			}
		}

		// Extract options
		if unique, ok := idx["unique"].(bool); ok {
			indexInfo.Unique = unique
		}
		if sparse, ok := idx["sparse"].(bool); ok {
			indexInfo.Sparse = sparse
		}

		indexes = append(indexes, indexInfo)
	}

	return indexes, nil
}

// Aggregate performs an aggregation operation
func (m *MongoDBAdapter) Aggregate(ctx context.Context, collection string, pipeline []map[string]interface{}) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, documents.ErrConnectionFailed
	}

	coll := m.database.Collection(collection)

	// Convert pipeline to bson
	bsonPipeline := make([]interface{}, len(pipeline))
	for i, stage := range pipeline {
		bsonPipeline[i] = stage
	}

	cursor, err := coll.Aggregate(ctx, bsonPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode aggregation result: %w", err)
		}
		results = append(results, m.convertBsonMToMap(doc))
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("aggregation cursor error: %w", err)
	}

	return results, nil
}

// GetType returns the database type
func (m *MongoDBAdapter) GetType() documents.DocDBType {
	return documents.DocDBTypeMongoDB
}

// IsConnected checks if the adapter is connected
func (m *MongoDBAdapter) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// Stats returns database statistics
func (m *MongoDBAdapter) Stats(ctx context.Context) (*documents.DBStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.connected {
		return nil, documents.ErrConnectionFailed
	}

	var result bson.M
	err := m.database.RunCommand(ctx, bson.D{{Key: "dbStats", Value: 1}}).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to get database stats: %w", err)
	}

	stats := &documents.DBStats{}

	if collections, ok := result["collections"].(int32); ok {
		stats.Collections = int64(collections)
	}
	if objects, ok := result["objects"].(int64); ok {
		stats.Documents = objects
	} else if objects, ok := result["objects"].(int32); ok {
		stats.Documents = int64(objects)
	}
	if dataSize, ok := result["dataSize"].(int64); ok {
		stats.DataSize = dataSize
	} else if dataSize, ok := result["dataSize"].(int32); ok {
		stats.DataSize = int64(dataSize)
	}
	if indexSize, ok := result["indexSize"].(int64); ok {
		stats.IndexSize = indexSize
	} else if indexSize, ok := result["indexSize"].(int32); ok {
		stats.IndexSize = int64(indexSize)
	}
	if storageSize, ok := result["storageSize"].(int64); ok {
		stats.StorageSize = storageSize
	} else if storageSize, ok := result["storageSize"].(int32); ok {
		stats.StorageSize = int64(storageSize)
	}

	return stats, nil
}

// Helper methods

// buildConnectionString builds MongoDB connection string
func (m *MongoDBAdapter) buildConnectionString() string {
	if m.config.Options != nil {
		if connStr, ok := m.config.Options["connection_string"]; ok {
			return connStr
		}
	}

	// Build connection string
	return fmt.Sprintf("mongodb://%s:%d", m.config.Host, m.config.Port)
}

// convertBsonMToMap converts bson.M to map[string]interface{}
func (m *MongoDBAdapter) convertBsonMToMap(bsonDoc bson.M) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range bsonDoc {
		switch val := v.(type) {
		case primitive.ObjectID:
			result[k] = val.Hex()
		case bson.M:
			result[k] = m.convertBsonMToMap(val)
		case primitive.A:
			arr := make([]interface{}, len(val))
			for i, item := range val {
				if bsonM, ok := item.(bson.M); ok {
					arr[i] = m.convertBsonMToMap(bsonM)
				} else {
					arr[i] = item
				}
			}
			result[k] = arr
		default:
			result[k] = v
		}
	}
	return result
}
