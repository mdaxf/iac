package funcgroup

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"database/sql"

	"github.com/mdaxf/iac/documents"
	funcs "github.com/mdaxf/iac/engine/function"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/signalrsrv/signalr"
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
}

func NewFGroup(DocDBCon *documents.DocDB, SignalRClient signalr.Client, dbTx *sql.Tx, fgobj types.FuncGroup, nextfuncgroup string, systemSession, userSession, externalinputs, externaloutputs map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc) *FGroup {
	log := logger.Log{}
	log.ModuleName = logger.TranCode
	log.ControllerName = "Function Group"
	if systemSession["User"] != nil {
		log.User = systemSession["User"].(string)
	} else {
		log.User = "System"
	}

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
	}

}
func (c *FGroup) Execute() {
	defer func() {
		if r := recover(); r != nil {
			c.iLog.Error(fmt.Sprintf("Panic: %s", r))
			c.ErrorMessage = fmt.Sprintf("Panic: %s", r)
			c.DBTx.Rollback()
			c.CtxCancel()
			return
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
		}

		f.Execute()
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

	c.iLog.Info(fmt.Sprintf("End process function group %s's %s ", c.FGobj.Name, reflect.ValueOf(c.Execute).Kind().String()))
	c.iLog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(c.SystemSession)))
	c.iLog.Debug(fmt.Sprintf("funcCachedVariables: %s", logger.ConvertJson(c.funcCachedVariables)))
	c.iLog.Debug(fmt.Sprintf("userSession: %s", logger.ConvertJson(c.UserSession)))
	c.iLog.Debug(fmt.Sprintf("externalinputs: %s", logger.ConvertJson(c.Externalinputs)))
	c.iLog.Debug(fmt.Sprintf("externaloutputs: %s", logger.ConvertJson(c.Externaloutputs)))
	c.iLog.Debug(fmt.Sprintf("nextfuncgroup: %s", c.Nextfuncgroup))

}

func (c *FGroup) CheckRouter(RouterDef types.RouterDef) string {
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
