package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// MCPServerService manages MCP server catalog entries and MCP client calls.
// CRUD operations now target MongoDB; FetchTools and CallTool remain unchanged.
type MCPServerService struct {
	db    *gorm.DB
	sqlDB *sql.DB
	docDB documents.DocumentDB
	iLog  logger.Log
}

var globalMCPServerService *MCPServerService

func GetGlobalMCPServerService() *MCPServerService  { return globalMCPServerService }
func SetGlobalMCPServerService(s *MCPServerService) { globalMCPServerService = s }

// NewMCPServerService creates the service backed by MongoDB for CRUD.
func NewMCPServerService(db *gorm.DB, sqlDB *sql.DB, docDB documents.DocumentDB) *MCPServerService {
	return &MCPServerService{
		db:    db,
		sqlDB: sqlDB,
		docDB: docDB,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "MCPServerService",
		},
	}
}

// ---- Request types ----

type CreateMCPServerRequest struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	TransportType string            `json:"transporttype"`
	URL           string            `json:"url"`
	MCPPath       string            `json:"mcppath"`
	Command       string            `json:"command"`
	Args          []string          `json:"args"`
	Headers       map[string]string `json:"headers"`
	Enabled       bool              `json:"enabled"`
	CreatedBy     string            `json:"createdby"`
}

type UpdateMCPServerRequest struct {
	Name          *string            `json:"name"`
	Description   *string            `json:"description"`
	TransportType *string            `json:"transporttype"`
	URL           *string            `json:"url"`
	MCPPath       *string            `json:"mcppath"`
	Command       *string            `json:"command"`
	Args          []string           `json:"args"`
	Headers       map[string]string  `json:"headers"`
	Enabled       *bool              `json:"enabled"`
}

// ensureDocDB returns an error if the document database connection is not initialized.
func (s *MCPServerService) ensureDocDB() error {
	if s.docDB == nil {
		return fmt.Errorf("document database not initialized for MCPServerService — ensure MongoDB is configured and reachable at startup")
	}
	return nil
}

// ---- CRUD ----

func (s *MCPServerService) CreateMCPServer(req CreateMCPServerRequest) (*models.MCPServer, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, fmt.Errorf("MCP server name is required")
	}
	if req.TransportType == "" {
		req.TransportType = "http"
	}
	if req.MCPPath == "" {
		req.MCPPath = "/mcp"
	}
	if req.Args == nil {
		req.Args = []string{}
	}
	if req.Headers == nil {
		req.Headers = map[string]string{}
	}

	now := time.Now()
	doc := &models.MCPServerDoc{
		ID:              uuid.New().String(),
		Name:            req.Name,
		Description:     req.Description,
		TransportType:   req.TransportType,
		URL:             req.URL,
		MCPPath:         req.MCPPath,
		Command:         req.Command,
		Args:            req.Args,
		Headers:         req.Headers,
		Enabled:         req.Enabled,
		Active:          true,
		CreatedBy:       req.CreatedBy,
		CreatedOn:       now,
		ModifiedBy:      req.CreatedBy,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	m, err := docToMap(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialise MCP server doc: %w", err)
	}

	ctx := context.Background()
	if _, err := s.docDB.InsertOne(ctx, models.CollectionMCPServers, m); err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Created MCP server '%s' (id=%s)", req.Name, doc.ID))
	return models.MCPServerDocToModel(doc), nil
}

func (s *MCPServerService) GetMCPServer(id string) (*models.MCPServer, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	filter := map[string]interface{}{"_id": id, "active": true}
	result, err := s.docDB.FindOne(ctx, models.CollectionMCPServers, filter)
	if err != nil {
		return nil, err
	}
	var doc models.MCPServerDoc
	if err := mapToDoc(result, &doc); err != nil {
		return nil, err
	}
	return models.MCPServerDocToModel(&doc), nil
}

func (s *MCPServerService) ListMCPServers() ([]models.MCPServer, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	filter := map[string]interface{}{"active": true}
	opts := &documents.FindOptions{Sort: map[string]int{"name": 1}}
	results, err := s.docDB.FindMany(ctx, models.CollectionMCPServers, filter, opts)
	if err != nil {
		return nil, err
	}

	servers := make([]models.MCPServer, 0, len(results))
	for _, r := range results {
		var doc models.MCPServerDoc
		if err := mapToDoc(r, &doc); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to decode MCP server: %v", err))
			continue
		}
		servers = append(servers, *models.MCPServerDocToModel(&doc))
	}
	return servers, nil
}

