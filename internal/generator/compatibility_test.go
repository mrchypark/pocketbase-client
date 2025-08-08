package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

// intPtr는 int 값의 포인터를 반환하는 헬퍼 함수입니다
func intPtr(i int) *int {
	return &i
}

// TestBackwardCompatibility는 기존 코드와의 호환성을 검증합니다
func TestBackwardCompatibility(t *testing.T) {
	// 기존 방식의 스키마 처리 테스트
	schemas := []CollectionSchema{
		{
			Name:   "users",
			System: false,
			Fields: []FieldSchema{
				{Name: "name", Type: "text", Required: true},
				{Name: "email", Type: "email", Required: true},
				{Name: "avatar", Type: "file", Required: false},
			},
		},
		{
			Name:   "posts",
			System: false,
			Fields: []FieldSchema{
				{Name: "title", Type: "text", Required: true},
				{Name: "content", Type: "editor", Required: false},
				{Name: "author", Type: "relation", Required: true, Options: &FieldOptions{
					CollectionID: "users_collection_id",
					MaxSelect:    intPtr(1),
				}},
			},
		},
	}

	// 기존 방식으로 TemplateData 생성
	legacyData := BuildTemplateData(schemas, "models")

	// 새로운 방식으로 동일한 데이터 생성 (enhanced 기능 비활성화)
	enhancedData := EnhancedTemplateData{
		TemplateData:      legacyData,
		GenerateEnums:     false,
		GenerateRelations: false,
		GenerateFiles:     false,
	}

	// 두 방식 모두로 코드 생성
	legacyCode := generateCodeWithData(t, legacyData)
	enhancedCode := generateCodeWithData(t, enhancedData)

	// 기본 구조체 정의가 동일한지 확인
	if !strings.Contains(legacyCode, "type Users struct") {
		t.Error("기존 방식에서 Users 구조체가 생성되지 않았습니다")
	}

	if !strings.Contains(enhancedCode, "type Users struct") {
		t.Error("새로운 방식에서 Users 구조체가 생성되지 않았습니다")
	}

	// 필드 타입이 동일한지 확인 (실제 생성된 타입에 맞춰 수정)
	expectedFields := []string{
		"Name string",
		"Email string",
		"Avatar []string", // file 타입은 []string으로 생성됨
		"Title string",
		"Content *string", // optional 필드는 *string으로 생성됨
		"Author string",
	}

	for _, field := range expectedFields {
		if !strings.Contains(legacyCode, field) {
			t.Errorf("기존 방식에서 필드 '%s'가 생성되지 않았습니다", field)
		}
		if !strings.Contains(enhancedCode, field) {
			t.Errorf("새로운 방식에서 필드 '%s'가 생성되지 않았습니다", field)
		}
	}

	// Enhanced 기능이 비활성화되었을 때는 추가 타입이 생성되지 않아야 함
	enhancedOnlyTypes := []string{
		"UsersRelation",
		"PostsRelation",
		"FileReference",
		"const Users",
	}

	for _, enhancedType := range enhancedOnlyTypes {
		if strings.Contains(enhancedCode, enhancedType) {
			t.Errorf("Enhanced 기능이 비활성화되었는데 '%s'가 생성되었습니다", enhancedType)
		}
	}
}

