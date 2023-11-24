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

var DB *sql.DB

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

func ConnectDB() error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Database"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("database.ConnectDB", elapsed)
	}()

	defer func() {
		if err := recover(); err != nil {
			iLog.Error(fmt.Sprintf("ConnectDB defer error: %s", err))
			//	ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		}
	}()
	iLog.Info(fmt.Sprintf("Connect Database: %s %s", DatabaseType, DatabaseConnection))

	once.Do(func() {
		DB, err = sql.Open(DatabaseType, DatabaseConnection)
		if err != nil {
			iLog.Error(fmt.Sprintf("Connect Database Error: %s", err.Error()))
			return
		}
		DB.SetMaxIdleConns(MaxIdleConns)
		DB.SetMaxOpenConns(MaxOpenConns)
	})
	/*DB, err = sql.Open(DatabaseType, DatabaseConnection)
	if err != nil {
		panic(err.Error())
	}

	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(10)
	*/
	return nil
}

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
