package schema

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"

	"github.com/mdaxf/iac/logger"
)

type SchemaController struct {
}

// DBColumn represents a database column
type DBColumn struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Nullable     bool        `json:"nullable"`
	PrimaryKey   bool        `json:"primary_key"`
	ForeignKey   *ForeignKey `json:"foreign_key,omitempty"`
	DefaultValue *string     `json:"default_value,omitempty"`
	Comment      *string     `json:"comment,omitempty"`
	Extra        string      `json:"extra"`
}

// ForeignKey represents a foreign key relationship
type ForeignKey struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

// DBTable represents a database table
type DBTable struct {
	Name    string     `json:"name"`
	Schema  string     `json:"schema,omitempty"`
	Columns []DBColumn `json:"columns"`
	Comment *string    `json:"comment,omitempty"`
}

// DBRelationship represents a table relationship
type DBRelationship struct {
	ID           string `json:"id"`
	SourceTable  string `json:"source_table"`
	SourceColumn string `json:"source_column"`
	TargetTable  string `json:"target_table"`
	TargetColumn string `json:"target_column"`
	Type         string `json:"type"` // "1:1", "1:N", "N:M"
}

// SchemaDiagram represents the complete diagram
type SchemaDiagram struct {
	Tables        []DBTable        `json:"tables"`
	Relationships []DBRelationship `json:"relationships"`
}

