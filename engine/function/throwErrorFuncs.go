package funcs

import (
	"fmt"
	"time"
)

type ThrowErrorFuncs struct {
}

func (cf *ThrowErrorFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.ThrowErrorFuncs.Execute", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err)

		}
	}()

	f.iLog.Debug(fmt.Sprintf("ThrowErrorFuncs Execute: %v", f))
	f.DBTx.Rollback()
	f.CtxCancel()
	f.Ctx.Done()
	f.ErrorMessage = "ThrowErrorFuncs Execute"
}
