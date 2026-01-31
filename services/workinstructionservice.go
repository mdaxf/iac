package services

import (
	"context"
	"fmt"
	"time"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CollectionWorkInstructions = "Work_Instructions"
)

// WorkInstructionService handles work instruction operations
type WorkInstructionService struct {
	docDB     documents.DocumentDB
	serverURL string // Server URL for resolving document URLs
}

// NewWorkInstructionService creates a new work instruction service
func NewWorkInstructionService(docDB documents.DocumentDB, serverURL string) *WorkInstructionService {
	return &WorkInstructionService{
		docDB:     docDB,
		serverURL: serverURL,
	}
}

// normalizeWorkInstruction ensures all arrays are initialized (not nil)
func (wis *WorkInstructionService) normalizeWorkInstruction(wi *models.WorkInstruction) {
	// Normalize root attachments
	if wi.Attachments == nil {
		wi.Attachments = []models.Attachment{}
	}
	for i := range wi.Attachments {
		if wi.Attachments[i].Annotations == nil {
			wi.Attachments[i].Annotations = []models.Annotation{}
		}
		if wi.Attachments[i].Views == nil {
			wi.Attachments[i].Views = []models.ThreeDView{}
		}
	}

	// Normalize steps and their attachments
	for i := range wi.Steps {
		if wi.Steps[i].Attachments == nil {
			wi.Steps[i].Attachments = []models.Attachment{}
		}
		for j := range wi.Steps[i].Attachments {
			if wi.Steps[i].Attachments[j].Annotations == nil {
				wi.Steps[i].Attachments[j].Annotations = []models.Annotation{}
			}
			if wi.Steps[i].Attachments[j].Views == nil {
				wi.Steps[i].Attachments[j].Views = []models.ThreeDView{}
			}
		}
		if wi.Steps[i].AttachmentOverrides == nil {
			wi.Steps[i].AttachmentOverrides = []models.AttachmentOverride{}
		}
		// Normalize arrays within attachment overrides
		for k := range wi.Steps[i].AttachmentOverrides {
			if wi.Steps[i].AttachmentOverrides[k].Annotations == nil {
				wi.Steps[i].AttachmentOverrides[k].Annotations = []models.Annotation{}
			}
			if wi.Steps[i].AttachmentOverrides[k].Views == nil {
				wi.Steps[i].AttachmentOverrides[k].Views = []models.ThreeDView{}
			}
		}
	}
}

// =====================================================
// Work Instruction Operations
// =====================================================

// ListWorkInstructions retrieves all work instructions
func (wis *WorkInstructionService) ListWorkInstructions(ctx context.Context) ([]models.WorkInstruction, error) {
	filter := bson.M{}
	results, err := wis.docDB.FindMany(ctx, CollectionWorkInstructions, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list work instructions: %w", err)
	}

	var workInstructions []models.WorkInstruction
	for _, result := range results {
		doc := bson.M(result)
		var wi models.WorkInstruction
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &wi); err != nil {
			continue
		}
		// Normalize arrays to prevent nil issues
		wis.normalizeWorkInstruction(&wi)
		// Resolve document URLs to full URLs
		wis.resolveDocumentURLs(&wi)
		workInstructions = append(workInstructions, wi)
	}

	return workInstructions, nil
}

// GetWorkInstruction retrieves a work instruction by ID
func (wis *WorkInstructionService) GetWorkInstruction(ctx context.Context, id string) (*models.WorkInstruction, error) {
	filter := bson.M{"_id": id}
	result, err := wis.docDB.FindOne(ctx, CollectionWorkInstructions, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get work instruction: %w", err)
	}

	doc := bson.M(result)
	var wi models.WorkInstruction
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal work instruction: %w", err)
	}
	if err := bson.Unmarshal(data, &wi); err != nil {
		return nil, fmt.Errorf("failed to unmarshal work instruction: %w", err)
	}
	// Normalize arrays to prevent nil issues
	wis.normalizeWorkInstruction(&wi)
	// Resolve document URLs to full URLs
	wis.resolveDocumentURLs(&wi)
	return &wi, nil
}

