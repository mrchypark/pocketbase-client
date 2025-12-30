package generator

// TemplateData represents the data structure used for code generation templates.
// It contains all the information needed to generate Go code from PocketBase schemas.
type TemplateData struct {
	PackageName string // Go package name for generated code
	JSONLibrary string // JSON library import path (e.g., "encoding/json")
	Collections []CollectionData

	// Enhanced 기능을 위한 필드들 (기본값은 빈 슬라이스)
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
	StructTag    string // Preformatted struct tag (e.g., `json:"is_published,omitempty"`)
	OmitEmpty    bool   // Whether to add omitempty tag for optional fields
	GetterMethod string // Getter method name (e.g., GetString, GetBool)
	IsPointer    bool   // Whether this field is a pointer type (for ValueOr method generation)
	BaseType     string // Base type without pointer (e.g., 'string' for '*string')
	ToMapBlock   string // Preformatted ToMap field block
	ValueOrBlock string // Preformatted ValueOr method block (empty for non-pointer fields)
}

// EnhancedFieldInfo contains additional information for enhanced code generation.
// It extends FieldSchema with metadata needed for generating enums, relations, and file types.
type EnhancedFieldInfo struct {
	FieldSchema

	// Select 필드용 - enum 생성을 위한 정보
	EnumValues   []string // select 필드의 values 옵션에서 추출
	EnumTypeName string   // 생성될 enum 타입명 (예: DeviceType)

	// Relation 필드용 - 관계 타입 생성을 위한 정보
	TargetCollection string // 참조하는 컬렉션명
	RelationTypeName string // 생성될 관계 타입명 (예: PlantRelation)
	IsMultiRelation  bool   // 다중 관계 여부 (maxSelect > 1)

	// File 필드용 - 파일 타입 생성을 위한 정보
	FileTypeName   string   // 생성될 파일 타입명 (예: ImageFile)
	IsMultiFile    bool     // 다중 파일 여부 (maxSelect > 1)
	HasThumbnails  bool     // 썸네일 지원 여부
	ThumbnailSizes []string // 썸네일 크기 목록
}

// EnhancedTemplateData extends TemplateData with enhanced generation features.
// It includes additional data for generating enums, relations, and file types.
type EnhancedTemplateData struct {
	TemplateData

	// 새로운 기능들
	Enums         []EnumData         // 생성할 enum 데이터
	RelationTypes []RelationTypeData // 생성할 관계 타입 데이터
	FileTypes     []FileTypeData     // 생성할 파일 타입 데이터

	// 생성 옵션 - CLI 플래그로 제어
	GenerateEnums     bool // enum 생성 여부
	GenerateRelations bool // 관계 타입 생성 여부
	GenerateFiles     bool // 파일 타입 생성 여부
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
	TypeName         string       // 관계 타입명 (예: PlantRelation)
	TargetCollection string       // 대상 컬렉션명 (예: plants)
	TargetTypeName   string       // 대상 타입명 (예: Plant)
	IsMulti          bool         // 다중 관계 여부
	Methods          []MethodData // 생성할 메서드 목록
}

// FileTypeData represents data for generating file types from file fields.
// It contains metadata for creating file reference helpers with URL generation.
type FileTypeData struct {
	TypeName       string       // 파일 타입명 (예: ImageFile)
	IsMulti        bool         // 다중 파일 여부
	HasThumbnails  bool         // 썸네일 지원 여부
	ThumbnailSizes []string     // 썸네일 크기 목록
	Methods        []MethodData // 생성할 메서드 목록
}

// MethodData represents a method to be generated for enhanced types.
// It contains the method signature and implementation details.
type MethodData struct {
	Name       string // 메서드명 (예: Load, URL)
	ReturnType string // 반환 타입 (예: (*Plant, error))
	Body       string // 메서드 본문 코드
}
