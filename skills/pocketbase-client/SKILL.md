---
name: pocketbase-client
description: Use when building a Go application on top of a PocketBase backend with the pocketbase-client Go library — installing, exporting the schema and generating typed models with pbc-gen, typed/dynamic CRUD, auth with auto-refresh, realtime subscriptions, batch operations, file handling, and structured error handling.
---

# PocketBase Client (Go)

Build Go applications against a PocketBase backend using the `github.com/mrchypark/pocketbase-client` library. It provides a type-safe CRUD API via generated models, plus a dynamic API for untyped access, realtime subscriptions, batch operations, file upload/download, and structured errors. Generated code depends **only** on this client — the PocketBase module itself is never pulled into your project.

## Install

```bash
# Client library
go get github.com/mrchypark/pocketbase-client

# Code generator (one of):
go install github.com/mrchypark/pocketbase-client/cmd/pbc-gen@latest
# or: curl -sL https://raw.githubusercontent.com/mrchypark/pocketbase-client/main/install.sh | sh
```

## One-time setup: export schema and generate models

```bash
# The raw paginated response of /api/collections is accepted directly.
curl "http://localhost:8090/api/collections?perPage=500" > schema.json

pbc-gen -schema schema.json -path models.gen.go -pkgname models
# Optional flags: -enums=false -relations=false -files=false -jsonlib=github.com/goccy/go-json
```

What you get:

- One struct per collection with system fields (`ID`, `CollectionID`, `CollectionName`, `Created`, `Updated`) plus generated fields, setters, and `ToMap()`.
- `New<Model>()` constructor and `New<Model>Service(client)` typed service constructor.
- Enum constants + `<EnumType>Values()` / `IsValid<EnumType>()` from `select` fields.
- Relation helpers (`<Target>Relation` with `ID()`/`Load()`/`IsEmpty()`) and `FileReference`/`FileReferences` helpers.

Do not hand-edit `models.gen.go`; regenerate it whenever the schema changes.

## Client initialization

```go
client := pocketbase.NewClient("http://127.0.0.1:8090",
    pocketbase.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
)
```

## Authentication

Once authenticated the client injects the `Authorization` header on every request automatically.

```go
ctx := context.Background()

// Admin (superuser) — convenience for WithPassword(_, "_superusers", ...)
client.WithAdminPassword(ctx, "admin@example.com", "password")

// Any auth collection
client.WithPassword(ctx, "users", "username_or_email", "password")

// Static token (preferred in production)
client.WithToken("pb_jwt_token")

// Custom strategy (env/keychain-backed), recommended for high-security environments
client.WithAuthStrategy(myAuthStrategy{})
```

Security note: `WithPassword`/`WithAdminPassword` store the plaintext credentials in memory so the token can be transparently refreshed when the JWT expires. Prefer `WithToken` or a custom `AuthStrategy` in production. `PasswordAuth` refreshes via `singleflight`, so concurrent requests share one refresh.

## Typed CRUD (recommended)

Generated services return your concrete model types and decode responses directly (single-pass, no JSON round-trip).

```go
posts := models.NewPostsService(client) // == pocketbase.NewTypedRecordService[models.Posts](client, "posts")

// Create — accepts *Posts, returns *Posts
post := models.NewPosts()
post.SetTitle("Hello")
created, err := posts.Create(ctx, post)

// Read
one, err := posts.GetOne(ctx, created.ID, nil)
list, err := posts.GetList(ctx, &pocketbase.ListOptions{
    Page:    1,
    PerPage: 20,
    Sort:    "-created",
    Filter:  "published = true",
    Expand:  "author",
})
all, err := posts.GetAll(ctx, &pocketbase.ListOptions{Filter: "published = true"}) // auto-pagination

// Update — PATCH semantics: empty optional fields are omitted via ToMap()
one.SetTitle("Updated")
updated, err := posts.Update(ctx, one.ID, one)

// Delete — via the embedded RecordService
err = posts.Delete(ctx, posts.Collection, one.ID)
```

## Dynamic API (no codegen)

```go
rec, err := client.Records.GetOne(ctx, "posts", id, nil)
list, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{Filter: "views > 100"})
created, err := client.Records.Create(ctx, "posts", map[string]any{"title": "New"})
updated, err := client.Records.Update(ctx, "posts", id, map[string]any{"title": "Changed"})
err = client.Records.Delete(ctx, "posts", id)
```

`*Record` provides type helpers: `GetString(key)`, `GetBool`, `GetFloat`, `GetDateTime`, `GetStringSlice`, `GetRawMessage`, plus pointer variants.

## Query options

```go
type ListOptions struct {
    Page        int
    PerPage     int
    Sort        string
    Filter      string
    Expand      string
    Fields      string
    SkipTotal   bool
    QueryParams map[string]string
}
type GetOneOptions struct { Expand, Fields string }
type WriteOptions  struct { Expand, Fields string }
```

