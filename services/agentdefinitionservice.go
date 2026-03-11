package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mdaxf/iac/documents"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/gorm"
)

// AgentDefinitionService manages agent definitions and their skill/MCP assignments.
// Agent skills and MCP servers are now stored in MongoDB (CollectionAgentSkills,
// CollectionAgentDefinitions, CollectionMCPServers); PostgreSQL is used only for
// schedule queries and legacy schema cleanup.
type AgentDefinitionService struct {
	db    *gorm.DB
	sqlDB *sql.DB
	docDB documents.DocumentDB
	iLog  logger.Log
}

var globalAgentDefinitionService *AgentDefinitionService

func GetGlobalAgentDefinitionService() *AgentDefinitionService { return globalAgentDefinitionService }
func SetGlobalAgentDefinitionService(s *AgentDefinitionService) { globalAgentDefinitionService = s }

// NewAgentDefinitionService creates the service, runs schema cleanup, and initialises MongoDB indexes.
func NewAgentDefinitionService(db *gorm.DB, sqlDB *sql.DB, docDB documents.DocumentDB) *AgentDefinitionService {
	svc := &AgentDefinitionService{
		db:    db,
		sqlDB: sqlDB,
		docDB: docDB,
		iLog: logger.Log{
			ModuleName:     logger.Framework,
			User:           "System",
			ControllerName: "AgentDefinitionService",
		},
	}
	svc.cleanupSchema()
	if docDB != nil {
		svc.initMongoIndexes()
	}
	return svc
}

// cleanupSchema removes stale tables created by a previous incorrect AutoMigrate call
// in the public schema and drops legacy columns from agent_definitions.
func (s *AgentDefinitionService) cleanupSchema() {
	s.db.Exec("DROP TABLE IF EXISTS public.agent_definition_skills CASCADE")
	s.db.Exec("DROP TABLE IF EXISTS public.agent_skills CASCADE")
	s.db.Exec("ALTER TABLE agent_definitions DROP COLUMN IF EXISTS system_user_id")
	s.db.Exec("ALTER TABLE agent_definitions DROP COLUMN IF EXISTS tools_enabled")
}

// initMongoIndexes creates indexes on the three agent runtime collections.
// Errors are silently ignored (indexes may already exist).
func (s *AgentDefinitionService) initMongoIndexes() {
	ctx := context.Background()

	// agent_skills
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentSkills,
		map[string]int{"name": 1},
		&documents.IndexOptions{Name: "idx_skill_name_unique", Unique: true, Sparse: true})
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentSkills,
		map[string]int{"category": 1, "active": 1},
		&documents.IndexOptions{Name: "idx_skill_category_active"})

	// agent_definitions
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentDefinitions,
		map[string]int{"name": 1},
		&documents.IndexOptions{Name: "idx_agent_name_unique", Unique: true})
	_ = s.docDB.CreateIndex(ctx, models.CollectionAgentDefinitions,
		map[string]int{"run_instance": 1, "active": 1, "enabled": 1},
		&documents.IndexOptions{Name: "idx_agent_run_instance"})

	// mcp_servers
	_ = s.docDB.CreateIndex(ctx, models.CollectionMCPServers,
		map[string]int{"name": 1},
		&documents.IndexOptions{Name: "idx_mcp_name_unique", Unique: true})

	s.iLog.Info("[MongoDB] indexes created on agent_skills, agent_definitions, mcp_servers")
}

// ensureDocDB returns an error if the document database connection is not initialized.
// Call at the top of any method that needs docDB to avoid nil-pointer panics.
func (s *AgentDefinitionService) ensureDocDB() error {
	if s.docDB == nil {
		return fmt.Errorf("document database not initialized for AgentDefinitionService — ensure MongoDB is configured and reachable at startup")
	}
	return nil
}

// ---- JSON round-trip helpers ----

