package funcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type WebServiceCallFunc struct{}

type Response struct {
	Data interface{} `json:"data"`
}

func (w *WebServiceCallFunc) Execute(f *Funcs) {
	// function execution start time
	startTime := time.Now()
	defer func() {
		// calculate elapsed time
		elapsed := time.Since(startTime)
		// log performance with duration
		f.iLog.PerformanceWithDuration("engine.funcs.WebServiceCallFunc.Execute", elapsed)
	}()

	defer func() {
		// recover from any panics
		if err := recover(); err != nil {
			// log the error
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.WebServiceCallFunc.Execute with error: %s", err))
			// cancel execution and set error message
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.WebServiceCallFunc.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.WebServiceCallFunc.Execute with error: %s", err)
		}
	}()

	f.iLog.Debug(fmt.Sprintf("Start process %s : %s", "WebServiceCallFunc.Execute", f.Fobj.Name))

	namelist, valuelist, _ := f.SetInputs()

	url := ""
	method := ""
	body := make(map[string]interface{})
	timeout := 30
	for i, name := range namelist {
		namestr := strings.ToUpper(name)
		if namestr == "URL" {
			url = valuelist[i]
		} else if namestr == "METHOD" {
			method = valuelist[i]
		} else {
			body[name] = valuelist[i]
		}
	}

	if url == "" {
		f.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", "URL is empty"))
		f.CancelExecution(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", "URL is empty"))
		f.ErrorMessage = fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", "URL is empty")
		return
	}

	if method == "" {
		method = "GET"
	}

	if timeout == 0 {
		timeout = 30
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.CancelExecution(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.ErrorMessage = fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.CancelExecution(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.ErrorMessage = fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.CancelExecution(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.ErrorMessage = fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err)
		return
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.CancelExecution(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.ErrorMessage = fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		f.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", string(resp.StatusCode)))
		f.CancelExecution(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", string(resp.StatusCode)))
		f.ErrorMessage = fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", string(resp.StatusCode))
		return
	}

	var response Response

	err = json.Unmarshal(respBody, &response)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.CancelExecution(fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err))
		f.ErrorMessage = fmt.Sprintf("Error in WebServiceCallFunc.Execute: %s", err)
		return
	}
	f.iLog.Debug(fmt.Sprintf("WebServiceCallFunc response: %v", response))

	outputs := make(map[string]interface{})
	outputs["Response"] = response.Data
	outputs["StatusCode"] = resp.StatusCode

	f.SetOutputs(outputs)
}
