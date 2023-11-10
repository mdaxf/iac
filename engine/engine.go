package engine

import (
	"fmt"

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

type Engine struct {
	trancode        types.TranCode
	externalinputs  map[string]interface{} // {sessionanme: value}
	externaloutputs map[string]interface{} // {sessionanme: value}
}

func NewEngine(trancode types.TranCode) *Engine {
	return &Engine{trancode: trancode, externalinputs: map[string]interface{}{}, externaloutputs: map[string]interface{}{}}
}

func (e *Engine) Execute() {
	defer func() {
		if r := recover(); r != nil {
			//outputs := make(map[string]interface{})
			return
		}
	}()
	iLog := logger.Log{}
	iLog.ModuleName = logger.TranCode
	iLog.ControllerName = "Engine"
	iLog.User = "System"

	tf := trancode.NewTranFlow(e.trancode, e.externalinputs, e.externaloutputs, nil, nil)
	_, err := tf.Execute()
	if err != nil {
		iLog.Error(fmt.Sprintf("Execute the trancode error: %v", err.Error()))
	}
}
