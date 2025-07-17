package generator

import (
	"bytes"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

func TestTemplateExecution(t *testing.T) {
	// 템플릿 파일 읽기
	templatePath := "../../cmd/pbc-gen/template.go.tpl"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}

	// 템플릿 파싱
	tmpl, err := template.New("test").Parse(string(templateContent))
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	tests := []struct {
		name string
		data EnhancedTemplateData
	}{
		{
			name: "basic template with all features",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "encoding/json",
					Collections: []CollectionData{
						{
							CollectionName: "users",
							StructName:     "User",
							Fields: []FieldData{
								{
									JSONName:     "name",
									GoName:       "Name",
									GoType:       "string",
									OmitEmpty:    false,
									GetterMethod: "GetString",
								},
								{
									JSONName:     "email",
									GoName:       "Email",
									GoType:       "*string",
									OmitEmpty:    true,
									GetterMethod: "GetStringPointer",
								},
								{
									JSONName:     "age",
									GoName:       "Age",
									GoType:       "*int",
									OmitEmpty:    true,
									GetterMethod: "GetIntPointer",
								},
							},
						},
						{
							CollectionName: "posts",
							StructName:     "Post",
							Fields: []FieldData{
								{
									JSONName:     "title",
									GoName:       "Title",
									GoType:       "string",
									OmitEmpty:    false,
									GetterMethod: "GetString",
								},
								{
									JSONName:     "content",
									GoName:       "Content",
									GoType:       "*string",
									OmitEmpty:    true,
									GetterMethod: "GetStringPointer",
								},
							},
						},
					},
				},
				Enums: []EnumData{
					{
						CollectionName: "users",
						FieldName:      "status",
						EnumTypeName:   "UserStatusType",
						Constants: []ConstantData{
							{Name: "UserStatusActive", Value: "active"},
							{Name: "UserStatusInactive", Value: "inactive"},
							{Name: "UserStatusPending", Value: "pending"},
						},
					},
				},
				RelationTypes: []RelationTypeData{
					{
						TypeName:         "AuthorRelation",
						TargetCollection: "users",
						TargetTypeName:   "User",
						IsMulti:          false,
					},
					{
						TypeName:         "CategoryRelation",
						TargetCollection: "categories",
						TargetTypeName:   "Category",
						IsMulti:          true,
					},
				},
				FileTypes: []FileTypeData{
					{
						TypeName:      "ImageFile",
						IsMulti:       false,
						HasThumbnails: true,
					},
				},
				GenerateEnums:     true,
				GenerateRelations: true,
				GenerateFiles:     true,
			},
		},
		{
			name: "template with enums only",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "encoding/json",
					Collections: []CollectionData{
						{
							CollectionName: "devices",
							StructName:     "Device",
							Fields: []FieldData{
								{
									JSONName:     "name",
									GoName:       "Name",
									GoType:       "string",
									OmitEmpty:    false,
									GetterMethod: "GetString",
								},
							},
						},
					},
				},
				Enums: []EnumData{
					{
						CollectionName: "devices",
						FieldName:      "type",
						EnumTypeName:   "DeviceTypeType",
						Constants: []ConstantData{
							{Name: "DeviceTypeSensor", Value: "sensor"},
							{Name: "DeviceTypeActuator", Value: "actuator"},
						},
					},
				},
				GenerateEnums:     true,
				GenerateRelations: false,
				GenerateFiles:     false,
			},
		},
		{
			name: "template with relations only",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "encoding/json",
					Collections: []CollectionData{
						{
							CollectionName: "posts",
							StructName:     "Post",
							Fields: []FieldData{
								{
									JSONName:     "title",
									GoName:       "Title",
									GoType:       "string",
									OmitEmpty:    false,
									GetterMethod: "GetString",
								},
							},
						},
					},
				},
				RelationTypes: []RelationTypeData{
					{
						TypeName:         "UserRelation",
						TargetCollection: "users",
						TargetTypeName:   "User",
						IsMulti:          false,
					},
				},
				GenerateEnums:     false,
				GenerateRelations: true,
				GenerateFiles:     false,
			},
		},
		{
			name: "template with files only",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "encoding/json",
					Collections: []CollectionData{
						{
							CollectionName: "gallery",
							StructName:     "Gallery",
							Fields: []FieldData{
								{
									JSONName:     "name",
									GoName:       "Name",
									GoType:       "string",
									OmitEmpty:    false,
									GetterMethod: "GetString",
								},
							},
						},
					},
				},
				FileTypes: []FileTypeData{
					{
						TypeName:       "ImageFile",
						IsMulti:        false,
						HasThumbnails:  true,
						ThumbnailSizes: []string{"100x100", "200x200"},
					},
				},
				GenerateEnums:     false,
				GenerateRelations: false,
				GenerateFiles:     true,
			},
		},
		{
			name: "minimal template without enhanced features",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "encoding/json",
					Collections: []CollectionData{
						{
							CollectionName: "simple",
							StructName:     "Simple",
							Fields: []FieldData{
								{
									JSONName:     "id",
									GoName:       "ID",
									GoType:       "string",
									OmitEmpty:    false,
									GetterMethod: "GetString",
								},
							},
						},
					},
				},
				GenerateEnums:     false,
				GenerateRelations: false,
				GenerateFiles:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 템플릿 실행
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, tt.data)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			generatedCode := buf.String()

			// 생성된 코드가 비어있지 않은지 확인
			if strings.TrimSpace(generatedCode) == "" {
				t.Fatal("Generated code is empty")
			}

			// 생성된 코드에 기본 구조가 포함되어 있는지 확인
			expectedParts := []string{
				"package " + tt.data.PackageName,
				"import (",
				tt.data.JSONLibrary,
				"github.com/mrchypark/pocketbase-client",
			}

			for _, part := range expectedParts {
				if !strings.Contains(generatedCode, part) {
					t.Errorf("Generated code missing expected part: %s", part)
				}
			}

			// 각 컬렉션에 대한 구조체가 생성되었는지 확인
			for _, collection := range tt.data.Collections {
				expectedStructParts := []string{
					"type " + collection.StructName + " struct",
					"func New" + collection.StructName + "()",
					"func To" + collection.StructName + "(",
					"func Get" + collection.StructName + "(",
					"func Get" + collection.StructName + "List(",
				}

				for _, part := range expectedStructParts {
					if !strings.Contains(generatedCode, part) {
						t.Errorf("Generated code missing expected struct part for %s: %s", collection.StructName, part)
					}
				}

				// 각 필드에 대한 getter/setter가 생성되었는지 확인
				for _, field := range collection.Fields {
					expectedFieldParts := []string{
						"func (m *" + collection.StructName + ") " + field.GoName + "()",
						"func (m *" + collection.StructName + ") Set" + field.GoName + "(",
					}

					for _, part := range expectedFieldParts {
						if !strings.Contains(generatedCode, part) {
							t.Errorf("Generated code missing expected field part for %s.%s: %s", collection.StructName, field.GoName, part)
						}
					}
				}
			}

			// Enum 생성 확인
			if tt.data.GenerateEnums {
				for _, enum := range tt.data.Enums {
					expectedEnumParts := []string{
						"const (",
						enum.EnumTypeName + "Values()",
						"IsValid" + enum.EnumTypeName + "(",
					}

					for _, part := range expectedEnumParts {
						if !strings.Contains(generatedCode, part) {
							t.Errorf("Generated code missing expected enum part for %s: %s", enum.EnumTypeName, part)
						}
					}

					// 각 상수가 생성되었는지 확인
					for _, constant := range enum.Constants {
						if !strings.Contains(generatedCode, constant.Name+" = \""+constant.Value+"\"") {
							t.Errorf("Generated code missing expected constant: %s = \"%s\"", constant.Name, constant.Value)
						}
					}
				}
			}

			// Relation 타입 생성 확인
			if tt.data.GenerateRelations {
				for _, relation := range tt.data.RelationTypes {
					expectedRelationParts := []string{
						"type " + relation.TypeName + " struct",
						"func (r " + relation.TypeName + ") ID()",
						"func (r " + relation.TypeName + ") Load(",
						"func (r " + relation.TypeName + ") IsEmpty()",
						"func New" + relation.TypeName + "(",
					}

					for _, part := range expectedRelationParts {
						if !strings.Contains(generatedCode, part) {
							t.Errorf("Generated code missing expected relation part for %s: %s", relation.TypeName, part)
						}
					}

					// 다중 관계 타입 확인
					if relation.IsMulti {
						multiTypeName := relation.TypeName + "s"
						expectedMultiParts := []string{
							"type " + multiTypeName + " []" + relation.TypeName,
							"func (r " + multiTypeName + ") IDs()",
							"func (r " + multiTypeName + ") LoadAll(",
						}

						for _, part := range expectedMultiParts {
							if !strings.Contains(generatedCode, part) {
								t.Errorf("Generated code missing expected multi-relation part for %s: %s", multiTypeName, part)
							}
						}
					}
				}
			}

			// File 타입 생성 확인
			if tt.data.GenerateFiles {
				expectedFileParts := []string{
					"type FileReference struct",
					"func (f FileReference) Filename()",
					"func (f FileReference) URL(",
					"func (f FileReference) ThumbURL(",
					"func (f FileReference) IsEmpty()",
					"func NewFileReference(",
					"type FileReferences []FileReference",
				}

				for _, part := range expectedFileParts {
					if !strings.Contains(generatedCode, part) {
						t.Errorf("Generated code missing expected file part: %s", part)
					}
				}
			}

			// Go 코드 포맷팅 테스트 (문법 검사)
			_, err = format.Source([]byte(generatedCode))
			if err != nil {
				t.Errorf("Generated code has syntax errors: %v\nGenerated code:\n%s", err, generatedCode)
			}
		})
	}
}

