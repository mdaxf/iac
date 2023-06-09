package funcgroup

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"database/sql"

	funcs "github.com/mdaxf/iac/engine/function"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
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
}

func NewFGroup(dbTx *sql.Tx, fgobj types.FuncGroup, nextfuncgroup string, systemSession, userSession, externalinputs, externaloutputs map[string]interface{}, ctx context.Context, ctxcancel context.CancelFunc) *FGroup {
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
	}

}
func (c *FGroup) Execute() {
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
		f := &funcs.Funcs{
			Fobj:                fobj,
			DBTx:                c.DBTx,
			SystemSession:       systemSession,
			UserSession:         userSession,
			Externalinputs:      externalinputs,
			Externaloutputs:     externaloutputs,
			FuncCachedVariables: funcCachedVariables,
		}

		f.Execute()
		userSession = f.UserSession
		funcCachedVariables = f.FuncCachedVariables
		externalinputs = f.Externalinputs
		externaloutputs = f.Externaloutputs
	}
	c.UserSession = userSession
	c.Externalinputs = externalinputs
	c.Externaloutputs = externaloutputs
	c.Nextfuncgroup = c.CheckRouter(c.FGobj.RouterDef)

	c.iLog.Info(fmt.Sprintf("End process function group %s's %s ", c.FGobj.Name, reflect.ValueOf(c.Execute).Kind().String()))
	c.iLog.Debug(fmt.Sprintf("systemSession: %s", logger.ConvertJson(c.SystemSession)))
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

	switch vartype {
	case "systemSession":
		if c.SystemSession[variable] != nil {
			for i, value := range values {
				if c.SystemSession[variable] == value {
					c.iLog.Info(fmt.Sprintf("End process function group %s's %s 's Next func group: %s", c.FGobj.Name, reflect.ValueOf(c.CheckRouter).Kind().String(), nextfuncgroups[i]))
					return nextfuncgroups[i]
				}
			}
		}
	case "userSession":
		if c.UserSession[variable] != nil {
			for i, value := range values {
				if c.UserSession[variable] == value {
					c.iLog.Info(fmt.Sprintf("End process function group %s's %s 's Next func group: %s", c.FGobj.Name, reflect.ValueOf(c.CheckRouter).Kind().String(), nextfuncgroups[i]))
					return nextfuncgroups[i]
				}
			}
		}
	case "funcCachedVariables":
		arr := strings.Split(variable, ".")
		if len(arr) == 2 {

			if c.funcCachedVariables[arr[0]] != nil {
				tempobj := c.funcCachedVariables[arr[0]].(map[string]interface{})
				if tempobj[arr[1]] != nil {
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