// CreateWorkInstruction creates a new work instruction
func (wis *WorkInstructionService) CreateWorkInstruction(ctx context.Context, req *models.WorkInstructionCreateRequest, user string) (*models.WorkInstruction, error) {
	now := time.Now()

	// Generate ID
	id := fmt.Sprintf("wi-%d", now.UnixNano())

	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = models.StatusDraft
	}

	// Initialize steps if nil
	steps := req.Steps
	if steps == nil {
		steps = []models.Step{}
	}

	// Initialize attachments if nil
	attachments := req.Attachments
	if attachments == nil {
		attachments = []models.Attachment{}
	}

	wi := &models.WorkInstruction{
		ID:           id,
		PartNumber:   req.PartNumber,
		Title:        req.Title,
		Version:      req.Version,
		Status:       status,
		LastModified: now,
		Author:       req.Author,
		Attachments:  attachments,
		Steps:        steps,
		Tags:         req.Tags,
		Category:     req.Category,
		Description:  req.Description,

		CreatedBy:       user,
		CreatedOn:       now,
		ModifiedBy:      user,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	// Normalize arrays to prevent nil issues
	wis.normalizeWorkInstruction(wi)

	// Convert to map for insertion
	data, err := bson.Marshal(wi)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal work instruction: %w", err)
	}

	var docMap map[string]interface{}
	if err := bson.Unmarshal(data, &docMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	_, err = wis.docDB.InsertOne(ctx, CollectionWorkInstructions, docMap)
	if err != nil {
		return nil, fmt.Errorf("failed to insert work instruction: %w", err)
	}

	return wi, nil
}

