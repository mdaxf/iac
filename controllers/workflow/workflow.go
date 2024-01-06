package workflow

import (
	"fmt"
	"time"

	//"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"

	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/workflow"
)

type WorkFlowController struct {
}

func (wf *WorkFlowController) GetWorkFlowbyUUID(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "workflow"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("GetTasksbyUser.workflow.ExplodeWorkFlow", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	requestobj, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.ClientID = clientid
	iLog.User = user
	data := make(map[string]interface{})
	data = requestobj["data"].(map[string]interface{})
	WorkFlowUUID := data["uuid"].(string)
	if WorkFlowUUID == "" {
		iLog.Error(fmt.Sprintf("failed to create the notification: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "WorkFlowUUID is required"})
		return
	}

	//_, WorkFlow, err := workflow.GetWorkFlowbyUUID(WorkFlowUUID, user, *documents.DocDBCon)
	WorkFlow, err := documents.DocDBCon.GetItembyUUID("WorkFlow", WorkFlowUUID)
	//err = json.Unmarshal(WorkFlowSchema, &WorkFlow)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error in getting workflow schema: %s", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": WorkFlow})
}

func (wf *WorkFlowController) ExplodeWorkFlow(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "workflow"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("GetTasksbyUser.workflow.ExplodeWorkFlow", elapsed)
	}()

	defer func() {
		err := recover()
		if err != nil {
			iLog.Error(fmt.Sprintf("Error: %v", err))
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	requestobj, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user
	data := make(map[string]interface{})
	data = requestobj["data"].(map[string]interface{})
	WorkFlowName := data["workflowname"].(string)
	EntityName := data["entityname"].(string)
	EntityType := data["entitytype"].(string)

	if WorkFlowName == "" || EntityName == "" {
		iLog.Error(fmt.Sprintf("failed to create the notification: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "WorkFlowName, EntityName are required"})
		return
	}

	wfe := workflow.NewExplosion(WorkFlowName, EntityName, EntityType, user, clientid)
	err = wfe.Explode()

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to create the notification: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
}

func (wf *WorkFlowController) GetTasksbyUser(ctx *gin.Context) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "workflow"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("WorkFlowController.workflow.GetTasksbyUser", elapsed)
	}()

	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	_, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	result, err := workflow.GetTasksbyUser(user)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to get the tasks for the user %s with error: %v", user, err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": result})

}

func (wf *WorkFlowController) GetWorkFlowTasks(ctx *gin.Context) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "workflow"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("WorkFlowController.workflow.GetWorkFlowTasks", elapsed)
	}()

	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	requestbody, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user
	data := make(map[string]interface{})
	data = requestbody["data"].(map[string]interface{})
	workflowentityid := int64(data["workflowentityid"].(float64))

	result, err := workflow.GetWorkFlowTasks(workflowentityid, user)

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to get the tasks for the user %s with error: %v", user, err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": result})

}

func (wf *WorkFlowController) StartTask(ctx *gin.Context) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "workflow"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("WorkFlowController.workflow.StartTask", elapsed)
	}()

	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	requestbody, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user
	data := make(map[string]interface{})
	data = requestbody["data"].(map[string]interface{})
	taskid := int64(data["taskid"].(float64))
	wft := workflow.NewWorkFlowTaskType(taskid, user)
	err = wft.StartTask()

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to start the tasks for the user %s with error: %v", user, err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": "OK"})

}

func (wf *WorkFlowController) CompleteTask(ctx *gin.Context) {

	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "workflow"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("WorkFlowController.workflow.CompleteTask", elapsed)
	}()

	/*	defer func() {
			err := recover()
			if err != nil {
				iLog.Error(fmt.Sprintf("Error: %v", err))
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
		}()
	*/
	requestbody, clientid, user, err := common.GetRequestBodyandUserbyJson(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get request information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user
	data := make(map[string]interface{})
	data = requestbody["data"].(map[string]interface{})
	taskid := int64(data["taskid"].(float64))
	wft := workflow.NewWorkFlowTaskType(taskid, user)
	err = wft.CompleteTask()

	if err != nil {

		iLog.Error(fmt.Sprintf("failed to complete the tasks for the user %s with error: %v", user, err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": "OK"})

}
