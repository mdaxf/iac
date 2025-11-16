package funcgroup

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"database/sql"

	"github.com/mdaxf/iac/documents"
	tcom "github.com/mdaxf/iac/engine/com"
	funcs "github.com/mdaxf/iac/engine/function"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac-signalr/signalr"
)

type FGroup struct {
	FGobj               types.FuncGroup
	DBTx                *sql.Tx
	Ctx                 context.Context
	CtxCancel           context.CancelFunc
	Nextfuncgroup       string
	SystemSession       map[string]interface{} // {sessionanme: value}
	UserSession         map[string]interface{} // {sessionanme: value}
	Externalinputs      map[string]interface{} // {sessionanme: value}
	Externaloutputs     map[string]interface{} // {sessionanme: value}
	funcCachedVariables map[string]interface{}
	iLog                logger.Log
	DocDBCon            *documents.DocDB
	SignalRClient       signalr.Client
	ErrorMessage        string
	TestwithSc          bool
	TestResults         map[string]interface{}
}

// NewFGroup creates a new instance of FGroup.
// It takes various parameters including DocDBCon, SignalRClient, dbTx, fgobj, nextfuncgroup, systemSession, userSession, externalinputs, externaloutputs, ctx, and ctxcancel.
// It initializes the FGroup struct with the provided values and returns a pointer to the newly created FGroup instance.

func NewFGroup(DocDBCon *documents.DocDB, SignalRClient signalr.Client, dbTx *sql.Tx, fgobj types.FuncGroup, nextfuncgroup string, systemSession, userSession, externalinputs, externaloutputs map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc) *FGroup {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Function Group"
	if systemSession["UserNo"] != nil {
		log.User = systemSession["UserNo"].(string)
	} else {
		log.User = "System"
	}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		log.PerformanceWithDuration("engine.funcgroup.NewFGroup", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				log.Error(fmt.Sprintf("There is error to engine.funcgroup.NewFGroup with error: %s", err))
				return
			}
		}()
	*/
	return &FGroup{
		FGobj:               fgobj,
		DBTx:                dbTx,
		Ctx:                 ctx,
		CtxCancel:           ctxcancel,
		Nextfuncgroup:       nextfuncgroup,
		SystemSession:       systemSession,
		UserSession:         userSession,
		Externalinputs:      externalinputs,
		Externaloutputs:     externaloutputs,
		funcCachedVariables: map[string]interface{}{},
		iLog:                log,
		DocDBCon:            DocDBCon,
		SignalRClient:       SignalRClient,
		ErrorMessage:        "",
		TestwithSc:          false,
		TestResults:         map[string]interface{}{},
	}

}

