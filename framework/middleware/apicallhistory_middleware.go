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

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/framework/auth"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services/apicallhistory"
)

// bodyLogWriter is a wrapper around gin.ResponseWriter that captures the response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body while writing
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// APICallHistoryMiddleware creates a middleware that records API calls
func APICallHistoryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get service
		service := apicallhistory.GetGlobalAPICallHistoryService()
		if service == nil {
			c.Next()
			return
		}

		// Get configuration
		cfg := service.GetConfig()
		if cfg == nil || !cfg.Enabled {
			c.Next()
			return
		}

		// Extract request information early
		method := c.Request.Method
		endpoint := c.Request.URL.Path
		fullPath := c.Request.URL.String()
		sourceIP := c.ClientIP()

		// Get user info (may not be available yet if auth middleware hasn't run)
		userID, userName, clientID, _ := auth.GetUserInformation(c)

		// Check if we should track this call (pre-response check)
		if !cfg.ShouldTrack(method, endpoint, sourceIP, userID, 0) {
			c.Next()
			return
		}

		// Apply sampling rate
		if cfg.SamplingRate < 1.0 && rand.Float64() > cfg.SamplingRate {
			c.Next()
			return
		}

		// Create history record
		history := models.NewAPICallHistory(method, endpoint, fullPath, sourceIP)
		history.UserID = userID
		history.UserName = userName
		history.ClientID = clientID
		history.UserAgent = c.Request.UserAgent()
		history.InstanceName = config.GetInstanceName()

		// Determine auth type
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer") {
				history.AuthType = "bearer"
			} else if strings.HasPrefix(strings.ToLower(authHeader), "apikey") {
				history.AuthType = "apikey"
			}
		} else {
			history.AuthType = "none"
		}

		// Capture request headers
		if cfg.CaptureHeaders {
			headers := make(map[string]string)
			for key, values := range c.Request.Header {
				// Skip sensitive headers
				lowerKey := strings.ToLower(key)
				if lowerKey == "authorization" || lowerKey == "cookie" {
					headers[key] = "[REDACTED]"
				} else {
					headers[key] = strings.Join(values, ", ")
				}
			}
			history.RequestHeaders = headers
		}

		// Capture query parameters
		queryParams := make(map[string]string)
		for key, values := range c.Request.URL.Query() {
			queryParams[key] = strings.Join(values, ", ")
		}
		history.QueryParams = queryParams

		// Capture request body
		if cfg.CaptureRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore body for handlers
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Limit body size
				if len(bodyBytes) > cfg.MaxBodySize {
					bodyBytes = bodyBytes[:cfg.MaxBodySize]
				}

				// Try to parse as JSON and mask sensitive fields
				var bodyData interface{}
				if json.Unmarshal(bodyBytes, &bodyData) == nil {
					bodyData = maskSensitiveData(bodyData, cfg.MaskSensitiveFields)
					history.RequestBody = bodyData
				} else {
					history.RequestBody = string(bodyBytes)
				}
			}
		}

		// Wrap response writer to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Capture response information
		history.StatusCode = c.Writer.Status()
		history.EndTime = time.Now()
		history.DurationMs = history.EndTime.Sub(history.StartTime).Milliseconds()

		// Check status code filters (post-response check)
		if cfg.OnlyErrors && history.StatusCode < 400 {
			return
		}
		if len(cfg.IncludeStatusCodes) > 0 && !containsInt(cfg.IncludeStatusCodes, history.StatusCode) {
			return
		}
		if containsInt(cfg.ExcludeStatusCodes, history.StatusCode) {
			return
		}

		// Capture response headers
		if cfg.CaptureHeaders {
			headers := make(map[string]string)
			for key, values := range c.Writer.Header() {
				headers[key] = strings.Join(values, ", ")
			}
			history.ResponseHeaders = headers
		}

		// Capture response body
		if cfg.CaptureResponseBody {
			responseBody := blw.body.Bytes()
			if len(responseBody) > cfg.MaxBodySize {
				responseBody = responseBody[:cfg.MaxBodySize]
			}

			// Try to parse as JSON and mask sensitive fields
			var bodyData interface{}
			if json.Unmarshal(responseBody, &bodyData) == nil {
				bodyData = maskSensitiveData(bodyData, cfg.MaskSensitiveFields)
				history.ResponseBody = bodyData
			} else {
				history.ResponseBody = string(responseBody)
			}
		}

		// Record error message if status indicates error
		if history.StatusCode >= 400 {
			if errVal, exists := c.Get("error"); exists {
				if errMsg, ok := errVal.(string); ok {
					history.ErrorMessage = errMsg
				}
			}
		}

		// Save asynchronously (non-blocking)
		service.RecordCall(history)
	}
}

// maskSensitiveData masks sensitive fields in the data
func maskSensitiveData(data interface{}, sensitiveFields []string) interface{} {
	if len(sensitiveFields) == 0 {
		return data
	}

	// Create regex patterns for sensitive fields
	patterns := make([]*regexp.Regexp, len(sensitiveFields))
	for i, field := range sensitiveFields {
		patterns[i] = regexp.MustCompile("(?i)" + regexp.QuoteMeta(field))
	}

	return maskDataRecursive(data, patterns)
}

// maskDataRecursive recursively masks sensitive data
func maskDataRecursive(data interface{}, patterns []*regexp.Regexp) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if isSensitiveField(key, patterns) {
				result[key] = "[MASKED]"
			} else {
				result[key] = maskDataRecursive(value, patterns)
			}
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = maskDataRecursive(item, patterns)
		}
		return result
	default:
		return data
	}
}

// isSensitiveField checks if a field name matches any sensitive pattern
func isSensitiveField(fieldName string, patterns []*regexp.Regexp) bool {
	for _, pattern := range patterns {
		if pattern.MatchString(fieldName) {
			return true
		}
	}
	return false
}

// containsInt checks if a slice contains an integer
func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SkipAPICallHistory marks a request to skip API call history recording
func SkipAPICallHistory(c *gin.Context) {
	c.Set("skip_api_history", true)
}

// IsAPICallHistoryEnabled returns true if API call history is enabled
func IsAPICallHistoryEnabled() bool {
	service := apicallhistory.GetGlobalAPICallHistoryService()
	if service == nil {
		return false
	}
	cfg := service.GetConfig()
	return cfg != nil && cfg.Enabled
}
