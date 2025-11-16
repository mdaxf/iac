// Copyright 2023 IAC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dbconn

import "errors"

// Configuration Errors
var (
	ErrInvalidDatabaseType = errors.New("invalid database type")
	ErrMissingHost         = errors.New("database host is required")
	ErrMissingDatabase     = errors.New("database name is required")
	ErrMissingUsername     = errors.New("database username is required")
	ErrInvalidPort         = errors.New("invalid database port")
	ErrInvalidConfig       = errors.New("invalid database configuration")
)

// Connection Errors
var (
	ErrConnectionFailed    = errors.New("failed to connect to database")
	ErrConnectionClosed    = errors.New("database connection is closed")
	ErrConnectionTimeout   = errors.New("database connection timeout")
	ErrPingFailed          = errors.New("database ping failed")
	ErrNotConnected        = errors.New("not connected to database")
)

// Query Errors
var (
	ErrQueryFailed         = errors.New("query execution failed")
	ErrInvalidQuery        = errors.New("invalid query")
	ErrNoRows              = errors.New("no rows returned")
	ErrDuplicateKey        = errors.New("duplicate key violation")
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
)

// Transaction Errors
var (
	ErrTransactionFailed = errors.New("transaction failed")
	ErrCommitFailed      = errors.New("transaction commit failed")
	ErrRollbackFailed    = errors.New("transaction rollback failed")
	ErrNoTransaction     = errors.New("no active transaction")
)

// Schema Errors
var (
	ErrTableNotFound   = errors.New("table not found")
	ErrColumnNotFound  = errors.New("column not found")
	ErrSchemaNotFound  = errors.New("schema not found")
	ErrInvalidSchema   = errors.New("invalid schema")
)

// Feature Errors
var (
	ErrFeatureNotSupported = errors.New("feature not supported by this database")
	ErrDialectNotFound     = errors.New("dialect not found for database type")
)

// DatabaseError wraps a database error with additional context
type DatabaseError struct {
	Operation string
	Err       error
	DBType    DBType
	Query     string
}

func (e *DatabaseError) Error() string {
	if e.Query != "" {
		return e.Operation + " failed on " + string(e.DBType) + ": " + e.Err.Error() + " (query: " + e.Query + ")"
	}
	return e.Operation + " failed on " + string(e.DBType) + ": " + e.Err.Error()
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(operation string, err error, dbType DBType, query string) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
		DBType:    dbType,
		Query:     query,
	}
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for connection errors
	if errors.Is(err, ErrConnectionTimeout) ||
		errors.Is(err, ErrConnectionClosed) ||
		errors.Is(err, ErrPingFailed) {
		return true
	}

	return false
}

// IsDuplicateKeyError checks if an error is a duplicate key error
func IsDuplicateKeyError(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}

// IsForeignKeyError checks if an error is a foreign key violation
func IsForeignKeyError(err error) bool {
	return errors.Is(err, ErrForeignKeyViolation)
}
