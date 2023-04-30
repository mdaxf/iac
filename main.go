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

	"log"
	"net/http"

	"plugin"
	"reflect"

	"github.com/gin-gonic/gin"
	dbconn "github.com/mdaxf/iac/databases"
)

func main() {
	// Load configuration from the file
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	initialize()

	defer dbconn.DB.Close()
	// Initialize the Gin router
	router := gin.Default()

	// Load controllers dynamically based on the configuration file
	plugincontrollers := make(map[string]interface{})
	for _, controllerConfig := range config.PluginControllers {

		jsonString, err := json.Marshal(controllerConfig)
		if err != nil {
			fmt.Println("Error marshaling json:", err)
			return
		}
		fmt.Println(string(jsonString))
		controllerModule, err := loadpluginControllerModule(controllerConfig.Path)
		if err != nil {
			fmt.Println(fmt.Sprintf("Failed to load controller module %s: %v", controllerConfig.Path, err))
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

	loadControllers(router, config.Controllers)

	// Start the portals
	log.Println("Starting portals")

	jsonString, err := json.Marshal(config.Portal)
	if err != nil {
		fmt.Println("Error marshaling json:", err)
		return
	}
	fmt.Println(string(jsonString))

	portal := config.Portal
	log.Println(fmt.Sprintf("Starting portal on port %d, page:%s, logon: %s", portal.Port, portal.Home, portal.Logon))
	router.Static("/portal", "./portal")
	router.LoadHTMLGlob("portal/*.html")
	router.LoadHTMLGlob("portal/scripts/*.js")
	router.GET(portal.Path, func(c *gin.Context) {
		c.HTML(http.StatusOK, portal.Home, gin.H{})
	})
	//		router.Run(fmt.Sprintf(":%d", portal.Port))

	// Start the server
	router.Run(fmt.Sprintf(":%d", config.Port))

	//defer dbconn.DB.Close()
}

func loadpluginControllerModule(controllerPath string) (interface{}, error) {

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
