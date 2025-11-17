package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/mdaxf/iac/integration/datahub"
	"github.com/sirupsen/logrus"
)

// GraphQLServer is a GraphQL server
type GraphQLServer struct {
	Port        int
	Schema      graphql.Schema
	Server      *http.Server
	Logger      *logrus.Logger
	enabled     bool
	datahubMode bool
}

// GraphQLServerConfig is the configuration for GraphQL server
type GraphQLServerConfig struct {
	Port        int  `json:"port"`
	DatahubMode bool `json:"datahub_mode"`
}

var (
	globalGraphQLServer *GraphQLServer
)

// NewGraphQLServer creates a new GraphQL server
func NewGraphQLServer(config GraphQLServerConfig, schema graphql.Schema, logger *logrus.Logger) *GraphQLServer {
	if logger == nil {
		logger = logrus.New()
	}

	return &GraphQLServer{
		Port:        config.Port,
		Schema:      schema,
		Logger:      logger,
		enabled:     false,
		datahubMode: config.DatahubMode,
	}
}

// InitGraphQLServer initializes the global GraphQL server from config file
func InitGraphQLServer(configFile string, schema graphql.Schema, logger *logrus.Logger) error {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config GraphQLServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	globalGraphQLServer = NewGraphQLServer(config, schema, logger)
	return nil
}

// GetGlobalGraphQLServer returns the global GraphQL server instance
func GetGlobalGraphQLServer() *GraphQLServer {
	return globalGraphQLServer
}

// SetSchema sets the GraphQL schema
func (s *GraphQLServer) SetSchema(schema graphql.Schema) {
	s.Schema = schema
	s.Logger.Info("GraphQL schema updated")
}

// Start starts the GraphQL server
func (s *GraphQLServer) Start() error {
	if s.enabled {
		return fmt.Errorf("GraphQL server already running")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", s.handleGraphQL)
	mux.HandleFunc("/graphiql", s.handleGraphiQL)

	s.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.enabled = true
	s.Logger.Infof("Starting GraphQL server on port %d", s.Port)

	go func() {
		if err := s.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.Logger.Errorf("GraphQL server error: %v", err)
		}
	}()

	return nil
}

