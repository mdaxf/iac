package funcs

import (
	"fmt"
)

type CollectionUpdateFuncs struct {
}

func (cf *CollectionUpdateFuncs) Execute(f *Funcs) {
	f.iLog.Debug(fmt.Sprintf("CollectionUpdateFuncs Execute: %v", f))

}
