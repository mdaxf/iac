// the package is exported from github.com/beego/beego/v2/core/berror

// Copyright 2023. All Rights Reserved.

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

package berror

import (
	"reflect"
	"testing"
)

func TestDefineCode(t *testing.T) {
	type args struct {
		code   uint32
		module string
		name   string
		desc   string
	}
	tests := []struct {
		name string
		args args
		want Code
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefineCode(tt.args.code, tt.args.module, tt.args.name, tt.args.desc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefineCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_codeRegistry_Get(t *testing.T) {
	type args struct {
		code uint32
	}
	tests := []struct {
		name  string
		cr    *codeRegistry
		args  args
		want  Code
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.cr.Get(tt.args.code)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("codeRegistry.Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("codeRegistry.Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
