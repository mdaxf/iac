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

package deployment

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	dbconn "github.com/mdaxf/iac/databases"
	deploymodels "github.com/mdaxf/iac/deployment/models"
	packagemgr "github.com/mdaxf/iac/deployment/package"
	"github.com/mdaxf/iac/deployment/repository"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"
)

const packageDefinitionsCollection = "package_definitions"

// PackageDefinitionController handles CRUD for reusable package definitions.
// Dual persistence: SQL (iacpackagedefs) stores stable metadata + version number;
// MongoDB (package_definitions) stores the full config per definition.
type PackageDefinitionController struct{}

func definitionCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func definitionUser(c *gin.Context) string {
	user, _ := c.Get("user")
	if s := fmt.Sprintf("%v", user); s != "" && s != "<nil>" {
		return s
	}
	return "System"
}

// ─── ListDefinitions ─────────────────────────────────────────────────────────

// ListDefinitions godoc
// @Summary List all active package definitions
// @Router /api/package-definitions [get]
func (pdc *PackageDefinitionController) ListDefinitions(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageDefinitionController.ListDefinitions"}

	if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Document database not available"})
		return
	}

	ctx, cancel := definitionCtx()
	defer cancel()

	coll := documents.DocDBCon.MongoDBDatabase.Collection(packageDefinitionsCollection)
	cursor, err := coll.Find(ctx, bson.M{"active": true})
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to list definitions: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var defs []deploymodels.PackageDefinitionDoc
	if err := cursor.All(ctx, &defs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if defs == nil {
		defs = []deploymodels.PackageDefinitionDoc{}
	}
	c.JSON(http.StatusOK, gin.H{"data": defs, "count": len(defs)})
}

// ─── GetDefinition ────────────────────────────────────────────────────────────

// GetDefinition godoc
// @Summary Get a package definition by ID
// @Router /api/package-definitions/:id [get]
func (pdc *PackageDefinitionController) GetDefinition(c *gin.Context) {
	id := c.Param("id")
	if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Document database not available"})
		return
	}

	ctx, cancel := definitionCtx()
	defer cancel()

	coll := documents.DocDBCon.MongoDBDatabase.Collection(packageDefinitionsCollection)
	var def deploymodels.PackageDefinitionDoc
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&def); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Definition not found: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": def})
}

// ─── CreateDefinition ────────────────────────────────────────────────────────

