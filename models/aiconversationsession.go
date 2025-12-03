package models

import (
	"database/sql"
)

// AIConversationSession represents a conversation session for AI agency
type AIConversationSession struct {
	ID          int          `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	SessionID   string       `json:"session_id" gorm:"column:session_id;type:varchar(36);not null;uniqueIndex"`
	UserID      string       `json:"user_id" gorm:"column:user_id;type:varchar(36);not null"`
	EditorType  string       `json:"editor_type" gorm:"column:editor_type;type:varchar(50)"` // bpm, page, view, workflow, whiteboard, report, general
	ContextData string       `json:"context_data" gorm:"column:context_data;type:text"`       // JSON serialized conversation context

	// Standard IAC audit fields (must be at end)
	Active          bool         `json:"active" gorm:"column:active;default:true"`
	ReferenceID     string       `json:"referenceid" gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby" gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon" gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby" gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon" gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
}

// TableName specifies the table name
func (AIConversationSession) TableName() string {
	return "aiconversationsessions"
}
