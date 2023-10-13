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

	//	iacmb "github.com/mdaxf/iac/framework/messagebus"

	//	"github.com/mdaxf/iac/integration/messagebus/nats"
	"github.com/mdaxf/iac/integration/mqttclient"

	//	"github.com/mdaxf/iac/integration/opcclient"
	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/integration/signalr"

	"github.com/mdaxf/iac/logger"
)

var err error
var ilog logger.Log

func initialize() {

	ilog = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Initialization"}
	initializeloger()
	ilog.Debug("initialize logger")
	config.SessionCacheTimeout = 1800
	initializecache()
	initializeDatabase()
	//	nats.MB_NATS_CONN, err = nats.ConnectNATSServer()

	initializedDocuments()
	initializeMqttClient()
	//	initializeOPCClient()
	//	initializeIACMessageBus()
	wg.Add(1)
	go func() {
		defer wg.Done()
		com.IACMessageBusClient, err = signalr.Connect(com.SingalRConfig)
		if err != nil {
			//	fmt.Errorf("Failed to connect to IAC Message Bus: %v", err)
			ilog.Error(fmt.Sprintf("Failed to connect to IAC Message Bus: %v", err))
		}
		fmt.Println("IAC Message Bus: ", com.IACMessageBusClient)
		ilog.Debug(fmt.Sprintf("IAC Message Bus: %v", com.IACMessageBusClient))
	}()
}

func initializeDatabase() {

	ilog.Debug("initialize Database")
	databaseconfig := config.GlobalConfiguration.DatabaseConfig

	if databaseconfig["type"] == nil {
		ilog.Error(fmt.Sprintf("initialize Database error: %s", "DatabaseType is missing"))
		return
	}
	if databaseconfig["connection"] == nil {
		ilog.Error(fmt.Sprintf("initialize Database error: %s", "DatabaseConnection is missing"))
		return
	}

	dbconn.DatabaseType = databaseconfig["type"].(string)
	dbconn.DatabaseConnection = databaseconfig["connection"].(string)

	err := dbconn.ConnectDB()
	if err != nil {
		ilog.Error(fmt.Sprintf("initialize Database error: %s", err.Error()))
	}
}