// docToMap serialises a typed doc struct to map[string]interface{} via JSON.
// Dates are stored as RFC 3339 strings.
func docToMap(v interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

// mapToDoc deserialises a docDB result map into a typed doc struct via JSON.
func mapToDoc(m map[string]interface{}, out interface{}) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

// ---- Request types ----

// CreateAgentRequest holds parameters for creating a new agent.
type CreateAgentRequest struct {
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	Avatar             string          `json:"avatar"`
	ModelProvider      string          `json:"modelprovider"`
	ModelName          string          `json:"modelname"`
	SystemPrompt       string          `json:"systemprompt"`
	SkillNames         []string        `json:"skillnames"`
	MCPServerIDs       []string        `json:"mcpserverids"`
	MaxIterations      int             `json:"maxiterations"`
	Temperature        float64         `json:"temperature"`
	IsDefault          bool            `json:"isdefault"`
	Enabled            bool            `json:"enabled"`
	RunInstance        string          `json:"runinstance"`
	NotificationConfig models.JSONBMap `json:"notificationconfig"`
	CreatedBy          string          `json:"createdby"`
}

// UpdateAgentRequest holds updatable fields for an agent.
// nil slice = no change; empty slice = remove all assignments.
type UpdateAgentRequest struct {
	Name               *string         `json:"name"`
	Description        *string         `json:"description"`
	Avatar             *string         `json:"avatar"`
	ModelProvider      *string         `json:"modelprovider"`
	ModelName          *string         `json:"modelname"`
	SystemPrompt       *string         `json:"systemprompt"`
	SkillNames         []string        `json:"skillnames"`
	MCPServerIDs       []string        `json:"mcpserverids"`
	MaxIterations      *int            `json:"maxiterations"`
	Temperature        *float64        `json:"temperature"`
	IsDefault          *bool           `json:"isdefault"`
	Enabled            *bool           `json:"enabled"`
	RunInstance        *string         `json:"runinstance"`
	NotificationConfig models.JSONBMap `json:"notificationconfig"`
}

// ---- Agent CRUD ----

// CreateAgent creates a new agent in MongoDB with the requested skill/MCP assignments.
func (s *AgentDefinitionService) CreateAgent(req CreateAgentRequest) (*models.AgentDefinition, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, fmt.Errorf("agent name is required")
	}
	if req.SystemPrompt == "" {
		return nil, fmt.Errorf("system prompt is required")
	}
	if req.ModelProvider == "" {
		req.ModelProvider = "openai"
	}
	if req.ModelName == "" {
		req.ModelName = "gpt-4o"
	}
	if req.MaxIterations <= 0 {
		req.MaxIterations = 10
	}
	if req.Temperature <= 0 {
		req.Temperature = 0.7
	}
	if req.RunInstance == "" {
		req.RunInstance = "app"
	}

	// Resolve skill names → AgentSkill records
	skills, err := s.resolveSkillsByName(req.SkillNames)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve skills: %w", err)
	}

	// Resolve MCP server IDs → MCPServer records
	mcpServers, err := s.resolveMCPServersByID(req.MCPServerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve MCP servers: %w", err)
	}

	// Build embedded arrays for the doc
	skillNames := make([]string, len(skills))
	for i, sk := range skills {
		skillNames[i] = sk.Name
	}
	mcpIDs := make([]string, len(mcpServers))
	for i, m := range mcpServers {
		mcpIDs[i] = m.ID
	}

	notifCfg := req.NotificationConfig
	if notifCfg == nil {
		notifCfg = models.JSONBMap{}
	}

	now := time.Now()
	doc := &models.AgentDefinitionDoc{
		ID:                 uuid.New().String(),
		Name:               req.Name,
		Description:        req.Description,
		Avatar:             req.Avatar,
		ModelProvider:      req.ModelProvider,
		ModelName:          req.ModelName,
		SystemPrompt:       req.SystemPrompt,
		MaxIterations:      req.MaxIterations,
		Temperature:        req.Temperature,
		IsDefault:          req.IsDefault,
		Enabled:            req.Enabled,
		RunInstance:        req.RunInstance,
		NotificationConfig: map[string]string(notifCfg),
		SkillNames:         skillNames,
		MCPServerIDs:       mcpIDs,
		Active:             true,
		CreatedBy:          req.CreatedBy,
		CreatedOn:          now,
		ModifiedBy:         req.CreatedBy,
		ModifiedOn:         now,
		RowVersionStamp:    1,
	}

	m, err := docToMap(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialise agent doc: %w", err)
	}

	ctx := context.Background()
	if _, err := s.docDB.InsertOne(ctx, models.CollectionAgentDefinitions, m); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	s.iLog.Info(fmt.Sprintf("Created agent '%s' (id=%s, skills=%d)", req.Name, doc.ID, len(skills)))
	return s.GetAgent(doc.ID)
}

