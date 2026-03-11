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

package signalrserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac-signalr/signalr"
	"github.com/mdaxf/iac/config"
	"github.com/mdaxf/iac/logger"
)

// SignalRServer manages the embedded SignalR server
type SignalRServer struct {
	config     *config.SignalRRoleConfig
	server     *http.Server
	hub        *IACMessageBus
	signalrSrv signalr.Server // stored so BroadcastToHub can use HubClients() for proper server push
	ilog       logger.Log
	mu         sync.RWMutex
	running    bool
	nodeData   map[string]interface{}
	cancelFunc context.CancelFunc
}

// NewSignalRServer creates a new SignalR server instance
func NewSignalRServer(cfg *config.SignalRRoleConfig) *SignalRServer {
	return &SignalRServer{
		config:   cfg,
		ilog:     logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "SignalRServer"},
		nodeData: make(map[string]interface{}),
	}
}

// Start starts the SignalR server
func (s *SignalRServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("SignalR server is already running")
	}

	fmt.Printf("[SignalR] Starting embedded SignalR server on port %d, hub: %s\n", s.config.Port, s.config.Hub)
	s.ilog.Info(fmt.Sprintf("Starting embedded SignalR server on port %d", s.config.Port))

	// Initialize node data
	s.nodeData["Name"] = "iac-signalr-embedded"
	s.nodeData["AppID"] = uuid.New().String()
	s.nodeData["Description"] = "IAC Embedded SignalR Server"
	s.nodeData["Type"] = "SignalR Server"
	s.nodeData["Version"] = "1.0.0"
	s.nodeData["Status"] = "Running"
	s.nodeData["StartTime"] = time.Now().UTC()

	// Get host info
	if hostInfo, err := s.getHostAndIPAddress(); err == nil {
		for k, v := range hostInfo {
			s.nodeData[k] = v
		}
	}

	// Create hub
	s.hub = NewIACMessageBus()

	// Configure server options
	keepAlive := s.config.KeepAliveInterval
	if keepAlive == 0 {
		keepAlive = 30 // default 30 seconds
	}

	timeout := s.config.TimeoutInterval
	if timeout == 0 {
		timeout = 60 // default 60 seconds
	}

	// Build allowed origin host patterns.
	// configuration.json may store origins as a single comma-separated string in
	// the first array element (e.g. "*,http://localhost:8080,...").
	// expandOrigins splits into individual entries; buildOriginPatterns extracts
	// the HOST part so nhooyr.io/websocket filepath.Match works correctly.
	normalizedOrigins := expandOrigins(s.config.AllowedOrigins)
	allowedPatterns := buildOriginPatterns(normalizedOrigins)

	// Create SignalR logger adapter
	logAdapter := NewLogAdapter(s.ilog)

	// Build transport options
	var transports []signalr.TransportType
	if s.config.EnableWebSocket {
		transports = append(transports, signalr.TransportWebSockets)
	}
	if s.config.EnableSSE {
		transports = append(transports, signalr.TransportServerSentEvents)
	}
	if len(transports) == 0 {
		// Default to WebSocket if nothing specified
		transports = []signalr.TransportType{signalr.TransportWebSockets}
	}

	// nhooyr.io/websocket (used internally by the signalr library) explicitly
	// recommends NOT using "*" as an OriginPattern and instead setting
	// InsecureSkipVerify when all origins should be allowed.
	// See: nhooyr.io/websocket accept.go — "Do not use * as a pattern to allow
	// any origin, prefer to use InsecureSkipVerify instead."
	hasWildcard := containsWildcard(normalizedOrigins)
	skipVerify := hasWildcard || s.config.InsecureSkipVerify

	// Create SignalR server.
	// UseHub(s.hub) shares the single IACMessageBus instance (initialised with ilog,
	// BroadcastToHub etc.) across every connection instead of creating blank
	// reflection-copies via SimpleHubFactory, which would lose the ilog field.
	signalrServer, err := signalr.NewServer(ctx,
		signalr.UseHub(s.hub),
		signalr.Logger(logAdapter, false),
		signalr.HTTPTransports(transports...),
		signalr.KeepAliveInterval(time.Duration(keepAlive)*time.Second),
		signalr.TimeoutInterval(time.Duration(timeout)*time.Second),
		signalr.HandshakeTimeout(15*time.Second),
		signalr.AllowOriginPatterns(allowedPatterns),
		signalr.InsecureSkipVerify(skipVerify))

	if err != nil {
		return fmt.Errorf("failed to create SignalR server: %v", err)
	}

	// Store the server so BroadcastToHub can use HubClients() for proper server-to-client push.
	// The hub's own Clients() method only works correctly inside a SignalR invocation context;
	// server.HubClients() works from any Go goroutine.
	s.signalrSrv = signalrServer

	// Set allowed clients globally (store the expanded list joined for display)
	signalr.AllowedClients = strings.Join(normalizedOrigins, ",")

	// Create HTTP router
	router := http.NewServeMux()

	// Map SignalR hub
	hubPath := s.config.Hub
	if hubPath == "" {
		hubPath = "/iacmessagebus"
	}
	signalrServer.MapHTTP(signalr.WithHTTPServeMux(router), hubPath)

	// Health check endpoint
	router.HandleFunc("/health", s.healthHandler)

	// Status endpoint
	router.HandleFunc("/status", s.statusHandler)

	// Create HTTP server.
	// NOTE: ReadTimeout and WriteTimeout must NOT be set for a server that handles
	// WebSocket connections. HTTP-level write timeouts kill the hijacked WebSocket
	// connection after the deadline, producing "Server timeout elapsed without
	// receiving a message from the server" on the client side.
	address := fmt.Sprintf(":%d", s.config.Port)
	s.server = &http.Server{
		Addr:        address,
		Handler:     s.corsMiddleware(s.logMiddleware(router)),
		IdleTimeout: 120 * time.Second,
	}

	// Configure TLS if specified
	if s.config.TLS != nil && s.config.TLS.CertFile != "" && s.config.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(s.config.TLS.CertFile, s.config.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificates: %v", err)
		}
		s.server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	// Create cancellable context
	serverCtx, cancel := context.WithCancel(ctx)
	s.cancelFunc = cancel

	// Start server in goroutine
	go func() {
		var err error
		if s.server.TLSConfig != nil {
			s.ilog.Info(fmt.Sprintf("SignalR server listening on %s (TLS enabled)", address))
			err = s.server.ListenAndServeTLS("", "")
		} else {
			s.ilog.Info(fmt.Sprintf("SignalR server listening on %s", address))
			err = s.server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			s.ilog.Error(fmt.Sprintf("SignalR server error: %v", err))
		}
	}()

	// Monitor context for shutdown
	go func() {
		<-serverCtx.Done()
		s.Stop(context.Background())
	}()

	s.running = true
	fmt.Printf("[SignalR] Server started - Hub: %s, Port: %d, Transports: %v, AllowedOrigins: %v, InsecureSkipVerify: %v\n",
		hubPath, s.config.Port, transports, allowedPatterns, skipVerify)
	s.ilog.Info(fmt.Sprintf("SignalR server started successfully - Hub: %s, Port: %d, Transports: %v, InsecureSkipVerify: %v",
		hubPath, s.config.Port, transports, skipVerify))

	return nil
}

