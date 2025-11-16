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

package documents_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mdaxf/iac/documents"
)

// DocumentTestConfig holds test configuration for document databases
type DocumentTestConfig struct {
	Type     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// getDocumentTestConfigs returns test configurations for document databases
func getDocumentTestConfigs() map[string]DocumentTestConfig {
	return map[string]DocumentTestConfig{
		"mongodb": {
			Type:     "mongodb",
			Host:     getEnv("TEST_MONGODB_HOST", "localhost"),
			Port:     getEnvInt("TEST_MONGODB_PORT", 27017),
			Database: getEnv("TEST_MONGODB_DATABASE", "iac_test"),
			Username: getEnv("TEST_MONGODB_USERNAME", "iac_user"),
			Password: getEnv("TEST_MONGODB_PASSWORD", "iac_pass"),
		},
		"postgres_jsonb": {
			Type:     "postgres",
			Host:     getEnv("TEST_POSTGRES_HOST", "localhost"),
			Port:     getEnvInt("TEST_POSTGRES_PORT", 5432),
			Database: getEnv("TEST_POSTGRES_DATABASE", "iac_test"),
			Username: getEnv("TEST_POSTGRES_USERNAME", "iac_user"),
			Password: getEnv("TEST_POSTGRES_PASSWORD", "iac_pass"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		fmt.Sscanf(value, "%d", &intValue)
		return intValue
	}
	return defaultValue
}

// TestDocumentDatabaseConnection tests connection to document databases
func TestDocumentDatabaseConnection(t *testing.T) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests", dbType)
			}

			db, err := createDocumentDB(config)
			if err != nil {
				t.Fatalf("Failed to create %s database: %v", dbType, err)
			}
			defer db.Close()

			// Test ping
			ctx := context.Background()
			if err := db.Ping(ctx); err != nil {
				t.Fatalf("Failed to ping %s: %v", dbType, err)
			}

			t.Logf("✓ %s connection successful", dbType)
		})
	}
}

// TestDocumentCRUDOperations tests basic CRUD operations
func TestDocumentCRUDOperations(t *testing.T) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests", dbType)
			}

			db, err := createDocumentDB(config)
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("test_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Test INSERT ONE
			t.Run("InsertOne", func(t *testing.T) {
				doc := map[string]interface{}{
					"name":  "John Doe",
					"email": "john@example.com",
					"age":   30,
					"tags":  []string{"developer", "golang"},
				}

				id, err := db.InsertOne(ctx, collectionName, doc)
				if err != nil {
					t.Fatalf("Failed to insert document: %v", err)
				}

				if id == "" {
					t.Error("Expected non-empty document ID")
				}

				t.Logf("✓ Inserted document with ID: %s", id)
			})

			// Test INSERT MANY
			t.Run("InsertMany", func(t *testing.T) {
				docs := []interface{}{
					map[string]interface{}{"name": "Alice", "email": "alice@example.com", "age": 25},
					map[string]interface{}{"name": "Bob", "email": "bob@example.com", "age": 35},
					map[string]interface{}{"name": "Charlie", "email": "charlie@example.com", "age": 40},
				}

				ids, err := db.InsertMany(ctx, collectionName, docs)
				if err != nil {
					t.Fatalf("Failed to insert documents: %v", err)
				}

				if len(ids) != 3 {
					t.Errorf("Expected 3 IDs, got %d", len(ids))
				}

				t.Logf("✓ Inserted %d documents", len(ids))
			})

			// Test FIND ONE
			t.Run("FindOne", func(t *testing.T) {
				filter := map[string]interface{}{"name": "John Doe"}

				result, err := db.FindOne(ctx, collectionName, filter)
				if err != nil {
					t.Fatalf("Failed to find document: %v", err)
				}

				if result == nil {
					t.Fatal("Expected document not found")
				}

				if name, ok := result["name"].(string); !ok || name != "John Doe" {
					t.Errorf("Unexpected name: %v", result["name"])
				}

				t.Logf("✓ Found document: %v", result)
			})

			// Test FIND
			t.Run("Find", func(t *testing.T) {
				filter := map[string]interface{}{
					"age": map[string]interface{}{"$gte": 30},
				}

				results, err := db.Find(ctx, collectionName, filter, nil)
				if err != nil {
					t.Fatalf("Failed to find documents: %v", err)
				}

				if len(results) < 2 {
					t.Errorf("Expected at least 2 documents, got %d", len(results))
				}

				t.Logf("✓ Found %d documents with age >= 30", len(results))
			})

			// Test UPDATE ONE
			t.Run("UpdateOne", func(t *testing.T) {
				filter := map[string]interface{}{"name": "John Doe"}
				update := map[string]interface{}{
					"$set": map[string]interface{}{
						"email": "newemail@example.com",
						"age":   31,
					},
				}

				count, err := db.UpdateOne(ctx, collectionName, filter, update)
				if err != nil {
					t.Fatalf("Failed to update document: %v", err)
				}

				if count != 1 {
					t.Errorf("Expected 1 updated document, got %d", count)
				}

				// Verify update
				result, _ := db.FindOne(ctx, collectionName, filter)
				if email, ok := result["email"].(string); !ok || email != "newemail@example.com" {
					t.Errorf("Update not applied correctly: %v", result["email"])
				}

				t.Logf("✓ Updated document successfully")
			})

			// Test UPDATE MANY
			t.Run("UpdateMany", func(t *testing.T) {
				filter := map[string]interface{}{
					"age": map[string]interface{}{"$gte": 30},
				}
				update := map[string]interface{}{
					"$set": map[string]interface{}{
						"category": "senior",
					},
				}

				count, err := db.UpdateMany(ctx, collectionName, filter, update)
				if err != nil {
					t.Fatalf("Failed to update documents: %v", err)
				}

				if count < 2 {
					t.Errorf("Expected at least 2 updated documents, got %d", count)
				}

				t.Logf("✓ Updated %d documents", count)
			})

			// Test DELETE ONE
			t.Run("DeleteOne", func(t *testing.T) {
				filter := map[string]interface{}{"name": "Alice"}

				count, err := db.DeleteOne(ctx, collectionName, filter)
				if err != nil {
					t.Fatalf("Failed to delete document: %v", err)
				}

				if count != 1 {
					t.Errorf("Expected 1 deleted document, got %d", count)
				}

				t.Logf("✓ Deleted document successfully")
			})

			// Test DELETE MANY
			t.Run("DeleteMany", func(t *testing.T) {
				filter := map[string]interface{}{
					"age": map[string]interface{}{"$lt": 30},
				}

				count, err := db.DeleteMany(ctx, collectionName, filter)
				if err != nil {
					t.Fatalf("Failed to delete documents: %v", err)
				}

				t.Logf("✓ Deleted %d documents with age < 30", count)
			})

			// Test COUNT
			t.Run("Count", func(t *testing.T) {
				filter := map[string]interface{}{}

				count, err := db.Count(ctx, collectionName, filter)
				if err != nil {
					t.Fatalf("Failed to count documents: %v", err)
				}

				t.Logf("✓ Total documents: %d", count)
			})
		})
	}
}

