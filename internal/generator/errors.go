package generator

import (
	"fmt"
	"strings"
)

// GenerationError represents different types of errors that can occur during code generation.
type GenerationError struct {
	Type    ErrorType
	Message string
	Details map[string]interface{}
	Cause   error
}

// ErrorType represents the category of generation error.
type ErrorType string

const (
	// Schema-related errors
	ErrorTypeSchemaLoad     ErrorType = "schema_load"
	ErrorTypeSchemaValidate ErrorType = "schema_validate"
	ErrorTypeSchemaParse    ErrorType = "schema_parse"

	// Template-related errors
	ErrorTypeTemplateParse   ErrorType = "template_parse"
	ErrorTypeTemplateExecute ErrorType = "template_execute"

	// File operation errors
	ErrorTypeFileCreate ErrorType = "file_create"
	ErrorTypeFileWrite  ErrorType = "file_write"
	ErrorTypeFileRead   ErrorType = "file_read"

	// Code generation errors
	ErrorTypeCodeFormat     ErrorType = "code_format"
	ErrorTypeCodeGeneration ErrorType = "code_generation"
	ErrorTypeTypeMapping    ErrorType = "type_mapping"
	ErrorTypeNameConflict   ErrorType = "name_conflict"

	// Configuration errors
	ErrorTypeInvalidConfig ErrorType = "invalid_config"
	ErrorTypeInvalidPath   ErrorType = "invalid_path"
)

// Error implements the error interface.
func (e *GenerationError) Error() string {
	var parts []string

	// Add error type prefix
	parts = append(parts, fmt.Sprintf("[%s]", strings.ToUpper(string(e.Type))))

	// Add main message
	parts = append(parts, e.Message)

	// Add details if available
	if len(e.Details) > 0 {
		var details []string
		for key, value := range e.Details {
			details = append(details, fmt.Sprintf("%s=%v", key, value))
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(details, ", ")))
	}

	return strings.Join(parts, " ")
}

// Unwrap returns the underlying cause error.
func (e *GenerationError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error type.
func (e *GenerationError) Is(target error) bool {
	if targetErr, ok := target.(*GenerationError); ok {
		return e.Type == targetErr.Type
	}
	return false
}

// NewGenerationError creates a new GenerationError.
func NewGenerationError(errorType ErrorType, message string, cause error) *GenerationError {
	return &GenerationError{
		Type:    errorType,
		Message: message,
		Details: make(map[string]interface{}),
		Cause:   cause,
	}
}

// WithDetail adds a detail to the error.
func (e *GenerationError) WithDetail(key string, value interface{}) *GenerationError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithDetails adds multiple details to the error.
func (e *GenerationError) WithDetails(details map[string]interface{}) *GenerationError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	for k, v := range details {
		e.Details[k] = v
	}
	return e
}

// ValidationError represents validation errors with multiple issues.
type ValidationError struct {
	Issues []ValidationIssue
}

// ValidationIssue represents a single validation issue.
type ValidationIssue struct {
	Type       string      `json:"type"`
	Message    string      `json:"message"`
	Path       string      `json:"path,omitempty"`
	Suggestion string      `json:"suggestion,omitempty"`
	Severity   Severity    `json:"severity"`
	Context    interface{} `json:"context,omitempty"`
}

// Severity represents the severity level of a validation issue.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Error implements the error interface for ValidationError.
func (ve *ValidationError) Error() string {
	if len(ve.Issues) == 0 {
		return "validation failed with no specific issues"
	}

	if len(ve.Issues) == 1 {
		return fmt.Sprintf("validation failed: %s", ve.Issues[0].Message)
	}

	var messages []string
	errorCount := 0
	warningCount := 0

	for _, issue := range ve.Issues {
		switch issue.Severity {
		case SeverityError:
			errorCount++
		case SeverityWarning:
			warningCount++
		}

		prefix := strings.ToUpper(string(issue.Severity))
		if issue.Path != "" {
			messages = append(messages, fmt.Sprintf("[%s] %s: %s", prefix, issue.Path, issue.Message))
		} else {
			messages = append(messages, fmt.Sprintf("[%s] %s", prefix, issue.Message))
		}
	}

	summary := fmt.Sprintf("validation failed with %d error(s)", errorCount)
	if warningCount > 0 {
		summary += fmt.Sprintf(" and %d warning(s)", warningCount)
	}

	return fmt.Sprintf("%s:\n%s", summary, strings.Join(messages, "\n"))
}

