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

func TestGinMiddleware(t *testing.T) {
	type args struct {
		headers map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		{
			name: "Test Case 1",
			args: args{
				headers: map[string]interface{}{},
			},
			want: nil, // Set the expected output here,
		}, // Add a comma here to fix the error
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
		{
			name: "Test Case 1",
			args: args{
				allowOrigin: "",
			},
			want: nil, // Set the expected output here,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CORSMiddleware(tt.args.allowOrigin); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CORSMiddleware() = %v, want %v", got, tt.want)
			}
		})
	}
}