// UpdateWorkInstruction updates an existing work instruction
func (wis *WorkInstructionService) UpdateWorkInstruction(ctx context.Context, id string, req *models.WorkInstructionUpdateRequest, user string) (*models.WorkInstruction, error) {
	// Get existing work instruction
	existing, err := wis.GetWorkInstruction(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Update fields if provided
	if req.PartNumber != "" {
		existing.PartNumber = req.PartNumber
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Version != "" {
		existing.Version = req.Version
	}
	if req.Status != "" {
		existing.Status = req.Status
	}
	if req.Author != "" {
		existing.Author = req.Author
	}
	if req.Attachments != nil {
		existing.Attachments = req.Attachments
	}
	if req.Steps != nil {
		existing.Steps = req.Steps
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}
	if req.Category != "" {
		existing.Category = req.Category
	}
	if req.Description != "" {
		existing.Description = req.Description
	}

	// Normalize arrays to prevent nil issues
	wis.normalizeWorkInstruction(existing)

	existing.LastModified = now
	existing.ModifiedBy = user
	existing.ModifiedOn = now
	existing.RowVersionStamp++

	// Convert to map for update
	data, err := bson.Marshal(existing)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal work instruction: %w", err)
	}

	var docMap map[string]interface{}
	if err := bson.Unmarshal(data, &docMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	filter := bson.M{"_id": id}
	if err := wis.docDB.UpdateOne(ctx, CollectionWorkInstructions, filter, bson.M{"$set": docMap}); err != nil {
		return nil, fmt.Errorf("failed to update work instruction: %w", err)
	}

	return existing, nil
}

// DeleteWorkInstruction deletes a work instruction by ID
func (wis *WorkInstructionService) DeleteWorkInstruction(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	if err := wis.docDB.DeleteOne(ctx, CollectionWorkInstructions, filter); err != nil {
		return fmt.Errorf("failed to delete work instruction: %w", err)
	}
	return nil
}

// resolveDocumentURLs resolves relative document URLs to full URLs in work instruction attachments
func (wis *WorkInstructionService) resolveDocumentURLs(wi *models.WorkInstruction) {
	if wi == nil || wis.serverURL == "" {
		return
	}

	// Remove trailing slash from serverURL
	serverURL := wis.serverURL
	if len(serverURL) > 0 && serverURL[len(serverURL)-1] == '/' {
		serverURL = serverURL[:len(serverURL)-1]
	}

	// Resolve URLs in root-level attachments
	for i := range wi.Attachments {
		attachment := &wi.Attachments[i]
		if attachment.URL != "" {
			// If already a full URL, skip
			if len(attachment.URL) > 7 && (attachment.URL[:7] == "http://" || attachment.URL[:8] == "https://") {
				continue
			}
			// Ensure URL starts with /
			url := attachment.URL
			if len(url) > 0 && url[0] != '/' {
				url = "/" + url
			}
			// Prepend server URL
			attachment.URL = serverURL + url
		}
	}

	// Resolve URLs in each step's attachments
	for i := range wi.Steps {
		for j := range wi.Steps[i].Attachments {
			attachment := &wi.Steps[i].Attachments[j]
			if attachment.URL != "" {
				// If already a full URL, skip
				if len(attachment.URL) > 7 && (attachment.URL[:7] == "http://" || attachment.URL[:8] == "https://") {
					continue
				}
				// Ensure URL starts with /
				url := attachment.URL
				if len(url) > 0 && url[0] != '/' {
					url = "/" + url
				}
				// Prepend server URL
				attachment.URL = serverURL + url
			}
		}
	}
}

// DuplicateWorkInstruction creates a copy of an existing work instruction
func (wis *WorkInstructionService) DuplicateWorkInstruction(ctx context.Context, id string, user string) (*models.WorkInstruction, error) {
	// Get existing work instruction
	existing, err := wis.GetWorkInstruction(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newID := fmt.Sprintf("wi-%d", now.UnixNano())

	// Create duplicate
	duplicate := &models.WorkInstruction{
		ID:           newID,
		PartNumber:   existing.PartNumber,
		Title:        existing.Title + " (Copy)",
		Version:      "1.0", // Reset version for duplicate
		Status:       models.StatusDraft,
		LastModified: now,
		Author:       existing.Author,
		Attachments:  existing.Attachments, // Copy root attachments
		Steps:        existing.Steps,        // Deep copy the steps
		Tags:         existing.Tags,
		Category:     existing.Category,
		Description:  existing.Description,

		CreatedBy:       user,
		CreatedOn:       now,
		ModifiedBy:      user,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	// Convert to map for insertion
	data, err := bson.Marshal(duplicate)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal work instruction: %w", err)
	}

	var docMap map[string]interface{}
	if err := bson.Unmarshal(data, &docMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	_, err = wis.docDB.InsertOne(ctx, CollectionWorkInstructions, docMap)
	if err != nil {
		return nil, fmt.Errorf("failed to insert work instruction: %w", err)
	}

	return duplicate, nil
}

// ChangeStatus updates the status of a work instruction
func (wis *WorkInstructionService) ChangeStatus(ctx context.Context, id string, status models.InstructionStatus, user string) error {
	req := &models.WorkInstructionUpdateRequest{
		Status: status,
	}
	_, err := wis.UpdateWorkInstruction(ctx, id, req, user)
	return err
}

// GetByStatus retrieves work instructions by status
func (wis *WorkInstructionService) GetByStatus(ctx context.Context, status models.InstructionStatus) ([]models.WorkInstruction, error) {
	filter := bson.M{"status": status}
	results, err := wis.docDB.FindMany(ctx, CollectionWorkInstructions, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get work instructions by status: %w", err)
	}

	var workInstructions []models.WorkInstruction
	for _, result := range results {
		doc := bson.M(result)
		var wi models.WorkInstruction
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &wi); err != nil {
			continue
		}
		// Normalize arrays to prevent nil issues
		wis.normalizeWorkInstruction(&wi)
		// Resolve document URLs to full URLs
		wis.resolveDocumentURLs(&wi)
		workInstructions = append(workInstructions, wi)
	}

	return workInstructions, nil
}

// SearchWorkInstructions searches work instructions by title, part number, or tags
func (wis *WorkInstructionService) SearchWorkInstructions(ctx context.Context, query string) ([]models.WorkInstruction, error) {
	// Create regex filter for case-insensitive search
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"partNumber": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	results, err := wis.docDB.FindMany(ctx, CollectionWorkInstructions, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search work instructions: %w", err)
	}

	var workInstructions []models.WorkInstruction
	for _, result := range results {
		doc := bson.M(result)
		var wi models.WorkInstruction
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &wi); err != nil {
			continue
		}
		// Normalize arrays to prevent nil issues
		wis.normalizeWorkInstruction(&wi)
		// Resolve document URLs to full URLs
		wis.resolveDocumentURLs(&wi)
		workInstructions = append(workInstructions, wi)
	}

	return workInstructions, nil
}
