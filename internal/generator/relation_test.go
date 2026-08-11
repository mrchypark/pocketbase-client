package generator

import (
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
		{
			CollectionName: "users",
		},
		{
			CollectionName: "categories",
		},
	}

	schemas := []CollectionSchema{
		{
			ID:   "posts_collection_id",
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
			ID:   "comments_collection_id",
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

func TestNewRelationGenerator(t *testing.T) {
	generator := NewRelationGenerator()
	if generator == nil {
		t.Error("NewRelationGenerator() returned nil")
	}
}
