package main

import (
	// 1. Add embed package.
	_ "embed"
	"flag"
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
		PackageName: *pkgName,
		JSONLibrary: *jsonLib,
		Collections: make([]generator.CollectionData, 0, len(schemas)),
	}

	for _, s := range schemas {
		if s.Type == "view" && len(s.Fields) == 0 {
			log.Printf("Collection '%s' is a view, generating an empty struct.", s.Name)
		}

		collectionData := generator.CollectionData{
			CollectionName: s.Name,
			StructName:     generator.ToPascalCase(s.Name),
			Fields:         make([]generator.FieldData, 0, len(s.Fields)),
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

	// Format generated code with better error handling
	formattedBytes, err := imports.Process(*outputPath, nil, nil)
	if err != nil {
		genErr := generator.NewGenerationError(generator.ErrorTypeCodeFormat,
			"failed to format generated code", err).
			WithDetail("file_path", *outputPath).
			WithDetail("suggestion", "check for syntax errors in generated code")
		log.Fatalf("Code formatting failed: %v", genErr)
	}

	// Write formatted code with better error handling
	err = os.WriteFile(*outputPath, formattedBytes, 0644)
	if err != nil {
		genErr := generator.WrapFileError(err, *outputPath, "write")
		log.Fatalf("Failed to write formatted code: %v", genErr)
	}

	log.Printf("Successfully generated models to %s\n", *outputPath)
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
				Context:    map[string]any{"path": schemaPath},
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
					Context:    map[string]any{"directory": outputDir},
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
				Context:  map[string]any{"path": outputPath},
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
			Context:    map[string]any{"name": pkgName},
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
