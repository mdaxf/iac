package debug

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/mdaxf/iac/logger"
)

// SSEHandler handles Server-Sent Events for debug event streaming
type SSEHandler struct {
	messageBus *MessageBus
	sessions   map[string]*DebugSession
	mu         sync.RWMutex
	log        logger.Log
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler(messageBus *MessageBus) *SSEHandler {
	return &SSEHandler{
		messageBus: messageBus,
		sessions:   make(map[string]*DebugSession),
		log: logger.Log{
			ModuleName:     "SSEHandler",
			ControllerName: "Debug",
		},
	}
}

// StreamEvents handles SSE connections for debug event streaming
// URL: /api/debug/stream/:sessionID
func (h *SSEHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get session ID from URL
	sessionID := r.URL.Query().Get("sessionID")
	if sessionID == "" {
		http.Error(w, "sessionID is required", http.StatusBadRequest)
		return
	}

	// Get optional subscriber ID (for multiple clients monitoring same session)
	subscriberID := r.URL.Query().Get("subscriberID")
	if subscriberID == "" {
		subscriberID = fmt.Sprintf("subscriber-%d", time.Now().UnixNano())
	}

	// Parse optional event filters
	filters := h.parseFilters(r)

	// Create subscriber
	subscriber := NewSubscriber(subscriberID, sessionID, 100)
	subscriber.Filters = filters

	// Subscribe to message bus
	h.messageBus.Subscribe(sessionID, subscriber)
	defer h.messageBus.Unsubscribe(sessionID, subscriberID)

	h.log.Info(fmt.Sprintf("SSE client connected: session=%s, subscriber=%s", sessionID, subscriberID))

	// Get flusher for SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Send initial connection confirmation
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"sessionID\":\"%s\",\"subscriberID\":\"%s\"}\n\n", sessionID, subscriberID)
	flusher.Flush()

	// Stream events until client disconnects or context is cancelled
	ctx := r.Context()
	for {
		select {
		case event, ok := <-subscriber.EventChannel:
			if !ok {
				// Channel closed
				return
			}

			// Format and send event
			eventData := event.ToSSEFormat()
			fmt.Fprintf(w, "%s\n\n", eventData)
			flusher.Flush()

		case <-ctx.Done():
			// Client disconnected
			h.log.Info(fmt.Sprintf("SSE client disconnected: session=%s, subscriber=%s", sessionID, subscriberID))
			return

		case <-subscriber.ctx.Done():
			// Subscriber closed
			return
		}
	}
}

// parseFilters parses event filters from query parameters
func (h *SSEHandler) parseFilters(r *http.Request) *EventFilter {
	query := r.URL.Query()

	filters := &EventFilter{
		EventTypes:    []EventType{},
		MinLevel:      query.Get("minLevel"),
		TranCodeNames: []string{},
		FunctionTypes: []string{},
	}

	// Parse event types
	if eventTypesParam := query.Get("eventTypes"); eventTypesParam != "" {
		var eventTypes []EventType
		if err := json.Unmarshal([]byte(eventTypesParam), &eventTypes); err == nil {
			filters.EventTypes = eventTypes
		}
	}

	// Parse trancode names
	if tranCodesParam := query.Get("tranCodes"); tranCodesParam != "" {
		var tranCodes []string
		if err := json.Unmarshal([]byte(tranCodesParam), &tranCodes); err == nil {
			filters.TranCodeNames = tranCodes
		}
	}

	// Parse function types
	if funcTypesParam := query.Get("functionTypes"); funcTypesParam != "" {
		var funcTypes []string
		if err := json.Unmarshal([]byte(funcTypesParam), &funcTypes); err == nil {
			filters.FunctionTypes = funcTypes
		}
	}

	return filters
}

// StartDebugSession starts a new debug session
// URL: POST /api/debug/sessions
func (h *SSEHandler) StartDebugSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID    string `json:"sessionID"`
		TranCodeName string `json:"tranCodeName"`
		UserID       string `json:"userID"`
		Description  string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		req.SessionID = fmt.Sprintf("debug-%d", time.Now().UnixNano())
	}

	// Create debug session
	session := NewDebugSession(req.SessionID, req.TranCodeName, req.UserID)
	session.Description = req.Description

	h.mu.Lock()
	h.sessions[req.SessionID] = session
	h.mu.Unlock()

	session.Start()

	h.log.Info(fmt.Sprintf("Debug session started: %s (trancode: %s, user: %s)",
		req.SessionID, req.TranCodeName, req.UserID))

	// Return session info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessionID":    session.SessionID,
		"tranCodeName": session.TranCodeName,
		"status":       session.Status,
		"startTime":    session.StartTime,
	})
}

// StopDebugSession stops a debug session
// URL: POST /api/debug/sessions/:sessionID/stop
func (h *SSEHandler) StopDebugSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionID")
	if sessionID == "" {
		http.Error(w, "sessionID is required", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	session, exists := h.sessions[sessionID]
	h.mu.Unlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	session.Stop()

	h.log.Info(fmt.Sprintf("Debug session stopped: %s", sessionID))

	// Return session summary
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.GetSummary())
}

// GetDebugSession gets information about a debug session
// URL: GET /api/debug/sessions/:sessionID
func (h *SSEHandler) GetDebugSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionID")
	if sessionID == "" {
		http.Error(w, "sessionID is required", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	session, exists := h.sessions[sessionID]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session.GetSummary())
}

// ListDebugSessions lists all active debug sessions
// URL: GET /api/debug/sessions
func (h *SSEHandler) ListDebugSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	sessions := make([]map[string]interface{}, 0, len(h.sessions))
	for _, session := range h.sessions {
		sessions = append(sessions, session.GetSummary())
	}
	h.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetExecutionTrace gets the full execution trace for a session
// URL: GET /api/debug/sessions/:sessionID/trace
func (h *SSEHandler) GetExecutionTrace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("sessionID")
	if sessionID == "" {
		http.Error(w, "sessionID is required", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	session, exists := h.sessions[sessionID]
	h.mu.RUnlock()

	if !exists {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	trace := session.GetExecutionTrace()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trace)
}

// RegisterRoutes registers all debug API routes
func (h *SSEHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/debug/stream", h.StreamEvents)
	mux.HandleFunc("/api/debug/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.StartDebugSession(w, r)
		} else if r.Method == http.MethodGet {
			// Check if sessionID is provided
			if r.URL.Query().Get("sessionID") != "" {
				h.GetDebugSession(w, r)
			} else {
				h.ListDebugSessions(w, r)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/debug/sessions/stop", h.StopDebugSession)
	mux.HandleFunc("/api/debug/sessions/trace", h.GetExecutionTrace)
}

// CleanupSessions removes completed sessions older than the specified duration
func (h *SSEHandler) CleanupSessions(maxAge time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	for sessionID, session := range h.sessions {
		if session.Status != "running" && now.Sub(session.EndTime) > maxAge {
			delete(h.sessions, sessionID)
			h.log.Info(fmt.Sprintf("Cleaned up old debug session: %s", sessionID))
		}
	}
}

// StartCleanupRoutine starts a background routine to clean up old sessions
func (h *SSEHandler) StartCleanupRoutine(interval, maxAge time.Duration) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.CleanupSessions(maxAge)
			case <-ctx.Done():
				return
			}
		}
	}()

	return cancel
}
