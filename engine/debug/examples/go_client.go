package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// DebugEvent represents a debug event from the SSE stream
type DebugEvent struct {
	ID              string                 `json:"id"`
	SessionID       string                 `json:"session_id"`
	Timestamp       time.Time              `json:"timestamp"`
	EventType       string                 `json:"event_type"`
	Level           string                 `json:"level"`
	TranCodeName    string                 `json:"trancode_name,omitempty"`
	TranCodeVersion string                 `json:"trancode_version,omitempty"`
	FuncGroupName   string                 `json:"funcgroup_name,omitempty"`
	FunctionName    string                 `json:"function_name,omitempty"`
	FunctionType    string                 `json:"function_type,omitempty"`
	ExecutionStep   int                    `json:"execution_step"`
	ExecutionTime   int64                  `json:"execution_time,omitempty"`
	Inputs          map[string]interface{} `json:"inputs,omitempty"`
	Outputs         map[string]interface{} `json:"outputs,omitempty"`
	RoutingValue    interface{}            `json:"routing_value,omitempty"`
	RoutingPath     string                 `json:"routing_path,omitempty"`
	Message         string                 `json:"message,omitempty"`
	Error           string                 `json:"error,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DebugClient is a client for consuming debug events via SSE
type DebugClient struct {
	BaseURL   string
	SessionID string
	events    chan *DebugEvent
	errors    chan error
}

// NewDebugClient creates a new debug client
func NewDebugClient(baseURL, sessionID string) *DebugClient {
	return &DebugClient{
		BaseURL:   baseURL,
		SessionID: sessionID,
		events:    make(chan *DebugEvent, 100),
		errors:    make(chan error, 10),
	}
}

// Connect connects to the SSE stream
func (dc *DebugClient) Connect() error {
	url := fmt.Sprintf("%s/api/debug/stream?sessionID=%s", dc.BaseURL, dc.SessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{
		Timeout: 0, // No timeout for SSE
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Start reading SSE stream
	go dc.readStream(resp)

	return nil
}

// readStream reads the SSE stream
func (dc *DebugClient) readStream(resp *http.Response) {
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var eventType string
	var data strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			// Empty line marks end of event
			if data.Len() > 0 {
				dc.processEvent(eventType, data.String())
				data.Reset()
				eventType = ""
			}
			continue
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			if data.Len() > 0 {
				data.WriteString("\n")
			}
			data.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "data:")))
		}
	}

	if err := scanner.Err(); err != nil {
		dc.errors <- fmt.Errorf("stream error: %w", err)
	}
}

// processEvent processes a received event
func (dc *DebugClient) processEvent(eventType, data string) {
	if eventType == "connected" {
		fmt.Printf("Connected to debug session: %s\n", data)
		return
	}

	var event DebugEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		dc.errors <- fmt.Errorf("failed to parse event: %w", err)
		return
	}

	dc.events <- &event
}

// Events returns the events channel
func (dc *DebugClient) Events() <-chan *DebugEvent {
	return dc.events
}

// Errors returns the errors channel
func (dc *DebugClient) Errors() <-chan error {
	return dc.errors
}

// StartDebugSession starts a new debug session
func (dc *DebugClient) StartDebugSession(tranCodeName, userID, description string) error {
	url := fmt.Sprintf("%s/api/debug/sessions", dc.BaseURL)

	payload := map[string]string{
		"sessionID":    dc.SessionID,
		"tranCodeName": tranCodeName,
		"userID":       userID,
		"description":  description,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// StopDebugSession stops the debug session
func (dc *DebugClient) StopDebugSession() error {
	url := fmt.Sprintf("%s/api/debug/sessions/stop?sessionID=%s", dc.BaseURL, dc.SessionID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to stop session: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Example usage
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run go_client.go <base-url> <session-id>")
		fmt.Println("Example: go run go_client.go http://localhost:8080 debug-session-123")
		os.Exit(1)
	}

	baseURL := os.Args[1]
	sessionID := os.Args[2]

	client := NewDebugClient(baseURL, sessionID)

	// Connect to SSE stream
	fmt.Printf("Connecting to debug session: %s\n", sessionID)
	if err := client.Connect(); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connected! Waiting for events...")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	// Process events
	for {
		select {
		case event := <-client.Events():
			displayEvent(event)

		case err := <-client.Errors():
			fmt.Printf("ERROR: %v\n", err)
		}
	}
}

// displayEvent displays a debug event
func displayEvent(event *DebugEvent) {
	timestamp := event.Timestamp.Format("15:04:05.000")

	// Color codes for different log levels
	colors := map[string]string{
		"DEBUG":   "\033[36m", // Cyan
		"INFO":    "\033[32m", // Green
		"WARNING": "\033[33m", // Yellow
		"ERROR":   "\033[31m", // Red
	}
	reset := "\033[0m"

	color := colors[event.Level]
	if color == "" {
		color = colors["INFO"]
	}

	fmt.Printf("%s[%s] [Step %d] %s%s\n",
		color, timestamp, event.ExecutionStep, event.EventType, reset)

	if event.Message != "" {
		fmt.Printf("  Message: %s\n", event.Message)
	}

	if event.FunctionName != "" {
		fmt.Printf("  Function: %s (%s)\n", event.FunctionName, event.FunctionType)
	}

	if event.FuncGroupName != "" {
		fmt.Printf("  FuncGroup: %s\n", event.FuncGroupName)
	}

	if event.ExecutionTime > 0 {
		ms := float64(event.ExecutionTime) / 1000000.0
		fmt.Printf("  Duration: %.2fms\n", ms)
	}

	if event.RoutingPath != "" {
		fmt.Printf("  Routing: %s (value: %v)\n", event.RoutingPath, event.RoutingValue)
	}

	if len(event.Inputs) > 0 {
		fmt.Println("  Inputs:")
		for key, value := range event.Inputs {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	if len(event.Outputs) > 0 {
		fmt.Println("  Outputs:")
		for key, value := range event.Outputs {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	if event.Error != "" {
		fmt.Printf("  %sError: %s%s\n", colors["ERROR"], event.Error, reset)
	}

	if len(event.Metadata) > 0 {
		fmt.Println("  Metadata:")
		for key, value := range event.Metadata {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	fmt.Println()
}
