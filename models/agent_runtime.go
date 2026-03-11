package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringSlice is a JSONB-backed string slice for GORM
type StringSlice []string

func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = StringSlice{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, s)
}

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(s)
}

// JSONBMap is a JSONB-backed map[string]string for GORM
type JSONBMap map[string]string

func (m *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*m = JSONBMap{}
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("JSONBMap: expected []byte, got %T", value)
	}
	return json.Unmarshal(b, m)
}

func (m JSONBMap) Value() (driver.Value, error) {
	if m == nil {
		return json.Marshal(map[string]string{})
	}
	return json.Marshal(m)
}

// AgentSkill represents a single tool/capability available to agents.
// Skills are defined once in the catalog and can be assigned to many agents.
type AgentSkill struct {
	ID          string       `json:"id"          gorm:"column:id;primaryKey;type:varchar(36)"`
	Name        string       `json:"name"        gorm:"column:name;type:varchar(100);not null;uniqueIndex"`
	DisplayName string       `json:"displayname" gorm:"column:display_name;type:varchar(100);not null"`
	Description string       `json:"description" gorm:"column:description;type:text"`
	Category    string       `json:"category"    gorm:"column:category;type:varchar(50)"`
	Active      bool         `json:"active"      gorm:"column:active;not null;default:true"`
	CreatedBy   string       `json:"createdby"   gorm:"column:createdby;type:varchar(45)"`
	CreatedOn   sql.NullTime `json:"createdon"   gorm:"column:createdon;autoCreateTime"`
	ModifiedBy  string       `json:"modifiedby"  gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn  sql.NullTime `json:"modifiedon"  gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int      `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
	// Installed skill metadata (not stored in PostgreSQL; populated from MongoDB only)
	SkillType   string `json:"skill_type"   gorm:"-"` // "builtin" | "installed"
	SourceURL   string `json:"source_url"   gorm:"-"` // installation source URL
	InstallPath string `json:"install_path" gorm:"-"` // on-disk directory path
	Runtime     string `json:"runtime"      gorm:"-"` // "python" | "node" | "bash" | "none"
	EntryPoint  string `json:"entry_point"  gorm:"-"` // relative path to entry script
	ToolDefs    string `json:"tool_defs"    gorm:"-"` // JSON array of ToolDefinition
	Version     string `json:"version"      gorm:"-"` // skill version string
	// Content holds extended LLM-facing instructions/schema for this skill.
	// When non-empty it is appended to the tool description seen by the LLM,
	// making the skill the authoritative source of its own spec/usage guide.
	Content     string `json:"content"      gorm:"-"` // extended instructions (MongoDB only)
}

func (AgentSkill) TableName() string { return "agent_skills" }

// ToolNames returns the skill names suitable for tool registry lookup.
func SkillNames(skills []AgentSkill) []string {
	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	return names
}

// AgentDefinition represents a configured AI agent.
// Skills are managed via the agent_definition_skills join table.
type AgentDefinition struct {
	ID            string       `json:"id"            gorm:"column:id;primaryKey;type:varchar(36)"`
	Name          string       `json:"name"          gorm:"column:name;type:varchar(100);not null"`
	Description   string       `json:"description"   gorm:"column:description;type:text"`
	Avatar        string       `json:"avatar"        gorm:"column:avatar;type:varchar(500)"`
	ModelProvider string       `json:"modelprovider" gorm:"column:model_provider;type:varchar(50);not null;default:'openai'"`
	ModelName     string       `json:"modelname"     gorm:"column:model_name;type:varchar(100);not null;default:'gpt-4o'"`
	SystemPrompt  string       `json:"systemprompt"  gorm:"column:system_prompt;type:text;not null"`
	MaxIterations int          `json:"maxiterations" gorm:"column:max_iterations;not null;default:10"`
	Temperature   float64      `json:"temperature"   gorm:"column:temperature;type:decimal(3,2);not null;default:0.70"`
	IsDefault     bool         `json:"isdefault"     gorm:"column:is_default;not null;default:false"`
	Enabled       bool         `json:"enabled"       gorm:"column:enabled;not null;default:true"`
	RunInstance   string       `json:"runinstance"   gorm:"column:run_instance;type:varchar(50);not null;default:'app'"`

	// NotificationConfig stores per-agent notification credentials and settings.
	// Keys: telegram_bot_token, telegram_chat_id, slack_webhook_url, smtp_from,
	//       teams_webhook_url, discord_webhook_url.
	// Values here override system-level defaults from configuration.json but are
	// overridden by parameters supplied per-call.
	NotificationConfig JSONBMap `json:"notificationconfig" gorm:"column:notification_config;type:jsonb;default:'{}'"`

	// Skills assigned to this agent — populated at runtime by AgentDefinitionService from MongoDB.
	// GORM does not manage this field; it is hydrated via MongoDB skill_names array.
	Skills []AgentSkill `json:"skills" gorm:"-"`

	// MCPServers assigned to this agent — populated at runtime by AgentDefinitionService from MongoDB.
	// GORM does not manage this field; it is hydrated via MongoDB mcp_server_ids array.
	MCPServers []MCPServer `json:"mcpservers" gorm:"-"`

	// SkillNames holds the raw skill name strings stored in the MongoDB document.
	// Used internally during conversion between AgentDefinitionDoc and AgentDefinition.
	SkillNames []string `json:"skillnames,omitempty" gorm:"-"`

	// Standard IAC audit fields
	Active          bool         `json:"active"          gorm:"column:active;not null;default:true"`
	ReferenceID     string       `json:"referenceid"     gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby"       gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon"       gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby"      gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon"      gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
}

