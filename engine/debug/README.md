# BPM Engine Remote Debugging

Real-time remote debugging system for the BPM execution engine with Server-Sent Events (SSE) and message bus support.

## Features

- **Real-time Event Streaming**: Stream execution events to multiple clients via SSE
- **Step-by-Step Execution Tracking**: Monitor every step of trancode execution
- **Detailed Execution Data**: Captures inputs, outputs, timing, and routing information
- **Multi-Client Support**: Multiple clients can monitor the same debug session
- **Event Filtering**: Filter events by type, log level, trancode, or function type
- **Flexible Architecture**: Uses message bus for async event distribution
- **Security**: Automatic sanitization of sensitive data (passwords, tokens, etc.)
- **Performance**: Minimal overhead when debug is disabled

## Architecture

```
┌─────────────┐
│  TranCode   │
│  Execution  │
└──────┬──────┘
       │ Emits Debug Events
       ↓
┌──────────────────┐
│  DebugHelper     │
│  (Event Builder) │
└──────┬───────────┘
       │
       ↓
┌──────────────────┐
│  MessageBus      │
│  (Distribution)  │
└──────┬───────────┘
       │
       ├─────────────┬─────────────┬─────────────┐
       ↓             ↓             ↓             ↓
  Subscriber 1  Subscriber 2  Subscriber 3  ...
       │             │             │
       ↓             ↓             ↓
   SSE Client   SSE Client   SSE Client
```

## Event Types

The following event types are emitted during trancode execution:

