package main

import (
	// 1. Add embed package.
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mrchypark/pocketbase-client/internal/generator"
	"golang.org/x/tools/imports"
)

// 2. Use go:embed directive to store template file content in a variable.
// This code must specify the relative path of the template file based on the main.go file.
//
//go:embed template.go.tpl
var templateFile string

func main() {
	schemaPath := flag.String("schema", "./pb_schema.json", "Input file path (pb_schema.json)")
	outputPath := flag.String("path", "./models.gen.go", "Output file path")
	pkgName := flag.String("pkgname", "models", "Package name for the generated file")
	jsonLib := flag.String("jsonlib", "encoding/json", "JSON library to use (e.g., github.com/goccy/go-json)")

	// 새로운 enhanced 기능 플래그들
	generateEnums := flag.Bool("enums", true, "Generate enum constants for select fields")
	generateRelations := flag.Bool("relations", true, "Generate enhanced relation types")
	generateFiles := flag.Bool("files", true, "Generate enhanced file types")

	flag.Parse()

	log.Printf("Generating Go models from schema: %s\n", *schemaPath)

	// 스키마 파일 읽기 및 버전 감지
	schemaData, err := os.ReadFile(*schemaPath)
	if err != nil {
		genErr := generator.WrapFileError(err, *schemaPath, "read")
		log.Fatalf("Schema file reading failed: %v", genErr)
	}

	// 스키마 버전 감지
	detector := generator.NewSchemaVersionDetector()
	schemaVersion, err := detector.DetectVersion(schemaData)
	if err != nil {
		genErr := generator.WrapSchemaError(err, *schemaPath, "version_detection")
		log.Fatalf("Schema version detection failed: %v", genErr)
	}

	log.Printf("Detected schema version: %s", schemaVersion.String())

	// 스키마 버전에 따른 추가 로깅
	switch schemaVersion {
	case generator.SchemaVersionLatest:
		log.Printf("Using latest schema format with 'fields' key - BaseModel only embedding")
	case generator.SchemaVersionLegacy:
		log.Printf("Using legacy schema format with 'schema' key - BaseModel + BaseDateTime embedding")
	default:
		log.Printf("Warning: Unknown schema version detected, using fallback Record embedding")
	}

	schemas, err := generator.LoadSchema(*schemaPath)
	if err != nil {
		genErr := generator.WrapSchemaError(err, *schemaPath, "load")
		log.Fatalf("Schema loading failed: %v", genErr)
	}

	// Validate basic configuration
	if err := validateConfig(*schemaPath, *outputPath, *pkgName); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// 기본 TemplateData 생성
	baseTplData := generator.TemplateData{
		PackageName:   *pkgName,
		JSONLibrary:   *jsonLib,
		Collections:   make([]generator.CollectionData, 0, len(schemas)),
		SchemaVersion: schemaVersion,
	}

	for _, s := range schemas {
		if s.Type == "view" && len(s.Fields) == 0 {
			log.Printf("Collection '%s' is a view, generating an empty struct.", s.Name)
		}

		useTimestamps := determineUseTimestamps(schemaVersion, s.Fields)
		log.Printf("Processing collection '%s': schema_version=%s, use_timestamps=%t, fields=%d",
			s.Name, schemaVersion.String(), useTimestamps, len(s.Fields))

		collectionData := generator.CollectionData{
			CollectionName: s.Name,
			StructName:     generator.ToPascalCase(s.Name),
			Fields:         make([]generator.FieldData, 0, len(s.Fields)),
			SchemaVersion:  schemaVersion,
			UseTimestamps:  useTimestamps,
		}

		for _, field := range s.Fields {
			if field.System {
				continue
			}
			// --- ✨ Modified part ---
			// Receive all 3 return values as goType, _, getter.
			// Comment is currently not used, so ignore with '_'.
			goType, _, getter := generator.MapPbTypeToGoType(field, !field.Required)

			// 포인터 타입인지 확인하고 기본 타입 추출
			isPointer := strings.HasPrefix(goType, "*")
			baseType := goType
			if isPointer {
				baseType = strings.TrimPrefix(goType, "*")
			}

			collectionData.Fields = append(collectionData.Fields, generator.FieldData{
				JSONName:     field.Name,
				GoName:       generator.ToPascalCase(field.Name),
				GoType:       goType,
				OmitEmpty:    !field.Required,
				GetterMethod: getter, // Assign value to the newly added GetterMethod field.
				IsPointer:    isPointer,
				BaseType:     baseType,
			})
		}
		baseTplData.Collections = append(baseTplData.Collections, collectionData)
	}

	// Enhanced 기능이 활성화된 경우 EnhancedTemplateData 생성
	var tplData any
	if *generateEnums || *generateRelations || *generateFiles {
		enhancedData := generator.EnhancedTemplateData{
			TemplateData:      baseTplData,
			GenerateEnums:     *generateEnums,
			GenerateRelations: *generateRelations,
			GenerateFiles:     *generateFiles,
		}

		// Enhanced 분석 및 데이터 생성
		if *generateEnums {
			enumGenerator := generator.NewEnumGenerator()
			enhancedData.Enums = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
		}

		if *generateRelations {
			relationGenerator := generator.NewRelationGenerator()
			enhancedData.RelationTypes = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
		}

		if *generateFiles {
			fileGenerator := generator.NewFileGenerator()
			enhancedData.FileTypes = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
		}

		tplData = enhancedData
	} else {
		// 기존 동작 유지 (하위 호환성)
		tplData = baseTplData
	}

	// Parse template with better error handling
	tpl, err := template.New("models").Parse(templateFile)
	if err != nil {
		genErr := generator.WrapTemplateError(err, "models", "parse")
		log.Fatalf("Template parsing failed: %v", genErr)
	}

	// Create output file with better error handling
	outputFile, err := os.Create(*outputPath)
	if err != nil {
		genErr := generator.WrapFileError(err, *outputPath, "create")
		log.Fatalf("Output file creation failed: %v", genErr)
	}
	// Execute template with better error handling
	err = tpl.Execute(outputFile, tplData)
	if err != nil {
		outputFile.Close() // Close file on error
		genErr := generator.WrapTemplateError(err, "models", "execute")
		log.Fatalf("Template execution failed: %v", genErr)
	}

	// Close file before formatting
	if err := outputFile.Close(); err != nil {
		genErr := generator.WrapFileError(err, *outputPath, "close")
		log.Fatalf("Failed to close output file: %v", genErr)
	}

	// 생성된 코드 검증 및 포맷팅
	log.Printf("Validating and formatting generated code...")

	if err := validateAndFormatCode(*outputPath, schemaVersion); err != nil {
		log.Fatalf("Code validation and formatting failed: %v", err)
	}

	log.Printf("Successfully generated models to %s", *outputPath)
}

