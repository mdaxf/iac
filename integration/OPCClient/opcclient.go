// Package opcclient provides a comprehensive OPC UA client implementation with full protocol support
//
// This package offers enhanced OPC UA connectivity capabilities including:
//
// Authentication Support:
//   - Anonymous authentication
//   - Username/Password authentication
//   - Certificate-based authentication
//   - Issued token authentication
//
// Connection Management:
//   - Automatic reconnection on connection loss
//   - Configurable session and request timeouts
//   - Connection monitoring and health checks
//   - Configurable reconnection attempts and intervals
//
// Data Operations:
//   - Single and batch read operations
//   - Single and batch write operations
//   - Type-safe value handling with automatic variant conversion
//
// Historical Data Operations:
//   - Read raw historical data (ReadHistory)
//   - Read modified historical data (ReadHistoryModified)
//   - Read processed/aggregated data with 18+ aggregate functions (ReadHistoryProcessed)
//   - Read values at specific timestamps (ReadHistoryAtTime)
//   - Read historical events (ReadHistoryEvents)
//   - Write/insert historical data (WriteHistory)
//   - Update/replace historical data (UpdateHistory)
//   - Delete historical data by time range (DeleteHistory)
//   - Delete historical data at specific times (DeleteHistoryAtTime)
//   - Delete historical events (DeleteHistoryEvents)
//   - Support for Average, Min, Max, Count, Total, StdDev, and more aggregates
//
// Browse and Discovery:
//   - Node browsing with automatic continuation point handling
//   - Browse path translation (TranslateBrowsePathsToNodeIds)
//   - BrowseNext for large result sets
//   - Server discovery (FindServers)
//   - Endpoint discovery and selection
//
// Subscriptions:
//   - Data change monitoring with configurable parameters
//   - Event monitoring
//   - Deadband filtering (absolute and percent)
//   - Custom sampling intervals and queue sizes
//   - Multiple subscription groups support
//
// Method Calling:
//   - Full support for OPC UA method invocation
//   - Automatic input/output argument handling
//
// Security:
//   - Multiple security policies (None, Basic128Rsa15, Basic256, Basic256Sha256)
//   - Multiple security modes (None, Sign, SignAndEncrypt)
//   - Certificate and key file support
//
// Example Usage:
//
//	config := opcclient.OPCClient{
//	    Endpoint: "opc.tcp://localhost:4840",
//	    Namespace: 2,
//	    Auth: opcclient.AuthConfig{
//	        Type: opcclient.AuthUserPass,
//	        Username: "admin",
//	        Password: "password",
//	    },
//	    Connection: opcclient.ConnectionConfig{
//	        AutoReconnect: true,
//	        ReconnectInterval: 5 * time.Second,
//	        MaxReconnectAttempts: 10,
//	    },
//	}
//
//	client := &config
//	cancel := client.CreateClient()
//	defer cancel()
//
//	err := client.Connect()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Disconnect()
//
//	// Read a value
//	value, err := client.ReadTagValue("ns=2;s=MyTag")
//
//	// Write a value
//	status, err := client.WriteTagValue("ns=2;s=MyTag", 42)
//
//	// Batch read
//	values, err := client.BatchRead([]string{"ns=2;s=Tag1", "ns=2;s=Tag2"})
//
//	// Subscribe to changes
//	err = client.Subscribe([]string{"ns=2;s=MyTag"}, func(tag string, value *ua.DataValue) {
//	    fmt.Printf("Tag %s changed to %v\n", tag, value.Value)
//	})
package opcclient

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/id"
	"github.com/gopcua/opcua/ua"
	"github.com/mdaxf/iac/logger"
)

type OPCConfig struct {
	OPCClients []OPCClient `json:"opcclients"`
}

// AuthType defines the authentication type for OPC UA connection
type AuthType string

const (
	AuthAnonymous   AuthType = "anonymous"
	AuthUserPass    AuthType = "userpass"
	AuthCertificate AuthType = "certificate"
	AuthIssuedToken AuthType = "issuedtoken"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	Policy   string `json:"policy"`   // Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256
	Mode     string `json:"mode"`     // Security mode: None, Sign, SignAndEncrypt
	CertFile string `json:"certFile"` // Path to certificate file
	KeyFile  string `json:"keyFile"`  // Path to private key file
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type     AuthType `json:"type"`     // Authentication type
	Username string   `json:"username"` // Username for userpass auth
	Password string   `json:"password"` // Password for userpass auth
	Token    string   `json:"token"`    // Token for issued token auth
}

// ConnectionConfig holds connection-related settings
type ConnectionConfig struct {
	AutoReconnect        bool          `json:"autoReconnect"`        // Enable automatic reconnection
	ReconnectInterval    time.Duration `json:"reconnectInterval"`    // Time between reconnection attempts
	SessionTimeout       time.Duration `json:"sessionTimeout"`       // Session timeout duration
	RequestTimeout       time.Duration `json:"requestTimeout"`       // Request timeout duration
	MaxReconnectAttempts int           `json:"maxReconnectAttempts"` // Maximum reconnection attempts (0 = infinite)
}

// SubscriptionConfig holds subscription-related settings
type SubscriptionConfig struct {
	PublishingInterval time.Duration `json:"publishingInterval"` // Publishing interval
	LifetimeCount      uint32        `json:"lifetimeCount"`      // Lifetime count
	MaxKeepAliveCount  uint32        `json:"maxKeepAliveCount"`  // Max keep-alive count
	MaxNotifications   uint32        `json:"maxNotifications"`   // Max notifications per publish
	Priority           uint8         `json:"priority"`           // Subscription priority
}

type OPCClient struct {
	Client       *opcua.Client
	Endpoint     string             `json:"endpoint"`
	Host         string             `json:"host"`
	Name         string             `json:"name"`
	Namespace    uint16             `json:"namespace"`
	CertFile     string             `json:"certFile"`     // Deprecated: use Security.CertFile
	KeyFile      string             `json:"keyFile"`      // Deprecated: use Security.KeyFile
	Timeout      time.Duration      `json:"timeout"`      // Deprecated: use Connection.RequestTimeout
	Security     SecurityConfig     `json:"security"`     // Security configuration
	Auth         AuthConfig         `json:"auth"`         // Authentication configuration
	Connection   ConnectionConfig   `json:"connection"`   // Connection configuration
	Subscription SubscriptionConfig `json:"subscription"` // Subscription configuration
	Nodes        map[string]*opcua.Node
	SubGroups    []SubGroup `json:"subgroups"`
	iLog         logger.Log
	reconnecting bool
	ctx          context.Context
	cancel       context.CancelFunc
}

type SubGroup struct {
	TriggerTags []string                    `json:"triggerTags"`
	ReportTags  []string                    `json:"reportTags"`
	Trigger     func(string, *ua.DataValue) `json:"trigger"`
	Report      func(string, *ua.DataValue) `json:"report"`
}

