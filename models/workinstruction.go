package models

import (
	"time"
)

// =====================================================
// Work Instruction Models
// =====================================================

// InstructionStatus represents the workflow status of a work instruction
type InstructionStatus string

const (
	StatusDraft    InstructionStatus = "DRAFT"
	StatusReview   InstructionStatus = "REVIEW"
	StatusReleased InstructionStatus = "RELEASED"
	StatusObsolete InstructionStatus = "OBSOLETE"
)

// MediaType represents the type of media attachment
type MediaType string

const (
	MediaImage   MediaType = "IMAGE"
	MediaPDF     MediaType = "PDF"
	MediaVideo   MediaType = "VIDEO"
	MediaOffice  MediaType = "OFFICE"
	MediaModel3D MediaType = "MODEL3D"
	MediaOther   MediaType = "OTHER"
)

// AnnotationType represents the type of annotation
type AnnotationType string

const (
	AnnotationInfo     AnnotationType = "info"
	AnnotationWarning  AnnotationType = "warning"
	AnnotationCritical AnnotationType = "critical"
	AnnotationArrow    AnnotationType = "arrow"
	AnnotationSquare   AnnotationType = "square"
	AnnotationLine     AnnotationType = "line"
	AnnotationPin      AnnotationType = "pin"
)

// WorldPosition represents 3D coordinates for 3D annotations
type WorldPosition struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
	Z float64 `json:"z" bson:"z"`
}

// Annotation represents an annotation on an attachment
type Annotation struct {
	ID       string         `json:"id" bson:"id"`
	X        float64        `json:"x" bson:"x"`                 // Percentage 0-100 (for 2D)
	Y        float64        `json:"y" bson:"y"`                 // Percentage 0-100 (for 2D)
	Width    float64        `json:"width,omitempty" bson:"width,omitempty"`
	Height   float64        `json:"height,omitempty" bson:"height,omitempty"`
	Label    string         `json:"label" bson:"label"`
	Text     string         `json:"text" bson:"text"`
	Type     AnnotationType `json:"type" bson:"type"`
	Rotation float64        `json:"rotation,omitempty" bson:"rotation,omitempty"`
	Page     int            `json:"page,omitempty" bson:"page,omitempty"` // For PDFs
	Color    string         `json:"color,omitempty" bson:"color,omitempty"`

	// 3D Annotation fields
	WorldPosition      *WorldPosition `json:"worldPosition,omitempty" bson:"worldPosition,omitempty"` // 3D world coordinates
	Size               float64        `json:"size,omitempty" bson:"size,omitempty"`                   // Size for 3D annotations
	Opacity            float64        `json:"opacity,omitempty" bson:"opacity,omitempty"`             // Opacity 0-1
	ShowConnectionLine bool           `json:"showConnectionLine,omitempty" bson:"showConnectionLine,omitempty"` // Show line to annotation
}

// ThreeDViewConfig represents configuration for 3D model views
type ThreeDViewConfig struct {
	HighlightedNodeNames []string `json:"highlightedNodeNames" bson:"highlightedNodeNames"`
	AssemblyNodeNames    []string `json:"assemblyNodeNames" bson:"assemblyNodeNames"` // For exploded view animation
	HiddenNodeNames      []string `json:"hiddenNodeNames" bson:"hiddenNodeNames"`
	AnimationSpeed       float64  `json:"animationSpeed" bson:"animationSpeed"`
}

// ThreeDView represents a saved camera view for 3D models
type ThreeDView struct {
	ID               string   `json:"id" bson:"id"`
	Name             string   `json:"name" bson:"name"`
	CameraOrbit      string   `json:"cameraOrbit" bson:"cameraOrbit"`
	CameraTarget     string   `json:"cameraTarget" bson:"cameraTarget"`
	HiddenNodeNames  []string `json:"hiddenNodeNames" bson:"hiddenNodeNames"`
	AssemblyNodeNames []string `json:"assemblyNodeNames,omitempty" bson:"assemblyNodeNames,omitempty"`
	AnimationSpeed   float64  `json:"animationSpeed,omitempty" bson:"animationSpeed,omitempty"`
	AutoPlayAssembly bool     `json:"autoPlayAssembly,omitempty" bson:"autoPlayAssembly,omitempty"`
	AssemblyDelay    float64  `json:"assemblyDelay,omitempty" bson:"assemblyDelay,omitempty"`
}