## Field type mapping

| PocketBase | Go (required) | Go (optional) |
|---|---|---|
| text, email, url, editor, password | `string` | `*string` |
| number | `float64` | `*float64` |
| bool | `bool` | `*bool` |
| date, autodate | `pocketbase.DateTime` | `*pocketbase.DateTime` |
| geoPoint | `pocketbase.GeoPoint` | `*pocketbase.GeoPoint` |
| json | `json.RawMessage` | `json.RawMessage` |
| select, relation, file (maxSelect > 1) | `[]string` | `[]string` |
| select, relation, file (single) | `string` | `*string` |

`pocketbase.DateTime` is wire-compatible with PocketBase's `tools/types.DateTime` (`"2006-01-02 15:04:05.000Z"` layout) and exposes `Time()`, `String()`, `IsZero()`, `Scan()`, `Value()`, `Compare()`, etc. `pocketbase.GeoPoint` is `{Lon, Lat float64}`.

## Batch operations (atomic)

```go
createReq, _ := client.Records.NewCreateRequest("posts", map[string]any{"title": "Batch 1"})
updateReq, _ := client.Records.NewUpdateRequest("posts", id, map[string]any{"title": "Batch 2"})
deleteReq, _ := client.Records.NewDeleteRequest("posts", oldID)
results, err := client.Batch.Execute(ctx, []*pocketbase.BatchRequest{createReq, updateReq, deleteReq})
```

## Files

```go
f, _ := os.Open("image.jpg")
defer f.Close()
rec, err := client.Files.Upload(ctx, "posts", recordID, "image", "image.jpg", f)

reader, err := client.Files.Download(ctx, "posts", recordID, "image.jpg", nil)
defer reader.Close()

url := client.Files.GetFileURL("posts", recordID, "image.jpg", nil)
err = client.Files.Delete(ctx, "posts", recordID, "image", "image.jpg")
```

## Realtime subscriptions (SSE)

```go
unsubscribe, err := client.Realtime.Subscribe(ctx, []string{"posts"}, func(e *pocketbase.RealtimeEvent, err error) {
    if err != nil { log.Printf("realtime: %v", err); return }
    fmt.Printf("%s on record %s\n", e.Action, e.Record.ID)
})
defer unsubscribe()
```

## Error handling

All API errors are `*pocketbase.Error` with `Status`, `Code` (stable alias), `Message`, and `Data map[string]FieldError`.

```go
if err != nil {
    var pbErr *pocketbase.Error
    if errors.As(err, &pbErr) {
        fmt.Printf("HTTP %d: %s (code: %s)\n", pbErr.Status, pbErr.Message, pbErr.Code)
    }
}

// Convenience checks
if pocketbase.IsNotFoundError(err) { /* 404 */ }
if pocketbase.IsAuthError(err)      { /* 401 */ }
if pocketbase.IsForbiddenError(err) { /* 403 */ }
if pocketbase.IsValidationError(err) {
    for field, fe := range pocketbase.GetFieldErrors(err) {
        fmt.Printf("  %s: %s\n", field, fe.Message)
    }
}
if pocketbase.HasErrorCode(err, "record_not_found") { /* ... */ }
if errors.Is(err, pocketbase.StatusNotFound)        { /* 404 constant */ }
```

## Health check

```go
status, err := client.HealthCheck(ctx)
```

## Migration notes

- **DateTime/GeoPoint (current)**: generated `date`/`autodate` fields use the client's own `pocketbase.DateTime` and `geoPoint` fields use `pocketbase.GeoPoint`, instead of `github.com/pocketbase/pocketbase/tools/types`. The JSON wire format is identical — only the type name/import changes. Update imports and the type name in existing code, then regenerate. Generated code and the client no longer require the PocketBase module.
- **Typed services** (`TypedRecordService`) are the supported API; the older generic `Service[T]` was removed.

## Examples

Runnable examples live in the repo's `examples/` directory (basic_crud, auth, list_options, file_management, realtime_subscriptions, batch, typed_crud, typed_batch, generic_usage).

## Pitfalls

- **Do not edit generated models by hand** — change the schema and rerun `pbc-gen`.
- **Update omits empty optional fields** (PATCH semantics). To clear a field, send a non-empty value or rely on the field type's zero semantics.
- **`perPage` default** for `/api/collections` is 30 — pass `?perPage=500` (or your max) so you don't silently miss collections during generation.
- **Plaintext credentials** are kept in memory with password auth; use `WithToken`/custom strategies in production.
- **Expanded relations** appear under `expand`; use `Expand` in options and read expanded data via the dynamic API or `Fields` selection.
