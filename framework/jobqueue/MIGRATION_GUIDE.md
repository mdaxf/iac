# Migration Guide: framework/queue to framework/jobqueue

## Overview

This guide helps migrate from the legacy `framework/queue` system to the new `framework/jobqueue` system.

## Feature Comparison

| Feature | Old (framework/queue) | New (framework/jobqueue) | Status |
|---------|----------------------|--------------------------|--------|
| **Queue Storage** | In-memory slice | Database + Cache | ✅ Enhanced |
| **Worker Pool** | Dynamic (10 workers per batch) | Fixed configurable workers | ✅ Improved |
| **Retry Logic** | Execute counter vs Retry limit | RetryCount vs MaxRetries | ✅ Compatible |
| **Transaction Execution** | trancode.ExecutebyExternal | trancode.ExecutebyExternal | ✅ Same |
| **Job History** | DocumentDB "Job_History" | Database + DocumentDB | ✅ Enhanced |
| **Distributed Processing** | Single instance only | Multi-instance with cache locking | ✅ New |
| **Scheduled Jobs** | Not supported | Cron & interval support | ✅ New |
| **Priority Queue** | FIFO only | Priority + FIFO | ✅ New |
| **Performance Tracking** | Logged durations | Saved to history (duration_ms) | ✅ Enhanced |
| **Graceful Shutdown** | Signal handling | Context-based cancellation | ✅ Improved |
| **Panic Recovery** | Per-method recovery | Per-job recovery | ✅ Same |

## Key Differences

### 1. Message Structure

**Old (framework/queue):**
```go
type Message struct {
    Id          string
    UUID        string
    Retry       int        // Max retry attempts
    Execute     int        // Current execution count
    Topic       string
    PayLoad     []byte
    Handler     string
    CreatedOn   time.Time
    ExecutedOn  time.Time
    CompletedOn time.Time
}
```

**New (framework/jobqueue):**
```go
type QueueJob struct {
    ID              string
    Handler         string
    Payload         string  // JSON string
    StatusID        int
    Priority        int     // NEW: Priority support
    MaxRetries      int     // Was "Retry"
    RetryCount      int     // Was "Execute"
    ScheduledAt     *time.Time  // NEW: Scheduled execution
    StartedAt       *time.Time  // Was "ExecutedOn"
    CompletedAt     *time.Time  // Was "CompletedOn"
    Metadata        JobMetadata // NEW: Flexible metadata
    // ... audit fields
}
```

### 2. Payload Format

**Old format (preserved in metadata):**
```json
{
    "Topic": "user.created",
    "Payload": "{\"user_id\": 123}"
}
```

**New format (in Payload field):**
```json
{
    "Topic": "user.created",
    "Payload": "{\"user_id\": 123}",
    "ID": "msg-123",
    "UUID": "uuid-456",
    "CreatedOn": "2024-01-01T00:00:00Z"
}
```

The new system automatically includes these fields in the handler execution data.

### 3. Job History Storage

**Old:**
- Stored only in DocumentDB collection "Job_History"
- Structure: `{message, executedon, executedby, status, errormessage, messagequeue, outputs}`

**New:**
- Primary storage: `job_histories` table (searchable, indexed)
- Secondary storage: DocumentDB "Job_History" collection (for compatibility)
- Enhanced fields: duration, retry_attempt, execution_id, etc.

### 4. Worker Model

**Old:**
- Dynamic workers created per batch
- 10 messages per worker batch
- Workers created on-demand when queue has messages
- 500ms polling interval

**New:**
- Fixed worker pool (configurable, default: 5)
- Continuous polling from database
- Configurable poll interval (default: 5 seconds)
- Distributed locking for multi-instance support

## Migration Steps

### Step 1: Update Configuration

Add to `configuration.json`:

```json
{
  "jobs": {
    "enabled": true,
    "workers": 10,
    "poll_interval": 1,
    "max_retries": 3,
    "scheduler_check_interval": 60,
    "use_redis": true,
    "job_history_retention_days": 90,
    "enable_metrics": true
  }
}
```

**Note:** Set `workers` to 10 and `poll_interval` to 1 second to match old system performance.

### Step 2: Run Database Migrations

```bash
# For MySQL
mysql -u user -p database < migrations/job_tables_mysql.sql

# For PostgreSQL
psql -U user -d database -f migrations/job_tables_postgresql.sql
```

### Step 3: Initialize Job System

In your initialization code:

```go
import "github.com/mdaxf/iac/framework/jobqueue"

// Initialize the new job system
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

defer jobqueue.ShutdownJobSystem()
```

### Step 4: Migrate Message Queue Usage

**Old usage:**
```go
import "github.com/mdaxf/iac/framework/queue"

// Create queue
mq := queue.NewMessageQueue("queue-1", "MyQueue")

// Push message
mq.Push(queue.Message{
    Id:      "msg-123",
    UUID:    "uuid-456",
    Retry:   3,
    Topic:   "user.created",
    PayLoad: []byte(`{"user_id": 123}`),
    Handler: "user.onboarding",
    CreatedOn: time.Now(),
})
```

**New usage:**
```go
import "github.com/mdaxf/iac/framework/jobqueue"

// Create job from message (via integration hook)
job, err := jobqueue.GlobalJobCreator.CreateJobFromMessage(
    ctx,
    "user.created",                    // topic
    `{"user_id": 123}`,               // payload
    "user.onboarding",                // handler
    "POST",                           // method
    "signalr",                        // protocol
    models.JobDirectionInbound,       // direction
    5,                                // priority
)
```

