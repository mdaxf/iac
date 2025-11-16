package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"database/sql"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/com"
	wftype "github.com/mdaxf/iac/workflow/types"

	"github.com/mdaxf/iac/notifications"

	"github.com/mdaxf/iac/framework/callback_mgr"
)

type WorkFlowTask struct {
	WorkFlowTaskID int64
	DocDBCon       *documents.DocDB
	DBTx           *sql.Tx
	Ctx            context.Context
	CtxCancel      context.CancelFunc
	UserName       string
	ClientID       string
	iLog           logger.Log
}

// NewWorkFlowTaskType creates a new instance of WorkFlowTask with the specified parameters.
// It initializes the log, context, and other properties of the WorkFlowTask.
// Parameters:
// - workflowtaskID: The ID of the workflow task.
// - UserName: The name of the user.
// Returns:
// - A pointer to the newly created WorkFlowTask instance.
func NewWorkFlowTaskType(workflowtaskID int64, UserName string) *WorkFlowTask {
	log := logger.Log{}
	log.ModuleName = logger.Framework
	log.ControllerName = "workflow Explosion"
	log.User = UserName

	Ctx, CtxCancel := context.WithTimeout(context.Background(), 10*time.Second)

	return &WorkFlowTask{
		WorkFlowTaskID: workflowtaskID,
		DocDBCon:       documents.DocDBCon,
		UserName:       UserName,
		iLog:           log,
		Ctx:            Ctx,
		CtxCancel:      CtxCancel,
	}
}

// UpdateTaskStatus updates the status of a workflow task.
// It takes a newstatus parameter of type int64, representing the new status value.
// The function returns an error if there was an issue updating the task status.
// Otherwise, it returns nil.
func (wft *WorkFlowTask) UpdateTaskStatus(newstatus int64) error {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		wft.iLog.PerformanceWithDuration("UpdateTaskStatus", elapsed)
	}()

	wft.iLog.Debug(fmt.Sprintf("UpdateTaskStatus by workflowtaskid: %d", wft.WorkFlowTaskID))

	DBTx, err := dbconn.DB.Begin()

	defer DBTx.Rollback()

	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
		return err
	}
	dbop := dbconn.NewDBOperation(wft.UserName, DBTx, logger.Framework)
	Columns := []string{"status"}
	Values := []string{fmt.Sprintf("%d", newstatus)}
	datatypes := []int{int(1)}

	if newstatus == 5 {
		Columns = []string{"status", "completedDate"}
		Values = []string{fmt.Sprintf("%d", newstatus), time.Now().UTC().Format("2006-01-02 15:04:05")}
		datatypes = []int{int(1), int(0)}
	} else if newstatus == 2 {
		Columns = []string{"status", "startedDate"}
		Values = []string{fmt.Sprintf("%d", newstatus), time.Now().UTC().Format("2006-01-02 15:04:05")}
		datatypes = []int{int(1), int(0)}
	}

	Where := fmt.Sprintf("id = %d", wft.WorkFlowTaskID)
	_, err = dbop.TableUpdate("workflow_tasks", Columns, Values, datatypes, Where)
	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in updating workflow tasks: %s", err))
		return err
	}

	DBTx.Commit()

	return nil
}

// StartTask starts the workflow task.
// It retrieves the workflow entity ID, workflow node ID, and notification UUID from the database based on the task ID.
// It then updates the task status to 2 (indicating that the task has started).
// If the task is being executed within an internal transaction, it commits the transaction.
// Finally, it asynchronously updates the notification with the "Task Started" message.
// Returns an error if any error occurs during the execution.

