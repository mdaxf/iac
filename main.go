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
	"github.com/google/uuid"
	configuration "github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/controllers/roles"
	dbconn "github.com/mdaxf/iac/databases"
	mongodb "github.com/mdaxf/iac/documents"
	funcs "github.com/mdaxf/iac/engine/function"
	"github.com/mdaxf/iac/framework/jobqueue"
	workflow "github.com/mdaxf/iac/workflow"
	"github.com/mdaxf/iac/framework/middleware"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/hubexecutor"
	"github.com/mdaxf/iac/services"
	"github.com/mdaxf/iac/services/inthubhistory"
	"github.com/mdaxf/iac/signalrserver"

	// Register database adapters
	_ "github.com/mdaxf/iac/databases/mssql"
	_ "github.com/mdaxf/iac/databases/mysql"
	_ "github.com/mdaxf/iac/databases/oracle"
	_ "github.com/mdaxf/iac/databases/postgres"

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

	// Initialize role-based execution
	rolesConfig := configuration.GetRolesConfig()
	roleInitializer := configuration.InitGlobalRoleInitializer(rolesConfig)
	log.Printf("Configured roles: %v", rolesConfig.GetEnabledRoles())

	// Load AI configuration
	log.Println("Loading AI configuration...")
	if _, err := configuration.LoadAIConfig(); err != nil {
		log.Printf("Warning: Failed to load AI configuration: %v", err)
		log.Println("AI features will be limited. Please ensure aiconfig.json exists in the working directory.")
	} else {
		log.Println("AI configuration loaded successfully")
	}

	// Initialize AI config accessor for workflow AI tasks
	initializeAITaskAccessor()

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

	// Set up role initializer callbacks
	setupRoleCallbacks(roleInitializer, config)

	// Log database connection status
	ilog.Info(fmt.Sprintf("Database connection status - Type: %s, Connected: %v", dbconn.DatabaseType, dbconn.DB != nil))

	if dbconn.DB != nil {
		defer dbconn.DB.Close()

		// Test the database connection
		if err := dbconn.DB.Ping(); err != nil {
			ilog.Error(fmt.Sprintf("Database connection test failed: %v", err))
		} else {
			ilog.Info("Database connection is alive")
		}

		// Initialize GORM database for new AI and reporting features
		ilog.Info("Attempting to initialize GORM database...")
		if err := gormdb.InitGormDB(); err != nil {
			ilog.Error(fmt.Sprintf("Failed to initialize GORM database: %v", err))
			ilog.Error("ChatController and AI features will not be available")
		} else {
			ilog.Info("GORM database initialized successfully")

			// Initialize vector database if configured in aiconfig.json
			ilog.Info("Checking for vector database configuration...")
			if err := services.InitializeVectorDatabase(gormdb.DB); err != nil {
				ilog.Error(fmt.Sprintf("Failed to initialize vector database: %v", err))
				ilog.Error("Vector embeddings features may not work properly")
			}
		}
	} else {
		//log.Fatalf("Failed to connect to database")
		ilog.Error("Failed to connect to database - dbconn.DB is nil")
		ilog.Error("GORM and ChatController will not be available")
	}

	// Cleanup legacy MongoDB connection (kept for backward compatibility)
	if mongodb.DocDBCon.MongoDBClient != nil {
		defer mongodb.DocDBCon.MongoDBClient.Disconnect(context.Background())
	} else {
		ilog.Error("Failed to connect to legacy document database")
	}

	// Cleanup global MongoDBClients array (kept for backward compatibility - deprecated)
	for _, dbclient := range com.MongoDBClients {
		if dbclient != nil {
			defer dbclient.Disconnect(context.Background())
		} else {
			ilog.Error("Failed to connect to the configured document database")
		}
	}

	// Cleanup all document database instances via unified factory
	defer func() {
		if err := mongodb.CloseAllDocDBs(); err != nil {
			ilog.Error(fmt.Sprintf("Failed to close all document database instances: %v", err))
		} else {
			ilog.Info("Successfully closed all document database instances via factory")
		}
	}()

	// Wait for SignalR to be ready (it's initialized in a goroutine)
	ilog.Info("Waiting for SignalR connection to complete...")
	select {
	case <-SignalRReady:
		ilog.Info("SignalR initialization complete")
	case <-time.After(10 * time.Second):
		ilog.Warn("SignalR connection timeout - continuing with job system initialization")
	}

	// Initialize background job system (only if job_executor role is enabled)
	if configuration.HasRole(configuration.RoleJobExecutor) {
		if dbconn.DB != nil && mongodb.DocDBCon != nil {
			ilog.Info("Initializing background job system (job_executor role enabled)...")
			ctx := context.Background()

			// Get cache instance for job queue (if available)
			var cacheInstance = configuration.SessionCache

			if err := jobqueue.InitializeJobSystem(ctx, dbconn.DB, mongodb.DocDBCon, cacheInstance, com.IACMessageBusClient); err != nil {
				ilog.Error(fmt.Sprintf("Failed to initialize job system: %v", err))
				ilog.Error("Background jobs and scheduled tasks will not be available")
			} else {
				ilog.Info("Background job system initialized successfully")

				// Ensure job system is shut down gracefully
				defer func() {
					ilog.Info("Shutting down job system...")
					if err := jobqueue.ShutdownJobSystem(); err != nil {
						ilog.Error(fmt.Sprintf("Error shutting down job system: %v", err))
					}
				}()
			}
		} else {
			ilog.Warn("Skipping job system initialization - database or document DB not available")
		}
	} else {
		ilog.Info("Job executor role not enabled - skipping job system initialization")
	}

	if com.IACMessageBusClient != nil {
		defer com.IACMessageBusClient.Stop()
	} else {
		//log.Fatalf("Failed to connect to database")
		ilog.Error("Failed to connect to the configured message bus")
	}

	ilog.Info("Initializing IAC application server...")

	// Use gin.New() + explicit Logger so we can suppress health-check / root-path
	// request noise.  gin.Default() would include the same Logger + Recovery but
	// with no path-skip capability.
	router = gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		// Suppress successful GET requests on paths that are polled by load-balancers
		// and health monitors – they flood logs without adding diagnostic value.
		// Non-2xx responses on these paths are still emitted (skip func returns false).
		SkipPaths: []string{"/", "/health/check", "/health", "/status"},
	}))
	router.Use(gin.Recovery())
	// Ensure CORS middleware is registered early so it runs even for preflight requests
	router.Use(CORSMiddleware("*"))

	// Add API call history middleware (records API calls for auditing)
	router.Use(middleware.APICallHistoryMiddleware())

	// Initialize roles with the router
	if err := roleInitializer.Initialize(context.Background(), router); err != nil {
		ilog.Error(fmt.Sprintf("Failed to initialize roles: %v", err))
	} else {
		ilog.Info("Roles initialized successfully")
	}

	// Ensure roles are shut down gracefully
	defer func() {
		ilog.Info("Shutting down roles...")
		if err := roleInitializer.Shutdown(context.Background()); err != nil {
			ilog.Error(fmt.Sprintf("Error shutting down roles: %v", err))
		}
	}()

	// Note: OPTIONS requests are handled by CORSMiddleware
	// Do NOT add a global OPTIONS("/*path") handler as it conflicts with other wildcard routes
	// Note: Controllers, portal, and static files are now loaded by the app role initializer
	// in setupRoleCallbacks. This ensures proper sequencing with other role components.

	// Start the HTTP server only if app role is enabled (API endpoints)
	if configuration.HasRole(configuration.RoleApp) {
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

		ilog.Info(fmt.Sprintf("Started IAC API server on port %d", config.Port))
	} else {
		ilog.Info("App role not enabled - API server not started")
	}

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
		origin := c.Request.Header.Get("Origin")

		// If allowOrigin is "*", use the request origin for credentials support
		// Otherwise use the configured allowOrigin
		if allowOrigin == "*" && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else if allowOrigin != "" {
			c.Header("Access-Control-Allow-Origin", allowOrigin)
			if allowOrigin != "*" {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, X-Client-Id, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

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

func initializeAITaskAccessor() {
	// Set up the global AI config accessor for workflow AI tasks
	// This allows engine/function to access config without creating import cycle
	log.Println("Initializing AI task accessor for workflows...")

	funcs.InitializeAITaskConfig(func() interface{} {
		return configuration.GetAIConfig()
	})

	// Wire up agent runtime callbacks — use lazy service lookup so the functions
	// are registered immediately but services are resolved at call time (after
	// controller initialization completes).
	funcs.InitializeAgentRuntime(
		func(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error) {
			runner := services.GetGlobalAgentRunnerService()
			if runner == nil {
				return "", fmt.Errorf("agent runner service not available")
			}
			return runner.RunSync(ctx, agentID, agentName, prompt, timeoutSecs)
		},
		func(ctx context.Context, skillName, argsJSON string) (string, error) {
			registry := services.GetGlobalToolRegistry()
			if registry == nil {
				return "", fmt.Errorf("tool registry not available")
			}
			return registry.ExecuteTool(ctx, skillName, argsJSON)
		},
	)

	// Wire up Integration Hub outbound send function for BPM engine and workflow nodes.
	// Uses lazy lookup of the global hub manager so it works regardless of init order.
	outboundSendFn := func(ctx context.Context, hubID, protocolGroupID, endpointID string, payload []byte, contentType, method string) (string, int, error) {
		mgr := hubexecutor.GetGlobalManager()
		if mgr == nil {
			return "", 0, fmt.Errorf("hub executor manager not initialized")
		}
		result, err := mgr.SendToEndpoint(ctx, hubID, protocolGroupID, endpointID, payload, contentType, method)
		if err != nil {
			return "", 0, err
		}
		return result.ResponseBody, result.StatusCode, nil
	}
	funcs.InitializeOutboundRuntime(outboundSendFn)
	workflow.InitializeOutboundRuntime(outboundSendFn)

	log.Println("AI task accessor and agent runtime callbacks initialized successfully")
}

func waitForTerminationSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("\nShutting down...")

	// Gracefully stop all channel pollers — each sends its configured stop message
	// to channel users before the goroutine exits.  Allow up to 15 s for delivery.
	if monitorSvc := services.GetGlobalAgentMonitorService(); monitorSvc != nil {
		fmt.Println("Stopping channel pollers and sending shutdown notifications...")
		monitorSvc.StopAll(15 * time.Second)
	}

	os.Exit(0)
}

// setupRoleCallbacks configures the role initializer with role-specific callbacks
func setupRoleCallbacks(roleInitializer *configuration.RoleInitializer, appConfig *configuration.Config) {
	log.Println("Setting up role callbacks...")

	// App role callback - registers API endpoints, portal, and static files
	log.Println("  - Setting app role callback")
	roleInitializer.SetAppInitializer(
		func(ctx context.Context, roleConfig *configuration.AppRoleConfig, router *gin.Engine) error {
			log.Println(">>> App role callback INVOKED <<<")
			ilog.Info("Initializing app role...")

			// Register roles API controller (fast, do synchronously)
			rolesController := roles.NewRolesController()
			rolesController.RegisterRoutes(router.Group("/api"))
			ilog.Info("Registered roles API controller")

			// Setup portal and static files if enabled (fast, do synchronously)
			if roleConfig != nil && roleConfig.EnablePortal {
				ilog.Info("Setting up portal...")
				setupPortal(router, appConfig)
			}

			if roleConfig != nil && roleConfig.EnableStaticFiles {
				ilog.Info("Setting up static files...")
				setupStaticFiles(router, appConfig)
			}

			// Setup proxy routes from webserver config (fast, do synchronously)
			if configuration.GlobalConfiguration != nil && configuration.GlobalConfiguration.WebServerConfig != nil {
				if proxy, ok := configuration.GlobalConfiguration.WebServerConfig["proxy"].(map[string]interface{}); ok && len(proxy) > 0 {
					ilog.Info("Setting up proxy routes...")
					renderproxy(proxy, router)
				}
			}

			// Load plugin controllers
			plugincontrollers := make(map[string]interface{})
			for _, controllerConfig := range appConfig.PluginControllers {
				controllerModule, err := loadpluginControllerModule(controllerConfig.Path)
				if err != nil {
					ilog.Error(fmt.Sprintf("Failed to load controller module %s: %v", controllerConfig.Path, err))
					continue
				}
				plugincontrollers[controllerConfig.Path] = controllerModule
			}

			// Create endpoints for plugin controllers
			for _, controllerConfig := range appConfig.PluginControllers {
				for _, endpointConfig := range controllerConfig.Endpoints {
					method := endpointConfig.Method
					path := fmt.Sprintf("/%s%s", controllerConfig.Path, endpointConfig.Path)
					if module, ok := plugincontrollers[controllerConfig.Path]; ok {
						if moduleMap, ok := module.(map[string]interface{}); ok {
							if handler, ok := moduleMap[endpointConfig.Handler].(func(*gin.Context)); ok {
								router.Handle(method, path, handler)
								ilog.Info(fmt.Sprintf("Registered plugin endpoint: %s %s", method, path))
							}
						}
					}
				}
			}

			// Load all API controllers
			ilog.Info("Loading API controllers...")
			loadControllers(router, appConfig.Controllers)
			ilog.Info("API controllers loaded successfully")
			return nil
		},
		func(ctx context.Context) error {
			ilog.Info("Shutting down app role...")
			return nil
		},
	)

	// Integration Hub role callback
	roleInitializer.SetIntegrationHubInitializer(
		func(ctx context.Context, config *configuration.IntegrationHubRoleConfig, router *gin.Engine) error {
			ilog.Info("Initializing integration hub role...")

			// Initialize the hub history service
			historyService, err := inthubhistory.NewIntHubHistoryService(nil)
			if err != nil {
				ilog.Error(fmt.Sprintf("Failed to initialize hub history service: %v", err))
			} else {
				inthubhistory.SetGlobalIntHubHistoryService(historyService)
				ilog.Info("Global hub history service initialized and set")
			}

			// Initialize the hub executor manager
			manager := hubexecutor.InitGlobalManager(hubexecutor.ManagerConfig{
				HubLoader: hubexecutor.NewMongoHubLoader(hubexecutor.MongoHubLoaderConfig{
					GetHubByHubName: services.GetDefaultHub,
					GetHubByID:      services.GetHub,
					GetAllHubs:      services.GetDefaultEnabledHubs,
				}),
				HistoryRecorder: historyService,
			})
			ilog.Info("Global hub executor manager initialized with history recorder")

			// Set job executor callback
			manager.SetJobExecutor(func(ctx context.Context, jobName string, params map[string]interface{}) error {
				// Execute job using job queue
				return jobqueue.ExecuteJobByName(ctx, jobName, params)
			})

			// Register hub executor API endpoints if enabled
			if config.EnableAPI {
				controller := hubexecutor.NewHubExecutorController(manager)
				controller.RegisterRoutes(router.Group("/api"))
				ilog.Info("Registered hub executor API controller")
			}

			// Auto-start hubs if configured
			if config.AutoStart {
				go func() {
					// Wait a bit for everything to be ready
					time.Sleep(2 * time.Second)

					if len(config.HubNames) > 0 {
						// Start specific hubs
						for _, hubName := range config.HubNames {
							ilog.Info(fmt.Sprintf("Auto-starting hub executor: %s", hubName))
							if err := manager.StartExecutor(hubName); err != nil {
								ilog.Error(fmt.Sprintf("Failed to auto-start hub executor %s: %v", hubName, err))
							}
						}
					} else {
						// Start all enabled hubs
						ilog.Info("Auto-starting all enabled hub executors")
						if results := manager.StartAllExecutors(); len(results) > 0 {
							for name, err := range results {
								if err != nil {
									ilog.Error(fmt.Sprintf("Failed to auto-start hub executor %s: %v", name, err))
								} else {
									ilog.Info(fmt.Sprintf("Auto-started hub executor: %s", name))
								}
							}
						}
					}
				}()
			}

			return nil
		},
		func(ctx context.Context) error {
			ilog.Info("Shutting down integration hub role...")
			manager := hubexecutor.GetGlobalManager()
			if manager != nil {
				manager.StopAll()
			}
			return nil
		},
	)

	// SignalR role callback - starts embedded SignalR server
	roleInitializer.SetSignalRInitializer(
		func(ctx context.Context, config *configuration.SignalRRoleConfig) error {
			ilog.Info("Initializing SignalR server role...")

			// Create and start SignalR server
			server := signalrserver.InitGlobalSignalRServer(config)
			if err := server.Start(ctx); err != nil {
				return fmt.Errorf("failed to start SignalR server: %v", err)
			}

			ilog.Info(fmt.Sprintf("SignalR server started on port %d", config.Port))
			return nil
		},
		func(ctx context.Context) error {
			ilog.Info("Shutting down SignalR server role...")
			server := signalrserver.GetGlobalSignalRServer()
			if server != nil {
				return server.Stop(ctx)
			}
			return nil
		},
	)

	// Job Executor role callback
	roleInitializer.SetJobExecutorInitializer(
		func(ctx context.Context, config *configuration.JobExecutorRoleConfig) error {
			ilog.Info("Initializing job executor role...")
			// Note: Job system is initialized separately in main() due to dependencies
			// This callback is for role-specific configuration
			ilog.Info(fmt.Sprintf("Job executor config: workers=%d, scheduler=%v", config.Workers, config.EnableScheduler))
			return nil
		},
		func(ctx context.Context) error {
			ilog.Info("Shutting down job executor role...")
			// Note: Job system shutdown is handled separately in main()
			return nil
		},
	)

	ilog.Info("Role callbacks configured")
}

// setupPortal sets up the portal routes and client configuration endpoint
func setupPortal(router *gin.Engine, config *configuration.Config) {
	portal := config.Portal

	clientconfig := make(map[string]interface{})
	clientconfig["signalrconfig"] = com.SingalRConfig
	clientconfig["instance"] = com.Instance
	clientconfig["instanceType"] = com.InstanceType
	clientconfig["instanceName"] = com.InstanceName
	clientconfig["dbtype"] = dbconn.DatabaseType

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

	ilog.Info(fmt.Sprintf("Portal configured: page=%s, logon=%s", portal.Home, portal.Logon))
}

// setupStaticFiles sets up static file serving routes
func setupStaticFiles(router *gin.Engine, config *configuration.Config) {
	// Serve static files for 3D models storage with CORS headers
	router.GET("/storage/3d_models/*filepath", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.File("./storage/3d_models/" + c.Param("filepath"))
	})

	// Serve static files for documents storage with CORS headers
	router.GET("/storage/documents/*filepath", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.File("./storage/documents/" + c.Param("filepath"))
	})

	ilog.Info("Static file routes configured")
}
