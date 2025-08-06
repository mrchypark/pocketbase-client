# Type-Safe Generator Example

This example demonstrates the type-safe code generation features of the PocketBase Go client, showing how to generate Go structs from PocketBase schema and use them with compile-time type safety.

## 🚀 Quick Start

```bash
# Start PocketBase server (from project root)
make pb_run

# Set up admin account at http://localhost:8090/_/
# Use: admin@example.com / 1q2w3e4r5t

# Run the example
cd examples/type_safe_generator
go run .
```

## 🎯 What You'll Learn

### Type Safety Benefits
- **Compile-time validation**: Field names and types verified at compile time
- **IDE support**: Auto-completion, refactoring, type hints
- **Runtime safety**: Prevents nil pointer dereference and type casting errors

### Code Generation Workflow
```bash
# 1. Export PocketBase schema
curl http://localhost:8090/api/collections > pb_schema.json

# 2. Generate type-safe Go models
go run ./cmd/pbc-gen -schema ./pb_schema.json -path ./models.gen.go

# 3. Use generated models
go run main.go
```

## 📋 Example Output

When you run the example, you'll see:

1. **Type Safety Concepts**: Comparison between old map-based and new type-safe approaches
2. **Generated Model Structure**: Overview of generated Go structs and service constructors
3. **Live CRUD Demo**: Real database operations using RelatedCollection
4. **Type Safety Benefits**: Compile-time validation examples
5. **Development Workflow**: Step-by-step process explanation
6. **Real-World Patterns**: Practical usage scenarios including batch processing and error handling

## 🏗️ Generated Code Features

### Type-Safe Service Constructors
```go
// Old way (not type-safe)
recordService := client.Records("related_collection")

// New way (type-safe)
service := NewRelatedCollectionService(client)
```

### Compile-Time Type Validation
```go
// ✅ Correct usage - compiles successfully
record := &RelatedCollection{
    Name: "example record",
}

// ❌ Wrong usage - compile error
record := &RelatedCollection{
    Name: 123,           // Compile error: type mismatch
    WrongField: "value", // Compile error: field doesn't exist
}
```

### Safe Optional Field Handling
```go
// Optional fields use pointers
var optionalText *string
if condition {
    text := "optional value"
    optionalText = &text
}

// Safe access
if record.TextOptional != nil {
    fmt.Println(*record.TextOptional)
}
```

## 🛠️ File Structure

```
examples/type_safe_generator/
├── pb_schema.json          # PocketBase schema definition
├── models.gen.go           # Generated type-safe models
├── main.go                 # Example demonstration code
└── README.md              # This file
```

## 💡 Real-World Usage

### Schema Change Workflow
```bash
# After changing schema in PocketBase admin UI
curl http://localhost:8090/api/collections > pb_schema.json
go run ./cmd/pbc-gen -schema ./pb_schema.json -path ./models.gen.go
go build .  # Verify no type errors
```

### Team Development Benefits
- **Consistency**: All developers use identical type definitions
- **Safety**: Schema changes reveal affected code at compile time
- **Productivity**: IDE support accelerates development

### CI/CD Integration
```yaml
- name: Generate PocketBase models
  run: |
    go run ./cmd/pbc-gen -schema ./pb_schema.json -path ./models.gen.go
    git diff --exit-code models.gen.go || (echo "Models need update" && exit 1)
```

## 🔧 Troubleshooting

**Authentication Failed**
- Ensure PocketBase server is running at `http://localhost:8090`
- Verify admin credentials: `admin@example.com` / `1q2w3e4r5t`

**Collection Not Found**
- Check that collections exist in PocketBase admin UI
- Verify schema file matches actual PocketBase collections

**Sort Field Errors**
- The example uses `sort=-id` since `related_collection` doesn't have `created` field
- For collections with `created`/`updated` fields, use `sort=-created`

**JSON Parsing Errors**
- Update schema file to latest state: `curl http://localhost:8090/api/collections > pb_schema.json`
- Regenerate models: `go run ./cmd/pbc-gen -schema ./pb_schema.json -path ./models.gen.go`

## 📚 Key Concepts Demonstrated

The example shows how type-safe code generation provides:

- **Compile-Time Safety**: Field name typos and type mismatches caught before runtime
- **IDE Integration**: Auto-completion, type hints, and refactoring support
- **Runtime Protection**: Prevents common errors like nil pointer dereference
- **Development Efficiency**: Faster coding with better tooling support
- **Team Collaboration**: Consistent type definitions across the team

This example effectively demonstrates why type-safe PocketBase development is superior to traditional map-based approaches, making it an excellent starting point for understanding the library's code generation capabilities.