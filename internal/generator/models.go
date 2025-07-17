package generator

type TemplateData struct {
	PackageName string
	JsonLibrary string
	Collections []CollectionData

	// Enhanced 기능을 위한 필드들 (기본값은 빈 슬라이스)
	Enums         []EnumData         `json:"enums,omitempty"`
	RelationTypes []RelationTypeData `json:"relationTypes,omitempty"`
	FileTypes     []FileTypeData     `json:"fileTypes,omitempty"`
}

type CollectionData struct {
	CollectionName string // 'posts'
	StructName     string // 'Post'
	Fields         []FieldData
}

type FieldData struct {
	JsonName     string // 'is_published'
	GoName       string // 'IsPublished'
	GoType       string // 'bool'
	OmitEmpty    bool   //
	GetterMethod string // Getter method name (e.g., GetString, GetBool)
}

// EnhancedFieldInfo contains additional information for enhanced code generation
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

// EnhancedTemplateData extends TemplateData with enhanced generation features
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

// EnumData represents data for generating enum constants
type EnumData struct {
	CollectionName string         // 컬렉션명 (예: devices)
	FieldName      string         // 필드명 (예: type)
	EnumTypeName   string         // enum 타입명 (예: DeviceType)
	Constants      []ConstantData // 상수 목록
}

// ConstantData represents a single enum constant
type ConstantData struct {
	Name  string // 상수명 (예: DeviceTypeM2)
	Value string // 상수값 (예: "m2")
}

// RelationTypeData represents data for generating relation types
type RelationTypeData struct {
	TypeName         string       // 관계 타입명 (예: PlantRelation)
	TargetCollection string       // 대상 컬렉션명 (예: plants)
	TargetTypeName   string       // 대상 타입명 (예: Plant)
	IsMulti          bool         // 다중 관계 여부
	Methods          []MethodData // 생성할 메서드 목록
}

// FileTypeData represents data for generating file types
type FileTypeData struct {
	TypeName       string       // 파일 타입명 (예: ImageFile)
	IsMulti        bool         // 다중 파일 여부
	HasThumbnails  bool         // 썸네일 지원 여부
	ThumbnailSizes []string     // 썸네일 크기 목록
	Methods        []MethodData // 생성할 메서드 목록
}

// MethodData represents a method to be generated
type MethodData struct {
	Name       string // 메서드명 (예: Load, URL)
	ReturnType string // 반환 타입 (예: (*Plant, error))
	Body       string // 메서드 본문 코드
}
