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

package models

import (
	"time"
)

// APICallHistory represents a single API call record
type APICallHistory struct {
	// Unique identifier
	ID string `json:"id" bson:"_id,omitempty"`

	// Request Information
	Method         string            `json:"method" bson:"method"`
	Endpoint       string            `json:"endpoint" bson:"endpoint"`
	FullPath       string            `json:"full_path" bson:"full_path"`
	RequestHeaders map[string]string `json:"request_headers,omitempty" bson:"request_headers,omitempty"`
	RequestBody    interface{}       `json:"request_body,omitempty" bson:"request_body,omitempty"`
	QueryParams    map[string]string `json:"query_params,omitempty" bson:"query_params,omitempty"`

	// Response Information
	StatusCode      int               `json:"status_code" bson:"status_code"`
	ResponseBody    interface{}       `json:"response_body,omitempty" bson:"response_body,omitempty"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" bson:"response_headers,omitempty"`

	// Source Information
	SourceIP      string `json:"source_ip" bson:"source_ip"`
	SourceMachine string `json:"source_machine,omitempty" bson:"source_machine,omitempty"`
	UserAgent     string `json:"user_agent,omitempty" bson:"user_agent,omitempty"`

	// User Information
	UserID   string `json:"user_id,omitempty" bson:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty" bson:"user_name,omitempty"`
	ClientID string `json:"client_id,omitempty" bson:"client_id,omitempty"`
	AuthType string `json:"auth_type,omitempty" bson:"auth_type,omitempty"` // bearer, apikey, none

	// Timing Information
	StartTime  time.Time `json:"start_time" bson:"start_time"`
	EndTime    time.Time `json:"end_time" bson:"end_time"`
	DurationMs int64     `json:"duration_ms" bson:"duration_ms"`

	// Instance Information
	InstanceID   string `json:"instance_id,omitempty" bson:"instance_id,omitempty"`
	InstanceName string `json:"instance_name,omitempty" bson:"instance_name,omitempty"`

	// Error Information
	ErrorMessage string `json:"error_message,omitempty" bson:"error_message,omitempty"`

	// Metadata
	Tags     []string               `json:"tags,omitempty" bson:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// APICallHistoryConfig represents the configuration for API call tracking
type APICallHistoryConfig struct {
	// Unique identifier
	ID string `json:"id" bson:"_id,omitempty"`

	// Configuration Metadata
	Name        string `json:"name" bson:"name"`               // Configuration name
	Description string `json:"description" bson:"description"` // Configuration description
	IsActive    bool   `json:"is_active" bson:"is_active"`     // Whether this configuration is currently active

	// Master Enable/Disable
	Enabled bool `json:"enabled" bson:"enabled"`

	// Endpoint Filters
	IncludeEndpoints []string `json:"include_endpoints" bson:"include_endpoints"`
	ExcludeEndpoints []string `json:"exclude_endpoints" bson:"exclude_endpoints"`
	IncludeMethods   []string `json:"include_methods" bson:"include_methods"`
	ExcludeMethods   []string `json:"exclude_methods" bson:"exclude_methods"`

	// Source Filters
	IncludeSourceIPs []string `json:"include_source_ips" bson:"include_source_ips"`
	ExcludeSourceIPs []string `json:"exclude_source_ips" bson:"exclude_source_ips"`
	IncludeUsers     []string `json:"include_users" bson:"include_users"`
	ExcludeUsers     []string `json:"exclude_users" bson:"exclude_users"`

	// Status Filters
	IncludeStatusCodes []int `json:"include_status_codes" bson:"include_status_codes"`
	ExcludeStatusCodes []int `json:"exclude_status_codes" bson:"exclude_status_codes"`
	OnlyErrors         bool  `json:"only_errors" bson:"only_errors"`

	// Data Capture Options
	CaptureRequestBody  bool `json:"capture_request_body" bson:"capture_request_body"`
	CaptureResponseBody bool `json:"capture_response_body" bson:"capture_response_body"`
	CaptureHeaders      bool `json:"capture_headers" bson:"capture_headers"`
	MaxBodySize         int  `json:"max_body_size" bson:"max_body_size"`

	// Sensitive Data Handling
	MaskSensitiveFields []string `json:"mask_sensitive_fields" bson:"mask_sensitive_fields"`

	// Retention
	RetentionDays int `json:"retention_days" bson:"retention_days"`

	// Sampling (for high-traffic endpoints)
	SamplingRate float64 `json:"sampling_rate" bson:"sampling_rate"`

	// Refresh interval for configuration updates (in minutes) - 0 means default (5)
	RefreshInterval int `json:"refresh_interval" bson:"refresh_interval"`

	// Audit Information
	UpdatedBy string    `json:"updated_by" bson:"updated_by"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// APICallHistoryListItem represents a summary item for listing
type APICallHistoryListItem struct {
	ID         string    `json:"id"`
	Method     string    `json:"method"`
	Endpoint   string    `json:"endpoint"`
	StatusCode int       `json:"status_code"`
	DurationMs int64     `json:"duration_ms"`
	SourceIP   string    `json:"source_ip"`
	UserID     string    `json:"user_id,omitempty"`
	UserName   string    `json:"user_name,omitempty"`
	StartTime  time.Time `json:"start_time"`
}

// APICallHistoryStats represents statistics for API calls
type APICallHistoryStats struct {
	TotalCalls        int64            `json:"total_calls"`
	SuccessfulCalls   int64            `json:"successful_calls"`
	FailedCalls       int64            `json:"failed_calls"`
	ErrorRate         float64          `json:"error_rate"`
	AverageDurationMs float64          `json:"average_duration_ms"`
	MinDurationMs     int64            `json:"min_duration_ms"`
	MaxDurationMs     int64            `json:"max_duration_ms"`
	CallsByMethod     map[string]int64 `json:"calls_by_method"`
	CallsByStatus     map[string]int64 `json:"calls_by_status"`
	TopEndpoints      []EndpointStat   `json:"top_endpoints"`
	TopUsers          []UserStat       `json:"top_users"`
}

// EndpointStat represents statistics for a specific endpoint
type EndpointStat struct {
	Endpoint      string  `json:"endpoint"`
	TotalCalls    int64   `json:"total_calls"`
	AvgDurationMs float64 `json:"avg_duration_ms"`
	ErrorRate     float64 `json:"error_rate"`
}

// UserStat represents statistics for a specific user
type UserStat struct {
	UserID     string `json:"user_id"`
	TotalCalls int64  `json:"total_calls"`
}

// ListHistoryParams represents parameters for listing history
type ListHistoryParams struct {
	Method     string    `json:"method"`
	Endpoint   string    `json:"endpoint"`
	StatusCode int       `json:"status_code"`
	UserID     string    `json:"user_id"`
	SourceIP   string    `json:"source_ip"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	OnlyErrors bool      `json:"only_errors"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

// ListHistoryResponse represents the response for listing history
type ListHistoryResponse struct {
	History []APICallHistoryListItem `json:"history"`
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	Pages   int64                    `json:"pages"`
}

// NewAPICallHistory creates a new API call history record
func NewAPICallHistory(method, endpoint, fullPath, sourceIP string) *APICallHistory {
	return &APICallHistory{
		Method:    method,
		Endpoint:  endpoint,
		FullPath:  fullPath,
		SourceIP:  sourceIP,
		StartTime: time.Now(),
	}
}

// Complete marks the API call as complete
func (h *APICallHistory) Complete(statusCode int, responseBody interface{}, errorMsg string) {
	h.EndTime = time.Now()
	h.StatusCode = statusCode
	h.ResponseBody = responseBody
	h.ErrorMessage = errorMsg
	h.DurationMs = h.EndTime.Sub(h.StartTime).Milliseconds()
}

// IsError returns true if the status code indicates an error
func (h *APICallHistory) IsError() bool {
	return h.StatusCode >= 400
}

// ToListItem converts APICallHistory to a list item
func (h *APICallHistory) ToListItem() APICallHistoryListItem {
	return APICallHistoryListItem{
		ID:         h.ID,
		Method:     h.Method,
		Endpoint:   h.Endpoint,
		StatusCode: h.StatusCode,
		DurationMs: h.DurationMs,
		SourceIP:   h.SourceIP,
		UserID:     h.UserID,
		UserName:   h.UserName,
		StartTime:  h.StartTime,
	}
}

// NewAPICallHistoryConfig creates a new configuration with default values
func NewAPICallHistoryConfig() *APICallHistoryConfig {
	return &APICallHistoryConfig{
		Enabled:             false,
		IncludeEndpoints:    []string{},
		ExcludeEndpoints:    []string{"/health", "/favicon.ico"},
		IncludeMethods:      []string{},
		ExcludeMethods:      []string{},
		IncludeSourceIPs:    []string{},
		ExcludeSourceIPs:    []string{},
		IncludeUsers:        []string{},
		ExcludeUsers:        []string{},
		IncludeStatusCodes:  []int{},
		ExcludeStatusCodes:  []int{},
		OnlyErrors:          false,
		CaptureRequestBody:  true,
		CaptureResponseBody: true,
		CaptureHeaders:      false,
		MaxBodySize:         10240, // 10KB default
		MaskSensitiveFields: []string{"password", "token", "secret", "apikey", "api_key", "authorization"},
		RetentionDays:       30,
		SamplingRate:        1.0, // 100% by default
		UpdatedAt:           time.Now(),
	}
}

// ShouldTrack determines if an API call should be tracked based on the configuration
func (c *APICallHistoryConfig) ShouldTrack(method, endpoint, sourceIP, userID string, statusCode int) bool {
	if !c.Enabled {
		return false
	}

	// Check method filters
	if len(c.IncludeMethods) > 0 && !contains(c.IncludeMethods, method) {
		return false
	}
	if contains(c.ExcludeMethods, method) {
		return false
	}

	// Check endpoint filters
	if len(c.IncludeEndpoints) > 0 && !matchesAny(c.IncludeEndpoints, endpoint) {
		return false
	}
	if matchesAny(c.ExcludeEndpoints, endpoint) {
		return false
	}

	// Check source IP filters
	if len(c.IncludeSourceIPs) > 0 && !contains(c.IncludeSourceIPs, sourceIP) {
		return false
	}
	if contains(c.ExcludeSourceIPs, sourceIP) {
		return false
	}

	// Check user filters
	if len(c.IncludeUsers) > 0 && !contains(c.IncludeUsers, userID) {
		return false
	}
	if contains(c.ExcludeUsers, userID) {
		return false
	}

	// Check status code filters (only applicable after response)
	if statusCode > 0 {
		if c.OnlyErrors && statusCode < 400 {
			return false
		}
		if len(c.IncludeStatusCodes) > 0 && !containsInt(c.IncludeStatusCodes, statusCode) {
			return false
		}
		if containsInt(c.ExcludeStatusCodes, statusCode) {
			return false
		}
	}

	return true
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInt(slice []int, item int) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func matchesAny(patterns []string, value string) bool {
	for _, pattern := range patterns {
		if matchPattern(pattern, value) {
			return true
		}
	}
	return false
}

// matchPattern performs simple wildcard matching (* matches any sequence)
func matchPattern(pattern, value string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == value {
		return true
	}
	// Simple prefix matching with *
	if len(pattern) > 1 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(value) >= len(prefix) && value[:len(prefix)] == prefix
	}
	// Simple suffix matching with *
	if len(pattern) > 1 && pattern[0] == '*' {
		suffix := pattern[1:]
		return len(value) >= len(suffix) && value[len(value)-len(suffix):] == suffix
	}
	return false
}
