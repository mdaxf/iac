package funcs

import (
	"fmt"
)

type CollectionDeleteFuncs struct {
}

func (cf *CollectionDeleteFuncs) Execute(f *Funcs) {
	f.iLog.Debug(fmt.Sprintf("CollectionDeleteFuncs Execute: %v", f))

}
