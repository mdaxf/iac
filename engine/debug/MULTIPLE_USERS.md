# Remote Debugging - Multiple Users Monitoring Same Session

## How It Works

Each debug session is identified by a **unique Session ID**. Multiple users can connect to the **same session ID** and monitor the execution simultaneously via SSE (Server-Sent Events).

## Architecture

```
┌─────────────────┐
│  TranCode       │
│  Execution      │  Session ID: "debug-session-123"
│  (SessionID in  │
│  SystemSession) │
└────────┬────────┘
         │
         │ Emits Debug Events
         ↓
┌────────────────────┐
│  MessageBus        │
│  (Routes by        │
│  SessionID)        │
└────────┬───────────┘
         │
         │ Publishes to all subscribers
         │ of "debug-session-123"
         │
    ┌────┴────┬─────────┬─────────┐
    ↓         ↓         ↓         ↓
Subscriber Subscriber Subscriber Subscriber
    1         2         3         N
    │         │         │         │
    ↓         ↓         ↓         ↓
  User A    User B    User C    User N
  (Web UI)  (CLI)     (Web UI)  (API)
```

All users see the **exact same events** in real-time.

## Session ID Flow

### 1. Session ID in TranCode Execution

The Session ID comes from the `SystemSession` when executing a trancode:

```go
systemSession := map[string]interface{}{
    "SessionID": "debug-session-123",  // Unique ID
    "UserNo":    "user-456",
    "ClientID":  "client-789",
}

outputs, err := trancode.Execute("OrderProcessing", inputs, systemSession)
```

### 2. Automatic Debug Event Emission

The trancode execution automatically detects the Session ID and emits debug events:

```go
// In trancode.go Execute() method
var debugHelper *debug.DebugHelper
if t.SystemSession["SessionID"] != nil {
    if sid, ok := t.SystemSession["SessionID"].(string); ok {
        sessionID = sid
        // Creates helper for THIS specific session
        debugHelper = debug.NewDebugHelper(sessionID, t.Tcode.Name, t.Tcode.Version)
    }
}

// All events from this execution will have SessionID = "debug-session-123"
debugHelper.EmitTranCodeStart()
```

### 3. Multiple Users Connect to Same Session

**User A** (Web Browser):
```javascript
const eventSource = new EventSource(
  'http://localhost:8080/api/debug/stream?sessionID=debug-session-123'
);

eventSource.onmessage = function(e) {
    const event = JSON.parse(e.data);
    console.log('User A sees:', event);
};
```

**User B** (Another Web Browser):
```javascript
const eventSource = new EventSource(
  'http://localhost:8080/api/debug/stream?sessionID=debug-session-123'  // SAME ID
);

eventSource.onmessage = function(e) {
    const event = JSON.parse(e.data);
    console.log('User B sees:', event);
};
```

**User C** (Go CLI):
```bash
go run go_client.go http://localhost:8080 debug-session-123
```

**All three users see the same events in real-time!**

## Example: Multiple Users Monitoring

### Scenario: Production Debugging

1. **Developer starts trancode execution** with debug session:
```go
systemSession := map[string]interface{}{
    "SessionID": "prod-issue-2024-001",  // Unique ID for this debug session
    "UserNo":    "system",
}

trancode.Execute("OrderProcessing", orderData, systemSession)
```

2. **Developer A** opens web UI:
   - URL: `http://localhost:8080/debug/client.html`
   - Enter Session ID: `prod-issue-2024-001`
   - Clicks "Connect"
   - Sees real-time execution events

3. **Developer B** (on different machine) opens same session:
   - Opens `http://localhost:8080/debug/client.html`
   - Enter **SAME** Session ID: `prod-issue-2024-001`
   - Clicks "Connect"
   - Sees **exact same events** as Developer A

4. **Senior Engineer** uses CLI from terminal:
```bash
go run go_client.go http://localhost:8080 prod-issue-2024-001
```
   - Sees same events in terminal with color coding

