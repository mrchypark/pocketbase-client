package main

import (
	"context"
	"fmt"
	"log"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// 1. Initialize the client
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	// 2. Authenticate as a user (optional)
	// Replace with your actual collection name, username, and password
	auth, err := client.WithPassword(ctx, "users", "testuser", "1234567890")
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	fmt.Printf("Authenticated as user: %s\n", auth.Record.ID)

	// 3. Define a struct for your record data
	type Post struct {
		pocketbase.BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	// 4. Create a new record
	newPost := Post{
		Title:   "My First Post",
		Content: "Hello, PocketBase!",
	}
	createdRecord, err := client.Records.Create(ctx, "posts", newPost)
	if err != nil {
		log.Fatalf("Failed to create record: %v", err)
	}
	fmt.Printf("Created record with ID: %s\n", createdRecord.ID)

	// 5. List records from the 'posts' collection
	records, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}
	fmt.Printf("Found %d records.\n", records.TotalItems)
	for _, record := range records.Items {
		fmt.Printf("- Record ID: %s, Title: %s\n", record.ID, record.GetString("title"))
	}

	// 6. Update the record
	updateData := map[string]any{
		"title": "My Updated Post Title",
	}
	updatedRecord, err := client.Records.Update(ctx, "posts", createdRecord.ID, updateData)
	if err != nil {
		log.Fatalf("Failed to update record: %v", err)
	}
	fmt.Printf("Updated record title: %s\n", updatedRecord.GetString("title"))

	// 7. Delete the record
	if err := client.Records.Delete(ctx, "posts", createdRecord.ID); err != nil {
		log.Fatalf("Failed to delete record: %v", err)
	}
	fmt.Println("Record deleted successfully.")
}