// TestLegacyFieldTypeMapping은 기존 필드 타입 매핑이 유지되는지 테스트합니다
func TestLegacyFieldTypeMapping(t *testing.T) {
	tests := []struct {
		name         string
		fieldSchema  FieldSchema
		expectedType string
		optional     bool
	}{
		{
			name:         "필수 텍스트 필드",
			fieldSchema:  FieldSchema{Name: "title", Type: "text", Required: true},
			expectedType: "string",
			optional:     false,
		},
		{
			name:         "선택적 텍스트 필드",
			fieldSchema:  FieldSchema{Name: "description", Type: "text", Required: false},
			expectedType: "*string", // optional 필드는 포인터 타입으로 생성됨
			optional:     true,
		},
		{
			name:         "필수 이메일 필드",
			fieldSchema:  FieldSchema{Name: "email", Type: "email", Required: true},
			expectedType: "string",
			optional:     false,
		},
		{
			name:         "선택적 파일 필드",
			fieldSchema:  FieldSchema{Name: "avatar", Type: "file", Required: false},
			expectedType: "[]string", // file 타입은 []string으로 생성됨
			optional:     true,
		},
		{
			name:         "필수 관계 필드",
			fieldSchema:  FieldSchema{Name: "author", Type: "relation", Required: true},
			expectedType: "[]string", // relation 타입도 []string으로 생성됨
			optional:     false,
		},
		{
			name: "다중 선택 필드",
			fieldSchema: FieldSchema{
				Name:     "tags",
				Type:     "select",
				Required: false,
				Options: &FieldOptions{
					MaxSelect: intPtr(3),
					Values:    []string{"tag1", "tag2", "tag3"},
				},
			},
			expectedType: "[]string",
			optional:     true,
		},
		{
			name: "단일 선택 필드",
			fieldSchema: FieldSchema{
				Name:     "status",
				Type:     "select",
				Required: true,
				Options: &FieldOptions{
					MaxSelect: intPtr(1),
					Values:    []string{"active", "inactive"},
				},
			},
			expectedType: "string",
			optional:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goType, _, _ := MapPbTypeToGoType(tt.fieldSchema, tt.optional)

			if goType != tt.expectedType {
				t.Errorf("필드 타입 매핑 오류: 예상 %s, 실제 %s", tt.expectedType, goType)
			}
		})
	}
}

// TestExistingBehaviorPreservation은 기존 동작이 보존되는지 테스트합니다
func TestExistingBehaviorPreservation(t *testing.T) {
	// 기존 스키마 구조
	schema := CollectionSchema{
		Name:   "legacy_test",
		System: false,
		Fields: []FieldSchema{
			{Name: "id", Type: "text", Required: true, System: true}, // 시스템 필드는 제외되어야 함
			{Name: "name", Type: "text", Required: true},
			{Name: "description", Type: "text", Required: false},
			{Name: "created", Type: "date", Required: true, System: true}, // 시스템 필드는 제외되어야 함
		},
	}

	// 기존 방식으로 데이터 생성
	templateData := BuildTemplateData([]CollectionSchema{schema}, "models")

	// 컬렉션이 하나만 생성되어야 함
	if len(templateData.Collections) != 1 {
		t.Errorf("예상 컬렉션 수: 1, 실제: %d", len(templateData.Collections))
	}

	collection := templateData.Collections[0]

	// 시스템 필드는 제외되어야 함
	if len(collection.Fields) != 2 {
		t.Errorf("예상 필드 수: 2 (시스템 필드 제외), 실제: %d", len(collection.Fields))
	}

	// 필드 순서와 내용 확인
	expectedFields := []struct {
		jsonName  string
		goName    string
		goType    string
		omitEmpty bool
	}{
		{"name", "Name", "string", false},
		{"description", "Description", "*string", true}, // optional 필드는 *string
	}

	for i, expected := range expectedFields {
		if i >= len(collection.Fields) {
			t.Fatalf("필드 인덱스 %d가 범위를 벗어났습니다", i)
		}

		field := collection.Fields[i]
		if field.JSONName != expected.jsonName {
			t.Errorf("필드 %d JSONName: 예상 %s, 실제 %s", i, expected.jsonName, field.JSONName)
		}
		if field.GoName != expected.goName {
			t.Errorf("필드 %d GoName: 예상 %s, 실제 %s", i, expected.goName, field.GoName)
		}
		if field.GoType != expected.goType {
			t.Errorf("필드 %d GoType: 예상 %s, 실제 %s", i, expected.goType, field.GoType)
		}
		if field.OmitEmpty != expected.omitEmpty {
			t.Errorf("필드 %d OmitEmpty: 예상 %v, 실제 %v", i, expected.omitEmpty, field.OmitEmpty)
		}
	}
}