5. **QA Tester** uses curl to save events:
```bash
curl -N http://localhost:8080/api/debug/stream?sessionID=prod-issue-2024-001 > debug.log
```

**All 4 people are watching the SAME execution in real-time!**

## Session ID Generation Strategies

### 1. Auto-Generated (Recommended for Production)

```go
import (
    "fmt"
    "time"
    "github.com/google/uuid"
)

// UUID-based
sessionID := uuid.New().String()
// Example: "550e8400-e29b-41d4-a716-446655440000"

// Timestamp-based
sessionID := fmt.Sprintf("debug-%d", time.Now().UnixNano())
// Example: "debug-1700000000123456789"

// Descriptive with timestamp
sessionID := fmt.Sprintf("order-processing-%s", time.Now().Format("20060102-150405"))
// Example: "order-processing-20241116-143022"
```

### 2. User-Defined (Good for Testing)

```go
// Easy to remember and share with team
sessionID := "fixing-payment-bug"
sessionID := "sprint-23-testing"
sessionID := "customer-123-issue"
```

### 3. Correlation ID (Best for Distributed Systems)

```go
// Use existing correlation ID from request
sessionID := request.Header.Get("X-Correlation-ID")
// Example: "corr-abc123xyz"
```

## Complete Working Example

### Server Setup

```go
package main

import (
    "fmt"
    "net/http"
    "github.com/mdaxf/iac/engine/debug"
)

func main() {
    // Enable debug globally
    debug.EnableGlobalDebug()

    // Create message bus and SSE handler
    messageBus := debug.GetGlobalMessageBus()
    sseHandler := debug.NewSSEHandler(messageBus)

    // Register routes
    mux := http.NewServeMux()
    sseHandler.RegisterRoutes(mux)

    // Serve static files (web UI)
    mux.Handle("/debug/", http.StripPrefix("/debug/",
        http.FileServer(http.Dir("engine/debug/examples"))))

    fmt.Println("Debug server running on :8080")
    fmt.Println("Multiple users can monitor the same session ID")
    http.ListenAndServe(":8080", mux)
}
```

### Execute TranCode with Session ID

```go
package main

import (
    "fmt"
    "github.com/mdaxf/iac/engine/trancode"
)

func main() {
    // Create unique session ID
    sessionID := "team-debugging-session-001"

    fmt.Printf("Starting debug session: %s\n", sessionID)
    fmt.Printf("Team members can connect at:\n")
    fmt.Printf("  Web: http://localhost:8080/debug/client.html\n")
    fmt.Printf("  Enter Session ID: %s\n\n", sessionID)

    // Execute trancode with this session ID
    systemSession := map[string]interface{}{
        "SessionID": sessionID,  // This enables debug for this session
        "UserNo":    "system",
        "ClientID":  "app-001",
    }

    inputs := map[string]interface{}{
        "orderId": "ORD-12345",
        "amount":  100.00,
    }

    outputs, err := trancode.Execute("OrderProcessing", inputs, systemSession)

    if err != nil {
        fmt.Printf("Error: %v\n", err)
    } else {
        fmt.Printf("Success: %v\n", outputs)
    }
}
```

### Users Connect (Multiple Browsers/Terminals)

**User 1 - Web Browser:**
1. Open `http://localhost:8080/debug/client.html`
2. Enter Session ID: `team-debugging-session-001`
3. Click "Connect"

**User 2 - Another Web Browser (different machine):**
1. Open `http://localhost:8080/debug/client.html`
2. Enter **SAME** Session ID: `team-debugging-session-001`
3. Click "Connect"

**User 3 - Terminal:**
```bash
cd engine/debug/examples
go run go_client.go http://localhost:8080 team-debugging-session-001
```

**User 4 - Curl:**
```bash
curl -N http://localhost:8080/api/debug/stream?sessionID=team-debugging-session-001
```

## Verification: Multiple Users See Same Events

When the trancode executes, **all connected users see**:

