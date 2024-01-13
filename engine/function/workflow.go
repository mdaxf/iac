package funcs

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/workflow"
)

type WorkFlowFunc struct{}

func (w *WorkFlowFunc) Execute_Explode(f *Funcs) {
	// function execution start time
	startTime := time.Now()
	defer func() {
		// calculate elapsed time
		elapsed := time.Since(startTime)
		// log performance with duration
		f.iLog.PerformanceWithDuration("engine.funcs.WorkFlow.Execute", elapsed)
	}()

	defer func() {
		// recover from any panics
		if err := recover(); err != nil {
			// log the error
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
			// cancel execution and set error message
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err)
		}
	}()

	namelist, valuelist, _ := f.SetInputs()

	EntityName := ""
	EntityType := ""
	WorkFlowName := ""
	Description := ""
	Data := make(map[string]interface{})

	for i, name := range namelist {
		Data[name] = valuelist[i]

		if name == "EntityName" {
			EntityName = valuelist[i]
		} else if name == "EntityType" {
			EntityType = valuelist[i]
		} else if name == "WorkFlowName" {
			WorkFlowName = valuelist[i]
		} else if name == "Description" {
			Description = valuelist[i]
		}
	}

	if WorkFlowName == "" || EntityName == "" {
		err := fmt.Errorf("There is value for the WorkFlowName and EntityName")
		f.iLog.Error(fmt.Sprintf("failed to create the notification: %v", err))
		f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
		return
	}
	f.iLog.Debug(fmt.Sprintf("%s, %s", Description, EntityType))

	wfe := workflow.NewExplosion(WorkFlowName, EntityName, EntityType, f.SystemSession["UserName"].(string), "")
	err := wfe.Explode(Description, Data)

	if err != nil {
		f.iLog.Error(fmt.Sprintf("failed to create the notification: %v", err))
		f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
		return
	}
}

func (w *WorkFlowFunc) Execute_StartTask(f *Funcs) {
	// function execution start time
	startTime := time.Now()
	defer func() {
		// calculate elapsed time
		elapsed := time.Since(startTime)
		// log performance with duration
		f.iLog.PerformanceWithDuration("engine.funcs.WorkFlow.Execute", elapsed)
	}()

	defer func() {
		// recover from any panics
		if err := recover(); err != nil {
			// log the error
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
			// cancel execution and set error message
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err)
		}
	}()

	namelist, valuelist, _ := f.SetInputs()

	TaskID := 0

	for i, name := range namelist {
		if name == "TaskID" {
			TaskID = com.ConverttoIntwithDefault(valuelist[i], 0)
		}
	}

	if TaskID == 0 {
		err := fmt.Errorf("TaskID cannot be 0 or no value")
		f.iLog.Error(fmt.Sprintf("failed to get the input : %v", err))
		f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
		return
	}
	user := f.SystemSession["UserName"].(string)

	wft := workflow.NewWorkFlowTaskType(int64(TaskID), user)
	err := wft.StartTask()

	if err != nil {

		f.iLog.Error(fmt.Sprintf("failed to start the tasks for the user %s with error: %v", user, err))
		f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
		return
	}
}

func (w *WorkFlowFunc) Execute_CompleteTask(f *Funcs) {
	// function execution start time
	startTime := time.Now()
	defer func() {
		// calculate elapsed time
		elapsed := time.Since(startTime)
		// log performance with duration
		f.iLog.PerformanceWithDuration("engine.funcs.WorkFlow.Execute", elapsed)
	}()

	defer func() {
		// recover from any panics
		if err := recover(); err != nil {
			// log the error
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
			// cancel execution and set error message
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err)
		}
	}()

	namelist, valuelist, _ := f.SetInputs()

	TaskID := 0

	for i, name := range namelist {
		if name == "TaskID" {
			TaskID = com.ConverttoIntwithDefault(valuelist[i], 0)
		}
	}

	if TaskID == 0 {
		err := fmt.Errorf("TaskID cannot be 0 or no value")
		f.iLog.Error(fmt.Sprintf("failed to get the input : %v", err))
		f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
		return
	}
	user := f.SystemSession["UserName"].(string)

	wft := workflow.NewWorkFlowTaskType(int64(TaskID), user)
	err := wft.CompleteTask()

	if err != nil {

		f.iLog.Error(fmt.Sprintf("failed to complete the tasks for the user %s with error: %v", user, err))
		f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WorkFlow.Execute with error: %s", err))
		return
	}
}
