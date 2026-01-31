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

package jobqueue

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services"
)

// AIScheduleBatchHandler handles AI schedule batch processing jobs
type AIScheduleBatchHandler struct {
	db      *sql.DB
	docdb   *documents.DocDB
	service *services.AIScheduleBatchService
	iLog    logger.Log
}

// NewAIScheduleBatchHandler creates a new batch handler
func NewAIScheduleBatchHandler(db *sql.DB, docdb *documents.DocDB) *AIScheduleBatchHandler {
	return &AIScheduleBatchHandler{
		db:      db,
		docdb:   docdb,
		service: services.NewAIScheduleBatchService(db, docdb),
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AIScheduleBatchHandler",
		},
	}
}

// Handle processes an AI schedule batch job
func (h *AIScheduleBatchHandler) Handle(ctx context.Context, job *models.QueueJob) error {
	h.iLog.Info(fmt.Sprintf("AIScheduleBatchHandler START - JobID: %s", job.ID))

	// Parse payload
	var payload services.BatchProcessPayload
	if err := json.Unmarshal([]byte(job.Payload), &payload); err != nil {
		h.iLog.Error(fmt.Sprintf("Failed to parse batch payload: %v", err))
		return fmt.Errorf("failed to parse batch payload: %w", err)
	}

	h.iLog.Info(fmt.Sprintf("Processing batch %d for job %s - Tasks: %d", payload.BatchNumber, payload.BatchJobID, len(payload.Tasks)))

	// Process the batch
	if err := h.service.ProcessBatch(ctx, payload, job.CreatedBy); err != nil {
		h.iLog.Error(fmt.Sprintf("Batch processing failed: %v", err))
		return err
	}

	h.iLog.Info(fmt.Sprintf("AIScheduleBatchHandler COMPLETE - Batch %d processed successfully", payload.BatchNumber))

	return nil
}

// RegisterAIScheduleBatchHandler registers the handler with the job system
func RegisterAIScheduleBatchHandler(db *sql.DB, docdb *documents.DocDB) {
	handler := NewAIScheduleBatchHandler(db, docdb)

	// Register with global handler registry
	if GlobalJobWorker != nil {
		GlobalJobWorker.RegisterHandler("ai.schedule.batch.process", handler.Handle)
	}
}
