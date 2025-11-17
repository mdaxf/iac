# Background Job System Implementation

## Overview

A comprehensive background job processing system has been implemented to handle messages from inbound and outbound integrations, manage job queues, and execute scheduled/interval jobs automatically.

## Key Features

1. **Job Queue Management**: All integration messages create jobs in `queue_jobs` table for asynchronous processing
2. **Distributed Processing**: Cache-based distributed locking enables multi-instance deployments (Redis, Memcache, etc.)
3. **Flexible Cache Support**: Works with any configured cache adapter or runs in single-instance mode
4. **Scheduled Jobs**: Configure recurring jobs with cron expressions or time intervals
5. **Job History**: Complete audit trail of all executions in `job_histories` table
6. **Retry Mechanism**: Automatic retry with configurable attempts
7. **Priority Support**: Jobs processed based on priority (higher first) and FIFO within priority
8. **Transaction Safety**: All jobs execute within database transactions

## Database Schema

### Tables Created

#### 1. queue_jobs
Stores all jobs to be executed (from integrations or scheduled jobs).

**Key Fields:**
- `id`: Unique job identifier
- `typeid`: Job type (0=Integration, 1=Scheduled, 2=Manual, 3=System)
- `method`: HTTP method or operation type
- `protocol`: Communication protocol (http, signalr, kafka, mqtt, activemq)
- `direction`: Message direction (inbound, outbound, internal)
- `handler`: Transaction code or command to execute
- `metadata`: Flexible JSON metadata
- `payload`: Job data/payload
- `result`: Execution result
- `statusid`: Job status (0=Pending, 1=Queued, 2=Processing, 3=Completed, 4=Failed, 5=Retrying, 6=Cancelled, 7=Scheduled)
- `priority`: Job priority (higher = processed first)
- `maxretries`: Maximum retry attempts
- `retrycount`: Current retry count
- `scheduledat`: When to execute (null = immediate)
- `startedat`: Execution start time
- `completedat`: Execution completion time
- `lasterror`: Last error message
- `parentjobid`: Parent job for chained jobs

#### 2. job_histories
Stores complete execution history for all job runs.

**Key Fields:**
- `id`: Unique history record identifier
- `jobid`: Reference to queue_jobs
- `executionid`: Unique execution identifier
- `statusid`: Execution status
- `startedat`: Execution start time
- `completedat`: Execution completion time
- `duration`: Execution duration in milliseconds
- `result`: Execution result
- `errormessage`: Error details if failed
- `retryattempt`: Which retry attempt
- `executedby`: Worker/instance that executed
- `inputdata`: Input data snapshot
- `outputdata`: Output data
- `metadata`: Execution metadata

#### 3. jobs
Stores scheduled/interval job configurations.

**Key Fields:**
- `id`: Unique job identifier
- `name`: Job name (unique)
- `description`: Job description
- `handler`: Transaction code or command to execute
- `cronexpression`: Cron expression (e.g., "0 8 * * *" = daily at 8 AM)
- `intervalseconds`: Interval in seconds (alternative to cron)
- `startat`: When to start execution
- `endat`: When to stop execution
- `maxexecutions`: Maximum number of executions (0 = unlimited)
- `executioncount`: Current execution count
- `enabled`: Whether job is active
- `condition`: SQL condition to evaluate before execution
- `priority`: Job priority
- `maxretries`: Maximum retry attempts for generated jobs
- `timeout`: Timeout in seconds
- `metadata`: Additional configuration
- `lastrunat`: Last execution time
- `nextrunat`: Next scheduled execution time

## Architecture Components

### 1. Models (`models/job.go`)
- `QueueJob`: Job queue record structure
- `JobHistory`: Execution history record structure
- `Job`: Scheduled job configuration structure
- `JobMetadata`: Flexible metadata map
- Job status and type enumerations

### 2. Job Service (`services/jobservice.go`)
Provides CRUD operations for jobs:
- `CreateQueueJob`: Create new job in queue
- `UpdateQueueJobStatus`: Update job status
- `GetNextPendingJob`: Retrieve next job to process
- `CreateJobHistory`: Record execution history
- `GetScheduledJobs`: Retrieve scheduled jobs to run
- `CreateJobFromIntegrationMessage`: Create job from integration message