```
Event 1: trancode.start
{
  "session_id": "team-debugging-session-001",
  "event_type": "trancode.start",
  "trancode_name": "OrderProcessing",
  "execution_step": 1,
  ...
}

Event 2: funcgroup.start
{
  "session_id": "team-debugging-session-001",
  "event_type": "funcgroup.start",
  "funcgroup_name": "ValidateOrder",
  "execution_step": 2,
  ...
}

Event 3: function.start
{
  "session_id": "team-debugging-session-001",
  "event_type": "function.start",
  "function_name": "CheckInventory",
  "inputs": {
    "orderId": "ORD-12345",
    "productId": "PROD-789"
  },
  "execution_step": 3,
  ...
}
```

## Benefits of Shared Session Monitoring

1. **Team Collaboration**:
   - Senior engineer can guide junior developer in real-time
   - Multiple team members can analyze the same execution

2. **Cross-Functional Debugging**:
   - Developer watches execution flow
   - DBA watches database queries
   - QA verifies test scenarios

3. **Training & Knowledge Transfer**:
   - Trainer executes trancode
   - Multiple trainees watch execution in real-time
   - Everyone sees inputs, outputs, timing

4. **Production Support**:
   - Support team monitors production issue
   - Development team watches simultaneously
   - Management can observe without interfering

5. **Audit & Compliance**:
   - Multiple auditors can review execution
   - All see the same immutable event stream

## Session Management

### List All Active Sessions

```bash
curl http://localhost:8080/api/debug/sessions
```

Response:
```json
{
  "sessions": [
    {
      "sessionID": "team-debugging-session-001",
      "tranCodeName": "OrderProcessing",
      "status": "running",
      "startTime": "2024-11-16T10:30:00Z",
      "eventCount": 145
    },
    {
      "sessionID": "prod-issue-2024-001",
      "tranCodeName": "PaymentProcessing",
      "status": "completed",
      "startTime": "2024-11-16T09:15:00Z",
      "eventCount": 89
    }
  ],
  "total": 2
}
```

### Share Session with Team

Send team members:
- **Session ID**: `team-debugging-session-001`
- **SSE URL**: `http://localhost:8080/api/debug/stream?sessionID=team-debugging-session-001`
- **Web UI**: `http://localhost:8080/debug/client.html` (then enter session ID)

## Security Considerations

1. **Session ID should be unpredictable** in production:
   ```go
   sessionID := uuid.New().String()  // Use UUID, not "session-1", "session-2"
   ```

2. **Add authentication** to SSE endpoint in production:
   ```go
   func (h *SSEHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
       // Verify user has permission to view this session
       if !isAuthorized(r, sessionID) {
           http.Error(w, "Unauthorized", http.StatusUnauthorized)
           return
       }
       // ... continue with SSE streaming
   }
   ```

3. **Limit session lifetime** (already implemented):
   - Sessions auto-cleanup after MaxSessionAge (default 1 hour)
   - Configurable in DebugConfig

## Troubleshooting

**Problem**: User B doesn't see User A's events

**Solution**: Verify both users connected to **exact same Session ID**:
```javascript
// User A
sessionID: "debug-session-123"

// User B (WRONG - different session)
sessionID: "debug-session-124"  // ❌ Typo!

// User B (CORRECT - same session)
sessionID: "debug-session-123"  // ✅ Same as User A
```

**Problem**: No events appearing

**Solution**:
1. Verify debug is enabled: `debug.IsGlobalDebugEnabled()`
2. Verify trancode execution includes SessionID in SystemSession
3. Check browser console for connection errors
4. Verify SSE endpoint is accessible

## Summary

✅ **Each debug session has a unique Session ID**
✅ **Multiple users can connect to the same Session ID**
✅ **All users see the same events in real-time**
✅ **Session ID is specified in SystemSession["SessionID"]**
✅ **Users connect via SSE with ?sessionID=<unique-id>**
✅ **Works with web UI, CLI, curl, or any SSE client**

The system is designed from the ground up for **collaborative debugging**!