// Stop stops the SignalR server
func (s *SignalRServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.ilog.Info("Stopping SignalR server...")

	if s.cancelFunc != nil {
		s.cancelFunc()
	}

	if s.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := s.server.Shutdown(shutdownCtx); err != nil {
			s.ilog.Error(fmt.Sprintf("Error shutting down SignalR server: %v", err))
			return err
		}
	}

	s.running = false
	s.ilog.Info("SignalR server stopped")

	return nil
}

// IsRunning returns whether the server is running
func (s *SignalRServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetStatus returns the server status
func (s *SignalRServer) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]interface{})
	status["running"] = s.running
	status["port"] = s.config.Port
	status["hub"] = s.config.Hub
	status["node"] = s.nodeData

	return status
}

// healthHandler handles health check requests
func (s *SignalRServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := map[string]interface{}{
		"status":    "healthy",
		"running":   s.running,
		"node":      s.nodeData,
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// statusHandler handles status requests
func (s *SignalRServer) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.GetStatus())
}

// corsMiddleware adds CORS headers
func (s *SignalRServer) corsMiddleware(next http.Handler) http.Handler {
	// Pre-expand comma-separated origins once at construction time.
	expanded := expandOrigins(s.config.AllowedOrigins)
	allowAll := len(expanded) == 0 || containsWildcard(expanded)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		if allowAll {
			allowed = true
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		} else {
			for _, allowedOrigin := range expanded {
				if origin == allowedOrigin {
					allowed = true
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			// x-signalr-user-agent is sent by the SignalR JS client on every request
			// and must be explicitly whitelisted or the preflight will fail.
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Requested-With, x-signalr-user-agent")
		}

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// logMiddleware logs requests (skips noisy health check endpoints)
func (s *SignalRServer) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// Skip logging for root path and health endpoints to reduce log noise
		if r.URL.Path != "/" && r.URL.Path != "/health" && r.URL.Path != "/status" {
			s.ilog.Debug(fmt.Sprintf("%s %s %v", r.Method, r.URL.Path, time.Since(start)))
		}
	})
}

