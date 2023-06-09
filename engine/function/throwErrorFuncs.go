package funcs

import (
	"fmt"
)

type ThrowErrorFuncs struct {
}

func (cf *ThrowErrorFuncs) Execute(f *Funcs) {
	f.iLog.Debug(fmt.Sprintf("ThrowErrorFuncs Execute: %v", f))
	f.DBTx.Rollback()
	f.CtxCancel()
	f.Ctx.Done()
	return
}
