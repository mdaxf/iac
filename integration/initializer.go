package integration

import (
	"fmt"
	"path/filepath"

	"github.com/mdaxf/iac/integration/datahub"
	"github.com/mdaxf/iac/integration/graphql"
	"github.com/mdaxf/iac/integration/rest"
	"github.com/mdaxf/iac/integration/soap"
	"github.com/mdaxf/iac/integration/tcp"
	"github.com/sirupsen/logrus"
)

// IntegrationConfig holds configuration for all integrations
type IntegrationConfig struct {
	ConfigPath  string `json:"config_path"`
	EnableREST  bool   `json:"enable_rest"`
	EnableSOAP  bool   `json:"enable_soap"`
	EnableTCP   bool   `json:"enable_tcp"`
	EnableGraphQL bool `json:"enable_graphql"`
	EnableDataHub bool `json:"enable_datahub"`

	// Existing integrations
	EnableMQTT     bool `json:"enable_mqtt"`
	EnableKafka    bool `json:"enable_kafka"`
	EnableActiveMQ bool `json:"enable_activemq"`
	EnableOPCUA    bool `json:"enable_opcua"`
	EnableSignalR  bool `json:"enable_signalr"`
}

// IntegrationManager manages all integration services
type IntegrationManager struct {
	Config      *IntegrationConfig
	Logger      *logrus.Logger

	// New integrations
	RESTClient     *rest.RESTClient
	RESTServer     *rest.RESTServer
	SOAPClient     *soap.SOAPClient
	SOAPServer     *soap.SOAPServer
	TCPClient      *tcp.TCPClient
	TCPServer      *tcp.TCPServer
	GraphQLClient  *graphql.GraphQLClient
	GraphQLServer  *graphql.GraphQLServer
	DataHub        *datahub.DataHub

	initialized    bool
}

var (
	globalManager *IntegrationManager
)

// NewIntegrationManager creates a new integration manager
func NewIntegrationManager(config *IntegrationConfig, logger *logrus.Logger) *IntegrationManager {
	if logger == nil {
		logger = logrus.New()
	}

	return &IntegrationManager{
		Config:      config,
		Logger:      logger,
		initialized: false,
	}
}

// GetGlobalManager returns the global integration manager
func GetGlobalManager() *IntegrationManager {
	return globalManager
}

// SetGlobalManager sets the global integration manager
func SetGlobalManager(manager *IntegrationManager) {
	globalManager = manager
}

// Initialize initializes all enabled integrations
func (im *IntegrationManager) Initialize() error {
	if im.initialized {
		return fmt.Errorf("integration manager already initialized")
	}

	im.Logger.Info("Initializing integration manager...")

	// Initialize DataHub first if enabled (other integrations may depend on it)
	if im.Config.EnableDataHub {
		if err := im.initializeDataHub(); err != nil {
			im.Logger.Errorf("Failed to initialize DataHub: %v", err)
			return err
		}
	}

	// Initialize REST if enabled
	if im.Config.EnableREST {
		if err := im.initializeREST(); err != nil {
			im.Logger.Errorf("Failed to initialize REST: %v", err)
			// Don't return error, continue with other integrations
		}
	}

	// Initialize SOAP if enabled
	if im.Config.EnableSOAP {
		if err := im.initializeSOAP(); err != nil {
			im.Logger.Errorf("Failed to initialize SOAP: %v", err)
		}
	}

	// Initialize TCP if enabled
	if im.Config.EnableTCP {
		if err := im.initializeTCP(); err != nil {
			im.Logger.Errorf("Failed to initialize TCP: %v", err)
		}
	}

	// Initialize GraphQL if enabled
	if im.Config.EnableGraphQL {
		if err := im.initializeGraphQL(); err != nil {
			im.Logger.Errorf("Failed to initialize GraphQL: %v", err)
		}
	}

	im.initialized = true
	im.Logger.Info("Integration manager initialized successfully")

	return nil
}

// initializeDataHub initializes the DataHub
func (im *IntegrationManager) initializeDataHub() error {
	im.Logger.Info("Initializing DataHub...")

	// Create DataHub instance
	im.DataHub = datahub.NewDataHub(im.Logger)

	// Load mappings if config file exists
	mappingFile := filepath.Join(im.Config.ConfigPath, "datahub", "mappings.json")
	if err := im.DataHub.LoadMappingsFromFile(mappingFile); err != nil {
		im.Logger.Warnf("Failed to load mappings from %s: %v", mappingFile, err)
		// Don't fail initialization if mappings file doesn't exist
	}

	im.Logger.Info("DataHub initialized")
	return nil
}

