package generator

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

// CodeGenerator는 스키마 버전에 따른 코드 생성을 담당합니다.
type CodeGenerator struct {
	schemaVersion SchemaVersion
	templateData  *TemplateData
}

// NewCodeGenerator는 새로운 CodeGenerator를 생성합니다.
func NewCodeGenerator(version SchemaVersion, templateData *TemplateData) *CodeGenerator {
	return &CodeGenerator{
		schemaVersion: version,
		templateData:  templateData,
	}
}

// GenerateStruct는 스키마 버전에 따라 구조체 코드를 생성합니다.
func (g *CodeGenerator) GenerateStruct(collection CollectionData) (string, error) {
	switch collection.SchemaVersion {
	case SchemaVersionLatest:
		return g.generateLatestStruct(collection)
	case SchemaVersionLegacy:
		return g.generateLegacyStruct(collection)
	default:
		return g.generateDefaultStruct(collection)
	}
}

// generateLatestStruct는 최신 스키마용 구조체를 생성합니다.
// BaseModel만 임베딩하고, created/updated는 스키마에 명시된 경우만 개별 필드로 추가합니다.
func (g *CodeGenerator) generateLatestStruct(collection CollectionData) (string, error) {
	var builder strings.Builder

	// 구조체 선언 시작
	builder.WriteString(fmt.Sprintf("// %s represents a record from the %s collection\n",
		collection.StructName, collection.CollectionName))
	builder.WriteString(fmt.Sprintf("type %s struct {\n", collection.StructName))

	// BaseModel 임베딩 (ID, CollectionID, CollectionName 포함)
	builder.WriteString("\tBaseModel\n")

	// 개별 필드들 추가
	for _, field := range collection.Fields {
		// 시스템 필드들은 BaseModel에 이미 포함되어 있으므로 제외
		if g.isSystemFieldInBaseModel(field.JSONName) {
			continue
		}

		// 필드 생성
		fieldLine := g.generateFieldLine(field)
		builder.WriteString(fmt.Sprintf("\t%s\n", fieldLine))
	}

	builder.WriteString("}\n\n")

	// 헬퍼 메서드들 생성
	helperMethods := g.generateHelperMethods(collection)
	builder.WriteString(helperMethods)

	return builder.String(), nil
}

// generateLegacyStruct는 구버전 스키마용 구조체를 생성합니다.
// BaseModel + BaseDateTime을 임베딩하여 기존 방식과 호환성을 유지합니다.
func (g *CodeGenerator) generateLegacyStruct(collection CollectionData) (string, error) {
	var builder strings.Builder

	// 구조체 선언 시작
	builder.WriteString(fmt.Sprintf("// %s represents a record from the %s collection\n",
		collection.StructName, collection.CollectionName))
	builder.WriteString(fmt.Sprintf("type %s struct {\n", collection.StructName))

	// BaseModel 임베딩 (ID, CollectionID, CollectionName 포함)
	builder.WriteString("\tBaseModel\n")

	// BaseDateTime 임베딩 (Created, Updated 포함)
	if collection.UseTimestamps {
		builder.WriteString("\tBaseDateTime\n")
	}

	// 개별 필드들 추가
	for _, field := range collection.Fields {
		// 시스템 필드들은 BaseModel이나 BaseDateTime에 이미 포함되어 있으므로 제외
		if g.isSystemFieldInBaseModel(field.JSONName) || g.isSystemFieldInBaseDateTime(field.JSONName) {
			continue
		}

		// 필드 생성
		fieldLine := g.generateFieldLine(field)
		builder.WriteString(fmt.Sprintf("\t%s\n", fieldLine))
	}

	builder.WriteString("}\n\n")

	// 헬퍼 메서드들 생성
	helperMethods := g.generateHelperMethods(collection)
	builder.WriteString(helperMethods)

	return builder.String(), nil
}

// generateDefaultStruct는 기본 구조체를 생성합니다 (최신 스키마와 동일).
func (g *CodeGenerator) generateDefaultStruct(collection CollectionData) (string, error) {
	return g.generateLatestStruct(collection)
}

// generateFieldLine은 개별 필드 라인을 생성합니다.
func (g *CodeGenerator) generateFieldLine(field FieldData) string {
	jsonTag := fmt.Sprintf(`json:"%s"`, field.JSONName)
	if field.OmitEmpty {
		jsonTag = fmt.Sprintf(`json:"%s,omitempty"`, field.JSONName)
	}

	return fmt.Sprintf("%s %s `%s`", field.GoName, field.GoType, jsonTag)
}

