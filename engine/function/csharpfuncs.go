package funcs

import (
	//	"bytes"
	//	"encoding/json"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	//	"os"
	"os/exec"
	//	"reflect"
	"github.com/mdaxf/iac/logger"
)

type CSharpFuncs struct {
}

func (cf *CSharpFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.func.CSharpFuncs.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.func.CSharpFuncs.Execute with error: %s", err))
			return
		}
	}()
	namelist, _, inputs := f.SetInputs()

	cmdArgs := []string{"-c", f.Fobj.Content}
	for i := range namelist {

		cmdArgs = append(cmdArgs, fmt.Sprintf("-p:%s=%s", namelist[i], inputs[namelist[i]]))
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
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.func.CSharpFuncs.Validate", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.func.CSharpFuncs.Validate with error: %s", err))
			return
		}
	}()

	return true, nil
}

func (cf *CSharpFuncs) Testfunction(content string, inputs interface{}, outputs []string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "CSharp Function"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("engine.func.CSharpFuncs.Testfunction", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to engine.func.CSharpFuncs.Testfunction with error: %s", err))
			return
		}
	}()

	iLog.Debug(fmt.Sprintf("Test Exec Function"))

	/* pass, err := cf.Validate(f)

	if !pass {
		return nil, err
	}
	*/
	namelist := make([]string, 0)
	valuelist := make([]interface{}, 0)

	for key, value := range inputs.(map[string]interface{}) {
		namelist = append(namelist, key)
		valuelist = append(valuelist, value)
	}
	iLog.Debug(fmt.Sprintf("Test Exec Function namelist: %s valuelist:%v", namelist, valuelist))
	iLog.Debug(fmt.Sprintf("Test Exec Function content: %s", content))

	cmdArgs := []string{"-c", content}
	for i := range namelist {

		cmdArgs = append(cmdArgs, fmt.Sprintf("-p:%s=%s", namelist[i], valuelist[i]))
	}

	iLog.Debug(fmt.Sprintf("Test Exec Function cmdArgs: %s", cmdArgs))

	cmd := exec.Command("dotnet", cmdArgs...)

	iLog.Debug(fmt.Sprintf("Test Exec Function cmd: %s", cmd))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		iLog.Error(fmt.Sprintf("Test Exec Function error: %s", err.Error()))
		return nil, err
	}

	// Capture standard output and error
	combinedoutput, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	iLog.Debug(fmt.Sprintf("Test Exec Function combinedoutput: %s", combinedoutput))
	var functionoutputs map[string]interface{}

	err = json.Unmarshal(combinedoutput, &functionoutputs)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return functionoutputs, nil
}
