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
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"plugin"
	"reflect"

	"github.com/gin-gonic/gin"
	configuration "github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"
	mongodb "github.com/mdaxf/iac/documents"

	//	kitlog "github.com/go-kit/log"
	"github.com/mdaxf/iac/com"
	//	msgbus "github.com/mdaxf/integration/signalr"
	//	"github.com/philippseith/signalr"
)

var wg sync.WaitGroup
var router *gin.Engine

func main() {
	Initialized = false
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		if Initialized {
			ilog.PerformanceWithDuration("main.initializeIACMessageBus", elapsed)
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			if Initialized {
				ilog.Error(fmt.Sprintf("Panic: %v", r))
			} else {
				log.Fatalf("Panic: %v", r)
			}
		}
	}()
	// Load configuration from the file

	config, err := configuration.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
		//	ilog.Error("Failed to load configuration: %v", err)
	}

	configuration.GlobalConfiguration, err = configuration.LoadGlobalConfig()

	if err != nil {
		log.Fatalf("Failed to load global configuration: %v", err)
		//	ilog.Error("Failed to load global configuration: %v", err)
	}

	initialize()

	if dbconn.DB != nil {
		defer dbconn.DB.Close()
	} else {
		//log.Fatalf("Failed to connect to database")
		ilog.Error("Failed to connect to database")
	}

	if mongodb.DocDBCon.MongoDBClient != nil {
		defer mongodb.DocDBCon.MongoDBClient.Disconnect(context.Background())
	} else {
		//log.Fatalf("Failed to connect to database")
		ilog.Error("Failed to connect to document database")
	}
	// Initialize the Gin router

	for _, dbclient := range com.MongoDBClients {
		if dbclient != nil {
			defer dbclient.Disconnect(context.Background())
		} else {
			//log.Fatalf("Failed to connect to database")
			ilog.Error("Failed to connect to the configured document database")
		}

	}

	if com.IACMessageBusClient != nil {
		defer com.IACMessageBusClient.Stop()
	} else {
		//log.Fatalf("Failed to connect to database")
		ilog.Error("Failed to connect to the configured message bus")
	}
	portal := config.Portal

	router = gin.Default()

	//router.Static("/portal", portal.Path)
	//router.StaticFile("/portal", portal.Home)
	//	router.StaticFile("/portal", portal.Logon)

	if configuration.GlobalConfiguration.WebServerConfig != nil {
		webserverconfig := configuration.GlobalConfiguration.WebServerConfig
		ilog.Debug(fmt.Sprintf("Webserver cconfig: %v", webserverconfig))

		paths := webserverconfig["paths"].(map[string]interface{})

		for key, value := range paths {
			ilog.Debug(fmt.Sprintf("Webserver path: %s configuration: %v", key, value))
			pathstr := value.(map[string]interface{})
			path := pathstr["path"].(string)
			home := pathstr["home"].(string)

			if path != "" {
				router.Static(fmt.Sprintf("/%s", key), path)
				ilog.Debug(fmt.Sprintf("Webserver path: /%s with %s", key, path))
			} else {
				ilog.Error(fmt.Sprintf("there is error in configuration %s, path cannot be empty!", key))
			}
			if home != "" {
				router.StaticFile(fmt.Sprintf("/%s", key), home)
			}
		}

		proxy := webserverconfig["proxy"].(map[string]interface{})

		if proxy != nil {
			go renderproxy(proxy, router)
		}

		headers := webserverconfig["headers"].(map[string]interface{})
		router.Use(GinMiddleware(headers))

	}

	// Load controllers dynamically based on the configuration file
	plugincontrollers := make(map[string]interface{})
	for _, controllerConfig := range config.PluginControllers {

		jsonString, err := json.Marshal(controllerConfig)
		if err != nil {

			ilog.Error(fmt.Sprintf("Error marshaling json: %v", err))
			return
		}
		fmt.Println(string(jsonString))
		controllerModule, err := loadpluginControllerModule(controllerConfig.Path)
		if err != nil {
			ilog.Error(fmt.Sprintf("Failed to load controller module %s: %v", controllerConfig.Path, err))
		}
		plugincontrollers[controllerConfig.Path] = controllerModule
	}

	go func() {
		// Create endpoints dynamically based on the configuration file
		for _, controllerConfig := range config.PluginControllers {
			for _, endpointConfig := range controllerConfig.Endpoints {
				method := endpointConfig.Method
				path := fmt.Sprintf("/%s%s", controllerConfig.Path, endpointConfig.Path)
				handler := plugincontrollers[controllerConfig.Path].(map[string]interface{})[endpointConfig.Handler].(func(*gin.Context))
				router.Handle(method, path, handler)
			}
		}
	}()
	// Load controllers statically based on the configuration file
	ilog.Info("Loading controllers")

	go func() {
		loadControllers(router, config.Controllers)
	}()
	// Start the portals
	ilog.Info("Starting portals")

	jsonString, err := json.Marshal(config.Portal)
	if err != nil {
		ilog.Error(fmt.Sprintf("Error marshaling json: %v", err))
		return
	}
	fmt.Println(string(jsonString))

	ilog.Info(fmt.Sprintf("Starting portal on port %d, page:%s, logon: %s", portal.Port, portal.Home, portal.Logon))

	clientconfig := make(map[string]interface{})
	clientconfig["signalrconfig"] = com.SingalRConfig
	clientconfig["instance"] = com.Instance
	clientconfig["instanceType"] = com.InstanceType
	clientconfig["instanceName"] = com.InstanceName

	router.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, clientconfig)
	})

	router.GET("/debug", func(c *gin.Context) {
		headers := c.Request.Header
		useragent := c.Request.Header.Get("User-Agent")
		ilog.Debug(fmt.Sprintf("User-Agent: %s, headers: %v", useragent, headers))
		debugInfo := map[string]interface{}{
			"Route":          c.FullPath(),
			"requestheader":  headers,
			"User-Agent":     useragent,
			"requestbody":    c.Request.Body,
			"responseheader": c.Writer.Header(),
			"Method":         c.Request.Method,
		}

		c.JSON(http.StatusOK, debugInfo)
	})
	/*
		router.Use(static.Serve("/portal", static.LocalFile("./portal", true)))
		router.Use(static.Serve("/portal/scripts", static.LocalFile("./portal/scripts", true)))*/
	/*

	 */
	// Start the server
	//go router.Run(fmt.Sprintf(":%d", config.Port))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port), // Set your desired port
		Handler:      router,
		ReadTimeout:  time.Duration(config.Timeout) * time.Millisecond,   // Set read timeout
		WriteTimeout: time.Duration(2*config.Timeout) * time.Millisecond, // Set write timeout
		IdleTimeout:  time.Duration(3*config.Timeout) * time.Millisecond, // Set idle timeout
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ilog.Error(fmt.Sprintf("Failed to start server: %v", err))
			panic(err)
		}
	}()

	ilog.Info(fmt.Sprintf("Started portal on port %d, page:%s, logon: %s", portal.Port, portal.Home, portal.Logon))

	elapsed := time.Since(startTime)
	ilog.PerformanceWithDuration("main.main", elapsed)

	wg.Wait()
}