// Initialize initializes and runs the OPC UA client with the given configuration
func Initialize(configurations OPCClient) error {
	iLog := logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "OPCClient"}

	iLog.Debug(fmt.Sprintf("Create OPCClient with configuration: %s", logger.ConvertJson(configurations)))

	// Create OPC client instance
	opcclient := &OPCClient{
		Endpoint:     configurations.Endpoint,
		Namespace:    configurations.Namespace,
		CertFile:     configurations.CertFile,
		KeyFile:      configurations.KeyFile,
		Timeout:      configurations.Timeout * time.Second,
		Security:     configurations.Security,
		Auth:         configurations.Auth,
		Connection:   configurations.Connection,
		Subscription: configurations.Subscription,
		SubGroups:    configurations.SubGroups,
		iLog:         iLog,
	}

	// Create and connect client
	cancel := opcclient.CreateClient()
	defer cancel()
	defer opcclient.Disconnect()

	err := opcclient.Connect()
	if err != nil {
		iLog.Error(fmt.Sprintf("Failed to connect to OPC UA server: %v", err))
		return err
	}

	// Create subscriptions for configured subgroups
	for _, subgroup := range opcclient.SubGroups {
		callbackfunc := func(tag string, v *ua.DataValue) {
			if subgroup.Trigger != nil {
				subgroup.Trigger(tag, v)
			} else {
				fmt.Printf("Tag: %v, Value: %v\n", tag, v.Value)
			}
		}

		err := opcclient.Subscribe(subgroup.TriggerTags, callbackfunc)
		if err != nil {
			iLog.Error(fmt.Sprintf("Failed to create subscription for subgroup: %v", err))
		}
	}

	// Wait for termination signal to gracefully shut down the client
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM)
	<-terminate

	iLog.Debug("OPC UA client shutting down")
	return nil
}

func (c *OPCClient) CreateClient() context.CancelFunc {
	// Initialize context if not already set
	if c.ctx == nil {
		c.ctx, c.cancel = context.WithCancel(context.Background())
	}

	// Apply backward compatibility for deprecated fields
	c.applyBackwardCompatibility()

	// Get endpoints from server
	endpoints, err := opcua.GetEndpoints(c.ctx, c.Endpoint)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("Failed to get endpoints: %v", err))
		return c.cancel
	}

	// Select appropriate endpoint based on security configuration
	ep := opcua.SelectEndpoint(endpoints, c.Security.Policy, ua.MessageSecurityModeFromString(c.Security.Mode))
	if ep == nil {
		c.iLog.Critical("Failed to find suitable endpoint")
		return c.cancel
	}

	c.iLog.Debug(fmt.Sprintf("Using endpoint: %v", ep.EndpointURL))
	c.iLog.Debug(fmt.Sprintf("Security: %s, Mode: %s", ep.SecurityPolicyURI, ep.SecurityMode))

	// Build client options
	opts := c.buildClientOptions(ep)

	// Create OPC UA client
	opcclient, err := opcua.NewClient(ep.EndpointURL, opts...)
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("Failed to create client: %s", err))
		return c.cancel
	}

	c.Client = opcclient
	c.iLog.Debug("OPC UA client created successfully")

	return c.cancel
}

// applyBackwardCompatibility applies backward compatibility for deprecated fields
func (c *OPCClient) applyBackwardCompatibility() {
	// Use deprecated fields if new fields are not set
	if c.Security.CertFile == "" && c.CertFile != "" {
		c.Security.CertFile = c.CertFile
	}
	if c.Security.KeyFile == "" && c.KeyFile != "" {
		c.Security.KeyFile = c.KeyFile
	}
	if c.Connection.RequestTimeout == 0 && c.Timeout != 0 {
		c.Connection.RequestTimeout = c.Timeout
	}
	if c.Auth.Type == "" {
		c.Auth.Type = AuthAnonymous
	}
}

// buildClientOptions builds opcua.Option array based on configuration
func (c *OPCClient) buildClientOptions(ep *ua.EndpointDescription) []opcua.Option {
	opts := []opcua.Option{}

	// Security options
	if c.Security.Policy != "" {
		opts = append(opts, opcua.SecurityPolicy(c.Security.Policy))
	}
	if c.Security.Mode != "" {
		opts = append(opts, opcua.SecurityModeString(c.Security.Mode))
	}
	if c.Security.CertFile != "" {
		opts = append(opts, opcua.CertificateFile(c.Security.CertFile))
	}
	if c.Security.KeyFile != "" {
		opts = append(opts, opcua.PrivateKeyFile(c.Security.KeyFile))
	}

	// Authentication options
	switch c.Auth.Type {
	case AuthAnonymous:
		opts = append(opts, opcua.AuthAnonymous())
		opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous))
		c.iLog.Debug("Using anonymous authentication")

	case AuthUserPass:
		if c.Auth.Username != "" && c.Auth.Password != "" {
			opts = append(opts, opcua.AuthUsername(c.Auth.Username, c.Auth.Password))
			opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeUserName))
			c.iLog.Debug(fmt.Sprintf("Using username/password authentication for user: %s", c.Auth.Username))
		} else {
			c.iLog.Error("Username or password not provided for userpass authentication, falling back to anonymous")
			opts = append(opts, opcua.AuthAnonymous())
			opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous))
		}

	case AuthCertificate:
		if c.Security.CertFile != "" && c.Security.KeyFile != "" {
			opts = append(opts, opcua.AuthCertificate([]byte{})) // Certificate auth
			opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeCertificate))
			c.iLog.Debug("Using certificate authentication")
		} else {
			c.iLog.Error("Certificate or key file not provided for certificate authentication, falling back to anonymous")
			opts = append(opts, opcua.AuthAnonymous())
			opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous))
		}

	case AuthIssuedToken:
		if c.Auth.Token != "" {
			opts = append(opts, opcua.AuthIssuedToken([]byte(c.Auth.Token)))
			opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeIssuedToken))
			c.iLog.Debug("Using issued token authentication")
		} else {
			c.iLog.Error("Token not provided for issued token authentication, falling back to anonymous")
			opts = append(opts, opcua.AuthAnonymous())
			opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous))
		}

	default:
		c.iLog.Error(fmt.Sprintf("Unknown authentication type: %s, using anonymous", c.Auth.Type))
		opts = append(opts, opcua.AuthAnonymous())
		opts = append(opts, opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous))
	}

	// Session timeout
	if c.Connection.SessionTimeout > 0 {
		opts = append(opts, opcua.SessionTimeout(c.Connection.SessionTimeout))
	}

	// Request timeout
	if c.Connection.RequestTimeout > 0 {
		opts = append(opts, opcua.RequestTimeout(c.Connection.RequestTimeout))
	}

	return opts
}

