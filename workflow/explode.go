package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"database/sql"

	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	wftype "github.com/mdaxf/iac/workflow/types"

	"github.com/google/uuid"
	//	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/notifications"
)

type ExplodionEngine struct {
	WorkflowName string
	EntityName   string
	Type         string
	Log          logger.Log
	workflow     wftype.WorkFlow
	DocDBCon     *documents.DocDB
	DBTx         *sql.Tx
	Ctx          context.Context
	CtxCancel    context.CancelFunc
	UserName     string
	ClientID     string
}

func NewExplosion(WorkFlowName string, EntityName string, Type string, UserName string, ClientID string) *ExplodionEngine {
	log := logger.Log{}
	log.ModuleName = logger.Framework
	log.ControllerName = "workflow Explosion"
	log.User = UserName
	log.ClientID = ClientID

	DBConn := documents.DocDBCon
	fmt.Print(log)
	return &ExplodionEngine{
		WorkflowName: WorkFlowName,
		EntityName:   EntityName,
		Type:         Type,
		Log:          log,
		UserName:     UserName,
		ClientID:     ClientID,
		DocDBCon:     DBConn,
	}

}

func (e *ExplodionEngine) Explode(Description string, EntityData map[string]interface{}) (int64, error) {
	/*if e.Log.ModuleName == "" {
		e.Log = logger.Log{ModuleName: logger.Framework, User: e.UserName, ControllerName: "workflow"}
	} */
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		e.Log.PerformanceWithDuration("engine.funcs.NewFuncs", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", r))
			return
		}
	}()

	e.Log.Debug(fmt.Sprintf("Start to explode workflow data %s's %s", e.WorkflowName, "Retrieven"))

	workflowM, err := e.getWorkFlowbyName()
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
		return 0, err
	}

	if workflowM == nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", "Workflow not found"))
		return 0, err
	}

	jsonString, err := json.Marshal(workflowM)
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
		return 0, err
	}

	var workflow wftype.WorkFlow
	err = json.Unmarshal(jsonString, &workflow)
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
		return 0, err
	}
	e.workflow = workflow

	e.Log.Debug(fmt.Sprintf("Workflow %s data %v ", e.WorkflowName, workflow))

	Nodes := workflow.Nodes
	Links := workflow.Links

	startNode := wftype.Node{}

	for _, node := range Nodes {
		if node.Type == "start" {
			e.Log.Debug(fmt.Sprintf("Workflow %s start node %s is %s ", e.WorkflowName, node.Name, e.Type))
			startNode = node
			break
		}
	}

	if startNode.ID == "" {
		err = fmt.Errorf("start node not found")
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", "start node not found"))
		return 0, err
	}

	e.Log.Debug(fmt.Sprintf("Workflow %s start node %s is %s ", e.WorkflowName, startNode.ID, e.Type))

	firstNodes := []wftype.Node{}

	for _, link := range Links {
		if link.Source == startNode.ID {
			e.Log.Debug(fmt.Sprintf("Workflow %s start node %s link %s ", e.WorkflowName, startNode.ID, link.Target))
			targetnode := e.getNodeByID(link.Target, Nodes)
			if targetnode.Type != "end" {
				firstNodes = append(firstNodes, targetnode)
			}
		}
	}

	if len(firstNodes) == 0 {
		err = fmt.Errorf("the first node not found")
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", "First node not found"))
		return 0, err
	}

	if e.DBTx == nil {
		e.DBTx, err = dbconn.DB.Begin()
		if err != nil {
			e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
			return 0, err
		}
		defer e.DBTx.Rollback()
	}

	jsonEntityData, err := json.Marshal(EntityData)
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode - convert the Entityata: %s", err))
		return 0, err
	}

	e.Log.Debug(fmt.Sprintf("Workflow %s first node %v ", e.WorkflowName, firstNodes))

	dbop := dbconn.NewDBOperation(e.UserName, e.DBTx, "Workflow.Explosion")

	//columns := []string{"Type", "Entity", "Status", "Description", "Data", "WorkflowUUID", "Workflow", "createdby", "createdon", "updatedby", "updatedon"}
	columns := []string{"typecode", "entity", "status", "description", "data", "workflowuuid", "workflow", "createdby", "createdon", "modifiedby", "modifiedon"}
	values := []string{e.Type, e.EntityName, "1", Description, string(jsonEntityData), workflow.UUID, string(jsonString), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05"), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05")}

	wfentityid, err := dbop.TableInsert("workflow_entities", columns, values)

	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
		return 0, err
	}

	pretaskdata := make(map[string]interface{})

	for _, node := range firstNodes {
		e.Log.Debug(fmt.Sprintf("Workflow %s first node %s explode ", e.WorkflowName, node.ID))
		e.explodeNode(node, wfentityid, e.DocDBCon, e.DBTx, pretaskdata)
	}

	err = e.DBTx.Commit()
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
		return 0, err
	}
	return wfentityid, nil

}

