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
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	"github.com/mdaxf/iac/vendor_skip/github.com/google/uuid"

	"github.com/mdaxf/iac/com"
)

var wg sync.WaitGroup
var router *gin.Engine

// main is the entry point of the program.
// It loads the configuration file, initializes the database connection, and starts the server.
// It also loads controllers dynamically and statically based on the configuration file.
// The server is started on the port specified in the configuration file.
// The server serves static files from the portal directory.
// The server also serves static files from the plugins directory.

func main() {

	Initialized = false
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		if Initialized {
			ilog.PerformanceWithDuration("main", elapsed)
		}
	}()
	/*
		defer func() {
			if r := recover(); r != nil {
				if Initialized {
					ilog.Error(fmt.Sprintf("Panic: %v", r))
				} else {
					log.Fatalf("Panic: %v", r)
				}
			}
		}()  */
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
	com.NodeHeartBeats = make(map[string]interface{})

	com.IACNode = make(map[string]interface{})
	appid := uuid.New().String()
	com.IACNode["Name"] = "iac"
	com.IACNode["AppID"] = appid
	com.IACNode["Type"] = "Application Server"
	com.IACNode["Version"] = "1.0.0"
	com.IACNode["Description"] = "IAC Application Server"
	com.IACNode["Status"] = "Started"
	com.IACNode["StartTime"] = time.Now()
	data := make(map[string]interface{})
	data["Node"] = com.IACNode
	data["Result"] = make(map[string]interface{})

	com.NodeHeartBeats[appid] = data

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

	router.GET("/app/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, clientconfig)
	})

	router.GET("/app/debug", func(c *gin.Context) {
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

	ilog.Info(fmt.Sprintf("Started iac endpoint server on port %d, page:%s, logon: %s", portal.Port, portal.Home, portal.Logon))
	waitForTerminationSignal()
	elapsed := time.Since(startTime)
	ilog.PerformanceWithDuration("main.main", elapsed)

	wg.Wait()

}

// loadpluginControllerModule loads a plugin controller module from the specified controllerPath.
// It returns the loaded module as an interface{} and an error if any.
// The module is loaded from the plugins directory.
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

// getpluginHandlerFunc is a function that returns a gin.HandlerFunc for loading and executing a plugin handler function.
// It takes a module reflect.Value and the name of the handler function as parameters.
// The function loads the plugin handler function from the module using reflection and returns a gin.HandlerFunc.
// The returned handler function takes a *gin.Context as a parameter and executes the plugin handler function with the context.
// If the plugin handler function returns an error, the handler function aborts the request with a 500 Internal Server Error.
// If the plugin handler function returns a []byte, the handler function sends the data as a JSON response with a 200 OK status code.
// If the plugin handler function does not return an error or []byte, the handler function sends a 200 OK status code.
// If the plugin handler function does not exist, the function returns nil.
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

// CORSMiddleware is a middleware function that adds Cross-Origin Resource Sharing (CORS) headers to the HTTP response.
// It allows requests from a specified origin and supports various HTTP methods.
// The allowOrigin parameter specifies the allowed origin for CORS requests.
// This middleware function also handles preflight requests by responding with appropriate headers.

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

// GinMiddleware is a middleware function that sets the specified headers in the HTTP response.
// It takes a map of headers as input and returns a gin.HandlerFunc.
// The middleware sets the headers in the response using the values provided in the headers map.
// If the HTTP request method is OPTIONS, it aborts the request with a status code of 204 (No Content).
// After setting the headers, it calls the next handler in the chain.
// The next handler can be a controller function or another middleware function.
// The next handler can also be a gin.HandlerFunc.
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

// renderproxy is a function that renders a proxy configuration by creating routes in a gin.Engine instance.
// It takes a map of proxy configurations and a pointer to a gin.Engine as parameters.
// Each key-value pair in the proxy map represents a route path and its corresponding target URL.
// The function iterates over the proxy map and creates a route for each key-value pair in the gin.Engine instance.
// When a request matches a route, the function sets up a reverse proxy to forward the request to the target URL.
// The function also updates the request URL path based on the "path" parameter in the route.
// Note that the ServeHTTP method of the reverse proxy is non-blocking and uses a goroutine under the hood.

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

func waitForTerminationSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")

	time.Sleep(2 * time.Second) // Add any cleanup or graceful shutdown logic here
	os.Exit(0)
}
