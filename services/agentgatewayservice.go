package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
)

// AgentGatewayService manages agent gateways and their A2A task lifecycle.
// All data is stored in MongoDB (agent_gateways + gateway_tasks collections).
type AgentGatewayService struct {
	docDB    documents.DocumentDB
	iLog     logger.Log
	runAgent func(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error)
}

var globalAgentGatewayService *AgentGatewayService

func GetGlobalAgentGatewayService() *AgentGatewayService { return globalAgentGatewayService }
func SetGlobalAgentGatewayService(s *AgentGatewayService) { globalAgentGatewayService = s }

// NewAgentGatewayService creates the service and initialises MongoDB indexes.
func NewAgentGatewayService(docDB documents.DocumentDB) *AgentGatewayService {
	svc := &AgentGatewayService{
		docDB: docDB,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentGatewayService",
		},
	}
	if docDB != nil {
		svc.initIndexes()
	}
	return svc
}

// SetRunAgentFunc injects the agent execution callback (avoids import cycle).
// Called from contollershandlers.go after AgentRunnerService is initialized.
func (s *AgentGatewayService) SetRunAgentFunc(fn func(ctx context.Context, agentID, agentName, prompt string, timeoutSecs int) (string, error)) {
	s.runAgent = fn
}

func (s *AgentGatewayService) initIndexes() {
	ctx := context.Background()
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentGateways,
		map[string]int{"name": 1},
		&documents.IndexOptions{Name: "idx_gateway_name", Unique: true})
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentGateways,
		map[string]int{"enabled": 1, "active": 1},
		&documents.IndexOptions{Name: "idx_gateway_enabled"})
	_ = s.docDB.CreateIndex(ctx, models.CollectionGatewayTasks,
		map[string]int{"gateway_id": 1},
		&documents.IndexOptions{Name: "idx_task_gateway"})
}

func (s *AgentGatewayService) ensureDocDB() error {
	if s.docDB == nil {
		return fmt.Errorf("document database not initialized for AgentGatewayService — ensure MongoDB is configured")
	}
	return nil
}

// ─── CRUD ─────────────────────────────────────────────────────────────────────

// Create inserts a new agent gateway document.
func (s *AgentGatewayService) Create(doc *models.AgentGatewayDoc) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	doc.CreatedOn = now
	doc.ModifiedOn = now
	doc.Active = true
	if doc.AgentIDs == nil {
		doc.AgentIDs = []string{}
	}
	if doc.MCPServerIDs == nil {
		doc.MCPServerIDs = []string{}
	}
	if doc.OutputConfig == nil {
		doc.OutputConfig = map[string]string{}
	}
	_, err := s.docDB.InsertOne(context.Background(), models.CollectionAgentGateways, doc)
	return err
}

// GetAll returns all active agent gateways.
func (s *AgentGatewayService) GetAll() ([]*models.AgentGatewayDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	filter := map[string]interface{}{"active": true}
	opts := &documents.FindOptions{
		Sort: map[string]int{"name": 1},
	}
	rows, err := s.docDB.FindMany(context.Background(), models.CollectionAgentGateways, filter, opts)
	if err != nil {
		return nil, err
	}
	result := make([]*models.AgentGatewayDoc, 0, len(rows))
	for _, row := range rows {
		doc := mapToGatewayDoc(row)
		result = append(result, doc)
	}
	return result, nil
}

// GetByID returns a single gateway by its ID.
func (s *AgentGatewayService) GetByID(id string) (*models.AgentGatewayDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	filter := map[string]interface{}{"_id": id}
	row, err := s.docDB.FindOne(context.Background(), models.CollectionAgentGateways, filter)
	if err != nil {
		return nil, err
	}
	return mapToGatewayDoc(row), nil
}

// Update replaces mutable fields on an existing gateway.
func (s *AgentGatewayService) Update(id string, doc *models.AgentGatewayDoc) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	doc.ModifiedOn = time.Now().UTC().Format(time.RFC3339)
	if doc.AgentIDs == nil {
		doc.AgentIDs = []string{}
	}
	if doc.MCPServerIDs == nil {
		doc.MCPServerIDs = []string{}
	}
	if doc.OutputConfig == nil {
		doc.OutputConfig = map[string]string{}
	}

	// Convert output_config to map[string]interface{} for safe BSON serialization.
	outputConfigI := make(map[string]interface{}, len(doc.OutputConfig))
	for k, v := range doc.OutputConfig {
		outputConfigI[k] = v
	}

	fields := map[string]interface{}{
		"name":               doc.Name,
		"description":        doc.Description,
		"workflow_id":        doc.WorkflowID,
		"agent_ids":          doc.AgentIDs,
		"mcp_server_ids":     doc.MCPServerIDs,
		"hub_id":             doc.HubID,
		"hub_route_id":       doc.HubRouteID,
		"output_type":        doc.OutputType,
		"output_config":      outputConfigI,
		"coordinator_mode":   doc.CoordinatorMode,
		"coordinator_prompt": doc.CoordinatorPrompt,
		"enabled":            doc.Enabled,
		"modifiedby":         doc.ModifiedBy,
		"modifiedon":         doc.ModifiedOn,
		"agent_card": map[string]interface{}{
			"display_name": doc.AgentCard.DisplayName,
			"description":  doc.AgentCard.Description,
			"provider":     doc.AgentCard.Provider,
			"provider_url": doc.AgentCard.ProviderURL,
			"skills":       doc.AgentCard.Skills,
		},
	}
	filter := map[string]interface{}{"_id": id}
	update := map[string]interface{}{"$set": fields}
	return s.docDB.UpdateOne(context.Background(), models.CollectionAgentGateways, filter, update)
}

