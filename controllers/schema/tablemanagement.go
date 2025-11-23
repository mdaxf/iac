package schema

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mdaxf/iac/controllers/common"
	"github.com/mdaxf/iac/logger"
)

// CreateTableRequest represents a request to create a new table
type CreateTableRequest struct {
	Alias       string      `json:"alias"`
	TableName   string      `json:"tableName" binding:"required"`
	Schema      string      `json:"schema,omitempty"`
	Columns     []ColumnDef `json:"columns" binding:"required,min=1"`
	Comment     *string     `json:"comment,omitempty"`
	Indexes     []IndexDef  `json:"indexes,omitempty"`
	Constraints []Constraint `json:"constraints,omitempty"`
}

// AlterTableRequest represents a request to alter an existing table
type AlterTableRequest struct {
	Alias      string              `json:"alias"`
	TableName  string              `json:"tableName" binding:"required"`
	Schema     string              `json:"schema,omitempty"`
	AddColumns []ColumnDef         `json:"addColumns,omitempty"`
	ModifyColumns []ColumnDef      `json:"modifyColumns,omitempty"`
	DropColumns []string           `json:"dropColumns,omitempty"`
	RenameColumns map[string]string `json:"renameColumns,omitempty"`
	AddIndexes  []IndexDef         `json:"addIndexes,omitempty"`
	DropIndexes []string           `json:"dropIndexes,omitempty"`
}

// IndexDef represents an index definition
type IndexDef struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type,omitempty"` // BTREE, HASH, etc.
}

// Constraint represents a table constraint
type Constraint struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"` // PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK
	Columns           []string `json:"columns"`
	ReferencedTable   string   `json:"referencedTable,omitempty"`
	ReferencedColumns []string `json:"referencedColumns,omitempty"`
	OnDelete          string   `json:"onDelete,omitempty"` // CASCADE, SET NULL, etc.
	OnUpdate          string   `json:"onUpdate,omitempty"`
}

// CreateTable creates a new database table
func (sc *SchemaController) CreateTable(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.CreateTable"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.CreateTable", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request CreateTableRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Alias == "" {
		request.Alias = "default"
	}

	iLog.Debug(fmt.Sprintf("Creating table: %s on alias: %s", request.TableName, request.Alias))

	// Get database type for proper SQL syntax
	dbType := sc.getDatabaseTypeForAlias(request.Alias, &iLog)

	// Build CREATE TABLE SQL
	sql := sc.buildCreateTableSQL(request, dbType)
	iLog.Debug(fmt.Sprintf("CREATE TABLE SQL: %s", sql))

	// Execute the SQL
	db := sc.getDBForAlias(request.Alias, &iLog)
	if db == nil {
		iLog.Error("Failed to get database connection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}

	_, err = db.Exec(sql)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error creating table: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create table: %v", err)})
		return
	}

	iLog.Debug(fmt.Sprintf("Table %s created successfully", request.TableName))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Table %s created successfully", request.TableName),
	})
}

// AlterTable modifies an existing database table
func (sc *SchemaController) AlterTable(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.AlterTable"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.AlterTable", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request AlterTableRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error unmarshaling request: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Alias == "" {
		request.Alias = "default"
	}

	iLog.Debug(fmt.Sprintf("Altering table: %s on alias: %s", request.TableName, request.Alias))

	// Get database type for proper SQL syntax
	dbType := sc.getDatabaseTypeForAlias(request.Alias, &iLog)

	// Build ALTER TABLE SQL statements
	sqlStatements := sc.buildAlterTableSQL(request, dbType)

	// Execute the SQL statements
	db := sc.getDBForAlias(request.Alias, &iLog)
	if db == nil {
		iLog.Error("Failed to get database connection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}

	for _, sql := range sqlStatements {
		iLog.Debug(fmt.Sprintf("ALTER TABLE SQL: %s", sql))
		_, err = db.Exec(sql)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error altering table: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to alter table: %v", err)})
			return
		}
	}

	iLog.Debug(fmt.Sprintf("Table %s altered successfully", request.TableName))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Table %s altered successfully", request.TableName),
	})
}

// DropTable drops a database table
func (sc *SchemaController) DropTable(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.DropTable"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.DropTable", elapsed)
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
		TableName string `json:"tableName" binding:"required"`
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

	iLog.Debug(fmt.Sprintf("Dropping table: %s on alias: %s", request.TableName, request.Alias))

	// Build DROP TABLE SQL
	fullTableName := request.TableName
	if request.Schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", request.Schema, request.TableName)
	}

	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", fullTableName)
	iLog.Debug(fmt.Sprintf("DROP TABLE SQL: %s", sql))

	// Execute the SQL
	db := sc.getDBForAlias(request.Alias, &iLog)
	if db == nil {
		iLog.Error("Failed to get database connection")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database connection"})
		return
	}

	_, err = db.Exec(sql)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error dropping table: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to drop table: %v", err)})
		return
	}

	iLog.Debug(fmt.Sprintf("Table %s dropped successfully", request.TableName))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Table %s dropped successfully", request.TableName),
	})
}

