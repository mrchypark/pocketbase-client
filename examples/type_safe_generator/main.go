package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mrchypark/pocketbase-client"
	"github.com/pocketbase/pocketbase/tools/types"
)

func main() {
	// Create PocketBase client
	client := pocketbase.NewClient("http://localhost:8090")
	ctx := context.Background()

	fmt.Println("🚀 PocketBase Type-Safe Generator Example")
	fmt.Println("=========================================")

	// 1. AllTypes collection usage example
	fmt.Println("\n📝 AllTypes Collection CRUD Example")
	if err := demonstrateAllTypes(ctx, client); err != nil {
		log.Printf("Error running AllTypes example: %v", err)
	}

	// 2. RelatedCollection collection usage example
	fmt.Println("\n🔗 RelatedCollection Collection Example")
	if err := demonstrateRelatedCollection(ctx, client); err != nil {
		log.Printf("Error running RelatedCollection example: %v", err)
	}

	// 3. Type safety demonstration
	fmt.Println("\n🛡️ Type Safety Demo")
	demonstrateTypeSafety()

	fmt.Println("\n✅ All examples completed!")
}

// AllTypes collection CRUD operations demo
func demonstrateAllTypes(ctx context.Context, client *pocketbase.Client) error {
	// Create type-safe service
	allTypesService := NewAllTypesService(client)

	// 1. Create new record
	fmt.Println("  📝 Creating new AllTypes record...")

	// Current time
	now, _ := types.ParseDateTime(time.Now().Format(time.RFC3339))

	newRecord := &AllTypes{
		TextRequired:         "Required text field",
		TextOptional:         &[]string{"Optional text"}[0],
		NumberRequired:       42.5,
		NumberOptional:       &[]float64{3.14}[0],
		BoolRequired:         true,
		BoolOptional:         &[]bool{false}[0],
		EmailRequired:        "test@example.com",
		EmailOptional:        &[]string{"optional@example.com"}[0],
		URLRequired:          "https://example.com",
		URLOptional:          &[]string{"https://optional.com"}[0],
		DateRequired:         now,
		DateOptional:         &now,
		SelectSingleRequired: "option1",
		SelectSingleOptional: &[]string{"option2"}[0],
		SelectMultiRequired:  []string{"multi1", "multi2"},
		SelectMultiOptional:  []string{"multi3", "multi4"},
		JSONRequired:         []byte(`{"required": true}`),
		JSONOptional:         []byte(`{"optional": true}`),
		FileSingle:           &[]string{"single_file.jpg"}[0],
		FileMulti:            []string{"file1.jpg", "file2.png"},
	}

	createdRecord, err := allTypesService.Create(ctx, newRecord)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	fmt.Printf("  ✅ Created record ID: %s\n", createdRecord.ID)

	// 2. Retrieve record
	fmt.Println("  🔍 Retrieving record...")
	retrievedRecord, err := allTypesService.GetOne(ctx, createdRecord.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve record: %w", err)
	}

	// Type-safe field access
	fmt.Printf("  📄 Text field: %s\n", retrievedRecord.TextRequired)
	fmt.Printf("  🔢 Number field: %.2f\n", retrievedRecord.NumberRequired)
	fmt.Printf("  ✅ Boolean field: %t\n", retrievedRecord.BoolRequired)
	fmt.Printf("  📧 Email field: %s\n", retrievedRecord.EmailRequired)

	// Safe access to optional fields
	if retrievedRecord.TextOptional != nil {
		fmt.Printf("  📝 Optional text: %s\n", *retrievedRecord.TextOptional)
	}
	if retrievedRecord.NumberOptional != nil {
		fmt.Printf("  🔢 Optional number: %.2f\n", *retrievedRecord.NumberOptional)
	}

	// 3. Update record
	fmt.Println("  ✏️ Updating record...")
	retrievedRecord.TextRequired = "Updated text"
	retrievedRecord.NumberRequired = 99.9

	updatedRecord, err := allTypesService.Update(ctx, retrievedRecord.ID, retrievedRecord)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	fmt.Printf("  ✅ Update completed: %s\n", updatedRecord.TextRequired)

	// 4. List records
	fmt.Println("  📋 Listing records...")
	listResult, err := allTypesService.GetList(ctx, &pocketbase.ListOptions{
		PerPage: 5,
		Sort:    "-created",
	})
	if err != nil {
		return fmt.Errorf("failed to list records: %w", err)
	}
	fmt.Printf("  📊 Retrieved %d out of %d total records\n", len(listResult.Items), listResult.TotalItems)

	// 5. Delete record
	fmt.Println("  🗑️ Deleting record...")
	if err := allTypesService.Delete(ctx, createdRecord.ID); err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	fmt.Println("  ✅ Deletion completed")

	return nil
}

