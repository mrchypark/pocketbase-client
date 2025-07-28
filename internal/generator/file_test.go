package generator

import (
	"reflect"
	"strings"
	"testing"
)

func TestFileGenerator_GenerateFileTypes(t *testing.T) {
	generator := NewFileGenerator()

	// Prepare test data
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

	// Expected result: 3 file types (users.avatar, users.documents, posts.featured_image)
	expectedCount := 3
	if len(result) != expectedCount {
		t.Errorf("Expected %d file types, got %d", expectedCount, len(result))
	}

	// Verify users.avatar file type (single file, with thumbnails)
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

	// Verify users.documents file type (multiple files, no thumbnails)
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

func TestFileGenerator_GenerateFileTypeData(t *testing.T) {
	generator := NewFileGenerator()

	enhanced := EnhancedFieldInfo{
		FieldSchema: FieldSchema{
			Name: "image",
			Type: "file",
		},
		FileTypeName:   "ImageFile",
		IsMultiFile:    false,
		HasThumbnails:  true,
		ThumbnailSizes: []string{"150x150", "300x300"},
	}

	result := generator.GenerateFileTypeData(enhanced, "gallery")

	expected := FileTypeData{
		TypeName:       "ImageFile",
		IsMulti:        false,
		HasThumbnails:  true,
		ThumbnailSizes: []string{"150x150", "300x300"},
		Methods:        generator.GenerateFileMethods(enhanced),
	}

	if result.TypeName != expected.TypeName {
		t.Errorf("Expected TypeName '%s', got '%s'", expected.TypeName, result.TypeName)
	}

	if result.IsMulti != expected.IsMulti {
		t.Errorf("Expected IsMulti %v, got %v", expected.IsMulti, result.IsMulti)
	}

	if result.HasThumbnails != expected.HasThumbnails {
		t.Errorf("Expected HasThumbnails %v, got %v", expected.HasThumbnails, result.HasThumbnails)
	}

	if !reflect.DeepEqual(result.ThumbnailSizes, expected.ThumbnailSizes) {
		t.Errorf("Expected ThumbnailSizes %v, got %v", expected.ThumbnailSizes, result.ThumbnailSizes)
	}
}

func TestFileGenerator_GenerateFileMethods(t *testing.T) {
	generator := NewFileGenerator()

	tests := []struct {
		name                string
		enhanced            EnhancedFieldInfo
		expectedMethodCount int
		expectedMethods     []string
	}{
		{
			name: "file without thumbnails",
			enhanced: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "document",
					Type: "file",
				},
				HasThumbnails: false,
			},
			expectedMethodCount: 3,
			expectedMethods:     []string{"Filename", "URL", "IsEmpty"},
		},
		{
			name: "file with thumbnails",
			enhanced: EnhancedFieldInfo{
				FieldSchema: FieldSchema{
					Name: "image",
					Type: "file",
				},
				HasThumbnails:  true,
				ThumbnailSizes: []string{"100x100"},
			},
			expectedMethodCount: 4,
			expectedMethods:     []string{"Filename", "URL", "ThumbURL", "IsEmpty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GenerateFileMethods(tt.enhanced)

			if len(result) != tt.expectedMethodCount {
				t.Errorf("Expected %d methods, got %d", tt.expectedMethodCount, len(result))
			}

			for i, expectedMethod := range tt.expectedMethods {
				if i >= len(result) {
					t.Errorf("Missing expected method: %s", expectedMethod)
					continue
				}
				if result[i].Name != expectedMethod {
					t.Errorf("Expected method name '%s', got '%s'", expectedMethod, result[i].Name)
				}
			}

			// Verify Filename method
			filenameMethod := result[0]
			if filenameMethod.ReturnType != "string" {
				t.Errorf("Expected Filename method return type 'string', got '%s'", filenameMethod.ReturnType)
			}

			// Verify URL method
			urlMethod := result[1]
			if urlMethod.ReturnType != "string" {
				t.Errorf("Expected URL method return type 'string', got '%s'", urlMethod.ReturnType)
			}

			// Verify IsEmpty method (last method)
			isEmptyMethod := result[len(result)-1]
			if isEmptyMethod.Name != "IsEmpty" {
				t.Errorf("Expected last method to be 'IsEmpty', got '%s'", isEmptyMethod.Name)
			}
			if isEmptyMethod.ReturnType != "bool" {
				t.Errorf("Expected IsEmpty method return type 'bool', got '%s'", isEmptyMethod.ReturnType)
			}
		})
	}
}