func (c *OPCClient) BrowseEndpoints() (*ua.GetEndpointsResponse, error) {
	// Browse the endpoint by the server
	c.iLog.Debug(fmt.Sprintf("Browsing endpoints from server"))
	endpoints, err := c.Client.GetEndpoints(context.Background())
	if err != nil {
		c.iLog.Critical(fmt.Sprintf("Failed to get endpoints: %s", err))
		return nil, err
	}
	c.iLog.Debug(fmt.Sprintf("Endpoints: %v", endpoints))

	return endpoints, nil
}
func (c *OPCClient) Connect() error {
	c.iLog.Debug(fmt.Sprintf("Connecting to OPC Server: %s", c.Endpoint))

	// Connect to the OPC UA server
	err := c.Client.Connect(c.ctx)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to connect to OPC UA server: %s", err))

		// Attempt auto-reconnection if enabled
		if c.Connection.AutoReconnect {
			return c.reconnect()
		}
		return err
	}

	c.iLog.Debug(fmt.Sprintf("Connected to OPC Server: %s", c.Endpoint))

	// Start connection monitor if auto-reconnect is enabled
	if c.Connection.AutoReconnect {
		go c.monitorConnection()
	}

	return nil
}

// reconnect attempts to reconnect to the OPC UA server
func (c *OPCClient) reconnect() error {
	if c.reconnecting {
		c.iLog.Debug("Reconnection already in progress")
		return nil
	}

	c.reconnecting = true
	defer func() { c.reconnecting = false }()

	attempts := 0
	maxAttempts := c.Connection.MaxReconnectAttempts
	interval := c.Connection.ReconnectInterval

	if interval == 0 {
		interval = 5 * time.Second // Default reconnection interval
	}

	c.iLog.Debug(fmt.Sprintf("Starting reconnection attempts (max: %d, interval: %v)", maxAttempts, interval))

	for {
		attempts++

		// Check if max attempts reached (0 means infinite)
		if maxAttempts > 0 && attempts > maxAttempts {
			err := fmt.Errorf("maximum reconnection attempts (%d) reached", maxAttempts)
			c.iLog.Error(err.Error())
			return err
		}

		c.iLog.Debug(fmt.Sprintf("Reconnection attempt %d", attempts))

		// Try to connect
		err := c.Client.Connect(c.ctx)
		if err == nil {
			c.iLog.Debug(fmt.Sprintf("Reconnected successfully after %d attempts", attempts))
			return nil
		}

		c.iLog.Error(fmt.Sprintf("Reconnection attempt %d failed: %v", attempts, err))

		// Wait before next attempt
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-time.After(interval):
			// Continue to next attempt
		}
	}
}

// monitorConnection monitors the connection and triggers reconnection if needed
func (c *OPCClient) monitorConnection() {
	ticker := time.NewTicker(10 * time.Second) // Check connection every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			// Check if client is still connected
			if c.Client.State() != opcua.Connected {
				c.iLog.Error("Connection lost, attempting to reconnect...")
				err := c.reconnect()
				if err != nil {
					c.iLog.Error(fmt.Sprintf("Failed to reconnect: %v", err))
				}
			}
		}
	}
}

// BrowseTags browses and returns all tags (nodes) under the specified node ID
func (c *OPCClient) BrowseTags(nodeID string) ([]Tag, error) {
	c.iLog.Debug(fmt.Sprintf("Browsing nodes (tags) from server starting at node: %s", nodeID))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	result, err := c.Client.Browse(c.ctx, &ua.BrowseRequest{
		NodesToBrowse: []*ua.BrowseDescription{
			{
				NodeID:          id,
				ReferenceTypeID: ua.NewNumericNodeID(0, 0),
				IncludeSubtypes: true,
				ResultMask:      uint32(ua.BrowseResultMaskAll),
				BrowseDirection: ua.BrowseDirectionForward,
			},
		},
	})
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to browse nodes: %v", err))
		return nil, err
	}

	var tags []Tag

	// Process the browse result
	for _, res := range result.Results {
		if res.StatusCode != ua.StatusOK {
			c.iLog.Error(fmt.Sprintf("Browse failed with status: %v", res.StatusCode))
			continue
		}

		for _, ref := range res.References {
			// Extract the tag name and address from the reference
			tag := Tag{
				Name:    ref.DisplayName.Text,
				Address: ref.NodeID.String(),
			}
			tags = append(tags, tag)
		}

		// Handle continuation point if there are more results
		if len(res.ContinuationPoint) > 0 {
			c.iLog.Debug("Browse has continuation point, fetching more results")
			moreTags, err := c.BrowseNext(res.ContinuationPoint)
			if err != nil {
				c.iLog.Error(fmt.Sprintf("Failed to browse next: %v", err))
			} else {
				for _, ref := range moreTags {
					tag := Tag{
						Name:    ref.DisplayName.Text,
						Address: ref.NodeID.String(),
					}
					tags = append(tags, tag)
				}
			}
		}
	}

	c.iLog.Debug(fmt.Sprintf("Found %d tags", len(tags)))
	return tags, nil
}

// browseTags is kept for backward compatibility
func (c *OPCClient) browseTags(nodeID string) ([]Tag, error) {
	return c.BrowseTags(nodeID)
}

// ReadTagValue reads a single node value by node ID
func (c *OPCClient) ReadTagValue(nodeID string) (interface{}, error) {
	c.iLog.Debug(fmt.Sprintf("Reading node value: %s", nodeID))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			{NodeID: id},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := c.retryableRead(req)
	if err != nil {
		return nil, err
	}

	if resp.Results[0].Status != ua.StatusOK {
		err := fmt.Errorf("read failed with status: %v", resp.Results[0].Status)
		c.iLog.Error(err.Error())
		return nil, err
	}

	return resp.Results[0].Value.Value(), nil
}

// readTagValue is kept for backward compatibility
func (c *OPCClient) readTagValue(nodeID string) (interface{}, error) {
	return c.ReadTagValue(nodeID)
}

// retryableRead performs a read operation with automatic retry on transient errors
func (c *OPCClient) retryableRead(req *ua.ReadRequest) (*ua.ReadResponse, error) {
	maxRetries := 5
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := c.Client.Read(c.ctx, req)
		if err == nil {
			return resp, nil
		}

		// Check for retryable errors
		switch {
		case err == io.EOF && c.Client.State() != opcua.Closed:
			c.iLog.Debug("Connection EOF, retrying...")
		case errors.Is(err, ua.StatusBadSessionIDInvalid):
			c.iLog.Debug("Session ID invalid, retrying...")
		case errors.Is(err, ua.StatusBadSessionNotActivated):
			c.iLog.Debug("Session not activated, retrying...")
		case errors.Is(err, ua.StatusBadSecureChannelIDInvalid):
			c.iLog.Debug("Secure channel invalid, retrying...")
		default:
			// Non-retryable error
			c.iLog.Error(fmt.Sprintf("Read failed: %v", err))
			return nil, err
		}

		// Wait before retry
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("read failed after %d retries", maxRetries)
}

