package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	// Initialize the PocketBase client
	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))

	// Authenticate as an admin to ensure we have access
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	// --- Setup: Create some sample records to work with ---
	fmt.Println("--- Creating sample records ---")
	sampleTitles := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}
	for _, title := range sampleTitles {
		_, err := client.Records.Create(context.Background(), "posts", map[string]any{
			"title":        title,
			"content":      "Some content here.",
			"is_published": true,
		})
		if err != nil {
			log.Fatalf("Failed to create sample record '%s': %v", title, err)
		}
	}
	fmt.Printf("Created %d sample records.\n\n", len(sampleTitles))

	// --- Example 1: Basic Pagination ---
	fmt.Println("--- Example 1: Basic Pagination (Page 2, 2 items per page) ---")
	paginatedResult, err := client.Records.GetList(context.Background(), "posts",
		&pocketbase.ListOptions{
			Page:    2,
			PerPage: 2,
		},
	)
	if err != nil {
		log.Fatalf("Failed to get paginated list: %v", err)
	}
	fmt.Printf("Page: %d, PerPage: %d, TotalItems: %d, TotalPages: %d\n",
		paginatedResult.Page, paginatedResult.PerPage, paginatedResult.TotalItems, paginatedResult.TotalPages)
	for _, record := range paginatedResult.Items {
		fmt.Printf("  - ID: %s, Title: %s\n", record.ID, record.GetString("title"))
	}
	fmt.Println()

	// --- Example 2: Sorting ---
	fmt.Println("--- Example 2: Sorting by title in descending order ---")
	sortedResult, err := client.Records.GetList(context.Background(), "posts",
		&pocketbase.ListOptions{
			Sort: "-title",
		},
	)
	if err != nil {
		log.Fatalf("Failed to get sorted list: %v", err)
	}
	for _, record := range sortedResult.Items {
		fmt.Printf("  - Title: %s\n", record.GetString("title"))
	}
	fmt.Println()

	// --- Example 3: Filtering ---
	// Filter syntax is the same as in the PocketBase API documentation.
	fmt.Println("--- Example 3: Filtering for titles starting with 'B' or 'C' ---")
	filteredResult, err := client.Records.GetList(context.Background(), "posts",
		&pocketbase.ListOptions{
			Filter: "title ~ 'B%' || title ~ 'C%'",
		},
	)
	if err != nil {
		log.Fatalf("Failed to get filtered list: %v", err)
	}
	for _, record := range filteredResult.Items {
		fmt.Printf("  - Title: %s\n", record.GetString("title"))
	}
	fmt.Println()

	// --- Example 4: Selecting specific fields ---
	fmt.Println("--- Example 4: Selecting only the 'title' and 'created' fields ---")
	// Note: The base fields (id, created, updated, collectionId, collectionName) are always returned.
	selectedFieldsResult, err := client.Records.GetList(context.Background(), "posts",
		&pocketbase.ListOptions{
			Fields:  "title,created",
			Sort:    "-title",
			PerPage: 2,
		},
	)
	if err != nil {
		log.Fatalf("Failed to get list with selected fields: %v", err)
	}
	for _, record := range selectedFieldsResult.Items {
		// The 'content' field will be empty (zero-value)
		fmt.Printf("  - Title: %s, Content: '%s', Created: %s\n",
			record.GetString("title"), record.GetString("content"), record.GetString("created"))
	}
	fmt.Println()

	// --- Cleanup: Delete all records from the 'posts' collection ---
	fmt.Println("--- Cleaning up all created records ---")
	allRecords, _ := client.Records.GetList(context.Background(), "posts", &pocketbase.ListOptions{PerPage: 100})
	for _, record := range allRecords.Items {
		if err := client.Records.Delete(context.Background(), "posts", record.ID); err != nil {
			log.Printf("Failed to delete record %s during cleanup: %v", record.ID, err)
		}
	}
	fmt.Println("Cleanup complete.")
}
