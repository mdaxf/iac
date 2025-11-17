package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/mdaxf/iac/integration/datahub"
	"github.com/sirupsen/logrus"
)

// GraphQLClient is a GraphQL client
type GraphQLClient struct {
	Endpoint    string
	Timeout     time.Duration
	Headers     map[string]string
	Client      *http.Client
	Logger      *logrus.Logger
	enabled     bool
	datahubMode bool
}

// GraphQLClientConfig is the configuration for GraphQL client
type GraphQLClientConfig struct {
	Endpoint    string            `json:"endpoint"`
	Timeout     int               `json:"timeout"` // seconds
	Headers     map[string]string `json:"headers"`
	DatahubMode bool              `json:"datahub_mode"`
}

var (
	globalGraphQLClient *GraphQLClient
)

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data   interface{}            `json:"data"`
	Errors []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                   `json:"message"`
	Locations  []GraphQLErrorLocation   `json:"locations,omitempty"`
	Path       []interface{}            `json:"path,omitempty"`
	Extensions map[string]interface{}   `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents an error location
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// NewGraphQLClient creates a new GraphQL client
func NewGraphQLClient(config GraphQLClientConfig, logger *logrus.Logger) *GraphQLClient {
	if logger == nil {
		logger = logrus.New()
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &GraphQLClient{
		Endpoint: config.Endpoint,
		Timeout:  timeout,
		Headers:  config.Headers,
		Client: &http.Client{
			Timeout: timeout,
		},
		Logger:      logger,
		enabled:     true,
		datahubMode: config.DatahubMode,
	}
}

// InitGraphQLClient initializes the global GraphQL client from config file
func InitGraphQLClient(configFile string, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config GraphQLClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalGraphQLClient = NewGraphQLClient(config, logger)
	return nil
}

// GetGlobalGraphQLClient returns the global GraphQL client instance
func GetGlobalGraphQLClient() *GraphQLClient {
	return globalGraphQLClient
}

// Query executes a GraphQL query
func (c *GraphQLClient) Query(query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	return c.Execute(&GraphQLRequest{
		Query:     query,
		Variables: variables,
	})
}

// Mutate executes a GraphQL mutation
func (c *GraphQLClient) Mutate(mutation string, variables map[string]interface{}) (*GraphQLResponse, error) {
	return c.Execute(&GraphQLRequest{
		Query:     mutation,
		Variables: variables,
	})
}

// Execute executes a GraphQL request
func (c *GraphQLClient) Execute(request *GraphQLRequest) (*GraphQLResponse, error) {
	if !c.enabled {
		return nil, fmt.Errorf("GraphQL client is disabled")
	}

	c.Logger.Infof("Executing GraphQL request to %s", c.Endpoint)
	c.Logger.Debugf("Query: %s", request.Query)

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", c.Endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	for k, v := range c.Headers {
		httpReq.Header.Set(k, v)
	}

	// Send request
	httpResp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	responseBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.Logger.Debugf("Response: %s", string(responseBody))

	// Parse response
	var response GraphQLResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for errors
	if len(response.Errors) > 0 {
		c.Logger.Warnf("GraphQL returned %d errors", len(response.Errors))
		for _, err := range response.Errors {
			c.Logger.Warnf("  - %s", err.Message)
		}
	}

	// If datahub mode is enabled, send to datahub
	if c.datahubMode {
		envelope := datahub.CreateEnvelope(
			"GraphQL",
			c.Endpoint,
			"",
			"application/json",
			response.Data,
		)
		envelope.Metadata["query"] = request.Query
		envelope.Metadata["operation_name"] = request.OperationName
		envelope.Metadata["has_errors"] = len(response.Errors) > 0
		envelope.Metadata["status_code"] = httpResp.StatusCode

		hub := datahub.GetGlobalDataHub()
		if hub.IsEnabled() {
			if err := hub.RouteMessage(envelope); err != nil {
				c.Logger.Warnf("Failed to route message to datahub: %v", err)
			}
		}
	}

	c.Logger.Infof("GraphQL request completed successfully")

	return &response, nil
}

// QueryWithOperationName executes a GraphQL query with operation name
func (c *GraphQLClient) QueryWithOperationName(query string, operationName string, variables map[string]interface{}) (*GraphQLResponse, error) {
	return c.Execute(&GraphQLRequest{
		Query:         query,
		Variables:     variables,
		OperationName: operationName,
	})
}

// Subscribe creates a GraphQL subscription (WebSocket-based)
// Note: This is a placeholder - full WebSocket implementation would be more complex
func (c *GraphQLClient) Subscribe(subscription string, variables map[string]interface{}, handler func(*GraphQLResponse)) error {
	c.Logger.Warn("GraphQL subscriptions require WebSocket support - not fully implemented")
	return fmt.Errorf("GraphQL subscriptions not yet implemented")
}

// Enable enables the GraphQL client
func (c *GraphQLClient) Enable() {
	c.enabled = true
	c.Logger.Info("GraphQL client enabled")
}

// Disable disables the GraphQL client
func (c *GraphQLClient) Disable() {
	c.enabled = false
	c.Logger.Info("GraphQL client disabled")
}

// IsEnabled returns whether the GraphQL client is enabled
func (c *GraphQLClient) IsEnabled() bool {
	return c.enabled
}

// Close closes the GraphQL client
func (c *GraphQLClient) Close() error {
	c.Client.CloseIdleConnections()
	c.Logger.Info("GraphQL client closed")
	return nil
}

// GetProtocolName returns the protocol name
func (c *GraphQLClient) GetProtocolName() string {
	return "GraphQL"
}

// Initialize initializes the adapter with configuration
func (c *GraphQLClient) Initialize(config map[string]interface{}) error {
	if endpoint, ok := config["endpoint"].(string); ok {
		c.Endpoint = endpoint
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

// Receive receives a message (not applicable for GraphQL client, returns error)
func (c *GraphQLClient) Receive(timeout time.Duration) (*datahub.MessageEnvelope, error) {
	return nil, fmt.Errorf("GraphQL client does not support receive operation")
}

// Health checks the health of the GraphQL client
func (c *GraphQLClient) Health() error {
	// Simple introspection query to check if server is alive
	query := `
		query {
			__schema {
				queryType {
					name
				}
			}
		}
	`

	_, err := c.Query(query, nil)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// Helper methods for common queries

// IntrospectSchema retrieves the GraphQL schema
func (c *GraphQLClient) IntrospectSchema() (*GraphQLResponse, error) {
	query := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
				directives {
					name
					description
					locations
					args {
						...InputValue
					}
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
			possibleTypes {
				...TypeRef
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	return c.Query(query, nil)
}
