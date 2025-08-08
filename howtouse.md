# Streaming API Usage Guide

This guide explains how to use the streaming capabilities of the PocketBase Go Client for handling large responses efficiently.

## Overview

The PocketBase Go Client provides streaming functionality through the `WithResponseWriter` option, which allows you to stream response data directly to an `io.Writer` instead of loading the entire response into memory. This is particularly useful for:

- Large file downloads
- Large dataset exports
- Real-time data streaming
- Memory-efficient data processing

## Basic Streaming Usage

### Using WithResponseWriter

The `WithResponseWriter` option streams the response body directly to the provided writer:

```go
package main

import (
    "bytes"
    "context"
    "fmt"
    "log"
    "os"

    pb "github.com/mrchypark/pocketbase-client"
)

func main() {
    client := pb.NewClient("http://127.0.0.1:8090")
    
    // Authenticate if required
    _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password")
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }

    // Stream response to a buffer
    var buf bytes.Buffer
    err = client.SendWithOptions(
        context.Background(),
        "GET",
        "/api/collections/posts/records",
        nil,
        nil, // responseData must be nil when using WithResponseWriter
        pb.WithResponseWriter(&buf),
    )
    if err != nil {
        log.Fatalf("Streaming failed: %v", err)
    }

    fmt.Printf("Streamed %d bytes\n", buf.Len())
    fmt.Printf("Content: %s\n", buf.String())
}
```

### Streaming to File

Stream large responses directly to a file:

```go
func streamToFile(client *pb.Client) error {
    // Create output file
    file, err := os.Create("large_export.json")
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()

    // Stream response directly to file
    err = client.SendWithOptions(
        context.Background(),
        "GET",
        "/api/collections/large_dataset/records?perPage=10000",
        nil,
        nil, // Must be nil when using WithResponseWriter
        pb.WithResponseWriter(file),
    )
    if err != nil {
        return fmt.Errorf("streaming to file failed: %w", err)
    }

    fmt.Println("Large dataset successfully streamed to file")
    return nil
}
```

### Streaming with Flush Support

If your writer implements `http.Flusher`, the client will automatically call `Flush()` after each write operation:

```go
type FlushingWriter struct {
    *os.File
}

func (fw *FlushingWriter) Flush() {
    // Custom flush logic
    fw.File.Sync()
}

func streamWithFlushing(client *pb.Client) error {
    file, err := os.Create("streamed_data.json")
    if err != nil {
        return err
    }
    defer file.Close()

    flushingWriter := &FlushingWriter{File: file}
    
    err = client.SendWithOptions(
        context.Background(),
        "GET",
        "/api/collections/posts/records",
        nil,
        nil,
        pb.WithResponseWriter(flushingWriter),
    )
    
    return err
}
```

## Advanced Streaming Patterns

### Progress Tracking

Track streaming progress by wrapping the writer:

```go
type ProgressWriter struct {
    writer      io.Writer
    totalBytes  int64
    onProgress  func(bytes int64)
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
    n, err = pw.writer.Write(p)
    pw.totalBytes += int64(n)
    if pw.onProgress != nil {
        pw.onProgress(pw.totalBytes)
    }
    return
}

func streamWithProgress(client *pb.Client) error {
    var buf bytes.Buffer
    
    progressWriter := &ProgressWriter{
        writer: &buf,
        onProgress: func(bytes int64) {
            fmt.Printf("\rStreamed: %d bytes", bytes)
        },
    }

    err := client.SendWithOptions(
        context.Background(),
        "GET",
        "/api/collections/large_collection/records?perPage=5000",
        nil,
        nil,
        pb.WithResponseWriter(progressWriter),
    )
    
    fmt.Println() // New line after progress
    return err
}
```

### Streaming with Compression

Handle compressed responses:

```go
import (
    "compress/gzip"
    "io"
)

func streamCompressed(client *pb.Client) error {
    file, err := os.Create("compressed_data.json.gz")
    if err != nil {
        return err
    }
    defer file.Close()

    // Create gzip writer
    gzipWriter := gzip.NewWriter(file)
    defer gzipWriter.Close()

    err = client.SendWithOptions(
        context.Background(),
        "GET",
        "/api/collections/posts/records",
        nil,
        nil,
        pb.WithResponseWriter(gzipWriter),
    )
    
    return err
}
```

### Streaming JSON Processing

Process large JSON responses line by line:

```go
import (
    "bufio"
    "encoding/json"
    "io"
)

type StreamProcessor struct {
    scanner *bufio.Scanner
    pipe    *io.PipeWriter
}

func NewStreamProcessor() (*StreamProcessor, io.Writer) {
    pr, pw := io.Pipe()
    return &StreamProcessor{
        scanner: bufio.NewScanner(pr),
        pipe:    pw,
    }, pw
}

func (sp *StreamProcessor) ProcessRecords(callback func(record map[string]any)) error {
    defer sp.pipe.Close()
    
    for sp.scanner.Scan() {
        line := sp.scanner.Text()
        if line == "" {
            continue
        }
        
        var record map[string]any
        if err := json.Unmarshal([]byte(line), &record); err != nil {
            continue // Skip invalid JSON
        }
        
        callback(record)
    }
    
    return sp.scanner.Err()
}

func streamAndProcess(client *pb.Client) error {
    processor, writer := NewStreamProcessor()
    
    // Start processing in a goroutine
    go func() {
        processor.ProcessRecords(func(record map[string]any) {
            fmt.Printf("Processing record: %s\n", record["id"])
            // Process each record as it arrives
        })
    }()

    // Stream data to processor
    return client.SendWithOptions(
        context.Background(),
        "GET",
        "/api/collections/posts/records?perPage=1000",
        nil,
        nil,
        pb.WithResponseWriter(writer),
    )
}
```

