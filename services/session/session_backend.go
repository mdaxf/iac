package session

import (
	"context"
	"errors"
	"time"

	"github.com/mdaxf/iac/models"
)

// Common errors for session operations
var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionLockFailed   = errors.New("failed to acquire session lock")
	ErrSessionLockTimeout  = errors.New("session lock timeout")
	ErrBackendUnavailable  = errors.New("session backend unavailable")
	ErrInvalidSession      = errors.New("invalid session data")
)

// SessionBackend defines the interface for distributed session storage
type SessionBackend interface {
	// Get retrieves a session by ID
	Get(ctx context.Context, sessionID string) (*models.DistributedSession, error)

	// Set stores or updates a session
	Set(ctx context.Context, session *models.DistributedSession) error

	// Delete removes a session
	Delete(ctx context.Context, sessionID string) error

	// Exists checks if a session exists
	Exists(ctx context.Context, sessionID string) (bool, error)

	// Extend extends the session TTL
	Extend(ctx context.Context, sessionID string, ttl time.Duration) error

	// GetByUserID retrieves all sessions for a user
	GetByUserID(ctx context.Context, userID string) ([]*models.DistributedSession, error)

	// DeleteByUserID deletes all sessions for a user
	DeleteByUserID(ctx context.Context, userID string) error

	// GetByToken retrieves a session by token hash
	GetByToken(ctx context.Context, tokenHash string) (*models.DistributedSession, error)

	// Lock acquires a distributed lock for a session (returns unlock function)
	Lock(ctx context.Context, sessionID string, ttl time.Duration) (func(), error)

	// Ping checks if the backend is available
	Ping(ctx context.Context) error

	// Close closes the backend connection
	Close() error

	// Name returns the backend name for logging
	Name() string
}

// SessionBackendConfig holds configuration for session backends
type SessionBackendConfig struct {
	// Common settings
	TTL     time.Duration `json:"ttl"`
	LockTTL time.Duration `json:"lock_ttl"`

	// Redis settings
	RedisURL      string `json:"redis_url"`
	RedisDB       int    `json:"redis_db"`
	RedisPassword string `json:"redis_password"`
	RedisPoolSize int    `json:"redis_pool_size"`
	RedisKeyPrefix string `json:"redis_key_prefix"`

	// MongoDB settings
	MongoCollection string `json:"mongo_collection"`
	MongoDatabase   string `json:"mongo_database"`

	// Instance identification
	InstanceID string `json:"instance_id"`
}

// DefaultSessionBackendConfig returns default configuration
func DefaultSessionBackendConfig() *SessionBackendConfig {
	return &SessionBackendConfig{
		TTL:             30 * time.Minute,
		LockTTL:         30 * time.Second,
		RedisDB:         0,
		RedisPoolSize:   10,
		RedisKeyPrefix:  "iac:session",
		MongoCollection: "iac_sessions",
	}
}

// SessionBackendType represents the type of session backend
type SessionBackendType string

const (
	BackendTypeRedis   SessionBackendType = "redis"
	BackendTypeMongoDB SessionBackendType = "mongodb"
	BackendTypeMemory  SessionBackendType = "memory"
)

// SessionBackendFactory creates session backends based on type
type SessionBackendFactory struct {
	config *SessionBackendConfig
}

// NewSessionBackendFactory creates a new factory
func NewSessionBackendFactory(config *SessionBackendConfig) *SessionBackendFactory {
	if config == nil {
		config = DefaultSessionBackendConfig()
	}
	return &SessionBackendFactory{config: config}
}

// Create creates a session backend of the specified type
func (f *SessionBackendFactory) Create(backendType SessionBackendType) (SessionBackend, error) {
	switch backendType {
	case BackendTypeRedis:
		return NewRedisSessionBackend(f.config)
	case BackendTypeMongoDB:
		return NewMongoSessionBackend(f.config)
	case BackendTypeMemory:
		return NewMemorySessionBackend(f.config)
	default:
		return nil, errors.New("unsupported session backend type: " + string(backendType))
	}
}

// CreateWithFallback creates a primary backend with a fallback
func (f *SessionBackendFactory) CreateWithFallback(primary, fallback SessionBackendType) (SessionBackend, error) {
	primaryBackend, err := f.Create(primary)
	if err != nil {
		return nil, err
	}

	// Try to ping the primary backend
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := primaryBackend.Ping(ctx); err != nil {
		// Primary unavailable, try fallback
		primaryBackend.Close()
		return f.Create(fallback)
	}

	return primaryBackend, nil
}

// FallbackSessionBackend wraps a primary backend with automatic fallback
type FallbackSessionBackend struct {
	primary  SessionBackend
	fallback SessionBackend
	config   *SessionBackendConfig
}

// NewFallbackSessionBackend creates a backend with fallback support
func NewFallbackSessionBackend(primary, fallback SessionBackend, config *SessionBackendConfig) *FallbackSessionBackend {
	return &FallbackSessionBackend{
		primary:  primary,
		fallback: fallback,
		config:   config,
	}
}

func (f *FallbackSessionBackend) getActiveBackend(ctx context.Context) SessionBackend {
	if f.primary != nil {
		if err := f.primary.Ping(ctx); err == nil {
			return f.primary
		}
	}
	return f.fallback
}

func (f *FallbackSessionBackend) Get(ctx context.Context, sessionID string) (*models.DistributedSession, error) {
	return f.getActiveBackend(ctx).Get(ctx, sessionID)
}

func (f *FallbackSessionBackend) Set(ctx context.Context, session *models.DistributedSession) error {
	return f.getActiveBackend(ctx).Set(ctx, session)
}

func (f *FallbackSessionBackend) Delete(ctx context.Context, sessionID string) error {
	return f.getActiveBackend(ctx).Delete(ctx, sessionID)
}

func (f *FallbackSessionBackend) Exists(ctx context.Context, sessionID string) (bool, error) {
	return f.getActiveBackend(ctx).Exists(ctx, sessionID)
}

func (f *FallbackSessionBackend) Extend(ctx context.Context, sessionID string, ttl time.Duration) error {
	return f.getActiveBackend(ctx).Extend(ctx, sessionID, ttl)
}

func (f *FallbackSessionBackend) GetByUserID(ctx context.Context, userID string) ([]*models.DistributedSession, error) {
	return f.getActiveBackend(ctx).GetByUserID(ctx, userID)
}

func (f *FallbackSessionBackend) DeleteByUserID(ctx context.Context, userID string) error {
	return f.getActiveBackend(ctx).DeleteByUserID(ctx, userID)
}

func (f *FallbackSessionBackend) GetByToken(ctx context.Context, tokenHash string) (*models.DistributedSession, error) {
	return f.getActiveBackend(ctx).GetByToken(ctx, tokenHash)
}

func (f *FallbackSessionBackend) Lock(ctx context.Context, sessionID string, ttl time.Duration) (func(), error) {
	return f.getActiveBackend(ctx).Lock(ctx, sessionID, ttl)
}

func (f *FallbackSessionBackend) Ping(ctx context.Context) error {
	if f.primary != nil {
		if err := f.primary.Ping(ctx); err == nil {
			return nil
		}
	}
	if f.fallback != nil {
		return f.fallback.Ping(ctx)
	}
	return ErrBackendUnavailable
}

func (f *FallbackSessionBackend) Close() error {
	var errs []error
	if f.primary != nil {
		if err := f.primary.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if f.fallback != nil {
		if err := f.fallback.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (f *FallbackSessionBackend) Name() string {
	return "fallback(" + f.primary.Name() + "/" + f.fallback.Name() + ")"
}