// generateHelperMethods는 구조체용 헬퍼 메서드들을 생성합니다.
func (g *CodeGenerator) generateHelperMethods(collection CollectionData) string {
	var builder strings.Builder

	// TableName 메서드 생성
	builder.WriteString(fmt.Sprintf("// TableName returns the table name for %s\n", collection.StructName))
	builder.WriteString(fmt.Sprintf("func (%s) TableName() string {\n", strings.ToLower(collection.StructName[:1])))
	builder.WriteString(fmt.Sprintf("\treturn \"%s\"\n", collection.CollectionName))
	builder.WriteString("}\n\n")

	// 포인터 타입 필드들에 대한 ValueOr 메서드 생성
	for _, field := range collection.Fields {
		if field.IsPointer {
			builder.WriteString(g.generateValueOrMethod(collection.StructName, field))
		}
	}

	return builder.String()
}

// generateValueOrMethod는 포인터 필드용 ValueOr 메서드를 생성합니다.
func (g *CodeGenerator) generateValueOrMethod(structName string, field FieldData) string {
	var builder strings.Builder

	methodName := fmt.Sprintf("%sValueOr", field.GoName)
	receiverName := strings.ToLower(structName[:1])

	builder.WriteString(fmt.Sprintf("// %s returns the value of %s or the default value if nil\n",
		methodName, field.GoName))
	builder.WriteString(fmt.Sprintf("func (%s *%s) %s(defaultValue %s) %s {\n",
		receiverName, structName, methodName, field.BaseType, field.BaseType))
	builder.WriteString(fmt.Sprintf("\tif %s.%s != nil {\n", receiverName, field.GoName))
	builder.WriteString(fmt.Sprintf("\t\treturn *%s.%s\n", receiverName, field.GoName))
	builder.WriteString("\t}\n")
	builder.WriteString("\treturn defaultValue\n")
	builder.WriteString("}\n\n")

	return builder.String()
}

// isSystemFieldInBaseModel은 BaseModel에 포함된 시스템 필드인지 확인합니다.
func (g *CodeGenerator) isSystemFieldInBaseModel(fieldName string) bool {
	baseModelFields := map[string]bool{
		"id":             true,
		"collectionId":   true,
		"collectionName": true,
	}
	return baseModelFields[fieldName]
}

// isSystemFieldInBaseDateTime은 BaseDateTime에 포함된 시스템 필드인지 확인합니다.
func (g *CodeGenerator) isSystemFieldInBaseDateTime(fieldName string) bool {
	baseDateTimeFields := map[string]bool{
		"created": true,
		"updated": true,
	}
	return baseDateTimeFields[fieldName]
}

// GenerateCode는 전체 코드를 생성합니다.
func (g *CodeGenerator) GenerateCode(templateContent string) (string, error) {
	// 템플릿 파싱
	tpl, err := template.New("models").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("template parsing failed: %w", err)
	}

	// 템플릿 데이터 준비
	templateData := g.prepareTemplateData()

	// 템플릿 실행
	var builder strings.Builder
	err = tpl.Execute(&builder, templateData)
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return builder.String(), nil
}

// GenerateGenericServices generates service code using the existing template
func (g *CodeGenerator) GenerateGenericServices() (string, error) {
	// 기존 템플릿 파일 읽기
	templatePath := "cmd/pbc-gen/template.go.tpl"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		// 템플릿 파일이 없으면 내장 템플릿 사용
		return g.generateWithBuiltinTemplate()
	}

	// 기존 템플릿 사용
	tpl, err := template.New("services").Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("template parsing failed: %w", err)
	}

	// 템플릿 데이터 준비
	templateData := g.prepareTemplateData()

	// 템플릿 실행
	var builder strings.Builder
	err = tpl.Execute(&builder, templateData)
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return builder.String(), nil
}

// generateWithBuiltinTemplate generates code using built-in template as fallback
func (g *CodeGenerator) generateWithBuiltinTemplate() (string, error) {
	// 내장 제네릭 서비스 템플릿 사용
	tpl, err := template.New("services").Parse(GenericServiceTemplate)
	if err != nil {
		return "", fmt.Errorf("builtin template parsing failed: %w", err)
	}

	// 템플릿 데이터 준비
	templateData := g.prepareTemplateData()

	// 템플릿 실행
	var builder strings.Builder
	err = tpl.Execute(&builder, templateData)
	if err != nil {
		return "", fmt.Errorf("builtin template execution failed: %w", err)
	}

	return builder.String(), nil
}

// GenerateStructs generates struct definitions using struct template
func (g *CodeGenerator) GenerateStructs() (string, error) {
	// 구조체 템플릿 사용
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tpl, err := template.New("structs").Funcs(funcMap).Parse(StructTemplate)
	if err != nil {
		return "", fmt.Errorf("struct template parsing failed: %w", err)
	}

	// 템플릿 데이터 준비
	templateData := g.prepareTemplateData()

	// 템플릿 실행
	var builder strings.Builder
	err = tpl.Execute(&builder, templateData)
	if err != nil {
		return "", fmt.Errorf("struct template execution failed: %w", err)
	}

	return builder.String(), nil
}

