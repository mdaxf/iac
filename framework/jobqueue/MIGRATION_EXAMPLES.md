# Migration Examples

## Using the Migration Adapter

### Example 1: Replace Old MessageQueue with New System

**Old Code (framework/queue):**
```go
import "github.com/mdaxf/iac/framework/queue"

// Create old message queue
mq := queue.NewMessageQueue("queue-1", "MyQueue")

// Push messages
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

**New Code (framework/jobqueue with adapter):**
```go
import "github.com/mdaxf/iac/framework/jobqueue"

// Create adapter (drop-in replacement)
adapter := jobqueue.NewMessageQueueAdapter(
    "queue-1",
    "MyQueue",
    db,
    docDB,
    signalRClient,
    jobqueue.GlobalQueueManager,
)

// Push messages (same interface)
adapter.Push(jobqueue.LegacyMessage{
    Id:      "msg-123",
    UUID:    "uuid-456",
    Retry:   3,
    Topic:   "user.created",
    PayLoad: []byte(`{"user_id": 123}`),
    Handler: "user.onboarding",
    CreatedOn: time.Now(),
})
```

### Example 2: Batch Migration

```go
import "github.com/mdaxf/iac/framework/jobqueue"

adapter := jobqueue.NewMessageQueueAdapter(
    "queue-1",
    "MyQueue",
    db,
    docDB,
    signalRClient,
    jobqueue.GlobalQueueManager,
)

// Prepare legacy messages
legacyMessages := []jobqueue.LegacyMessage{
    {
        Id:      "msg-1",
        UUID:    "uuid-1",
        Retry:   3,
        Topic:   "order.placed",
        PayLoad: []byte(`{"order_id": 100}`),
        Handler: "order.process",
        CreatedOn: time.Now(),
    },
    {
        Id:      "msg-2",
        UUID:    "uuid-2",
        Retry:   3,
        Topic:   "order.shipped",
        PayLoad: []byte(`{"order_id": 101}`),
        Handler: "order.notify",
        CreatedOn: time.Now(),
    },
}

// Migrate batch
err := adapter.PushBatch(legacyMessages)
if err != nil {
    log.Printf("Batch migration had errors: %v", err)
}
```

### Example 3: Migrate Existing Job_History Collection

```go
import (
    "context"
    "github.com/mdaxf/iac/framework/jobqueue"
)

adapter := jobqueue.NewMessageQueueAdapter(
    "queue-1",
    "MyQueue",
    db,
    docDB,
    signalRClient,
    jobqueue.GlobalQueueManager,
)

// Migrate up to 1000 jobs from DocumentDB Job_History collection
ctx := context.Background()
migratedCount, err := adapter.MigrateFromLegacyJobHistory(ctx, 1000)
if err != nil {
    log.Printf("Migration failed: %v", err)
} else {
    log.Printf("Successfully migrated %d jobs", migratedCount)
}

// To migrate all jobs, set limit to 0
migratedCount, err := adapter.MigrateFromLegacyJobHistory(ctx, 0)
```

### Example 4: Gradual Migration Strategy

```go
import (
    "github.com/mdaxf/iac/framework/queue"
    "github.com/mdaxf/iac/framework/jobqueue"
)

type QueueManager struct {
    useNewSystem bool
    oldQueue     *queue.MessageQueue
    adapter      *jobqueue.MessageQueueAdapter
}

func NewQueueManager(useNew bool) *QueueManager {
    qm := &QueueManager{
        useNewSystem: useNew,
    }

    if useNew {
        // Use new job system
        qm.adapter = jobqueue.NewMessageQueueAdapter(
            "queue-1",
            "MyQueue",
            db,
            docDB,
            signalRClient,
            jobqueue.GlobalQueueManager,
        )
    } else {
        // Use old system
        qm.oldQueue = queue.NewMessageQueue("queue-1", "MyQueue")
    }

    return qm
}

func (qm *QueueManager) PushMessage(msg queue.Message) error {
    if qm.useNewSystem {
        // Convert and push to new system
        return qm.adapter.Push(jobqueue.LegacyMessage{
            Id:        msg.Id,
            UUID:      msg.UUID,
            Retry:     msg.Retry,
            Execute:   msg.Execute,
            Topic:     msg.Topic,
            PayLoad:   msg.PayLoad,
            Handler:   msg.Handler,
            CreatedOn: msg.CreatedOn,
        })
    } else {
        // Push to old system
        qm.oldQueue.Push(msg)
        return nil
    }
}

// Usage
func main() {
    // Start with old system
    qm := NewQueueManager(false)

    // ... operate with old system ...

    // Later, switch to new system
    qm = NewQueueManager(true)

    // Same code, now using new system
    qm.PushMessage(queue.Message{...})
}
```

### Example 5: Direct Job Creation (No Adapter)

If you want to bypass the adapter and use the new system directly:

```go
import (
    "context"
    "github.com/mdaxf/iac/framework/jobqueue"
    "github.com/mdaxf/iac/models"
)

// Create job directly using integration creator
job, err := jobqueue.GlobalJobCreator.CreateJobFromMessage(
    context.Background(),
    "user.created",                    // topic
    `{"user_id": 123}`,               // payload
    "user.onboarding",                // handler
    "POST",                           // method
    "signalr",                        // protocol
    models.JobDirectionInbound,       // direction
    5,                                // priority
)
if err != nil {
    log.Printf("Failed to create job: %v", err)
}

