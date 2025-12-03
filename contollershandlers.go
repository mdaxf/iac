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
	"net/http"
	"reflect"
	"strings"
	"time"

	//	"github.com/gin-contrib/timeout"
	config "github.com/mdaxf/iac/config"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/ai"
	"github.com/mdaxf/iac/controllers/aiconfig"
	"github.com/mdaxf/iac/controllers/bpmcontroller"
	"github.com/mdaxf/iac/controllers/collectionop"
	"github.com/mdaxf/iac/controllers/component"
	"github.com/mdaxf/iac/controllers/configmng"
	"github.com/mdaxf/iac/controllers/databaseop"
	"github.com/mdaxf/iac/controllers/function"
	healthcheck "github.com/mdaxf/iac/controllers/health"
	"github.com/mdaxf/iac/controllers/iacai"
	"github.com/mdaxf/iac/controllers/lngcodes"
	"github.com/mdaxf/iac/controllers/models3d"
	"github.com/mdaxf/iac/controllers/notifications"
	"github.com/mdaxf/iac/controllers/processplan"
	"github.com/mdaxf/iac/controllers/report"
	"github.com/mdaxf/iac/controllers/role"
	"github.com/mdaxf/iac/controllers/schema"
	"github.com/mdaxf/iac/controllers/trans"
	"github.com/mdaxf/iac/controllers/user"
	"github.com/mdaxf/iac/controllers/workflow"
	"github.com/mdaxf/iac/framework/auth"
	"github.com/mdaxf/iac/gormdb"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/services"
)

// GetMongoDBConnection returns the MongoDB DocumentDB connection using the unified factory pattern
func GetMongoDBConnection() documents.DocumentDB {
	// Use unified factory pattern exclusively
	factory := documents.GetDocFactory()
	if factory != nil {
		// Try to get the "default" instance first
		if db, err := factory.GetDB("default"); err == nil && db != nil {
			return db
		}

		// If default not found, log warning and return nil
		ilog.Warn("No default MongoDB instance found in factory - document database may not be initialized")

		// List available instances for debugging
		instances := factory.ListInstances()
		if len(instances) > 0 {
			ilog.Info(fmt.Sprintf("Available database instances: %v", instances))
		}
	}

	ilog.Warn("No MongoDB connection available - factory returned nil")
	return nil
}

// loadControllers loads the specified controllers into the router.
// It iterates over the controllers and calls createEndpoints to create the endpoints for each controller.
// The performance duration of the function is logged using ilog.PerformanceWithDuration.
// If an error occurs while loading a controller module, an error message is logged using ilog.Error.
// The function returns the error returned by createEndpoints.
// The function is called by main.
func loadControllers(router *gin.Engine, controllers []config.Controller) {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.loadControllers", elapsed)
	}()

	for _, controllerConfig := range controllers {
		ilog.Info(fmt.Sprintf("loadControllers:%s", controllerConfig.Module))

		err := createEndpoints(router, controllerConfig.Module, controllerConfig.Path, controllerConfig.Endpoints, controllerConfig)
		if err != nil {
			ilog.Error(fmt.Sprintf("Failed to load controller module %s: %v", controllerConfig.Module, err))
		}
	}
}