// Attachment represents a media attachment in a step
type Attachment struct {
	ID           string            `json:"id" bson:"id"`
	Type         MediaType         `json:"type" bson:"type"`
	URL          string            `json:"url" bson:"url"`
	Name         string            `json:"name" bson:"name"`
	Annotations  []Annotation      `json:"annotations" bson:"annotations"`
	CameraOrbit  string            `json:"cameraOrbit,omitempty" bson:"cameraOrbit,omitempty"`
	CameraTarget string            `json:"cameraTarget,omitempty" bson:"cameraTarget,omitempty"`
	ThreeDConfig *ThreeDViewConfig `json:"threeDConfig,omitempty" bson:"threeDConfig,omitempty"`
	Views        []ThreeDView      `json:"views,omitempty" bson:"views,omitempty"`
}

// AttachmentOverride represents step-specific customizations for root-level attachments
type AttachmentOverride struct {
	ID            string            `json:"id" bson:"id"`
	AttachmentID  string            `json:"attachmentId" bson:"attachmentId"` // Reference to root attachment
	DefaultViewID string            `json:"defaultViewId,omitempty" bson:"defaultViewId,omitempty"`
	Annotations   []Annotation      `json:"annotations,omitempty" bson:"annotations,omitempty"`
	Hidden        bool              `json:"hidden,omitempty" bson:"hidden,omitempty"`
	Views         []ThreeDView      `json:"views,omitempty" bson:"views,omitempty"`
	ThreeDConfig  *ThreeDViewConfig `json:"threeDConfig,omitempty" bson:"threeDConfig,omitempty"`
	CameraOrbit   string            `json:"cameraOrbit,omitempty" bson:"cameraOrbit,omitempty"`
	CameraTarget  string            `json:"cameraTarget,omitempty" bson:"cameraTarget,omitempty"`
}

// Step represents a single step in a work instruction
type Step struct {
	ID                  string               `json:"id" bson:"id"`
	Order               int                  `json:"order" bson:"order"`
	Title               string               `json:"title" bson:"title"`
	Description         string               `json:"description" bson:"description"` // HTML or Markdown
	Attachments         []Attachment         `json:"attachments" bson:"attachments"`
	AttachmentOverrides []AttachmentOverride `json:"attachmentOverrides,omitempty" bson:"attachmentOverrides,omitempty"`
}

// WorkInstruction represents a complete work instruction document
type WorkInstruction struct {
	ID           string            `json:"_id" bson:"_id"`
	PartNumber   string            `json:"partNumber" bson:"partNumber"`
	Title        string            `json:"title" bson:"title"`
	Version      string            `json:"version" bson:"version"`
	Status       InstructionStatus `json:"status" bson:"status"`
	LastModified time.Time         `json:"lastModified" bson:"lastModified"`
	Author       string            `json:"author" bson:"author"`
	Attachments  []Attachment      `json:"attachments,omitempty" bson:"attachments,omitempty"` // Root-level shared attachments
	Steps        []Step            `json:"steps" bson:"steps"`

	// Additional metadata
	Tags        []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	Category    string    `json:"category,omitempty" bson:"category,omitempty"`
	Description string    `json:"description,omitempty" bson:"description,omitempty"`

	// Audit Fields
	CreatedBy       string    `json:"createdby" bson:"createdby"`
	CreatedOn       time.Time `json:"createdon" bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby" bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon" bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// WorkInstructionCreateRequest represents the request to create a new work instruction
type WorkInstructionCreateRequest struct {
	PartNumber  string            `json:"partNumber" binding:"required"`
	Title       string            `json:"title" binding:"required"`
	Version     string            `json:"version" binding:"required"`
	Status      InstructionStatus `json:"status"`
	Author      string            `json:"author" binding:"required"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Steps       []Step            `json:"steps"`
	Tags        []string          `json:"tags,omitempty"`
	Category    string            `json:"category,omitempty"`
	Description string            `json:"description,omitempty"`
}

// WorkInstructionUpdateRequest represents the request to update an existing work instruction
type WorkInstructionUpdateRequest struct {
	PartNumber  string            `json:"partNumber,omitempty"`
	Title       string            `json:"title,omitempty"`
	Version     string            `json:"version,omitempty"`
	Status      InstructionStatus `json:"status,omitempty"`
	Author      string            `json:"author,omitempty"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Steps       []Step            `json:"steps,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Category    string            `json:"category,omitempty"`
	Description string            `json:"description,omitempty"`
}
