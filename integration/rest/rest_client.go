package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mdaxf/iac/integration/datahub"
	"github.com/sirupsen/logrus"
)

// RESTClient is a REST web service client
type RESTClient struct {
	BaseURL     string
	Timeout     time.Duration
	Headers     map[string]string
	Client      *http.Client
	Logger      *logrus.Logger
	enabled     bool
	datahubMode bool
}

// RESTClientConfig is the configuration for REST client
type RESTClientConfig struct {
	BaseURL     string            `json:"base_url"`
	Timeout     int               `json:"timeout"` // seconds
	Headers     map[string]string `json:"headers"`
	DatahubMode bool              `json:"datahub_mode"`
}

var (
	globalRESTClient *RESTClient
)

// NewRESTClient creates a new REST client
func NewRESTClient(config RESTClientConfig, logger *logrus.Logger) *RESTClient {
	if logger == nil {
		logger = logrus.New()
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &RESTClient{
		BaseURL: config.BaseURL,
		Timeout: timeout,
		Headers: config.Headers,
		Client: &http.Client{
			Timeout: timeout,
		},
		Logger:      logger,
		enabled:     true,
		datahubMode: config.DatahubMode,
	}
}

// InitRESTClient initializes the global REST client from config file
func InitRESTClient(configFile string, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config RESTClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalRESTClient = NewRESTClient(config, logger)
	return nil
}

// GetGlobalRESTClient returns the global REST client instance
func GetGlobalRESTClient() *RESTClient {
	return globalRESTClient
}

// Request represents a REST request
type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
	Query   map[string]string
}

// Response represents a REST response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	BodyJSON   map[string]interface{}
}

// Send sends a REST request
func (c *RESTClient) Send(req *Request) (*Response, error) {
	if !c.enabled {
		return nil, fmt.Errorf("REST client is disabled")
	}

	url := c.BaseURL + req.Path

	// Add query parameters
	if len(req.Query) > 0 {
		url += "?"
		first := true
		for k, v := range req.Query {
			if !first {
				url += "&"
			}
			url += fmt.Sprintf("%s=%s", k, v)
			first = false
		}
	}

	c.Logger.Infof("Sending REST %s request to %s", req.Method, url)

	// Prepare request body
	var bodyReader io.Reader
	if req.Body != nil {
		bodyData, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for k, v := range c.Headers {
		httpReq.Header.Set(k, v)
	}
	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	// Set default content type if not specified
	if httpReq.Header.Get("Content-Type") == "" && req.Body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Send request
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response body
	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	resp := &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       respBody,
	}

	// Try to parse as JSON
	if len(respBody) > 0 {
		var jsonBody map[string]interface{}
		if err := json.Unmarshal(respBody, &jsonBody); err == nil {
			resp.BodyJSON = jsonBody
		}
	}

	c.Logger.Infof("Received REST response with status code %d", httpResp.StatusCode)

	// If datahub mode is enabled, send to datahub
	if c.datahubMode {
		envelope := datahub.CreateEnvelope(
			"REST",
			url,
			"",
			"application/json",
			resp.BodyJSON,
		)
		envelope.Metadata["status_code"] = httpResp.StatusCode
		envelope.Metadata["method"] = req.Method

		hub := datahub.GetGlobalDataHub()
		if hub.IsEnabled() {
			if err := hub.RouteMessage(envelope); err != nil {
				c.Logger.Warnf("Failed to route message to datahub: %v", err)
			}
		}
	}

	return resp, nil
}

// GET sends a GET request
func (c *RESTClient) GET(path string, query map[string]string) (*Response, error) {
	return c.Send(&Request{
		Method: "GET",
		Path:   path,
		Query:  query,
	})
}

// POST sends a POST request
func (c *RESTClient) POST(path string, body interface{}) (*Response, error) {
	return c.Send(&Request{
		Method: "POST",
		Path:   path,
		Body:   body,
	})
}

// PUT sends a PUT request
func (c *RESTClient) PUT(path string, body interface{}) (*Response, error) {
	return c.Send(&Request{
		Method: "PUT",
		Path:   path,
		Body:   body,
	})
}

// PATCH sends a PATCH request
func (c *RESTClient) PATCH(path string, body interface{}) (*Response, error) {
	return c.Send(&Request{
		Method: "PATCH",
		Path:   path,
		Body:   body,
	})
}

// DELETE sends a DELETE request
func (c *RESTClient) DELETE(path string) (*Response, error) {
	return c.Send(&Request{
		Method: "DELETE",
		Path:   path,
	})
}

// Enable enables the REST client
func (c *RESTClient) Enable() {
	c.enabled = true
	c.Logger.Info("REST client enabled")
}

// Disable disables the REST client
func (c *RESTClient) Disable() {
	c.enabled = false
	c.Logger.Info("REST client disabled")
}

// IsEnabled returns whether the REST client is enabled
func (c *RESTClient) IsEnabled() bool {
	return c.enabled
}

// Close closes the REST client
func (c *RESTClient) Close() error {
	c.Client.CloseIdleConnections()
	c.Logger.Info("REST client closed")
	return nil
}

// GetProtocolName returns the protocol name
func (c *RESTClient) GetProtocolName() string {
	return "REST"
}

// Initialize initializes the adapter with configuration
func (c *RESTClient) Initialize(config map[string]interface{}) error {
	if baseURL, ok := config["base_url"].(string); ok {
		c.BaseURL = baseURL
	}
	if timeout, ok := config["timeout"].(float64); ok {
		c.Timeout = time.Duration(timeout) * time.Second
		c.Client.Timeout = c.Timeout
	}
	if headers, ok := config["headers"].(map[string]string); ok {
		c.Headers = headers
	}
	return nil
}

// Receive receives a message (not applicable for REST client, returns error)
func (c *RESTClient) Receive(timeout time.Duration) (*datahub.MessageEnvelope, error) {
	return nil, fmt.Errorf("REST client does not support receive operation")
}

// Health checks the health of the REST client
func (c *RESTClient) Health() error {
	// Try a simple GET request to check connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}

	return nil
}