func (wft *WorkFlowTask) StartTask() error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		wft.iLog.PerformanceWithDuration("StartTask", elapsed)
	}()

	wft.iLog.Debug(fmt.Sprintf("StartTask by workflowtaskid: %d", wft.WorkFlowTaskID))

	DBTx := wft.DBTx
	err := error(nil)
	internaltransaction := false

	if DBTx == nil {
		DBTx, err = dbconn.DB.Begin()
		internaltransaction = true
		defer DBTx.Rollback()

		if err != nil {
			wft.iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
			return err
		}
	}
	dbop := dbconn.NewDBOperation(wft.UserName, DBTx, logger.Framework)

	//rows, err := dbop.Query_Json(fmt.Sprintf("select WorkflowEntityID, WorkflowNodeID, NotificationUUID from workflow_tasks where ID = %d", wft.WorkFlowTaskID))
	rows, err := dbop.Query_Json(fmt.Sprintf("select workflowentityid, workflownodeid, notificationuuid from workflow_tasks where id = %d", wft.WorkFlowTaskID))

	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return err
	}

	if len(rows) == 0 {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return err
	}

	err = wft.UpdateTaskStatus(2)

	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Update the task status to %d error!", 2))
	}

	if internaltransaction {
		DBTx.Commit()
	}

	if rows[0]["NotificationUUID"] != nil {
		go func() {

			NotificationUUID := rows[0]["NotificationUUID"].(string)
			notifications.UpdateNotificationbyUUID(NotificationUUID, wft.UserName, "Task Started")

		}()
	}

	return nil

}

func (wft *WorkFlowTask) UpdateProcessData(extprocessdata map[string]interface{}) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		wft.iLog.PerformanceWithDuration("UpdateProcessData", elapsed)
	}()

	wft.iLog.Debug(fmt.Sprintf("UpdateProcessData by workflowtaskid: %d with new process data: %v", wft.WorkFlowTaskID, extprocessdata))

	DBTx := wft.DBTx
	err := error(nil)
	internaltransaction := false

	if DBTx == nil {
		DBTx, err = dbconn.DB.Begin()
		internaltransaction = true
		defer DBTx.Rollback()

		if err != nil {
			wft.iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
			return err
		}
	}
	dbop := dbconn.NewDBOperation(wft.UserName, DBTx, logger.Framework)

	//rows, err := dbop.Query_Json(fmt.Sprintf("select WorkflowEntityID, WorkflowNodeID, ProcessData, NotificationUUID from workflow_tasks where ID = %d", wft.WorkFlowTaskID))
	rows, err := dbop.Query_Json(fmt.Sprintf("select workflowentityid, workflownodeid, processdata, notificationuuid from workflow_tasks where id = %d", wft.WorkFlowTaskID))

	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in getting cureent process data: %s", err))
		return err
	}

	if len(rows) == 0 {
		wft.iLog.Error(fmt.Sprintf("Error in getting cureent process data %s", err))
		return err
	}
	ProcessData := map[string]interface{}{}

	if rows[0]["processdata"] != nil {
		err := json.Unmarshal([]byte(rows[0]["processdata"].(string)), &ProcessData)
		if err != nil {
			wft.iLog.Error(fmt.Sprintf("Error in getting process data: %s", err))
			return err
		}
	}

	wft.iLog.Debug(fmt.Sprintf("the current process data: %v", ProcessData))
	for key, value := range extprocessdata {
		ProcessData[key] = value
	}

	wft.iLog.Debug(fmt.Sprintf("the new process data: %v", ProcessData))

	jsonData, err := json.Marshal(ProcessData)
	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode: %s", err))
		return err
	}

	Columns := []string{"processdata"}
	Values := []string{string(jsonData)}
	datatypes := []int{int(0)}
	Where := fmt.Sprintf("ID = %d", wft.WorkFlowTaskID)
	_, err = dbop.TableUpdate("workflow_tasks", Columns, Values, datatypes, Where)
	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in updating workflow tasks: %s", err))
		return err
	}

	if internaltransaction {
		DBTx.Commit()
	}
	return nil

}