func loadpluginControllerModule(controllerPath string) (interface{}, error) {

	ilog.Info(fmt.Sprintf("Loading plugin controllers %s", controllerPath))

	modulePath := fmt.Sprintf("./plugins/%s/%s.so", controllerPath, controllerPath)
	print(modulePath)
	module, err := plugin.Open(modulePath)
	if err != nil {

		return nil, fmt.Errorf("Failed to open controller module %s: %v", modulePath, err)
	}
	sym, err := module.Lookup(controllerPath + "Controller")
	if err != nil {
		return nil, fmt.Errorf("Failed to lookup symbol in controller module %s: %v", modulePath, err)
	}
	return sym, nil
}

func getpluginHandlerFunc(module reflect.Value, name string) gin.HandlerFunc {

	ilog.Info(fmt.Sprintf("Loading plugin handler function: %s", name))

	method := module.MethodByName(name)
	if !method.IsValid() {
		return nil
	}

	return func(c *gin.Context) {
		in := make([]reflect.Value, 1)
		in[0] = reflect.ValueOf(c)
		out := method.Call(in)
		if len(out) > 0 {
			if err, ok := out[0].Interface().(error); ok {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			if data, ok := out[0].Interface().([]byte); ok {
				c.Data(http.StatusOK, "application/json", data)
				return
			}
		}
		c.Status(http.StatusOK)
	}
}
func CORSMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		//  c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Origin")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		//	ilog.Debug(fmt.Sprintf("CORSMiddleware: %s", allowOrigin))
		//	ilog.Debug(fmt.Sprintf("CORSMiddleware header: %s", c.Request.Header))
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func GinMiddleware(headers map[string]interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {

		//	ilog.Debug(fmt.Sprintf("GinMiddleware: %v", headers))
		//	ilog.Debug(fmt.Sprintf("GinMiddleware header: %s", c.Request.Header))

		for key, value := range headers {
			c.Header(key, value.(string))
			//	c.Writer.Header().Set(key, value.(string))
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func renderproxy(proxy map[string]interface{}, router *gin.Engine) {
	ilog.Debug(fmt.Sprintf("renderproxy: %v", proxy))

	for key, value := range proxy {
		ilog.Debug(fmt.Sprintf("renderproxy key: %s, value: %s", key, value))

		nextURL, _ := url.Parse((value).(string))
		ilog.Debug(fmt.Sprintf("renderproxy nextURL: %v", nextURL))

		router.Any(fmt.Sprintf("/%s/*path", key), func(c *gin.Context) {

			ilog.Debug(fmt.Sprintf("renderproxy path: %s, target: %s", c.Request.URL.Path, nextURL))

			proxy := httputil.NewSingleHostReverseProxy(nextURL)

			// Update the headers to allow for SSL redirection
			//	req := c.Request
			//	req.URL.Host = nextURL.Host
			//	req.URL.Scheme = nextURL.Scheme
			//req.Header["X-Forwarded-Host"] = req.Header["Host"]

			c.Request.URL.Path = c.Param("path")

			ilog.Debug(fmt.Sprintf("request: %v", c.Request))
			// Note that ServeHttp is non blocking and uses a go routine under the hood
			proxy.ServeHTTP(c.Writer, c.Request)

		})

	}
}