// Delete soft-deletes a gateway by setting active=false.
func (s *AgentGatewayService) Delete(id string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	filter := map[string]interface{}{"_id": id}
	update := map[string]interface{}{"$set": map[string]interface{}{
		"active":      false,
		"enabled":     false,
		"modifiedon":  time.Now().UTC().Format(time.RFC3339),
	}}
	return s.docDB.UpdateOne(context.Background(), models.CollectionAgentGateways, filter, update)
}

// ─── A2A Task Management ──────────────────────────────────────────────────────

// CreateTask creates a new gateway task document with status "submitted".
func (s *AgentGatewayService) CreateTask(ctx context.Context, gatewayID, input string) (*models.GatewayTaskDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	task := &models.GatewayTaskDoc{
		ID:        uuid.New().String(),
		GatewayID: gatewayID,
		Status:    "submitted",
		Input:     input,
		CreatedOn: now,
		UpdatedOn: now,
	}
	_, err := s.docDB.InsertOne(ctx, models.CollectionGatewayTasks, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// GetTask retrieves a gateway task by gatewayID + taskID.
func (s *AgentGatewayService) GetTask(ctx context.Context, gatewayID, taskID string) (*models.GatewayTaskDoc, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	filter := map[string]interface{}{"_id": taskID, "gateway_id": gatewayID}
	row, err := s.docDB.FindOne(ctx, models.CollectionGatewayTasks, filter)
	if err != nil {
		return nil, err
	}
	return mapToTaskDoc(row), nil
}

// CancelTask transitions a task to "canceled" if it is still submitted/working.
func (s *AgentGatewayService) CancelTask(ctx context.Context, gatewayID, taskID string) error {
	if err := s.ensureDocDB(); err != nil {
		return err
	}
	filter := map[string]interface{}{"_id": taskID, "gateway_id": gatewayID}
	update := map[string]interface{}{"$set": map[string]interface{}{
		"status":    "canceled",
		"updatedon": time.Now().UTC().Format(time.RFC3339),
	}}
	return s.docDB.UpdateOne(ctx, models.CollectionGatewayTasks, filter, update)
}

func (s *AgentGatewayService) updateTaskStatus(ctx context.Context, taskID, status, output, errMsg string) error {
	filter := map[string]interface{}{"_id": taskID}
	fields := map[string]interface{}{
		"status":    status,
		"updatedon": time.Now().UTC().Format(time.RFC3339),
	}
	if output != "" {
		fields["output"] = output
	}
	if errMsg != "" {
		fields["error"] = errMsg
	}
	update := map[string]interface{}{"$set": fields}
	return s.docDB.UpdateOne(ctx, models.CollectionGatewayTasks, filter, update)
}

// ─── Execution ────────────────────────────────────────────────────────────────

// ExecuteAsync creates a GatewayTask, starts a goroutine, and returns the taskID immediately.
func (s *AgentGatewayService) ExecuteAsync(ctx context.Context, gatewayID, input string) (string, error) {
	task, err := s.CreateTask(ctx, gatewayID, input)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	go func() {
		bgCtx := context.Background()
		if updateErr := s.updateTaskStatus(bgCtx, task.ID, "working", "", ""); updateErr != nil {
			s.iLog.Error(fmt.Sprintf("updateTaskStatus(working) error: %v", updateErr))
		}

		output, execErr := s.ExecuteSync(bgCtx, gatewayID, input)
		if execErr != nil {
			_ = s.updateTaskStatus(bgCtx, task.ID, "failed", "", execErr.Error())
			return
		}
		_ = s.updateTaskStatus(bgCtx, task.ID, "completed", output, "")
	}()

	return task.ID, nil
}

// ExecuteSync runs the gateway's workflow synchronously and returns the final response.
func (s *AgentGatewayService) ExecuteSync(ctx context.Context, gatewayID, input string) (string, error) {
	gateway, err := s.GetByID(gatewayID)
	if err != nil {
		return "", fmt.Errorf("gateway %s not found: %w", gatewayID, err)
	}
	if !gateway.Enabled {
		return "", fmt.Errorf("gateway %s is disabled", gatewayID)
	}
	return s.executeWorkflow(ctx, gateway, input)
}

// executeWorkflow traverses the workflow and invokes aiagent nodes sequentially.
// If no WorkflowID is set the first agent in AgentIDs is run directly.
// When CoordinatorMode is enabled the coordinator prompt is injected into the
// execution context so the runner prepends it to the agent system prompt.
func (s *AgentGatewayService) executeWorkflow(ctx context.Context, gateway *models.AgentGatewayDoc, input string) (string, error) {
	if s.runAgent == nil {
		return "", fmt.Errorf("runAgent callback not set on AgentGatewayService")
	}

	// Inject coordinator prompt when coordinator mode is on.
	if gateway.CoordinatorMode {
		coordinatorPrompt := gateway.CoordinatorPrompt
		if coordinatorPrompt == "" {
			coordinatorPrompt = models.DefaultGatewayCoordinatorPrompt
		}
		// Append a list of available agent IDs so the coordinator can reference them.
		if len(gateway.AgentIDs) > 1 {
			coordinatorPrompt += fmt.Sprintf("\n\nAvailable specialist agents (use send_to_agent with these IDs): %s",
				strings.Join(gateway.AgentIDs[1:], ", "))
		}
		ctx = context.WithValue(ctx, models.AgentCoordinatorPromptKey{}, coordinatorPrompt)
	}

	// No workflow → run first agent directly
	if gateway.WorkflowID == "" {
		if len(gateway.AgentIDs) == 0 {
			return "", fmt.Errorf("gateway %s has no agents configured", gateway.ID)
		}
		return s.runAgent(ctx, gateway.AgentIDs[0], "", input, 300)
	}

	// Fetch workflow document from MongoDB
	if err := s.ensureDocDB(); err != nil {
		return "", err
	}
	filter := map[string]interface{}{"_id": gateway.WorkflowID}
	wfRow, err := s.docDB.FindOne(ctx, "workflows", filter)
	if err != nil {
		return "", fmt.Errorf("failed to fetch workflow %s: %w", gateway.WorkflowID, err)
	}

	// Parse nodes + edges from the workflow JSON
	nodesRaw, _ := wfRow["nodes"]
	edgesRaw, _ := wfRow["edges"]

	nodesJSON, _ := json.Marshal(nodesRaw)
	edgesJSON, _ := json.Marshal(edgesRaw)

	var nodes []workflowNode
	var edges []workflowEdge
	if err := json.Unmarshal(nodesJSON, &nodes); err != nil {
		return "", fmt.Errorf("failed to parse workflow nodes: %w", err)
	}
	if err := json.Unmarshal(edgesJSON, &edges); err != nil {
		return "", fmt.Errorf("failed to parse workflow edges: %w", err)
	}

	// Build adjacency list: nodeID → []targetNodeID
	adj := make(map[string][]string)
	for _, e := range edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
	}

	// Find start node
	startID := ""
	for _, n := range nodes {
		if strings.EqualFold(n.Type, "start") || strings.EqualFold(n.ID, "start") {
			startID = n.ID
			break
		}
	}
	if startID == "" && len(nodes) > 0 {
		startID = nodes[0].ID
	}

	// Index nodes by ID
	nodeMap := make(map[string]workflowNode, len(nodes))
	for _, n := range nodes {
		nodeMap[n.ID] = n
	}

	// DFS traversal — execute aiagent/aitask nodes in order
	output := input
	visited := make(map[string]bool)
	queue := []string{startID}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if visited[cur] {
			continue
		}
		visited[cur] = true

		node := nodeMap[cur]
		nodeType := strings.ToLower(node.Type)
		if nodeType == "aiagent" || nodeType == "aitask" {
			prompt := node.Data.UserPrompt
			if prompt == "" {
				prompt = output
			} else {
				prompt = strings.ReplaceAll(prompt, "{{previousOutput}}", output)
			}
			agentID := node.Data.AgentID
			agentName := node.Data.AgentName
			result, err := s.runAgent(ctx, agentID, agentName, prompt, 300)
			if err != nil {
				s.iLog.Error(fmt.Sprintf("executeWorkflow: agent node %s error: %v", cur, err))
				return "", fmt.Errorf("agent node %s failed: %w", cur, err)
			}
			output = result
		}

		for _, next := range adj[cur] {
			if !visited[next] {
				queue = append(queue, next)
			}
		}
	}

	return output, nil
}

