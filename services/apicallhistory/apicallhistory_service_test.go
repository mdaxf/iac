package apicallhistory

import (
	"fmt"
	"testing"
)

// TestGetPlaceholder tests the placeholder generation for different database types
func TestGetPlaceholder(t *testing.T) {
	// Create a mock service instance
	service := &APICallHistoryService{}

	tests := []struct {
		dbType   string
		index    int
		expected string
	}{
		{"postgres", 1, "$1"},
		{"postgres", 2, "$2"},
		{"postgres", 10, "$10"},
		{"mysql", 1, "?"},
		{"mysql", 5, "?"},
		{"mssql", 1, "?"},
		{"oracle", 2, "?"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_index_%d", tt.dbType, tt.index), func(t *testing.T) {
			result := service.getPlaceholder(tt.dbType, tt.index)
			if result != tt.expected {
				t.Errorf("getPlaceholder(%q, %d) = %q; want %q", tt.dbType, tt.index, result, tt.expected)
			}
		})
	}
}

// TestQueryBuildExamples demonstrates how queries are built for different databases
func TestQueryBuildExamples(t *testing.T) {
	service := &APICallHistoryService{
		historyTable: "iac_api_call_history",
	}

	testCases := []struct {
		dbType      string
		description string
	}{
		{"postgres", "PostgreSQL with numbered placeholders"},
		{"mysql", "MySQL with question mark placeholders"},
		{"mssql", "SQL Server with OFFSET/FETCH syntax"},
		{"oracle", "Oracle with OFFSET/FETCH syntax"},
	}

	for _, tc := range testCases {
		t.Run(tc.dbType, func(t *testing.T) {
			var selectQuery string
			whereClause := "WHERE method = " + service.getPlaceholder(tc.dbType, 1)
			placeholderIndex := 2

			switch tc.dbType {
			case "postgres", "mysql":
				selectQuery = fmt.Sprintf(`
			SELECT id, method, endpoint, status_code, duration_ms, user_id, user_name, source_ip, start_time
			FROM %s %s
			ORDER BY start_time DESC
			LIMIT %s OFFSET %s
		`, service.historyTable, whereClause, service.getPlaceholder(tc.dbType, placeholderIndex), service.getPlaceholder(tc.dbType, placeholderIndex+1))

			case "mssql", "oracle":
				selectQuery = fmt.Sprintf(`
			SELECT id, method, endpoint, status_code, duration_ms, user_id, user_name, source_ip, start_time
			FROM %s %s
			ORDER BY start_time DESC
			OFFSET %s ROWS FETCH NEXT %s ROWS ONLY
		`, service.historyTable, whereClause, service.getPlaceholder(tc.dbType, placeholderIndex), service.getPlaceholder(tc.dbType, placeholderIndex+1))
			}

			// Log the generated query for inspection
			t.Logf("%s:\n%s", tc.description, selectQuery)

			// Verify the query contains appropriate placeholders
			if tc.dbType == "postgres" {
				if !contains(selectQuery, "$2") || !contains(selectQuery, "$3") {
					t.Errorf("PostgreSQL query should contain $2 and $3 placeholders")
				}
			} else if tc.dbType == "mysql" {
				if !contains(selectQuery, "?") {
					t.Errorf("MySQL query should contain ? placeholders")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
