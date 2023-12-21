package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/health/checks"
	"github.com/mdaxf/iac/logger"
)

// CheckSystemHealth is a function that checks the health of the system.
// It takes a gin.Context as input and returns a map[string]interface{} and an error.
// The function registers various health checks for different components of the system,
// such as HTTP, MongoDB, MySQL, MQTT, and SignalR.
// It measures the health of the system and returns the result as a JSON-encoded map.

func CheckSystemHealth(c *gin.Context) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "System Status Check"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("framework.health.CheckSystemHealth", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			c.JSON(500, gin.H{"error": r})
		}
	}()

	//	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	ctx := c
	h, _ := New(WithComponent(Component{
		Name:         "IAC Service",
		Instance:     com.Instance,
		InstanceName: com.InstanceName,
		InstanceType: com.InstanceType,
		Version:      "v1.0",
	}), WithChecks(Config{
		Name:      "http",
		Timeout:   time.Second * 5,
		SkipOnErr: true,
		Check: func(context context.Context) error {
			checks.CheckHttpStatus(ctx, "/portal/uipage.html", time.Second*5)
			return nil
		},
	}))

	h.systemInfoEnabled = true

	documentsConfig := config.GlobalConfiguration.DocumentConfig

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

	db := dbconn.DB
	if db != nil {
		h.Register(Config{
			Name: "mysql",
			Check: func(ctx context.Context) error {
				return checks.CheckMySQLStatus(ctx, db, "")
			},
		})
	}
	fmt.Println("MQTT Clients:", config.MQTTClients)
	for key, value := range config.MQTTClients {
		fmt.Println(key)
		client := value
		h.Register(Config{
			Name: "mqtt." + key,
			Check: func(ctx context.Context) error {
				return checks.CheckMqttClientStatus(ctx, client.Client)
			},
		})
	}

	if com.SingalRConfig != nil {
		SAddress := com.SingalRConfig["server"].(string)
		WcAddress := com.SingalRConfig["serverwc"].(string)
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
	/*	h.Register(Config{
		Name: "ping",
		Check: func(ctx context.Context) error {
			return checks.CheckPingStatus(ctx, "", time.Second*5)
			},
		})  */

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
