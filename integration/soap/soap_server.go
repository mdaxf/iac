package soap

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/mdaxf/iac/integration/datahub"
	"github.com/sirupsen/logrus"
)

// SOAPServer is a SOAP web service server
type SOAPServer struct {
	Port        int
	Server      *http.Server
	Logger      *logrus.Logger
	Handlers    map[string]SOAPHandlerFunc
	enabled     bool
	datahubMode bool
	mu          sync.RWMutex
}

// SOAPHandlerFunc is a SOAP handler function
type SOAPHandlerFunc func(ctx *SOAPContext) error

// SOAPContext represents a SOAP request context
type SOAPContext struct {
	Action   string
	Request  interface{}
	Response interface{}
	Envelope *SOAPEnvelope
	Logger   *logrus.Logger
}

// SOAPServerConfig is the configuration for SOAP server
type SOAPServerConfig struct {
	Port        int  `json:"port"`
	DatahubMode bool `json:"datahub_mode"`
}

var (
	globalSOAPServer *SOAPServer
)

// NewSOAPServer creates a new SOAP server
func NewSOAPServer(config SOAPServerConfig, logger *logrus.Logger) *SOAPServer {
	if logger == nil {
		logger = logrus.New()
	}

	return &SOAPServer{
		Port:        config.Port,
		Logger:      logger,
		Handlers:    make(map[string]SOAPHandlerFunc),
		enabled:     false,
		datahubMode: config.DatahubMode,
	}
}

// InitSOAPServer initializes the global SOAP server from config file
func InitSOAPServer(configFile string, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config SOAPServerConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalSOAPServer = NewSOAPServer(config, logger)
	return nil
}

// GetGlobalSOAPServer returns the global SOAP server instance
func GetGlobalSOAPServer() *SOAPServer {
	return globalSOAPServer
}

// RegisterHandler registers a handler for a specific SOAP action
func (s *SOAPServer) RegisterHandler(action string, handler SOAPHandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Handlers[action] = handler
	s.Logger.Infof("Registered SOAP handler for action: %s", action)
}

// Start starts the SOAP server
func (s *SOAPServer) Start() error {
	if s.enabled {
		return fmt.Errorf("SOAP server already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleSOAPRequest)

	s.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.enabled = true
	s.Logger.Infof("Starting SOAP server on port %d", s.Port)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Errorf("SOAP server error: %v", err)
		}
	}()

	return nil
}

// handleSOAPRequest handles incoming SOAP requests
func (s *SOAPServer) handleSOAPRequest(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Errorf("SOAP handler panic: %v", err)
			s.sendSOAPFault(w, "Server", fmt.Sprintf("Internal server error: %v", err))
		}
	}()

	// Only accept POST requests
	if r.Method != "POST" {
		s.sendSOAPFault(w, "Client", "Only POST method is allowed")
		return
	}

	// Read request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.Logger.Errorf("Failed to read request body: %v", err)
		s.sendSOAPFault(w, "Server", "Failed to read request body")
		return
	}

	s.Logger.Debugf("Received SOAP request:\n%s", string(body))

	// Parse SOAP envelope
	var envelope SOAPEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		s.Logger.Errorf("Failed to parse SOAP envelope: %v", err)
		s.sendSOAPFault(w, "Client", "Invalid SOAP envelope")
		return
	}

	// Get SOAPAction from header
	action := r.Header.Get("SOAPAction")
	action = trimQuotes(action)

	s.Logger.Infof("Processing SOAP action: %s", action)

	// Find handler
	s.mu.RLock()
	handler, exists := s.Handlers[action]
	s.mu.RUnlock()

	if !exists {
		s.Logger.Warnf("No handler found for action: %s", action)
		s.sendSOAPFault(w, "Client", fmt.Sprintf("Unknown action: %s", action))
		return
	}

	// Create context
	ctx := &SOAPContext{
		Action:   action,
		Envelope: &envelope,
		Logger:   s.Logger,
	}

	// If datahub mode is enabled, send to datahub first
	if s.datahubMode {
		dhEnvelope := datahub.CreateEnvelope(
			"SOAP",
			r.URL.Path,
			"",
			"application/soap+xml",
			envelope.Body.Content,
		)
		dhEnvelope.Metadata["action"] = action
		dhEnvelope.Metadata["remote_addr"] = r.RemoteAddr
		dhEnvelope.OriginalBody = body

		hub := datahub.GetGlobalDataHub()
		if hub.IsEnabled() {
			if err := hub.RouteMessage(dhEnvelope); err != nil {
				s.Logger.Warnf("Failed to route message to datahub: %v", err)
			}
		}
	}

	// Call handler
	if err := handler(ctx); err != nil {
		s.Logger.Errorf("Handler error: %v", err)
		s.sendSOAPFault(w, "Server", err.Error())
		return
	}

	// Create response envelope
	respEnvelope := &SOAPEnvelope{
		XMLNS: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: SOAPBody{
			Content: ctx.Response,
		},
	}

	// Marshal response
	responseXML, err := xml.MarshalIndent(respEnvelope, "", "  ")
	if err != nil {
		s.Logger.Errorf("Failed to marshal SOAP response: %v", err)
		s.sendSOAPFault(w, "Server", "Failed to create response")
		return
	}

	// Add XML declaration
	responseXML = append([]byte(xml.Header), responseXML...)

	s.Logger.Debugf("Sending SOAP response:\n%s", string(responseXML))

	// Send response
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(responseXML)
}

// sendSOAPFault sends a SOAP fault response
func (s *SOAPServer) sendSOAPFault(w http.ResponseWriter, faultCode, faultString string) {
	fault := SOAPFault{
		FaultCode:   faultCode,
		FaultString: faultString,
	}

	envelope := &SOAPEnvelope{
		XMLNS: "http://schemas.xmlsoap.org/soap/envelope/",
		Body: SOAPBody{
			Content: fault,
		},
	}

	responseXML, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		s.Logger.Errorf("Failed to marshal SOAP fault: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Add XML declaration
	responseXML = append([]byte(xml.Header), responseXML...)

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write(responseXML)
}

// Stop stops the SOAP server
func (s *SOAPServer) Stop() error {
	if !s.enabled {
		return fmt.Errorf("SOAP server not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.enabled = false
	s.Logger.Info("SOAP server stopped")
	return nil
}

// IsEnabled returns whether the SOAP server is enabled
func (s *SOAPServer) IsEnabled() bool {
	return s.enabled
}

// trimQuotes removes quotes from a string
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// SOAPContext helper methods

// SetResponse sets the response for the SOAP call
func (c *SOAPContext) SetResponse(response interface{}) {
	c.Response = response
}

// GetRequest gets the request data
func (c *SOAPContext) GetRequest() interface{} {
	return c.Request
}

// ParseRequest parses the request body into a struct
func (c *SOAPContext) ParseRequest(v interface{}) error {
	bodyContent, ok := c.Envelope.Body.Content.(string)
	if !ok {
		return fmt.Errorf("invalid body content type")
	}

	if err := xml.Unmarshal([]byte(bodyContent), v); err != nil {
		return fmt.Errorf("failed to parse request: %w", err)
	}

	c.Request = v
	return nil
}