// getModule returns a reflect.Value of the specified module.
// It measures the performance of the function and logs the elapsed time.
// The module parameter specifies the name of the module to retrieve.
// The function returns a reflect.Value of the module instance.
// If the module is not found, it returns an empty reflect.Value.
// The function is called by loadControllers.
func getModule(module string) reflect.Value {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.getModule", elapsed)
	}()

	ilog.Info(fmt.Sprintf("loadControllers get controller instance:%s", module))

	switch module {
	case "RoleController":
		moduleInstance := &role.RoleController{}
		return reflect.ValueOf(moduleInstance)

	case "UserController":
		moduleInstance := &user.UserController{}
		return reflect.ValueOf(moduleInstance)

	case "TranCodeController":
		moduleInstance := &trans.TranCodeController{}
		return reflect.ValueOf(moduleInstance)

	case "CollectionController":
		moduleInstance := &collectionop.CollectionController{}
		return reflect.ValueOf(moduleInstance)
	case "DBController":
		moduleInstance := &databaseop.DBController{}
		return reflect.ValueOf(moduleInstance)
	case "FunctionController":
		moduleInstance := &function.FunctionController{}
		return reflect.ValueOf(moduleInstance)
	case "LCController":
		moduleInstance := &lngcodes.LCController{}
		return reflect.ValueOf(moduleInstance)
	case "HealthController":
		moduleInstance := &healthcheck.HealthController{}
		return reflect.ValueOf(moduleInstance)
	case "NotificationController":
		moduleInstance := &notifications.NotificationController{}
		return reflect.ValueOf(moduleInstance)

	case "WorkFlowController":
		moduleInstance := &workflow.WorkFlowController{}
		return reflect.ValueOf(moduleInstance)

	case "BPMController":
		moduleInstance := &bpmcontroller.BPMController{}
		return reflect.ValueOf(moduleInstance)

	case "IACComponentController":
		moduleInstance := &component.IACComponentController{}
		return reflect.ValueOf(moduleInstance)

	case "IACAIController":
		// Initialize AI Agency Service when IACAIController is loaded
		// This happens once during application startup

		// Get services needed for AI Agency
		schemaMetadataService := services.NewSchemaMetadataService(gormdb.DB)
		schemaEmbeddingService := services.NewSchemaEmbeddingService(gormdb.DB, config.OpenAiKey)
		aiReportService := services.NewAIReportService(config.OpenAiKey, config.OpenAiModel)
		chatService := services.NewChatService(gormdb.DB, config.OpenAiKey, config.OpenAiModel, schemaMetadataService)

		// Initialize AI Agency Service
		aiAgencyService := services.NewAIAgencyService(
			gormdb.DB,
			config.OpenAiKey,
			config.OpenAiModel,
			chatService,
			aiReportService,
			schemaMetadataService,
			schemaEmbeddingService,
		)

		// Set the service for the controller
		iacai.SetAIAgencyService(aiAgencyService)

		ilog.Info("AI Agency Service initialized successfully")

		moduleInstance := &iacai.IACAIController{}
		return reflect.ValueOf(moduleInstance)

	case "ProcessPlanController":
		moduleInstance := &processplan.ProcessPlanController{}
		return reflect.ValueOf(moduleInstance)
	case "SchemaController":
		moduleInstance := &schema.SchemaController{}
		return reflect.ValueOf(moduleInstance)

	case "Models3DController":
		moduleInstance := &models3d.Models3DController{}
		return reflect.ValueOf(moduleInstance)

	case "ConfigController":
		moduleInstance := &configmng.ConfigController{}
		return reflect.ValueOf(moduleInstance)

	case "ReportController":
		moduleInstance := report.NewReportController()
		return reflect.ValueOf(moduleInstance)

	case "ReportDocController":
		// Get MongoDB connection from document DB registry
		docDB := GetMongoDBConnection()
		if docDB != nil {
			// Create SchemaMetadataService for query execution
			// Use the global GORM DB connection
			schemaMetadataService := services.NewSchemaMetadataService(gormdb.DB)
			moduleInstance := report.NewReportDocController(docDB, schemaMetadataService)
			return reflect.ValueOf(moduleInstance)
		}
		ilog.Warn("MongoDB not available, ReportDocController not loaded")
		return reflect.Value{}

	case "ChatController":
		moduleInstance := report.NewChatController()
		return reflect.ValueOf(moduleInstance)

	case "SchemaMetadataController":
		moduleInstance := ai.NewSchemaMetadataController()
		return reflect.ValueOf(moduleInstance)

	case "BusinessEntityController":
		moduleInstance := ai.NewBusinessEntityController()
		return reflect.ValueOf(moduleInstance)

	case "QueryTemplateController":
		moduleInstance := ai.NewQueryTemplateController()
		return reflect.ValueOf(moduleInstance)

	case "AIConfigController":
		moduleInstance := &aiconfig.AIConfigController{}
		return reflect.ValueOf(moduleInstance)

	case "AIEmbeddingController":
		moduleInstance := ai.NewAIEmbeddingController()
		return reflect.ValueOf(moduleInstance)

	case "SchemaEmbeddingController":
		// Get OpenAI API key from AI configuration
		openAIKey := getEmbeddingAPIKey()
		moduleInstance := ai.NewSchemaEmbeddingController(gormdb.DB, openAIKey)
		return reflect.ValueOf(moduleInstance)

	case "SchemaContextController":
		// Get OpenAI API key from AI configuration
		openAIKey := getEmbeddingAPIKey()
		moduleInstance := ai.NewSchemaContextController(gormdb.DB, openAIKey)
		return reflect.ValueOf(moduleInstance)

	}
	return reflect.Value{}
}

