package generator

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

// TestEndToEndCodeGeneration은 전체 코드 생성 파이프라인을 테스트합니다
func TestEndToEndCodeGeneration(t *testing.T) {
	tests := []struct {
		name           string
		schemaFile     string
		generateEnums  bool
		generateRels   bool
		generateFiles  bool
		expectedTypes  []string
		expectedConsts []string
	}{
		{
			name:           "모든 기능 활성화",
			schemaFile:     "testdata/complex_schema.json",
			generateEnums:  true,
			generateRels:   true,
			generateFiles:  true,
			expectedTypes:  []string{"Users", "Posts", "Comments", "UsersRelation", "PostsRelation", "FileReference"},
			expectedConsts: []string{"UsersRoleAdmin", "PostsTagsTech", "PostsStatusDraft"},
		},
		{
			name:           "Enum만 활성화",
			schemaFile:     "testdata/complex_schema.json",
			generateEnums:  true,
			generateRels:   false,
			generateFiles:  false,
			expectedTypes:  []string{"Users", "Posts", "Comments"},
			expectedConsts: []string{"UsersRoleAdmin", "PostsTagsTech", "PostsStatusDraft"},
		},
		{
			name:           "기본 기능만",
			schemaFile:     "testdata/complex_schema.json",
			generateEnums:  false,
			generateRels:   false,
			generateFiles:  false,
			expectedTypes:  []string{"Users", "Posts", "Comments"},
			expectedConsts: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 임시 출력 파일 생성
			tempDir := t.TempDir()
			outputFile := filepath.Join(tempDir, "models.gen.go")

			// 스키마 로드
			schemas, err := LoadSchema(tt.schemaFile)
			if err != nil {
				t.Fatalf("스키마 로드 실패: %v", err)
			}

			// 기본 TemplateData 생성
			baseTplData := TemplateData{
				PackageName: "models",
				JsonLibrary: "encoding/json",
				Collections: make([]CollectionData, 0, len(schemas)),
			}

			for _, s := range schemas {
				collectionData := CollectionData{
					CollectionName: s.Name,
					StructName:     ToPascalCase(s.Name),
					Fields:         make([]FieldData, 0, len(s.Fields)),
				}

				for _, field := range s.Fields {
					if field.System {
						continue
					}
					goType, _, getter := MapPbTypeToGoType(field, !field.Required)
					collectionData.Fields = append(collectionData.Fields, FieldData{
						JsonName:     field.Name,
						GoName:       ToPascalCase(field.Name),
						GoType:       goType,
						OmitEmpty:    !field.Required,
						GetterMethod: getter,
					})
				}
				baseTplData.Collections = append(baseTplData.Collections, collectionData)
			}

			// Enhanced 기능 처리
			var tplData any
			if tt.generateEnums || tt.generateRels || tt.generateFiles {
				enhancedData := EnhancedTemplateData{
					TemplateData:      baseTplData,
					GenerateEnums:     tt.generateEnums,
					GenerateRelations: tt.generateRels,
					GenerateFiles:     tt.generateFiles,
				}

				if tt.generateEnums {
					enumGenerator := NewEnumGenerator()
					enhancedData.Enums = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)
				}

				if tt.generateRels {
					relationGenerator := NewRelationGenerator()
					enhancedData.RelationTypes = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)
				}

				if tt.generateFiles {
					fileGenerator := NewFileGenerator()
					enhancedData.FileTypes = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)
				}

				tplData = enhancedData
			} else {
				tplData = baseTplData
			}

			// 템플릿 실행
			templateContent := getTestTemplate()
			tpl, err := template.New("models").Parse(templateContent)
			if err != nil {
				t.Fatalf("템플릿 파싱 실패: %v", err)
			}

			var buf bytes.Buffer
			err = tpl.Execute(&buf, tplData)
			if err != nil {
				t.Fatalf("템플릿 실행 실패: %v", err)
			}

			// 생성된 코드를 파일에 저장
			err = os.WriteFile(outputFile, buf.Bytes(), 0644)
			if err != nil {
				t.Fatalf("파일 저장 실패: %v", err)
			}

			// 생성된 코드 검증
			generatedCode := buf.String()

			// 예상 타입들이 생성되었는지 확인
			for _, expectedType := range tt.expectedTypes {
				if !strings.Contains(generatedCode, fmt.Sprintf("type %s struct", expectedType)) {
					t.Errorf("예상 타입 %s가 생성되지 않았습니다", expectedType)
				}
			}

			// 예상 상수들이 생성되었는지 확인
			for _, expectedConst := range tt.expectedConsts {
				if !strings.Contains(generatedCode, expectedConst) {
					t.Errorf("예상 상수 %s가 생성되지 않았습니다", expectedConst)
				}
			}

			// Go 문법 검증
			err = validateGoSyntax(outputFile)
			if err != nil {
				t.Errorf("생성된 코드의 Go 문법 오류: %v", err)
			}

			// 컴파일 가능성 검증
			err = validateCompilation(outputFile)
			if err != nil {
				t.Errorf("생성된 코드 컴파일 실패: %v", err)
			}
		})
	}
}