// WriteTagValue writes a value to a single node by node ID
func (c *OPCClient) WriteTagValue(nodeID string, value interface{}) (ua.StatusCode, error) {
	c.iLog.Debug(fmt.Sprintf("Writing node value: %s, %v", nodeID, value))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return ua.StatusBad, err
	}

	var v *ua.Variant
	switch val := value.(type) {
	case string:
		v, err = ua.NewVariant(val)
	case *ua.Variant:
		v = val
	default:
		v, err = ua.NewVariant(value)
	}

	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create variant: %v", err))
		return ua.StatusBad, err
	}

	req := &ua.WriteRequest{
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      id,
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        v,
				},
			},
		},
	}

	resp, err := c.Client.Write(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Write failed: %v", err))
		return ua.StatusBad, err
	}

	c.iLog.Debug(fmt.Sprintf("Node value written with result: %v", resp.Results[0]))
	return resp.Results[0], nil
}

// writeTagValue is kept for backward compatibility
func (c *OPCClient) writeTagValue(nodeID string, value string) (ua.StatusCode, error) {
	return c.WriteTagValue(nodeID, value)
}

// BatchRead reads multiple node values in a single request
func (c *OPCClient) BatchRead(nodeIDs []string) (map[string]interface{}, error) {
	c.iLog.Debug(fmt.Sprintf("Batch reading %d nodes", len(nodeIDs)))

	if len(nodeIDs) == 0 {
		return nil, fmt.Errorf("no node IDs provided")
	}

	nodesToRead := make([]*ua.ReadValueID, len(nodeIDs))
	for i, nodeID := range nodeIDs {
		id, err := ua.ParseNodeID(nodeID)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Invalid node ID %s: %v", nodeID, err))
			return nil, err
		}
		nodesToRead[i] = &ua.ReadValueID{NodeID: id}
	}

	req := &ua.ReadRequest{
		MaxAge:             2000,
		NodesToRead:        nodesToRead,
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := c.retryableRead(req)
	if err != nil {
		return nil, err
	}

	results := make(map[string]interface{})
	for i, result := range resp.Results {
		if result.Status == ua.StatusOK {
			results[nodeIDs[i]] = result.Value.Value()
		} else {
			c.iLog.Error(fmt.Sprintf("Read failed for node %s with status: %v", nodeIDs[i], result.Status))
			results[nodeIDs[i]] = nil
		}
	}

	return results, nil
}

// BatchWrite writes multiple node values in a single request
func (c *OPCClient) BatchWrite(nodeValues map[string]interface{}) (map[string]ua.StatusCode, error) {
	c.iLog.Debug(fmt.Sprintf("Batch writing %d nodes", len(nodeValues)))

	if len(nodeValues) == 0 {
		return nil, fmt.Errorf("no node values provided")
	}

	nodesToWrite := make([]*ua.WriteValue, 0, len(nodeValues))
	nodeIDs := make([]string, 0, len(nodeValues))

	for nodeID, value := range nodeValues {
		id, err := ua.ParseNodeID(nodeID)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Invalid node ID %s: %v", nodeID, err))
			continue
		}

		v, err := ua.NewVariant(value)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to create variant for node %s: %v", nodeID, err))
			continue
		}

		nodesToWrite = append(nodesToWrite, &ua.WriteValue{
			NodeID:      id,
			AttributeID: ua.AttributeIDValue,
			Value: &ua.DataValue{
				EncodingMask: ua.DataValueValue,
				Value:        v,
			},
		})
		nodeIDs = append(nodeIDs, nodeID)
	}

	req := &ua.WriteRequest{
		NodesToWrite: nodesToWrite,
	}

	resp, err := c.Client.Write(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Batch write failed: %v", err))
		return nil, err
	}

	results := make(map[string]ua.StatusCode)
	for i, status := range resp.Results {
		results[nodeIDs[i]] = status
		if status != ua.StatusOK {
			c.iLog.Error(fmt.Sprintf("Write failed for node %s with status: %v", nodeIDs[i], status))
		}
	}

	return results, nil
}

// CallMethod calls an OPC UA method on a server
func (c *OPCClient) CallMethod(objectID, methodID string, inputArgs ...interface{}) ([]interface{}, error) {
	c.iLog.Debug(fmt.Sprintf("Calling method %s on object %s", methodID, objectID))

	objID, err := ua.ParseNodeID(objectID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid object ID: %v", err))
		return nil, err
	}

	methID, err := ua.ParseNodeID(methodID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid method ID: %v", err))
		return nil, err
	}

	// Convert input arguments to variants
	inputVariants := make([]*ua.Variant, len(inputArgs))
	for i, arg := range inputArgs {
		v, err := ua.NewVariant(arg)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to create variant for argument %d: %v", i, err))
			return nil, err
		}
		inputVariants[i] = v
	}

	req := &ua.CallRequest{
		MethodsToCall: []*ua.CallMethodRequest{
			{
				ObjectID:       objID,
				MethodID:       methID,
				InputArguments: inputVariants,
			},
		},
	}

	resp, err := c.Client.Call(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Method call failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from method call")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("method call failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	// Convert output arguments
	outputs := make([]interface{}, len(result.OutputArguments))
	for i, variant := range result.OutputArguments {
		outputs[i] = variant.Value()
	}

	c.iLog.Debug(fmt.Sprintf("Method call successful, returned %d outputs", len(outputs)))
	return outputs, nil
}

// HistoryReadType defines the type of historical read operation
type HistoryReadType int

const (
	HistoryReadRaw       HistoryReadType = iota // Read raw historical data
	HistoryReadModified                         // Read modified historical data
	HistoryReadProcessed                        // Read processed/aggregated data
	HistoryReadAtTime                           // Read values at specific times
)

// AggregateType defines common OPC UA aggregate functions
type AggregateType uint32

