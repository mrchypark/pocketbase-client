package main

import (
	"os"
	"testing"
	"text/template"

	"github.com/mrchypark/pocketbase-client/internal/generator"
)

// BenchmarkCodeGeneration 전체 코드 생성 프로세스의 성능을 측정합니다
func BenchmarkCodeGeneration(b *testing.B) {
	// 테스트용 스키마 파일 경로
	schemaPath := "../../internal/generator/testdata/complex_schema.json"

	// 스키마 로드 (벤치마크 외부에서 한 번만 수행)
	schemas, err := generator.LoadSchema(schemaPath)
	if err != nil {
		b.Fatalf("스키마 로드 실패: %v", err)
	}

	// 템플릿 파싱 (벤치마크 외부에서 한 번만 수행)
	tpl, err := template.New("models").Parse(templateFile)
	if err != nil {
		b.Fatalf("템플릿 파싱 실패: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// 기본 TemplateData 생성
		baseTplData := generator.TemplateData{
			PackageName: "models",
			JSONLibrary: "encoding/json",
			Collections: make([]generator.CollectionData, 0, len(schemas)),
		}

		for _, s := range schemas {
			collectionData := generator.CollectionData{
				CollectionName: s.Name,
				StructName:     generator.ToPascalCase(s.Name),
				Fields:         make([]generator.FieldData, 0, len(s.Fields)),
			}

			for _, field := range s.Fields {
				if field.System {
					continue
				}
				goType, getter := generator.MapPbTypeToGoType(field, !field.Required)
				collectionData.Fields = append(collectionData.Fields, generator.FieldData{
					JSONName:     field.Name,
					GoName:       generator.ToPascalCase(field.Name),
					GoType:       goType,
					OmitEmpty:    !field.Required,
					GetterMethod: getter,
				})
			}
			baseTplData.Collections = append(baseTplData.Collections, collectionData)
		}

		// Enhanced 기능 포함 데이터 생성
		enhancedData := generator.EnhancedTemplateData{
			TemplateData:      baseTplData,
			GenerateEnums:     true,
			GenerateRelations: true,
			GenerateFiles:     true,
		}

		// Enhanced 분석 및 데이터 생성
		enumGenerator := generator.NewEnumGenerator()
		enhancedData.Enums = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)

		relationGenerator := generator.NewRelationGenerator()
		enhancedData.RelationTypes = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)

		fileGenerator := generator.NewFileGenerator()
		enhancedData.FileTypes = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)

		// 임시 파일에 템플릿 실행
		tmpFile, err := os.CreateTemp("", "benchmark_*.go")
		if err != nil {
			b.Fatalf("임시 파일 생성 실패: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		err = tpl.Execute(tmpFile, enhancedData)
		if err != nil {
			b.Fatalf("템플릿 실행 실패: %v", err)
		}
	}
}

// BenchmarkSchemaLoading 스키마 로딩 성능을 측정합니다
func BenchmarkSchemaLoading(b *testing.B) {
	schemaPath := "../../internal/generator/testdata/complex_schema.json"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := generator.LoadSchema(schemaPath)
		if err != nil {
			b.Fatalf("스키마 로드 실패: %v", err)
		}
	}
}

