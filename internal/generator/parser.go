package generator

import (
	"os"
	"strings"

	"github.com/goccy/go-json"
)

// LoadSchema reads a JSON file from the given path and unmarshals it into a slice of CollectionSchema.
func LoadSchema(filePath string) ([]CollectionSchema, error) {
	// Validate file path
	if filePath == "" {
		return nil, NewGenerationError(ErrorTypeInvalidPath,
			"schema file path cannot be empty", nil)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, NewGenerationError(ErrorTypeSchemaLoad,
			"schema file does not exist", err).
			WithDetail("file_path", filePath).
			WithDetail("suggestion", "ensure the schema file exists and is accessible")
	}

	// Read file with better error context
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, NewGenerationError(ErrorTypeFileRead,
			"failed to read schema file", err).
			WithDetail("file_path", filePath)
	}

	// Validate file is not empty
	if len(data) == 0 {
		return nil, NewGenerationError(ErrorTypeSchemaValidate,
			"schema file is empty", nil).
			WithDetail("file_path", filePath).
			WithDetail("suggestion", "ensure the schema file contains valid JSON data")
	}

	// Parse JSON with better error context
	var schemas []CollectionSchema
	err = json.Unmarshal(data, &schemas)
	if err != nil {
		return nil, NewGenerationError(ErrorTypeSchemaParse,
			"failed to parse schema JSON", err).
			WithDetail("file_path", filePath).
			WithDetail("suggestion", "ensure the schema file contains valid JSON format")
	}

	// Validate parsed schemas
	if len(schemas) == 0 {
		return nil, NewGenerationError(ErrorTypeSchemaValidate,
			"no collections found in schema", nil).
			WithDetail("file_path", filePath).
			WithDetail("suggestion", "ensure the schema contains at least one collection")
	}

	return schemas, nil
}

// BuildTemplateData generates template data from parsed schemas.
func BuildTemplateData(schemas []CollectionSchema, packageName string) TemplateData {
	var collections []CollectionData

	for _, s := range schemas {
		// System collections can be skipped. (e.g., _superusers)
		if s.System && s.Name == "_superusers" {
			continue
		}

		structName := ToPascalCase(s.Name)
		var fields []FieldData
		// cs.Fields is always populated thanks to custom UnmarshalJSON logic.
		for _, f := range s.Fields {
			// System fields or hidden fields can be skipped as needed.
			if f.System || f.Hidden {
				continue
			}

			goType, getterMethod := MapPbTypeToGoType(f, !f.Required)

			// 포인터 타입인지 확인하고 기본 타입 추출
			isPointer := strings.HasPrefix(goType, "*")
			baseType := goType
			if isPointer {
				baseType = strings.TrimPrefix(goType, "*")
			}

			goName := ToPascalCase(f.Name)
			fields = append(fields, FieldData{
				JSONName:     f.Name,
				GoName:       goName,
				GoType:       goType,
				StructTag:    BuildJSONTag(f.Name, !f.Required),
				OmitEmpty:    !f.Required, // Use 'required' field directly
				GetterMethod: getterMethod,
				IsPointer:    isPointer,
				BaseType:     baseType,
				ToMapBlock:   BuildToMapBlock(f.Name, goName, !f.Required),
				ValueOrBlock: BuildValueOrBlock(structName, goName, f.Name, baseType, isPointer),
			})
		}

		collections = append(collections, CollectionData{
			CollectionName: s.Name,
			StructName:     structName,
			Fields:         fields,
		})
	}

	return TemplateData{
		PackageName: packageName,
		Collections: collections,
	}
}
