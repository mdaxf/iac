package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// registerMessagingTools registers IAC conversation and messaging tools
func (s *AgentToolRegistryService) registerMessagingTools() {
	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "list_conversations",
			Description: "Lists recent IAC chat conversations with their title, owner, and type.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"limit": {Type: "integer", Description: "Maximum number of conversations to return (default 10)"},
				},
			},
		},
	}, s.listConversationsHandler)

	s.register(ToolDefinition{
		Type: "function",
		Function: ToolFunctionDef{
			Name:        "send_message",
			Description: "Posts a message into an IAC conversation on behalf of the agent.",
			Parameters: ToolParameterSchema{
				Type: "object",
				Properties: map[string]ToolPropertySchema{
					"conversation_id": {Type: "string", Description: "The ID of the conversation to post the message to"},
					"message":         {Type: "string", Description: "The message text to send"},
				},
				Required: []string{"conversation_id", "message"},
			},
		},
	}, s.sendMessageHandler)
}

func (s *AgentToolRegistryService) listConversationsHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	limit := 10
	if v, ok := args["limit"].(float64); ok && v > 0 {
		limit = int(v)
	}
	rows, err := s.sqlDB.QueryContext(ctx,
		`SELECT id, title, userid, conversation_type, agent_definition_id, createdon
		 FROM conversations WHERE active = true ORDER BY createdon DESC LIMIT $1`, limit)
	if err != nil {
		return "", fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()
	return s.rowsToJSON(rows)
}

func (s *AgentToolRegistryService) sendMessageHandler(ctx context.Context, args map[string]interface{}) (string, error) {
	conversationID, _ := args["conversation_id"].(string)
	message, _ := args["message"].(string)
	if conversationID == "" || message == "" {
		return "", fmt.Errorf("conversation_id and message are required")
	}

	var exists bool
	err := s.sqlDB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM conversations WHERE id = $1 AND active = true)`,
		conversationID).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("failed to verify conversation: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("conversation %q not found", conversationID)
	}

	msgID := uuid.New().String()
	now := time.Now().UTC()
	_, err = s.sqlDB.ExecContext(ctx,
		`INSERT INTO chatmessages
		 (id, conversationid, messagetype, text, active, createdby, createdon, modifiedby, modifiedon, rowversionstamp)
		 VALUES ($1, $2, 'assistant', $3, true, 'agent', $4, 'agent', $4, 1)`,
		msgID, conversationID, message, now)
	if err != nil {
		return "", fmt.Errorf("failed to insert message: %w", err)
	}

	result := map[string]interface{}{
		"id":              msgID,
		"conversation_id": conversationID,
		"sent":            true,
		"timestamp":       now.Format(time.RFC3339),
	}
	b, _ := json.Marshal(result)
	return string(b), nil
}