// TablePosition represents table position in diagram
type TablePosition struct {
	Table string  `json:"table"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

// SchemaGenerationRequest is the request for generating diagram
type SchemaGenerationRequest struct {
	Tables          []string `json:"tables"`
	IncludeChildren bool     `json:"includeChildren"`
	Schema          string   `json:"schema"` // Optional: database schema name, defaults to current_schema()
}

// SchemaGenerationResponse is the response for diagram generation
type SchemaGenerationResponse struct {
	Diagram   SchemaDiagram   `json:"diagram"`
	Positions []TablePosition `json:"positions"`
}

// DatasetSchema represents the JSON schema format
type DatasetSchema struct {
	Schema         string                 `json:"$schema"`
	Ref            string                 `json:"$ref"`
	DatasourceType string                 `json:"datasourcetype"`
	Datasource     string                 `json:"datasource"`
	ListFields     []string               `json:"listfields"`
	HiddenFields   []string               `json:"hiddenfields"`
	KeyField       string                 `json:"keyfield"`
	DetailPage     map[string]interface{} `json:"detailpage,omitempty"`
	Definitions    map[string]interface{} `json:"definitions"`
}

// PropertyDefinition represents a field definition
type PropertyDefinition struct {
	Type      string                 `json:"type"`
	Format    string                 `json:"format,omitempty"`
	Readonly  bool                   `json:"readonly,omitempty"`
	Lng       map[string]interface{} `json:"lng"`
	Nullvalue string                 `json:"nullvalue"`
	External  bool                   `json:"external"`
}

// ListTables returns all table names from the database
func (sc *SchemaController) ListTables(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.ListTables"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.ListTables", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	// Get optional schema parameter from query string
	schemaName := c.Query("schema")
	iLog.Debug(fmt.Sprintf("Listing database tables for schema: %s", schemaName))

	// Get tables using helper
	tables, err := sc.getAllTableNames(schemaName, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting tables: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Found %d tables", len(tables)))
	c.JSON(http.StatusOK, tables)
}

// GetTableMetadata returns metadata for a specific table
func (sc *SchemaController) GetTableMetadata(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GetTableMetadata"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GetTableMetadata", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	tableName := c.Param("tableName")
	schemaName := c.Query("schema")
	iLog.Debug(fmt.Sprintf("Getting metadata for table: %s, schema: %s", tableName, schemaName))

	table, err := sc.getTableStructure(tableName, schemaName, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting table metadata: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, table)
}

// GetTablesMetadata returns metadata for multiple tables
func (sc *SchemaController) GetTablesMetadata(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GetTablesMetadata"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GetTablesMetadata", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request struct {
		Tables []string `json:"tables"`
		Schema string   `json:"schema"`
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Getting metadata for %d tables, schema: %s", len(request.Tables), request.Schema))

	var tables []DBTable
	for _, tableName := range request.Tables {
		table, err := sc.getTableStructure(tableName, request.Schema, &iLog)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error getting metadata for table %s: %v", tableName, err))
			continue
		}
		tables = append(tables, *table)
	}

	c.JSON(http.StatusOK, tables)
}

// GenerateDiagram generates a schema diagram with relationships
func (sc *SchemaController) GenerateDiagram(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GenerateDiagram"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GenerateDiagram", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request SchemaGenerationRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generating diagram for tables: %v, includeChildren: %v, schema: %s", request.Tables, request.IncludeChildren, request.Schema))

	// Get tables to include
	var tableNames []string
	if len(request.Tables) == 0 {
		// Get all tables
		tableNames, err = sc.getAllTableNames(request.Schema, &iLog)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		tableNames = request.Tables
		// If includeChildren, add related tables
		if request.IncludeChildren {
			tableNames, err = sc.getTablesWithChildren(tableNames, request.Schema, &iLog)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Get table structures
	var tables []DBTable
	for _, tableName := range tableNames {
		table, err := sc.getTableStructure(tableName, request.Schema, &iLog)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error getting table structure for %s: %v", tableName, err))
			continue
		}
		tables = append(tables, *table)
	}

	// Get relationships
	relationships, err := sc.getRelationships(tableNames, request.Schema, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting relationships: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate positions (simple grid layout)
	positions := sc.calculatePositions(tables)

	response := SchemaGenerationResponse{
		Diagram: SchemaDiagram{
			Tables:        tables,
			Relationships: relationships,
		},
		Positions: positions,
	}

	iLog.Debug(fmt.Sprintf("Generated diagram with %d tables and %d relationships", len(tables), len(relationships)))
	c.JSON(http.StatusOK, response)
}

// GenerateDatasetSchema generates a dataset schema file for a table
func (sc *SchemaController) GenerateDatasetSchema(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GenerateDatasetSchema"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GenerateDatasetSchema", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	tableName := c.Param("tableName")
	schemaName := c.Query("schema")
	iLog.Debug(fmt.Sprintf("Generating dataset schema for table: %s, schema: %s", tableName, schemaName))

	schema, err := sc.generateDatasetSchemaForTable(tableName, schemaName, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating dataset schema: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, schema)
}

// GenerateDatasetSchemas generates dataset schemas for multiple tables
func (sc *SchemaController) GenerateDatasetSchemas(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GenerateDatasetSchemas"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GenerateDatasetSchemas", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request struct {
		Tables []string `json:"tables"`
		Schema string   `json:"schema"`
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Generating dataset schemas for %d tables, schema: %s", len(request.Tables), request.Schema))

	schemas := make(map[string]interface{})
	for _, tableName := range request.Tables {
		schema, err := sc.generateDatasetSchemaForTable(tableName, request.Schema, &iLog)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error generating schema for table %s: %v", tableName, err))
			continue
		}
		schemas[tableName] = schema
	}

	c.JSON(http.StatusOK, schemas)
}

// GetDatabaseAliases returns all available database connection aliases
func (sc *SchemaController) GetDatabaseAliases(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GetDatabaseAliases"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GetDatabaseAliases", elapsed)
	}()

	_, user, clientid, err := common.GetRequestUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Get user information Error: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	iLog.Debug("Fetching database aliases")

	// Access the database cache from orm package
	aliases := sc.getDatabaseAliases(&iLog)

	iLog.Debug(fmt.Sprintf("Found %d database aliases", len(aliases)))
	c.JSON(http.StatusOK, aliases)
}

// GetDatabaseTables returns all tables for a specific database alias
func (sc *SchemaController) GetDatabaseTables(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GetDatabaseTables"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GetDatabaseTables", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request struct {
		Alias  string `json:"alias"`
		Schema string `json:"schema,omitempty"`
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Alias == "" {
		request.Alias = "default"
	}

	iLog.Debug(fmt.Sprintf("Fetching tables for database alias: %s, schema: %s", request.Alias, request.Schema))

	tables, err := sc.getAllTableNamesWithAlias(request.Alias, request.Schema, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting tables: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Found %d tables", len(tables)))
	c.JSON(http.StatusOK, tables)
}

// GetDatabaseColumns returns columns for a specific table in a database alias
func (sc *SchemaController) GetDatabaseColumns(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GetDatabaseColumns"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GetDatabaseColumns", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request struct {
		Alias     string `json:"alias"`
		TableName string `json:"tableName"`
		Schema    string `json:"schema,omitempty"`
	}

	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Alias == "" {
		request.Alias = "default"
	}

	if request.TableName == "" {
		iLog.Error("Table name is required")
		c.JSON(http.StatusBadRequest, gin.H{"error": "tableName is required"})
		return
	}

	iLog.Debug(fmt.Sprintf("Fetching columns for table: %s, alias: %s, schema: %s", request.TableName, request.Alias, request.Schema))

	table, err := sc.getTableStructureWithAlias(request.Alias, request.TableName, request.Schema, &iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting table structure: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	iLog.Debug(fmt.Sprintf("Found %d columns for table %s", len(table.Columns), request.TableName))
	c.JSON(http.StatusOK, table.Columns)
}
