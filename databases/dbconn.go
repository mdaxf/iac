// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbconn

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

// DB is the interface for database connection
var DB *sql.DB

// ConnectDB is the function to connect to database
// DatabaseType: mysql
// DatabaseConnection: user:password@tcp(localhost:3306)/mydb
// DatabaseName: iac

var (
	/*
		mysql,sqlserver, goracle
	*/
	DatabaseType = "mysql"

	/*
		user:password@tcp(localhost:3306)/mydb
		server=%s;port=%d;user id=%s;password=%s;database=%s
	*/
	//DatabaseConnection = "server=xxx;user id=xx;password=xxx;database=xxx"  //sqlserver
	DatabaseConnection = "user:iacf12345678@tcp(localhost:3306)/iac"
	MaxIdleConns       = 5
	MaxOpenConns       = 10
	once               sync.Once
	err                error
)

// ConnectDB establishes a connection to the database.
// It returns an error if the connection fails.

func ConnectDB() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.ConnectDB", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("ConnectDB defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	// Log the database connection details
	iLog.Info(fmt.Sprintf("Connect Database: %s %s", DatabaseType, DatabaseConnection))

	// Establish the database connection if it hasn't been done before
	once.Do(func() {
		DB, err = sql.Open(DatabaseType, DatabaseConnection)
		if err != nil {
			iLog.Error(fmt.Sprintf("Connect Database Error: %s", err.Error()))
			return
		}
		DB.SetMaxIdleConns(MaxIdleConns)
		DB.SetMaxOpenConns(MaxOpenConns)
	})

	return nil
}

// DBPing pings the database to check if it is still alive.
// It returns an error if the ping fails.

func DBPing() error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.DBPing", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("DBPing defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	return DB.Ping()

}
