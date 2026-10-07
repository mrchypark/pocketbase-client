# Changelog

## [v0.5.0] - 2026-10-07

### Breaking Changes

- **Collection API**: replace `Collection.Schema` with `Fields []SchemaField`, string `Indexes` with `[]string`, and string access rules with `*string`. Nil rules serialize as `null` (locked); pointers to empty strings mean public access. Collection ID/timestamps are concrete fields rather than an embedded `BaseModel`. Collection and field `Options` now serialize as flattened properties. Replace removed `SchemaField.Unique` with a unique SQL index in `Collection.Indexes`.
- **Generated auth models**: regenerate models from modern PocketBase schemas. Readable auth email, emailVisibility, and verified fields are retained as pointers for sparse PATCH requests. `Password`, `PasswordConfirm`, and `OldPassword` are write-only pointers with `json:"-"`; setters and `ToMap` send them only when set. Ordinary required base fields retain their existing behavior.
- **Generator system policy**: all system collections are skipped. Custom autodate fields are decoded but excluded from writes; standard ID, metadata, created, and updated fields remain declared once and excluded from `ToMap`. Relation helpers are generated only for visible fields whose targets have generated models.

### Fixes

- Align collection field metadata, indexes, nullable rules, and flattened auth/view options with PocketBase v0.39.10. Import sends `{collections, deleteMissing}` and accepts the server's 204 response, returning no collection items.
- Adapt the deprecated Admin service to `_superusers` record endpoints. `WithAdminPassword` returns both modern `Record` and legacy `Admin` data.
- Isolate authentication transport wrapping from caller-owned HTTP clients and clone requests before injecting tokens. Shared password authentication respects each caller's cancellation without canceling other waiters, with a bounded refresh timeout.
- Normalize single-object and array relation expansions to `Record.Expand` slices, clear stale expansions on reuse, and report malformed expansion data.
- Continue automatic pagination using the server's effective page size when requested `perPage` exceeds its cap, including `skipTotal` queries.
- Delete single and multiple files with the server's `field-` modifier, avoiding a read/rewrite race; reject empty filenames.
- Re-register realtime subscriptions after reconnects without blocking on repeated connection notifications; suppress expected cancellation errors.
- Preserve explicit zero PATCH values and omitted auth fields in regenerated models. Retain multi-relation helpers when single and multiple relations share a target; avoid references to skipped system models.
- Verify release checksums before replacing `pbc-gen`; support Linux/macOS installation. Download and verify pinned PocketBase in temporary extraction directories instead of overwriting or deleting repository documentation. Format checks no longer modify source files.
- Replace deprecated GoReleaser archive keys with `ids` and `formats`, preserving raw binary asset names and checksum output.
- Clarify that direct Users OAuth2, OTP, and refresh calls require `UseAuthResponse` to install the returned authentication state.

### Validation

- Full Go tests, race tests, vet, and build passed. Isolated PocketBase v0.39.10 contract tests cover authentication, Admin compatibility, collection rules/import/indexes, generated auth CRUD and sparse/zero PATCH, relation expansion, pagination beyond 1000 records, and single/multiple file deletion.
- Installer regressions verify successful installation and preservation of an existing binary on mismatched or missing checksums. CI runs both live contract tests against the pinned server; local contract tests skip when `POCKETBASE_BIN` is unset.

## [v0.4.0] - 2026-08-11

### Breaking Changes

- **DateTime/GeoPoint types**: generated `date`/`autodate` fields changed from `github.com/pocketbase/pocketbase/tools/types.DateTime` to the client's own `pocketbase.DateTime`; `geoPoint` fields changed to `pocketbase.GeoPoint`. The JSON wire format is identical — only the type name and import change. The client and generated code no longer require the PocketBase module.
- **Service[T] removed**: the older generic `Service[T]`/`NewService` API was removed. Use `TypedRecordService[T]`/`NewTypedRecordService` instead.
- **GetX/GetXList helpers removed**: top-level `Get{{X}}`/`Get{{X}}List` helpers were removed. Use the typed service methods directly (`NewXService`). Relation `Load` now takes `*pocketbase.Client` instead of `RecordServiceAPI`.
- **Legacy schema format dropped**: the `schema` key (pre-v0.23) and nested `options` objects are no longer accepted. Re-export your schema from PocketBase v0.23+.
- **Go 1.26+ required**.

### Features

- **Codegen single pipeline**: `BuildTemplateData` is now the sole entry point for schema-to-template-data. Dead code removed (~2.4k lines).
- **Zero-cost typed reads**: `TypedRecordService` decodes raw HTTP responses directly into `T` (single `json.Unmarshal` pass) via new `Client.SendRaw`. Removed the old `Record` marshal→unmarshal round-trip.
- **PocketBase dependency removed**: self-contained `pocketbase.DateTime` (wire-compatible with `tools/types.DateTime`) and `pocketbase.GeoPoint`. Generated code depends only on the client library — no `github.com/pocketbase/pocketbase` in `go.mod`.
- **PocketBase v0.39.10 support**: `geoPoint` field type mapped to `pocketbase.GeoPoint`; `help` key parsed from field objects.
- **Schema parsing**: `LoadSchema` now accepts the actual API output shapes: plain array, paginated `{items:[...]}` response, or single collection object.
- **Development/usage skills** added under `skills/`.

## [v0.3.2] - 2026-01-20

### Bug Fixes

- **Fixed autodate field duplication**: autodate fields (created, updated) with `system: false` are now properly skipped during code generation to avoid duplication with BaseModel fields
- **Fixed relation field type for maxSelect=0**: Relations with `maxSelect: 0` are now correctly generated as `string` (single relation) instead of `[]string` (multi relation)

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
