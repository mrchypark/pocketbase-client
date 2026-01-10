# Changelog

## [v0.3.1] - 2026-01-10

### Breaking Changes

- **pbc-gen Template Redesign**: Generated models now use direct struct fields instead of embedding `pocketbase.Record`
  - Models implement `RecordModel` interface for seamless client integration
  - Fields are directly accessible (e.g., `post.Title` instead of `post.GetString("title")`)
  - Setter methods generated for all fields (e.g., `post.SetTitle("value")`)
  - Pointer fields accept value types in setters for convenience (e.g., `post.SetContent("text")` instead of `post.Content = &text`)

### Features

- **RecordModel Interface**: New interface for generated record types
  - Extends `BaseModel` with `SetID()`, `SetCollectionID()`, `SetCollectionName()` methods
  - Enables automatic conversion in `TypedRecordService`
- **Setter Methods for All Fields**: Generated models include setter methods
  - Pointer fields: `SetField(v BaseType)` - automatically converts to pointer
  - Value fields: `SetField(v Type)` - direct assignment
- **Service Factory Functions**: Each collection gets `New{Collection}Service(client)` factory

### Improvements

- `convertRecord[T]` now supports `RecordModel` interface for seamless type conversion
- Generated code includes `var _ pocketbase.RecordModel = (*Type)(nil)` for compile-time interface verification
- Template tests updated for new structure

### Migration Guide

#### Before (v0.3.0)
```go
type Post struct {
    pocketbase.Record
}

func (p *Post) Title() string { return p.GetString("title") }
func (p *Post) SetTitle(v string) { p.Set("title", v) }

// Usage
title := post.Title()
post.SetTitle("New Title")
```

#### After (v0.3.1)
```go
type Post struct {
    ID             string         `json:"id"`
    CollectionID   string         `json:"collectionId"`
    CollectionName string         `json:"collectionName"`
    Created        types.DateTime `json:"created"`
    Updated        types.DateTime `json:"updated"`
    Title          string         `json:"title"`
    Content        *string        `json:"content,omitempty"`
}

// Usage - direct field access!
title := post.Title
post.Title = "New Title"
post.SetTitle("New Title")      // Also works
post.SetContent("My content")   // Pointer field - no &v needed!
```

## [Unreleased]

### Breaking Changes

- `BaseModel` is now an interface instead of a struct. Models should embed fields directly or implement `GetID()` and `GetCollectionName()` methods.
- Generated models from `pbc-gen` may need to be regenerated for compatibility.

### Features

- **Type-Safe Generic Services**: New `TypedRecordService[T]` provides compile-time type-safe CRUD operations
  - `NewTypedRecordService[T](client, collection)` creates a typed service
  - `GetOne()`, `Create()`, `Update()` return `*T` instead of `*Record`
  - `GetList()` returns `*TypedListResult[T]` with typed `Items`
  - `GetAll()` performs auto-pagination with typed results
- **BaseModel Interface**: New interface for type-safe model operations
  - `GetID() string` - returns the model ID
  - `GetCollectionName() string` - returns the collection name
- Added `examples/generic_usage/main.go` demonstrating new patterns

### Improvements

- Cleaned up documentation comments throughout the codebase
- Removed duplicate package documentation
- Fixed `NewDeleteRequest` to properly set `Body: nil`

### Migration Guide

#### Before (Legacy API)
```go
service := pocketbase.NewRecordService[Post](client, "posts")
record, err := service.GetOne(ctx, "posts", "id", nil)
title := record.GetString("title")
```

#### After (New Generic API)
```go
postService := pocketbase.NewTypedRecordService[Post](client, "posts")
post, err := postService.GetOne(ctx, "id", nil)
title := post.Title() // Using generated getter
```

#### Model Definition

**Before:**
```go
type Post struct {
    pocketbase.BaseModel
    Title string `json:"title"`
}
```

**After:**
```go
type Post struct {
    pocketbase.Record
    Title string `json:"title"`
}

func (p *Post) Title() string { return p.GetString("title") }
func (p *Post) SetTitle(v string) { p.Set("title", v) }
```

Or use `pbc-gen` to regenerate models automatically.
