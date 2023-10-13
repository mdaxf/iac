package logger

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

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

func Init(config map[string]interface{}) {

	FrameworkLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//FrameworkLogger.SetLogger(logs.AdapterConsole)
	setLogger(FrameworkLogger, config, "Framework")

	PortalLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//PortalLogger.SetLogger(logs.AdapterConsole)
	setLogger(PortalLogger, config, "Portal")

	//	PortalLogger.Debug("this is a debug message")

	APILogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//APILogger.SetLogger(logs.AdapterConsole)
	setLogger(APILogger, config, "API")
	//	APILogger.Debug("this is a debug message")

	DatabaseLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//DatabaseLogger.SetLogger(logs.AdapterConsole)
	setLogger(DatabaseLogger, config, "Database")
	//	DatabaseLogger.Debug("this is a debug message")

	TranCodeLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//TranCodeLogger.SetLogger(logs.AdapterConsole)
	setLogger(TranCodeLogger, config, "BPM")
	//	TranCodeLogger.Debug("this is a debug message")

	JobLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//JobLogger.SetLogger(logs.AdapterConsole)
	setLogger(JobLogger, config, "Job")
	//	JobLogger.Debug("this is a debug message")

	Logger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//Logger.SetLogger(logs.AdapterConsole)
	setLogger(Logger, config, "Log")
	//logs.SetLogger(logs.AdapterMultiFile, ``{"filename":"test.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}``)
}

func setLogger(loger *logs.IACLogger, config map[string]interface{}, logtype string) {
	logadapter := config["adapter"]
	if logadapter == nil {
		logadapter = "console"
	}
	level := 3
	levelstr := config["level"]
	if levelstr == nil {
		levelstr = "debug"
	}
	switch levelstr {
	case "emergency":
		level = 0
	case "alert":
		level = 1
	case "critical":
		level = 2
	case "error":
		level = 3
	case "warning":
		level = 4
	case "notice":
		level = 5
	case "info":
		level = 6
	case "debug":
		level = 7
	default:
		level = 3
	}
	fullfilename := ""
	maxlines := 1000000
	maxsize := 1024 * 1024 * 1024
	//	fmt.Println(fmt.Sprintf(`{"level":%d}, %d`, level, logadapter))
	if logadapter == "file" || logadapter == "multifile" {
		adapterconfig := make(map[string]interface{})
		if config["adapterconfig"] != nil {
			adapterconfig = config["adapterconfig"].(map[string]interface{})
		} else {
			adapterconfig = config
		}

		filename := adapterconfig["file"]
		if filename == nil {
			filename = "iac.log"
		}
		suffix := filepath.Ext(filename.(string))
		fileNameOnly := strings.TrimSuffix(filename.(string), suffix)

		folder := adapterconfig["folder"]
		if folder == nil {
			folder = "c:\\\\temp"
		}

		if suffix == "" {
			fullfilename = fmt.Sprintf("%s\\\\%s_%s.log", folder, logtype, fileNameOnly)
		} else {
			fullfilename = fmt.Sprintf("%s\\\\%s_%s%s", folder, logtype, fileNameOnly, suffix)
		}
		if adapterconfig["maxlines"] != nil {
			maxlines = adapterconfig["maxlines"].(int) //maxlines := config["maxlines"].(int)
		}

		if adapterconfig["maxsize"] != nil {
			maxsize = adapterconfig["maxsize"].(int)
		}
	} else if logadapter == "documentdb" {
		conn := "mongodb://localhost:27017"
		db := "IAC_Cache"
		collection := "cache"
		adapterconfig := make(map[string]interface{})
		if config["adapterconfig"] != nil {
			adapterconfig = config["adapterconfig"].(map[string]interface{})
		} else {
			adapterconfig = config
		}

		if adapterconfig["documentdb"] != nil {
			documentdbcfg := adapterconfig["documentdb"].(map[string]interface{})
			if documentdbcfg["conn"] != nil {
				conn = documentdbcfg["conn"].(string)
			}
			if documentdbcfg["db"] != nil {
				db = documentdbcfg["db"].(string)
			}
			if documentdbcfg["collection"] != nil {
				collection = documentdbcfg["collection"].(string)
			}
		}
		loger.SetLogger(logs.AdapterDocumentDB, fmt.Sprintf(`{"level":"%d", "conn":"%s", "db":"%s", "collection":"%s"}`, level, conn, db, collection))
		return
	}

	switch logadapter {
	case "console":
		loger.SetLogger(logs.AdapterConsole, fmt.Sprintf(`{"level":%d}`, level))
	case "file":
		//	logs.SetLogger(logs.AdapterFile, `{"filename":"test.log"}`)
		loger.SetLogger(logs.AdapterFile, fmt.Sprintf(`{"level":"%d","filename":"%s","maxlines":"%d","maxsize":"%d"}`, level, fullfilename, maxlines, maxsize))
	case "multifile":
		loger.SetLogger(logs.AdapterMultiFile, fmt.Sprintf(`{"filename":"%s","level":"%s", }`, fullfilename, level))
	case "smtp":
		loger.SetLogger(logs.AdapterMail, fmt.Sprintf(`{"username":"%s","password":"%s","host":"%s","subject":"%s","sendTos":"%s","level":"%d"}`, config["username"], config["password"], config["host"], config["subject"], config["sendTos"], level))
	case "conn":
		loger.SetLogger(logs.AdapterConn, fmt.Sprintf(`{"net":"%s","addr":"%s","level":"%d"}`, config["net"], config["addr"], level))
	case "documentdb":
		loger.SetLogger(logs.AdapterDocumentDB, fmt.Sprintf(`{"level":"%d"}`, level))
	default:
		loger.SetLogger(logs.AdapterConsole)
	}
	//loger.SetLogger(logs.AdapterFile, `{"filename":"test.log","level":7}`)
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
