# pocketbase-client - Agent Guidelines

This document provides comprehensive guidelines for AI agents working with the pocketbase-client Go library.

## 🚀 Quick Start Commands

### Build & Test
```bash
# Build the generator and client
make gen          # Generate models from schema (uses pbc-gen)
go build -o bin/pbc-gen ./cmd/pbc-gen
go test ./...     # Run all tests
go test ./internal/generator -v  # Generator tests only

# Start PocketBase for integration testing
make pb_run       # Starts PocketBase on localhost:8090
make pb_clean      # Clean PocketBase data

# Generate models and test
./bin/pbc-gen -schema database/pb_schema.json -path models/generated.go -pkgname models
```

### Lint & Format
```bash
# Format all Go files
go fmt ./...
goimports -w .

# Run static analysis
go vet ./...
golangci-lint run

# Check for code smells
gosec ./...
```

### Run Single Test
```bash
# Run specific test
go test -run TestSpecificFunction ./internal/generator
go test -v ./internal/generator -run TestTemplateExecution

# Run performance benchmarks
go test -bench=. ./internal/generator
go test -bench=BenchmarkMarshalJSON ./...
```

## 📋 Code Style Guidelines

### Import Organization
```go
// Standard imports - group in logical order
import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/mrchypark/pocketbase-client"
)
```

### Naming Conventions
- **Variables**: `camelCase` for local variables, `PascalCase` for exported
- **Functions**: `PascalCase` for exported, `camelCase` for private
- **Types**: `PascalCase` for structs/interfaces, `camelCase` for generic parameters
- **Constants**: `PascalCase` with descriptive prefixes: `CollectionNameFieldName`
- **Files**: `snake_case.go` for files, `kebab-case` for commands

### Struct Design
```go
// Always include system fields in generated models
type ModelName struct {
    ID             string             `json:"id"`
    CollectionID   string             `json:"collectionId"`
    CollectionName string             `json:"collectionName"`
    Created        pocketbase.DateTime `json:"created"`
    Updated        pocketbase.DateTime `json:"updated"`
    
    // Generated fields follow
    FieldName  FieldType `json:"fieldName,omitempty"`
}
```

### Error Handling Patterns
```go
// Always wrap errors with context and details
return nil, NewGenerationError(ErrorTypeSchemaLoad,
    "failed to load schema", err).
    WithDetail("file_path", filePath).
    WithDetail("suggestion", "ensure file exists and is readable")

// Use structured errors, not fmt.Errorf for library code
if err != nil {
    return nil, fmt.Errorf("pocketbase: failed to marshal record: %w", err)
}
```

### Template Generation
- Use `{{range .Collections}}` for iterating collections
- Access fields via `{{.FieldName}}` syntax
- Include conditional blocks with `{{if .GenerateSomething}}`
- Use helper functions from `internal/generator/mapper.go`

## 🏗️ Architecture Overview

### Core Components
```
pocketbase-client/
├── client.go                 # Main client with auth management
├── records.go                # RecordServiceAPI + TypedRecordService
├── models.go                  # Base models and interfaces
├── generic_client.go          # Generic Service[T] implementation
├── cmd/pbc-gen/              # Code generator CLI
├── internal/generator/        # Generation logic
│   ├── schema.go            # PocketBase schema parsing
│   ├── mapper.go            # Type mapping utilities
│   ├── models.go            # Template data structures
│   ├── parser.go            # Schema loading logic
│   └── template.go.tpl       # Code generation template
└── examples/                  # Usage examples
```

### Key Interfaces
```go
// RecordServiceAPI - Base CRUD operations
type RecordServiceAPI interface {
    GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error)
    GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error)
    Create(ctx context.Context, collection string, body any) (*Record, error)
    // ... other CRUD operations
}

// RecordModel - Generated model interface
type RecordModel interface {
    BaseModel
    SetID(id string)
    SetCollectionID(id string)
    SetCollectionName(name string)
}

// Model - Generic constraint for type safety
type Model = Mappable
```

## 🔧 Generator Development

### Template Modification Guidelines
```go
// For type-safe field population
func Get{{.StructName}}(client pocketbase.RecordServiceAPI, id string, opts *pocketbase.GetOneOptions) (*{{.StructName}}, error) {
    r, err := client.GetOne(context.Background(), "{{.CollectionName}}", id, opts)
    if err != nil {
        return nil, err
    }
    result := &{{.StructName}}{}
    result.SetID(r.ID)
    result.SetCollectionID(r.CollectionID)
    result.SetCollectionName(r.CollectionName)

    // Marshal record to JSON and unmarshal into struct to populate all fields
    data, err := json.Marshal(r)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal record: %w", err)
    }
    if err := json.Unmarshal(data, result); err != nil {
        return nil, fmt.Errorf("failed to unmarshal into {{.StructName}}: %w", err)
    }

    return result, nil
}
```

