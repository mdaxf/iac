package funcs

import (
	"encoding/json"
	"log"
)

type InputMapFuncs struct{}

func (cf *InputMapFuncs) Execute(f *Funcs) {
	namelist, valuelist, _ := f.SetInputs()

	mapstr := f.Fobj.Content

	var data map[string]interface{}

	// Convert the string to JSON format
	err := json.Unmarshal([]byte(mapstr), &data)
	if err != nil {
		log.Println("Error:", err)
	}
	outputs := make(map[string]interface{})

	for key := range data {
		mapedinput := data[key]
		for i := range namelist {
			if mapedinput == namelist[i] {
				outputs[key] = valuelist[i]
			}
		}
	}

	f.SetOutputs(outputs)
}
