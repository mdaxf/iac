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

package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"github.com/mdaxf/iac/services/session"
)

var log = logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DistributedAuth"}

// CreateDistributedSession creates a new distributed session for a user
func CreateDistributedSession(userID, username, loginName, clientID string, c *gin.Context) (*models.DistributedSession, error) {
	sessionService := session.GetGlobalSessionService()
	if sessionService == nil || !config.IsDistributedSessionEnabled() {
		// Distributed sessions not enabled, return nil (caller should use regular JWT)
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create new session
	sess, err := sessionService.CreateSession(ctx, userID, username, loginName, clientID)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to create distributed session: %v", err))
		return nil, err
	}

	// Set additional metadata from request
	if c != nil {
		sess.UserAgent = c.GetHeader("User-Agent")
		sess.IPAddress = c.ClientIP()
		sess.Language = c.GetHeader("Accept-Language")
	}

	// Update session with metadata
	if err := sessionService.UpdateSession(ctx, sess); err != nil {
		log.Error(fmt.Sprintf("Failed to update session metadata: %v", err))
	}

	log.Debug(fmt.Sprintf("Created distributed session %s for user %s", sess.SessionID, userID))
	return sess, nil
}

// SetSessionToken associates a JWT token with the distributed session
func SetSessionToken(sessionID, token string) error {
	sessionService := session.GetGlobalSessionService()
	if sessionService == nil || !config.IsDistributedSessionEnabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sessionService.SetToken(ctx, sessionID, token)
}

// ValidateDistributedSession validates a session from the distributed store
func ValidateDistributedSession(sessionID string) (*models.DistributedSession, error) {
	sessionService := session.GetGlobalSessionService()
	if sessionService == nil || !config.IsDistributedSessionEnabled() {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sessionService.ValidateSession(ctx, sessionID)
}

// ValidateDistributedToken validates a token from the distributed store
func ValidateDistributedToken(token string) (*models.DistributedSession, error) {
	sessionService := session.GetGlobalSessionService()
	if sessionService == nil || !config.IsDistributedSessionEnabled() {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sessionService.ValidateToken(ctx, token)
}

// DeleteDistributedSession removes a session from the distributed store
func DeleteDistributedSession(sessionID string) error {
	sessionService := session.GetGlobalSessionService()
	if sessionService == nil || !config.IsDistributedSessionEnabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sessionService.DeleteSession(ctx, sessionID)
}

// DeleteUserDistributedSessions removes all sessions for a user (logout everywhere)
func DeleteUserDistributedSessions(userID string) error {
	sessionService := session.GetGlobalSessionService()
	if sessionService == nil || !config.IsDistributedSessionEnabled() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return sessionService.DeleteUserSessions(ctx, userID)
}

// GetUserDistributedSessions returns all active sessions for a user
func GetUserDistributedSessions(userID string) ([]*models.DistributedSession, error) {
	sessionService := session.GetGlobalSessionService()
	if sessionService == nil || !config.IsDistributedSessionEnabled() {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sessionService.GetUserSessions(ctx, userID)
}

// HashToken creates a SHA256 hash of a token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// DistributedAuthMiddleware returns a middleware that validates distributed sessions
// This should be used in addition to or instead of the regular AuthMiddleware when
// distributed sessions are enabled
func DistributedAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip certain paths
		if shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get the token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || (strings.ToLower(bearerToken[0]) != "bearer" && strings.ToLower(bearerToken[0]) != "apikey") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		authType := strings.ToLower(bearerToken[0])
		tokenString := bearerToken[1]

		// Handle API key authentication
		if authType == "apikey" {
			if tokenString != config.ApiKey {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid APIKey"})
				return
			}
			c.Next()
			return
		}

		// First, validate the JWT token locally
		ok, err := ValidateToken(tokenString)
		if err != nil || !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		// If distributed sessions are enabled, also check the distributed session
		if config.IsDistributedSessionEnabled() {
			sess, err := ValidateDistributedToken(tokenString)
			if err != nil {
				log.Error(fmt.Sprintf("Failed to validate distributed session: %v", err))
				// Allow the request if it's just a backend issue
				// The JWT is still valid
				if err == session.ErrSessionNotFound || err == session.ErrSessionExpired {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session not found or expired"})
					return
				}
			}

			if sess != nil {
				// Set session info in context
				c.Set("distributed_session", sess)
				c.Set("session_id", sess.SessionID)
				c.Set("user_id", sess.UserID)
				c.Set("username", sess.Username)
				c.Set("login_name", sess.LoginName)
				c.Set("client_id", sess.ClientID)
				c.Set("roles", sess.Roles)
				c.Set("permissions", sess.Permissions)
			}
		}

		c.Next()
	}
}

// shouldSkipAuth checks if a path should skip authentication
func shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/favicon.ico",
		"/user/login",
		"/user/changepwd",
		"/health",
		"/roles/status",
	}

	for _, skip := range skipPaths {
		if path == skip {
			return true
		}
	}

	// Check path prefixes
	skipPrefixes := []string{
		"/user/image",
		"/portal",
		"/static",
	}

	for _, prefix := range skipPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// GetSessionFromContext retrieves the distributed session from the gin context
func GetSessionFromContext(c *gin.Context) *models.DistributedSession {
	sess, exists := c.Get("distributed_session")
	if !exists {
		return nil
	}
	if s, ok := sess.(*models.DistributedSession); ok {
		return s
	}
	return nil
}

// SetSessionData sets a value in the distributed session
func SetSessionData(c *gin.Context, key string, value interface{}) error {
	sess := GetSessionFromContext(c)
	if sess == nil {
		return fmt.Errorf("no distributed session in context")
	}

	sessionService := session.GetGlobalSessionService()
	if sessionService == nil {
		return fmt.Errorf("session service not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sessionService.SetSessionData(ctx, sess.SessionID, key, value)
}

// GetSessionData retrieves a value from the distributed session
func GetSessionData(c *gin.Context, key string) (interface{}, error) {
	sess := GetSessionFromContext(c)
	if sess == nil {
		return nil, fmt.Errorf("no distributed session in context")
	}

	sessionService := session.GetGlobalSessionService()
	if sessionService == nil {
		return nil, fmt.Errorf("session service not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sessionService.GetSessionData(ctx, sess.SessionID, key)
}
