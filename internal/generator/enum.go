package generator

import (
	"strings"
)

// EnumGenerator handles enum constant generation for select fields
type EnumGenerator struct{}

// NewEnumGenerator creates a new EnumGenerator instance
func NewEnumGenerator() *EnumGenerator {
	return &EnumGenerator{}
}

// GenerateEnums generates enum data for all collections
func (g *EnumGenerator) GenerateEnums(collections []CollectionData, schemas []CollectionSchema) []EnumData {
	// 성능 최적화: 예상 크기로 슬라이스 미리 할당
	estimatedEnums := 0
	for _, schema := range schemas {
		for _, field := range schema.Fields {
			if !field.System && field.Type == "select" {
				estimatedEnums++
			}
		}
	}

	enums := make([]EnumData, 0, estimatedEnums)

	// 성능 최적화: 스키마를 맵으로 변환하여 O(1) 조회
	schemaMap := make(map[string]CollectionSchema, len(schemas))
	for _, schema := range schemas {
		schemaMap[schema.Name] = schema
	}

	for _, collection := range collections {
		schema, exists := schemaMap[collection.CollectionName]
		if !exists {
			continue
		}

		for _, field := range schema.Fields {
			if field.System || field.Type != "select" {
				continue
			}

			// 성능 최적화: 직접 enum 값 추출하여 불필요한 함수 호출 제거
			if field.Options != nil && len(field.Options.Values) > 0 {
				enumData := g.generateEnumDataOptimized(field, collection.CollectionName)
				enums = append(enums, enumData)
			}
		}
	}

	return enums
}

// generateEnumDataOptimized 성능 최적화된 enum 데이터 생성 함수
func (g *EnumGenerator) generateEnumDataOptimized(field FieldSchema, collectionName string) EnumData {
	values := field.Options.Values
	constants := make([]ConstantData, 0, len(values))

	// 성능 최적화: 문자열 빌더 사용하여 메모리 할당 최소화
	var nameBuilder strings.Builder
	basePrefix := ToPascalCase(collectionName) + ToPascalCase(field.Name)

	for _, value := range values {
		nameBuilder.Reset()
		nameBuilder.WriteString(basePrefix)
		nameBuilder.WriteString(ToPascalCase(value))

		constants = append(constants, ConstantData{
			Name:  nameBuilder.String(),
			Value: value,
		})
	}

	return EnumData{
		CollectionName: collectionName,
		FieldName:      field.Name,
		EnumTypeName:   basePrefix + "Type",
		Constants:      constants,
	}
}