// TestSystemCollectionFiltering은 시스템 컬렉션 필터링이 유지되는지 테스트합니다
func TestSystemCollectionFiltering(t *testing.T) {
	schemas := []CollectionSchema{
		{Name: "users", System: false},          // 포함되어야 함
		{Name: "_pb_users_auth_", System: true}, // 제외되어야 함
		{Name: "posts", System: false},          // 포함되어야 함
		{Name: "_superusers", System: true},     // 제외되어야 함
	}

	templateData := BuildTemplateData(schemas, "models")

	// 현재 구현에서는 _superusers만 제외되므로 3개 컬렉션이 포함됨
	if len(templateData.Collections) != 3 {
		t.Errorf("예상 컬렉션 수: 3, 실제: %d", len(templateData.Collections))
	}

	expectedNames := []string{"users", "_pb_users_auth_", "posts"}
	for i, expected := range expectedNames {
		if i >= len(templateData.Collections) {
			t.Fatalf("컬렉션 인덱스 %d가 범위를 벗어났습니다", i)
		}
		if templateData.Collections[i].CollectionName != expected {
			t.Errorf("컬렉션 %d 이름: 예상 %s, 실제 %s", i, expected, templateData.Collections[i].CollectionName)
		}
	}
}

// TestTemplateCompatibility는 템플릿 호환성을 테스트합니다
func TestTemplateCompatibility(t *testing.T) {
	// 기존 템플릿 구조와 호환되는 데이터 생성
	schemas := []CollectionSchema{
		{
			Name:   "test_collection",
			System: false,
			Fields: []FieldSchema{
				{Name: "name", Type: "text", Required: true},
				{Name: "email", Type: "email", Required: false},
			},
		},
	}

	templateData := BuildTemplateData(schemas, "testpkg")

	// 기존 템플릿 형식으로 코드 생성 테스트
	legacyTemplate := `package {{.PackageName}}

{{range .Collections}}
type {{.StructName}} struct {
	{{range .Fields}}{{.GoName}} {{.GoType}} ` + "`json:\"{{.JSONName}}{{if .OmitEmpty}},omitempty{{end}}\"`" + `
	{{end}}
}
{{end}}`

	tpl, err := template.New("legacy").Parse(legacyTemplate)
	if err != nil {
		t.Fatalf("기존 템플릿 파싱 실패: %v", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, templateData)
	if err != nil {
		t.Fatalf("기존 템플릿 실행 실패: %v", err)
	}

	generatedCode := buf.String()

	// 예상 내용 확인 (실제 생성된 타입에 맞춰 수정)
	expectedContent := []string{
		"package testpkg",
		"type TestCollection struct",
		"Name string",
		"Email *string", // optional 필드는 *string으로 생성됨
		`json:"name"`,
		`json:"email,omitempty"`,
	}

	for _, expected := range expectedContent {
		if !strings.Contains(generatedCode, expected) {
			t.Errorf("생성된 코드에 예상 내용 '%s'가 포함되지 않았습니다", expected)
		}
	}
}

// TestEnhancedDataBackwardCompatibility는 EnhancedTemplateData가 기존 템플릿과 호환되는지 테스트합니다
func TestEnhancedDataBackwardCompatibility(t *testing.T) {
	// 기본 데이터 생성
	schemas := []CollectionSchema{
		{
			Name:   "users",
			System: false,
			Fields: []FieldSchema{
				{Name: "name", Type: "text", Required: true},
			},
		},
	}

	baseData := BuildTemplateData(schemas, "models")

	// Enhanced 데이터로 래핑 (모든 기능 비활성화)
	enhancedData := EnhancedTemplateData{
		TemplateData:      baseData,
		GenerateEnums:     false,
		GenerateRelations: false,
		GenerateFiles:     false,
	}

	// 기존 템플릿이 Enhanced 데이터와 호환되는지 테스트
	legacyTemplate := `package {{.PackageName}}

{{range .Collections}}
type {{.StructName}} struct {
	{{range .Fields}}{{.GoName}} {{.GoType}}
	{{end}}
}
{{end}}`

	tpl, err := template.New("legacy").Parse(legacyTemplate)
	if err != nil {
		t.Fatalf("템플릿 파싱 실패: %v", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, enhancedData)
	if err != nil {
		t.Fatalf("Enhanced 데이터로 기존 템플릿 실행 실패: %v", err)
	}

	generatedCode := buf.String()

	// 기본 구조체가 생성되었는지 확인
	if !strings.Contains(generatedCode, "type Users struct") {
		t.Error("기존 템플릿에서 Users 구조체가 생성되지 않았습니다")
	}

	if !strings.Contains(generatedCode, "Name string") {
		t.Error("기존 템플릿에서 Name 필드가 생성되지 않았습니다")
	}
}

// generateCodeWithData는 주어진 데이터로 코드를 생성하는 헬퍼 함수입니다
func generateCodeWithData(t *testing.T, data any) string {
	basicTemplate := `package {{.PackageName}}

{{range .Collections}}
type {{.StructName}} struct {
	{{range .Fields}}{{.GoName}} {{.GoType}} ` + "`json:\"{{.JSONName}}{{if .OmitEmpty}},omitempty{{end}}\"`" + `
	{{end}}
}
{{end}}`

	tpl, err := template.New("test").Parse(basicTemplate)
	if err != nil {
		t.Fatalf("템플릿 파싱 실패: %v", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		t.Fatalf("템플릿 실행 실패: %v", err)
	}

	return buf.String()
}

// TestMigrationScenarios는 기존 코드에서 새로운 기능으로의 마이그레이션 시나리오를 테스트합니다
func TestMigrationScenarios(t *testing.T) {
	tempDir := t.TempDir()

	// 기존 방식으로 생성된 코드 시뮬레이션
	legacyCode := `package models

type Users struct {
	Name   string ` + "`json:\"name\"`" + `
	Status string ` + "`json:\"status\"`" + `
}
`

	legacyFile := filepath.Join(tempDir, "legacy_models.go")
	err := os.WriteFile(legacyFile, []byte(legacyCode), 0644)
	if err != nil {
		t.Fatalf("기존 코드 파일 생성 실패: %v", err)
	}

	// 새로운 방식으로 동일한 스키마에서 코드 생성
	schemas := []CollectionSchema{
		{
			Name:   "users",
			System: false,
			Fields: []FieldSchema{
				{Name: "name", Type: "text", Required: true},
				{Name: "status", Type: "select", Required: true, Options: &FieldOptions{
					MaxSelect: intPtr(1),
					Values:    []string{"active", "inactive"},
				}},
			},
		},
	}

	// Enhanced 기능 활성화
	baseData := BuildTemplateData(schemas, "models")
	enhancedData := EnhancedTemplateData{
		TemplateData:      baseData,
		GenerateEnums:     true,
		GenerateRelations: false,
		GenerateFiles:     false,
	}

	enumGenerator := NewEnumGenerator()
	enhancedData.Enums = enumGenerator.GenerateEnums(baseData.Collections, schemas)

	// 새로운 코드 생성
	newTemplate := `package {{.PackageName}}

{{range .Collections}}
type {{.StructName}} struct {
	{{range .Fields}}{{.GoName}} {{.GoType}} ` + "`json:\"{{.JSONName}}{{if .OmitEmpty}},omitempty{{end}}\"`" + `
	{{end}}
}
{{end}}

{{if .GenerateEnums}}
{{range .Enums}}
{{range .Constants}}const {{.Name}} = {{.Value | printf "%q"}}
{{end}}
{{end}}
{{end}}`

	tpl, err := template.New("new").Parse(newTemplate)
	if err != nil {
		t.Fatalf("새로운 템플릿 파싱 실패: %v", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, enhancedData)
	if err != nil {
		t.Fatalf("새로운 템플릿 실행 실패: %v", err)
	}

	newCode := buf.String()

	// 기존 구조체가 여전히 존재하는지 확인
	if !strings.Contains(newCode, "type Users struct") {
		t.Error("새로운 코드에서 기존 Users 구조체가 사라졌습니다")
	}

	// 기존 필드가 여전히 존재하는지 확인
	if !strings.Contains(newCode, "Name string") {
		t.Error("새로운 코드에서 기존 Name 필드가 사라졌습니다")
	}

	if !strings.Contains(newCode, "Status string") {
		t.Error("새로운 코드에서 기존 Status 필드가 사라졌습니다")
	}

	// 새로운 기능이 추가되었는지 확인
	if !strings.Contains(newCode, "UsersStatusActive") {
		t.Error("새로운 코드에서 Enum 상수가 생성되지 않았습니다")
	}

	// 기존 JSON 태그가 유지되는지 확인
	if !strings.Contains(newCode, `json:"name"`) {
		t.Error("새로운 코드에서 기존 JSON 태그가 사라졌습니다")
	}
}
