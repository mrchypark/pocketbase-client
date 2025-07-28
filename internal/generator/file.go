package generator

import (
	"fmt"
	"strings"
)

// FileGenerator handles file type generation for file fields
type FileGenerator struct{}

// NewFileGenerator creates a new FileGenerator instance
func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

// GenerateFileTypes generates file type data for all collections
func (g *FileGenerator) GenerateFileTypes(collections []CollectionData, schemas []CollectionSchema) []FileTypeData {
	// Performance optimization: pre-allocate slice with estimated size
	estimatedFiles := 0
	for _, schema := range schemas {
		for _, field := range schema.Fields {
			if !field.System && field.Type == "file" {
				estimatedFiles++
			}
		}
	}

	fileTypes := make([]FileTypeData, 0, estimatedFiles)

	// Performance optimization: convert schemas to map for O(1) lookup
	schemaMap := make(map[string]CollectionSchema, len(schemas))
	for _, schema := range schemas {
		schemaMap[schema.Name] = schema
	}

	for _, collection := range collections {
		schema, exists := schemaMap[collection.CollectionName]
		if !exists {
			continue
		}

		for _, field := range schema.Fields {
			if field.System || field.Type != "file" {
				continue
			}

			// Performance optimization: directly extract file information
			fileType := g.generateFileTypeDataOptimized(field, collection.CollectionName)
			fileTypes = append(fileTypes, fileType)
		}
	}

	return fileTypes
}

// GenerateFileTypeData generates file type data for a single file field
func (g *FileGenerator) GenerateFileTypeData(enhanced EnhancedFieldInfo, _ string) FileTypeData {
	methods := g.GenerateFileMethods(enhanced)

	return FileTypeData{
		TypeName:       enhanced.FileTypeName,
		IsMulti:        enhanced.IsMultiFile,
		HasThumbnails:  enhanced.HasThumbnails,
		ThumbnailSizes: enhanced.ThumbnailSizes,
		Methods:        methods,
	}
}

// GenerateFileMethods generates methods for a file type
func (g *FileGenerator) GenerateFileMethods(enhanced EnhancedFieldInfo) []MethodData {
	var methods []MethodData

	// Filename method
	filenameMethod := MethodData{
		Name:       "Filename",
		ReturnType: "string",
		Body:       "return f.filename",
	}
	methods = append(methods, filenameMethod)

	// URL method
	urlMethod := MethodData{
		Name:       "URL",
		ReturnType: "string",
		Body: `if f.filename == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/files/%s/%s/%s", baseURL, f.collection, f.recordID, f.filename)`,
	}
	methods = append(methods, urlMethod)

	// ThumbURL method (if thumbnails are supported)
	if enhanced.HasThumbnails {
		thumbURLMethod := MethodData{
			Name:       "ThumbURL",
			ReturnType: "string",
			Body: `if f.filename == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/files/%s/%s/%s?thumb=%s", baseURL, f.collection, f.recordID, f.filename, thumb)`,
		}
		methods = append(methods, thumbURLMethod)
	}

	// IsEmpty method
	isEmptyMethod := MethodData{
		Name:       "IsEmpty",
		ReturnType: "bool",
		Body:       `return f.filename == ""`,
	}
	methods = append(methods, isEmptyMethod)

	return methods
}

// GenerateFileTypeCode generates the complete Go code for a file type
func (g *FileGenerator) GenerateFileTypeCode(fileType FileTypeData) string {
	var code strings.Builder

	// Type definition
	code.WriteString("// FileReference represents a file reference\n")
	code.WriteString("type FileReference struct {\n")
	code.WriteString("\tfilename   string\n")
	code.WriteString("\trecordID   string\n")
	code.WriteString("\tcollection string\n")
	code.WriteString("\tfieldName  string\n")
	code.WriteString("}\n\n")

	// Methods
	for _, method := range fileType.Methods {
		code.WriteString(g.GenerateFileMethodCode(method))
		code.WriteString("\n")
	}

	// Constructor
	code.WriteString(g.GenerateFileConstructorCode())

	return code.String()
}

// GenerateFileMethodCode generates Go code for a single file method
func (g *FileGenerator) GenerateFileMethodCode(method MethodData) string {
	var code strings.Builder

	// Method signature
	switch method.Name {
	case "Filename":
		code.WriteString("// Filename returns the filename\n")
		code.WriteString("func (f FileReference) Filename() string {\n")
	case "URL":
		code.WriteString("// URL generates the file URL\n")
		code.WriteString("func (f FileReference) URL(baseURL string) string {\n")
	case "ThumbURL":
		code.WriteString("// ThumbURL generates thumbnail URL\n")
		code.WriteString("func (f FileReference) ThumbURL(baseURL, thumb string) string {\n")
	case "IsEmpty":
		code.WriteString("// IsEmpty returns true if the file reference is empty\n")
		code.WriteString("func (f FileReference) IsEmpty() bool {\n")
	default:
		code.WriteString(fmt.Sprintf("// %s method\n", method.Name))
		code.WriteString(fmt.Sprintf("func (f FileReference) %s() %s {\n", method.Name, method.ReturnType))
	}

	// Method body
	lines := strings.Split(method.Body, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			code.WriteString(fmt.Sprintf("\t%s\n", line))
		}
	}

	code.WriteString("}\n")

	return code.String()
}