// CreateDefinition godoc
// @Summary Create a new package definition (SQL + MongoDB)
// @Router /api/package-definitions [post]
func (pdc *PackageDefinitionController) CreateDefinition(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageDefinitionController.CreateDefinition"}

	if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Document database not available"})
		return
	}

	var def deploymodels.PackageDefinitionDoc
	if err := c.ShouldBindJSON(&def); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	userName := definitionUser(c)
	now := time.Now().Format(time.RFC3339)
	defID := uuid.New().String()

	def.ID = defID
	def.VersionNumber = 1
	def.Active = true
	def.CreatedBy = userName
	def.CreatedOn = now
	def.ModifiedBy = userName
	def.ModifiedOn = now

	// 1. Persist config in MongoDB
	ctx, cancel := definitionCtx()
	defer cancel()

	coll := documents.DocDBCon.MongoDBDatabase.Collection(packageDefinitionsCollection)
	if _, err := coll.InsertOne(ctx, def); err != nil {
		iLog.Error(fmt.Sprintf("Failed to create definition in MongoDB: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 2. Persist metadata in SQL (best-effort: MongoDB doc is the source of truth)
	if dbconn.DB != nil {
		dbTx, txErr := dbconn.DB.Begin()
		if txErr == nil {
			sqlRepo := repository.NewPackageDefinitionRepository(userName, dbTx)
			sqlRec := &repository.PackageDefinitionRecord{
				ID:          defID,
				Name:        def.Name,
				PackageType: def.PackageType,
				Description: def.Description,
				Environment: def.Environment,
				LatestVersion: 1,
				Active:      true,
				CreatedBy:   userName,
				ModifiedBy:  userName,
			}
			if saveErr := sqlRepo.SaveDefinition(sqlRec); saveErr != nil {
				iLog.Error(fmt.Sprintf("Failed to save definition to SQL (non-fatal): %v", saveErr))
				dbTx.Rollback()
			} else {
				dbTx.Commit()
			}
		}
	}

	iLog.Info(fmt.Sprintf("Package definition created: %s (%s)", def.Name, def.ID))
	c.JSON(http.StatusCreated, gin.H{"data": def, "message": "Package definition created successfully"})
}

// ─── UpdateDefinition ────────────────────────────────────────────────────────

// UpdateDefinition godoc
// @Summary Update a package definition — increments version_number in MongoDB and SQL
// @Router /api/package-definitions/:id [put]
func (pdc *PackageDefinitionController) UpdateDefinition(c *gin.Context) {
	id := c.Param("id")
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageDefinitionController.UpdateDefinition"}

	if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Document database not available"})
		return
	}

	var update deploymodels.PackageDefinitionDoc
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	userName := definitionUser(c)

	// Load the current MongoDB doc to get the current version_number
	ctx, cancel := definitionCtx()
	defer cancel()

	coll := documents.DocDBCon.MongoDBDatabase.Collection(packageDefinitionsCollection)
	var current deploymodels.PackageDefinitionDoc
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&current); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Definition not found: %v", err)})
		return
	}

	// Increment version
	newVersion := current.VersionNumber + 1
	if newVersion < 2 {
		newVersion = 2
	}

	now := time.Now().Format(time.RFC3339)
	update.ID = id
	update.VersionNumber = newVersion
	update.Active = true
	update.CreatedBy = current.CreatedBy
	update.CreatedOn = current.CreatedOn
	update.ModifiedBy = userName
	update.ModifiedOn = now

	// Replace the MongoDB doc in-place (same _id, new version_number)
	if _, err := coll.ReplaceOne(ctx, bson.M{"_id": id}, update); err != nil {
		iLog.Error(fmt.Sprintf("Failed to update definition %s: %v", id, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update SQL metadata
	if dbconn.DB != nil {
		dbTx, txErr := dbconn.DB.Begin()
		if txErr == nil {
			sqlRepo := repository.NewPackageDefinitionRepository(userName, dbTx)
			sqlRec := &repository.PackageDefinitionRecord{
				ID:            id,
				Name:          update.Name,
				PackageType:   update.PackageType,
				Description:   update.Description,
				Environment:   update.Environment,
				LatestVersion: newVersion,
				ModifiedBy:    userName,
				ModifiedOn:    time.Now(),
			}
			if updErr := sqlRepo.UpdateDefinition(sqlRec); updErr != nil {
				iLog.Error(fmt.Sprintf("Failed to update definition in SQL (non-fatal): %v", updErr))
				dbTx.Rollback()
			} else {
				dbTx.Commit()
			}
		}
	}

	iLog.Info(fmt.Sprintf("Package definition updated: %s (v%d)", id, newVersion))
	c.JSON(http.StatusOK, gin.H{"data": update, "message": "Package definition updated successfully"})
}

// ─── DeleteDefinition ────────────────────────────────────────────────────────

// DeleteDefinition godoc
// @Summary Soft-delete a package definition
// @Router /api/package-definitions/:id [delete]
func (pdc *PackageDefinitionController) DeleteDefinition(c *gin.Context) {
	id := c.Param("id")
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageDefinitionController.DeleteDefinition"}

	if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Document database not available"})
		return
	}

	userName := definitionUser(c)

	ctx, cancel := definitionCtx()
	defer cancel()

	coll := documents.DocDBCon.MongoDBDatabase.Collection(packageDefinitionsCollection)
	_, err := coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
			"active":     false,
			"modifiedby": userName,
			"modifiedon": time.Now().Format(time.RFC3339),
		},
	})
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to delete definition %s: %v", id, err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Soft-delete SQL record
	if dbconn.DB != nil {
		dbTx, txErr := dbconn.DB.Begin()
		if txErr == nil {
			sqlRepo := repository.NewPackageDefinitionRepository(userName, dbTx)
			if delErr := sqlRepo.DeleteDefinition(id, userName); delErr != nil {
				iLog.Error(fmt.Sprintf("Failed to delete definition from SQL (non-fatal): %v", delErr))
				dbTx.Rollback()
			} else {
				dbTx.Commit()
			}
		}
	}

	iLog.Info(fmt.Sprintf("Package definition deleted: %s", id))
	c.JSON(http.StatusOK, gin.H{"message": "Package definition deleted successfully", "id": id})
}

