package generator

import (
	"testing"
)

func TestMapPbTypeToGoType(t *testing.T) {
	tests := []struct {
		name       string
		field      FieldSchema
		omitEmpty  bool
		wantGoType string
	}{
		{
			name:       "text type, not omitEmpty",
			field:      FieldSchema{Type: "text"},
			omitEmpty:  false,
			wantGoType: "string",
		},
		{
			name:       "text type, omitEmpty",
			field:      FieldSchema{Type: "text"},
			omitEmpty:  true,
			wantGoType: "*string",
		},
		{
			name:       "number type, not omitEmpty",
			field:      FieldSchema{Type: "number"},
			omitEmpty:  false,
			wantGoType: "float64",
		},
		{
			name:       "number type, omitEmpty",
			field:      FieldSchema{Type: "number"},
			omitEmpty:  true,
			wantGoType: "*float64",
		},
		{
			name:       "bool type, not omitEmpty",
			field:      FieldSchema{Type: "bool"},
			omitEmpty:  false,
			wantGoType: "bool",
		},
		{
			name:       "bool type, omitEmpty",
			field:      FieldSchema{Type: "bool"},
			omitEmpty:  true,
			wantGoType: "*bool",
		},
		{
			name:       "date type, not omitEmpty",
			field:      FieldSchema{Type: "date"},
			omitEmpty:  false,
			wantGoType: "types.DateTime",
		},
		{
			name:       "date type, omitEmpty",
			field:      FieldSchema{Type: "date"},
			omitEmpty:  true,
			wantGoType: "*types.DateTime",
		},
		{
			name:       "json type, not omitEmpty",
			field:      FieldSchema{Type: "json"},
			omitEmpty:  false,
			wantGoType: "json.RawMessage",
		},
		{
			name:       "json type, omitEmpty",
			field:      FieldSchema{Type: "json"},
			omitEmpty:  true,
			wantGoType: "json.RawMessage", // json.RawMessage should not be a pointer
		},
		{
			name:       "relation type (single), not omitEmpty",
			field:      FieldSchema{Type: "relation", Options: &FieldOptions{MaxSelect: func() *int { i := 1; return &i }()}},
			omitEmpty:  false,
			wantGoType: "string",
		},
		{
			name:       "relation type (single), omitEmpty",
			field:      FieldSchema{Type: "relation", Options: &FieldOptions{MaxSelect: func() *int { i := 1; return &i }()}},
			omitEmpty:  true,
			wantGoType: "*string",
		},
		{
			name:       "relation type (multiple), not omitEmpty",
			field:      FieldSchema{Type: "relation", Options: &FieldOptions{MaxSelect: func() *int { i := 2; return &i }()}},
			omitEmpty:  false,
			wantGoType: "[]string",
		},
		{
			name:       "relation type (multiple), omitEmpty",
			field:      FieldSchema{Type: "relation", Options: &FieldOptions{MaxSelect: func() *int { i := 2; return &i }()}},
			omitEmpty:  true,
			wantGoType: "[]string", // slices should not be pointers
		},
		{
			name:       "select type (single), not omitEmpty",
			field:      FieldSchema{Type: "select", Options: &FieldOptions{MaxSelect: func() *int { i := 1; return &i }()}},
			omitEmpty:  false,
			wantGoType: "string",
		},
		{
			name:       "select type (single), omitEmpty",
			field:      FieldSchema{Type: "select", Options: &FieldOptions{MaxSelect: func() *int { i := 1; return &i }()}},
			omitEmpty:  true,
			wantGoType: "*string",
		},
		{
			name:       "select type (multiple), not omitEmpty",
			field:      FieldSchema{Type: "select", Options: &FieldOptions{MaxSelect: func() *int { i := 2; return &i }()}},
			omitEmpty:  false,
			wantGoType: "[]string",
		},
		{
			name:       "select type (multiple), omitEmpty",
			field:      FieldSchema{Type: "select", Options: &FieldOptions{MaxSelect: func() *int { i := 2; return &i }()}},
			omitEmpty:  true,
			wantGoType: "[]string", // slices should not be pointers
		},
		{
			name:       "unknown type, not omitEmpty",
			field:      FieldSchema{Type: "unknown"},
			omitEmpty:  false,
			wantGoType: "interface{}",
		},
		{
			name:       "unknown type, omitEmpty",
			field:      FieldSchema{Type: "unknown"},
			omitEmpty:  true,
			wantGoType: "interface{}", // interface{} should not be a pointer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goType, _ := MapPbTypeToGoType(tt.field, tt.omitEmpty)
			if goType != tt.wantGoType {
				t.Errorf("MapPbTypeToGoType() got %q, want %q", goType, tt.wantGoType)
			}
		})
	}
}
