package health

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"

	"github.com/mdaxf/iac/health/checks"
)

func CheckSystemHealth(c *gin.Context) (map[string]interface{}, error) {

	defer func() {
		if r := recover(); r != nil {
			c.JSON(500, gin.H{"error": r})
		}
	}()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

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
			checks.CheckHttpStatus(ctx, "/health", time.Second*5)
			return nil
		},
	}))

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
