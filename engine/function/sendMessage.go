package funcs

import (
	"encoding/json"
	"fmt"

	iacmb "github.com/mdaxf/iac/framework/messagebus"
)

type SendMessageFuncs struct {
}

func (cf *SendMessageFuncs) Execute(f *Funcs) {
	f.iLog.Debug(fmt.Sprintf("SendMessageFuncs Execute: %v", f))

	namelist, valuelist, _ := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("SendMessageFuncs Execute: %v, %v", namelist, valuelist))

	data := make(map[string]interface{})
	for i, name := range namelist {
		data[name] = valuelist[i]
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		f.iLog.Error(fmt.Sprintf("Error:", err))
		return
	}
	// Convert JSON byte array to string
	jsonString := string(jsonData)

	iacmb.IACMB.Channel.Write(jsonString)
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