// initializeREST initializes REST client and server
func (im *IntegrationManager) initializeREST() error {
	im.Logger.Info("Initializing REST integration...")

	// Initialize REST Client
	clientConfigFile := filepath.Join(im.Config.ConfigPath, "rest", "client.json")
	if err := rest.InitRESTClient(clientConfigFile, im.Logger); err != nil {
		im.Logger.Warnf("Failed to initialize REST client from %s: %v", clientConfigFile, err)
	} else {
		im.RESTClient = rest.GetGlobalRESTClient()

		// Register with DataHub if enabled
		if im.Config.EnableDataHub && im.DataHub != nil {
			adapter := datahub.NewRESTAdapter(im.RESTClient)
			im.DataHub.RegisterAdapter("REST", adapter)
		}
	}

	// Initialize REST Server
	serverConfigFile := filepath.Join(im.Config.ConfigPath, "rest", "server.json")
	if err := rest.InitRESTServer(serverConfigFile, im.Logger); err != nil {
		im.Logger.Warnf("Failed to initialize REST server from %s: %v", serverConfigFile, err)
	} else {
		im.RESTServer = rest.GetGlobalRESTServer()
	}

	im.Logger.Info("REST integration initialized")
	return nil
}

// initializeSOAP initializes SOAP client and server
func (im *IntegrationManager) initializeSOAP() error {
	im.Logger.Info("Initializing SOAP integration...")

	// Initialize SOAP Client
	clientConfigFile := filepath.Join(im.Config.ConfigPath, "soap", "client.json")
	if err := soap.InitSOAPClient(clientConfigFile, im.Logger); err != nil {
		im.Logger.Warnf("Failed to initialize SOAP client from %s: %v", clientConfigFile, err)
	} else {
		im.SOAPClient = soap.GetGlobalSOAPClient()

		// Register with DataHub if enabled
		if im.Config.EnableDataHub && im.DataHub != nil {
			adapter := datahub.NewSOAPAdapter(im.SOAPClient)
			im.DataHub.RegisterAdapter("SOAP", adapter)
		}
	}

	// Initialize SOAP Server
	serverConfigFile := filepath.Join(im.Config.ConfigPath, "soap", "server.json")
	if err := soap.InitSOAPServer(serverConfigFile, im.Logger); err != nil {
		im.Logger.Warnf("Failed to initialize SOAP server from %s: %v", serverConfigFile, err)
	} else {
		im.SOAPServer = soap.GetGlobalSOAPServer()
	}

	im.Logger.Info("SOAP integration initialized")
	return nil
}

// initializeTCP initializes TCP client and server
func (im *IntegrationManager) initializeTCP() error {
	im.Logger.Info("Initializing TCP integration...")

	// Initialize TCP Client
	clientConfigFile := filepath.Join(im.Config.ConfigPath, "tcp", "client.json")
	if err := tcp.InitTCPClient(clientConfigFile, im.Logger); err != nil {
		im.Logger.Warnf("Failed to initialize TCP client from %s: %v", clientConfigFile, err)
	} else {
		im.TCPClient = tcp.GetGlobalTCPClient()

		// Register with DataHub if enabled
		if im.Config.EnableDataHub && im.DataHub != nil {
			adapter := datahub.NewTCPAdapter(im.TCPClient)
			im.DataHub.RegisterAdapter("TCP", adapter)
		}
	}

	// Initialize TCP Server
	serverConfigFile := filepath.Join(im.Config.ConfigPath, "tcp", "server.json")
	if err := tcp.InitTCPServer(serverConfigFile, im.Logger); err != nil {
		im.Logger.Warnf("Failed to initialize TCP server from %s: %v", serverConfigFile, err)
	} else {
		im.TCPServer = tcp.GetGlobalTCPServer()
	}

	im.Logger.Info("TCP integration initialized")
	return nil
}

// initializeGraphQL initializes GraphQL client and server
func (im *IntegrationManager) initializeGraphQL() error {
	im.Logger.Info("Initializing GraphQL integration...")

	// Initialize GraphQL Client
	clientConfigFile := filepath.Join(im.Config.ConfigPath, "graphql", "client.json")
	if err := graphql.InitGraphQLClient(clientConfigFile, im.Logger); err != nil {
		im.Logger.Warnf("Failed to initialize GraphQL client from %s: %v", clientConfigFile, err)
	} else {
		im.GraphQLClient = graphql.GetGlobalGraphQLClient()

		// Register with DataHub if enabled
		if im.Config.EnableDataHub && im.DataHub != nil {
			adapter := datahub.NewGraphQLAdapter(im.GraphQLClient)
			im.DataHub.RegisterAdapter("GraphQL", adapter)
		}
	}

	// Note: GraphQL server requires schema configuration
	// This should be configured separately by the application
	im.Logger.Info("GraphQL integration initialized (server requires separate schema configuration)")

	return nil
}

