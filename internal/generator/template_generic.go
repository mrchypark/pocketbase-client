package generator

// GenericServiceTemplate provides template for generating generic services
const GenericServiceTemplate = `package {{.PackageName}}

import (
	"context"
	"strings"
	{{if .JSONLibrary}}"{{.JSONLibrary}}"{{end}}
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/tools/types"
)

{{if .UseGeneric}}
// GenericRecordService provides generic CRUD operations for any collection
type GenericRecordService[T any] interface {
	Get(ctx context.Context, id string, expand ...string) (*T, error)
	List(ctx context.Context, filter string, sort string, page int, perPage int, expand ...string) ([]*T, error)
	Create(ctx context.Context, data *T) (*T, error)
	Update(ctx context.Context, id string, data *T) (*T, error)
	Delete(ctx context.Context, id string) error
}

// BaseGenericService provides a base implementation for generic record services
type BaseGenericService[T any] struct {
	client         pocketbase.RecordServiceAPI
	collectionName string
}

// NewBaseGenericService creates a new BaseGenericService instance
func NewBaseGenericService[T any](client pocketbase.RecordServiceAPI, collectionName string) *BaseGenericService[T] {
	return &BaseGenericService[T]{
		client:         client,
		collectionName: collectionName,
	}
}

// Get retrieves a record by ID using generic approach
func (s *BaseGenericService[T]) Get(ctx context.Context, id string, expand ...string) (*T, error) {
	record, err := s.client.Collection(s.collectionName).GetOne(id, &pocketbase.RecordGetOptions{
		Expand: strings.Join(expand, ","),
	})
	if err != nil {
		return nil, err
	}
	
	var result T
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// List retrieves records with generic filtering
func (s *BaseGenericService[T]) List(ctx context.Context, filter string, sort string, page int, perPage int, expand ...string) ([]*T, error) {
	records, err := s.client.Collection(s.collectionName).GetList(page, perPage, &pocketbase.RecordListOptions{
		Filter: filter,
		Sort:   sort,
		Expand: strings.Join(expand, ","),
	})
	if err != nil {
		return nil, err
	}
	
	var results []*T
	for _, record := range records.Items {
		var item T
		if err := record.Unmarshal(&item); err != nil {
			return nil, err
		}
		results = append(results, &item)
	}
	
	return results, nil
}

// Create creates a new record using generic approach
func (s *BaseGenericService[T]) Create(ctx context.Context, data *T) (*T, error) {
	record, err := s.client.Collection(s.collectionName).Create(data)
	if err != nil {
		return nil, err
	}
	
	var result T
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// Update updates a record using generic approach
func (s *BaseGenericService[T]) Update(ctx context.Context, id string, data *T) (*T, error) {
	record, err := s.client.Collection(s.collectionName).Update(id, data)
	if err != nil {
		return nil, err
	}
	
	var result T
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// Delete deletes a record
func (s *BaseGenericService[T]) Delete(ctx context.Context, id string) error {
	return s.client.Collection(s.collectionName).Delete(id)
}

{{range .Collections}}
{{$collection := .}}
// {{.StructName}}Service provides operations for {{.CollectionName}} collection using generics
type {{.StructName}}Service struct {
	*BaseGenericService[{{.StructName}}]
}

// New{{.StructName}}Service creates a new {{.StructName}}Service instance
func New{{.StructName}}Service(client pocketbase.RecordServiceAPI) *{{.StructName}}Service {
	return &{{.StructName}}Service{
		BaseGenericService: NewBaseGenericService[{{.StructName}}](client, "{{.CollectionName}}"),
	}
}

{{if .Fields}}
// Field-specific helper methods for {{.StructName}}
{{range .Fields}}
{{if .IsPointer}}
// Get{{.GoName}}ValueOr returns the value of {{.JSONName}} field or default value
func (s *{{$collection.StructName}}Service) Get{{.GoName}}ValueOr(record *{{$collection.StructName}}, defaultValue {{.BaseType}}) {{.BaseType}} {
	if record.{{.GoName}} != nil {
		return *record.{{.GoName}}
	}
	return defaultValue
}
{{end}}

// Get{{.GoName}} gets {{.JSONName}} field value
func (s *{{$collection.StructName}}Service) Get{{.GoName}}(record *{{$collection.StructName}}) {{.GoType}} {
	return record.{{.GoName}}
}
{{end}}
{{end}}
{{end}}

{{else}}
// Non-generic service implementations
{{range .Collections}}
{{$collection := .}}
// {{.StructName}}Service provides operations for {{.CollectionName}} collection
type {{.StructName}}Service struct {
	client pocketbase.RecordServiceAPI
}

// New{{.StructName}}Service creates a new {{.StructName}}Service instance
func New{{.StructName}}Service(client pocketbase.RecordServiceAPI) *{{.StructName}}Service {
	return &{{.StructName}}Service{client: client}
}

// Get retrieves a {{.StructName}} record by ID
func (s *{{.StructName}}Service) Get(ctx context.Context, id string, expand ...string) (*{{.StructName}}, error) {
	record, err := s.client.Collection("{{.CollectionName}}").GetOne(id, &pocketbase.RecordGetOptions{
		Expand: strings.Join(expand, ","),
	})
	if err != nil {
		return nil, err
	}
	
	var result {{.StructName}}
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// List retrieves {{.StructName}} records with filtering
func (s *{{.StructName}}Service) List(ctx context.Context, filter string, sort string, page int, perPage int, expand ...string) ([]*{{.StructName}}, error) {
	records, err := s.client.Collection("{{.CollectionName}}").GetList(page, perPage, &pocketbase.RecordListOptions{
		Filter: filter,
		Sort:   sort,
		Expand: strings.Join(expand, ","),
	})
	if err != nil {
		return nil, err
	}
	
	var results []*{{.StructName}}
	for _, record := range records.Items {
		var item {{.StructName}}
		if err := record.Unmarshal(&item); err != nil {
			return nil, err
		}
		results = append(results, &item)
	}
	
	return results, nil
}

// Create creates a new {{.StructName}} record
func (s *{{.StructName}}Service) Create(ctx context.Context, data *{{.StructName}}) (*{{.StructName}}, error) {
	record, err := s.client.Collection("{{.CollectionName}}").Create(data)
	if err != nil {
		return nil, err
	}
	
	var result {{.StructName}}
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// Update updates a {{.StructName}} record
func (s *{{.StructName}}Service) Update(ctx context.Context, id string, data *{{.StructName}}) (*{{.StructName}}, error) {
	record, err := s.client.Collection("{{.CollectionName}}").Update(id, data)
	if err != nil {
		return nil, err
	}
	
	var result {{.StructName}}
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

// Delete deletes a {{.StructName}} record
func (s *{{.StructName}}Service) Delete(ctx context.Context, id string) error {
	return s.client.Collection("{{.CollectionName}}").Delete(id)
}

{{if .Fields}}
// Field-specific getter methods for {{.StructName}}
{{range .Fields}}
{{if .IsPointer}}
// Get{{.GoName}}ValueOr returns the value of {{.JSONName}} field or default value
func (s *{{$collection.StructName}}Service) Get{{.GoName}}ValueOr(record *{{$collection.StructName}}, defaultValue {{.BaseType}}) {{.BaseType}} {
	if record.{{.GoName}} != nil {
		return *record.{{.GoName}}
	}
	return defaultValue
}
{{end}}

// Get{{.GoName}} gets {{.JSONName}} field value using {{.GetterMethod}}
func (s *{{$collection.StructName}}Service) Get{{.GoName}}(record *{{$collection.StructName}}) {{.GoType}} {
	{{if eq .GetterMethod "Get[string]" "Get[float64]" "Get[bool]" "Get[types.DateTime]" "Get[json.RawMessage]" "Get[[]string]" "Get[any]"}}
	// Generic getter method would be: record.{{.GetterMethod}}(ctx, "{{.JSONName}}")
	// For now, return direct field access
	return record.{{.GoName}}
	{{else}}
	// Traditional getter method: {{.GetterMethod}}
	return record.{{.GoName}}
	{{end}}
}
{{end}}
{{end}}
{{end}}
{{end}}
`

