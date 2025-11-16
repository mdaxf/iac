// Copyright 2023 IAC. All Rights Reserved.

package databases

import (
	"os"
	"testing"
)

func TestValidateIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantErr    bool
	}{
		{"Valid simple name", "users", false},
		{"Valid with underscore", "user_table", false},
		{"Valid starting with underscore", "_internal", false},
		{"Valid with numbers", "table123", false},
		{"Invalid empty", "", true},
		{"Invalid starting with number", "123table", true},
		{"Invalid with space", "user table", true},
		{"Invalid with dash", "user-table", true},
		{"Invalid with dot", "schema.table", true},
		{"Invalid SQL keyword DROP", "DROP", true},
		{"Invalid SQL keyword DELETE", "DELETE", true},
		{"Invalid too long", string(make([]byte, 64)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIdentifier(tt.identifier)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIdentifier() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSSLConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  DBConfig
		env     string
		wantErr bool
	}{
		{
			name: "PostgreSQL production with SSL",
			config: DBConfig{
				Type:    "postgres",
				Options: map[string]string{"sslmode": "require"},
			},
			env:     "production",
			wantErr: false,
		},
		{
			name: "PostgreSQL production without SSL",
			config: DBConfig{
				Type:    "postgres",
				Options: map[string]string{"sslmode": "disable"},
			},
			env:     "production",
			wantErr: true,
		},
		{
			name: "PostgreSQL development without SSL",
			config: DBConfig{
				Type:    "postgres",
				Options: map[string]string{"sslmode": "disable"},
			},
			env:     "development",
			wantErr: false,
		},
		{
			name: "MySQL production",
			config: DBConfig{
				Type:    "mysql",
				Options: map[string]string{"tls": "true"},
			},
			env:     "production",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ENV", tt.env)
			defer os.Unsetenv("ENV")

			err := ValidateSSLConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSSLConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConnectionConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  DBConfig
		wantErr bool
	}{
		{
			name: "Valid configuration",
			config: DBConfig{
				Type:         "mysql",
				Host:         "localhost",
				Port:         3306,
				Database:     "testdb",
				Username:     "testuser",
				Password:     "testpassword123",
				ConnTimeout:  30,
				MaxIdleConns: 5,
				MaxOpenConns: 10,
				Options:      make(map[string]string),
			},
			wantErr: false,
		},
		{
			name: "Invalid database name",
			config: DBConfig{
				Type:         "mysql",
				Database:     "test-db", // Invalid: contains dash
				Username:     "testuser",
				Password:     "testpassword",
				ConnTimeout:  30,
				MaxIdleConns: 5,
				MaxOpenConns: 10,
				Options:      make(map[string]string),
			},
			wantErr: true,
		},
		{
			name: "Empty username",
			config: DBConfig{
				Type:         "mysql",
				Database:     "testdb",
				Username:     "", // Invalid
				Password:     "testpassword",
				ConnTimeout:  30,
				MaxIdleConns: 5,
				MaxOpenConns: 10,
				Options:      make(map[string]string),
			},
			wantErr: true,
		},
		{
			name: "Invalid connection pool",
			config: DBConfig{
				Type:         "mysql",
				Database:     "testdb",
				Username:     "testuser",
				Password:     "testpassword",
				ConnTimeout:  30,
				MaxIdleConns: 10, // Greater than MaxOpenConns
				MaxOpenConns: 5,
				Options:      make(map[string]string),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectionConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConnectionConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeErrorMessage(t *testing.T) {
	config := DBConfig{
		Username: "testuser",
		Password: "secretpassword",
	}

	tests := []struct {
		name     string
		err      error
		want     string
		notWant  string
	}{
		{
			name:    "Password in error",
			err:     &DatabaseError{Message: "connection failed: password 'secretpassword' incorrect"},
			notWant: "secretpassword",
		},
		{
			name:    "Username and password in error",
			err:     &DatabaseError{Message: "auth failed for testuser:secretpassword"},
			notWant: "secretpassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeErrorMessage(tt.err, config)
			if tt.notWant != "" && containsString(got, tt.notWant) {
				t.Errorf("SanitizeErrorMessage() contains sensitive data: %s", tt.notWant)
			}
		})
	}
}

func TestDefaultSecurityConfig(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want *SecurityConfig
	}{
		{
			name: "Production environment",
			env:  "production",
			want: &SecurityConfig{
				EnforceSSL:          true,
				ValidateIdentifiers: true,
				LogQueries:          false,
				LogParameters:       false,
				EnableAuditLog:      true,
				Environment:         "production",
			},
		},
		{
			name: "Development environment",
			env:  "development",
			want: &SecurityConfig{
				EnforceSSL:          false,
				ValidateIdentifiers: true,
				LogQueries:          true,
				LogParameters:       false,
				EnableAuditLog:      false,
				Environment:         "development",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ENV", tt.env)
			defer os.Unsetenv("ENV")

			got := DefaultSecurityConfig()
			if got.EnforceSSL != tt.want.EnforceSSL {
				t.Errorf("EnforceSSL = %v, want %v", got.EnforceSSL, tt.want.EnforceSSL)
			}
			if got.EnableAuditLog != tt.want.EnableAuditLog {
				t.Errorf("EnableAuditLog = %v, want %v", got.EnableAuditLog, tt.want.EnableAuditLog)
			}
		})
	}
}

func TestValidateSecurityConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *SecurityConfig
		wantErr bool
	}{
		{
			name:    "Nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "Valid config",
			config: &SecurityConfig{
				MaxQueryDuration: 30,
				RateLimitQPS:     100,
			},
			wantErr: false,
		},
		{
			name: "Invalid max query duration",
			config: &SecurityConfig{
				MaxQueryDuration: 0,
				RateLimitQPS:     100,
			},
			wantErr: true,
		},
		{
			name: "Invalid rate limit",
			config: &SecurityConfig{
				MaxQueryDuration: 30,
				RateLimitQPS:     -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecurityConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecurityConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(substr) > 0 && len(s) > 0 && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
