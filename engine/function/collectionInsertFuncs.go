package funcs

import (
	"fmt"
)

type CollectionInsertFuncs struct {
}

func (cf *CollectionInsertFuncs) Execute(f *Funcs) {
	f.iLog.Debug(fmt.Sprintf("CollectionInsertFuncs Execute: %v", f))

}
