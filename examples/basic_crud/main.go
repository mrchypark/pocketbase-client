package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/cypark/pocketbase-client"
	"github.com/cypark/pocketbase-client/tools/list"
)

// Define a struct that matches your collection's schema.
// The `json` tags are important for serialization.
type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func main() {
	// Initialize the PocketBase client
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// Authenticate as an admin (or user) to have permission to modify data
	if _, err := client.AuthWithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	// --- 1. Create a new record ---
	fmt.Println("--- Creating a new record ---")
	newPost := Post{
		Title:   "My First Post",
		Content: "Hello from the Go SDK!",
	}
	createdRecord, err := client.Records.Create(context.Background(), "posts", newPost)
	if err != nil {
		log.Fatalf("Failed to create record: %v", err)
	}
	fmt.Printf("Created record with ID: %s and title: '%s'\n\n", createdRecord.ID, createdRecord.GetString("title"))

	// --- 2. Get a list of records ---
	fmt.Println("--- Listing records ---")
	// Use list.NewOptions() for default pagination, or customize it.
	// Example: list.NewOptions(list.WithPage(1), list.WithPerPage(10), list.WithSort("-created"))
	records, err := client.Records.GetList(context.Background(), "posts", list.NewOptions())
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}
	fmt.Printf("Found %d total record(s).\n", records.TotalItems)
	for i, record := range records.Items {
		fmt.Printf("  %d: ID=%s, Title='%s'\n", i+1, record.ID, record.GetString("title"))
	}
	fmt.Println()

	// --- 3. Get a single record ---
	fmt.Println("--- Getting a single record ---")
	recordID := createdRecord.ID
	retrievedRecord, err := client.Records.GetOne(context.Background(), "posts", recordID, nil)
	if err != nil {
		log.Fatalf("Failed to get record %s: %v", recordID, err)
	}
	fmt.Printf("Retrieved record with title: '%s' and content: '%s'\n\n", retrievedRecord.GetString("title"), retrievedRecord.GetString("content"))

	// --- 4. Update a record ---
	fmt.Println("--- Updating a record ---")
	updateData := map[string]interface{}{
		"title": "My Updated Post Title",
	}
	updatedRecord, err := client.Records.Update(context.Background(), "posts", recordID, updateData)
	if err != nil {
		log.Fatalf("Failed to update record %s: %v", recordID, err)
	}
	fmt.Printf("Updated record title to: '%s'\n\n", updatedRecord.GetString("title"))

	// --- 5. Delete a record ---
	fmt.Println("--- Deleting a record ---")
	if err := client.Records.Delete(context.Background(), "posts", recordID); err != nil {
		log.Fatalf("Failed to delete record %s: %v", recordID, err)
	}
	fmt.Printf("Successfully deleted record %s.\n", recordID)
}