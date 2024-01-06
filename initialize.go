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
	"time"

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

	"github.com/mdaxf/iac/engine/trancode"
	"github.com/mdaxf/iac/framework/callback_mgr"
)

var err error
var ilog logger.Log
var Initialized bool

// initialize is a function that performs the initialization process of the application.
// It sets up the logger, initializes the cache, database, documents, MQTT client, and IAC message bus.
// It also connects to the IAC Message Bus using SignalR.
// The function measures the performance duration and logs any errors that occur during the initialization process.
// Finally, it sets the Initialized flag to true.

func initialize() {
	ilog = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "Initialization"}
	startTime := time.Now()
	fmt.Printf("initialize starttime: %v", startTime)
	defer func() {
		fmt.Printf("initialize defer time: %v", time.Now())
		elapsed := time.Since(startTime)
		if Initialized {
			ilog.PerformanceWithDuration("main.initialize", elapsed)
		} else {
			fmt.Printf("initialize defer time: %v, duration: %v", time.Now(), elapsed)
		}
	}()

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
		fmt.Printf("IAC Message Bus: %v", com.IACMessageBusClient)
		ilog.Debug(fmt.Sprintf("IAC Message Bus: %v", com.IACMessageBusClient))
	}()

	go func() {

		ilog.Debug("Register the trancode execution interface")
		tfr := trancode.TranFlowstr{}
		callback_mgr.RegisterCallBack("TranCode_Execute", tfr.Execute)

	}()

	fmt.Printf("initialize end time: %v", time.Now())
	Initialized = true
}

// initializeDatabase initializes the database connection based on the configuration provided in the global configuration.
// It sets the database type, connection string, maximum idle connections, and maximum open connections.
// If any required configuration is missing or if there is an error connecting to the database, an error is logged.
func initializeDatabase() {
	// function execution start time
	startTime := time.Now()

	// defer function to log the performance duration of initializeDatabase
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.initializeDatabase", elapsed)
	}()

	ilog.Debug("initialize Database")
	databaseconfig := config.GlobalConfiguration.DatabaseConfig

	// check if database type is missing
	if databaseconfig["type"] == nil {
		ilog.Error(fmt.Sprintf("initialize Database error: %s", "DatabaseType is missing"))
		return
	}

	// check if database connection is missing
	if databaseconfig["connection"] == nil {
		ilog.Error(fmt.Sprintf("initialize Database error: %s", "DatabaseConnection is missing"))
		return
	}

	// set the database type and connection string
	dbconn.DatabaseType = databaseconfig["type"].(string)
	dbconn.DatabaseConnection = databaseconfig["connection"].(string)

	// set the maximum idle connections, default to 5 if not provided or if the value is not a float64
	if databaseconfig["maxidleconns"] == nil {
		dbconn.MaxIdleConns = 5
	} else {
		if v, ok := databaseconfig["maxidleconns"].(float64); ok {
			dbconn.MaxIdleConns = int(v)
		} else {
			dbconn.MaxIdleConns = 5
		}
	}

	// set the maximum open connections, default to 10 if not provided or if the value is not a float64
	if databaseconfig["maxopenconns"] == nil {
		dbconn.MaxOpenConns = 10
	} else {
		if v, ok := databaseconfig["maxopenconns"].(float64); ok {
			dbconn.MaxOpenConns = int(v)
		} else {
			dbconn.MaxOpenConns = 10
		}
	}

	// connect to the database
	err := dbconn.ConnectDB()
	if err != nil {
		ilog.Error(fmt.Sprintf("initialize Database error: %s", err.Error()))
	}
}

// initializecache initializes the cache based on the configuration provided in the global configuration.
// It checks the cache adapter and interval in the configuration and sets the session cache timeout accordingly.
// Depending on the cache adapter, it creates a new cache instance with the corresponding configuration.
// If the cache adapter is not recognized, it falls back to the default cache adapter with the session cache timeout.
// If there is an error initializing the cache, an error is logged.
func initializecache() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.initializecache", elapsed)
	}()
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
		config.SessionCache, err = cache.NewCache(cacheConfig["adapter"].(string), fmt.Sprintf(`{"interval":%v}`, interval))
		if err != nil {
			ilog.Error(fmt.Sprintf("initialize cache error: %s", err.Error()))
		}
	}

}

// initializeloger initializes the logger based on the global configuration.
// It checks if the log configuration is missing and prints the log configuration.
// Then, it calls the logger.Init function with the log configuration.
func initializeloger() error {
	if config.GlobalConfiguration.LogConfig == nil {
		return fmt.Errorf("log configuration is missing")
	}
	fmt.Printf("log configuration: %v", config.GlobalConfiguration.LogConfig)
	logger.Init(config.GlobalConfiguration.LogConfig)
	return nil
}

// initializedDocuments initializes the documents for the application.
// It retrieves the document configuration from the global configuration and connects to the specified database.
// If any required configuration is missing, it logs an error and returns.

func initializedDocuments() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.initializedDocuments", elapsed)
	}()

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

// initializeMqttClient initializes the MQTT clients by reading the configuration file "mqttconfig.json" and creating MqttClient instances based on the configuration.
// It populates the config.MQTTClients map with the created MqttClient instances.
// The function also logs debug information about the configuration and created MQTT clients.
// It measures the performance duration of the function using the ilog.PerformanceWithDuration function.

func initializeMqttClient() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.initializeMqttClient", elapsed)
	}()

	config.MQTTClients = make(map[string]*mqttclient.MqttClient)

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
		i := 1
		for _, mqttcfg := range mqttconfig.Mqtts {
			ilog.Debug(fmt.Sprintf("MQTT Client configuration: %s", logger.ConvertJson(mqttcfg)))
			mqtc := mqttclient.NewMqttClient(mqttcfg)
			//	fmt.Println("MQTT Client: %v", mqtc)
			//	ilog.Debug(fmt.Sprintf("MQTT Client: %v", mqtc))
			config.MQTTClients[fmt.Sprintf("mqttclient_%d", i)] = mqtc
			mqtc.Initialize_mqttClient()
			//	fmt.Sprintln("MQTT Client: %v", config.MQTTClients)
			//	ilog.Debug(fmt.Sprintf("MQTT Client: %v", config.MQTTClients))
			i++
		}

		fmt.Println("MQTT Clients: %v, %d", config.MQTTClients, i)
		ilog.Debug(fmt.Sprintf("MQTT Clients: %v, %d", config.MQTTClients, i))
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

// initializeIACMessageBus initializes the IAC message bus.
// It starts a goroutine to handle the initialization process.
// The function measures the performance duration and logs it using ilog.PerformanceWithDuration.
// The function also logs debug information about the IAC message bus configuration and channel.
func initializeIACMessageBus() {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.initializeIACMessageBus", elapsed)
	}()

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
