package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/models"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	CollectionDocuments = "Documents"
)

// DocumentService handles document storage and management
type DocumentService struct {
	docDB       documents.DocumentDB
	storagePath string // Base path for document storage
	baseURL     string // Base URL path for document access (e.g., /storage/documents)
	serverURL   string // Full server URL (e.g., http://localhost:8080)
}

// NewDocumentService creates a new document service
func NewDocumentService(docDB documents.DocumentDB, storagePath string, baseURL string, serverURL string) *DocumentService {
	// Ensure storage path exists
	if storagePath != "" {
		os.MkdirAll(storagePath, 0755)
	}

	return &DocumentService{
		docDB:       docDB,
		storagePath: storagePath,
		baseURL:     baseURL,
		serverURL:   serverURL,
	}
}

// UploadDocument handles file upload and creates document record
func (ds *DocumentService) UploadDocument(ctx context.Context, file multipart.File, header *multipart.FileHeader, user string, description string, tags []string) (*models.Document, error) {
	// Generate unique ID and filename
	docID := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	uniqueFilename := fmt.Sprintf("%s%s", docID, ext)

	// Determine document type based on MIME type and file extension
	mimeType := header.Header.Get("Content-Type")
	fmt.Printf("DEBUG: Uploading file: %s, MIME type: %s\n", header.Filename, mimeType)
	docType := ds.getDocumentType(mimeType, header.Filename)
	fmt.Printf("DEBUG: Detected document type: %s\n", docType)

	// Create subdirectory based on type
	typeDir := strings.ToLower(string(docType))
	targetDir := filepath.Join(ds.storagePath, typeDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(targetDir, uniqueFilename)

	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Generate relative URL (for database storage - portable across environments)
	relativeURL := fmt.Sprintf("%s/%s/%s", ds.baseURL, typeDir, uniqueFilename)

	// Create document record
	now := time.Now()
	doc := &models.Document{
		ID:              docID,
		Name:            header.Filename,
		Type:            docType,
		FilePath:        filePath,
		URL:             relativeURL,
		Size:            size,
		MimeType:        header.Header.Get("Content-Type"),
		Description:     description,
		Tags:            tags,
		CreatedBy:       user,
		CreatedOn:       now,
		ModifiedBy:      user,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	// Save to database
	data, err := bson.Marshal(doc)
	if err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	var docMap map[string]interface{}
	if err := bson.Unmarshal(data, &docMap); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	if _, err := ds.docDB.InsertOne(ctx, CollectionDocuments, docMap); err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to insert document record: %w", err)
	}

	// Resolve URL to full URL before returning to frontend
	ds.resolveDocumentURLs(doc)

	return doc, nil
}

// ListDocuments retrieves all documents
func (ds *DocumentService) ListDocuments(ctx context.Context) ([]models.Document, error) {
	filter := bson.M{}
	results, err := ds.docDB.FindMany(ctx, CollectionDocuments, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	var docs []models.Document
	for _, result := range results {
		doc := bson.M(result)
		var d models.Document
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &d); err != nil {
			continue
		}
		// Resolve URL to full URL
		ds.resolveDocumentURLs(&d)
		docs = append(docs, d)
	}

	return docs, nil
}

// GetDocument retrieves a document by ID
func (ds *DocumentService) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	filter := bson.M{"_id": id}
	result, err := ds.docDB.FindOne(ctx, CollectionDocuments, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	doc := bson.M(result)
	var d models.Document
	data, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}
	if err := bson.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	// Resolve URL to full URL
	ds.resolveDocumentURLs(&d)

	return &d, nil
}

// UpdateDocument updates document metadata
func (ds *DocumentService) UpdateDocument(ctx context.Context, id string, req *models.DocumentUpdateRequest, user string) (*models.Document, error) {
	// Get existing document
	existing, err := ds.GetDocument(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Update fields if provided
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Tags != nil {
		existing.Tags = req.Tags
	}

	existing.ModifiedBy = user
	existing.ModifiedOn = now
	existing.RowVersionStamp++

	// Convert to map for update
	data, err := bson.Marshal(existing)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}

	var docMap map[string]interface{}
	if err := bson.Unmarshal(data, &docMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to map: %w", err)
	}

	filter := bson.M{"_id": id}
	if err := ds.docDB.UpdateOne(ctx, CollectionDocuments, filter, bson.M{"$set": docMap}); err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return existing, nil
}

// DeleteDocument deletes a document and its file
func (ds *DocumentService) DeleteDocument(ctx context.Context, id string) error {
	// Get document to find file path
	doc, err := ds.GetDocument(ctx, id)
	if err != nil {
		return err
	}

	// Delete from database first
	filter := bson.M{"_id": id}
	if err := ds.docDB.DeleteOne(ctx, CollectionDocuments, filter); err != nil {
		return fmt.Errorf("failed to delete document record: %w", err)
	}

	// Delete physical file (ignore errors - file might not exist)
	if doc.FilePath != "" {
		os.Remove(doc.FilePath)
	}

	return nil
}

