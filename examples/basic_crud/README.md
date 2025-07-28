# Basic CRUD Operations Example (Type-Safe)

This example demonstrates type-safe CRUD operations using custom structs in the PocketBase Go client.

## Key Features

### 1. Type-Safe Struct Definition
```go
type Post struct {
    pocketbase.BaseModel
    Title   string `json:"title"`
    Content string `json:"content"`
}
```

### 2. Generic Record Service Usage
```go
postsService := pocketbase.NewRecordService[Post](client, "posts")
```

### 3. Type-Safe CRUD Operations
- **Create**: `postsService.Create(ctx, &post)`
- **List**: `postsService.GetList(ctx, options)`
- **Get One**: `postsService.GetOne(ctx, id, options)`
- **Update**: `postsService.Update(ctx, id, &post)`
- **Delete**: `postsService.Delete(ctx, id)`

## Advantages

### Type Safety
- Compile-time type error detection
- IDE autocomplete support
- Refactoring safety guarantee

### Code Readability
- Clear data structure
- Intuitive field access (`post.Title`)
- Separation of business logic and data models

## How to Run

```bash
# Set environment variables
export POCKETBASE_URL="http://127.0.0.1:8090"

# Run example
go run examples/basic_crud/main.go
```

## Required PocketBase Setup

1. Create `posts` collection
2. Add the following fields:
   - `title` (text)
   - `content` (text)
3. Create admin account (`admin@example.com` / `password123`)

## Related Examples

- [Direct Record Usage](../record_direct/) - Direct Record object usage for dynamic schemas
- [Type-Safe Generator](../type_safe_generator/) - Automatically generate types from schema