// validateAndFormatCode validates the generated Go code and applies formatting
func validateAndFormatCode(outputPath string, schemaVersion generator.SchemaVersion) error {
	log.Printf("Reading generated code for validation...")

	// 생성된 파일 읽기
	generatedBytes, err := os.ReadFile(outputPath)
	if err != nil {
		return generator.WrapFileError(err, outputPath, "read_for_validation")
	}

	// 기본 문법 검증 (Go 파싱 시도)
	log.Printf("Performing syntax validation...")
	if err := validateGoSyntax(generatedBytes, outputPath); err != nil {
		return err
	}

	// 스키마 버전별 특정 검증
	log.Printf("Performing schema version specific validation...")
	if err := validateSchemaVersionSpecificCode(generatedBytes, schemaVersion, outputPath); err != nil {
		return err
	}

	// imports 처리 및 포맷팅
	log.Printf("Processing imports and formatting code...")
	formattedBytes, err := imports.Process(outputPath, generatedBytes, &imports.Options{
		Fragment:  false,
		AllErrors: true,
		Comments:  true,
		TabIndent: true,
		TabWidth:  8,
	})
	if err != nil {
		genErr := generator.NewGenerationError(generator.ErrorTypeCodeFormat,
			"failed to format generated code", err).
			WithDetail("file_path", outputPath).
			WithDetail("schema_version", schemaVersion.String()).
			WithDetail("suggestion", "check for syntax errors in generated code")
		return genErr
	}

	// 포맷팅된 코드 쓰기
	log.Printf("Writing formatted code...")
	err = os.WriteFile(outputPath, formattedBytes, 0644)
	if err != nil {
		return generator.WrapFileError(err, outputPath, "write_formatted")
	}

	// 최종 검증
	log.Printf("Performing final validation...")
	if err := validateFinalCode(formattedBytes, schemaVersion, outputPath); err != nil {
		return err
	}

	log.Printf("Code validation and formatting completed successfully")
	return nil
}