// getEmbeddingAPIKey retrieves the API key for embedding from AI configuration
// It looks for the embedding use case configuration and returns the appropriate vendor's API key
func getEmbeddingAPIKey() string {
	if config.AIConf == nil {
		ilog.Warn("AI configuration not loaded, attempting to load...")
		if _, err := config.LoadAIConfig(); err != nil {
			ilog.Error(fmt.Sprintf("Failed to load AI config: %v", err))
			return ""
		}
	}

	// Check for embedding use case configuration
	if embeddingConfig, exists := config.AIConf.UseCases["embedding"]; exists {
		vendorName := embeddingConfig.Vendor
		if vendor, vendorExists := config.AIConf.AIVendors[vendorName]; vendorExists && vendor.Enabled {
			ilog.Info(fmt.Sprintf("Using %s for embeddings with model: %s", vendorName, embeddingConfig.ModelOverride))
			return vendor.APIKey
		}
	}

	// Fallback: check for OpenAI vendor if it's enabled
	if openAI, exists := config.AIConf.AIVendors["openai"]; exists && openAI.Enabled {
		ilog.Info("Using OpenAI for embeddings (fallback)")
		return openAI.APIKey
	}

	ilog.Warn("No valid AI vendor configuration found for embeddings")
	return ""
}

// createEndpoints creates API endpoints for a given module using the provided router.
// It takes the module name, module path, list of endpoints, and controller configuration as input.
// The function adds the API endpoints to the router based on the HTTP method specified for each endpoint.
// It also applies authentication middleware to the appropriate endpoints.
// The function returns an error if there is any issue creating the endpoints.
// The function is called by loadControllers.
func createEndpoints(router *gin.Engine, module string, modulepath string, endpoints []config.Endpoint, controllercfg config.Controller) error {

	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.createEndpoints", elapsed)
	}()

	ilog.Info(fmt.Sprintf("Create the endpoints for the module:%s", module))

	//moduleValue := reflect.ValueOf(module)

	moduleValue := getModule(module)

	for _, endpoint := range endpoints {
		// Get the handler function for the endpoint method

		//handlermethod := reflect.ValueOf(moduleValue).MethodByName(endpoint.Handler);

		handler, err := getHandlerFunc(moduleValue, endpoint.Handler, controllercfg)
		if err != nil {
			return fmt.Errorf("error creating endpoint '%s': %v", endpoint.Path, err)
		}
		ilog.Debug(fmt.Sprintf("modulepath:%s, method:%s module:%s", modulepath, endpoint.Method, module))
		// Add the API endpoint to the router
		//auth.AuthMiddleware(),

		switch endpoint.Method {
		case http.MethodGet:
			router.GET(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), auth.AuthMiddleware(), handler)
		case http.MethodPost:

			if strings.Contains(modulepath, "/user/login") {
				router.POST(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), handler)
			} else {
				router.POST(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), auth.AuthMiddleware(), handler)
			}
		case http.MethodPut:
			router.PUT(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), auth.AuthMiddleware(), handler)
		case http.MethodPatch:
			router.PATCH(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), auth.AuthMiddleware(), handler)
		case http.MethodDelete:
			router.DELETE(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), auth.AuthMiddleware(), handler)
		default:
			return fmt.Errorf("unsupported HTTP method '%s'", endpoint.Method)
		}
	}

	return nil
}

