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

// GORMGenerationRequest represents a request to generate GORM struct
type GORMGenerationRequest struct {
	Alias      string       `json:"alias"`
	TableName  string       `json:"tableName"`
	Schema     string       `json:"schema,omitempty"`
	Columns    []ColumnDef  `json:"columns,omitempty"` // If provided, use these instead of DB
	StructName string       `json:"structName,omitempty"`
}

// ColumnDef represents a column definition for GORM generation
type ColumnDef struct {
	Name            string  `json:"name"`
	DataType        string  `json:"data_type"`
	Length          *int    `json:"length,omitempty"`
	Precision       *int    `json:"precision,omitempty"`
	Scale           *int    `json:"scale,omitempty"`
	IsNullable      bool    `json:"is_nullable"`
	IsPrimaryKey    bool    `json:"is_primary_key"`
	IsUnique        bool    `json:"is_unique"`
	IsAutoIncrement bool    `json:"is_auto_increment"`
	DefaultValue    *string `json:"default_value,omitempty"`
	Comment         *string `json:"comment,omitempty"`
}

// GORMGenerationResponse represents the GORM struct generation response
type GORMGenerationResponse struct {
	StructCode string `json:"struct_code"`
	StructName string `json:"struct_name"`
	Package    string `json:"package"`
}

// GenerateGORMStruct generates a GORM struct from table definition
func (sc *SchemaController) GenerateGORMStruct(c *gin.Context) {
	iLog := logger.Log{ModuleName: logger.API, User: "System", ControllerName: "schema.GenerateGORMStruct"}
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		iLog.PerformanceWithDuration("controllers.schema.GenerateGORMStruct", elapsed)
	}()

	body, clientid, user, err := common.GetRequestBodyandUser(c)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error reading body: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	iLog.ClientID = clientid
	iLog.User = user

	var request GORMGenerationRequest
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

	iLog.Debug(fmt.Sprintf("Generating GORM struct for table: %s, alias: %s", request.TableName, request.Alias))

	var columns []ColumnDef

	// If columns provided in request, use them; otherwise fetch from DB
	if len(request.Columns) > 0 {
		columns = request.Columns
	} else {
		// Fetch table structure from database
		table, err := sc.getTableStructureWithAlias(request.Alias, request.TableName, request.Schema, &iLog)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error getting table structure: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Convert DBColumn to ColumnDef
		for _, col := range table.Columns {
			colDef := ColumnDef{
				Name:            col.Name,
				DataType:        col.Type,
				IsNullable:      col.Nullable,
				IsPrimaryKey:    col.PrimaryKey,
				IsAutoIncrement: strings.Contains(strings.ToLower(col.Extra), "auto_increment"),
				DefaultValue:    col.DefaultValue,
				Comment:         col.Comment,
			}
			columns = append(columns, colDef)
		}
	}

	// Get database type for proper type mapping
	dbType := sc.getDatabaseTypeForAlias(request.Alias, &iLog)

	// Generate GORM struct
	structName := request.StructName
	if structName == "" {
		structName = toPascalCase(request.TableName)
	}

	structCode := sc.buildGORMStruct(structName, request.TableName, columns, dbType)

	response := GORMGenerationResponse{
		StructCode: structCode,
		StructName: structName,
		Package:    "models",
	}

	iLog.Debug(fmt.Sprintf("Generated GORM struct for %s", request.TableName))
	c.JSON(http.StatusOK, response)
}

// buildGORMStruct constructs the GORM struct code
func (sc *SchemaController) buildGORMStruct(structName, tableName string, columns []ColumnDef, dbType string) string {
	var sb strings.Builder

	// Add package comment
	sb.WriteString(fmt.Sprintf("// %s represents the %s table\n", structName, tableName))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	// Add columns
	for _, col := range columns {
		fieldName := toPascalCase(col.Name)
		goType := sc.sqlTypeToGoType(col.DataType, col.IsNullable, dbType)
		tags := sc.buildGORMTags(col)

		// Add comment if exists
		if col.Comment != nil && *col.Comment != "" {
			sb.WriteString(fmt.Sprintf("\t// %s\n", *col.Comment))
		}

		sb.WriteString(fmt.Sprintf("\t%s %s `%s`\n", fieldName, goType, tags))
	}

	sb.WriteString("}\n\n")

	// Add TableName method
	sb.WriteString(fmt.Sprintf("// TableName returns the table name for %s\n", structName))
	sb.WriteString(fmt.Sprintf("func (%s) TableName() string {\n", structName))
	sb.WriteString(fmt.Sprintf("\treturn \"%s\"\n", tableName))
	sb.WriteString("}\n")

	return sb.String()
}

