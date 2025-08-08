// Package main demonstrates file management operations with PocketBase Go client.
// This example shows how to upload, download, generate URLs, and delete files
// associated with PocketBase records.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// Create PocketBase client
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	// Admin authentication (using environment variables is recommended in production)
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password123")
	if err != nil {
		log.Fatalf("Admin authentication failed: %v", err)
	}

	fmt.Println("=== File Management Example ===")

	// 1. File upload example
	fmt.Println("\n1. File Upload")
	err = uploadFileExample(ctx, client)
	if err != nil {
		log.Printf("File upload failed: %v", err)
	}

	// 2. File download example
	fmt.Println("\n2. File Download")
	err = downloadFileExample(ctx, client)
	if err != nil {
		log.Printf("File download failed: %v", err)
	}

	// 3. File URL generation example
	fmt.Println("\n3. File URL Generation")
	fileURLExample(client)

	// 4. File deletion example
	fmt.Println("\n4. File Deletion")
	err = deleteFileExample(ctx, client)
	if err != nil {
		log.Printf("File deletion failed: %v", err)
	}
}

func uploadFileExample(ctx context.Context, client *pocketbase.Client) error {
	// Create test file content
	fileContent := strings.NewReader("This is test file content.")

	// First create a record (assuming posts collection exists)
	recordData := map[string]interface{}{
		"title":   "File Upload Test",
		"content": "This is a post with an attached file.",
	}

	record, err := client.Records.Create(ctx, "posts", recordData)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	fmt.Printf("Record created: ID = %s\n", record.ID)

	// Upload file to the record's 'image' field
	updatedRecord, err := client.Files.Upload(ctx, "posts", record.ID, "image", "test.txt", fileContent)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	fmt.Printf("File upload successful: %s\n", updatedRecord.ID)
	if imageField := updatedRecord.Get("image"); imageField != nil {
		fmt.Printf("Uploaded file: %v\n", imageField)
	}

	return nil
}

func downloadFileExample(ctx context.Context, client *pocketbase.Client) error {
	// First find a record with a file
	records, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{
		Page:    1,
		PerPage: 1,
		Filter:  "image != ''", // Records with non-empty image field
	})
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	if len(records.Items) == 0 {
		fmt.Println("No records with files to download found.")
		return nil
	}

	record := records.Items[0]
	imageField := record.Get("image")
	if imageField == nil {
		fmt.Println("No image field found.")
		return nil
	}

	var filename string
	switch v := imageField.(type) {
	case string:
		filename = v
	case []any:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				filename = str
			}
		}
	}

	if filename == "" {
		fmt.Println("Filename not found.")
		return nil
	}

	fmt.Printf("Starting file download: %s\n", filename)

	// Download file
	reader, err := client.Files.Download(ctx, "posts", record.ID, filename, nil)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer reader.Close()

	// Read file content
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read file content: %w", err)
	}

	fmt.Printf("Downloaded file size: %d bytes\n", len(content))
	fmt.Printf("File content (first 100 chars): %s\n", string(content[:min(100, len(content))]))

	return nil
}

func fileURLExample(client *pocketbase.Client) {
	// Generate example file URLs
	collection := "posts"
	recordID := "example_record_id"
	filename := "example.jpg"

	// Basic file URL
	basicURL := client.Files.GetFileURL(collection, recordID, filename, nil)
	fmt.Printf("Basic file URL: %s\n", basicURL)

	// Thumbnail URL
	thumbURL := client.Files.GetFileURL(collection, recordID, filename, &pocketbase.FileDownloadOptions{
		Thumb: "100x100",
	})
	fmt.Printf("Thumbnail URL: %s\n", thumbURL)

	// Force download URL
	downloadURL := client.Files.GetFileURL(collection, recordID, filename, &pocketbase.FileDownloadOptions{
		Download: true,
	})
	fmt.Printf("Download URL: %s\n", downloadURL)

	// Thumbnail + download URL
	combinedURL := client.Files.GetFileURL(collection, recordID, filename, &pocketbase.FileDownloadOptions{
		Thumb:    "200x200",
		Download: true,
	})
	fmt.Printf("Thumbnail + download URL: %s\n", combinedURL)
}

func deleteFileExample(ctx context.Context, client *pocketbase.Client) error {
	// Find a record with a file
	records, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{
		Page:    1,
		PerPage: 1,
		Filter:  "image != ''", // Records with non-empty image field
	})
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	if len(records.Items) == 0 {
		fmt.Println("No records with files to delete found.")
		return nil
	}

	record := records.Items[0]
	imageField := record.Get("image")
	if imageField == nil {
		fmt.Println("No image field found.")
		return nil
	}

	var filename string
	switch v := imageField.(type) {
	case string:
		filename = v
	case []any:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				filename = str
			}
		}
	}

	if filename == "" {
		fmt.Println("No filename to delete found.")
		return nil
	}

	fmt.Printf("Starting file deletion: %s\n", filename)

	// Delete file
	updatedRecord, err := client.Files.Delete(ctx, "posts", record.ID, "image", filename)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fmt.Printf("File deletion completed. Updated record ID: %s\n", updatedRecord.ID)
	if imageField := updatedRecord.Get("image"); imageField != nil {
		fmt.Printf("Image field after deletion: %v\n", imageField)
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
