package models

// MongoDB collection name constants for agent gateway entities.
const (
	CollectionAgentGateways = "agent_gateways"
	CollectionGatewayTasks  = "gateway_tasks"
)

// AgentGatewayDoc is the MongoDB document representation of an Agent Gateway.
// An Agent Gateway packages the agent ReAct execution behind a standard A2A
// protocol endpoint and can be triggered by IntegrationHub routes.
type AgentGatewayDoc struct {
	ID                string            `json:"_id"                bson:"_id"`
	Name              string            `json:"name"               bson:"name"`
	Description       string            `json:"description"        bson:"description"`
	WorkflowID        string            `json:"workflow_id"        bson:"workflow_id"`        // optional: existing workflow doc ID
	AgentIDs          []string          `json:"agent_ids"          bson:"agent_ids"`          // agents available to this gateway
	MCPServerIDs      []string          `json:"mcp_server_ids"     bson:"mcp_server_ids"`
	HubID             string            `json:"hub_id"             bson:"hub_id"`             // IntegrationHub trigger
	HubRouteID        string            `json:"hub_route_id"       bson:"hub_route_id"`       // specific route within hub
	AgentCard         AgentCardConfig   `json:"agent_card"         bson:"agent_card"`         // Google A2A metadata
	OutputType        string            `json:"output_type"        bson:"output_type"`        // "response"|"webhook"|"email"
	OutputConfig      map[string]string `json:"output_config"      bson:"output_config"`
	// CoordinatorMode enables multi-agent/multi-tool coordination at the gateway.
	// When true the first (coordinator) agent receives a prompt prefix that lists
	// all available agents and instructs it to route/delegate via send_to_agent.
	CoordinatorMode   bool              `json:"coordinator_mode"   bson:"coordinator_mode"`
	CoordinatorPrompt string            `json:"coordinator_prompt" bson:"coordinator_prompt"`
	Enabled           bool              `json:"enabled"            bson:"enabled"`
	Active            bool              `json:"active"             bson:"active"`
	CreatedBy         string            `json:"createdby"          bson:"createdby"`
	CreatedOn         string            `json:"createdon"          bson:"createdon"` // RFC 3339 string
	ModifiedBy        string            `json:"modifiedby"         bson:"modifiedby"`
	ModifiedOn        string            `json:"modifiedon"         bson:"modifiedon"`
}

// DefaultGatewayCoordinatorPrompt is injected as a system-prompt prefix when
// CoordinatorMode is true and no custom CoordinatorPrompt is provided.
// Unlike the channel variant this prompt is aware that multiple specialist
// agents may be registered and should be delegated to via send_to_agent.
const DefaultGatewayCoordinatorPrompt = `You are an intelligent gateway coordinator.

You have access to one or more specialist agents. When a request arrives:
1. Analyse the request to understand what type of task it represents.
2. Identify which agent(s) are best suited to handle it — use send_to_agent to delegate sub-tasks.
3. For complex requests that span multiple domains, delegate each part to the appropriate specialist and combine their responses.
4. Use your own tools directly for any parts that do not require specialist knowledge.
5. Synthesise all results into a single, coherent response.

Always prefer the most capable specialist for each sub-task over a generic answer.`

// AgentCardConfig holds the Google A2A agent card metadata for a gateway.
type AgentCardConfig struct {
	DisplayName string           `json:"display_name" bson:"display_name"`
	Description string           `json:"description"  bson:"description"`
	Provider    string           `json:"provider"     bson:"provider"`
	ProviderURL string           `json:"provider_url" bson:"provider_url"`
	Skills      []AgentCardSkill `json:"skills"       bson:"skills"`
}

// AgentCardSkill represents a capability advertised on the A2A agent card.
type AgentCardSkill struct {
	ID          string `json:"id"          bson:"id"`
	Name        string `json:"name"        bson:"name"`
	Description string `json:"description" bson:"description"`
}

// GatewayTaskDoc is the MongoDB document for a single A2A task execution.
// Status values follow the Google A2A spec: submitted | working | completed | failed | canceled.
type GatewayTaskDoc struct {
	ID        string `json:"_id"        bson:"_id"`
	GatewayID string `json:"gateway_id" bson:"gateway_id"`
	Status    string `json:"status"     bson:"status"`
	Input     string `json:"input"      bson:"input"`
	Output    string `json:"output"     bson:"output"`
	Error     string `json:"error"      bson:"error"`
	CreatedOn string `json:"createdon"  bson:"createdon"` // RFC 3339 string
	UpdatedOn string `json:"updatedon"  bson:"updatedon"` // RFC 3339 string
}