### 3. Distributed Queue Manager (`framework/jobqueue/queue_manager.go`)
Cache-based distributed coordination (works with Redis, Memcache, or any cache adapter):
- `EnqueueJob`: Add job to distributed queue
- `DequeueJob`: Remove and return next job
- `AcquireLock`: Distributed lock for job processing
- `ReleaseLock`: Release distributed lock
- `SetJobStatus`: Cache job status
- Prevents duplicate processing across multiple instances
- Automatically disabled if no cache is configured (single-instance mode)

### 4. Job Worker (`framework/jobqueue/worker.go`)
Background worker pool that processes jobs:
- Configurable number of worker goroutines
- Polls database for pending jobs
- Acquires distributed locks before processing
- Executes job handlers within transactions
- Handles retries for failed jobs
- Records execution history
- Graceful shutdown support

### 5. Job Scheduler (`framework/jobqueue/scheduler.go`)
Manages scheduled and interval jobs:
- Cron expression support (using robfig/cron/v3)
- Interval-based scheduling
- Condition evaluation before execution
- Generates queue jobs from scheduled jobs
- Automatic reload of configuration changes
- Start/end time windows
- Maximum execution limits

### 6. Integration Job Creator (`framework/jobqueue/integration_hooks.go`)
Creates jobs from integration messages:
- `CreateJobFromSignalRMessage`: SignalR message handler
- `CreateJobFromKafkaMessage`: Kafka message handler
- `CreateJobFromMQTTMessage`: MQTT message handler
- `CreateJobFromActiveMQMessage`: ActiveMQ message handler
- `CreateJobFromHTTPRequest`: HTTP request handler
- `BatchCreateJobs`: Batch job creation
- Automatic direction and protocol detection

### 7. Job System Initializer (`framework/jobqueue/initializer.go`)
Initialization and lifecycle management:
- `InitializeJobSystem`: Start all components
- `ShutdownJobSystem`: Graceful shutdown
- Configuration validation
- Health checks
- Status reporting

## Configuration

Add to `configuration.json`:

```json
{
  "jobs": {
    "enabled": true,
    "workers": 5,
    "poll_interval": 5,
    "max_retries": 3,
    "scheduler_check_interval": 60,
    "use_redis": true,
    "job_history_retention_days": 90,
    "enable_metrics": true
  }
}
```

**Configuration Parameters:**
- `enabled`: Enable/disable the job system (default: false)
- `workers`: Number of worker goroutines (default: 5)
- `poll_interval`: Poll interval in seconds (default: 5)
- `max_retries`: Default maximum retry attempts (default: 3)
- `scheduler_check_interval`: Scheduler check interval in seconds (default: 60)
- `use_redis`: Informational flag indicating cache type - system uses any configured cache (default: false)
- `job_history_retention_days`: Days to retain job history (default: 90)
- `enable_metrics`: Enable metrics collection (default: false)

**Cache Behavior**: The system automatically detects and uses whatever cache is configured in the cache configuration (Redis, Memcache, etc.). If no cache is available, it runs in single-instance mode without distributed locking. This allows the system to work in any deployment scenario.

## Installation Steps

### 1. Run Database Migrations

**For MySQL:**
```bash
mysql -u user -p database < migrations/job_tables_mysql.sql
```

**For PostgreSQL:**
```bash
psql -U user -d database -f migrations/job_tables_postgresql.sql
```

The migrations create:
- All three tables (queue_jobs, job_histories, jobs)
- Appropriate indexes for performance
- Foreign key constraints
- Sample scheduled jobs

### 2. Update Configuration

Add the jobs configuration to your `configuration.json` file (see Configuration section above).

### 3. Initialize in Application

Add to your initialization code (e.g., `initialize.go`):

```go
import "github.com/mdaxf/iac/framework/jobqueue"

// Initialize job system
err := jobqueue.InitializeJobSystem(
    context.Background(),
    dbconn.DB,
    docDBconn,
    config.ObjectCache,  // Redis cache
    signalRClient,
)
if err != nil {
    logger.Error(fmt.Sprintf("Failed to initialize job system: %v", err))
}

// Register cleanup on shutdown
defer jobqueue.ShutdownJobSystem()
```