// sqlTypeToGoType converts SQL data type to Go type
func (sc *SchemaController) sqlTypeToGoType(sqlType string, isNullable bool, dbType string) string {
	typeUpper := strings.ToUpper(sqlType)
	var goType string

	// Handle different SQL types
	if strings.Contains(typeUpper, "INT") {
		if strings.Contains(typeUpper, "TINYINT(1)") || strings.Contains(typeUpper, "BOOL") {
			goType = "bool"
		} else if strings.Contains(typeUpper, "BIGINT") {
			goType = "int64"
		} else {
			goType = "int"
		}
	} else if strings.Contains(typeUpper, "FLOAT") || strings.Contains(typeUpper, "DOUBLE") || strings.Contains(typeUpper, "REAL") {
		goType = "float64"
	} else if strings.Contains(typeUpper, "DECIMAL") || strings.Contains(typeUpper, "NUMERIC") {
		goType = "float64"
	} else if strings.Contains(typeUpper, "BOOL") || strings.Contains(typeUpper, "BIT") {
		goType = "bool"
	} else if strings.Contains(typeUpper, "TIME") || strings.Contains(typeUpper, "DATE") {
		goType = "time.Time"
	} else if strings.Contains(typeUpper, "BLOB") || strings.Contains(typeUpper, "BINARY") || strings.Contains(typeUpper, "BYTEA") {
		goType = "[]byte"
	} else if strings.Contains(typeUpper, "JSON") {
		goType = "datatypes.JSON" // GORM JSON type
	} else {
		goType = "string"
	}

	// Add pointer for nullable fields (except time.Time which GORM handles specially)
	if isNullable && goType != "time.Time" {
		return "*" + goType
	}

	return goType
}

// buildGORMTags constructs GORM struct tags
func (sc *SchemaController) buildGORMTags(col ColumnDef) string {
	var gormTags []string
	var jsonTag string

	// Column name
	gormTags = append(gormTags, fmt.Sprintf("column:%s", col.Name))

	// Primary key
	if col.IsPrimaryKey {
		gormTags = append(gormTags, "primaryKey")
	}

	// Data type with length/precision
	typeTag := fmt.Sprintf("type:%s", col.DataType)
	if col.Length != nil && *col.Length > 0 {
		typeTag = fmt.Sprintf("type:%s(%d)", col.DataType, *col.Length)
	} else if col.Precision != nil && *col.Precision > 0 {
		if col.Scale != nil && *col.Scale > 0 {
			typeTag = fmt.Sprintf("type:%s(%d,%d)", col.DataType, *col.Precision, *col.Scale)
		} else {
			typeTag = fmt.Sprintf("type:%s(%d)", col.DataType, *col.Precision)
		}
	}
	gormTags = append(gormTags, typeTag)

	// Not null
	if !col.IsNullable {
		gormTags = append(gormTags, "not null")
	}

	// Unique
	if col.IsUnique {
		gormTags = append(gormTags, "unique")
	}

	// Auto increment
	if col.IsAutoIncrement {
		gormTags = append(gormTags, "autoIncrement")
	}

	// Default value
	if col.DefaultValue != nil && *col.DefaultValue != "" {
		gormTags = append(gormTags, fmt.Sprintf("default:%s", *col.DefaultValue))
	}

	// Comment
	if col.Comment != nil && *col.Comment != "" {
		// Escape quotes in comment
		comment := strings.ReplaceAll(*col.Comment, "\"", "\\\"")
		gormTags = append(gormTags, fmt.Sprintf("comment:%s", comment))
	}

	// JSON tag
	jsonTag = fmt.Sprintf("json:\"%s\"", col.Name)

	return fmt.Sprintf("gorm:\"%s\" %s", strings.Join(gormTags, ";"), jsonTag)
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// getDatabaseTypeForAlias gets the database type (mysql, postgres, mssql, oracle) for an alias
func (sc *SchemaController) getDatabaseTypeForAlias(alias string, iLog *logger.Log) string {
	db := sc.getDBForAlias(alias, iLog)
	if db == nil {
		return "mysql" // default
	}

	// Try to detect database type from driver name
	// This is a simple detection, you may need to adjust based on your setup
	var driverName string
	err := db.QueryRow("SELECT VERSION()").Scan(&driverName)
	if err == nil {
		lowerDriver := strings.ToLower(driverName)
		if strings.Contains(lowerDriver, "postgres") {
			return "postgres"
		} else if strings.Contains(lowerDriver, "mysql") || strings.Contains(lowerDriver, "mariadb") {
			return "mysql"
		} else if strings.Contains(lowerDriver, "microsoft") || strings.Contains(lowerDriver, "mssql") {
			return "mssql"
		} else if strings.Contains(lowerDriver, "oracle") {
			return "oracle"
		}
	}

	// Fallback: try database-specific queries
	var testVal string

	// Test for PostgreSQL
	err = db.QueryRow("SELECT current_schema()").Scan(&testVal)
	if err == nil {
		return "postgres"
	}

	// Test for MySQL
	err = db.QueryRow("SELECT DATABASE()").Scan(&testVal)
	if err == nil {
		return "mysql"
	}

	return "mysql" // default fallback
}
