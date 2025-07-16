// Package main demonstrates streaming API operations using a test HTTP server.
// This example creates a mock PocketBase API server and tests streaming functionality
// without requiring an actual PocketBase instance.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	pb "github.com/mrchypark/pocketbase-client"
)

// ProgressWriter wraps an io.Writer to track progress
type ProgressWriter struct {
	writer     *os.File
	totalBytes int64
	onProgress func(int64)
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.totalBytes += int64(n)
	if pw.onProgress != nil {
		pw.onProgress(pw.totalBytes)
	}

	return n, nil
}

// Test streaming with actual HTTP server
func main() {
	fmt.Println("=== HTTP Test Server Streaming Example ===")

	// Create mock PocketBase API server
	server := createMockPocketBaseServer()
	defer server.Close()

	// Create client (using test server URL)
	client := pb.NewClient(server.URL)

	fmt.Printf("Test server URL: %s\n", server.URL)

	// Run streaming tests
	runStreamingTests(client)

	fmt.Println("\n=== Tests completed ===")
}

// createMockPocketBaseServer creates a mock PocketBase API server
func createMockPocketBaseServer() *httptest.Server {
	mux := http.NewServeMux()

	// Collections endpoint
	mux.HandleFunc("/api/collections", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Simulate large response
		response := map[string]interface{}{
			"page":       1,
			"perPage":    30,
			"totalItems": 1000,
			"totalPages": 34,
			"items":      generateMockCollections(100), // Generate 100 collections
		}

		// Encode JSON and respond
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(response)
	})

	// Records endpoint (large data)
	mux.HandleFunc("/api/collections/posts/records", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := map[string]interface{}{
			"page":       1,
			"perPage":    100,
			"totalItems": 10000,
			"totalPages": 10,
			"items":      generateMockRecords(1000), // Generate 1000 records
		}

		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		encoder.Encode(response)
	})

	return httptest.NewServer(mux)
}

// generateMockCollections generates mock collection data
func generateMockCollections(count int) []map[string]interface{} {
	collections := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		collections[i] = map[string]interface{}{
			"id":      fmt.Sprintf("collection_%d", i+1),
			"name":    fmt.Sprintf("Collection %d", i+1),
			"type":    "base",
			"system":  false,
			"schema":  []interface{}{},
			"created": "2023-01-01T00:00:00.000Z",
			"updated": "2023-01-01T00:00:00.000Z",
		}
	}
	return collections
}

// generateMockRecords generates mock record data
func generateMockRecords(count int) []map[string]interface{} {
	records := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		records[i] = map[string]interface{}{
			"id":      fmt.Sprintf("record_%d", i+1),
			"title":   fmt.Sprintf("Post %d", i+1),
			"content": fmt.Sprintf("This is the content of post %d", i+1),
			"created": "2023-01-01T00:00:00.000Z",
			"updated": "2023-01-01T00:00:00.000Z",
		}
	}
	return records
}

// runStreamingTests runs streaming tests
func runStreamingTests(client *pb.Client) {
	ctx := context.Background()

	// Test 1: Basic streaming
	fmt.Println("\n1. Basic streaming test...")
	if err := testBasicStreaming(ctx, client); err != nil {
		log.Printf("Basic streaming failed: %v", err)
	}

	// Test 2: Large data streaming
	fmt.Println("\n2. Large data streaming test...")
	if err := testLargeDataStreaming(ctx, client); err != nil {
		log.Printf("Large data streaming failed: %v", err)
	}

	// Test 3: Progress tracking streaming
	fmt.Println("\n3. Progress tracking streaming test...")
	if err := testProgressStreaming(ctx, client); err != nil {
		log.Printf("Progress streaming failed: %v", err)
	}
}

// testBasicStreaming tests basic streaming
func testBasicStreaming(ctx context.Context, client *pb.Client) error {
	var buf bytes.Buffer

	err := client.SendWithOptions(
		ctx,
		"GET",
		"/api/collections",
		nil,
		nil,
		pb.WithResponseWriter(&buf),
	)
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	fmt.Printf("   Streamed data size: %d bytes\n", buf.Len())

	// Parse JSON to verify data
	var response map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &response); err != nil {
		return fmt.Errorf("JSON parsing failed: %w", err)
	}

	if items, ok := response["items"].([]interface{}); ok {
		fmt.Printf("   Collection count: %d\n", len(items))
	}

	return nil
}

// testLargeDataStreaming tests large data streaming
func testLargeDataStreaming(ctx context.Context, client *pb.Client) error {
	file, err := os.Create("large_data_stream.json")
	if err != nil {
		return fmt.Errorf("file creation failed: %w", err)
	}
	defer file.Close()

	start := time.Now()

	err = client.SendWithOptions(
		ctx,
		"GET",
		"/api/collections/posts/records",
		nil,
		nil,
		pb.WithResponseWriter(file),
	)
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	duration := time.Since(start)
	fileInfo, _ := file.Stat()

	fmt.Printf("   Large data streaming completed: large_data_stream.json\n")
	fmt.Printf("   File size: %d bytes\n", fileInfo.Size())
	fmt.Printf("   Duration: %v\n", duration)
	fmt.Printf("   Processing speed: %.2f MB/s\n", float64(fileInfo.Size())/duration.Seconds()/1024/1024)

	return nil
}

// testProgressStreaming tests progress tracking streaming
func testProgressStreaming(ctx context.Context, client *pb.Client) error {
	file, err := os.Create("progress_test_stream.json")
	if err != nil {
		return fmt.Errorf("file creation failed: %w", err)
	}
	defer file.Close()

	var lastProgress int64
	progressWriter := &ProgressWriter{
		writer: file,
		onProgress: func(bytes int64) {
			// Output progress every 1KB (to avoid too frequent output)
			if bytes-lastProgress >= 1024 || bytes == 0 {
				fmt.Printf("\r   Progress: %d bytes (%.2f KB)", bytes, float64(bytes)/1024)
				lastProgress = bytes
			}
		},
	}

	err = client.SendWithOptions(
		ctx,
		"GET",
		"/api/collections/posts/records",
		nil,
		nil,
		pb.WithResponseWriter(progressWriter),
	)

	fmt.Println() // New line after progress output

	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	fmt.Printf("   Progress streaming completed: progress_test_stream.json (%d bytes)\n", progressWriter.totalBytes)

	return nil
}