// GetAgent retrieves an agent by ID with Skills and MCPServers hydrated from MongoDB.
func (s *AgentDefinitionService) GetAgent(id string) (*models.AgentDefinition, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	filter := map[string]interface{}{"_id": id, "active": true}
	result, err := s.docDB.FindOne(ctx, models.CollectionAgentDefinitions, filter)
	if err != nil {
		return nil, err
	}
	return s.hydrateAgent(result)
}

// ListAgents returns all active agents with Skills and MCPServers hydrated.
func (s *AgentDefinitionService) ListAgents() ([]models.AgentDefinition, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	filter := map[string]interface{}{"active": true}
	opts := &documents.FindOptions{Sort: map[string]int{"name": 1}}
	results, err := s.docDB.FindMany(ctx, models.CollectionAgentDefinitions, filter, opts)
	if err != nil {
		return nil, err
	}

	agents := make([]models.AgentDefinition, 0, len(results))
	for _, r := range results {
		agent, err := s.hydrateAgent(r)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to hydrate agent: %v", err))
			continue
		}
		agents = append(agents, *agent)
	}
	return agents, nil
}

// ListByRunInstance returns all active, enabled agents with a specific run_instance value.
func (s *AgentDefinitionService) ListByRunInstance(runInstance string) ([]models.AgentDefinition, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	// "runinstance" matches the AgentDefinitionDoc JSON tag (no underscore).
	// The alternative "run_instance" key is set by UpdateAgent's $set patch;
	// query with $or so both originally-created and updated documents are found.
	filter := map[string]interface{}{
		"active":  true,
		"enabled": true,
		"$or": []interface{}{
			map[string]interface{}{"runinstance": runInstance},
			map[string]interface{}{"run_instance": runInstance},
		},
	}
	opts := &documents.FindOptions{Sort: map[string]int{"name": 1}}
	results, err := s.docDB.FindMany(ctx, models.CollectionAgentDefinitions, filter, opts)
	if err != nil {
		return nil, err
	}
	agents := make([]models.AgentDefinition, 0, len(results))
	for _, r := range results {
		agent, err := s.hydrateAgent(r)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to hydrate agent: %v", err))
			continue
		}
		agents = append(agents, *agent)
	}
	return agents, nil
}