// Execute executes the function group by iterating over its functions and executing each one.
// It also handles error recovery and logs performance metrics.
// It takes no parameters and returns no values.
func (c *FGroup) Execute() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		c.iLog.PerformanceWithDuration("engine.funcgroup.Execute", elapsed)
	}()

	// ROLLBACK DESIGN: This defer/recover pattern is intentional.
	// When a function within this group fails, we catch the panic and propagate it up
	// to ensure the entire transaction is rolled back.
	defer func() {
		if r := recover(); r != nil {
			// Check if this is a structured BPMError
			if bpmErr, ok := r.(*types.BPMError); ok {
				// Add function group context if not already present
				if bpmErr.Context == nil {
					bpmErr.Context = &types.ExecutionContext{}
				}
				bpmErr.Context.FunctionGroup = c.FGobj.Name

				// Log the formatted error
				c.iLog.Error(bpmErr.GetFormattedError())
				c.ErrorMessage = bpmErr.Error()

				// Update rollback reason
				if bpmErr.RollbackReason == "" {
					bpmErr.WithRollbackReason(fmt.Sprintf("Function group %s failed", c.FGobj.Name))
				}
			} else {
				// Handle non-structured errors
				errMsg := fmt.Sprintf("Panic in function group %s: %v", c.FGobj.Name, r)
				c.iLog.Error(errMsg)
				c.ErrorMessage = errMsg

				// Create a structured error for better tracking
				execContext := &types.ExecutionContext{
					FunctionGroup: c.FGobj.Name,
					ExecutionTime: startTime,
				}
				if userNo, ok := c.SystemSession["UserNo"].(string); ok {
					execContext.UserNo = userNo
				}
				if clientID, ok := c.SystemSession["ClientID"].(string); ok {
					execContext.ClientID = clientID
				}

				structuredErr := types.NewExecutionError(errMsg, nil).
					WithContext(execContext).
					WithRollbackReason(fmt.Sprintf("Unexpected error in function group %s", c.FGobj.Name))

				c.iLog.Error(structuredErr.GetFormattedError())
			}

			// Rollback and propagate the panic upward
			c.iLog.Info(fmt.Sprintf("Rolling back transaction due to error in function group %s", c.FGobj.Name))
			if c.DBTx != nil {
				c.DBTx.Rollback()
			}
			if c.CtxCancel != nil {
				c.CtxCancel()
			}

			if c.TestwithSc {
				tcom.SendTestResultMessageBus("", c.FGobj.ID, "", "End", "",
					c.Externalinputs, c.Externaloutputs, c.SystemSession, c.UserSession, fmt.Errorf(c.ErrorMessage), c.SystemSession["ClientID"].(string), c.SystemSession["UserNo"].(string))
			}

			// Re-panic to propagate to parent transcode
			panic(r)
		}
	}()

	c.iLog.Info(fmt.Sprintf("Start process function group %s's %s ", c.FGobj.Name, reflect.ValueOf(c.Execute).Kind().String()))
	c.iLog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(c.SystemSession)))
	c.iLog.Debug(fmt.Sprintf("userSession: %s", logger.ConvertJson(c.UserSession)))
	c.iLog.Debug(fmt.Sprintf("externalinputs: %s", logger.ConvertJson(c.Externalinputs)))
	c.iLog.Debug(fmt.Sprintf("externaloutputs: %s", logger.ConvertJson(c.Externaloutputs)))

	systemSession := c.SystemSession
	userSession := c.UserSession
	funcCachedVariables := map[string]interface{}{}
	externalinputs := c.Externalinputs
	externaloutputs := c.Externaloutputs

	if c.TestwithSc {

		c.TestResults["Name"] = c.FGobj.Name
		c.TestResults["Type"] = "FunctionGroup"
		c.TestResults["Inputs"] = c.Externalinputs
		c.TestResults["Outputs"] = c.Externaloutputs
		c.TestResults["UserSession"] = c.UserSession
		c.TestResults["SystemSession"] = systemSession
		c.TestResults["StartTime"] = time.Now()
		c.TestResults["Functions"] = []map[string]interface{}{}

		tcom.SendTestResultMessageBus("", c.FGobj.ID, "", "Start", "",
			externalinputs, externaloutputs, systemSession, userSession, nil, c.SystemSession["ClientID"].(string), c.SystemSession["UserNo"].(string))
	}

	for _, fobj := range c.FGobj.Functions {
		//	f := *(funcs.NewFuncs(fobj, systemSession, userSession, externalinputs, externaloutputs, funcCachedVariables))
		c.iLog.Info(fmt.Sprintf("Start process function %s", fobj.Name))
		c.iLog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(systemSession)))
		c.iLog.Debug(fmt.Sprintf("funcCachedVariables: %s", logger.ConvertJson(c.funcCachedVariables)))
		c.iLog.Debug(fmt.Sprintf("userSession: %s", logger.ConvertJson(userSession)))
		c.iLog.Debug(fmt.Sprintf("externalinputs: %s", logger.ConvertJson(externalinputs)))
		c.iLog.Debug(fmt.Sprintf("externaloutputs: %s", logger.ConvertJson(externaloutputs)))

		f := &funcs.Funcs{
			Fobj:                fobj,
			Ctx:                 c.Ctx,
			CtxCancel:           c.CtxCancel,
			DBTx:                c.DBTx,
			DocDBCon:            c.DocDBCon,
			SignalRClient:       c.SignalRClient,
			SystemSession:       systemSession,
			UserSession:         userSession,
			Externalinputs:      externalinputs,
			Externaloutputs:     externaloutputs,
			FuncCachedVariables: funcCachedVariables,
			ErrorMessage:        "",
			TestwithSc:          c.TestwithSc,
			TestResults:         make([]map[string]interface{}, 0),
		}

		f.Execute()

		if c.TestwithSc {
			funcTestResults := c.TestResults["Functions"].([]map[string]interface{})

			for _, ft := range f.TestResults {
				funcTestResults = append(funcTestResults, ft)
			}

			c.TestResults["Functions"] = funcTestResults

			tcom.SendTestResultMessageBus("", c.FGobj.ID, fobj.ID, "End", "",
				f.Externalinputs, f.Externaloutputs, f.SystemSession, f.UserSession, fmt.Errorf(f.ErrorMessage), c.SystemSession["ClientID"].(string), c.SystemSession["UserNo"].(string))
		}

		if f.ErrorMessage != "" {
			c.ErrorMessage = f.ErrorMessage
			c.iLog.Error(fmt.Sprintf("Error: %s", c.ErrorMessage))
			c.CtxCancel()
			return
		}

		c.iLog.Info(fmt.Sprintf("End process function %s", fobj.Name))
		//c.iLog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(systemSession)))
		c.iLog.Debug(fmt.Sprintf("funcCachedVariables: %s", logger.ConvertJson(funcCachedVariables)))
		c.iLog.Debug(fmt.Sprintf("userSession: %s", logger.ConvertJson(userSession)))
		c.iLog.Debug(fmt.Sprintf("externalinputs: %s", logger.ConvertJson(externalinputs)))
		c.iLog.Debug(fmt.Sprintf("externaloutputs: %s", logger.ConvertJson(externaloutputs)))

		userSession = f.UserSession
		funcCachedVariables = f.FuncCachedVariables
		externalinputs = f.Externalinputs
		externaloutputs = f.Externaloutputs
	}
	c.UserSession = userSession
	c.Externalinputs = externalinputs
	c.Externaloutputs = externaloutputs
	c.funcCachedVariables = funcCachedVariables
	c.Nextfuncgroup = c.CheckRouter(c.FGobj.RouterDef)

	if c.TestwithSc {
		c.TestResults["EndTime"] = time.Now()
		c.TestResults["Outputs"] = c.Externaloutputs
		c.TestResults["Error"] = c.ErrorMessage
		tcom.SendTestResultMessageBus("", c.FGobj.ID, "", "End", "",
			externalinputs, externaloutputs, systemSession, userSession, nil, c.SystemSession["ClientID"].(string), c.SystemSession["UserNo"].(string))

	}
	c.iLog.Info(fmt.Sprintf("End process function group %s's %s ", c.FGobj.Name, reflect.ValueOf(c.Execute).Kind().String()))
	c.iLog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(c.SystemSession)))
	c.iLog.Debug(fmt.Sprintf("funcCachedVariables: %s", logger.ConvertJson(c.funcCachedVariables)))
	c.iLog.Debug(fmt.Sprintf("userSession: %s", logger.ConvertJson(c.UserSession)))
	c.iLog.Debug(fmt.Sprintf("externalinputs: %s", logger.ConvertJson(c.Externalinputs)))
	c.iLog.Debug(fmt.Sprintf("externaloutputs: %s", logger.ConvertJson(c.Externaloutputs)))
	c.iLog.Debug(fmt.Sprintf("nextfuncgroup: %s", c.Nextfuncgroup))

}

