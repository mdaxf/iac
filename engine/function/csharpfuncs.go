package funcs

import (
	//	"bytes"
	//	"encoding/json"
	"fmt"
	"log"

	//	"os"
	"os/exec"
	//	"reflect"
)

type CSharpFuncs struct {
}

func (cf *CSharpFuncs) Execute(f *Funcs) {
	namelist, valuelist, _ := f.SetInputs()

	cmdArgs := []string{"-c", f.Fobj.Content}
	for i := range namelist {

		cmdArgs = append(cmdArgs, fmt.Sprintf("-p:%s=%s", namelist[i], valuelist[i]))
	}
	cmd := exec.Command("dotnet", cmdArgs...)

	// Capture standard output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	// Decode the output object from the command output
	f.SetOutputs(f.ConvertfromBytes(output))
}

func (cf *CSharpFuncs) Validate(f *Funcs) (bool, error) {

	return true, nil
}

func (cf *CSharpFuncs) Testfunction(f *Funcs) (bool, error) {

	return true, nil
}