func (wft *WorkFlowTask) ExecuteTaskTranCode(TranCode string) error {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		wft.iLog.PerformanceWithDuration("ExecuteTaskTranCode", elapsed)
	}()

	wft.iLog.Debug(fmt.Sprintf("ExecuteTaskTranCode by workflowtaskid: %d", wft.WorkFlowTaskID))

	if TranCode == "" {
		wft.iLog.Debug("Trancode canot be empty!")
		return fmt.Errorf("Trancode is empty!")
	}

	DBTx := wft.DBTx
	err := error(nil)
	internaltransaction := false

	if DBTx == nil {
		DBTx, err = dbconn.DB.Begin()
		internaltransaction = true
		defer DBTx.Rollback()

		if err != nil {
			wft.iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
			return err
		}
	}
	dbop := dbconn.NewDBOperation(wft.UserName, DBTx, logger.Framework)

	//rows, err := dbop.Query_Json(fmt.Sprintf("select WorkflowEntityID, WorkflowNodeID, ProcessData, NotificationUUID from workflow_tasks where ID = %d", wft.WorkFlowTaskID))
	rows, err := dbop.Query_Json(fmt.Sprintf("select workflowentityid, workflownodeid, processdata, notificationuuid from workflow_tasks where id = %d", wft.WorkFlowTaskID))

	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return err
	}

	if len(rows) == 0 {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return err
	}

	//	var WorkflowEntityID int64 = 0
	//	WorkflowNodeID := ""
	ProcessData := map[string]interface{}{}
	//	NotificationUUID := ""
	/*
		if rows[0]["WorkflowEntityID"] != nil {
			WorkflowEntityID = rows[0]["WorkflowEntityID"].(int64)
		}

		if rows[0]["WorkflowNodeID"] != nil {
			WorkflowNodeID = rows[0]["WorkflowNodeID"].(string)
		}
	*/
	if rows[0]["processdata"] != nil {
		err := json.Unmarshal([]byte(rows[0]["processdata"].(string)), &ProcessData)
		if err != nil {
			wft.iLog.Error(fmt.Sprintf("Error in getting process data: %s", err))
			return err
		}
	}

	_, err = ExecuteTaskTranCode(wft.WorkFlowTaskID, TranCode, ProcessData, DBTx, wft.DocDBCon, wft.UserName)

	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error during executing the trancode with error: %v", err))
		return err
	}

	if internaltransaction {
		DBTx.Commit()
	}
	return nil
}

// CompleteTask completes the workflow task.
// It updates the status and completed date of the task in the database.
// If the task is a gateway, it checks the routing table to determine the next nodes.
// If the task is a task node, it follows the links to find the next nodes.
// If the task is an end node, it marks the workflow as completed and triggers the validation and completion process.
// After completing the task, it updates the notification associated with the task.
// If the task is part of an internal transaction, it commits the transaction.
// If there are next nodes, it spawns goroutines to handle each next node concurrently.
// Returns an error if any operation fails.

