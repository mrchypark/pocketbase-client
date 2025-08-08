# PocketBase Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/mrchypark/pocketbase-client.svg)](https://pkg.go.dev/github.com/mrchypark/pocketbase-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrchypark/pocketbase-client)](https://goreportcard.com/report/github.com/mrchypark/pocketbase-client)
[![CI](https://github.com/mrchypark/pocketbase-client/actions/workflows/go.yml/badge.svg)](https://github.com/mrchypark/pocketbase-client/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A robust, type-safe Go client for the [PocketBase API](https://pocketbase.io/). Provides compile-time type safety, automatic pagination, real-time subscriptions, and code generation from PocketBase schema.

## ✨ Key Features

  * **Full API Coverage**: Interact with all PocketBase endpoints, including Records, Collections, Admins, Users, Logs, and Settings.
  * **Type-Safe**: Go structs for all PocketBase models (`Record`, `Admin`, `Collection`, etc.) with proper JSON tagging.
  * **Fluent API**: A logically structured client that is easy to read and use.
  * **Automatic Auth Handling**: The client automatically manages and injects authentication tokens for all relevant requests.
  * **Real-time Subscriptions**: Subscribe to real-time events on your collections with a simple callback function.
  * **File Management**: Easy-to-use methods for uploading, downloading, and managing files.
  * **Batch Operations**: Execute multiple create, update, or delete operations in a single atomic request.
  * **Context-Aware**: All network requests use `context.Context` for timeouts, deadlines, and cancellation.

## 💾 Installation

```bash
go get github.com/mrchypark/pocketbase-client
```

## 🚀 Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
    client := pocketbase.NewClient("http://127.0.0.1:8090")

    // Authenticate as an admin
    _, err := client.AuthenticateAsAdmin(context.Background(), "admin@example.com", "password")
    if err != nil {
        log.Fatalf("Failed to authenticate: %v", err)
    }
    fmt.Println("Authenticated successfully!")

    // Fetch a list of records
    list, err := client.Records.GetList(context.Background(), "posts", &pocketbase.ListOptions{
        Page:    1,
        PerPage: 10,
    })
    if err != nil {
        log.Fatalf("Failed to get record list: %v", err)
    }

    fmt.Printf("Retrieved %d records:\n", len(list.Items))
    for _, record := range list.Items {
        fmt.Printf("- ID: %s, Title: %v\n", record.ID, record.Data["title"])
    }
}
```

## 📚 Usage

### Client Initialization

Create a new client by providing the URL of your PocketBase instance.

```go
client := pocketbase.NewClient("http://127.0.0.1:8090")
```

You can also provide a custom `http.Client` for more advanced configurations, such as setting a timeout.

```go
httpClient := &http.Client{Timeout: 10 * time.Second}
client := pocketbase.NewClient("http://127.0.0.1:8090", pocketbase.WithHTTPClient(httpClient))
```

### Authentication

The client supports authentication for both admins and regular users. Once authenticated, the client will automatically handle token refreshes and include the auth token in subsequent requests.

```go
ctx := context.Background()

// Authenticate as an admin
adminAuth, err := client.AuthenticateAsAdmin(ctx, "admin@example.com", "password")
if err != nil { /* ... */ }

// Authenticate as a user from the 'users' collection
userAuth, err := client.AuthenticateWithPassword(ctx, "users", "username_or_email", "password")
if err != nil { /* ... */ }
```

### Record Operations (CRUD)

Perform Create, Read, Update, and Delete operations on your records.

```bash
# Install pbc-gen
curl -sL https://raw.githubusercontent.com/mrchypark/pocketbase-client/main/install.sh | sh

# Export schema and generate models
curl http://localhost:8090/api/collections > schema.json
pbc-gen -schema schema.json -path models.gen.go -pkgname models
```

**Generated code example:**
```go
type Post struct {
    pocketbase.BaseModel
    Title     string `json:"title"`
    Content   string `json:"content"`
    Published bool   `json:"published"`
}

func NewPostService(client *pocketbase.Client) pocketbase.RecordServiceAPI[Post] {
    return pocketbase.NewRecordService[Post](client, "posts")
}
```

## 📚 Core Operations

### Authentication
```go
ctx := context.Background()

// Admin auth
_, err := client.WithAdminPassword(ctx, "admin@example.com", "password")

// User auth
_, err := client.WithPassword(ctx, "users", "user@example.com", "password")

// Token auth
client.WithToken("your-auth-token")
```

### CRUD Operations
```go
service := pocketbase.NewRecordService[Post](client, "posts")

// Create
post := &Post{Title: "New Post", Content: "Content here"}
created, err := service.Create(ctx, post)

// Read one
post, err := service.GetOne(ctx, "RECORD_ID", nil)

// Read list with options
posts, err := service.GetList(ctx, &pocketbase.ListOptions{
    Page:    1,
    PerPage: 20,
    Sort:    "-created",
    Filter:  "published = true",
    Expand:  "author",
})

// Get all records (auto-pagination)
allPosts, err := service.GetAll(ctx, &pocketbase.ListOptions{
    Filter: "published = true",
})

// Update
post.Title = "Updated Title"
updated, err := service.Update(ctx, post.ID, post)

// Delete
err = service.Delete(ctx, post.ID)
```

### File Management
```go
// Upload
file, _ := os.Open("image.jpg")
defer file.Close()
record, err := client.Files.Upload(ctx, "posts", recordID, "image", "image.jpg", file)

// Download
reader, err := client.Files.Download(ctx, "posts", recordID, "image.jpg", nil)
defer reader.Close()
```

### Real-time Subscriptions
```go
unsubscribe, err := client.Realtime.Subscribe(ctx, []string{"posts"}, func(e *pocketbase.RealtimeEvent, err error) {
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    fmt.Printf("Event: %s on record %s\n", e.Action, e.Record.ID)
})
defer unsubscribe()
```

### Batch Operations
```go
createReq, _ := service.NewCreateRequest(&Post{Title: "Batch Post"})
updateReq, _ := service.NewUpdateRequest("ID", &Post{Title: "Updated"})

results, err := client.Batch.Execute(ctx, []*pocketbase.BatchRequest{
    createReq, updateReq,
})
```

## 🚨 Error Handling

The client provides comprehensive error handling with type-safe error checking and detailed error information.

### Error Types

All PocketBase API errors are wrapped in a structured `*Error` type:

```go
type Error struct {
    Status  int                    // HTTP status code (404, 401, etc.)
    Code    string                 // Stable alias code ("collection_not_found")
    Message string                 // Server error message
    Data    map[string]FieldError  // Field validation errors
    
    // Debugging fields
    Endpoint   string     // API endpoint that failed
    RawHeaders http.Header // Original response headers
    RawBody    []byte     // Original response body
}
```

### Standard Go Error Patterns

Use Go's standard `errors.Is` and `errors.As` patterns:

```go
import (
    "errors"
    "net/http"
    pocketbase "github.com/mrchypark/pocketbase-client"
)

// Check by HTTP status code
if errors.Is(err, pocketbase.HTTPStatus(http.StatusNotFound)) {
    fmt.Println("Resource not found")
}

// Use convenience constants
if errors.Is(err, pocketbase.StatusNotFound) {
    fmt.Println("Resource not found")
}

// Extract detailed error information
var pbErr *pocketbase.Error
if errors.As(err, &pbErr) {
    fmt.Printf("API Error: %d %s (code: %s)\n", 
        pbErr.Status, pbErr.Message, pbErr.Code)
}
```

### Convenience Functions

For common error types, use helper functions:

```go
if pocketbase.IsNotFoundError(err) {
    // Handle 404 errors
}

if pocketbase.IsAuthError(err) {
    // Handle 401 authentication errors
}

if pocketbase.IsValidationError(err) {
    // Handle validation errors with field details
    fieldErrors := pocketbase.GetFieldErrors(err)
    for field, fieldErr := range fieldErrors {
        fmt.Printf("Field %s: %s\n", field, fieldErr.Message)
    }
}

if pocketbase.IsForbiddenError(err) {
    // Handle 403 authorization errors
}

if pocketbase.IsRateLimitedError(err) {
    // Handle 429 rate limit errors
}
```

### Error Code Checking

Check for specific error conditions using stable alias codes:

```go
if pocketbase.HasErrorCode(err, "collection_not_found") {
    // Handle collection not found specifically
}

if pocketbase.HasErrorCode(err, "record_not_found") {
    // Handle record not found specifically
}

// Get error code for logging
code := pocketbase.GetErrorCode(err)
if code != "" {
    log.Printf("PocketBase error code: %s", code)
}
```

### HTTP Status Code Checking

Work directly with HTTP status codes:

```go
// Generic HTTP status checking
if pocketbase.HasHTTPStatus(err, http.StatusNotFound) {
    // Handle any 404 error
}

// Get HTTP status for custom logic
status := pocketbase.GetHTTPStatus(err)
if status >= 500 {
    // Handle server errors
}
```

### Complete Error Handling Example

```go
func handlePocketBaseError(err error) {
    if err == nil {
        return
    }

    // Check for specific error types first
    switch {
    case pocketbase.IsNotFoundError(err):
        log.Println("Resource not found - may need to create it")
        
    case pocketbase.IsAuthError(err):
        log.Println("Authentication failed - check credentials")
        
    case pocketbase.IsValidationError(err):
        log.Println("Validation failed:")
        fieldErrors := pocketbase.GetFieldErrors(err)
        for field, fieldErr := range fieldErrors {
            log.Printf("  %s: %s", field, fieldErr.Message)
        }
        
    case pocketbase.IsForbiddenError(err):
        log.Println("Access denied - insufficient permissions")
        
    case pocketbase.IsRateLimitedError(err):
        log.Println("Rate limited - slow down requests")
        
    default:
        // Extract detailed error information
        var pbErr *pocketbase.Error
        if errors.As(err, &pbErr) {
            log.Printf("PocketBase API error: %d %s", pbErr.Status, pbErr.Message)
            if pbErr.Code != "" {
                log.Printf("Error code: %s", pbErr.Code)
            }
        } else {
            log.Printf("Non-PocketBase error: %v", err)
        }
    }
}

// Usage in your application
post, err := service.GetOne(ctx, "invalid-id", nil)
if err != nil {
    handlePocketBaseError(err)
    return
}
```

### Available Error Codes

Common error codes you can check for:

- `collection_not_found` - Collection doesn't exist
- `record_not_found` - Record doesn't exist
- `invalid_auth_token` - Authentication token is invalid
- `forbidden_generic` - Generic permission denied
- `validation_failed` - Request validation failed
- `invalid_request_payload` - Malformed request data
- `too_many_requests` - Rate limit exceeded
- `internal_generic` - Server internal error

See the [full list of error codes](./errors.md) in the source code.

## 📖 Examples

Comprehensive examples in the [`examples/`](./examples/) directory:

- **[`basic_crud`](./examples/basic_crud/)** - Essential CRUD operations
- **[`auth`](./examples/auth/)** - Authentication patterns  
- **[`batch`](./examples/batch/)** - Batch operations
- **[`file_management`](./examples/file_management/)** - File upload/download
- **[`realtime_subscriptions`](./examples/realtime_subscriptions/)** - Real-time events
- **[`error_handling`](./examples/error_handling/)** - Comprehensive error handling patterns
- **[`type_safe_generator`](./examples/type_safe_generator/)** - Code generation demo

## 🔧 Configuration

```go
client := pocketbase.NewClient("http://localhost:8090",
    pocketbase.WithHTTPClient(&http.Client{
        Timeout: 30 * time.Second,
    }),
)
```

## 🚨 Migration from v0.2.x

**Breaking Changes:**
- Pagination helpers removed → Use `GetAll()` method
- Generic services introduced → Recommended over legacy Record service

**Migration:**
```go
// Old (v0.2.x)
records, err := client.Records().GetList(ctx, "posts", opts)

// New (v0.3.0+)
service := pocketbase.NewRecordService[Post](client, "posts")
records, err := service.GetList(ctx, opts)
allRecords, err := service.GetAll(ctx, opts) // Replaces pagination helpers
```

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

**Development:**
```bash
go test ./...              # Run tests
go test -race ./...        # Race detection
go test -bench=. ./...     # Benchmarks
```

## 📜 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [PocketBase](https://pocketbase.io/) - Amazing backend-as-a-service
- [goccy/go-json](https://github.com/goccy/go-json) - High-performance JSON
- All contributors

---

**⭐ Star this repository if you find it helpful!**