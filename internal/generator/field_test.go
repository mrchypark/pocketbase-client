package generator

import (
	"fmt"
	"testing"
)

// TestLatestFieldProcessor_ProcessFields는 최신 스키마 필드 처리기를 테스트합니다.
func TestLatestFieldProcessor_ProcessFields(t *testing.T) {
	processor := NewLatestFieldProcessor()

	tests := []struct {
		name           string
		fields         []FieldSchema
		collectionName string
		expectedCount  int
		expectedFields map[string]string // fieldName -> expectedGoType
		wantErr        bool
	}{
		{
			name: "기본 필드들 처리",
			fields: []FieldSchema{
				{Name: "title", Type: "text", Required: true},
				{Name: "content", Type: "editor", Required: false},
				{Name: "published", Type: "bool", Required: false},
			},
			collectionName: "posts",
			expectedCount:  3,
			expectedFields: map[string]string{
				"title":     "string",
				"content":   "*string",
				"published": "*bool",
			},
			wantErr: false,
		},
		{
			name: "시스템 필드가 명시된 경우",
			fields: []FieldSchema{
				{Name: "id", Type: "text", Required: true},
				{Name: "created", Type: "date", Required: true},
				{Name: "updated", Type: "date", Required: true},
				{Name: "title", Type: "text", Required: true},
			},
			collectionName: "posts",
			expectedCount:  4,
			expectedFields: map[string]string{
				"id":      "string",
				"created": "types.DateTime",
				"updated": "types.DateTime",
				"title":   "string",
			},
			wantErr: false,
		},
		{
			name: "관계 필드 처리",
			fields: []FieldSchema{
				{Name: "author", Type: "relation", Required: true},
				{Name: "tags", Type: "relation", Required: false},
			},
			collectionName: "posts",
			expectedCount:  2,
			expectedFields: map[string]string{
				"author": "[]string",
				"tags":   "[]string",
			},
			wantErr: false,
		},
		{
			name: "파일 필드 처리",
			fields: []FieldSchema{
				{Name: "avatar", Type: "file", Required: false},
				{Name: "attachments", Type: "file", Required: false},
			},
			collectionName: "users",
			expectedCount:  2,
			expectedFields: map[string]string{
				"avatar":      "[]string",
				"attachments": "[]string",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ProcessFields(tt.fields, tt.collectionName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ProcessFields() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ProcessFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("ProcessFields() returned %d fields, expected %d", len(result), tt.expectedCount)
			}

			// 각 필드의 타입 검증
			for _, field := range result {
				expectedType, exists := tt.expectedFields[field.JSONName]
				if !exists {
					t.Errorf("Unexpected field %s in result", field.JSONName)
					continue
				}

				if field.GoType != expectedType {
					t.Errorf("Field %s has type %s, expected %s", field.JSONName, field.GoType, expectedType)
				}

				// GoName이 올바르게 변환되었는지 확인
				expectedGoName := ToPascalCase(field.JSONName)
				if field.GoName != expectedGoName {
					t.Errorf("Field %s has GoName %s, expected %s", field.JSONName, field.GoName, expectedGoName)
				}
			}
		})
	}
}

// TestLatestFieldProcessor_ShouldEmbedBaseDateTime는 최신 스키마에서 BaseDateTime 임베딩 여부를 테스트합니다.
func TestLatestFieldProcessor_ShouldEmbedBaseDateTime(t *testing.T) {
	processor := NewLatestFieldProcessor()

	if processor.ShouldEmbedBaseDateTime() {
		t.Error("LatestFieldProcessor should not embed BaseDateTime")
	}
}