func (wft *WorkFlowTask) CompleteTask() error {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		wft.iLog.PerformanceWithDuration("CompleteTask", elapsed)
	}()

	wft.iLog.Debug(fmt.Sprintf("CompleteTask by workflowtaskid: %d", wft.WorkFlowTaskID))

	DBTx := wft.DBTx
	err := error(nil)
	internaltransaction := false

	if DBTx == nil {
		DBTx, err = dbconn.DB.Begin()
		internaltransaction = true
		defer DBTx.Rollback()

		if err != nil {
			wft.iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
			return err
		}
	}
	dbop := dbconn.NewDBOperation(wft.UserName, DBTx, logger.Framework)

	Columns := []string{"status", "completedDate"}
	Values := []string{fmt.Sprintf("%d", 5), time.Now().UTC().Format("2006-01-02 15:04:05")}
	datatypes := []int{int(1), int(0)}
	Where := fmt.Sprintf("id = %d", wft.WorkFlowTaskID)
	_, err = dbop.TableUpdate("workflow_tasks", Columns, Values, datatypes, Where)
	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in updating workflow tasks: %s", err))
		return err
	}

	rows, err := dbop.Query_Json(fmt.Sprintf("select workflowentityid, workflownodeid, processdata, notificationuuid from workflow_tasks where id = %d", wft.WorkFlowTaskID))
	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return err
	}

	if len(rows) == 0 {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return err
	}

	var WorkflowEntityID int64 = 0
	WorkflowNodeID := ""
	ProcessData := map[string]interface{}{}
	NotificationUUID := ""

	if rows[0]["workflowentityid"] != nil {
		WorkflowEntityID = rows[0]["workflowentityid"].(int64)
	}

	if rows[0]["workflownodeid"] != nil {
		WorkflowNodeID = rows[0]["workflownodeid"].(string)
	}

	if rows[0]["processdata"] != nil {
		err := json.Unmarshal([]byte(rows[0]["processdata"].(string)), &ProcessData)
		if err != nil {
			wft.iLog.Error(fmt.Sprintf("Error in getting process data: %s", err))
			return err
		}
	}

	if rows[0]["notificationuuid"] != nil {
		NotificationUUID = rows[0]["notificationuuid"].(string)
	}

	if WorkflowEntityID == 0 || WorkflowNodeID == "" {
		err = fmt.Errorf("Error in getting workflow entity id: %s", err)
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return err
	}

	rows, err = dbop.Query_Json(fmt.Sprintf("select workflowuuid from workflow_entities where id = %d", WorkflowEntityID))
	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow uuid: %s", err))
		return err
	}

	if len(rows) == 0 {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow uuid: %s length of result is 0", err))
		return err
	}

	var WorkFlow wftype.WorkFlow

	WorkflowUUID := ""
	if rows[0]["workflowuuid"] != nil {
		WorkflowUUID = rows[0]["workflowuuid"].(string)
	}
	//WorkFlowSchema := rows[0]["WorkFlow"].([]byte)

	if WorkflowUUID == "" {
		err = fmt.Errorf("Error in getting workflow uuid: %s", err)
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow uuid: %s", err))
		return err
	}

	WorkFlow, _, err = GetWorkFlowbyUUID(WorkflowUUID, wft.UserName, *wft.DocDBCon)
	//err = json.Unmarshal(WorkFlowSchema, &WorkFlow)
	if err != nil {
		wft.iLog.Error(fmt.Sprintf("Error in getting workflow schema: %s", err))
		return err
	}

	Nodes := WorkFlow.Nodes
	Links := WorkFlow.Links

	// Get the current node routing table
	currentNode := wftype.Node{}
	for _, node := range Nodes {
		if node.ID == WorkflowNodeID {
			currentNode = node
			break
		}
	}

	nextNodes := []wftype.Node{}

	//RoutingTable := currentNode.RoutingTable
	/*
		{
			"sequence": 10,
			"data": "Status",
			"operator": "eq",
			"value": "Approved",
			"target": "nodeid"
		}


	*/
	if currentNode.Type == "gateway" {
		// check the routign table
		RoutingTables := currentNode.RoutingTables
		for _, routing := range RoutingTables {
			if CheckRoutingCondition(routing, ProcessData) {
				targetNodeID := routing.Target
				for _, node := range Nodes {
					if node.ID == targetNodeID {
						nextNodes = append(nextNodes, node)
						break
					}
				}
				if len(nextNodes) == 0 {
					err = fmt.Errorf("Error in getting next node with routing: %v", routing)
					wft.iLog.Error(fmt.Sprintf("Error in getting next node: %s", err))
					return err
				}
			}
		}

		if len(nextNodes) == 0 {
			err = fmt.Errorf("Error in getting next node with routing table: %v", RoutingTables)
			wft.iLog.Error(fmt.Sprintf("Error in getting next node: %s", err))
			return err
		}
	} else if currentNode.Type == "task" {
		for _, link := range Links {
			if link.Source == WorkflowNodeID {
				for _, node := range Nodes {
					if node.ID == link.Target {
						nextNodes = append(nextNodes, node)
					}
				}
			}
		}
		if len(nextNodes) == 0 {
			err = fmt.Errorf("Error in getting next node for the task: %v", currentNode)
			wft.iLog.Error(fmt.Sprintf("Error in getting next node: %s", err))
			return err
		}

	} else if currentNode.Type == "end" {
		wft.iLog.Debug(fmt.Sprintf("Workflow completed for workflowtaskid: %d", wft.WorkFlowTaskID))

		go func() {
			ValidateAndCompleteWorkFlow(WorkflowEntityID, DBTx, wft.DocDBCon, wft.UserName)
		}()
	}

	go func() {
		notifications.UpdateNotificationbyUUID(NotificationUUID, wft.UserName, "Task Completed")
	}()

	if len(nextNodes) > 0 {

		var wg sync.WaitGroup

		// Add 1 to the wait group
		wg.Add(1)

		go ExplodeNextNodes(&wg, WorkflowEntityID, nextNodes, wft.UserName, ProcessData)

		wg.Wait()
	}

	if internaltransaction {
		DBTx.Commit()
	}

	return nil
}