// prepareTemplateData는 템플릿 실행용 데이터를 준비합니다.
func (g *CodeGenerator) prepareTemplateData() interface{} {
	// 스키마 버전 정보를 포함한 템플릿 데이터 반환
	enhancedData := EnhancedTemplateData{
		TemplateData: *g.templateData,
	}

	// 각 컬렉션에 스키마 버전 정보 설정
	for i := range enhancedData.Collections {
		enhancedData.Collections[i].SchemaVersion = g.schemaVersion

		// 필드 처리기를 사용하여 필드 처리 및 UseTimestamps 설정
		processor := CreateFieldProcessorWithGeneric(g.schemaVersion, g.templateData.UseGeneric)
		enhancedData.Collections[i].UseTimestamps = processor.ShouldEmbedBaseDateTime()
	}

	return enhancedData
}

// GenerateStructsOnly는 구조체 코드만 생성합니다 (템플릿 없이).
func (g *CodeGenerator) GenerateStructsOnly() (string, error) {
	var builder strings.Builder

	// 패키지 선언
	builder.WriteString(fmt.Sprintf("package %s\n\n", g.templateData.PackageName))

	// 필요한 import 추가
	imports := g.getRequiredImports()
	if len(imports) > 0 {
		builder.WriteString("import (\n")
		for _, imp := range imports {
			builder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		builder.WriteString(")\n\n")
	}

	// 각 컬렉션에 대한 구조체 생성
	for _, collection := range g.templateData.Collections {
		structCode, err := g.GenerateStruct(collection)
		if err != nil {
			return "", fmt.Errorf("failed to generate struct for %s: %w", collection.CollectionName, err)
		}
		builder.WriteString(structCode)
	}

	return builder.String(), nil
}

// GenerateWithServices는 구조체와 서비스 코드를 모두 생성합니다.
func (g *CodeGenerator) GenerateWithServices() (string, error) {
	var builder strings.Builder

	// 패키지 선언
	builder.WriteString(fmt.Sprintf("package %s\n\n", g.templateData.PackageName))

	// 필요한 import 추가
	imports := g.getRequiredImportsWithServices()
	if len(imports) > 0 {
		builder.WriteString("import (\n")
		for _, imp := range imports {
			builder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		builder.WriteString(")\n\n")
	}

	// 각 컬렉션에 대한 구조체 생성
	for _, collection := range g.templateData.Collections {
		structCode, err := g.GenerateStruct(collection)
		if err != nil {
			return "", fmt.Errorf("failed to generate struct for %s: %w", collection.CollectionName, err)
		}
		builder.WriteString(structCode)
	}

	// 제네릭 사용 시 서비스 코드 생성
	if g.templateData.UseGeneric {
		serviceGenerator := NewServiceGenerator()
		serviceCode := serviceGenerator.GenerateAllServices(g.templateData.Collections, true)
		builder.WriteString(serviceCode)
	}

	return builder.String(), nil
}

// getRequiredImports는 필요한 import 목록을 반환합니다.
func (g *CodeGenerator) getRequiredImports() []string {
	processor := CreateFieldProcessor(g.schemaVersion)
	imports := processor.GetRequiredImports()

	// JSON 라이브러리 추가
	if g.templateData.JSONLibrary != "" {
		imports = append(imports, g.templateData.JSONLibrary)
	}

	// 중복 제거
	uniqueImports := make(map[string]bool)
	var result []string
	for _, imp := range imports {
		if !uniqueImports[imp] {
			uniqueImports[imp] = true
			result = append(result, imp)
		}
	}

	return result
}

// getRequiredImportsWithServices는 서비스 포함 시 필요한 import 목록을 반환합니다.
func (g *CodeGenerator) getRequiredImportsWithServices() []string {
	imports := g.getRequiredImports()

	// 서비스 관련 import 추가
	serviceImports := []string{
		"context",
		"strings",
		"github.com/pocketbase/pocketbase",
	}

	// 제네릭 사용 시 추가 import
	if g.templateData.UseGeneric {
		serviceImports = append(serviceImports, "fmt")
	}

	// 모든 import 합치기
	allImports := append(imports, serviceImports...)

	// 중복 제거
	uniqueImports := make(map[string]bool)
	var result []string
	for _, imp := range allImports {
		if !uniqueImports[imp] {
			uniqueImports[imp] = true
			result = append(result, imp)
		}
	}

	return result
}

// SetSchemaVersion은 스키마 버전을 설정합니다.
func (g *CodeGenerator) SetSchemaVersion(version SchemaVersion) {
	g.schemaVersion = version
}

// GetSchemaVersion은 현재 스키마 버전을 반환합니다.
func (g *CodeGenerator) GetSchemaVersion() SchemaVersion {
	return g.schemaVersion
}
