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

package devops

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ObjectType represents types of objects that can be version controlled
type ObjectType string

const (
	ObjectTypeTranCode   ObjectType = "TranCode"
	ObjectTypeWorkflow   ObjectType = "Workflow"
	ObjectTypeUIPage     ObjectType = "UI_Page"
	ObjectTypeUIView     ObjectType = "UI_View"
	ObjectTypeWhiteboard ObjectType = "Whiteboard"
	ObjectTypeProcess    ObjectType = "Process_Plan"
)

// VersionControl manages version control for document objects
type VersionControl struct {
	docDB  *documents.DocDB
	logger logger.Log
}

// NewVersionControl creates a new version control manager
func NewVersionControl(docDB *documents.DocDB, user string) *VersionControl {
	iLog := logger.Log{ModuleName: logger.Framework, User: user, ControllerName: "VersionControl"}

	return &VersionControl{
		docDB:  docDB,
		logger: iLog,
	}
}

// VersionRecord represents a version of an object
type VersionRecord struct {
	ID              string                 `bson:"_id" json:"id"`
	ObjectID        string                 `bson:"object_id" json:"object_id"`
	ObjectType      ObjectType             `bson:"object_type" json:"object_type"`
	Version         string                 `bson:"version" json:"version"`
	VersionNumber   int                    `bson:"version_number" json:"version_number"`
	Content         map[string]interface{} `bson:"content" json:"content"`
	ContentHash     string                 `bson:"content_hash" json:"content_hash"`
	CreatedAt       time.Time              `bson:"created_at" json:"created_at"`
	CreatedBy       string                 `bson:"created_by" json:"created_by"`
	CommitMessage   string                 `bson:"commit_message" json:"commit_message"`
	Tags            []string               `bson:"tags,omitempty" json:"tags,omitempty"`
	Branch          string                 `bson:"branch" json:"branch"`
	ParentVersion   string                 `bson:"parent_version,omitempty" json:"parent_version,omitempty"`
	MergedFrom      []string               `bson:"merged_from,omitempty" json:"merged_from,omitempty"`
	IsDeleted       bool                   `bson:"is_deleted" json:"is_deleted"`
	Metadata        map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// Branch represents a development branch
type Branch struct {
	ID          string    `bson:"_id" json:"id"`
	Name        string    `bson:"name" json:"name"`
	ObjectID    string    `bson:"object_id" json:"object_id"`
	ObjectType  ObjectType `bson:"object_type" json:"object_type"`
	BaseVersion string    `bson:"base_version" json:"base_version"`
	HeadVersion string    `bson:"head_version" json:"head_version"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	CreatedBy   string    `bson:"created_by" json:"created_by"`
	IsActive    bool      `bson:"is_active" json:"is_active"`
	IsMerged    bool      `bson:"is_merged" json:"is_merged"`
}

// ChangeLog represents a changelog entry
type ChangeLog struct {
	ID              string                 `bson:"_id" json:"id"`
	ObjectID        string                 `bson:"object_id" json:"object_id"`
	ObjectType      ObjectType             `bson:"object_type" json:"object_type"`
	Action          string                 `bson:"action" json:"action"` // "create", "update", "delete", "merge"
	FromVersion     string                 `bson:"from_version,omitempty" json:"from_version,omitempty"`
	ToVersion       string                 `bson:"to_version" json:"to_version"`
	Changes         []FieldChange          `bson:"changes" json:"changes"`
	ChangedBy       string                 `bson:"changed_by" json:"changed_by"`
	ChangedAt       time.Time              `bson:"changed_at" json:"changed_at"`
	CommitMessage   string                 `bson:"commit_message" json:"commit_message"`
	Metadata        map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// FieldChange represents a change to a specific field
type FieldChange struct {
	Field    string      `bson:"field" json:"field"`
	OldValue interface{} `bson:"old_value,omitempty" json:"old_value,omitempty"`
	NewValue interface{} `bson:"new_value,omitempty" json:"new_value,omitempty"`
	Action   string      `bson:"action" json:"action"` // "add", "modify", "delete"
}

// Commit commits a new version of an object
func (vc *VersionControl) Commit(objectType ObjectType, objectID string, content map[string]interface{}, commitMessage, branch, createdBy string) (*VersionRecord, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		vc.logger.PerformanceWithDuration("VersionControl.Commit", elapsed)
	}()

	vc.logger.Info(fmt.Sprintf("Committing version for %s: %s", objectType, objectID))

	// Get collection
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Versions")

	// Get latest version
	latestVersion, err := vc.GetLatestVersion(objectType, objectID, branch)
	var versionNumber int
	var parentVersion string

	if err == nil && latestVersion != nil {
		versionNumber = latestVersion.VersionNumber + 1
		parentVersion = latestVersion.ID
	} else {
		versionNumber = 1
	}

	// Calculate content hash
	contentHash := vc.calculateHash(content)

	// Check if content has changed
	if latestVersion != nil && latestVersion.ContentHash == contentHash {
		vc.logger.Info("No changes detected, skipping commit")
		return latestVersion, nil
	}

	// Create version record
	versionRecord := &VersionRecord{
		ID:            uuid.New().String(),
		ObjectID:      objectID,
		ObjectType:    objectType,
		Version:       fmt.Sprintf("v%d", versionNumber),
		VersionNumber: versionNumber,
		Content:       content,
		ContentHash:   contentHash,
		CreatedAt:     time.Now(),
		CreatedBy:     createdBy,
		CommitMessage: commitMessage,
		Branch:        branch,
		ParentVersion: parentVersion,
		IsDeleted:     false,
		Metadata:      make(map[string]interface{}),
	}

	// Insert version record
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, versionRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to insert version record: %w", err)
	}

	// Create changelog
	if latestVersion != nil {
		changes := vc.compareVersions(latestVersion.Content, content)
		changelog := &ChangeLog{
			ID:            uuid.New().String(),
			ObjectID:      objectID,
			ObjectType:    objectType,
			Action:        "update",
			FromVersion:   latestVersion.Version,
			ToVersion:     versionRecord.Version,
			Changes:       changes,
			ChangedBy:     createdBy,
			ChangedAt:     time.Now(),
			CommitMessage: commitMessage,
		}

		if err := vc.saveChangelog(changelog); err != nil {
			vc.logger.Warn(fmt.Sprintf("Failed to save changelog: %v", err))
		}
	} else {
		// First version
		changelog := &ChangeLog{
			ID:            uuid.New().String(),
			ObjectID:      objectID,
			ObjectType:    objectType,
			Action:        "create",
			ToVersion:     versionRecord.Version,
			Changes:       []FieldChange{},
			ChangedBy:     createdBy,
			ChangedAt:     time.Now(),
			CommitMessage: commitMessage,
		}

		if err := vc.saveChangelog(changelog); err != nil {
			vc.logger.Warn(fmt.Sprintf("Failed to save changelog: %v", err))
		}
	}

	// Update branch head
	if err := vc.updateBranchHead(objectType, objectID, branch, versionRecord.ID, createdBy); err != nil {
		vc.logger.Warn(fmt.Sprintf("Failed to update branch head: %v", err))
	}

	vc.logger.Info(fmt.Sprintf("Version committed: %s", versionRecord.ID))
	return versionRecord, nil
}

// GetLatestVersion retrieves the latest version of an object
func (vc *VersionControl) GetLatestVersion(objectType ObjectType, objectID, branch string) (*VersionRecord, error) {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Versions")

	filter := bson.M{
		"object_id":   objectID,
		"object_type": objectType,
		"branch":      branch,
		"is_deleted":  false,
	}

	opts := options.FindOne().SetSort(bson.D{{"version_number", -1}})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var version VersionRecord
	err := collection.FindOne(ctx, filter, opts).Decode(&version)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &version, nil
}

// GetVersion retrieves a specific version
func (vc *VersionControl) GetVersion(versionID string) (*VersionRecord, error) {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Versions")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var version VersionRecord
	err := collection.FindOne(ctx, bson.M{"_id": versionID}).Decode(&version)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

// ListVersions lists all versions of an object
func (vc *VersionControl) ListVersions(objectType ObjectType, objectID, branch string) ([]*VersionRecord, error) {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Versions")

	filter := bson.M{
		"object_id":   objectID,
		"object_type": objectType,
		"is_deleted":  false,
	}

	if branch != "" {
		filter["branch"] = branch
	}

	opts := options.Find().SetSort(bson.D{{"version_number", -1}})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	versions := make([]*VersionRecord, 0)
	for cursor.Next(ctx) {
		var version VersionRecord
		if err := cursor.Decode(&version); err != nil {
			continue
		}
		versions = append(versions, &version)
	}

	return versions, nil
}

// CreateBranch creates a new branch
func (vc *VersionControl) CreateBranch(objectType ObjectType, objectID, branchName, baseVersion, createdBy string) (*Branch, error) {
	vc.logger.Info(fmt.Sprintf("Creating branch %s for %s: %s", branchName, objectType, objectID))

	collection := vc.docDB.MongoDBDatabase.Collection("Object_Branches")

	branch := &Branch{
		ID:          uuid.New().String(),
		Name:        branchName,
		ObjectID:    objectID,
		ObjectType:  objectType,
		BaseVersion: baseVersion,
		HeadVersion: baseVersion,
		CreatedAt:   time.Now(),
		CreatedBy:   createdBy,
		IsActive:    true,
		IsMerged:    false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, branch)
	if err != nil {
		return nil, fmt.Errorf("failed to create branch: %w", err)
	}

	return branch, nil
}

// MergeBranch merges a branch into another branch
func (vc *VersionControl) MergeBranch(objectType ObjectType, objectID, sourceBranch, targetBranch, mergedBy string) (*VersionRecord, error) {
	vc.logger.Info(fmt.Sprintf("Merging %s into %s for %s: %s", sourceBranch, targetBranch, objectType, objectID))

	// Get latest version from source branch
	sourceVersion, err := vc.GetLatestVersion(objectType, objectID, sourceBranch)
	if err != nil || sourceVersion == nil {
		return nil, fmt.Errorf("failed to get source branch version: %w", err)
	}

	// Get latest version from target branch
	targetVersion, err := vc.GetLatestVersion(objectType, objectID, targetBranch)
	if err != nil {
		return nil, fmt.Errorf("failed to get target branch version: %w", err)
	}

	// Merge content (simplified - in practice would need conflict resolution)
	mergedContent := vc.mergeContent(targetVersion.Content, sourceVersion.Content)

	// Commit merged version to target branch
	mergedVersion, err := vc.Commit(
		objectType,
		objectID,
		mergedContent,
		fmt.Sprintf("Merge %s into %s", sourceBranch, targetBranch),
		targetBranch,
		mergedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to commit merge: %w", err)
	}

	// Mark source branch as merged
	if err := vc.markBranchMerged(objectType, objectID, sourceBranch); err != nil {
		vc.logger.Warn(fmt.Sprintf("Failed to mark branch as merged: %v", err))
	}

	// Update merged version record
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Versions")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$set": bson.M{
			"merged_from": []string{sourceVersion.ID},
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"_id": mergedVersion.ID}, update)
	if err != nil {
		vc.logger.Warn(fmt.Sprintf("Failed to update merge metadata: %v", err))
	}

	return mergedVersion, nil
}

// TagVersion tags a version with a label
func (vc *VersionControl) TagVersion(versionID string, tags []string) error {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Versions")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	update := bson.M{
		"$addToSet": bson.M{
			"tags": bson.M{"$each": tags},
		},
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": versionID}, update)
	return err
}

// Diff compares two versions
func (vc *VersionControl) Diff(version1ID, version2ID string) ([]FieldChange, error) {
	v1, err := vc.GetVersion(version1ID)
	if err != nil {
		return nil, err
	}

	v2, err := vc.GetVersion(version2ID)
	if err != nil {
		return nil, err
	}

	return vc.compareVersions(v1.Content, v2.Content), nil
}

// GetChangelog retrieves changelog for an object
func (vc *VersionControl) GetChangelog(objectType ObjectType, objectID string, limit int) ([]*ChangeLog, error) {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Changelogs")

	filter := bson.M{
		"object_id":   objectID,
		"object_type": objectType,
	}

	opts := options.Find().SetSort(bson.D{{"changed_at", -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	changelogs := make([]*ChangeLog, 0)
	for cursor.Next(ctx) {
		var changelog ChangeLog
		if err := cursor.Decode(&changelog); err != nil {
			continue
		}
		changelogs = append(changelogs, &changelog)
	}

	return changelogs, nil
}

// Helper functions

func (vc *VersionControl) calculateHash(content map[string]interface{}) string {
	data, _ := json.Marshal(content)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (vc *VersionControl) compareVersions(old, new map[string]interface{}) []FieldChange {
	changes := make([]FieldChange, 0)

	// Check for modified and deleted fields
	for key, oldVal := range old {
		if newVal, ok := new[key]; ok {
			// Field exists in both - check if modified
			if !vc.isEqual(oldVal, newVal) {
				changes = append(changes, FieldChange{
					Field:    key,
					OldValue: oldVal,
					NewValue: newVal,
					Action:   "modify",
				})
			}
		} else {
			// Field deleted
			changes = append(changes, FieldChange{
				Field:    key,
				OldValue: oldVal,
				Action:   "delete",
			})
		}
	}

	// Check for added fields
	for key, newVal := range new {
		if _, ok := old[key]; !ok {
			changes = append(changes, FieldChange{
				Field:    key,
				NewValue: newVal,
				Action:   "add",
			})
		}
	}

	return changes
}

func (vc *VersionControl) isEqual(a, b interface{}) bool {
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

func (vc *VersionControl) mergeContent(base, incoming map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})

	// Copy base
	for k, v := range base {
		merged[k] = v
	}

	// Apply incoming changes (simple strategy - incoming wins)
	for k, v := range incoming {
		merged[k] = v
	}

	return merged
}

func (vc *VersionControl) saveChangelog(changelog *ChangeLog) error {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Changelogs")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, changelog)
	return err
}

func (vc *VersionControl) updateBranchHead(objectType ObjectType, objectID, branch, headVersion, updatedBy string) error {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Branches")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"object_id":   objectID,
		"object_type": objectType,
		"name":        branch,
		"is_active":   true,
	}

	update := bson.M{
		"$set": bson.M{
			"head_version": headVersion,
		},
	}

	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (vc *VersionControl) markBranchMerged(objectType ObjectType, objectID, branch string) error {
	collection := vc.docDB.MongoDBDatabase.Collection("Object_Branches")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"object_id":   objectID,
		"object_type": objectType,
		"name":        branch,
	}

	update := bson.M{
		"$set": bson.M{
			"is_merged": true,
			"is_active": false,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// Revert reverts an object to a specific version
func (vc *VersionControl) Revert(objectType ObjectType, objectID, versionID, revertedBy string) (*VersionRecord, error) {
	vc.logger.Info(fmt.Sprintf("Reverting %s: %s to version %s", objectType, objectID, versionID))

	// Get the version to revert to
	version, err := vc.GetVersion(versionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	// Commit the reverted content as a new version
	return vc.Commit(
		objectType,
		objectID,
		version.Content,
		fmt.Sprintf("Revert to version %s", version.Version),
		version.Branch,
		revertedBy,
	)
}

// AutoCommit automatically commits changes from the main collection
func (vc *VersionControl) AutoCommit(objectType ObjectType, objectID, commitMessage, createdBy string) (*VersionRecord, error) {
	// Get current document from main collection
	collection := vc.docDB.MongoDBDatabase.Collection(string(objectType))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var objID interface{}
	if oid, err := primitive.ObjectIDFromHex(objectID); err == nil {
		objID = oid
	} else {
		objID = objectID
	}

	var doc map[string]interface{}
	err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&doc)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	// Commit to version control
	return vc.Commit(objectType, objectID, doc, commitMessage, "main", createdBy)
}
