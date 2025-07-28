# Type-Safe Generator Example

This example demonstrates the type-safe code generation features of the PocketBase Go Client.

## 🚀 How to Run

### 1. Generate Models
```bash
# Run from current directory
go run ../../cmd/pbc-gen -schema ./schema.json -path ./models.gen.go -pkgname main
```

### 2. Run Example
```bash
# PocketBase server must be running on localhost:8090
go mod tidy
go run .
```

## 📋 Included Features

### ✅ Complete Type Safety
- All fields mapped to appropriate Go types
- Compile-time type checking
- IDE auto-completion support

### 🔧 CRUD Operations
- `Create`: Create new records
- `GetOne`: Retrieve single record
- `GetList`: Paginated list retrieval
- `GetAll`: Get all records (automatic pagination)
- `Update`: Update records
- `Delete`: Delete records

### 🛡️ Safe Field Access
```go
// Required fields - direct access
fmt.Println(record.TextRequired)

// Optional fields - nil check
if record.TextOptional != nil {
    fmt.Println(*record.TextOptional)
}

// Array fields - safe length check
fmt.Printf("Item count: %d\n", len(record.SelectMultiRequired))
```

### 📊 Supported Field Types

| PocketBase Type | Go Type (Required) | Go Type (Optional) |
|----------------|-------------------|-------------------|
| text | `string` | `*string` |
| number | `float64` | `*float64` |
| bool | `bool` | `*bool` |
| email | `string` | `*string` |
| url | `string` | `*string` |
| date | `types.DateTime` | `*types.DateTime` |
| select (single) | `string` | `*string` |
| select (multi) | `[]string` | `[]string` |
| json | `json.RawMessage` | `json.RawMessage` |
| file (single) | `string` | `*string` |
| file (multi) | `[]string` | `[]string` |
| relation (single) | `string` | `*string` |
| relation (multi) | `[]string` | `[]string` |

## 🎯 Usage Patterns

### Basic Usage
```go
// Create service
service := NewAllTypesService(client)

// Create new record
record := &AllTypes{
    TextRequired: "Required value",
    NumberRequired: 42.0,
    BoolRequired: true,
}

// Create
created, err := service.Create(ctx, record)
if err != nil {
    log.Fatal(err)
}

// Retrieve
found, err := service.GetOne(ctx, created.ID, nil)
if err != nil {
    log.Fatal(err)
}

// Update
found.TextRequired = "New value"
updated, err := service.Update(ctx, found.ID, found)
```

### Advanced Queries
```go
// Filtering and sorting
list, err := service.GetList(ctx, &pocketbase.ListOptions{
    Filter:  "text_required != '' && number_required > 0",
    Sort:    "-created,text_required",
    PerPage: 20,
    Page:    1,
})

// All records (automatic pagination)
all, err := service.GetAll(ctx, &pocketbase.ListOptions{
    Filter: "bool_required = true",
})
```

## 🔧 Schema Structure

This example uses the following collections:

- **all_types**: Test collection containing all PocketBase field types
- **related_collection**: Simple collection for relation testing

## 📝 Notes

- PocketBase server must be running
- Appropriate collections and fields must be configured
- If authentication is required, set authentication info on the client