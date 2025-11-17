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

// TCPClient is a TCP client
type TCPClient struct {
	Host        string
	Port        int
	Timeout     time.Duration
	conn        net.Conn
	Logger      *logrus.Logger
	enabled     bool
	datahubMode bool
	mu          sync.RWMutex
	autoReconnect bool
	delimiter   byte
}

// TCPClientConfig is the configuration for TCP client
type TCPClientConfig struct {
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Timeout       int    `json:"timeout"` // seconds
	DatahubMode   bool   `json:"datahub_mode"`
	AutoReconnect bool   `json:"auto_reconnect"`
	Delimiter     string `json:"delimiter"` // Message delimiter (newline, null, etc.)
}

var (
	globalTCPClient *TCPClient
)

// NewTCPClient creates a new TCP client
func NewTCPClient(config TCPClientConfig, logger *logrus.Logger) *TCPClient {
	if logger == nil {
		logger = logrus.New()
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
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

	return &TCPClient{
		Host:          config.Host,
		Port:          config.Port,
		Timeout:       timeout,
		Logger:        logger,
		enabled:       false,
		datahubMode:   config.DatahubMode,
		autoReconnect: config.AutoReconnect,
		delimiter:     delimiter,
	}
}

// InitTCPClient initializes the global TCP client from config file
func InitTCPClient(configFile string, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config TCPClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalTCPClient = NewTCPClient(config, logger)
	return nil
}

// GetGlobalTCPClient returns the global TCP client instance
func GetGlobalTCPClient() *TCPClient {
	return globalTCPClient
}

// Connect connects to the TCP server
func (c *TCPClient) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return fmt.Errorf("already connected")
	}

	address := fmt.Sprintf("%s:%d", c.Host, c.Port)
	c.Logger.Infof("Connecting to TCP server at %s", address)

	conn, err := net.DialTimeout("tcp", address, c.Timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	c.enabled = true
	c.Logger.Infof("Connected to TCP server at %s", address)

	return nil
}

// Disconnect disconnects from the TCP server
func (c *TCPClient) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	c.conn = nil
	c.enabled = false
	c.Logger.Info("Disconnected from TCP server")

	return nil
}

// Send sends data to the TCP server
func (c *TCPClient) Send(data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.enabled || c.conn == nil {
		if c.autoReconnect {
			c.mu.RUnlock()
			if err := c.Connect(); err != nil {
				c.mu.RLock()
				return fmt.Errorf("failed to reconnect: %w", err)
			}
			c.mu.RLock()
		} else {
			return fmt.Errorf("not connected")
		}
	}

	c.Logger.Debugf("Sending %d bytes to TCP server", len(data))

	// Set write deadline
	if err := c.conn.SetWriteDeadline(time.Now().Add(c.Timeout)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	// Write data
	n, err := c.conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	c.Logger.Infof("Sent %d bytes to TCP server", n)

	return nil
}

// SendString sends a string message to the TCP server
func (c *TCPClient) SendString(message string) error {
	return c.Send([]byte(message))
}

// SendWithDelimiter sends data with delimiter to the TCP server
func (c *TCPClient) SendWithDelimiter(data []byte) error {
	dataWithDelimiter := append(data, c.delimiter)
	return c.Send(dataWithDelimiter)
}

// Receive receives data from the TCP server
func (c *TCPClient) Receive(bufferSize int) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.enabled || c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	buffer := make([]byte, bufferSize)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to receive data: %w", err)
	}

	data := buffer[:n]
	c.Logger.Debugf("Received %d bytes from TCP server", n)

	// If datahub mode is enabled, send to datahub
	if c.datahubMode {
		envelope := datahub.CreateEnvelope(
			"TCP",
			fmt.Sprintf("%s:%d", c.Host, c.Port),
			"",
			"application/octet-stream",
			data,
		)
		envelope.OriginalBody = data

		hub := datahub.GetGlobalDataHub()
		if hub.IsEnabled() {
			if err := hub.RouteMessage(envelope); err != nil {
				c.Logger.Warnf("Failed to route message to datahub: %v", err)
			}
		}
	}

	return data, nil
}

// ReceiveUntilDelimiter receives data until delimiter is found
func (c *TCPClient) ReceiveUntilDelimiter() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.enabled || c.conn == nil {
		return nil, fmt.Errorf("not connected")
	}

	// Set read deadline
	if err := c.conn.SetReadDeadline(time.Now().Add(c.Timeout)); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	reader := bufio.NewReader(c.conn)
	data, err := reader.ReadBytes(c.delimiter)
	if err != nil {
		return nil, fmt.Errorf("failed to receive data: %w", err)
	}

	// Remove delimiter
	if len(data) > 0 && data[len(data)-1] == c.delimiter {
		data = data[:len(data)-1]
	}

	c.Logger.Debugf("Received %d bytes from TCP server (until delimiter)", len(data))

	// If datahub mode is enabled, send to datahub
	if c.datahubMode {
		envelope := datahub.CreateEnvelope(
			"TCP",
			fmt.Sprintf("%s:%d", c.Host, c.Port),
			"",
			"application/octet-stream",
			data,
		)
		envelope.OriginalBody = data

		hub := datahub.GetGlobalDataHub()
		if hub.IsEnabled() {
			if err := hub.RouteMessage(envelope); err != nil {
				c.Logger.Warnf("Failed to route message to datahub: %v", err)
			}
		}
	}

	return data, nil
}

// SendReceive sends data and receives response
func (c *TCPClient) SendReceive(data []byte, bufferSize int) ([]byte, error) {
	if err := c.Send(data); err != nil {
		return nil, fmt.Errorf("failed to send: %w", err)
	}

	return c.Receive(bufferSize)
}

// IsConnected returns whether the client is connected
func (c *TCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn != nil
}

// Enable enables the TCP client
func (c *TCPClient) Enable() {
	c.enabled = true
	c.Logger.Info("TCP client enabled")
}

// Disable disables the TCP client
func (c *TCPClient) Disable() {
	c.enabled = false
	c.Logger.Info("TCP client disabled")
}

// IsEnabled returns whether the TCP client is enabled
func (c *TCPClient) IsEnabled() bool {
	return c.enabled
}

// Close closes the TCP client
func (c *TCPClient) Close() error {
	if c.IsConnected() {
		return c.Disconnect()
	}
	return nil
}

// GetProtocolName returns the protocol name
func (c *TCPClient) GetProtocolName() string {
	return "TCP"
}

// Initialize initializes the adapter with configuration
func (c *TCPClient) Initialize(config map[string]interface{}) error {
	if host, ok := config["host"].(string); ok {
		c.Host = host
	}
	if port, ok := config["port"].(float64); ok {
		c.Port = int(port)
	}
	if timeout, ok := config["timeout"].(float64); ok {
		c.Timeout = time.Duration(timeout) * time.Second
	}
	return nil
}

// Health checks the health of the TCP client
func (c *TCPClient) Health() error {
	if !c.IsConnected() {
		return fmt.Errorf("not connected")
	}
	return nil
}
