package funcs

import (
	"fmt"
	"time"
)

type CollectionUpdateFuncs struct {
}

func (cf *CollectionUpdateFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.func.CollectionUpdateFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.func.CollectionUpdateFuncs.Execute with error: %s", err))
			return
		}
	}()
	f.iLog.Debug(fmt.Sprintf("CollectionUpdateFuncs Execute: %v", f))

}
