package soap

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mdaxf/iac/integration/datahub"
	"github.com/sirupsen/logrus"
)

// SOAPClient is a SOAP web service client
type SOAPClient struct {
	URL         string
	Namespace   string
	Timeout     time.Duration
	Headers     map[string]string
	Client      *http.Client
	Logger      *logrus.Logger
	enabled     bool
	datahubMode bool
}

// SOAPClientConfig is the configuration for SOAP client
type SOAPClientConfig struct {
	URL         string            `json:"url"`
	Namespace   string            `json:"namespace"`
	Timeout     int               `json:"timeout"` // seconds
	Headers     map[string]string `json:"headers"`
	DatahubMode bool              `json:"datahub_mode"`
}

var (
	globalSOAPClient *SOAPClient
)

// SOAPEnvelope represents a SOAP envelope
type SOAPEnvelope struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	XMLNS   string   `xml:"xmlns:soap,attr"`
	Header  *SOAPHeader
	Body    SOAPBody
}

// SOAPHeader represents a SOAP header
type SOAPHeader struct {
	XMLName xml.Name `xml:"soap:Header"`
	Content interface{}
}

// SOAPBody represents a SOAP body
type SOAPBody struct {
	XMLName xml.Name `xml:"soap:Body"`
	Content interface{} `xml:",innerxml"`
}

// SOAPFault represents a SOAP fault
type SOAPFault struct {
	XMLName     xml.Name `xml:"Fault"`
	FaultCode   string   `xml:"faultcode"`
	FaultString string   `xml:"faultstring"`
	FaultActor  string   `xml:"faultactor,omitempty"`
	Detail      string   `xml:"detail,omitempty"`
}

// NewSOAPClient creates a new SOAP client
func NewSOAPClient(config SOAPClientConfig, logger *logrus.Logger) *SOAPClient {
	if logger == nil {
		logger = logrus.New()
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	namespace := config.Namespace
	if namespace == "" {
		namespace = "http://schemas.xmlsoap.org/soap/envelope/"
	}

	return &SOAPClient{
		URL:       config.URL,
		Namespace: namespace,
		Timeout:   timeout,
		Headers:   config.Headers,
		Client: &http.Client{
			Timeout: timeout,
		},
		Logger:      logger,
		enabled:     true,
		datahubMode: config.DatahubMode,
	}
}

// InitSOAPClient initializes the global SOAP client from config file
func InitSOAPClient(configFile string, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config SOAPClientConfig
	if err := xml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalSOAPClient = NewSOAPClient(config, logger)
	return nil
}

// GetGlobalSOAPClient returns the global SOAP client instance
func GetGlobalSOAPClient() *SOAPClient {
	return globalSOAPClient
}

// Call makes a SOAP call
func (c *SOAPClient) Call(action string, request interface{}, response interface{}) error {
	if !c.enabled {
		return fmt.Errorf("SOAP client is disabled")
	}

	c.Logger.Infof("Making SOAP call to %s with action %s", c.URL, action)

	// Create SOAP envelope
	envelope := c.createEnvelope(request)

	// Marshal to XML
	requestXML, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	// Add XML declaration
	requestXML = append([]byte(xml.Header), requestXML...)

	c.Logger.Debugf("SOAP request:\n%s", string(requestXML))

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.URL, bytes.NewReader(requestXML))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	if action != "" {
		httpReq.Header.Set("SOAPAction", action)
	}

	for k, v := range c.Headers {
		httpReq.Header.Set(k, v)
	}

	// Send request
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send SOAP request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	responseXML, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read SOAP response: %w", err)
	}

	c.Logger.Debugf("SOAP response:\n%s", string(responseXML))

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("SOAP call failed with HTTP status %d: %s", httpResp.StatusCode, string(responseXML))
	}

	// Parse response envelope
	var respEnvelope SOAPEnvelope
	if err := xml.Unmarshal(responseXML, &respEnvelope); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP response: %w", err)
	}

	// Check for SOAP fault
	var fault SOAPFault
	if err := xml.Unmarshal([]byte(respEnvelope.Body.Content.(string)), &fault); err == nil {
		if fault.FaultCode != "" {
			return fmt.Errorf("SOAP fault: %s - %s", fault.FaultCode, fault.FaultString)
		}
	}

	// Unmarshal response body
	if response != nil {
		if err := xml.Unmarshal([]byte(respEnvelope.Body.Content.(string)), response); err != nil {
			return fmt.Errorf("failed to unmarshal SOAP response body: %w", err)
		}
	}

	c.Logger.Infof("SOAP call completed successfully")

	// If datahub mode is enabled, send to datahub
	if c.datahubMode {
		envelope := datahub.CreateEnvelope(
			"SOAP",
			c.URL,
			"",
			"application/soap+xml",
			response,
		)
		envelope.Metadata["action"] = action
		envelope.Metadata["status_code"] = httpResp.StatusCode
		envelope.OriginalBody = responseXML

		hub := datahub.GetGlobalDataHub()
		if hub.IsEnabled() {
			if err := hub.RouteMessage(envelope); err != nil {
				c.Logger.Warnf("Failed to route message to datahub: %v", err)
			}
		}
	}

	return nil
}