// handleGraphQL handles GraphQL requests
func (s *GraphQLServer) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			s.Logger.Errorf("GraphQL handler panic: %v", err)
			s.sendError(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
		}
	}()

	// Only accept POST and GET requests
	if r.Method != "POST" && r.Method != "GET" {
		s.sendError(w, "Only POST and GET methods are allowed", http.StatusMethodNotAllowed)
		return
	}

	var request GraphQLRequest

	if r.Method == "POST" {
		// Read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			s.Logger.Errorf("Failed to read request body: %v", err)
			s.sendError(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Parse request
		if err := json.Unmarshal(body, &request); err != nil {
			s.Logger.Errorf("Failed to parse request: %v", err)
			s.sendError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		// GET request - parse query parameters
		request.Query = r.URL.Query().Get("query")
		request.OperationName = r.URL.Query().Get("operationName")

		variablesStr := r.URL.Query().Get("variables")
		if variablesStr != "" {
			if err := json.Unmarshal([]byte(variablesStr), &request.Variables); err != nil {
				s.Logger.Errorf("Failed to parse variables: %v", err)
				s.sendError(w, "Invalid variables JSON", http.StatusBadRequest)
				return
			}
		}
	}

	s.Logger.Debugf("GraphQL query: %s", request.Query)

	// Execute query
	params := graphql.Params{
		Schema:         s.Schema,
		RequestString:  request.Query,
		VariableValues: request.Variables,
		OperationName:  request.OperationName,
		Context:        r.Context(),
	}

	result := graphql.Do(params)

	// If datahub mode is enabled, send to datahub
	if s.datahubMode {
		envelope := datahub.CreateEnvelope(
			"GraphQL",
			r.URL.Path,
			"",
			"application/json",
			result.Data,
		)
		envelope.Metadata["query"] = request.Query
		envelope.Metadata["operation_name"] = request.OperationName
		envelope.Metadata["has_errors"] = len(result.Errors) > 0
		envelope.Metadata["remote_addr"] = r.RemoteAddr

		hub := datahub.GetGlobalDataHub()
		if hub.IsEnabled() {
			if err := hub.RouteMessage(envelope); err != nil {
				s.Logger.Warnf("Failed to route message to datahub: %v", err)
			}
		}
	}

	// Convert result to response format
	response := GraphQLResponse{
		Data: result.Data,
	}

	if len(result.Errors) > 0 {
		response.Errors = make([]GraphQLError, len(result.Errors))
		for i, err := range result.Errors {
			response.Errors[i] = GraphQLError{
				Message: err.Error(),
			}
		}
	}

	// Send response
	s.sendJSON(w, response, http.StatusOK)
}

// handleGraphiQL serves the GraphiQL IDE
func (s *GraphQLServer) handleGraphiQL(w http.ResponseWriter, r *http.Request) {
	graphiqlHTML := `
<!DOCTYPE html>
<html>
<head>
	<title>GraphiQL</title>
	<link href="https://unpkg.com/graphiql/graphiql.min.css" rel="stylesheet" />
	<script crossorigin src="https://unpkg.com/react/umd/react.production.min.js"></script>
	<script crossorigin src="https://unpkg.com/react-dom/umd/react-dom.production.min.js"></script>
	<script crossorigin src="https://unpkg.com/graphiql/graphiql.min.js"></script>
</head>
<body style="margin: 0;">
	<div id="graphiql" style="height: 100vh;"></div>
	<script>
		const fetcher = GraphiQL.createFetcher({
			url: '/graphql',
		});

		ReactDOM.render(
			React.createElement(GraphiQL, { fetcher: fetcher }),
			document.getElementById('graphiql'),
		);
	</script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(graphiqlHTML))
}

// sendJSON sends a JSON response
func (s *GraphQLServer) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.Logger.Errorf("Failed to encode response: %v", err)
	}
}

// sendError sends an error response
func (s *GraphQLServer) sendError(w http.ResponseWriter, message string, statusCode int) {
	response := GraphQLResponse{
		Errors: []GraphQLError{
			{
				Message: message,
			},
		},
	}
	s.sendJSON(w, response, statusCode)
}

// Stop stops the GraphQL server
func (s *GraphQLServer) Stop() error {
	if !s.enabled {
		return fmt.Errorf("GraphQL server not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	s.enabled = false
	s.Logger.Info("GraphQL server stopped")
	return nil
}

// IsEnabled returns whether the GraphQL server is enabled
func (s *GraphQLServer) IsEnabled() bool {
	return s.enabled
}

// Close closes the GraphQL server
func (s *GraphQLServer) Close() error {
	return s.Stop()
}

// SchemaBuilder helps build GraphQL schemas
type SchemaBuilder struct {
	QueryFields    graphql.Fields
	MutationFields graphql.Fields
	Types          map[string]*graphql.Object
	logger         *logrus.Logger
}

// NewSchemaBuilder creates a new schema builder
func NewSchemaBuilder(logger *logrus.Logger) *SchemaBuilder {
	return &SchemaBuilder{
		QueryFields:    graphql.Fields{},
		MutationFields: graphql.Fields{},
		Types:          make(map[string]*graphql.Object),
		logger:         logger,
	}
}

// AddQueryField adds a query field
func (sb *SchemaBuilder) AddQueryField(name string, field *graphql.Field) {
	sb.QueryFields[name] = field
	sb.logger.Infof("Added query field: %s", name)
}

// AddMutationField adds a mutation field
func (sb *SchemaBuilder) AddMutationField(name string, field *graphql.Field) {
	sb.MutationFields[name] = field
	sb.logger.Infof("Added mutation field: %s", name)
}

// AddType adds a custom type
func (sb *SchemaBuilder) AddType(name string, typeObj *graphql.Object) {
	sb.Types[name] = typeObj
	sb.logger.Infof("Added type: %s", name)
}

// Build builds the GraphQL schema
func (sb *SchemaBuilder) Build() (graphql.Schema, error) {
	rootQuery := graphql.ObjectConfig{
		Name:   "Query",
		Fields: sb.QueryFields,
	}

	schemaConfig := graphql.SchemaConfig{
		Query: graphql.NewObject(rootQuery),
	}

	// Add mutations if any
	if len(sb.MutationFields) > 0 {
		rootMutation := graphql.ObjectConfig{
			Name:   "Mutation",
			Fields: sb.MutationFields,
		}
		schemaConfig.Mutation = graphql.NewObject(rootMutation)
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		return graphql.Schema{}, fmt.Errorf("failed to create schema: %w", err)
	}

	sb.logger.Info("GraphQL schema built successfully")
	return schema, nil
}
