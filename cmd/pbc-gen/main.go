package main

import (
	// 1. embed 패키지를 추가합니다.
	_ "embed"
	"flag"
	"log"
	"os"
	"text/template"

	"github.com/mrchypark/pocketbase-client/internal/generator"
	"golang.org/x/tools/imports"
)

// 2. go:embed 지시어를 사용해 템플릿 파일의 내용을 변수에 저장합니다.
// 이 코드는 main.go 파일 기준으로 템플릿 파일의 상대 경로를 지정해야 합니다.
//
//go:embed template.go.tpl
var templateFile string

func main() {
	schemaPath := flag.String("schema", "./pb_schema.json", "Input file path (pb_schema.json)")
	outputPath := flag.String("path", "./models.gen.go", "Output file path")
	pkgName := flag.String("pkgname", "models", "Package name for the generated file")
	jsonLib := flag.String("jsonlib", "encoding/json", "JSON library to use (e.g., github.com/goccy/go-json)")

	flag.Parse()

	log.Printf("Generating Go models from schema: %s\n", *schemaPath)

	schemas, err := generator.LoadSchema(*schemaPath)
	if err != nil {
		log.Fatalf("Error loading schema: %v", err)
	}

	tplData := generator.TemplateData{
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
			goType := generator.MapPbTypeToGoType(field)
			collectionData.Fields = append(collectionData.Fields, generator.FieldData{
				JsonName:  field.Name,
				GoName:    generator.ToPascalCase(field.Name),
				GoType:    goType,
				OmitEmpty: !field.Required,
			})
		}
		tplData.Collections = append(tplData.Collections, collectionData)
	}

	// 3. 파일 시스템에서 읽는 대신, embed로 주입된 변수(templateFile)를 파싱합니다.
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