// createEnvelope creates a SOAP envelope from request
func (c *SOAPClient) createEnvelope(request interface{}) *SOAPEnvelope {
	return &SOAPEnvelope{
		XMLNS: c.Namespace,
		Body: SOAPBody{
			Content: request,
		},
	}
}

// CallWithHeader makes a SOAP call with custom header
func (c *SOAPClient) CallWithHeader(action string, header interface{}, request interface{}, response interface{}) error {
	if !c.enabled {
		return fmt.Errorf("SOAP client is disabled")
	}

	c.Logger.Infof("Making SOAP call with header to %s with action %s", c.URL, action)

	// Create SOAP envelope with header
	envelope := &SOAPEnvelope{
		XMLNS: c.Namespace,
		Header: &SOAPHeader{
			Content: header,
		},
		Body: SOAPBody{
			Content: request,
		},
	}

	// Marshal to XML
	requestXML, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SOAP envelope: %w", err)
	}

	// Add XML declaration
	requestXML = append([]byte(xml.Header), requestXML...)

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.URL, bytes.NewReader(requestXML))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "text/xml; charset=utf-8")
	if action != "" {
		httpReq.Header.Set("SOAPAction", action)
	}

	for k, v := range c.Headers {
		httpReq.Header.Set(k, v)
	}

	// Send request
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send SOAP request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	responseXML, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read SOAP response: %w", err)
	}

	// Check for HTTP errors
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("SOAP call failed with HTTP status %d: %s", httpResp.StatusCode, string(responseXML))
	}

	// Parse response envelope
	var respEnvelope SOAPEnvelope
	if err := xml.Unmarshal(responseXML, &respEnvelope); err != nil {
		return fmt.Errorf("failed to unmarshal SOAP response: %w", err)
	}

	// Check for SOAP fault
	var fault SOAPFault
	if err := xml.Unmarshal([]byte(respEnvelope.Body.Content.(string)), &fault); err == nil {
		if fault.FaultCode != "" {
			return fmt.Errorf("SOAP fault: %s - %s", fault.FaultCode, fault.FaultString)
		}
	}

	// Unmarshal response body
	if response != nil {
		if err := xml.Unmarshal([]byte(respEnvelope.Body.Content.(string)), response); err != nil {
			return fmt.Errorf("failed to unmarshal SOAP response body: %w", err)
		}
	}

	c.Logger.Infof("SOAP call with header completed successfully")

	return nil
}

// Enable enables the SOAP client
func (c *SOAPClient) Enable() {
	c.enabled = true
	c.Logger.Info("SOAP client enabled")
}

// Disable disables the SOAP client
func (c *SOAPClient) Disable() {
	c.enabled = false
	c.Logger.Info("SOAP client disabled")
}

// IsEnabled returns whether the SOAP client is enabled
func (c *SOAPClient) IsEnabled() bool {
	return c.enabled
}

// Close closes the SOAP client
func (c *SOAPClient) Close() error {
	c.Client.CloseIdleConnections()
	c.Logger.Info("SOAP client closed")
	return nil
}

// GetProtocolName returns the protocol name
func (c *SOAPClient) GetProtocolName() string {
	return "SOAP"
}

// Initialize initializes the adapter with configuration
func (c *SOAPClient) Initialize(config map[string]interface{}) error {
	if url, ok := config["url"].(string); ok {
		c.URL = url
	}
	if namespace, ok := config["namespace"].(string); ok {
		c.Namespace = namespace
	}
	if timeout, ok := config["timeout"].(float64); ok {
		c.Timeout = time.Duration(timeout) * time.Second
		c.Client.Timeout = c.Timeout
	}
	return nil
}

// Receive receives a message (not applicable for SOAP client, returns error)
func (c *SOAPClient) Receive(timeout time.Duration) (*datahub.MessageEnvelope, error) {
	return nil, fmt.Errorf("SOAP client does not support receive operation")
}

// Health checks the health of the SOAP client
func (c *SOAPClient) Health() error {
	// Simple health check - try to connect to the endpoint
	httpReq, err := http.NewRequest("POST", c.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer httpResp.Body.Close()

	return nil
}
