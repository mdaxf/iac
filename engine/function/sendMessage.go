package funcs

import (
	//	"context"
	"encoding/json"
	"fmt"

	//	dapr "github.com/dapr/go-sdk/client"
	"github.com/mdaxf/iac/com"
)

type SendMessageFuncs struct {
}

func (cf *SendMessageFuncs) Execute(f *Funcs) {
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
		f.iLog.Error(fmt.Sprintf("Error:", err))
		return
	}
	// Convert JSON byte array to string
	jsonString := string(jsonData)

	f.iLog.Debug(fmt.Sprintf("SendMessageFuncs Execute: %v", jsonString))
	//iacmb.IACMB.Channel.Write(jsonString)

	com.IACMessageBusClient.Send(Topic, jsonString)
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

}

func (cf *SendMessageFuncs) Validate(f *Funcs) (bool, error) {
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