// RelatedCollection collection usage demo
func demonstrateRelatedCollection(ctx context.Context, client *pocketbase.Client) error {
	// Create type-safe service
	relatedService := NewRelatedCollectionService(client)

	// Create new related collection record
	fmt.Println("  📝 Creating new RelatedCollection record...")

	newRelated := &RelatedCollection{
		Name: "Related collection example",
	}

	createdRelated, err := relatedService.Create(ctx, newRelated)
	if err != nil {
		return fmt.Errorf("failed to create related record: %w", err)
	}
	fmt.Printf("  ✅ Created related record ID: %s, Name: %s\n", createdRelated.ID, createdRelated.Name)

	// List records
	relatedList, err := relatedService.GetList(ctx, &pocketbase.ListOptions{
		PerPage: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list related records: %w", err)
	}

	fmt.Printf("  📋 Found %d related collection records\n", len(relatedList.Items))
	for i, item := range relatedList.Items {
		fmt.Printf("    %d. %s (ID: %s)\n", i+1, item.Name, item.ID)
	}

	return nil
}

// Type safety demonstration
func demonstrateTypeSafety() {
	fmt.Println("  🛡️ Compile-time type safety example")

	// Create type-safe struct
	record := &AllTypes{
		TextRequired:   "Type-safe text",
		NumberRequired: 123.45,
		BoolRequired:   true,
		EmailRequired:  "safe@example.com",
		URLRequired:    "https://safe.example.com",
	}

	// IDE auto-completion and type checking support
	fmt.Printf("  📝 Text: %s\n", record.TextRequired)
	fmt.Printf("  🔢 Number: %.2f\n", record.NumberRequired)
	fmt.Printf("  ✅ Boolean: %t\n", record.BoolRequired)

	// Safe handling of optional fields
	if record.TextOptional != nil {
		fmt.Printf("  📝 Optional text: %s\n", *record.TextOptional)
	} else {
		fmt.Println("  📝 Optional text: nil (safely handled)")
	}

	// Safe access to array fields
	fmt.Printf("  📋 Multi-select field count: %d\n", len(record.SelectMultiRequired))

	fmt.Println("  ✅ All field access is validated at compile time!")
}

// Error handling helper function
func handleError(operation string, err error) {
	if err != nil {
		log.Printf("❌ %s failed: %v", operation, err)
	}
}

// Usage example output function
func printUsageExample() {
	fmt.Println(`
🎯 Usage Examples:

1. Basic CRUD operations:
   service := NewAllTypesService(client)
   record := &AllTypes{TextRequired: "value", NumberRequired: 42}
   created, err := service.Create(ctx, record)

2. Type-safe field access:
   fmt.Println(record.TextRequired)  // Compile-time check
   if record.TextOptional != nil {   // Safe nil check
       fmt.Println(*record.TextOptional)
   }

3. List retrieval with filtering:
   list, err := service.GetList(ctx, &pocketbase.ListOptions{
       Filter: "text_required != ''",
       Sort: "-created",
   })

4. Get all records (automatic pagination):
   all, err := service.GetAll(ctx, nil)
`)
}
