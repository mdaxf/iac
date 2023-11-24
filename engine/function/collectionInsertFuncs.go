package funcs

import (
	"fmt"
	"time"
)

type CollectionInsertFuncs struct {
}

func (cf *CollectionInsertFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.func.CollectionInsertFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.func.CollectionInsertFuncs.Execute with error: %s", err))
			return
		}
	}()
	f.iLog.Debug(fmt.Sprintf("CollectionInsertFuncs Execute: %v", f))

}
