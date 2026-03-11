// migrate_agent_to_mongo is a one-off migration binary that copies
// agent_skills, mcp_servers, and agent_definitions from PostgreSQL to MongoDB.
//
// Usage:
//   go run ./cmd/migrate_agent_to_mongo
//
// The binary reads configuration.json from the current working directory.
// It is idempotent: if a MongoDB collection already contains documents, that
// collection is skipped.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/mdaxf/iac/documents"
	_ "github.com/mdaxf/iac/documents/mongodb"
	"github.com/mdaxf/iac/logger"
	"github.com/mdaxf/iac/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ---- Local PostgreSQL structs (keep many2many tags for legacy PG reads) ----

type pgAgentSkill struct {
	ID              string       `gorm:"column:id;primaryKey"`
	Name            string       `gorm:"column:name"`
	DisplayName     string       `gorm:"column:display_name"`
	Description     string       `gorm:"column:description"`
	Category        string       `gorm:"column:category"`
	Active          bool         `gorm:"column:active"`
	CreatedBy       string       `gorm:"column:createdby"`
	CreatedOn       sql.NullTime `gorm:"column:createdon"`
	ModifiedBy      string       `gorm:"column:modifiedby"`
	ModifiedOn      sql.NullTime `gorm:"column:modifiedon"`
	RowVersionStamp int          `gorm:"column:rowversionstamp"`
}

func (pgAgentSkill) TableName() string { return "agent_skills" }

type pgMCPServer struct {
	ID              string       `gorm:"column:id;primaryKey"`
	Name            string       `gorm:"column:name"`
	Description     string       `gorm:"column:description"`
	TransportType   string       `gorm:"column:transport_type"`
	URL             string       `gorm:"column:url"`
	MCPPath         string       `gorm:"column:mcp_path"`
	Command         string       `gorm:"column:command"`
	Args            models.StringSlice `gorm:"column:args;type:jsonb"`
	Headers         models.JSONBMap    `gorm:"column:headers;type:jsonb"`
	Enabled         bool         `gorm:"column:enabled"`
	Active          bool         `gorm:"column:active"`
	CreatedBy       string       `gorm:"column:createdby"`
	CreatedOn       sql.NullTime `gorm:"column:createdon"`
	ModifiedBy      string       `gorm:"column:modifiedby"`
	ModifiedOn      sql.NullTime `gorm:"column:modifiedon"`
	RowVersionStamp int          `gorm:"column:rowversionstamp"`
}

func (pgMCPServer) TableName() string { return "mcp_servers" }

type pgAgentDefinition struct {
	ID                 string             `gorm:"column:id;primaryKey"`
	Name               string             `gorm:"column:name"`
	Description        string             `gorm:"column:description"`
	Avatar             string             `gorm:"column:avatar"`
	ModelProvider      string             `gorm:"column:model_provider"`
	ModelName          string             `gorm:"column:model_name"`
	SystemPrompt       string             `gorm:"column:system_prompt"`
	MaxIterations      int                `gorm:"column:max_iterations"`
	Temperature        float64            `gorm:"column:temperature"`
	IsDefault          bool               `gorm:"column:is_default"`
	Enabled            bool               `gorm:"column:enabled"`
	RunInstance        string             `gorm:"column:run_instance"`
	NotificationConfig models.JSONBMap    `gorm:"column:notification_config;type:jsonb"`
	Active             bool               `gorm:"column:active"`
	ReferenceID        string             `gorm:"column:referenceid"`
	CreatedBy          string             `gorm:"column:createdby"`
	CreatedOn          sql.NullTime       `gorm:"column:createdon"`
	ModifiedBy         string             `gorm:"column:modifiedby"`
	ModifiedOn         sql.NullTime       `gorm:"column:modifiedon"`
	RowVersionStamp    int                `gorm:"column:rowversionstamp"`

	// Legacy many2many associations — used only for reading from PG during migration.
	Skills     []pgAgentSkill `gorm:"many2many:agent_definition_skills;joinForeignKey:agent_definition_id;joinReferences:agent_skill_id"`
	MCPServers []pgMCPServer  `gorm:"many2many:agent_mcp_servers;joinForeignKey:agent_definition_id;joinReferences:mcp_server_id"`
}

