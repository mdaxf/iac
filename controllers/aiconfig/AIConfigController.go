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

package aiconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
)

type AIConfigController struct{}

// GetAIConfig retrieves the current AI configuration
// GET /api/ai-config
func (c *AIConfigController) GetAIConfig(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIConfigController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIConfigController.GetAIConfig", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Get AI configuration
	aiConfig := config.GetAIConfig()
	if aiConfig == nil {
		iLog.Error(fmt.Sprintf("AI configuration not loaded"))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "AI configuration not available"})
		return
	}

	// Mask sensitive data (API keys) before sending to client
	maskedConfig := maskSensitiveData(aiConfig)

	ctx.JSON(http.StatusOK, maskedConfig)

	iLog.Info(fmt.Sprintf("AI configuration retrieved successfully"))
}

// UpdateAIConfig updates the AI configuration
// PUT /api/ai-config
func (c *AIConfigController) UpdateAIConfig(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIConfigController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIConfigController.UpdateAIConfig", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse request body
	var newConfig config.AIConfig
	if err := ctx.ShouldBindJSON(&newConfig); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Write to aiconfig.local.json
	configData, err := json.MarshalIndent(newConfig, "", "  ")
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to marshal config: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	if err := ioutil.WriteFile("aiconfig.local.json", configData, 0644); err != nil {
		iLog.Error(fmt.Sprintf("Failed to write config file: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	// Reload configuration
	if err := config.ReloadAIConfig(); err != nil {
		iLog.Error(fmt.Sprintf("Failed to reload config: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload configuration"})
		return
	}

	// Get updated config and mask sensitive data
	updatedConfig := config.GetAIConfig()
	maskedConfig := maskSensitiveData(updatedConfig)

	ctx.JSON(http.StatusOK, maskedConfig)

	iLog.Info(fmt.Sprintf("AI configuration updated successfully"))
}

// TestConnection tests connection to an AI vendor
// POST /api/ai-config/test-connection
func (c *AIConfigController) TestConnection(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIConfigController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIConfigController.TestConnection", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Parse request body
	var req struct {
		Vendor string                  `json:"vendor"`
		Config config.AIVendorConfig   `json:"config"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse request body: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Test connection based on vendor
	success, errorMsg := testVendorConnection(req.Vendor, &req.Config)

	response := map[string]interface{}{
		"success": success,
		"vendor":  req.Vendor,
	}
	if !success {
		response["error"] = errorMsg
	}

	ctx.JSON(http.StatusOK, response)

	if success {
		iLog.Info(fmt.Sprintf("Connection test successful for vendor: %s", req.Vendor))
	} else {
		iLog.Warn(fmt.Sprintf("Connection test failed for vendor: %s, error: %s", req.Vendor, errorMsg))
	}
}

// GetVectorDBStatus gets the status of the configured vector database
// GET /api/ai-config/vector-db-status
func (c *AIConfigController) GetVectorDBStatus(ctx *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "AIConfigController"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("AIConfigController.GetVectorDBStatus", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(ctx)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	dbType, dbConfig, err := config.GetVectorDatabaseConfig()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to get vector DB config: %v", err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"type":      dbType,
		"config":    dbConfig,
		"connected": true, // TODO: Implement actual connection check
	}

	ctx.JSON(http.StatusOK, response)

	iLog.Info(fmt.Sprintf("Vector DB status retrieved: %s", dbType))
}

// Helper function to mask sensitive data
func maskSensitiveData(config *config.AIConfig) *config.AIConfig {
	if config == nil {
		return nil
	}

	// Create a copy of the config
	maskedConfig := *config

	// Mask API keys in vendors
	for vendor, vendorConfig := range maskedConfig.AIVendors {
		if vendorConfig.APIKey != "" {
			vendorConfig.APIKey = maskAPIKey(vendorConfig.APIKey)
			maskedConfig.AIVendors[vendor] = vendorConfig
		}
	}

	// Mask vector database credentials
	if maskedConfig.VectorDatabase.PostgresPGVector.ConnectionString != "" {
		maskedConfig.VectorDatabase.PostgresPGVector.ConnectionString = "***masked***"
	}
	if maskedConfig.VectorDatabase.Qdrant.APIKey != "" {
		maskedConfig.VectorDatabase.Qdrant.APIKey = maskAPIKey(maskedConfig.VectorDatabase.Qdrant.APIKey)
	}
	if maskedConfig.VectorDatabase.Pinecone.APIKey != "" {
		maskedConfig.VectorDatabase.Pinecone.APIKey = maskAPIKey(maskedConfig.VectorDatabase.Pinecone.APIKey)
	}
	if maskedConfig.VectorDatabase.Weaviate.APIKey != "" {
		maskedConfig.VectorDatabase.Weaviate.APIKey = maskAPIKey(maskedConfig.VectorDatabase.Weaviate.APIKey)
	}

	return &maskedConfig
}

// Helper function to mask API key
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// Helper function to test vendor connection
func testVendorConnection(vendor string, vendorConfig *config.AIVendorConfig) (bool, string) {
	if vendorConfig.APIKey == "" {
		return false, "API key is required"
	}

	switch vendor {
	case "openai":
		return testOpenAIConnection(vendorConfig)
	case "anthropic":
		return testAnthropicConnection(vendorConfig)
	case "azure_openai":
		return testAzureOpenAIConnection(vendorConfig)
	case "google":
		return testGoogleConnection(vendorConfig)
	case "ollama":
		return testOllamaConnection(vendorConfig)
	default:
		return false, fmt.Sprintf("Unknown vendor: %s", vendor)
	}
}

// Test OpenAI connection
func testOpenAIConnection(vendorConfig *config.AIVendorConfig) (bool, string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", vendorConfig.APIBaseURL+"/models", nil)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+vendorConfig.APIKey)
	if vendorConfig.Organization != "" {
		req.Header.Set("OpenAI-Organization", vendorConfig.Organization)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return false, fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return true, ""
}

// Test Anthropic connection
func testAnthropicConnection(vendorConfig *config.AIVendorConfig) (bool, string) {
	// Simple check: verify API key format
	if len(vendorConfig.APIKey) < 10 || vendorConfig.APIKey[:7] != "sk-ant-" {
		return false, "Invalid Anthropic API key format"
	}

	// TODO: Implement actual API call to test connection
	return true, ""
}

// Test Azure OpenAI connection
func testAzureOpenAIConnection(vendorConfig *config.AIVendorConfig) (bool, string) {
	if vendorConfig.Deployment == "" {
		return false, "Deployment name is required for Azure OpenAI"
	}

	// TODO: Implement actual API call to test connection
	return true, ""
}

// Test Google connection
func testGoogleConnection(vendorConfig *config.AIVendorConfig) (bool, string) {
	if vendorConfig.ProjectID == "" {
		return false, "Project ID is required for Google"
	}

	// TODO: Implement actual API call to test connection
	return true, ""
}

// Test Ollama connection
func testOllamaConnection(vendorConfig *config.AIVendorConfig) (bool, string) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", vendorConfig.APIBaseURL+"/api/tags", nil)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("Ollama server returned status %d", resp.StatusCode)
	}

	return true, ""
}
