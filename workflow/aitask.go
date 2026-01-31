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

// ExecuteAITask executes an AI task for a workflow task node.
// It retrieves the process data, builds the AI prompt from the node configuration,
// calls the AI service, extracts outputs, and updates the process data.
func ExecuteAITask(workflowtaskid int64, NodeData wftype.Node, idbTx *sql.Tx, DocDBCon *documents.DocDB, UserName string) error {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow AI task execution", User: UserName}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("ExecuteAITask", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Recovered in ExecuteAITask: %s", r))
		}
	}()

	iLog.Debug(fmt.Sprintf("Execute AI Task for workflowtaskid: %d", workflowtaskid))

	// Check if AI configuration exists
	if NodeData.AIConfig == nil {
		err := fmt.Errorf("AI configuration is missing for AI task node")
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

	// Build prompt from template using process data
	prompt := NodeData.AIConfig.UserPrompt
	if prompt == "" {
		err := fmt.Errorf("user prompt is required for AI task")
		iLog.Error(err.Error())
		return err
	}

	// Replace {{variableName}} placeholders with values from ProcessData
	for key, value := range ProcessData {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		prompt = strings.ReplaceAll(prompt, placeholder, valueStr)
	}

	iLog.Debug(fmt.Sprintf("Built AI prompt: %s", prompt))

	// Prepare AI task data in the format expected by the AI task function
	aiTaskData := map[string]interface{}{
		"useCaseName":   getStringValue(NodeData.AIConfig.UseCaseName, "workflow_ai_task"),
		"aiVendor":      getStringValue(NodeData.AIConfig.AIVendor, ""),
		"systemPrompt":  getStringValue(NodeData.AIConfig.SystemPrompt, ""),
		"userPrompt":    prompt,
		"temperature":   NodeData.AIConfig.Temperature,
		"maxTokens":     NodeData.AIConfig.MaxTokens,
		"timeout":       NodeData.AIConfig.Timeout,
		"retryAttempts": getIntValue(NodeData.AIConfig.RetryAttempts, 1),
	}

	iLog.Debug(fmt.Sprintf("AI task configuration: %v", aiTaskData))

	// Execute AI task through callback manager
	result, err := callback_mgr.CallBackFunc("AI_Task_Execute", aiTaskData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in executing AI task: %s", err))
		return err
	}

	iLog.Debug(fmt.Sprintf("AI task result: %v", result))

	// Extract AI response - CallBackFunc returns []interface{}
	var aiResponse string
	if len(result) > 0 {
		// Get the first element from the result slice
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
		// Empty result
		aiResponse = ""
		iLog.Warn("AI task returned empty result")
	}

	// Store AI response in process data
	ProcessData["ai_response"] = aiResponse
	ProcessData["ai_task_completed"] = true
	ProcessData["ai_task_timestamp"] = time.Now().UTC().Format("2006-01-02 15:04:05")

	iLog.Debug(fmt.Sprintf("Updated process data with AI response: %v", ProcessData))

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

	iLog.Info(fmt.Sprintf("AI task executed successfully for workflowtaskid: %d", workflowtaskid))

	return nil
}

// Helper functions

func getStringValue(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntValue(value int, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}
