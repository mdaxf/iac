package engine

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

type Engine struct {
	trancode        types.TranCode
	externalinputs  map[string]interface{} // {sessionanme: value}
	externaloutputs map[string]interface{} // {sessionanme: value}
	systemsessions  map[string]interface{} // {sessionanme: value}
}

func NewEngine(trancode types.TranCode, systemSession map[string]interface{}) *Engine {
	//systemSession := make(map[string]interface{})
	systemSession["UTCTime"] = time.Now().UTC()
	systemSession["LocalTime"] = time.Now()
	if systemSession["UserNo"] == nil {
		systemSession["UserNo"] = "System"
	}
	if systemSession["UserID"] == nil {
		systemSession["UserID"] = 0
	}

	if systemSession["WorkSpace"] == nil {
		systemSession["WorkSpace"] = ""
	}
	if systemSession["ClientID"] == nil {
		systemSession["ClientID"] = ""
	}

	return &Engine{
		trancode:        trancode,
		externalinputs:  map[string]interface{}{},
		externaloutputs: map[string]interface{}{},
		systemsessions:  systemSession,
	}

}

func (e *Engine) Execute() {
	iLog := logger.Log{ModuleName: logger.TranCode, User: e.systemsessions["UserNo"].(string), ControllerName: "TranCode.Engine"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.TranCode.compareMap", elapsed)
	}()
	defer func() {
		if r := recover(); r != nil {
			iLog.Error(fmt.Sprintf("Error in Trancode.compareMap: %s", r))
			return
		}
	}()

	tf := trancode.NewTranFlow(e.trancode, e.externalinputs, e.systemsessions, nil, nil)
	_, err := tf.Execute()
	if err != nil {
		iLog.Error(fmt.Sprintf("Execute the trancode error: %v", err.Error()))
	}
}
