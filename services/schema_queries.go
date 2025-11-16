package services

import (
	"fmt"
)

// SchemaQuery represents a database schema discovery query
type SchemaQuery struct {
	Dialect string
	Query   string
}

// TableInfo represents table metadata structure
type TableInfo struct {
	TableName    string
	TableComment string
	TableSchema  string
}

// ColumnInfo represents column metadata structure
type ColumnInfo struct {
	ColumnName    string
	DataType      string
	IsNullable    string
	ColumnKey     string
	ColumnComment string
	OrdinalPos    int
}

// GetTablesQuery returns the appropriate query to discover tables for each database type
func GetTablesQuery(dialect string, schemaName string) string {
	switch dialect {
	case "mysql":
		return fmt.Sprintf(`
			SELECT
				TABLE_NAME as table_name,
				COALESCE(TABLE_COMMENT, '') as table_comment,
				TABLE_SCHEMA as table_schema
			FROM information_schema.TABLES
			WHERE TABLE_SCHEMA = '%s'
			AND TABLE_TYPE = 'BASE TABLE'
			ORDER BY TABLE_NAME
		`, schemaName)

	case "postgres":
		return fmt.Sprintf(`
			SELECT
				t.table_name as table_name,
				COALESCE(pg_catalog.obj_description(pgc.oid, 'pg_class'), '') as table_comment,
				t.table_schema as table_schema
			FROM information_schema.tables t
			LEFT JOIN pg_catalog.pg_class pgc ON pgc.relname = t.table_name
			LEFT JOIN pg_catalog.pg_namespace pgn ON pgn.oid = pgc.relnamespace AND pgn.nspname = t.table_schema
			WHERE t.table_schema = '%s'
			AND t.table_type = 'BASE TABLE'
			ORDER BY t.table_name
		`, schemaName)

	case "mssql":
		return fmt.Sprintf(`
			SELECT
				t.TABLE_NAME as table_name,
				COALESCE(CAST(ep.value AS NVARCHAR(MAX)), '') as table_comment,
				t.TABLE_SCHEMA as table_schema
			FROM INFORMATION_SCHEMA.TABLES t
			LEFT JOIN sys.tables st ON st.name = t.TABLE_NAME
			LEFT JOIN sys.extended_properties ep ON ep.major_id = st.object_id
				AND ep.minor_id = 0
				AND ep.name = 'MS_Description'
			WHERE t.TABLE_SCHEMA = '%s'
			AND t.TABLE_TYPE = 'BASE TABLE'
			ORDER BY t.TABLE_NAME
		`, schemaName)

	case "oracle":
		return fmt.Sprintf(`
			SELECT
				table_name as table_name,
				COALESCE(
					(SELECT comments FROM all_tab_comments
					 WHERE owner = '%s' AND table_name = t.table_name),
					''
				) as table_comment,
				'%s' as table_schema
			FROM all_tables t
			WHERE owner = '%s'
			ORDER BY table_name
		`, schemaName, schemaName, schemaName)

	default:
		return fmt.Sprintf(`
			SELECT
				TABLE_NAME as table_name,
				COALESCE(TABLE_COMMENT, '') as table_comment,
				TABLE_SCHEMA as table_schema
			FROM information_schema.TABLES
			WHERE TABLE_SCHEMA = '%s'
			AND TABLE_TYPE = 'BASE TABLE'
			ORDER BY TABLE_NAME
		`, schemaName)
	}
}

