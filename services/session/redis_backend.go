package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// RedisSessionBackend implements SessionBackend using Redis
type RedisSessionBackend struct {
	pool       *redis.Pool
	config     *SessionBackendConfig
	keyPrefix  string
	log        logger.Log
}

// NewRedisSessionBackend creates a new Redis session backend
func NewRedisSessionBackend(config *SessionBackendConfig) (*RedisSessionBackend, error) {
	if config.RedisURL == "" {
		config.RedisURL = "localhost:6379"
	}
	if config.RedisKeyPrefix == "" {
		config.RedisKeyPrefix = "iac:session"
	}
	if config.RedisPoolSize <= 0 {
		config.RedisPoolSize = 10
	}

	pool := &redis.Pool{
		MaxIdle:     config.RedisPoolSize,
		MaxActive:   config.RedisPoolSize * 2,
		IdleTimeout: 180 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.RedisURL,
				redis.DialConnectTimeout(5*time.Second),
				redis.DialReadTimeout(3*time.Second),
				redis.DialWriteTimeout(3*time.Second),
			)
			if err != nil {
				return nil, err
			}

			if config.RedisPassword != "" {
				if _, err := c.Do("AUTH", config.RedisPassword); err != nil {
					c.Close()
					return nil, err
				}
			}

			if config.RedisDB > 0 {
				if _, err := c.Do("SELECT", config.RedisDB); err != nil {
					c.Close()
					return nil, err
				}
			}

			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	backend := &RedisSessionBackend{
		pool:      pool,
		config:    config,
		keyPrefix: config.RedisKeyPrefix,
		log:       logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "RedisSessionBackend"},
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := backend.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	backend.log.Info("Redis session backend initialized successfully")
	return backend, nil
}

// Key helpers
func (r *RedisSessionBackend) sessionKey(sessionID string) string {
	return fmt.Sprintf("%s:%s", r.keyPrefix, sessionID)
}

func (r *RedisSessionBackend) userSessionsKey(userID string) string {
	return fmt.Sprintf("%s:user:%s", r.keyPrefix, userID)
}

func (r *RedisSessionBackend) tokenKey(tokenHash string) string {
	return fmt.Sprintf("%s:token:%s", r.keyPrefix, tokenHash)
}

func (r *RedisSessionBackend) lockKey(sessionID string) string {
	return fmt.Sprintf("%s:lock:%s", r.keyPrefix, sessionID)
}