// GenerateFileConstructorCode generates constructor function for file type
func (g *FileGenerator) GenerateFileConstructorCode() string {
	return `// NewFileReference creates a new FileReference
func NewFileReference(filename, recordID, collection, fieldName string) FileReference {
	return FileReference{
		filename:   filename,
		recordID:   recordID,
		collection: collection,
		fieldName:  fieldName,
	}
}`
}

// GenerateMultiFileTypeCode generates code for multi-file types
func (g *FileGenerator) GenerateMultiFileTypeCode(fileType FileTypeData) string {
	if !fileType.IsMulti {
		return g.GenerateFileTypeCode(fileType)
	}

	var code strings.Builder

	// Multi-file type definition
	multiTypeName := "FileReferences" // Pluralize
	code.WriteString("// FileReferences represents multiple file references\n")
	code.WriteString("type FileReferences []FileReference\n\n")

	// Multi-file methods
	code.WriteString(g.GenerateMultiFileMethods(fileType, multiTypeName))

	return code.String()
}

// GenerateMultiFileMethods generates methods for multi-file types
func (g *FileGenerator) GenerateMultiFileMethods(fileType FileTypeData, multiTypeName string) string {
	var code strings.Builder

	// Filenames method
	code.WriteString(`// Filenames returns all filenames
func (f FileReferences) Filenames() []string {
	names := make([]string, len(f))
	for i, file := range f {
		names[i] = file.Filename()
	}
	return names
}

`)

	// URLs method
	code.WriteString(`// URLs generates URLs for all files
func (f FileReferences) URLs(baseURL string) []string {
	urls := make([]string, len(f))
	for i, file := range f {
		urls[i] = file.URL(baseURL)
	}
	return urls
}

`)

	// ThumbURLs method (if thumbnails are supported)
	if fileType.HasThumbnails {
		code.WriteString(`// ThumbURLs generates thumbnail URLs for all files
func (f FileReferences) ThumbURLs(baseURL, thumb string) []string {
	urls := make([]string, len(f))
	for i, file := range f {
		urls[i] = file.ThumbURL(baseURL, thumb)
	}
	return urls
}

`)
	}

	// IsEmpty method
	code.WriteString(`// IsEmpty returns true if there are no file references
func (f FileReferences) IsEmpty() bool {
	return len(f) == 0
}

`)

	// Filter method
	code.WriteString(`// Filter returns non-empty file references
func (f FileReferences) Filter() FileReferences {
	var filtered FileReferences
	for _, file := range f {
		if !file.IsEmpty() {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

`)

	return code.String()
}

// ValidateFileName ensures the file name is safe for use
func (g *FileGenerator) ValidateFileName(name string) string {
	// Remove any invalid characters and ensure it starts with a letter
	cleaned := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			return r
		}
		return -1
	}, name)

	if cleaned == "" {
		cleaned = "file"
	}

	return cleaned
}

// GenerateThumbnailSizeConstants generates constants for thumbnail sizes
func (g *FileGenerator) GenerateThumbnailSizeConstants(fileType FileTypeData) string {
	if !fileType.HasThumbnails || len(fileType.ThumbnailSizes) == 0 {
		return ""
	}

	var code strings.Builder

	code.WriteString("// Thumbnail size constants\n")
	code.WriteString("const (\n")

	for i, size := range fileType.ThumbnailSizes {
		constantName := fmt.Sprintf("ThumbSize%d", i+1)
		code.WriteString(fmt.Sprintf("\t%s = \"%s\"\n", constantName, size))
	}

	code.WriteString(")\n\n")

	return code.String()
}

// generateFileTypeDataOptimized performance-optimized file type data generation function
func (g *FileGenerator) generateFileTypeDataOptimized(field FieldSchema, _ string) FileTypeData {
	// Performance optimization: use pre-calculated values
	fileTypeName := ToPascalCase(field.Name) + "File"

	// Check multi-file status and thumbnail information
	isMulti := field.Options != nil && field.Options.MaxSelect != nil && *field.Options.MaxSelect > 1
	hasThumbnails := field.Options != nil && len(field.Options.Thumbs) > 0
	var thumbnailSizes []string
	if hasThumbnails {
		thumbnailSizes = field.Options.Thumbs
	}

	// Generate methods (using pre-allocated slice)
	methodCount := 3 // Filename, URL, IsEmpty
	if hasThumbnails {
		methodCount++ // ThumbURL
	}
	methods := make([]MethodData, 0, methodCount)

	// Filename method
	methods = append(methods, MethodData{
		Name:       "Filename",
		ReturnType: "string",
		Body:       "return f.filename",
	})

	// URL method
	methods = append(methods, MethodData{
		Name:       "URL",
		ReturnType: "string",
		Body: `if f.filename == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/files/%s/%s/%s", baseURL, f.collection, f.recordID, f.filename)`,
	})

	// ThumbURL method (if thumbnails are supported)
	if hasThumbnails {
		methods = append(methods, MethodData{
			Name:       "ThumbURL",
			ReturnType: "string",
			Body: `if f.filename == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/files/%s/%s/%s?thumb=%s", baseURL, f.collection, f.recordID, f.filename, thumb)`,
		})
	}

	// IsEmpty method
	methods = append(methods, MethodData{
		Name:       "IsEmpty",
		ReturnType: "bool",
		Body:       `return f.filename == ""`,
	})

	return FileTypeData{
		TypeName:       fileTypeName,
		IsMulti:        isMulti,
		HasThumbnails:  hasThumbnails,
		ThumbnailSizes: thumbnailSizes,
		Methods:        methods,
	}
}
