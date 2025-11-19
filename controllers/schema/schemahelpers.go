package schema

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mdaxf/iac/config"
	dbconn "github.com/mdaxf/iac/databases"
	"github.com/mdaxf/iac/databases/orm"
	"github.com/mdaxf/iac/logger"
)

// getCurrentSchema gets the current database schema/database name
// For MySQL: uses DATABASE(), for PostgreSQL: uses current_schema()
func (sc *SchemaController) getCurrentSchema(iLog *logger.Log) string {
	db := dbconn.DB
	if db == nil {
		return "public"
	}

	// Try MySQL first (DATABASE())
	var schemaName string
	err := db.QueryRow("SELECT DATABASE()").Scan(&schemaName)
	if err != nil {
		// If MySQL fails, try PostgreSQL (current_schema())
		err = db.QueryRow("SELECT current_schema()").Scan(&schemaName)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error getting current schema: %v, using 'public'", err))
			return "public"
		}
	}

	if schemaName == "" {
		return "public"
	}

	return schemaName
}

// getTableStructure retrieves the structure of a table
func (sc *SchemaController) getTableStructure(tableName string, schemaName string, iLog *logger.Log) (*DBTable, error) {
	db := dbconn.DB
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// If no schema provided, get current schema
	if schemaName == "" {
		schemaName = sc.getCurrentSchema(iLog)
	}

	table := &DBTable{
		Name:   tableName,
		Schema: schemaName,
	}

	// Get columns - using ? placeholders for MySQL/PostgreSQL compatibility
	columnsQuery := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			coalesce(c.extra,'') as extra
		FROM information_schema.columns c
		WHERE c.table_name = ?
		AND c.table_schema = ?
		ORDER BY c.ordinal_position
	`

	rows, err := db.Query(columnsQuery, tableName, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error querying columns: %v", err)
	}
	defer rows.Close()

	var columns []DBColumn
	for rows.Next() {
		var col DBColumn
		var dataType string
		var isNullable string
		var defaultValue, charMaxLength, numericPrecision, numericScale interface{}
		var extra string

		err := rows.Scan(&col.Name, &dataType, &isNullable, &defaultValue, &charMaxLength, &numericPrecision, &numericScale, &extra)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error scanning column: %v", err))
			continue
		}

		// Build complete type string
		col.Type = sc.formatDataType(dataType, charMaxLength, numericPrecision, numericScale)
		col.Nullable = isNullable == "YES"
		col.Extra = extra

		if defaultValue != nil {
			defStr := fmt.Sprintf("%v", defaultValue)
			col.DefaultValue = &defStr
		}

		columns = append(columns, col)
	}

	// Get primary keys
	pkQuery := `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.table_name = ?
			AND tc.table_schema = ?
			AND tc.constraint_type = 'PRIMARY KEY'
	`

	pkRows, err := db.Query(pkQuery, tableName, schemaName)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error querying primary keys: %v", err))
	} else {
		defer pkRows.Close()
		primaryKeys := make(map[string]bool)
		for pkRows.Next() {
			var colName string
			if err := pkRows.Scan(&colName); err == nil {
				primaryKeys[colName] = true
			}
		}

		// Mark primary key columns
		for i := range columns {
			if primaryKeys[columns[i].Name] {
				columns[i].PrimaryKey = true
			}
		}
	}

	// Get foreign keys
	fkQuery := `
		SELECT
			kcu.column_name,
			kcu.referenced_table_name AS foreign_table_name,
			kcu.referenced_column_name AS foreign_column_name
		FROM information_schema.key_column_usage AS kcu
		WHERE kcu.table_name = ?
			AND kcu.table_schema = ?
			AND kcu.referenced_table_name IS NOT NULL
	`

	fkRows, err := db.Query(fkQuery, tableName, schemaName)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error querying foreign keys: %v", err))
	} else {
		defer fkRows.Close()
		foreignKeys := make(map[string]*ForeignKey)
		for fkRows.Next() {
			var colName, foreignTable, foreignColumn string
			if err := fkRows.Scan(&colName, &foreignTable, &foreignColumn); err == nil {
				foreignKeys[colName] = &ForeignKey{
					Table:  foreignTable,
					Column: foreignColumn,
				}
			}
		}

		// Mark foreign key columns
		for i := range columns {
			if fk, exists := foreignKeys[columns[i].Name]; exists {
				columns[i].ForeignKey = fk
			}
		}
	}

	table.Columns = columns
	return table, nil
}

// formatDataType formats the data type with precision/length
func (sc *SchemaController) formatDataType(dataType string, charMaxLength, numericPrecision, numericScale interface{}) string {
	switch dataType {
	case "character varying", "varchar":
		if charMaxLength != nil {
			return fmt.Sprintf("varchar(%v)", charMaxLength)
		}
		return "varchar"
	case "character", "char":
		if charMaxLength != nil {
			return fmt.Sprintf("char(%v)", charMaxLength)
		}
		return "char"
	case "numeric", "decimal":
		if numericPrecision != nil && numericScale != nil {
			return fmt.Sprintf("numeric(%v,%v)", numericPrecision, numericScale)
		}
		return "numeric"
	default:
		return dataType
	}
}

// getAllTableNames retrieves all table names
func (sc *SchemaController) getAllTableNames(schemaName string, iLog *logger.Log) ([]string, error) {
	db := dbconn.DB
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// If no schema provided, get current schema
	if schemaName == "" {
		schemaName = sc.getCurrentSchema(iLog)
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := db.Query(query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error querying tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// getTablesWithChildren returns tables and their related children via foreign keys
func (sc *SchemaController) getTablesWithChildren(tableNames []string, schemaName string, iLog *logger.Log) ([]string, error) {
	db := dbconn.DB
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// If no schema provided, get current schema
	if schemaName == "" {
		schemaName = sc.getCurrentSchema(iLog)
	}

	// Use a map to track unique tables
	tablesMap := make(map[string]bool)
	for _, t := range tableNames {
		tablesMap[t] = true
	}

	// Build IN clause for MySQL compatibility
	placeholders := make([]string, len(tableNames))
	args := make([]interface{}, 0)

	for i := range tableNames {
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	// Add arguments for first CASE IN clause
	for _, t := range tableNames {
		args = append(args, t)
	}
	// Add arguments for second CASE IN clause
	for _, t := range tableNames {
		args = append(args, t)
	}
	// Add schema for WHERE clause
	args = append(args, schemaName)
	// Add arguments for third IN clause (table_name)
	for _, t := range tableNames {
		args = append(args, t)
	}
	// Add arguments for fourth IN clause (referenced_table_name)
	for _, t := range tableNames {
		args = append(args, t)
	}

	// Query to find related tables (both directions of FK relationships)
	// Simplified for MySQL using key_column_usage table
	query := fmt.Sprintf(`
		SELECT DISTINCT
			CASE
				WHEN kcu.table_name IN (%s) THEN kcu.referenced_table_name
				WHEN kcu.referenced_table_name IN (%s) THEN kcu.table_name
			END as related_table
		FROM information_schema.key_column_usage AS kcu
		WHERE kcu.referenced_table_name IS NOT NULL
			AND kcu.table_schema = ?
			AND (kcu.table_name IN (%s) OR kcu.referenced_table_name IN (%s))
	`, inClause, inClause, inClause, inClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error querying related tables: %v", err))
		return tableNames, nil // Return original tables if query fails
	}
	defer rows.Close()

	for rows.Next() {
		var relatedTable string
		if err := rows.Scan(&relatedTable); err == nil && relatedTable != "" {
			tablesMap[relatedTable] = true
		}
	}

	// Convert map back to slice
	result := make([]string, 0, len(tablesMap))
	for table := range tablesMap {
		result = append(result, table)
	}

	iLog.Debug(fmt.Sprintf("Expanded %d tables to %d with children", len(tableNames), len(result)))
	return result, nil
}

// getRelationships retrieves foreign key relationships between tables
func (sc *SchemaController) getRelationships(tableNames []string, schemaName string, iLog *logger.Log) ([]DBRelationship, error) {
	if len(tableNames) == 0 {
		return []DBRelationship{}, nil
	}

	db := dbconn.DB
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	// If no schema provided, get current schema
	if schemaName == "" {
		schemaName = sc.getCurrentSchema(iLog)
	}

	// Build IN clause for MySQL compatibility
	placeholders := make([]string, len(tableNames))
	args := make([]interface{}, 0)

	for i := range tableNames {
		placeholders[i] = "?"
	}
	inClause := strings.Join(placeholders, ",")

	// Add arguments for first IN clause (table_name)
	for _, tableName := range tableNames {
		args = append(args, tableName)
	}
	// Add schema parameter
	args = append(args, schemaName)
	// Add arguments for second IN clause (referenced_table_name)
	for _, tableName := range tableNames {
		args = append(args, tableName)
	}

	// Query to find relationships in BOTH directions
	// - Where source table is in the list (table_name IN)
	// - Where target table is in the list (referenced_table_name IN)
	query := fmt.Sprintf(`
		SELECT DISTINCT
			kcu.table_name AS source_table,
			kcu.column_name AS source_column,
			kcu.referenced_table_name AS target_table,
			kcu.referenced_column_name AS target_column,
			kcu.constraint_name
		FROM information_schema.key_column_usage kcu
		WHERE kcu.referenced_table_name IS NOT NULL
			AND (kcu.table_schema = database() OR kcu.table_schema =?)
			AND (kcu.table_name IN (%s) OR kcu.referenced_table_name IN (%s))
		ORDER BY kcu.table_name, kcu.column_name
	`, inClause, inClause)

	iLog.Debug(fmt.Sprintf("Querying relationships for %d tables in schema '%s'", len(tableNames), schemaName))
	iLog.Debug(fmt.Sprintf("Tables: %v", tableNames))
	iLog.Debug(fmt.Sprintf("Query: %s", query))
	iLog.Debug(fmt.Sprintf("Args: %v", args))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying relationships: %v", err)
	}
	defer rows.Close()

	// First, let's check if there are ANY foreign keys in this schema
	var testCount int
	testQuery := "SELECT COUNT(*) FROM information_schema.key_column_usage WHERE table_schema = ? AND referenced_table_name IS NOT NULL"
	db.QueryRow(testQuery, schemaName).Scan(&testCount)
	iLog.Debug(fmt.Sprintf("Total FK constraints in schema '%s': %d", schemaName, testCount))

	var relationships []DBRelationship
	for rows.Next() {
		var sourceTable, sourceColumn, targetTable, targetColumn, constraintName string
		err := rows.Scan(&sourceTable, &sourceColumn, &targetTable, &targetColumn, &constraintName)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error scanning relationship: %v", err))
			continue
		}

		rel := DBRelationship{
			ID:           fmt.Sprintf("rel-%s", constraintName),
			SourceTable:  sourceTable,
			SourceColumn: sourceColumn,
			TargetTable:  targetTable,
			TargetColumn: targetColumn,
			Type:         "N:1", // Most foreign keys are many-to-one
		}

		iLog.Debug(fmt.Sprintf("Found relationship: %s.%s -> %s.%s", sourceTable, sourceColumn, targetTable, targetColumn))
		relationships = append(relationships, rel)
	}

	iLog.Debug(fmt.Sprintf("Found %d relationships", len(relationships)))
	return relationships, nil
}

// calculatePositions calculates simple grid positions for tables
func (sc *SchemaController) calculatePositions(tables []DBTable) []TablePosition {
	positions := make([]TablePosition, len(tables))
	columns := 5 // Tables per row
	spacing := 350.0

	for i, table := range tables {
		x := float64(i%columns) * spacing
		y := float64(i/columns) * spacing
		positions[i] = TablePosition{
			Table: table.Name,
			X:     x,
			Y:     y,
		}
	}

	return positions
}

// generateDetailPage generates the detailpage structure based on table relationships
func (sc *SchemaController) generateDetailPage(tableName string, schemaName string, table *DBTable, keyField string, iLog *logger.Log) (map[string]interface{}, error) {
	detailPage := make(map[string]interface{})
	tabs := make(map[string]interface{})
	tabIndex := 0

	// Always add General tab
	tabs["General"] = map[string]interface{}{
		"lng": map[string]interface{}{
			"code":    "General",
			"default": "General",
		},
		"index": tabIndex,
	}
	tabIndex++

	// Build General tab content with form fields
	generalFields := []interface{}{}
	systemFields := []string{"createdby", "createdon", "modifiedby", "modifiedon"}
	hasSystemFields := false
	excludedfields := []string{"metadata", "extensionid", "referenceid", "rowversionstamp", "uuid"}
	hasextensionid := false

	for _, col := range table.Columns {
		isSystemField := false
		isExcludedfield := false
		for _, sf := range systemFields {
			if col.Name == sf {
				isSystemField = true
				hasSystemFields = true
				break
			}
		}
		for _, sf := range excludedfields {
			if col.Name == sf {
				isExcludedfield = true
				break
			}
		}

		if col.Name == "extensionid" {
			hasextensionid = true
		}
		if !isExcludedfield && !isSystemField && col.Name != keyField {
			fieldstr := col.Name

			if col.Extra == "auto_increment" {
				continue
			}
			if col.Name == "lngcodeid" {
				data := map[string]interface{}{
					"lngcodeid": map[string]interface{}{
						"translation": "",
						"fields":      []string{"shorttext", "mediumtext_"},
					},
				}
				generalFields = append(generalFields, data)
				generalFields = append(generalFields, "dummy")
				continue
			}

			if col.ForeignKey != nil {

				data := map[string]interface{}{}
				item := map[string]interface{}{}
				item["link"] = col.ForeignKey.Table
				item["schema"] = col.ForeignKey.Table
				item["field"] = "name"
				item["keyfield"] = col.ForeignKey.Column
				data[fieldstr] = item
				generalFields = append(generalFields, data)
				continue
			}

			generalFields = append(generalFields, fieldstr)
		}
	}

	// Create General tab content
	detailPage["General"] = map[string]interface{}{
		"tables": []map[string]interface{}{
			{
				"cols":   2,
				"fields": generalFields,
			},
		},
	}

	// Get relationships FROM this table (foreign keys in this table pointing to other tables)
	outgoingRels := make(map[string]*ForeignKey)
	for _, col := range table.Columns {
		if col.ForeignKey != nil {
			outgoingRels[col.Name] = col.ForeignKey
		}
	}

	// Get relationships TO this table (other tables pointing to this table)
	incomingRels, err := sc.getIncomingRelationships(tableName, schemaName, iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error getting incoming relationships: %v", err))
	}

	// Add tabs for incoming relationships (sublink type)
	for _, rel := range incomingRels {
		tabName := sc.toPascalCase(rel.SourceTable)
		tabs[tabName] = map[string]interface{}{
			"available": map[string]interface{}{
				keyField: map[string]interface{}{">": 0},
			},
			"lng": map[string]interface{}{
				"code":    tabName,
				"default": sc.toTitleCase(rel.SourceTable),
			},
			"index": tabIndex,
		}
		tabIndex++

		// Create sublink content
		detailPage[tabName] = map[string]interface{}{
			"type": "subitem",
			"linkfields": []map[string]interface{}{
				{keyField: fmt.Sprintf("%s.%s", rel.SourceTable, rel.SourceColumn)},
			},
			"keyfield":       keyField,
			"tablename":      rel.SourceTable,
			"schema":         rel.SourceTable,
			"tablekeyfields": []string{rel.SourceColumn},
			"lng": map[string]interface{}{
				"code":    tabName,
				"default": sc.toTitleCase(rel.SourceTable),
			},
			"actions": []string{"Add", "Save", "Delete"},
		}
	}

	if hasextensionid {
		tabName := "Extensions"
		tabs[tabName] = map[string]interface{}{
			"available": map[string]interface{}{
				keyField: map[string]interface{}{">": 0},
			},
			"lng": map[string]interface{}{
				"code":    tabName,
				"default": "Extensions",
			},
			"index": tabIndex,
		}
		tabIndex++
		// Create extension content

		detailPage[tabName] = map[string]interface{}{
			"type": "extensionlink",
			"linkfields": []map[string]interface{}{
				{"extensionid": "extension_values.extensionid"},
			},
			"masterschema":      "extensions",
			"keyfield":          "extensionid",
			"targetkeyfield":    "extensionid",
			"tablename":         "extension_values",
			"maintable":         tableName,
			"maintablekeyfield": "id",
			"schema":            "extensionvalues",
			"lng": map[string]interface{}{
				"code":    "Extensions",
				"default": "Extensions",
			},
			"actions": []string{"Add", "Delete"},
		}

	}

	// Add System tab if system fields exist
	if hasSystemFields {
		tabs["System"] = map[string]interface{}{
			"lng": map[string]interface{}{
				"code":    "System",
				"default": "System",
			},
			"index": tabIndex,
		}

		detailPage["System"] = map[string]interface{}{
			"tables": []map[string]interface{}{
				{
					"cols":   2,
					"fields": systemFields,
				},
			},
		}
	}

	detailPage["tabs"] = tabs
	return detailPage, nil
}

// getIncomingRelationships finds all tables that have foreign keys pointing to this table
func (sc *SchemaController) getIncomingRelationships(tableName string, schemaName string, iLog *logger.Log) ([]DBRelationship, error) {
	db := dbconn.DB
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	if schemaName == "" {
		schemaName = sc.getCurrentSchema(iLog)
	}

	query := `
		SELECT DISTINCT
			kcu.table_name AS source_table,
			kcu.column_name AS source_column,
			kcu.referenced_table_name AS target_table,
			kcu.referenced_column_name AS target_column,
			kcu.constraint_name
		FROM information_schema.key_column_usage kcu
		WHERE kcu.referenced_table_name IS NOT NULL
			AND (kcu.table_schema = database() OR kcu.table_schema = ?)
			AND kcu.referenced_table_name = ?
		ORDER BY kcu.table_name, kcu.column_name
	`

	rows, err := db.Query(query, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("error querying incoming relationships: %v", err)
	}
	defer rows.Close()

	var relationships []DBRelationship
	for rows.Next() {
		var sourceTable, sourceColumn, targetTable, targetColumn, constraintName string
		err := rows.Scan(&sourceTable, &sourceColumn, &targetTable, &targetColumn, &constraintName)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error scanning relationship: %v", err))
			continue
		}

		rel := DBRelationship{
			ID:           fmt.Sprintf("rel-%s", constraintName),
			SourceTable:  sourceTable,
			SourceColumn: sourceColumn,
			TargetTable:  targetTable,
			TargetColumn: targetColumn,
			Type:         "N:1",
		}
		relationships = append(relationships, rel)
	}

	return relationships, nil
}

// generateDatasetSchemaForTable generates a dataset schema JSON for a table
func (sc *SchemaController) generateDatasetSchemaForTable(tableName string, schemaName string, iLog *logger.Log) (*DatasetSchema, error) {
	table, err := sc.getTableStructure(tableName, schemaName, iLog)
	if err != nil {
		return nil, err
	}

	// Find primary key
	keyField := "id"
	for _, col := range table.Columns {
		if col.PrimaryKey {
			keyField = col.Name
			break
		}
	}

	// Build list fields (non-hidden fields)
	var listFields []string
	var hiddenFields []string
	excludedfields := []string{"metadata", "referenceid", "rowversionstamp", "uuid"}
	isExcludedfield := false
	haslngcodeid := false
	for _, col := range table.Columns {
		if col.Name == "id" || strings.HasSuffix(col.Name, "id") {
			if col.Name != keyField {
				hiddenFields = append(hiddenFields, col.Name)
			}
		}

		for _, sf := range excludedfields {
			if col.Name == sf {
				isExcludedfield = true
				break
			}
		}

		if !isExcludedfield {
			listFields = append(listFields, col.Name)
		}

		if col.Name == "lngcodeid" {
			haslngcodeid = true
		}
	}

	// Add common fields if they exist
	commonFields := []string{"createdby", "createdon", "modifiedby", "modifiedon"}
	for _, field := range commonFields {
		for _, col := range table.Columns {
			if col.Name == field {
				listFields = append(listFields, field)
				break
			}
		}
	}

	// Build properties
	properties := make(map[string]interface{})
	var requiredFields []string

	for _, col := range table.Columns {
		propDef := PropertyDefinition{
			Type: sc.mapDBTypeToJSONType(col.Type),
			Lng: map[string]interface{}{
				"code":    sc.toPascalCase(col.Name),
				"default": sc.toTitleCase(col.Name),
			},
		}
		if col.ForeignKey != nil {
			if col.Type == "integer" {
				propDef.Nullvalue = "0"
			} else {
				propDef.Nullvalue = ""
			}
		}
		// Set format for datetime types
		if strings.Contains(strings.ToLower(col.Type), "timestamp") ||
			strings.Contains(strings.ToLower(col.Type), "date") {
			propDef.Format = "datetime"
		}

		// Set readonly for system fields
		systemFields := []string{"id", "createdby", "createdon", "modifiedby", "modifiedon"}
		for _, field := range systemFields {
			if col.Name == field {
				propDef.Readonly = true
				break
			}
		}

		properties[col.Name] = propDef

		// Add to required if not nullable and not a system field
		if !col.Nullable && col.Name != keyField && !propDef.Readonly {
			requiredFields = append(requiredFields, col.Name)
		}
	}

	if haslngcodeid {
		propDef := PropertyDefinition{
			Type: "string",
			Lng: map[string]interface{}{
				"code":    "ShortText",
				"default": "Short Description",
			},
			External: true,
		}
		properties["shorttext"] = propDef

		propDef = PropertyDefinition{
			Type: "string",
			Lng: map[string]interface{}{
				"code":    "MediumText",
				"default": "Description",
			},
			External: true,
		}
		properties["mediumtext_"] = propDef
	}

	// Build definition
	definition := map[string]interface{}{
		"type":                 "object",
		"additionalProperties": false,
		"properties":           properties,
		"required":             requiredFields,
		"title":                fmt.Sprintf("%s Maintenance", sc.toTitleCase(tableName)),
	}

	// Generate detailpage based on relationships
	detailPage, err := sc.generateDetailPage(tableName, schemaName, table, keyField, iLog)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error generating detail page: %v", err))
		detailPage = nil
	}

	schema := &DatasetSchema{
		Schema:         "http://json-schema.org/draft-06/schema#",
		Ref:            fmt.Sprintf("#/definitions/%s", sc.toPascalCase(tableName)),
		DatasourceType: "table",
		Datasource:     tableName,
		ListFields:     listFields,
		HiddenFields:   hiddenFields,
		KeyField:       keyField,
		DetailPage:     detailPage,
		Definitions: map[string]interface{}{
			sc.toPascalCase(tableName): definition,
		},
	}

	return schema, nil
}

// mapDBTypeToJSONType maps database types to JSON schema types
func (sc *SchemaController) mapDBTypeToJSONType(dbType string) string {
	dbType = strings.ToLower(dbType)

	switch {
	case strings.Contains(dbType, "tinyint"):
		return "boolean"
	case strings.Contains(dbType, "int"):
		return "integer"
	case strings.Contains(dbType, "numeric"), strings.Contains(dbType, "decimal"),
		strings.Contains(dbType, "real"), strings.Contains(dbType, "double"),
		strings.Contains(dbType, "float"):
		return "number"
	case strings.Contains(dbType, "bool"):
		return "boolean"
	case strings.Contains(dbType, "json"):
		return "object"
	case strings.Contains(dbType, "array"):
		return "array"
	default:
		return "string"
	}
}

// toPascalCase converts a string to PascalCase
func (sc *SchemaController) toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, "")
}

// toTitleCase converts a string to Title Case
func (sc *SchemaController) toTitleCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i, word := range words {
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}

// getCurrentSchemaWithDB gets the current database schema/database name for a specific DB connection
func (sc *SchemaController) getCurrentSchemaWithDB(db *sql.DB, iLog *logger.Log) string {
	if db == nil {
		return "public"
	}

	// Try MySQL first (DATABASE())
	var schemaName string
	err := db.QueryRow("SELECT DATABASE()").Scan(&schemaName)
	if err != nil {
		// If MySQL fails, try PostgreSQL (current_schema())
		err = db.QueryRow("SELECT current_schema()").Scan(&schemaName)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error getting current schema: %v, using 'public'", err))
			return "public"
		}
	}

	if schemaName == "" {
		return "public"
	}

	return schemaName
}

// getAllTableNamesWithAlias retrieves all table names from a specific database alias
func (sc *SchemaController) getAllTableNamesWithAlias(alias string, schemaName string, iLog *logger.Log) ([]string, error) {
	// Get database connection for alias
	db, err := orm.GetDB(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting database for alias '%s': %v", alias, err)
	}

	// If no schema provided, get current schema
	if schemaName == "" {
		schemaName = sc.getCurrentSchemaWithDB(db, iLog)
	}

	query := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ?
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := db.Query(query, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error querying tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// getTableStructureWithAlias retrieves the structure of a table from a specific database alias
func (sc *SchemaController) getTableStructureWithAlias(alias string, tableName string, schemaName string, iLog *logger.Log) (*DBTable, error) {
	// Get database connection for alias
	db, err := orm.GetDB(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting database for alias '%s': %v", alias, err)
	}

	// If no schema provided, get current schema
	if schemaName == "" {
		schemaName = sc.getCurrentSchemaWithDB(db, iLog)
	}

	table := &DBTable{
		Name:   tableName,
		Schema: schemaName,
	}

	// Get columns
	columnsQuery := `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale
		FROM information_schema.columns c
		WHERE c.table_name = ?
		AND c.table_schema = ?
		ORDER BY c.ordinal_position
	`

	rows, err := db.Query(columnsQuery, tableName, schemaName)
	if err != nil {
		return nil, fmt.Errorf("error querying columns: %v", err)
	}
	defer rows.Close()

	var columns []DBColumn
	for rows.Next() {
		var col DBColumn
		var dataType string
		var isNullable string
		var defaultValue, charMaxLength, numericPrecision, numericScale interface{}

		err := rows.Scan(&col.Name, &dataType, &isNullable, &defaultValue, &charMaxLength, &numericPrecision, &numericScale)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error scanning column: %v", err))
			continue
		}

		// Build complete type string
		col.Type = sc.formatDataType(dataType, charMaxLength, numericPrecision, numericScale)
		col.Nullable = isNullable == "YES"

		if defaultValue != nil {
			defStr := fmt.Sprintf("%v", defaultValue)
			col.DefaultValue = &defStr
		}

		columns = append(columns, col)
	}

	// Get primary keys
	pkQuery := `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.table_name = ?
			AND tc.table_schema = ?
			AND tc.constraint_type = 'PRIMARY KEY'
	`

	pkRows, err := db.Query(pkQuery, tableName, schemaName)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error querying primary keys: %v", err))
	} else {
		defer pkRows.Close()
		primaryKeys := make(map[string]bool)
		for pkRows.Next() {
			var colName string
			if err := pkRows.Scan(&colName); err == nil {
				primaryKeys[colName] = true
			}
		}

		// Mark primary key columns
		for i := range columns {
			if primaryKeys[columns[i].Name] {
				columns[i].PrimaryKey = true
			}
		}
	}

	// Get foreign keys
	fkQuery := `
		SELECT
			kcu.column_name,
			kcu.referenced_table_name AS foreign_table_name,
			kcu.referenced_column_name AS foreign_column_name
		FROM information_schema.key_column_usage AS kcu
		WHERE kcu.table_name = ?
			AND kcu.table_schema = ?
			AND kcu.referenced_table_name IS NOT NULL
	`

	fkRows, err := db.Query(fkQuery, tableName, schemaName)
	if err != nil {
		iLog.Error(fmt.Sprintf("Error querying foreign keys: %v", err))
	} else {
		defer fkRows.Close()
		foreignKeys := make(map[string]*ForeignKey)
		for fkRows.Next() {
			var colName, foreignTable, foreignColumn string
			if err := fkRows.Scan(&colName, &foreignTable, &foreignColumn); err == nil {
				foreignKeys[colName] = &ForeignKey{
					Table:  foreignTable,
					Column: foreignColumn,
				}
			}
		}

		// Mark foreign key columns
		for i := range columns {
			if fk, exists := foreignKeys[columns[i].Name]; exists {
				columns[i].ForeignKey = fk
			}
		}
	}

	table.Columns = columns
	return table, nil
}

// getDatabaseAliases returns all registered database alias names from configuration
func (sc *SchemaController) getDatabaseAliases(iLog *logger.Log) []string {
	aliases := []string{}

	// Get global configuration
	globalConfig := config.GlobalConfiguration
	if globalConfig == nil {
		iLog.Error("Global configuration is nil, returning default alias only")
		return []string{"default"}
	}

	// Always include the default database
	aliases = append(aliases, "default")

	// Get alternative databases from configuration
	if globalConfig.AltDatabasesConfig != nil && len(globalConfig.AltDatabasesConfig) > 0 {
		iLog.Debug(fmt.Sprintf("Found %d alternative databases in configuration", len(globalConfig.AltDatabasesConfig)))

		for _, altDB := range globalConfig.AltDatabasesConfig {
			// Get the "name" field from the map
			if name, ok := altDB["name"].(string); ok && name != "" {
				// Skip if it's "default" as we already added it
				if name != "default" {
					// Verify the alias is actually registered in ORM
					if _, err := orm.GetDB(name); err == nil {
						aliases = append(aliases, name)
						iLog.Debug(fmt.Sprintf("Added alias '%s' from configuration", name))
					} else {
						iLog.Warn(fmt.Sprintf("Alias '%s' found in config but not registered in ORM: %v", name, err))
					}
				}
			}
		}
	} else {
		iLog.Debug("No alternative databases configured")
	}

	iLog.Info(fmt.Sprintf("Database aliases available: %v", aliases))
	return aliases
}

// executeQuery executes a SQL query with parameters and returns results
func (sc *SchemaController) executeQuery(alias string, sqlQuery string, parameters map[string]interface{}, limit int, iLog *logger.Log) (map[string]interface{}, error) {
	startTime := time.Now()

	// Get database connection for alias
	db, err := orm.GetDB(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting database for alias '%s': %v", alias, err)
	}

	// Safety check - only allow SELECT queries
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))
	if !strings.HasPrefix(normalizedSQL, "SELECT") {
		return nil, fmt.Errorf("only SELECT queries are allowed")
	}

	// Check for dangerous operations
	dangerousOps := []string{"DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE", "ALTER", "CREATE", "EXEC", "EXECUTE"}
	for _, op := range dangerousOps {
		if strings.Contains(normalizedSQL, op) {
			return nil, fmt.Errorf("query contains forbidden operation: %s", op)
		}
	}

	// Replace parameter placeholders (@paramName) with actual values
	processedSQL, args, err := sc.replaceParameters(sqlQuery, parameters)
	if err != nil {
		return nil, fmt.Errorf("error processing parameters: %v", err)
	}

	// Add LIMIT clause if not present
	if !strings.Contains(normalizedSQL, "LIMIT") && !strings.Contains(normalizedSQL, "TOP") {
		processedSQL = fmt.Sprintf("%s LIMIT %d", processedSQL, limit)
	}

	iLog.Debug(fmt.Sprintf("Executing query: %s", processedSQL))

	// Execute query with timeout
	rows, err := db.Query(processedSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Get column names and types
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %v", err)
	}

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("error getting column types: %v", err)
	}

	// Build fields array with metadata
	fields := make([]map[string]interface{}, 0, len(columns))
	for i, col := range columns {
		fieldInfo := map[string]interface{}{
			"name":     col,
			"dataType": columnTypes[i].DatabaseTypeName(),
		}
		fields = append(fields, fieldInfo)
	}

	// Read all rows
	var resultRows []map[string]interface{}
	for rows.Next() {
		// Create slice of interface{} to hold each column
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			iLog.Error(fmt.Sprintf("Error scanning row: %v", err))
			continue
		}

		// Create a map for this row
		rowData := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			rowData[col] = v
		}
		resultRows = append(resultRows, rowData)
	}

	executionTime := time.Since(startTime).Milliseconds()

	// Build response
	result := map[string]interface{}{
		"columns":         columns,
		"rows":            resultRows,
		"totalRows":       len(resultRows),
		"executionTimeMs": executionTime,
		"fields":          fields,
	}

	return result, nil
}

// replaceParameters replaces @paramName placeholders with ? and returns ordered args
func (sc *SchemaController) replaceParameters(sqlQuery string, parameters map[string]interface{}) (string, []interface{}, error) {
	if parameters == nil || len(parameters) == 0 {
		return sqlQuery, []interface{}{}, nil
	}

	processedSQL := sqlQuery
	args := []interface{}{}

	// Build replacements
	for paramName, paramValue := range parameters {
		placeholder := "@" + paramName
		if strings.Contains(processedSQL, placeholder) {
			processedSQL = strings.Replace(processedSQL, placeholder, "?", -1)
			args = append(args, paramValue)
		}
	}

	return processedSQL, args, nil
}

// validateQuery validates SQL syntax without executing
func (sc *SchemaController) validateQuery(alias string, sqlQuery string, iLog *logger.Log) map[string]interface{} {
	result := map[string]interface{}{
		"valid":              true,
		"errors":             []string{},
		"warnings":           []string{},
		"detectedParameters": []string{},
	}

	// Basic validation
	normalizedSQL := strings.ToUpper(strings.TrimSpace(sqlQuery))

	// Check if it's a SELECT query
	if !strings.HasPrefix(normalizedSQL, "SELECT") {
		result["valid"] = false
		result["errors"] = append(result["errors"].([]string), "Only SELECT queries are allowed")
		return result
	}

	// Check for dangerous operations
	dangerousOps := []string{"DROP", "DELETE", "TRUNCATE", "INSERT", "UPDATE", "ALTER", "CREATE", "EXEC", "EXECUTE"}
	for _, op := range dangerousOps {
		if strings.Contains(normalizedSQL, op) {
			result["valid"] = false
			result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Query contains forbidden operation: %s", op))
		}
	}

	// Detect parameters
	params := []string{}
	for _, match := range strings.Split(sqlQuery, "@") {
		if len(match) > 0 {
			// Extract parameter name (alphanumeric and underscore)
			paramName := ""
			for _, char := range match {
				if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_' {
					paramName += string(char)
				} else {
					break
				}
			}
			if paramName != "" && !contains(params, paramName) {
				params = append(params, paramName)
			}
		}
	}
	result["detectedParameters"] = params

	// Warning if no LIMIT clause
	if !strings.Contains(normalizedSQL, "LIMIT") && !strings.Contains(normalizedSQL, "TOP") {
		result["warnings"] = append(result["warnings"].([]string), "Query does not include LIMIT clause, may return large result set")
	}

	// Try to use database EXPLAIN to validate syntax (optional)
	db, err := orm.GetDB(alias)
	if err != nil {
		result["warnings"] = append(result["warnings"].([]string), fmt.Sprintf("Could not connect to database for syntax validation: %v", err))
		return result
	}

	// Try EXPLAIN (works for most SQL databases)
	explainSQL := "EXPLAIN " + sqlQuery
	_, err = db.Query(explainSQL)
	if err != nil {
		result["valid"] = false
		result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Syntax error: %v", err))
	}

	return result
}

// getRelationshipsWithAlias retrieves foreign key relationships for tables from a specific database alias
func (sc *SchemaController) getRelationshipsWithAlias(alias string, tables []string, schemaName string, iLog *logger.Log) ([]map[string]interface{}, error) {
	// Get database connection for alias
	db, err := orm.GetDB(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting database for alias '%s': %v", alias, err)
	}

	// If no schema provided, get current schema
	if schemaName == "" {
		var currentSchema string
		err := db.QueryRow("SELECT DATABASE()").Scan(&currentSchema)
		if err != nil {
			// Try PostgreSQL syntax
			err = db.QueryRow("SELECT current_schema()").Scan(&currentSchema)
			if err != nil {
				iLog.Warn(fmt.Sprintf("Could not determine current schema: %v, using 'public'", err))
				schemaName = "public"
			} else {
				schemaName = currentSchema
			}
		} else {
			schemaName = currentSchema
		}
	}

	// Build query to get foreign key relationships
	query := `
		SELECT
			rc.constraint_name,
			kcu.table_name AS source_table,
			kcu.column_name AS source_column,
			kcu.referenced_table_name AS target_table,
			kcu.referenced_column_name AS target_column
		FROM information_schema.referential_constraints rc
		JOIN information_schema.key_column_usage kcu
			ON rc.constraint_name = kcu.constraint_name
			AND rc.constraint_schema = kcu.constraint_schema
		WHERE kcu.table_schema = ?
	`

	args := []interface{}{schemaName}

	// Filter by specific tables if provided
	if len(tables) > 0 {
		placeholders := make([]string, len(tables))
		for i := range tables {
			placeholders[i] = "?"
			args = append(args, tables[i])
		}
		query += fmt.Sprintf(" AND (kcu.table_name IN (%s) OR kcu.referenced_table_name IN (%s))",
			strings.Join(placeholders, ","),
			strings.Join(placeholders, ","))
		// Duplicate the tables for the second IN clause
		for _, table := range tables {
			args = append(args, table)
		}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying relationships: %v", err)
	}
	defer rows.Close()

	relationships := []map[string]interface{}{}
	for rows.Next() {
		var constraintName, sourceTable, sourceColumn, targetTable, targetColumn string
		err := rows.Scan(&constraintName, &sourceTable, &sourceColumn, &targetTable, &targetColumn)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error scanning relationship: %v", err))
			continue
		}

		relationship := map[string]interface{}{
			"constraintName": constraintName,
			"sourceTable":    sourceTable,
			"sourceColumn":   sourceColumn,
			"targetTable":    targetTable,
			"targetColumn":   targetColumn,
		}
		relationships = append(relationships, relationship)
	}

	return relationships, nil
}

// contains checks if a string slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// getStoredProcedures retrieves list of stored procedures from a database
func (sc *SchemaController) getStoredProcedures(alias string, schemaName string, iLog *logger.Log) ([]map[string]interface{}, error) {
	// Get database connection for alias
	db, err := orm.GetDB(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting database for alias '%s': %v", alias, err)
	}

	// Determine database type and schema
	dbType, currentSchema := sc.detectDatabaseType(db, schemaName, iLog)

	var query string
	var args []interface{}

	switch dbType {
	case "postgresql":
		// PostgreSQL: Query information_schema.routines for functions (PostgreSQL uses functions instead of procedures)
		query = `
			SELECT
				routine_name as name,
				routine_schema as schema,
				routine_schema || '.' || routine_name as full_name,
				COALESCE(routine_comment, '') as description,
				created as created,
				last_altered as modified
			FROM information_schema.routines
			WHERE routine_type = 'FUNCTION'
				AND routine_schema = ?
			ORDER BY routine_name
		`
		args = []interface{}{currentSchema}

	case "mysql":
		// MySQL: Use SHOW PROCEDURE STATUS
		query = `
			SELECT
				name,
				db as schema,
				CONCAT(db, '.', name) as full_name,
				COALESCE(comment, '') as description,
				created,
				modified
			FROM mysql.proc
			WHERE db = ?
				AND type = 'PROCEDURE'
			ORDER BY name
		`
		args = []interface{}{currentSchema}

	case "sqlserver":
		// SQL Server: Query sys.procedures
		if schemaName == "" {
			schemaName = "dbo"
		}
		query = `
			SELECT
				p.name,
				SCHEMA_NAME(p.schema_id) as schema,
				SCHEMA_NAME(p.schema_id) + '.' + p.name as full_name,
				COALESCE(CAST(ep.value AS NVARCHAR(MAX)), '') as description,
				p.create_date as created,
				p.modify_date as modified
			FROM sys.procedures p
			LEFT JOIN sys.extended_properties ep
				ON ep.major_id = p.object_id
				AND ep.minor_id = 0
				AND ep.name = 'MS_Description'
			WHERE SCHEMA_NAME(p.schema_id) = ?
			ORDER BY p.name
		`
		args = []interface{}{schemaName}

	default:
		// Fallback: Try standard INFORMATION_SCHEMA
		query = `
			SELECT
				routine_name as name,
				routine_schema as schema,
				routine_schema || '.' || routine_name as full_name,
				'' as description,
				created as created,
				last_altered as modified
			FROM information_schema.routines
			WHERE routine_type IN ('PROCEDURE', 'FUNCTION')
				AND routine_schema = ?
			ORDER BY routine_name
		`
		args = []interface{}{currentSchema}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying stored procedures: %v", err)
	}
	defer rows.Close()

	procedures := []map[string]interface{}{}
	for rows.Next() {
		var name, schema, fullName, description string
		var created, modified interface{}

		err := rows.Scan(&name, &schema, &fullName, &description, &created, &modified)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error scanning procedure: %v", err))
			continue
		}

		procedure := map[string]interface{}{
			"name":        name,
			"schema":      schema,
			"fullName":    fullName,
			"description": description,
			"created":     created,
			"modified":    modified,
		}
		procedures = append(procedures, procedure)
	}

	return procedures, nil
}

// getProcedureMetadata retrieves parameters and metadata for a stored procedure
func (sc *SchemaController) getProcedureMetadata(alias string, procedureName string, schemaName string, iLog *logger.Log) (map[string]interface{}, error) {
	// Get database connection for alias
	db, err := orm.GetDB(alias)
	if err != nil {
		return nil, fmt.Errorf("error getting database for alias '%s': %v", alias, err)
	}

	// Determine database type and schema
	dbType, currentSchema := sc.detectDatabaseType(db, schemaName, iLog)

	// Parse procedure name (handle schema.procedure format)
	procSchema := currentSchema
	procName := procedureName
	if strings.Contains(procedureName, ".") {
		parts := strings.SplitN(procedureName, ".", 2)
		procSchema = parts[0]
		procName = parts[1]
	}

	var query string
	var args []interface{}

	switch dbType {
	case "postgresql":
		// PostgreSQL: Query information_schema.parameters
		query = `
			SELECT
				parameter_name as name,
				data_type,
				parameter_mode as direction,
				CASE WHEN is_nullable = 'YES' THEN true ELSE false END as is_nullable,
				parameter_default as default_value,
				ordinal_position,
				character_maximum_length as max_length
			FROM information_schema.parameters
			WHERE specific_schema = ?
				AND specific_name = ?
			ORDER BY ordinal_position
		`
		args = []interface{}{procSchema, procName}

	case "mysql":
		// MySQL: Query information_schema.parameters
		query = `
			SELECT
				PARAMETER_NAME as name,
				DATA_TYPE as data_type,
				PARAMETER_MODE as direction,
				CASE WHEN IS_NULLABLE = 'YES' THEN true ELSE false END as is_nullable,
				NULL as default_value,
				ORDINAL_POSITION as ordinal_position,
				CHARACTER_MAXIMUM_LENGTH as max_length
			FROM information_schema.parameters
			WHERE SPECIFIC_SCHEMA = ?
				AND SPECIFIC_NAME = ?
			ORDER BY ORDINAL_POSITION
		`
		args = []interface{}{procSchema, procName}

	case "sqlserver":
		// SQL Server: Query sys.parameters
		query = `
			SELECT
				p.name,
				TYPE_NAME(p.system_type_id) as data_type,
				CASE WHEN p.is_output = 1 THEN 'OUT' ELSE 'IN' END as direction,
				CASE WHEN p.is_nullable = 1 THEN true ELSE false END as is_nullable,
				p.default_value,
				p.parameter_id as ordinal_position,
				p.max_length
			FROM sys.parameters p
			WHERE p.object_id = OBJECT_ID(?)
			ORDER BY p.parameter_id
		`
		fullProcName := procSchema + "." + procName
		args = []interface{}{fullProcName}

	default:
		// Fallback: Try standard INFORMATION_SCHEMA
		query = `
			SELECT
				parameter_name as name,
				data_type,
				parameter_mode as direction,
				CASE WHEN is_nullable = 'YES' THEN 1 ELSE 0 END as is_nullable,
				parameter_default as default_value,
				ordinal_position,
				character_maximum_length as max_length
			FROM information_schema.parameters
			WHERE specific_schema = ?
				AND specific_name = ?
			ORDER BY ordinal_position
		`
		args = []interface{}{procSchema, procName}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error querying procedure parameters: %v", err)
	}
	defer rows.Close()

	parameters := []map[string]interface{}{}
	for rows.Next() {
		var name, dataType, direction string
		var isNullable interface{}
		var defaultValue, maxLength interface{}
		var ordinalPosition int

		err := rows.Scan(&name, &dataType, &direction, &isNullable, &defaultValue, &ordinalPosition, &maxLength)
		if err != nil {
			iLog.Error(fmt.Sprintf("Error scanning parameter: %v", err))
			continue
		}

		// Normalize direction (IN, OUT, INOUT)
		if direction == "" {
			direction = "IN"
		}

		parameter := map[string]interface{}{
			"name":            name,
			"dataType":        dataType,
			"direction":       direction,
			"isNullable":      isNullable,
			"defaultValue":    defaultValue,
			"ordinalPosition": ordinalPosition,
			"maxLength":       maxLength,
		}
		parameters = append(parameters, parameter)
	}

	// Build metadata response
	metadata := map[string]interface{}{
		"name":          procName,
		"schema":        procSchema,
		"fullName":      procSchema + "." + procName,
		"description":   "",
		"parameters":    parameters,
		"returnType":    "TABLE",
		"resultColumns": []map[string]interface{}{}, // Could be enhanced to detect result columns
	}

	return metadata, nil
}

// detectDatabaseType detects the database type and returns current schema
func (sc *SchemaController) detectDatabaseType(db *sql.DB, schemaName string, iLog *logger.Log) (string, string) {
	// Try to detect database type
	var dbType string
	var currentSchema string

	// Try MySQL detection
	var mysqlVersion string
	err := db.QueryRow("SELECT VERSION()").Scan(&mysqlVersion)
	if err == nil && strings.Contains(strings.ToLower(mysqlVersion), "mysql") {
		dbType = "mysql"
		if schemaName == "" {
			db.QueryRow("SELECT DATABASE()").Scan(&currentSchema)
		} else {
			currentSchema = schemaName
		}
		return dbType, currentSchema
	}

	// Try PostgreSQL detection
	var pgVersion string
	err = db.QueryRow("SELECT version()").Scan(&pgVersion)
	if err == nil && strings.Contains(strings.ToLower(pgVersion), "postgresql") {
		dbType = "postgresql"
		if schemaName == "" {
			db.QueryRow("SELECT current_schema()").Scan(&currentSchema)
		} else {
			currentSchema = schemaName
		}
		if currentSchema == "" {
			currentSchema = "public"
		}
		return dbType, currentSchema
	}

	// Try SQL Server detection
	var serverProperty string
	err = db.QueryRow("SELECT SERVERPROPERTY('ProductVersion')").Scan(&serverProperty)
	if err == nil {
		dbType = "sqlserver"
		if schemaName == "" {
			currentSchema = "dbo"
		} else {
			currentSchema = schemaName
		}
		return dbType, currentSchema
	}

	// Default fallback
	dbType = "unknown"
	if schemaName == "" {
		currentSchema = "public"
	} else {
		currentSchema = schemaName
	}

	iLog.Warn(fmt.Sprintf("Could not detect database type, using fallback. Schema: %s", currentSchema))
	return dbType, currentSchema
}