// BenchmarkTemplateExecution 템플릿 실행 성능을 측정합니다
func BenchmarkTemplateExecution(b *testing.B) {
	// 테스트 데이터 준비
	testData := generator.EnhancedTemplateData{
		TemplateData: generator.TemplateData{
			PackageName: "models",
			JSONLibrary: "encoding/json",
			Collections: []generator.CollectionData{
				{
					CollectionName: "test_collection",
					StructName:     "TestCollection",
					Fields: []generator.FieldData{
						{
							JSONName:     "name",
							GoName:       "Name",
							GoType:       "string",
							OmitEmpty:    false,
							GetterMethod: "GetString",
						},
						{
							JSONName:     "count",
							GoName:       "Count",
							GoType:       "int",
							OmitEmpty:    false,
							GetterMethod: "GetInt",
						},
					},
				},
			},
		},
		GenerateEnums:     true,
		GenerateRelations: true,
		GenerateFiles:     true,
		Enums: []generator.EnumData{
			{
				CollectionName: "test_collection",
				FieldName:      "status",
				EnumTypeName:   "TestCollectionStatusType",
				Constants: []generator.ConstantData{
					{Name: "TestCollectionStatusActive", Value: "active"},
					{Name: "TestCollectionStatusInactive", Value: "inactive"},
				},
			},
		},
	}

	// 템플릿 파싱
	tpl, err := template.New("models").Parse(templateFile)
	if err != nil {
		b.Fatalf("템플릿 파싱 실패: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		tmpFile, err := os.CreateTemp("", "benchmark_template_*.go")
		if err != nil {
			b.Fatalf("임시 파일 생성 실패: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		defer tmpFile.Close()

		err = tpl.Execute(tmpFile, testData)
		if err != nil {
			b.Fatalf("템플릿 실행 실패: %v", err)
		}
	}
}

// BenchmarkEnumGeneration enum 생성 성능을 측정합니다
func BenchmarkEnumGeneration(b *testing.B) {
	// 테스트용 스키마 데이터
	schemas := []generator.CollectionSchema{
		{
			Name: "devices",
			Fields: []generator.FieldSchema{
				{
					Name: "type",
					Type: "select",
					Options: &generator.FieldOptions{
						Values: []string{"m2", "d2", "s2", "pro", "mini"},
					},
				},
				{
					Name: "status",
					Type: "select",
					Options: &generator.FieldOptions{
						Values: []string{"active", "inactive", "maintenance", "retired"},
					},
				},
			},
		},
	}

	collections := []generator.CollectionData{
		{
			CollectionName: "devices",
			StructName:     "Device",
		},
	}

	enumGenerator := generator.NewEnumGenerator()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = enumGenerator.GenerateEnums(collections, schemas)
	}
}

// BenchmarkRelationGeneration relation 타입 생성 성능을 측정합니다
func BenchmarkRelationGeneration(b *testing.B) {
	// 테스트용 스키마 데이터
	schemas := []generator.CollectionSchema{
		{
			ID:   "plants_id",
			Name: "plants",
		},
		{
			Name: "devices",
			Fields: []generator.FieldSchema{
				{
					Name: "plant",
					Type: "relation",
					Options: &generator.FieldOptions{
						CollectionID: "plants_id",
						MaxSelect:    intPtr(1),
					},
				},
			},
		},
	}

	collections := []generator.CollectionData{
		{
			CollectionName: "devices",
			StructName:     "Device",
		},
	}

	relationGenerator := generator.NewRelationGenerator()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = relationGenerator.GenerateRelationTypes(collections, schemas)
	}
}

// BenchmarkFileGeneration file 타입 생성 성능을 측정합니다
func BenchmarkFileGeneration(b *testing.B) {
	// 테스트용 스키마 데이터
	schemas := []generator.CollectionSchema{
		{
			Name: "posts",
			Fields: []generator.FieldSchema{
				{
					Name: "image",
					Type: "file",
					Options: &generator.FieldOptions{
						MaxSelect: intPtr(1),
						Thumbs:    []string{"400x0", "100x0"},
					},
				},
				{
					Name: "attachments",
					Type: "file",
					Options: &generator.FieldOptions{
						MaxSelect: intPtr(5),
					},
				},
			},
		},
	}

	collections := []generator.CollectionData{
		{
			CollectionName: "posts",
			StructName:     "Post",
		},
	}

	fileGenerator := generator.NewFileGenerator()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = fileGenerator.GenerateFileTypes(collections, schemas)
	}
}

// BenchmarkMemoryUsage 메모리 사용량을 측정합니다
func BenchmarkMemoryUsage(b *testing.B) {
	schemaPath := "../../internal/generator/testdata/complex_schema.json"

	schemas, err := generator.LoadSchema(schemaPath)
	if err != nil {
		b.Fatalf("스키마 로드 실패: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs() // 메모리 할당 정보 리포트

	for i := 0; i < b.N; i++ {
		// 전체 코드 생성 프로세스
		baseTplData := generator.TemplateData{
			PackageName: "models",
			JSONLibrary: "encoding/json",
			Collections: make([]generator.CollectionData, 0, len(schemas)),
		}

		for _, s := range schemas {
			collectionData := generator.CollectionData{
				CollectionName: s.Name,
				StructName:     generator.ToPascalCase(s.Name),
				Fields:         make([]generator.FieldData, 0, len(s.Fields)),
			}

			for _, field := range s.Fields {
				if field.System {
					continue
				}
				goType, getter := generator.MapPbTypeToGoType(field, !field.Required)
				collectionData.Fields = append(collectionData.Fields, generator.FieldData{
					JSONName:     field.Name,
					GoName:       generator.ToPascalCase(field.Name),
					GoType:       goType,
					OmitEmpty:    !field.Required,
					GetterMethod: getter,
				})
			}
			baseTplData.Collections = append(baseTplData.Collections, collectionData)
		}

		// Enhanced 기능들
		enumGenerator := generator.NewEnumGenerator()
		relationGenerator := generator.NewRelationGenerator()
		fileGenerator := generator.NewFileGenerator()

		_ = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
		_ = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
		_ = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
	}
}

// 헬퍼 함수
func intPtr(i int) *int {
	return &i
}
