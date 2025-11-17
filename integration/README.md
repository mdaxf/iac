# IAC Integration Framework

A comprehensive integration framework supporting modern protocols and a powerful data hub for message transformation and routing.

## Table of Contents

- [Overview](#overview)
- [Supported Protocols](#supported-protocols)
- [Data Hub](#data-hub)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Examples](#examples)
- [Architecture](#architecture)

## Overview

The IAC Integration Framework provides a unified approach to integrating with various protocols and services. It includes:

- **Protocol Adapters**: REST, SOAP, GraphQL, TCP, MQTT, Kafka, ActiveMQ, OPC UA, SignalR
- **Data Hub**: Central message transformation and routing engine
- **Message Mapping**: Transform messages between different schemas and protocols
- **Routing Rules**: Intelligent message routing based on conditions
- **Job Integration**: Zero-code, configuration-driven automatic message processing
- **Background Workers**: Distributed job processing with retry and error handling

## Supported Protocols

### 1. REST Web Services

**Client Features:**
- HTTP methods: GET, POST, PUT, PATCH, DELETE
- Custom headers and query parameters
- JSON request/response handling
- Configurable timeouts
- DataHub integration

**Server Features:**
- HTTP server with routing
- Middleware support (logging, recovery, CORS, auth)
- RESTful API endpoints
- Request/Response context handling

**Configuration Example:**
```json
{
  "client": {
    "base_url": "https://api.example.com",
    "timeout": 30,
    "headers": {
      "Authorization": "Bearer TOKEN"
    },
    "datahub_mode": true
  },
  "server": {
    "port": 8080,
    "datahub_mode": true
  }
}
```

**Usage:**
```go
// Client
client := rest.NewRESTClient(config, logger)
response, err := client.GET("/api/users", nil)
response, err := client.POST("/api/orders", orderData)

// Server
server := rest.NewRESTServer(config, logger)
server.POST("/api/orders", func(ctx *rest.Context) error {
    return ctx.JSON(200, map[string]string{"status": "created"})
})
server.Start()
```

### 2. SOAP Web Services

**Client Features:**
- SOAP 1.1/1.2 support
- WSDL-based or manual envelope construction
- Custom SOAP headers
- Fault handling
- DataHub integration

**Server Features:**
- SOAP action routing
- Automatic envelope parsing
- SOAP fault generation
- XML marshalling/unmarshalling

**Configuration Example:**
```json
{
  "client": {
    "url": "http://soapservice.example.com/OrderService",
    "namespace": "http://schemas.xmlsoap.org/soap/envelope/",
    "timeout": 30,
    "datahub_mode": true
  },
  "server": {
    "port": 8081,
    "datahub_mode": true
  }
}
```

**Usage:**
```go
// Client
client := soap.NewSOAPClient(config, logger)
var response OrderResponse
err := client.Call("CreateOrder", &orderRequest, &response)

// Server
server := soap.NewSOAPServer(config, logger)
server.RegisterHandler("CreateOrder", func(ctx *soap.SOAPContext) error {
    var req OrderRequest
    ctx.ParseRequest(&req)
    ctx.SetResponse(&OrderResponse{Status: "Created"})
    return nil
})
server.Start()
```

### 3. TCP Integration

**Client Features:**
- Raw TCP socket communication
- Delimiter-based message framing
- Auto-reconnect support
- Send/Receive operations
- Binary and text protocols

**Server Features:**
- Multi-client TCP server
- Connection pooling
- Broadcast capabilities
- Custom message handlers
- Delimiter-based message parsing

**Configuration Example:**
```json
{
  "client": {
    "host": "localhost",
    "port": 9000,
    "timeout": 30,
    "delimiter": "\\n",
    "auto_reconnect": true,
    "datahub_mode": true
  },
  "server": {
    "host": "0.0.0.0",
    "port": 9000,
    "delimiter": "\\n",
    "datahub_mode": true
  }
}
```

**Usage:**
```go
// Client
client := tcp.NewTCPClient(config, logger)
client.Connect()
client.SendWithDelimiter([]byte("Hello"))
data, err := client.ReceiveUntilDelimiter()

// Server
server := tcp.NewTCPServer(config, logger)
server.SetHandler(func(conn net.Conn, data []byte, logger *logrus.Logger) ([]byte, error) {
    return []byte("Response"), nil
})
server.Start()
```

### 4. GraphQL

**Client Features:**
- Query and mutation execution
- Variable support
- Operation naming
- Schema introspection
- Subscription support (WebSocket)
- DataHub integration

**Server Features:**
- GraphQL schema builder
- Query/Mutation/Subscription support
- GraphiQL IDE
- Custom resolvers
- Context-aware execution

**Configuration Example:**
```json
{
  "client": {
    "endpoint": "https://api.example.com/graphql",
    "timeout": 30,
    "headers": {
      "Authorization": "Bearer TOKEN"
    },
    "datahub_mode": true
  },
  "server": {
    "port": 4000,
    "datahub_mode": true
  }
}
```

**Usage:**
```go
// Client
client := graphql.NewGraphQLClient(config, logger)
response, err := client.Query(`
    query GetUser($id: ID!) {
        user(id: $id) {
            id
            email
            name
        }
    }
`, map[string]interface{}{"id": "123"})

// Server
builder := graphql.NewSchemaBuilder(logger)
builder.AddQueryField("user", &graphql.Field{
    Type: userType,
    Args: graphql.FieldConfigArgument{
        "id": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.ID)},
    },
    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
        // Fetch user logic
        return user, nil
    },
})
schema, _ := builder.Build()
server := graphql.NewGraphQLServer(config, schema, logger)
server.Start()
```

### 5. MQTT

See existing implementation in `integration/mqttclient/`

### 6. Kafka

See existing implementation in `integration/kafka/`

### 7. ActiveMQ

See existing implementation in `integration/activemq/`

## Data Hub

The Data Hub is a central message transformation and routing engine that enables seamless integration between different protocols and message schemas.

### Features

- **Protocol Agnostic**: Works with any protocol adapter
- **Message Transformation**: Transform messages between different schemas
- **Schema Mapping**: Define field-level mappings with JSONPath/XPath
- **Array Mapping**: Advanced nested array iteration and transformation
- **Built-in Functions**: 20+ transformation functions (date, string, number, binary)
- **Custom Scripts**: JavaScript/Lua support for complex transformations
- **Routing Rules**: Conditional message routing
- **Message History**: Audit trail of all transformations
- **Health Monitoring**: Built-in health checks for all adapters

### Message Envelope

All messages in the Data Hub are wrapped in a universal envelope:

```go
type MessageEnvelope struct {
    ID            string                 // Unique message ID
    Protocol      string                 // Source protocol (REST, SOAP, etc.)
    Source        string                 // Source endpoint
    Destination   string                 // Destination endpoint
    Timestamp     time.Time              // Message timestamp
    ContentType   string                 // Content type
    Headers       map[string]interface{} // Protocol headers
    Body          interface{}            // Message body
    OriginalBody  []byte                 // Original binary body
    Metadata      map[string]interface{} // Additional metadata
    TransformPath []string               // Transformation history
}
```

### Mapping Definition

Define how messages should be transformed:

```json
{
  "id": "rest-to-soap-order",
  "name": "REST Order to SOAP Order Transformation",
  "source_protocol": "REST",
  "source_schema": "OrderV1",
  "target_protocol": "SOAP",
  "target_schema": "OrderServiceV2",
  "mappings": [
    {
      "source_path": "$.order.id",
      "target_path": "//OrderRequest/OrderID",
      "data_type": "string",
      "required": true
    },
    {
      "source_path": "$.order.total_amount",
      "target_path": "//OrderRequest/TotalAmount",
      "data_type": "float",
      "required": true,
      "transform_func": "round_to_2_decimals"
    }
  ],
  "transformations": [
    {
      "type": "enrich",
      "description": "Add processing timestamp",
      "config": {
        "target_path": "//OrderRequest/ProcessedAt",
        "value": "{{current_timestamp}}"
      }
    }
  ]
}
```

### Built-in Transform Functions

**String Functions:**
- `to_upper`, `to_lower`, `trim`, `substring`, `concat`, `replace`

**Number Functions:**
- `round_to_2_decimals`, `round_to_n_decimals`, `to_int`, `to_float`

**Date/Time Functions:**
- `iso8601_to_soap_datetime`, `unix_timestamp_to_iso8601`, `current_timestamp_iso8601`, `format_date`

**Binary Functions:**
- `bytes_to_hex`, `bytes_to_float32`, `bytes_to_string`

**Protocol Functions:**
- `mqtt_topic_to_channel`, `soap_envelope_wrap`, `rest_to_graphql_query`

**Array/Object Functions:**
- `array_join`, `array_filter`, `object_merge`

### Advanced Array Mapping

The DataHub supports comprehensive array mapping for complex nested data structures. This is essential for transforming messages with multiple levels of arrays, such as Orders → Operations → Parts → WIS/Tools.

**Array Mapping Modes:**
- **iterate**: Process each array item with item-specific mappings
- **flatten**: Flatten nested arrays into a single array
- **filter**: Filter array items based on conditions
- **merge**: Merge array of objects into a single object
- **expand**: Expand objects into array items

**Key Features:**
- Unlimited nesting depth (arrays within arrays)
- Optional node handling (gracefully handle missing fields)
- Filtering, sorting, and limiting
- Grouping and aggregation
- Relative and absolute path references

**Example: Nested Array Mapping**

Source structure: Orders → Operations → Parts → WIS

```json
{
  "source_path": "$.orders",
  "target_path": "$.ProcessOrders",
  "data_type": "array",
  "array_mapping": {
    "mode": "iterate",
    "item_mappings": [
      {
        "source_path": ".order_id",
        "target_path": "$.OrderNumber",
        "data_type": "string",
        "required": true
      },
      {
        "source_path": ".operations",
        "target_path": "$.Operations",
        "data_type": "array",
        "optional": true,
        "array_mapping": {
          "mode": "iterate",
          "item_mappings": [
            {
              "source_path": ".operation_id",
              "target_path": "$.OpCode",
              "data_type": "string"
            },
            {
              "source_path": ".parts",
              "target_path": "$.Parts",
              "data_type": "array",
              "optional": true,
              "array_mapping": {
                "mode": "iterate",
                "item_mappings": [
                  {
                    "source_path": ".part_id",
                    "target_path": "$.PartNumber",
                    "data_type": "string"
                  },
                  {
                    "source_path": ".wis",
                    "target_path": "$.WorkInstructions",
                    "data_type": "array",
                    "optional": true,
                    "default_value": []
                  }
                ]
              }
            }
          ]
        }
      }
    ]
  }
}
```

**Handling Missing/Optional Nodes:**

Use `optional: true` and `default_value` for fields that may not exist:

```json
{
  "source_path": ".tools",
  "target_path": "$.ToolsRequired",
  "data_type": "array",
  "optional": true,
  "default_value": []
}
```

**Additional Array Operations:**

```json
{
  "array_mapping": {
    "mode": "iterate",
    "sort_by": "priority",
    "sort_order": "desc",
    "limit": 10,
    "filter_condition": {
      "field": "status",
      "operator": "eq",
      "value": "active"
    },
    "group_by": "part_id",
    "aggregate_func": "sum"
  }
}
```

For detailed array mapping documentation and complex examples, see:
- `integration/datahub/ARRAY_MAPPING_GUIDE.md` - Comprehensive guide
- `integration/datahub/complex_array_mapping_example.json` - Real-world examples

### Routing Rules

Define how messages should be routed:

```json
{
  "id": "route-orders-to-erp",
  "name": "Route Orders to ERP System",
  "source": "REST:/api/orders",
  "destination": "SOAP:http://erp.company.com/OrderService",
  "conditions": [
    {
      "field": "$.order.status",
      "operator": "eq",
      "value": "confirmed"
    }
  ],
  "mapping_id": "rest-to-soap-order",
  "priority": 100,
  "active": true
}
```

### Job Integration - Zero-Code Automation

The DataHub includes **zero-code job integration** that automatically creates and processes jobs when messages are received or need to be sent. **No programming required!**

**Key Features:**
- Automatic job creation from incoming messages
- Configuration-driven transformation and routing
- Scheduled jobs with cron expressions
- Built-in retry and error handling
- Distributed worker processing
- Full audit trail

**Example Configuration:**

Define jobs in `config/integration/datahub/jobs.json`:

```json
{
  "jobs": [
    {
      "id": "rest-to-soap-orders",
      "name": "REST Orders to SOAP ERP",
      "enabled": true,
      "type": "transform",
      "trigger": {
        "type": "on_receive",
        "protocol": "REST",
        "topic": "/api/orders"
      },
      "protocol": "REST",
      "destination": "SOAP:http://erp.example.com/OrderService",
      "mapping_id": "rest-to-soap-order",
      "priority": 10,
      "max_retries": 3,
      "auto_route": true
    }
  ]
}
```

**What This Does:**
1. Monitors REST endpoint `/api/orders` for incoming messages
2. Automatically creates a job when a message arrives
3. Applies the `rest-to-soap-order` transformation mapping
4. Routes the transformed message to the SOAP ERP system
5. Retries up to 3 times on failure
6. All without writing any code!

**Job Types:**
- **transform**: Transform and route messages
- **receive**: Poll for new messages
- **send**: Send messages to destinations
- **route**: Route through DataHub rules

**Trigger Types:**
- **on_receive**: When message arrives
- **on_schedule**: Cron-based scheduling
- **on_event**: System events
- **manual**: API triggered

**Priority Levels:**
- 1-3: Low priority (background tasks)
- 4-6: Normal priority (regular operations)
- 7-9: High priority (important transactions)
- 10+: Critical priority (urgent processing)

For complete job integration documentation, see:
- `integration/datahub/JOB_INTEGRATION_GUIDE.md` - Comprehensive guide
- `integration/datahub/job_config_example.json` - Configuration examples

### Usage

```go
// Initialize Data Hub
hub := datahub.NewDataHub(logger)

// Register protocol adapters
hub.RegisterAdapter("REST", datahub.NewRESTAdapter(restClient))
hub.RegisterAdapter("SOAP", datahub.NewSOAPAdapter(soapClient))
hub.RegisterAdapter("GraphQL", datahub.NewGraphQLAdapter(graphqlClient))

// Load mappings from file
hub.LoadMappingsFromFile("config/mappings.json")

// Route a message
envelope := datahub.CreateEnvelope("REST", "/api/orders", "", "application/json", orderData)
err := hub.RouteMessage(envelope)

// Get message history
history := hub.GetMessageHistory(10)
```

## Configuration

### Global Configuration Structure

```
iac/integration/
├── datahub/
│   ├── schema.go
│   ├── datahub.go
│   ├── transform_engine.go
│   ├── adapters.go
│   └── mapping_example.json
├── rest/
│   ├── rest_client.go
│   ├── rest_server.go
│   └── config_example.json
├── soap/
│   ├── soap_client.go
│   ├── soap_server.go
│   └── config_example.json
├── tcp/
│   ├── tcp_client.go
│   ├── tcp_server.go
│   └── config_example.json
├── graphql/
│   ├── graphql_client.go
│   ├── graphql_server.go
│   └── config_example.json
└── README.md
```

## Examples

### Example 1: REST to SOAP Integration

```go
// Setup REST server to receive orders
restServer := rest.NewRESTServer(restConfig, logger)
restServer.POST("/api/orders", func(ctx *rest.Context) error {
    // Order received via REST
    order := ctx.BodyJSON

    // Create envelope for DataHub
    envelope := datahub.CreateEnvelope("REST", "/api/orders", "", "application/json", order)

    // Route through DataHub (will transform to SOAP)
    hub := datahub.GetGlobalDataHub()
    err := hub.RouteMessage(envelope)

    return ctx.JSON(200, map[string]string{"status": "routed"})
})
```

### Example 2: TCP to REST Integration

```go
// Setup TCP server to receive sensor data
tcpServer := tcp.NewTCPServer(tcpConfig, logger)
tcpServer.SetHandler(func(conn net.Conn, data []byte, logger *logrus.Logger) ([]byte, error) {
    // Parse binary sensor data
    sensorData := parseBinarySensorData(data)

    // Create envelope for DataHub
    envelope := datahub.CreateEnvelope("TCP", conn.RemoteAddr().String(), "", "application/octet-stream", sensorData)

    // Route through DataHub (will transform to REST API call)
    hub := datahub.GetGlobalDataHub()
    hub.RouteMessage(envelope)

    return []byte("OK"), nil
})
```

### Example 3: GraphQL to Kafka Integration

```go
// Setup GraphQL mutation to publish events
mutation := &graphql.Field{
    Type: userType,
    Args: graphql.FieldConfigArgument{
        "email": &graphql.ArgumentConfig{Type: graphql.String},
    },
    Resolve: func(p graphql.ResolveParams) (interface{}, error) {
        // Create user
        user := createUser(p.Args["email"].(string))

        // Publish to Kafka via DataHub
        envelope := datahub.CreateEnvelope("GraphQL", "/graphql", "Kafka:user-events", "application/json", user)
        hub := datahub.GetGlobalDataHub()
        hub.RouteMessage(envelope)

        return user, nil
    },
}
```

## Architecture

### Message Flow

```
┌─────────────┐
│   Source    │ (REST, SOAP, TCP, GraphQL, MQTT, Kafka, etc.)
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────────────┐
│              Protocol Adapter                        │
│  - Normalizes protocol-specific messages             │
│  - Creates MessageEnvelope                           │
└──────┬──────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────┐
│                  Data Hub                            │
│  ┌────────────────────────────────────────────────┐ │
│  │  Routing Engine                                │ │
│  │  - Matches routing rules                       │ │
│  │  - Evaluates conditions                        │ │
│  └──────┬─────────────────────────────────────────┘ │
│         │                                            │
│         ▼                                            │
│  ┌────────────────────────────────────────────────┐ │
│  │  Transform Engine                              │ │
│  │  - Applies field mappings                      │ │
│  │  - Executes transform functions                │ │
│  │  - Enriches/filters data                       │ │
│  └──────┬─────────────────────────────────────────┘ │
│         │                                            │
│         ▼                                            │
│  ┌────────────────────────────────────────────────┐ │
│  │  Message History                               │ │
│  │  - Records transformation events               │ │
│  │  - Audit trail                                 │ │
│  └────────────────────────────────────────────────┘ │
└──────┬──────────────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────────────┐
│           Destination Adapter                        │
│  - Converts envelope to target protocol              │
│  - Sends to destination                              │
└──────┬──────────────────────────────────────────────┘
       │
       ▼
┌─────────────┐
│ Destination │ (REST, SOAP, TCP, GraphQL, MQTT, Kafka, etc.)
└─────────────┘
```

### Key Components

1. **Protocol Adapters**: Convert protocol-specific messages to/from MessageEnvelope
2. **Data Hub**: Central routing and transformation engine
3. **Transform Engine**: Executes field mappings and transformations
4. **Routing Rules**: Determines message flow based on conditions
5. **Message History**: Tracks all transformations for auditing

## Best Practices

1. **Use DataHub Mode**: Enable `datahub_mode` for automatic message routing
2. **Define Clear Mappings**: Create explicit field mappings for better maintainability
3. **Version Your Schemas**: Include version info in schema names
4. **Monitor Health**: Regularly check adapter health status
5. **Review History**: Use message history for debugging and auditing
6. **Test Transformations**: Validate mappings with sample data before deployment
7. **Handle Errors**: Implement proper error handling in custom handlers
8. **Secure Connections**: Use TLS/SSL for production deployments

## Dependencies

Required Go packages:
- `github.com/sirupsen/logrus` - Logging
- `github.com/google/uuid` - UUID generation
- `github.com/tidwall/gjson` - JSON path queries
- `github.com/tidwall/sjson` - JSON path updates
- `github.com/gorilla/mux` - HTTP routing
- `github.com/graphql-go/graphql` - GraphQL support

## License

This integration framework is part of the IAC project.

## Support

For issues and questions, please refer to the main IAC documentation or contact the development team.
