package health

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/health/checks"
	"github.com/mdaxf/iac/logger"
)

// CheckSystemHealth is a function that checks the health of the system.
// It takes a gin.Context as input and returns a map[string]interface{} and an error.
// The function registers various health checks for different components of the system,
// such as HTTP, MongoDB, MySQL, MQTT, and SignalR.
// It measures the health of the system and returns the result as a JSON-encoded map.

func CheckSignalRServerHealth(Node map[string]interface{}, url string, wc string) (map[string]interface{}, error) {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "System Status Check"}

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("framework.health.CheckSystemHealth", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error:", r)
		}
	}()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	h, _ := New(WithComponent(Component{
		Name:         Node["Name"].(string),
		Instance:     Node["AppID"].(string),
		InstanceName: Node["Description"].(string),
		InstanceType: Node["Type"].(string),
		Version:      Node["Version"].(string),
	}))

	SAddress := url
	WcAddress := wc

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

	h.systemInfoEnabled = true

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