// UpdateAgent updates mutable fields and optionally replaces skill/MCP assignments.
func (s *AgentDefinitionService) UpdateAgent(id string, req UpdateAgentRequest, modifiedBy string) (*models.AgentDefinition, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	// Verify agent exists
	ctx := context.Background()
	existFilter := map[string]interface{}{"_id": id, "active": true}
	if _, err := s.docDB.FindOne(ctx, models.CollectionAgentDefinitions, existFilter); err != nil {
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
	if req.Avatar != nil {
		setFields["avatar"] = *req.Avatar
	}
	if req.ModelProvider != nil {
		setFields["model_provider"] = *req.ModelProvider
	}
	if req.ModelName != nil {
		setFields["model_name"] = *req.ModelName
	}
	if req.SystemPrompt != nil {
		setFields["system_prompt"] = *req.SystemPrompt
	}
	if req.MaxIterations != nil {
		setFields["max_iterations"] = *req.MaxIterations
	}
	if req.Temperature != nil {
		setFields["temperature"] = *req.Temperature
	}
	if req.IsDefault != nil {
		setFields["is_default"] = *req.IsDefault
	}
	if req.Enabled != nil {
		setFields["enabled"] = *req.Enabled
	}
	if req.RunInstance != nil {
		// Write both keys so the document is found regardless of which key is queried.
		setFields["runinstance"] = *req.RunInstance
		setFields["run_instance"] = *req.RunInstance
	}
	if req.NotificationConfig != nil {
		setFields["notification_config"] = map[string]string(req.NotificationConfig)
	}

	// Replace skill assignments when SkillNames is explicitly provided
	if req.SkillNames != nil {
		skills, err := s.resolveSkillsByName(req.SkillNames)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve skills: %w", err)
		}
		names := make([]string, len(skills))
		for i, sk := range skills {
			names[i] = sk.Name
		}
		setFields["skill_names"] = names
	}

	// Replace MCP server assignments when MCPServerIDs is explicitly provided
	if req.MCPServerIDs != nil {
		mcpServers, err := s.resolveMCPServersByID(req.MCPServerIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve MCP servers: %w", err)
		}
		ids := make([]string, len(mcpServers))
		for i, m := range mcpServers {
			ids[i] = m.ID
		}
		setFields["mcp_server_ids"] = ids
	}

	update := map[string]interface{}{
		"$set": setFields,
		"$inc": map[string]interface{}{"rowversionstamp": 1},
	}

	if err := s.docDB.UpdateOne(ctx, models.CollectionAgentDefinitions,
		map[string]interface{}{"_id": id}, update); err != nil {
		return nil, err
	}

	return s.GetAgent(id)
}

// ToggleAgent enables or disables an agent.
func (s *AgentDefinitionService) ToggleAgent(id string, enabled bool, modifiedBy string) error {
	ctx := context.Background()
	filter := map[string]interface{}{"_id": id}
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"enabled":    enabled,
			"modifiedby": modifiedBy,
			"modifiedon": time.Now().Format(time.RFC3339),
		},
	}
	return s.docDB.UpdateOne(ctx, models.CollectionAgentDefinitions, filter, update)
}

// DeleteAgent soft-deletes an agent by setting active=false.
func (s *AgentDefinitionService) DeleteAgent(id string, deletedBy string) error {
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
	return s.docDB.UpdateOne(ctx, models.CollectionAgentDefinitions, filter, update)
}

// ---- Skill catalog ----

type CreateSkillRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayname"`
	Description string `json:"description"`
	Category    string `json:"category"`
	CreatedBy   string `json:"createdby"`
}

