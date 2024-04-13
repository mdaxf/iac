package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"database/sql"

	"github.com/mdaxf/iac/com"

	"github.com/mdaxf/iac/health/checks"
	"github.com/mdaxf/iac/logger"
)

// CheckSystemHealth is a function that checks the health of the system.
// It takes a gin.Context as input and returns a map[string]interface{} and an error.
// The function registers various health checks for different components of the system,
// such as HTTP, MongoDB, MySQL, MQTT, and SignalR.
// It measures the health of the system and returns the result as a JSON-encoded map.

func CheckNodeHealth(Node map[string]interface{}, db *sql.DB, DocDBConn, documentsConfig map[string]interface{}, signalrcfg map[string]interface{} ) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Node Status Check"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("framework.health.CheckNodeHealth", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()


	h, _ := New(WithComponent(Component{
		Name:      Node["Name"].(string),
		AppID:     Node["AppID"].(string),
		Description: Node["Description"].(string),
	}))

	h.systemInfoEnabled = true

	if documentsConfig != nil {

		var DatabaseType = documentsConfig["type"].(string)             // "mongodb"
		var DatabaseConnection = documentsConfig["connection"].(string) //"mongodb://localhost:27017"
		var DatabaseName = documentsConfig["database"].(string)         //"IAC_CFG"

		if DatabaseType == "mongodb" && DatabaseConnection != "" && DatabaseName != "" {
			h.Register(Config{
				Name: "mongodb",
				Check: func(ctx context.Context) error {
					return checks.CheckMongoDBStatus(ctx, DatabaseConnection, time.Second*5, time.Second*5, time.Second*5)
				},
			})
		}
	}

	if db != nil {
		h.Register(Config{
			Name: "mysql",
			Check: func(ctx context.Context) error {
				return checks.CheckMySQLStatus(ctx, db, "")
			},
		})
	}

	if signalrcfg != nil {
		SAddress := signalrcfg["server"].(string)
		WcAddress := signalrcfg["serverwc"].(string)
		fmt.Println("SignalR Address:", SAddress)
		fmt.Println("SignalR Websocket Address:", WcAddress)

		if WcAddress != "" {

			h.Register(Config{
				Name: "signalr Websocket Server",
				Check: func(ctx context.Context) error {
					return checks.CheckSignalRSrvStatus(ctx, SAddress, WcAddress)
				},
			})
		}
		if SAddress != "" {
			h.Register(Config{
				Name: "signalr Http Server",
				Check: func(ctx context.Context) error {
					return checks.CheckSignalRSrvHttpStatus(ctx, SAddress, WcAddress)
				},
			})
		}
	}

	m := h.Measure(ctx)
	data, err := json.Marshal(m)
	if err != nil {
		return make(map[string]interface{}), err
	}

	jdata, err := com.ConvertbytesToMap(data)

	if err != nil {
		return make(map[string]interface{}), err
	}

	return jdata, nil
}