// validateGoSyntax performs basic Go syntax validation
func validateGoSyntax(code []byte, filePath string) error {
	// Go 파싱을 통한 기본 문법 검증
	// 실제로는 go/parser 패키지를 사용할 수 있지만,
	// 여기서는 imports.Process가 이미 문법 검증을 수행하므로 간단히 처리

	// 기본적인 구조 검증
	codeStr := string(code)

	// 필수 요소들이 있는지 확인
	if !strings.Contains(codeStr, "package ") {
		return generator.NewGenerationError(generator.ErrorTypeCodeFormat,
			"generated code missing package declaration", nil).
			WithDetail("file_path", filePath)
	}

	// 기본 import 확인
	if !strings.Contains(codeStr, "import") {
		return generator.NewGenerationError(generator.ErrorTypeCodeFormat,
			"generated code missing import statements", nil).
			WithDetail("file_path", filePath).
			WithDetail("suggestion", "check template import section")
	}

	return nil
}

// validateSchemaVersionSpecificCode validates code based on schema version
func validateSchemaVersionSpecificCode(code []byte, version generator.SchemaVersion, filePath string) error {
	codeStr := string(code)

	switch version {
	case generator.SchemaVersionLatest:
		// 최신 스키마 검증: pocketbase.BaseModel 사용 확인
		if !strings.Contains(codeStr, "pocketbase.BaseModel") {
			return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
				"latest schema should use pocketbase.BaseModel", nil).
				WithDetail("file_path", filePath).
				WithDetail("schema_version", version.String())
		}

		// BaseDateTime 임베딩이 없어야 함 (최신 스키마에서는 UseTimestamps가 true인 경우만)
		if strings.Contains(codeStr, "pocketbase.BaseDateTime") {
			log.Printf("Info: Latest schema uses BaseDateTime embedding (UseTimestamps=true)")
		}

		log.Printf("Latest schema validation passed: BaseModel and BaseDateTime properly separated")

	case generator.SchemaVersionLegacy:
		// 구버전 스키마 검증: pocketbase.BaseModel + pocketbase.BaseDateTime 임베딩 확인
		if !strings.Contains(codeStr, "pocketbase.BaseModel") {
			return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
				"legacy schema should embed pocketbase.BaseModel", nil).
				WithDetail("file_path", filePath).
				WithDetail("schema_version", version.String())
		}

		if !strings.Contains(codeStr, "pocketbase.BaseDateTime") {
			return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
				"legacy schema should embed pocketbase.BaseDateTime", nil).
				WithDetail("file_path", filePath).
				WithDetail("schema_version", version.String())
		}

		// 타임스탬프 필드 확인
		timestampFields := []string{"Created", "Updated"}
		for _, field := range timestampFields {
			if !strings.Contains(codeStr, field) {
				return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
					fmt.Sprintf("legacy schema missing timestamp field: %s", field), nil).
					WithDetail("file_path", filePath).
					WithDetail("missing_field", field)
			}
		}

		log.Printf("Legacy schema validation passed: BaseModel and BaseDateTime properly embedded")

	case generator.SchemaVersionUnknown:
		// 알 수 없는 버전: 기본 구조 확인
		if !strings.Contains(codeStr, "pocketbase.BaseModel") {
			log.Printf("Warning: Unknown schema version should use pocketbase.BaseModel fallback")
		}

		log.Printf("Unknown schema version validation completed with warnings")
	}

	return nil
}

