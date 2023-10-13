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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/mdaxf/iac/com"
)

type Controller struct {
	Path      string     `json:"path"`
	Module    string     `json:"module"`
	Endpoints []Endpoint `json:"endpoints"`
}

type PluginController struct {
	Path      string     `json:"path"`
	Endpoints []Endpoint `json:"endpoints"`
}
type Endpoint struct {
	Path    string `json:"path"`
	Method  string `json:"method"`
	Handler string `json:"handler"`
}
type Config struct {
	Port              int                `json:"port"`
	Controllers       []Controller       `json:"controllers"`
	PluginControllers []PluginController `json:"plugins"`
	Portal            Portal             `json:"portal"`
}

type Portal struct {
	Port  int    `json:"port"`
	Path  string `json:"path"`
	Home  string `json:"home"`
	Logon string `json:"logon"`
}

var apiconfig = "apiconfig.json"
var gconfig = "configuration.json"

func LoadConfig() (*Config, error) {
	data, err := ioutil.ReadFile(apiconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %v", err)
	}
	fmt.Println("loaded portal and api configuration:", config)
	return &config, nil
}

func LoadGlobalConfig() (*GlobalConfig, error) {
	jsonFile, err := ioutil.ReadFile(gconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %v", err)
	}

	// Create a map to hold the JSON data
	var jsonData GlobalConfig

	// Unmarshal the JSON data into the map
	if err := json.Unmarshal(jsonFile, &jsonData); err != nil {

		return nil, fmt.Errorf("failed to parse configuration file: %v", err)
	}
	//fmt.Println(jsonFile, jsonData)

	com.Instance = jsonData.Instance
	com.InstanceType = jsonData.InstanceType
	com.InstanceName = jsonData.InstanceName
	com.SingalRConfig = jsonData.SingalRConfig
	//fmt.Println(com.SingalRConfig, com.Instance)

	fmt.Println("loaded global configuration:", jsonData)
	return &jsonData, nil
}
