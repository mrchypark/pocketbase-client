package generator

import (
	"reflect"
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
			wantGoType: "any",
		},
		{
			name:       "unknown type, omitEmpty",
			field:      FieldSchema{Type: "unknown"},
			omitEmpty:  true,
			wantGoType: "any", // any should not be a pointer
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goType, _, _ := MapPbTypeToGoType(tt.field, tt.omitEmpty)
			if goType != tt.wantGoType {
				t.Errorf("MapPbTypeToGoType() got %q, want %q", goType, tt.wantGoType)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty string", "", ""},
		{"simple word", "hello", "Hello"},
		{"snake_case", "hello_world", "HelloWorld"},
		{"kebab-case", "hello-world", "HelloWorld"},
		{"space separated", "hello world", "HelloWorld"},
		{"already PascalCase", "HelloWorld", "HelloWorld"},
		{"special acronym id", "id", "ID"},
		{"special acronym url", "url", "URL"},
		{"special acronym html", "html", "HTML"},
		{"special acronym json", "json", "JSON"},
		{"mixed case with acronym", "some_id_field", "SomeIDField"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToPascalCase(tt.input); got != tt.want {
				t.Errorf("ToPascalCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAnalyzeEnhancedField tests the AnalyzeEnhancedField function
func TestAnalyzeEnhancedField(t *testing.T) {
	// Create test collection schema
	allCollections := []CollectionSchema{
		{
			ID:   "plants_collection_id",
			Name: "plants",
		},
		{
			ID:   "users_collection_id",
			Name: "users",
		},
	}

	tests := []struct {
		name           string
		field          FieldSchema
		collectionName string
		want           EnhancedFieldInfo
	}{
		{
			name: "select field analysis",
			field: FieldSchema{
				Name: "status",
				Type: "select",
				Options: &FieldOptions{
					Values: []string{"active", "inactive", "pending"},
				},
			},
			collectionName: "devices",
			want: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "status",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"active", "inactive", "pending"},
					},
				},
				EnumValues:   []string{"active", "inactive", "pending"},
				EnumTypeName: "DevicesStatusType",
			},
		},
		{
			name: "relation field analysis",
			field: FieldSchema{
				Name: "plant",
				Type: "relation",
				Options: &FieldOptions{
					CollectionID: "plants_collection_id",
					MaxSelect:    func() *int { i := 1; return &i }(),
				},
			},
			collectionName: "devices",
			want: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "plant",
					Type: "relation",
					Options: &FieldOptions{
						CollectionID: "plants_collection_id",
						MaxSelect:    func() *int { i := 1; return &i }(),
					},
				},
				TargetCollection: "plants",
				RelationTypeName: "PlantsRelation",
				IsMultiRelation:  false,
			},
		},
		{
			name: "multi relation field analysis",
			field: FieldSchema{
				Name: "users",
				Type: "relation",
				Options: &FieldOptions{
					CollectionID: "users_collection_id",
					MaxSelect:    func() *int { i := 3; return &i }(),
				},
			},
			collectionName: "projects",
			want: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "users",
					Type: "relation",
					Options: &FieldOptions{
						CollectionID: "users_collection_id",
						MaxSelect:    func() *int { i := 3; return &i }(),
					},
				},
				TargetCollection: "users",
				RelationTypeName: "UsersRelation",
				IsMultiRelation:  true,
			},
		},
		{
			name: "file field analysis",
			field: FieldSchema{
				Name: "avatar",
				Type: "file",
				Options: &FieldOptions{
					MaxSelect: func() *int { i := 1; return &i }(),
					Thumbs:    []string{"100x100", "200x200"},
				},
			},
			collectionName: "users",
			want: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "avatar",
					Type: "file",
					Options: &FieldOptions{
						MaxSelect: func() *int { i := 1; return &i }(),
						Thumbs:    []string{"100x100", "200x200"},
					},
				},
				FileTypeName:   "AvatarFile",
				IsMultiFile:    false,
				HasThumbnails:  true,
				ThumbnailSizes: []string{"100x100", "200x200"},
			},
		},
		{
			name: "multi file field analysis",
			field: FieldSchema{
				Name: "images",
				Type: "file",
				Options: &FieldOptions{
					MaxSelect: func() *int { i := 5; return &i }(),
				},
			},
			collectionName: "gallery",
			want: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "images",
					Type: "file",
					Options: &FieldOptions{
						MaxSelect: func() *int { i := 5; return &i }(),
					},
				},
				FileTypeName:   "ImagesFile",
				IsMultiFile:    true,
				HasThumbnails:  false,
				ThumbnailSizes: nil,
			},
		},
		{
			name: "text field analysis (no enhancement)",
			field: FieldSchema{
				Name: "title",
				Type: "text",
			},
			collectionName: "posts",
			want: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "title",
					Type: "text",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnalyzeEnhancedField(tt.field, tt.collectionName, allCollections)

			// Use reflect.DeepEqual to compare structs
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AnalyzeEnhancedField() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestToConstantName tests the ToConstantName function
func TestToConstantName(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		fieldName      string
		value          string
		want           string
	}{
		{
			name:           "simple values",
			collectionName: "devices",
			fieldName:      "status",
			value:          "active",
			want:           "DevicesStatusActive",
		},
		{
			name:           "value with spaces",
			collectionName: "users",
			fieldName:      "role",
			value:          "admin user",
			want:           "UsersRoleAdminUser",
		},
		{
			name:           "value with hyphens",
			collectionName: "products",
			fieldName:      "category",
			value:          "home-garden",
			want:           "ProductsCategoryHomeGarden",
		},
		{
			name:           "value with underscores",
			collectionName: "orders",
			fieldName:      "status",
			value:          "pending_payment",
			want:           "OrdersStatusPendingPayment",
		},
		{
			name:           "mixed special characters",
			collectionName: "items",
			fieldName:      "type",
			value:          "type-a_special item",
			want:           "ItemsTypeTypeASpecialItem",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToConstantName(tt.collectionName, tt.fieldName, tt.value)
			if got != tt.want {
				t.Errorf("ToConstantName() = %v, want %v", got, tt.want)
			}
		})
	}
}
