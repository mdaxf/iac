package debug

import (
	"context"
	"sync"
	"time"
)

// MessageBus manages event distribution to multiple subscribers
type MessageBus struct {
	subscribers map[string][]*Subscriber
	mu          sync.RWMutex
	buffer      int
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewMessageBus creates a new message bus
func NewMessageBus(bufferSize int) *MessageBus {
	ctx, cancel := context.WithCancel(context.Background())
	return &MessageBus{
		subscribers: make(map[string][]*Subscriber),
		buffer:      bufferSize,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Subscriber represents a client subscribing to debug events
type Subscriber struct {
	ID           string
	SessionID    string
	EventChannel chan *DebugEvent
	Created      time.Time
	LastActivity time.Time
	Filters      *EventFilter
	ctx          context.Context
	cancel       context.CancelFunc
}

// EventFilter defines what events a subscriber wants to receive
type EventFilter struct {
	EventTypes    []EventType // Only receive these event types
	MinLevel      string      // Minimum log level (DEBUG, INFO, WARNING, ERROR)
	TranCodeNames []string    // Only events from these trancodes
	FunctionTypes []string    // Only events from these function types
}

// NewSubscriber creates a new subscriber
func NewSubscriber(id, sessionID string, bufferSize int) *Subscriber {
	ctx, cancel := context.WithCancel(context.Background())
	return &Subscriber{
		ID:           id,
		SessionID:    sessionID,
		EventChannel: make(chan *DebugEvent, bufferSize),
		Created:      time.Now(),
		LastActivity: time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Close closes the subscriber
func (s *Subscriber) Close() {
	s.cancel()
	close(s.EventChannel)
}

// Subscribe subscribes a client to a debug session
func (mb *MessageBus) Subscribe(sessionID string, subscriber *Subscriber) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.subscribers[sessionID] == nil {
		mb.subscribers[sessionID] = make([]*Subscriber, 0)
	}

	mb.subscribers[sessionID] = append(mb.subscribers[sessionID], subscriber)
}

// Unsubscribe removes a subscriber from a session
func (mb *MessageBus) Unsubscribe(sessionID, subscriberID string) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	subscribers := mb.subscribers[sessionID]
	if subscribers == nil {
		return
	}

	// Find and remove the subscriber
	for i, sub := range subscribers {
		if sub.ID == subscriberID {
			sub.Close()
			mb.subscribers[sessionID] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}

	// Clean up empty session
	if len(mb.subscribers[sessionID]) == 0 {
		delete(mb.subscribers, sessionID)
	}
}

// Publish publishes an event to all subscribers of a session
func (mb *MessageBus) Publish(event *DebugEvent) {
	mb.mu.RLock()
	subscribers := mb.subscribers[event.SessionID]
	mb.mu.RUnlock()

	if subscribers == nil {
		return
	}

	// Send to all subscribers in parallel
	var wg sync.WaitGroup
	for _, sub := range subscribers {
		wg.Add(1)
		go func(s *Subscriber) {
			defer wg.Done()
			mb.sendToSubscriber(s, event)
		}(sub)
	}

	wg.Wait()
}

// sendToSubscriber sends an event to a specific subscriber
func (mb *MessageBus) sendToSubscriber(subscriber *Subscriber, event *DebugEvent) {
	// Update last activity
	subscriber.LastActivity = time.Now()

	// Apply filters
	if !mb.matchesFilter(subscriber.Filters, event) {
		return
	}

	// Try to send with timeout to prevent blocking
	select {
	case subscriber.EventChannel <- event:
		// Event sent successfully
	case <-time.After(100 * time.Millisecond):
		// Timeout - subscriber is slow, skip this event
	case <-subscriber.ctx.Done():
		// Subscriber closed
	}
}

// matchesFilter checks if an event matches the subscriber's filter
func (mb *MessageBus) matchesFilter(filter *EventFilter, event *DebugEvent) bool {
	if filter == nil {
		return true // No filter, accept all
	}

	// Check event types
	if len(filter.EventTypes) > 0 {
		matched := false
		for _, et := range filter.EventTypes {
			if et == event.EventType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check log level
	if filter.MinLevel != "" {
		if !meetsMinLevel(event.Level, filter.MinLevel) {
			return false
		}
	}

	// Check trancode names
	if len(filter.TranCodeNames) > 0 {
		matched := false
		for _, name := range filter.TranCodeNames {
			if name == event.TranCodeName {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check function types
	if len(filter.FunctionTypes) > 0 {
		matched := false
		for _, ft := range filter.FunctionTypes {
			if ft == event.FunctionType {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

// meetsMinLevel checks if event level meets minimum level requirement
func meetsMinLevel(eventLevel, minLevel string) bool {
	levels := map[string]int{
		"DEBUG":   0,
		"INFO":    1,
		"WARNING": 2,
		"ERROR":   3,
	}

	eventLevelValue, ok1 := levels[eventLevel]
	minLevelValue, ok2 := levels[minLevel]

	if !ok1 || !ok2 {
		return true // If unknown level, allow
	}

	return eventLevelValue >= minLevelValue
}

// GetSubscribers returns all subscribers for a session
func (mb *MessageBus) GetSubscribers(sessionID string) []*Subscriber {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	return mb.subscribers[sessionID]
}

// GetSessionCount returns the number of active debug sessions
func (mb *MessageBus) GetSessionCount() int {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	return len(mb.subscribers)
}

// GetSubscriberCount returns the total number of active subscribers
func (mb *MessageBus) GetSubscriberCount() int {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	count := 0
	for _, subs := range mb.subscribers {
		count += len(subs)
	}

	return count
}

// CleanupInactiveSubscribers removes subscribers that haven't been active
func (mb *MessageBus) CleanupInactiveSubscribers(timeout time.Duration) int {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	removed := 0
	now := time.Now()

	for sessionID, subscribers := range mb.subscribers {
		activeSubscribers := make([]*Subscriber, 0)

		for _, sub := range subscribers {
			if now.Sub(sub.LastActivity) > timeout {
				sub.Close()
				removed++
			} else {
				activeSubscribers = append(activeSubscribers, sub)
			}
		}

		if len(activeSubscribers) > 0 {
			mb.subscribers[sessionID] = activeSubscribers
		} else {
			delete(mb.subscribers, sessionID)
		}
	}

	return removed
}

// Shutdown shuts down the message bus
func (mb *MessageBus) Shutdown() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// Close all subscribers
	for _, subscribers := range mb.subscribers {
		for _, sub := range subscribers {
			sub.Close()
		}
	}

	mb.subscribers = make(map[string][]*Subscriber)
	mb.cancel()
}

// StartCleanupRoutine starts a background routine to clean up inactive subscribers
func (mb *MessageBus) StartCleanupRoutine(interval, timeout time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				removed := mb.CleanupInactiveSubscribers(timeout)
				if removed > 0 {
					// Log cleanup activity (could be sent to logger)
					_ = removed
				}

			case <-mb.ctx.Done():
				return
			}
		}
	}()
}

// Global message bus instance
var globalMessageBus *MessageBus
var globalMessageBusOnce sync.Once

// GetGlobalMessageBus returns the global message bus instance
func GetGlobalMessageBus() *MessageBus {
	globalMessageBusOnce.Do(func() {
		globalMessageBus = NewMessageBus(100)
		// Start cleanup routine - remove subscribers inactive for more than 5 minutes
		globalMessageBus.StartCleanupRoutine(1*time.Minute, 5*time.Minute)
	})
	return globalMessageBus
}
