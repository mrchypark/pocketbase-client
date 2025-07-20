package generator

import (
	"os"
	"strings"

	"github.com/goccy/go-json"
)

// LoadSchemaResult contains the result of loading and parsing a schema file
type LoadSchemaResult struct {
	Schemas       []CollectionSchema
	SchemaVersion SchemaVersion
}

// LoadSchema reads a JSON file from the given path and unmarshals it into a slice of CollectionSchema.
// It also detects the schema version and sets it on each collection.
func LoadSchema(filePath string) (*LoadSchemaResult, error) {
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

	// Detect schema version first
	detector := NewSchemaVersionDetector()
	schemaVersion, err := detector.DetectVersion(data)
	if err != nil {
		return nil, NewGenerationError(ErrorTypeSchemaValidate,
			"failed to detect schema version", err).
			WithDetail("file_path", filePath).
			WithDetail("suggestion", "ensure the schema file uses either 'fields' (latest) or 'schema' (legacy) format consistently")
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

	// Set schema version on each collection for later use
	for i := range schemas {
		schemas[i].SchemaVersion = schemaVersion
	}

	return &LoadSchemaResult{
		Schemas:       schemas,
		SchemaVersion: schemaVersion,
	}, nil
}

// BuildTemplateData generates template data from parsed schemas with schema version information.
func BuildTemplateData(schemas []CollectionSchema, packageName string, schemaVersion SchemaVersion) TemplateData {
	var collections []CollectionData

	for _, s := range schemas {
		// System collections can be skipped. (e.g., _superusers)
		if s.System && s.Name == "_superusers" {
			continue
		}

		var fields []FieldData
		// cs.Fields is always populated thanks to custom UnmarshalJSON logic.
		for _, f := range s.Fields {
			// System fields or hidden fields can be skipped as needed.
			if f.System || f.Hidden {
				continue
			}

			goType, _, getterMethod := MapPbTypeToGoType(f, !f.Required)

			// 포인터 타입인지 확인하고 기본 타입 추출
			isPointer := strings.HasPrefix(goType, "*")
			baseType := goType
			if isPointer {
				baseType = strings.TrimPrefix(goType, "*")
			}

			fields = append(fields, FieldData{
				JSONName:     f.Name,
				GoName:       ToPascalCase(f.Name),
				GoType:       goType,
				OmitEmpty:    !f.Required, // Use 'required' field directly
				GetterMethod: getterMethod,
				IsPointer:    isPointer,
				BaseType:     baseType,
			})
		}

		// UseTimestamps 플래그 설정: 구버전 스키마에서는 BaseDateTime 임베딩 필요
		useTimestamps := (schemaVersion == SchemaVersionLegacy)

		collections = append(collections, CollectionData{
			CollectionName: s.Name,
			StructName:     ToPascalCase(s.Name),
			Fields:         fields,
			SchemaVersion:  schemaVersion,
			UseTimestamps:  useTimestamps,
		})
	}

	return TemplateData{
		PackageName:   packageName,
		JSONLibrary:   "github.com/goccy/go-json", // 기본 JSON 라이브러리 설정
		Collections:   collections,
		SchemaVersion: schemaVersion,
	}
}