func TestFileGenerator_GenerateFileTypeCode(t *testing.T) {
	generator := NewFileGenerator()

	fileType := FileTypeData{
		TypeName:      "ImageFile",
		IsMulti:       false,
		HasThumbnails: true,
		Methods: []MethodData{
			{Name: "Filename", ReturnType: "string", Body: "return f.filename"},
			{Name: "URL", ReturnType: "string", Body: "return f.url"},
			{Name: "IsEmpty", ReturnType: "bool", Body: `return f.filename == ""`},
		},
	}

	result := generator.GenerateFileTypeCode(fileType)

	// Verify that generated code has correct format
	expectedParts := []string{
		"type FileReference struct",
		"filename   string",
		"recordID   string",
		"collection string",
		"fieldName  string",
		"func (f FileReference) Filename() string",
		"func (f FileReference) URL(baseURL string) string",
		"func (f FileReference) IsEmpty() bool",
		"func NewFileReference(filename, recordID, collection, fieldName string) FileReference",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Generated file type code missing expected part: %s\nGenerated:\n%s", part, result)
		}
	}
}

func TestFileGenerator_GenerateFileMethodCode(t *testing.T) {
	generator := NewFileGenerator()

	tests := []struct {
		name     string
		method   MethodData
		expected []string
	}{
		{
			name:   "Filename method",
			method: MethodData{Name: "Filename", ReturnType: "string", Body: "return f.filename"},
			expected: []string{
				"func (f FileReference) Filename() string",
				"return f.filename",
				"// Filename returns the filename",
			},
		},
		{
			name:   "URL method",
			method: MethodData{Name: "URL", ReturnType: "string", Body: "return f.url"},
			expected: []string{
				"func (f FileReference) URL(baseURL string) string",
				"return f.url",
				"// URL generates the file URL",
			},
		},
		{
			name:   "ThumbURL method",
			method: MethodData{Name: "ThumbURL", ReturnType: "string", Body: "return f.thumbUrl"},
			expected: []string{
				"func (f FileReference) ThumbURL(baseURL, thumb string) string",
				"return f.thumbUrl",
				"// ThumbURL generates thumbnail URL",
			},
		},
		{
			name:   "IsEmpty method",
			method: MethodData{Name: "IsEmpty", ReturnType: "bool", Body: `return f.filename == ""`},
			expected: []string{
				"func (f FileReference) IsEmpty() bool",
				`return f.filename == ""`,
				"// IsEmpty returns true if the file reference is empty",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GenerateFileMethodCode(tt.method)

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Generated method code missing expected part: %s\nGenerated:\n%s", expected, result)
				}
			}
		})
	}
}

func TestFileGenerator_GenerateFileConstructorCode(t *testing.T) {
	generator := NewFileGenerator()

	result := generator.GenerateFileConstructorCode()

	expectedParts := []string{
		"func NewFileReference(filename, recordID, collection, fieldName string) FileReference",
		"return FileReference{",
		"filename:   filename,",
		"recordID:   recordID,",
		"collection: collection,",
		"fieldName:  fieldName,",
		"// NewFileReference creates a new FileReference",
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Generated constructor code missing expected part: %s\nGenerated:\n%s", part, result)
		}
	}
}

func TestFileGenerator_GenerateMultiFileTypeCode(t *testing.T) {
	generator := NewFileGenerator()

	// Single file test
	singleFileType := FileTypeData{
		TypeName:      "DocumentFile",
		IsMulti:       false,
		HasThumbnails: false,
		Methods: []MethodData{
			{Name: "Filename", ReturnType: "string", Body: "return f.filename"},
		},
	}

	singleResult := generator.GenerateMultiFileTypeCode(singleFileType)
	if !strings.Contains(singleResult, "type FileReference struct") {
		t.Error("Single file should generate normal file type")
	}

	// Multiple file test
	multiFileType := FileTypeData{
		TypeName:      "ImageFile",
		IsMulti:       true,
		HasThumbnails: true,
	}

	multiResult := generator.GenerateMultiFileTypeCode(multiFileType)

	expectedParts := []string{
		"type FileReferences []FileReference",
		"func (f FileReferences) Filenames() []string",
		"func (f FileReferences) URLs(baseURL string) []string",
		"func (f FileReferences) ThumbURLs(baseURL, thumb string) []string",
		"func (f FileReferences) IsEmpty() bool",
		"func (f FileReferences) Filter() FileReferences",
	}

	for _, part := range expectedParts {
		if !strings.Contains(multiResult, part) {
			t.Errorf("Generated multi-file code missing expected part: %s\nGenerated:\n%s", part, multiResult)
		}
	}
}

