package generator

import (
	"testing"
)

// BenchmarkEnumGenerator_GenerateEnums enum 생성 전체 프로세스 성능을 측정합니다
func BenchmarkEnumGenerator_GenerateEnums(b *testing.B) {
	generator := NewEnumGenerator()

	// 테스트용 컬렉션 데이터
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

	// 테스트용 스키마 데이터 (다양한 select 필드 포함)
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

// BenchmarkEnumGenerator_GenerateEnumConstants 단일 enum 상수 생성 성능을 측정합니다
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

// BenchmarkEnumGenerator_LargeValueSet 많은 값을 가진 enum 생성 성능을 측정합니다
func BenchmarkEnumGenerator_LargeValueSet(b *testing.B) {
	generator := NewEnumGenerator()

	// 100개의 값을 가진 큰 enum
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

// BenchmarkEnumGenerator_MultipleCollections 여러 컬렉션의 enum 생성 성능을 측정합니다
func BenchmarkEnumGenerator_MultipleCollections(b *testing.B) {
	generator := NewEnumGenerator()

	// 10개 컬렉션, 각각 5개의 select 필드
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

// BenchmarkEnumGenerator_SpecialCharacters 특수문자가 포함된 값들의 enum 생성 성능을 측정합니다
func BenchmarkEnumGenerator_SpecialCharacters(b *testing.B) {
	generator := NewEnumGenerator()

	// 특수문자가 포함된 값들
	specialValues := []string{
		"value-with-dashes",
		"value with spaces",
		"value_with_underscores",
		"value.with.dots",
		"value@with@symbols",
		"value#with#hash",
		"value$with$dollar",
		"value%with%percent",
		"한글값",
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

// BenchmarkEnumGenerator_Memory enum 생성의 메모리 사용량을 측정합니다
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
	b.ReportAllocs() // 메모리 할당 정보 리포트

	for i := 0; i < b.N; i++ {
		_ = generator.GenerateEnums(collections, schemas)
	}
}