// getHostAndIPAddress returns host and IP information
func (s *SignalRServer) getHostAndIPAddress() (map[string]interface{}, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var ipAddress string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipAddress = ipnet.IP.String()
				break
			}
		}
	}

	if ipAddress == "" {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				ipAddress = ipnet.IP.String()
				break
			}
		}
	}

	if ipAddress == "" {
		ipAddress = "127.0.0.1"
	}

	return map[string]interface{}{
		"Host":      hostname,
		"OS":        runtime.GOOS,
		"IPAddress": ipAddress,
	}, nil
}

// LogAdapter adapts IAC logger to SignalR logger interface
type LogAdapter struct {
	ilog logger.Log
}

// NewLogAdapter creates a new log adapter
func NewLogAdapter(ilog logger.Log) *LogAdapter {
	return &LogAdapter{ilog: ilog}
}

// Log implements signalr.StructuredLogger
func (l *LogAdapter) Log(keyVals ...interface{}) error {
	msg := ""
	for i := 0; i < len(keyVals); i += 2 {
		if i+1 < len(keyVals) {
			msg += fmt.Sprintf("%v=%v ", keyVals[i], keyVals[i+1])
		}
	}
	l.ilog.Debug(msg)
	return nil
}

// Global SignalR server instance
var (
	globalServer *SignalRServer
	serverMu     sync.RWMutex
)

// GetGlobalSignalRServer returns the global SignalR server instance
func GetGlobalSignalRServer() *SignalRServer {
	serverMu.RLock()
	defer serverMu.RUnlock()
	return globalServer
}

// SetGlobalSignalRServer sets the global SignalR server instance
func SetGlobalSignalRServer(server *SignalRServer) {
	serverMu.Lock()
	defer serverMu.Unlock()
	globalServer = server
}

// InitGlobalSignalRServer initializes and sets the global SignalR server
func InitGlobalSignalRServer(cfg *config.SignalRRoleConfig) *SignalRServer {
	server := NewSignalRServer(cfg)
	SetGlobalSignalRServer(server)
	return server
}

// BroadcastToHub sends a message to all connected SignalR clients on the given topic.
// Uses server.HubClients().All() which iterates the clients sync.Map — populated
// synchronously in onConnected() — so it is reliable even immediately after a new
// connection (no goroutine timing race like group membership has).
// The topic-based on() handlers on the JS side ensure only interested clients process it.
func (s *SignalRServer) BroadcastToHub(topic, message string) {
	s.mu.RLock()
	srv := s.signalrSrv
	running := s.running
	s.mu.RUnlock()
	if srv == nil || !running {
		s.ilog.Info(fmt.Sprintf("BroadcastToHub: skipped (srv=%v, running=%v) topic=%s", srv != nil, running, topic))
		return
	}
	s.ilog.Info(fmt.Sprintf("BroadcastToHub: broadcasting topic=%s", topic))
	srv.HubClients().All().Send(topic, message)
}

// expandOrigins flattens the AllowedOrigins slice, splitting any comma-separated
// entries into individual origins and trimming whitespace.
func expandOrigins(origins []string) []string {
	var result []string
	for _, entry := range origins {
		for _, part := range strings.Split(entry, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				result = append(result, part)
			}
		}
	}
	return result
}

// containsWildcard reports whether the slice contains a bare "*" wildcard.
func containsWildcard(origins []string) bool {
	for _, o := range origins {
		if o == "*" {
			return true
		}
	}
	return false
}

// buildOriginPatterns converts expanded origin strings into host patterns suitable
// for signalr.AllowOriginPatterns, which passes them to nhooyr.io/websocket's
// AcceptOptions.OriginPatterns.  That library uses filepath.Match against the
// HOST part of the Origin header (e.g. "localhost:3000"), NOT a full URL regex:
//
//   - A bare "*" wildcard stays as "*" — filepath.Match("*", "localhost:3000") matches.
//   - A full URL like "http://localhost:3000" is parsed; only the host "localhost:3000"
//     is kept, so the match works correctly.
//   - If any entry is "*" the returned slice contains only ["*"] (allow all).
func buildOriginPatterns(origins []string) []string {
	if len(origins) == 0 || containsWildcard(origins) {
		return []string{"*"}
	}
	patterns := make([]string, 0, len(origins))
	for _, o := range origins {
		// If origin looks like a full URL, extract the host so filepath.Match works.
		if strings.Contains(o, "://") {
			if u, err := url.Parse(o); err == nil && u.Host != "" {
				patterns = append(patterns, u.Host)
				continue
			}
		}
		patterns = append(patterns, o)
	}
	return patterns
}
