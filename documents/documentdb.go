package documents

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
)

var DocDBCon *DocDB

func ConnectDB(DatabaseType string, DatabaseConnection string, DatabaseName string) {
	iLog := logger.Log{ModuleName: logger.Database, User: "System", ControllerName: "DocumentDatabase"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documents.ConnectDB", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("There is error to documents.ConnectDB with error: %s", err))
			return
		}
	}()

	if DatabaseType == "mongodb" {
		var err error
		DocDBCon, err = InitMongoDB(DatabaseConnection, DatabaseName)

		if err != nil {
			iLog.Error(fmt.Sprintf("initialize Documents Database (MongoDB) error: %s", err.Error()))
		}
	}

}
