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

package mysqldb

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	com "github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/logger"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

type DBConn com.DBConn

var once sync.Once

func (db *DBConn) Connect() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database.Connect"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.Connect", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Connect Database defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	// Log the database connection details
	iLog.Info(fmt.Sprintf("Connect Database: %s %s", db.DBType, db.DBConnection))

	// Establish the database connection if it hasn't been done before
	once.Do(func() {
		DB, err := sql.Open(db.DBType, db.DBConnection)
		if err != nil {
			iLog.Error(fmt.Sprintf("Connect Database Error: %s", err.Error()))
			return
		}
		DB.SetMaxIdleConns(db.MaxIdleConns)
		DB.SetMaxOpenConns(db.MaxOpenConns)

		db.DB = DB
	})

	return nil
}

func (db *DBConn) Disconnect() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database.Close"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.Close", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Close Database defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	// Close the database connection
	err := db.DB.Close()
	if err != nil {
		iLog.Error(fmt.Sprintf("Close Database Error: %s", err.Error()))
		return err
	}

	return nil
}

func (db *DBConn) ReConnect() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database.ReConnect"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.ReConnect", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("ReConnect Database defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	// Reconnect to the database
	err := db.Connect()
	if err != nil {
		iLog.Error(fmt.Sprintf("ReConnect Database Error: %s", err.Error()))
		return err
	}

	return nil
}

func (db *DBConn) Ping() error {
	// Function execution logging
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database.Ping"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.Ping", elapsed)
	}()

	// Recover from any panics and log the error
	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("Ping Database defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()

	iLog.Debug(fmt.Sprintf("Database.Ping: %v", db))
	// Ping the database to check if it is still alive
	err := db.DB.Ping()
	if err != nil {
		iLog.Error(fmt.Sprintf("Ping Database Error: %s", err.Error()))
		return err
	}

	return nil
}