// TestLegacyFieldProcessor_ProcessFields는 구버전 스키마 필드 처리기를 테스트합니다.
func TestLegacyFieldProcessor_ProcessFields(t *testing.T) {
	processor := NewLegacyFieldProcessor()

	tests := []struct {
		name           string
		fields         []FieldSchema
		collectionName string
		expectedCount  int
		expectedFields map[string]string // fieldName -> expectedGoType
		wantErr        bool
	}{
		{
			name: "기본 필드들 처리",
			fields: []FieldSchema{
				{Name: "title", Type: "text", Required: true},
				{Name: "content", Type: "editor", Required: false},
			},
			collectionName: "posts",
			expectedCount:  2,
			expectedFields: map[string]string{
				"title":   "string",
				"content": "*string",
			},
			wantErr: false,
		},
		{
			name: "시스템 필드가 명시된 경우 (기존 정의 우선)",
			fields: []FieldSchema{
				{Name: "id", Type: "text", Required: true},
				{Name: "created", Type: "date", Required: true},
				{Name: "updated", Type: "date", Required: true},
				{Name: "title", Type: "text", Required: true},
			},
			collectionName: "posts",
			expectedCount:  4,
			expectedFields: map[string]string{
				"id":      "string",
				"created": "types.DateTime",
				"updated": "types.DateTime",
				"title":   "string",
			},
			wantErr: false,
		},
		{
			name: "중복 필드 방지",
			fields: []FieldSchema{
				{Name: "title", Type: "text", Required: true},
				{Name: "title", Type: "text", Required: true}, // 중복
			},
			collectionName: "posts",
			expectedCount:  1, // 중복 제거되어 1개만
			expectedFields: map[string]string{
				"title": "string",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ProcessFields(tt.fields, tt.collectionName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ProcessFields() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("ProcessFields() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(result) != tt.expectedCount {
				t.Errorf("ProcessFields() returned %d fields, expected %d", len(result), tt.expectedCount)
			}

			// 각 필드의 타입 검증
			for _, field := range result {
				expectedType, exists := tt.expectedFields[field.JSONName]
				if !exists {
					t.Errorf("Unexpected field %s in result", field.JSONName)
					continue
				}

				if field.GoType != expectedType {
					t.Errorf("Field %s has type %s, expected %s", field.JSONName, field.GoType, expectedType)
				}
			}
		})
	}
}

// TestLegacyFieldProcessor_ShouldEmbedBaseDateTime는 구버전 스키마에서 BaseDateTime 임베딩 여부를 테스트합니다.
func TestLegacyFieldProcessor_ShouldEmbedBaseDateTime(t *testing.T) {
	processor := NewLegacyFieldProcessor()

	if !processor.ShouldEmbedBaseDateTime() {
		t.Error("LegacyFieldProcessor should embed BaseDateTime")
	}
}

// TestCreateFieldProcessor는 스키마 버전에 따른 필드 처리기 생성을 테스트합니다.
func TestCreateFieldProcessor(t *testing.T) {
	tests := []struct {
		name            string
		version         SchemaVersion
		expectedType    string
		shouldEmbedBase bool
	}{
		{
			name:            "최신 스키마 처리기 생성",
			version:         SchemaVersionLatest,
			expectedType:    "*generator.LatestFieldProcessor",
			shouldEmbedBase: false,
		},
		{
			name:            "구버전 스키마 처리기 생성",
			version:         SchemaVersionLegacy,
			expectedType:    "*generator.LegacyFieldProcessor",
			shouldEmbedBase: true,
		},
		{
			name:            "알 수 없는 버전 (기본값으로 최신 처리기)",
			version:         SchemaVersionUnknown,
			expectedType:    "*generator.LatestFieldProcessor",
			shouldEmbedBase: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := CreateFieldProcessor(tt.version)

			if processor == nil {
				t.Error("CreateFieldProcessor() returned nil")
				return
			}

			// BaseDateTime 임베딩 여부 확인
			if processor.ShouldEmbedBaseDateTime() != tt.shouldEmbedBase {
				t.Errorf("CreateFieldProcessor() ShouldEmbedBaseDateTime() = %v, expected %v",
					processor.ShouldEmbedBaseDateTime(), tt.shouldEmbedBase)
			}
		})
	}
}

// TestProcessFieldsWithVersion는 버전별 필드 처리 헬퍼 함수를 테스트합니다.
func TestProcessFieldsWithVersion(t *testing.T) {
	fields := []FieldSchema{
		{Name: "title", Type: "text", Required: true},
		{Name: "content", Type: "editor", Required: false},
		{Name: "created", Type: "date", Required: true},
		{Name: "updated", Type: "date", Required: true},
	}

	tests := []struct {
		name                  string
		version               SchemaVersion
		expectedUseTimestamps bool
		expectedFieldCount    int
	}{
		{
			name:                  "최신 스키마 버전",
			version:               SchemaVersionLatest,
			expectedUseTimestamps: false,
			expectedFieldCount:    4, // 모든 필드 포함 (created, updated도 개별 필드로)
		},
		{
			name:                  "구버전 스키마 버전",
			version:               SchemaVersionLegacy,
			expectedUseTimestamps: true,
			expectedFieldCount:    4, // 모든 필드 포함
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processedFields, useTimestamps, err := ProcessFieldsWithVersion(fields, "posts", tt.version)

			if err != nil {
				t.Errorf("ProcessFieldsWithVersion() error = %v", err)
				return
			}

			if useTimestamps != tt.expectedUseTimestamps {
				t.Errorf("ProcessFieldsWithVersion() useTimestamps = %v, expected %v",
					useTimestamps, tt.expectedUseTimestamps)
			}

			if len(processedFields) != tt.expectedFieldCount {
				t.Errorf("ProcessFieldsWithVersion() returned %d fields, expected %d",
					len(processedFields), tt.expectedFieldCount)
			}
		})
	}
}

