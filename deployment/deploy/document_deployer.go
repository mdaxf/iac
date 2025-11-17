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

package deploymgr

import (
	"context"
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

// DocumentDeployer handles deployment of document packages
type DocumentDeployer struct {
	docDB      *documents.DocDB
	logger     logger.Log
	idMappings map[string]map[interface{}]interface{} // Collection -> OldID -> NewID
}

// NewDocumentDeployer creates a new document deployer
func NewDocumentDeployer(docDB *documents.DocDB, user string) *DocumentDeployer {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "DocumentDeployer"}

	return &DocumentDeployer{
		docDB:      docDB,
		logger:     iLog,
		idMappings: make(map[string]map[interface{}]interface{}),
	}
}

// Deploy deploys a document package
func (dd *DocumentDeployer) Deploy(pkg *models.Package, options models.DeploymentOptions) (*models.DeploymentRecord, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		dd.logger.PerformanceWithDuration("DocumentDeployer.Deploy", elapsed)
	}()

	dd.logger.Info(fmt.Sprintf("Starting document deployment: %s v%s", pkg.Name, pkg.Version))

	// Create deployment record
	record := &models.DeploymentRecord{
		ID:              uuid.New().String(),
		PackageID:       pkg.ID,
		PackageName:     pkg.Name,
		PackageVersion:  pkg.Version,
		TargetDatabase:  dd.docDB.DatabaseName,
		DeployedAt:      time.Now(),
		Status:          "in_progress",
		IDMappingResult: make(map[string]map[interface{}]interface{}),
		ErrorLog:        make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}

	// Validate package
	if pkg.DocumentData == nil {
		return record, fmt.Errorf("package does not contain document data")
	}

	// Check dry run
	if options.DryRun {
		dd.logger.Info("Dry run mode - validating only")
		if err := dd.validatePackage(pkg, options); err != nil {
			record.Status = "failed"
			record.ErrorLog = append(record.ErrorLog, err.Error())
			return record, err
		}
		record.Status = "validated"
		return record, nil
	}

	// Deploy each collection
	for _, collData := range pkg.DocumentData.Collections {
		dd.logger.Debug(fmt.Sprintf("Deploying collection: %s", collData.CollectionName))

		idMapping := pkg.DocumentData.IDMappings[collData.CollectionName]
		if err := dd.deployCollection(collData, idMapping, options); err != nil {
			errMsg := fmt.Sprintf("Failed to deploy collection %s: %v", collData.CollectionName, err)
			dd.logger.Error(errMsg)
			record.ErrorLog = append(record.ErrorLog, errMsg)

			if !options.ContinueOnError {
				record.Status = "failed"
				return record, err
			}
		}

		// Store ID mappings in record
		if collMappings, ok := dd.idMappings[collData.CollectionName]; ok {
			record.IDMappingResult[collData.CollectionName] = collMappings
		}
	}

	// Rebuild references
	if err := dd.rebuildReferences(pkg.DocumentData.References, options); err != nil {
		errMsg := fmt.Sprintf("Failed to rebuild references: %v", err)
		dd.logger.Error(errMsg)
		record.ErrorLog = append(record.ErrorLog, errMsg)

		if !options.ContinueOnError {
			record.Status = "failed"
			return record, err
		}
	}

	// Rebuild indexes if requested
	if options.RebuildIndexes {
		for _, collData := range pkg.DocumentData.Collections {
			if err := dd.rebuildIndexes(collData); err != nil {
				dd.logger.Warn(fmt.Sprintf("Failed to rebuild indexes for %s: %v", collData.CollectionName, err))
			}
		}
	}

	record.Status = "completed"
	dd.logger.Info(fmt.Sprintf("Document deployment completed: %s", record.ID))
	return record, nil
}

// deployCollection deploys a single collection
func (dd *DocumentDeployer) deployCollection(collData models.CollectionData, idMapping models.IDMapping, options models.DeploymentOptions) error {
	collection := dd.docDB.MongoDBDatabase.Collection(collData.CollectionName)

	// Initialize ID mapping for this collection
	if _, ok := dd.idMappings[collData.CollectionName]; !ok {
		dd.idMappings[collData.CollectionName] = make(map[interface{}]interface{})
	}

	batchSize := options.BatchSize
	if batchSize <= 0 {
		batchSize = 100
	}

	// Process documents in batches
	for i := 0; i < len(collData.Documents); i += batchSize {
		end := i + batchSize
		if end > len(collData.Documents) {
			end = len(collData.Documents)
		}

		batch := collData.Documents[i:end]
		if err := dd.deployBatch(collection, batch, idMapping, options); err != nil {
			return fmt.Errorf("failed to deploy batch %d-%d: %w", i, end, err)
		}
	}

	return nil
}

