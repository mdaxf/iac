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
	"sync"
)

// AIConfig represents the complete AI configuration
type AIConfig struct {
	AIVendors       map[string]AIVendorConfig `json:"ai_vendors"`
	UseCases        map[string]UseCaseConfig  `json:"use_cases"`
	VectorDatabase  VectorDatabaseConfig      `json:"vector_database"`
	Thresholds      ThresholdsConfig          `json:"thresholds"`
	Features        FeaturesConfig            `json:"features"`
	Monitoring      MonitoringConfig          `json:"monitoring"`
}

// AIVendorConfig represents configuration for an AI vendor
type AIVendorConfig struct {
	Enabled      bool                       `json:"enabled"`
	APIKey       string                     `json:"api_key"`
	APIBaseURL   string                     `json:"api_base_url"`
	Organization string                     `json:"organization_id,omitempty"`
	APIVersion   string                     `json:"api_version,omitempty"`
	ProjectID    string                     `json:"project_id,omitempty"`
	Location     string                     `json:"location,omitempty"`
	Deployment   string                     `json:"deployment_name,omitempty"`
	Models       map[string]string          `json:"models"`
	Parameters   map[string]interface{}     `json:"parameters"`
}

// UseCaseConfig represents configuration for a specific use case
type UseCaseConfig struct {
	Vendor            string                 `json:"vendor"`
	ModelOverride     string                 `json:"model_override"`
	SystemPrompt      string                 `json:"system_prompt,omitempty"`
	Temperature       float64                `json:"temperature,omitempty"`
	MaxTokens         int                    `json:"max_tokens,omitempty"`
	BatchSize         int                    `json:"batch_size,omitempty"`
	Dimension         int                    `json:"dimension,omitempty"`
	ConfidenceThreshold float64              `json:"confidence_threshold,omitempty"`
	CustomParams      map[string]interface{} `json:"custom_params,omitempty"`
}

// VectorDatabaseConfig represents vector database configuration
type VectorDatabaseConfig struct {
	Type            string                        `json:"type"`
	PostgresPGVector PostgresPGVectorConfig        `json:"postgres_pgvector"`
	Qdrant          QdrantConfig                  `json:"qdrant"`
	Pinecone        PineconeConfig                `json:"pinecone"`
	Weaviate        WeaviateConfig                `json:"weaviate"`
	ChromaDB        ChromaDBConfig                `json:"chromadb"`
}

// PostgresPGVectorConfig for pgvector extension
type PostgresPGVectorConfig struct {
	Enabled          bool   `json:"enabled"`
	UseMainDB        bool   `json:"use_main_db"`
	ConnectionString string `json:"connection_string"`
	Schema           string `json:"schema"`
	TablePrefix      string `json:"table_prefix"`
	Dimension        int    `json:"dimension"`
}

// QdrantConfig for Qdrant vector database
type QdrantConfig struct {
	Enabled        bool   `json:"enabled"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	APIKey         string `json:"api_key"`
	CollectionName string `json:"collection_name"`
	Dimension      int    `json:"dimension"`
	DistanceMetric string `json:"distance_metric"`
}

// PineconeConfig for Pinecone vector database
type PineconeConfig struct {
	Enabled     bool   `json:"enabled"`
	APIKey      string `json:"api_key"`
	Environment string `json:"environment"`
	IndexName   string `json:"index_name"`
	Dimension   int    `json:"dimension"`
	Metric      string `json:"metric"`
}

// WeaviateConfig for Weaviate vector database
type WeaviateConfig struct {
	Enabled   bool   `json:"enabled"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	APIKey    string `json:"api_key"`
	Scheme    string `json:"scheme"`
	ClassName string `json:"class_name"`
}

