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
	"testing"

	"github.com/gin-gonic/gin"
	config "github.com/mdaxf/iac/config"
)

func Test_loadControllers(t *testing.T) {
	type args struct {
		router      *gin.Engine
		controllers []config.Controller
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test Case 1",
			args: args{
				router:      nil, // Set the expected output here,
				controllers: nil, // Set whether an error is expected or not
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadControllers(tt.args.router, tt.args.controllers)
		})
	}
}

/*
func Test_getModule(t *testing.T) {
	type args struct {
		module string
	}
	tests := []struct {
		name string
		args args
		want reflect.Value
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getModule(tt.args.module); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHandlerFunc(t *testing.T) {
	type args struct {
		module reflect.Value
		name   string
	}
	tests := []struct {
		name    string
		args    args
		want    gin.HandlerFunc
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHandlerFunc(tt.args.module, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHandlerFunc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getHandlerFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createEndpoints(t *testing.T) {
	type args struct {
		router     *gin.Engine
		module     string
		modulepath string
		endpoints  []config.Endpoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := createEndpoints(tt.args.router, tt.args.module, tt.args.modulepath, tt.args.endpoints); (err != nil) != tt.wantErr {
				t.Errorf("createEndpoints() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}*/