func initializecache() {
	/*
		type cacheconfigstruct struct {
			Adapter  string
			Interval int
		}
		var cache cacheconfigstruct

		err := json.Unmarshal(cacheConfig, &cache)
		if err != nil {
			ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
		} */
	ilog.Debug("initialize Chche")

	cacheConfig := config.GlobalConfiguration.CacheConfig

	ilog.Debug(fmt.Sprintf("initialize cache, %v", cacheConfig))

	if cacheConfig["adapter"] == nil {
		ilog.Error(fmt.Sprintf("initialize cache error: %s", "CacheType is missing"))
		cacheConfig["adapter"] = "memory"
	}
	if cacheConfig["interval"] == nil {
		ilog.Error(fmt.Sprintf("initialize cache error: %s", "CacheTTL is missing"))
		cacheConfig["interval"] = 3600
	}

	if v, ok := cacheConfig["interval"].(float64); ok {

		config.SessionCacheTimeout = v
	} else {
		config.SessionCacheTimeout = 3600
	}

	//	fmt.Printf("CacheType: %s, CacheTTL: %d", cacheConfig["adapter"], config.SessionCacheTimeout)
	ilog.Debug(fmt.Sprintf("initialize cache with the configuration, %v", cacheConfig))

	switch cacheConfig["adapter"] {
	case "memcache":
		conn := "127.0.0.1:11211"
		if cacheConfig["memcache"] != nil {
			memcachecfg := cacheConfig["memcache"].(map[string]interface{})
			if memcachecfg["conn"] != nil {
				conn = memcachecfg["conn"].(string)
			}
		}
		config.SessionCache, err = cache.NewCache(cacheConfig["adapter"].(string), fmt.Sprintf(`{"conn":"%s"}`, conn))
		if err != nil {
			ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
		}
	case "redis":
		key := "IAC_Cache"
		conn := "6379"
		dbNum := 0
		password := ""
		if cacheConfig["redis"] != nil {
			rediscfg := cacheConfig["redis"].(map[string]interface{})
			if rediscfg["key"] != nil {
				key = rediscfg["key"].(string)
			}
			if rediscfg["conn"] != nil {
				conn = rediscfg["conn"].(string)
			}
			if rediscfg["dbNum"] != nil {
				dbNum = int(rediscfg["dbNum"].(float64))
			}
			if rediscfg["password"] != nil {
				password = rediscfg["password"].(string)
			}
		}
		config.SessionCache, err = cache.NewCache(cacheConfig["adapter"].(string), fmt.Sprintf(`{"key":"%s","conn":"%s","dbNum":%d,"password":"%s"}`, key, conn, dbNum, password))
		if err != nil {
			ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
		}
	case "file":
		CachePath := ""
		DirectoryLevel := 2
		FileSuffix := ".cache"
		EmbedExpiry := 0
		if cacheConfig["file"] != nil {
			filecfg := cacheConfig["file"].(map[string]interface{})

			if filecfg["CachePath"] != nil {
				CachePath = filecfg["CachePath"].(string)
			}

			if filecfg["DirectoryLevel"] != nil {
				DirectoryLevel = int(filecfg["DirectoryLevel"].(float64))
			}

			if filecfg["FileSuffix"] != nil {
				FileSuffix = filecfg["FileSuffix"].(string)
			}

			if filecfg["EmbedExpiry"] != nil {
				EmbedExpiry = int(filecfg["EmbedExpiry"].(float64))
			}
		}

		config.SessionCache, err = cache.NewCache(cacheConfig["adapter"].(string), fmt.Sprintf(`{"CachePath":"%s","FileSuffix":"%s","DirectoryLevel":"%d","EmbedExpiry":"%d"}`, CachePath, FileSuffix, DirectoryLevel, EmbedExpiry))
		//	ilog.Debug(fmt.Sprintf("initialize cache with the configuration, %v", cacheConfig))

		if err != nil {
			ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
		}
	case "documentdb":
		conn := "mongodb://localhost:27017"
		db := "IAC_Cache"
		collection := "cache"
		if cacheConfig["documentdb"] != nil {
			documentdbcfg := cacheConfig["documentdb"].(map[string]interface{})
			if documentdbcfg["conn"] != nil {
				conn = documentdbcfg["conn"].(string)
			}
			if documentdbcfg["db"] != nil {
				db = documentdbcfg["db"].(string)
			}
			if documentdbcfg["collection"] != nil {
				collection = documentdbcfg["collection"].(string)
			}
		}
		//cachedb.Initalize()
		config.SessionCache, err = cache.NewCache(cacheConfig["adapter"].(string), fmt.Sprintf(`{"conn":"%s","db":"%s","collection":"%s"}`, conn, db, collection))
		if err != nil {
			ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
		}

	default:
		interval := config.SessionCacheTimeout
		/*
			if cacheConfig["memory"] != nil {
				memorycfg := cacheConfig["memory"].(map[string]interface{})
				if memorycfg["interval"] != nil {
					interval = int(memorycfg["interval"].(float64))
				}
			}  */
		config.SessionCache, err = cache.NewCache(cacheConfig["adapter"].(string), fmt.Sprintf(`{"interval":"%d"}`, interval))
		//config.SessionCache, err = cache.NewCache("memory", `{"interval":60}`)
		if err != nil {
			ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
		}
	}

}

func initializeloger() {
	if config.GlobalConfiguration.LogConfig == nil {
		fmt.Errorf("log configuration is missing")
	}
	fmt.Println("log configuration", config.GlobalConfiguration.LogConfig)
	logger.Init(config.GlobalConfiguration.LogConfig)

}

func initializedDocuments() {

	if config.GlobalConfiguration.DocumentConfig == nil {
		fmt.Errorf("documentdb configuration is missing")
		return
	}

	documentsConfig := config.GlobalConfiguration.DocumentConfig

	ilog.Debug(fmt.Sprintf("initialize Documents, %v", documentsConfig))

	var DatabaseType = documentsConfig["type"].(string)             // "mongodb"
	var DatabaseConnection = documentsConfig["connection"].(string) //"mongodb://localhost:27017"
	var DatabaseName = documentsConfig["database"].(string)         //"IAC_CFG"

	if DatabaseType == "" {
		ilog.Error(fmt.Sprintf("initialize Documents error: %s", "DatabaseType is missing"))
		return
	}
	if DatabaseConnection == "" {
		ilog.Error(fmt.Sprintf("initialize Documents error: %s", "DatabaseConnection is missing"))
		return
	}
	if DatabaseName == "" {
		ilog.Error(fmt.Sprintf("initialize Documents error: %s", "DatabaseName is missing"))
		return
	}

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

/*
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
} */

func initializeIACMessageBus() {
	wg.Add(1)
	go func() {
		ilog.Debug("initialize IAC Message Bus")
		defer wg.Done()

		//	iacmb.Initialize()

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
