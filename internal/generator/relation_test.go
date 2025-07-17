package generator

import (
	"strings"
	"testing"
)

func TestRelationGenerator_GenerateRelationTypes(t *testing.T) {
	generator := NewRelationGenerator()

	// 테스트용 데이터 준비
	collections := []CollectionData{
		{
			CollectionName: "posts",
		},
		{
			CollectionName: "comments",
		},
	}

	schemas := []CollectionSchema{
		{
			Name: "posts",
			Fields: []FieldSchema{
				{
					Name:   "author",
					Type:   "relation",
					System: false,
					Options: &FieldOptions{
						CollectionID: "users_collection_id",
						MaxSelect:    func() *int { i := 1; return &i }(),
					},
				},
				{
					Name:   "categories",
					Type:   "relation",
					System: false,
					Options: &FieldOptions{
						CollectionID: "categories_collection_id",
						MaxSelect:    func() *int { i := 3; return &i }(),
					},
				},
				{
					Name:   "title",
					Type:   "text",
					System: false,
				},
			},
		},
		{
			Name: "comments",
			Fields: []FieldSchema{
				{
					Name:   "post",
					Type:   "relation",
					System: false,
					Options: &FieldOptions{
						CollectionID: "posts_collection_id",
						MaxSelect:    func() *int { i := 1; return &i }(),
					},
				},
				{
					Name:   "id",
					Type:   "text",
					System: true, // System field should be ignored
				},
			},
		},
		{
			ID:   "users_collection_id",
			Name: "users",
		},
		{
			ID:   "categories_collection_id",
			Name: "categories",
		},
		{
			ID:   "posts_collection_id",
			Name: "posts",
		},
	}

	result := generator.GenerateRelationTypes(collections, schemas)

	// 예상 결과: 3개의 relation type (posts.author, posts.categories, comments.post)
	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("Expected %d relation types, got %d", expectedCount, len(result))
	}

	// posts.author relation 검증 (단일 관계)
	var authorRelation *RelationTypeData
	for i := range result {
		if result[i].TypeName == "UsersRelation" {
			authorRelation = &result[i]
			break
		}
	}

	if authorRelation == nil {
		t.Fatal("posts.author relation not found")
	}

	if authorRelation.IsMulti {
		t.Error("posts.author should be single relation, got multi")
	}

	if authorRelation.TargetCollection != "users" {
		t.Errorf("Expected target collection 'users', got '%s'", authorRelation.TargetCollection)
	}

	// posts.categories relation 검증 (다중 관계)
	var categoriesRelation *RelationTypeData
	for i := range result {
		if result[i].TypeName == "CategoriesRelation" {
			categoriesRelation = &result[i]
			break
		}
	}

	if categoriesRelation == nil {
		t.Fatal("posts.categories relation not found")
	}

	if !categoriesRelation.IsMulti {
		t.Error("posts.categories should be multi relation, got single")
	}
}

func TestRelationGenerator_GenerateRelationTypeData(t *testing.T) {
	generator := NewRelationGenerator()

	enhanced := EnhancedFieldInfo{
		FieldSchema: FieldSchema{
			Name: "author",
			Type: "relation",
		},
		TargetCollection: "users",
		RelationTypeName: "UsersRelation",
		IsMultiRelation:  false,
	}

	result := generator.GenerateRelationTypeData(enhanced, "posts")

	expected := RelationTypeData{
		TypeName:         "UsersRelation",
		TargetCollection: "users",
		TargetTypeName:   "Users",
		IsMulti:          false,
		Methods:          generator.GenerateRelationMethods(enhanced),
	}

	if result.TypeName != expected.TypeName {
		t.Errorf("Expected TypeName '%s', got '%s'", expected.TypeName, result.TypeName)
	}

	if result.TargetCollection != expected.TargetCollection {
		t.Errorf("Expected TargetCollection '%s', got '%s'", expected.TargetCollection, result.TargetCollection)
	}

	if result.IsMulti != expected.IsMulti {
		t.Errorf("Expected IsMulti %v, got %v", expected.IsMulti, result.IsMulti)
	}

	if len(result.Methods) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(result.Methods))
	}
}

func TestRelationGenerator_GenerateRelationMethods(t *testing.T) {
	generator := NewRelationGenerator()

	enhanced := EnhancedFieldInfo{
		FieldSchema: FieldSchema{
			Name: "category",
			Type: "relation",
		},
		TargetCollection: "categories",
		RelationTypeName: "CategoriesRelation",
		IsMultiRelation:  false,
	}

	result := generator.GenerateRelationMethods(enhanced)

	expectedMethodNames := []string{"ID", "Load", "IsEmpty"}
	if len(result) != len(expectedMethodNames) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethodNames), len(result))
	}

	for i, method := range result {
		if method.Name != expectedMethodNames[i] {
			t.Errorf("Expected method name '%s', got '%s'", expectedMethodNames[i], method.Name)
		}
	}

	// ID method 검증
	idMethod := result[0]
	if idMethod.ReturnType != "string" {
		t.Errorf("Expected ID method return type 'string', got '%s'", idMethod.ReturnType)
	}

	// Load method 검증
	loadMethod := result[1]
	expectedLoadReturnType := "(*Categories, error)"
	if loadMethod.ReturnType != expectedLoadReturnType {
		t.Errorf("Expected Load method return type '%s', got '%s'", expectedLoadReturnType, loadMethod.ReturnType)
	}

	// IsEmpty method 검증
	isEmptyMethod := result[2]
	if isEmptyMethod.ReturnType != "bool" {
		t.Errorf("Expected IsEmpty method return type 'bool', got '%s'", isEmptyMethod.ReturnType)
	}
}

