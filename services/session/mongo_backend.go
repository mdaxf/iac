package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/mdaxf/iac/com"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// MongoSessionBackend implements SessionBackend using MongoDB
type MongoSessionBackend struct {
	collection *mongo.Collection
	config     *SessionBackendConfig
	log        logger.Log
	locks      map[string]*sync.Mutex
	locksMutex sync.RWMutex
}

// NewMongoSessionBackend creates a new MongoDB session backend
func NewMongoSessionBackend(config *SessionBackendConfig) (*MongoSessionBackend, error) {
	if config.MongoCollection == "" {
		config.MongoCollection = "iac_sessions"
	}

	// Get MongoDB client from global connection
	if com.IACDocDBConn.MongoDBClient == nil {
		return nil, fmt.Errorf("MongoDB client not available")
	}

	db := com.IACDocDBConn.MongoDBClient.Database(com.IACDocDBConn.DatabaseName)
	collection := db.Collection(config.MongoCollection)

	backend := &MongoSessionBackend{
		collection: collection,
		config:     config,
		log:        logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MongoSessionBackend"},
		locks:      make(map[string]*sync.Mutex),
	}

	// Create indexes
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Index on session_id (unique)
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "session_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		backend.log.Error(fmt.Sprintf("Failed to create session_id index: %v", err))
	}

	// Index on user_id
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	if err != nil {
		backend.log.Error(fmt.Sprintf("Failed to create user_id index: %v", err))
	}

	// Index on token_hash
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "token_hash", Value: 1}},
	})
	if err != nil {
		backend.log.Error(fmt.Sprintf("Failed to create token_hash index: %v", err))
	}

	// TTL index on expires_at for automatic cleanup
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		backend.log.Error(fmt.Sprintf("Failed to create TTL index: %v", err))
	}

	backend.log.Info("MongoDB session backend initialized successfully")
	return backend, nil
}

// Get retrieves a session by ID
func (m *MongoSessionBackend) Get(ctx context.Context, sessionID string) (*models.DistributedSession, error) {
	var session models.DistributedSession

	err := m.collection.FindOne(ctx, bson.M{"session_id": sessionID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.IsExpired() {
		// Clean up expired session
		go m.Delete(context.Background(), sessionID)
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// Set stores or updates a session
func (m *MongoSessionBackend) Set(ctx context.Context, session *models.DistributedSession) error {
	filter := bson.M{"session_id": session.SessionID}
	update := bson.M{"$set": session}
	opts := options.Update().SetUpsert(true)

	_, err := m.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to set session: %w", err)
	}

	return nil
}

// Delete removes a session
func (m *MongoSessionBackend) Delete(ctx context.Context, sessionID string) error {
	_, err := m.collection.DeleteOne(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// Exists checks if a session exists
func (m *MongoSessionBackend) Exists(ctx context.Context, sessionID string) (bool, error) {
	count, err := m.collection.CountDocuments(ctx, bson.M{"session_id": sessionID})
	if err != nil {
		return false, fmt.Errorf("failed to check session existence: %w", err)
	}

	return count > 0, nil
}

// Extend extends the session TTL
func (m *MongoSessionBackend) Extend(ctx context.Context, sessionID string, ttl time.Duration) error {
	session, err := m.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Extend(ttl)
	return m.Set(ctx, session)
}

// GetByUserID retrieves all sessions for a user
func (m *MongoSessionBackend) GetByUserID(ctx context.Context, userID string) ([]*models.DistributedSession, error) {
	cursor, err := m.collection.Find(ctx, bson.M{
		"user_id": userID,
		"expires_at": bson.M{"$gt": time.Now()},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer cursor.Close(ctx)

	var sessions []*models.DistributedSession
	for cursor.Next(ctx) {
		var session models.DistributedSession
		if err := cursor.Decode(&session); err != nil {
			m.log.Error(fmt.Sprintf("Failed to decode session: %v", err))
			continue
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// DeleteByUserID deletes all sessions for a user
func (m *MongoSessionBackend) DeleteByUserID(ctx context.Context, userID string) error {
	_, err := m.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// GetByToken retrieves a session by token hash
func (m *MongoSessionBackend) GetByToken(ctx context.Context, tokenHash string) (*models.DistributedSession, error) {
	var session models.DistributedSession

	err := m.collection.FindOne(ctx, bson.M{
		"token_hash": tokenHash,
		"expires_at": bson.M{"$gt": time.Now()},
	}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}

	return &session, nil
}

// Lock acquires a distributed lock for a session (using in-memory locks for MongoDB backend)
// Note: For true distributed locking with MongoDB, consider using a separate locks collection
// with findAndModify operations or a dedicated locking service
func (m *MongoSessionBackend) Lock(ctx context.Context, sessionID string, ttl time.Duration) (func(), error) {
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

// Ping checks if MongoDB is available
func (m *MongoSessionBackend) Ping(ctx context.Context) error {
	return com.IACDocDBConn.MongoDBClient.Ping(ctx, nil)
}

// Close closes the backend (MongoDB connection is managed globally)
func (m *MongoSessionBackend) Close() error {
	// MongoDB connection is managed globally, no need to close here
	return nil
}

// Name returns the backend name
func (m *MongoSessionBackend) Name() string {
	return "mongodb"
}