// StartServers starts all enabled servers
func (im *IntegrationManager) StartServers() error {
	im.Logger.Info("Starting integration servers...")

	// Start REST server
	if im.RESTServer != nil {
		if err := im.RESTServer.Start(); err != nil {
			im.Logger.Errorf("Failed to start REST server: %v", err)
		}
	}

	// Start SOAP server
	if im.SOAPServer != nil {
		if err := im.SOAPServer.Start(); err != nil {
			im.Logger.Errorf("Failed to start SOAP server: %v", err)
		}
	}

	// Start TCP server
	if im.TCPServer != nil {
		if err := im.TCPServer.Start(); err != nil {
			im.Logger.Errorf("Failed to start TCP server: %v", err)
		}
	}

	// Start GraphQL server
	if im.GraphQLServer != nil {
		if err := im.GraphQLServer.Start(); err != nil {
			im.Logger.Errorf("Failed to start GraphQL server: %v", err)
		}
	}

	im.Logger.Info("Integration servers started")
	return nil
}

// StopServers stops all running servers
func (im *IntegrationManager) StopServers() error {
	im.Logger.Info("Stopping integration servers...")

	// Stop REST server
	if im.RESTServer != nil && im.RESTServer.IsEnabled() {
		if err := im.RESTServer.Stop(); err != nil {
			im.Logger.Errorf("Failed to stop REST server: %v", err)
		}
	}

	// Stop SOAP server
	if im.SOAPServer != nil && im.SOAPServer.IsEnabled() {
		if err := im.SOAPServer.Stop(); err != nil {
			im.Logger.Errorf("Failed to stop SOAP server: %v", err)
		}
	}

	// Stop TCP server
	if im.TCPServer != nil && im.TCPServer.IsEnabled() {
		if err := im.TCPServer.Stop(); err != nil {
			im.Logger.Errorf("Failed to stop TCP server: %v", err)
		}
	}

	// Stop GraphQL server
	if im.GraphQLServer != nil && im.GraphQLServer.IsEnabled() {
		if err := im.GraphQLServer.Stop(); err != nil {
			im.Logger.Errorf("Failed to stop GraphQL server: %v", err)
		}
	}

	im.Logger.Info("Integration servers stopped")
	return nil
}

// Shutdown shuts down all integrations
func (im *IntegrationManager) Shutdown() error {
	im.Logger.Info("Shutting down integration manager...")

	// Stop all servers
	if err := im.StopServers(); err != nil {
		im.Logger.Errorf("Failed to stop servers: %v", err)
	}

	// Close clients
	if im.RESTClient != nil {
		im.RESTClient.Close()
	}
	if im.SOAPClient != nil {
		im.SOAPClient.Close()
	}
	if im.TCPClient != nil {
		im.TCPClient.Close()
	}
	if im.GraphQLClient != nil {
		im.GraphQLClient.Close()
	}

	// Close DataHub
	if im.DataHub != nil {
		im.DataHub.Close()
	}

	im.initialized = false
	im.Logger.Info("Integration manager shut down")

	return nil
}

// HealthCheck checks the health of all integrations
func (im *IntegrationManager) HealthCheck() map[string]bool {
	health := make(map[string]bool)

	// Check REST client
	if im.RESTClient != nil && im.RESTClient.IsEnabled() {
		health["rest_client"] = im.RESTClient.Health() == nil
	}

	// Check SOAP client
	if im.SOAPClient != nil && im.SOAPClient.IsEnabled() {
		health["soap_client"] = im.SOAPClient.Health() == nil
	}

	// Check TCP client
	if im.TCPClient != nil && im.TCPClient.IsEnabled() {
		health["tcp_client"] = im.TCPClient.Health() == nil
	}

	// Check GraphQL client
	if im.GraphQLClient != nil && im.GraphQLClient.IsEnabled() {
		health["graphql_client"] = im.GraphQLClient.Health() == nil
	}

	// Check servers
	if im.RESTServer != nil {
		health["rest_server"] = im.RESTServer.IsEnabled()
	}
	if im.SOAPServer != nil {
		health["soap_server"] = im.SOAPServer.IsEnabled()
	}
	if im.TCPServer != nil {
		health["tcp_server"] = im.TCPServer.IsEnabled()
	}
	if im.GraphQLServer != nil {
		health["graphql_server"] = im.GraphQLServer.IsEnabled()
	}

	// Check DataHub
	if im.DataHub != nil {
		health["datahub"] = im.DataHub.IsEnabled()
	}

	return health
}

// GetStatus returns the status of all integrations
func (im *IntegrationManager) GetStatus() map[string]interface{} {
	status := make(map[string]interface{})

	status["initialized"] = im.initialized
	status["health"] = im.HealthCheck()

	if im.DataHub != nil {
		status["datahub_enabled"] = im.DataHub.IsEnabled()
		history := im.DataHub.GetMessageHistory(10)
		status["recent_transformations"] = len(history)
	}

	return status
}