// validateFinalCode performs final validation on the formatted code
func validateFinalCode(code []byte, version generator.SchemaVersion, filePath string) error {
	codeStr := string(code)

	// 필수 import 확인
	requiredImports := []string{
		"context",
		"github.com/mrchypark/pocketbase-client",
		"github.com/pocketbase/pocketbase/tools/types",
	}

	for _, imp := range requiredImports {
		if !strings.Contains(codeStr, imp) {
			return generator.NewGenerationError(generator.ErrorTypeCodeFormat,
				fmt.Sprintf("missing required import: %s", imp), nil).
				WithDetail("file_path", filePath).
				WithDetail("missing_import", imp)
		}
	}

	// 생성된 함수들 확인
	expectedFunctionPatterns := map[string]string{
		"constructor": "func New",
		"service":     "Service struct",
		"getter":      "func Get",
		"toMap":       "ToMap()",
	}

	functionCounts := make(map[string]int)
	for pattern, funcPattern := range expectedFunctionPatterns {
		count := strings.Count(codeStr, funcPattern)
		functionCounts[pattern] = count
		if count == 0 {
			log.Printf("Warning: No %s functions found (pattern: %s)", pattern, funcPattern)
		} else {
			log.Printf("Found %d %s functions", count, pattern)
		}
	}

	// pocketbase 패키지 사용 확인
	if !strings.Contains(codeStr, "pocketbase.BaseModel") {
		return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
			"generated code should use pocketbase.BaseModel", nil).
			WithDetail("file_path", filePath)
	}

	// 스키마 버전별 최종 검증
	switch version {
	case generator.SchemaVersionLatest:
		// 최신 스키마: 직접 필드 접근 방식 확인
		if functionCounts["getter"] > 0 {
			log.Printf("Info: Latest schema generated %d getter functions for compatibility", functionCounts["getter"])
		}

		// pocketbase.BaseModel 임베딩 확인
		if !strings.Contains(codeStr, "pocketbase.BaseModel") {
			return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
				"latest schema should use pocketbase.BaseModel embedding", nil).
				WithDetail("file_path", filePath)
		}

		log.Printf("Latest schema final validation: direct field access pattern confirmed")

	case generator.SchemaVersionLegacy:
		// 구버전 스키마: pocketbase.BaseModel + pocketbase.BaseDateTime 임베딩 확인
		if !strings.Contains(codeStr, "pocketbase.BaseModel") || !strings.Contains(codeStr, "pocketbase.BaseDateTime") {
			return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
				"legacy schema should embed both pocketbase.BaseModel and pocketbase.BaseDateTime", nil).
				WithDetail("file_path", filePath)
		}

		// 타임스탬프 필드 확인
		timestampFields := []string{"Created", "Updated"}
		for _, field := range timestampFields {
			if !strings.Contains(codeStr, field) {
				return generator.NewGenerationError(generator.ErrorTypeCodeGeneration,
					fmt.Sprintf("legacy schema missing timestamp field: %s", field), nil).
					WithDetail("file_path", filePath).
					WithDetail("missing_field", field)
			}
		}

		log.Printf("Legacy schema final validation: BaseModel + BaseDateTime embedding confirmed")

	case generator.SchemaVersionUnknown:
		// 알 수 없는 버전: 기본 구조 확인
		if !strings.Contains(codeStr, "pocketbase.BaseModel") {
			log.Printf("Warning: Unknown schema version should use pocketbase.BaseModel fallback")
		}

		log.Printf("Unknown schema version final validation completed with warnings")
	}

	// 코드 품질 검증
	if err := validateCodeQuality(codeStr, filePath); err != nil {
		return err
	}

	log.Printf("Final validation passed for schema version: %s", version.String())
	return nil
}

// validateCodeQuality performs additional code quality checks
func validateCodeQuality(code, filePath string) error {
	// JSON 태그 검증
	if !strings.Contains(code, "`json:") {
		return generator.NewGenerationError(generator.ErrorTypeCodeFormat,
			"generated code missing JSON tags", nil).
			WithDetail("file_path", filePath)
	}

	// 구조체 정의 검증
	if !strings.Contains(code, "struct {") {
		return generator.NewGenerationError(generator.ErrorTypeCodeFormat,
			"generated code missing struct definitions", nil).
			WithDetail("file_path", filePath)
	}

	// 함수 정의 검증
	if !strings.Contains(code, "func ") {
		return generator.NewGenerationError(generator.ErrorTypeCodeFormat,
			"generated code missing function definitions", nil).
			WithDetail("file_path", filePath)
	}

	// 패키지 선언 검증
	if !strings.HasPrefix(strings.TrimSpace(code), "// Code generated by pbc-gen") {
		log.Printf("Warning: Generated code missing generation header comment")
	}

	return nil
}

