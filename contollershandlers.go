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
	"github.com/mdaxf/iac/controllers/collectionop"
	"github.com/mdaxf/iac/controllers/databaseop"
	"github.com/mdaxf/iac/controllers/function"
	healthcheck "github.com/mdaxf/iac/controllers/health"
	"github.com/mdaxf/iac/controllers/lngcodes"
	"github.com/mdaxf/iac/controllers/role"
	"github.com/mdaxf/iac/controllers/trans"
	"github.com/mdaxf/iac/controllers/user"
	"github.com/mdaxf/iac/framework/auth"
)

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
	}

	return reflect.Value{}
}

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
			router.Use(auth.AuthMiddleware())
			router.GET(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), handler)
		case http.MethodPost:

			if strings.Contains(modulepath, "/user/login") {
				router.POST(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), handler)
			} else {
				router.Use(auth.AuthMiddleware())
				router.POST(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), handler)
			}
		case http.MethodPut:
			router.Use(auth.AuthMiddleware())
			router.PUT(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), handler)
		case http.MethodPatch:
			router.Use(auth.AuthMiddleware())
			router.PATCH(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), handler)
		case http.MethodDelete:
			router.Use(auth.AuthMiddleware())
			router.DELETE(fmt.Sprintf("%s/%s", modulepath, endpoint.Path), handler)
		default:
			return fmt.Errorf("unsupported HTTP method '%s'", endpoint.Method)
		}
	}

	return nil
}

func getHandlerFunc(module reflect.Value, name string, controllercfg config.Controller) (gin.HandlerFunc, error) {
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		ilog.PerformanceWithDuration("main.getHandlerFunc", elapsed)
	}()

	defer func() {
		if r := recover(); r != nil {
			ilog.Error(fmt.Sprintf("Panic: %s", r))
			return
		}
	}()

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

func callWithTimeout(method reflect.Value, args []reflect.Value, timeout time.Duration) ([]reflect.Value, error) {
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
