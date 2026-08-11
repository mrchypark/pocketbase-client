package generator

// RelationGenerator handles relation type generation for relation fields
type RelationGenerator struct{}

// NewRelationGenerator creates a new RelationGenerator instance
func NewRelationGenerator() *RelationGenerator {
	return &RelationGenerator{}
}

// GenerateRelationTypes generates relation type data for all collections
func (g *RelationGenerator) GenerateRelationTypes(collections []CollectionData, schemas []CollectionSchema) []RelationTypeData {
	// 성능 최적화: 예상 크기로 슬라이스 미리 할당
	estimatedRelations := 0
	for _, schema := range schemas {
		for _, field := range schema.Fields {
			if !field.System && field.Type == "relation" {
				estimatedRelations++
			}
		}
	}

	relationTypes := make([]RelationTypeData, 0, estimatedRelations)
	seenRelations := make(map[string]bool) // 중복 제거를 위한 맵

	// 성능 최적화: 스키마를 맵으로 변환하여 O(1) 조회
	schemaMap := make(map[string]CollectionSchema, len(schemas))
	collectionIDMap := make(map[string]string, len(schemas)) // ID -> Name 매핑
	for _, schema := range schemas {
		schemaMap[schema.Name] = schema
		if schema.ID != "" {
			collectionIDMap[schema.ID] = schema.Name
		}
	}

	for _, collection := range collections {
		schema, exists := schemaMap[collection.CollectionName]
		if !exists {
			continue
		}

		for _, field := range schema.Fields {
			if field.System || field.Type != "relation" {
				continue
			}

			// 성능 최적화: 직접 relation 정보 추출
			if field.Options != nil && field.Options.CollectionID != "" {
				targetCollection, exists := collectionIDMap[field.Options.CollectionID]
				if exists {
					relationType := g.generateRelationTypeDataOptimized(field, collection.CollectionName, targetCollection)
					// 중복 제거: 같은 타입명이 이미 존재하는지 확인
					if !seenRelations[relationType.TypeName] {
						seenRelations[relationType.TypeName] = true
						relationTypes = append(relationTypes, relationType)
					}
				}
			}
		}
	}

	return relationTypes
}

// generateRelationTypeDataOptimized 성능 최적화된 relation 타입 데이터 생성 함수
func (g *RelationGenerator) generateRelationTypeDataOptimized(field FieldSchema, _, targetCollection string) RelationTypeData {
	// 성능 최적화: 미리 계산된 값들 사용
	relationTypeName := ToPascalCase(targetCollection) + "Relation"
	targetTypeName := ToPascalCase(targetCollection)

	// 다중 관계 여부 확인
	isMulti := field.Options.MaxSelect != nil && *field.Options.MaxSelect > 1

	return RelationTypeData{
		TypeName:         relationTypeName,
		TargetCollection: targetCollection,
		TargetTypeName:   targetTypeName,
		IsMulti:          isMulti,
	}
}
