package main

import (
	// 1. Add embed package.
	_ "embed"
	"flag"
	"log"
	"os"
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
		log.Fatalf("Error loading schema: %v", err)
	}

	// 기본 TemplateData 생성
	baseTplData := generator.TemplateData{
		PackageName: *pkgName,
		JsonLibrary: *jsonLib,
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
			collectionData.Fields = append(collectionData.Fields, generator.FieldData{
				JsonName:     field.Name,
				GoName:       generator.ToPascalCase(field.Name),
				GoType:       goType,
				OmitEmpty:    !field.Required,
				GetterMethod: getter, // Assign value to the newly added GetterMethod field.
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

	// 3. Parse the embed-injected variable (templateFile) instead of reading from file system.
	tpl, err := template.New("models").Parse(templateFile)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outputFile.Close()

	err = tpl.Execute(outputFile, tplData)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
	outputFile.Close()

	formattedBytes, err := imports.Process(*outputPath, nil, nil)
	if err != nil {
		log.Fatalf("Error formatting generated code: %v", err)
	}

	err = os.WriteFile(*outputPath, formattedBytes, 0644)
	if err != nil {
		log.Fatalf("Error writing formatted code to file: %v", err)
	}

	log.Printf("Successfully generated models to %s\n", *outputPath)
}
