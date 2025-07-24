package generator

import (
	"strings"
	"testing"
)

func TestGenericServiceGeneration(t *testing.T) {
	// 테스트용 스키마 데이터
	schemas := []CollectionSchema{
		{
			Name:   "posts",
			System: false,
			Fields: []FieldSchema{
				{Name: "title", Type: "text", Required: true},
				{Name: "content", Type: "editor", Required: false},
				{Name: "published", Type: "bool", Required: false},
			},
		},
		{
			Name:   "users",
			System: false,
			Fields: []FieldSchema{
				{Name: "name", Type: "text", Required: true},
				{Name: "email", Type: "email", Required: true},
				{Name: "age", Type: "number", Required: false},
			},
		},
	}

	tests := []struct {
		name         string
		useGeneric   bool
		wantContains []string
	}{
		{
			name:       "제네릭 서비스 생성",
			useGeneric: true,
			wantContains: []string{
				"GenericRecordService[T any]",
				"BaseGenericService[T any]",
				"PostsService struct",
				"BaseGenericService[Posts]",
				"UsersService struct",
				"BaseGenericService[Users]",
			},
		},
		{
			name:       "기존 방식 서비스 생성",
			useGeneric: false,
			wantContains: []string{
				"PostsService struct",
				"UsersService struct",
				"GetString", // 기존 getter 메서드는 주석으로만 표시
				"GetBool",
				"GetFloat",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TemplateData 생성
			templateData := BuildTemplateData(schemas, "models", SchemaVersionLatest, tt.useGeneric)

			// CodeGenerator 생성
			generator := NewCodeGenerator(SchemaVersionLatest, &templateData)

			// 제네릭 서비스 코드 생성
			serviceCode, err := generator.GenerateGenericServices()
			if err != nil {
				t.Fatalf("제네릭 서비스 생성 실패: %v", err)
			}

			// 생성된 코드 검증
			for _, want := range tt.wantContains {
				if !strings.Contains(serviceCode, want) {
					t.Errorf("생성된 코드에 '%s'가 포함되지 않음", want)
				}
			}

			// 기본적인 구조 검증
			if !strings.Contains(serviceCode, "package models") {
				t.Error("패키지 선언이 없음")
			}

			if !strings.Contains(serviceCode, "import (") {
				t.Error("import 구문이 없음")
			}

			// 각 컬렉션에 대한 서비스가 생성되었는지 확인
			for _, schema := range schemas {
				serviceName := ToPascalCase(schema.Name) + "Service"
				if !strings.Contains(serviceCode, serviceName) {
					t.Errorf("%s 서비스가 생성되지 않음", serviceName)
				}
			}

			t.Logf("생성된 서비스 코드 (처음 500자):\n%s", serviceCode[:min(500, len(serviceCode))])
		})
	}
}

func TestStructGeneration(t *testing.T) {
	// 테스트용 스키마 데이터
	schemas := []CollectionSchema{
		{
			Name:   "posts",
			System: false,
			Fields: []FieldSchema{
				{Name: "title", Type: "text", Required: true},
				{Name: "content", Type: "editor", Required: false},
				{Name: "published", Type: "bool", Required: false},
			},
		},
	}

	// TemplateData 생성
	templateData := BuildTemplateData(schemas, "models", SchemaVersionLatest, false)

	// CodeGenerator 생성
	generator := NewCodeGenerator(SchemaVersionLatest, &templateData)

	// 구조체 코드 생성
	structCode, err := generator.GenerateStructs()
	if err != nil {
		t.Fatalf("구조체 생성 실패: %v", err)
	}

	// 생성된 코드 검증
	wantContains := []string{
		"package models",
		"type BaseModel struct",
		"type Posts struct",
		"BaseModel",
		"Title string",
		"Content *string",
		"Published *bool",
		"TableName() string",
		"ContentValueOr(defaultValue string) string",
		"PublishedValueOr(defaultValue bool) bool",
	}

	for _, want := range wantContains {
		if !strings.Contains(structCode, want) {
			t.Errorf("생성된 코드에 '%s'가 포함되지 않음", want)
		}
	}

	t.Logf("생성된 구조체 코드 (처음 800자):\n%s", structCode[:min(800, len(structCode))])
}

func TestGenericFieldProcessor(t *testing.T) {
	// GenericFieldProcessor 테스트
	processor := NewGenericFieldProcessor(SchemaVersionLatest)

	fields := []FieldSchema{
		{Name: "title", Type: "text", Required: true},
		{Name: "content", Type: "editor", Required: false},
		{Name: "user_id", Type: "relation", Required: true, System: true}, // 시스템 필드는 제외되어야 함
	}

	processedFields, err := processor.ProcessFields(fields, "posts")
	if err != nil {
		t.Fatalf("필드 처리 실패: %v", err)
	}

	// 시스템 필드가 제외되었는지 확인
	if len(processedFields) != 2 {
		t.Errorf("예상 필드 수: 2, 실제: %d", len(processedFields))
	}

	// 제네릭 getter 메서드가 설정되었는지 확인
	for _, field := range processedFields {
		if !strings.Contains(field.GetterMethod, "Get[") {
			t.Errorf("필드 %s에 제네릭 getter 메서드가 설정되지 않음: %s", field.JSONName, field.GetterMethod)
		}
	}
}

func TestMapPbTypeToGoTypeWithGeneric(t *testing.T) {
	tests := []struct {
		name       string
		field      FieldSchema
		omitEmpty  bool
		useGeneric bool
		wantType   string
		wantGetter string
	}{
		{
			name:       "텍스트 필드 - 제네릭",
			field:      FieldSchema{Type: "text", Required: true},
			omitEmpty:  false,
			useGeneric: true,
			wantType:   "string",
			wantGetter: "Get[string]",
		},
		{
			name:       "텍스트 필드 - 기존 방식",
			field:      FieldSchema{Type: "text", Required: true},
			omitEmpty:  false,
			useGeneric: false,
			wantType:   "string",
			wantGetter: "GetString",
		},
		{
			name:       "선택적 텍스트 필드 - 제네릭",
			field:      FieldSchema{Type: "text", Required: false},
			omitEmpty:  true,
			useGeneric: true,
			wantType:   "string", // 제네릭에서는 포인터 타입 적용 안함
			wantGetter: "Get[string]",
		},
		{
			name:       "선택적 텍스트 필드 - 기존 방식",
			field:      FieldSchema{Type: "text", Required: false},
			omitEmpty:  true,
			useGeneric: false,
			wantType:   "*string",
			wantGetter: "GetStringPointer",
		},
		{
			name:       "불린 필드 - 제네릭",
			field:      FieldSchema{Type: "bool", Required: true},
			omitEmpty:  false,
			useGeneric: true,
			wantType:   "bool",
			wantGetter: "Get[bool]",
		},
		{
			name:       "숫자 필드 - 제네릭",
			field:      FieldSchema{Type: "number", Required: true},
			omitEmpty:  false,
			useGeneric: true,
			wantType:   "float64",
			wantGetter: "Get[float64]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goType, _, getterMethod := MapPbTypeToGoTypeWithGeneric(tt.field, tt.omitEmpty, tt.useGeneric)

			if goType != tt.wantType {
				t.Errorf("GoType = %v, want %v", goType, tt.wantType)
			}

			if getterMethod != tt.wantGetter {
				t.Errorf("GetterMethod = %v, want %v", getterMethod, tt.wantGetter)
			}
		})
	}
}

// min 함수 (Go 1.21 이전 버전 호환성)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