func (AgentDefinition) TableName() string { return "agent_definitions" }

// AgentSchedule represents a scheduled task for an agent
type AgentSchedule struct {
	ID                string        `json:"id"                gorm:"column:id;primaryKey;type:varchar(36)"`
	AgentDefinitionID string        `json:"agentdefinitionid" gorm:"column:agent_definition_id;type:varchar(36);not null"`
	Name              string        `json:"name"              gorm:"column:name;type:varchar(100);not null"`
	Description       string        `json:"description"       gorm:"column:description;type:text"`
	TaskPrompt        string        `json:"taskprompt"        gorm:"column:task_prompt;type:text;not null"`
	CronExpression    string        `json:"cronexpression"    gorm:"column:cron_expression;type:varchar(100)"`
	RunOnceAt         *sql.NullTime `json:"runonceat"         gorm:"column:run_once_at"`
	LastRunAt         *sql.NullTime `json:"lastrunat"         gorm:"column:last_run_at"`
	LastRunStatus     string        `json:"lastrunstatus"     gorm:"column:last_run_status;type:varchar(20)"`
	LastRunOutput     string        `json:"lastrunoutput"     gorm:"column:last_run_output;type:text"`
	JobID             string        `json:"jobid"             gorm:"column:job_id;type:varchar(36)"`
	Enabled           bool          `json:"enabled"           gorm:"column:enabled;not null;default:true"`

	// Standard IAC audit fields
	Active          bool         `json:"active"          gorm:"column:active;not null;default:true"`
	ReferenceID     string       `json:"referenceid"     gorm:"column:referenceid;type:varchar(36)"`
	CreatedBy       string       `json:"createdby"       gorm:"column:createdby;type:varchar(45)"`
	CreatedOn       sql.NullTime `json:"createdon"       gorm:"column:createdon;autoCreateTime"`
	ModifiedBy      string       `json:"modifiedby"      gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn      sql.NullTime `json:"modifiedon"      gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int          `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
}

func (AgentSchedule) TableName() string { return "agent_schedules" }

// MCPServer represents a Model Context Protocol server configuration.
type MCPServer struct {
	ID            string       `json:"id"              gorm:"column:id;primaryKey;type:varchar(36)"`
	Name          string       `json:"name"            gorm:"column:name;type:varchar(100);not null"`
	Description   string       `json:"description"     gorm:"column:description;type:text"`
	TransportType string       `json:"transporttype"   gorm:"column:transport_type;type:varchar(20);not null;default:'http'"`
	URL           string       `json:"url"             gorm:"column:url;type:varchar(500)"`
	// MCPPath is the JSON-RPC 2.0 endpoint path (e.g. "/mcp"). Default "/mcp".
	MCPPath       string       `json:"mcppath"         gorm:"column:mcp_path;type:varchar(200);not null;default:'/mcp'"`
	Command       string       `json:"command"         gorm:"column:command;type:varchar(500)"`
	Args          StringSlice  `json:"args"            gorm:"column:args;type:jsonb;default:'[]'"`
	Headers       JSONBMap     `json:"headers"         gorm:"column:headers;type:jsonb;default:'{}'"`
	Enabled       bool         `json:"enabled"         gorm:"column:enabled;not null;default:true"`
	Active        bool         `json:"active"          gorm:"column:active;not null;default:true"`
	CreatedBy     string       `json:"createdby"       gorm:"column:createdby;type:varchar(45)"`
	CreatedOn     sql.NullTime `json:"createdon"       gorm:"column:createdon;autoCreateTime"`
	ModifiedBy    string       `json:"modifiedby"      gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn    sql.NullTime `json:"modifiedon"      gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp int        `json:"rowversionstamp" gorm:"column:rowversionstamp;default:1"`
}

func (MCPServer) TableName() string { return "mcp_servers" }

// AgentMemory stores a single memory item for an agent, using a 3-layer system.
// L0 = lightweight index (title + brief), L1 = summary, L2 = full content.
// Priority: P0 = critical/permanent, P1 = normal, P2 = low/disposable.
// Retention: permanent | interval (auto-archive after N days) | temporary (expires_at).
type AgentMemory struct {
	ID                 string        `json:"id"                 gorm:"column:id;primaryKey;type:varchar(36)"`
	AgentDefinitionID  string        `json:"agentdefinitionid"  gorm:"column:agent_definition_id;type:varchar(36);not null"`
	Title              string        `json:"title"              gorm:"column:title;type:varchar(200);not null"`
	Summary            string        `json:"summary"            gorm:"column:summary;type:text"`
	Content            string        `json:"content"            gorm:"column:content;type:text"`
	ContentType        string        `json:"contenttype"        gorm:"column:content_type;type:varchar(50);not null;default:'text'"`
	Tags               StringSlice   `json:"tags"               gorm:"column:tags;type:jsonb;default:'[]'"`
	Layer              string        `json:"layer"              gorm:"column:layer;type:varchar(2);not null;default:'L1'"`
	Priority           string        `json:"priority"           gorm:"column:priority;type:varchar(2);not null;default:'P1'"`
	RetentionType      string        `json:"retentiontype"      gorm:"column:retention_type;type:varchar(20);not null;default:'permanent'"`
	RetentionDays      *int          `json:"retentiondays"      gorm:"column:retention_days"`
	ExpiresAt          *sql.NullTime `json:"expiresat"          gorm:"column:expires_at"`
	AccessCount        int           `json:"accesscount"        gorm:"column:access_count;not null;default:0"`
	LastAccessedOn     *sql.NullTime `json:"lastaccessedon"     gorm:"column:last_accessed_on"`
	Active             bool          `json:"active"             gorm:"column:active;not null;default:true"`
	Archived           bool          `json:"archived"           gorm:"column:archived;not null;default:false"`
	CreatedBy          string        `json:"createdby"          gorm:"column:createdby;type:varchar(45)"`
	CreatedOn          sql.NullTime  `json:"createdon"          gorm:"column:createdon;autoCreateTime"`
	ModifiedBy         string        `json:"modifiedby"         gorm:"column:modifiedby;type:varchar(45)"`
	ModifiedOn         sql.NullTime  `json:"modifiedon"         gorm:"column:modifiedon;autoUpdateTime"`
	RowVersionStamp    int           `json:"rowversionstamp"    gorm:"column:rowversionstamp;default:1"`
}

func (AgentMemory) TableName() string { return "agent_memory" }

// AgentJobPayload is the JSON payload stored in QueueJob.Payload for agent runner jobs
type AgentJobPayload struct {
	AgentDefinitionID string `json:"agentdefinitionid"`
	AgentScheduleID   string `json:"agentscheduleid,omitempty"`
	ConversationID    string `json:"conversationid"`
	TaskPrompt        string `json:"taskprompt"`
	RequestedBy       string `json:"requestedby"`
}
