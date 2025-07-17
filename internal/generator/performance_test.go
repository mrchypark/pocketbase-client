package generator

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"text/template"
	"time"
)

// BenchmarkLargeSchemaProcessing 대용량 스키마 처리 성능을 측정합니다
func BenchmarkLargeSchemaProcessing(b *testing.B) {
	schemaPath := "testdata/large_schema.json"

	// 스키마 로드 (벤치마크 외부에서 한 번만 수행)
	schemas, err := LoadSchema(schemaPath)
	if err != nil {
		b.Fatalf("대용량 스키마 로드 실패: %v", err)
	}

	b.Logf("로드된 컬렉션 수: %d", len(schemas))

	// 총 필드 수 계산
	totalFields := 0
	for _, schema := range schemas {
		totalFields += len(schema.Fields)
	}
	b.Logf("총 필드 수: %d", totalFields)

	b.ResetTimer()
	b.ReportAllocs() // 메모리 할당 정보 리포트

	for i := 0; i < b.N; i++ {
		// 기본 TemplateData 생성
		baseTplData := TemplateData{
			PackageName: "models",
			JSONLibrary: "encoding/json",
			Collections: make([]CollectionData, 0, len(schemas)),
		}

		for _, s := range schemas {
			collectionData := CollectionData{
				CollectionName: s.Name,
				StructName:     ToPascalCase(s.Name),
				Fields:         make([]FieldData, 0, len(s.Fields)),
			}

			for _, field := range s.Fields {
				if field.System {
					continue
				}
				goType, _, getter := MapPbTypeToGoType(field, !field.Required)
				collectionData.Fields = append(collectionData.Fields, FieldData{
					JSONName:     field.Name,
					GoName:       ToPascalCase(field.Name),
					GoType:       goType,
					OmitEmpty:    !field.Required,
					GetterMethod: getter,
				})
			}
			baseTplData.Collections = append(baseTplData.Collections, collectionData)
		}

		// Enhanced 기능들 생성
		enumGenerator := NewEnumGenerator()
		relationGenerator := NewRelationGenerator()
		fileGenerator := NewFileGenerator()

		_ = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
		_ = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
		_ = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
	}
}

// BenchmarkScalabilityTest 확장성 테스트 - 컬렉션 수에 따른 성능 변화를 측정합니다
func BenchmarkScalabilityTest(b *testing.B) {
	collectionCounts := []int{1, 5, 10, 20, 50, 100}

	for _, count := range collectionCounts {
		b.Run(fmt.Sprintf("Collections_%d", count), func(b *testing.B) {
			// 동적으로 스키마 생성
			schemas := generateTestSchemas(count)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				processSchemas(schemas)
			}
		})
	}
}

// BenchmarkFieldScalabilityTest 필드 수에 따른 성능 변화를 측정합니다
func BenchmarkFieldScalabilityTest(b *testing.B) {
	fieldCounts := []int{5, 10, 20, 50, 100, 200}

	for _, count := range fieldCounts {
		b.Run(fmt.Sprintf("Fields_%d", count), func(b *testing.B) {
			// 많은 필드를 가진 단일 컬렉션 생성
			schemas := generateTestSchemasWithFields(1, count)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				processSchemas(schemas)
			}
		})
	}
}

// BenchmarkComplexRelationships 복잡한 관계 구조의 성능을 측정합니다
func BenchmarkComplexRelationships(b *testing.B) {
	// 상호 참조가 많은 복잡한 스키마 생성
	schemas := generateComplexRelationSchemas()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processSchemas(schemas)
	}
}

// BenchmarkMemoryUsageBySize 스키마 크기별 메모리 사용량을 측정합니다
func BenchmarkMemoryUsageBySize(b *testing.B) {
	sizes := []struct {
		collections int
		fields      int
	}{
		{10, 10},
		{20, 20},
		{50, 30},
		{100, 50},
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("C%d_F%d", size.collections, size.fields), func(b *testing.B) {
			schemas := generateTestSchemasWithFields(size.collections, size.fields)

			var m1, m2 runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&m1)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				processSchemas(schemas)
			}

			runtime.GC()
			runtime.ReadMemStats(&m2)

			b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
		})
	}
}

