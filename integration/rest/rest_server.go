package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/mdaxf/iac/integration/datahub"
	"github.com/sirupsen/logrus"
)

// RESTServer is a REST web service server
type RESTServer struct {
	Port        int
	Router      *mux.Router
	Server      *http.Server
	Logger      *logrus.Logger
	Handlers    map[string]HandlerFunc
	Middlewares []MiddlewareFunc
	enabled     bool
	datahubMode bool
	mu          sync.RWMutex
}

// HandlerFunc is a REST handler function
type HandlerFunc func(ctx *Context) error

// MiddlewareFunc is a middleware function
type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc

// Context represents a REST request context
type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
	Vars     map[string]string
	Body     []byte
	BodyJSON map[string]interface{}
	Logger   *logrus.Logger
}

// RESTServerConfig is the configuration for REST server
type RESTServerConfig struct {
	Port        int  `json:"port"`
	DatahubMode bool `json:"datahub_mode"`
}

var (
	globalRESTServer *RESTServer
)

// NewRESTServer creates a new REST server
func NewRESTServer(config RESTServerConfig, logger *logrus.Logger) *RESTServer {
	if logger == nil {
		logger = logrus.New()
	}

	router := mux.NewRouter()

	server := &RESTServer{
		Port:        config.Port,
		Router:      router,
		Logger:      logger,
		Handlers:    make(map[string]HandlerFunc),
		Middlewares: make([]MiddlewareFunc, 0),
		enabled:     false,
		datahubMode: config.DatahubMode,
	}

	// Add default middlewares
	server.Use(loggingMiddleware(logger))
	server.Use(recoveryMiddleware(logger))

	return server
}

// InitRESTServer initializes the global REST server from config file
func InitRESTServer(configFile string, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config RESTServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalRESTServer = NewRESTServer(config, logger)
	return nil
}

// GetGlobalRESTServer returns the global REST server instance
func GetGlobalRESTServer() *RESTServer {
	return globalRESTServer
}

// Use adds a middleware to the server
func (s *RESTServer) Use(middleware MiddlewareFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Middlewares = append(s.Middlewares, middleware)
}

// RegisterHandler registers a handler for a specific route
func (s *RESTServer) RegisterHandler(method, path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := fmt.Sprintf("%s:%s", method, path)
	s.Handlers[key] = handler

	// Wrap handler with middlewares
	httpHandler := s.wrapHandler(handler)
	for i := len(s.Middlewares) - 1; i >= 0; i-- {
		httpHandler = s.Middlewares[i](httpHandler)
	}

	s.Router.HandleFunc(path, httpHandler).Methods(method)
	s.Logger.Infof("Registered REST handler: %s %s", method, path)
}

// wrapHandler wraps a HandlerFunc into http.HandlerFunc
func (s *RESTServer) wrapHandler(handler HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse JSON body if content type is JSON
		var bodyJSON map[string]interface{}
		if r.Header.Get("Content-Type") == "application/json" && len(body) > 0 {
			if err := json.Unmarshal(body, &bodyJSON); err != nil {
				http.Error(w, "Failed to parse JSON body", http.StatusBadRequest)
				return
			}
		}

		// Create context
		ctx := &Context{
			Request:  r,
			Response: w,
			Vars:     mux.Vars(r),
			Body:     body,
			BodyJSON: bodyJSON,
			Logger:   s.Logger,
		}

		// If datahub mode is enabled, send to datahub first
		if s.datahubMode {
			envelope := datahub.CreateEnvelope(
				"REST",
				r.URL.Path,
				"",
				r.Header.Get("Content-Type"),
				bodyJSON,
			)
			envelope.Metadata["method"] = r.Method
			envelope.Metadata["remote_addr"] = r.RemoteAddr

			hub := datahub.GetGlobalDataHub()
			if hub.IsEnabled() {
				if err := hub.RouteMessage(envelope); err != nil {
					s.Logger.Warnf("Failed to route message to datahub: %v", err)
				}
			}
		}

		// Call handler
		if err := handler(ctx); err != nil {
			s.Logger.Errorf("Handler error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// GET registers a GET handler
func (s *RESTServer) GET(path string, handler HandlerFunc) {
	s.RegisterHandler("GET", path, handler)
}

// POST registers a POST handler
func (s *RESTServer) POST(path string, handler HandlerFunc) {
	s.RegisterHandler("POST", path, handler)
}

// PUT registers a PUT handler
func (s *RESTServer) PUT(path string, handler HandlerFunc) {
	s.RegisterHandler("PUT", path, handler)
}

// PATCH registers a PATCH handler
func (s *RESTServer) PATCH(path string, handler HandlerFunc) {
	s.RegisterHandler("PATCH", path, handler)
}

// DELETE registers a DELETE handler
func (s *RESTServer) DELETE(path string, handler HandlerFunc) {
	s.RegisterHandler("DELETE", path, handler)
}

// Start starts the REST server
func (s *RESTServer) Start() error {
	if s.enabled {
		return fmt.Errorf("REST server already running")
	}

	s.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      s.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.enabled = true
	s.Logger.Infof("Starting REST server on port %d", s.Port)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Errorf("REST server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the REST server
func (s *RESTServer) Stop() error {
	if !s.enabled {
		return fmt.Errorf("REST server not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.enabled = false
	s.Logger.Info("REST server stopped")
	return nil
}

// IsEnabled returns whether the REST server is enabled
func (s *RESTServer) IsEnabled() bool {
	return s.enabled
}

// Context helper methods

// JSON sends a JSON response
func (c *Context) JSON(statusCode int, data interface{}) error {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(statusCode)

	if data != nil {
		return json.NewEncoder(c.Response).Encode(data)
	}
	return nil
}

// String sends a plain text response
func (c *Context) String(statusCode int, message string) error {
	c.Response.Header().Set("Content-Type", "text/plain")
	c.Response.WriteHeader(statusCode)
	_, err := c.Response.Write([]byte(message))
	return err
}

// Error sends an error response
func (c *Context) Error(statusCode int, message string) error {
	return c.JSON(statusCode, map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"path":    c.Request.URL.Path,
		"method":  c.Request.Method,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetParam gets a URL parameter
func (c *Context) GetParam(key string) string {
	return c.Vars[key]
}

// GetQuery gets a query parameter
func (c *Context) GetQuery(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetHeader gets a header value
func (c *Context) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// Middleware functions

// loggingMiddleware logs HTTP requests
func loggingMiddleware(logger *logrus.Logger) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Infof("REST %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			next(w, r)
			logger.Infof("REST %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
		}
	}
}

// recoveryMiddleware recovers from panics
func recoveryMiddleware(logger *logrus.Logger) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Errorf("REST handler panic: %v", err)
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			next(w, r)
		}
	}
}

// corsMiddleware adds CORS headers
func CORSMiddleware(origins []string) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, o := range origins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}
}

// authMiddleware adds authentication
func AuthMiddleware(authFunc func(token string) bool) MiddlewareFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !authFunc(token) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}