### 4. Install Dependencies

```bash
go get github.com/robfig/cron/v3
go get github.com/google/uuid
```

## Usage Examples

### Creating Jobs from Integration Messages

```go
import "github.com/mdaxf/iac/framework/jobqueue"

// Example 1: Create job from SignalR message
job, err := jobqueue.GlobalJobCreator.CreateJobFromSignalRMessage(
    ctx,
    "user.created",
    messagePayload,
    "user.onboarding.process",
)

// Example 2: Create job from Kafka inbound message
job, err := jobqueue.GlobalJobCreator.CreateJobFromKafkaMessage(
    ctx,
    "orders.placed",
    orderData,
    "order.processing.handler",
    models.JobDirectionInbound,
)

// Example 3: Create job for outbound HTTP call
job, err := jobqueue.GlobalJobCreator.CreateJobFromHTTPRequest(
    ctx,
    "/api/external/notify",
    notificationData,
    "notification.send.handler",
    "POST",
    models.JobDirectionOutbound,
)
```

### Creating Scheduled Jobs

```sql
-- Daily report at 8 AM
INSERT INTO jobs (
    id, name, description, handler, cronexpression,
    enabled, priority, maxretries, createdby
) VALUES (
    'daily-sales-report',
    'Daily Sales Report',
    'Generate and send daily sales report',
    'reports.daily.sales',
    '0 8 * * *',
    TRUE,
    5,
    3,
    'system'
);

-- Health check every 5 minutes
INSERT INTO jobs (
    id, name, description, handler, intervalseconds,
    enabled, priority, createdby
) VALUES (
    'health-check',
    'System Health Check',
    'Check system health and send alerts',
    'system.health.check',
    300,
    TRUE,
    10,
    'system'
);

-- Monthly cleanup on first day at midnight
INSERT INTO jobs (
    id, name, description, handler, cronexpression,
    enabled, priority, createdby
) VALUES (
    'monthly-cleanup',
    'Monthly Data Cleanup',
    'Archive old data and clean up temporary files',
    'system.cleanup.monthly',
    '0 0 1 * *',
    TRUE,
    3,
    'system'
);
```

### Manual Job Creation

```go
import (
    "github.com/mdaxf/iac/models"
    "github.com/mdaxf/iac/services"
)

jobService := services.NewJobService(db)

// Create a high-priority manual job
job := &models.QueueJob{
    TypeID:     int(models.JobTypeManual),
    Handler:    "data.export.handler",
    Payload:    `{"format": "csv", "filters": {"date": "2024-01-01"}}`,
    Priority:   15,  // High priority
    MaxRetries: 5,
    Metadata: models.JobMetadata{
        "requested_by": "user@example.com",
        "export_type": "full",
    },
    CreatedBy: "admin",
}

err := jobService.CreateQueueJob(ctx, job)
if err != nil {
    // Handle error
}

// Enqueue for distributed processing
if jobqueue.GlobalQueueManager != nil {
    jobqueue.GlobalQueueManager.EnqueueJob(ctx, job.ID, job.Priority)
}
```

## Monitoring and Management

### Check System Status

```go
// Get overall system status
status := jobqueue.GetJobSystemStatus()
fmt.Printf("System Status: %+v\n", status)

// Check if running
isRunning := jobqueue.IsJobSystemRunning()
fmt.Printf("Job System Running: %v\n", isRunning)
```

### Query Job Statistics

```go
jobService := services.NewJobService(db)
stats, err := jobService.GetJobStatistics(ctx)
if err == nil {
    fmt.Printf("Job Statistics: %+v\n", stats)
}
```

### View Job History