// ─── PackFromDefinition ───────────────────────────────────────────────────────

// PackFromDefinition godoc
// @Summary Trigger packaging from a saved definition; tracks build number in SQL
// @Router /api/package-definitions/:id/pack [post]
func (pdc *PackageDefinitionController) PackFromDefinition(c *gin.Context) {
	id := c.Param("id")
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "PackageDefinitionController.PackFromDefinition"}
	startTime := time.Now()

	if documents.DocDBCon == nil || documents.DocDBCon.MongoDBDatabase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Document database not available"})
		return
	}

	userName := definitionUser(c)

	var body struct {
		Version string `json:"version"`
	}
	_ = c.ShouldBindJSON(&body)

	// Load definition from MongoDB
	ctx, cancel := definitionCtx()
	defer cancel()

	coll := documents.DocDBCon.MongoDBDatabase.Collection(packageDefinitionsCollection)
	var def deploymodels.PackageDefinitionDoc
	if err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&def); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Definition not found: %v", err)})
		return
	}

	// Determine version
	version := def.Version
	if body.Version != "" {
		version = body.Version
	} else if def.AutoVersion {
		version = def.Version + "-" + time.Now().Format("20060102150405")
	}

	// Resolve build number from SQL
	buildNumber := 1
	var pkgTx interface{ Commit() error; Rollback() error }
	var sqlRepo *repository.PackageDefinitionRepository

	if dbconn.DB != nil {
		dbTx, txErr := dbconn.DB.Begin()
		if txErr == nil {
			pkgTx = dbTx
			sqlRepo = repository.NewPackageDefinitionRepository(userName, dbTx)
			if bn, err := sqlRepo.GetNextBuildNumber(id); err == nil {
				buildNumber = bn
			}
		}
	}

	// Build error and result tracking
	var packageID, packageName, packageVersion string
	tablesProcessed, collectionsProcessed := 0, 0
	var packErr error

	if def.PackageType == "database" {
		dbTx2, err := dbconn.DB.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
			return
		}
		defer dbTx2.Rollback()

		packager := packagemgr.NewDatabasePackager(userName, dbTx2, dbconn.DatabaseType)
		p, err := packager.PackageTables(def.Name, version, userName, def.Filter)
		if err != nil {
			packErr = err
		} else {
			repo := repository.NewPackageRepository(userName, dbTx2)
			pkgRecord, err := repo.SavePackage(p, def.Environment, &def.Filter)
			if err != nil {
				packErr = err
			} else {
				packageID = pkgRecord.ID
				packageName = pkgRecord.Name
				packageVersion = pkgRecord.Version
				tablesProcessed = len(p.DatabaseData.Tables)

				now := time.Now()
				_ = repo.SaveAction(&repository.PackageActionRecord{
					PackageID:         p.ID,
					ActionType:        repository.ActionTypePack,
					ActionStatus:      repository.ActionStatusCompleted,
					SourceEnvironment: def.Environment,
					PerformedBy:       userName,
					PerformedAt:       now,
					StartedAt:         &startTime,
					CompletedAt:       &now,
					TablesProcessed:   tablesProcessed,
				})
				if commitErr := dbTx2.Commit(); commitErr != nil {
					packErr = fmt.Errorf("failed to commit package: %w", commitErr)
					packageID = ""
				}
			}
		}

	} else if def.PackageType == "document" {
		packager := packagemgr.NewDocumentPackager(documents.DocDBCon, userName)
		p, err := packager.PackageCollections(def.Name, version, userName, def.Filter)
		if err != nil {
			packErr = err
		} else {
			dbTx2, err := dbconn.DB.Begin()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to begin transaction: %v", err)})
				return
			}
			defer dbTx2.Rollback()

			repo := repository.NewPackageRepository(userName, dbTx2)
			pkgRecord, err := repo.SavePackage(p, def.Environment, &def.Filter)
			if err != nil {
				packErr = err
			} else {
				packageID = pkgRecord.ID
				packageName = pkgRecord.Name
				packageVersion = pkgRecord.Version
				collectionsProcessed = len(p.DocumentData.Collections)

				now := time.Now()
				_ = repo.SaveAction(&repository.PackageActionRecord{
					PackageID:            p.ID,
					ActionType:           repository.ActionTypePack,
					ActionStatus:         repository.ActionStatusCompleted,
					SourceEnvironment:    def.Environment,
					PerformedBy:          userName,
					PerformedAt:          now,
					StartedAt:            &startTime,
					CompletedAt:          &now,
					CollectionsProcessed: collectionsProcessed,
				})
				if commitErr := dbTx2.Commit(); commitErr != nil {
					packErr = fmt.Errorf("failed to commit package: %w", commitErr)
					packageID = ""
				}
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid package_type in definition"})
		return
	}

	if packErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": packErr.Error()})
		return
	}

	// Save build record in SQL
	now := time.Now()
	elapsed := int(now.Sub(startTime).Seconds())
	buildStatus := "completed"

	if sqlRepo != nil && pkgTx != nil {
		buildRec := &repository.PackageBuildRecord{
			ID:             uuid.New().String(),
			DefinitionID:   id,
			DefinitionName: def.Name,
			DefVersion:     def.VersionNumber,
			BuildNumber:    buildNumber,
			PackageID:      packageID,
			PackageName:    packageName,
			PackageVersion: packageVersion,
			Status:         buildStatus,
			StartedAt:      &startTime,
			CompletedAt:    &now,
			DurationSeconds: elapsed,
			TriggeredBy:    userName,
			Active:         true,
			CreatedBy:      userName,
			ModifiedBy:     userName,
			RowVersionStamp: 1,
		}
		_ = sqlRepo.SaveBuild(buildRec)
		pkgTx.Commit()
	}

	iLog.Info(fmt.Sprintf("Packaged from definition %s (%s v%s) — build #%d, tables=%d, collections=%d",
		def.ID, def.Name, version, buildNumber, tablesProcessed, collectionsProcessed))

	c.JSON(http.StatusCreated, gin.H{
		"message":      fmt.Sprintf("Package '%s' v%s created (build #%d)", def.Name, version, buildNumber),
		"package_id":   packageID,
		"package_name": packageName,
		"version":      version,
		"build_number": buildNumber,
	})
}

// ─── ListBuilds ───────────────────────────────────────────────────────────────

// ListBuilds godoc
// @Summary List build history for a package definition
// @Router /api/package-definitions/:id/builds [get]
func (pdc *PackageDefinitionController) ListBuilds(c *gin.Context) {
	id := c.Param("id")

	if dbconn.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SQL database not available"})
		return
	}

	dbTx, err := dbconn.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open DB transaction"})
		return
	}
	defer dbTx.Rollback()

	userName := definitionUser(c)
	repo := repository.NewPackageDefinitionRepository(userName, dbTx)
	builds, err := repo.ListBuilds(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": builds, "count": len(builds)})
}
