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
	config "github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/framework/cache"
	"github.com/mdaxf/iac/integration/messagebus/nats"
	"github.com/mdaxf/iac/logger"
)

var err error

func initialize() {

	config.SessionCacheTimeout = 1800
	initializecache()
	initializeDatabase()
	nats.MB_NATS_CONN, err = nats.ConnectNATSServer()
	initializeloger()
}

func initializeDatabase() {

	err := dbconn.ConnectDB()
	if err != nil {
		panic(err.Error())
	}
}

func initializecache() {
	var err error

	config.SessionCache, err = cache.NewCache("memory", `{"interval":60}`)
	if err != nil {
		panic(err.Error())
	}
}

func initializeloger() {

	logger.Init()
	logger.Debug("start logger")
}
