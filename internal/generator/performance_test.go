package generator

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"text/template"
	"time"
)

// BenchmarkLargeSchemaProcessing measures performance of large schema processing
func BenchmarkLargeSchemaProcessing(b *testing.B) {
	schemaPath := "testdata/large_schema.json"

	// Load schema (performed only once outside benchmark)
	schemas, err := LoadSchema(schemaPath)
	if err != nil {
		b.Fatalf("Large schema load failed: %v", err)
	}

	b.Logf("Number of loaded collections: %d", len(schemas))

	// Calculate total number of fields
	totalFields := 0
	for _, schema := range schemas {
		totalFields += len(schema.Fields)
	}
	b.Logf("Total number of fields: %d", totalFields)

	b.ResetTimer()
	b.ReportAllocs() // Report memory allocation information

	for i := 0; i < b.N; i++ {
		// Generate basic TemplateData
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

		// Generate Enhanced features
		enumGenerator := NewEnumGenerator()
		relationGenerator := NewRelationGenerator()
		fileGenerator := NewFileGenerator()

		_ = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
		_ = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
		_ = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
	}
}

// BenchmarkScalabilityTest scalability test - measures performance changes based on number of collections
func BenchmarkScalabilityTest(b *testing.B) {
	collectionCounts := []int{1, 5, 10, 20, 50, 100}

	for _, count := range collectionCounts {
		b.Run(fmt.Sprintf("Collections_%d", count), func(b *testing.B) {
			// Generate schema dynamically
			schemas := generateTestSchemas(count)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				processSchemas(schemas)
			}
		})
	}
}

// BenchmarkFieldScalabilityTest measures performance changes based on number of fields
func BenchmarkFieldScalabilityTest(b *testing.B) {
	fieldCounts := []int{5, 10, 20, 50, 100, 200}

	for _, count := range fieldCounts {
		b.Run(fmt.Sprintf("Fields_%d", count), func(b *testing.B) {
			// Create single collection with many fields
			schemas := generateTestSchemasWithFields(1, count)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				processSchemas(schemas)
			}
		})
	}
}

// BenchmarkComplexRelationships measures performance of complex relationship structures
func BenchmarkComplexRelationships(b *testing.B) {
	// Generate complex schema with many cross-references
	schemas := generateComplexRelationSchemas()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processSchemas(schemas)
	}
}

// BenchmarkMemoryUsageBySize measures memory usage by schema size
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

// TestLargeSchemaGeneration tests actual code generation with large schema
func TestLargeSchemaGeneration(t *testing.T) {
	schemaPath := "testdata/large_schema.json"

	schemas, err := LoadSchema(schemaPath)
	if err != nil {
		t.Fatalf("Large schema load failed: %v", err)
	}

	start := time.Now()

	// Generate basic TemplateData
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

	// Generate Enhanced features
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
	t.Logf("Data processing time: %v", processingTime)

	// Template execution test
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
		t.Fatalf("Template parsing failed: %v", err)
	}

	start = time.Now()

	tmpFile, err := os.CreateTemp("", "large_schema_test_*.go")
	if err != nil {
		t.Fatalf("Temporary file creation failed: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	err = tpl.Execute(tmpFile, enhancedData)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	templateTime := time.Since(start)
	t.Logf("Template execution time: %v", templateTime)

	// Check generated file size
	fileInfo, err := tmpFile.Stat()
	if err != nil {
		t.Fatalf("File info query failed: %v", err)
	}

	t.Logf("Generated file size: %d bytes", fileInfo.Size())
	t.Logf("Total processing time: %v", processingTime+templateTime)

	// Performance threshold validation
	totalTime := processingTime + templateTime
	if totalTime > 5*time.Second {
		t.Errorf("Processing time is too long: %v (threshold: 5 seconds)", totalTime)
	}

	if fileInfo.Size() > 10*1024*1024 { // 10MB
		t.Errorf("Generated file is too large: %d bytes (threshold: 10MB)", fileInfo.Size())
	}
}

// TestPerformanceBottlenecks identifies performance bottlenecks
func TestPerformanceBottlenecks(t *testing.T) {
	schemas := generateTestSchemasWithFields(50, 50) // 50 collections, 50 fields each

	// Measure time for each step
	times := make(map[string]time.Duration)

	// 1. Generate basic data
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

	// 2. Generate Enum
	start = time.Now()
	enumGenerator := NewEnumGenerator()
	enums := enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
	times["enum_generation"] = time.Since(start)

	// 3. Generate Relation
	start = time.Now()
	relationGenerator := NewRelationGenerator()
	relations := relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
	times["relation_generation"] = time.Since(start)

	// 4. Generate File
	start = time.Now()
	fileGenerator := NewFileGenerator()
	files := fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
	times["file_generation"] = time.Since(start)

	// Output and analyze results
	t.Logf("Performance analysis results:")
	for phase, duration := range times {
		t.Logf("  %s: %v", phase, duration)
	}

	t.Logf("Generated data statistics:")
	t.Logf("  Collections: %d", len(baseTplData.Collections))
	t.Logf("  Enum: %d", len(enums))
	t.Logf("  Relation: %d", len(relations))
	t.Logf("  File: %d", len(files))

	// Identify bottlenecks
	maxTime := time.Duration(0)
	bottleneck := ""
	for phase, duration := range times {
		if duration > maxTime {
			maxTime = duration
			bottleneck = phase
		}
	}

	t.Logf("Longest step: %s (%v)", bottleneck, maxTime)
}

// Helper functions

// generateTestSchemas dynamically generates test schemas
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

// generateTestSchemasWithFields generates schemas with specified number of fields
func generateTestSchemasWithFields(collectionCount, fieldCount int) []CollectionSchema {
	schemas := make([]CollectionSchema, collectionCount)

	for i := 0; i < collectionCount; i++ {
		fields := make([]FieldSchema, fieldCount)

		for j := 0; j < fieldCount; j++ {
			fieldType := "text"
			var options *FieldOptions

			// Distribute field types diversely
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
				Required: j%3 == 0, // 1/3 probability of being required
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

// generateComplexRelationSchemas generates schemas with complex relationship structures
func generateComplexRelationSchemas() []CollectionSchema {
	schemas := make([]CollectionSchema, 10)

	for i := 0; i < 10; i++ {
		fields := make([]FieldSchema, 0)

		// Basic fields
		fields = append(fields, FieldSchema{
			Name:     "name",
			Type:     "text",
			Required: true,
		})

		// Add relationship fields with other collections
		for j := 0; j < 10; j++ {
			if i != j { // exclude self
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

// processSchemas executes schema processing logic
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

	// Enhanced features
	enumGenerator := NewEnumGenerator()
	relationGenerator := NewRelationGenerator()
	fileGenerator := NewFileGenerator()

	_ = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
	_ = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
	_ = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
}
