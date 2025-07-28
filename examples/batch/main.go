// Package main demonstrates batch operations using PocketBase Go client.
// This example shows how to execute multiple CRUD operations in a single API call.
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
	// Initialize PocketBase client
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// Authenticate as admin to have permission to modify data
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Create Record service for posts collection
	postsService := client.Records("posts")

	// --- Prepare Batch Requests ---
	// We will create two new posts and update one existing post in a single API call.

	// Request 1: Create the first post
	post1 := &pocketbase.Record{}
	post1.Set("title", "Batch Post 1")
	post1.Set("content", "Content for the first batch post.")

	createReq1, err := postsService.NewCreateRequest(post1)
	if err != nil {
		log.Fatalf("Failed to prepare first create request: %v", err)
	}

	// Request 2: Create the second post
	post2 := &pocketbase.Record{}
	post2.Set("title", "Batch Post 2")
	post2.Set("content", "Content for the second batch post.")

	createReq2, err := postsService.NewCreateRequest(post2)
	if err != nil {
		log.Fatalf("Failed to prepare second create request: %v", err)
	}

	// First, create a record that we can update in the batch operation.
	initialRecord := &pocketbase.Record{}
	initialRecord.Set("title", "Record to be updated")
	initialRecord.Set("content", "This record will be updated in the batch.")

	createdInitial, err := postsService.Create(context.Background(), initialRecord)
	if err != nil {
		log.Fatalf("Failed to create initial record for update: %v", err)
	}
	fmt.Printf("Initial record created with ID: %s\n", createdInitial.ID)

	// Request 3: Update the record we just created
	updateRecord := &pocketbase.Record{}
	updateRecord.Set("title", "Title Updated via Batch")
	updateRecord.Set("content", "Content updated via batch operation.")

	updateReq, err := postsService.NewUpdateRequest(createdInitial.ID, updateRecord)
	if err != nil {
		log.Fatalf("Failed to prepare update request: %v", err)
	}

	// --- Execute the Batch Request ---
	fmt.Println("\n--- Executing batch request ---")
	requests := []*pocketbase.BatchRequest{createReq1, createReq2, updateReq}
	responses, err := client.Batch.Execute(context.Background(), requests)
	if err != nil {
		log.Fatalf("Batch execution failed: %v", err)
	}

	// --- Process Batch Responses ---
	fmt.Println("--- Processing batch responses ---")
	var createdIDs []string

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
				recordID := bodyMap["id"].(string)
				recordTitle := bodyMap["title"].(string)
				fmt.Printf("  Record ID: %s\n", recordID)
				fmt.Printf("  Record Title: %s\n", recordTitle)

				// Collect created record IDs for cleanup
				if i < 2 { // First two requests are create requests
					createdIDs = append(createdIDs, recordID)
				}
			}
		}
		fmt.Println()
	}

	// --- Cleanup: Delete created records ---
	fmt.Println("--- Cleaning up created records ---")

	// Add the updated record to deletion list as well
	allIDsToDelete := append(createdIDs, createdInitial.ID)

	for _, id := range allIDsToDelete {
		if err := postsService.Delete(context.Background(), id); err != nil {
			log.Printf("Failed to delete record %s during cleanup: %v", id, err)
		} else {
			fmt.Printf("Deleted record %s\n", id)
		}
	}

	fmt.Println("\nBatch operation example completed!")

	// --- Additional Example: Advanced batch operations ---
	fmt.Println("\n--- Advanced Batch Operations Example ---")

	// Create multiple records
	var advancedRequests []*pocketbase.BatchRequest

	for i := 1; i <= 3; i++ {
		record := &pocketbase.Record{}
		record.Set("title", fmt.Sprintf("Advanced Batch Post %d", i))
		record.Set("content", fmt.Sprintf("This is the %d post created by advanced batch operation.", i))

		req, err := postsService.NewCreateRequest(record)
		if err != nil {
			log.Printf("Failed to prepare advanced batch request %d: %v", i, err)
			continue
		}
		advancedRequests = append(advancedRequests, req)
	}

	if len(advancedRequests) > 0 {
		advancedResponses, err := client.Batch.Execute(context.Background(), advancedRequests)
		if err != nil {
			log.Printf("Advanced batch execution failed: %v", err)
		} else {
			fmt.Printf("Advanced batch processed %d records.\n", len(advancedResponses))

			// Clean up created records
			for _, res := range advancedResponses {
				if res.Status == http.StatusOK {
					if bodyMap, ok := res.Body.(map[string]any); ok {
						if id, ok := bodyMap["id"].(string); ok {
							postsService.Delete(context.Background(), id)
						}
					}
				}
			}
		}
	}
}
