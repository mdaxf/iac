package models

import (
	"time"
)

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeImage   DocumentType = "IMAGE"
	DocumentTypePDF     DocumentType = "PDF"
	DocumentTypeVideo   DocumentType = "VIDEO"
	DocumentTypeOffice  DocumentType = "OFFICE"
	DocumentTypeModel3D DocumentType = "MODEL3D"
	DocumentTypeOther   DocumentType = "OTHER"
)

// Document represents a stored document/file in the system
type Document struct {
	ID           string       `json:"_id" bson:"_id"`
	Name         string       `json:"name" bson:"name"`                   // Original filename
	Type         DocumentType `json:"type" bson:"type"`                   // Document type
	FilePath     string       `json:"filePath" bson:"filePath"`           // Physical file path on server
	URL          string       `json:"url" bson:"url"`                     // Relative URL (no server name)
	Size         int64        `json:"size" bson:"size"`                   // File size in bytes
	MimeType     string       `json:"mimeType" bson:"mimeType"`           // MIME type
	Description  string       `json:"description,omitempty" bson:"description,omitempty"`
	Tags         []string     `json:"tags,omitempty" bson:"tags,omitempty"`

	// Audit Fields
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// DocumentCreateRequest represents the request to create a new document record
type DocumentCreateRequest struct {
	Name        string       `json:"name" binding:"required"`
	Type        DocumentType `json:"type" binding:"required"`
	FilePath    string       `json:"filePath" binding:"required"`
	URL         string       `json:"url" binding:"required"`
	Size        int64        `json:"size"`
	MimeType    string       `json:"mimeType"`
	Description string       `json:"description,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
}

// DocumentUpdateRequest represents the request to update an existing document
type DocumentUpdateRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// DocumentListResponse represents the response for listing documents
type DocumentListResponse struct {
	Data  []Document `json:"data"`
	Count int        `json:"count"`
}
