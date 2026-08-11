package generator

import (
	"reflect"
	"testing"
)

func TestFileGenerator_GenerateFileTypes(t *testing.T) {
	generator := NewFileGenerator()

	// 테스트용 데이터 준비
	collections := []CollectionData{
		{
			CollectionName: "users",
		},
		{
			CollectionName: "posts",
		},
	}

	schemas := []CollectionSchema{
		{
			Name: "users",
			Fields: []FieldSchema{
				{
					Name:   "avatar",
					Type:   "file",
					System: false,
					Options: &FieldOptions{
						MaxSelect: func() *int { i := 1; return &i }(),
						Thumbs:    []string{"100x100", "200x200"},
					},
				},
				{
					Name:   "documents",
					Type:   "file",
					System: false,
					Options: &FieldOptions{
						MaxSelect: func() *int { i := 5; return &i }(),
					},
				},
				{
					Name:   "name",
					Type:   "text",
					System: false,
				},
			},
		},
		{
			Name: "posts",
			Fields: []FieldSchema{
				{
					Name:   "featured_image",
					Type:   "file",
					System: false,
					Options: &FieldOptions{
						MaxSelect: func() *int { i := 1; return &i }(),
						Thumbs:    []string{"400x0", "800x0"},
					},
				},
				{
					Name:   "id",
					Type:   "text",
					System: true, // System field should be ignored
				},
			},
		},
	}

	result := generator.GenerateFileTypes(collections, schemas)

	// 예상 결과: 3개의 file type (users.avatar, users.documents, posts.featured_image)
	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("Expected %d file types, got %d", expectedCount, len(result))
	}

	// users.avatar file type 검증 (단일 파일, 썸네일 있음)
	var avatarFile *FileTypeData
	for i := range result {
		if result[i].TypeName == "AvatarFile" {
			avatarFile = &result[i]
			break
		}
	}

	if avatarFile == nil {
		t.Fatal("users.avatar file type not found")
	}

	if avatarFile.IsMulti {
		t.Error("users.avatar should be single file, got multi")
	}

	if !avatarFile.HasThumbnails {
		t.Error("users.avatar should have thumbnails")
	}

	expectedThumbnails := []string{"100x100", "200x200"}
	if !reflect.DeepEqual(avatarFile.ThumbnailSizes, expectedThumbnails) {
		t.Errorf("Expected thumbnails %v, got %v", expectedThumbnails, avatarFile.ThumbnailSizes)
	}

	// users.documents file type 검증 (다중 파일, 썸네일 없음)
	var documentsFile *FileTypeData
	for i := range result {
		if result[i].TypeName == "DocumentsFile" {
			documentsFile = &result[i]
			break
		}
	}

	if documentsFile == nil {
		t.Fatal("users.documents file type not found")
	}

	if !documentsFile.IsMulti {
		t.Error("users.documents should be multi file, got single")
	}

	if documentsFile.HasThumbnails {
		t.Error("users.documents should not have thumbnails")
	}
}

func TestNewFileGenerator(t *testing.T) {
	generator := NewFileGenerator()
	if generator == nil {
		t.Error("NewFileGenerator() returned nil")
	}
}