func TestTemplateCompilation(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir, err := os.MkdirTemp("", "pbc-gen-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 템플릿 파일 읽기
	templatePath := "../../cmd/pbc-gen/template.go.tpl"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}

	// 템플릿 파싱
	tmpl, err := template.New("test").Parse(string(templateContent))
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	// 테스트 데이터
	testData := EnhancedTemplateData{
		TemplateData: TemplateData{
			PackageName: "testmodels",
			JSONLibrary: "encoding/json",
			Collections: []CollectionData{
				{
					CollectionName: "users",
					StructName:     "User",
					Fields: []FieldData{
						{
							JSONName:     "name",
							GoName:       "Name",
							GoType:       "string",
							OmitEmpty:    false,
							GetterMethod: "GetString",
						},
						{
							JSONName:     "email",
							GoName:       "Email",
							GoType:       "*string",
							OmitEmpty:    true,
							GetterMethod: "GetStringPointer",
						},
					},
				},
			},
		},
		Enums: []EnumData{
			{
				CollectionName: "users",
				FieldName:      "status",
				EnumTypeName:   "UserStatusType",
				Constants: []ConstantData{
					{Name: "UserStatusActive", Value: "active"},
					{Name: "UserStatusInactive", Value: "inactive"},
				},
			},
		},
		RelationTypes: []RelationTypeData{
			{
				TypeName:         "ProfileRelation",
				TargetCollection: "profiles",
				TargetTypeName:   "Profile",
				IsMulti:          false,
			},
		},
		FileTypes: []FileTypeData{
			{
				TypeName:       "AvatarFile",
				IsMulti:        false,
				HasThumbnails:  true,
				ThumbnailSizes: []string{"50x50", "100x100"},
			},
		},
		GenerateEnums:     true,
		GenerateRelations: true,
		GenerateFiles:     true,
	}

	// 템플릿 실행
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, testData)
	if err != nil {
		t.Fatalf("Template execution failed: %v", err)
	}

	// 생성된 코드를 파일로 저장
	generatedFile := filepath.Join(tempDir, "models.go")
	err = os.WriteFile(generatedFile, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("Failed to write generated file: %v", err)
	}

	// go.mod 파일 생성 (컴파일을 위해 필요)
	goModContent := `module testmodels

go 1.21

require (
	github.com/mrchypark/pocketbase-client v0.0.0
	github.com/pocketbase/pocketbase v0.0.0
)

replace github.com/mrchypark/pocketbase-client => ../../../
replace github.com/pocketbase/pocketbase => github.com/pocketbase/pocketbase v0.22.0
`
	goModFile := filepath.Join(tempDir, "go.mod")
	err = os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write go.mod file: %v", err)
	}

	// 생성된 코드 읽기
	generatedCode, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	// Go 문법 검사 (format.Source를 사용하여 문법 오류 확인)
	_, err = format.Source(generatedCode)
	if err != nil {
		t.Errorf("Generated code has syntax errors: %v", err)
		t.Logf("Generated code:\n%s", string(generatedCode))
	}

	// 생성된 코드에 예상되는 구조들이 포함되어 있는지 확인
	codeStr := string(generatedCode)
	expectedStructures := []string{
		"package testmodels",
		"type User struct",
		"func NewUser()",
		"func GetUser(",
		"func GetUserList(",
		"UserStatusActive",
		"UserStatusInactive",
		"type ProfileRelation struct",
		"func (r ProfileRelation) Load(",
		"type FileReference struct",
		"func (f FileReference) URL(",
	}

	for _, expected := range expectedStructures {
		if !strings.Contains(codeStr, expected) {
			t.Errorf("Generated code missing expected structure: %s", expected)
		}
	}
}

