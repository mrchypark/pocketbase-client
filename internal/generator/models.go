package generator

// TemplateData represents the data structure used for code generation templates.
// It contains all the information needed to generate Go code from PocketBase schemas.
type TemplateData struct {
	PackageName string // Go package name for generated code
	JSONLibrary string // JSON library import path (e.g., "encoding/json")
	Collections []CollectionData

	// Fields for Enhanced features (default is empty slice)
	Enums         []EnumData         `json:"enums,omitempty"`
	RelationTypes []RelationTypeData `json:"relationTypes,omitempty"`
	FileTypes     []FileTypeData     `json:"fileTypes,omitempty"`
}

// CollectionData represents a single PocketBase collection and its metadata
// for code generation purposes.
type CollectionData struct {
	CollectionName string // PocketBase collection name (e.g., 'posts')
	StructName     string // Generated Go struct name (e.g., 'Post')
	Fields         []FieldData
}

// FieldData represents a single field within a collection and its
// corresponding Go type information.
type FieldData struct {
	JSONName     string // JSON field name as it appears in PocketBase (e.g., 'is_published')
	GoName       string // Go field name in PascalCase (e.g., 'IsPublished')
	GoType       string // Go type for the field (e.g., 'bool')
	OmitEmpty    bool   // Whether to add omitempty tag for optional fields
	GetterMethod string // Getter method name (e.g., GetString, GetBool)
	IsPointer    bool   // Whether this field is a pointer type (for ValueOr method generation)
	BaseType     string // Base type without pointer (e.g., 'string' for '*string')
}

// EnhancedFieldInfo contains additional information for enhanced code generation.
// It extends FieldSchema with metadata needed for generating enums, relations, and file types.
type EnhancedFieldInfo struct {
	FieldSchema

	// For Select fields - information for enum generation
	EnumValues   []string // extracted from values option of select field
	EnumTypeName string   // enum type name to be generated (e.g., DeviceType)

	// For Relation fields - information for relation type generation
	TargetCollection string // referenced collection name
	RelationTypeName string // relation type name to be generated (e.g., PlantRelation)
	IsMultiRelation  bool   // whether it's a multiple relationship (maxSelect > 1)

	// For File fields - information for file type generation
	FileTypeName   string   // file type name to be generated (e.g., ImageFile)
	IsMultiFile    bool     // whether it's multiple files (maxSelect > 1)
	HasThumbnails  bool     // whether thumbnails are supported
	ThumbnailSizes []string // list of thumbnail sizes
}

// EnhancedTemplateData extends TemplateData with enhanced generation features.
// It includes additional data for generating enums, relations, and file types.
type EnhancedTemplateData struct {
	TemplateData

	// New features
	Enums         []EnumData         // enum data to generate
	RelationTypes []RelationTypeData // relation type data to generate
	FileTypes     []FileTypeData     // file type data to generate

	// Generation options - controlled by CLI flags
	GenerateEnums     bool // whether to generate enums
	GenerateRelations bool // whether to generate relation types
	GenerateFiles     bool // whether to generate file types
}

// EnumData represents data for generating enum constants from select fields.
// It contains all information needed to generate type-safe enum constants.
type EnumData struct {
	CollectionName string         // collection name (e.g., devices)
	FieldName      string         // field name (e.g., type)
	EnumTypeName   string         // enum type name (e.g., DeviceType)
	Constants      []ConstantData // list of constants
}

// ConstantData represents a single enum constant with its name and value.
type ConstantData struct {
	Name  string // constant name (e.g., DeviceTypeM2)
	Value string // constant value (e.g., "m2")
}

// RelationTypeData represents data for generating relation types from relation fields.
// It contains metadata for creating type-safe relation helpers.
type RelationTypeData struct {
	TypeName         string       // relation type name (e.g., PlantRelation)
	TargetCollection string       // target collection name (e.g., plants)
	TargetTypeName   string       // target type name (e.g., Plant)
	IsMulti          bool         // whether it's a multiple relationship
	Methods          []MethodData // list of methods to generate
}

// FileTypeData represents data for generating file types from file fields.
// It contains metadata for creating file reference helpers with URL generation.
type FileTypeData struct {
	TypeName       string       // file type name (e.g., ImageFile)
	IsMulti        bool         // whether it's multiple files
	HasThumbnails  bool         // whether thumbnails are supported
	ThumbnailSizes []string     // list of thumbnail sizes
	Methods        []MethodData // list of methods to generate
}

// MethodData represents a method to be generated for enhanced types.
// It contains the method signature and implementation details.
type MethodData struct {
	Name       string // method name (e.g., Load, URL)
	ReturnType string // return type (e.g., (*Plant, error))
	Body       string // method body code
}
