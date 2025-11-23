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
	"os"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/integration/activemq"
	"github.com/mdaxf/iac/integration/kafka"
	"github.com/mdaxf/iac/integration/mqttclient"
)

type Controller struct {
	Path      string     `json:"path"`
	Module    string     `json:"module"`
	Timeout   int        `json:"timeout"`
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
	Timeout           int                `json:"timeout"`
	Controllers       []Controller       `json:"controllers"`
	PluginControllers []PluginController `json:"plugins"`
	Portal            Portal             `json:"portal"`
	ApiKey            string             `json:"apikey"`
	OpenAiKey         string             `json:"openaikey"`
	OpenAiModel       string             `json:"openaimodel"`
}

type Portal struct {
	Port  int    `json:"port"`
	Path  string `json:"path"`
	Home  string `json:"home"`
	Logon string `json:"logon"`
}

var apiconfig = "apiconfig.json"
var apiconfigLocal = "apiconfig.local.json"
var gconfig = "configuration.json"

var MQTTClients map[string]*mqttclient.MqttClient
var Kakfas map[string]*kafka.KafkaConsumer
var ActiveMQs map[string]*activemq.ActiveMQ

func LoadConfig() (*Config, error) {
	// Load base configuration from apiconfig.json
	data, err := ioutil.ReadFile(apiconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %v", err)
	}

	// Check if local override exists
	if localData, err := ioutil.ReadFile(apiconfigLocal); err == nil {
		var localConfig Config
		if err := json.Unmarshal(localData, &localConfig); err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", apiconfigLocal, err)
		} else {
			// Override only root-level configuration, preserve controllers
			if localConfig.Port != 0 {
				config.Port = localConfig.Port
			}
			if localConfig.Timeout != 0 {
				config.Timeout = localConfig.Timeout
			}
			if localConfig.ApiKey != "" {
				config.ApiKey = localConfig.ApiKey
			}
			if localConfig.OpenAiKey != "" {
				config.OpenAiKey = localConfig.OpenAiKey
			}
			if localConfig.OpenAiModel != "" {
				config.OpenAiModel = localConfig.OpenAiModel
			}
			// Override Portal if any field is set
			if localConfig.Portal.Port != 0 || localConfig.Portal.Path != "" ||
				localConfig.Portal.Home != "" || localConfig.Portal.Logon != "" {
				config.Portal = localConfig.Portal
			}
			fmt.Printf("  - Applied local configuration overrides from %s\n", apiconfigLocal)
		}
	}

	// Load API key from environment variable first, fallback to config file
	if envApiKey := os.Getenv("IAC_API_KEY"); envApiKey != "" {
		ApiKey = envApiKey
	} else {
		ApiKey = config.ApiKey
	}

	// Load OpenAI key from environment variable first, fallback to config file
	if envOpenAiKey := os.Getenv("OPENAI_KEY"); envOpenAiKey != "" {
		OpenAiKey = envOpenAiKey
	} else {
		OpenAiKey = config.OpenAiKey
	}

	// Load OpenAI model from environment variable first, fallback to config file
	if envOpenAiModel := os.Getenv("OPENAI_MODEL"); envOpenAiModel != "" {
		OpenAiModel = envOpenAiModel
	} else if config.OpenAiModel != "" {
		OpenAiModel = config.OpenAiModel
	} else {
		// Default model if neither env var nor config is set
		OpenAiModel = "gpt-4o"
	}

	fmt.Println("loaded portal and api configuration")
	fmt.Printf("  - Port: %d\n", config.Port)
	fmt.Printf("  - API Key: %s\n", maskSecret(ApiKey))
	fmt.Printf("  - OpenAI Key: %s\n", maskSecret(OpenAiKey))
	fmt.Printf("  - OpenAI Model: %s\n", OpenAiModel)
	fmt.Printf("  - Controllers: %d\n", len(config.Controllers))

	// Load AI configuration (new multi-vendor support)
	_, aiErr := LoadAIConfig()
	if aiErr != nil {
		fmt.Printf("Warning: failed to load AI configuration: %v\n", aiErr)
		fmt.Println("  - Falling back to legacy OpenAI configuration from apiconfig.json")
	}

	return &config, nil
}

// maskSecret masks sensitive values for logging
func maskSecret(secret string) string {
	if secret == "" {
		return "[not set]"
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
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

	Transaction := jsonData.Transaction

	com.TransactionTimeout = com.ConverttoIntwithDefault(Transaction["timeout"], 15)
	com.DBTransactionTimeout = com.ConverttoIntwithDefault(jsonData.DatabaseConfig["timeout"], 5)

	fmt.Println("loaded global configuration:", jsonData)
	return &jsonData, nil
}