// HasErrors returns true if there are any error-level issues.
func (ve *ValidationError) HasErrors() bool {
	for _, issue := range ve.Issues {
		if issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warning-level issues.
func (ve *ValidationError) HasWarnings() bool {
	for _, issue := range ve.Issues {
		if issue.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// GetErrors returns only error-level issues.
func (ve *ValidationError) GetErrors() []ValidationIssue {
	var errors []ValidationIssue
	for _, issue := range ve.Issues {
		if issue.Severity == SeverityError {
			errors = append(errors, issue)
		}
	}
	return errors
}

// GetWarnings returns only warning-level issues.
func (ve *ValidationError) GetWarnings() []ValidationIssue {
	var warnings []ValidationIssue
	for _, issue := range ve.Issues {
		if issue.Severity == SeverityWarning {
			warnings = append(warnings, issue)
		}
	}
	return warnings
}

// AddIssue adds a validation issue.
func (ve *ValidationError) AddIssue(issue ValidationIssue) {
	ve.Issues = append(ve.Issues, issue)
}

// AddError adds an error-level validation issue.
func (ve *ValidationError) AddError(message, path string) {
	ve.AddIssue(ValidationIssue{
		Type:     "validation_error",
		Message:  message,
		Path:     path,
		Severity: SeverityError,
	})
}

// AddWarning adds a warning-level validation issue.
func (ve *ValidationError) AddWarning(message, path string) {
	ve.AddIssue(ValidationIssue{
		Type:     "validation_warning",
		Message:  message,
		Path:     path,
		Severity: SeverityWarning,
	})
}

// NewValidationError creates a new ValidationError.
func NewValidationError() *ValidationError {
	return &ValidationError{
		Issues: make([]ValidationIssue, 0),
	}
}

// RecoveryInfo provides information for error recovery.
type RecoveryInfo struct {
	CanRecover    bool                   `json:"can_recover"`
	Suggestions   []string               `json:"suggestions"`
	PartialResult interface{}            `json:"partial_result,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

// RecoverableError represents an error that might be recoverable.
type RecoverableError struct {
	*GenerationError
	Recovery *RecoveryInfo
}

// NewRecoverableError creates a new RecoverableError.
func NewRecoverableError(genErr *GenerationError, recovery *RecoveryInfo) *RecoverableError {
	return &RecoverableError{
		GenerationError: genErr,
		Recovery:        recovery,
	}
}

// GetRecoveryInfo returns recovery information.
func (re *RecoverableError) GetRecoveryInfo() *RecoveryInfo {
	return re.Recovery
}

// Common error creation helpers

// WrapSchemaError wraps a schema-related error with additional context.
func WrapSchemaError(err error, schemaPath string, operation string) *GenerationError {
	return NewGenerationError(ErrorTypeSchemaLoad,
		fmt.Sprintf("failed to %s schema", operation), err).
		WithDetail("schema_path", schemaPath).
		WithDetail("operation", operation)
}

// WrapTemplateError wraps a template-related error with additional context.
func WrapTemplateError(err error, templateName string, operation string) *GenerationError {
	return NewGenerationError(ErrorTypeTemplateParse,
		fmt.Sprintf("failed to %s template", operation), err).
		WithDetail("template_name", templateName).
		WithDetail("operation", operation)
}

// WrapFileError wraps a file operation error with additional context.
func WrapFileError(err error, filePath string, operation string) *GenerationError {
	errorType := ErrorTypeFileRead
	switch operation {
	case "create":
		errorType = ErrorTypeFileCreate
	case "write":
		errorType = ErrorTypeFileWrite
	}

	return NewGenerationError(errorType,
		fmt.Sprintf("failed to %s file", operation), err).
		WithDetail("file_path", filePath).
		WithDetail("operation", operation)
}

// CreateConfigError creates a configuration error.
func CreateConfigError(message string, details map[string]interface{}) *GenerationError {
	return NewGenerationError(ErrorTypeInvalidConfig, message, nil).
		WithDetails(details)
}
