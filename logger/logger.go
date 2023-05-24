package logger

import (
	"encoding/json"
	"fmt"

	"github.com/mdaxf/iac/framework/logs"
)

var (
	Logger          *logs.IACLogger
	PortalLogger    *logs.IACLogger
	APILogger       *logs.IACLogger
	DatabaseLogger  *logs.IACLogger
	TranCodeLogger  *logs.IACLogger
	JobLogger       *logs.IACLogger
	FrameworkLogger *logs.IACLogger
)

const (
	Framework string = "Framework"
	Portal    string = "Portal"
	API       string = "API"
	Database  string = "Database"
	TranCode  string = "TranCode"
	Job       string = "Job"
)

type Log struct {
	ModuleName     string
	ControllerName string
	User           string
}

func Init() {

	FrameworkLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	FrameworkLogger.SetLogger(logs.AdapterConsole)

	PortalLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	PortalLogger.SetLogger(logs.AdapterConsole)

	//	PortalLogger.Debug("this is a debug message")

	APILogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	APILogger.SetLogger(logs.AdapterConsole)

	//	APILogger.Debug("this is a debug message")

	DatabaseLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	DatabaseLogger.SetLogger(logs.AdapterConsole)

	//	DatabaseLogger.Debug("this is a debug message")

	TranCodeLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	TranCodeLogger.SetLogger(logs.AdapterConsole)

	//	TranCodeLogger.Debug("this is a debug message")

	JobLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	JobLogger.SetLogger(logs.AdapterConsole)

	//	JobLogger.Debug("this is a debug message")

	Logger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	Logger.SetLogger(logs.AdapterConsole)

	//logs.SetLogger(logs.AdapterMultiFile, ``{"filename":"test.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}``)
}

func customFormatter(lm *logs.LogMsg) string {
	return fmt.Sprintf("[CUSTOM FILE LOGGING] %s", lm.Msg)
}

func GlobalFormatter(lm *logs.LogMsg) string {
	return fmt.Sprintf("[GLOBAL] %s", lm.Msg)
}

func (l *Log) Debug(logmsg string) {

	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Debug(logmsg)
	case API:
		APILogger.Debug(logmsg)
	case Database:
		DatabaseLogger.Debug(logmsg)
	case TranCode:
		TranCodeLogger.Debug(logmsg)
	case Job:
		JobLogger.Debug(logmsg)
	case Framework:
		FrameworkLogger.Debug(logmsg)
	default:
		Logger.Debug(logmsg)
	}

}

func (l *Log) Info(logmsg string) {
	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Info(logmsg)
	case API:
		APILogger.Info(logmsg)
	case Database:
		DatabaseLogger.Info(logmsg)
	case TranCode:
		TranCodeLogger.Info(logmsg)
	case Job:
		JobLogger.Info(logmsg)
	case Framework:
		FrameworkLogger.Info(logmsg)
	default:
		Logger.Info(logmsg)
	}
}

func (l *Log) Warn(logmsg string) {
	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Warn(logmsg)
	case API:
		APILogger.Warn(logmsg)
	case Database:
		DatabaseLogger.Warn(logmsg)
	case TranCode:
		TranCodeLogger.Warn(logmsg)
	case Job:
		JobLogger.Warn(logmsg)
	case Framework:
		FrameworkLogger.Warn(logmsg)
	default:
		Logger.Warn(logmsg)
	}
}

func (l *Log) Error(logmsg string) {
	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Error(logmsg)
	case API:
		APILogger.Error(logmsg)
	case Database:
		DatabaseLogger.Error(logmsg)
	case TranCode:
		TranCodeLogger.Error(logmsg)
	case Job:
		JobLogger.Error(logmsg)
	case Framework:
		FrameworkLogger.Error(logmsg)
	default:
		Logger.Error(logmsg)
	}
}

func (l *Log) Notice(logmsg string) {
	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Notice(logmsg)
	case API:
		APILogger.Notice(logmsg)
	case Database:
		DatabaseLogger.Notice(logmsg)
	case TranCode:
		TranCodeLogger.Notice(logmsg)
	case Job:
		JobLogger.Notice(logmsg)
	case Framework:
		FrameworkLogger.Notice(logmsg)
	default:
		Logger.Notice(logmsg)
	}

}

func (l *Log) Critical(logmsg string) {
	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Critical(logmsg)
	case API:
		APILogger.Critical(logmsg)
	case Database:
		DatabaseLogger.Critical(logmsg)
	case TranCode:
		TranCodeLogger.Critical(logmsg)
	case Job:
		JobLogger.Critical(logmsg)
	case Framework:
		FrameworkLogger.Critical(logmsg)
	default:
		Logger.Critical(logmsg)
	}

}

func (l *Log) Alert(logmsg string) {
	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Alert(logmsg)
	case API:
		APILogger.Alert(logmsg)
	case Database:
		DatabaseLogger.Alert(logmsg)
	case TranCode:
		TranCodeLogger.Alert(logmsg)
	case Job:
		JobLogger.Alert(logmsg)
	case Framework:
		FrameworkLogger.Alert(logmsg)
	default:
		Logger.Alert(logmsg)
	}
}

func (l *Log) Emergency(logmsg string) {
	logmsg = (fmt.Sprintf("%s  %s  %s", l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Emergency(logmsg)
	case API:
		APILogger.Emergency(logmsg)
	case Database:
		DatabaseLogger.Emergency(logmsg)
	case TranCode:
		TranCodeLogger.Emergency(logmsg)
	case Job:
		JobLogger.Emergency(logmsg)
	case Framework:
		FrameworkLogger.Emergency(logmsg)
	default:
		Logger.Emergency(logmsg)
	}
}

func ConvertJson(jobj interface{}) string {
	jsonString, err := json.Marshal(jobj)
	if err != nil {
		fmt.Println("Error marshaling json:", err)
		return ""
	}
	return string(jsonString)
}