| Event Type | Description |
|------------|-------------|
| `trancode.start` | TranCode execution started |
| `trancode.complete` | TranCode execution completed |
| `funcgroup.start` | FuncGroup execution started |
| `funcgroup.complete` | FuncGroup execution completed |
| `funcgroup.routing` | FuncGroup routing decision |
| `function.start` | Function execution started |
| `function.complete` | Function execution completed |
| `input.mapping` | Input mapping performed |
| `output.mapping` | Output mapping performed |
| `database.query` | Database query executed |
| `script.execution` | Script executed (Python, C#, JS, etc.) |
| `transaction.begin` | Database transaction started |
| `transaction.commit` | Database transaction committed |
| `transaction.rollback` | Database transaction rolled back |

## Event Data Structure

Each debug event contains:

```json
{
  "id": "event-unique-id",
  "session_id": "debug-session-123",
  "timestamp": "2025-11-16T10:30:45.123Z",
  "event_type": "function.complete",
  "level": "INFO",

  "trancode_name": "OrderProcessing",
  "trancode_version": "v1.0",
  "funcgroup_name": "ValidateOrder",
  "function_name": "CheckInventory",
  "function_type": "Query",

  "execution_step": 5,
  "execution_time": 1234567,
  "start_time": "2025-11-16T10:30:45.000Z",
  "end_time": "2025-11-16T10:30:45.001Z",

  "inputs": {
    "productId": "P123",
    "quantity": 10
  },
  "outputs": {
    "available": true,
    "stock": 50
  },
  "routing_value": "success",
  "routing_path": "ValidateOrder.CheckPayment",

  "message": "Function 'CheckInventory' completed successfully",
  "error": null,
  "metadata": {
    "query": "SELECT * FROM inventory WHERE product_id = ?",
    "duration_ms": 1.23
  }
}
```

## Usage

### 1. Enable Debug Mode

```go
import "github.com/mdaxf/iac/engine/debug"

// Enable debug globally
debug.EnableGlobalDebug()

// Or configure custom settings
config := &debug.DebugConfig{
    Enabled:              true,
    MaxEventsPerSession:  10000,
    SanitizeSensitiveData: true,
}
debug.SetGlobalDebugConfig(config)
```

### 2. Start a Debug Session

**Via REST API:**

```bash
curl -X POST http://localhost:8080/api/debug/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "sessionID": "debug-session-123",
    "tranCodeName": "OrderProcessing",
    "userID": "user-456",
    "description": "Debugging order processing flow"
  }'
```

**Programmatically:**

```go
manager := debug.GetGlobalDebugSessionManager()
session := manager.CreateSession("debug-session-123", "OrderProcessing", "user-456")
session.Description = "Debugging order processing flow"
session.Start()
```

### 3. Connect to SSE Stream

**JavaScript Client:**

```javascript
const eventSource = new EventSource(
  'http://localhost:8080/api/debug/stream?sessionID=debug-session-123'
);

eventSource.onmessage = function(e) {
  const event = JSON.parse(e.data);
  console.log('Event:', event);
};

eventSource.onerror = function(err) {
  console.error('Connection error:', err);
};
```

**Go Client:**

```go
client := NewDebugClient("http://localhost:8080", "debug-session-123")

if err := client.Connect(); err != nil {
    log.Fatal(err)
}

for event := range client.Events() {
    fmt.Printf("Event: %s - %s\n", event.EventType, event.Message)
}
```

**Using the HTML Client:**

Open `examples/client.html` in a browser:
1. Enter API URL (e.g., `http://localhost:8080`)
2. Enter Session ID
3. Click "Connect"
4. Monitor real-time events

### 4. Instrument Your Code

**In TranCode Execution:**

```go
import "github.com/mdaxf/iac/engine/debug"

func (t *TranCode) Execute() (map[string]interface{}, error) {
    // Create debug helper
    debugHelper := debug.NewDebugHelper(
        t.SessionID,
        t.Tcode.Name,
        t.Tcode.Version,
    )

    // Emit trancode start
    debugHelper.EmitTranCodeStart()

    startTime := time.Now()

    // ... execute trancode logic ...

    // Emit trancode complete
    duration := time.Since(startTime)
    debugHelper.EmitTranCodeComplete(duration, outputs)

    return outputs, nil
}
```

**In Function Execution:**

```go
func ExecuteFunction(funcDef *types.Function, inputs map[string]interface{}) (map[string]interface{}, error) {
    debugHelper := debug.NewDebugHelper(sessionID, tranCodeName, version)

    // Emit function start with inputs
    debugHelper.EmitFunctionStart(
        funcGroupName,
        funcDef.Name,
        string(funcDef.Functype),
        inputs,
    )

    startTime := time.Now()

    // Execute function
    outputs, err := doExecute(funcDef, inputs)

    // Emit function complete with outputs and timing
    debugHelper.EmitFunctionComplete(
        funcGroupName,
        funcDef.Name,
        string(funcDef.Functype),
        outputs,
        startTime,
        time.Now(),
    )

    return outputs, err
}
```

**Emit Routing Events:**

```go
debugHelper.EmitFuncGroupRouting(
    funcGroupName,
    routingValue,      // The value used for routing decision
    nextFuncGroupName, // The path taken
)
```

**Emit Script Execution:**

```go
debugHelper.EmitScriptExecution(
    funcGroupName,
    functionName,
    "Python",          // Script type
    scriptContent,
    executionDuration,
)
```

### 5. Event Filtering

Filter events on the client side:

```javascript
// Filter by event types
const url = 'http://localhost:8080/api/debug/stream?' +
  'sessionID=debug-session-123&' +
  'eventTypes=["function.start","function.complete"]&' +
  'minLevel=INFO';

const eventSource = new EventSource(url);
```

Available filters:
- `eventTypes`: Array of event types to include
- `minLevel`: Minimum log level (DEBUG, INFO, WARNING, ERROR)
- `tranCodes`: Array of trancode names to include
- `functionTypes`: Array of function types to include

### 6. Stop Debug Session

```bash
curl -X POST http://localhost:8080/api/debug/sessions/stop?sessionID=debug-session-123
```

### 7. Get Execution Trace

Retrieve the complete execution trace after session completes:

```bash
curl http://localhost:8080/api/debug/sessions/trace?sessionID=debug-session-123
```

Returns:

```json
{
  "session_id": "debug-session-123",
  "trancode_name": "OrderProcessing",
  "start_time": "2025-11-16T10:30:00Z",
  "end_time": "2025-11-16T10:30:05Z",
  "status": "completed",
  "event_count": 145,
  "events": [
    {
      "id": "event-1",
      "event_type": "trancode.start",
      ...
    },
    ...
  ]
}
```

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/debug/stream` | GET | SSE stream for debug events |
| `/api/debug/sessions` | POST | Start a new debug session |
| `/api/debug/sessions` | GET | List all debug sessions |
| `/api/debug/sessions` | GET | Get session details (with `sessionID` param) |
| `/api/debug/sessions/stop` | POST | Stop a debug session |
| `/api/debug/sessions/trace` | GET | Get execution trace |

## Configuration

```go
config := &debug.DebugConfig{
    // Enable/disable debug globally
    Enabled: true,

    // Maximum events to store per session
    MaxEventsPerSession: 10000,

    // Buffer size for each subscriber
    SubscriberBufferSize: 100,

    // Timeout for inactive subscribers
    SubscriberTimeout: 5 * time.Minute,

    // Cleanup intervals
    CleanupInterval: 1 * time.Minute,
    MaxSessionAge: 1 * time.Hour,

    // Sanitize sensitive data
    SanitizeSensitiveData: true,
    SensitiveFields: []string{
        "password", "token", "api_key", "secret",
    },

    // Maximum data size in events
    MaxDataSize: 10 * 1024 * 1024, // 10MB

    // Event filtering
    ExcludedEventTypes: []EventType{},
    MinLogLevel: "DEBUG",
}
```

## Security Considerations

1. **Sensitive Data**: Automatically sanitizes passwords, tokens, API keys, etc.
2. **Data Size Limits**: Prevents memory exhaustion from large inputs/outputs
3. **Session Limits**: Maximum concurrent sessions configurable
4. **Subscriber Timeout**: Inactive subscribers automatically cleaned up
5. **Authentication**: Add authentication middleware to SSE endpoints in production

## Performance Impact

When debug is **disabled** (default):
- **Zero overhead**: All `IsEnabled()` checks short-circuit immediately
- No event creation or serialization
- No message bus publishing

When debug is **enabled**:
- **Minimal overhead**: Event creation and publishing is non-blocking
- Events sent to subscribers in parallel goroutines
- Slow subscribers don't block execution (buffered channels with timeout)
- Configurable buffer sizes and timeouts

## Examples

See the `examples/` directory for:
- `client.html`: Interactive web-based debug monitor
- `go_client.go`: Command-line Go client for monitoring events

## Integration with Existing Code

The debug system is designed to integrate seamlessly:

1. **Non-intrusive**: All debug calls check `IsEnabled()` first
2. **Zero dependencies**: Core execution code doesn't depend on debug package
3. **Backward compatible**: Works with existing trancode execution
4. **Optional**: Can be completely disabled with zero performance impact

## Troubleshooting

**No events appearing:**
- Check that debug is enabled: `debug.IsGlobalDebugEnabled()`
- Verify debug session is started and running
- Check session ID matches between client and server

**Connection timeout:**
- Ensure SSE endpoint is accessible
- Check for network issues or proxies blocking SSE
- Verify CORS headers if accessing from different origin

**Missing event data:**
- Check `MaxDataSize` configuration
- Verify sensitive data isn't being filtered out
- Check event type filters

**High memory usage:**
- Reduce `MaxEventsPerSession`
- Enable session cleanup
- Reduce `SubscriberBufferSize`

## Best Practices

1. **Development**: Enable debug for all test environments
2. **Production**: Only enable debug for specific troubleshooting sessions
3. **Event Filtering**: Use filters to reduce noise and focus on relevant events
4. **Session Management**: Always stop debug sessions when done
5. **Monitoring**: Monitor subscriber count and event rates
6. **Data Sanitization**: Always enable in production environments

## License

Part of the IAC (Intelligent Automation Core) BPM Engine.
