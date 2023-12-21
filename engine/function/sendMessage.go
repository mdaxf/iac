package funcs

import (
	//	"context"
	"encoding/json"
	"fmt"
	"time"

	//	dapr "github.com/dapr/go-sdk/client"
	"github.com/mdaxf/iac/com"
)

type SendMessageFuncs struct {
}

// Execute executes the SendMessageFuncs function.
// It sets the start time, defers the calculation of elapsed time and logging of performance,
// and recovers from any panics that occur during execution.
// It retrieves the inputs, sets the topic and data, and marshals the data into JSON format.
// If the topic is empty, it logs an error and returns.
// It then invokes the SignalRClient or IACMessageBusClient to send the message.
// Finally, it sets the outputs of the function.
func (cf *SendMessageFuncs) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendEmessage.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendEmessage.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.SendEmessage.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendEmessage.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("SendMessageFuncs Execute: %v", f))

	namelist, valuelist, _ := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("SendMessageFuncs Execute: %v, %v", namelist, valuelist))

	Topic := ""
	data := make(map[string]interface{})
	for i, name := range namelist {
		if name == "Topic" {
			Topic = valuelist[i]
			continue
		}
		data[name] = valuelist[i]
	}

	if Topic == "" {
		f.iLog.Error(fmt.Sprintf("SendMessageFuncs validate wrong: %v", "Topic is empty"))
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error:%v", err))
		return
	}
	// Convert JSON byte array to string
	jsonString := string(jsonData)

	f.iLog.Debug(fmt.Sprintf("SendMessageFuncs Execute: topic, %s, message: %v", Topic, jsonString))
	//iacmb.IACMB.Channel.Write(jsonString)

	//c.Invoke("send", "Test", "this is a message from the GO client", "")
	if f.SignalRClient != nil {
		f.SignalRClient.Invoke("send", Topic, jsonString, "")
	} else {
		com.IACMessageBusClient.Invoke("send", Topic, jsonString, "")
		<-time.After(time.Microsecond * 100)
	}
	/*client, err := dapr.NewClient()

	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error creating Dapr client for client '%s': %v\n", "clientID", err))
		return
	}

	defer client.Close()

	// Publish the message to the client's topic
	err = client.PublishEvent(context.Background(), "IACF-DAPR-BackGround-Function-clientID", Topic, jsonData)

	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error publishing message to client '%s': %v\n", "IACF-DAPR-clientID", err))
		return
	}
	*/
	outputs := make(map[string][]interface{})
	f.SetOutputs(f.convertMap(outputs))
}

// Validate validates the SendMessageFuncs function.
// It checks if the namelist and valuelist are empty,
// and if the "Topic" name is present in the namelist.
// Returns true if the validation passes, otherwise returns false with an error.
// It also logs the performance of the function.
func (cf *SendMessageFuncs) Validate(f *Funcs) (bool, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendEmessage.Validate", elapsed)
	}()
	/*	defer func() {
			if err := recover(); err != nil {
				f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendEmessage.Validate with error: %s", err))
				f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendEmessage.Validate with error: %s", err)
				return
			}
		}()
	*/
	f.iLog.Debug(fmt.Sprintf("SendMessageFuncs validate: %v", f))
	namelist, valuelist, _ := f.SetInputs()

	if len(namelist) == 0 {
		return false, fmt.Errorf("SendMessageFuncs validate: %v", "namelist is empty")
	}

	if len(valuelist) == 0 {
		return false, fmt.Errorf("SendMessageFuncs validate: %v", "valuelist is empty")
	}
	found := false
	for _, name := range namelist {
		if name == "" {
			return false, fmt.Errorf("SendMessageFuncs validate: %v", "name is empty")
		}

		if name == "Topic" {
			found = true
		}
	}
	if !found {
		return false, fmt.Errorf("SendMessageFuncs validate: %v", "Topic is not found")
	}

	return true, nil
}