func TestTemplateWithDifferentSchemaPatterns(t *testing.T) {
	// 템플릿 파일 읽기
	templatePath := "../../cmd/pbc-gen/template.go.tpl"
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template file: %v", err)
	}

	// 템플릿 파싱
	tmpl, err := template.New("test").Parse(string(templateContent))
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	tests := []struct {
		name        string
		data        EnhancedTemplateData
		description string
	}{
		{
			name:        "complex_schema_with_all_field_types",
			description: "복잡한 스키마 패턴 - 모든 필드 타입 포함",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "github.com/goccy/go-json",
					Collections: []CollectionData{
						{
							CollectionName: "complex_records",
							StructName:     "ComplexRecord",
							Fields: []FieldData{
								{JSONName: "text_field", GoName: "TextField", GoType: "string", OmitEmpty: false, GetterMethod: "GetString"},
								{JSONName: "number_field", GoName: "NumberField", GoType: "*float64", OmitEmpty: true, GetterMethod: "GetFloatPointer"},
								{JSONName: "bool_field", GoName: "BoolField", GoType: "*bool", OmitEmpty: true, GetterMethod: "GetBoolPointer"},
								{JSONName: "date_field", GoName: "DateField", GoType: "*types.DateTime", OmitEmpty: true, GetterMethod: "GetDateTimePointer"},
								{JSONName: "json_field", GoName: "JsonField", GoType: "json.RawMessage", OmitEmpty: false, GetterMethod: "GetRawMessage"},
								{JSONName: "select_single", GoName: "SelectSingle", GoType: "*string", OmitEmpty: true, GetterMethod: "GetStringPointer"},
								{JSONName: "select_multi", GoName: "SelectMulti", GoType: "[]string", OmitEmpty: false, GetterMethod: "GetStringSlice"},
								{JSONName: "relation_single", GoName: "RelationSingle", GoType: "*string", OmitEmpty: true, GetterMethod: "GetStringPointer"},
								{JSONName: "relation_multi", GoName: "RelationMulti", GoType: "[]string", OmitEmpty: false, GetterMethod: "GetStringSlice"},
								{JSONName: "file_single", GoName: "FileSingle", GoType: "*string", OmitEmpty: true, GetterMethod: "GetStringPointer"},
								{JSONName: "file_multi", GoName: "FileMulti", GoType: "[]string", OmitEmpty: false, GetterMethod: "GetStringSlice"},
							},
						},
					},
				},
				Enums: []EnumData{
					{
						CollectionName: "complex_records",
						FieldName:      "select_single",
						EnumTypeName:   "ComplexRecordsSelectSingleType",
						Constants: []ConstantData{
							{Name: "ComplexRecordsSelectSingleOption1", Value: "option1"},
							{Name: "ComplexRecordsSelectSingleOption2", Value: "option2"},
							{Name: "ComplexRecordsSelectSingleSpecialChars", Value: "special-chars_test"},
						},
					},
					{
						CollectionName: "complex_records",
						FieldName:      "select_multi",
						EnumTypeName:   "ComplexRecordsSelectMultiType",
						Constants: []ConstantData{
							{Name: "ComplexRecordsSelectMultiValueA", Value: "value_a"},
							{Name: "ComplexRecordsSelectMultiValueB", Value: "value_b"},
							{Name: "ComplexRecordsSelectMultiValueC", Value: "value_c"},
						},
					},
				},
				RelationTypes: []RelationTypeData{
					{
						TypeName:         "SingleRelation",
						TargetCollection: "target_collection",
						TargetTypeName:   "TargetCollection",
						IsMulti:          false,
					},
					{
						TypeName:         "MultiRelation",
						TargetCollection: "multi_target",
						TargetTypeName:   "MultiTarget",
						IsMulti:          true,
					},
				},
				GenerateEnums:     true,
				GenerateRelations: true,
				GenerateFiles:     true,
			},
		},
		{
			name:        "edge_case_empty_collections",
			description: "엣지 케이스 - 빈 컬렉션",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "encoding/json",
					Collections: []CollectionData{},
				},
				GenerateEnums:     false,
				GenerateRelations: false,
				GenerateFiles:     false,
			},
		},
		{
			name:        "special_characters_in_names",
			description: "특수 문자가 포함된 이름들",
			data: EnhancedTemplateData{
				TemplateData: TemplateData{
					PackageName: "models",
					JSONLibrary: "encoding/json",
					Collections: []CollectionData{
						{
							CollectionName: "special_chars_test",
							StructName:     "SpecialCharsTest",
							Fields: []FieldData{
								{JSONName: "field_with_underscores", GoName: "FieldWithUnderscores", GoType: "string", OmitEmpty: false, GetterMethod: "GetString"},
								{JSONName: "field-with-hyphens", GoName: "FieldWithHyphens", GoType: "string", OmitEmpty: false, GetterMethod: "GetString"},
								{JSONName: "fieldWithCamelCase", GoName: "FieldWithCamelCase", GoType: "string", OmitEmpty: false, GetterMethod: "GetString"},
							},
						},
					},
				},
				Enums: []EnumData{
					{
						CollectionName: "special_chars_test",
						FieldName:      "status",
						EnumTypeName:   "SpecialCharsTestStatusType",
						Constants: []ConstantData{
							{Name: "SpecialCharsTestStatusValueWithSpaces", Value: "value with spaces"},
							{Name: "SpecialCharsTestStatusValueWithHyphens", Value: "value-with-hyphens"},
							{Name: "SpecialCharsTestStatusValueWithUnderscores", Value: "value_with_underscores"},
						},
					},
				},
				GenerateEnums:     true,
				GenerateRelations: false,
				GenerateFiles:     false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("Testing: %s", tt.description)

			// 템플릿 실행
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, tt.data)
			if err != nil {
				t.Fatalf("Template execution failed: %v", err)
			}

			generatedCode := buf.String()

			// 기본 구조 확인
			if tt.data.PackageName != "" {
				if !strings.Contains(generatedCode, "package "+tt.data.PackageName) {
					t.Errorf("Generated code missing package declaration")
				}
			}

			// Go 문법 검사
			if strings.TrimSpace(generatedCode) != "" {
				_, err = format.Source([]byte(generatedCode))
				if err != nil {
					t.Errorf("Generated code has syntax errors: %v", err)
				}
			}
		})
	}
}