func (pgAgentDefinition) TableName() string { return "agent_definitions" }

// ---- Configuration parsing ----

type rawConfig struct {
	Database  map[string]interface{} `json:"database"`
	DocumentDB map[string]interface{} `json:"documentdb"`
}

func loadConfig(path string) (*rawConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg rawConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &cfg, nil
}

func str(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// ---- MongoDB doc helpers ----

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

// ---- main ----

func main() {
	// Initialize the IAC logger framework (required by MongoDB adapter).
	// Use console output at info level; no file logging needed for a migration tool.
	logger.Init(map[string]interface{}{"level": "info"})

	cfgPath := "configuration.json"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// ── Connect to PostgreSQL ──────────────────────────────────────────────
	pgDSN := str(cfg.Database, "connection")
	if pgDSN == "" {
		log.Fatal("No PostgreSQL connection string found in configuration.json (.database.connection)")
	}
	log.Printf("Connecting to PostgreSQL: %s", maskPassword(pgDSN))

	pgDB, err := gorm.Open(postgres.Open(pgDSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	log.Println("PostgreSQL connection OK")

	// ── Connect to MongoDB ─────────────────────────────────────────────────
	mongoConnStr := str(cfg.DocumentDB, "connection")
	mongoDBName  := str(cfg.DocumentDB, "database")
	if mongoConnStr == "" {
		log.Fatal("No MongoDB connection string found in configuration.json (.documentdb.connection)")
	}
	if mongoDBName == "" {
		mongoDBName = "IAC_CFG"
	}
	log.Printf("Connecting to MongoDB: %s / %s", mongoConnStr, mongoDBName)

	mongoHost, mongoPort := parseMongoHostPort(mongoConnStr)
	docConfig := &documents.DocDBConfig{
		Type:     documents.DocDBTypeMongoDB,
		Host:     mongoHost,
		Port:     mongoPort,
		Database: mongoDBName,
		Options:  map[string]string{"connection_string": mongoConnStr},
	}

	docDB, err := documents.NewDocumentDB(docConfig)
	if err != nil {
		log.Fatalf("Failed to create MongoDB connection: %v", err)
	}
	if err := docDB.Connect(docConfig); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	log.Println("MongoDB connection OK")

	ctx := context.Background()

	// ── Migrate agent_skills ───────────────────────────────────────────────
	skillCount, err := migrateSkills(ctx, pgDB, docDB)
	if err != nil {
		log.Fatalf("Failed to migrate agent_skills: %v", err)
	}

	// ── Migrate mcp_servers ────────────────────────────────────────────────
	mcpCount, err := migrateMCPServers(ctx, pgDB, docDB)
	if err != nil {
		log.Fatalf("Failed to migrate mcp_servers: %v", err)
	}

	// ── Migrate agent_definitions ──────────────────────────────────────────
	agentCount, err := migrateAgentDefinitions(ctx, pgDB, docDB)
	if err != nil {
		log.Fatalf("Failed to migrate agent_definitions: %v", err)
	}

	// ── Create indexes ─────────────────────────────────────────────────────
	createIndexes(ctx, docDB)

	fmt.Printf("\n=== Migration Summary ===\n")
	fmt.Printf("  agent_skills:      %d documents\n", skillCount)
	fmt.Printf("  mcp_servers:       %d documents\n", mcpCount)
	fmt.Printf("  agent_definitions: %d documents\n", agentCount)
	fmt.Println("Migration complete.")
}

func migrateSkills(ctx context.Context, pgDB *gorm.DB, docDB documents.DocumentDB) (int, error) {
	existing, err := docDB.CountDocuments(ctx, models.CollectionAgentSkills, map[string]interface{}{})
	if err != nil {
		return 0, fmt.Errorf("count agent_skills: %w", err)
	}
	if existing > 0 {
		log.Printf("[agent_skills] already has %d documents — skipping", existing)
		return int(existing), nil
	}

	var pgSkills []pgAgentSkill
	if err := pgDB.Find(&pgSkills).Error; err != nil {
		return 0, fmt.Errorf("read PG agent_skills: %w", err)
	}

	if len(pgSkills) == 0 {
		log.Println("[agent_skills] no rows in PostgreSQL — nothing to migrate")
		return 0, nil
	}

	docs := make([]interface{}, 0, len(pgSkills))
	for _, sk := range pgSkills {
		doc := models.AgentSkillDoc{
			ID:              sk.ID,
			Name:            sk.Name,
			DisplayName:     sk.DisplayName,
			Description:     sk.Description,
			Category:        sk.Category,
			Active:          sk.Active,
			CreatedBy:       sk.CreatedBy,
			ModifiedBy:      sk.ModifiedBy,
			RowVersionStamp: sk.RowVersionStamp,
		}
		if sk.CreatedOn.Valid {
			doc.CreatedOn = sk.CreatedOn.Time
		}
		if sk.ModifiedOn.Valid {
			doc.ModifiedOn = sk.ModifiedOn.Time
		}
		m, err := docToMap(doc)
		if err != nil {
			return 0, fmt.Errorf("serialise skill %s: %w", sk.ID, err)
		}
		docs = append(docs, m)
	}

	if _, err := docDB.InsertMany(ctx, models.CollectionAgentSkills, docs); err != nil {
		return 0, fmt.Errorf("insert agent_skills: %w", err)
	}

	log.Printf("[agent_skills] migrated %d documents", len(pgSkills))
	return len(pgSkills), nil
}

func migrateMCPServers(ctx context.Context, pgDB *gorm.DB, docDB documents.DocumentDB) (int, error) {
	existing, err := docDB.CountDocuments(ctx, models.CollectionMCPServers, map[string]interface{}{})
	if err != nil {
		return 0, fmt.Errorf("count mcp_servers: %w", err)
	}
	if existing > 0 {
		log.Printf("[mcp_servers] already has %d documents — skipping", existing)
		return int(existing), nil
	}

	var pgServers []pgMCPServer
	if err := pgDB.Find(&pgServers).Error; err != nil {
		return 0, fmt.Errorf("read PG mcp_servers: %w", err)
	}

	if len(pgServers) == 0 {
		log.Println("[mcp_servers] no rows in PostgreSQL — nothing to migrate")
		return 0, nil
	}

	docs := make([]interface{}, 0, len(pgServers))
	for _, srv := range pgServers {
		args := []string(srv.Args)
		if args == nil {
			args = []string{}
		}
		headers := map[string]string(srv.Headers)
		if headers == nil {
			headers = map[string]string{}
		}
		doc := models.MCPServerDoc{
			ID:              srv.ID,
			Name:            srv.Name,
			Description:     srv.Description,
			TransportType:   srv.TransportType,
			URL:             srv.URL,
			MCPPath:         srv.MCPPath,
			Command:         srv.Command,
			Args:            args,
			Headers:         headers,
			Enabled:         srv.Enabled,
			Active:          srv.Active,
			CreatedBy:       srv.CreatedBy,
			ModifiedBy:      srv.ModifiedBy,
			RowVersionStamp: srv.RowVersionStamp,
		}
		if srv.CreatedOn.Valid {
			doc.CreatedOn = srv.CreatedOn.Time
		}
		if srv.ModifiedOn.Valid {
			doc.ModifiedOn = srv.ModifiedOn.Time
		}
		m, err := docToMap(doc)
		if err != nil {
			return 0, fmt.Errorf("serialise mcp_server %s: %w", srv.ID, err)
		}
		docs = append(docs, m)
	}

	if _, err := docDB.InsertMany(ctx, models.CollectionMCPServers, docs); err != nil {
		return 0, fmt.Errorf("insert mcp_servers: %w", err)
	}

	log.Printf("[mcp_servers] migrated %d documents", len(pgServers))
	return len(pgServers), nil
}

func migrateAgentDefinitions(ctx context.Context, pgDB *gorm.DB, docDB documents.DocumentDB) (int, error) {
	existing, err := docDB.CountDocuments(ctx, models.CollectionAgentDefinitions, map[string]interface{}{})
	if err != nil {
		return 0, fmt.Errorf("count agent_definitions: %w", err)
	}
	if existing > 0 {
		log.Printf("[agent_definitions] already has %d documents — skipping", existing)
		return int(existing), nil
	}

	var pgAgents []pgAgentDefinition
	if err := pgDB.Preload("Skills").Preload("MCPServers").Find(&pgAgents).Error; err != nil {
		return 0, fmt.Errorf("read PG agent_definitions: %w", err)
	}

	if len(pgAgents) == 0 {
		log.Println("[agent_definitions] no rows in PostgreSQL — nothing to migrate")
		return 0, nil
	}

	docs := make([]interface{}, 0, len(pgAgents))
	for _, ag := range pgAgents {
		skillNames := make([]string, len(ag.Skills))
		for i, s := range ag.Skills {
			skillNames[i] = s.Name
		}
		mcpIDs := make([]string, len(ag.MCPServers))
		for i, m := range ag.MCPServers {
			mcpIDs[i] = m.ID
		}
		notifCfg := map[string]string(ag.NotificationConfig)
		if notifCfg == nil {
			notifCfg = map[string]string{}
		}

		doc := models.AgentDefinitionDoc{
			ID:                 ag.ID,
			Name:               ag.Name,
			Description:        ag.Description,
			Avatar:             ag.Avatar,
			ModelProvider:      ag.ModelProvider,
			ModelName:          ag.ModelName,
			SystemPrompt:       ag.SystemPrompt,
			MaxIterations:      ag.MaxIterations,
			Temperature:        ag.Temperature,
			IsDefault:          ag.IsDefault,
			Enabled:            ag.Enabled,
			RunInstance:        ag.RunInstance,
			NotificationConfig: notifCfg,
			SkillNames:         skillNames,
			MCPServerIDs:       mcpIDs,
			Active:             ag.Active,
			ReferenceID:        ag.ReferenceID,
			CreatedBy:          ag.CreatedBy,
			ModifiedBy:         ag.ModifiedBy,
			RowVersionStamp:    ag.RowVersionStamp,
		}
		if ag.CreatedOn.Valid {
			doc.CreatedOn = ag.CreatedOn.Time
		}
		if ag.ModifiedOn.Valid {
			doc.ModifiedOn = ag.ModifiedOn.Time
		}

		m, err := docToMap(doc)
		if err != nil {
			return 0, fmt.Errorf("serialise agent %s: %w", ag.ID, err)
		}
		docs = append(docs, m)
	}

	if _, err := docDB.InsertMany(ctx, models.CollectionAgentDefinitions, docs); err != nil {
		return 0, fmt.Errorf("insert agent_definitions: %w", err)
	}

	log.Printf("[agent_definitions] migrated %d documents", len(pgAgents))
	return len(pgAgents), nil
}

func createIndexes(ctx context.Context, docDB documents.DocumentDB) {
	_ = docDB.CreateIndex(ctx, models.CollectionAgentSkills,
		map[string]int{"name": 1},
		&documents.IndexOptions{Name: "idx_skill_name_unique", Unique: true, Sparse: true})
	_ = docDB.CreateIndex(ctx, models.CollectionAgentDefinitions,
		map[string]int{"name": 1},
		&documents.IndexOptions{Name: "idx_agent_name_unique", Unique: true})
	_ = docDB.CreateIndex(ctx, models.CollectionMCPServers,
		map[string]int{"name": 1},
		&documents.IndexOptions{Name: "idx_mcp_name_unique", Unique: true})
	log.Println("MongoDB indexes created/verified")
}

// parseMongoHostPort extracts host and port from a mongodb:// URI.
func parseMongoHostPort(connStr string) (string, int) {
	u, err := url.Parse(connStr)
	if err != nil {
		return "localhost", 27017
	}
	host := u.Hostname()
	if host == "" {
		host = "localhost"
	}
	port := 27017
	if p := u.Port(); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	return host, port
}

// maskPassword redacts the password in a connection URL for safe logging.
func maskPassword(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return "(unparseable DSN)"
	}
	if u.User != nil {
		u.User = url.UserPassword(u.User.Username(), "***")
	}
	return u.String()
}

// Ensure time package is used (imported for sql.NullTime in struct tags).
var _ = time.Now
