# PocketBase Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/mrchypark/pocketbase-client.svg)](https://pkg.go.dev/github.com/mrchypark/pocketbase-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrchypark/pocketbase-client)](https://goreportcard.com/report/github.com/mrchypark/pocketbase-client)
[![CI](https://github.com/mrchypark/pocketbase-client/actions/workflows/go.yml/badge.svg)](https://github.com/mrchypark/pocketbase-client/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A robust, type-safe Go client for the [PocketBase API](https://pocketbase.io/). Provides compile-time type safety, automatic pagination, real-time subscriptions, and code generation from PocketBase schema.

## ✨ Key Features

- **Type-Safe Generic Services**: `RecordService[T]` with compile-time type checking
- **Code Generation**: Generate Go models from PocketBase schema with `pbc-gen`
- **Full API Coverage**: Records, Collections, Admins, Users, Logs, Settings, Files
- **Real-time Subscriptions**: Event-driven updates with callback functions
- **Automatic Pagination**: `GetAll()` handles pagination automatically
- **Context-Aware**: All operations support `context.Context`
- **File Management**: Upload/download with streaming support

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

// Define your record type
type Post struct {
    pocketbase.BaseModel
    Title   string `json:"title"`
    Content string `json:"content"`
}

func main() {
    client := pocketbase.NewClient("http://localhost:8090")
    
    // Authenticate
    _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }

    // Create type-safe service
    posts := pocketbase.NewRecordService[Post](client, "posts")

    // Create record
    newPost := &Post{Title: "Hello World", Content: "My first post"}
    created, err := posts.Create(context.Background(), newPost)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Created: %s (ID: %s)\n", created.Title, created.ID)
}
```

## 🛠️ Code Generation

Generate type-safe models from PocketBase schema:

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

## 📖 Examples

Comprehensive examples in the [`examples/`](./examples/) directory:

- **[`basic_crud`](./examples/basic_crud/)** - Essential CRUD operations
- **[`auth`](./examples/auth/)** - Authentication patterns  
- **[`batch`](./examples/batch/)** - Batch operations
- **[`file_management`](./examples/file_management/)** - File upload/download
- **[`realtime_subscriptions`](./examples/realtime_subscriptions/)** - Real-time events
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