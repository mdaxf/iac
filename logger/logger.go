package logger

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mdaxf/iac/com"
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
	ClientID       string
}

// Init initializes the logger with the provided configuration.
// The configuration is a map[string]interface{} containing the following optional keys:
// - "performance": a boolean indicating whether performance logging should be enabled (default: false)
// - "performancethread": an integer specifying the performance threshold (default: 10)
// The logger is initialized with different loggers for different components, such as "Framework", "Portal", "API", etc.
// Each logger is configured using the setLogger function with the corresponding component name.
// The logger's performance and threshold settings are also set based on the configuration.
// Note: This function assumes that the logs package is imported as "logs" and the com package is imported as "com".

func Init(config map[string]interface{}) {

	performance := false
	performancestr := config["performance"]
	if performancestr != nil && performancestr.(bool) == true {
		performance = true
	}

	performancethrehold := 10

	if config["performancethread"] != nil {
		performancethrehold = com.ConverttoIntwithDefault(config["performancethread"], 10)
	}

	fmt.Println("performance:", performance)

	FrameworkLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//FrameworkLogger.SetLogger(logs.AdapterConsole)
	setLogger(FrameworkLogger, config, "Framework")
	FrameworkLogger.Perf = performance
	FrameworkLogger.Threhold = performancethrehold

	PortalLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//PortalLogger.SetLogger(logs.AdapterConsole)
	setLogger(PortalLogger, config, "Portal")
	PortalLogger.Perf = performance
	//	PortalLogger.Debug("this is a debug message")
	PortalLogger.Threhold = performancethrehold

	APILogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//APILogger.SetLogger(logs.AdapterConsole)
	setLogger(APILogger, config, "API")
	//	APILogger.Debug("this is a debug message")
	APILogger.Perf = performance

	DatabaseLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//DatabaseLogger.SetLogger(logs.AdapterConsole)
	setLogger(DatabaseLogger, config, "Database")
	//	DatabaseLogger.Debug("this is a debug message")
	DatabaseLogger.Perf = performance
	DatabaseLogger.Threhold = performancethrehold

	TranCodeLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//TranCodeLogger.SetLogger(logs.AdapterConsole)
	setLogger(TranCodeLogger, config, "BPM")
	//	TranCodeLogger.Debug("this is a debug message")
	TranCodeLogger.Perf = performance
	TranCodeLogger.Threhold = performancethrehold

	JobLogger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//JobLogger.SetLogger(logs.AdapterConsole)
	setLogger(JobLogger, config, "Job")
	//	JobLogger.Debug("this is a debug message")
	JobLogger.Perf = performance
	JobLogger.Threhold = performancethrehold

	Logger = logs.NewLogger()
	//logs.SetGlobalFormatter(GlobalFormatter())
	//Logger.SetLogger(logs.AdapterConsole)
	setLogger(Logger, config, "Log")
	Logger.Perf = performance
	Logger.Threhold = performancethrehold
	//logs.SetLogger(logs.AdapterMultiFile, ``{"filename":"test.log","separate":["emergency", "alert", "critical", "error", "warning", "notice", "info", "debug"]}``)
}

