package generator

import (
	"testing"
)

// BenchmarkEnumGenerator_GenerateEnums measures performance of the entire enum generation process
func BenchmarkEnumGenerator_GenerateEnums(b *testing.B) {
	generator := NewEnumGenerator()

	// Test collection data
	collections := []CollectionData{
		{
			CollectionName: "devices",
			StructName:     "Device",
		},
		{
			CollectionName: "users",
			StructName:     "User",
		},
		{
			CollectionName: "posts",
			StructName:     "Post",
		},
	}

	// Test schema data (including various select fields)
	schemas := []CollectionSchema{
		{
			Name: "devices",
			Fields: []FieldSchema{
				{
					Name: "type",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"m2", "d2", "s2", "pro", "mini"},
					},
				},
				{
					Name: "status",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"active", "inactive", "maintenance", "retired"},
					},
				},
				{
					Name: "priority",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"low", "medium", "high", "critical"},
					},
				},
			},
		},
		{
			Name: "users",
			Fields: []FieldSchema{
				{
					Name: "role",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"admin", "moderator", "user", "guest"},
					},
				},
				{
					Name: "subscription",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"free", "basic", "premium", "enterprise"},
					},
				},
			},
		},
		{
			Name: "posts",
			Fields: []FieldSchema{
				{
					Name: "visibility",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"public", "private", "draft", "archived"},
					},
				},
				{
					Name: "category",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"tech", "lifestyle", "business", "education", "entertainment"},
					},
				},
			},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateEnums(collections, schemas)
	}
}

// BenchmarkEnumGenerator_GenerateEnumConstants measures performance of single enum constant generation
func BenchmarkEnumGenerator_GenerateEnumConstants(b *testing.B) {
	generator := NewEnumGenerator()

	enhanced := EnhancedFieldInfo{
		FieldSchema: FieldSchema{
			Name: "status",
			Type: "select",
			Options: &FieldOptions{
				Values: []string{"active", "inactive", "pending", "archived", "deleted", "suspended"},
			},
		},
		EnumValues:   []string{"active", "inactive", "pending", "archived", "deleted", "suspended"},
		EnumTypeName: "DeviceStatusType",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateEnumConstants(enhanced, "devices")
	}
}

// BenchmarkEnumGenerator_LargeValueSet measures performance of enum generation with many values
func BenchmarkEnumGenerator_LargeValueSet(b *testing.B) {
	generator := NewEnumGenerator()

	// Large enum with 100 values
	largeValues := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeValues[i] = "value_" + string(rune('a'+i%26)) + string(rune('0'+i/26))
	}

	enhanced := EnhancedFieldInfo{
		FieldSchema: FieldSchema{
			Name: "large_enum",
			Type: "select",
			Options: &FieldOptions{
				Values: largeValues,
			},
		},
		EnumValues:   largeValues,
		EnumTypeName: "LargeEnumType",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateEnumConstants(enhanced, "test_collection")
	}
}

// BenchmarkEnumGenerator_MultipleCollections measures performance of enum generation for multiple collections
func BenchmarkEnumGenerator_MultipleCollections(b *testing.B) {
	generator := NewEnumGenerator()

	// 10 collections, each with 5 select fields
	collections := make([]CollectionData, 10)
	schemas := make([]CollectionSchema, 10)

	for i := 0; i < 10; i++ {
		collectionName := "collection_" + string(rune('a'+i))
		collections[i] = CollectionData{
			CollectionName: collectionName,
			StructName:     ToPascalCase(collectionName),
		}

		fields := make([]FieldSchema, 5)
		for j := 0; j < 5; j++ {
			fieldName := "field_" + string(rune('a'+j))
			fields[j] = FieldSchema{
				Name: fieldName,
				Type: "select",
				Options: &FieldOptions{
					Values: []string{"value1", "value2", "value3", "value4", "value5"},
				},
			}
		}

		schemas[i] = CollectionSchema{
			Name:   collectionName,
			Fields: fields,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateEnums(collections, schemas)
	}
}

// BenchmarkEnumGenerator_SpecialCharacters measures performance of enum generation with special characters in values
func BenchmarkEnumGenerator_SpecialCharacters(b *testing.B) {
	generator := NewEnumGenerator()

	// Values containing special characters
	specialValues := []string{
		"value-with-dashes",
		"value with spaces",
		"value_with_underscores",
		"value.with.dots",
		"value@with@symbols",
		"value#with#hash",
		"value$with$dollar",
		"value%with%percent",
		"korean_value",
		"日本語値",
	}

	enhanced := EnhancedFieldInfo{
		FieldSchema: FieldSchema{
			Name: "special_enum",
			Type: "select",
			Options: &FieldOptions{
				Values: specialValues,
			},
		},
		EnumValues:   specialValues,
		EnumTypeName: "SpecialEnumType",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateEnumConstants(enhanced, "test_collection")
	}
}

// BenchmarkEnumGenerator_Memory measures memory usage of enum generation
func BenchmarkEnumGenerator_Memory(b *testing.B) {
	generator := NewEnumGenerator()

	collections := []CollectionData{
		{
			CollectionName: "devices",
			StructName:     "Device",
		},
	}

	schemas := []CollectionSchema{
		{
			Name: "devices",
			Fields: []FieldSchema{
				{
					Name: "type",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"m2", "d2", "s2", "pro", "mini", "ultra", "max", "lite"},
					},
				},
				{
					Name: "status",
					Type: "select",
					Options: &FieldOptions{
						Values: []string{"active", "inactive", "maintenance", "retired", "pending", "error"},
					},
				},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs() // Report memory allocation information

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateEnums(collections, schemas)
	}
}
