package generator

import (
	"fmt"
	"strings"
)

// ServiceGenerator handles generic service generation for collections
type ServiceGenerator struct{}

// NewServiceGenerator creates a new ServiceGenerator instance
func NewServiceGenerator() *ServiceGenerator {
	return &ServiceGenerator{}
}

// GenerateGenericService generates a generic service for a collection
func (g *ServiceGenerator) GenerateGenericService(collection CollectionData) string {
	var code strings.Builder

	// Service struct definition
	code.WriteString(g.generateServiceStruct(collection))
	code.WriteString("\n")

	// Constructor
	code.WriteString(g.generateServiceConstructor(collection))
	code.WriteString("\n")

	// Generic CRUD methods
	code.WriteString(g.generateGenericCRUDMethods(collection))

	// Field-specific generic getters
	code.WriteString(g.generateGenericFieldMethods(collection))

	return code.String()
}

// generateServiceStruct generates the service struct definition
func (g *ServiceGenerator) generateServiceStruct(collection CollectionData) string {
	serviceName := collection.StructName + "Service"

	return fmt.Sprintf(`// %s provides generic operations for %s collection
type %s struct {
	client pocketbase.RecordServiceAPI
}`, serviceName, collection.CollectionName, serviceName)
}

// generateServiceConstructor generates the service constructor
func (g *ServiceGenerator) generateServiceConstructor(collection CollectionData) string {
	serviceName := collection.StructName + "Service"

	return fmt.Sprintf(`// New%s creates a new %s instance
func New%s(client pocketbase.RecordServiceAPI) *%s {
	return &%s{client: client}
}`, serviceName, serviceName, serviceName, serviceName, serviceName)
}

// generateGenericCRUDMethods generates generic CRUD methods
func (g *ServiceGenerator) generateGenericCRUDMethods(collection CollectionData) string {
	var code strings.Builder
	serviceName := collection.StructName + "Service"
	structName := collection.StructName
	collectionName := collection.CollectionName

	// Get method
	code.WriteString(fmt.Sprintf(`
// Get retrieves a %s record by ID using generic approach
func (s *%s) Get(ctx context.Context, id string, expand ...string) (*%s, error) {
	record, err := s.client.Collection("%s").GetOne(id, &pocketbase.RecordGetOptions{
		Expand: strings.Join(expand, ","),
	})
	if err != nil {
		return nil, err
	}
	
	var result %s
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

`, structName, serviceName, structName, collectionName, structName))

	// List method
	code.WriteString(fmt.Sprintf(`// List retrieves %s records with generic filtering
func (s *%s) List(ctx context.Context, filter string, sort string, page int, perPage int, expand ...string) ([]*%s, error) {
	records, err := s.client.Collection("%s").GetList(page, perPage, &pocketbase.RecordListOptions{
		Filter: filter,
		Sort:   sort,
		Expand: strings.Join(expand, ","),
	})
	if err != nil {
		return nil, err
	}
	
	var results []*%s
	for _, record := range records.Items {
		var item %s
		if err := record.Unmarshal(&item); err != nil {
			return nil, err
		}
		results = append(results, &item)
	}
	
	return results, nil
}

`, structName, serviceName, structName, collectionName, structName, structName))

	// Create method
	code.WriteString(fmt.Sprintf(`// Create creates a new %s record using generic approach
func (s *%s) Create(ctx context.Context, data *%s) (*%s, error) {
	record, err := s.client.Collection("%s").Create(data)
	if err != nil {
		return nil, err
	}
	
	var result %s
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

`, structName, serviceName, structName, structName, collectionName, structName))

	// Update method
	code.WriteString(fmt.Sprintf(`// Update updates a %s record using generic approach
func (s *%s) Update(ctx context.Context, id string, data *%s) (*%s, error) {
	record, err := s.client.Collection("%s").Update(id, data)
	if err != nil {
		return nil, err
	}
	
	var result %s
	if err := record.Unmarshal(&result); err != nil {
		return nil, err
	}
	
	return &result, nil
}

`, structName, serviceName, structName, structName, collectionName, structName))

	// Delete method
	code.WriteString(fmt.Sprintf(`// Delete deletes a %s record
func (s *%s) Delete(ctx context.Context, id string) error {
	return s.client.Collection("%s").Delete(id)
}

`, structName, serviceName, collectionName))

	return code.String()
}

