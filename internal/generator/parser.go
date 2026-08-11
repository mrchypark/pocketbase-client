package generator

import (
	"fmt"
	"os"
	"strings"

	"github.com/goccy/go-json"
)

// GenerateOptions controls which enhanced constructs (enums, relation types, and
// file types) are generated alongside the base collection models.
type GenerateOptions struct {
	// JSONLibrary is the import path of the JSON package used by the generated
	// code. Defaults to "encoding/json" when empty.
	JSONLibrary string

	Enums     bool
	Relations bool
	Files     bool
}

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
	schemas, err := parseSchemas(data)
	if err != nil {
		return nil, NewGenerationError(ErrorTypeSchemaParse,
			"failed to parse schema JSON", err).
			WithDetail("file_path", filePath).
			WithDetail("suggestion", "ensure the schema file contains a JSON array of collections, a paginated {items:[...]} response from GET /api/collections, or a single collection object")
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

// parseSchemas decodes collection schema JSON into a slice of CollectionSchema,
// accepting the shapes produced by PocketBase:
//
//  1. A plain JSON array of collection objects (the format used by the testdata).
//  2. The paginated response of GET /api/collections:
//     {"items": [...], "page": 1, "perPage": 30, "totalItems": N, "totalPages": 1}
//  3. A single collection object (the response of GET /api/collections/{id}).
//
// The legacy (pre-v0.23) format, which nests fields under a "schema" key, is
// deliberately not supported and returns a clear error instead of silently
// generating incorrect models.
func parseSchemas(data []byte) ([]CollectionSchema, error) {
	// 1. Plain array of collections.
	var schemas []CollectionSchema
	if err := json.Unmarshal(data, &schemas); err == nil {
		if err := rejectLegacyCollections(data); err != nil {
			return nil, err
		}
		return schemas, nil
	}

	// 2. Paginated wrapper: {"items": [...]}.
	var wrapper struct {
		Items json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Items != nil && string(wrapper.Items) != "null" {
		if err := rejectLegacyCollections(wrapper.Items); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(wrapper.Items, &schemas); err != nil {
			return nil, err
		}
		return schemas, nil
	}

	// 3. Single collection object.
	var single CollectionSchema
	if err := json.Unmarshal(data, &single); err == nil && (single.Name != "" || single.ID != "" || len(single.Fields) > 0) {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(data, &obj); err == nil {
			if _, ok := obj["schema"]; ok {
				return nil, legacyFormatError()
			}
		}
		return []CollectionSchema{single}, nil
	}

	return nil, fmt.Errorf("unsupported schema JSON: expected an array of collections, a paginated {items:[...]} response, or a single collection object")
}

// rejectLegacyCollections returns an error if any collection in the given JSON
// array uses the legacy "schema" key.
func rejectLegacyCollections(data []byte) error {
	var collections []map[string]json.RawMessage
	if err := json.Unmarshal(data, &collections); err != nil {
		return nil
	}
	for _, c := range collections {
		if _, ok := c["schema"]; ok {
			return legacyFormatError()
		}
	}
	return nil
}

// legacyFormatError describes the removed legacy schema format.
func legacyFormatError() error {
	return fmt.Errorf("legacy schema format is not supported: collections use a \"schema\" key, but only the latest \"fields\" format (PocketBase v0.23+) is accepted; re-export the schema from a current PocketBase instance")
}

// BuildTemplateData is the single entry point for transforming parsed PocketBase
// schemas into template data. It builds the base collection models and, depending
// on the provided GenerateOptions, the enhanced enums/relation types/file types.
//
// The first entry of opts (if any) is used; without opts a base-only TemplateData
// with the default "encoding/json" JSON library is returned.
func BuildTemplateData(schemas []CollectionSchema, packageName string, opts ...GenerateOptions) TemplateData {
	opt := GenerateOptions{JSONLibrary: "encoding/json"}
	if len(opts) > 0 {
		opt = opts[0]
		if opt.JSONLibrary == "" {
			opt.JSONLibrary = "encoding/json"
		}
	}

	data := TemplateData{
		PackageName:       packageName,
		JSONLibrary:       opt.JSONLibrary,
		GenerateEnums:     opt.Enums,
		GenerateRelations: opt.Relations,
		GenerateFiles:     opt.Files,
		Collections:       make([]CollectionData, 0, len(schemas)),
	}

	for _, s := range schemas {
		// System collections can be skipped. (e.g., _superusers)
		if s.System && s.Name == "_superusers" {
			continue
		}

		collection := CollectionData{
			CollectionName: s.Name,
			StructName:     ToPascalCase(s.Name),
			Fields:         make([]FieldData, 0, len(s.Fields)),
		}

		for _, f := range s.Fields {
			if isSkippedField(f) {
				continue
			}

			goType, _ := MapPbTypeToGoType(f, !f.Required)

			collection.Fields = append(collection.Fields, FieldData{
				JSONName:  f.Name,
				GoName:    ToPascalCase(f.Name),
				GoType:    goType,
				StructTag: BuildJSONTag(f.Name, !f.Required),
				OmitEmpty: !f.Required,
				IsPointer: strings.HasPrefix(goType, "*"),
				BaseType:  strings.TrimPrefix(goType, "*"),
			})
		}

		data.Collections = append(data.Collections, collection)
	}

	data.HasJSONFields = hasJSONField(data.Collections)

	if opt.Enums {
		data.Enums = NewEnumGenerator().GenerateEnums(data.Collections, schemas)
	}
	if opt.Relations {
		data.RelationTypes = NewRelationGenerator().GenerateRelationTypes(data.Collections, schemas)
	}
	if opt.Files {
		data.FileTypes = NewFileGenerator().GenerateFileTypes(data.Collections, schemas)
	}

	return data
}

// hasJSONField reports whether any generated field is a json.RawMessage.
func hasJSONField(collections []CollectionData) bool {
	for _, c := range collections {
		for _, f := range c.Fields {
			if f.GoType == "json.RawMessage" {
				return true
			}
		}
	}
	return false
}

// isSkippedField reports whether a field should be excluded from generated models.
// System fields, hidden fields, and autodate fields are always skipped. Additionally,
// the standard PocketBase fields that are hardcoded in the template (id, collectionId,
// collectionName, created, updated) are skipped by name as a safety net.
func isSkippedField(f FieldSchema) bool {
	if f.System || f.Hidden || f.Type == "autodate" {
		return true
	}
	switch strings.ToLower(f.Name) {
	case "id", "collectionid", "collectionname", "created", "updated":
		return true
	}
	return false
}
