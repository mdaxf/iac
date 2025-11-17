package tcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/mdaxf/iac/integration/datahub"
	"github.com/sirupsen/logrus"
)

// TCPServer is a TCP server
type TCPServer struct {
	Host        string
	Port        int
	listener    net.Listener
	Logger      *logrus.Logger
	Handler     TCPHandlerFunc
	enabled     bool
	datahubMode bool
	mu          sync.RWMutex
	connections map[string]net.Conn
	delimiter   byte
}

// TCPHandlerFunc is a TCP connection handler function
type TCPHandlerFunc func(conn net.Conn, data []byte, logger *logrus.Logger) ([]byte, error)

// TCPServerConfig is the configuration for TCP server
type TCPServerConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	DatahubMode bool   `json:"datahub_mode"`
	Delimiter   string `json:"delimiter"` // Message delimiter
}

var (
	globalTCPServer *TCPServer
)

// NewTCPServer creates a new TCP server
func NewTCPServer(config TCPServerConfig, logger *logrus.Logger) *TCPServer {
	if logger == nil {
		logger = logrus.New()
	}

	host := config.Host
	if host == "" {
		host = "0.0.0.0"
	}

	delimiter := byte('\n')
	if config.Delimiter != "" {
		if config.Delimiter == "\\n" {
			delimiter = '\n'
		} else if config.Delimiter == "\\0" {
			delimiter = '\x00'
		} else if len(config.Delimiter) > 0 {
			delimiter = config.Delimiter[0]
		}
	}

	return &TCPServer{
		Host:        host,
		Port:        config.Port,
		Logger:      logger,
		enabled:     false,
		datahubMode: config.DatahubMode,
		connections: make(map[string]net.Conn),
		delimiter:   delimiter,
	}
}

// InitTCPServer initializes the global TCP server from config file
func InitTCPServer(configFile string, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config TCPServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalTCPServer = NewTCPServer(config, logger)
	return nil
}

// GetGlobalTCPServer returns the global TCP server instance
func GetGlobalTCPServer() *TCPServer {
	return globalTCPServer
}

// SetHandler sets the connection handler
func (s *TCPServer) SetHandler(handler TCPHandlerFunc) {
	s.Handler = handler
}

// Start starts the TCP server
func (s *TCPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.enabled {
		return fmt.Errorf("TCP server already running")
	}

	address := fmt.Sprintf("%s:%d", s.Host, s.Port)
	s.Logger.Infof("Starting TCP server on %s", address)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	s.listener = listener
	s.enabled = true

	// Start accepting connections
	go s.acceptConnections()

	s.Logger.Infof("TCP server started on %s", address)

	return nil
}

// acceptConnections accepts incoming connections
func (s *TCPServer) acceptConnections() {
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Errorf("TCP server panic: %v", r)
		}
	}()

	for s.enabled {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.enabled {
				s.Logger.Errorf("Failed to accept connection: %v", err)
			}
			continue
		}

		s.Logger.Infof("Accepted connection from %s", conn.RemoteAddr().String())

		// Store connection
		s.mu.Lock()
		s.connections[conn.RemoteAddr().String()] = conn
		s.mu.Unlock()

		// Handle connection in goroutine
		go s.handleConnection(conn)
	}
}

// handleConnection handles a single connection
func (s *TCPServer) handleConnection(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			s.Logger.Errorf("TCP connection handler panic: %v", r)
		}

		// Remove connection
		s.mu.Lock()
		delete(s.connections, conn.RemoteAddr().String())
		s.mu.Unlock()

		conn.Close()
		s.Logger.Infof("Closed connection from %s", conn.RemoteAddr().String())
	}()

	reader := bufio.NewReader(conn)

	for {
		// Read data until delimiter
		data, err := reader.ReadBytes(s.delimiter)
		if err != nil {
			s.Logger.Debugf("Connection closed or error reading: %v", err)
			break
		}

		// Remove delimiter
		if len(data) > 0 && data[len(data)-1] == s.delimiter {
			data = data[:len(data)-1]
		}

		s.Logger.Debugf("Received %d bytes from %s", len(data), conn.RemoteAddr().String())

		// If datahub mode is enabled, send to datahub
		if s.datahubMode {
			envelope := datahub.CreateEnvelope(
				"TCP",
				conn.RemoteAddr().String(),
				"",
				"application/octet-stream",
				data,
			)
			envelope.OriginalBody = data

			hub := datahub.GetGlobalDataHub()
			if hub.IsEnabled() {
				if err := hub.RouteMessage(envelope); err != nil {
					s.Logger.Warnf("Failed to route message to datahub: %v", err)
				}
			}
		}

		// Call handler if set
		if s.Handler != nil {
			response, err := s.Handler(conn, data, s.Logger)
			if err != nil {
				s.Logger.Errorf("Handler error: %v", err)
				continue
			}

			// Send response if provided
			if response != nil {
				// Add delimiter to response
				responseWithDelimiter := append(response, s.delimiter)

				if _, err := conn.Write(responseWithDelimiter); err != nil {
					s.Logger.Errorf("Failed to send response: %v", err)
					break
				}

				s.Logger.Debugf("Sent %d bytes to %s", len(response), conn.RemoteAddr().String())
			}
		}
	}
}

// Stop stops the TCP server
func (s *TCPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return fmt.Errorf("TCP server not running")
	}

	s.enabled = false

	// Close all connections
	for addr, conn := range s.connections {
		s.Logger.Infof("Closing connection to %s", addr)
		conn.Close()
	}
	s.connections = make(map[string]net.Conn)

	// Close listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("failed to close listener: %w", err)
		}
	}

	s.Logger.Info("TCP server stopped")

	return nil
}

// IsEnabled returns whether the TCP server is enabled
func (s *TCPServer) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// Broadcast sends data to all connected clients
func (s *TCPServer) Broadcast(data []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return fmt.Errorf("TCP server not running")
	}

	// Add delimiter
	dataWithDelimiter := append(data, s.delimiter)

	count := 0
	for addr, conn := range s.connections {
		if _, err := conn.Write(dataWithDelimiter); err != nil {
			s.Logger.Errorf("Failed to send to %s: %v", addr, err)
			continue
		}
		count++
	}

	s.Logger.Infof("Broadcasted %d bytes to %d clients", len(data), count)

	return nil
}

// SendToClient sends data to a specific client
func (s *TCPServer) SendToClient(remoteAddr string, data []byte) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.enabled {
		return fmt.Errorf("TCP server not running")
	}

	conn, exists := s.connections[remoteAddr]
	if !exists {
		return fmt.Errorf("client %s not found", remoteAddr)
	}

	// Add delimiter
	dataWithDelimiter := append(data, s.delimiter)

	if _, err := conn.Write(dataWithDelimiter); err != nil {
		return fmt.Errorf("failed to send to %s: %w", remoteAddr, err)
	}

	s.Logger.Infof("Sent %d bytes to %s", len(data), remoteAddr)

	return nil
}

// GetConnectedClients returns a list of connected client addresses
func (s *TCPServer) GetConnectedClients() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]string, 0, len(s.connections))
	for addr := range s.connections {
		clients = append(clients, addr)
	}

	return clients
}

// Close closes the TCP server
func (s *TCPServer) Close() error {
	return s.Stop()
}