// buildCreateTableSQL builds CREATE TABLE SQL statement
func (sc *SchemaController) buildCreateTableSQL(request CreateTableRequest, dbType string) string {
	var sb strings.Builder

	// Table name with schema
	fullTableName := request.TableName
	if request.Schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", request.Schema, request.TableName)
	}

	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", fullTableName))

	// Add columns
	columnDefs := make([]string, 0, len(request.Columns))
	for _, col := range request.Columns {
		columnDefs = append(columnDefs, sc.buildColumnDefinition(col, dbType))
	}
	sb.WriteString("  " + strings.Join(columnDefs, ",\n  "))

	// Add constraints
	if len(request.Constraints) > 0 {
		sb.WriteString(",\n")
		constraintDefs := make([]string, 0, len(request.Constraints))
		for _, constraint := range request.Constraints {
			constraintDefs = append(constraintDefs, sc.buildConstraintDefinition(constraint, dbType))
		}
		sb.WriteString("  " + strings.Join(constraintDefs, ",\n  "))
	}

	sb.WriteString("\n)")

	// Add table comment for MySQL
	if request.Comment != nil && *request.Comment != "" && dbType == "mysql" {
		sb.WriteString(fmt.Sprintf(" COMMENT='%s'", strings.ReplaceAll(*request.Comment, "'", "''")))
	}

	return sb.String()
}

// buildColumnDefinition builds a column definition string
func (sc *SchemaController) buildColumnDefinition(col ColumnDef, dbType string) string {
	var parts []string

	// Column name
	parts = append(parts, col.Name)

	// Data type with length/precision
	dataType := col.DataType
	if col.Length != nil && *col.Length > 0 {
		dataType = fmt.Sprintf("%s(%d)", dataType, *col.Length)
	} else if col.Precision != nil && *col.Precision > 0 {
		if col.Scale != nil && *col.Scale > 0 {
			dataType = fmt.Sprintf("%s(%d,%d)", dataType, *col.Precision, *col.Scale)
		} else {
			dataType = fmt.Sprintf("%s(%d)", dataType, *col.Precision)
		}
	}
	parts = append(parts, dataType)

	// NULL/NOT NULL
	if !col.IsNullable {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}

	// Auto increment
	if col.IsAutoIncrement {
		if dbType == "postgres" {
			// PostgreSQL uses SERIAL or GENERATED
			parts = append(parts, "GENERATED ALWAYS AS IDENTITY")
		} else {
			parts = append(parts, "AUTO_INCREMENT")
		}
	}

	// Default value
	if col.DefaultValue != nil && *col.DefaultValue != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", *col.DefaultValue))
	}

	// Comment
	if col.Comment != nil && *col.Comment != "" && dbType == "mysql" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", strings.ReplaceAll(*col.Comment, "'", "''")))
	}

	return strings.Join(parts, " ")
}

// buildConstraintDefinition builds a constraint definition string
func (sc *SchemaController) buildConstraintDefinition(constraint Constraint, dbType string) string {
	var sb strings.Builder

	if constraint.Name != "" {
		sb.WriteString(fmt.Sprintf("CONSTRAINT %s ", constraint.Name))
	}

	switch strings.ToUpper(constraint.Type) {
	case "PRIMARY KEY":
		sb.WriteString(fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(constraint.Columns, ", ")))
	case "FOREIGN KEY":
		sb.WriteString(fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s (%s)",
			strings.Join(constraint.Columns, ", "),
			constraint.ReferencedTable,
			strings.Join(constraint.ReferencedColumns, ", ")))
		if constraint.OnDelete != "" {
			sb.WriteString(fmt.Sprintf(" ON DELETE %s", constraint.OnDelete))
		}
		if constraint.OnUpdate != "" {
			sb.WriteString(fmt.Sprintf(" ON UPDATE %s", constraint.OnUpdate))
		}
	case "UNIQUE":
		sb.WriteString(fmt.Sprintf("UNIQUE (%s)", strings.Join(constraint.Columns, ", ")))
	case "CHECK":
		// CHECK constraints need special handling
		sb.WriteString(fmt.Sprintf("CHECK (%s)", constraint.Columns[0]))
	}

	return sb.String()
}

// buildAlterTableSQL builds ALTER TABLE SQL statements
func (sc *SchemaController) buildAlterTableSQL(request AlterTableRequest, dbType string) []string {
	var statements []string

	fullTableName := request.TableName
	if request.Schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", request.Schema, request.TableName)
	}

	// Add columns
	for _, col := range request.AddColumns {
		sql := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s",
			fullTableName,
			sc.buildColumnDefinition(col, dbType))
		statements = append(statements, sql)
	}

	// Modify columns
	for _, col := range request.ModifyColumns {
		if dbType == "postgres" {
			// PostgreSQL requires multiple ALTER TABLE statements
			sql := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s",
				fullTableName,
				col.Name,
				col.DataType)
			statements = append(statements, sql)
		} else {
			sql := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s",
				fullTableName,
				sc.buildColumnDefinition(col, dbType))
			statements = append(statements, sql)
		}
	}

	// Drop columns
	for _, colName := range request.DropColumns {
		sql := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", fullTableName, colName)
		statements = append(statements, sql)
	}

	// Rename columns
	for oldName, newName := range request.RenameColumns {
		if dbType == "postgres" {
			sql := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
				fullTableName, oldName, newName)
			statements = append(statements, sql)
		} else {
			sql := fmt.Sprintf("ALTER TABLE %s CHANGE COLUMN %s %s",
				fullTableName, oldName, newName)
			statements = append(statements, sql)
		}
	}

	// Add indexes
	for _, idx := range request.AddIndexes {
		uniqueStr := ""
		if idx.Unique {
			uniqueStr = "UNIQUE "
		}
		sql := fmt.Sprintf("CREATE %sINDEX %s ON %s (%s)",
			uniqueStr,
			idx.Name,
			fullTableName,
			strings.Join(idx.Columns, ", "))
		statements = append(statements, sql)
	}

	// Drop indexes
	for _, idxName := range request.DropIndexes {
		if dbType == "postgres" {
			sql := fmt.Sprintf("DROP INDEX %s", idxName)
			statements = append(statements, sql)
		} else {
			sql := fmt.Sprintf("DROP INDEX %s ON %s", idxName, fullTableName)
			statements = append(statements, sql)
		}
	}

	return statements
}
