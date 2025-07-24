package generator

import (
	"fmt"
	"strings"
)

// FieldProcessor 인터페이스는 스키마 버전에 따른 필드 처리 로직을 정의합니다.
type FieldProcessor interface {
	// ProcessFields는 스키마 필드들을 처리하여 최종 FieldData 슬라이스를 반환합니다.
	ProcessFields(fields []FieldSchema, collectionName string) ([]FieldData, error)

	// ShouldEmbedBaseDateTime은 BaseDateTime 구조체를 임베딩해야 하는지 여부를 반환합니다.
	ShouldEmbedBaseDateTime() bool

	// GetRequiredImports는 필요한 import 목록을 반환합니다.
	GetRequiredImports() []string
}

// LatestFieldProcessor는 최신 스키마(fields 키)용 필드 처리기입니다.
type LatestFieldProcessor struct{}

// NewLatestFieldProcessor는 최신 스키마용 필드 처리기를 생성합니다.
func NewLatestFieldProcessor() *LatestFieldProcessor {
	return &LatestFieldProcessor{}
}

// ProcessFields는 최신 스키마의 필드들을 처리합니다.
// 최신 스키마에서는 created/updated 필드가 명시된 경우만 개별 필드로 추가합니다.
func (p *LatestFieldProcessor) ProcessFields(fields []FieldSchema, collectionName string) ([]FieldData, error) {
	var result []FieldData
	fieldNames := make(map[string]bool) // 중복 방지용

	for _, field := range fields {
		// System fields or hidden fields can be skipped
		if field.System || field.Hidden {
			continue
		}

		// 기본 필드들(id, created, updated, collectionId, collectionName)은
		// 스키마에 명시된 경우만 개별 필드로 추가
		if p.isSystemField(field.Name) {
			// 시스템 필드는 스키마에 명시되어 있으면 개별 필드로 처리
			if fieldNames[field.Name] {
				continue // 중복 방지
			}
			fieldNames[field.Name] = true
		}

		// 필드 데이터 생성
		fieldData, err := p.createFieldData(field)
		if err != nil {
			return nil, fmt.Errorf("failed to create field data for %s: %w", field.Name, err)
		}

		result = append(result, fieldData)
	}

	return result, nil
}

// ShouldEmbedBaseDateTime은 최신 스키마에서는 BaseDateTime을 임베딩하지 않습니다.
func (p *LatestFieldProcessor) ShouldEmbedBaseDateTime() bool {
	return false
}

// GetRequiredImports는 필요한 import 목록을 반환합니다.
func (p *LatestFieldProcessor) GetRequiredImports() []string {
	return []string{
		"github.com/pocketbase/pocketbase/tools/types",
	}
}

// isSystemField는 시스템 필드인지 확인합니다.
func (p *LatestFieldProcessor) isSystemField(fieldName string) bool {
	systemFields := map[string]bool{
		"id":             true,
		"created":        true,
		"updated":        true,
		"collectionId":   true,
		"collectionName": true,
	}
	return systemFields[fieldName]
}

// createFieldData는 FieldSchema로부터 FieldData를 생성합니다.
func (p *LatestFieldProcessor) createFieldData(field FieldSchema) (FieldData, error) {
	goType, _, getterMethod := MapPbTypeToGoTypeWithGeneric(field, !field.Required, false)

	// 포인터 타입인지 확인
	isPointer := strings.HasPrefix(goType, "*")
	baseType := goType
	if isPointer {
		baseType = strings.TrimPrefix(goType, "*")
	}

	// omitEmpty는 필수가 아닌 필드에 대해 true
	omitEmpty := !field.Required

	return FieldData{
		JSONName:     field.Name,
		GoName:       ToPascalCase(field.Name),
		GoType:       goType,
		OmitEmpty:    omitEmpty,
		GetterMethod: getterMethod,
		IsPointer:    isPointer,
		BaseType:     baseType,
	}, nil
}

// LegacyFieldProcessor는 구버전 스키마(schema 키)용 필드 처리기입니다.
type LegacyFieldProcessor struct{}

// NewLegacyFieldProcessor는 구버전 스키마용 필드 처리기를 생성합니다.
func NewLegacyFieldProcessor() *LegacyFieldProcessor {
	return &LegacyFieldProcessor{}
}

// ProcessFields는 구버전 스키마의 필드들을 처리합니다.
// 구버전 스키마에서는 BaseDateTime 임베딩을 자동 적용하고,
// 기본 필드들(id, created, updated)이 이미 정의되어 있으면 기존 정의를 우선 사용합니다.
func (p *LegacyFieldProcessor) ProcessFields(fields []FieldSchema, collectionName string) ([]FieldData, error) {
	var result []FieldData
	fieldNames := make(map[string]bool) // 중복 방지용

	// 먼저 스키마에 정의된 필드들을 처리
	for _, field := range fields {
		// System fields or hidden fields can be skipped
		if field.System || field.Hidden {
			continue
		}

		if fieldNames[field.Name] {
			continue // 중복 방지
		}
		fieldNames[field.Name] = true

		// 시스템 필드가 스키마에 명시되어 있으면 해당 정의를 사용
		// (BaseDateTime 임베딩으로 자동 추가되는 것보다 우선)
		fieldData, err := p.createFieldData(field)
		if err != nil {
			return nil, fmt.Errorf("failed to create field data for %s: %w", field.Name, err)
		}

		result = append(result, fieldData)
	}

	// 구버전에서는 BaseDateTime 임베딩으로 created, updated가 자동 제공되므로
	// 별도로 추가할 필요 없음 (BaseDateTime 구조체에 포함됨)

	return result, nil
}

