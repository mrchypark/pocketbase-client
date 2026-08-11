# Usage Scenarios

This guide covers common usage patterns and scenarios to help you get the most out of the PocketBase Go Client.

## 🔐 Authentication

PocketBase supports multiple authentication methods. The client handles token storage and auto-refreshing for you.

### 1. Admin Authentication
Use this for server-side operations requiring full access.

```go
// Login as admin
auth, err := client.WithAdminPassword(ctx, "admin@example.com", "secret123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Logged in as admin: %s\n", auth.Admin.Id)
```

### 2. User Authentication (Password)
Authenticate regular users from the `users` collection (or any auth collection).

```go
// Login as a user
auth, err := client.WithPassword(ctx, "users", "user@example.com", "password123")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Logged in user: %s\n", auth.Record.ID)
```

### 3. API Key / Static Token
If you already have a valid token (e.g., passed from a frontend), you can set it directly.

```go
client.WithToken("YOUR_JWT_TOKEN")
```

---

## 🔍 Fetching Data (List Options)

When fetching lists (`GetList`), you can control the output using `ListOptions`.

### Filtering & Sorting
PocketBase uses a SQL-like syntax for filtering.

```go
opts := &pocketbase.ListOptions{
    // Filter: "published" is true AND "views" > 100
    Filter: "is_published = true && views > 100",
    
    // Sort: Newest first (descending created), then by title
    Sort:   "-created,title",
    
    Page:    1,
    PerPage: 20,
}

result, err := client.Records.GetList(ctx, "posts", opts)
```

### Expanding Relations
If your record has relation fields (e.g., `author`), you can expand them to get the full record details in a single request.

```go
opts := &pocketbase.ListOptions{
    // Expand the 'author' relation and 'comments' (if it's a relation)
    Expand: "author,comments",
}

record, _ := client.Records.GetOne(ctx, "posts", "RECORD_ID", &pocketbase.GetOneOptions{
    Expand: "author",
})

// Access expanded data
authors := record.Expand["author"]
if len(authors) > 0 {
    fmt.Println("Author Name:", authors[0].GetString("name"))
}
```

---

## ⚡ Batch Operations (Transactions)

Perform multiple operations atomically. If one fails, they all fail.

```go
// 1. Prepare requests (NOTE: These do NOT execute immediately)
createReq, _ := client.Records.NewCreateRequest("posts", map[string]any{
    "title": "My New Post",
})

updateReq, _ := client.Records.NewUpdateRequest("users", "USER_ID", map[string]any{
    "name": "Updated Name",
})

deleteReq, _ := client.Records.NewDeleteRequest("comments", "COMMENT_ID")

// 2. Execute them in a single transaction
results, err := client.Batch.Execute(ctx, []*pocketbase.BatchRequest{
    createReq,
    updateReq,
    deleteReq,
})

if err != nil {
    log.Fatal("Batch failed:", err)
}
fmt.Println("All operations succeeded!")
```

---

## 📁 File Uploads

Uploading files requires a `multipart/form-data` request, which the client handles easily.

```go
// Open the file
file, err := os.Open("profile.jpg")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// Upload to 'users' collection, record 'USER_ID', field 'avatar'
record, err := client.Files.Upload(ctx, "users", "USER_ID", "avatar", "profile.jpg", file)

if err != nil {
    log.Fatal("Upload failed:", err)
}
fmt.Println("File uploaded successfully")
```
