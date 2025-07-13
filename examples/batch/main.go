package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// Initialize the PocketBase client
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// Authenticate as an admin to have permission to modify data
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	// --- Prepare Batch Requests ---
	// We will create two new posts and update one existing post in a single API call.

	// Request 1: Create the first post
	createReq1, _ := client.Records.NewCreateRequest("posts", map[string]any{
		"title":   "Batch Post 1",
		"content": "Content for the first batch post.",
	})

	// Request 2: Create the second post
	createReq2, _ := client.Records.NewCreateRequest("posts", map[string]any{
		"title":   "Batch Post 2",
		"content": "Content for the second batch post.",
	})

	// First, create a record that we can update in the batch operation.
	initialRecord, err := client.Records.Create(context.Background(), "posts", map[string]any{
		"title": "Record to be updated",
	})
	if err != nil {
		log.Fatalf("Failed to create initial record for update: %v", err)
	}
	fmt.Printf("Created initial record with ID: %s\n", initialRecord.ID)

	// Request 3: Update the record we just created
	updateReq, _ := client.Records.NewUpdateRequest("posts", initialRecord.ID, map[string]any{
		"title": "Title Updated via Batch",
	})

	// --- Execute the Batch Request ---
	fmt.Println("\n--- Executing batch request ---")
	requests := []*pocketbase.BatchRequest{createReq1, createReq2, updateReq}
	responses, err := client.Batch.Execute(context.Background(), requests)
	if err != nil {
		log.Fatalf("Batch execution failed: %v", err)
	}

	// --- Process Batch Responses ---
	fmt.Println("--- Processing batch responses ---")
	for i, res := range responses {
		fmt.Printf("Response for Request #%d:\n", i+1)
		fmt.Printf("  Status: %d\n", res.Status)

		if res.Status >= http.StatusBadRequest {
			// If the request failed, ParsedError will contain structured error info.
			// The raw body is also available in res.Body.
			fmt.Printf("  Error: %v\n", res.ParsedError)
		} else {
			// If successful, the response body is in res.Body.
			// We can assert it to a map[string]any to inspect it.
			if bodyMap, ok := res.Body.(map[string]any); ok {
				fmt.Printf("  Record ID: %s\n", bodyMap["id"])
				fmt.Printf("  Record Title: %s\n", bodyMap["title"])
			}
		}
		fmt.Println()
	}

	// --- Cleanup: Delete the records we created ---
	fmt.Println("--- Cleaning up created records ---")
	// We need to find the IDs of the newly created records from the successful responses.
	var idsToDelete []string
	idsToDelete = append(idsToDelete, initialRecord.ID) // The one we updated
	for _, res := range responses {
		if res.Status == http.StatusOK {
			if bodyMap, ok := res.Body.(map[string]any); ok {
				// Only add IDs from the CREATE operations
				if id, ok := bodyMap["id"].(string); ok && bodyMap["title"] != "Title Updated via Batch" {
					idsToDelete = append(idsToDelete, id)
				}
			}
		}
	}

	for _, id := range idsToDelete {
		if err := client.Records.Delete(context.Background(), "posts", id); err != nil {
			log.Printf("Failed to delete record %s during cleanup: %v", id, err)
		} else {
			fmt.Printf("Deleted record %s\n", id)
		}
	}
}
