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
	"encoding/json"
	"fmt"
	"io/ioutil"

	config "github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/framework/cache"
	iacmb "github.com/mdaxf/iac/framework/messagebus"
	"github.com/mdaxf/iac/integration/messagebus/nats"
	"github.com/mdaxf/iac/integration/mqttclient"
	"github.com/mdaxf/iac/integration/opcclient"
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
	initializeMqttClient()
	//	initializeOPCClient()
	//	initializeIACMessageBus()
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

func initializeMqttClient() {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ilog.Debug("initialize MQTT Client")

		data, err := ioutil.ReadFile("mqttconfig.json")
		if err != nil {
			ilog.Debug(fmt.Sprintf("failed to read configuration file: %v", err))

		}
		ilog.Debug(fmt.Sprintf("MQTT Clients configuration file: %s", string(data)))
		var mqttconfig mqttclient.MqttConfig
		err = json.Unmarshal(data, &mqttconfig)
		if err != nil {
			ilog.Debug(fmt.Sprintf("failed to unmarshal the configuration file: %v", err))

		}
		ilog.Debug(fmt.Sprintf("MQTT Clients configuration: %v", logger.ConvertJson(mqttconfig)))

		for _, mqttcfg := range mqttconfig.Mqtts {
			ilog.Debug(fmt.Sprintf("MQTT Client configuration: %s", logger.ConvertJson(mqttcfg)))
			mqttclient.NewMqttClient(mqttcfg)

		}

	}()

}

func initializeOPCClient() {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ilog.Debug("initialize OPC Client")
		data, err := ioutil.ReadFile("opcuaclient.json")
		if err != nil {
			ilog.Debug(fmt.Sprintf("failed to read configuration file: %v", err))

		}
		ilog.Debug(fmt.Sprintf("OPC UA Clients configuration file: %s", string(data)))

		var config opcclient.OPCConfig

		err = json.Unmarshal(data, &config)

		if err != nil {
			ilog.Debug(fmt.Sprintf("failed to unmarshal the configuration file: %v", err))
		}
		for _, opcuaclient := range config.OPCClients {
			ilog.Debug(fmt.Sprintf("OPC UA Client configuration: %s", logger.ConvertJson(opcuaclient)))
			opcclient.Initialize(opcuaclient)
		}
	}()
}

func initializeIACMessageBus() {
	wg.Add(1)
	go func() {
		ilog.Debug("initialize IAC Message Bus")
		defer wg.Done()

		iacmb.Initialize()

		/*	iacmb.Initialize(8888, "IAC")

			ilog.Debug(fmt.Sprintf("IAC Message bus: %v", iacmb.IACMB))

			iacmb.IACMB.Channel.OnRead(func(data string) {
				ilog.Debug(fmt.Sprintf("IAC Message bus channel read: %s", data))
			})

			iacmb.IACMB.Channel.Write("Start the Message bus channel IAC")

			data := map[string]interface{}{"name": "IAC", "type": "messagebus", "port": 8888, "channel": "IAC"}
			iacmb.IACMB.Channel.Write(logger.ConvertJson(data))
			iacmb.IACMB.Channel.Write("Start the Message bus channel IAC") */
	}()
}
