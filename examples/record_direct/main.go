// Package main demonstrates using Record objects directly with PocketBase Go client.
// This example shows dynamic record manipulation without predefined structs.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// Initialize PocketBase client
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// Authenticate as an admin (or user) to have permission to modify data
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Create Record service for posts collection
	recordsService := client.Records("posts")

	// --- 1. Create new record using Record object directly ---
	fmt.Println("--- Creating new record with Record object ---")
	newRecord := &pocketbase.Record{}
	newRecord.Set("title", "Direct Record Post")
	newRecord.Set("content", "This is an example using Record objects directly!")

	createdRecord, err := recordsService.Create(context.Background(), newRecord)
	if err != nil {
		log.Fatalf("Failed to create record: %v", err)
	}
	fmt.Printf("Created record ID: %s, Title: '%s'\n\n", createdRecord.ID, createdRecord.GetString("title"))

	// --- 2. List records ---
	fmt.Println("--- Listing records ---")
	records, err := recordsService.GetList(context.Background(), &pocketbase.ListOptions{
		PerPage: 10,
		Sort:    "-created", // Latest first
	})
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}
	fmt.Printf("Found %d total record(s).\n", records.TotalItems)
	for i, record := range records.Items {
		fmt.Printf("  %d: ID=%s, Title='%s', Created=%s\n",
			i+1,
			record.ID,
			record.GetString("title"),
			record.GetDateTime("created").String())
	}
	fmt.Println()

	// --- 3. Get a single record ---
	fmt.Println("--- Getting a single record ---")
	recordID := createdRecord.ID
	retrievedRecord, err := recordsService.GetOne(context.Background(), recordID, &pocketbase.GetOneOptions{
		Expand: "user", // Expand related records if any
	})
	if err != nil {
		log.Fatalf("Failed to get record %s: %v", recordID, err)
	}
	fmt.Printf("Retrieved record:\n")
	fmt.Printf("  ID: %s\n", retrievedRecord.ID)
	fmt.Printf("  Title: %s\n", retrievedRecord.GetString("title"))
	fmt.Printf("  Content: %s\n", retrievedRecord.GetString("content"))
	fmt.Printf("  Created: %s\n", retrievedRecord.GetDateTime("created").String())
	fmt.Printf("  Updated: %s\n\n", retrievedRecord.GetDateTime("updated").String())

	// --- 4. Update record using Record object ---
	fmt.Println("--- Updating record with Record object ---")
	updateRecord := &pocketbase.Record{}
	updateRecord.Set("title", "Updated Title")
	updateRecord.Set("content", "Content updated using Record object.")

	updatedRecord, err := recordsService.Update(context.Background(), recordID, updateRecord)
	if err != nil {
		log.Fatalf("Failed to update record %s: %v", recordID, err)
	}
	fmt.Printf("Updated record:\n")
	fmt.Printf("  Title: %s\n", updatedRecord.GetString("title"))
	fmt.Printf("  Content: %s\n\n", updatedRecord.GetString("content"))

	// --- 5. Filtering records ---
	fmt.Println("--- Filtering records ---")
	filteredRecords, err := recordsService.GetList(context.Background(), &pocketbase.ListOptions{
		Filter: "title ~ 'Updated'", // Records with 'Updated' in title
		Fields: "id,title,created",  // Specific fields only
	})
	if err != nil {
		log.Fatalf("Failed to get filtered records: %v", err)
	}
	fmt.Printf("Filter results: %d records\n", len(filteredRecords.Items))
	for _, record := range filteredRecords.Items {
		fmt.Printf("  ID: %s, Title: %s\n", record.ID, record.GetString("title"))
	}
	fmt.Println()

	// --- 6. Example with various data types ---
	fmt.Println("--- Example with various data types ---")
	complexRecord := &pocketbase.Record{}
	complexRecord.Set("title", "Complex Data Example")
	complexRecord.Set("content", "Contains various data types")
	complexRecord.Set("is_published", true)
	complexRecord.Set("view_count", 42)
	complexRecord.Set("rating", 4.5)
	complexRecord.Set("tags", []string{"golang", "pocketbase", "example"})

	complexCreated, err := recordsService.Create(context.Background(), complexRecord)
	if err != nil {
		log.Fatalf("Failed to create complex record: %v", err)
	}

	fmt.Printf("Complex record created:\n")
	fmt.Printf("  Title: %s\n", complexCreated.GetString("title"))
	fmt.Printf("  Published: %t\n", complexCreated.GetBool("is_published"))
	fmt.Printf("  View count: %.0f\n", complexCreated.GetFloat("view_count"))
	fmt.Printf("  Rating: %.1f\n", complexCreated.GetFloat("rating"))
	fmt.Printf("  Tags: %v\n\n", complexCreated.GetStringSlice("tags"))

	// --- 7. Delete records ---
	fmt.Println("--- Deleting records ---")
	if err := recordsService.Delete(context.Background(), recordID); err != nil {
		log.Fatalf("Failed to delete record %s: %v", recordID, err)
	}
	fmt.Printf("Record %s successfully deleted.\n", recordID)

	if err := recordsService.Delete(context.Background(), complexCreated.ID); err != nil {
		log.Fatalf("Failed to delete complex record %s: %v", complexCreated.ID, err)
	}
	fmt.Printf("Complex record %s successfully deleted.\n", complexCreated.ID)
}
