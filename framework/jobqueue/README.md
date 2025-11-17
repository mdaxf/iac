# Background Job System

A comprehensive background job processing system with queue management, scheduled jobs, and distributed processing capabilities.

## Features

- **Queue Job Processing**: Process jobs asynchronously with priority support
- **Scheduled Jobs**: Configure recurring jobs with cron expressions or intervals
- **Distributed Processing**: Redis-based distributed locking for multi-instance deployments
- **Job History**: Complete audit trail of all job executions
- **Retry Mechanism**: Automatic retry with configurable attempts
- **Integration Hooks**: Create jobs from various integration sources (SignalR, Kafka, MQTT, HTTP, etc.)
- **Transaction Support**: Jobs execute within database transactions

## Architecture

### Components

1. **QueueJob** - Individual jobs to be executed
2. **Job** - Scheduled/interval job configurations
3. **JobHistory** - Execution history and audit trail
4. **JobWorker** - Background workers that process jobs
5. **JobScheduler** - Manages scheduled and interval jobs
6. **DistributedQueueManager** - Redis-based distributed job coordination
7. **IntegrationJobCreator** - Creates jobs from integration messages

### Database Tables

#### queue_jobs
Stores all jobs to be executed (from integrations or scheduled jobs).

Key fields:
- `id`: Job unique identifier
- `typeid`: Job type (Integration, Scheduled, Manual, System)
- `handler`: Transaction code or command to execute
- `statusid`: Current job status (Pending, Processing, Completed, Failed, etc.)
- `priority`: Job priority (higher = processed first)
- `maxretries`: Maximum retry attempts
- `scheduledat`: When to execute (null = immediate)
- `metadata`: Flexible JSON metadata

#### job_histories
Stores execution history for all job runs.

Key fields:
- `jobid`: Reference to queue_jobs
- `executionid`: Unique execution identifier
- `startedat`, `completedat`: Execution timestamps
- `duration`: Execution duration in milliseconds
- `result`: Execution result
- `errormessage`: Error details if failed

#### jobs
Stores scheduled/interval job configurations.

Key fields:
- `name`: Job name (unique)
- `handler`: Transaction code or command to execute
- `cronexpression`: Cron expression for scheduling
- `intervalseconds`: Interval in seconds (alternative to cron)
- `startat`, `endat`: Execution window
- `maxexecutions`: Maximum number of executions
- `condition`: SQL condition to evaluate before execution
- `enabled`: Whether job is active

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

### Configuration Options

- `enabled`: Enable/disable the job system
- `workers`: Number of worker goroutines
- `poll_interval`: How often to poll for new jobs (seconds)
- `max_retries`: Default maximum retry attempts
- `scheduler_check_interval`: How often to check for new scheduled jobs (seconds)
- `use_redis`: Use Redis for distributed job coordination
- `job_history_retention_days`: How long to keep job history
- `enable_metrics`: Enable job metrics collection

## Installation

### 1. Run Database Migrations

**MySQL:**
```bash
mysql -u user -p database < migrations/job_tables_mysql.sql
```

**PostgreSQL:**
```bash
psql -U user -d database -f migrations/job_tables_postgresql.sql
```

### 2. Install Dependencies

```bash
go get github.com/robfig/cron/v3
go get github.com/google/uuid
```

### 3. Initialize in Application

Add to `initialize.go`:

```go
import "iac/framework/jobqueue"

// In initialization function
err := jobqueue.InitializeJobSystem(
    context.Background(),
    dbconn.DB,
    docDBconn,
    config.ObjectCache,
    signalRClient,
)
if err != nil {
    logger.Error(fmt.Sprintf("Failed to initialize job system: %v", err))
}

// Register shutdown handler
defer jobqueue.ShutdownJobSystem()
```

## Usage

### Creating Jobs from Integration Messages

```go
// Create job from SignalR message
job, err := jobqueue.GlobalJobCreator.CreateJobFromSignalRMessage(
    ctx,
    "user.created",
    messagePayload,
    "user.onboarding",
)

// Create job from Kafka message
job, err := jobqueue.GlobalJobCreator.CreateJobFromKafkaMessage(
    ctx,
    "orders.placed",
    messagePayload,
    "order.processing",
    models.JobDirectionInbound,
)

// Create job from HTTP request
job, err := jobqueue.GlobalJobCreator.CreateJobFromHTTPRequest(
    ctx,
    "/api/process",
    requestData,
    "api.handler",
    "POST",
    models.JobDirectionInbound,
)
```