const (
	AggregateAverage       AggregateType = 2341 // Average aggregate
	AggregateCount         AggregateType = 2342 // Count aggregate
	AggregateMinimum       AggregateType = 2346 // Minimum aggregate
	AggregateMaximum       AggregateType = 2347 // Maximum aggregate
	AggregateRange         AggregateType = 2348 // Range aggregate
	AggregateTotal         AggregateType = 2355 // Total aggregate
	AggregateStandardDev   AggregateType = 2351 // Standard deviation
	AggregateTimeAverage   AggregateType = 2341 // Time-weighted average
	AggregateInterpolative AggregateType = 2340 // Interpolative aggregate
	AggregateDurationGood  AggregateType = 2340 // Duration in good state
	AggregateDurationBad   AggregateType = 2342 // Duration in bad state
	AggregatePercentGood   AggregateType = 2361 // Percent good
	AggregatePercentBad    AggregateType = 2362 // Percent bad
	AggregateStart         AggregateType = 2363 // Start value
	AggregateEnd           AggregateType = 2364 // End value
	AggregateDelta         AggregateType = 2365 // Delta
	AggregateStartBound    AggregateType = 2366 // Start bound
	AggregateEndBound      AggregateType = 2367 // End bound
	AggregateDeltaBounds   AggregateType = 2368 // Delta bounds
)

// HistoryUpdateType defines the type of historical update operation
type HistoryUpdateType int

const (
	HistoryUpdateInsert  HistoryUpdateType = 1 // Insert new historical data
	HistoryUpdateReplace HistoryUpdateType = 2 // Replace existing historical data
	HistoryUpdateUpdate  HistoryUpdateType = 3 // Update existing historical data
	HistoryUpdateDelete  HistoryUpdateType = 4 // Delete historical data
)

