package generator

type TemplateData struct {
	PackageName string
	JsonLibrary string
	Collections []CollectionData
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
	GetterMethod string // Getter 메서드 이름 (예: GetString, GetBool)
}
