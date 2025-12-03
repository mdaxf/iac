package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// EmbeddingJobService handles async embedding generation jobs
type EmbeddingJobService struct {
	DB                     *gorm.DB
	VectorDB               *gorm.DB // Vector database connection
	SchemaEmbeddingService *SchemaEmbeddingService
	iLog                   logger.Log
}

// NewEmbeddingJobService creates a new embedding job service
func NewEmbeddingJobService(db *gorm.DB, openAIKey string) *EmbeddingJobService {
	// Get vector database connection
	vectorDB, err := GetVectorDB(db)
	if err != nil {
		// Fallback to main DB if vector DB connection fails
		vectorDB = db
	}

	return &EmbeddingJobService{
		DB:                     db,
		VectorDB:               vectorDB,
		SchemaEmbeddingService: NewSchemaEmbeddingService(db, openAIKey),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "EmbeddingJobService",
		},
	}
}

// JobStatus constants
const (
	JobStatusPending   = "pending"
	JobStatusRunning   = "running"
	JobStatusCompleted = "completed"
	JobStatusFailed    = "failed"
	JobStatusCancelled = "cancelled"
)

// JobType constants
const (
	JobTypeGenerateEmbeddings = "generate_embeddings"
	JobTypeRegenerateEmbeddings = "regenerate_embeddings"
)

// CreateEmbeddingJob creates a new embedding generation job
func (s *EmbeddingJobService) CreateEmbeddingJob(databaseAlias string, jobType string, tableNames []string, force bool) (*models.EmbeddingGenerationJob, error) {
	s.iLog.Info(fmt.Sprintf("Creating embedding job for database: %s, type: %s, tables: %v, force: %v",
		databaseAlias, jobType, tableNames, force))

	job := &models.EmbeddingGenerationJob{
		ConfigID:      s.SchemaEmbeddingService.ConfigID,
		JobType:       jobType,
		DatabaseAlias: &databaseAlias,
		Status:        JobStatusPending,
		CreatedBy:     "System",
	}

	if err := s.VectorDB.Create(job).Error; err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to create job in vector DB: %v", err))
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Created embedding job with ID: %d in vector database", job.ID))
	return job, nil
}

// GetJob retrieves a job by ID
func (s *EmbeddingJobService) GetJob(jobID int) (*models.EmbeddingGenerationJob, error) {
	var job models.EmbeddingGenerationJob
	if err := s.VectorDB.First(&job, jobID).Error; err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	return &job, nil
}

// GetJobByUUID retrieves a job by UUID
func (s *EmbeddingJobService) GetJobByUUID(uuid string) (*models.EmbeddingGenerationJob, error) {
	var job models.EmbeddingGenerationJob
	if err := s.VectorDB.Where("uuid = ?", uuid).First(&job).Error; err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}
	return &job, nil
}