// ExplodeNextNodes explodes the next nodes in a workflow task.
// It creates new workflow tasks for each next node, associating them with the given WorkflowEntityID.
// The function uses a sync.WaitGroup to wait for all tasks to complete before returning.
// It also logs performance metrics and any errors that occur during the process.
//
// Parameters:
// - wg: A pointer to a sync.WaitGroup used to wait for all tasks to complete.
// - WorkflowEntityID: The ID of the workflow entity.
// - nextNodes: A slice of wftype.Node representing the next nodes in the workflow.
// - UserName: The name of the user performing the operation.
//
// Returns:
// - An error if any error occurs during the process, otherwise nil.
func ExplodeNextNodes(wg *sync.WaitGroup, WorkflowEntityID int64, nextNodes []wftype.Node, UserName string, ProcessData map[string]interface{}) error {

	defer wg.Done()

	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow tasks explode next node"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("ExplodeNextNode", elapsed)
	}()

	iLog.Debug(fmt.Sprintf("ExplodeNextNode by workflowentityid: %d", WorkflowEntityID))

	err := error(nil)

	idbTx, err := dbconn.DB.Begin()
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
		return err
	}

	defer idbTx.Rollback()

	DocDBCon := documents.DocDBCon
	//	defer DocDBCon.MongoDBClient.Disconnect(context.Background())

	wfexplode := NewExplosion("", "", "", UserName, "")
	for _, node := range nextNodes {
		// create new workflow task
		wfexplode.explodeNode(node, WorkflowEntityID, DocDBCon, idbTx, ProcessData)
	}

	idbTx.Commit()

	return nil
}

// ValidateAndCompleteWorkFlow validates and completes a workflow based on the given parameters.
// It takes the following parameters:
// - WorkFlowEntityID: The ID of the workflow entity.
// - idbTx: The SQL transaction object.
// - DocDBCon: The document database connection object.
// - UserName: The name of the user.
// It returns a boolean value indicating whether the workflow was successfully completed, and an error if any.
// If the workflow was successfully completed, it returns true, otherwise false.

