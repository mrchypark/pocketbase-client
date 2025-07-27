package generator

import (
	"reflect"
	"strings"
	"testing"
)

func TestEnumGenerator_GenerateEnums(t *testing.T) {
	generator := NewEnumGenerator()

	// Prepare test data
	collections := []CollectionData{
		{
			CollectionName: "devices",
		},
		{
			CollectionName: "users",
		},
	}

	schemas := []CollectionSchema{
		{
			Name: "devices",
			Fields: []FieldSchema{
				{
					Name:   "status",
					Type:   "select",
					System: false,
					Options: &FieldOptions{
						Values: []string{"active", "inactive", "pending"},
					},
				},
				{
					Name:   "type",
					Type:   "select",
					System: false,
					Options: &FieldOptions{
						Values: []string{"sensor", "actuator"},
					},
				},
				{
					Name:   "name",
					Type:   "text",
					System: false,
				},
			},
		},
		{
			Name: "users",
			Fields: []FieldSchema{
				{
					Name:   "role",
					Type:   "select",
					System: false,
					Options: &FieldOptions{
						Values: []string{"admin", "user", "guest"},
					},
				},
				{
					Name:   "id",
					Type:   "text",
					System: true, // System field should be ignored
				},
			},
		},
	}

	result := generator.GenerateEnums(collections, schemas)

	// Expected result: 3 enums (devices.status, devices.type, users.role)
	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("Expected %d enums, got %d", expectedCount, len(result))
	}

	// Verify devices.status enum
	var deviceStatusEnum *EnumData
	for i := range result {
		if result[i].CollectionName == "devices" && result[i].FieldName == "status" {
			deviceStatusEnum = &result[i]
			break
		}
	}

	if deviceStatusEnum == nil {
		t.Fatal("devices.status enum not found")
	}

	expectedConstants := []ConstantData{
		{Name: "DevicesStatusActive", Value: "active"},
		{Name: "DevicesStatusInactive", Value: "inactive"},
		{Name: "DevicesStatusPending", Value: "pending"},
	}

	if !reflect.DeepEqual(deviceStatusEnum.Constants, expectedConstants) {
		t.Errorf("devices.status constants mismatch.\nGot: %+v\nWant: %+v",
			deviceStatusEnum.Constants, expectedConstants)
	}
}

func TestEnumGenerator_GenerateEnumData(t *testing.T) {
	generator := NewEnumGenerator()

	enhanced := EnhancedFieldInfo{
		FieldSchema: FieldSchema{
			Name: "priority",
			Type: "select",
		},
		EnumValues:   []string{"high", "medium", "low"},
		EnumTypeName: "TasksPriorityType",
	}

	result := generator.GenerateEnumData(enhanced, "tasks")

	expected := EnumData{
		CollectionName: "tasks",
		FieldName:      "priority",
		EnumTypeName:   "TasksPriorityType",
		Constants: []ConstantData{
			{Name: "TasksPriorityHigh", Value: "high"},
			{Name: "TasksPriorityMedium", Value: "medium"},
			{Name: "TasksPriorityLow", Value: "low"},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("GenerateEnumData() = %+v, want %+v", result, expected)
	}
}

func TestEnumGenerator_GenerateEnumConstants(t *testing.T) {
	generator := NewEnumGenerator()

	tests := []struct {
		name           string
		field          EnhancedFieldInfo
		collectionName string
		want           []ConstantData
	}{
		{
			name: "normal enum field",
			field: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "status",
				},
				EnumValues: []string{"draft", "published", "archived"},
			},
			collectionName: "posts",
			want: []ConstantData{
				{Name: "PostsStatusDraft", Value: "draft"},
				{Name: "PostsStatusPublished", Value: "published"},
				{Name: "PostsStatusArchived", Value: "archived"},
			},
		},
		{
			name: "empty enum values",
			field: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "category",
				},
				EnumValues: []string{},
			},
			collectionName: "items",
			want:           nil,
		},
		{
			name: "special characters in values",
			field: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "type",
				},
				EnumValues: []string{"type-a", "type_b", "type c"},
			},
			collectionName: "products",
			want: []ConstantData{
				{Name: "ProductsTypeTypeA", Value: "type-a"},
				{Name: "ProductsTypeTypeB", Value: "type_b"},
				{Name: "ProductsTypeTypeC", Value: "type c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generator.GenerateEnumConstants(tt.field, tt.collectionName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateEnumConstants() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnumGenerator_GenerateEnumHelperFunction(t *testing.T) {
	generator := NewEnumGenerator()

	enumData := EnumData{
		EnumTypeName: "UserRoleType",
		Constants: []ConstantData{
			{Name: "UserRoleAdmin", Value: "admin"},
			{Name: "UserRoleUser", Value: "user"},
			{Name: "UserRoleGuest", Value: "guest"},
		},
	}

	result := generator.GenerateEnumHelperFunction(enumData)

	// Verify that generated function has correct format
	expectedParts := []string{
		"func UserRoleTypeValues() []string",
		"return []string{UserRoleAdmin, UserRoleUser, UserRoleGuest}",
		"// UserRoleTypeValues returns all possible values for UserRoleType",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Generated helper function missing expected part: %s\nGenerated:\n%s", part, result)
		}
	}
}

func TestEnumGenerator_GenerateEnumValidationFunction(t *testing.T) {
	generator := NewEnumGenerator()

	enumData := EnumData{
		EnumTypeName: "StatusType",
		Constants: []ConstantData{
			{Name: "StatusActive", Value: "active"},
			{Name: "StatusInactive", Value: "inactive"},
		},
	}

	result := generator.GenerateEnumValidationFunction(enumData)

	// Verify that generated function has correct format
	expectedParts := []string{
		"func IsValidStatusType(value string) bool",
		"case StatusActive:",
		"case StatusInactive:",
		"return true",
		"return false",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Generated validation function missing expected part: %s\nGenerated:\n%s", part, result)
		}
	}
}

func TestEnumGenerator_ValidateEnumName(t *testing.T) {
	generator := NewEnumGenerator()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid name",
			input: "UserStatus",
			want:  "UserStatus",
		},
		{
			name:  "name with special characters",
			input: "User-Status_Type!",
			want:  "UserStatusType",
		},
		{
			name:  "name starting with number",
			input: "123Status",
			want:  "Enum123Status",
		},
		{
			name:  "empty name",
			input: "",
			want:  "EnumValue",
		},
		{
			name:  "only special characters",
			input: "!@#$%",
			want:  "EnumValue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generator.ValidateEnumName(tt.input)
			if got != tt.want {
				t.Errorf("ValidateEnumName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnumGenerator_SanitizeConstantValue(t *testing.T) {
	generator := NewEnumGenerator()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple value",
			input: "active",
			want:  `"active"`,
		},
		{
			name:  "value with quotes",
			input: `value with "quotes"`,
			want:  `"value with \"quotes\""`,
		},
		{
			name:  "value with backslashes",
			input: `path\to\file`,
			want:  `"path\\to\\file"`,
		},
		{
			name:  "value with newlines",
			input: "line1\nline2",
			want:  `"line1\nline2"`,
		},
		{
			name:  "value with tabs",
			input: "col1\tcol2",
			want:  `"col1\tcol2"`,
		},
		{
			name:  "value with carriage returns",
			input: "line1\rline2",
			want:  `"line1\rline2"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generator.SanitizeConstantValue(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeConstantValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEnumGenerator(t *testing.T) {
	generator := NewEnumGenerator()
	if generator == nil {
		t.Error("NewEnumGenerator() returned nil")
	}
}