// TestFieldProcessor_GetRequiredImports는 필요한 import 목록을 테스트합니다.
func TestFieldProcessor_GetRequiredImports(t *testing.T) {
	tests := []struct {
		name      string
		processor FieldProcessor
		expected  []string
	}{
		{
			name:      "최신 스키마 처리기 imports",
			processor: NewLatestFieldProcessor(),
			expected:  []string{"github.com/pocketbase/pocketbase/tools/types"},
		},
		{
			name:      "구버전 스키마 처리기 imports",
			processor: NewLegacyFieldProcessor(),
			expected:  []string{"github.com/pocketbase/pocketbase/tools/types"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imports := tt.processor.GetRequiredImports()

			if len(imports) != len(tt.expected) {
				t.Errorf("GetRequiredImports() returned %d imports, expected %d",
					len(imports), len(tt.expected))
				return
			}

			for i, imp := range imports {
				if imp != tt.expected[i] {
					t.Errorf("GetRequiredImports()[%d] = %s, expected %s", i, imp, tt.expected[i])
				}
			}
		})
	}
}

// TestFieldData_IsPointer는 FieldData의 포인터 타입 감지를 테스트합니다.
func TestFieldData_IsPointer(t *testing.T) {
	processor := NewLatestFieldProcessor()

	tests := []struct {
		name             string
		field            FieldSchema
		expectedPointer  bool
		expectedBaseType string
	}{
		{
			name:             "필수 필드 (포인터 아님)",
			field:            FieldSchema{Name: "title", Type: "text", Required: true},
			expectedPointer:  false,
			expectedBaseType: "string",
		},
		{
			name:             "선택적 필드 (포인터)",
			field:            FieldSchema{Name: "description", Type: "text", Required: false},
			expectedPointer:  true,
			expectedBaseType: "string",
		},
		{
			name:             "선택적 불린 필드 (포인터)",
			field:            FieldSchema{Name: "published", Type: "bool", Required: false},
			expectedPointer:  true,
			expectedBaseType: "bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fieldData, err := processor.createFieldData(tt.field)
			if err != nil {
				t.Errorf("createFieldData() error = %v", err)
				return
			}

			if fieldData.IsPointer != tt.expectedPointer {
				t.Errorf("FieldData.IsPointer = %v, expected %v", fieldData.IsPointer, tt.expectedPointer)
			}

			if fieldData.BaseType != tt.expectedBaseType {
				t.Errorf("FieldData.BaseType = %s, expected %s", fieldData.BaseType, tt.expectedBaseType)
			}
		})
	}
}

// TestFieldProcessor_SystemFields는 시스템 필드 처리를 테스트합니다.
func TestFieldProcessor_SystemFields(t *testing.T) {
	latestProcessor := NewLatestFieldProcessor()

	systemFields := []string{"id", "created", "updated", "collectionId", "collectionName"}

	for _, fieldName := range systemFields {
		t.Run(fmt.Sprintf("시스템 필드 %s 감지", fieldName), func(t *testing.T) {
			if !latestProcessor.isSystemField(fieldName) {
				t.Errorf("isSystemField(%s) = false, expected true", fieldName)
			}
		})
	}

	// 일반 필드는 시스템 필드가 아님
	normalFields := []string{"title", "content", "author", "tags"}
	for _, fieldName := range normalFields {
		t.Run(fmt.Sprintf("일반 필드 %s는 시스템 필드 아님", fieldName), func(t *testing.T) {
			if latestProcessor.isSystemField(fieldName) {
				t.Errorf("isSystemField(%s) = true, expected false", fieldName)
			}
		})
	}
}
