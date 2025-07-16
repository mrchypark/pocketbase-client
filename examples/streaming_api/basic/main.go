// Package main demonstrates streaming API operations with PocketBase Go client.
// This example shows how to stream large responses to memory buffers, files,
// and track progress during streaming operations.
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "github.com/mrchypark/pocketbase-client"
)

func main() {
	fmt.Println("=== PocketBase Streaming API Example ===")

	// PocketBase server connection check
	client := pb.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	// Server connection test
	fmt.Println("1. Checking PocketBase server connection...")
	if !testConnection(client) {
		fmt.Println("   PocketBase server is not running.")
		fmt.Println("   Using mock server to demonstrate streaming functionality.")
		runMockExamples()
		return
	}

	fmt.Println("   Server connection successful!")

	// Admin authentication (optional)
	fmt.Println("2. Attempting admin authentication...")
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password")
	if err != nil {
		fmt.Printf("   Authentication failed (continuing anyway): %v\n", err)
	} else {
		fmt.Println("   Authentication successful!")
	}

	// Run examples with real server
	runRealExamples(client)
}

// streamToBuffer streams response to memory buffer
func streamToBuffer(client *pb.Client) error {
	var buf bytes.Buffer

	// Use SendWithOptions for streaming
	err := client.SendWithOptions(
		context.Background(),
		"GET",
		"/api/collections",
		nil,
		nil, // responseData should be nil when using WithResponseWriter
		pb.WithResponseWriter(&buf),
	)
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	fmt.Printf("   Streamed data size: %d bytes\n", buf.Len())

	// Show first 200 characters only
	content := buf.String()
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	fmt.Printf("   Content preview: %s\n", content)

	return nil
}