// TestGeneratedCodeUsability는 생성된 코드의 실제 사용 가능성을 검증합니다
func TestGeneratedCodeUsability(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "models.gen.go")

	// 복잡한 스키마로 코드 생성
	schemas, err := LoadSchema("testdata/complex_schema.json")
	if err != nil {
		t.Fatalf("스키마 로드 실패: %v", err)
	}

	// 모든 기능을 활성화하여 코드 생성
	baseTplData := TemplateData{
		PackageName: "models",
		JsonLibrary: "encoding/json",
		Collections: make([]CollectionData, 0, len(schemas)),
	}

	for _, s := range schemas {
		collectionData := CollectionData{
			CollectionName: s.Name,
			StructName:     ToPascalCase(s.Name),
			Fields:         make([]FieldData, 0, len(s.Fields)),
		}

		for _, field := range s.Fields {
			if field.System {
				continue
			}
			goType, _, getter := MapPbTypeToGoType(field, !field.Required)
			collectionData.Fields = append(collectionData.Fields, FieldData{
				JsonName:     field.Name,
				GoName:       ToPascalCase(field.Name),
				GoType:       goType,
				OmitEmpty:    !field.Required,
				GetterMethod: getter,
			})
		}
		baseTplData.Collections = append(baseTplData.Collections, collectionData)
	}

	enhancedData := EnhancedTemplateData{
		TemplateData:      baseTplData,
		GenerateEnums:     true,
		GenerateRelations: true,
		GenerateFiles:     true,
	}

	enumGenerator := NewEnumGenerator()
	enhancedData.Enums = enumGenerator.GenerateEnums(baseTplData.Collections, schemas)

	relationGenerator := NewRelationGenerator()
	enhancedData.RelationTypes = relationGenerator.GenerateRelationTypes(baseTplData.Collections, schemas)

	fileGenerator := NewFileGenerator()
	enhancedData.FileTypes = fileGenerator.GenerateFileTypes(baseTplData.Collections, schemas)

	// 템플릿 실행
	templateContent := getTestTemplate()
	tpl, err := template.New("models").Parse(templateContent)
	if err != nil {
		t.Fatalf("템플릿 파싱 실패: %v", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, enhancedData)
	if err != nil {
		t.Fatalf("템플릿 실행 실패: %v", err)
	}

	err = os.WriteFile(outputFile, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("파일 저장 실패: %v", err)
	}

	// 사용성 테스트 코드 생성
	usageTestFile := filepath.Join(tempDir, "usage_test.go")
	usageTestCode := `package models

import (
	"context"
	"testing"
)

func TestGeneratedCodeUsage(t *testing.T) {
	// Enum 상수 사용 테스트
	role := UsersRoleAdmin
	if role != "admin" {
		t.Errorf("Expected admin, got %s", role)
	}

	// 구조체 생성 및 필드 설정 테스트
	user := Users{
		Name: "Test User",
		Role: UsersRoleAdmin,
	}

	if user.Name != "Test User" {
		t.Errorf("Expected Test User, got %s", user.Name)
	}

	// FileReference 사용 테스트
	fileRef := NewFileReference("test.jpg", "record123", "users", "avatar")
	if fileRef.Filename() != "test.jpg" {
		t.Errorf("Expected test.jpg, got %s", fileRef.Filename())
	}

	url := fileRef.URL("http://localhost:8090")
	expectedURL := "http://localhost:8090/api/files/users/record123/test.jpg"
	if url != expectedURL {
		t.Errorf("Expected %s, got %s", expectedURL, url)
	}

	// Relation 타입 사용 테스트
	userRel := NewUsersRelation("user123")
	if userRel.ID() != "user123" {
		t.Errorf("Expected user123, got %s", userRel.ID())
	}

	if userRel.IsEmpty() {
		t.Error("Expected relation to not be empty")
	}
}
`

	err = os.WriteFile(usageTestFile, []byte(usageTestCode), 0644)
	if err != nil {
		t.Fatalf("사용성 테스트 파일 생성 실패: %v", err)
	}

	// 사용성 테스트 실행
	cmd := exec.Command("go", "test", "-v", usageTestFile, outputFile)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("사용성 테스트 실행 실패: %v\n출력: %s", err, string(output))
	}
}

