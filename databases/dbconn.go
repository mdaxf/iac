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

	//	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/logger"
)

// DB is the interface for database connection
var DB *sql.DB
var monitoring = false
var connectionerr error

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
	DatabaseConnection = "user:iacf12345678@tcp(localhost:3306)/iac?charset=utf8mb4&parseTime=True&loc=Local"
	MaxIdleConns       = 5
	MaxOpenConns       = 10
	once               sync.Once
	err                error
)

// ConnectDB establishes a connection to the database.
// It returns an error if the connection fails.

func ConnectDB() error {
	// Function execution logging
	connectionerr = nil
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
			connectionerr = err.(error)
			return
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	/*
		dbconn := &com.DBConn{
			DBType:       DatabaseType,
			DBConnection: DatabaseConnection,
			DBName:       "",
			MaxIdleConns: MaxIdleConns,
			MaxOpenConns: MaxOpenConns,
		}

		mydbconn := &DBConn{
			DBType:       DatabaseType,
			DBConnection: DatabaseConnection,
			DBName:       "",
			DB:           nil,
			MaxIdleConns: MaxIdleConns,
			MaxOpenConns: MaxOpenConns,
		}

		err := mydbconn.Connect()
		if err != nil {
			iLog.Error(fmt.Sprintf("initialize Database (MySQL) error: %s", err.Error()))
			return err
		}

		DB = mydbconn.DB
		dbconn.DB = DB
		iLog.Debug(fmt.Sprintf("Connect to database:%v", DB))
		com.IACDBConn = dbconn

		return nil */

	// Log the database connection details
	iLog.Info(fmt.Sprintf("Connect Database: %s %s", DatabaseType, DatabaseConnection))

	// Establish the database connection if it hasn't been done before
	once.Do(func() {
		DB, err = sql.Open(DatabaseType, DatabaseConnection)
		if err != nil {
			iLog.Error(fmt.Sprintf("Connect Database Error: %v", err))
			connectionerr = err
			return
		}
		DB.SetMaxIdleConns(MaxIdleConns)
		DB.SetMaxOpenConns(MaxOpenConns)
	})

	if monitoring == false {
		go func() {
			monitorAndReconnectMySQL()
		}()
	}

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

func monitorAndReconnectMySQL() {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database.monitorAndReconnectMySQL"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.monitorAndReconnectMySQL", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("monitorAndReconnectMySQL defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	monitoring = true
	for {
		err := DB.Ping()
		if err != nil {
			iLog.Error(fmt.Sprintf("MySQL connection lost, reconnecting..."))

			ConnectDB()

			if connectionerr != nil {
				iLog.Error(fmt.Sprintf("Failed to reconnect to MySQL:%v", connectionerr))
				time.Sleep(60 * time.Second) // Wait before retrying
				continue
			} else {
				time.Sleep(5 * 60 * time.Second)
				iLog.Debug(fmt.Sprintf("MySQL reconnected successfully"))
				continue
			}
		} else {
			time.Sleep(5 * 60 * time.Second) // Check connection every 60 seconds
			continue
		}
	}

}
