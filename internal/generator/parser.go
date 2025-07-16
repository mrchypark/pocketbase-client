package generator

import (
	"os"

	"github.com/goccy/go-json"
)

// LoadSchema reads a JSON file from the given path and unmarshals it into a slice of CollectionSchema.
func LoadSchema(filePath string) ([]CollectionSchema, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var schemas []CollectionSchema
	err = json.Unmarshal(data, &schemas)
	if err != nil {
		return nil, err
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

		var fields []FieldData
		// cs.Fields is always populated thanks to custom UnmarshalJSON logic.
		for _, f := range s.Fields {
			// System fields or hidden fields can be skipped as needed.
			if f.System || f.Hidden {
				continue
			}

			goType, _, _ := MapPbTypeToGoType(f, !f.Required)
			fields = append(fields, FieldData{
				JsonName:  f.Name,
				GoName:    ToPascalCase(f.Name),
				GoType:    goType,
				OmitEmpty: !f.Required, // Use 'required' field directly
			})
		}

		collections = append(collections, CollectionData{
			CollectionName: s.Name,
			StructName:     ToPascalCase(s.Name),
			Fields:         fields,
		})
	}

	return TemplateData{
		PackageName: packageName,
		Collections: collections,
	}
}
