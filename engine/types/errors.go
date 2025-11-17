package types

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"
)

// TransactionState represents the current state of a transaction
type TransactionState int

const (
	TransactionRunning TransactionState = iota
	TransactionCommitted
	TransactionRolledBack
	TransactionFailed
)

func (s TransactionState) String() string {
	switch s {
	case TransactionRunning:
		return "Running"
	case TransactionCommitted:
		return "Committed"
	case TransactionRolledBack:
		return "RolledBack"
	case TransactionFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// ErrorSeverity indicates the severity level of an error
type ErrorSeverity int

const (
	ErrorSeverityInfo ErrorSeverity = iota
	ErrorSeverityWarning
	ErrorSeverityError
	ErrorSeverityCritical
)

func (s ErrorSeverity) String() string {
	switch s {
	case ErrorSeverityInfo:
		return "INFO"
	case ErrorSeverityWarning:
		return "WARNING"
	case ErrorSeverityError:
		return "ERROR"
	case ErrorSeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// ErrorCategory categorizes different types of errors
type ErrorCategory string

const (
	ErrorCategoryValidation    ErrorCategory = "VALIDATION"
	ErrorCategoryTypeAssertion ErrorCategory = "TYPE_ASSERTION"
	ErrorCategoryDatabase      ErrorCategory = "DATABASE"
	ErrorCategoryExecution     ErrorCategory = "EXECUTION"
	ErrorCategoryTimeout       ErrorCategory = "TIMEOUT"
	ErrorCategoryScript        ErrorCategory = "SCRIPT"
	ErrorCategoryNetwork       ErrorCategory = "NETWORK"
	ErrorCategorySystem        ErrorCategory = "SYSTEM"
	ErrorCategoryBusiness      ErrorCategory = "BUSINESS"
)

// ExecutionContext contains information about where an error occurred
type ExecutionContext struct {
	TranCodeName    string    `json:"trancode_name"`
	TranCodeVersion string    `json:"trancode_version"`
	FunctionGroup   string    `json:"function_group"`
	FunctionName    string    `json:"function_name"`
	FunctionType    string    `json:"function_type"`
	ExecutionTime   time.Time `json:"execution_time"`
	UserNo          string    `json:"user_no"`
	ClientID        string    `json:"client_id"`
}

// BPMError represents a structured error in the BPM engine
type BPMError struct {
	Category       ErrorCategory     `json:"category"`
	Severity       ErrorSeverity     `json:"severity"`
	Message        string            `json:"message"`
	Context        *ExecutionContext `json:"context,omitempty"`
	OriginalError  error             `json:"original_error,omitempty"`
	StackTrace     string            `json:"stack_trace,omitempty"`
	Timestamp      time.Time         `json:"timestamp"`
	RollbackReason string            `json:"rollback_reason,omitempty"`
	Details        map[string]string `json:"details,omitempty"`
}

// Error implements the error interface
func (e *BPMError) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s][%s] %s", e.Severity, e.Category, e.Message))

	if e.Context != nil {
		sb.WriteString(fmt.Sprintf(" | TranCode: %s", e.Context.TranCodeName))
		if e.Context.FunctionGroup != "" {
			sb.WriteString(fmt.Sprintf(" | FuncGroup: %s", e.Context.FunctionGroup))
		}
		if e.Context.FunctionName != "" {
			sb.WriteString(fmt.Sprintf(" | Func: %s", e.Context.FunctionName))
		}
	}

	if e.OriginalError != nil {
		sb.WriteString(fmt.Sprintf(" | Original: %s", e.OriginalError.Error()))
	}

	return sb.String()
}

// Unwrap returns the original error for error wrapping
func (e *BPMError) Unwrap() error {
	return e.OriginalError
}

// NewBPMError creates a new BPM error with context
func NewBPMError(category ErrorCategory, severity ErrorSeverity, message string, originalErr error) *BPMError {
	return &BPMError{
		Category:      category,
		Severity:      severity,
		Message:       message,
		OriginalError: originalErr,
		Timestamp:     time.Now(),
		StackTrace:    string(debug.Stack()),
	}
}

// WithContext adds execution context to the error
func (e *BPMError) WithContext(ctx *ExecutionContext) *BPMError {
	e.Context = ctx
	return e
}

// WithRollbackReason adds a rollback reason to the error
func (e *BPMError) WithRollbackReason(reason string) *BPMError {
	e.RollbackReason = reason
	return e
}

// WithDetail adds a detail key-value pair to the error
func (e *BPMError) WithDetail(key, value string) *BPMError {
	if e.Details == nil {
		e.Details = make(map[string]string)
	}
	e.Details[key] = value
	return e
}

// GetFormattedError returns a formatted error message with full context
func (e *BPMError) GetFormattedError() string {
	var sb strings.Builder

	sb.WriteString("============================================================\n")
	sb.WriteString(fmt.Sprintf("BPM Engine Error - %s\n", e.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString("============================================================\n")
	sb.WriteString(fmt.Sprintf("Severity:  %s\n", e.Severity))
	sb.WriteString(fmt.Sprintf("Category:  %s\n", e.Category))
	sb.WriteString(fmt.Sprintf("Message:   %s\n", e.Message))

	if e.Context != nil {
		sb.WriteString("\nExecution Context:\n")
		sb.WriteString(fmt.Sprintf("  Transaction Code: %s (v%s)\n", e.Context.TranCodeName, e.Context.TranCodeVersion))
		if e.Context.FunctionGroup != "" {
			sb.WriteString(fmt.Sprintf("  Function Group:   %s\n", e.Context.FunctionGroup))
		}
		if e.Context.FunctionName != "" {
			sb.WriteString(fmt.Sprintf("  Function Name:    %s\n", e.Context.FunctionName))
			sb.WriteString(fmt.Sprintf("  Function Type:    %s\n", e.Context.FunctionType))
		}
		sb.WriteString(fmt.Sprintf("  User:             %s\n", e.Context.UserNo))
		sb.WriteString(fmt.Sprintf("  Client ID:        %s\n", e.Context.ClientID))
	}

	if e.RollbackReason != "" {
		sb.WriteString(fmt.Sprintf("\nRollback Reason: %s\n", e.RollbackReason))
	}

	if e.OriginalError != nil {
		sb.WriteString(fmt.Sprintf("\nOriginal Error: %s\n", e.OriginalError.Error()))
	}

	if len(e.Details) > 0 {
		sb.WriteString("\nAdditional Details:\n")
		for k, v := range e.Details {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	if e.StackTrace != "" {
		sb.WriteString("\nStack Trace:\n")
		sb.WriteString(e.StackTrace)
	}

	sb.WriteString("============================================================\n")

	return sb.String()
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(message string, originalErr error) *BPMError {
	return NewBPMError(ErrorCategoryValidation, ErrorSeverityError, message, originalErr)
}

// NewTypeAssertionError creates a type assertion error
func NewTypeAssertionError(expectedType, actualType, variableName string) *BPMError {
	err := NewBPMError(
		ErrorCategoryTypeAssertion,
		ErrorSeverityError,
		fmt.Sprintf("Type assertion failed for variable '%s'", variableName),
		nil,
	)
	err.WithDetail("expected_type", expectedType)
	err.WithDetail("actual_type", actualType)
	err.WithDetail("variable_name", variableName)
	return err
}

// NewDatabaseError creates a database error
func NewDatabaseError(operation, message string, originalErr error) *BPMError {
	err := NewBPMError(ErrorCategoryDatabase, ErrorSeverityCritical, message, originalErr)
	err.WithDetail("operation", operation)
	return err
}

// NewExecutionError creates an execution error
func NewExecutionError(message string, originalErr error) *BPMError {
	return NewBPMError(ErrorCategoryExecution, ErrorSeverityError, message, originalErr)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, timeout time.Duration) *BPMError {
	err := NewBPMError(
		ErrorCategoryTimeout,
		ErrorSeverityError,
		fmt.Sprintf("Operation '%s' timed out", operation),
		nil,
	)
	err.WithDetail("timeout", timeout.String())
	return err
}

// NewScriptError creates a script execution error
func NewScriptError(scriptType, message string, originalErr error) *BPMError {
	err := NewBPMError(ErrorCategoryScript, ErrorSeverityError, message, originalErr)
	err.WithDetail("script_type", scriptType)
	return err
}

// NewBusinessError creates a business logic error
func NewBusinessError(message string) *BPMError {
	return NewBPMError(ErrorCategoryBusiness, ErrorSeverityWarning, message, nil)
}