// ─── Workflow types ───────────────────────────────────────────────────────────

type workflowNode struct {
	ID   string           `json:"id"`
	Type string           `json:"type"`
	Data workflowNodeData `json:"data"`
}

type workflowNodeData struct {
	AgentID    string `json:"agentId"`
	AgentName  string `json:"agentName"`
	UserPrompt string `json:"userPrompt"`
}

type workflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// ─── Map helpers ─────────────────────────────────────────────────────────────

func mapToGatewayDoc(row map[string]interface{}) *models.AgentGatewayDoc {
	doc := &models.AgentGatewayDoc{}
	if v, ok := row["_id"].(string); ok {
		doc.ID = v
	}
	if v, ok := row["name"].(string); ok {
		doc.Name = v
	}
	if v, ok := row["description"].(string); ok {
		doc.Description = v
	}
	if v, ok := row["workflow_id"].(string); ok {
		doc.WorkflowID = v
	}
	if v, ok := row["hub_id"].(string); ok {
		doc.HubID = v
	}
	if v, ok := row["hub_route_id"].(string); ok {
		doc.HubRouteID = v
	}
	if v, ok := row["output_type"].(string); ok {
		doc.OutputType = v
	}
	if v, ok := row["enabled"].(bool); ok {
		doc.Enabled = v
	}
	if v, ok := row["active"].(bool); ok {
		doc.Active = v
	}
	if v, ok := row["createdby"].(string); ok {
		doc.CreatedBy = v
	}
	if v, ok := row["createdon"].(string); ok {
		doc.CreatedOn = v
	}
	if v, ok := row["modifiedby"].(string); ok {
		doc.ModifiedBy = v
	}
	if v, ok := row["modifiedon"].(string); ok {
		doc.ModifiedOn = v
	}
	// agent_ids
	if v, ok := row["agent_ids"]; ok {
		doc.AgentIDs = toStringSlice(v)
	}
	// mcp_server_ids
	if v, ok := row["mcp_server_ids"]; ok {
		doc.MCPServerIDs = toStringSlice(v)
	}
	// output_config
	if v, ok := row["output_config"]; ok {
		doc.OutputConfig = toStringMap(v)
	}
	// coordinator
	if v, ok := row["coordinator_mode"].(bool); ok {
		doc.CoordinatorMode = v
	}
	if v, ok := row["coordinator_prompt"].(string); ok {
		doc.CoordinatorPrompt = v
	}
	// agent_card
	if v, ok := row["agent_card"]; ok {
		if cardMap, ok := v.(map[string]interface{}); ok {
			doc.AgentCard = mapToAgentCard(cardMap)
		}
	}
	return doc
}

