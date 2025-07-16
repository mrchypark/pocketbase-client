# Streaming API Example

This example demonstrates the streaming capabilities of the PocketBase Go Client.

## Features

- **Memory Buffer Streaming**: Stream responses to memory buffers
- **File Streaming**: Stream large responses directly to files
- **Progress Tracking**: Track streaming progress in real-time
- **Real-time Streaming**: Stream real-time events

## How to Run

### Basic Example (with mock data)
```bash
cd examples/streaming_api
go run main.go
```

### Real Streaming Test with HTTP Test Server
```bash
cd examples/streaming_api/server_test
go run main.go
```

### Run with Actual PocketBase Server
```bash
# 1. Start PocketBase server (http://127.0.0.1:8090)
make pb_run

# 2. Run example in another terminal
cd examples/streaming_api
go run main.go
```

## Key Features

### WithResponseWriter Option

The `WithResponseWriter` option allows streaming response data directly to an `io.Writer` without loading it into memory:

```go
var buf bytes.Buffer
err := client.SendWithOptions(
    ctx,
    "GET",
    "/api/collections",
    nil,
    nil, // responseData must be nil
    pb.WithResponseWriter(&buf),
)
```

### Progress Tracking

Implement a custom Writer to track streaming progress:

```go
type ProgressWriter struct {
    writer     io.Writer
    totalBytes int64
    onProgress func(bytes int64)
}
```

### Real-time Streaming

Stream real-time events through the Realtime service:

```go
unsubscribe, err := client.Realtime.Subscribe(
    ctx,
    []string{"*"}, // All collections
    func(event *pb.RealtimeEvent, err error) {
        // Handle real-time events
    },
)
```

## Generated Files

- `collections_stream.json`: File with streamed collection data
- `progress_stream.json`: File streamed with progress tracking

## Important Notes

1. `WithResponseWriter` and `responseData` parameters cannot be used together
2. Always use context with timeouts when streaming
3. For large file streaming, using buffered Writers is recommended