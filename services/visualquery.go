package services

import (
	"fmt"
	"strings"
)

// VisualQueryField represents a field in the visual query
type VisualQueryField struct {
	Table      string `json:"table"`
	Column     string `json:"column"`
	Alias      string `json:"alias,omitempty"`
	Aggregation string `json:"aggregation,omitempty"` // SUM, AVG, COUNT, MAX, MIN
}

// VisualQueryFilter represents a filter condition
type VisualQueryFilter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // =, !=, >, <, >=, <=, LIKE, IN, BETWEEN
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic,omitempty"` // AND, OR
}

// VisualQueryJoin represents a table join
type VisualQueryJoin struct {
	Type        string `json:"type"` // INNER, LEFT, RIGHT, FULL
	Table       string `json:"table"`
	LeftColumn  string `json:"leftColumn"`
	RightColumn string `json:"rightColumn"`
}

// VisualQuerySort represents a sort order
type VisualQuerySort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // ASC, DESC
}

// VisualQuery represents the complete visual query structure
type VisualQuery struct {
	Tables  []string            `json:"tables"`
	Fields  []VisualQueryField  `json:"fields"`
	Filters []VisualQueryFilter `json:"filters,omitempty"`
	Joins   []VisualQueryJoin   `json:"joins,omitempty"`
	Sorts   []VisualQuerySort   `json:"sorts,omitempty"`
	GroupBy []string            `json:"groupBy,omitempty"`
	Limit   int                 `json:"limit,omitempty"`
	Offset  int                 `json:"offset,omitempty"`
}

// ParseVisualQuery parses a visual query map into structured VisualQuery
func ParseVisualQuery(queryMap map[string]interface{}) (*VisualQuery, error) {
	vq := &VisualQuery{}

	// Parse tables
	if tables, ok := queryMap["tables"].([]interface{}); ok {
		for _, t := range tables {
			if tableName, ok := t.(string); ok {
				vq.Tables = append(vq.Tables, tableName)
			}
		}
	}

	// Parse fields
	if fields, ok := queryMap["fields"].([]interface{}); ok {
		for _, f := range fields {
			if fieldMap, ok := f.(map[string]interface{}); ok {
				field := VisualQueryField{}
				if table, ok := fieldMap["table"].(string); ok {
					field.Table = table
				}
				if column, ok := fieldMap["column"].(string); ok {
					field.Column = column
				}
				if alias, ok := fieldMap["alias"].(string); ok {
					field.Alias = alias
				}
				if agg, ok := fieldMap["aggregation"].(string); ok {
					field.Aggregation = agg
				}
				vq.Fields = append(vq.Fields, field)
			}
		}
	}

	// Parse filters
	if filters, ok := queryMap["filters"].([]interface{}); ok {
		for _, f := range filters {
			if filterMap, ok := f.(map[string]interface{}); ok {
				filter := VisualQueryFilter{}
				if field, ok := filterMap["field"].(string); ok {
					filter.Field = field
				}
				if op, ok := filterMap["operator"].(string); ok {
					filter.Operator = op
				}
				if val, ok := filterMap["value"]; ok {
					filter.Value = val
				}
				if logic, ok := filterMap["logic"].(string); ok {
					filter.Logic = logic
				}
				vq.Filters = append(vq.Filters, filter)
			}
		}
	}

	// Parse joins
	if joins, ok := queryMap["joins"].([]interface{}); ok {
		for _, j := range joins {
			if joinMap, ok := j.(map[string]interface{}); ok {
				join := VisualQueryJoin{}
				if joinType, ok := joinMap["type"].(string); ok {
					join.Type = joinType
				}
				if table, ok := joinMap["table"].(string); ok {
					join.Table = table
				}
				if left, ok := joinMap["leftColumn"].(string); ok {
					join.LeftColumn = left
				}
				if right, ok := joinMap["rightColumn"].(string); ok {
					join.RightColumn = right
				}
				vq.Joins = append(vq.Joins, join)
			}
		}
	}

	// Parse sorts
	if sorts, ok := queryMap["sorts"].([]interface{}); ok {
		for _, s := range sorts {
			if sortMap, ok := s.(map[string]interface{}); ok {
				sort := VisualQuerySort{}
				if field, ok := sortMap["field"].(string); ok {
					sort.Field = field
				}
				if dir, ok := sortMap["direction"].(string); ok {
					sort.Direction = dir
				}
				vq.Sorts = append(vq.Sorts, sort)
			}
		}
	}

	// Parse groupBy
	if groupBy, ok := queryMap["groupBy"].([]interface{}); ok {
		for _, g := range groupBy {
			if field, ok := g.(string); ok {
				vq.GroupBy = append(vq.GroupBy, field)
			}
		}
	}

	// Parse limit and offset
	if limit, ok := queryMap["limit"].(float64); ok {
		vq.Limit = int(limit)
	}
	if offset, ok := queryMap["offset"].(float64); ok {
		vq.Offset = int(offset)
	}

	return vq, nil
}

