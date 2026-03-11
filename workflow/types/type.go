package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type WorkFlow struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `json:"name"`
	UUID        string             `json:"uuid"`
	Version     string             `json:"version"`
	Description string             `json:"description"`
	ISDefault   bool               `json:"isDefault"`
	Type        string             `json:"type"`
	Nodes       []Node             `json:"nodes"`
	Links       []Link             `json:"links"`
}

type Node struct {
	Name          string                 `json:"name"`
	ID            string                 `json:"id"`
	Description   string                 `json:"description"`
	Type          string                 `json:"type"`
	Page          string                 `json:"page"`
	TranCode      string                 `json:"trancode"`
	Roles         []string               `json:"roles"`
	Users         []string               `json:"users"`
	Roleids       []int64                `json:"roleids"`
	Userids       []int64                `json:"userids"`
	PreCondition  map[string]interface{} `json:"precondition"`
	PostCondition map[string]interface{} `json:"postcondition"`
	ProcessData   map[string]interface{} `json:"processdata"`
	RoutingTables []RoutingTable         `json:"routingtables"`
	AIConfig        *AITaskConfig          `json:"aiconfig,omitempty"`
	SubflowId       string                 `json:"subflowId,omitempty"` // ID/UUID of the workflow to execute as subflow
	AIAgentConfig   *AIAgentConfig         `json:"aiAgentConfig,omitempty"`   // AI Agent/Skills configuration
	OutboundConfig  *OutboundConfig        `json:"outboundConfig,omitempty"`  // Outbound hub call configuration
}

// AIAgentConfig represents the configuration for an AI Agent/Skills call
type AIAgentConfig struct {
	AgentName    string   `json:"agentName,omitempty"`     // Name of the AI agent to call
	Skills       []string `json:"skills,omitempty"`        // Skills to use
	AIModel      string   `json:"aiModel,omitempty"`       // AI model from aiconfig.json
	AIVendor     string   `json:"aiVendor,omitempty"`      // AI vendor override
	SystemPrompt string   `json:"systemPrompt,omitempty"`  // System prompt for the agent
	UserPrompt   string   `json:"userPrompt,omitempty"`    // User prompt template
	Temperature  float64  `json:"temperature,omitempty"`   // Temperature setting
	MaxTokens    int      `json:"maxTokens,omitempty"`     // Max tokens
	Timeout      int      `json:"timeout,omitempty"`       // Timeout in seconds
}

// OutboundConfig holds the Integration Hub outbound endpoint configuration for a workflow node.
type OutboundConfig struct {
	HubID             string `json:"hubId,omitempty"`
	HubName           string `json:"hubName,omitempty"`
	ProtocolGroupID   string `json:"protocolGroupId,omitempty"`
	ProtocolGroupName string `json:"protocolGroupName,omitempty"`
	EndpointID        string `json:"endpointId,omitempty"`
	EndpointName      string `json:"endpointName,omitempty"`
	OutboundURL       string `json:"outboundUrl,omitempty"`  // Resolved URL cached at design time
	Method            string `json:"method,omitempty"`       // HTTP method, default POST
	ContentType       string `json:"contentType,omitempty"`  // default application/json
	PayloadDataKey    string `json:"payloadDataKey,omitempty"`  // ProcessData key to use as body; empty = all ProcessData
	ResponseDataKey   string `json:"responseDataKey,omitempty"` // ProcessData key to store response; default "outbound_response"
	DynamicUrl        bool   `json:"dynamicUrl,omitempty"`      // If true, read URL from ProcessData["outbound_url"]
}

type AITaskConfig struct {
	UseCaseName   string              `json:"useCaseName,omitempty"`
	AIVendor      string              `json:"aiVendor,omitempty"`
	Agent         string              `json:"agent,omitempty"`
	Skills        []string            `json:"skills,omitempty"`
	MCPServers    []string            `json:"mcpServers,omitempty"`
	SystemPrompt  string              `json:"systemPrompt,omitempty"`
	UserPrompt    string              `json:"userPrompt,omitempty"`
	InputMapping  []AIInputMapping    `json:"inputMapping,omitempty"`
	OutputMapping []AIOutputMapping   `json:"outputMapping,omitempty"`
	Temperature   float64             `json:"temperature,omitempty"`
	MaxTokens     int                 `json:"maxTokens,omitempty"`
	Timeout       int                 `json:"timeout,omitempty"`
	RetryAttempts int                 `json:"retryAttempts,omitempty"`
}

type AIInputMapping struct {
	ID             string `json:"id"`
	InputName      string `json:"inputName"`
	PromptVariable string `json:"promptVariable"`
	Transformation string `json:"transformation,omitempty"`
}

type AIOutputMapping struct {
	ID             string `json:"id"`
	OutputName     string `json:"outputName"`
	AIResponsePath string `json:"aiResponsePath"`
	Transformation string `json:"transformation,omitempty"`
	DefaultValue   string `json:"defaultValue,omitempty"`
}

type Link struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Type   string `json:"type"`
	Label  string `json:"label"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type RoutingTable struct {
	Default  bool   `json:"default"`
	Sequence int    `json:"sequence"`
	Data     string `json:"data"`
	Value    string `json:"value"`
	Target   string `json:"target"`
}
