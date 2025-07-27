// Package main demonstrates a quick start guide for PocketBase Go client.
// This example shows basic authentication and CRUD operations using both typed and Record approaches.
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

	// 3. Define a struct for your record data (type-safe approach)
	type Post struct {
		pocketbase.BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	// Create type-safe service for posts collection
	postsService := pocketbase.NewRecordService[Post](client, "posts")

	// 4. Create a new record (type-safe approach)
	newPost := &Post{
		Title:   "My First Post",
		Content: "Hello, PocketBase!",
	}
	createdRecord, err := postsService.Create(ctx, newPost)
	if err != nil {
		log.Fatalf("Failed to create record: %v", err)
	}
	fmt.Printf("Created record with ID: %s\n", createdRecord.ID)

	// 5. List records from the 'posts' collection
	records, err := postsService.GetList(ctx, &pocketbase.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}
	fmt.Printf("Found %d records.\n", records.TotalItems)
	for _, record := range records.Items {
		fmt.Printf("- Record ID: %s, Title: %s\n", record.ID, record.Title)
	}

	// 6. Update the record (type-safe approach)
	updatePost := &Post{
		BaseModel: pocketbase.BaseModel{ID: createdRecord.ID},
		Title:     "My Updated Post Title",
		Content:   createdRecord.Content, // Keep existing content
	}
	updatedRecord, err := postsService.Update(ctx, createdRecord.ID, updatePost)
	if err != nil {
		log.Fatalf("Failed to update record: %v", err)
	}
	fmt.Printf("Updated record title: %s\n", updatedRecord.Title)

	// 7. Delete the record
	if err := postsService.Delete(ctx, createdRecord.ID); err != nil {
		log.Fatalf("Failed to delete record: %v", err)
	}
	fmt.Println("Record deleted successfully.")

	fmt.Println("\n--- Using Record objects directly ---")

	// Create Record service (dynamic approach)
	recordsService := client.Records("posts")

	// Create new record using Record object
	newRecord := &pocketbase.Record{}
	newRecord.Set("title", "Direct Record Post")
	newRecord.Set("content", "This is an example using Record objects directly!")

	createdDirectRecord, err := recordsService.Create(ctx, newRecord)
	if err != nil {
		log.Fatalf("Failed to create record using Record approach: %v", err)
	}
	fmt.Printf("Created record using Record approach - ID: %s, Title: %s\n",
		createdDirectRecord.ID,
		createdDirectRecord.GetString("title"))

	// Delete record using Record approach
	if err := recordsService.Delete(ctx, createdDirectRecord.ID); err != nil {
		log.Fatalf("Failed to delete record using Record approach: %v", err)
	}
	fmt.Println("Record deleted successfully using Record approach.")

	fmt.Println("\nQuick start guide completed!")
}