// Get retrieves a session by ID
func (r *RedisSessionBackend) Get(ctx context.Context, sessionID string) (*models.DistributedSession, error) {
	conn := r.pool.Get()
	defer conn.Close()

	data, err := redis.Bytes(conn.Do("GET", r.sessionKey(sessionID)))
	if err != nil {
		if err == redis.ErrNil {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session models.DistributedSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	if session.IsExpired() {
		// Clean up expired session
		go r.Delete(context.Background(), sessionID)
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// Set stores or updates a session
func (r *RedisSessionBackend) Set(ctx context.Context, session *models.DistributedSession) error {
	conn := r.pool.Get()
	defer conn.Close()

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = r.config.TTL
	}

	// Store session data
	_, err = conn.Do("SETEX", r.sessionKey(session.SessionID), int(ttl.Seconds()), data)
	if err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	// Add to user's session set
	_, err = conn.Do("SADD", r.userSessionsKey(session.UserID), session.SessionID)
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to add session to user set: %v", err))
	}

	// Set expiry on user sessions set
	_, err = conn.Do("EXPIRE", r.userSessionsKey(session.UserID), int(ttl.Seconds())+3600)
	if err != nil {
		r.log.Error(fmt.Sprintf("Failed to set expiry on user sessions set: %v", err))
	}

	// Store token mapping if token hash exists
	if session.TokenHash != "" {
		_, err = conn.Do("SETEX", r.tokenKey(session.TokenHash), int(ttl.Seconds()), session.SessionID)
		if err != nil {
			r.log.Error(fmt.Sprintf("Failed to set token mapping: %v", err))
		}
	}

	return nil
}

// Delete removes a session
func (r *RedisSessionBackend) Delete(ctx context.Context, sessionID string) error {
	conn := r.pool.Get()
	defer conn.Close()

	// Get session first to clean up related keys
	session, err := r.Get(ctx, sessionID)
	if err != nil && err != ErrSessionNotFound && err != ErrSessionExpired {
		return err
	}

	// Delete session
	_, err = conn.Do("DEL", r.sessionKey(sessionID))
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Clean up related keys if session was found
	if session != nil {
		conn.Do("SREM", r.userSessionsKey(session.UserID), sessionID)
		if session.TokenHash != "" {
			conn.Do("DEL", r.tokenKey(session.TokenHash))
		}
	}

	return nil
}

// Exists checks if a session exists
func (r *RedisSessionBackend) Exists(ctx context.Context, sessionID string) (bool, error) {
	conn := r.pool.Get()
	defer conn.Close()

	exists, err := redis.Bool(conn.Do("EXISTS", r.sessionKey(sessionID)))
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return exists, nil
}

// Extend extends the session TTL
func (r *RedisSessionBackend) Extend(ctx context.Context, sessionID string, ttl time.Duration) error {
	session, err := r.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Extend(ttl)
	return r.Set(ctx, session)
}

// GetByUserID retrieves all sessions for a user
func (r *RedisSessionBackend) GetByUserID(ctx context.Context, userID string) ([]*models.DistributedSession, error) {
	conn := r.pool.Get()
	defer conn.Close()

	sessionIDs, err := redis.Strings(conn.Do("SMEMBERS", r.userSessionsKey(userID)))
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}

	var sessions []*models.DistributedSession
	for _, sessionID := range sessionIDs {
		session, err := r.Get(ctx, sessionID)
		if err != nil {
			if err == ErrSessionNotFound || err == ErrSessionExpired {
				// Clean up stale reference
				conn.Do("SREM", r.userSessionsKey(userID), sessionID)
				continue
			}
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// DeleteByUserID deletes all sessions for a user
func (r *RedisSessionBackend) DeleteByUserID(ctx context.Context, userID string) error {
	sessions, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if err := r.Delete(ctx, session.SessionID); err != nil {
			r.log.Error(fmt.Sprintf("Failed to delete session %s: %v", session.SessionID, err))
		}
	}

	conn := r.pool.Get()
	defer conn.Close()

	// Delete user sessions set
	conn.Do("DEL", r.userSessionsKey(userID))

	return nil
}

// GetByToken retrieves a session by token hash
func (r *RedisSessionBackend) GetByToken(ctx context.Context, tokenHash string) (*models.DistributedSession, error) {
	conn := r.pool.Get()
	defer conn.Close()

	sessionID, err := redis.String(conn.Do("GET", r.tokenKey(tokenHash)))
	if err != nil {
		if err == redis.ErrNil {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return r.Get(ctx, sessionID)
}

// Lock acquires a distributed lock for a session
func (r *RedisSessionBackend) Lock(ctx context.Context, sessionID string, ttl time.Duration) (func(), error) {
	conn := r.pool.Get()
	defer conn.Close()

	lockKey := r.lockKey(sessionID)
	lockValue := fmt.Sprintf("%s:%d", r.config.InstanceID, time.Now().UnixNano())

	// Try to acquire lock with NX (only if not exists) and PX (expiry in milliseconds)
	result, err := redis.String(conn.Do("SET", lockKey, lockValue, "NX", "PX", int(ttl.Milliseconds())))
	if err != nil {
		if err == redis.ErrNil {
			return nil, ErrSessionLockFailed
		}
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if result != "OK" {
		return nil, ErrSessionLockFailed
	}

	// Return unlock function
	unlock := func() {
		c := r.pool.Get()
		defer c.Close()

		// Only delete if we still own the lock
		script := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`
		c.Do("EVAL", script, 1, lockKey, lockValue)
	}

	return unlock, nil
}

// Ping checks if Redis is available
func (r *RedisSessionBackend) Ping(ctx context.Context) error {
	conn := r.pool.Get()
	defer conn.Close()

	_, err := conn.Do("PING")
	if err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// Close closes the Redis connection pool
func (r *RedisSessionBackend) Close() error {
	return r.pool.Close()
}

// Name returns the backend name
func (r *RedisSessionBackend) Name() string {
	return "redis"
}