// getHandlerFunc is a function that returns a gin.HandlerFunc based on the provided module, name, and controller configuration.
// It measures the performance duration of the handler function and recovers from any panics that occur.
// If the controller configuration specifies a timeout, the handler function is executed with a timeout.
// The handler function takes a *gin.Context as input and returns an HTTP response.
// If an error occurs during the execution of the handler function, it is returned as an HTTP error response.
// If the handler function returns a []byte, it is returned as the response body with the "application/json" content type.
// If no error or data is returned, the handler function sets the HTTP status code to 200.
// If the module value is invalid or the method name is invalid, an error is returned.
// The function is called by createEndpoints.
func getHandlerFunc(module reflect.Value, name string, controllercfg config.Controller) (gin.HandlerFunc, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.getHandlerFunc", elapsed)
	}()

	/*	defer func() {
			if r := recover(); r != nil {
				ilog.Error(fmt.Sprintf("Panic: %s", r))
				return
			}
		}()
	*/
	ilog.Info(fmt.Sprintf("Get Handler Function:%s", name))

	if module.Kind() != reflect.Ptr || module.IsNil() {
		return nil, fmt.Errorf("invalid module value: %v", module)
	}

	method := module.MethodByName(name)
	if !method.IsValid() {
		return nil, fmt.Errorf("invalid method name: %s", name)
	}

	if controllercfg.Timeout > 0 {
		return func(c *gin.Context) {
			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(c)
			out, err := callWithTimeout(method, in, time.Duration(controllercfg.Timeout)*time.Millisecond)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
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
		}, nil
	} else {
		return func(c *gin.Context) {

			defer func() {
				if r := recover(); r != nil {
					ilog.Error(fmt.Sprintf("Panic: %s", r))
					c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("Panic: %s", r))
					return
				}
			}()

			in := make([]reflect.Value, 1)
			in[0] = reflect.ValueOf(c)
			out := method.Call(in)
			if len(out) > 0 {
				if err, ok := out[0].Interface().(error); ok {
					ilog.Error(fmt.Sprintf("%s %s Error: %s", module, name, err))
					c.AbortWithError(http.StatusInternalServerError, err)
					return
				}
				if data, ok := out[0].Interface().([]byte); ok {
					ilog.Debug(fmt.Sprintf("%s %s Data: %s", module, name, data))
					c.Data(http.StatusOK, "application/json", data)
					return
				}
			}
			c.Status(http.StatusOK)
		}, nil
	}
}

// callWithTimeout is a function that executes a given method with a timeout.
// It takes a reflect.Value representing the method to be called, an array of reflect.Value arguments,
// and a timeout duration. It returns an array of reflect.Value results and an error.
// If the method execution completes within the timeout, the results are returned.
// If the timeout is exceeded, an error is returned.
// The function is called by getHandlerFunc.
func callWithTimeout(method reflect.Value, args []reflect.Value, timeout time.Duration) ([]reflect.Value, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration(fmt.Sprintf("main.callWithTimeout: %v", method), elapsed)
	}()

	resultChan := make(chan []reflect.Value, 1)
	//errChan := make(chan error, 1)

	go func() {
		result := method.Call(args)
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		ilog.Debug(fmt.Sprintf("callWithTimeout result: %s", result))
		return result, nil
	case <-time.After(timeout):
		ilog.Error(fmt.Sprintf("callWithTimeout timeout: %s", timeout))
		return nil, fmt.Errorf("timeout exceeded")
	}
}