func TestRelationGenerator_GenerateRelationTypeCode(t *testing.T) {
	generator := NewRelationGenerator()

	relationType := RelationTypeData{
		TypeName:         "UserRelation",
		TargetCollection: "users",
		TargetTypeName:   "User",
		IsMulti:          false,
		Methods: []MethodData{
			{Name: "ID", ReturnType: "string", Body: "return r.id"},
			{Name: "IsEmpty", ReturnType: "bool", Body: `return r.id == ""`},
		},
	}

	result := generator.GenerateRelationTypeCode(relationType)

	// 생성된 코드가 올바른 형식인지 확인
	expectedParts := []string{
		"type UserRelation struct",
		"id string",
		"func (r UserRelation) ID() string",
		"func (r UserRelation) IsEmpty() bool",
		"func NewUserRelation(id string) UserRelation",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Generated relation type code missing expected part: %s\nGenerated:\n%s", part, result)
		}
	}
}

func TestRelationGenerator_GenerateMethodCode(t *testing.T) {
	generator := NewRelationGenerator()

	tests := []struct {
		name     string
		typeName string
		method   MethodData
		expected []string
	}{
		{
			name:     "ID method",
			typeName: "UserRelation",
			method:   MethodData{Name: "ID", ReturnType: "string", Body: "return r.id"},
			expected: []string{
				"func (r UserRelation) ID() string",
				"return r.id",
			},
		},
		{
			name:     "Load method",
			typeName: "UserRelation",
			method: MethodData{
				Name:       "Load",
				ReturnType: "(*User, error)",
				Body:       "if r.id == \"\"\nreturn nil, nil",
			},
			expected: []string{
				"func (r UserRelation) Load(ctx context.Context, client pocketbase.RecordServiceAPI) (*User, error)",
				"if r.id == \"\"",
				"return nil, nil",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GenerateMethodCode(tt.typeName, tt.method)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Generated method code missing expected part: %s\nGenerated:\n%s", expected, result)
				}
			}
		})
	}
}

func TestRelationGenerator_GenerateConstructorCode(t *testing.T) {
	generator := NewRelationGenerator()

	relationType := RelationTypeData{
		TypeName: "CategoryRelation",
	}

	result := generator.GenerateConstructorCode(relationType)

	expectedParts := []string{
		"func NewCategoryRelation(id string) CategoryRelation",
		"return CategoryRelation{id: id}",
		"// NewCategoryRelation creates a new CategoryRelation",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Generated constructor code missing expected part: %s\nGenerated:\n%s", part, result)
		}
	}
}

func TestRelationGenerator_GenerateMultiRelationTypeCode(t *testing.T) {
	generator := NewRelationGenerator()

	// 단일 관계 테스트
	singleRelationType := RelationTypeData{
		TypeName:         "UserRelation",
		TargetCollection: "users",
		IsMulti:          false,
		Methods: []MethodData{
			{Name: "ID", ReturnType: "string", Body: "return r.id"},
		},
	}

	singleResult := generator.GenerateMultiRelationTypeCode(singleRelationType)
	if !strings.Contains(singleResult, "type UserRelation struct") {
		t.Error("Single relation should generate normal relation type")
	}

	// 다중 관계 테스트
	multiRelationType := RelationTypeData{
		TypeName:         "CategoryRelation",
		TargetCollection: "categories",
		TargetTypeName:   "Category",
		IsMulti:          true,
	}

	multiResult := generator.GenerateMultiRelationTypeCode(multiRelationType)

	expectedParts := []string{
		"type CategoryRelations []CategoryRelation",
		"func (r CategoryRelations) IDs() []string",
		"func (r CategoryRelations) LoadAll",
		"func (r CategoryRelations) IsEmpty() bool",
	}

	for _, part := range expectedParts {
		if !strings.Contains(multiResult, part) {
			t.Errorf("Generated multi-relation code missing expected part: %s\nGenerated:\n%s", part, multiResult)
		}
	}
}

func TestRelationGenerator_GenerateMultiRelationMethods(t *testing.T) {
	generator := NewRelationGenerator()

	relationType := RelationTypeData{
		TypeName:         "TagRelation",
		TargetCollection: "tags",
		TargetTypeName:   "Tag",
		IsMulti:          true,
	}

	result := generator.GenerateMultiRelationMethods(relationType, "TagRelations")

	expectedParts := []string{
		"func (r TagRelations) IDs() []string",
		"func (r TagRelations) LoadAll(ctx context.Context, client pocketbase.RecordServiceAPI) ([]*Tag, error)",
		"func (r TagRelations) IsEmpty() bool",
		"return len(r) == 0",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Generated multi-relation methods missing expected part: %s\nGenerated:\n%s", part, result)
		}
	}
}

func TestRelationGenerator_ValidateRelationName(t *testing.T) {
	generator := NewRelationGenerator()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid name",
			input: "UserRelation",
			want:  "UserRelation",
		},
		{
			name:  "name with special characters",
			input: "User-Relation_Type!",
			want:  "UserRelationType",
		},
		{
			name:  "name starting with number",
			input: "123Relation",
			want:  "Relation123Relation",
		},
		{
			name:  "empty name",
			input: "",
			want:  "RelationType",
		},
		{
			name:  "only special characters",
			input: "!@#$%",
			want:  "RelationType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generator.ValidateRelationName(tt.input)
			if got != tt.want {
				t.Errorf("ValidateRelationName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRelationGenerator(t *testing.T) {
	generator := NewRelationGenerator()
	if generator == nil {
		t.Error("NewRelationGenerator() returned nil")
	}
}