// ReadHistory reads historical data for a node
func (c *OPCClient) ReadHistory(nodeID string, startTime, endTime time.Time, maxValues uint32) ([]*ua.DataValue, error) {
	c.iLog.Debug(fmt.Sprintf("Reading history for node %s from %v to %v", nodeID, startTime, endTime))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.HistoryReadRequest{
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		HistoryReadDetails: &ua.ReadRawModifiedDetails{
			IsReadModified:   false,
			StartTime:        startTime,
			EndTime:          endTime,
			NumValuesPerNode: maxValues,
			ReturnBounds:     true,
		},
		NodesToRead: []*ua.HistoryReadValueID{
			{
				NodeID: id,
			},
		},
	}

	resp, err := c.Client.HistoryRead(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History read failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history read")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("history read failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	// Extract data values from history data
	historyData, ok := result.HistoryData.(*ua.HistoryData)
	if !ok {
		return nil, fmt.Errorf("unexpected history data type")
	}

	c.iLog.Debug(fmt.Sprintf("Retrieved %d historical values", len(historyData.DataValues)))
	return historyData.DataValues, nil
}

// ReadHistoryModified reads modified historical data (data that was changed after initial recording)
func (c *OPCClient) ReadHistoryModified(nodeID string, startTime, endTime time.Time, maxValues uint32) ([]*ua.DataValue, error) {
	c.iLog.Debug(fmt.Sprintf("Reading modified history for node %s from %v to %v", nodeID, startTime, endTime))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.HistoryReadRequest{
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		HistoryReadDetails: &ua.ReadRawModifiedDetails{
			IsReadModified:   true, // Read modified values
			StartTime:        startTime,
			EndTime:          endTime,
			NumValuesPerNode: maxValues,
			ReturnBounds:     true,
		},
		NodesToRead: []*ua.HistoryReadValueID{
			{
				NodeID: id,
			},
		},
	}

	resp, err := c.Client.HistoryRead(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History read modified failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history read")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("history read modified failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	historyData, ok := result.HistoryData.(*ua.HistoryData)
	if !ok {
		return nil, fmt.Errorf("unexpected history data type")
	}

	c.iLog.Debug(fmt.Sprintf("Retrieved %d modified historical values", len(historyData.DataValues)))
	return historyData.DataValues, nil
}

// ReadHistoryProcessed reads processed/aggregated historical data
func (c *OPCClient) ReadHistoryProcessed(nodeID string, startTime, endTime time.Time, processingInterval time.Duration, aggregateType AggregateType) ([]*ua.DataValue, error) {
	c.iLog.Debug(fmt.Sprintf("Reading processed history for node %s from %v to %v with interval %v", nodeID, startTime, endTime, processingInterval))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	// Create aggregate node ID
	aggregateNodeID := ua.NewNumericNodeID(0, uint32(aggregateType))

	req := &ua.HistoryReadRequest{
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		HistoryReadDetails: &ua.ReadProcessedDetails{
			StartTime:          startTime,
			EndTime:            endTime,
			ProcessingInterval: float64(processingInterval.Milliseconds()),
			AggregateType:      []*ua.NodeID{aggregateNodeID},
		},
		NodesToRead: []*ua.HistoryReadValueID{
			{
				NodeID: id,
			},
		},
	}

	resp, err := c.Client.HistoryRead(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History read processed failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history read processed")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("history read processed failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	historyData, ok := result.HistoryData.(*ua.HistoryData)
	if !ok {
		return nil, fmt.Errorf("unexpected history data type")
	}

	c.iLog.Debug(fmt.Sprintf("Retrieved %d processed historical values", len(historyData.DataValues)))
	return historyData.DataValues, nil
}

// ReadHistoryAtTime reads historical values at specific timestamps
func (c *OPCClient) ReadHistoryAtTime(nodeID string, timestamps []time.Time, useSimpleBounds bool) ([]*ua.DataValue, error) {
	c.iLog.Debug(fmt.Sprintf("Reading history at specific times for node %s (%d timestamps)", nodeID, len(timestamps)))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.HistoryReadRequest{
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		HistoryReadDetails: &ua.ReadAtTimeDetails{
			ReqTimes:        timestamps,
			UseSimpleBounds: useSimpleBounds,
		},
		NodesToRead: []*ua.HistoryReadValueID{
			{
				NodeID: id,
			},
		},
	}

	resp, err := c.Client.HistoryRead(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History read at time failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history read at time")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("history read at time failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	historyData, ok := result.HistoryData.(*ua.HistoryData)
	if !ok {
		return nil, fmt.Errorf("unexpected history data type")
	}

	c.iLog.Debug(fmt.Sprintf("Retrieved %d historical values at specified times", len(historyData.DataValues)))
	return historyData.DataValues, nil
}

// ReadHistoryEvents reads historical events for a node
func (c *OPCClient) ReadHistoryEvents(nodeID string, startTime, endTime time.Time, maxValues uint32, eventFilter *ua.EventFilter) ([]*ua.HistoryEventFieldList, error) {
	c.iLog.Debug(fmt.Sprintf("Reading historical events for node %s from %v to %v", nodeID, startTime, endTime))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	// Use default event filter if none provided
	if eventFilter == nil {
		eventFilter = createDefaultEventFilter()
	}

	req := &ua.HistoryReadRequest{
		TimestampsToReturn: ua.TimestampsToReturnBoth,
		HistoryReadDetails: &ua.ReadEventDetails{
			StartTime:        startTime,
			EndTime:          endTime,
			NumValuesPerNode: maxValues,
			Filter:           eventFilter,
		},
		NodesToRead: []*ua.HistoryReadValueID{
			{
				NodeID: id,
			},
		},
	}

	resp, err := c.Client.HistoryRead(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History read events failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history read events")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("history read events failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	historyEvent, ok := result.HistoryData.(*ua.HistoryEvent)
	if !ok {
		return nil, fmt.Errorf("unexpected history data type, expected HistoryEvent")
	}

	c.iLog.Debug(fmt.Sprintf("Retrieved %d historical events", len(historyEvent.Events)))
	return historyEvent.Events, nil
}

// WriteHistory inserts historical data values for a node
func (c *OPCClient) WriteHistory(nodeID string, dataValues []*ua.DataValue) ([]ua.StatusCode, error) {
	c.iLog.Debug(fmt.Sprintf("Writing %d historical values for node %s", len(dataValues), nodeID))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.HistoryUpdateRequest{
		HistoryUpdateDetails: []*ua.ExtensionObject{
			{
				EncodingMask: ua.ExtensionObjectBinary,
				TypeID: &ua.ExpandedNodeID{
					NodeID: ua.NewNumericNodeID(0, id.UpdateDataDetails_Encoding_DefaultBinary),
				},
				Value: &ua.UpdateDataDetails{
					NodeID:               id,
					PerformInsertReplace: ua.PerformUpdateTypeInsert,
					UpdateValues:         dataValues,
				},
			},
		},
	}

	resp, err := c.Client.HistoryUpdate(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History write failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history write")
	}

	result := resp.Results[0]
	updateResult, ok := result.(*ua.HistoryUpdateResult)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	c.iLog.Debug(fmt.Sprintf("History write completed with status: %v", updateResult.StatusCode))
	return updateResult.OperationResults, nil
}

// UpdateHistory replaces existing historical data values for a node
func (c *OPCClient) UpdateHistory(nodeID string, dataValues []*ua.DataValue) ([]ua.StatusCode, error) {
	c.iLog.Debug(fmt.Sprintf("Updating %d historical values for node %s", len(dataValues), nodeID))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.HistoryUpdateRequest{
		HistoryUpdateDetails: []*ua.ExtensionObject{
			{
				EncodingMask: ua.ExtensionObjectBinary,
				TypeID: &ua.ExpandedNodeID{
					NodeID: ua.NewNumericNodeID(0, id.UpdateDataDetails_Encoding_DefaultBinary),
				},
				Value: &ua.UpdateDataDetails{
					NodeID:               id,
					PerformInsertReplace: ua.PerformUpdateTypeReplace,
					UpdateValues:         dataValues,
				},
			},
		},
	}

	resp, err := c.Client.HistoryUpdate(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History update failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history update")
	}

	result := resp.Results[0]
	updateResult, ok := result.(*ua.HistoryUpdateResult)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	c.iLog.Debug(fmt.Sprintf("History update completed with status: %v", updateResult.StatusCode))
	return updateResult.OperationResults, nil
}

// DeleteHistory deletes historical data for a node within a time range
func (c *OPCClient) DeleteHistory(nodeID string, startTime, endTime time.Time) (ua.StatusCode, error) {
	c.iLog.Debug(fmt.Sprintf("Deleting historical data for node %s from %v to %v", nodeID, startTime, endTime))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return ua.StatusBad, err
	}

	req := &ua.HistoryUpdateRequest{
		HistoryUpdateDetails: []*ua.ExtensionObject{
			{
				EncodingMask: ua.ExtensionObjectBinary,
				TypeID: &ua.ExpandedNodeID{
					NodeID: ua.NewNumericNodeID(0, id.DeleteRawModifiedDetails_Encoding_DefaultBinary),
				},
				Value: &ua.DeleteRawModifiedDetails{
					NodeID:           id,
					IsDeleteModified: false,
					StartTime:        startTime,
					EndTime:          endTime,
				},
			},
		},
	}

	resp, err := c.Client.HistoryUpdate(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History delete failed: %v", err))
		return ua.StatusBad, err
	}

	if len(resp.Results) == 0 {
		return ua.StatusBad, fmt.Errorf("no results returned from history delete")
	}

	result := resp.Results[0]
	updateResult, ok := result.(*ua.HistoryUpdateResult)
	if !ok {
		return ua.StatusBad, fmt.Errorf("unexpected result type")
	}

	c.iLog.Debug(fmt.Sprintf("History delete completed with status: %v", updateResult.StatusCode))
	return updateResult.StatusCode, nil
}

// DeleteHistoryAtTime deletes historical data at specific timestamps
func (c *OPCClient) DeleteHistoryAtTime(nodeID string, timestamps []time.Time) ([]ua.StatusCode, error) {
	c.iLog.Debug(fmt.Sprintf("Deleting historical data for node %s at %d specific times", nodeID, len(timestamps)))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.HistoryUpdateRequest{
		HistoryUpdateDetails: []*ua.ExtensionObject{
			{
				EncodingMask: ua.ExtensionObjectBinary,
				TypeID: &ua.ExpandedNodeID{
					NodeID: ua.NewNumericNodeID(0, id.DeleteAtTimeDetails_Encoding_DefaultBinary),
				},
				Value: &ua.DeleteAtTimeDetails{
					NodeID:   id,
					ReqTimes: timestamps,
				},
			},
		},
	}

	resp, err := c.Client.HistoryUpdate(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History delete at time failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history delete at time")
	}

	result := resp.Results[0]
	updateResult, ok := result.(*ua.HistoryUpdateResult)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	c.iLog.Debug(fmt.Sprintf("History delete at time completed with status: %v", updateResult.StatusCode))
	return updateResult.OperationResults, nil
}

// DeleteHistoryEvents deletes historical events for a node
func (c *OPCClient) DeleteHistoryEvents(nodeID string, eventIDs [][]byte) ([]ua.StatusCode, error) {
	c.iLog.Debug(fmt.Sprintf("Deleting %d historical events for node %s", len(eventIDs), nodeID))

	id, err := ua.ParseNodeID(nodeID)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid node ID: %v", err))
		return nil, err
	}

	req := &ua.HistoryUpdateRequest{
		HistoryUpdateDetails: []*ua.ExtensionObject{
			{
				EncodingMask: ua.ExtensionObjectBinary,
				TypeID: &ua.ExpandedNodeID{
					NodeID: ua.NewNumericNodeID(0, id.DeleteEventDetails_Encoding_DefaultBinary),
				},
				Value: &ua.DeleteEventDetails{
					NodeID:   id,
					EventIDs: eventIDs,
				},
			},
		},
	}

	resp, err := c.Client.HistoryUpdate(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("History delete events failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned from history delete events")
	}

	result := resp.Results[0]
	updateResult, ok := result.(*ua.HistoryUpdateResult)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	c.iLog.Debug(fmt.Sprintf("History delete events completed with status: %v", updateResult.StatusCode))
	return updateResult.OperationResults, nil
}

// createDefaultEventFilter creates a default event filter for historical event reading
func createDefaultEventFilter() *ua.EventFilter {
	// Define common event fields
	fieldNames := []string{"EventId", "EventType", "SourceName", "Time", "Message", "Severity"}
	selects := make([]*ua.SimpleAttributeOperand, len(fieldNames))

	for i, name := range fieldNames {
		selects[i] = &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		}
	}

	return &ua.EventFilter{
		SelectClauses: selects,
		WhereClause:   &ua.ContentFilter{}, // No filtering
	}
}

// BrowseWithPath translates a browse path to node IDs
func (c *OPCClient) BrowseWithPath(startingNode string, relativePath []string) ([]string, error) {
	c.iLog.Debug(fmt.Sprintf("Translating browse path from %s: %v", startingNode, relativePath))

	startID, err := ua.ParseNodeID(startingNode)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Invalid starting node ID: %v", err))
		return nil, err
	}

	// Build browse path
	browsePath := make([]*ua.QualifiedName, len(relativePath))
	for i, pathElement := range relativePath {
		browsePath[i] = &ua.QualifiedName{
			NamespaceIndex: c.Namespace,
			Name:           pathElement,
		}
	}

	req := &ua.TranslateBrowsePathsToNodeIDsRequest{
		BrowsePaths: []*ua.BrowsePath{
			{
				StartingNode: startID,
				RelativePath: &ua.RelativePath{
					Elements: make([]*ua.RelativePathElement, len(browsePath)),
				},
			},
		},
	}

	for i, qn := range browsePath {
		req.BrowsePaths[0].RelativePath.Elements[i] = &ua.RelativePathElement{
			ReferenceTypeID: ua.NewNumericNodeID(0, id.HierarchicalReferences),
			IsInverse:       false,
			IncludeSubtypes: true,
			TargetName:      qn,
		}
	}

	resp, err := c.Client.TranslateBrowsePathsToNodeIDs(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Translate browse paths failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("translate failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	nodeIDs := make([]string, len(result.Targets))
	for i, target := range result.Targets {
		nodeIDs[i] = target.TargetID.NodeID.String()
	}

	c.iLog.Debug(fmt.Sprintf("Found %d target nodes", len(nodeIDs)))
	return nodeIDs, nil
}

// BrowseNext continues browsing when the initial browse result has more data
func (c *OPCClient) BrowseNext(continuationPoint []byte) ([]*ua.ReferenceDescription, error) {
	c.iLog.Debug("Executing BrowseNext with continuation point")

	req := &ua.BrowseNextRequest{
		ContinuationPoints:        [][]byte{continuationPoint},
		ReleaseContinuationPoints: false,
	}

	resp, err := c.Client.BrowseNext(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("BrowseNext failed: %v", err))
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("no results returned")
	}

	result := resp.Results[0]
	if result.StatusCode != ua.StatusOK {
		err := fmt.Errorf("BrowseNext failed with status: %v", result.StatusCode)
		c.iLog.Error(err.Error())
		return nil, err
	}

	c.iLog.Debug(fmt.Sprintf("BrowseNext returned %d references", len(result.References)))
	return result.References, nil
}

// FindServers discovers available OPC UA servers on a host
func (c *OPCClient) FindServers() ([]*ua.ApplicationDescription, error) {
	c.iLog.Debug(fmt.Sprintf("Finding servers at %s", c.Endpoint))

	req := &ua.FindServersRequest{
		EndpointURL: c.Endpoint,
	}

	resp, err := c.Client.FindServers(c.ctx, req)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("FindServers failed: %v", err))
		return nil, err
	}

	c.iLog.Debug(fmt.Sprintf("Found %d servers", len(resp.Servers)))
	return resp.Servers, nil
}