### Schema Handling
```go
// Support both legacy and new PocketBase schema formats
func (cs *CollectionSchema) UnmarshalJSON(data []byte) error {
    type Alias CollectionSchema
    aux := &struct {
        SchemaRaw json.RawMessage `json:"schema"`
        FieldsRaw json.RawMessage `json:"fields"`
        *Alias
    }{
        Alias: (*Alias)(cs),
    }
    
    // Handle both format variants
    if len(aux.SchemaRaw) > 0 && string(aux.SchemaRaw) != "null" {
        return json.Unmarshal(aux.SchemaRaw, &cs.Fields)
    } else if len(aux.FieldsRaw) > 0 && string(aux.FieldsRaw) != "null" {
        return json.Unmarshal(aux.FieldsRaw, &cs.Fields)
    }
}
```

## 🧪 Testing Guidelines

### Unit Test Patterns
```go
// Table-driven tests for comprehensive coverage
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
    }{
        {
            name:     "valid text field",
            input:    FieldSchema{Type: "text", Required: false},
            expected: "string",
            wantErr:  false,
        },
        {
            name:     "required relation field",
            input:    FieldSchema{Type: "relation", Required: true},
            expected: "[]string",
            wantErr:  false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MapPbTypeToGoType(tt.input, !tt.input.Required)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, got)
            }
        })
    }
}
```

### Integration Test Setup
```go
// Use Makefile for PocketBase setup
func TestMain(m *testing.M) {
    // Start test PocketBase instance
    if err := setupTestDatabase(); err != nil {
        log.Fatal(err)
    }
    
    // Clean up after tests
    m.Cleanup(func() {
        cleanupTestDatabase()
    })
    
    os.Exit(m.Run())
}

// Helper for test data creation
func createTestUser(client *pocketbase.Client) *models.Users {
    user := models.NewUsers()
    user.SetName("Test User")
    user.SetEmail("test@example.com")
    return user
}
```

## ⚡ Performance Considerations

### JSON Marshal/Unmarshal
```go
// For high-performance record conversion
func convertRecord[T any](rec *Record) (*T, error) {
    var t T
    
    if model, ok := any(&t).(RecordModel); ok {
        // Fast path: direct field assignment for system fields
        model.SetID(rec.ID)
        model.SetCollectionID(rec.CollectionID)
        model.SetCollectionName(rec.CollectionName)
        
        // Bulk data transfer via JSON
        data, err := json.Marshal(rec)
        if err != nil {
            return nil, err
        }
        if err := json.Unmarshal(data, &t); err != nil {
            return nil, err
        }
        return &t, nil
    }
    
    return nil, fmt.Errorf("unsupported type: %T", t)
}
```

### Memory Management
```go
// Use object pooling for high-frequency operations
var recordPool = sync.Pool{
    New: func() interface{} {
        return new(Record)
    },
}

// Efficient slice operations
func processRecords(records []*Record) {
    // Pre-allocate slices with known capacity
    result := make([]*ProcessedRecord, 0, len(records))
    
    for _, rec := range records {
        processed := recordPool.Get().(*Record)
        *processed = *rec  // Copy instead of allocate
        recordPool.Put(processed)
        result = append(result, processed)
    }
    
    return result
}
```

## 🔍 Debugging Guidelines

### Generator Debugging
```bash
# Enable verbose generation
./pbc-gen -schema schema.json -path models.go -v

# Check template generation step-by-step
go test -run TestTemplateExecution -v

# Validate generated code compiles
go build -o /tmp/test models.go

# Inspect generated AST
go list -json . | grep '"Name":"Models"'
```

### Runtime Debugging
```go
// Enable debug logging in client
client := pocketbase.NewClient(
    "http://localhost:8090",
    pocketbase.WithDebugLogging(true),
)

// Use structured errors with context
if err != nil {
    return fmt.Errorf("operation failed: %w (context: %v)", err, ctx)
}

// Add debug markers
log.Printf("DEBUG: Processing record %+v", record)
```

## 🚨 Common Pitfalls

### Generator Issues
- **"Hollow" structs**: Fixed by JSON marshal/unmarshal in GetOne/GetList helpers
- **Missing imports**: Always add required packages in template header
- **Type mismatches**: Ensure generated types match template expectations
- **Schema evolution**: Handle new PocketBase schema versions gracefully

### Runtime Issues
- **Auth token leaks**: Use WithToken() instead of password strategies in production
- **Context cancellation**: Always pass context through call chains
- **Race conditions**: Use proper synchronization in concurrent scenarios
- **Memory leaks**: Ensure proper cleanup of resources

## 📚 Key References

### Internal Documentation
- `cmd/pbc-gen/template.go.tpl` - Master template for all generated code
- `internal/generator/schema.go` - PocketBase schema parsing and compatibility
- `internal/generator/mapper.go` - Type mapping and naming utilities
- `internal/generator/models.go` - Template data structures

### API Documentation
- [PocketBase API Docs](https://pocketbase.io/docs/api-overview/) - REST API reference
- [Go Client Docs](https://pkg.go.dev/github.com/mrchypark/pocketbase-client) - Generated Go docs
- [Examples Directory](./examples/) - Comprehensive usage examples

### Testing References
- [Testing Best Practices](https://golang.org/doc/testing) - Official Go testing guide
- [Table-driven Tests](https://dave.cheney.net/practical-go/presentations/63.html) - Test pattern guide
- [Testify Guidelines](https://pkg.go.dev/github.com/stretchr/testify) - Assertion library patterns

---

*This document is maintained alongside the codebase. Update it when making significant architectural changes or when adding new agent capabilities.*