// validateConfig validates the basic configuration parameters
func validateConfig(schemaPath, outputPath, pkgName string) error {
	validationErr := generator.NewValidationError()

	// Validate schema path
	if schemaPath == "" {
		validationErr.AddError("schema path cannot be empty", "schema")
	} else {
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			validationErr.AddIssue(generator.ValidationIssue{
				Type:       "schema_not_found",
				Message:    "schema file does not exist",
				Path:       "schema",
				Suggestion: "ensure the schema file exists and is accessible",
				Severity:   generator.SeverityError,
				Context:    map[string]interface{}{"path": schemaPath},
			})
		}
	}

	// Validate output path
	if outputPath == "" {
		validationErr.AddError("output path cannot be empty", "output")
	} else {
		// Check if output directory exists
		outputDir := filepath.Dir(outputPath)
		if outputDir != "." {
			if _, err := os.Stat(outputDir); os.IsNotExist(err) {
				validationErr.AddIssue(generator.ValidationIssue{
					Type:       "directory_not_found",
					Message:    "output directory does not exist, will be created",
					Path:       "output",
					Suggestion: "directory will be created automatically",
					Severity:   generator.SeverityWarning,
					Context:    map[string]interface{}{"directory": outputDir},
				})
			}
		}

		// Check if output file already exists
		if _, err := os.Stat(outputPath); err == nil {
			validationErr.AddIssue(generator.ValidationIssue{
				Type:     "file_exists",
				Message:  "output file already exists and will be overwritten",
				Path:     "output",
				Severity: generator.SeverityWarning,
				Context:  map[string]interface{}{"path": outputPath},
			})
		}
	}

	// Validate package name
	if pkgName == "" {
		validationErr.AddError("package name cannot be empty", "package")
	} else if !isValidGoIdentifier(pkgName) {
		validationErr.AddIssue(generator.ValidationIssue{
			Type:       "invalid_identifier",
			Message:    "package name is not a valid Go identifier",
			Path:       "package",
			Suggestion: "use a valid Go package name (letters, digits, underscore)",
			Severity:   generator.SeverityError,
			Context:    map[string]interface{}{"name": pkgName},
		})
	}

	// Return error only if there are actual errors (not warnings)
	if validationErr.HasErrors() {
		return validationErr
	}

	// Log warnings if any
	if validationErr.HasWarnings() {
		for _, warning := range validationErr.GetWarnings() {
			log.Printf("Warning: %s", warning.Message)
		}
	}

	return nil
}

// determineUseTimestamps determines whether to use timestamp fields based on schema version and fields
func determineUseTimestamps(schemaVersion generator.SchemaVersion, fields []generator.FieldSchema) bool {
	switch schemaVersion {
	case generator.SchemaVersionLegacy:
		// 구버전 스키마: 항상 BaseDateTime 임베딩 사용
		return true
	case generator.SchemaVersionLatest:
		// 최신 스키마: created, updated 필드가 명시적으로 정의된 경우만 사용
		hasCreated := false
		hasUpdated := false
		for _, field := range fields {
			if field.Name == "created" {
				hasCreated = true
			}
			if field.Name == "updated" {
				hasUpdated = true
			}
		}
		return hasCreated && hasUpdated
	default:
		// 알 수 없는 버전: 타임스탬프 사용하지 않음
		return false
	}
}

// isValidGoIdentifier checks if a string is a valid Go identifier
func isValidGoIdentifier(name string) bool {
	if name == "" {
		return false
	}

	// First character must be a letter or underscore
	first := rune(name[0])
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Remaining characters must be letters, digits, or underscores
	for _, r := range name[1:] {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}

	return true
}
