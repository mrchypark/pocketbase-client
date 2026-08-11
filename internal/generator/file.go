package generator

// FileGenerator handles file type generation for file fields
type FileGenerator struct{}

// NewFileGenerator creates a new FileGenerator instance
func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

// GenerateFileTypes generates file type data for all collections
func (g *FileGenerator) GenerateFileTypes(collections []CollectionData, schemas []CollectionSchema) []FileTypeData {
	// 성능 최적화: 예상 크기로 슬라이스 미리 할당
	estimatedFiles := 0
	for _, schema := range schemas {
		for _, field := range schema.Fields {
			if !field.System && field.Type == "file" {
				estimatedFiles++
			}
		}
	}

	fileTypes := make([]FileTypeData, 0, estimatedFiles)

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
			if field.System || field.Type != "file" {
				continue
			}

			// 성능 최적화: 직접 file 정보 추출
			fileType := g.generateFileTypeDataOptimized(field, collection.CollectionName)
			fileTypes = append(fileTypes, fileType)
		}
	}

	return fileTypes
}

// generateFileTypeDataOptimized 성능 최적화된 file 타입 데이터 생성 함수
func (g *FileGenerator) generateFileTypeDataOptimized(field FieldSchema, _ string) FileTypeData {
	// 성능 최적화: 미리 계산된 값들 사용
	fileTypeName := ToPascalCase(field.Name) + "File"

	// 다중 파일 여부 및 썸네일 정보 확인
	isMulti := field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect > 1
	hasThumbnails := field.Options != nil && len(field.Options.Thumbs) > 0
	var thumbnailSizes []string
	if hasThumbnails {
		thumbnailSizes = field.Options.Thumbs
	}

	return FileTypeData{
		TypeName:       fileTypeName,
		IsMulti:        isMulti,
		HasThumbnails:  hasThumbnails,
		ThumbnailSizes: thumbnailSizes,
	}
}
