package services

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// AgentChatService manages conversations between users and agents
type AgentChatService struct {
	db    *gorm.DB
	sqlDB *sql.DB
	iLog  logger.Log
}

var (
	globalAgentChatService *AgentChatService
)

func GetGlobalAgentChatService() *AgentChatService {
	return globalAgentChatService
}

func SetGlobalAgentChatService(s *AgentChatService) {
	globalAgentChatService = s
}

func NewAgentChatService(db *gorm.DB, sqlDB *sql.DB) *AgentChatService {
	return &AgentChatService{
		db:    db,
		sqlDB: sqlDB,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentChatService",
		},
	}
}

// CreateAgentConversation creates a new agent conversation linked to an agent definition
func (s *AgentChatService) CreateAgentConversation(agentDefinitionID, userID, title string, createdBy string) (*models.Conversation, error) {
	convID := uuid.New().String()
	if title == "" {
		title = "Agent Chat"
	}

	conv := &models.Conversation{
		ID:               convID,
		Title:            title,
		UserID:           userID,
		DatabaseAlias:    "",
		AutoExecuteQuery: false,
		Active:           true,
		CreatedBy:        createdBy,
		ModifiedBy:       createdBy,
	}

	if err := s.db.Create(conv).Error; err != nil {
		return nil, fmt.Errorf("failed to create agent conversation: %w", err)
	}

	// Set agent_definition_id and conversation_type via raw SQL (column added via migration)
	_, err := s.sqlDB.Exec(
		`UPDATE conversations SET agent_definition_id = $1, conversation_type = 'agent' WHERE id = $2`,
		agentDefinitionID, convID,
	)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("Failed to set agent_definition_id on conversation %s: %v", convID, err))
	}

	return conv, nil
}

// GetConversation retrieves a conversation with messages
func (s *AgentChatService) GetConversation(id string) (*models.Conversation, error) {
	var conv models.Conversation
	err := s.db.Preload("Messages", func(db *gorm.DB) *gorm.DB {
		return db.Where("active = ?", true).Order("createdon ASC")
	}).First(&conv, "id = ? AND active = ?", id, true).Error
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListAgentConversations returns conversations for a specific agent and user
func (s *AgentChatService) ListAgentConversations(agentDefinitionID, userID string) ([]models.Conversation, error) {
	var convs []models.Conversation
	query := s.db.Where("active = ?", true)
	if agentDefinitionID != "" {
		query = query.Where("agent_definition_id = ?", agentDefinitionID)
	}
	if userID != "" {
		query = query.Where("userid = ?", userID)
	}
	err := query.Order("createdon DESC").Find(&convs).Error
	return convs, err
}

// SaveUserMessage saves a user message to the conversation
func (s *AgentChatService) SaveUserMessage(conversationID, text, createdBy string) (*models.ChatMessage, error) {
	return s.saveMessage(conversationID, models.MessageTypeUser, text, createdBy)
}

// SaveAgentMessage saves an agent (assistant) message to the conversation
func (s *AgentChatService) SaveAgentMessage(conversationID, text, createdBy string) (*models.ChatMessage, error) {
	return s.saveMessage(conversationID, models.MessageTypeAssistant, text, createdBy)
}

// SaveAgentMessageWithTools saves an agent message together with any rich tool results
// (ui_render, generate_html_report, etc.) produced during the run.
// toolResults is a slice of {tool, result} maps; stored as {"items":[...]} in the toolresults column.
func (s *AgentChatService) SaveAgentMessageWithTools(conversationID, text string, toolResults []map[string]interface{}, createdBy string) (*models.ChatMessage, error) {
	msg := &models.ChatMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		MessageType:    models.MessageTypeAssistant,
		Text:           text,
		Active:         true,
		CreatedBy:      createdBy,
		ModifiedBy:     createdBy,
	}
	if len(toolResults) > 0 {
		msg.ToolResults = models.JSONMap{"items": toolResults}
	}
	if err := s.db.Create(msg).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}
	return msg, nil
}

func (s *AgentChatService) saveMessage(conversationID string, msgType models.MessageType, text, createdBy string) (*models.ChatMessage, error) {
	msg := &models.ChatMessage{
		ID:             uuid.New().String(),
		ConversationID: conversationID,
		MessageType:    msgType,
		Text:           text,
		Active:         true,
		CreatedBy:      createdBy,
		ModifiedBy:     createdBy,
	}
	if err := s.db.Create(msg).Error; err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}
	return msg, nil
}

// GetMessageHistory retrieves conversation messages in chronological order
func (s *AgentChatService) GetMessageHistory(conversationID string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := s.db.Where("conversationid = ? AND active = ?", conversationID, true).
		Order("createdon ASC").
		Find(&messages).Error
	return messages, err
}

// DeleteConversation soft-deletes a conversation and its messages
func (s *AgentChatService) DeleteConversation(conversationID, deletedBy string) error {
	if err := s.db.Model(&models.ChatMessage{}).
		Where("conversationid = ?", conversationID).
		Updates(map[string]interface{}{"active": false, "modifiedby": deletedBy}).Error; err != nil {
		return err
	}
	return s.db.Model(&models.Conversation{}).
		Where("id = ?", conversationID).
		Updates(map[string]interface{}{"active": false, "modifiedby": deletedBy}).Error
}