func (s *MCPServerService) UpdateMCPServer(id string, req UpdateMCPServerRequest, modifiedBy string) (*models.MCPServer, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	setFields := map[string]interface{}{
		"modifiedby": modifiedBy,
		"modifiedon": time.Now().Format(time.RFC3339),
	}
	if req.Name != nil {
		setFields["name"] = *req.Name
	}
	if req.Description != nil {
		setFields["description"] = *req.Description
	}
	if req.TransportType != nil {
		setFields["transport_type"] = *req.TransportType
	}
	if req.URL != nil {
		setFields["url"] = *req.URL
	}
	if req.MCPPath != nil {
		setFields["mcp_path"] = *req.MCPPath
	}
	if req.Command != nil {
		setFields["command"] = *req.Command
	}
	if req.Args != nil {
		setFields["args"] = req.Args
	}
	if req.Headers != nil {
		setFields["headers"] = req.Headers
	}
	if req.Enabled != nil {
		setFields["enabled"] = *req.Enabled
	}

	ctx := context.Background()
	update := map[string]interface{}{
		"$set": setFields,
		"$inc": map[string]interface{}{"rowversionstamp": 1},
	}
	if err := s.docDB.UpdateOne(ctx, models.CollectionMCPServers,
		map[string]interface{}{"_id": id}, update); err != nil {
		return nil, err
	}
	return s.GetMCPServer(id)
}

func (s *MCPServerService) DeleteMCPServer(id, deletedBy string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	ctx := context.Background()
	filter := map[string]interface{}{"_id": id}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"active":     false,
			"modifiedby": deletedBy,
			"modifiedon": time.Now().Format(time.RFC3339),
		},
	}
	return s.docDB.UpdateOne(ctx, models.CollectionMCPServers, filter, update)
}

// ─── JSON-RPC 2.0 MCP Protocol Implementation ─────────────────────────────

// idCounter provides monotonically increasing JSON-RPC request IDs.
var idCounter atomic.Int64

func nextRPCID() int64 {
	return idCounter.Add(1)
}

