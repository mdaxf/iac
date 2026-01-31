package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/callback_mgr"
	"github.com/mdaxf/iac/logger"
	wftype "github.com/mdaxf/iac/workflow/types"
)

// ExecuteAIAgent executes an AI Agent/Skills for a workflow task node.
// It retrieves the process data, calls the AI agent with skills,
// and updates the process data with the results.
func ExecuteAIAgent(workflowtaskid int64, NodeData wftype.Node, idbTx *sql.Tx, DocDBCon *documents.DocDB, UserName string) error {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow AI agent execution", User: UserName}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("ExecuteAIAgent", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Recovered in ExecuteAIAgent: %s", r))
		}
	}()

	iLog.Debug(fmt.Sprintf("Execute AI Agent for workflowtaskid: %d", workflowtaskid))

	// Check if AI agent configuration exists
	if NodeData.AIAgentConfig == nil {
		err := fmt.Errorf("AI agent configuration is missing for AI agent node")
		iLog.Error(err.Error())
		return err
	}

	internaltransaction := false
	err := error(nil)

	if idbTx == nil {
		idbTx, err = dbconn.DB.Begin()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
			return err
		}
		internaltransaction = true
		defer idbTx.Rollback()
	}

	if DocDBCon == nil {
		DocDBCon = documents.DocDBCon
		defer DocDBCon.MongoDBClient.Disconnect(context.Background())
	}

	dbop := dbconn.NewDBOperation(UserName, idbTx, logger.Framework)

	// Get current process data from workflow task
	rows, err := dbop.Query_Json(fmt.Sprintf("select workflowentityid, workflownodeid, processdata from workflow_tasks where id = %d", workflowtaskid))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting workflow task data: %s", err))
		return err
	}

	if len(rows) == 0 {
		err = fmt.Errorf("workflow task not found")
		iLog.Error(fmt.Sprintf("Error in getting workflow task data: %s", err))
		return err
	}

	ProcessData := map[string]interface{}{}
	if rows[0]["processdata"] != nil {
		err := json.Unmarshal([]byte(rows[0]["processdata"].(string)), &ProcessData)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error in parsing process data: %s", err))
			return err
		}
	}

	// Merge node's processdata with workflow task's processdata
	// Node's processdata takes precedence (allows node-level configuration)
	if NodeData.ProcessData != nil {
		for key, value := range NodeData.ProcessData {
			ProcessData[key] = value
		}
	}

	iLog.Debug(fmt.Sprintf("Current process data (after merge): %v", ProcessData))

	// Build prompts from templates using process data
	systemPrompt := NodeData.AIAgentConfig.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a helpful AI assistant"
	}

	// Add agent name and skills to system prompt
	if NodeData.AIAgentConfig.AgentName != "" {
		systemPrompt = fmt.Sprintf("You are an AI agent named '%s'. %s", NodeData.AIAgentConfig.AgentName, systemPrompt)
	}

	if len(NodeData.AIAgentConfig.Skills) > 0 {
		skillsList := strings.Join(NodeData.AIAgentConfig.Skills, ", ")
		systemPrompt = fmt.Sprintf("%s\n\nYou have access to the following skills: %s", systemPrompt, skillsList)
	}

	// Build user prompt from template
	userPrompt := NodeData.AIAgentConfig.UserPrompt
	if userPrompt == "" {
		// Create a default prompt from process data
		dataJSON, _ := json.MarshalIndent(ProcessData, "", "  ")
		userPrompt = fmt.Sprintf("Process the following data:\n\n%s", string(dataJSON))
	} else {
		// Replace {{variableName}} placeholders with values from ProcessData
		for key, value := range ProcessData {
			placeholder := fmt.Sprintf("{{%s}}", key)
			valueStr := fmt.Sprintf("%v", value)
			userPrompt = strings.ReplaceAll(userPrompt, placeholder, valueStr)
		}
	}

	iLog.Debug(fmt.Sprintf("System prompt: %s", systemPrompt))
	iLog.Debug(fmt.Sprintf("User prompt: %s", userPrompt))

	// Prepare AI agent data for callback
	aiAgentData := map[string]interface{}{
		"agentName":    getStringValue(NodeData.AIAgentConfig.AgentName, "default_agent"),
		"skills":       NodeData.AIAgentConfig.Skills,
		"aiModel":      getStringValue(NodeData.AIAgentConfig.AIModel, ""),
		"aiVendor":     getStringValue(NodeData.AIAgentConfig.AIVendor, ""),
		"systemPrompt": systemPrompt,
		"userPrompt":   userPrompt,
		"temperature":  NodeData.AIAgentConfig.Temperature,
		"maxTokens":    NodeData.AIAgentConfig.MaxTokens,
		"timeout":      getIntValue(NodeData.AIAgentConfig.Timeout, 60),
		"inputData":    ProcessData,
	}

	iLog.Debug(fmt.Sprintf("AI agent configuration: %v", aiAgentData))

	// Execute AI agent through callback manager
	result, err := callback_mgr.CallBackFunc("AI_Agent_Execute", aiAgentData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in executing AI agent: %s", err))
		return err
	}

	iLog.Debug(fmt.Sprintf("AI agent result: %v", result))

	// Extract AI response
	var aiResponse string
	if len(result) > 0 {
		firstResult := result[0]

		if resultMap, ok := firstResult.(map[string]interface{}); ok {
			if response, exists := resultMap["response"]; exists {
				aiResponse = fmt.Sprintf("%v", response)
			} else if response, exists := resultMap["result"]; exists {
				aiResponse = fmt.Sprintf("%v", response)
			} else {
				// Try to convert entire result to JSON string
				jsonBytes, err := json.Marshal(resultMap)
				if err == nil {
					aiResponse = string(jsonBytes)
				} else {
					aiResponse = fmt.Sprintf("%v", firstResult)
				}
			}
		} else if resultStr, ok := firstResult.(string); ok {
			aiResponse = resultStr
		} else {
			aiResponse = fmt.Sprintf("%v", firstResult)
		}
	} else {
		aiResponse = ""
		iLog.Warn("AI agent returned empty result")
	}

	// Store AI response in process data
	ProcessData["ai_agent_response"] = aiResponse
	ProcessData["ai_agent_completed"] = true
	ProcessData["ai_agent_name"] = NodeData.AIAgentConfig.AgentName
	ProcessData["ai_agent_timestamp"] = time.Now().UTC().Format("2006-01-02 15:04:05")

	// Try to parse response as JSON and merge structured data
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(aiResponse), &responseData); err == nil {
		for key, value := range responseData {
			ProcessData[fmt.Sprintf("ai_agent_%s", key)] = value
		}
		iLog.Debug("Parsed AI agent response as JSON and merged into process data")
	}

	iLog.Debug(fmt.Sprintf("Updated process data with AI agent response: %v", ProcessData))

	// Update workflow task with new process data
	jsonData, err := json.Marshal(ProcessData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in marshaling process data: %s", err))
		return err
	}

	Columns := []string{"processdata"}
	Values := []string{string(jsonData)}
	datatypes := []int{int(0)}
	idColumn := dbop.QuoteIdentifier("id")
	Where := fmt.Sprintf("%s = %d", idColumn, workflowtaskid)
	_, err = dbop.TableUpdate("workflow_tasks", Columns, Values, datatypes, Where)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in updating workflow tasks: %s", err))
		return err
	}

	if internaltransaction {
		idbTx.Commit()
	}

	iLog.Info(fmt.Sprintf("AI agent executed successfully for workflowtaskid: %d", workflowtaskid))

	return nil
}
