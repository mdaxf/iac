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

package packagemgr

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/deployment/models"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DocumentPackager handles packaging of document database data
type DocumentPackager struct {
	docDB  *documents.DocDB
	logger logger.Log
}

// NewDocumentPackager creates a new document packager
func NewDocumentPackager(docDB *documents.DocDB, user string) *DocumentPackager {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "DocumentPackager"}

	return &DocumentPackager{
		docDB:  docDB,
		logger: iLog,
	}
}

// PackageCollections packages specified collections into a deployable package
func (dp *DocumentPackager) PackageCollections(packageName, version, createdBy string, filter models.PackageFilter) (*models.Package, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		dp.logger.PerformanceWithDuration("DocumentPackager.PackageCollections", elapsed)
	}()

	dp.logger.Info(fmt.Sprintf("Starting document packaging: %s v%s", packageName, version))

	pkg := &models.Package{
		ID:          uuid.New().String(),
		Name:        packageName,
		Version:     version,
		PackageType: "document",
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		Metadata:    make(map[string]interface{}),
		DocumentData: &models.DocumentPackage{
			Collections:  make([]models.CollectionData, 0),
			IDMappings:   make(map[string]models.IDMapping),
			References:   make([]models.DocumentReference, 0),
			SkipIDs:      filter.ExcludeFields["_id"] != nil,
			DatabaseType: "mongodb",
			DatabaseName: dp.docDB.DatabaseName,
		},
	}

	// Package each collection
	for _, collectionName := range filter.Collections {
		if err := dp.packageCollection(pkg, collectionName, filter); err != nil {
			dp.logger.Error(fmt.Sprintf("Error packaging collection %s: %v", collectionName, err))
			return nil, err
		}
	}

	// Build reference graph
	if err := dp.buildReferences(pkg, filter); err != nil {
		dp.logger.Error(fmt.Sprintf("Error building references: %v", err))
		// Non-fatal, continue
	}

	dp.logger.Info(fmt.Sprintf("Package created: %s with %d collections", pkg.ID, len(pkg.DocumentData.Collections)))
	return pkg, nil
}

// packageCollection packages a single collection's data
func (dp *DocumentPackager) packageCollection(pkg *models.Package, collectionName string, filter models.PackageFilter) error {
	dp.logger.Debug(fmt.Sprintf("Packaging collection: %s", collectionName))

	collection := dp.docDB.MongoDBDatabase.Collection(collectionName)

	collData := models.CollectionData{
		CollectionName: collectionName,
		Documents:      make([]map[string]interface{}, 0),
		IDField:        "_id",
		IndexInfo:      make([]models.IndexInfo, 0),
	}

	// Build filter query
	filterQuery := bson.M{}
	if whereClause, ok := filter.WhereClause[collectionName]; ok && whereClause != "" {
		// Parse WHERE clause as BSON
		if err := json.Unmarshal([]byte(whereClause), &filterQuery); err != nil {
			dp.logger.Warn(fmt.Sprintf("Failed to parse filter for %s: %v", collectionName, err))
		}
	}

	// Query documents
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filterQuery)
	if err != nil {
		return fmt.Errorf("failed to query collection %s: %w", collectionName, err)
	}
	defer cursor.Close(ctx)

	// Get excluded fields for this collection
	excludeFields := filter.ExcludeFields[collectionName]

	// Process documents
	for cursor.Next(ctx) {
		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			dp.logger.Error(fmt.Sprintf("Error decoding document: %v", err))
			continue
		}

		// Skip or preserve ID based on strategy
		if pkg.DocumentData.SkipIDs {
			delete(doc, "_id")
		}

		// Remove excluded fields
		for _, field := range excludeFields {
			delete(doc, field)
		}

		// Convert ObjectID to string for JSON serialization
		dp.convertObjectIDs(doc)

		collData.Documents = append(collData.Documents, doc)
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	collData.DocumentCount = len(collData.Documents)

	// Get index information
	indexInfo, err := dp.getIndexInfo(collection)
	if err != nil {
		dp.logger.Warn(fmt.Sprintf("Failed to get index info for %s: %v", collectionName, err))
	} else {
		collData.IndexInfo = indexInfo
	}

	pkg.DocumentData.Collections = append(pkg.DocumentData.Collections, collData)

	// Create ID mapping
	idMapping := models.IDMapping{
		CollectionName: collectionName,
		IDField:        "_id",
		IDType:         "objectid",
		Strategy:       dp.determineIDStrategy(pkg.DocumentData.SkipIDs),
	}
	pkg.DocumentData.IDMappings[collectionName] = idMapping

	return nil
}

