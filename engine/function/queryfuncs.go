package funcs

import (
	"fmt"

	dbconn "github.com/mdaxf/iac/databases"
)

type QueryFuncs struct {
}

func (cf *QueryFuncs) Execute(f *Funcs) {
	namelist, valuelist, _ := f.SetInputs()

	// Create SELECT clause with aliases
	outputs, err := dbconn.QuerybyList(f.Fobj.Content, namelist, valuelist)
	if err != nil {
		fmt.Println(err)
	}

	f.SetOutputs(f.convertMap(outputs))
}

func (cf *QueryFuncs) Validate(f *Funcs) (bool, error) {

	return true, nil
}

func (cf *QueryFuncs) Testfunction(f *Funcs) (bool, error) {

	return true, nil
}