// TestLargeSchemaGeneration 대용량 스키마로 실제 코드 생성 테스트
func TestLargeSchemaGeneration(t *testing.T) {
	schemaPath := "testdata/large_schema.json"

	schemas, err := LoadSchema(schemaPath)
	if err != nil {
		t.Fatalf("대용량 스키마 로드 실패: %v", err)
	}

	start := time.Now()

	// 기본 TemplateData 생성
	baseTplData := TemplateData{
		PackageName: "models",
		JSONLibrary: "encoding/json",
		Collections: make([]CollectionData, 0, len(schemas)),
	}

	for _, s := range schemas {
		collectionData := CollectionData{
			CollectionName: s.Name,
			StructName:     ToPascalCase(s.Name),
			Fields:         make([]FieldData, 0, len(s.Fields)),
		}

		for _, field := range s.Fields {
			if field.System {
				continue
			}
			goType, _, getter := MapPbTypeToGoType(field, !field.Required)
			collectionData.Fields = append(collectionData.Fields, FieldData{
				JSONName:     field.Name,
				GoName:       ToPascalCase(field.Name),
				GoType:       goType,
				OmitEmpty:    !field.Required,
				GetterMethod: getter,
			})
		}
		baseTplData.Collections = append(baseTplData.Collections, collectionData)
	}

	// Enhanced 기능들 생성
	enhancedData := EnhancedTemplateData{
		TemplateData:      baseTplData,
		GenerateEnums:     true,
		GenerateRelations: true,
		GenerateFiles:     true,
	}

	enumGenerator := NewEnumGenerator()
	enhancedData.Enums = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)

	relationGenerator := NewRelationGenerator()
	enhancedData.RelationTypes = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)

	fileGenerator := NewFileGenerator()
	enhancedData.FileTypes = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)

	processingTime := time.Since(start)
	t.Logf("데이터 처리 시간: %v", processingTime)

	// 템플릿 실행 테스트
	templateContent := `package {{.PackageName}}

// Generated code - do not edit

{{range .Collections}}
type {{.StructName}} struct {
{{range .Fields}}    {{.GoName}} {{.GoType}} ` + "`json:\"{{.JSONName}}{{if .OmitEmpty}},omitempty{{end}}\"`" + `
{{end}}}
{{end}}

{{if .GenerateEnums}}
{{range .Enums}}
// {{.EnumTypeName}} enum constants
const (
{{range .Constants}}    {{.Name}} = "{{.Value}}"
{{end}})
{{end}}
{{end}}

{{if .GenerateRelations}}
{{range .RelationTypes}}
// {{.TypeName}} represents a relation to {{.TargetCollection}}
type {{.TypeName}} struct {
    id string
}
{{end}}
{{end}}

{{if .GenerateFiles}}
{{range .FileTypes}}
// {{.TypeName}} represents a file reference
type {{.TypeName}} struct {
    filename string
}
{{end}}
{{end}}
`

	tpl, err := template.New("test").Parse(templateContent)
	if err != nil {
		t.Fatalf("템플릿 파싱 실패: %v", err)
	}

	start = time.Now()

	tmpFile, err := os.CreateTemp("", "large_schema_test_*.go")
	if err != nil {
		t.Fatalf("임시 파일 생성 실패: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	err = tpl.Execute(tmpFile, enhancedData)
	if err != nil {
		t.Fatalf("템플릿 실행 실패: %v", err)
	}

	templateTime := time.Since(start)
	t.Logf("템플릿 실행 시간: %v", templateTime)

	// 생성된 파일 크기 확인
	fileInfo, err := tmpFile.Stat()
	if err != nil {
		t.Fatalf("파일 정보 조회 실패: %v", err)
	}

	t.Logf("생성된 파일 크기: %d bytes", fileInfo.Size())
	t.Logf("총 처리 시간: %v", processingTime+templateTime)

	// 성능 임계값 검증
	totalTime := processingTime + templateTime
	if totalTime > 5*time.Second {
		t.Errorf("처리 시간이 너무 오래 걸립니다: %v (임계값: 5초)", totalTime)
	}

	if fileInfo.Size() > 10*1024*1024 { // 10MB
		t.Errorf("생성된 파일이 너무 큽니다: %d bytes (임계값: 10MB)", fileInfo.Size())
	}
}

// TestPerformanceBottlenecks 성능 병목 지점을 식별합니다
func TestPerformanceBottlenecks(t *testing.T) {
	schemas := generateTestSchemasWithFields(50, 50) // 50개 컬렉션, 각각 50개 필드

	// 각 단계별 시간 측정
	times := make(map[string]time.Duration)

	// 1. 기본 데이터 생성
	start := time.Now()
	baseTplData := TemplateData{
		PackageName: "models",
		JSONLibrary: "encoding/json",
		Collections: make([]CollectionData, 0, len(schemas)),
	}

	for _, s := range schemas {
		collectionData := CollectionData{
			CollectionName: s.Name,
			StructName:     ToPascalCase(s.Name),
			Fields:         make([]FieldData, 0, len(s.Fields)),
		}

		for _, field := range s.Fields {
			if field.System {
				continue
			}
			goType, _, getter := MapPbTypeToGoType(field, !field.Required)
			collectionData.Fields = append(collectionData.Fields, FieldData{
				JSONName:     field.Name,
				GoName:       ToPascalCase(field.Name),
				GoType:       goType,
				OmitEmpty:    !field.Required,
				GetterMethod: getter,
			})
		}
		baseTplData.Collections = append(baseTplData.Collections, collectionData)
	}
	times["basic_data_generation"] = time.Since(start)

	// 2. Enum 생성
	start = time.Now()
	enumGenerator := NewEnumGenerator()
	enums := enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
	times["enum_generation"] = time.Since(start)

	// 3. Relation 생성
	start = time.Now()
	relationGenerator := NewRelationGenerator()
	relations := relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
	times["relation_generation"] = time.Since(start)

	// 4. File 생성
	start = time.Now()
	fileGenerator := NewFileGenerator()
	files := fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
	times["file_generation"] = time.Since(start)

	// 결과 출력 및 분석
	t.Logf("성능 분석 결과:")
	for phase, duration := range times {
		t.Logf("  %s: %v", phase, duration)
	}

	t.Logf("생성된 데이터 통계:")
	t.Logf("  컬렉션: %d개", len(baseTplData.Collections))
	t.Logf("  Enum: %d개", len(enums))
	t.Logf("  Relation: %d개", len(relations))
	t.Logf("  File: %d개", len(files))

	// 병목 지점 식별
	maxTime := time.Duration(0)
	bottleneck := ""
	for phase, duration := range times {
		if duration > maxTime {
			maxTime = duration
			bottleneck = phase
		}
	}

	t.Logf("가장 오래 걸린 단계: %s (%v)", bottleneck, maxTime)
}

// 헬퍼 함수들

// generateTestSchemas 테스트용 스키마를 동적으로 생성합니다
func generateTestSchemas(collectionCount int) []CollectionSchema {
	schemas := make([]CollectionSchema, collectionCount)

	for i := 0; i < collectionCount; i++ {
		schemas[i] = CollectionSchema{
			ID:   fmt.Sprintf("collection_%03d", i+1),
			Name: fmt.Sprintf("collection_%d", i+1),
			Type: "base",
			Fields: []FieldSchema{
				{
					Name:     "name",
					Type:     "text",
					Required: true,
				},
				{
					Name: "status",
					Type: "select",
					Options: &FieldOptions{
						MaxSelect: intPtr(1),
						Values:    []string{"active", "inactive", "pending"},
					},
				},
				{
					Name: "tags",
					Type: "select",
					Options: &FieldOptions{
						MaxSelect: intPtr(5),
						Values:    []string{"tag1", "tag2", "tag3", "tag4", "tag5"},
					},
				},
				{
					Name: "image",
					Type: "file",
					Options: &FieldOptions{
						MaxSelect: intPtr(1),
						Thumbs:    []string{"400x0", "200x0"},
					},
				},
			},
		}
	}

	return schemas
}

// generateTestSchemasWithFields 지정된 수의 필드를 가진 스키마를 생성합니다
func generateTestSchemasWithFields(collectionCount, fieldCount int) []CollectionSchema {
	schemas := make([]CollectionSchema, collectionCount)

	for i := 0; i < collectionCount; i++ {
		fields := make([]FieldSchema, fieldCount)

		for j := 0; j < fieldCount; j++ {
			fieldType := "text"
			var options *FieldOptions

			// 필드 타입을 다양하게 분배
			switch j % 6 {
			case 0:
				fieldType = "text"
			case 1:
				fieldType = "number"
			case 2:
				fieldType = "select"
				options = &FieldOptions{
					MaxSelect: intPtr(1),
					Values:    []string{"option1", "option2", "option3"},
				}
			case 3:
				fieldType = "relation"
				options = &FieldOptions{
					CollectionID: fmt.Sprintf("collection_%03d", (i+1)%collectionCount+1),
					MaxSelect:    intPtr(1),
				}
			case 4:
				fieldType = "file"
				options = &FieldOptions{
					MaxSelect: intPtr(1),
					Thumbs:    []string{"400x0"},
				}
			case 5:
				fieldType = "json"
			}

			fields[j] = FieldSchema{
				Name:     fmt.Sprintf("field_%d", j+1),
				Type:     fieldType,
				Required: j%3 == 0, // 1/3 확률로 required
				Options:  options,
			}
		}

		schemas[i] = CollectionSchema{
			ID:     fmt.Sprintf("collection_%03d", i+1),
			Name:   fmt.Sprintf("collection_%d", i+1),
			Type:   "base",
			Fields: fields,
		}
	}

	return schemas
}

// generateComplexRelationSchemas 복잡한 관계 구조를 가진 스키마를 생성합니다
func generateComplexRelationSchemas() []CollectionSchema {
	schemas := make([]CollectionSchema, 10)

	for i := 0; i < 10; i++ {
		fields := make([]FieldSchema, 0)

		// 기본 필드들
		fields = append(fields, FieldSchema{
			Name:     "name",
			Type:     "text",
			Required: true,
		})

		// 다른 컬렉션들과의 관계 필드들 추가
		for j := 0; j < 10; j++ {
			if i != j { // 자기 자신 제외
				fields = append(fields, FieldSchema{
					Name: fmt.Sprintf("relation_to_%d", j+1),
					Type: "relation",
					Options: &FieldOptions{
						CollectionID: fmt.Sprintf("collection_%03d", j+1),
						MaxSelect:    intPtr(5),
					},
				})
			}
		}

		schemas[i] = CollectionSchema{
			ID:     fmt.Sprintf("collection_%03d", i+1),
			Name:   fmt.Sprintf("collection_%d", i+1),
			Type:   "base",
			Fields: fields,
		}
	}

	return schemas
}

// processSchemas 스키마 처리 로직을 실행합니다
func processSchemas(schemas []CollectionSchema) {
	baseTplData := TemplateData{
		PackageName: "models",
		JSONLibrary: "encoding/json",
		Collections: make([]CollectionData, 0, len(schemas)),
	}

	for _, s := range schemas {
		collectionData := CollectionData{
			CollectionName: s.Name,
			StructName:     ToPascalCase(s.Name),
			Fields:         make([]FieldData, 0, len(s.Fields)),
		}

		for _, field := range s.Fields {
			if field.System {
				continue
			}
			goType, _, getter := MapPbTypeToGoType(field, !field.Required)
			collectionData.Fields = append(collectionData.Fields, FieldData{
				JSONName:     field.Name,
				GoName:       ToPascalCase(field.Name),
				GoType:       goType,
				OmitEmpty:    !field.Required,
				GetterMethod: getter,
			})
		}
		baseTplData.Collections = append(baseTplData.Collections, collectionData)
	}

	// Enhanced 기능들
	enumGenerator := NewEnumGenerator()
	relationGenerator := NewRelationGenerator()
	fileGenerator := NewFileGenerator()

	_ = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
	_ = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
	_ = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
}
