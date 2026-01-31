package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// DistributedSessionService manages sessions across distributed IAC instances
type DistributedSessionService struct {
	backend      SessionBackend
	config       *DistributedSessionConfig
	instanceID   string
	log          logger.Log
	mutex        sync.RWMutex
	initialized  bool
}

// DistributedSessionConfig holds configuration for distributed session management
type DistributedSessionConfig struct {
	// Enabled determines if distributed sessions are active
	Enabled bool `json:"enabled"`

	// BackendType specifies the primary backend (redis, mongodb, memory)
	BackendType SessionBackendType `json:"backend_type"`

	// FallbackType specifies the fallback backend if primary fails
	FallbackType SessionBackendType `json:"fallback_type"`

	// SessionTTL is the default session duration
	SessionTTL time.Duration `json:"session_ttl"`

	// TokenTTL is the default token duration
	TokenTTL time.Duration `json:"token_ttl"`

	// LockTTL is the timeout for acquiring session locks
	LockTTL time.Duration `json:"lock_ttl"`

	// AutoExtend determines if sessions should be auto-extended on access
	AutoExtend bool `json:"auto_extend"`

	// ExtendThreshold is the remaining time before auto-extending
	ExtendThreshold time.Duration `json:"extend_threshold"`

	// InstanceID identifies this IAC instance
	InstanceID string `json:"instance_id"`

	// Backend-specific configuration
	BackendConfig *SessionBackendConfig `json:"backend_config"`
}

// DefaultDistributedSessionConfig returns default configuration
func DefaultDistributedSessionConfig() *DistributedSessionConfig {
	return &DistributedSessionConfig{
		Enabled:         false,
		BackendType:     BackendTypeMemory,
		FallbackType:    BackendTypeMemory,
		SessionTTL:      30 * time.Minute,
		TokenTTL:        30 * time.Minute,
		LockTTL:         30 * time.Second,
		AutoExtend:      true,
		ExtendThreshold: 5 * time.Minute,
		InstanceID:      "default",
		BackendConfig:   DefaultSessionBackendConfig(),
	}
}

// Global service instance
var (
	globalSessionService *DistributedSessionService
	globalServiceMutex   sync.RWMutex
)

// GetGlobalSessionService returns the global session service instance
func GetGlobalSessionService() *DistributedSessionService {
	globalServiceMutex.RLock()
	defer globalServiceMutex.RUnlock()
	return globalSessionService
}

// SetGlobalSessionService sets the global session service instance
func SetGlobalSessionService(service *DistributedSessionService) {
	globalServiceMutex.Lock()
	defer globalServiceMutex.Unlock()
	globalSessionService = service
}

// NewDistributedSessionService creates a new distributed session service
func NewDistributedSessionService(config *DistributedSessionConfig) (*DistributedSessionService, error) {
	if config == nil {
		config = DefaultDistributedSessionConfig()
	}

	if config.BackendConfig == nil {
		config.BackendConfig = DefaultSessionBackendConfig()
	}

	// Copy common settings to backend config
	config.BackendConfig.TTL = config.SessionTTL
	config.BackendConfig.LockTTL = config.LockTTL
	config.BackendConfig.InstanceID = config.InstanceID

	service := &DistributedSessionService{
		config:     config,
		instanceID: config.InstanceID,
		log:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "DistributedSessionService"},
	}

	// Create backend based on configuration
	factory := NewSessionBackendFactory(config.BackendConfig)

	var backend SessionBackend
	var err error

	if config.FallbackType != "" && config.FallbackType != config.BackendType {
		// Create with fallback
		backend, err = factory.CreateWithFallback(config.BackendType, config.FallbackType)
	} else {
		// Create primary only
		backend, err = factory.Create(config.BackendType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create session backend: %w", err)
	}

	service.backend = backend
	service.initialized = true

	service.log.Info(fmt.Sprintf("Distributed session service initialized with backend: %s", backend.Name()))

	return service, nil
}