## Real-time Streaming

For real-time data streaming, use the Realtime service:

```go
func realtimeStreaming(client *pb.Client) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    unsubscribe, err := client.Realtime.Subscribe(
        ctx,
        []string{"posts", "comments"},
        func(event *pb.RealtimeEvent, err error) {
            if err != nil {
                log.Printf("Realtime error: %v", err)
                return
            }
            
            // Stream real-time events to a file or processor
            fmt.Printf("Real-time event: %s on %s\n", 
                event.Action, event.Record.CollectionName)
        },
    )
    if err != nil {
        return err
    }
    defer unsubscribe()

    // Keep streaming until context is cancelled
    <-ctx.Done()
    return nil
}
```

## Important Notes

### Limitations

1. **Mutual Exclusivity**: You cannot use `WithResponseWriter` together with a `responseData` parameter. Doing so will result in an error.

```go
// ❌ This will fail
var result map[string]any
err := client.SendWithOptions(ctx, "GET", "/path", nil, &result, WithResponseWriter(&buf))
// Error: "WithResponseWriter and responseData cannot be used together"

// ✅ This is correct
err := client.SendWithOptions(ctx, "GET", "/path", nil, nil, WithResponseWriter(&buf))
```

2. **Context Handling**: Always use context for timeout and cancellation control:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := client.SendWithOptions(ctx, "GET", "/path", nil, nil, WithResponseWriter(writer))
```

3. **Error Handling**: Streaming errors are returned immediately, not through the writer:

```go
err := client.SendWithOptions(ctx, "GET", "/path", nil, nil, WithResponseWriter(writer))
if err != nil {
    // Handle streaming error
    log.Printf("Streaming failed: %v", err)
}
```

### Performance Tips

1. **Buffer Size**: The internal buffer size is 32KB. For better performance with large files, consider using a buffered writer:

```go
file, _ := os.Create("output.json")
bufferedWriter := bufio.NewWriterSize(file, 64*1024) // 64KB buffer
defer bufferedWriter.Flush()

client.SendWithOptions(ctx, "GET", "/path", nil, nil, WithResponseWriter(bufferedWriter))
```

2. **Memory Usage**: Streaming keeps memory usage constant regardless of response size, making it ideal for large datasets.

3. **Network Efficiency**: The client automatically flushes data when the writer implements `http.Flusher`, ensuring efficient network utilization.

## Complete Example

Here's a complete example that demonstrates various streaming patterns:

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"
    "time"

    pb "github.com/mrchypark/pocketbase-client"
)

func main() {
    client := pb.NewClient("http://127.0.0.1:8090")
    
    // Authenticate
    _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password")
    if err != nil {
        log.Fatalf("Authentication failed: %v", err)
    }

    // Example 1: Stream to file
    if err := streamLargeDataset(client); err != nil {
        log.Printf("Stream to file failed: %v", err)
    }

    // Example 2: Stream with progress
    if err := streamWithProgress(client); err != nil {
        log.Printf("Stream with progress failed: %v", err)
    }

    // Example 3: Real-time streaming
    if err := setupRealtimeStream(client); err != nil {
        log.Printf("Real-time streaming failed: %v", err)
    }
}

func streamLargeDataset(client *pb.Client) error {
    file, err := os.Create("large_dataset.json")
    if err != nil {
        return err
    }
    defer file.Close()

    bufferedWriter := bufio.NewWriter(file)
    defer bufferedWriter.Flush()

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    return client.SendWithOptions(
        ctx,
        "GET",
        "/api/collections/posts/records?perPage=10000",
        nil,
        nil,
        pb.WithResponseWriter(bufferedWriter),
    )
}

func streamWithProgress(client *pb.Client) error {
    file, err := os.Create("progress_stream.json")
    if err != nil {
        return err
    }
    defer file.Close()

    progressWriter := &ProgressWriter{
        writer: file,
        onProgress: func(bytes int64) {
            fmt.Printf("\rProgress: %d bytes streamed", bytes)
        },
    }

    err = client.SendWithOptions(
        context.Background(),
        "GET",
        "/api/collections/posts/records",
        nil,
        nil,
        pb.WithResponseWriter(progressWriter),
    )
    
    fmt.Println() // New line after progress
    return err
}

func setupRealtimeStream(client *pb.Client) error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    unsubscribe, err := client.Realtime.Subscribe(
        ctx,
        []string{"posts"},
        func(event *pb.RealtimeEvent, err error) {
            if err != nil {
                log.Printf("Realtime error: %v", err)
                return
            }
            fmt.Printf("Real-time: %s %s\n", event.Action, event.Record.ID)
        },
    )
    if err != nil {
        return err
    }
    defer unsubscribe()

    fmt.Println("Listening for real-time events for 30 seconds...")
    <-ctx.Done()
    return nil
}

type ProgressWriter struct {
    writer     io.Writer
    totalBytes int64
    onProgress func(bytes int64)
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
    n, err = pw.writer.Write(p)
    pw.totalBytes += int64(n)
    if pw.onProgress != nil {
        pw.onProgress(pw.totalBytes)
    }
    return
}
```

This streaming API provides efficient, memory-conscious ways to handle large responses and real-time data in your PocketBase applications.