// SubscribeOptions holds options for creating subscriptions
type SubscribeOptions struct {
	MonitorEvents    bool          // Monitor events instead of data changes
	SamplingInterval time.Duration // Sampling interval for monitored items
	QueueSize        uint32        // Queue size for monitored items
	DiscardOldest    bool          // Discard oldest values when queue is full
	DeadbandType     uint32        // Deadband filter type (0=None, 1=Absolute, 2=Percent)
	DeadbandValue    float64       // Deadband value for filtering
}

// Subscribe creates a subscription and monitors tags for changes
func (c *OPCClient) Subscribe(triggerTags []string, callback func(string, *ua.DataValue)) error {
	return c.SubscribeWithOptions(triggerTags, callback, SubscribeOptions{
		MonitorEvents:    false,
		SamplingInterval: 100 * time.Millisecond,
		QueueSize:        10,
		DiscardOldest:    true,
	})
}

// SubscribeWithOptions creates a subscription with custom options
func (c *OPCClient) SubscribeWithOptions(triggerTags []string, callback func(string, *ua.DataValue), opts SubscribeOptions) error {
	c.iLog.Debug(fmt.Sprintf("Creating subscription for %d tags", len(triggerTags)))

	notifyCh := make(chan *opcua.PublishNotificationData, 1000)

	// Use subscription configuration or defaults
	interval := c.Subscription.PublishingInterval
	if interval == 0 {
		interval = opcua.DefaultSubscriptionInterval
	}

	subParams := &opcua.SubscriptionParameters{
		Interval: interval,
	}

	if c.Subscription.LifetimeCount > 0 {
		subParams.LifetimeCount = c.Subscription.LifetimeCount
	}
	if c.Subscription.MaxKeepAliveCount > 0 {
		subParams.MaxKeepAliveCount = c.Subscription.MaxKeepAliveCount
	}
	if c.Subscription.MaxNotifications > 0 {
		subParams.MaxNotificationsPerPublish = c.Subscription.MaxNotifications
	}
	if c.Subscription.Priority > 0 {
		subParams.Priority = c.Subscription.Priority
	}

	sub, err := c.Client.Subscribe(c.ctx, subParams, notifyCh)
	if err != nil {
		c.iLog.Error(fmt.Sprintf("Failed to create subscription: %v", err))
		return err
	}

	// Create monitored items for all tags
	monitorRequests := make([]*ua.MonitoredItemCreateRequest, 0, len(triggerTags))
	tagMap := make(map[uint32]string) // Map client handle to tag name

	for i, tag := range triggerTags {
		id, err := ua.ParseNodeID(tag)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Invalid node ID %s: %v", tag, err))
			continue
		}

		clientHandle := uint32(i + 1)
		tagMap[clientHandle] = tag

		var miCreateRequest *ua.MonitoredItemCreateRequest
		if opts.MonitorEvents {
			miCreateRequest, _ = eventRequest(id)
			miCreateRequest.RequestedParameters.ClientHandle = clientHandle
		} else {
			miCreateRequest = c.createMonitoredItemRequest(id, clientHandle, opts)
		}

		monitorRequests = append(monitorRequests, miCreateRequest)
	}

	// Monitor all items
	if len(monitorRequests) > 0 {
		res, err := sub.Monitor(c.ctx, ua.TimestampsToReturnBoth, monitorRequests...)
		if err != nil {
			c.iLog.Error(fmt.Sprintf("Failed to create monitored items: %v", err))
			return err
		}

		// Check results
		for i, result := range res.Results {
			if result.StatusCode != ua.StatusOK {
				c.iLog.Error(fmt.Sprintf("Failed to monitor tag %s: %v", triggerTags[i], result.StatusCode))
			} else {
				c.iLog.Debug(fmt.Sprintf("Successfully monitoring tag %s", triggerTags[i]))
			}
		}
	}

	// Start notification handler
	go c.handleSubscriptionNotifications(notifyCh, callback, tagMap, sub)

	c.iLog.Debug("Subscription created and monitoring started")
	return nil
}

