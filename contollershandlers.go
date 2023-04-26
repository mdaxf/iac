package main

import (
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/role"
	"github.com/mdaxf/iac/controllers/user"
)

func loadControllers(router *gin.Engine, controllers []Controller) {

	for _, controllerConfig := range controllers {
		log.Println("loadControllers:%s", controllerConfig.Module)
		err := createEndpoints(router, controllerConfig.Module, controllerConfig.Endpoints)
		if err != nil {
			log.Println(err)
		}
	}

	//return nil
}

func getModule(module string) reflect.Value {
	switch module {
	case "RoleController":
		moduleInstance := &role.RoleController{}
		return reflect.ValueOf(moduleInstance)

	case "UserController":
		moduleInstance := &user.UserController{}
		return reflect.ValueOf(moduleInstance)

	}
	return reflect.Value{}
}

func createEndpoints(router *gin.Engine, module string, endpoints []Endpoint) error {

	log.Println("createEndpoints:%s", module)

	//moduleValue := reflect.ValueOf(module)

	moduleValue := getModule(module)

	for _, endpoint := range endpoints {
		// Get the handler function for the endpoint method

		//handlermethod := reflect.ValueOf(moduleValue).MethodByName(endpoint.Handler);

		handler, err := getHandlerFunc(moduleValue, endpoint.Handler)
		if err != nil {
			return fmt.Errorf("error creating endpoint '%s': %v", endpoint.Path, err)
		}

		// Add the API endpoint to the router
		switch endpoint.Method {
		case http.MethodGet:
			router.GET(endpoint.Path, handler)
		case http.MethodPost:
			router.POST(endpoint.Path, handler)
		case http.MethodPut:
			router.PUT(endpoint.Path, handler)
		case http.MethodPatch:
			router.PATCH(endpoint.Path, handler)
		case http.MethodDelete:
			router.DELETE(endpoint.Path, handler)
		default:
			return fmt.Errorf("unsupported HTTP method '%s'", endpoint.Method)
		}
	}

	return nil
}

func getHandlerFunc(module reflect.Value, name string) (gin.HandlerFunc, error) {
	log.Println("getHandlerFunc:%s", name)

	if module.Kind() != reflect.Ptr || module.IsNil() {
		return nil, fmt.Errorf("invalid module value: %v", module)
	}

	method := module.MethodByName(name)
	if !method.IsValid() {
		return nil, fmt.Errorf("invalid method name: %s", name)
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
	}, nil
}

/*
func createEndpoints(router *gin.Engine, module string, endpoints []Endpoint) error {
	for _, endpoint := range endpoints {
		// Get the handler function for the endpoint method

		handler := getHandlerFunc(reflect.ValueOf(module), endpoint.Handler)

		// Add the API endpoint to the router
		switch endpoint.Method {
		case http.MethodGet:
			router.GET(endpoint.Path, handler)
		case http.MethodPost:
			router.POST(endpoint.Path, handler)
		case http.MethodPut:
			router.PUT(endpoint.Path, handler)
		case http.MethodPatch:
			router.PATCH(endpoint.Path, handler)
		case http.MethodDelete:
			router.DELETE(endpoint.Path, handler)
		default:
			return fmt.Errorf("unsupported HTTP method '%s'", endpoint.Method)
		}
	}

	return nil
}

func getHandlerFunc(module reflect.Value, name string) gin.HandlerFunc {
	log.Println("getHandlerFunc:%s, %s", module.Pointer(), name)

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
*/