// ChromaDBConfig for ChromaDB vector database
type ChromaDBConfig struct {
	Enabled        bool   `json:"enabled"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	CollectionName string `json:"collection_name"`
}

// ThresholdsConfig represents various threshold configurations
type ThresholdsConfig struct {
	SQLGenerationConfidence float64 `json:"sql_generation_confidence"`
	SchemaMatchingConfidence float64 `json:"schema_matching_confidence"`
	MinSimilarityScore      float64 `json:"min_similarity_score"`
	AutoExecuteThreshold    float64 `json:"auto_execute_threshold"`
	ReviewThreshold         float64 `json:"review_threshold"`
	MaxRetryAttempts        int     `json:"max_retry_attempts"`
}

// FeaturesConfig represents feature flags
type FeaturesConfig struct {
	EnableCaching       bool   `json:"enable_caching"`
	CacheTTLSeconds     int    `json:"cache_ttl_seconds"`
	EnableLogging       bool   `json:"enable_logging"`
	LogPrompts          bool   `json:"log_prompts"`
	LogResponses        bool   `json:"log_responses"`
	EnableMetrics       bool   `json:"enable_metrics"`
	EnableCostTracking  bool   `json:"enable_cost_tracking"`
	EnableRateLimiting  bool   `json:"enable_rate_limiting"`
	RequestsPerMinute   int    `json:"requests_per_minute"`
	EnableFallback      bool   `json:"enable_fallback"`
	FallbackVendor      string `json:"fallback_vendor"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	EnableTelemetry       bool    `json:"enable_telemetry"`
	TelemetryEndpoint     string  `json:"telemetry_endpoint"`
	AlertOnErrors         bool    `json:"alert_on_errors"`
	AlertEmail            string  `json:"alert_email"`
	CostAlertThresholdUSD float64 `json:"cost_alert_threshold_usd"`
	DailyRequestLimit     int     `json:"daily_request_limit"`
}

var (
	aiConfigInstance *AIConfig
	aiConfigOnce     sync.Once
	aiConfigMutex    sync.RWMutex
)

const (
	aiConfigFile      = "aiconfig.json"
	aiConfigLocalFile = "aiconfig.local.json"
)

// Global exported variables for backward compatibility
var (
	// AIConf is the global AI configuration
	AIConf *AIConfig
)

// LoadAIConfig loads AI configuration from file with local override support
func LoadAIConfig() (*AIConfig, error) {
	var err error
	aiConfigOnce.Do(func() {
		aiConfigInstance, err = loadAIConfigInternal()
		if err == nil {
			AIConf = aiConfigInstance
		}
	})
	return aiConfigInstance, err
}

func loadAIConfigInternal() (*AIConfig, error) {
	// Load base configuration from aiconfig.json
	data, err := ioutil.ReadFile(aiConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read AI configuration file: %v", err)
	}

	var config AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse AI configuration file: %v", err)
	}

	// Check if local override exists
	if localData, err := ioutil.ReadFile(aiConfigLocalFile); err == nil {
		var localConfig AIConfig
		if err := json.Unmarshal(localData, &localConfig); err != nil {
			fmt.Printf("Warning: failed to parse %s: %v\n", aiConfigLocalFile, err)
		} else {
			// Merge local config with base config
			mergeAIConfig(&config, &localConfig)
			fmt.Printf("  - Applied local AI configuration overrides from %s\n", aiConfigLocalFile)
		}
	}

	// Override with environment variables
	applyEnvOverrides(&config)

	// Validate configuration
	if err := validateAIConfig(&config); err != nil {
		return nil, fmt.Errorf("AI configuration validation failed: %v", err)
	}

	fmt.Println("loaded AI configuration")
	fmt.Printf("  - Enabled vendors: %s\n", getEnabledVendors(&config))
	fmt.Printf("  - Vector DB type: %s\n", config.VectorDatabase.Type)
	fmt.Printf("  - Chatbot vendor: %s\n", config.UseCases["chatbot"].Vendor)
	fmt.Printf("  - Text2SQL vendor: %s\n", config.UseCases["text2sql"].Vendor)

	return &config, nil
}