func (e *ExplodionEngine) explodeNode(node wftype.Node, workflowentityid int64, DBConn *documents.DocDB, DBTx *sql.Tx, PreTaskData map[string]interface{}) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		e.Log.PerformanceWithDuration("WorkFlow.Explosion.explodeNode", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode: %s", r))
			return
		}
	}()

	jsonData, err := json.Marshal(node.ProcessData)
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode - convert the node processdata: %s", err))
		return
	}

	PreTaskjsonData, err := json.Marshal(PreTaskData)
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode - convert the pretaskdata: %s", err))
		return
	}

	dbop := dbconn.NewDBOperation(e.UserName, DBTx, "Workflow.Explosion")

	//columns := []string{"WorkflowEntityID", "Type", "Status", "WorkflowNodeID", "PreTaskData", "ProcessData", "Page", "TranCode", "createdby", "createdon", "updatedby", "updatedon"}

	columns := []string{"workflowentityid", "type", "status", "workflownodeid", "pretaskdata", "processdata", "page", "trancode", "createdby", "createdon", "modifiedby", "modifiedon"}

	values := []string{fmt.Sprintf("%d", workflowentityid), node.Type, "1", node.ID, string(PreTaskjsonData), string(jsonData), node.Page, node.TranCode, e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05"), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05")}

	taskid, err := dbop.TableInsert("workflow_tasks", columns, values)

	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode - insert the data to database: %s", err))
		return
	}

	e.Log.Debug(fmt.Sprintf("Workflow %s node %s explode taskid %d ", e.WorkflowName, node.ID, taskid))
	roleids := []int64{}

	notification := make(map[string]interface{})
	notroles := make(map[string]interface{})
	notusers := make(map[string]interface{})

	for _, role := range node.Roles {

		if role != "" {
			rows, err := dbop.Query_Json(fmt.Sprintf("select id from roles where name = '%s'", role))

			if err != nil {
				e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode to get the role assignment: %s", err))

			} else if len(rows) == 0 {
				e.Log.Error(fmt.Sprintf("System does not find the role: %s", role))
			} else {
				roleid := rows[0]["id"].(int64)
				roleids = append(roleids, roleid)
				//columns = []string{"WorkflowTaskID", "RoleID", "createdby", "createdon", "updatedby", "updatedon"}
				columns = []string{"workflowtaskid", "roleid", "createdby", "createdon", "modifiedby", "modifiedon"}
				values = []string{fmt.Sprintf("%d", taskid), fmt.Sprintf("%d", roleid), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05"), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05")}

				_, err = dbop.TableInsert("workflow_task_assignments", columns, values)

				if err != nil {
					e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode during adding assignment: %s", err))
					return
				}

				notroles[role] = 1
			}
		}
	}

	userids := []int64{}

	for _, user := range node.Users {
		if user != "" {
			rows, err := dbop.Query_Json(fmt.Sprintf("select id, loginname from users where loginname = '%s' OR name = '%s'", user, user))

			if err != nil {
				e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode during gettign userid: %s", err))
			} else if len(rows) == 0 {
				e.Log.Error(fmt.Sprintf("System does not find the user: %s", user))
			} else {
				userid := rows[0]["id"].(int64)
				loginname := rows[0]["loginname"].(string)
				userids = append(userids, userid)
				//columns = []string{"WorkflowTaskID", "UserID", "createdby", "createdon", "updatedby", "updatedon"}
				columns = []string{"workflowtaskid", "userid", "createdby", "createdon", "modifiedby", "modifiedon"}
				values = []string{fmt.Sprintf("%d", taskid), fmt.Sprintf("%d", userid), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05"), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05")}

				_, err = dbop.TableInsert("workflow_task_assignments", columns, values)

				if err != nil {
					e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode during adding assignment: %s", err))
					return
				}
				notusers[loginname] = 1
			}
		}
	}
	node.Roleids = roleids
	node.Userids = userids

	if (node.Type == "task" || node.Type == "gateway") && node.Page != "" {

		e.Log.Debug(fmt.Sprintf("Workflow %s node %s explode page %s and send notification", e.WorkflowName, node.ID, node.Page))
		notification["type"] = "workflow"
		notification["entity"] = e.EntityName
		notification["workflow"] = e.WorkflowName
		notification["workflownode"] = node.Name
		notification["workflownodeid"] = node.ID
		notification["workflowtaskid"] = taskid
		notification["workflowentityid"] = workflowentityid
		notification["status"] = "1"
		notification["roles"] = notroles
		notification["receipts"] = notusers
		notification["sender"] = e.UserName
		notification["topic"] = "workflow task created for " + e.EntityName
		notification["message"] = "workflow task created for " + e.EntityName
		notification["uuid"] = uuid.New().String()

		err = notifications.CreateNewNotification(notification, e.UserName)
		if err != nil {
			e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode during creating notification: %s", err))
		}

		columns = []string{"notificationuuid"}
		values = []string{notification["uuid"].(string)}
		datatypes := []int{int(0)}
		Where := fmt.Sprintf("id = %d", taskid)
		_, err = dbop.TableUpdate("workflow_tasks", columns, values, datatypes, Where)
		if err != nil {
			e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode during updating notification: %s", err))
		}

	}

	//columns = []string{"WorkflowEntityID", "WorkflowTaskID", "Type", "Status", "createdby", "createdon", "updatedby", "updatedon"}
	columns = []string{"workflowentityid", "workflowtaskid", "typecode", "status", "createdby", "createdon", "modifiedby", "modifiedon"}
	values = []string{fmt.Sprintf("%d", workflowentityid), fmt.Sprintf("%d", taskid), "create task", "1", e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05"), e.UserName, time.Now().UTC().Format("2006-01-02 15:04:05")}

	_, err = dbop.TableInsert("workflow_task_histories", columns, values)

	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.explodeNode during adding history records: %s", err))
		return
	}

	ExecuteTask(taskid, node, DBTx, DBConn, e.UserName)

}