### Creating Scheduled Jobs

Insert into `jobs` table:

```sql
INSERT INTO jobs (
    id, name, handler, cronexpression, enabled, priority
) VALUES (
    'daily-report',
    'Daily Report Generation',
    'reports.daily',
    '0 8 * * *',  -- Every day at 8 AM
    TRUE,
    5
);

-- Or with interval
INSERT INTO jobs (
    id, name, handler, intervalseconds, enabled, priority
) VALUES (
    'health-check',
    'Health Check',
    'system.health.check',
    300,  -- Every 5 minutes
    TRUE,
    3
);
```

### Manual Job Creation

```go
import "iac/services"

jobService := services.NewJobService(db)

job := &models.QueueJob{
    TypeID:     int(models.JobTypeManual),
    Handler:    "custom.handler",
    Payload:    `{"key": "value"}`,
    Priority:   10,
    MaxRetries: 5,
    CreatedBy:  "user123",
}

err := jobService.CreateQueueJob(ctx, job)

// Enqueue for distributed processing
jobqueue.GlobalQueueManager.EnqueueJob(ctx, job.ID, job.Priority)
```

## Job Handlers

Job handlers are transaction codes that will be executed. They should:

1. Accept a data map as input
2. Execute within a database transaction
3. Return outputs and error

Example transaction code structure:
```go
func ProcessUserOnboarding(data map[string]interface{}, tx *sql.Tx, docDB *documents.DocDB) (map[string]interface{}, error) {
    // Extract job data
    userID := data["user_id"].(string)

    // Process the job
    // ...

    // Return results
    return map[string]interface{}{
        "status": "completed",
        "user_id": userID,
    }, nil
}
```

## Job Statuses

- `JobStatusPending` (0): Job created, waiting to be processed
- `JobStatusQueued` (1): Job queued in Redis
- `JobStatusProcessing` (2): Job currently being processed
- `JobStatusCompleted` (3): Job completed successfully
- `JobStatusFailed` (4): Job failed permanently
- `JobStatusRetrying` (5): Job failed but will retry
- `JobStatusCancelled` (6): Job cancelled
- `JobStatusScheduled` (7): Job scheduled for future execution

## Monitoring

### Check Job System Status

```go
status := jobqueue.GetJobSystemStatus()
fmt.Printf("Job System Status: %+v\n", status)
```

### Query Job Statistics

```go
jobService := services.NewJobService(db)
stats, err := jobService.GetJobStatistics(ctx)
```

### View Job History

```sql
SELECT * FROM job_histories
WHERE jobid = 'job-id'
ORDER BY startedat DESC;
```

## Cron Expression Examples

- `0 */5 * * * *` - Every 5 minutes
- `0 0 * * * *` - Every hour
- `0 0 8 * * *` - Every day at 8 AM
- `0 0 0 * * 0` - Every Sunday at midnight
- `0 0 0 1 * *` - First day of every month

## Best Practices

1. **Idempotency**: Design job handlers to be idempotent
2. **Timeouts**: Set appropriate timeouts for long-running jobs
3. **Error Handling**: Properly handle and log errors
4. **Retry Logic**: Use retries for transient failures only
5. **Monitoring**: Regularly monitor job queue depth and failure rates
6. **Cleanup**: Periodically clean up old job history
7. **Priority**: Use priority wisely to avoid starvation

## Troubleshooting

### Jobs Not Processing

1. Check if job system is enabled in configuration
2. Verify database connectivity
3. Check worker status: `jobqueue.GlobalJobWorker.IsRunning()`
4. Review logs for errors

### Jobs Failing

1. Check job_histories table for error messages
2. Verify handler exists and is correct
3. Check transaction code for errors
4. Review database transaction logs

### Redis Connection Issues

1. Verify Redis configuration in cache settings
2. Check Redis connectivity
3. System will fall back to single-instance mode if Redis unavailable

## License

Copyright 2023 IAC. All Rights Reserved.