func ValidateAndCompleteWorkFlow(WorkFlowEntityID int64, idbTx *sql.Tx, DocDBCon *documents.DocDB, UserName string) (bool, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: UserName, ControllerName: "ValidateAndCompleteWorkFlow"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("ValidateAndCompleteWorkFlow", elapsed)
	}()

	iLog.Debug(fmt.Sprintf("ValidateAndCompleteWorkFlow by workflowentityid: %d", WorkFlowEntityID))

	internaltransaction := false
	err := error(nil)

	if idbTx == nil {
		idbTx, err = dbconn.DB.Begin()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
			return false, err
		}
		internaltransaction = true
		defer idbTx.Rollback()
	}

	if DocDBCon == nil {
		DocDBCon = documents.DocDBCon
		//		defer DocDBCon.MongoDBClient.Disconnect(context.Background())
	}

	dbop := dbconn.NewDBOperation(UserName, idbTx, logger.Framework)

	tasks, err := dbop.Query_Json(fmt.Sprintf("select * from workflow_tasks where workflowentityid = %d AND status != %d", WorkFlowEntityID, 5))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting workflow tasks: %s", err))

		if internaltransaction {
			idbTx.Commit()
		}

		return false, err
	}

	if len(tasks) == 0 {
		iLog.Debug(fmt.Sprintf("Workflow completed for workflowentityid: %d", WorkFlowEntityID))

		Columns := []string{"status", "completeddate"}
		Values := []string{fmt.Sprintf("%d", 5), time.Now().UTC().Format("2006-01-02 15:04:05")}
		datatypes := []int{int(1), int(0)}
		Where := fmt.Sprintf("id = %d", WorkFlowEntityID)
		_, err = dbop.TableUpdate("workflow_entities", Columns, Values, datatypes, Where)

		if err != nil {
			iLog.Error(fmt.Sprintf("Error in updating workflow entities: %s", err))
			return false, err
		}
		if internaltransaction {
			idbTx.Commit()
		}
		return true, nil
	} else {
		iLog.Debug(fmt.Sprintf("Workflow not completed for workflowentityid: %d", WorkFlowEntityID))

		if internaltransaction {
			idbTx.Commit()
		}

		return false, nil
	}

}

// CheckRoutingCondition checks the routing condition based on the provided RoutingTable and ProcessData.
// It returns true if the routing condition is met, otherwise false.
// The routing condition is met if the ProcessData contains the data specified in the RoutingTable and the value of the data is equal to the value specified in the RoutingTable.
// If the RoutingTable is the default routing table, it returns true.
// If the ProcessData does not contain the data specified in the RoutingTable, it returns false.
func CheckRoutingCondition(Routing wftype.RoutingTable, ProcessData map[string]interface{}) bool {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow tasks check routing condition"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("CheckRoutingCondition", elapsed)
	}()

	iLog.Debug(fmt.Sprintf("CheckRoutingCondition by routing: %v with data %v", Routing, ProcessData))

	if Routing.Default {
		return true
	}

	if ProcessData[Routing.Data] == nil {
		return false
	} else if ProcessData[Routing.Data] == Routing.Value {
		return true
	} else {
		return false
	}

}

// ExecuteTask executes a workflow task.
// It takes the following parameters:
// - workflowtaskid: the ID of the workflow task
// - NodeData: the data associated with the workflow task node
// - idbTx: the SQL transaction object
// - DocDBCon: the document database connection object
// - UserName: the name of the user executing the task
// It returns a map[string]interface{} and an error.
// The map[string]interface{} contains the result of the task execution.
func ExecuteTask(workflowtaskid int64, NodeData wftype.Node, idbTx *sql.Tx, DocDBCon *documents.DocDB, UserName string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow tasks execution"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("ExecuteTaskTranCode", elapsed)
	}()

	iLog.Debug(fmt.Sprintf("ExecuteTask by workflowtaskid: %d", workflowtaskid))

	if NodeData.Type == "start" {
		wft := NewWorkFlowTaskType(workflowtaskid, UserName)
		wft.DBTx = idbTx
		wft.DocDBCon = DocDBCon
		wft.UpdateTaskStatus(2) // In Progress / started
		wft.CompleteTask()
		return nil, nil
	} else if NodeData.Type == "end" {
		wft := NewWorkFlowTaskType(workflowtaskid, UserName)
		wft.DBTx = idbTx
		wft.DocDBCon = DocDBCon
		wft.UpdateTaskStatus(2) // In Progress / started
		wft.CompleteTask()
		return nil, nil
	}

	if NodeData.Page == "" {

		internaltransaction := false
		err := error(nil)

		if idbTx == nil {
			idbTx, err = dbconn.DB.Begin()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
				return nil, err
			}
			internaltransaction = true
			defer idbTx.Rollback()
		}

		//	internalDoctransaction := false
		if DocDBCon == nil {
			DocDBCon = documents.DocDBCon
			//	internalDoctransaction = true
			defer DocDBCon.MongoDBClient.Disconnect(context.Background())
		}

		wft := NewWorkFlowTaskType(workflowtaskid, UserName)
		wft.DBTx = idbTx
		wft.DocDBCon = DocDBCon

		wft.UpdateTaskStatus(2) // In Progress / started

		if NodeData.TranCode != "" {

			//_, err = callback_mgr.CallBackFunc("TranCode_Execute", workflowtaskid, NodeData.TranCode, NodeData.ProcessData, idbTx, DocDBCon, UserName)

			_, err = ExecuteTaskTranCode(workflowtaskid, NodeData.TranCode, NodeData.ProcessData, idbTx, DocDBCon, UserName)

			if err != nil {
				wft.UpdateTaskStatus(4) // executed with Error
				return nil, err
			}
		}
		wft.CompleteTask()

		if internaltransaction {
			idbTx.Commit()
		}
	}

	return nil, nil
}

