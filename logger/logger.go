package logger

import (
	"encoding/json"
	"fmt"

	"github.com/mdaxf/iac/framework/logs"
)

var (
	Logger *logs.IACLogger
)

func Init() {

	Logger = logs.NewLogger()
	//	logs.SetGlobalFormatter(GlobalFormatter)
	Logger.SetLogger(logs.AdapterConsole, `{"level":1}`)

	Logger.Debug("this is a debug message")

	//logs.SetLogger(logs.AdapterMultiFile, ``{"filename":"test.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}``)
}

func customFormatter(lm *logs.LogMsg) string {
	return fmt.Sprintf("[CUSTOM FILE LOGGING] %s", lm.Msg)
}

func GlobalFormatter(lm *logs.LogMsg) string {
	return fmt.Sprintf("[GLOBAL] %s", lm.Msg)
}

func Debug(logmsg string) {
	Logger.Debug(logmsg)
}

func Info(logmsg string) {
	Logger.Info(logmsg)
}

func Warn(logmsg string) {
	Logger.Warn(logmsg)
}

func Error(logmsg string) {
	Logger.Error(logmsg)
}

func Notice(logmsg string) {
	Logger.Notice(logmsg)
}

func Critical(logmsg string) {
	Logger.Critical(logmsg)
}

func Alert(logmsg string) {
	Logger.Alert(logmsg)
}

func Emergency(logmsg string) {
	Logger.Emergency(logmsg)
}

func ConvertJson(jobj interface{}) string {
	jsonString, err := json.Marshal(jobj)
	if err != nil {
		fmt.Println("Error marshaling json:", err)
		return ""
	}
	return string(jsonString)
}
