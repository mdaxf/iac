package funcgroup

import (
	"strings"

	"database/sql"

	funcs "github.com/mdaxf/iac/engine/function"
	"github.com/mdaxf/iac/engine/types"
)

type FGroup struct {
	FGobj               types.FuncGroup
	DBTx                *sql.Tx
	Nextfuncgroup       string
	SystemSession       map[string]interface{} // {sessionanme: value}
	UserSession         map[string]interface{} // {sessionanme: value}
	Externalinputs      map[string]interface{} // {sessionanme: value}
	Externaloutputs     map[string]interface{} // {sessionanme: value}
	funcCachedVariables map[string]interface{}
}

func NewFGroup(dbTx *sql.Tx, fgobj types.FuncGroup, nextfuncgroup string, systemSession, userSession, externalinputs, externaloutputs map[string]interface{}) *FGroup {
	return &FGroup{
		FGobj:               fgobj,
		DBTx:                dbTx,
		Nextfuncgroup:       nextfuncgroup,
		SystemSession:       systemSession,
		UserSession:         userSession,
		Externalinputs:      externalinputs,
		Externaloutputs:     externaloutputs,
		funcCachedVariables: map[string]interface{}{},
	}

}
func (c *FGroup) Execute() {

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
}

func (c *FGroup) CheckRouter(RouterDef types.RouterDef) string {
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
					return nextfuncgroups[i]
				}
			}
		}
	case "userSession":
		if c.UserSession[variable] != nil {
			for i, value := range values {
				if c.UserSession[variable] == value {
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
							return nextfuncgroups[i]
						}
					}
				}
			}
		}
	}
	return defaultfuncgroup
}
