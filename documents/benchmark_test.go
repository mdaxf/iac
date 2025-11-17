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
	"testing"
	"time"
)

// BenchmarkDocumentInsertOne benchmarks single document inserts
func BenchmarkDocumentInsertOne(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_insert_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				doc := map[string]interface{}{
					"name":  fmt.Sprintf("User%d", i),
					"email": fmt.Sprintf("user%d@example.com", i),
					"age":   20 + (i % 50),
				}

				_, err := db.InsertOne(ctx, collectionName, doc)
				if err != nil {
					b.Fatalf("Insert failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentInsertMany benchmarks bulk document inserts
func BenchmarkDocumentInsertMany(b *testing.B) {
	configs := getDocumentTestConfigs()
	batchSize := 100

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_bulk_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				docs := make([]interface{}, batchSize)
				for j := 0; j < batchSize; j++ {
					docs[j] = map[string]interface{}{
						"name":  fmt.Sprintf("User%d", i*batchSize+j),
						"email": fmt.Sprintf("user%d@example.com", i*batchSize+j),
						"age":   20 + ((i*batchSize + j) % 50),
					}
				}

				_, err := db.InsertMany(ctx, collectionName, docs)
				if err != nil {
					b.Fatalf("Bulk insert failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentFindOne benchmarks single document lookups
func BenchmarkDocumentFindOne(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_find_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Insert test data
			docs := make([]interface{}, 1000)
			for i := 0; i < 1000; i++ {
				docs[i] = map[string]interface{}{
					"name":  fmt.Sprintf("User%d", i),
					"email": fmt.Sprintf("user%d@example.com", i),
					"age":   20 + (i % 50),
				}
			}
			db.InsertMany(ctx, collectionName, docs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				filter := map[string]interface{}{
					"name": fmt.Sprintf("User%d", i%1000),
				}

				_, err := db.FindOne(ctx, collectionName, filter)
				if err != nil {
					b.Fatalf("FindOne failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentFind benchmarks multi-document queries
func BenchmarkDocumentFind(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_query_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Insert test data
			docs := make([]interface{}, 1000)
			for i := 0; i < 1000; i++ {
				docs[i] = map[string]interface{}{
					"name":     fmt.Sprintf("User%d", i),
					"email":    fmt.Sprintf("user%d@example.com", i),
					"age":      20 + (i % 50),
					"category": string(rune('A' + (i % 10))),
				}
			}
			db.InsertMany(ctx, collectionName, docs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				filter := map[string]interface{}{
					"age": map[string]interface{}{
						"$gte": 30 + (i % 20),
					},
				}

				_, err := db.Find(ctx, collectionName, filter, nil)
				if err != nil {
					b.Fatalf("Find failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentUpdateOne benchmarks single document updates
func BenchmarkDocumentUpdateOne(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_update_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Insert test data
			docs := make([]interface{}, 100)
			for i := 0; i < 100; i++ {
				docs[i] = map[string]interface{}{
					"name":  fmt.Sprintf("User%d", i),
					"email": fmt.Sprintf("user%d@example.com", i),
					"age":   20 + (i % 50),
				}
			}
			db.InsertMany(ctx, collectionName, docs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				filter := map[string]interface{}{
					"name": fmt.Sprintf("User%d", i%100),
				}
				update := map[string]interface{}{
					"$set": map[string]interface{}{
						"email": fmt.Sprintf("updated%d@example.com", i),
					},
				}

				_, err := db.UpdateOne(ctx, collectionName, filter, update)
				if err != nil {
					b.Fatalf("UpdateOne failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentUpdateMany benchmarks bulk document updates
func BenchmarkDocumentUpdateMany(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_update_many_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Insert test data
			docs := make([]interface{}, 1000)
			for i := 0; i < 1000; i++ {
				docs[i] = map[string]interface{}{
					"name":     fmt.Sprintf("User%d", i),
					"email":    fmt.Sprintf("user%d@example.com", i),
					"age":      20 + (i % 50),
					"category": string(rune('A' + (i % 10))),
				}
			}
			db.InsertMany(ctx, collectionName, docs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				filter := map[string]interface{}{
					"category": string(rune('A' + (i % 10))),
				}
				update := map[string]interface{}{
					"$set": map[string]interface{}{
						"updated": true,
					},
				}

				_, err := db.UpdateMany(ctx, collectionName, filter, update)
				if err != nil {
					b.Fatalf("UpdateMany failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentAggregate benchmarks aggregation pipeline
func BenchmarkDocumentAggregate(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_agg_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Insert test data
			docs := make([]interface{}, 1000)
			for i := 0; i < 1000; i++ {
				docs[i] = map[string]interface{}{
					"category": string(rune('A' + (i % 10))),
					"value":    i % 100,
				}
			}
			db.InsertMany(ctx, collectionName, docs)

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

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := db.Aggregate(ctx, collectionName, pipeline)
				if err != nil {
					b.Fatalf("Aggregate failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentIndexedQuery benchmarks queries with indexes
func BenchmarkDocumentIndexedQuery(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_indexed_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Create index
			db.CreateIndex(ctx, collectionName, []string{"email"}, false, false)

			// Insert test data
			docs := make([]interface{}, 10000)
			for i := 0; i < 10000; i++ {
				docs[i] = map[string]interface{}{
					"name":  fmt.Sprintf("User%d", i),
					"email": fmt.Sprintf("user%d@example.com", i),
					"age":   20 + (i % 50),
				}
			}
			db.InsertMany(ctx, collectionName, docs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				filter := map[string]interface{}{
					"email": fmt.Sprintf("user%d@example.com", i%10000),
				}

				_, err := db.FindOne(ctx, collectionName, filter)
				if err != nil {
					b.Fatalf("Indexed query failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkDocumentConcurrentReads benchmarks concurrent read operations
func BenchmarkDocumentConcurrentReads(b *testing.B) {
	configs := getDocumentTestConfigs()

	for dbType, config := range configs {
		b.Run(dbType, func(b *testing.B) {
			if skipDocTest(dbType) {
				b.Skip("Database not available")
			}

			db, err := createDocumentDB(config)
			if err != nil {
				b.Fatalf("Failed to create database: %v", err)
			}
			defer db.Close()

			ctx := context.Background()
			collectionName := fmt.Sprintf("bench_concurrent_%s_%d", dbType, time.Now().Unix())
			defer cleanupCollection(db, collectionName)

			// Insert test data
			docs := make([]interface{}, 100)
			for i := 0; i < 100; i++ {
				docs[i] = map[string]interface{}{
					"name":  fmt.Sprintf("User%d", i),
					"email": fmt.Sprintf("user%d@example.com", i),
				}
			}
			db.InsertMany(ctx, collectionName, docs)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					filter := map[string]interface{}{
						"name": fmt.Sprintf("User%d", time.Now().UnixNano()%100),
					}

					_, err := db.FindOne(ctx, collectionName, filter)
					if err != nil {
						b.Fatalf("Concurrent read failed: %v", err)
					}
				}
			})
		})
	}
}

// Helper function to skip document database tests
func skipDocTest(dbType string) bool {
	skipEnv := fmt.Sprintf("SKIP_%s_TESTS", dbType)
	return getEnv(skipEnv, "false") == "true"
}