// validateGoSyntax는 생성된 Go 코드의 문법을 검증합니다
func validateGoSyntax(filename string) error {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	return err
}

// validateCompilation은 생성된 코드가 컴파일 가능한지 검증합니다
func validateCompilation(filename string) error {
	tempDir := filepath.Dir(filename)

	// go.mod 파일 생성
	goModContent := `module testmodule

go 1.21

require (
	github.com/mrchypark/pocketbase-client v0.0.0
)

replace github.com/mrchypark/pocketbase-client => ../../../..
`

	goModFile := filepath.Join(tempDir, "go.mod")
	err := os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		return fmt.Errorf("go.mod 파일 생성 실패: %v", err)
	}

	// go build 실행
	cmd := exec.Command("go", "build", "-o", "/dev/null", filename)
	cmd.Dir = tempDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("컴파일 실패: %v\n출력: %s", err, string(output))
	}

	return nil
}

// getTestTemplate은 테스트용 템플릿을 반환합니다
func getTestTemplate() string {
	return `package {{.PackageName}}

import (
	"context"
	"fmt"
	{{.JsonLibrary | printf "%q"}}
)

// BaseModel은 모든 PocketBase 레코드의 기본 필드를 포함합니다
type BaseModel struct {
	Id      string ` + "`json:\"id\"`" + `
	Created string ` + "`json:\"created\"`" + `
	Updated string ` + "`json:\"updated\"`" + `
}

{{range .Collections}}
// {{.StructName}}은 {{.CollectionName}} 컬렉션의 구조체입니다
type {{.StructName}} struct {
	BaseModel
	{{range .Fields}}{{.GoName}} {{.GoType}} ` + "`json:\"{{.JsonName}}{{if .OmitEmpty}},omitempty{{end}}\"`" + `
	{{end}}
}
{{end}}

{{if .GenerateEnums}}
{{range .Enums}}
// {{.CollectionName}} {{.FieldName}} enum constants
{{range .Constants}}const {{.Name}} = {{.Value | printf "%q"}}
{{end}}

// {{.CollectionName}}{{.FieldName}}Values returns all possible values
func {{.CollectionName}}{{.FieldName}}Values() []string {
	return []string{ {{range $i, $c := .Constants}}{{if $i}}, {{end}}{{$c.Name}}{{end}} }
}
{{end}}
{{end}}

{{if .GenerateRelations}}
{{range .RelationTypes}}
// {{.TypeName}}은 {{.TargetCollection}} 컬렉션에 대한 관계를 나타냅니다
type {{.TypeName}} struct {
	id string
}

// ID returns the relation ID
func (r {{.TypeName}}) ID() string {
	return r.id
}

// IsEmpty returns true if the relation is empty
func (r {{.TypeName}}) IsEmpty() bool {
	return r.id == ""
}

// New{{.TypeName}} creates a new {{.TypeName}}
func New{{.TypeName}}(id string) {{.TypeName}} {
	return {{.TypeName}}{id: id}
}
{{end}}
{{end}}

{{if .GenerateFiles}}
// FileReference represents a file reference
type FileReference struct {
	filename   string
	recordID   string
	collection string
	fieldName  string
}

// Filename returns the filename
func (f FileReference) Filename() string {
	return f.filename
}

// URL generates the file URL
func (f FileReference) URL(baseURL string) string {
	if f.filename == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/files/%s/%s/%s", baseURL, f.collection, f.recordID, f.filename)
}

// ThumbURL generates thumbnail URL
func (f FileReference) ThumbURL(baseURL, thumb string) string {
	if f.filename == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/files/%s/%s/%s?thumb=%s", baseURL, f.collection, f.recordID, f.filename, thumb)
}

// IsEmpty returns true if the file reference is empty
func (f FileReference) IsEmpty() bool {
	return f.filename == ""
}

// NewFileReference creates a new FileReference
func NewFileReference(filename, recordID, collection, fieldName string) FileReference {
	return FileReference{
		filename:   filename,
		recordID:   recordID,
		collection: collection,
		fieldName:  fieldName,
	}
}
{{end}}
`
}