// deployBatch deploys a batch of documents
func (dd *DocumentDeployer) deployBatch(collection *mongo.Collection, docs []map[string]interface{}, idMapping models.IDMapping, options models.DeploymentOptions) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, doc := range docs {
		if err := dd.deployDocument(ctx, collection, doc, idMapping, options); err != nil {
			if !options.ContinueOnError {
				return err
			}
			dd.logger.Warn(fmt.Sprintf("Failed to deploy document: %v", err))
		}
	}

	return nil
}

// deployDocument deploys a single document
func (dd *DocumentDeployer) deployDocument(ctx context.Context, collection *mongo.Collection, doc map[string]interface{}, idMapping models.IDMapping, options models.DeploymentOptions) error {
	// Store old ID value
	oldID := doc[idMapping.IDField]

	// Handle ID based on strategy
	newID, err := dd.handleIDStrategy(doc, idMapping, options)
	if err != nil {
		return err
	}

	// Check if document exists
	if newID != nil {
		exists, err := dd.documentExists(ctx, collection, idMapping.IDField, newID)
		if err != nil {
			return err
		}

		if exists {
			if options.SkipExisting {
				dd.logger.Debug(fmt.Sprintf("Skipping existing document in %s", collection.Name()))
				// Still map the ID even if skipping
				dd.idMappings[collection.Name()][oldID] = newID
				return nil
			}

			if options.UpdateExisting {
				return dd.updateDocument(ctx, collection, doc, idMapping.IDField, newID)
			}

			return fmt.Errorf("document already exists in %s", collection.Name())
		}
	}

	// Insert new document
	result, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	// Map old ID to new ID
	if oldID != nil {
		dd.idMappings[collection.Name()][oldID] = result.InsertedID
	}

	return nil
}

// handleIDStrategy handles ID generation based on strategy
func (dd *DocumentDeployer) handleIDStrategy(doc map[string]interface{}, idMapping models.IDMapping, options models.DeploymentOptions) (interface{}, error) {
	switch idMapping.Strategy {
	case "regenerate":
		// Generate new ObjectID
		newID := primitive.NewObjectID()
		doc[idMapping.IDField] = newID
		return newID, nil

	case "skip":
		// Remove ID field and let MongoDB generate it
		delete(doc, idMapping.IDField)
		return nil, nil

	case "preserve":
		// Keep original ID value
		if idStr, ok := doc[idMapping.IDField].(string); ok {
			// Convert string back to ObjectID if needed
			if idMapping.IDType == "objectid" {
				oid, err := primitive.ObjectIDFromHex(idStr)
				if err != nil {
					return nil, fmt.Errorf("failed to convert ID: %w", err)
				}
				doc[idMapping.IDField] = oid
				return oid, nil
			}
		}
		return doc[idMapping.IDField], nil

	default:
		return nil, fmt.Errorf("unknown ID strategy: %s", idMapping.Strategy)
	}
}