func mapToTaskDoc(row map[string]interface{}) *models.GatewayTaskDoc {
	doc := &models.GatewayTaskDoc{}
	if v, ok := row["_id"].(string); ok {
		doc.ID = v
	}
	if v, ok := row["gateway_id"].(string); ok {
		doc.GatewayID = v
	}
	if v, ok := row["status"].(string); ok {
		doc.Status = v
	}
	if v, ok := row["input"].(string); ok {
		doc.Input = v
	}
	if v, ok := row["output"].(string); ok {
		doc.Output = v
	}
	if v, ok := row["error"].(string); ok {
		doc.Error = v
	}
	if v, ok := row["createdon"].(string); ok {
		doc.CreatedOn = v
	}
	if v, ok := row["updatedon"].(string); ok {
		doc.UpdatedOn = v
	}
	return doc
}

func mapToAgentCard(m map[string]interface{}) models.AgentCardConfig {
	card := models.AgentCardConfig{}
	if v, ok := m["display_name"].(string); ok {
		card.DisplayName = v
	}
	if v, ok := m["description"].(string); ok {
		card.Description = v
	}
	if v, ok := m["provider"].(string); ok {
		card.Provider = v
	}
	if v, ok := m["provider_url"].(string); ok {
		card.ProviderURL = v
	}
	if v, ok := m["skills"]; ok {
		if skills, ok := v.([]interface{}); ok {
			for _, s := range skills {
				if sm, ok := s.(map[string]interface{}); ok {
					skill := models.AgentCardSkill{}
					if id, ok := sm["id"].(string); ok {
						skill.ID = id
					}
					if name, ok := sm["name"].(string); ok {
						skill.Name = name
					}
					if desc, ok := sm["description"].(string); ok {
						skill.Description = desc
					}
					card.Skills = append(card.Skills, skill)
				}
			}
		}
	}
	return card
}

func toStringSlice(v interface{}) []string {
	if v == nil {
		return []string{}
	}
	if arr, ok := v.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{}
}

func toStringMap(v interface{}) map[string]string {
	result := map[string]string{}
	if v == nil {
		return result
	}
	if m, ok := v.(map[string]interface{}); ok {
		for k, val := range m {
			if s, ok := val.(string); ok {
				result[k] = s
			}
		}
	}
	return result
}