// CreateSession creates a new session for a user
func (s *DistributedSessionService) CreateSession(ctx context.Context, userID, username, loginName, clientID string) (*models.DistributedSession, error) {
	if !s.initialized {
		return nil, fmt.Errorf("session service not initialized")
	}

	// Create new session
	session := models.NewDistributedSession(
		userID,
		username,
		loginName,
		clientID,
		s.instanceID,
		s.config.SessionTTL,
	)

	// Set token expiry
	session.TokenExpiry = time.Now().Add(s.config.TokenTTL)

	// Store session
	if err := s.backend.Set(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	s.log.Debug(fmt.Sprintf("Created session %s for user %s", session.SessionID, userID))

	return session, nil
}

// GetSession retrieves a session by ID
func (s *DistributedSessionService) GetSession(ctx context.Context, sessionID string) (*models.DistributedSession, error) {
	if !s.initialized {
		return nil, fmt.Errorf("session service not initialized")
	}

	session, err := s.backend.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Auto-extend if enabled and near expiry
	if s.config.AutoExtend {
		remaining := time.Until(session.ExpiresAt)
		if remaining > 0 && remaining < s.config.ExtendThreshold {
			session.Extend(s.config.SessionTTL)
			session.UpdateLastAccess(s.instanceID)
			if err := s.backend.Set(ctx, session); err != nil {
				s.log.Error(fmt.Sprintf("Failed to auto-extend session %s: %v", sessionID, err))
			}
		}
	}

	return session, nil
}

// ValidateSession validates a session and returns user info
func (s *DistributedSessionService) ValidateSession(ctx context.Context, sessionID string) (*models.DistributedSession, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsExpired() {
		return nil, ErrSessionExpired
	}

	// Update last access
	session.UpdateLastAccess(s.instanceID)
	go s.backend.Set(context.Background(), session)

	return session, nil
}

// UpdateSession updates a session
func (s *DistributedSessionService) UpdateSession(ctx context.Context, session *models.DistributedSession) error {
	if !s.initialized {
		return fmt.Errorf("session service not initialized")
	}

	session.UpdateLastAccess(s.instanceID)
	return s.backend.Set(ctx, session)
}

// DeleteSession removes a session
func (s *DistributedSessionService) DeleteSession(ctx context.Context, sessionID string) error {
	if !s.initialized {
		return fmt.Errorf("session service not initialized")
	}

	return s.backend.Delete(ctx, sessionID)
}

// GetUserSessions retrieves all sessions for a user
func (s *DistributedSessionService) GetUserSessions(ctx context.Context, userID string) ([]*models.DistributedSession, error) {
	if !s.initialized {
		return nil, fmt.Errorf("session service not initialized")
	}

	return s.backend.GetByUserID(ctx, userID)
}

// DeleteUserSessions removes all sessions for a user (logout everywhere)
func (s *DistributedSessionService) DeleteUserSessions(ctx context.Context, userID string) error {
	if !s.initialized {
		return fmt.Errorf("session service not initialized")
	}

	return s.backend.DeleteByUserID(ctx, userID)
}

// SetToken associates a token hash with a session
func (s *DistributedSessionService) SetToken(ctx context.Context, sessionID, token string) error {
	session, err := s.backend.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Hash the token
	hash := sha256.Sum256([]byte(token))
	session.TokenHash = hex.EncodeToString(hash[:])
	session.TokenExpiry = time.Now().Add(s.config.TokenTTL)

	return s.backend.Set(ctx, session)
}

// ValidateToken validates a token and returns the session
func (s *DistributedSessionService) ValidateToken(ctx context.Context, token string) (*models.DistributedSession, error) {
	if !s.initialized {
		return nil, fmt.Errorf("session service not initialized")
	}

	// Hash the token
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	session, err := s.backend.GetByToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if session.IsTokenExpired() {
		return nil, ErrSessionExpired
	}

	// Update last access
	session.UpdateLastAccess(s.instanceID)
	go s.backend.Set(context.Background(), session)

	return session, nil
}

// ExtendSession extends a session's TTL
func (s *DistributedSessionService) ExtendSession(ctx context.Context, sessionID string, duration time.Duration) error {
	if !s.initialized {
		return fmt.Errorf("session service not initialized")
	}

	if duration == 0 {
		duration = s.config.SessionTTL
	}

	return s.backend.Extend(ctx, sessionID, duration)
}

// LockSession acquires a distributed lock for a session
func (s *DistributedSessionService) LockSession(ctx context.Context, sessionID string) (func(), error) {
	if !s.initialized {
		return nil, fmt.Errorf("session service not initialized")
	}

	return s.backend.Lock(ctx, sessionID, s.config.LockTTL)
}

// SetSessionData sets a key-value pair in the session data
func (s *DistributedSessionService) SetSessionData(ctx context.Context, sessionID, key string, value interface{}) error {
	unlock, err := s.LockSession(ctx, sessionID)
	if err != nil {
		return err
	}
	defer unlock()

	session, err := s.backend.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.SetData(key, value)
	session.UpdateLastAccess(s.instanceID)

	return s.backend.Set(ctx, session)
}

// GetSessionData retrieves a value from the session data
func (s *DistributedSessionService) GetSessionData(ctx context.Context, sessionID, key string) (interface{}, error) {
	session, err := s.backend.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	value, ok := session.GetData(key)
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return value, nil
}

// MigrateSession records a session migration between instances
func (s *DistributedSessionService) MigrateSession(ctx context.Context, sessionID, fromInstance, reason string) error {
	session, err := s.backend.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	// Update instance info
	session.LastInstance = s.instanceID
	session.UpdateLastAccess(s.instanceID)

	// Store migration info in session data
	migration := models.SessionMigration{
		SessionID:    sessionID,
		FromInstance: fromInstance,
		ToInstance:   s.instanceID,
		Reason:       reason,
		MigratedAt:   time.Now(),
		Success:      true,
	}
	session.SetData("last_migration", migration)

	return s.backend.Set(ctx, session)
}

// Ping checks if the session service is healthy
func (s *DistributedSessionService) Ping(ctx context.Context) error {
	if !s.initialized {
		return fmt.Errorf("session service not initialized")
	}

	return s.backend.Ping(ctx)
}

// Close shuts down the session service
func (s *DistributedSessionService) Close() error {
	if s.backend != nil {
		return s.backend.Close()
	}
	return nil
}

// GetBackendName returns the name of the current backend
func (s *DistributedSessionService) GetBackendName() string {
	if s.backend != nil {
		return s.backend.Name()
	}
	return "none"
}

// IsEnabled returns whether distributed sessions are enabled
func (s *DistributedSessionService) IsEnabled() bool {
	return s.initialized && s.config.Enabled
}

// GetInstanceID returns the current instance ID
func (s *DistributedSessionService) GetInstanceID() string {
	return s.instanceID
}

// Stats returns statistics about the session service
func (s *DistributedSessionService) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":      s.config.Enabled,
		"backend":      s.GetBackendName(),
		"instance_id":  s.instanceID,
		"session_ttl":  s.config.SessionTTL.String(),
		"token_ttl":    s.config.TokenTTL.String(),
		"auto_extend":  s.config.AutoExtend,
		"initialized":  s.initialized,
	}

	// Add backend-specific stats if available
	if memBackend, ok := s.backend.(*MemorySessionBackend); ok {
		for k, v := range memBackend.Stats() {
			stats[k] = v
		}
	}

	return stats
}
