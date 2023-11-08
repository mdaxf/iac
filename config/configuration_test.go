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

package config

import (
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	type Database struct {
		Host string
		Port int
	}

	type Config struct {
		Database Database
	}

	tests := []struct {
		name    string
		want    *Config
		wantErr bool
	}{
		{
			name: "Test Case 1",
			want: &Config{
				Database: Database{
					Host: "localhost",
					Port: 5432,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadGlobalConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *GlobalConfig
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Test Case 1",
			want: &GlobalConfig{
				InstanceName: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadGlobalConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGlobalConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadGlobalConfig() = %v, want %v", got, tt.want)
			}
		})
	}

}