// CheckRouter checks the router definition and determines the next function group to execute based on the provided RouterDef.
// It returns the name of the next function group.
func (c *FGroup) CheckRouter(RouterDef types.RouterDef) string {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		c.iLog.PerformanceWithDuration("engine.funcgroup.CheckRouter", elapsed)
	}()

	c.iLog.Info(fmt.Sprintf("Start process function group %s's %s ", c.FGobj.Name, reflect.ValueOf(c.CheckRouter).Kind().String()))
	c.iLog.Debug(fmt.Sprintf("RouterDef: %s", logger.ConvertJson(RouterDef)))

	variable := RouterDef.Variable
	vartype := RouterDef.Vartype
	values := RouterDef.Values
	nextfuncgroups := RouterDef.Nextfuncgroups
	defaultfuncgroup := RouterDef.Defaultfuncgroup

	c.iLog.Debug(fmt.Sprintf("variable: %s, vartype: %s, values: %s, nextfg:%s, defaultfg:%s", variable, vartype, logger.ConvertJson(values), logger.ConvertJson(nextfuncgroups), defaultfuncgroup))
	switch vartype {
	case "systemSession":
		c.iLog.Debug("start systemSession:")
		if c.SystemSession[variable] != nil {
			for i, value := range values {
				if c.SystemSession[variable] == value {
					c.iLog.Info(fmt.Sprintf("End process function group %s's %s 's Next func group: %s", c.FGobj.Name, reflect.ValueOf(c.CheckRouter).Kind().String(), nextfuncgroups[i]))
					return nextfuncgroups[i]
				}
			}
		}
	case "userSession":
		c.iLog.Debug("start userSession:")
		if c.UserSession[variable] != nil {
			for i, value := range values {
				if c.UserSession[variable] == value {
					c.iLog.Info(fmt.Sprintf("End process function group %s's %s 's Next func group: %s", c.FGobj.Name, reflect.ValueOf(c.CheckRouter).Kind().String(), nextfuncgroups[i]))
					return nextfuncgroups[i]
				}
			}
		}
	/*case "":
	case "funcCachedVariables": */
	default:
		c.iLog.Debug("start default:")
		arr := strings.Split(variable, ".")
		c.iLog.Debug(fmt.Sprintf("variable: %s arr: %s", variable, logger.ConvertJson(arr)))
		if len(arr) == 2 {
			c.iLog.Debug(fmt.Sprintf("function variables: %s", logger.ConvertJson(c.funcCachedVariables)))
			if c.funcCachedVariables[arr[0]] != nil {
				tempobj := c.funcCachedVariables[arr[0]].(map[string]interface{})
				c.iLog.Debug(fmt.Sprintf("function variables: %s", logger.ConvertJson(tempobj)))
				if tempobj[arr[1]] != nil {
					c.iLog.Debug(fmt.Sprintf("function %s variable %s value: %s", arr[0], arr[1], logger.ConvertJson(tempobj[arr[1]])))
					for i, value := range values {
						if tempobj[arr[1]] == value {
							c.iLog.Info(fmt.Sprintf("End process function group %s's %s 's Next func group: %s", c.FGobj.Name, reflect.ValueOf(c.CheckRouter).Kind().String(), nextfuncgroups[i]))
							return nextfuncgroups[i]
						}
					}
				}
			}
		}
	}

	c.iLog.Info(fmt.Sprintf("End process function group %s's %s 's Next func group: %s", c.FGobj.Name, reflect.ValueOf(c.CheckRouter).Kind().String(), defaultfuncgroup))

	return defaultfuncgroup
}