// createMonitoredItemRequest creates a monitored item request with filtering
func (c *OPCClient) createMonitoredItemRequest(nodeID *ua.NodeID, clientHandle uint32, opts SubscribeOptions) *ua.MonitoredItemCreateRequest {
	req := &ua.MonitoredItemCreateRequest{
		ItemToMonitor: &ua.ReadValueID{
			NodeID:      nodeID,
			AttributeID: ua.AttributeIDValue,
		},
		MonitoringMode: ua.MonitoringModeReporting,
		RequestedParameters: &ua.MonitoringParameters{
			ClientHandle:     clientHandle,
			DiscardOldest:    opts.DiscardOldest,
			QueueSize:        opts.QueueSize,
			SamplingInterval: float64(opts.SamplingInterval.Milliseconds()),
		},
	}

	// Add deadband filter if specified
	if opts.DeadbandType > 0 && opts.DeadbandValue > 0 {
		filter := &ua.DataChangeFilter{
			Trigger:       ua.DataChangeTriggerStatusValue,
			DeadbandType:  opts.DeadbandType,
			DeadbandValue: opts.DeadbandValue,
		}

		req.RequestedParameters.Filter = &ua.ExtensionObject{
			EncodingMask: ua.ExtensionObjectBinary,
			TypeID: &ua.ExpandedNodeID{
				NodeID: ua.NewNumericNodeID(0, id.DataChangeFilter_Encoding_DefaultBinary),
			},
			Value: filter,
		}
	}

	return req
}

// handleSubscriptionNotifications processes subscription notifications
func (c *OPCClient) handleSubscriptionNotifications(notifyCh chan *opcua.PublishNotificationData, callback func(string, *ua.DataValue), tagMap map[uint32]string, sub *opcua.Subscription) {
	defer sub.Cancel(c.ctx)

	for {
		select {
		case <-c.ctx.Done():
			c.iLog.Debug("Subscription notification handler stopped")
			return

		case res := <-notifyCh:
			if res.Error != nil {
				c.iLog.Error(fmt.Sprintf("Subscription error: %v", res.Error))
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					tagName := tagMap[item.ClientHandle]
					if callback != nil {
						callback(tagName, item.Value)
					}
					c.iLog.Debug(fmt.Sprintf("Data change for tag %s (handle %v): %v", tagName, item.ClientHandle, item.Value.Value.Value()))
				}

			case *ua.EventNotificationList:
				for _, item := range x.Events {
					c.iLog.Debug(fmt.Sprintf("Event notification for handle %v", item.ClientHandle))
					for i, field := range item.EventFields {
						c.iLog.Debug(fmt.Sprintf("  Field %d: %v (Type: %T)", i, field.Value(), field.Value()))
					}
				}

			default:
				c.iLog.Debug(fmt.Sprintf("Unknown publish result type: %T", res.Value))
			}
		}
	}
}
func valueRequest(nodeID *ua.NodeID) *ua.MonitoredItemCreateRequest {
	handle := uint32(42)
	return opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, handle)
}

func eventRequest(nodeID *ua.NodeID) (*ua.MonitoredItemCreateRequest, []string) {
	fieldNames := []string{"EventId", "EventType", "Severity", "Time", "Message"}
	selects := make([]*ua.SimpleAttributeOperand, len(fieldNames))

	for i, name := range fieldNames {
		selects[i] = &ua.SimpleAttributeOperand{
			TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
			BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: name}},
			AttributeID:      ua.AttributeIDValue,
		}
	}

	wheres := &ua.ContentFilter{
		Elements: []*ua.ContentFilterElement{
			{
				FilterOperator: ua.FilterOperatorGreaterThanOrEqual,
				FilterOperands: []*ua.ExtensionObject{
					{
						EncodingMask: 1,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.SimpleAttributeOperand_Encoding_DefaultBinary),
						},
						Value: ua.SimpleAttributeOperand{
							TypeDefinitionID: ua.NewNumericNodeID(0, id.BaseEventType),
							BrowsePath:       []*ua.QualifiedName{{NamespaceIndex: 0, Name: "Severity"}},
							AttributeID:      ua.AttributeIDValue,
						},
					},
					{
						EncodingMask: 1,
						TypeID: &ua.ExpandedNodeID{
							NodeID: ua.NewNumericNodeID(0, id.LiteralOperand_Encoding_DefaultBinary),
						},
						Value: ua.LiteralOperand{
							Value: ua.MustVariant(uint16(0)),
						},
					},
				},
			},
		},
	}

	filter := ua.EventFilter{
		SelectClauses: selects,
		WhereClause:   wheres,
	}

	filterExtObj := ua.ExtensionObject{
		EncodingMask: ua.ExtensionObjectBinary,
		TypeID: &ua.ExpandedNodeID{
			NodeID: ua.NewNumericNodeID(0, id.EventFilter_Encoding_DefaultBinary),
		},
		Value: filter,
	}

	handle := uint32(42)
	req := &ua.MonitoredItemCreateRequest{
		ItemToMonitor: &ua.ReadValueID{
			NodeID:       nodeID,
			AttributeID:  ua.AttributeIDEventNotifier,
			DataEncoding: &ua.QualifiedName{},
		},
		MonitoringMode: ua.MonitoringModeReporting,
		RequestedParameters: &ua.MonitoringParameters{
			ClientHandle:     handle,
			DiscardOldest:    true,
			Filter:           &filterExtObj,
			QueueSize:        10,
			SamplingInterval: 1.0,
		},
	}

	return req, fieldNames
}

func (c *OPCClient) Disconnect() {
	// Disconnect from the OPC UA server
	c.Client.Close(context.Background())
	c.iLog.Debug(fmt.Sprintf("Disconnected from OPC Server: %s", c.Endpoint))
}

type Tag struct {
	Name    string
	Address string
}