// ExecuteTaskTranCode executes a transaction code for a workflow task.
// It takes the workflow task ID, transaction code name, data map, database transaction, document database connection, and user name as input parameters.
// It returns the result of the transaction code execution and an error, if any.
// The result of the transaction code execution is a map[string]interface{}.
func ExecuteTaskTranCode(workflowtaskid int64, TranCodeName string, data map[string]interface{}, idbTx *sql.Tx, DocDBCon *documents.DocDB, UserName string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow tasks execute tran code"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("ExecuteTaskTranCode", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Recovered in ExecuteTaskTranCode: %s", r))
		}
	}()

	iLog.Debug(fmt.Sprintf("Execute the Trancode for workflowtaskid: %d", workflowtaskid))

	internaltransaction := false
	err := error(nil)

	if idbTx == nil {
		idbTx, err = dbconn.DB.Begin()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
			return nil, err
		}
		internaltransaction = true
		defer idbTx.Rollback()
	}

	//	internalDoctransaction := false
	if DocDBCon == nil {
		DocDBCon = documents.DocDBCon
		//	internalDoctransaction = true
		defer DocDBCon.MongoDBClient.Disconnect(context.Background())
	}

	//data["WorkFlowTaskID"] = workflowtaskid
	data["workflow_taskid"] = workflowtaskid
	sc := com.IACMessageBusClient

	//result, err := trancode.ExecutebyExternal(TranCodeName, data, idbTx, DocDBCon, sc)
	result, err := callback_mgr.CallBackFunc("TranCode_Execute", TranCodeName, data, idbTx, DocDBCon, sc)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in executing tran code: %s", err))
		return nil, err
	}

	dbop := dbconn.NewDBOperation(UserName, idbTx, logger.Framework)

	rows, err := dbop.Query_Json(fmt.Sprintf("select workflowentityid, workflownodeid, processdata from workflow_tasks where id = %d", workflowtaskid))
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return nil, err
	}

	if len(rows) == 0 {
		iLog.Error(fmt.Sprintf("Error in getting workflow entity id: %s", err))
		return nil, err
	}

	ProcessData := rows[0]["processdata"].(map[string]interface{})

	outputs := callback_mgr.ConvertSliceToMap(result)
	for key, value := range outputs {
		ProcessData[key] = value
	}

	Columns := []string{"processdata"}
	Values := []string{fmt.Sprintf("%s", ProcessData)}
	datatypes := []int{int(0)}
	Where := fmt.Sprintf("ID = %d", workflowtaskid)
	_, err = dbop.TableUpdate("workflow_tasks", Columns, Values, datatypes, Where)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in updating workflow tasks: %s", err))
		return nil, err
	}

	if internaltransaction {
		idbTx.Commit()
	}

	//if internalDoctransaction {
	//	DocDBCon.Commit()
	//}

	return outputs, nil
}