// getIndexInfo retrieves index information for a collection
func (dp *DocumentPackager) getIndexInfo(collection *mongo.Collection) ([]models.IndexInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	indexInfo := make([]models.IndexInfo, 0)
	for cursor.Next(ctx) {
		var idx bson.M
		if err := cursor.Decode(&idx); err != nil {
			continue
		}

		info := models.IndexInfo{
			Name:   idx["name"].(string),
			Keys:   make(map[string]interface{}),
			Unique: false,
		}

		if keys, ok := idx["key"].(bson.M); ok {
			for k, v := range keys {
				info.Keys[k] = v
			}
		}

		if unique, ok := idx["unique"].(bool); ok {
			info.Unique = unique
		}

		indexInfo = append(indexInfo, info)
	}

	return indexInfo, nil
}

// buildReferences builds the reference graph between documents
func (dp *DocumentPackager) buildReferences(pkg *models.Package, filter models.PackageFilter) error {
	// Define known reference patterns for IAC collections
	referencePatterns := map[string][]ReferencePattern{
		"TranCode": {
			{Field: "workflow_id", TargetCollection: "Workflow", ReferenceType: "single"},
		},
		"Workflow": {
			{Field: "nodes.trancode_id", TargetCollection: "TranCode", ReferenceType: "array"},
		},
		"UI_Page": {
			{Field: "actions.trancode_id", TargetCollection: "TranCode", ReferenceType: "array"},
		},
		"UI_View": {
			{Field: "page_id", TargetCollection: "UI_Page", ReferenceType: "single"},
		},
	}

	for _, collData := range pkg.DocumentData.Collections {
		if patterns, ok := referencePatterns[collData.CollectionName]; ok {
			for _, pattern := range patterns {
				ref := models.DocumentReference{
					ID:               uuid.New().String(),
					SourceCollection: collData.CollectionName,
					SourceField:      pattern.Field,
					TargetCollection: pattern.TargetCollection,
					TargetIDField:    "_id",
					ReferenceType:    pattern.ReferenceType,
				}
				pkg.DocumentData.References = append(pkg.DocumentData.References, ref)
			}
		}
	}

	return nil
}

// ReferencePattern defines a reference pattern
type ReferencePattern struct {
	Field            string
	TargetCollection string
	ReferenceType    string
}

// convertObjectIDs recursively converts ObjectIDs to strings
func (dp *DocumentPackager) convertObjectIDs(doc map[string]interface{}) {
	for key, value := range doc {
		switch v := value.(type) {
		case primitive.ObjectID:
			doc[key] = v.Hex()
		case map[string]interface{}:
			dp.convertObjectIDs(v)
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					dp.convertObjectIDs(itemMap)
				} else if oid, ok := item.(primitive.ObjectID); ok {
					v[i] = oid.Hex()
				}
			}
		}
	}
}

// determineIDStrategy determines the ID generation strategy
func (dp *DocumentPackager) determineIDStrategy(skipIDs bool) string {
	if skipIDs {
		return "skip"
	}
	return "regenerate"
}

// ExportPackage exports package to JSON
func (dp *DocumentPackager) ExportPackage(pkg *models.Package) ([]byte, error) {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package: %w", err)
	}
	return data, nil
}

// ImportPackage imports package from JSON
func (dp *DocumentPackager) ImportPackage(data []byte) (*models.Package, error) {
	var pkg models.Package
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal package: %w", err)
	}
	return &pkg, nil
}

// PackageDevOpsObjects packages specific DevOps objects (TranCode, Workflow, etc.)
func (dp *DocumentPackager) PackageDevOpsObjects(objectType string, objectIDs []string, packageName, version, createdBy string) (*models.Package, error) {
	dp.logger.Info(fmt.Sprintf("Packaging DevOps objects: %s", objectType))

	filter := models.PackageFilter{
		Collections: []string{objectType},
		WhereClause: make(map[string]string),
	}

	// Build filter for specific IDs
	if len(objectIDs) > 0 {
		idsQuery := bson.M{"_id": bson.M{"$in": objectIDs}}
		queryBytes, _ := json.Marshal(idsQuery)
		filter.WhereClause[objectType] = string(queryBytes)
	}

	return dp.PackageCollections(packageName, version, createdBy, filter)
}

// CompareVersions compares two versions of a document for DevOps diff
func (dp *DocumentPackager) CompareVersions(collectionName string, docID string, version1, version2 time.Time) (map[string]interface{}, error) {
	// This would integrate with the DevOps versioning system
	// For now, return a placeholder
	return map[string]interface{}{
		"collection": collectionName,
		"document_id": docID,
		"version1":   version1,
		"version2":   version2,
		"changes":    []string{"Not yet implemented"},
	}, nil
}
