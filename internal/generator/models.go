package generator

// TemplateData represents the data structure used for code generation templates.
// It contains all the information needed to generate Go code from PocketBase schemas,
// including the base collection models and any enhanced constructs (enums, relation
// types, and file types) selected via GenerateOptions.
type TemplateData struct {
	PackageName string // Go package name for generated code
	JSONLibrary string // JSON library import path (e.g., "encoding/json")
	Collections []CollectionData

	// Enhanced constructs. Only populated when the matching Generate* option is enabled.
	Enums         []EnumData
	RelationTypes []RelationTypeData
	FileTypes     []FileTypeData

	// Generate* flags control which enhanced constructs are emitted by the template.
	GenerateEnums     bool
	GenerateRelations bool
	GenerateFiles     bool

	// HasJSONFields is true when any generated field is a json.RawMessage.
	// It controls whether the JSON library import is emitted.
	HasJSONFields bool
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
	JSONName  string // JSON field name as it appears in PocketBase (e.g., 'is_published')
	GoName    string // Go field name in PascalCase (e.g., 'IsPublished')
	GoType    string // Go type for the field (e.g., 'bool')
	StructTag string // Preformatted struct tag (e.g., `json:"is_published,omitempty"`)
	OmitEmpty bool   // Whether to add omitempty tag for optional fields
	IsPointer bool   // Whether this field is a pointer type (for ValueOr method generation)
	BaseType  string // Base type without pointer (e.g., 'string' for '*string')
}

// EnumData represents data for generating enum constants from select fields.
// It contains all information needed to generate type-safe enum constants.
type EnumData struct {
	CollectionName string         // 컬렉션명 (예: devices)
	FieldName      string         // 필드명 (예: type)
	EnumTypeName   string         // enum 타입명 (예: DeviceType)
	Constants      []ConstantData // 상수 목록
}

// ConstantData represents a single enum constant with its name and value.
type ConstantData struct {
	Name  string // 상수명 (예: DeviceTypeM2)
	Value string // 상수값 (예: "m2")
}

// RelationTypeData represents data for generating relation types from relation fields.
// It contains metadata for creating type-safe relation helpers.
type RelationTypeData struct {
	TypeName         string // 관계 타입명 (예: PlantRelation)
	TargetCollection string // 대상 컬렉션명 (예: plants)
	TargetTypeName   string // 대상 타입명 (예: Plant)
	IsMulti          bool   // 다중 관계 여부
}

// FileTypeData represents data for generating file types from file fields.
// It contains metadata for creating file reference helpers with URL generation.
type FileTypeData struct {
	TypeName       string   // 파일 타입명 (예: ImageFile)
	IsMulti        bool     // 다중 파일 여부
	HasThumbnails  bool     // 썸네일 지원 여부
	ThumbnailSizes []string // 썸네일 크기 목록
}