// mergeAIConfig merges local config into base config
func mergeAIConfig(base, local *AIConfig) {
	// Merge AI Vendors
	for vendor, localVendorConfig := range local.AIVendors {
		if baseVendorConfig, exists := base.AIVendors[vendor]; exists {
			// Merge individual fields
			if localVendorConfig.APIKey != "" {
				baseVendorConfig.APIKey = localVendorConfig.APIKey
			}
			if localVendorConfig.APIBaseURL != "" {
				baseVendorConfig.APIBaseURL = localVendorConfig.APIBaseURL
			}
			if localVendorConfig.Organization != "" {
				baseVendorConfig.Organization = localVendorConfig.Organization
			}
			baseVendorConfig.Enabled = localVendorConfig.Enabled
			// Merge models
			for modelKey, modelValue := range localVendorConfig.Models {
				baseVendorConfig.Models[modelKey] = modelValue
			}
			// Merge parameters
			for paramKey, paramValue := range localVendorConfig.Parameters {
				baseVendorConfig.Parameters[paramKey] = paramValue
			}
			base.AIVendors[vendor] = baseVendorConfig
		} else {
			base.AIVendors[vendor] = localVendorConfig
		}
	}

	// Merge Use Cases
	for useCase, localUseCaseConfig := range local.UseCases {
		base.UseCases[useCase] = localUseCaseConfig
	}

	// Merge Vector Database config
	if local.VectorDatabase.Type != "" {
		base.VectorDatabase.Type = local.VectorDatabase.Type
	}
	if local.VectorDatabase.PostgresPGVector.Enabled {
		base.VectorDatabase.PostgresPGVector = local.VectorDatabase.PostgresPGVector
	}
	if local.VectorDatabase.Qdrant.Enabled {
		base.VectorDatabase.Qdrant = local.VectorDatabase.Qdrant
	}
	if local.VectorDatabase.Pinecone.Enabled {
		base.VectorDatabase.Pinecone = local.VectorDatabase.Pinecone
	}
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(config *AIConfig) {
	// OpenAI overrides
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		if openai, exists := config.AIVendors["openai"]; exists {
			openai.APIKey = key
			config.AIVendors["openai"] = openai
		}
	}

	// Anthropic overrides
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		if anthropic, exists := config.AIVendors["anthropic"]; exists {
			anthropic.APIKey = key
			config.AIVendors["anthropic"] = anthropic
		}
	}

	// Azure OpenAI overrides
	if key := os.Getenv("AZURE_OPENAI_API_KEY"); key != "" {
		if azure, exists := config.AIVendors["azure_openai"]; exists {
			azure.APIKey = key
			config.AIVendors["azure_openai"] = azure
		}
	}
	if endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT"); endpoint != "" {
		if azure, exists := config.AIVendors["azure_openai"]; exists {
			azure.APIBaseURL = endpoint
			config.AIVendors["azure_openai"] = azure
		}
	}

	// Google overrides
	if key := os.Getenv("GOOGLE_API_KEY"); key != "" {
		if google, exists := config.AIVendors["google"]; exists {
			google.APIKey = key
			config.AIVendors["google"] = google
		}
	}
}

// validateAIConfig validates the configuration
func validateAIConfig(config *AIConfig) error {
	// Check that at least one vendor is enabled
	hasEnabled := false
	for _, vendor := range config.AIVendors {
		if vendor.Enabled {
			hasEnabled = true
			break
		}
	}
	if !hasEnabled {
		return fmt.Errorf("at least one AI vendor must be enabled")
	}

	// Validate use cases reference valid vendors
	for useCaseName, useCase := range config.UseCases {
		if _, exists := config.AIVendors[useCase.Vendor]; !exists {
			return fmt.Errorf("use case '%s' references unknown vendor '%s'", useCaseName, useCase.Vendor)
		}
		vendor := config.AIVendors[useCase.Vendor]
		if !vendor.Enabled {
			return fmt.Errorf("use case '%s' references disabled vendor '%s'", useCaseName, useCase.Vendor)
		}
	}

	return nil
}

