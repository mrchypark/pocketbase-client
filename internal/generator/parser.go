package generator

import (
	"encoding/json"
	"os"
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

// BuildTemplateData는 파싱된 스키마로부터 템플릿 데이터를 생성합니다.
func BuildTemplateData(schemas []CollectionSchema, packageName string) TemplateData {
	var collections []CollectionData

	for _, s := range schemas {
		// 시스템 컬렉션은 건너뛸 수 있습니다. (예: _superusers)
		if s.System && s.Name == "_superusers" {
			continue
		}

		var fields []FieldData
		// cs.Fields는 UnmarshalJSON 커스텀 로직 덕분에 항상 채워져 있습니다.
		for _, f := range s.Fields {
			// 시스템 필드나 숨겨진 필드는 필요에 따라 건너뛸 수 있습니다.
			if f.System || f.Hidden {
				continue
			}

			goType := MapPbTypeToGoType(f)
			fields = append(fields, FieldData{
				JsonName:  f.Name,
				GoName:    ToPascalCase(f.Name),
				GoType:    goType,
				OmitEmpty: !f.Required, // 'required' 필드를 직접 사용
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
