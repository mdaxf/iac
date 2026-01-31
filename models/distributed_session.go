package models

import (
	"time"
)

// DistributedSession represents a session shareable across IAC instances
type DistributedSession struct {
	SessionID     string                 `json:"session_id" bson:"session_id"`
	UserID        string                 `json:"user_id" bson:"user_id"`
	Username      string                 `json:"username" bson:"username"`
	LoginName     string                 `json:"login_name" bson:"login_name"`
	ClientID      string                 `json:"client_id" bson:"client_id"`

	// Token information
	TokenHash     string                 `json:"token_hash" bson:"token_hash"`
	TokenExpiry   time.Time              `json:"token_expiry" bson:"token_expiry"`
	RefreshToken  string                 `json:"refresh_token,omitempty" bson:"refresh_token,omitempty"`

	// Instance tracking
	CreatedInstance string               `json:"created_instance" bson:"created_instance"`
	LastInstance    string               `json:"last_instance" bson:"last_instance"`

	// User permissions and roles
	Roles         []string               `json:"roles,omitempty" bson:"roles,omitempty"`
	Permissions   []string               `json:"permissions,omitempty" bson:"permissions,omitempty"`

	// Session data (arbitrary key-value storage)
	Data          map[string]interface{} `json:"data,omitempty" bson:"data,omitempty"`

	// Timing
	CreatedAt     time.Time              `json:"created_at" bson:"created_at"`
	LastAccessedAt time.Time             `json:"last_accessed_at" bson:"last_accessed_at"`
	ExpiresAt     time.Time              `json:"expires_at" bson:"expires_at"`

	// Client metadata
	UserAgent     string                 `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
	IPAddress     string                 `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	Language      string                 `json:"language,omitempty" bson:"language,omitempty"`
	Timezone      string                 `json:"timezone,omitempty" bson:"timezone,omitempty"`
}

// SessionMigration tracks session failover/migration between instances
type SessionMigration struct {
	SessionID    string    `json:"session_id" bson:"session_id"`
	FromInstance string    `json:"from_instance" bson:"from_instance"`
	ToInstance   string    `json:"to_instance" bson:"to_instance"`
	Reason       string    `json:"reason" bson:"reason"` // failover, load_balance, manual
	MigratedAt   time.Time `json:"migrated_at" bson:"migrated_at"`
	Success      bool      `json:"success" bson:"success"`
}

// SessionStatus represents the current status of a session
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusExpired  SessionStatus = "expired"
	SessionStatusRevoked  SessionStatus = "revoked"
	SessionStatusMigrated SessionStatus = "migrated"
)

// IsExpired checks if the session has expired
func (s *DistributedSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsTokenExpired checks if the token has expired
func (s *DistributedSession) IsTokenExpired() bool {
	return time.Now().After(s.TokenExpiry)
}

// Extend extends the session expiry time
func (s *DistributedSession) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
	s.LastAccessedAt = time.Now()
}

// ExtendToken extends the token expiry time
func (s *DistributedSession) ExtendToken(duration time.Duration) {
	s.TokenExpiry = time.Now().Add(duration)
}

// UpdateLastAccess updates the last accessed time and instance
func (s *DistributedSession) UpdateLastAccess(instanceID string) {
	s.LastAccessedAt = time.Now()
	s.LastInstance = instanceID
}

// SetData sets a key-value pair in the session data
func (s *DistributedSession) SetData(key string, value interface{}) {
	if s.Data == nil {
		s.Data = make(map[string]interface{})
	}
	s.Data[key] = value
}

// GetData retrieves a value from the session data
func (s *DistributedSession) GetData(key string) (interface{}, bool) {
	if s.Data == nil {
		return nil, false
	}
	val, ok := s.Data[key]
	return val, ok
}

// ToUserInfo converts session to a map compatible with existing user info structure
func (s *DistributedSession) ToUserInfo() map[string]interface{} {
	return map[string]interface{}{
		"user_id":      s.UserID,
		"username":     s.Username,
		"login_name":   s.LoginName,
		"client_id":    s.ClientID,
		"session_id":   s.SessionID,
		"roles":        s.Roles,
		"permissions":  s.Permissions,
		"language":     s.Language,
		"timezone":     s.Timezone,
		"token_expiry": s.TokenExpiry,
		"expires_at":   s.ExpiresAt,
	}
}

// NewDistributedSession creates a new distributed session
func NewDistributedSession(userID, username, loginName, clientID, instanceID string, ttl time.Duration) *DistributedSession {
	now := time.Now()
	return &DistributedSession{
		SessionID:       generateSessionID(),
		UserID:          userID,
		Username:        username,
		LoginName:       loginName,
		ClientID:        clientID,
		CreatedInstance: instanceID,
		LastInstance:    instanceID,
		Data:            make(map[string]interface{}),
		CreatedAt:       now,
		LastAccessedAt:  now,
		ExpiresAt:       now.Add(ttl),
		TokenExpiry:     now.Add(ttl),
	}
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(16)
}

// randomString generates a random string of specified length
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
