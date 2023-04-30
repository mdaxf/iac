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
	"sync"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB
var (
	/*
		mysql,sqlserver, goracle
	*/
	DatabaseType = "sqlserver"

	/*
		user:password@tcp(localhost:3306)/mydb
		server=%s;port=%d;user id=%s;password=%s;database=%s
	*/
	DatabaseConnection = "server=delsrv000039;user id=flxadmin;password=DS@3ds23;database=LPM23"

	once sync.Once
	err  error
)

func ConnectDB() error {
	once.Do(func() {
		DB, err = sql.Open(DatabaseType, DatabaseConnection)
		if err != nil {
			panic(err.Error())
		}
		DB.SetMaxIdleConns(1000)
		DB.SetMaxOpenConns(10000)
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

	return DB.Ping()

}
