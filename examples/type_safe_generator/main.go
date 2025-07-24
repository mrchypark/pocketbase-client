// Package main demonstrates the usage of type-safe generated models with PocketBase client.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mrchypark/pocketbase-client"
	"github.com/pocketbase/pocketbase/tools/types"
)

func main() {
	client := pocketbase.NewClient("http://127.0.0.1:8090") // PocketBase server URL

	ctx := context.Background()

	// 1. Authenticate with admin account.
	// In production environments, use environment variables for secure management.
	if _, err := client.WithAdminPassword(ctx, "admin@example.com", "1q2w3e4r5t"); err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}
	fmt.Println("--- Admin authenticated successfully. ---")

	// --- Create RelatedCollection record (for testing AllTypes Relation Single/Multi) ---
	fmt.Println("\n--- Creating a RelatedCollection record for testing relations ---")
	newRelated := NewRelatedCollection()
	newRelated.Name = "Example Related Item" // Direct field access

	// Use generic service to create record
	relatedService := NewRelatedCollectionService(client)
	createdRelated, err := relatedService.Create(ctx, newRelated, nil)
	if err != nil {
		log.Fatalf("Failed to create related record: %v", err)
	}
	fmt.Printf("Created RelatedCollection with ID: %s, Name: '%s'\n", createdRelated.ID, createdRelated.Name)

	// --- Create AllTypes record using Type-Safe models ---
	fmt.Println("\n--- Creating a new AllTypes record using generated types ---")

	newAllTypes := NewAllTypes() // Create new AllTypes instance

	// Set required fields (direct field access)
	newAllTypes.TextRequired = "This is a required text."
	newAllTypes.NumberRequired = 123.45
	newAllTypes.BoolRequired = true
	newAllTypes.EmailRequired = "test@example.com"
	newAllTypes.URLRequired = "https://example.com"
	newAllTypes.DateRequired = types.NowDateTime()

	// Assign valid values defined in schema to required select fields
	newAllTypes.SelectSingleRequired = "a"               // One of "a", "b", "c"
	newAllTypes.SelectMultiRequired = []string{"a", "b"} // One or more of "a", "b", "c"
	newAllTypes.JSONRequired = json.RawMessage(`{"key": "value", "number": 123}`)

	// Set optional fields (direct field access - no pointers needed in new structure)
	newAllTypes.TextOptional = "This is an optional text."
	newAllTypes.NumberOptional = 987.65
	newAllTypes.BoolOptional = false
	newAllTypes.EmailOptional = "optional@example.com"
	newAllTypes.URLOptional = "https://optional.dev"
	newAllTypes.DateOptional = types.NowDateTime().Add(24 * time.Hour)
	newAllTypes.SelectSingleOptional = "x"          // One of "x", "y", "z"
	newAllTypes.SelectMultiOptional = []string{"y"} // One or more of "x", "y", "z"
	newAllTypes.JSONOptional = json.RawMessage(`{"another_key": "another_value"}`)

	// File and Relation fields (in this example, actual file upload/ID references are omitted, using empty slices or generated IDs)
	// In production environments, use pocketbase-client's File-related methods.
	newAllTypes.RelationSingle = createdRelated.ID // Reference to RelatedCollection ID created above
	newAllTypes.RelationMulti = []string{}         // Add multiple RelatedCollection IDs here

	// Create record using generic service
	allTypesService := NewAllTypesService(client)
	createdAllTypeRecord, err := allTypesService.Create(ctx, newAllTypes, nil)
	if err != nil {
		log.Fatalf("Failed to create type-safe AllTypes record: %v", err)
	}

	// Verify created record (using direct field access)
	fmt.Printf("Created AllTypes record with ID: %s\n", createdAllTypeRecord.ID)
	fmt.Printf("  TextRequired: '%s'\n", createdAllTypeRecord.TextRequired)
	fmt.Printf("  TextOptional: '%s'\n", createdAllTypeRecord.TextOptional)
	fmt.Printf("  NumberRequired: %f\n", createdAllTypeRecord.NumberRequired)
	fmt.Printf("  NumberOptional: %f\n", createdAllTypeRecord.NumberOptional)
	fmt.Printf("  BoolRequired: %t\n", createdAllTypeRecord.BoolRequired)
	fmt.Printf("  BoolOptional: %t\n", createdAllTypeRecord.BoolOptional)
	fmt.Printf("  EmailRequired: '%s'\n", createdAllTypeRecord.EmailRequired)
	fmt.Printf("  EmailOptional: '%s'\n", createdAllTypeRecord.EmailOptional)
	fmt.Printf("  URLRequired: '%s'\n", createdAllTypeRecord.URLRequired)
	fmt.Printf("  URLOptional: '%s'\n", createdAllTypeRecord.URLOptional)
	fmt.Printf("  DateRequired: '%s'\n", createdAllTypeRecord.DateRequired.String())
	fmt.Printf("  DateOptional: '%s'\n", createdAllTypeRecord.DateOptional.String())
	fmt.Printf("  SelectSingleRequired: %s\n", createdAllTypeRecord.SelectSingleRequired)
	fmt.Printf("  SelectSingleOptional: '%s'\n", createdAllTypeRecord.SelectSingleOptional)
	fmt.Printf("  SelectMultiRequired: %v\n", createdAllTypeRecord.SelectMultiRequired)
	fmt.Printf("  SelectMultiOptional: %v\n", createdAllTypeRecord.SelectMultiOptional)
	fmt.Printf("  JSONRequired: %s\n", string(createdAllTypeRecord.JSONRequired))
	fmt.Printf("  JSONOptional: %s\n", string(createdAllTypeRecord.JSONOptional))
	fmt.Printf("  FileSingle (IDs): '%s'\n", createdAllTypeRecord.FileSingle)
	fmt.Printf("  FileMulti (IDs): %v\n", createdAllTypeRecord.FileMulti)
	fmt.Printf("  RelationSingle (IDs): '%s'\n", createdAllTypeRecord.RelationSingle)
	fmt.Printf("  RelationMulti (IDs): %v\n", createdAllTypeRecord.RelationMulti)

	// --- List AllTypes records using Type-Safe helpers ---
	fmt.Println("\n--- Listing AllTypes records using generated helper ---")

	allTypesCollection, err := GetAllTypesList(client, &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
		Sort:    "-created", // Sort by newest created records first
	})
	if err != nil {
		log.Fatalf("Failed to list AllTypes: %v", err)
	}

	fmt.Printf("Found %d AllTypes records. (Page %d/%d)\n", allTypesCollection.TotalItems, allTypesCollection.Page, allTypesCollection.TotalPages)
	for i, item := range allTypesCollection.Items {
		// Use direct field access to safely access fields even within loops.
		fmt.Printf("  [%d] ID: %s, TextRequired: '%s'\n", i+1, item.ID, item.TextRequired)
		// Add more fields here if you want to output more
	}

	// --- Fetch single AllTypes record ---
	fmt.Println("\n--- Fetching a single AllTypes record by ID ---")
	fetchedAllTypes, err := GetAllTypes(client, createdAllTypeRecord.ID, nil)
	if err != nil {
		log.Fatalf("Failed to fetch single AllTypes record: %v", err)
	}
	fmt.Printf("Fetched AllTypes record (ID: %s) TextRequired: '%s'\n", fetchedAllTypes.ID)

	// --- Delete created records (cleanup) ---
	fmt.Printf("\n--- Cleaning up created AllTypes record (ID: %s) ---\n", createdAllTypeRecord.ID)
	if err := client.Records.Delete(ctx, "all_types", createdAllTypeRecord.ID); err != nil {
		log.Printf("Failed to delete AllTypes record %s during cleanup: %v", createdAllTypeRecord.ID, err)
	} else {
		fmt.Println("AllTypes record cleanup complete.")
	}

	fmt.Printf("\n--- Cleaning up created RelatedCollection record (ID: %s) ---\n", createdRelated.ID)
	if err := client.Records.Delete(ctx, "related_collection", createdRelated.ID); err != nil {
		log.Printf("Failed to delete RelatedCollection record %s during cleanup: %v", createdRelated.ID, err)
	} else {
		fmt.Println("RelatedCollection record cleanup complete.")
	}

	fmt.Println("\n--- Example execution finished. ---")
}
