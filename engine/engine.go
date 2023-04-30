package engine

import (
	"log"

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/engine/types"
)

type Engine struct {
	trancode types.TranCode
}

func NewEngine(trancode types.TranCode) *Engine {
	return &Engine{trancode: trancode}
}

func (e *Engine) Execute() {
	tf := trancode.NewTranFlow(e.trancode, map[string]interface{}{}, map[string]interface{}{})
	_, err := tf.Execute()
	if err != nil {
		log.Println(err)
	}
}
