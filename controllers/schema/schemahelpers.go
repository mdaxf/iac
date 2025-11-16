package schema

import (
	"database/sql"
	"fmt"
	"strings"

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
