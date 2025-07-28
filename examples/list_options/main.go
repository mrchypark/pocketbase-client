// Package main demonstrates various list options and query features with PocketBase Go client.
// This example shows pagination, sorting, filtering, and field selection using Record objects.
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

	// Authenticate as admin to ensure we have access
	if _, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123"); err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	// Create Record service for posts collection
	postsService := client.Records("posts")

	// --- Setup: Create some sample records to work with ---
	fmt.Println("--- Creating sample records ---")
	sampleTitles := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}
	var createdIDs []string

	for _, title := range sampleTitles {
		record := &pocketbase.Record{}
		record.Set("title", title)
		record.Set("content", "Some content here.")
		record.Set("is_published", true)

		created, err := postsService.Create(context.Background(), record)
		if err != nil {
			log.Fatalf("Failed to create sample record '%s': %v", title, err)
		}
		createdIDs = append(createdIDs, created.ID)
	}
	fmt.Printf("Created %d sample records.\n\n", len(sampleTitles))

	// --- Example 1: Basic Pagination ---
	fmt.Println("--- Example 1: Basic Pagination (Page 2, 2 items per page) ---")
	paginatedResult, err := postsService.GetList(context.Background(),
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
	sortedResult, err := postsService.GetList(context.Background(),
		&pocketbase.ListOptions{
			Sort: "-title", // Descending order
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
	filteredResult, err := postsService.GetList(context.Background(),
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
	selectedFieldsResult, err := postsService.GetList(context.Background(),
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
			record.GetString("title"), record.GetString("content"), record.GetDateTime("created").String()[:19])
	}
	fmt.Println()

	// --- Example 5: Complex query options ---
	fmt.Println("--- Example 5: Complex query options (filtering + sorting + pagination) ---")
	complexResult, err := postsService.GetList(context.Background(),
		&pocketbase.ListOptions{
			Filter:  "is_published = true",
			Sort:    "title",
			Page:    1,
			PerPage: 3,
			Fields:  "id,title,is_published,created",
		},
	)
	if err != nil {
		log.Fatalf("Complex query failed: %v", err)
	}
	fmt.Printf("Showing %d of %d published records:\n", len(complexResult.Items), complexResult.TotalItems)
	for i, record := range complexResult.Items {
		fmt.Printf("  %d. %s (Published: %t, Created: %s)\n",
			i+1,
			record.GetString("title"),
			record.GetBool("is_published"),
			record.GetDateTime("created").String()[:10]) // Date only
	}
	fmt.Println()

	// --- Example 6: Advanced filtering ---
	fmt.Println("--- Example 6: Advanced filtering (text search) ---")
	advancedResult, err := postsService.GetList(context.Background(),
		&pocketbase.ListOptions{
			Filter: "title ~ 'A%' || content ~ 'content'",
			Sort:   "-created",
		},
	)
	if err != nil {
		log.Printf("Advanced filtering failed: %v", err)
	} else {
		fmt.Printf("Text search results: %d records\n", len(advancedResult.Items))
		for _, record := range advancedResult.Items {
			fmt.Printf("  - %s\n", record.GetString("title"))
		}
	}
	fmt.Println()

	// --- Cleanup: Delete all created records ---
	fmt.Println("--- Cleaning up all created records ---")
	for _, id := range createdIDs {
		if err := postsService.Delete(context.Background(), id); err != nil {
			log.Printf("Failed to delete record %s during cleanup: %v", id, err)
		}
	}
	fmt.Println("Cleanup complete.")

	fmt.Println("\nList options example completed!")
}