// updateDocument updates an existing document
func (dd *DocumentDeployer) updateDocument(ctx context.Context, collection *mongo.Collection, doc map[string]interface{}, idField string, idValue interface{}) error {
	// Validate idField to prevent NoSQL injection
	if idField != "_id" && idField != "id" {
		return fmt.Errorf("invalid ID field name: %s (only '_id' and 'id' are allowed)", idField)
	}

	filter := bson.M{idField: idValue}

	// Remove ID from update document
	updateDoc := make(map[string]interface{})
	for k, v := range doc {
		if k != idField {
			updateDoc[k] = v
		}
	}

	update := bson.M{"$set": updateDoc}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// documentExists checks if a document exists
func (dd *DocumentDeployer) documentExists(ctx context.Context, collection *mongo.Collection, idField string, idValue interface{}) (bool, error) {
	// Validate idField to prevent NoSQL injection
	// Only allow standard MongoDB ID field names
	if idField != "_id" && idField != "id" {
		return false, fmt.Errorf("invalid ID field name: %s (only '_id' and 'id' are allowed)", idField)
	}

	filter := bson.M{idField: idValue}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// rebuildReferences rebuilds document references
func (dd *DocumentDeployer) rebuildReferences(references []models.DocumentReference, options models.DeploymentOptions) error {
	dd.logger.Info("Rebuilding document references")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, ref := range references {
		// Validate field names to prevent NoSQL injection
		if err := dd.validateFieldName(ref.SourceField); err != nil {
			dd.logger.Error(fmt.Sprintf("Invalid source field in reference: %v", err))
			continue
		}

		// Get source collection mappings
		sourceMappings, ok := dd.idMappings[ref.SourceCollection]
		if !ok {
			continue
		}

		// Get target collection mappings
		targetMappings, ok := dd.idMappings[ref.TargetCollection]
		if !ok {
			continue
		}

		collection := dd.docDB.MongoDBDatabase.Collection(ref.SourceCollection)

		// Update references based on type
		switch ref.ReferenceType {
		case "single":
			// Update single reference field
			for oldID, newID := range sourceMappings {
				// Find documents with old reference
				filter := bson.M{ref.SourceField: oldID}

				// Update to new reference
				if mappedTarget, ok := targetMappings[oldID]; ok {
					update := bson.M{"$set": bson.M{ref.SourceField: mappedTarget}}
					_, err := collection.UpdateMany(ctx, filter, update)
					if err != nil {
						dd.logger.Warn(fmt.Sprintf("Failed to update reference: %v", err))
					}
				}

				_ = newID // Suppress unused warning
			}

		case "array":
			// Update array reference field
			for oldID, _ := range sourceMappings {
				// Find documents with old reference in array
				filter := bson.M{ref.SourceField: oldID}

				// Replace old ID with new ID in array
				if mappedTarget, ok := targetMappings[oldID]; ok {
					update := bson.M{"$set": bson.M{ref.SourceField + ".$": mappedTarget}}
					_, err := collection.UpdateMany(ctx, filter, update)
					if err != nil {
						dd.logger.Warn(fmt.Sprintf("Failed to update array reference: %v", err))
					}
				}
			}
		}
	}

	return nil
}

// rebuildIndexes rebuilds indexes for a collection
func (dd *DocumentDeployer) rebuildIndexes(collData models.CollectionData) error {
	dd.logger.Debug(fmt.Sprintf("Rebuilding indexes for %s", collData.CollectionName))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := dd.docDB.MongoDBDatabase.Collection(collData.CollectionName)

	for _, indexInfo := range collData.IndexInfo {
		// Skip default _id index
		if indexInfo.Name == "_id_" {
			continue
		}

		indexModel := mongo.IndexModel{
			Keys: indexInfo.Keys,
		}

		if indexInfo.Unique {
			indexModel.Options = options.Index().SetUnique(true)
		}

		_, err := collection.Indexes().CreateOne(ctx, indexModel)
		if err != nil {
			return fmt.Errorf("failed to create index %s: %w", indexInfo.Name, err)
		}
	}

	return nil
}

// validatePackage validates a package before deployment
func (dd *DocumentDeployer) validatePackage(pkg *models.Package, options models.DeploymentOptions) error {
	dd.logger.Info("Validating document package")

	// Check package structure
	if pkg.DocumentData == nil {
		return fmt.Errorf("package contains no document data")
	}

	// Validate collections
	for _, collData := range pkg.DocumentData.Collections {
		dd.logger.Debug(fmt.Sprintf("Validating collection: %s", collData.CollectionName))

		// Validate document structure
		for i, doc := range collData.Documents {
			if doc == nil {
				return fmt.Errorf("null document at index %d in collection %s", i, collData.CollectionName)
			}
		}
	}

	// Validate references
	if options.ValidateReferences {
		for _, ref := range pkg.DocumentData.References {
			dd.logger.Debug(fmt.Sprintf("Validating reference: %s.%s -> %s",
				ref.SourceCollection, ref.SourceField, ref.TargetCollection))
		}
	}

	return nil
}

// validateFieldName validates MongoDB field names to prevent NoSQL injection
// Field names must not start with $ (MongoDB operator) and should not contain null bytes
func (dd *DocumentDeployer) validateFieldName(fieldName string) error {
	if fieldName == "" {
		return fmt.Errorf("field name cannot be empty")
	}

	// Check for MongoDB operators (fields starting with $)
	if len(fieldName) > 0 && fieldName[0] == '$' {
		return fmt.Errorf("field name cannot start with '$': %s", fieldName)
	}

	// Check for null bytes
	for i := 0; i < len(fieldName); i++ {
		if fieldName[i] == 0 {
			return fmt.Errorf("field name cannot contain null bytes: %s", fieldName)
		}
	}

	return nil
}

// Rollback rolls back a deployment
func (dd *DocumentDeployer) Rollback(record *models.DeploymentRecord) error {
	dd.logger.Info(fmt.Sprintf("Rolling back document deployment: %s", record.ID))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Delete inserted documents using the ID mappings
	for collectionName, idMappings := range record.IDMappingResult {
		collection := dd.docDB.MongoDBDatabase.Collection(collectionName)

		// Collect new IDs
		newIDs := make([]interface{}, 0)
		for _, newID := range idMappings {
			newIDs = append(newIDs, newID)
		}

		// Delete documents
		filter := bson.M{"_id": bson.M{"$in": newIDs}}
		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			dd.logger.Error(fmt.Sprintf("Failed to rollback collection %s: %v", collectionName, err))
			continue
		}

		dd.logger.Info(fmt.Sprintf("Rolled back %d documents from %s", result.DeletedCount, collectionName))
	}

	record.Status = "rolled_back"
	return nil
}
