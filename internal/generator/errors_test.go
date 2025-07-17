package generator

import (
	"errors"
	"testing"
)

func TestGenerationError(t *testing.T) {
	t.Run("기본 에러 생성", func(t *testing.T) {
		err := NewGenerationError(ErrorTypeSchemaLoad, "test error", nil)

		if err.Type != ErrorTypeSchemaLoad {
			t.Errorf("expected error type %s, got %s", ErrorTypeSchemaLoad, err.Type)
		}

		if err.Message != "test error" {
			t.Errorf("expected message 'test error', got '%s'", err.Message)
		}

		expected := "[SCHEMA_LOAD] test error"
		if err.Error() != expected {
			t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
		}
	})

	t.Run("세부 정보가 있는 에러", func(t *testing.T) {
		err := NewGenerationError(ErrorTypeFileRead, "failed to read", nil).
			WithDetail("file_path", "/test/path").
			WithDetail("operation", "read")

		errorStr := err.Error()
		if !contains(errorStr, "file_path=/test/path") {
			t.Errorf("error string should contain file_path detail: %s", errorStr)
		}

		if !contains(errorStr, "operation=read") {
			t.Errorf("error string should contain operation detail: %s", errorStr)
		}
	})

	t.Run("원인 에러가 있는 경우", func(t *testing.T) {
		cause := errors.New("original error")
		err := NewGenerationError(ErrorTypeTemplateParse, "template failed", cause)

		if err.Unwrap() != cause {
			t.Error("Unwrap should return the original cause")
		}
	})

	t.Run("에러 타입 비교", func(t *testing.T) {
		err1 := NewGenerationError(ErrorTypeSchemaLoad, "error 1", nil)
		err2 := NewGenerationError(ErrorTypeSchemaLoad, "error 2", nil)
		err3 := NewGenerationError(ErrorTypeFileRead, "error 3", nil)

		if !err1.Is(err2) {
			t.Error("errors with same type should match")
		}

		if err1.Is(err3) {
			t.Error("errors with different types should not match")
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Run("빈 검증 에러", func(t *testing.T) {
		ve := NewValidationError()

		if ve.HasErrors() {
			t.Error("new validation error should not have errors")
		}

		if ve.HasWarnings() {
			t.Error("new validation error should not have warnings")
		}
	})

	t.Run("에러 추가", func(t *testing.T) {
		ve := NewValidationError()
		ve.AddError("test error", "test.path")

		if !ve.HasErrors() {
			t.Error("should have errors after adding error")
		}

		errors := ve.GetErrors()
		if len(errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(errors))
		}

		if errors[0].Message != "test error" {
			t.Errorf("expected message 'test error', got '%s'", errors[0].Message)
		}

		if errors[0].Path != "test.path" {
			t.Errorf("expected path 'test.path', got '%s'", errors[0].Path)
		}
	})

	t.Run("경고 추가", func(t *testing.T) {
		ve := NewValidationError()
		ve.AddWarning("test warning", "test.path")

		if !ve.HasWarnings() {
			t.Error("should have warnings after adding warning")
		}

		warnings := ve.GetWarnings()
		if len(warnings) != 1 {
			t.Errorf("expected 1 warning, got %d", len(warnings))
		}
	})

	t.Run("복합 검증 에러", func(t *testing.T) {
		ve := NewValidationError()
		ve.AddError("error 1", "path1")
		ve.AddError("error 2", "path2")
		ve.AddWarning("warning 1", "path3")

		if !ve.HasErrors() {
			t.Error("should have errors")
		}

		if !ve.HasWarnings() {
			t.Error("should have warnings")
		}

		errorStr := ve.Error()
		if !contains(errorStr, "validation failed with 2 error(s) and 1 warning(s)") {
			t.Errorf("error string should contain summary: %s", errorStr)
		}
	})

	t.Run("단일 에러 메시지", func(t *testing.T) {
		ve := NewValidationError()
		ve.AddError("single error", "path")

		errorStr := ve.Error()
		expected := "validation failed: single error"
		if errorStr != expected {
			t.Errorf("expected '%s', got '%s'", expected, errorStr)
		}
	})
}

func TestRecoverableError(t *testing.T) {
	t.Run("복구 가능한 에러 생성", func(t *testing.T) {
		genErr := NewGenerationError(ErrorTypeSchemaLoad, "test error", nil)
		recovery := &RecoveryInfo{
			CanRecover:  true,
			Suggestions: []string{"suggestion 1", "suggestion 2"},
		}

		recErr := NewRecoverableError(genErr, recovery)

		if recErr.GetRecoveryInfo() != recovery {
			t.Error("recovery info should match")
		}

		if !recErr.GetRecoveryInfo().CanRecover {
			t.Error("should be recoverable")
		}

		if len(recErr.GetRecoveryInfo().Suggestions) != 2 {
			t.Errorf("expected 2 suggestions, got %d", len(recErr.GetRecoveryInfo().Suggestions))
		}
	})
}

func TestErrorHelpers(t *testing.T) {
	t.Run("WrapSchemaError", func(t *testing.T) {
		cause := errors.New("original error")
		err := WrapSchemaError(cause, "/test/schema.json", "load")

		if err.Type != ErrorTypeSchemaLoad {
			t.Errorf("expected type %s, got %s", ErrorTypeSchemaLoad, err.Type)
		}

		if err.Cause != cause {
			t.Error("cause should be preserved")
		}

		if err.Details["schema_path"] != "/test/schema.json" {
			t.Error("schema_path detail should be set")
		}

		if err.Details["operation"] != "load" {
			t.Error("operation detail should be set")
		}
	})

	t.Run("WrapTemplateError", func(t *testing.T) {
		cause := errors.New("template error")
		err := WrapTemplateError(cause, "test.tpl", "parse")

		if err.Type != ErrorTypeTemplateParse {
			t.Errorf("expected type %s, got %s", ErrorTypeTemplateParse, err.Type)
		}

		if err.Details["template_name"] != "test.tpl" {
			t.Error("template_name detail should be set")
		}
	})

	t.Run("WrapFileError", func(t *testing.T) {
		cause := errors.New("file error")

		// Test create operation
		createErr := WrapFileError(cause, "/test/file.go", "create")
		if createErr.Type != ErrorTypeFileCreate {
			t.Errorf("expected type %s for create, got %s", ErrorTypeFileCreate, createErr.Type)
		}

		// Test write operation
		writeErr := WrapFileError(cause, "/test/file.go", "write")
		if writeErr.Type != ErrorTypeFileWrite {
			t.Errorf("expected type %s for write, got %s", ErrorTypeFileWrite, writeErr.Type)
		}

		// Test read operation (default)
		readErr := WrapFileError(cause, "/test/file.go", "read")
		if readErr.Type != ErrorTypeFileRead {
			t.Errorf("expected type %s for read, got %s", ErrorTypeFileRead, readErr.Type)
		}
	})

	t.Run("CreateConfigError", func(t *testing.T) {
		details := map[string]interface{}{
			"config_key": "config_value",
			"number":     42,
		}

		err := CreateConfigError("invalid configuration", details)

		if err.Type != ErrorTypeInvalidConfig {
			t.Errorf("expected type %s, got %s", ErrorTypeInvalidConfig, err.Type)
		}

		if err.Details["config_key"] != "config_value" {
			t.Error("config_key detail should be set")
		}

		if err.Details["number"] != 42 {
			t.Error("number detail should be set")
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