// setLogger sets the logger configuration based on the provided parameters.
// It takes in a loger pointer, a config map, and a logtype string.
// The function determines the log adapter, log level, performance settings, and other parameters based on the config map.
// It then sets the logger accordingly using the loger.SetLogger() function.
// If the log adapter is "documentdb", it sets the logger to use the DocumentDB adapter with the specified connection, database, and collection.
// If the log adapter is "console", "file", "multifile", "smtp", or "conn", it sets the logger to use the corresponding adapter with the specified configuration.
// If the log adapter is not recognized, it sets the logger to use the Console adapter as a fallback.
// The function returns nothing.
// Note: This function assumes that the logs package is imported as "logs" and the com package is imported as "com".
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
	performance := false
	performancestr := config["performance"]
	if performancestr != nil && performancestr.(bool) == true {
		performance = true
	}
	performancethrehold := 10

	if config["performancethread"] != nil {
		performancethrehold = com.ConverttoIntwithDefault(config["performancethread"], 10)
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
	case "performance":
		level = 8
	default:
		level = 3
	}
	fullfilename := ""
	// More reasonable defaults for production use:
	// - 100MB max file size (instead of 1GB) for better manageability
	// - 500K lines max (instead of 1M) to keep files searchable
	// Note: maxsize is in KB, will be converted to bytes when passed to logger
	maxlines := 500000
	maxsizeKB := 102400    // 100 MB in KB (100 * 1024)
	maxsizeBytes := 0      // Will be calculated from maxsizeKB
	//	fmt.Println(fmt.Sprintf(`{"level":%d}, %d`, level, logadapter))
	if logadapter == "file" || logadapter == "multifile" {
		adapterconfig := make(map[string]interface{})

		// Try to get adapter-specific config in order of preference:
		// 1. config["adapterconfig"] (new format)
		// 2. config["file"] or config["multifile"] (legacy format matching adapter name)
		// 3. Fall back to entire config
		if config["adapterconfig"] != nil {
			// Safe type assertion with check
			if cfg, ok := config["adapterconfig"].(map[string]interface{}); ok {
				adapterconfig = cfg
			}
		} else if config[logadapter.(string)] != nil {
			// Check for config under adapter name (e.g., config["file"])
			if cfg, ok := config[logadapter.(string)].(map[string]interface{}); ok {
				adapterconfig = cfg
			}
		}

		// If still empty, fall back to entire config
		if len(adapterconfig) == 0 {
			adapterconfig = config
		}

		// Debug output to help diagnose configuration issues
		fmt.Printf("Logger %s config - adapterconfig keys: %v\n", logtype, getMapKeys(adapterconfig))

		// Safe type conversion for filename
		filenameStr := "iac.log"
		if filename := adapterconfig["file"]; filename != nil {
			if fn, ok := filename.(string); ok {
				filenameStr = fn
			}
		}
		suffix := filepath.Ext(filenameStr)
		fileNameOnly := strings.TrimSuffix(filenameStr, suffix)

		// Safe type conversion for folder
		folderStr := ""
		if folder := adapterconfig["folder"]; folder != nil {
			if f, ok := folder.(string); ok {
				folderStr = f
			}
		}
		if folderStr == "" {
			// Use OS-appropriate default folder
			if runtime.GOOS == "windows" {
				folderStr = "C:\\temp"
			} else {
				folderStr = "/tmp"
			}
		}

		// Use cross-platform path handling
		if suffix == "" {
			fullfilename = filepath.Join(folderStr, fmt.Sprintf("%s_%s.log", logtype, fileNameOnly))
		} else {
			fullfilename = filepath.Join(folderStr, fmt.Sprintf("%s_%s%s", logtype, fileNameOnly, suffix))
		}

		// Safe type conversion for maxlines (handles int, float64, string)
		if adapterconfig["maxlines"] != nil {
			maxlines = com.ConverttoIntwithDefault(adapterconfig["maxlines"], maxlines)
			// Validate maxlines: warn if too large
			if maxlines > 5000000 {
				fmt.Printf("Warning: maxlines=%d is very large, consider using a smaller value (recommended: 500000)\n", maxlines)
			}
		}

		// Safe type conversion for maxsize (handles int, float64, string)
		// Note: maxsize in config is in KB, we convert to bytes for the logger
		if adapterconfig["maxsize"] != nil {
			maxsizeKB = com.ConverttoIntwithDefault(adapterconfig["maxsize"], maxsizeKB)
			// Validate maxsize: warn if too large (> 500 MB = 512000 KB)
			if maxsizeKB > 512000 {
				fmt.Printf("Warning: maxsize=%d KB (%.0f MB) is very large, consider using a smaller value (recommended: 102400 KB = 100 MB)\n",
					maxsizeKB, float64(maxsizeKB)/1024)
			}
		}

		// Convert maxsize from KB to bytes for the logger
		maxsizeBytes = maxsizeKB * 1024

		// Debug output showing the resolved log file path
		fmt.Printf("Logger %s will write to: %s (maxlines=%d, maxsize=%d KB / %d bytes)\n",
			logtype, fullfilename, maxlines, maxsizeKB, maxsizeBytes)
	} else if logadapter == "documentdb" {
		conn := "mongodb://localhost:27017"
		db := "IAC_Cache"
		collection := "cache"
		adapterconfig := make(map[string]interface{})

		// Try to get adapter-specific config in order of preference:
		// 1. config["adapterconfig"]
		// 2. config["documentdb"] (legacy format)
		// 3. Fall back to entire config
		if config["adapterconfig"] != nil {
			// Safe type assertion with check
			if cfg, ok := config["adapterconfig"].(map[string]interface{}); ok {
				adapterconfig = cfg
			}
		} else if config["documentdb"] != nil {
			// Check for config under "documentdb" key
			if cfg, ok := config["documentdb"].(map[string]interface{}); ok {
				adapterconfig = cfg
			}
		}

		// If still empty, fall back to entire config
		if len(adapterconfig) == 0 {
			adapterconfig = config
		}

		// For documentdb, check if there's a nested documentdb config
		documentdbcfg := adapterconfig
		if adapterconfig["documentdb"] != nil {
			// Safe type assertion with check
			if cfg, ok := adapterconfig["documentdb"].(map[string]interface{}); ok {
				documentdbcfg = cfg
			}
		}

		// Extract connection parameters
		if documentdbcfg["conn"] != nil {
			if c, ok := documentdbcfg["conn"].(string); ok {
				conn = c
			}
		}
		if documentdbcfg["db"] != nil {
			if d, ok := documentdbcfg["db"].(string); ok {
				db = d
			}
		}
		if documentdbcfg["collection"] != nil {
			if col, ok := documentdbcfg["collection"].(string); ok {
				collection = col
			}
		}

		// Create config using json.Marshal to properly handle special characters
		docdbConfig := map[string]interface{}{
			"level":      level,
			"conn":       conn,
			"db":         db,
			"collection": collection,
			"Perf":       performance,
			"Threhold":   performancethrehold,
		}
		if configJSON, err := json.Marshal(docdbConfig); err == nil {
			loger.SetLogger(logs.AdapterDocumentDB, string(configJSON))
		} else {
			fmt.Printf("Error marshaling documentdb logger config: %v\n", err)
		}
		return
	}

	switch logadapter {
	case "console":
		loger.SetLogger(logs.AdapterConsole, fmt.Sprintf(`{"level":%d, "Perf": "%b", "Threhold": %d}`, level, performance, performancethrehold))
	case "file":
		// Create config using json.Marshal to properly escape paths
		// Note: maxsizeBytes is already in bytes (converted from KB in config)
		fileConfig := map[string]interface{}{
			"level":    level,
			"filename": fullfilename,
			"maxlines": maxlines,
			"maxsize":  maxsizeBytes,
			"Perf":     performance,
			"Threhold": performancethrehold,
		}
		if configJSON, err := json.Marshal(fileConfig); err == nil {
			loger.SetLogger(logs.AdapterFile, string(configJSON))
		} else {
			fmt.Printf("Error marshaling file logger config: %v\n", err)
		}
	case "multifile":
		// Create config using json.Marshal to properly escape paths
		multiFileConfig := map[string]interface{}{
			"filename": fullfilename,
			"level":    level,
			"Perf":     performance,
			"Threhold": performancethrehold,
		}
		if configJSON, err := json.Marshal(multiFileConfig); err == nil {
			loger.SetLogger(logs.AdapterMultiFile, string(configJSON))
		} else {
			fmt.Printf("Error marshaling multifile logger config: %v\n", err)
		}
	case "smtp":
		loger.SetLogger(logs.AdapterMail, fmt.Sprintf(`{"username":"%s","password":"%s","host":"%s","subject":"%s","sendTos":"%s","level":"%d","Perf": "%b", "Threhold": %d}`, config["username"], config["password"], config["host"], config["subject"], config["sendTos"], level, performance, performancethrehold))
	case "conn":
		loger.SetLogger(logs.AdapterConn, fmt.Sprintf(`{"net":"%s","addr":"%s","level":"%d", "Perf": "%b", "Threhold": %d}`, config["net"], config["addr"], level, performance, performancethrehold))
	case "documentdb":
		loger.SetLogger(logs.AdapterDocumentDB, fmt.Sprintf(`{"level":"%d","Perf": "%b", "Threhold": %d}`, level, performance, performancethrehold))
	default:
		loger.SetLogger(logs.AdapterConsole, fmt.Sprintf(`{"level":%d, "Perf": "%b", "Threhold": %d}`, level, performance, performancethrehold))
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

	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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
func (l *Log) ErrorLog(err error) {
	l.Error(fmt.Sprintf("%v", err))
}

func (l *Log) Error(logmsg string) {
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

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

func (l *Log) performance(elapsed time.Duration, logmsg string) {
	logmsg = (fmt.Sprintf("%s %s  %s  %s", l.ClientID, l.User, l.ControllerName, logmsg))

	switch l.ModuleName {
	case Portal:
		PortalLogger.Performance(elapsed, logmsg)
	case API:
		APILogger.Performance(elapsed, logmsg)
	case Database:
		DatabaseLogger.Performance(elapsed, logmsg)
	case TranCode:
		TranCodeLogger.Performance(elapsed, logmsg)
	case Job:
		JobLogger.Performance(elapsed, logmsg)
	case Framework:
		FrameworkLogger.Performance(elapsed, logmsg)
	default:
		Logger.Performance(elapsed, logmsg)
	}
}

func (l *Log) PerformanceWithDuration(function string, elapsed time.Duration) {

	logmsg := fmt.Sprintf(" %s execution elapsed time: %v", function, elapsed)

	l.performance(elapsed, logmsg)

}

func ConvertJson(jobj interface{}) string {
	jsonString, err := json.Marshal(jobj)
	if err != nil {
		fmt.Println("Error marshaling json:", err)
		return ""
	}
	return string(jsonString)
}

// Helper function to get map keys for debugging
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