// GenerateSQL generates SQL from a visual query
func (vq *VisualQuery) GenerateSQL() (string, []interface{}, error) {
	if len(vq.Tables) == 0 {
		return "", nil, fmt.Errorf("at least one table is required")
	}

	if len(vq.Fields) == 0 {
		return "", nil, fmt.Errorf("at least one field is required")
	}

	var sql strings.Builder
	var args []interface{}

	// SELECT clause
	sql.WriteString("SELECT ")
	for i, field := range vq.Fields {
		if i > 0 {
			sql.WriteString(", ")
		}

		if field.Aggregation != "" {
			sql.WriteString(strings.ToUpper(field.Aggregation))
			sql.WriteString("(")
		}

		// Use table.column format if table is specified
		if field.Table != "" {
			sql.WriteString(sanitizeIdentifier(field.Table))
			sql.WriteString(".")
		}
		sql.WriteString(sanitizeIdentifier(field.Column))

		if field.Aggregation != "" {
			sql.WriteString(")")
		}

		if field.Alias != "" {
			sql.WriteString(" AS ")
			sql.WriteString(sanitizeIdentifier(field.Alias))
		}
	}

	// FROM clause
	sql.WriteString(" FROM ")
	sql.WriteString(sanitizeIdentifier(vq.Tables[0]))

	// JOIN clauses
	for _, join := range vq.Joins {
		sql.WriteString(" ")
		sql.WriteString(strings.ToUpper(join.Type))
		sql.WriteString(" JOIN ")
		sql.WriteString(sanitizeIdentifier(join.Table))
		sql.WriteString(" ON ")
		sql.WriteString(sanitizeIdentifier(join.LeftColumn))
		sql.WriteString(" = ")
		sql.WriteString(sanitizeIdentifier(join.RightColumn))
	}

	// WHERE clause
	if len(vq.Filters) > 0 {
		sql.WriteString(" WHERE ")
		for i, filter := range vq.Filters {
			if i > 0 {
				if filter.Logic != "" {
					sql.WriteString(" ")
					sql.WriteString(strings.ToUpper(filter.Logic))
					sql.WriteString(" ")
				} else {
					sql.WriteString(" AND ")
				}
			}

			sql.WriteString(sanitizeIdentifier(filter.Field))
			sql.WriteString(" ")
			sql.WriteString(filter.Operator)
			sql.WriteString(" ?")
			args = append(args, filter.Value)
		}
	}

	// GROUP BY clause
	if len(vq.GroupBy) > 0 {
		sql.WriteString(" GROUP BY ")
		for i, field := range vq.GroupBy {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(sanitizeIdentifier(field))
		}
	}

	// ORDER BY clause
	if len(vq.Sorts) > 0 {
		sql.WriteString(" ORDER BY ")
		for i, sort := range vq.Sorts {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(sanitizeIdentifier(sort.Field))
			sql.WriteString(" ")
			sql.WriteString(strings.ToUpper(sort.Direction))
		}
	}

	// LIMIT and OFFSET
	if vq.Limit > 0 {
		sql.WriteString(" LIMIT ?")
		args = append(args, vq.Limit)
	}
	if vq.Offset > 0 {
		sql.WriteString(" OFFSET ?")
		args = append(args, vq.Offset)
	}

	return sql.String(), args, nil
}

// sanitizeIdentifier sanitizes a database identifier (table or column name)
func sanitizeIdentifier(identifier string) string {
	// Remove any SQL injection attempts
	// Allow only alphanumeric characters, underscore, and dot
	var result strings.Builder
	for _, ch := range identifier {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') || ch == '_' || ch == '.' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
