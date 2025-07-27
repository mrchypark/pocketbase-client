# Direct Record Object Usage Example

This example demonstrates how to use the `Record` object directly in the PocketBase Go client. You can dynamically create and manipulate records without creating custom structs.

## Key Features

### 1. Record Object Creation and Data Setting
```go
newRecord := &pocketbase.Record{}
newRecord.Set("title", "Direct Record Post")
newRecord.Set("content", "This is an example using Record object directly!")
```

### 2. Various Data Type Support
```go
record.Set("title", "String")
record.Set("is_published", true)           // Boolean
record.Set("view_count", 42)               // Number
record.Set("rating", 4.5)                  // Float
record.Set("tags", []string{"tag1", "tag2"}) // String array
```

### 3. Data Retrieval Methods
- `GetString(key)` - Retrieve string value
- `GetBool(key)` - Retrieve boolean value  
- `GetFloat(key)` - Retrieve float value
- `GetStringSlice(key)` - Retrieve string array
- `GetDateTime(key)` - Retrieve date/time value
- `Get(key)` - Retrieve raw value

### 4. Advanced Query Options
- **Filtering**: `Filter: "title ~ 'search_term'"` 
- **Sorting**: `Sort: "-created"` (newest first)
- **Field Selection**: `Fields: "id,title,created"`
- **Relation Expansion**: `Expand: "user"`

## When to Use?

### Direct Record Object Usage is Suitable When:
- Handling dynamic schemas
- Prototyping or rapid development
- Fields are determined at runtime
- Simple CRUD operations

### Custom Structs are Suitable When:
- Type safety is important
- Complex business logic exists
- Want to leverage IDE autocomplete and type checking
- Large-scale application development

## How to Run

```bash
# Set environment variables
export POCKETBASE_URL="http://127.0.0.1:8090"

# Run example
go run examples/record_direct/main.go
```

## Required PocketBase Setup

Before running this example, you need to configure the following in PocketBase:

1. Create `posts` collection
2. Add the following fields:
   - `title` (text)
   - `content` (text)  
   - `is_published` (bool, optional)
   - `view_count` (number, optional)
   - `rating` (number, optional)
   - `tags` (json, optional)
3. Create admin account (`admin@example.com` / `password123`)