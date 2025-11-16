package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mdaxf/iac/engine/debug"
)

// Example of integrating the debug API into your HTTP server
func main() {
	// 1. Configure debug settings
	config := &debug.DebugConfig{
		Enabled:                true,
		MaxEventsPerSession:    10000,
		SubscriberBufferSize:   100,
		SubscriberTimeout:      5 * time.Minute,
		CleanupInterval:        1 * time.Minute,
		MaxSessionAge:          1 * time.Hour,
		SessionCleanupInterval: 10 * time.Minute,
		MaxConcurrentSessions:  100,
		SanitizeSensitiveData:  true,
		SensitiveFields: []string{
			"password", "token", "api_key", "secret",
		},
		MaxDataSize:    10 * 1024 * 1024, // 10MB
		MinLogLevel:    "DEBUG",
	}

	// Set global debug configuration
	debug.SetGlobalDebugConfig(config)

	// 2. Create HTTP server mux
	mux := http.NewServeMux()

	// 3. Create SSE handler
	messageBus := debug.GetGlobalMessageBus()
	sseHandler := debug.NewSSEHandler(messageBus)

	// 4. Register debug API routes
	sseHandler.RegisterRoutes(mux)

	// 5. Start cleanup routines
	sessionCleanupCancel := sseHandler.StartCleanupRoutine(
		10*time.Minute, // cleanup interval
		1*time.Hour,    // max session age
	)
	defer sessionCleanupCancel()

	// 6. Add your application routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("BPM Engine Server\n"))
		w.Write([]byte("Debug API available at /api/debug/\n"))
	})

	// 7. Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 8. Start HTTP server
	addr := ":8080"
	fmt.Printf("Starting server on %s\n", addr)
	fmt.Println("Debug endpoints:")
	fmt.Println("  - POST   /api/debug/sessions          - Start a debug session")
	fmt.Println("  - GET    /api/debug/sessions          - List all debug sessions")
	fmt.Println("  - GET    /api/debug/sessions?sessionID=X - Get session details")
	fmt.Println("  - POST   /api/debug/sessions/stop     - Stop a debug session")
	fmt.Println("  - GET    /api/debug/sessions/trace    - Get execution trace")
	fmt.Println("  - GET    /api/debug/stream            - SSE stream for events")
	fmt.Println()
	fmt.Println("Open examples/client.html in a browser to monitor debug sessions")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// Alternative: Integrate into existing http.ServeMux
func IntegrateIntoExistingServer(existingMux *http.ServeMux) {
	// Get or create message bus
	messageBus := debug.GetGlobalMessageBus()

	// Create SSE handler
	sseHandler := debug.NewSSEHandler(messageBus)

	// Register routes
	sseHandler.RegisterRoutes(existingMux)

	// Start cleanup routines
	sseHandler.StartCleanupRoutine(10*time.Minute, 1*time.Hour)

	fmt.Println("Debug API integrated successfully")
}

// Alternative: Using a custom router (e.g., gorilla/mux, chi, etc.)
func IntegrateWithCustomRouter() {
	// Example with standard library patterns
	// You can adapt this for your router

	messageBus := debug.GetGlobalMessageBus()
	sseHandler := debug.NewSSEHandler(messageBus)

	// Manual route registration if not using http.ServeMux
	http.HandleFunc("/api/debug/stream", sseHandler.StreamEvents)
	http.HandleFunc("/api/debug/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			sseHandler.StartDebugSession(w, r)
		case http.MethodGet:
			if r.URL.Query().Get("sessionID") != "" {
				sseHandler.GetDebugSession(w, r)
			} else {
				sseHandler.ListDebugSessions(w, r)
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/debug/sessions/stop", sseHandler.StopDebugSession)
	http.HandleFunc("/api/debug/sessions/trace", sseHandler.GetExecutionTrace)

	fmt.Println("Debug API routes registered with custom router")
}

// Example: Enable debug for specific session
func EnableDebugForSession(sessionID, tranCodeName, userID string) error {
	// Get debug session manager
	manager := debug.GetGlobalDebugSessionManager()

	// Create and start debug session
	session := manager.CreateSession(sessionID, tranCodeName, userID)
	session.Description = "Debug session for troubleshooting"
	session.Start()

	fmt.Printf("Debug session started: %s\n", sessionID)
	return nil
}

// Example: Programmatically emit events
func EmitCustomDebugEvents(sessionID, tranCodeName string) {
	// Create debug helper
	helper := debug.NewDebugHelper(sessionID, tranCodeName, "v1.0")

	// Check if debug is enabled (prevents overhead when disabled)
	if !helper.IsEnabled() {
		return
	}

	// Emit various events
	helper.EmitTranCodeStart()

	inputs := map[string]interface{}{
		"orderId":   "ORD-123",
		"productId": "PROD-456",
		"quantity":  10,
	}

	helper.EmitFunctionStart("ValidateOrder", "CheckInventory", "Query", inputs)

	// ... function execution ...

	outputs := map[string]interface{}{
		"available": true,
		"stock":     50,
	}

	startTime := time.Now()
	endTime := startTime.Add(15 * time.Millisecond)

	helper.EmitFunctionComplete(
		"ValidateOrder",
		"CheckInventory",
		"Query",
		outputs,
		startTime,
		endTime,
	)

	// Emit routing
	helper.EmitFuncGroupRouting("ValidateOrder", "approved", "ProcessPayment")

	// Complete trancode
	helper.EmitTranCodeComplete(time.Second, outputs)
}

// Example: CORS configuration for SSE (if needed for web clients)
func WithCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
