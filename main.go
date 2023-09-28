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

	"log"
	"net/http"

	"plugin"
	"reflect"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	configuration "github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"
	mongodb "github.com/mdaxf/iac/documents"
)

var wg sync.WaitGroup
var router *gin.Engine

func main() {
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

	defer dbconn.DB.Close()
	defer mongodb.DocDBCon.MongoDBClient.Disconnect(context.Background())
	// Initialize the Gin router

	//defer config.SessionCache.MongoDBClient.Disconnect(context.Background())

	router = gin.Default()

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

	// Create endpoints dynamically based on the configuration file
	for _, controllerConfig := range config.PluginControllers {
		for _, endpointConfig := range controllerConfig.Endpoints {
			method := endpointConfig.Method
			path := fmt.Sprintf("/%s%s", controllerConfig.Path, endpointConfig.Path)
			handler := plugincontrollers[controllerConfig.Path].(map[string]interface{})[endpointConfig.Handler].(func(*gin.Context))
			router.Handle(method, path, handler)
		}
	}

	// Load controllers statically based on the configuration file
	ilog.Info("Loading controllers")
	loadControllers(router, config.Controllers)

	// Start the portals
	ilog.Info("Starting portals")

	jsonString, err := json.Marshal(config.Portal)
	if err != nil {
		ilog.Error(fmt.Sprintf("Error marshaling json: %v", err))
		return
	}
	fmt.Println(string(jsonString))

	portal := config.Portal

	ilog.Info(fmt.Sprintf("Starting portal on port %d, page:%s, logon: %s", portal.Port, portal.Home, portal.Logon))

	router.Use(static.Serve("/portal", static.LocalFile("./portal", true)))
	router.Use(static.Serve("/portal/scripts", static.LocalFile("./portal/scripts", true)))
	router.LoadHTMLGlob("portal/Scripts/UIForm.js")
	router.LoadHTMLGlob("portal/Scripts/UIFramework.js")

	/*router.Static("/portal", "./portal")
	router.LoadHTMLGlob("portal/*.html")
	router.LoadHTMLGlob("portal/Scripts/UIFramework.js") */
	/*
		corsconfig := cors.DefaultConfig()
		corsconfig.AllowAllOrigins = true
		//corsconfig.AllowOrigins = []string{"http://localhost:8888"} // Replace with your origin
		corsconfig.AllowedMethods = []string{"GET", "POST", "PUT"}
		corsconfig.AllowedHeaders = []string{"Content-Type", "Authorization"}
		corsconfig.AllowCredentials = true
		router.Use(cors.New(corsconfig))  */
	/*
		router.Use(func(c *gin.Context) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "127.0.0.1:8888") // Replace "*" with the specific origin you want to allow
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(200)
				return
			}
			c.Next()
		}) */

	router.GET(portal.Path, func(c *gin.Context) {
		c.HTML(http.StatusOK, portal.Home, gin.H{})
	})
	//		router.Run(fmt.Sprintf(":%d", portal.Port))

	// Start the server
	router.Run(fmt.Sprintf(":%d", config.Port))

	ilog.Info(fmt.Sprintf("Started portal on port %d, page:%s, logon: %s", portal.Port, portal.Home, portal.Logon))

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

func GinMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}