log.Printf("Created job: %s", job.ID)
```

### Example 6: Testing Both Systems Side-by-Side

```go
import (
    "testing"
    "github.com/mdaxf/iac/framework/queue"
    "github.com/mdaxf/iac/framework/jobqueue"
)

func TestMigration(t *testing.T) {
    // Create test message
    testMsg := queue.Message{
        Id:      "test-123",
        UUID:    "test-uuid",
        Retry:   3,
        Topic:   "test.topic",
        PayLoad: []byte(`{"test": true}`),
        Handler: "test.handler",
        CreatedOn: time.Now(),
    }

    // Push to old system
    oldQueue := queue.NewMessageQueue("test-queue", "TestQueue")
    oldQueue.Push(testMsg)

    // Push to new system via adapter
    adapter := jobqueue.NewMessageQueueAdapter(
        "test-queue",
        "TestQueue",
        db,
        docDB,
        signalRClient,
        jobqueue.GlobalQueueManager,
    )

    err := adapter.Push(jobqueue.LegacyMessage{
        Id:        testMsg.Id,
        UUID:      testMsg.UUID,
        Retry:     testMsg.Retry,
        Topic:     testMsg.Topic,
        PayLoad:   testMsg.PayLoad,
        Handler:   testMsg.Handler,
        CreatedOn: testMsg.CreatedOn,
    })

    if err != nil {
        t.Fatalf("Failed to push to new system: %v", err)
    }

    // Verify job was created
    jobService := services.NewJobService(db)
    job, err := jobService.GetJobByID(context.Background(), testMsg.Id)
    if err != nil {
        t.Fatalf("Failed to retrieve job: %v", err)
    }

    if job.Handler != testMsg.Handler {
        t.Errorf("Handler mismatch: got %s, want %s", job.Handler, testMsg.Handler)
    }
}
```

### Example 7: Custom Conversion Logic

If you need custom logic during migration:

```go
import "github.com/mdaxf/iac/framework/jobqueue"

func MigrateWithCustomLogic(oldMsg queue.Message) error {
    // Convert using helper function
    job := jobqueue.ConvertLegacyMessageToJob(
        jobqueue.LegacyMessage{
            Id:        oldMsg.Id,
            UUID:      oldMsg.UUID,
            Retry:     oldMsg.Retry,
            Execute:   oldMsg.Execute,
            Topic:     oldMsg.Topic,
            PayLoad:   oldMsg.PayLoad,
            Handler:   oldMsg.Handler,
            CreatedOn: oldMsg.CreatedOn,
        },
        "MyQueue",
    )

    // Apply custom logic
    if oldMsg.Topic == "critical.alert" {
        job.Priority = 10 // High priority
    }

    // Add custom metadata
    job.Metadata["migration_date"] = time.Now().Format(time.RFC3339)
    job.Metadata["migrated_from"] = "old_system"

    // Save manually
    jobService := services.NewJobService(db)
    return jobService.CreateQueueJob(context.Background(), job)
}
```

### Example 8: Monitoring Migration Progress

```go
import (
    "context"
    "time"
    "github.com/mdaxf/iac/framework/jobqueue"
)

func MonitorMigration(adapter *jobqueue.MessageQueueAdapter) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        // Get job statistics
        jobService := services.NewJobService(db)
        stats, err := jobService.GetJobStatistics(context.Background())
        if err != nil {
            log.Printf("Failed to get stats: %v", err)
            continue
        }

        log.Printf("Migration Progress:")
        log.Printf("  Queue: %s (%s)", adapter.GetQueueName(), adapter.GetQueueID())
        log.Printf("  Statistics: %+v", stats)

        // Check if migration complete
        statusCounts := stats["status_counts"].(map[int]int)
        pendingCount := statusCounts[int(models.JobStatusPending)]

        if pendingCount == 0 {
            log.Printf("Migration complete - no pending jobs")
            return
        }

        log.Printf("  Pending jobs: %d", pendingCount)
    }
}
```

## Best Practices

1. **Test in Development First**: Always test migration in a development environment
2. **Gradual Migration**: Migrate one queue at a time, not all at once
3. **Monitor Performance**: Watch for any performance degradation
4. **Keep History**: Don't delete Job_History collection until migration is verified
5. **Verify Results**: Confirm jobs are executing correctly in the new system
6. **Have Rollback Plan**: Keep old system available for quick rollback if needed

## Troubleshooting

### Problem: Messages not appearing in new system

**Solution:**
```go
// Check if job was created
jobService := services.NewJobService(db)
job, err := jobService.GetJobByID(ctx, messageID)
if err != nil {
    log.Printf("Job not found: %v", err)
}

// Check job status
log.Printf("Job Status: %d", job.StatusID)
log.Printf("Job Handler: %s", job.Handler)
```

### Problem: Legacy payload format not working

**Solution:** The adapter automatically creates the Topic/Payload structure. Verify:
```go
var payload map[string]interface{}
json.Unmarshal([]byte(job.Payload), &payload)

fmt.Printf("Topic: %v\n", payload["Topic"])
fmt.Printf("Payload: %v\n", payload["Payload"])
```

### Problem: Job_History not being saved to DocumentDB

**Solution:** Ensure DocumentDB is connected when initializing:
```go
err := jobqueue.InitializeJobSystem(
    ctx,
    db,
    docDB,  // Must not be nil
    cache,
    signalRClient,
)
```
