package main

import (
	"flag"
	"log"
	"os"
	"text/template"

	"github.com/mrchypark/pocketbase-client/internal/generator" // 여러분의 실제 경로로 수정하세요
	"golang.org/x/tools/imports"
)

func main() {
	schemaPath := flag.String("schema", "./pb_schema.json", "Input file path (pb_schema.json)")
	outputPath := flag.String("path", "./models.gen.go", "Output file path")
	pkgName := flag.String("pkgname", "models", "Package name for the generated file")

	flag.Parse()

	log.Printf("Generating Go models from schema: %s\n", *schemaPath)

	schemas, err := generator.LoadSchema(*schemaPath)
	if err != nil {
		log.Fatalf("Error loading schema: %v", err)
	}

	tplData := generator.TemplateData{
		PackageName: *pkgName,
		Collections: make([]generator.CollectionData, 0, len(schemas)),
	}

	for _, s := range schemas {
		// 'view' 타입 컬렉션은 SQL 쿼리로 정의되므로 실제 필드가 없을 수 있습니다.
		// 이 경우에도 빈 구조체를 생성하도록 합니다.
		if s.Type == "view" && len(s.Fields) == 0 {
			log.Printf("Collection '%s' is a view, generating an empty struct.", s.Name)
		}

		collectionData := generator.CollectionData{
			CollectionName: s.Name,
			StructName:     generator.ToPascalCase(s.Name),
			Fields:         make([]generator.FieldData, 0, len(s.Fields)),
		}

		for _, field := range s.Fields {
			// 시스템 필드는 일반적으로 클라이언트 모델에 포함하지 않습니다.
			if field.System {
				continue
			}
			goType, _ := generator.MapPbTypeToGoType(field)
			collectionData.Fields = append(collectionData.Fields, generator.FieldData{
				JsonName:  field.Name,
				GoName:    generator.ToPascalCase(field.Name),
				GoType:    goType,
				OmitEmpty: !field.Required,
			})
		}
		tplData.Collections = append(tplData.Collections, collectionData)
	}

	// 템플릿 파일 경로는 실제 프로젝트 구조에 맞게 수정해주세요.
	tpl, err := template.ParseFiles("internal/generator/template.go.tpl")
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	outputFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Error creating output file: %v", err)
	}
	defer outputFile.Close()

	// 템플릿 실행 전에 임시 파일에 쓰고 포매팅 후 최종 파일에 쓰는 것이 더 안전합니다.
	// 여기서는 간결함을 위해 바로 씁니다.
	err = tpl.Execute(outputFile, tplData)
	if err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
	outputFile.Close() // 파일을 닫아야 다음 단계에서 읽을 수 있습니다.

	// 생성된 코드 포매팅 및 임포트 정리
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
