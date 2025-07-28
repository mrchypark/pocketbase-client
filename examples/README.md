# PocketBase Go Client Examples

This directory contains examples demonstrating various usage patterns of the PocketBase Go client.

## 📚 Example List

### Basic Usage
- **[quick_start](quick_start/)** - Quick start guide
- **[basic_crud](basic_crud/)** - Type-safe CRUD operations (recommended)
- **[record_direct](record_direct/)** - Direct Record object usage

### Advanced Features
- **[auth](auth/)** - Authentication and user management
- **[batch](batch/)** - Batch operations
- **[file_management](file_management/)** - File upload/download
- **[list_options](list_options/)** - Advanced query options
- **[realtime_subscriptions](realtime_subscriptions/)** - Real-time subscriptions
- **[realtime_chat](realtime_chat/)** - Real-time chat
- **[streaming_api](streaming_api/)** - Streaming API
- **[type_safe_generator](type_safe_generator/)** - Type-safe code generation

## 🔄 Record Usage Pattern Comparison

### 1. Type-Safe Struct Usage (Recommended)
```go
// Define struct
type Post struct {
    pocketbase.BaseModel
    Title   string `json:"title"`
    Content string `json:"content"`
}

// Create service
postsService := pocketbase.NewRecordService[Post](client, "posts")

// Usage
post := &Post{Title: "Title", Content: "Content"}
created, err := postsService.Create(ctx, post)
fmt.Println(created.Title) // Type-safe access
```

**Advantages:**
- ✅ Compile-time type checking
- ✅ IDE autocomplete support
- ✅ Refactoring safety
- ✅ Clear data structure

**Disadvantages:**
- ❌ Requires struct definition upfront
- ❌ Difficult to handle dynamic schemas

### 2. Direct Record Object Usage
```go
// Create service
recordsService := client.Records("posts")

// Usage
record := &pocketbase.Record{}
record.Set("title", "Title")
record.Set("content", "Content")
created, err := recordsService.Create(ctx, record)
fmt.Println(created.GetString("title")) // Runtime type conversion
```

**Advantages:**
- ✅ Dynamic schema support
- ✅ Fast prototyping
- ✅ Runtime field determination
- ✅ No struct definition required

**Disadvantages:**
- ❌ Possible runtime type errors
- ❌ Limited IDE support
- ❌ Risk of bugs from typos

## 🎯 When to Use Which Approach?

### Use Type-Safe Structs When:
- Building production applications
- Working with complex business logic
- Developing in teams
- Code requires long-term maintenance

### Use Direct Record When:
- Prototyping
- Handling dynamic schemas
- Writing simple scripts
- Schema changes frequently during early development

## 🚀 Getting Started

1. **Beginners**: Start with [quick_start](quick_start/) example
2. **General usage**: Refer to [basic_crud](basic_crud/) example
3. **Dynamic processing needed**: Check [record_direct](record_direct/) example

## 📋 Common Setup

Before running any examples:

```bash
# Run PocketBase server
make pb_run

# Set environment variables
export POCKETBASE_URL="http://127.0.0.1:8090"
```