// StructTemplate provides template for generating struct definitions
const StructTemplate = `package {{.PackageName}}

import (
	{{if .JSONLibrary}}"{{.JSONLibrary}}"{{end}}
	"github.com/pocketbase/pocketbase/tools/types"
)

// BaseModel contains common fields for all PocketBase records
type BaseModel struct {
	ID             string ` + "`json:\"id\"`" + `
	CollectionID   string ` + "`json:\"collectionId\"`" + `
	CollectionName string ` + "`json:\"collectionName\"`" + `
}

{{if eq .SchemaVersion 1}}
// BaseDateTime contains timestamp fields for legacy schema
type BaseDateTime struct {
	Created types.DateTime ` + "`json:\"created\"`" + `
	Updated types.DateTime ` + "`json:\"updated\"`" + `
}
{{end}}

{{range .Collections}}
// {{.StructName}} represents a record from the {{.CollectionName}} collection
type {{.StructName}} struct {
	BaseModel
	{{if .UseTimestamps}}BaseDateTime{{end}}
	{{range .Fields}}{{.GoName}} {{.GoType}} ` + "`json:\"{{.JSONName}}{{if .OmitEmpty}},omitempty{{end}}\"`" + `
	{{end}}
}

// TableName returns the table name for {{.StructName}}
func ({{.StructName | printf "%.1s" | ToLower}}) TableName() string {
	return "{{.CollectionName}}"
}

{{range .Fields}}
{{if .IsPointer}}
// {{.GoName}}ValueOr returns the value of {{.JSONName}} or the default value if nil
func ({{$.StructName | printf "%.1s" | ToLower}} *{{$.StructName}}) {{.GoName}}ValueOr(defaultValue {{.BaseType}}) {{.BaseType}} {
	if {{$.StructName | printf "%.1s" | ToLower}}.{{.GoName}} != nil {
		return *{{$.StructName | printf "%.1s" | ToLower}}.{{.GoName}}
	}
	return defaultValue
}
{{end}}
{{end}}
{{end}}
`
