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
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}

func Test_loadpluginControllerModule(t *testing.T) {
	type args struct {
		controllerPath string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadpluginControllerModule(tt.args.controllerPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadpluginControllerModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadpluginControllerModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getpluginHandlerFunc(t *testing.T) {
	type args struct {
		module reflect.Value
		name   string
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getpluginHandlerFunc(tt.args.module, tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getpluginHandlerFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGinMiddleware(t *testing.T) {
	type args struct {
		headers map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GinMiddleware(tt.args.headers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GinMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	type args struct {
		allowOrigin string
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CORSMiddleware(tt.args.allowOrigin); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CORSMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}