// ShouldEmbedBaseDateTime은 구버전 스키마에서는 BaseDateTime을 임베딩합니다.
func (p *LegacyFieldProcessor) ShouldEmbedBaseDateTime() bool {
	return true
}

// GetRequiredImports는 필요한 import 목록을 반환합니다.
func (p *LegacyFieldProcessor) GetRequiredImports() []string {
	return []string{
		"github.com/pocketbase/pocketbase/tools/types",
	}
}

// createFieldData는 FieldSchema로부터 FieldData를 생성합니다.
func (p *LegacyFieldProcessor) createFieldData(field FieldSchema) (FieldData, error) {
	goType, _, getterMethod := MapPbTypeToGoTypeWithGeneric(field, !field.Required, false)

	// 포인터 타입인지 확인
	isPointer := strings.HasPrefix(goType, "*")
	baseType := goType
	if isPointer {
		baseType = strings.TrimPrefix(goType, "*")
	}

	// omitEmpty는 필수가 아닌 필드에 대해 true
	omitEmpty := !field.Required

	return FieldData{
		JSONName:     field.Name,
		GoName:       ToPascalCase(field.Name),
		GoType:       goType,
		OmitEmpty:    omitEmpty,
		GetterMethod: getterMethod,
		IsPointer:    isPointer,
		BaseType:     baseType,
	}, nil
}

// GenericFieldProcessor는 제네릭 방식을 사용하는 필드 처리기입니다.
type GenericFieldProcessor struct {
	baseProcessor FieldProcessor
}

// NewGenericFieldProcessor는 제네릭 필드 처리기를 생성합니다.
func NewGenericFieldProcessor(version SchemaVersion) *GenericFieldProcessor {
	var baseProcessor FieldProcessor
	switch version {
	case SchemaVersionLatest:
		baseProcessor = NewLatestFieldProcessor()
	case SchemaVersionLegacy:
		baseProcessor = NewLegacyFieldProcessor()
	default:
		baseProcessor = NewLatestFieldProcessor()
	}

	return &GenericFieldProcessor{
		baseProcessor: baseProcessor,
	}
}

// ProcessFields는 제네릭 방식으로 필드들을 처리합니다.
func (p *GenericFieldProcessor) ProcessFields(fields []FieldSchema, collectionName string) ([]FieldData, error) {
	var result []FieldData
	fieldNames := make(map[string]bool) // 중복 방지용

	for _, field := range fields {
		// System fields or hidden fields can be skipped
		if field.System || field.Hidden {
			continue
		}

		// 중복 방지
		if fieldNames[field.Name] {
			continue
		}
		fieldNames[field.Name] = true

		// 제네릭 방식으로 필드 데이터 생성
		fieldData, err := p.createGenericFieldData(field)
		if err != nil {
			return nil, fmt.Errorf("failed to create generic field data for %s: %w", field.Name, err)
		}

		result = append(result, fieldData)
	}

	return result, nil
}

// ShouldEmbedBaseDateTime은 기본 처리기의 설정을 따릅니다.
func (p *GenericFieldProcessor) ShouldEmbedBaseDateTime() bool {
	return p.baseProcessor.ShouldEmbedBaseDateTime()
}

// GetRequiredImports는 제네릭 방식에 필요한 import 목록을 반환합니다.
func (p *GenericFieldProcessor) GetRequiredImports() []string {
	imports := p.baseProcessor.GetRequiredImports()
	// 제네릭 방식에 필요한 추가 import들
	imports = append(imports, "context")
	return imports
}

// createGenericFieldData는 제네릭 방식으로 FieldData를 생성합니다.
func (p *GenericFieldProcessor) createGenericFieldData(field FieldSchema) (FieldData, error) {
	// 제네릭 방식 사용
	goType, _, getterMethod := MapPbTypeToGoTypeWithGeneric(field, !field.Required, true)

	// 포인터 타입인지 확인
	isPointer := strings.HasPrefix(goType, "*")
	baseType := goType
	if isPointer {
		baseType = strings.TrimPrefix(goType, "*")
	}

	// omitEmpty는 필수가 아닌 필드에 대해 true
	omitEmpty := !field.Required

	return FieldData{
		JSONName:     field.Name,
		GoName:       ToPascalCase(field.Name),
		GoType:       goType,
		OmitEmpty:    omitEmpty,
		GetterMethod: getterMethod,
		IsPointer:    isPointer,
		BaseType:     baseType,
	}, nil
}

// CreateFieldProcessor는 스키마 버전에 따라 적절한 필드 처리기를 생성합니다.
func CreateFieldProcessor(version SchemaVersion) FieldProcessor {
	switch version {
	case SchemaVersionLatest:
		return NewLatestFieldProcessor()
	case SchemaVersionLegacy:
		return NewLegacyFieldProcessor()
	default:
		// 기본값으로 최신 버전 처리기 사용
		return NewLatestFieldProcessor()
	}
}

// CreateFieldProcessorWithGeneric는 제네릭 지원 여부에 따라 필드 처리기를 생성합니다.
func CreateFieldProcessorWithGeneric(version SchemaVersion, useGeneric bool) FieldProcessor {
	if useGeneric {
		return NewGenericFieldProcessor(version)
	}
	return CreateFieldProcessor(version)
}

// ProcessFieldsWithVersion은 스키마 버전에 따라 필드를 처리하는 헬퍼 함수입니다.
func ProcessFieldsWithVersion(fields []FieldSchema, collectionName string, version SchemaVersion) ([]FieldData, bool, error) {
	processor := CreateFieldProcessor(version)

	processedFields, err := processor.ProcessFields(fields, collectionName)
	if err != nil {
		return nil, false, err
	}

	useTimestamps := processor.ShouldEmbedBaseDateTime()

	return processedFields, useTimestamps, nil
}
