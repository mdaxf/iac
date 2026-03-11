package models

import "time"

// MongoDB collection name constants for agent runtime entities.
const (
	CollectionAgentSkills      = "agent_skills"
	CollectionAgentDefinitions = "agent_definitions"
	CollectionMCPServers       = "mcp_servers"
)

// AgentSkillDoc is the MongoDB document representation of an agent skill.
// The _id field holds the same UUID used in the legacy PostgreSQL table.
type AgentSkillDoc struct {
	ID              string    `json:"_id"             bson:"_id"`
	Name            string    `json:"name"            bson:"name"`
	DisplayName     string    `json:"display_name"    bson:"display_name"`
	Description     string    `json:"description"     bson:"description"`
	Category        string    `json:"category"        bson:"category"`
	Active          bool      `json:"active"          bson:"active"`
	CreatedBy       string    `json:"createdby"       bson:"createdby"`
	CreatedOn       time.Time `json:"createdon"       bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby"      bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon"      bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
	// Installed skill metadata (empty for built-in skills)
	SkillType   string `json:"skill_type,omitempty"   bson:"skill_type,omitempty"`
	SourceURL   string `json:"source_url,omitempty"   bson:"source_url,omitempty"`
	InstallPath string `json:"install_path,omitempty" bson:"install_path,omitempty"`
	Runtime     string `json:"runtime,omitempty"      bson:"runtime,omitempty"`
	EntryPoint  string `json:"entry_point,omitempty"  bson:"entry_point,omitempty"`
	ToolDefs    string `json:"tool_defs,omitempty"    bson:"tool_defs,omitempty"`
	Version     string `json:"version,omitempty"      bson:"version,omitempty"`
	// Content holds extended LLM-facing instructions/schema for this skill.
	// Appended to the tool description at run time when non-empty.
	Content     string `json:"content,omitempty"      bson:"content,omitempty"`
}

// AgentDefinitionDoc is the MongoDB document representation of an agent definition.
// SkillNames and MCPServerIDs are embedded arrays replacing the join tables
// agent_definition_skills and agent_mcp_servers.
type AgentDefinitionDoc struct {
	ID                 string            `json:"_id"                          bson:"_id"`
	Name               string            `json:"name"                         bson:"name"`
	Description        string            `json:"description"                  bson:"description"`
	Avatar             string            `json:"avatar"                       bson:"avatar"`
	ModelProvider      string            `json:"model_provider"               bson:"model_provider"`
	ModelName          string            `json:"model_name"                   bson:"model_name"`
	SystemPrompt       string            `json:"system_prompt"                bson:"system_prompt"`
	MaxIterations      int               `json:"max_iterations"               bson:"max_iterations"`
	Temperature        float64           `json:"temperature"                  bson:"temperature"`
	IsDefault          bool              `json:"is_default"                   bson:"is_default"`
	Enabled            bool              `json:"enabled"                      bson:"enabled"`
	RunInstance        string            `json:"run_instance"                 bson:"run_instance"`
	NotificationConfig map[string]string `json:"notification_config,omitempty" bson:"notification_config,omitempty"`
	// SkillNames stores tool names assigned to this agent (replaces join table).
	SkillNames   []string `json:"skill_names"    bson:"skill_names"`
	// MCPServerIDs stores MCP server UUIDs assigned to this agent (replaces join table).
	MCPServerIDs []string `json:"mcp_server_ids" bson:"mcp_server_ids"`
	Active          bool      `json:"active"          bson:"active"`
	ReferenceID     string    `json:"referenceid"     bson:"referenceid"`
	CreatedBy       string    `json:"createdby"       bson:"createdby"`
	CreatedOn       time.Time `json:"createdon"       bson:"createdon"`
	ModifiedBy      string    `json:"modifiedby"      bson:"modifiedby"`
	ModifiedOn      time.Time `json:"modifiedon"      bson:"modifiedon"`
	RowVersionStamp int       `json:"rowversionstamp" bson:"rowversionstamp"`
}

// MCPServerDoc is the MongoDB document representation of an MCP server.
type MCPServerDoc struct {
	ID              string            `json:"_id"             bson:"_id"`
	Name            string            `json:"name"            bson:"name"`
	Description     string            `json:"description"     bson:"description"`
	TransportType   string            `json:"transport_type"  bson:"transport_type"`
	URL             string            `json:"url"             bson:"url"`
	MCPPath         string            `json:"mcp_path"        bson:"mcp_path"`
	Command         string            `json:"command"         bson:"command"`
	Args            []string          `json:"args"            bson:"args"`
	Headers         map[string]string `json:"headers"         bson:"headers"`
	Enabled         bool              `json:"enabled"         bson:"enabled"`
	Active          bool              `json:"active"          bson:"active"`
	CreatedBy       string            `json:"createdby"       bson:"createdby"`
	CreatedOn       time.Time         `json:"createdon"       bson:"createdon"`
	ModifiedBy      string            `json:"modifiedby"      bson:"modifiedby"`
	ModifiedOn      time.Time         `json:"modifiedon"      bson:"modifiedon"`
	RowVersionStamp int               `json:"rowversionstamp" bson:"rowversionstamp"`
}

// ---- Converters: GORM Model → MongoDB Doc ----

// AgentSkillModelToDoc converts a GORM AgentSkill to an AgentSkillDoc for MongoDB storage.
func AgentSkillModelToDoc(s *AgentSkill) *AgentSkillDoc {
	doc := &AgentSkillDoc{
		ID:              s.ID,
		Name:            s.Name,
		DisplayName:     s.DisplayName,
		Description:     s.Description,
		Category:        s.Category,
		Active:          s.Active,
		CreatedBy:       s.CreatedBy,
		ModifiedBy:      s.ModifiedBy,
		RowVersionStamp: s.RowVersionStamp,
		Content:         s.Content,
	}
	if s.CreatedOn.Valid {
		doc.CreatedOn = s.CreatedOn.Time
	}
	if s.ModifiedOn.Valid {
		doc.ModifiedOn = s.ModifiedOn.Time
	}
	return doc
}

// MCPServerModelToDoc converts a GORM MCPServer to an MCPServerDoc for MongoDB storage.
func MCPServerModelToDoc(m *MCPServer) *MCPServerDoc {
	doc := &MCPServerDoc{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		TransportType:   m.TransportType,
		URL:             m.URL,
		MCPPath:         m.MCPPath,
		Command:         m.Command,
		Args:            []string(m.Args),
		Headers:         map[string]string(m.Headers),
		Enabled:         m.Enabled,
		Active:          m.Active,
		CreatedBy:       m.CreatedBy,
		ModifiedBy:      m.ModifiedBy,
		RowVersionStamp: m.RowVersionStamp,
	}
	if m.CreatedOn.Valid {
		doc.CreatedOn = m.CreatedOn.Time
	}
	if m.ModifiedOn.Valid {
		doc.ModifiedOn = m.ModifiedOn.Time
	}
	return doc
}

// AgentDefinitionModelToDoc converts a GORM AgentDefinition (with preloaded Skills and
// MCPServers) to an AgentDefinitionDoc for MongoDB storage. Skills and MCP server
// assignments are embedded as string arrays.
func AgentDefinitionModelToDoc(a *AgentDefinition) *AgentDefinitionDoc {
	doc := &AgentDefinitionDoc{
		ID:              a.ID,
		Name:            a.Name,
		Description:     a.Description,
		Avatar:          a.Avatar,
		ModelProvider:   a.ModelProvider,
		ModelName:       a.ModelName,
		SystemPrompt:    a.SystemPrompt,
		MaxIterations:   a.MaxIterations,
		Temperature:     a.Temperature,
		IsDefault:       a.IsDefault,
		Enabled:         a.Enabled,
		RunInstance:     a.RunInstance,
		Active:          a.Active,
		ReferenceID:     a.ReferenceID,
		CreatedBy:       a.CreatedBy,
		ModifiedBy:      a.ModifiedBy,
		RowVersionStamp: a.RowVersionStamp,
		SkillNames:      make([]string, 0, len(a.Skills)),
		MCPServerIDs:    make([]string, 0, len(a.MCPServers)),
	}
	if a.NotificationConfig != nil {
		doc.NotificationConfig = map[string]string(a.NotificationConfig)
	}
	if a.CreatedOn.Valid {
		doc.CreatedOn = a.CreatedOn.Time
	}
	if a.ModifiedOn.Valid {
		doc.ModifiedOn = a.ModifiedOn.Time
	}
	for _, s := range a.Skills {
		doc.SkillNames = append(doc.SkillNames, s.Name)
	}
	for _, m := range a.MCPServers {
		doc.MCPServerIDs = append(doc.MCPServerIDs, m.ID)
	}
	return doc
}

// ---- Converters: MongoDB Doc → GORM Model ----
// Note: Skills and MCPServers on AgentDefinition are NOT populated by these converters.
// The AgentDefinitionService hydrates them separately after resolving from MongoDB.

// AgentSkillDocToModel converts an AgentSkillDoc to an AgentSkill model.
func AgentSkillDocToModel(d *AgentSkillDoc) *AgentSkill {
	return &AgentSkill{
		ID:              d.ID,
		Name:            d.Name,
		DisplayName:     d.DisplayName,
		Description:     d.Description,
		Category:        d.Category,
		Active:          d.Active,
		CreatedBy:       d.CreatedBy,
		ModifiedBy:      d.ModifiedBy,
		RowVersionStamp: d.RowVersionStamp,
		SkillType:       d.SkillType,
		SourceURL:       d.SourceURL,
		InstallPath:     d.InstallPath,
		Runtime:         d.Runtime,
		EntryPoint:      d.EntryPoint,
		ToolDefs:        d.ToolDefs,
		Version:         d.Version,
		Content:         d.Content,
	}
}

// MCPServerDocToModel converts an MCPServerDoc to an MCPServer model.
func MCPServerDocToModel(d *MCPServerDoc) *MCPServer {
	return &MCPServer{
		ID:              d.ID,
		Name:            d.Name,
		Description:     d.Description,
		TransportType:   d.TransportType,
		URL:             d.URL,
		MCPPath:         d.MCPPath,
		Command:         d.Command,
		Args:            StringSlice(d.Args),
		Headers:         JSONBMap(d.Headers),
		Enabled:         d.Enabled,
		Active:          d.Active,
		CreatedBy:       d.CreatedBy,
		ModifiedBy:      d.ModifiedBy,
		RowVersionStamp: d.RowVersionStamp,
	}
}

// AgentDefinitionDocToModel converts an AgentDefinitionDoc to an AgentDefinition model.
// Skills and MCPServers are left empty; the caller must hydrate them.
func AgentDefinitionDocToModel(d *AgentDefinitionDoc) *AgentDefinition {
	return &AgentDefinition{
		ID:                 d.ID,
		Name:               d.Name,
		Description:        d.Description,
		Avatar:             d.Avatar,
		ModelProvider:      d.ModelProvider,
		ModelName:          d.ModelName,
		SystemPrompt:       d.SystemPrompt,
		MaxIterations:      d.MaxIterations,
		Temperature:        d.Temperature,
		IsDefault:          d.IsDefault,
		Enabled:            d.Enabled,
		RunInstance:        d.RunInstance,
		NotificationConfig: JSONBMap(d.NotificationConfig),
		Active:             d.Active,
		ReferenceID:        d.ReferenceID,
		CreatedBy:          d.CreatedBy,
		ModifiedBy:         d.ModifiedBy,
		RowVersionStamp:    d.RowVersionStamp,
		// Skills and MCPServers intentionally left nil; service hydrates them.
	}
}
