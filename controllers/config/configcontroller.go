package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	//"os"
	"path/filepath"
	"time"

	"github.com/mdaxf/iac/logger"

	"github.com/gin-gonic/gin"
)

type ConfigController struct{}

// GetConfiguration retrieves the main configuration.json
func (cc *ConfigController) GetConfiguration(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "GetConfiguration"}
	iLog.Info("Getting configuration.json")

	configPath := "./configuration.json"
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to read configuration.json: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read configuration", "details": err.Error()})
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse configuration.json: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse configuration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
		"path":   configPath,
	})
}

// UpdateConfiguration updates the main configuration.json
func (cc *ConfigController) UpdateConfiguration(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "UpdateConfiguration"}
	iLog.Info("Updating configuration.json")

	var request struct {
		Config map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	configPath := "./configuration.json"

	// Create backup before updating
	backupPath := fmt.Sprintf("./configuration.backup.%s.json", time.Now().Format("20060102-150405"))
	if err := cc.createBackup(configPath, backupPath); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to create backup: %v", err))
	}

	// Marshal the new configuration
	data, err := json.MarshalIndent(request.Config, "", "    ")
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to marshal configuration: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal configuration", "details": err.Error()})
		return
	}

	// Write the new configuration
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		iLog.Error(fmt.Sprintf("Failed to write configuration: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write configuration", "details": err.Error()})
		return
	}

	iLog.Info("Configuration updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":    "Configuration updated successfully",
		"backupPath": backupPath,
	})
}

// GetAPIConfig retrieves the apiconfig.json
func (cc *ConfigController) GetAPIConfig(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "GetAPIConfig"}
	iLog.Info("Getting apiconfig.json")

	configPath := "./apiconfig.json"
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to read apiconfig.json: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read API configuration", "details": err.Error()})
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse apiconfig.json: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse API configuration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"config": config,
		"path":   configPath,
	})
}

// UpdateAPIConfig updates the apiconfig.json
func (cc *ConfigController) UpdateAPIConfig(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "UpdateAPIConfig"}
	iLog.Info("Updating apiconfig.json")

	var request struct {
		Config map[string]interface{} `json:"config"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	configPath := "./apiconfig.json"

	// Create backup before updating
	backupPath := fmt.Sprintf("./apiconfig.backup.%s.json", time.Now().Format("20060102-150405"))
	if err := cc.createBackup(configPath, backupPath); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to create backup: %v", err))
	}

	// Marshal the new configuration
	data, err := json.MarshalIndent(request.Config, "", "  ")
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to marshal API configuration: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal API configuration", "details": err.Error()})
		return
	}

	// Write the new configuration
	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		iLog.Error(fmt.Sprintf("Failed to write API configuration: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write API configuration", "details": err.Error()})
		return
	}

	iLog.Info("API Configuration updated successfully")
	c.JSON(http.StatusOK, gin.H{
		"message":    "API Configuration updated successfully",
		"backupPath": backupPath,
	})
}

// GetWebServerConfig retrieves web server specific configuration
func (cc *ConfigController) GetWebServerConfig(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "GetWebServerConfig"}
	iLog.Info("Getting web server configuration")

	configPath := "./configuration.json"
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to read configuration.json: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read configuration", "details": err.Error()})
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		iLog.Error(fmt.Sprintf("Failed to parse configuration.json: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse configuration", "details": err.Error()})
		return
	}

	// Extract web server specific config
	webserverConfig := config["webserver"]

	c.JSON(http.StatusOK, gin.H{
		"config": webserverConfig,
	})
}

// ListBackups lists all backup configuration files
func (cc *ConfigController) ListBackups(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "ListBackups"}
	iLog.Info("Listing backup files")

	currentDir := "."
	files, err := ioutil.ReadDir(currentDir)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to read directory: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory", "details": err.Error()})
		return
	}

	var backups []map[string]interface{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		// Look for backup files
		if filepath.Ext(name) == ".json" && (filepath.Base(name)[:len("configuration.backup")] == "configuration.backup" || filepath.Base(name)[:len("apiconfig.backup")] == "apiconfig.backup") {
			backups = append(backups, map[string]interface{}{
				"name":    name,
				"size":    file.Size(),
				"modTime": file.ModTime().Format("2006-01-02 15:04:05"),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"backups": backups,
	})
}

// ValidateConfiguration validates the configuration format
func (cc *ConfigController) ValidateConfiguration(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "ValidateConfiguration"}
	iLog.Info("Validating configuration")

	var request struct {
		Config map[string]interface{} `json:"config"`
		Type   string                 `json:"type"` // "main" or "api"
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	errors := []string{}

	if request.Type == "main" {
		// Validate main configuration required fields
		requiredFields := []string{"instance", "name", "version", "database", "documentdb", "webserver"}
		for _, field := range requiredFields {
			if _, exists := request.Config[field]; !exists {
				errors = append(errors, fmt.Sprintf("Missing required field: %s", field))
			}
		}
	} else if request.Type == "api" {
		// Validate API configuration required fields
		requiredFields := []string{"port", "timeout", "controllers"}
		for _, field := range requiredFields {
			if _, exists := request.Config[field]; !exists {
				errors = append(errors, fmt.Sprintf("Missing required field: %s", field))
			}
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":  false,
			"errors": errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Configuration is valid",
	})
}

// Helper function to create backup
func (cc *ConfigController) createBackup(sourcePath, backupPath string) error {
	data, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(backupPath, data, 0644)
}

// RestoreBackup restores configuration from a backup file
func (cc *ConfigController) RestoreBackup(c *gin.Context) {
	iLog := logger.Log{ModuleName: "ConfigController", ControllerName: "RestoreBackup"}

	var request struct {
		BackupFile string `json:"backupFile"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		iLog.Error(fmt.Sprintf("Invalid request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Determine target file based on backup filename
	var targetPath string
	if filepath.Base(request.BackupFile)[:len("configuration.backup")] == "configuration.backup" {
		targetPath = "./configuration.json"
	} else if filepath.Base(request.BackupFile)[:len("apiconfig.backup")] == "apiconfig.backup" {
		targetPath = "./apiconfig.json"
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup file name"})
		return
	}

	// Read backup file
	data, err := ioutil.ReadFile(request.BackupFile)
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to read backup file: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read backup file", "details": err.Error()})
		return
	}

	// Create a backup of current configuration before restoring
	currentBackupPath := fmt.Sprintf("%s.before-restore.%s.json", filepath.Base(targetPath), time.Now().Format("20060102-150405"))
	if err := cc.createBackup(targetPath, currentBackupPath); err != nil {
		iLog.Warn(fmt.Sprintf("Failed to create pre-restore backup: %v", err))
	}

	// Write backup data to target
	if err := ioutil.WriteFile(targetPath, data, 0644); err != nil {
		iLog.Error(fmt.Sprintf("Failed to restore configuration: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore configuration", "details": err.Error()})
		return
	}

	iLog.Info(fmt.Sprintf("Configuration restored from %s", request.BackupFile))
	c.JSON(http.StatusOK, gin.H{
		"message":  "Configuration restored successfully",
		"restored": request.BackupFile,
		"target":   targetPath,
	})
}
