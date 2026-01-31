package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// MemorySessionBackend implements SessionBackend using in-memory storage
// This is suitable for single-instance deployments or development/testing
type MemorySessionBackend struct {
	sessions   map[string]*models.DistributedSession
	userIndex  map[string][]string // userID -> []sessionID
	tokenIndex map[string]string   // tokenHash -> sessionID
	locks      map[string]*sync.Mutex
	config     *SessionBackendConfig
	log        logger.Log
	mutex      sync.RWMutex
	locksMutex sync.RWMutex
	stopCh     chan struct{}
}

// NewMemorySessionBackend creates a new in-memory session backend
func NewMemorySessionBackend(config *SessionBackendConfig) (*MemorySessionBackend, error) {
	backend := &MemorySessionBackend{
		sessions:   make(map[string]*models.DistributedSession),
		userIndex:  make(map[string][]string),
		tokenIndex: make(map[string]string),
		locks:      make(map[string]*sync.Mutex),
		config:     config,
		log:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MemorySessionBackend"},
		stopCh:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go backend.cleanupLoop()

	backend.log.Info("Memory session backend initialized successfully")
	return backend, nil
}

// cleanupLoop periodically removes expired sessions
func (m *MemorySessionBackend) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stopCh:
			return
		}
	}
}

// cleanup removes expired sessions
func (m *MemorySessionBackend) cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for sessionID, session := range m.sessions {
		if now.After(session.ExpiresAt) {
			// Remove from user index
			m.removeFromUserIndex(session.UserID, sessionID)

			// Remove from token index
			if session.TokenHash != "" {
				delete(m.tokenIndex, session.TokenHash)
			}

			// Remove session
			delete(m.sessions, sessionID)
			expiredCount++
		}
	}

	if expiredCount > 0 {
		m.log.Debug(fmt.Sprintf("Cleaned up %d expired sessions", expiredCount))
	}
}

// removeFromUserIndex removes a session from user's session list
func (m *MemorySessionBackend) removeFromUserIndex(userID, sessionID string) {
	sessions := m.userIndex[userID]
	for i, id := range sessions {
		if id == sessionID {
			m.userIndex[userID] = append(sessions[:i], sessions[i+1:]...)
			break
		}
	}
	if len(m.userIndex[userID]) == 0 {
		delete(m.userIndex, userID)
	}
}

// Get retrieves a session by ID
func (m *MemorySessionBackend) Get(ctx context.Context, sessionID string) (*models.DistributedSession, error) {
	m.mutex.RLock()
	session, exists := m.sessions[sessionID]
	m.mutex.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	if session.IsExpired() {
		// Clean up expired session asynchronously
		go m.Delete(context.Background(), sessionID)
		return nil, ErrSessionExpired
	}

	// Return a copy to prevent mutation
	sessionCopy := *session
	return &sessionCopy, nil
}

// Set stores or updates a session
func (m *MemorySessionBackend) Set(ctx context.Context, session *models.DistributedSession) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if session already exists
	existing, exists := m.sessions[session.SessionID]

	// Store session copy
	sessionCopy := *session
	m.sessions[session.SessionID] = &sessionCopy

	// Update user index
	if !exists {
		m.userIndex[session.UserID] = append(m.userIndex[session.UserID], session.SessionID)
	} else if existing.UserID != session.UserID {
		// User changed (unlikely but handle it)
		m.removeFromUserIndex(existing.UserID, session.SessionID)
		m.userIndex[session.UserID] = append(m.userIndex[session.UserID], session.SessionID)
	}

	// Update token index
	if session.TokenHash != "" {
		// Remove old token mapping if different
		if exists && existing.TokenHash != "" && existing.TokenHash != session.TokenHash {
			delete(m.tokenIndex, existing.TokenHash)
		}
		m.tokenIndex[session.TokenHash] = session.SessionID
	}

	return nil
}

// Delete removes a session
func (m *MemorySessionBackend) Delete(ctx context.Context, sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil // Not an error if doesn't exist
	}

	// Remove from user index
	m.removeFromUserIndex(session.UserID, sessionID)

	// Remove from token index
	if session.TokenHash != "" {
		delete(m.tokenIndex, session.TokenHash)
	}

	// Remove session
	delete(m.sessions, sessionID)

	return nil
}

// Exists checks if a session exists
func (m *MemorySessionBackend) Exists(ctx context.Context, sessionID string) (bool, error) {
	m.mutex.RLock()
	session, exists := m.sessions[sessionID]
	m.mutex.RUnlock()

	if !exists {
		return false, nil
	}

	// Check expiry
	if session.IsExpired() {
		return false, nil
	}

	return true, nil
}

// Extend extends the session TTL
func (m *MemorySessionBackend) Extend(ctx context.Context, sessionID string, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	if session.IsExpired() {
		return ErrSessionExpired
	}

	session.Extend(ttl)
	return nil
}

// GetByUserID retrieves all sessions for a user
func (m *MemorySessionBackend) GetByUserID(ctx context.Context, userID string) ([]*models.DistributedSession, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessionIDs := m.userIndex[userID]
	var sessions []*models.DistributedSession

	for _, sessionID := range sessionIDs {
		session, exists := m.sessions[sessionID]
		if exists && !session.IsExpired() {
			sessionCopy := *session
			sessions = append(sessions, &sessionCopy)
		}
	}

	return sessions, nil
}

// DeleteByUserID deletes all sessions for a user
func (m *MemorySessionBackend) DeleteByUserID(ctx context.Context, userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessionIDs := m.userIndex[userID]
	for _, sessionID := range sessionIDs {
		session, exists := m.sessions[sessionID]
		if exists {
			// Remove from token index
			if session.TokenHash != "" {
				delete(m.tokenIndex, session.TokenHash)
			}
			delete(m.sessions, sessionID)
		}
	}

	// Remove user index
	delete(m.userIndex, userID)

	return nil
}

// GetByToken retrieves a session by token hash
func (m *MemorySessionBackend) GetByToken(ctx context.Context, tokenHash string) (*models.DistributedSession, error) {
	m.mutex.RLock()
	sessionID, exists := m.tokenIndex[tokenHash]
	m.mutex.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	return m.Get(ctx, sessionID)
}

// Lock acquires a lock for a session
func (m *MemorySessionBackend) Lock(ctx context.Context, sessionID string, ttl time.Duration) (func(), error) {
	m.locksMutex.Lock()
	lock, exists := m.locks[sessionID]
	if !exists {
		lock = &sync.Mutex{}
		m.locks[sessionID] = lock
	}
	m.locksMutex.Unlock()

	// Try to acquire lock with timeout
	done := make(chan struct{})
	go func() {
		lock.Lock()
		close(done)
	}()

	select {
	case <-done:
		// Lock acquired
		unlock := func() {
			lock.Unlock()
		}
		return unlock, nil
	case <-time.After(ttl):
		return nil, ErrSessionLockTimeout
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Ping checks if the backend is available (always true for memory)
func (m *MemorySessionBackend) Ping(ctx context.Context) error {
	return nil
}

// Close stops the cleanup goroutine
func (m *MemorySessionBackend) Close() error {
	close(m.stopCh)
	return nil
}

// Name returns the backend name
func (m *MemorySessionBackend) Name() string {
	return "memory"
}

// Stats returns statistics about the session store (for monitoring)
func (m *MemorySessionBackend) Stats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"total_sessions": len(m.sessions),
		"total_users":    len(m.userIndex),
		"total_tokens":   len(m.tokenIndex),
	}
}