// getEnabledVendors returns a comma-separated list of enabled vendors
func getEnabledVendors(config *AIConfig) string {
	var vendors []string
	for name, vendor := range config.AIVendors {
		if vendor.Enabled {
			vendors = append(vendors, name)
		}
	}
	if len(vendors) == 0 {
		return "[none]"
	}
	result := ""
	for i, v := range vendors {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}

// GetAIConfig returns the current AI configuration (thread-safe)
func GetAIConfig() *AIConfig {
	aiConfigMutex.RLock()
	defer aiConfigMutex.RUnlock()
	return aiConfigInstance
}

// ReloadAIConfig reloads the AI configuration
func ReloadAIConfig() error {
	aiConfigMutex.Lock()
	defer aiConfigMutex.Unlock()

	newConfig, err := loadAIConfigInternal()
	if err != nil {
		return err
	}

	aiConfigInstance = newConfig
	AIConf = newConfig
	return nil
}

// GetModelForUseCase returns the model to use for a specific use case
func GetModelForUseCase(useCase string) (vendor string, model string, err error) {
	config := GetAIConfig()
	if config == nil {
		return "", "", fmt.Errorf("AI configuration not loaded")
	}

	useCaseConfig, exists := config.UseCases[useCase]
	if !exists {
		return "", "", fmt.Errorf("unknown use case: %s", useCase)
	}

	vendor = useCaseConfig.Vendor
	vendorConfig, exists := config.AIVendors[vendor]
	if !exists {
		return "", "", fmt.Errorf("vendor %s not found in configuration", vendor)
	}

	// Use model override if specified
	if useCaseConfig.ModelOverride != "" {
		model = useCaseConfig.ModelOverride
	} else {
		// Get model from vendor's use case mapping
		model, exists = vendorConfig.Models[useCase]
		if !exists {
			return "", "", fmt.Errorf("no model configured for use case %s with vendor %s", useCase, vendor)
		}
	}

	return vendor, model, nil
}

// GetVectorDatabaseConfig returns the active vector database configuration
func GetVectorDatabaseConfig() (dbType string, config interface{}, err error) {
	aiConfig := GetAIConfig()
	if aiConfig == nil {
		return "", nil, fmt.Errorf("AI configuration not loaded")
	}

	dbType = aiConfig.VectorDatabase.Type

	// Auto-detect based on what's enabled
	if dbType == "auto" {
		if aiConfig.VectorDatabase.PostgresPGVector.Enabled {
			return "postgres_pgvector", aiConfig.VectorDatabase.PostgresPGVector, nil
		}
		if aiConfig.VectorDatabase.Qdrant.Enabled {
			return "qdrant", aiConfig.VectorDatabase.Qdrant, nil
		}
		if aiConfig.VectorDatabase.Pinecone.Enabled {
			return "pinecone", aiConfig.VectorDatabase.Pinecone, nil
		}
		if aiConfig.VectorDatabase.Weaviate.Enabled {
			return "weaviate", aiConfig.VectorDatabase.Weaviate, nil
		}
		if aiConfig.VectorDatabase.ChromaDB.Enabled {
			return "chromadb", aiConfig.VectorDatabase.ChromaDB, nil
		}
		return "", nil, fmt.Errorf("no vector database enabled")
	}

	// Return specific config based on type
	switch dbType {
	case "postgres_pgvector":
		return dbType, aiConfig.VectorDatabase.PostgresPGVector, nil
	case "qdrant":
		return dbType, aiConfig.VectorDatabase.Qdrant, nil
	case "pinecone":
		return dbType, aiConfig.VectorDatabase.Pinecone, nil
	case "weaviate":
		return dbType, aiConfig.VectorDatabase.Weaviate, nil
	case "chromadb":
		return dbType, aiConfig.VectorDatabase.ChromaDB, nil
	default:
		return "", nil, fmt.Errorf("unknown vector database type: %s", dbType)
	}
}
