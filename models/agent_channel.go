package models

// MongoDB collection name constants for agent channel entities.
const (
	CollectionAgentChannels       = "agent_channels"
	CollectionChannelUserMappings = "channel_user_mappings"
	CollectionAgentChannelSessions = "agent_channel_sessions"
)

// Channel type string constants
const (
	ChannelTypeTelegram      = "telegram"
	ChannelTypeSlack         = "slack"
	ChannelTypeDiscord       = "discord"
	ChannelTypeMSTeams       = "msteams"
	ChannelTypeGoogleChat    = "googlechat"
	ChannelTypeFeishu        = "feishu"
	ChannelTypeWhatsApp      = "whatsapp"
	ChannelTypeIMessage      = "imessage"
	ChannelTypeSignal        = "signal"
	ChannelTypeLine          = "line"
	ChannelTypeMattermost    = "mattermost"
	ChannelTypeMatrix        = "matrix"
	ChannelTypeIRC           = "irc"
	ChannelTypeZalo          = "zalo"
	ChannelTypeSynologyChat  = "synology-chat"
	ChannelTypeNextcloudTalk = "nextcloud-talk"
	ChannelTypeNostr         = "nostr"
	ChannelTypeWeChatWork    = "wechatwork"
)

// AgentChannelDoc is the MongoDB document for an agent-channel binding.
// One IAC agent can have multiple channel bindings.
type AgentChannelDoc struct {
	ID               string                     `json:"_id"               bson:"_id"`
	AgentID          string                     `json:"agent_id"          bson:"agent_id"`
	AgentName        string                     `json:"agent_name"        bson:"agent_name"`
	ChannelType      string                     `json:"channel_type"      bson:"channel_type"`
	DisplayName      string                     `json:"display_name"      bson:"display_name"`
	ChannelConfig    map[string]string          `json:"channel_config"    bson:"channel_config"`
	SecurityPolicy   AgentChannelSecurityPolicy `json:"security_policy"   bson:"security_policy"`
	WelcomeMessage   string                     `json:"welcome_message"   bson:"welcome_message"`
	StartMessage     string                     `json:"start_message"     bson:"start_message"`
	StopMessage      string                     `json:"stop_message"      bson:"stop_message"`
	// CoordinatorMode enables multi-tool/multi-agent coordination behaviour.
	// When true the bound agent receives a coordinator system-prompt prefix that
	// directs it to analyse the incoming message, select the right tools
	// (possibly more than one), execute them in sequence, and return an
	// aggregated answer. CoordinatorPrompt overrides the default prefix.
	CoordinatorMode   bool   `json:"coordinator_mode"   bson:"coordinator_mode"`
	CoordinatorPrompt string `json:"coordinator_prompt" bson:"coordinator_prompt"`
	Enabled          bool                       `json:"enabled"           bson:"enabled"`
	Active           bool                       `json:"active"            bson:"active"`
	CreatedBy        string                     `json:"createdby"         bson:"createdby"`
	CreatedOn        string                     `json:"createdon"         bson:"createdon"` // RFC 3339 string
	ModifiedBy       string                     `json:"modifiedby"        bson:"modifiedby"`
	ModifiedOn       string                     `json:"modifiedon"        bson:"modifiedon"`
}

// DefaultCoordinatorPrompt is injected as a system-prompt prefix when
// CoordinatorMode is true and no custom CoordinatorPrompt is set.
const DefaultCoordinatorPrompt = `You are an intelligent agent coordinator for this channel.

When a user sends a message:
1. Analyse the request carefully to understand what information or actions are needed.
2. Select the most appropriate tool(s) from your available toolset. For complex or multi-part questions, use several tools in sequence.
3. If a specialist agent is better suited for part of the work, delegate using send_to_agent.
4. Synthesise all results into a clear, concise answer before replying to the user.

Always prefer accuracy over speed. Use multiple tools when a single tool cannot fully address the request.`

// AgentChannelSecurityPolicy controls what channel users can do via this channel.
type AgentChannelSecurityPolicy struct {
	RequireUserAuth      bool     `json:"require_user_auth"      bson:"require_user_auth"`
	AllowedTools         []string `json:"allowed_tools"          bson:"allowed_tools"`
	BlockedTools         []string `json:"blocked_tools"          bson:"blocked_tools"`
	MaxIterations        int      `json:"max_iterations"         bson:"max_iterations"`
	MaxResponseLength    int      `json:"max_response_length"    bson:"max_response_length"`
	RateLimitPerHour     int      `json:"rate_limit_per_hour"    bson:"rate_limit_per_hour"`
	SystemPromptOverride string   `json:"system_prompt_override" bson:"system_prompt_override"`
}

// ChannelUserMappingDoc links external channel identities to IAC users.
type ChannelUserMappingDoc struct {
	ID              string `json:"_id"               bson:"_id"`
	AgentChannelID  string `json:"agent_channel_id"  bson:"agent_channel_id"`
	ChannelType     string `json:"channel_type"      bson:"channel_type"`
	ChannelUserID   string `json:"channel_user_id"   bson:"channel_user_id"`
	ChannelUsername string `json:"channel_username"  bson:"channel_username"`
	IACUserID       string `json:"iac_user_id"       bson:"iac_user_id"`
	IACUsername     string `json:"iac_username"      bson:"iac_username"`
	Authorized      bool   `json:"authorized"        bson:"authorized"`
	AuthorizedBy    string `json:"authorized_by"     bson:"authorized_by"`
	AuthorizedOn    string `json:"authorized_on"     bson:"authorized_on"`
	LastSeenOn      string `json:"last_seen_on"      bson:"last_seen_on"`
	Active          bool   `json:"active"            bson:"active"`
	CreatedOn       string `json:"createdon"         bson:"createdon"`
}

// AgentChannelSessionDoc tracks conversation continuity per channel user.
type AgentChannelSessionDoc struct {
	ID             string `json:"_id"               bson:"_id"`
	AgentChannelID string `json:"agent_channel_id"  bson:"agent_channel_id"`
	ChannelUserID  string `json:"channel_user_id"   bson:"channel_user_id"`
	ConversationID string `json:"conversation_id"   bson:"conversation_id"`
	MessageCount   int    `json:"message_count"     bson:"message_count"`
	LastMessageOn  string `json:"last_message_on"   bson:"last_message_on"`
	Active         bool   `json:"active"            bson:"active"`
	CreatedOn      string `json:"createdon"         bson:"createdon"`
}

// agentChannelPolicyKey is the context key used to pass security policy to AgentRunnerService.
type AgentChannelPolicyKey struct{}

// AgentCoordinatorPromptKey carries the coordinator prompt prefix into AgentRunnerService.
type AgentCoordinatorPromptKey struct{}

// AuthorizeUserRequest is used to authorize a channel user.
type AuthorizeUserRequest struct {
	ChannelUserID   string `json:"channel_user_id"`
	ChannelUsername string `json:"channel_username"`
	IACUserID       string `json:"iac_user_id"`
	IACUsername     string `json:"iac_username"`
}
