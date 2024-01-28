package com

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/engine/types"
	"github.com/mdaxf/iac/logger"
)

func SendTestResultMessageBus(trancode string, funcgroup string, function string,
	type_ string, status string,
	inputs map[string]interface{}, outputs map[string]interface{},
	systemsession map[string]interface{}, usersession map[string]interface{},
	err error, ClientID string, User string) {

	go func() {
		iLog := logger.Log{ModuleName: logger.TranCode, User: "System", ControllerName: "TransCode"}

		startTime := time.Now()
		defer func() {
			elapsed := time.Since(startTime)
			iLog.PerformanceWithDuration("engine.TranCode.SendTestResultMessageBus", elapsed)
		}()

		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("There is error to engine.TranCode.SendTestResultMessageBus with error: %s", err))
				//	f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.ThrowErrorFuncs.Execute with error: %s", err)

			}
		}()

		// Send message to IAC_TRANCODE_TEST_RESULT

		// 1. Create message
		Message := make(map[string]interface{})

		Message["ClientID"] = ClientID
		Message["trancode"] = trancode
		Message["funcgroup"] = funcgroup
		Message["function"] = function
		Message["type"] = type_
		Message["status"] = status
		Message["inputs"] = inputs
		Message["outputs"] = outputs
		Message["systemsession"] = systemsession
		Message["usersession"] = usersession
		Message["err"] = err

		jsonData, err := json.Marshal(Message)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error:%v", err))
			return
		}

		jsonString := string(jsonData)

		iLog.Debug(fmt.Sprintf("SendMessageFuncs Execute: topic, %s, message: %v", types.TranCodeTestResultMessageBus, jsonString))

		// 2. Send message

		com.IACMessageBusClient.Invoke("send", types.TranCodeTestResultMessageBus, jsonString)
		<-time.After(time.Microsecond * 100)
	}()
}
