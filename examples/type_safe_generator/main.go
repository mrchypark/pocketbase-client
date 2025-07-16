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
	newRelated.SetName("Example Related Item")
	createdRelated, err := client.Records.Create(ctx, "related_collection", newRelated)
	if err != nil {
		log.Fatalf("Failed to create related record: %v", err)
	}
	fmt.Printf("Created RelatedCollection with ID: %s, Name: '%s'\n", createdRelated.ID, newRelated.Name())

	// --- Create AllTypes record using Type-Safe models ---
	fmt.Println("\n--- Creating a new AllTypes record using generated types ---")

	newAllTypes := NewAllTypes() // Create new AllTypes instance

	// Set required fields
	newAllTypes.SetTextRequired("This is a required text.")
	newAllTypes.SetNumberRequired(123.45)
	newAllTypes.SetBoolRequired(true)
	newAllTypes.SetEmailRequired("test@example.com")
	newAllTypes.SetURLRequired("https://example.com")
	newAllTypes.SetDateRequired(types.NowDateTime())

	// Assign valid values defined in schema to required select fields
	newAllTypes.SetSelectSingleRequired("a")               // One of "a", "b", "c"
	newAllTypes.SetSelectMultiRequired([]string{"a", "b"}) // One or more of "a", "b", "c"
	newAllTypes.SetJSONRequired(json.RawMessage(`{"key": "value", "number": 123}`))

	// Set optional fields (using pointer values)
	optionalText := "This is an optional text."
	newAllTypes.SetTextOptional(&optionalText)
	optionalNumber := 987.65
	newAllTypes.SetNumberOptional(&optionalNumber)
	optionalBool := false
	newAllTypes.SetBoolOptional(&optionalBool)
	optionalEmail := "optional@example.com"
	newAllTypes.SetEmailOptional(&optionalEmail)
	optionalURL := "https://optional.dev"
	newAllTypes.SetURLOptional(&optionalURL)
	optionalDate := types.NowDateTime().Add(24 * time.Hour)
	newAllTypes.SetDateOptional(&optionalDate)
	so := "x"
	newAllTypes.SetSelectSingleOptional(&so)          // One of "x", "y", "z"
	newAllTypes.SetSelectMultiOptional([]string{"y"}) // One or more of "x", "y", "z"
	optionalJSONContent := json.RawMessage(`{"another_key": "another_value"}`)
	newAllTypes.SetJSONOptional(optionalJSONContent)

	// File and Relation fields (in this example, actual file upload/ID references are omitted, using empty slices or generated IDs)
	// In production environments, use pocketbase-client's File-related methods.
	id := createdRelated.ID
	newAllTypes.SetRelationSingle(&id)       // Reference to RelatedCollection ID created above
	newAllTypes.SetRelationMulti([]string{}) // Add multiple RelatedCollection IDs here

	// Create record
	createdAllTypeRecord, err := client.Records.Create(ctx, "all_types", newAllTypes)
	if err != nil {
		log.Fatalf("Failed to create type-safe AllTypes record: %v", err)
	}

	// Verify created record (using type-safe Getters)
	createdAllTypes := ToAllTypes(createdAllTypeRecord)
	fmt.Printf("Created AllTypes record with ID: %s\n", createdAllTypes.ID)
	fmt.Printf("  TextRequired: '%s'\n", createdAllTypes.TextRequired())
	if txt := createdAllTypes.TextOptional(); txt != nil {
		fmt.Printf("  TextOptional: '%s'\n", *txt) // Pointer dereference
	} else {
		fmt.Printf("  TextOptional: <nil>\n") // When nil
	}
	fmt.Printf("  NumberRequired: %f\n", createdAllTypes.NumberRequired())
	if num := createdAllTypes.NumberOptional(); num != nil {
		fmt.Printf("  NumberOptional: %f\n", *num) // Pointer dereference
	} else {
		fmt.Printf("  NumberOptional: <nil>\n")
	}
	fmt.Printf("  BoolRequired: %t\n", createdAllTypes.BoolRequired())
	if b := createdAllTypes.BoolOptional(); b != nil {
		fmt.Printf("  BoolOptional: %t\n", *b) // Pointer dereference
	} else {
		fmt.Printf("  BoolOptional: <nil>\n")
	}
	fmt.Printf("  EmailRequired: '%s'\n", createdAllTypes.EmailRequired())
	if email := createdAllTypes.EmailOptional(); email != nil {
		fmt.Printf("  EmailOptional: '%s'\n", *email) // Pointer dereference
	} else {
		fmt.Printf("  EmailOptional: <nil>\n")
	}
	fmt.Printf("  URLRequired: '%s'\n", createdAllTypes.URLRequired())
	if url := createdAllTypes.URLOptional(); url != nil {
		fmt.Printf("  URLOptional: '%s'\n", *url) // Pointer dereference
	} else {
		fmt.Printf("  URLOptional: <nil>\n")
	}
	fmt.Printf("  DateRequired: '%s'\n", createdAllTypes.DateRequired().String())
	if dt := createdAllTypes.DateOptional(); dt != nil {
		fmt.Printf("  DateOptional: '%s'\n", dt.String()) // Pointer dereference (String() method returns value)
	} else {
		fmt.Printf("  DateOptional: <nil>\n")
	}

	// SelectSingleRequired is string, SelectSingleOptional is *string, so handle branching
	fmt.Printf("  SelectSingleRequired: %s\n", createdAllTypes.SelectSingleRequired())
	if sso := createdAllTypes.SelectSingleOptional(); sso != nil {
		fmt.Printf("  SelectSingleOptional: '%s'\n", *sso) // *string dereference
	} else {
		fmt.Printf("  SelectSingleOptional: <nil>\n")
	}

	fmt.Printf("  SelectMultiRequired: %v\n", createdAllTypes.SelectMultiRequired())
	fmt.Printf("  SelectMultiOptional: %v\n", createdAllTypes.SelectMultiOptional())
	fmt.Printf("  JSONRequired: %s\n", string(createdAllTypes.JSONRequired()))
	fmt.Printf("  JSONOptional: %s\n", string(createdAllTypes.JSONOptional()))

	// FileSingle is *string, FileMulti is []string, so handle branching
	if fs := createdAllTypes.FileSingle(); fs != nil {
		fmt.Printf("  FileSingle (IDs): '%s'\n", *fs) // *string dereference
	} else {
		fmt.Printf("  FileSingle (IDs): <nil>\n")
	}
	fmt.Printf("  FileMulti (IDs): %v\n", createdAllTypes.FileMulti())

	// RelationSingle is *string, RelationMulti is []string, so handle branching
	if rs := createdAllTypes.RelationSingle(); rs != nil {
		fmt.Printf("  RelationSingle (IDs): '%s'\n", *rs) // *string dereference
	} else {
		fmt.Printf("  RelationSingle (IDs): <nil>\n")
	}
	fmt.Printf("  RelationMulti (IDs): %v\n", createdAllTypes.RelationMulti())

	// --- List AllTypes records using Type-Safe helpers ---
	fmt.Println("\n--- Listing AllTypes records using generated helper ---")

	allTypesCollection, err := GetAllTypesList(client.Records, &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
		Sort:    "-created", // Sort by newest created records first
	})
	if err != nil {
		log.Fatalf("Failed to list AllTypes: %v", err)
	}

	fmt.Printf("Found %d AllTypes records. (Page %d/%d)\n", allTypesCollection.TotalItems, allTypesCollection.Page, allTypesCollection.TotalPages)
	for i, item := range allTypesCollection.Items {
		// Use type-safe Getters to safely access fields even within loops.
		fmt.Printf("  [%d] ID: %s, TextRequired: '%s'\n", i+1, item.ID, item.TextRequired())
		// Add more fields here if you want to output more
	}

	// --- Fetch single AllTypes record ---
	fmt.Println("\n--- Fetching a single AllTypes record by ID ---")
	fetchedAllTypes, err := GetAllTypes(client.Records, createdAllTypeRecord.ID, nil)
	if err != nil {
		log.Fatalf("Failed to fetch single AllTypes record: %v", err)
	}
	fmt.Printf("Fetched AllTypes record (ID: %s) TextRequired: '%s'\n", fetchedAllTypes.ID, fetchedAllTypes.TextRequired())

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
