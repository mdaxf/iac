# DataHub Job Integration Guide

## Overview

The DataHub Job Integration system provides **zero-code, configuration-driven** message processing and routing. When messages are received or need to be sent, jobs are automatically created and executed based on your configuration - no code required!

## Table of Contents

- [Key Features](#key-features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Job Types](#job-types)
- [Trigger Types](#trigger-types)
- [Configuration](#configuration)
- [Examples](#examples)
- [Integration with Protocols](#integration-with-protocols)
- [Best Practices](#best-practices)

## Key Features

### ✅ Zero-Code Configuration

Define your integration logic entirely in JSON configuration files. No programming required!

### ✅ Automatic Job Creation

Jobs are automatically created when:
- Messages arrive from any protocol (REST, SOAP, TCP, GraphQL, MQTT, Kafka, ActiveMQ)
- Scheduled times occur (cron-based)
- System events happen
- Manual triggers are invoked

### ✅ Built-in Message Transformation

Apply DataHub mappings automatically to transform messages between different schemas and protocols.

### ✅ Automatic Routing

Route messages to destinations based on configurable rules - no custom code needed.

### ✅ Retry and Error Handling

Built-in retry logic with configurable max attempts and failure handling.

### ✅ Distributed Processing

Jobs are processed by background workers with distributed locking for high availability.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│         Message Source (REST, SOAP, TCP, etc.)          │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│         IntegrationJobCreator (Auto-Detect)             │
│  - Matches incoming message to job configurations       │
│  - Creates job(s) in queue automatically                │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│              Job Queue (Database + Cache)                │
│  - Jobs stored in database                              │
│  - Distributed queue for worker coordination            │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│                 Job Workers (Pool)                       │
│  - Poll for pending jobs                                │
│  - Process jobs in parallel                             │
│  - Retry on failure                                     │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│              DataHub Job Handler                         │
│  - ExecuteTransformJob()                                │
│  - ExecuteReceiveJob()                                  │
│  - ExecuteSendJob()                                     │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│                    DataHub                               │
│  - Apply transformations                                │
│  - Route to destinations                                │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────┐
│       Destination (SOAP, Kafka, REST, etc.)             │
└─────────────────────────────────────────────────────────┘
```

## Quick Start

### Step 1: Create Job Configuration

Create `/config/integration/datahub/jobs.json`:

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
      "source": "/api/orders",
      "destination": "SOAP:http://erp.example.com/OrderService",
      "mapping_id": "rest-to-soap-order",
      "routing_rule": "route-orders-to-erp",
      "priority": 10,
      "max_retries": 3,
      "timeout": 60,
      "auto_route": true
    }
  ]
}
```

### Step 2: Initialize System

```go
// Initialize DataHub
hub := datahub.NewDataHub(logger)
hub.LoadMappingsFromFile("config/datahub/mappings.json")

// Initialize Job Config Manager
jobConfigMgr := datahub.NewJobConfigManager(hub, logger)
jobConfigMgr.LoadFromFile("config/integration/datahub/jobs.json")

// Initialize Integration Job Creator
jobCreator := datahub.NewIntegrationJobCreator(jobConfigMgr, db, logger)

// Initialize Job Handler
jobHandler := datahub.NewJobHandler(hub, db, docDB, logger)
```

### Step 3: That's It!

Messages will now be automatically:
1. Received from REST endpoint `/api/orders`
2. Jobs created automatically based on configuration
3. Processed by workers
4. Transformed using the mapping
5. Routed to SOAP ERP system

**No code required!**

## Job Types

### 1. Transform Jobs

Transform messages using DataHub mappings and optionally route them.

```json
{
  "type": "transform",
  "mapping_id": "rest-to-soap-order",
  "routing_rule": "route-orders-to-erp"
}
```

**Use Cases:**
- Convert REST to SOAP
- Transform JSON to XML
- Map fields between schemas
- Enrich data

### 2. Receive Jobs

Poll protocol adapters for new messages.

```json
{
  "type": "receive",
  "protocol": "Kafka",
  "timeout": 120,
  "auto_route": true
}
```

**Use Cases:**
- Poll Kafka topics
- Check TCP connections
- Pull from message queues
- Monitor endpoints

### 3. Send Jobs

Send messages to destinations with optional transformation.

```json
{
  "type": "send",
  "protocol": "REST",
  "destination": "http://api.example.com/data",
  "mapping_id": "transform-before-send"
}
```

**Use Cases:**
- Push data to REST APIs
- Publish to Kafka topics
- Send SOAP requests
- Transmit via TCP

### 4. Route Jobs

Route messages through DataHub routing rules.

```json
{
  "type": "route",
  "routing_rule": "conditional-routing"
}
```

**Use Cases:**
- Conditional routing
- Fan-out to multiple destinations
- Load balancing
- Content-based routing

## Trigger Types

### 1. on_receive

Triggered when a message is received.

```json
{
  "trigger": {
    "type": "on_receive",
    "protocol": "REST",
    "topic": "/api/orders",
    "condition": {
      "field": "order.status",
      "operator": "eq",
      "value": "confirmed"
    }
  }
}
```

**Operators:**
- `eq`: Equal to
- `ne`: Not equal to
- `contains`: String contains
- `exists`: Field exists

### 2. on_schedule

Triggered based on cron schedule.

```json
{
  "trigger": {
    "type": "on_schedule",
    "schedule": "*/5 * * * *",
    "config": {
      "timezone": "UTC"
    }
  }
}
```

**Cron Format:**
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
* * * * *
```

**Examples:**
- `*/5 * * * *` - Every 5 minutes
- `0 */6 * * *` - Every 6 hours
- `0 0 * * *` - Daily at midnight
- `0 9 * * 1` - Every Monday at 9 AM

### 3. on_event

Triggered by system events (future enhancement).

```json
{
  "trigger": {
    "type": "on_event",
    "event": "user.created"
  }
}
```

### 4. manual

Manually triggered via API.

```json
{
  "trigger": {
    "type": "manual"
  }
}
```

## Configuration

### Complete Job Definition

```json
{
  "id": "unique-job-id",
  "name": "Human Readable Name",
  "description": "What this job does",
  "enabled": true,
  "type": "transform|receive|send|route",

  "trigger": {
    "type": "on_receive|on_schedule|on_event|manual",
    "protocol": "REST|SOAP|TCP|GraphQL|MQTT|Kafka|ActiveMQ",
    "topic": "topic/endpoint/queue name",
    "schedule": "cron expression",
    "condition": {
      "field": "path.to.field",
      "operator": "eq|ne|contains|exists",
      "value": "comparison value"
    }
  },

  "protocol": "Protocol name",
  "source": "Source endpoint/topic",
  "destination": "Destination endpoint/topic",
  "mapping_id": "DataHub mapping ID",
  "routing_rule": "DataHub routing rule ID",

  "priority": 10,
  "max_retries": 3,
  "timeout": 60,
  "auto_route": true,

  "metadata": {
    "custom": "metadata"
  },
  "options": {
    "custom": "options"
  }
}
```

### Priority Levels

- **1-3**: Low priority (batch processing, reports)
- **4-6**: Normal priority (regular transactions)
- **7-9**: High priority (important business logic)
- **10+**: Critical priority (urgent, time-sensitive)

## Examples

### Example 1: REST to SOAP with Transformation

```json
{
  "id": "rest-order-to-soap-erp",
  "name": "REST Order to SOAP ERP",
  "enabled": true,
  "type": "transform",
  "trigger": {
    "type": "on_receive",
    "protocol": "REST",
    "topic": "/api/orders",
    "condition": {
      "field": "order.status",
      "operator": "eq",
      "value": "confirmed"
    }
  },
  "protocol": "REST",
  "source": "/api/orders",
  "destination": "SOAP:http://erp.company.com/OrderService",
  "mapping_id": "rest-to-soap-order",
  "priority": 10,
  "max_retries": 3,
  "timeout": 60,
  "auto_route": true
}
```

**What This Does:**
1. Watches for REST POST requests to `/api/orders`
2. Checks if `order.status == "confirmed"`
3. Creates a job automatically
4. Applies `rest-to-soap-order` transformation mapping
5. Sends transformed message to SOAP ERP system
6. Retries up to 3 times on failure

**No code required!**

### Example 2: Scheduled Data Sync

```json
{
  "id": "hourly-inventory-sync",
  "name": "Hourly Inventory Synchronization",
  "enabled": true,
  "type": "send",
  "trigger": {
    "type": "on_schedule",
    "schedule": "0 * * * *"
  },
  "protocol": "REST",
  "destination": "http://warehouse.company.com/api/inventory",
  "mapping_id": "inventory-sync-mapping",
  "priority": 5,
  "max_retries": 2,
  "timeout": 300,
  "metadata": {
    "sync_type": "incremental"
  }
}
```

**What This Does:**
1. Runs every hour (top of the hour)
2. Fetches data (implementation-specific)
3. Applies inventory mapping
4. Sends to warehouse API
5. Handles retries automatically

### Example 3: MQTT Sensor to REST Analytics

```json
{
  "id": "mqtt-sensor-to-analytics",
  "name": "MQTT Sensor Data to Analytics",
  "enabled": true,
  "type": "transform",
  "trigger": {
    "type": "on_receive",
    "protocol": "MQTT",
    "topic": "sensors/+/temperature"
  },
  "protocol": "MQTT",
  "source": "sensors/#",
  "destination": "REST:http://analytics.company.com/api/data",
  "mapping_id": "mqtt-to-rest-sensor",
  "priority": 8,
  "max_retries": 3,
  "timeout": 30,
  "auto_route": true
}
```

**What This Does:**
1. Subscribes to MQTT topic `sensors/+/temperature`
2. Creates job for each message received
3. Transforms binary/JSON data
4. Sends to REST analytics API
5. All automatic - no code!

### Example 4: Conditional Routing

```json
{
  "id": "priority-based-routing",
  "name": "Route by Priority",
  "enabled": true,
  "type": "route",
  "trigger": {
    "type": "on_receive",
    "protocol": "REST",
    "topic": "/api/events",
    "condition": {
      "field": "event.priority",
      "operator": "eq",
      "value": "high"
    }
  },
  "protocol": "REST",
  "source": "/api/events",
  "routing_rule": "priority-routing",
  "priority": 10,
  "max_retries": 5,
  "auto_route": true
}
```

**What This Does:**
1. Monitors `/api/events` endpoint
2. Only processes events with `priority == "high"`
3. Routes through DataHub rules
4. Different destinations based on content
5. High priority (10) for fast processing

## Integration with Protocols

### REST Integration

```go
// REST server automatically calls job creator
server := rest.NewRESTServer(config, logger)
server.POST("/api/orders", func(ctx *rest.Context) error {
    // Job is created automatically based on configuration
    // No need to manually create job!
    return ctx.JSON(200, map[string]string{"status": "accepted"})
})
```

### MQTT Integration

```go
// MQTT client calls job creator on message receive
mqttClient.OnMessage(func(topic string, payload []byte) {
    // Job created automatically
    // Transformation and routing happen via job system
})
```

### Kafka Integration

```go
// Kafka consumer creates jobs automatically
kafkaConsumer.Subscribe(topics, func(message *kafka.Message) {
    // Job created and queued
    // Workers process asynchronously
})
```

### TCP Integration

```go
// TCP server creates jobs for incoming data
tcpServer.SetHandler(func(conn net.Conn, data []byte, logger *logrus.Logger) ([]byte, error) {
    // Job created automatically
    // DataHub handles transformation and routing
    return []byte("ACK"), nil
})
```

## Best Practices

### 1. Use Descriptive IDs and Names

```json
{
  "id": "rest-order-to-soap-erp",
  "name": "REST Order to SOAP ERP Transform"
}
```

### 2. Set Appropriate Priorities

- Critical business transactions: 10+
- Standard operations: 5-7
- Background tasks: 1-3

### 3. Configure Retries Wisely

- Idempotent operations: 3-5 retries
- Non-idempotent: 1-2 retries
- External APIs: Consider their rate limits

### 4. Use Conditions to Filter

```json
{
  "condition": {
    "field": "order.amount",
    "operator": "gt",
    "value": 10000
  }
}
```

Only process high-value orders, reducing unnecessary job creation.

### 5. Set Realistic Timeouts

- Simple transforms: 30 seconds
- External API calls: 60 seconds
- Batch operations: 300 seconds (5 minutes)

### 6. Add Metadata for Tracking

```json
{
  "metadata": {
    "department": "sales",
    "business_unit": "retail",
    "sla": "4_hours"
  }
}
```

### 7. Test with `enabled: false`

Set `enabled: false` while testing, then enable when ready.

### 8. Monitor Job Status

Check job history and logs:
- `queue_jobs` table - Job records
- `job_history` table - Execution history
- `Job_History` collection - DocumentDB audit trail

### 9. Use Auto-Route When Possible

```json
{
  "auto_route": true
}
```

Let DataHub handle routing automatically based on rules.

### 10. Organize by Environment

```
config/
  integration/
    datahub/
      dev/jobs.json
      staging/jobs.json
      production/jobs.json
```

## Monitoring and Debugging

### Check Job Status

```sql
SELECT id, handler, status_id, retry_count, created_on, updated_on
FROM queue_jobs
WHERE handler LIKE 'DataHub_%'
ORDER BY created_on DESC
LIMIT 100;
```

### View Job History

```sql
SELECT jh.job_id, jh.status_id, jh.start_time, jh.end_time,
       jh.duration, jh.error_message, jh.result
FROM job_history jh
WHERE jh.job_id = 'your-job-id'
ORDER BY jh.start_time DESC;
```

### Get Job Statistics

```go
stats := jobHandler.GetStats()
// Returns: enabled, hub_enabled, adapters, mappings, routing_rules

configStats := jobConfigMgr.GetStats()
// Returns: enabled_jobs, by_type, by_trigger, by_protocol
```

## Troubleshooting

### Job Not Created

**Check:**
1. Is job `enabled: true`?
2. Does trigger match (protocol, topic)?
3. Does condition evaluate to true?
4. Is IntegrationJobCreator enabled?

### Job Fails Immediately

**Check:**
1. Is mapping ID valid?
2. Does destination adapter exist?
3. Are credentials/permissions correct?
4. Check timeout settings

### Job Retrying Forever

**Check:**
1. Max retries setting
2. Error in transformation logic
3. Destination availability
4. Network connectivity

### Transformation Not Applied

**Check:**
1. Mapping ID specified?
2. Mapping loaded into DataHub?
3. Source/target paths correct?
4. Data types compatible?

## Performance Tuning

### Worker Count

Increase workers for higher throughput:

```json
{
  "workers": 10
}
```

### Poll Interval

Adjust how often workers check for jobs:

```json
{
  "poll_interval": 5
}
```

### Batch Size

Process multiple messages in batch (future enhancement).

### Caching

Enable Redis for distributed queue management.

## Summary

The DataHub Job Integration System provides:

✅ **Zero-Code Configuration** - Define everything in JSON
✅ **Automatic Job Creation** - Jobs created when messages arrive
✅ **Built-in Transformation** - Apply DataHub mappings automatically
✅ **Automatic Routing** - Route based on configurable rules
✅ **Retry & Error Handling** - Built-in resilience
✅ **Distributed Processing** - Scalable worker pools
✅ **Multiple Protocols** - REST, SOAP, TCP, GraphQL, MQTT, Kafka, ActiveMQ
✅ **Flexible Triggers** - On receive, scheduled, events, manual
✅ **Comprehensive Monitoring** - Full audit trail

**No code required - just configuration!**

## Additional Resources

- [DataHub Documentation](./README.md)
- [Array Mapping Guide](./ARRAY_MAPPING_GUIDE.md)
- [Job Configuration Examples](./job_config_example.json)
- [Complex Array Examples](./complex_array_mapping_example.json)