### Step 5: Backward Compatibility

The new system maintains backward compatibility:

1. **Job_History DocumentDB Collection**: Still populated with same structure
2. **Payload Format**: Topic/Payload structure preserved in handler data
3. **Transaction Execution**: Same `trancode.ExecutebyExternal` method
4. **Retry Logic**: Same behavior, different field names

## Coexistence Strategy

Both systems can run simultaneously during migration:

1. **Keep old system running** for existing queues
2. **Route new messages** to new job system
3. **Gradually migrate** queue by queue
4. **Monitor both systems** during transition
5. **Remove old system** after full migration

## Performance Considerations

### Old System
- In-memory (fast but not persistent)
- No distributed support
- Dynamic workers (overhead)
- 500ms polling interval

### New System
- Database-backed (persistent, recoverable)
- Distributed with cache locking
- Fixed worker pool (efficient)
- Configurable polling interval

**Performance Tuning:**

For similar performance to old system:
```json
{
  "workers": 10,
  "poll_interval": 1
}
```

For better throughput:
```json
{
  "workers": 20,
  "poll_interval": 5,
  "use_redis": true
}
```

## API Reference

### Creating Jobs

**From SignalR Message:**
```go
job, err := jobqueue.GlobalJobCreator.CreateJobFromSignalRMessage(
    ctx, topic, payload, handler,
)
```

**From Kafka Message:**
```go
job, err := jobqueue.GlobalJobCreator.CreateJobFromKafkaMessage(
    ctx, topic, payload, handler, direction,
)
```

**From HTTP Request:**
```go
job, err := jobqueue.GlobalJobCreator.CreateJobFromHTTPRequest(
    ctx, endpoint, payload, handler, method, direction,
)
```

**Direct Job Creation (equivalent to old Push):**
```go
jobService := services.NewJobService(db)

job := &models.QueueJob{
    Handler:    "user.onboarding",
    Payload:    `{"Topic":"user.created","Payload":"..."}`,
    Priority:   5,
    MaxRetries: 3,
    Metadata: models.JobMetadata{
        "topic": "user.created",
        "queue_name": "MyQueue",
    },
    CreatedBy: "system",
}

err := jobService.CreateQueueJob(ctx, job)

// Enqueue for distributed processing
if jobqueue.GlobalQueueManager != nil {
    jobqueue.GlobalQueueManager.EnqueueJob(ctx, job.ID, job.Priority)
}
```

## Troubleshooting

### Issue: Jobs not processing

**Check:**
1. Job system initialized: `jobqueue.IsJobSystemRunning()`
2. Workers running: Check logs for "Started job worker with N workers"
3. Database connection: Verify `queue_jobs` table exists
4. Job status: Query `SELECT * FROM queue_jobs WHERE statusid IN (0,1)`

### Issue: Jobs processing slowly

**Solutions:**
1. Increase workers: `"workers": 20`
2. Decrease poll interval: `"poll_interval": 1`
3. Enable cache: Configure Redis in cache config
4. Check database indexes: Ensure migrations ran successfully

### Issue: Jobs failing silently

**Check:**
1. Job histories: `SELECT * FROM job_histories WHERE statusid = 4`
2. Error messages: `SELECT jobid, errormessage FROM job_histories WHERE errormessage != ''`
3. Handler exists: Verify transaction code handler is registered

## Best Practices

1. **Use Integration Hooks**: Let the system create jobs automatically from messages
2. **Set Appropriate Priorities**: Use priority to ensure critical jobs run first
3. **Monitor Job History**: Regularly review `job_histories` for failures
4. **Clean Old History**: Use retention settings to manage database size
5. **Use Metadata**: Store additional context in metadata for debugging
6. **Test Handlers**: Ensure transaction codes handle failures gracefully

## Rollback Plan

If needed to rollback to old system:

1. Stop new job system: `jobqueue.ShutdownJobSystem()`
2. Reactivate old MessageQueue instances
3. Drain remaining jobs from `queue_jobs` table
4. Keep database tables for historical records

## Support

- Documentation: See `framework/jobqueue/README.md`
- System Status: Use `jobqueue.GetJobSystemStatus()`
- Job Statistics: Use `jobService.GetJobStatistics(ctx)`

## Appendix: Field Mapping

| Old Field | New Field | Notes |
|-----------|-----------|-------|
| Message.Id | QueueJob.ID | Same |
| Message.UUID | Metadata["uuid"] | Stored in metadata |
| Message.Retry | QueueJob.MaxRetries | Max attempts |
| Message.Execute | QueueJob.RetryCount | Current attempt |
| Message.Topic | Metadata["topic"] | Stored in metadata |
| Message.PayLoad | QueueJob.Payload | Converted to JSON string |
| Message.Handler | QueueJob.Handler | Same |
| Message.CreatedOn | QueueJob.CreatedOn | Same |
| Message.ExecutedOn | QueueJob.StartedAt | Renamed |
| Message.CompletedOn | QueueJob.CompletedAt | Same |
| N/A | QueueJob.Priority | New field |
| N/A | QueueJob.StatusID | New field |
| N/A | QueueJob.ScheduledAt | New field for scheduled jobs |