func TestFileGenerator_GenerateMultiFileMethods(t *testing.T) {
	generator := NewFileGenerator()

	tests := []struct {
		name          string
		fileType      FileTypeData
		expectedParts []string
	}{
		{
			name: "multi-file without thumbnails",
			fileType: FileTypeData{
				TypeName:      "DocumentFile",
				IsMulti:       true,
				HasThumbnails: false,
			},
			expectedParts: []string{
				"func (f FileReferences) Filenames() []string",
				"func (f FileReferences) URLs(baseURL string) []string",
				"func (f FileReferences) IsEmpty() bool",
				"func (f FileReferences) Filter() FileReferences",
			},
		},
		{
			name: "multi-file with thumbnails",
			fileType: FileTypeData{
				TypeName:      "ImageFile",
				IsMulti:       true,
				HasThumbnails: true,
			},
			expectedParts: []string{
				"func (f FileReferences) Filenames() []string",
				"func (f FileReferences) URLs(baseURL string) []string",
				"func (f FileReferences) ThumbURLs(baseURL, thumb string) []string",
				"func (f FileReferences) IsEmpty() bool",
				"func (f FileReferences) Filter() FileReferences",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GenerateMultiFileMethods(tt.fileType, "FileReferences")

			for _, part := range tt.expectedParts {
				if !strings.Contains(result, part) {
					t.Errorf("Generated multi-file methods missing expected part: %s\nGenerated:\n%s", part, result)
				}
			}
		})
	}
}

func TestFileGenerator_ValidateFileName(t *testing.T) {
	generator := NewFileGenerator()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "valid filename",
			input: "image.jpg",
			want:  "image.jpg",
		},
		{
			name:  "filename with underscores and hyphens",
			input: "my_file-name.pdf",
			want:  "my_file-name.pdf",
		},
		{
			name:  "filename with invalid characters",
			input: "file@name#with$special%.txt",
			want:  "filenamewithspecial.txt",
		},
		{
			name:  "empty filename",
			input: "",
			want:  "file",
		},
		{
			name:  "only invalid characters",
			input: "!@#$%^&*()",
			want:  "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generator.ValidateFileName(tt.input)
			if got != tt.want {
				t.Errorf("ValidateFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileGenerator_GenerateThumbnailSizeConstants(t *testing.T) {
	generator := NewFileGenerator()

	tests := []struct {
		name     string
		fileType FileTypeData
		expected []string
	}{
		{
			name: "file without thumbnails",
			fileType: FileTypeData{
				HasThumbnails: false,
			},
			expected: []string{}, // Should return empty string
		},
		{
			name: "file with thumbnails",
			fileType: FileTypeData{
				HasThumbnails:  true,
				ThumbnailSizes: []string{"100x100", "200x200", "400x400"},
			},
			expected: []string{
				"const (",
				`ThumbSize1 = "100x100"`,
				`ThumbSize2 = "200x200"`,
				`ThumbSize3 = "400x400"`,
				")",
			},
		},
		{
			name: "file with empty thumbnail sizes",
			fileType: FileTypeData{
				HasThumbnails:  true,
				ThumbnailSizes: []string{},
			},
			expected: []string{}, // Should return empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GenerateThumbnailSizeConstants(tt.fileType)

			if len(tt.expected) == 0 {
				if result != "" {
					t.Errorf("Expected empty string, got: %s", result)
				}
				return
			}

			for _, expected := range tt.expected {
				if !strings.Contains(result, expected) {
					t.Errorf("Generated thumbnail constants missing expected part: %s\nGenerated:\n%s", expected, result)
				}
			}
		})
	}
}

func TestNewFileGenerator(t *testing.T) {
	generator := NewFileGenerator()
	if generator == nil {
		t.Error("NewFileGenerator() returned nil")
	}
}