// GetColumnsQuery returns the appropriate query to discover columns for each database type
func GetColumnsQuery(dialect string, schemaName string, tableName string) string {
	switch dialect {
	case "mysql":
		return fmt.Sprintf(`
			SELECT
				COLUMN_NAME as column_name,
				DATA_TYPE as data_type,
				IS_NULLABLE as is_nullable,
				COLUMN_KEY as column_key,
				COALESCE(COLUMN_COMMENT, '') as column_comment,
				ORDINAL_POSITION as ordinal_pos
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = '%s'
			AND TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION
		`, schemaName, tableName)

	case "postgres":
		return fmt.Sprintf(`
			SELECT
				c.column_name as column_name,
				c.data_type as data_type,
				c.is_nullable as is_nullable,
				CASE WHEN pk.column_name IS NOT NULL THEN 'PRI' ELSE '' END as column_key,
				COALESCE(pgd.description, '') as column_comment,
				c.ordinal_position as ordinal_pos
			FROM information_schema.columns c
			LEFT JOIN (
				SELECT ku.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage ku
					ON tc.constraint_name = ku.constraint_name
					AND tc.table_schema = ku.table_schema
				WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_schema = '%s'
				AND tc.table_name = '%s'
			) pk ON pk.column_name = c.column_name
			LEFT JOIN pg_catalog.pg_statio_all_tables st
				ON st.schemaname = c.table_schema
				AND st.relname = c.table_name
			LEFT JOIN pg_catalog.pg_description pgd
				ON pgd.objoid = st.relid
				AND pgd.objsubid = c.ordinal_position
			WHERE c.table_schema = '%s'
			AND c.table_name = '%s'
			ORDER BY c.ordinal_position
		`, schemaName, tableName, schemaName, tableName)

	case "mssql":
		return fmt.Sprintf(`
			SELECT
				c.COLUMN_NAME as column_name,
				c.DATA_TYPE as data_type,
				c.IS_NULLABLE as is_nullable,
				CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 'PRI' ELSE '' END as column_key,
				COALESCE(CAST(ep.value AS NVARCHAR(MAX)), '') as column_comment,
				c.ORDINAL_POSITION as ordinal_pos
			FROM INFORMATION_SCHEMA.COLUMNS c
			LEFT JOIN (
				SELECT ku.COLUMN_NAME
				FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
				JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
					ON tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
					AND tc.TABLE_SCHEMA = ku.TABLE_SCHEMA
					AND tc.TABLE_NAME = ku.TABLE_NAME
				WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
				AND tc.TABLE_SCHEMA = '%s'
				AND tc.TABLE_NAME = '%s'
			) pk ON pk.COLUMN_NAME = c.COLUMN_NAME
			LEFT JOIN sys.columns sc
				ON sc.object_id = OBJECT_ID('%s.%s')
				AND sc.name = c.COLUMN_NAME
			LEFT JOIN sys.extended_properties ep
				ON ep.major_id = sc.object_id
				AND ep.minor_id = sc.column_id
				AND ep.name = 'MS_Description'
			WHERE c.TABLE_SCHEMA = '%s'
			AND c.TABLE_NAME = '%s'
			ORDER BY c.ORDINAL_POSITION
		`, schemaName, tableName, schemaName, tableName, schemaName, tableName)

	case "oracle":
		return fmt.Sprintf(`
			SELECT
				column_name as column_name,
				data_type as data_type,
				CASE WHEN nullable = 'Y' THEN 'YES' ELSE 'NO' END as is_nullable,
				CASE WHEN EXISTS (
					SELECT 1 FROM all_constraints c
					JOIN all_cons_columns cc ON c.constraint_name = cc.constraint_name
						AND c.owner = cc.owner
					WHERE c.constraint_type = 'P'
					AND c.owner = '%s'
					AND c.table_name = '%s'
					AND cc.column_name = t.column_name
				) THEN 'PRI' ELSE '' END as column_key,
				COALESCE(
					(SELECT comments FROM all_col_comments
					 WHERE owner = '%s' AND table_name = '%s' AND column_name = t.column_name),
					''
				) as column_comment,
				column_id as ordinal_pos
			FROM all_tab_columns t
			WHERE owner = '%s'
			AND table_name = '%s'
			ORDER BY column_id
		`, schemaName, tableName, schemaName, tableName, schemaName, tableName)

	default:
		return fmt.Sprintf(`
			SELECT
				COLUMN_NAME as column_name,
				DATA_TYPE as data_type,
				IS_NULLABLE as is_nullable,
				COLUMN_KEY as column_key,
				COALESCE(COLUMN_COMMENT, '') as column_comment,
				ORDINAL_POSITION as ordinal_pos
			FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = '%s'
			AND TABLE_NAME = '%s'
			ORDER BY ORDINAL_POSITION
		`, schemaName, tableName)
	}
}

// GetDatabaseListQuery returns the appropriate query to list all databases/schemas
func GetDatabaseListQuery(dialect string) string {
	switch dialect {
	case "mysql":
		return `
			SELECT SCHEMA_NAME as database_name
			FROM information_schema.SCHEMATA
			WHERE SCHEMA_NAME NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
			ORDER BY SCHEMA_NAME
		`

	case "postgres":
		return `
			SELECT datname as database_name
			FROM pg_database
			WHERE datistemplate = false
			AND datname NOT IN ('postgres', 'template0', 'template1')
			ORDER BY datname
		`

	case "mssql":
		return `
			SELECT name as database_name
			FROM sys.databases
			WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb')
			AND state_desc = 'ONLINE'
			ORDER BY name
		`

	case "oracle":
		return `
			SELECT DISTINCT owner as database_name
			FROM all_tables
			WHERE owner NOT IN ('SYS', 'SYSTEM', 'OUTLN', 'XDB', 'CTXSYS', 'MDSYS', 'ORDSYS', 'WMSYS')
			ORDER BY owner
		`

	default:
		return `
			SELECT SCHEMA_NAME as database_name
			FROM information_schema.SCHEMATA
			ORDER BY SCHEMA_NAME
		`
	}
}