// UpdateJobStatus updates the job status and related fields
func (s *EmbeddingJobService) UpdateJobStatus(jobID int, status string, errorMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	now := time.Now()
	if status == JobStatusRunning && errorMsg == nil {
		updates["started_at"] = now
	}
	if status == JobStatusCompleted || status == JobStatusFailed {
		updates["completed_at"] = now
	}
	if errorMsg != nil {
		updates["error_message"] = *errorMsg
	}

	if err := s.VectorDB.Model(&models.EmbeddingGenerationJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

// UpdateJobProgress updates the job progress
func (s *EmbeddingJobService) UpdateJobProgress(jobID int, totalItems, processedItems, failedItems int) error {
	updates := map[string]interface{}{
		"total_items":     totalItems,
		"processed_items": processedItems,
		"failed_items":    failedItems,
	}

	if err := s.VectorDB.Model(&models.EmbeddingGenerationJob{}).Where("id = ?", jobID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update job progress: %w", err)
	}

	return nil
}

// ProcessEmbeddingJobAsync starts an async job to generate embeddings
func (s *EmbeddingJobService) ProcessEmbeddingJobAsync(jobID int, databaseAlias string, tableNames []string, force bool) {
	// Run in a goroutine to not block the API response
	go func() {
		s.processEmbeddingJob(jobID, databaseAlias, tableNames, force)
	}()
}

// processEmbeddingJob processes the embedding generation job
func (s *EmbeddingJobService) processEmbeddingJob(jobID int, databaseAlias string, tableNames []string, force bool) {
	s.iLog.Info(fmt.Sprintf("Starting background job processing for job ID: %d, tables: %v, force: %v", jobID, tableNames, force))

	// Update job status to running
	if err := s.UpdateJobStatus(jobID, JobStatusRunning, nil); err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to update job status to running: %v", err))
		return
	}

	ctx := context.Background()

	// Count total items before starting
	var totalTables, totalColumns int64
	tableQuery := s.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ?", databaseAlias, "table")

	// Filter by specific tables if provided
	if len(tableNames) > 0 {
		tableQuery = tableQuery.Where("tablename IN ?", tableNames)
	}
	tableQuery.Count(&totalTables)

	columnQuery := s.DB.Table("databaseschemametadata").
		Where("databasealias = ? AND metadatatype = ?", databaseAlias, "column")

	// Filter by specific tables if provided
	if len(tableNames) > 0 {
		columnQuery = columnQuery.Where("tablename IN ?", tableNames)
	}
	columnQuery.Count(&totalColumns)

	totalItems := int(totalTables + totalColumns)
	s.iLog.Info(fmt.Sprintf("Job %d: Found %d total items to process (%d tables, %d columns)",
		jobID, totalItems, totalTables, totalColumns))

	// If no metadata exists, trigger schema discovery first
	if totalItems == 0 {
		s.iLog.Info(fmt.Sprintf("Job %d: No metadata found, triggering schema discovery", jobID))
		dbName := databaseAlias
		err := s.SchemaEmbeddingService.SchemaMetadataService.DiscoverDatabaseSchema(ctx, databaseAlias, dbName)
		if err != nil {
			errorMsg := fmt.Sprintf("Schema discovery failed: %v", err)
			s.iLog.Error(fmt.Sprintf("Job %d: %s", jobID, errorMsg))
			s.UpdateJobStatus(jobID, JobStatusFailed, &errorMsg)
			return
		}

		// Recount after discovery
		tableQuery = s.DB.Table("databaseschemametadata").
			Where("databasealias = ? AND metadatatype = ?", databaseAlias, "table")
		if len(tableNames) > 0 {
			tableQuery = tableQuery.Where("tablename IN ?", tableNames)
		}
		tableQuery.Count(&totalTables)

		columnQuery = s.DB.Table("databaseschemametadata").
			Where("databasealias = ? AND metadatatype = ?", databaseAlias, "column")
		if len(tableNames) > 0 {
			columnQuery = columnQuery.Where("tablename IN ?", tableNames)
		}
		columnQuery.Count(&totalColumns)
		totalItems = int(totalTables + totalColumns)
		s.iLog.Info(fmt.Sprintf("Job %d: After discovery, found %d items", jobID, totalItems))
	}

	// Initialize progress
	s.UpdateJobProgress(jobID, totalItems, 0, 0)

	// Process tables
	var tableMetadata []models.DatabaseSchemaMetadata
	tableMetadataQuery := s.DB.Where("databasealias = ? AND metadatatype = ?",
		databaseAlias, models.MetadataTypeTable)

	// Filter by specific tables if provided
	if len(tableNames) > 0 {
		tableMetadataQuery = tableMetadataQuery.Where("tablename IN ?", tableNames)
	}

	err := tableMetadataQuery.Find(&tableMetadata).Error
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch table metadata: %v", err)
		s.iLog.Error(fmt.Sprintf("Job %d: %s", jobID, errorMsg))
		s.UpdateJobStatus(jobID, JobStatusFailed, &errorMsg)
		return
	}

	processedItems := 0
	failedItems := 0

	// Process tables
	for i, meta := range tableMetadata {
		// Check if embedding already exists (in vector DB)
		var existingEmbedding models.DatabaseSchemaEmbedding
		err := s.VectorDB.Where("config_id = ? AND database_alias = ? AND table_name = ? AND column_name IS NULL",
			s.SchemaEmbeddingService.ConfigID, databaseAlias, meta.Table).First(&existingEmbedding).Error

		if err == gorm.ErrRecordNotFound || force {
			// Delete existing embedding if force mode
			if force && err != gorm.ErrRecordNotFound {
				s.VectorDB.Delete(&existingEmbedding)
				s.iLog.Info(fmt.Sprintf("Job %d: Deleted existing embedding for table %s (force mode)", jobID, meta.Table))
			}

			// Generate new embedding
			if err := s.SchemaEmbeddingService.GenerateAndStoreTableEmbedding(ctx, databaseAlias, &meta); err != nil {
				s.iLog.Error(fmt.Sprintf("Job %d: Failed to generate table embedding for %s: %v", jobID, meta.Table, err))
				failedItems++
			} else {
				processedItems++
			}

			// Update progress every 10 items
			if (i+1)%10 == 0 {
				s.UpdateJobProgress(jobID, totalItems, processedItems, failedItems)
				s.iLog.Info(fmt.Sprintf("Job %d: Progress - %d/%d items processed, %d failed",
					jobID, processedItems, totalItems, failedItems))
			}

			// Rate limiting
			time.Sleep(100 * time.Millisecond)
		} else if err != nil {
			s.iLog.Error(fmt.Sprintf("Job %d: Error checking existing embedding: %v", jobID, err))
			failedItems++
		} else {
			// Already exists, skip
			processedItems++
		}
	}

	// Process columns
	var columnMetadata []models.DatabaseSchemaMetadata
	columnMetadataQuery := s.DB.Where("databasealias = ? AND metadatatype = ?",
		databaseAlias, models.MetadataTypeColumn)

	// Filter by specific tables if provided
	if len(tableNames) > 0 {
		columnMetadataQuery = columnMetadataQuery.Where("tablename IN ?", tableNames)
	}

	err = columnMetadataQuery.Find(&columnMetadata).Error
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to fetch column metadata: %v", err)
		s.iLog.Error(fmt.Sprintf("Job %d: %s", jobID, errorMsg))
		s.UpdateJobStatus(jobID, JobStatusFailed, &errorMsg)
		return
	}

	for i, meta := range columnMetadata {
		// Check if embedding already exists (in vector DB)
		var existingEmbedding models.DatabaseSchemaEmbedding
		err := s.VectorDB.Where("config_id = ? AND database_alias = ? AND table_name = ? AND column_name = ?",
			s.SchemaEmbeddingService.ConfigID, databaseAlias, meta.Table, meta.Column).First(&existingEmbedding).Error

		if err == gorm.ErrRecordNotFound || force {
			// Delete existing embedding if force mode
			if force && err != gorm.ErrRecordNotFound {
				s.VectorDB.Delete(&existingEmbedding)
				s.iLog.Info(fmt.Sprintf("Job %d: Deleted existing embedding for column %s.%s (force mode)", jobID, meta.Table, meta.Column))
			}

			// Generate new embedding
			if err := s.SchemaEmbeddingService.GenerateAndStoreColumnEmbedding(ctx, databaseAlias, &meta); err != nil {
				s.iLog.Error(fmt.Sprintf("Job %d: Failed to generate column embedding for %s.%s: %v",
					jobID, meta.Table, meta.Column, err))
				failedItems++
			} else {
				processedItems++
			}

			// Update progress every 10 items
			if (i+1)%10 == 0 {
				s.UpdateJobProgress(jobID, totalItems, processedItems, failedItems)
				s.iLog.Info(fmt.Sprintf("Job %d: Progress - %d/%d items processed, %d failed",
					jobID, processedItems, totalItems, failedItems))
			}

			// Rate limiting
			time.Sleep(100 * time.Millisecond)
		} else if err != nil {
			s.iLog.Error(fmt.Sprintf("Job %d: Error checking existing embedding: %v", jobID, err))
			failedItems++
		} else {
			// Already exists, skip
			processedItems++
		}
	}

	// Final progress update
	s.UpdateJobProgress(jobID, totalItems, processedItems, failedItems)

	// Update job status
	if failedItems == totalItems && totalItems > 0 {
		errorMsg := fmt.Sprintf("All %d items failed to process", totalItems)
		s.UpdateJobStatus(jobID, JobStatusFailed, &errorMsg)
		s.iLog.Error(fmt.Sprintf("Job %d: %s", jobID, errorMsg))
	} else if failedItems > 0 {
		s.UpdateJobStatus(jobID, JobStatusCompleted, nil)
		s.iLog.Warn(fmt.Sprintf("Job %d: Completed with %d failures out of %d items", jobID, failedItems, totalItems))
	} else {
		s.UpdateJobStatus(jobID, JobStatusCompleted, nil)
		s.iLog.Info(fmt.Sprintf("Job %d: Completed successfully, processed %d items", jobID, processedItems))
	}
}
