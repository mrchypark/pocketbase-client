// Package main demonstrates basic CRUD operations using type-safe structs with PocketBase Go client.
// This example shows how to use generic RecordService for type-safe database operations.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// Post defines a struct that matches your collection's schema.
// The `json` tags are important for serialization.
type Post struct {
	pocketbase.BaseModel
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	// Initialize PocketBase client
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// Authenticate as an admin (or user) to have permission to modify data
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Create generic record service for posts collection
	postsService := pocketbase.NewRecordService[Post](client, "posts")

	// --- 1. Create a new record ---
	fmt.Println("--- Creating a new record ---")
	newPost := &Post{
		Title:   "My First Post",
		Content: "Hello from the Go SDK!",
	}
	createdRecord, err := postsService.Create(context.Background(), newPost)
	if err != nil {
		log.Fatalf("Failed to create record: %v", err)
	}
	fmt.Printf("Created record ID: %s, Title: '%s'\n\n", createdRecord.ID, createdRecord.Title)

	// --- 2. Get a list of records ---
	fmt.Println("--- Listing records ---")
	records, err := postsService.GetList(context.Background(), &pocketbase.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}
	fmt.Printf("Found %d total record(s).\n", records.TotalItems)
	for i, record := range records.Items {
		fmt.Printf("  %d: ID=%s, Title='%s'\n", i+1, record.ID, record.Title)
	}
	fmt.Println()

	// --- 3. Get a single record ---
	fmt.Println("--- Getting a single record ---")
	recordID := createdRecord.ID
	retrievedRecord, err := postsService.GetOne(context.Background(), recordID, nil)
	if err != nil {
		log.Fatalf("Failed to get record %s: %v", recordID, err)
	}
	fmt.Printf("Retrieved record title: '%s', content: '%s'\n\n", retrievedRecord.Title, retrievedRecord.Content)

	// --- 4. Update a record ---
	fmt.Println("--- Updating a record ---")
	updatePost := &Post{
		BaseModel: pocketbase.BaseModel{ID: recordID},
		Title:     "My Updated Post Title",
	}
	updatedRecord, err := postsService.Update(context.Background(), recordID, updatePost)
	if err != nil {
		log.Fatalf("Failed to update record %s: %v", recordID, err)
	}
	fmt.Printf("Record title updated to: '%s'\n\n", updatedRecord.Title)

	// --- 5. Delete a record ---
	fmt.Println("--- Deleting a record ---")
	if err := postsService.Delete(context.Background(), recordID); err != nil {
		log.Fatalf("Failed to delete record %s: %v", recordID, err)
	}
	fmt.Printf("Record %s successfully deleted.\n", recordID)
}
