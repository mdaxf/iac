package funcs

import (
	//	"context"

	"fmt"
	"time"
)

type SendMessagebyActiveMQ struct {
}

func (w *SendMessagebyActiveMQ) Execute(f *Funcs) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		f.iLog.PerformanceWithDuration("engine.funcs.SendMessagebyActiveMQ.Execute", elapsed)
	}()
	defer func() {
		if err := recover(); err != nil {
			f.iLog.Error(fmt.Sprintf("There is error to engine.funcs.SendMessagebyActiveMQ.Execute with error: %s", err))
			f.CancelExecution(fmt.Sprintf("There is error to engine.funcs.SendMessagebyActiveMQ.Execute with error: %s", err))
			f.ErrorMessage = fmt.Sprintf("There is error to engine.funcs.SendMessagebyActiveMQ.Execute with error: %s", err)
			return
		}
	}()

	f.iLog.Debug(fmt.Sprintf("SendMessagebyActiveMQ Execute: %v", f))

	namelist, valuelist, _ := f.SetInputs()
	f.iLog.Debug(fmt.Sprintf("SendMessagebyActiveMQ Execute: %v, %v", namelist, valuelist))
	activeMQCfg := ""
	Topic := ""
	data := make(map[string]interface{})
	for i, name := range namelist {
		if name == "Topic" {
			Topic = valuelist[i]

			continue
		} else if name == "ActiveMQ" {
			activeMQCfg = valuelist[i]
		}
		data[name] = valuelist[i]
	}

	if Topic == "" {
		f.iLog.Error(fmt.Sprintf("SendMessagebyActiveMQ validate wrong: %v", "Topic is empty"))
		return
	}

	if activeMQCfg == "" {
		f.iLog.Error(fmt.Sprintf("SendMessagebyActiveMQ validate wrong: %v", "ActiveMQ is empty"))
		return
	}

	// get activeMQ connection

	// product the data

	// send the data to activeMQ

}