// generateGenericFieldMethods generates generic field accessor methods
func (g *ServiceGenerator) generateGenericFieldMethods(collection CollectionData) string {
	var code strings.Builder
	serviceName := collection.StructName + "Service"

	code.WriteString(fmt.Sprintf(`// Generic field accessors for %s
`, collection.StructName))

	for _, field := range collection.Fields {
		if field.IsPointer {
			// Generate ValueOr method for pointer fields
			code.WriteString(g.generateGenericValueOrMethod(serviceName, field))
		}

		// Generate generic getter method
		code.WriteString(g.generateGenericGetterMethod(serviceName, field))
	}

	return code.String()
}

// generateGenericValueOrMethod generates a generic ValueOr method for pointer fields
func (g *ServiceGenerator) generateGenericValueOrMethod(serviceName string, field FieldData) string {
	methodName := fmt.Sprintf("Get%sValueOr", field.GoName)

	return fmt.Sprintf(`// %s returns the value of %s field or default value using generics
func (s *%s) %s(record *%s, defaultValue %s) %s {
	if record.%s != nil {
		return *record.%s
	}
	return defaultValue
}

`, methodName, field.JSONName, serviceName, methodName,
		strings.TrimSuffix(serviceName, "Service"), field.BaseType, field.BaseType,
		field.GoName, field.GoName)
}

// generateGenericGetterMethod generates a generic getter method
func (g *ServiceGenerator) generateGenericGetterMethod(serviceName string, field FieldData) string {
	methodName := fmt.Sprintf("Get%s", field.GoName)
	structName := strings.TrimSuffix(serviceName, "Service")

	// For generic approach, we use Get[T] pattern
	if strings.Contains(field.GetterMethod, "Get[") {
		return fmt.Sprintf(`// %s gets %s field value using generic Get method
func (s *%s) %s(record *%s) %s {
	return record.%s(ctx, "%s")
}

`, methodName, field.JSONName, serviceName, methodName, structName, field.GoType, field.GetterMethod, field.JSONName)
	}

	// For non-generic fields, provide direct access
	return fmt.Sprintf(`// %s gets %s field value directly
func (s *%s) %s(record *%s) %s {
	return record.%s
}

`, methodName, field.JSONName, serviceName, methodName, structName, field.GoType, field.GoName)
}

// GenerateGenericServiceInterface generates a generic service interface
func (g *ServiceGenerator) GenerateGenericServiceInterface() string {
	return `// GenericRecordService provides generic CRUD operations for any collection
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
`
}

// GenerateAllServices generates services for all collections
func (g *ServiceGenerator) GenerateAllServices(collections []CollectionData, useGeneric bool) string {
	var code strings.Builder

	// Add imports
	code.WriteString(`import (
	"context"
	"strings"
	"github.com/pocketbase/pocketbase"
)

`)

	if useGeneric {
		// Generate generic interface and base service
		code.WriteString(g.GenerateGenericServiceInterface())
		code.WriteString("\n")
	}

	// Generate individual services for each collection
	for _, collection := range collections {
		if useGeneric {
			code.WriteString(g.generateGenericCollectionService(collection))
		} else {
			code.WriteString(g.GenerateGenericService(collection))
		}
		code.WriteString("\n")
	}

	return code.String()
}

// generateGenericCollectionService generates a collection-specific service using generics
func (g *ServiceGenerator) generateGenericCollectionService(collection CollectionData) string {
	serviceName := collection.StructName + "Service"
	structName := collection.StructName
	collectionName := collection.CollectionName

	return fmt.Sprintf(`// %s provides operations for %s collection using generics
type %s struct {
	*BaseGenericService[%s]
}

// New%s creates a new %s instance
func New%s(client pocketbase.RecordServiceAPI) *%s {
	return &%s{
		BaseGenericService: NewBaseGenericService[%s](client, "%s"),
	}
}
`, serviceName, collectionName, serviceName, structName, serviceName, serviceName, serviceName, serviceName, serviceName, structName, collectionName)
}
