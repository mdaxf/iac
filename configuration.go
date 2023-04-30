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
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var gconfig = "config.json"

func loadConfig() (*Config, error) {
	data, err := ioutil.ReadFile(gconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %v", err)
	}

	return &config, nil
}
