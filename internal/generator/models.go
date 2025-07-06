package generator

type TemplateData struct {
	PackageName string
	Collections []CollectionData
}

type CollectionData struct {
	CollectionName string // 'posts'
	StructName     string // 'Post'
	Fields         []FieldData
}

type FieldData struct {
	JsonName  string // 'is_published'
	GoName    string // 'IsPublished'
	GoType    string // 'bool'
	OmitEmpty bool   // 'required: false'일 경우 true
}
