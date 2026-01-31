package workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	wftype "github.com/mdaxf/iac/workflow/types"
)

// ExecuteSubflow executes a subflow (child workflow) as part of a parent workflow task.
// It retrieves the subflow definition, creates a workflow entity for the subflow,
// executes it, and merges the results back into the parent workflow's process data.
func ExecuteSubflow(workflowtaskid int64, NodeData wftype.Node, idbTx *sql.Tx, DocDBCon *documents.DocDB, UserName string) error {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow subflow execution", User: UserName}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("ExecuteSubflow", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Recovered in ExecuteSubflow: %s", r))
		}
	}()

	iLog.Debug(fmt.Sprintf("Execute Subflow for workflowtaskid: %d", workflowtaskid))

	// Check if subflow ID is provided
	if NodeData.SubflowId == "" {
		err := fmt.Errorf("subflow ID is missing for subflow node")
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

	// Get current process data from parent workflow task
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

	// Get parent workflow entity ID
	var parentWorkflowEntityID int64 = 0
	if rows[0]["workflowentityid"] != nil {
		parentWorkflowEntityID = rows[0]["workflowentityid"].(int64)
	}

	// Parse parent process data
	ProcessData := map[string]interface{}{}
	if rows[0]["processdata"] != nil {
		err := json.Unmarshal([]byte(rows[0]["processdata"].(string)), &ProcessData)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error in parsing process data: %s", err))
			return err
		}
	}

	// Merge node's processdata with parent workflow task's processdata
	// Node's processdata takes precedence (allows node-level configuration)
	if NodeData.ProcessData != nil {
		for key, value := range NodeData.ProcessData {
			ProcessData[key] = value
		}
	}

	iLog.Debug(fmt.Sprintf("Parent process data (after merge): %v", ProcessData))

	// Retrieve the subflow workflow definition
	subflowWorkflow, subflowBSON, err := GetWorkFlowbyUUID(NodeData.SubflowId, UserName, *DocDBCon)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting subflow workflow: %s", err))
		return err
	}

	iLog.Info(fmt.Sprintf("Executing subflow: %s (UUID: %s)", subflowWorkflow.Name, subflowWorkflow.UUID))

	// Prepare subflow entity data
	entityData := ProcessData // Pass parent process data to subflow

	// Create workflow entity for the subflow
	subflowEntityDescription := fmt.Sprintf("Subflow execution from parent task %d", workflowtaskid)

	jsonEntityData, err := json.Marshal(entityData)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in marshaling entity data: %s", err))
		return err
	}

	jsonWorkflowData, err := json.Marshal(subflowBSON)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in marshaling workflow data: %s", err))
		return err
	}

	// Insert subflow workflow entity
	columns := []string{"typecode", "entity", "status", "description", "data", "workflowuuid", "workflow", "parentworkflowentityid", "createdby", "createdon", "modifiedby", "modifiedon"}
	values := []string{
		"subflow",
		NodeData.Name,
		"1", // Status: Active
		subflowEntityDescription,
		string(jsonEntityData),
		subflowWorkflow.UUID,
		string(jsonWorkflowData),
		fmt.Sprintf("%d", parentWorkflowEntityID),
		UserName,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
		UserName,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
	}

	subflowEntityID, err := dbop.TableInsert("workflow_entities", columns, values)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in creating subflow entity: %s", err))
		return err
	}

	iLog.Debug(fmt.Sprintf("Created subflow entity with ID: %d", subflowEntityID))

	// Execute the subflow by exploding its nodes
	wfexplode := NewExplosion("", "", "", UserName, "")

	// Find start node
	var startNode wftype.Node
	for _, node := range subflowWorkflow.Nodes {
		if node.Type == "start" {
			startNode = node
			break
		}
	}

	if startNode.ID == "" {
		err = fmt.Errorf("start node not found in subflow")
		iLog.Error(err.Error())
		return err
	}

	// Find first nodes after start
	firstNodes := []wftype.Node{}
	for _, link := range subflowWorkflow.Links {
		if link.Source == startNode.ID {
			for _, node := range subflowWorkflow.Nodes {
				if node.ID == link.Target && node.Type != "end" {
					firstNodes = append(firstNodes, node)
					break
				}
			}
		}
	}

	if len(firstNodes) == 0 {
		err = fmt.Errorf("no nodes found after start node in subflow")
		iLog.Error(err.Error())
		return err
	}

	// Explode first nodes with parent process data
	pretaskdata := make(map[string]interface{})
	for _, node := range firstNodes {
		iLog.Debug(fmt.Sprintf("Exploding subflow first node: %s", node.ID))
		wfexplode.explodeNode(node, subflowEntityID, DocDBCon, idbTx, pretaskdata)
	}

	// Wait for subflow to complete by checking workflow entity status
	maxWaitTime := 300 // 5 minutes max wait
	checkInterval := 2 // Check every 2 seconds
	elapsedTime := 0

	subflowCompleted := false
	for elapsedTime < maxWaitTime {
		// Check subflow entity status
		statusRows, err := dbop.Query_Json(fmt.Sprintf("select status from workflow_entities where id = %d", subflowEntityID))
		if err != nil {
			iLog.Error(fmt.Sprintf("Error checking subflow status: %s", err))
			break
		}

		if len(statusRows) > 0 && statusRows[0]["status"] != nil {
			status := statusRows[0]["status"].(int64)
			if status == 5 { // Completed
				subflowCompleted = true
				iLog.Info(fmt.Sprintf("Subflow entity %d completed", subflowEntityID))
				break
			}
		}

		time.Sleep(time.Duration(checkInterval) * time.Second)
		elapsedTime += checkInterval
	}

	if !subflowCompleted {
		iLog.Warn(fmt.Sprintf("Subflow did not complete within %d seconds, continuing anyway", maxWaitTime))
	}

	// Retrieve subflow results from workflow entity data
	resultRows, err := dbop.Query_Json(fmt.Sprintf("select data from workflow_entities where id = %d", subflowEntityID))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error retrieving subflow results: %s", err))
		return err
	}

	if len(resultRows) > 0 && resultRows[0]["data"] != nil {
		var subflowResults map[string]interface{}
		err := json.Unmarshal([]byte(resultRows[0]["data"].(string)), &subflowResults)
		if err != nil {
			iLog.Warn(fmt.Sprintf("Error parsing subflow results: %s", err))
		} else {
			// Merge subflow results back into parent process data
			// Prefix subflow results to avoid key conflicts
			for key, value := range subflowResults {
				ProcessData[fmt.Sprintf("subflow_%s", key)] = value
			}
			ProcessData["subflow_completed"] = true
			ProcessData["subflow_entity_id"] = subflowEntityID

			iLog.Debug(fmt.Sprintf("Merged subflow results into parent process data: %v", ProcessData))
		}
	}

	// Update parent workflow task with merged process data
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

	iLog.Info(fmt.Sprintf("Subflow executed successfully for workflowtaskid: %d", workflowtaskid))

	return nil
}