// TestDocumentIndexOperations tests index creation and management
func TestDocumentIndexOperations(t *testing.T) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests", dbType)
			}

			db, err := createDocumentDB(config)
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("test_index_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Create single field index
			t.Run("CreateIndex", func(t *testing.T) {
				indexName, err := db.CreateIndex(ctx, collectionName, []string{"email"}, false, false)
				if err != nil {
					t.Fatalf("Failed to create index: %v", err)
				}

				t.Logf("✓ Created index: %s", indexName)
			})

			// Create unique index
			t.Run("CreateUniqueIndex", func(t *testing.T) {
				indexName, err := db.CreateIndex(ctx, collectionName, []string{"username"}, true, false)
				if err != nil {
					t.Fatalf("Failed to create unique index: %v", err)
				}

				t.Logf("✓ Created unique index: %s", indexName)
			})

			// Create compound index
			t.Run("CreateCompoundIndex", func(t *testing.T) {
				indexName, err := db.CreateIndex(ctx, collectionName, []string{"name", "age"}, false, false)
				if err != nil {
					t.Fatalf("Failed to create compound index: %v", err)
				}

				t.Logf("✓ Created compound index: %s", indexName)
			})

			// List indexes
			t.Run("ListIndexes", func(t *testing.T) {
				indexes, err := db.ListIndexes(ctx, collectionName)
				if err != nil {
					t.Fatalf("Failed to list indexes: %v", err)
				}

				if len(indexes) < 3 {
					t.Errorf("Expected at least 3 indexes, got %d", len(indexes))
				}

				for _, idx := range indexes {
					t.Logf("  Index: %s", idx)
				}

				t.Logf("✓ Listed %d indexes", len(indexes))
			})

			// Drop index
			t.Run("DropIndex", func(t *testing.T) {
				err := db.DropIndex(ctx, collectionName, "email_1")
				if err != nil {
					// Index name might be different based on database type
					t.Logf("Note: Could not drop index 'email_1': %v", err)
				} else {
					t.Logf("✓ Dropped index successfully")
				}
			})
		})
	}
}

// TestDocumentAggregation tests aggregation pipeline
func TestDocumentAggregation(t *testing.T) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		t.Run(dbType, func(t *testing.T) {
			if os.Getenv(fmt.Sprintf("SKIP_%s_TESTS", dbType)) == "true" {
				t.Skipf("Skipping %s tests", dbType)
			}

			db, err := createDocumentDB(config)
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("test_agg_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Insert test data
			docs := []interface{}{
				map[string]interface{}{"category": "A", "value": 10},
				map[string]interface{}{"category": "A", "value": 20},
				map[string]interface{}{"category": "B", "value": 15},
				map[string]interface{}{"category": "B", "value": 25},
				map[string]interface{}{"category": "C", "value": 30},
			}

			_, err = db.InsertMany(ctx, collectionName, docs)
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			// Test aggregation
			pipeline := []interface{}{
				map[string]interface{}{
					"$group": map[string]interface{}{
						"_id":   "$category",
						"total": map[string]interface{}{"$sum": "$value"},
						"count": map[string]interface{}{"$sum": 1},
					},
				},
				map[string]interface{}{
					"$sort": map[string]interface{}{"_id": 1},
				},
			}

			results, err := db.Aggregate(ctx, collectionName, pipeline)
			if err != nil {
				t.Fatalf("Failed to aggregate: %v", err)
			}

			if len(results) != 3 {
				t.Errorf("Expected 3 groups, got %d", len(results))
			}

			for _, result := range results {
				category := result["_id"]
				total := result["total"]
				count := result["count"]
				t.Logf("  Category %v: total=%v, count=%v", category, total, count)
			}

			t.Logf("✓ Aggregation successful")
		})
	}
}

// Helper functions

func createDocumentDB(config DocumentTestConfig) (documents.DocumentDB, error) {
	connStr := ""

	switch config.Type {
	case "mongodb":
		connStr = fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
			config.Username, config.Password, config.Host, config.Port, config.Database)
	case "postgres":
		connStr = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.Username, config.Password, config.Database)
	}

	return documents.NewDocumentDB(config.Type, connStr, config.Database)
}

func cleanupCollection(db documents.DocumentDB, collectionName string) {
	ctx := context.Background()
	db.DropCollection(ctx, collectionName)
}