```sql
-- Recent job executions
SELECT
    jh.id,
    jh.jobid,
    qj.handler,
    jh.statusid,
    jh.startedat,
    jh.completedat,
    jh.duration,
    jh.errormessage
FROM job_histories jh
JOIN queue_jobs qj ON jh.jobid = qj.id
ORDER BY jh.startedat DESC
LIMIT 100;

-- Failed jobs needing attention
SELECT
    qj.id,
    qj.handler,
    qj.statusid,
    qj.retrycount,
    qj.maxretries,
    qj.lasterror,
    qj.createdon
FROM queue_jobs qj
WHERE qj.statusid = 4  -- Failed
  AND qj.active = TRUE
ORDER BY qj.priority DESC, qj.createdon ASC;

-- Job execution statistics
SELECT
    qj.handler,
    COUNT(*) as total_executions,
    SUM(CASE WHEN jh.statusid = 3 THEN 1 ELSE 0 END) as successful,
    SUM(CASE WHEN jh.statusid = 4 THEN 1 ELSE 0 END) as failed,
    AVG(jh.duration) as avg_duration_ms
FROM job_histories jh
JOIN queue_jobs qj ON jh.jobid = qj.id
WHERE jh.startedat > DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY qj.handler
ORDER BY total_executions DESC;
```

## Integration Points

The job system integrates with all existing integration methods:

1. **SignalR/WebSocket**: Real-time message handling
2. **Kafka**: Message queue integration
3. **MQTT**: IoT and messaging
4. **ActiveMQ**: Enterprise messaging
5. **HTTP/REST**: API calls and webhooks
6. **Database**: Direct job creation

## Job Flow

### Inbound Message Flow
1. Message received via integration (SignalR, Kafka, etc.)
2. Integration hook creates job in `queue_jobs`
3. Job enqueued in Redis (if enabled)
4. Worker picks up job and acquires lock
5. Handler executed within transaction
6. Result saved to `queue_jobs`
7. Execution history saved to `job_histories`
8. Lock released

### Scheduled Job Flow
1. Scheduler checks `jobs` table periodically
2. Evaluates cron expression or interval
3. Checks conditions (if specified)
4. Creates job in `queue_jobs`
5. Updates next run time
6. Job processed by worker (same as inbound)

### Retry Flow
1. Job fails during execution
2. Error saved to `lasterror`
3. Retry count incremented
4. If retries remaining:
   - Status set to Retrying
   - Job re-enqueued with lower priority
5. If max retries exceeded:
   - Status set to Failed
   - No further processing

## Performance Considerations

1. **Database Indexes**: Optimized for common queries
2. **Connection Pooling**: Reuses database connections
3. **Worker Pool**: Parallel processing with configurable workers
4. **Priority Queue**: High-priority jobs processed first
5. **Distributed Locking**: Prevents duplicate processing
6. **Batch Operations**: Supports batch job creation
7. **Async Processing**: Non-blocking integration handlers

## Error Handling

- All errors logged with context
- Failed jobs retain error messages
- Automatic retry for transient failures
- Transaction rollback on errors
- Graceful degradation without Redis

## Security Considerations

- All jobs execute within database transactions
- Input validation in handlers
- User context preserved in audit fields
- Configurable job isolation
- Role-based job creation (via CreatedBy field)

## Future Enhancements

Potential future improvements:
- Job dependencies and chaining
- Job progress tracking
- Web UI for job management
- Metrics and alerting
- Dead letter queue
- Job cancellation API
- Dynamic worker scaling
- Job templates

## Files Created/Modified

### New Files
- `models/job.go`: Job data models
- `services/jobservice.go`: Job service layer
- `framework/jobqueue/queue_manager.go`: Distributed queue management
- `framework/jobqueue/worker.go`: Background job worker
- `framework/jobqueue/scheduler.go`: Job scheduler
- `framework/jobqueue/integration_hooks.go`: Integration message handlers
- `framework/jobqueue/initializer.go`: System initialization
- `framework/jobqueue/README.md`: Detailed documentation
- `migrations/job_tables_mysql.sql`: MySQL schema
- `migrations/job_tables_postgresql.sql`: PostgreSQL schema
- `migrations/job_system_config_example.json`: Configuration example

### Modified Files
- `config/global_variable.go`: Added JobsConfiguration structure
- `go.mod`: Added cron dependency
- `go.sum`: Dependency checksums

## License

Copyright 2023 IAC. All Rights Reserved.

Licensed under the Apache License, Version 2.0.