// GetWorkFlowTasks retrieves the workflow tasks associated with a given workflow entity ID and user name.
// It returns a slice of maps, where each map represents a workflow task with its corresponding attributes.
// If an error occurs during the retrieval process, it returns nil and the error.
// The attributes of a workflow task are:
// - ID: the ID of the workflow task
// - WorkflowEntityID: the ID of the workflow entity
// - WorkflowNodeID: the ID of the workflow node
// - ProcessData: the data associated with the workflow task
// - Status: the status of the workflow task
// - CreatedOn: the date and time when the workflow task was created
// - StartedDate: the date and time when the workflow task was started
// - CompletedDate: the date and time when the workflow task was completed
// - NotificationUUID: the UUID of the notification associated with the workflow task

func GetWorkFlowTasks(workflowentityid int64, UserName string) ([]map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow tasks"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("GetWorkFlowTasks", elapsed)
	}()

	iLog.Debug(fmt.Sprintf("GetWorkFlowTasks by workflowentityid: %d", workflowentityid))

	DBTx, err := dbconn.DB.Begin()

	defer DBTx.Rollback()

	if err != nil {
		iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
		return nil, err
	}

	dbop := dbconn.NewDBOperation(UserName, DBTx, logger.Framework)

	// Get workflow entity
	result, err := dbop.Query_Json(fmt.Sprintf("select * from workflow_tasks where workflowentityid = %d", workflowentityid))

	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting workflow tasks: %s", err))
		return nil, err
	}

	DBTx.Commit()

	return result, nil
}

func GetTasksbyUser(UserName string) ([]map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "workflow tasks"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("GetTasksbyUser", elapsed)
	}()

	iLog.Debug(fmt.Sprintf("GetTasksbyUser by username: %s", UserName))

	DBTx, err := dbconn.DB.Begin()

	defer DBTx.Rollback()

	if err != nil {
		iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
		return nil, err
	}

	dbop := dbconn.NewDBOperation(UserName, DBTx, logger.Framework)

	// Get workflow entity
	querytemp := `SELECT * FROM workflow_tasks wt 
			WHERE exists (Select 1 FROM workflow_task_assignments wts 
				LEFT JOIN user_roles ur on ur.roleid = wts.roleid 
				INNER JOIN users u on u.ID = wts.userid OR ur.userid = u.id 
				where wts.workflowtaskid = wt.id AND u.loginname = '%s')`

	result, err := dbop.Query_Json(fmt.Sprintf(querytemp, UserName))

	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting workflow tasks: %s", err))
		return nil, err
	}

	DBTx.Commit()

	return result, nil
}

func GetTaskPreTaskData(taskid int64, UserName string) (map[string]interface{}, error) {

	iLog := logger.Log{ModuleName: logger.Framework, ControllerName: "GetTaskPreTaskData"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("GetTaskPreTaskData", elapsed)
	}()

	iLog.Debug(fmt.Sprintf("GetTaskPreTaskData by username: %s", UserName))

	DBTx, err := dbconn.DB.Begin()

	defer DBTx.Rollback()

	if err != nil {
		iLog.Error(fmt.Sprintf("Error in creating DB connection: %s", err))
		return nil, err
	}

	dbop := dbconn.NewDBOperation(UserName, DBTx, logger.Framework)

	// Get workflow entity
	rows, err := dbop.Query_Json(fmt.Sprintf("SELECT pretaskdata FROM workflow_tasks WHERE id = %d", taskid))

	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting workflow task's PreTaskData: %s", err))
		return nil, err
	}

	if len(rows) == 0 {
		iLog.Error(fmt.Sprintf("Error in getting workflow uuid: %s length of result is 0", err))
		return nil, err
	}

	PreTaskData := make(map[string]interface{})

	if rows[0]["pretaskdata"] != nil {
		err := json.Unmarshal([]byte(rows[0]["pretaskdata"].(string)), &PreTaskData)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error in getting pretask data: %s", err))
			return nil, err
		}
	}

	DBTx.Commit()

	return PreTaskData, nil

}
