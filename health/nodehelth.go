package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"database/sql"

	"github.com/mdaxf/signalrsrv/signalr"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mdaxf/iac/com"

	"github.com/mdaxf/iac/health/checks"
	"github.com/mdaxf/iac/logger"
)

// CheckSystemHealth is a function that checks the health of the system.
// It takes a gin.Context as input and returns a map[string]interface{} and an error.
// The function registers various health checks for different components of the system,
// such as HTTP, MongoDB, MySQL, MQTT, and SignalR.
// It measures the health of the system and returns the result as a JSON-encoded map.

func CheckNodeHealth(Node map[string]interface{}, db *sql.DB, mongoClient *mongo.Client, signalRClient signalr.Client) (map[string]interface{}, error) {
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
		Name:         Node["Name"].(string),
		Instance:     Node["AppID"].(string),
		InstanceName: Node["Description"].(string),
		InstanceType: Node["Type"].(string),
		Version:      Node["Version"].(string),
	}))

	h.systemInfoEnabled = true

	if mongoClient != nil {
		h.Register(Config{
			Name: "mongoDB",
			Check: func(ctx context.Context) error {
				return checks.CheckMongoClientStatus(ctx, mongoClient)
			},
		})
	}
	if db != nil {
		h.Register(Config{
			Name: "mysql",
			Check: func(ctx context.Context) error {
				return checks.CheckMySQLStatus(ctx, db, "")
			},
		})
	}
	if signalRClient != nil {
		h.Register(Config{
			Name: "signalR",
			Check: func(ctx context.Context) error {
				return checks.CheckSignalClientStatus(ctx, signalRClient)
			},
		})
	}

	m := h.Measure(context.Background())
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
