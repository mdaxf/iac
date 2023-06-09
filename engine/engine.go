package engine

import (
	"fmt"

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

type Engine struct {
	trancode types.TranCode
}

func NewEngine(trancode types.TranCode) *Engine {
	return &Engine{trancode: trancode}
}

func (e *Engine) Execute() {
	iLog := logger.Log{}
	iLog.ModuleName = logger.TranCode
	iLog.ControllerName = "Engine"
	iLog.User = "System"

	tf := trancode.NewTranFlow(e.trancode, map[string]interface{}{}, map[string]interface{}{}, nil, nil)
	_, err := tf.Execute()
	if err != nil {
		iLog.Error(fmt.Sprintf("Execute the trancode error: %v", err.Error()))
	}
}
