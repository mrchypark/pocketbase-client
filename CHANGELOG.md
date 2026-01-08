# Changelog

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