// jsonrpcRequest is the standard JSON-RPC 2.0 request envelope.
type jsonrpcRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int64       `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// jsonrpcResponse is the standard JSON-RPC 2.0 response envelope.
type jsonrpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

// jsonrpcError represents a JSON-RPC 2.0 error object.
type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// mcpEndpoint returns the full MCP JSON-RPC endpoint URL for a server.
func mcpEndpoint(server models.MCPServer) string {
	path := server.MCPPath
	if path == "" {
		path = "/mcp"
	}
	return server.URL + path
}

// doJSONRPC sends a JSON-RPC 2.0 request to an HTTP MCP server and returns the raw result.
func doJSONRPC(ctx context.Context, server models.MCPServer, method string, params interface{}, sessionID string) (json.RawMessage, error) {
	rpc := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      nextRPCID(),
		Method:  method,
		Params:  params,
	}
	body, err := json.Marshal(rpc)
	if err != nil {
		return nil, fmt.Errorf("MCP marshal request: %w", err)
	}

	endpoint := mcpEndpoint(server)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("MCP build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	if sessionID != "" {
		req.Header.Set("Mcp-Session-Id", sessionID)
	}
	for k, v := range server.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("MCP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("MCP HTTP %d: %s", resp.StatusCode, string(b))
	}

	var rpcResp jsonrpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("MCP decode response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("MCP error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

// InitializeSession sends an MCP initialize request and returns the session ID.
func (s *MCPServerService) InitializeSession(ctx context.Context, server models.MCPServer) (string, error) {
	if server.TransportType != "http" || server.URL == "" {
		return "", nil
	}

	params := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]interface{}{},
		"clientInfo": map[string]interface{}{
			"name":    "iac-agent",
			"version": "1.0.0",
		},
	}

	rpc := jsonrpcRequest{
		JSONRPC: "2.0",
		ID:      nextRPCID(),
		Method:  "initialize",
		Params:  params,
	}
	body, err := json.Marshal(rpc)
	if err != nil {
		return "", fmt.Errorf("MCP marshal initialize: %w", err)
	}

	endpoint := mcpEndpoint(server)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("MCP build initialize request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range server.Headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("MCP initialize request: %w", err)
	}
	defer resp.Body.Close()

	sessionID := resp.Header.Get("Mcp-Session-Id")

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("MCP initialize HTTP %d: %s", resp.StatusCode, string(b))
	}

	io.ReadAll(resp.Body)
	return sessionID, nil
}

// ─── Tool Discovery ────────────────────────────────────────────────────────

type mcpToolsListResult struct {
	Tools      []mcpToolEntry `json:"tools"`
	NextCursor string         `json:"nextCursor,omitempty"`
}

type mcpToolEntry struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// FetchTools calls the MCP server via JSON-RPC 2.0 "tools/list" and returns ToolDefinitions.
func (s *MCPServerService) FetchTools(server models.MCPServer) ([]ToolDefinition, error) {
	if server.TransportType != "http" || server.URL == "" {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sessionID, _ := s.InitializeSession(ctx, server)

	result, err := doJSONRPC(ctx, server, "tools/list", map[string]interface{}{}, sessionID)
	if err != nil {
		return nil, fmt.Errorf("MCP FetchTools: %w", err)
	}

	var listResult mcpToolsListResult
	if err := json.Unmarshal(result, &listResult); err != nil {
		return nil, fmt.Errorf("MCP FetchTools decode result: %w", err)
	}

	defs := make([]ToolDefinition, 0, len(listResult.Tools))
	for _, t := range listResult.Tools {
		var params ToolParameterSchema
		if len(t.InputSchema) > 0 {
			_ = json.Unmarshal(t.InputSchema, &params)
		}
		if params.Type == "" {
			params.Type = "object"
		}
		defs = append(defs, ToolDefinition{
			Type: "function",
			Function: ToolFunctionDef{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}
	return defs, nil
}

// ─── Tool Invocation ──────────────────────────────────────────────────────

type mcpCallToolResult struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	IsError bool `json:"isError"`
}

// CallTool calls a named tool on an HTTP MCP server via JSON-RPC 2.0.
func (s *MCPServerService) CallTool(ctx context.Context, server models.MCPServer, toolName, argsJSON string) (string, error) {
	if server.TransportType != "http" || server.URL == "" {
		return "", fmt.Errorf("MCP CallTool: server %s does not support http transport", server.Name)
	}

	var arguments interface{}
	if argsJSON != "" && argsJSON != "null" {
		if err := json.Unmarshal([]byte(argsJSON), &arguments); err != nil {
			arguments = map[string]interface{}{}
		}
	} else {
		arguments = map[string]interface{}{}
	}

	params := map[string]interface{}{
		"name":      toolName,
		"arguments": arguments,
	}

	result, err := doJSONRPC(ctx, server, "tools/call", params, "")
	if err != nil {
		return "", fmt.Errorf("MCP CallTool: %w", err)
	}

	var callResult mcpCallToolResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		return "", fmt.Errorf("MCP CallTool decode result: %w", err)
	}
	if callResult.IsError {
		if len(callResult.Content) > 0 {
			return "", fmt.Errorf("MCP tool error: %s", callResult.Content[0].Text)
		}
		return "", fmt.Errorf("MCP tool returned error")
	}

	out := ""
	for _, c := range callResult.Content {
		if c.Type == "text" {
			out += c.Text
		}
	}
	return out, nil
}

// ─── Hub-level convenience: call any tool by server URL + path ────────────

// CallToolDirect invokes a tool on an arbitrary MCP server without a catalog entry.
func CallToolDirect(ctx context.Context, serverURL, mcpPath, toolName string, headers map[string]string, argsJSON string) (string, error) {
	if mcpPath == "" {
		mcpPath = "/mcp"
	}
	srv := models.MCPServer{
		URL:           serverURL,
		MCPPath:       mcpPath,
		TransportType: "http",
		Headers:       models.JSONBMap(headers),
	}
	svc := &MCPServerService{iLog: logger.Log{ModuleName: logger.Framework, User: "System", ControllerName: "MCPServerService"}}
	return svc.CallTool(ctx, srv, toolName, argsJSON)
}