// GetIndexesQuery returns the appropriate query to discover indexes for each database type
func GetIndexesQuery(dialect string, schemaName string, tableName string) string {
	switch dialect {
	case "mysql":
		return fmt.Sprintf(`
			SELECT
				INDEX_NAME as index_name,
				COLUMN_NAME as column_name,
				NON_UNIQUE as non_unique,
				SEQ_IN_INDEX as seq_in_index,
				INDEX_TYPE as index_type
			FROM information_schema.STATISTICS
			WHERE TABLE_SCHEMA = '%s'
			AND TABLE_NAME = '%s'
			ORDER BY INDEX_NAME, SEQ_IN_INDEX
		`, schemaName, tableName)

	case "postgres":
		return fmt.Sprintf(`
			SELECT
				i.relname as index_name,
				a.attname as column_name,
				CASE WHEN ix.indisunique THEN 0 ELSE 1 END as non_unique,
				a.attnum as seq_in_index,
				am.amname as index_type
			FROM pg_class t
			JOIN pg_index ix ON t.oid = ix.indrelid
			JOIN pg_class i ON i.oid = ix.indexrelid
			JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
			JOIN pg_am am ON i.relam = am.oid
			JOIN pg_namespace n ON t.relnamespace = n.oid
			WHERE n.nspname = '%s'
			AND t.relname = '%s'
			ORDER BY i.relname, a.attnum
		`, schemaName, tableName)

	case "mssql":
		return fmt.Sprintf(`
			SELECT
				i.name as index_name,
				c.name as column_name,
				CASE WHEN i.is_unique = 1 THEN 0 ELSE 1 END as non_unique,
				ic.key_ordinal as seq_in_index,
				i.type_desc as index_type
			FROM sys.indexes i
			JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
			JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
			JOIN sys.tables t ON i.object_id = t.object_id
			JOIN sys.schemas s ON t.schema_id = s.schema_id
			WHERE s.name = '%s'
			AND t.name = '%s'
			ORDER BY i.name, ic.key_ordinal
		`, schemaName, tableName)

	case "oracle":
		return fmt.Sprintf(`
			SELECT
				i.index_name as index_name,
				ic.column_name as column_name,
				CASE WHEN i.uniqueness = 'UNIQUE' THEN 0 ELSE 1 END as non_unique,
				ic.column_position as seq_in_index,
				i.index_type as index_type
			FROM all_indexes i
			JOIN all_ind_columns ic ON i.index_name = ic.index_name AND i.owner = ic.index_owner
			WHERE i.owner = '%s'
			AND i.table_name = '%s'
			ORDER BY i.index_name, ic.column_position
		`, schemaName, tableName)

	default:
		return fmt.Sprintf(`
			SELECT
				INDEX_NAME as index_name,
				COLUMN_NAME as column_name,
				NON_UNIQUE as non_unique,
				SEQ_IN_INDEX as seq_in_index,
				INDEX_TYPE as index_type
			FROM information_schema.STATISTICS
			WHERE TABLE_SCHEMA = '%s'
			AND TABLE_NAME = '%s'
			ORDER BY INDEX_NAME, SEQ_IN_INDEX
		`, schemaName, tableName)
	}
}

// NormalizeDataType normalizes database-specific data types to common types
func NormalizeDataType(dialect string, dataType string) string {
	// Common type mappings
	typeMap := map[string]map[string]string{
		"mysql": {
			"int":        "integer",
			"bigint":     "bigint",
			"varchar":    "varchar",
			"text":       "text",
			"datetime":   "timestamp",
			"timestamp":  "timestamp",
			"decimal":    "decimal",
			"float":      "float",
			"double":     "double",
			"tinyint":    "boolean",
			"mediumtext": "text",
			"longtext":   "text",
		},
		"postgres": {
			"integer":             "integer",
			"bigint":              "bigint",
			"character varying":   "varchar",
			"text":                "text",
			"timestamp":           "timestamp",
			"timestamp with time zone": "timestamp",
			"numeric":             "decimal",
			"real":                "float",
			"double precision":    "double",
			"boolean":             "boolean",
		},
		"mssql": {
			"int":         "integer",
			"bigint":      "bigint",
			"varchar":     "varchar",
			"nvarchar":    "varchar",
			"text":        "text",
			"ntext":       "text",
			"datetime":    "timestamp",
			"datetime2":   "timestamp",
			"decimal":     "decimal",
			"float":       "float",
			"real":        "float",
			"bit":         "boolean",
		},
		"oracle": {
			"NUMBER":      "integer",
			"VARCHAR2":    "varchar",
			"CLOB":        "text",
			"DATE":        "timestamp",
			"TIMESTAMP":   "timestamp",
			"FLOAT":       "float",
			"BINARY_FLOAT": "float",
			"BINARY_DOUBLE": "double",
		},
	}

	if dialectMap, ok := typeMap[dialect]; ok {
		if normalizedType, exists := dialectMap[dataType]; exists {
			return normalizedType
		}
	}

	return dataType
}
