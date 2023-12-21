package documents

import (
	"fmt"
	"time"

	"github.com/mdaxf/iac/logger"
)

// DocDB is the interface for document database

var DocDBCon *DocDB

// ConnectDB is the function to connect to document database
// DatabaseType: mongodb
// DatabaseConnection: mongodb://localhost:27017
// DatabaseName: iac

func ConnectDB(DatabaseType string, DatabaseConnection string, DatabaseName string) {
	iLog := logger.Log{ModuleName: logger.Database, User: "System", ControllerName: "DocumentDatabase"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("documents.ConnectDB", elapsed)
	}()
	/*
		defer func() {
			if err := recover(); err != nil {
				iLog.Error(fmt.Sprintf("There is error to documents.ConnectDB with error: %s", err))
				return
			}
		}()
	*/
	if DatabaseType == "mongodb" {
		var err error
		DocDBCon, err = InitMongoDB(DatabaseConnection, DatabaseName)

		if err != nil {
			iLog.Error(fmt.Sprintf("initialize Documents Database (MongoDB) error: %s", err.Error()))
		}
	}

}