type UpdateSkillRequest struct {
	DisplayName *string `json:"displayname"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
}

// ListSkills returns all active skills in the catalog ordered by category then name.
func (s *AgentDefinitionService) ListSkills() ([]models.AgentSkill, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	filter := map[string]interface{}{"active": true}
	// Use SortFields (not Sort) for compound sorts — the MongoDB driver requires
	// an ordered document (bson.D) when sorting by multiple fields.
	opts := &documents.FindOptions{
		SortFields: []documents.SortField{
			{Field: "category", Order: 1},
			{Field: "name", Order: 1},
		},
	}
	results, err := s.docDB.FindMany(ctx, models.CollectionAgentSkills, filter, opts)
	if err != nil {
		return nil, err
	}

	skills := make([]models.AgentSkill, 0, len(results))
	for _, r := range results {
		var doc models.AgentSkillDoc
		if err := mapToDoc(r, &doc); err != nil {
			s.iLog.Error(fmt.Sprintf("Failed to decode skill: %v", err))
			continue
		}
		skills = append(skills, *models.AgentSkillDocToModel(&doc))
	}
	return skills, nil
}

// GetSkill retrieves a single active skill by ID.
func (s *AgentDefinitionService) GetSkill(id string) (*models.AgentSkill, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	ctx := context.Background()
	filter := map[string]interface{}{"_id": id, "active": true}
	result, err := s.docDB.FindOne(ctx, models.CollectionAgentSkills, filter)
	if err != nil {
		return nil, err
	}
	var doc models.AgentSkillDoc
	if err := mapToDoc(result, &doc); err != nil {
		return nil, err
	}
	return models.AgentSkillDocToModel(&doc), nil
}

// CreateSkill adds a new skill to the catalog.
func (s *AgentDefinitionService) CreateSkill(req CreateSkillRequest) (*models.AgentSkill, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	if req.Name == "" {
		return nil, fmt.Errorf("skill name is required")
	}
	if req.DisplayName == "" {
		req.DisplayName = req.Name
	}

	now := time.Now()
	doc := &models.AgentSkillDoc{
		ID:              uuid.New().String(),
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		Description:     req.Description,
		Category:        req.Category,
		Active:          true,
		CreatedBy:       req.CreatedBy,
		CreatedOn:       now,
		ModifiedBy:      req.CreatedBy,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}

	m, err := docToMap(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to serialise skill doc: %w", err)
	}

	ctx := context.Background()
	if _, err := s.docDB.InsertOne(ctx, models.CollectionAgentSkills, m); err != nil {
		return nil, fmt.Errorf("failed to create skill: %w", err)
	}
	return models.AgentSkillDocToModel(doc), nil
}

// UpdateSkill updates display metadata for a skill.
func (s *AgentDefinitionService) UpdateSkill(id string, req UpdateSkillRequest, modifiedBy string) (*models.AgentSkill, error) {
	if err := s.ensureDocDB(); err != nil {
		return nil, err
	}
	setFields := map[string]interface{}{
		"modifiedby": modifiedBy,
		"modifiedon": time.Now().Format(time.RFC3339),
	}
	if req.DisplayName != nil {
		setFields["display_name"] = *req.DisplayName
	}
	if req.Description != nil {
		setFields["description"] = *req.Description
	}
	if req.Category != nil {
		setFields["category"] = *req.Category
	}

	ctx := context.Background()
	update := map[string]interface{}{
		"$set": setFields,
		"$inc": map[string]interface{}{"rowversionstamp": 1},
	}
	if err := s.docDB.UpdateOne(ctx, models.CollectionAgentSkills,
		map[string]interface{}{"_id": id}, update); err != nil {
		return nil, err
	}
	return s.GetSkill(id)
}

// DeleteSkill soft-deletes a skill from the catalog.
func (s *AgentDefinitionService) DeleteSkill(id string, deletedBy string) error {
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
	return s.docDB.UpdateOne(ctx, models.CollectionAgentSkills, filter, update)
}

// ---- Internal helpers ----

// hydrateAgent decodes a raw docDB result map into a fully populated AgentDefinition
// with Skills and MCPServers resolved from MongoDB.
func (s *AgentDefinitionService) hydrateAgent(result map[string]interface{}) (*models.AgentDefinition, error) {
	var doc models.AgentDefinitionDoc
	if err := mapToDoc(result, &doc); err != nil {
		return nil, fmt.Errorf("failed to decode agent doc: %w", err)
	}

	agent := models.AgentDefinitionDocToModel(&doc)

	skills, err := s.resolveSkillsByName(doc.SkillNames)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve agent skills: %w", err)
	}
	agent.Skills = skills

	mcpServers, err := s.resolveMCPServersByID(doc.MCPServerIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve agent MCP servers: %w", err)
	}
	agent.MCPServers = mcpServers

	return agent, nil
}

// resolveSkillsByName looks up AgentSkill records by their tool name from MongoDB.
// Unknown names are silently skipped.
func (s *AgentDefinitionService) resolveSkillsByName(names []string) ([]models.AgentSkill, error) {
	if len(names) == 0 {
		return []models.AgentSkill{}, nil
	}

	nameSlice := make([]interface{}, len(names))
	for i, n := range names {
		nameSlice[i] = n
	}

	ctx := context.Background()
	filter := map[string]interface{}{
		"name":   map[string]interface{}{"$in": nameSlice},
		"active": true,
	}
	results, err := s.docDB.FindMany(ctx, models.CollectionAgentSkills, filter, nil)
	if err != nil {
		return nil, err
	}

	skills := make([]models.AgentSkill, 0, len(results))
	for _, r := range results {
		var doc models.AgentSkillDoc
		if err := mapToDoc(r, &doc); err != nil {
			continue
		}
		skills = append(skills, *models.AgentSkillDocToModel(&doc))
	}
	return skills, nil
}

// resolveMCPServersByID looks up MCPServer records by their UUID from MongoDB.
// Unknown IDs are silently skipped.
func (s *AgentDefinitionService) resolveMCPServersByID(ids []string) ([]models.MCPServer, error) {
	if len(ids) == 0 {
		return []models.MCPServer{}, nil
	}

	idSlice := make([]interface{}, len(ids))
	for i, id := range ids {
		idSlice[i] = id
	}

	ctx := context.Background()
	filter := map[string]interface{}{
		"_id":    map[string]interface{}{"$in": idSlice},
		"active": true,
	}
	results, err := s.docDB.FindMany(ctx, models.CollectionMCPServers, filter, nil)
	if err != nil {
		return nil, err
	}

	servers := make([]models.MCPServer, 0, len(results))
	for _, r := range results {
		var doc models.MCPServerDoc
		if err := mapToDoc(r, &doc); err != nil {
			continue
		}
		servers = append(servers, *models.MCPServerDocToModel(&doc))
	}
	return servers, nil
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

func sanitizeLoginName(name string) string {
	clean := nonAlphaNum.ReplaceAllString(strings.ToLower(name), "_")
	if len(clean) > 30 {
		clean = clean[:30]
	}
	return strings.Trim(clean, "_")
}

// builtinSkillDef holds the catalog metadata for a tool-backed built-in skill.
type builtinSkillDef struct {
	Name        string
	DisplayName string
	Description string // short UI description
	Content     string // extended LLM-facing instructions/schema (appended to tool description at run time)
	Category    string
}

// SyncBuiltinWebSkills ensures that each registered web/search tool has a
// corresponding skill catalog entry in MongoDB so agents can assign them.
// Existing entries are not modified. New entries are created with skill_type="builtin".
// Called at startup after both the tool registry and definition service are initialised.
func (s *AgentDefinitionService) SyncBuiltinWebSkills(registry *AgentToolRegistryService) {
	if err := s.ensureDocDB(); err != nil {
		s.iLog.Error(fmt.Sprintf("SyncBuiltinWebSkills: %v", err))
		return
	}
	if registry == nil {
		return
	}

	builtins := []builtinSkillDef{
		{
			Name:        "web_search",
			DisplayName: "Web Search (auto)",
			Description: "Searches the web using the configured primary provider with automatic fallback. The active provider is set in aiconfig.json under search.provider.",
			Category:    "Web & Search",
		},
		{
			Name:        "web_fetch",
			DisplayName: "Web Fetch",
			Description: "Fetches a public web page and returns its readable text content. Use after a search to read the full content of a specific URL.",
			Category:    "Web & Search",
		},
		{
			Name:        "http_request",
			DisplayName: "HTTP Request",
			Description: "Makes HTTP requests (GET/POST/PUT/DELETE/PATCH) to external APIs or services.",
			Category:    "Web & Search",
		},
		{
			Name:        "search_duckduckgo",
			DisplayName: "Search: DuckDuckGo",
			Description: "Searches using DuckDuckGo Instant Answers (free, no API key). Returns Wikipedia summaries and related topics. Always available as a free fallback.",
			Category:    "Web & Search",
		},
		{
			Name:        "search_brave",
			DisplayName: "Search: Brave",
			Description: "Searches the web using the Brave Search API — an independent, privacy-focused index. Requires brave.api_key in aiconfig.json.",
			Category:    "Web & Search",
		},
		{
			Name:        "search_tavily",
			DisplayName: "Search: Tavily",
			Description: "Searches using Tavily — an AI-optimised search API that returns pre-cleaned content ideal for LLM agents. Requires tavily.api_key in aiconfig.json.",
			Category:    "Web & Search",
		},
		{
			Name:        "search_serpapi",
			DisplayName: "Search: SerpAPI (Google/Bing)",
			Description: "Searches via SerpAPI covering Google, Bing, and other engines. Specify engine parameter: google (default), bing, yahoo. Requires serpapi.api_key in aiconfig.json.",
			Category:    "Web & Search",
		},
		{
			Name:        "search_google",
			DisplayName: "Search: Google CSE",
			Description: "Searches using Google Custom Search Engine. Returns authoritative Google results. Requires google_cse.api_key and google_cse.cx in aiconfig.json.",
			Category:    "Web & Search",
		},
	}

	ctx := context.Background()
	now := time.Now()

	for _, def := range builtins {
		// Check if skill already exists
		existing, _ := s.docDB.FindOne(ctx, models.CollectionAgentSkills,
			map[string]interface{}{"name": def.Name})
		if existing != nil {
			continue // already in catalog, don't overwrite user edits
		}

		doc := models.AgentSkillDoc{
			ID:              uuid.New().String(),
			Name:            def.Name,
			DisplayName:     def.DisplayName,
			Description:     def.Description,
			Content:         def.Content,
			Category:        def.Category,
			Active:          true,
			SkillType:       "builtin",
			CreatedBy:       "system",
			ModifiedBy:      "system",
			CreatedOn:       now,
			ModifiedOn:      now,
			RowVersionStamp: 1,
		}
		m, err := docToMap(doc)
		if err != nil {
			s.iLog.Error(fmt.Sprintf("SyncBuiltinWebSkills: failed to serialise %s: %v", def.Name, err))
			continue
		}
		if _, err := s.docDB.InsertOne(ctx, models.CollectionAgentSkills, m); err != nil {
			s.iLog.Error(fmt.Sprintf("SyncBuiltinWebSkills: failed to insert %s: %v", def.Name, err))
		} else {
			s.iLog.Info(fmt.Sprintf("SyncBuiltinWebSkills: registered skill catalog entry for '%s'", def.Name))
		}
	}
}

// uiRenderSkillContent is the authoritative spec for the ui_render tool, stored in the
// MongoDB skill catalog. GetToolDefinitionsForSkills appends this to the tool description
// so any LLM using an agent with this skill sees the full schema.
const uiRenderSkillContent = `## ui_render — JSON Spec Reference

The spec_json parameter must be a JSON object. All fields are optional except "sections".

### Top-level fields
| field    | type   | description                          |
|----------|--------|--------------------------------------|
| title    | string | Panel header title                   |
| subtitle | string | Panel header subtitle                |
| sections | array  | Ordered list of UI sections (required)|

### Section types

#### kpi_row — metric cards in a horizontal row
` + "```json" + `
{"type":"kpi_row","items":[
  {"label":"Revenue","value":"$1.2M","trend":"+12%","trend_up":true,"color":"green"},
  {"label":"Orders","value":"4821","trend":"+8%","trend_up":true,"color":"blue"},
  {"label":"Returns","value":"142","trend":"-3%","trend_up":false,"color":"red"}
]}
` + "```" + `
color: "blue" | "green" | "red" | "orange" | "purple" | "gray"

#### chart — Chart.js chart
` + "```json" + `
{"type":"chart","chart_type":"bar","title":"Monthly Revenue",
 "labels":["Jan","Feb","Mar"],
 "datasets":[{"label":"Revenue","data":[400000,380000,420000],"color":"#667eea"}]}
` + "```" + `
chart_type: "bar" | "pie" | "line" | "doughnut"

#### table — sortable data table (user can click column headers to sort)
` + "```json" + `
{"type":"table","title":"Top Products",
 "columns":["Product","Units","Revenue"],
 "rows":[["Widget A","1200","$48,000"],["Widget B","980","$39,200"]]}
` + "```" + `

#### form — interactive input form (submitted values are sent back as your next message)
` + "```json" + `
{"type":"form","title":"Apply Filters","submit_label":"Apply",
 "fields":[
   {"id":"month","label":"Month","type":"select","options":["Jan","Feb","Mar"]},
   {"id":"q","label":"Search","type":"text","placeholder":"keyword"},
   {"id":"start","label":"From","type":"date"},
   {"id":"amount","label":"Min Amount","type":"number"}
 ]}
` + "```" + `
field types: "text" | "number" | "select" | "date" | "textarea"

#### text — formatted paragraph
` + "```json" + `
{"type":"text","content":"Analysis: Widget A drove Q1 growth by 12% QoQ due to seasonal demand."}
` + "```" + `

### Complete example
` + "```json" + `
{
  "title": "Q1 2026 Sales Dashboard",
  "subtitle": "Performance overview across all regions",
  "sections": [
    {"type":"kpi_row","items":[
      {"label":"Revenue","value":"$1.2M","trend":"+12%","trend_up":true,"color":"green"},
      {"label":"Orders","value":"4821","trend":"+8%","trend_up":true,"color":"blue"},
      {"label":"Avg Order","value":"$249","trend":"+4%","trend_up":true,"color":"purple"}
    ]},
    {"type":"chart","chart_type":"bar","title":"Monthly Revenue",
     "labels":["Jan","Feb","Mar"],
     "datasets":[{"label":"Revenue","data":[380000,400000,420000]}]},
    {"type":"table","title":"Top Products",
     "columns":["Product","Units","Revenue"],
     "rows":[["Widget A","1200","$48,000"],["Widget B","980","$39,200"],["Widget C","760","$30,400"]]},
    {"type":"form","title":"Drill Down","submit_label":"Load Data",
     "fields":[
       {"id":"region","label":"Region","type":"select","options":["All","North","South","East","West"]},
       {"id":"month","label":"Month","type":"select","options":["Jan","Feb","Mar"]}
     ]},
    {"type":"text","content":"Widget A outperformed projections driven by a marketing campaign in January."}
  ]
}
` + "```"

// SyncBuiltinUISkills ensures the ui_render skill catalog entry exists in MongoDB.
// The skill's Content field holds the full JSON spec so any agent assigned this skill
// receives the complete schema in the LLM tool description at run time.
// Existing entries are not modified.
func (s *AgentDefinitionService) SyncBuiltinUISkills() {
	if err := s.ensureDocDB(); err != nil {
		s.iLog.Error(fmt.Sprintf("SyncBuiltinUISkills: %v", err))
		return
	}

	def := builtinSkillDef{
		Name:        "ui_render",
		DisplayName: "UI Render",
		Description: "Renders rich interactive UI panels (dashboards, KPI cards, charts, tables, forms) directly inside the agent chat window. Prefer this over plain text for any structured data presentation.",
		Content:     uiRenderSkillContent,
		Category:    "UI & Display",
	}

	ctx := context.Background()
	existing, _ := s.docDB.FindOne(ctx, models.CollectionAgentSkills,
		map[string]interface{}{"name": def.Name})
	if existing != nil {
		s.iLog.Info("SyncBuiltinUISkills: ui_render skill already in catalog")
		return
	}

	now := time.Now()
	doc := models.AgentSkillDoc{
		ID:              uuid.New().String(),
		Name:            def.Name,
		DisplayName:     def.DisplayName,
		Description:     def.Description,
		Content:         def.Content,
		Category:        def.Category,
		Active:          true,
		SkillType:       "builtin",
		CreatedBy:       "system",
		ModifiedBy:      "system",
		CreatedOn:       now,
		ModifiedOn:      now,
		RowVersionStamp: 1,
	}
	m, err := docToMap(doc)
	if err != nil {
		s.iLog.Error(fmt.Sprintf("SyncBuiltinUISkills: failed to serialise ui_render: %v", err))
		return
	}
	if _, err := s.docDB.InsertOne(ctx, models.CollectionAgentSkills, m); err != nil {
		s.iLog.Error(fmt.Sprintf("SyncBuiltinUISkills: failed to insert ui_render: %v", err))
	} else {
		s.iLog.Info("SyncBuiltinUISkills: registered ui_render skill in catalog")
	}
}
