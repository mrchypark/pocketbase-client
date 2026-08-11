package generator

import (
	"testing"
)

// BenchmarkMapPbTypeToGoType 타입 매핑 성능을 측정합니다
func BenchmarkMapPbTypeToGoType(b *testing.B) {
	testFields := []FieldSchema{
		{Name: "text_field", Type: "text", Required: true},
		{Name: "number_field", Type: "number", Required: false},
		{Name: "bool_field", Type: "bool", Required: true},
		{Name: "date_field", Type: "date", Required: false},
		{Name: "json_field", Type: "json", Required: true},
		{
			Name: "select_field",
			Type: "select",
			Options: &FieldOptions{
				MaxSelect: intPtr(1),
				Values:    []string{"option1", "option2", "option3"},
			},
		},
		{
			Name: "multi_select_field",
			Type: "select",
			Options: &FieldOptions{
				MaxSelect: intPtr(3),
				Values:    []string{"option1", "option2", "option3", "option4"},
			},
		},
		{
			Name: "relation_field",
			Type: "relation",
			Options: &FieldOptions{
				CollectionID: "test_collection_id",
				MaxSelect:    intPtr(1),
			},
		},
		{
			Name: "multi_relation_field",
			Type: "relation",
			Options: &FieldOptions{
				CollectionID: "test_collection_id",
				MaxSelect:    intPtr(5),
			},
		},
		{
			Name: "file_field",
			Type: "file",
			Options: &FieldOptions{
				MaxSelect: intPtr(1),
				Thumbs:    []string{"400x0", "100x0"},
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, field := range testFields {
			_, _ = MapPbTypeToGoType(field, !field.Required)
		}
	}
}

// BenchmarkToPascalCase PascalCase 변환 성능을 측정합니다
func BenchmarkToPascalCase(b *testing.B) {
	testStrings := []string{
		"simple_field",
		"very_long_field_name_with_many_underscores",
		"field-with-dashes",
		"field with spaces",
		"mixed_field-name with spaces",
		"id",
		"url",
		"html",
		"json",
		"field_id",
		"base_url",
		"html_content",
		"json_data",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, str := range testStrings {
			_ = ToPascalCase(str)
		}
	}
}

// BenchmarkToConstantName 상수명 생성 성능을 측정합니다
func BenchmarkToConstantName(b *testing.B) {
	testCases := []struct {
		collection string
		field      string
		value      string
	}{
		{"devices", "type", "m2"},
		{"devices", "type", "d2"},
		{"devices", "status", "active"},
		{"devices", "status", "inactive"},
		{"user_profiles", "account_type", "premium"},
		{"user_profiles", "account_type", "basic"},
		{"content_items", "visibility", "public"},
		{"content_items", "visibility", "private"},
		{"system_logs", "log_level", "error"},
		{"system_logs", "log_level", "warning"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = ToConstantName(tc.collection, tc.field, tc.value)
		}
	}
}