// streamToFile streams response directly to file
func streamToFile(client *pb.Client) error {
	// Create output file
	file, err := os.Create("collections_stream.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Use buffered writer for better performance
	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stream directly to file
	err = client.SendWithOptions(
		ctx,
		"GET",
		"/api/collections",
		nil,
		nil,
		pb.WithResponseWriter(bufferedWriter),
	)
	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	// Check file size
	fileInfo, _ := file.Stat()
	fmt.Printf("   File streaming completed: collections_stream.json (%d bytes)\n", fileInfo.Size())

	return nil
}

// streamWithProgress streams with progress tracking
func streamWithProgress(client *pb.Client) error {
	file, err := os.Create("progress_stream.json")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create progress tracking writer
	progressWriter := &ProgressWriter{
		writer: file,
		onProgress: func(bytes int64) {
			fmt.Printf("\r   Progress: %d bytes streamed", bytes)
		},
	}

	err = client.SendWithOptions(
		context.Background(),
		"GET",
		"/api/collections",
		nil,
		nil,
		pb.WithResponseWriter(progressWriter),
	)

	fmt.Println() // New line after progress output

	if err != nil {
		return fmt.Errorf("streaming failed: %w", err)
	}

	fmt.Printf("   Progress streaming completed: progress_stream.json (%d bytes)\n", progressWriter.totalBytes)
	return nil
}

// setupRealtimeStream sets up real-time streaming
func setupRealtimeStream(client *pb.Client) error {
	// Set 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set up real-time subscription
	unsubscribe, err := client.Realtime.Subscribe(
		ctx,
		[]string{"*"}, // Subscribe to all collections
		func(event *pb.RealtimeEvent, err error) {
			if err != nil {
				log.Printf("   Real-time error: %v", err)
				return
			}
			fmt.Printf("   Real-time event: %s action occurred on record %s\n",
				event.Action, event.Record.ID)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	defer unsubscribe()

	fmt.Println("   Receiving real-time events... (10 seconds)")

	// Wait until context is done
	<-ctx.Done()

	fmt.Println("   Real-time streaming completed")
	return nil
}

// testConnection tests PocketBase server connection
func testConnection(client *pb.Client) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Simple GET request to check server connection (endpoint that doesn't require auth)
	err := client.SendWithOptions(ctx, "GET", "/api/health", nil, nil)
	return err == nil
}

// runMockExamples demonstrates streaming functionality with mock data
func runMockExamples() {
	fmt.Println("\n=== Mock Streaming Examples ===")

	// Example 1: Memory buffer streaming demonstration
	fmt.Println("1. Memory buffer streaming demonstration...")
	mockStreamToBuffer()

	// Example 2: File streaming demonstration
	fmt.Println("\n2. File streaming demonstration...")
	mockStreamToFile()

	// Example 3: Progress tracking demonstration
	fmt.Println("\n3. Progress tracking streaming demonstration...")
	mockStreamWithProgress()

	fmt.Println("\n=== Mock examples completed ===")
	fmt.Println("\nTo run with actual PocketBase server:")
	fmt.Println("1. Download PocketBase binary: make pb")
	fmt.Println("2. Run server: make pb_run")
	fmt.Println("3. Run this example again: go run main.go")
}

// runRealExamples runs examples with actual PocketBase server
func runRealExamples(client *pb.Client) {
	fmt.Println("\n=== Real Server Streaming Examples ===")

	// Example 1: Stream to memory buffer
	fmt.Println("3. Streaming to memory buffer...")
	if err := streamToBuffer(client); err != nil {
		log.Printf("Buffer streaming failed: %v", err)
	}

	// Example 2: Stream to file
	fmt.Println("\n4. Streaming to file...")
	if err := streamToFile(client); err != nil {
		log.Printf("File streaming failed: %v", err)
	}

	// Example 3: Stream with progress tracking
	fmt.Println("\n5. Streaming with progress tracking...")
	if err := streamWithProgress(client); err != nil {
		log.Printf("Progress streaming failed: %v", err)
	}

	// Example 4: Real-time streaming (short duration)
	fmt.Println("\n6. Real-time streaming (5 seconds)...")
	if err := setupRealtimeStreamSafe(client); err != nil {
		log.Printf("Real-time streaming failed: %v", err)
	}

	fmt.Println("\n=== All examples completed ===")
}

// mockStreamToBuffer demonstrates buffer streaming with mock data
func mockStreamToBuffer() {
	// Generate mock JSON data
	mockData := `{
  "page": 1,
  "perPage": 30,
  "totalItems": 100,
  "totalPages": 4,
  "items": [
    {"id": "1", "name": "Collection 1", "type": "base"},
    {"id": "2", "name": "Collection 2", "type": "auth"},
    {"id": "3", "name": "Collection 3", "type": "base"}
  ]
}`

	var buf bytes.Buffer
	buf.WriteString(mockData)

	fmt.Printf("   Mock streamed data size: %d bytes\n", buf.Len())
	fmt.Printf("   Content preview: %s...\n", mockData[:100])
}

// mockStreamToFile demonstrates file streaming with mock data
func mockStreamToFile() {
	mockData := `{
  "page": 1,
  "perPage": 30,
  "totalItems": 1000,
  "items": [`

	// Generate mock large data
	for i := 0; i < 100; i++ {
		if i > 0 {
			mockData += ","
		}
		mockData += fmt.Sprintf(`
    {
      "id": "%d",
      "name": "Collection %d",
      "type": "base",
      "created": "2023-01-01 00:00:00.000Z",
      "updated": "2023-01-01 00:00:00.000Z"
    }`, i+1, i+1)
	}

	mockData += `
  ]
}`

	// Write to file
	file, err := os.Create("mock_collections_stream.json")
	if err != nil {
		fmt.Printf("   File creation failed: %v\n", err)
		return
	}
	defer file.Close()

	bufferedWriter := bufio.NewWriter(file)
	defer bufferedWriter.Flush()

	_, err = bufferedWriter.WriteString(mockData)
	if err != nil {
		fmt.Printf("   File write failed: %v\n", err)
		return
	}

	fileInfo, _ := file.Stat()
	fmt.Printf("   Mock file streaming completed: mock_collections_stream.json (%d bytes)\n", fileInfo.Size())
}

// mockStreamWithProgress demonstrates progress tracking with mock data
func mockStreamWithProgress() {
	file, err := os.Create("mock_progress_stream.json")
	if err != nil {
		fmt.Printf("   File creation failed: %v\n", err)
		return
	}
	defer file.Close()

	progressWriter := &ProgressWriter{
		writer: file,
		onProgress: func(bytes int64) {
			fmt.Printf("\r   Progress: %d bytes streamed", bytes)
		},
	}

	// Write mock data in chunks to demonstrate progress
	mockChunks := []string{
		`{"page": 1, "perPage": 30, "totalItems": 500, "items": [`,
		`{"id": "1", "name": "Item 1"},`,
		`{"id": "2", "name": "Item 2"},`,
		`{"id": "3", "name": "Item 3"}`,
		`]}`,
	}

	for _, chunk := range mockChunks {
		progressWriter.Write([]byte(chunk))
		time.Sleep(200 * time.Millisecond) // Delay for progress demonstration
	}

	fmt.Println() // New line after progress output
	fmt.Printf("   Mock progress streaming completed: mock_progress_stream.json (%d bytes)\n", progressWriter.totalBytes)
}

// setupRealtimeStreamSafe sets up safe real-time streaming (prevents deadlock)
func setupRealtimeStreamSafe(client *pb.Client) error {
	// Set 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create error channel
	errChan := make(chan error, 1)

	go func() {
		// Set up real-time subscription
		unsubscribe, err := client.Realtime.Subscribe(
			ctx,
			[]string{"*"},
			func(event *pb.RealtimeEvent, err error) {
				if err != nil {
					log.Printf("   Real-time error: %v", err)
					return
				}
				fmt.Printf("   Real-time event: %s action occurred on record %s\n",
					event.Action, event.Record.ID)
			},
		)
		if err != nil {
			errChan <- err
			return
		}
		defer unsubscribe()

		fmt.Println("   Receiving real-time events... (5 seconds)")

		// Wait until context is done
		select {
		case <-ctx.Done():
			errChan <- nil
		}
	}()

	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
		fmt.Println("   Real-time streaming completed")
		return nil
	case <-time.After(6 * time.Second):
		fmt.Println("   Real-time streaming timeout")
		return nil
	}
}

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
