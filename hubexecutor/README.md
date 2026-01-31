# Integration Hub Executor

Backend execution engine for integration hub configurations.

## Overview

The Hub Executor system can execute tasks per integration hub configuration by instance name. One executor instance runs all configured inbound and outbound protocols.

## Architecture

```
+------------------------+
|  HubExecutorManager    |  Manages multiple executors
+------------------------+
         |
         v
+------------------------+
|     HubExecutor        |  Per-instance execution
+------------------------+
         |
    +----+----+----+----+
    |    |    |    |    |
    v    v    v    v    v
+------+------+------+------+------+
| Web  | MQTT | Kafka| OPC  | TCP  |  Protocol Handlers
|Server| Hdlr | Hdlr | UA   | Hdlr |
+------+------+------+------+------+
         |
         v
+------------------------+
| DestinationExecutor    |  Execute configured actions
+------------------------+
         |
    +----+----+----+----+
    v    v    v    v    v
  Job  Transcode  Outbound  Script
```

## Files

| File | Description |
|------|-------------|
| `types.go` | Core types, interfaces, and status enums |
| `destination.go` | Destination executor for Job, Transcode, Outbound, Script |
| `handlers.go` | Protocol handlers (WebServer, MQTT, Kafka, ActiveMQ, OPC UA, TCP, SignalR) |
| `hub_executor.go` | Main executor that manages handlers per hub |
| `manager.go` | Manager for multiple hub executors |

## Protocol Handlers

| Handler | Protocol | Status |
|---------|----------|--------|
| WebServerHandler | REST, SOAP, GraphQL | ✅ Implemented |
| MQTTHandler | MQTT | ✅ Implemented |
| KafkaHandler | Kafka | 🔧 Skeleton |
| ActiveMQHandler | ActiveMQ | 🔧 Skeleton |
| OPCUAHandler | OPC UA | ✅ Implemented (needs OPC UA client integration) |
| TCPHandler | TCP | 🔧 Skeleton |
| SignalRHandler | SignalR | 🔧 Skeleton |

## Destination Types

| Type | Description |
|------|-------------|
| Job | Execute IAC jobs with parameters |
| Execute Transcode | Transform data using mapping definitions |
| Route to Outbound | Forward to external HTTP endpoints |
| Custom Script | Execute JavaScript using otto VM |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/hubexecutor/start/:instanceName` | Start executor by instance name |
| POST | `/hubexecutor/start-by-id/:hubId` | Start executor by hub ID |
| POST | `/hubexecutor/stop/:instanceName` | Stop executor |
| POST | `/hubexecutor/restart/:instanceName` | Restart executor |
| GET | `/hubexecutor/status/:instanceName` | Get executor status |
| GET | `/hubexecutor/info/:instanceName` | Get executor detailed info |
| GET | `/hubexecutor/list` | List all running executors |
| POST | `/hubexecutor/start-all` | Start all enabled hubs |
| POST | `/hubexecutor/stop-all` | Stop all executors |

## Usage

```go
// Initialize manager
manager := hubexecutor.InitGlobalManager(hubexecutor.ManagerConfig{
    HubLoader: hubexecutor.NewMongoHubLoader(hubexecutor.MongoHubLoaderConfig{
        GetHubByInstanceName: service.GetDefaultHub,
        GetHubByID:           service.GetHub,
        GetAllHubs:           service.GetDefaultEnabledHubs,
    }),
})

// Set job executor
manager.SetJobExecutor(func(ctx context.Context, jobName string, params map[string]interface{}) error {
    return jobService.ExecuteJob(ctx, jobName, params)
})

// Start executor for instance
err := manager.StartExecutor("my-hub-instance")

// Get status
info, _ := manager.GetExecutorInfo("my-hub-instance")

// Stop all
manager.StopAll()
```

## Route Registration

```go
controller := hubexecutor.NewHubExecutorController(manager)
controller.RegisterRoutes(router.Group("/api"))
```

## OPC UA Handler Features

- Parse OPC tags from protocol group configuration
- Action trigger support: interval, startup, shutdown, datachange
- Condition script evaluation (JavaScript)
- Monitored tags configuration

## Configuration

Hub configuration uses the `IntegrationHub` model:
- `Inbound.ProtocolGroups[]` - Inbound protocol configurations
- `Outbound.ProtocolGroups[]` - Outbound protocol configurations
- `ProtocolGroup.BrokerConfig` - Message broker settings
- `ProtocolGroup.Endpoints[]` - Individual endpoints
- `Endpoint.Handler.RouteIDs[]` - Destination job references
- `Endpoint.OverrideConfig.destinations[]` - Custom destinations
- `Endpoint.OverrideConfig.opc_tags[]` - OPC UA tags
- `Endpoint.OverrideConfig.action_trigger` - Action trigger config

## Instance Roles

The hub executor integrates with the IAC instance roles system. Configure the `integration_hub` role in `configuration.json`:

```json
{
  "roles": {
    "roles": [
      {
        "type": "integration_hub",
        "enabled": true,
        "integration_hub": {
          "instance_names": ["my-hub-1", "my-hub-2"],
          "auto_start": true,
          "enable_api": true
        }
      }
    ]
  }
}
```

### Role Configuration Options

| Field | Type | Description |
|-------|------|-------------|
| `instance_names` | string[] | Hub instance names to run (empty = all enabled defaults) |
| `auto_start` | bool | Auto-start hub executors on instance startup |
| `enable_api` | bool | Enable the hub executor management API endpoints |

### Available Roles

| Role | Description |
|------|-------------|
| `main` | Main API endpoints and portal |
| `integration_hub` | Integration hub execution |
| `signalr` | SignalR server |
| `job_executor` | Background job processing |

One instance can run multiple roles. At least one role must be enabled.

## Role-Based Execution

The hub executor is now fully integrated with the IAC instance roles system. When the `integration_hub` role is enabled:

1. The `HubExecutorManager` is automatically initialized with the proper hub loader
2. Hub executor API endpoints are registered if `enable_api` is true
3. Hub executors auto-start if `auto_start` is true

### How It Works

The role initialization happens in `main.go` via the `RoleInitializer`:

```go
// Integration Hub role callback
roleInitializer.SetIntegrationHubInitializer(
    func(ctx context.Context, config *configuration.IntegrationHubRoleConfig, router *gin.Engine) error {
        // Initialize the hub executor manager
        manager := hubexecutor.InitGlobalManager(hubexecutor.ManagerConfig{
            HubLoader: hubexecutor.NewMongoHubLoader(hubexecutor.MongoHubLoaderConfig{
                GetHubByInstanceName: services.GetDefaultHub,
                GetHubByID:           services.GetHub,
                GetAllHubs:           services.GetDefaultEnabledHubs,
            }),
        })

        // Set job executor callback for executing IAC jobs
        manager.SetJobExecutor(func(ctx context.Context, jobName string, params map[string]interface{}) error {
            return jobqueue.ExecuteJobByName(ctx, jobName, params)
        })

        // ... register routes and auto-start
    },
    func(ctx context.Context) error {
        // Shutdown logic
    },
)
```
