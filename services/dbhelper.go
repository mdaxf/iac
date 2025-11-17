package services

import (
	"context"
	"database/sql"
	"fmt"

	dbconn "github.com/mdaxf/iac/databases"
	"gorm.io/gorm"
)

// DatabaseHelper provides multi-database support for services
// It wraps the database selector and provides dialect-aware operations
type DatabaseHelper struct {
	selector *dbconn.DatabaseSelector
	gormDB   *gorm.DB // GORM DB for IAC application's own tables
}

// NewDatabaseHelper creates a new database helper
func NewDatabaseHelper(selector *dbconn.DatabaseSelector, gormDB *gorm.DB) *DatabaseHelper {
	return &DatabaseHelper{
		selector: selector,
		gormDB:   gormDB,
	}
}

// GetAppDB returns the GORM database for IAC application's own tables
func (h *DatabaseHelper) GetAppDB() *gorm.DB {
	return h.gormDB
}

// GetUserDB gets a database connection for a user's database
// This uses the database selector to choose the appropriate database
func (h *DatabaseHelper) GetUserDB(ctx context.Context, databaseAlias string) (dbconn.RelationalDB, error) {
	dbCtx := &dbconn.DatabaseContext{
		Operation: dbconn.OperationRead,
		Metadata: map[string]interface{}{
			"database_alias": databaseAlias,
		},
	}

	return h.selector.SelectDatabase(dbCtx)
}

// GetUserDBForWrite gets a database connection for write operations
func (h *DatabaseHelper) GetUserDBForWrite(ctx context.Context, databaseAlias string) (dbconn.RelationalDB, error) {
	dbCtx := &dbconn.DatabaseContext{
		Operation: dbconn.OperationWrite,
		Metadata: map[string]interface{}{
			"database_alias": databaseAlias,
		},
	}

	return h.selector.SelectDatabase(dbCtx)
}

// ExecuteDialectQuery executes a query using the appropriate dialect
func (h *DatabaseHelper) ExecuteDialectQuery(ctx context.Context, db dbconn.RelationalDB, queryFunc func(dialectName string) string, args ...interface{}) (*sql.Rows, error) {
	// Get dialect name from database type
	dialectName := string(db.GetType())
	query := queryFunc(dialectName)
	return db.Query(ctx, query, args...)
}

// CurrentTimestampExpr returns the appropriate current timestamp expression for the dialect
func CurrentTimestampExpr(dialect string) string {
	switch dialect {
	case "postgres":
		return "CURRENT_TIMESTAMP"
	case "mysql":
		return "CURRENT_TIMESTAMP"
	case "mssql":
		return "GETDATE()"
	case "oracle":
		return "SYSTIMESTAMP"
	default:
		return "CURRENT_TIMESTAMP"
	}
}

// LikeOperator returns the appropriate LIKE operator for the dialect
func LikeOperator(dialect string, caseSensitive bool) string {
	if caseSensitive {
		return "LIKE"
	}

	switch dialect {
	case "postgres":
		return "ILIKE"
	case "mysql":
		return "LIKE" // MySQL LIKE is case-insensitive by default
	case "mssql":
		return "LIKE" // Use collation for case sensitivity
	case "oracle":
		return "LIKE" // Use UPPER() for case-insensitive
	default:
		return "LIKE"
	}
}

// CoalesceExpr returns the appropriate COALESCE expression
func CoalesceExpr(column string, defaultValue string) string {
	return fmt.Sprintf("COALESCE(%s, %s)", column, defaultValue)
}

// LimitOffsetClause returns the appropriate pagination clause for the dialect
func LimitOffsetClause(dialect string, limit, offset int) string {
	switch dialect {
	case "postgres", "mysql":
		if offset > 0 {
			return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
		}
		return fmt.Sprintf("LIMIT %d", limit)
	case "mssql":
		// MSSQL requires ORDER BY with OFFSET/FETCH
		if offset > 0 {
			return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
		}
		return fmt.Sprintf("FETCH FIRST %d ROWS ONLY", limit)
	case "oracle":
		// Oracle 12c+ syntax
		if offset > 0 {
			return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
		}
		return fmt.Sprintf("FETCH FIRST %d ROWS ONLY", limit)
	default:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}
}

// StringConcatExpr returns the appropriate string concatenation expression
func StringConcatExpr(dialect string, parts ...string) string {
	switch dialect {
	case "postgres", "oracle":
		// PostgreSQL and Oracle use ||
		return joinWith(" || ", parts)
	case "mysql":
		// MySQL uses CONCAT()
		return fmt.Sprintf("CONCAT(%s)", joinWith(", ", parts))
	case "mssql":
		// MSSQL uses +
		return joinWith(" + ", parts)
	default:
		return fmt.Sprintf("CONCAT(%s)", joinWith(", ", parts))
	}
}

// joinWith is a helper to join strings with a separator
func joinWith(separator string, parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += separator
		}
		result += part
	}
	return result
}

// BooleanLiteral returns the appropriate boolean literal for the dialect
func BooleanLiteral(dialect string, value bool) string {
	switch dialect {
	case "postgres":
		if value {
			return "TRUE"
		}
		return "FALSE"
	case "mysql":
		if value {
			return "1"
		}
		return "0"
	case "mssql", "oracle":
		if value {
			return "1"
		}
		return "0"
	default:
		if value {
			return "TRUE"
		}
		return "FALSE"
	}
}

// JSONExtractExpr returns the appropriate JSON extract expression
func JSONExtractExpr(dialect string, column string, path string) string {
	switch dialect {
	case "postgres":
		return fmt.Sprintf("%s->>'%s'", column, path)
	case "mysql":
		return fmt.Sprintf("JSON_EXTRACT(%s, '$.%s')", column, path)
	case "mssql":
		return fmt.Sprintf("JSON_VALUE(%s, '$.%s')", column, path)
	case "oracle":
		return fmt.Sprintf("JSON_VALUE(%s, '$.%s')", column, path)
	default:
		return fmt.Sprintf("JSON_EXTRACT(%s, '$.%s')", column, path)
	}
}