func (e *ExplodionEngine) getNodeByID(ID string, Nodes []wftype.Node) wftype.Node {
	for _, node := range Nodes {
		if node.ID == ID {
			return node
		}
	}
	return wftype.Node{}
}

func (e *ExplodionEngine) getWorkFlowbyName() (primitive.M, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		e.Log.PerformanceWithDuration("WorkFlow.Explosion.getWorkFlowbyName", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.getWorkFlowbyName: %s", r))
			return
		}
	}()

	if e.WorkflowName == "" {
		return nil, fmt.Errorf("Workflow name is empty")
	}

	e.Log.Info(fmt.Sprintf("Start to get workflow data %s's %s", e.WorkflowName, "Retrieven"))

	filter := bson.M{"name": e.WorkflowName, "isdefault": true}

	workflowM, err := e.DocDBCon.QueryCollection("WorkFlow", filter, nil)
	if err != nil {
		e.Log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.getWorkFlowbyName: %s", err))
		return nil, err
	}
	e.Log.Debug(fmt.Sprintf("Workflow %s data %v ", e.WorkflowName, workflowM))
	e.Log.Info(fmt.Sprintf("End to get workflow data %s's %s", e.WorkflowName, "Retrieven"))
	return workflowM[0], nil

}

func convertToMap(m primitive.M) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range m {
		result[key] = value
	}

	return result
}

func convertWorkFlowNodeToJson(node wftype.Node) []byte {
	/*	result := make(map[string]interface{})
		result["name"] = node.Name
		result["id"] = node.ID
		result["description"] = node.Description
		result["type"] = node.Type
		result["page"] = node.Page
		result["trancode"] = node.TranCode
		result["roles"] = node.Roles
		result["users"] = node.Users
		result["roleids"] = node.Roleids
		result["userids"] = node.Userids
		result["precondition"] = node.PreCondition
		result["postcondition"] = node.PostCondition
		result["processdata"] = node.ProcessData
		result["routingtables"] = node.RoutingTables */

	result, err := json.Marshal(node)
	if err != nil {
		return nil
	}
	/*
		var resultMap map[string]interface{}
		err = json.Unmarshal(result, &resultMap)
		if err != nil {
			return nil
		}
	*/
	return result
}

func GetWorkFlowbyUUID(uuid string, UserName string, DocDBCon documents.DocDB) (wftype.WorkFlow, primitive.M, error) {
	log := logger.Log{}
	log.ModuleName = logger.Framework
	log.ControllerName = "workflow"
	log.User = UserName

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.funcs.NewFuncs", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", r))
			return
		}
	}()

	filter := bson.M{"uuid": uuid}

	workflowM, err := DocDBCon.QueryCollection("WorkFlow", filter, nil)
	if err != nil {
		log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.getWorkFlowbyUUID: %s", err))
		return wftype.WorkFlow{}, nil, err
	}

	if workflowM == nil {
		log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", "Workflow not found"))
		return wftype.WorkFlow{}, nil, err
	}

	jsonString, err := json.Marshal(workflowM[0])
	if err != nil {
		log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
		return wftype.WorkFlow{}, nil, err
	}

	var workflow wftype.WorkFlow
	err = json.Unmarshal(jsonString, &workflow)
	if err != nil {
		log.Error(fmt.Sprintf("Error in WorkFlow.Explosion.Explode: %s", err))
		return wftype.WorkFlow{}, nil, err
	}
	return workflow, workflowM[0], nil
}
