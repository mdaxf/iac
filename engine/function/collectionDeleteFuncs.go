package funcs

import (
	"fmt"
	"time"
)

type CollectionDeleteFuncs struct {
}

func (cf *CollectionDeleteFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.func.CollectionDeleeFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.func.CollectionDeleeFuncs.Execute with error: %s", err))
			return
		}
	}()
	f.iLog.Debug(fmt.Sprintf("CollectionDeleteFuncs Execute: %v", f))

}