// SearchDocuments searches documents by name or tags
func (ds *DocumentService) SearchDocuments(ctx context.Context, query string) ([]models.Document, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"name": bson.M{"$regex": query, "$options": "i"}},
			{"tags": bson.M{"$regex": query, "$options": "i"}},
			{"description": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	results, err := ds.docDB.FindMany(ctx, CollectionDocuments, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents: %w", err)
	}

	var docs []models.Document
	for _, result := range results {
		doc := bson.M(result)
		var d models.Document
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &d); err != nil {
			continue
		}
		// Resolve URL to full URL
		ds.resolveDocumentURLs(&d)
		docs = append(docs, d)
	}

	return docs, nil
}

// GetDocumentsByType retrieves documents by type
func (ds *DocumentService) GetDocumentsByType(ctx context.Context, docType models.DocumentType) ([]models.Document, error) {
	filter := bson.M{"type": docType}
	results, err := ds.docDB.FindMany(ctx, CollectionDocuments, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents by type: %w", err)
	}

	var docs []models.Document
	for _, result := range results {
		doc := bson.M(result)
		var d models.Document
		data, err := bson.Marshal(doc)
		if err != nil {
			continue
		}
		if err := bson.Unmarshal(data, &d); err != nil {
			continue
		}
		// Resolve URL to full URL
		ds.resolveDocumentURLs(&d)
		docs = append(docs, d)
	}

	return docs, nil
}

// ResolveDocumentURL converts a relative document URL to a full URL
// This is used by frontend to get actual accessible URLs
func (ds *DocumentService) ResolveDocumentURL(relativeURL string) string {
	if relativeURL == "" {
		return ""
	}

	// If already a full URL, return as-is
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}

	// Remove trailing slash from serverURL if present
	serverURL := ds.serverURL
	if strings.HasSuffix(serverURL, "/") {
		serverURL = serverURL[:len(serverURL)-1]
	}

	// Ensure relativeURL starts with /
	if !strings.HasPrefix(relativeURL, "/") {
		relativeURL = "/" + relativeURL
	}

	return serverURL + relativeURL
}

// resolveDocumentURLs resolves the URL field in a document to full URL
func (ds *DocumentService) resolveDocumentURLs(doc *models.Document) {
	if doc != nil {
		doc.URL = ds.ResolveDocumentURL(doc.URL)
	}
}

// getDocumentType determines document type from MIME type and file extension
func (ds *DocumentService) getDocumentType(mimeType string, filename string) models.DocumentType {
	// Normalize filename to lowercase for extension checking
	lowerFilename := strings.ToLower(filename)

	fmt.Printf("DEBUG getDocumentType: MIME='%s', filename='%s'\n", mimeType, filename)

	// Check MIME type first
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		fmt.Println("DEBUG: Matched as IMAGE (MIME)")
		return models.DocumentTypeImage
	case mimeType == "application/pdf":
		fmt.Println("DEBUG: Matched as PDF (MIME)")
		return models.DocumentTypePDF
	case strings.HasPrefix(mimeType, "video/"):
		fmt.Println("DEBUG: Matched as VIDEO (MIME)")
		return models.DocumentTypeVideo
	case strings.Contains(mimeType, "word") || strings.Contains(mimeType, "excel") || strings.Contains(mimeType, "powerpoint") ||
		strings.Contains(mimeType, "officedocument"):
		fmt.Println("DEBUG: Matched as OFFICE (MIME)")
		return models.DocumentTypeOffice
	case strings.Contains(mimeType, "model") || mimeType == "model/gltf-binary" || mimeType == "model/gltf+json":
		fmt.Println("DEBUG: Matched as MODEL3D (MIME)")
		return models.DocumentTypeModel3D
	}

	fmt.Println("DEBUG: No MIME match, checking file extension...")

	// Fallback to file extension for ambiguous MIME types (like application/octet-stream)
	switch {
	case strings.HasSuffix(lowerFilename, ".glb") || strings.HasSuffix(lowerFilename, ".gltf"):
		fmt.Println("DEBUG: Matched as MODEL3D (extension)")
		return models.DocumentTypeModel3D
	case strings.HasSuffix(lowerFilename, ".pdf"):
		return models.DocumentTypePDF
	case strings.HasSuffix(lowerFilename, ".jpg") || strings.HasSuffix(lowerFilename, ".jpeg") ||
		strings.HasSuffix(lowerFilename, ".png") || strings.HasSuffix(lowerFilename, ".gif") ||
		strings.HasSuffix(lowerFilename, ".webp") || strings.HasSuffix(lowerFilename, ".svg"):
		return models.DocumentTypeImage
	case strings.HasSuffix(lowerFilename, ".mp4") || strings.HasSuffix(lowerFilename, ".webm") ||
		strings.HasSuffix(lowerFilename, ".mov") || strings.HasSuffix(lowerFilename, ".avi"):
		return models.DocumentTypeVideo
	case strings.HasSuffix(lowerFilename, ".docx") || strings.HasSuffix(lowerFilename, ".doc") ||
		strings.HasSuffix(lowerFilename, ".xlsx") || strings.HasSuffix(lowerFilename, ".xls") ||
		strings.HasSuffix(lowerFilename, ".pptx") || strings.HasSuffix(lowerFilename, ".ppt"):
		return models.DocumentTypeOffice
	default:
		return models.DocumentTypeOther
	}
}
