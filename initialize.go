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

package main

import (
	"fmt"

	config "github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/cache"
	"github.com/mdaxf/iac/integration/messagebus/nats"
	"github.com/mdaxf/iac/logger"
)

var err error
var ilog logger.Log

func initialize() {
	ilog = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Initialization"}
	initializeloger()
	config.SessionCacheTimeout = 1800
	initializecache()
	initializeDatabase()
	nats.MB_NATS_CONN, err = nats.ConnectNATSServer()

	initializedDocuments()

}

func initializeDatabase() {

	ilog.Debug("initialize Database")

	err := dbconn.ConnectDB()
	if err != nil {
		ilog.Error(fmt.Sprintf("initialize Database error: %s", err.Error()))
	}
}

func initializecache() {

	ilog.Debug("initialize Chche")

	config.SessionCache, err = cache.NewCache("memory", `{"interval":60}`)
	if err != nil {
		ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
	}
}

func initializeloger() {
	logger.Init()
	ilog.Debug("initialize logger")
}

func initializedDocuments() {
	ilog.Debug("initialize Documents")
	var DatabaseType = "mongodb"
	var DatabaseConnection = "mongodb://localhost:27017"
	var DatabaseName = "IAC_CFG"
	documents.ConnectDB(DatabaseType, DatabaseConnection, DatabaseName)
}
