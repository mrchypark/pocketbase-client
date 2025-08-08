// Package main demonstrates type-safe PocketBase client usage with generated models
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/mrchypark/pocketbase-client"
)

func main() {
	// Create PocketBase client
	client := pocketbase.NewClient("http://localhost:8090")
	ctx := context.Background()

	// Admin authentication
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "1q2w3e4r5t")
	if err != nil {
		log.Fatalf("Admin authentication failed: %v", err)
	}

	fmt.Println("🚀 PocketBase Type-Safe Generator Example")
	fmt.Println("=========================================")
	fmt.Println()

	// 1. Explain type safety concepts
	fmt.Println("📚 1. What is Type Safety?")
	explainTypeSafety()

	// 2. Explain generated model structure
	fmt.Println("\n🏗️ 2. Generated Model Structure")
	explainGeneratedModels()

	// 3. Demonstrate actual CRUD operations (using RelatedCollection)
	fmt.Println("\n💼 3. Live CRUD Operations Demo")
	if err := demonstrateCRUD(ctx, client); err != nil {
		log.Printf("CRUD demo error: %v", err)
	}

	// 4. Show actual benefits of type safety
	fmt.Println("\n✨ 4. Real Benefits of Type Safety")
	demonstrateTypeSafetyBenefits()

	// 5. Explain development workflow
	fmt.Println("\n🔄 5. Development Workflow")
	explainDevelopmentWorkflow()

	// 6. Show real-world usage patterns
	fmt.Println("\n💡 6. Real-World Usage Patterns")
	demonstrateRealWorldPatterns(ctx, client)

	fmt.Println("\n✅ Example Complete!")
	fmt.Println("\n📖 See README.md for more detailed information.")
}

// Explain type safety concepts
func explainTypeSafety() {
	fmt.Println("  Type safety means validating data types at compile time.")
	fmt.Println()
	fmt.Println("  🔴 Old approach (not type-safe):")
	fmt.Println(`    record := map[string]any{
        "text_field": "value",
        "number_field": "123", // 🚨 String assigned to number field
        "wrong_field": "value", // 🚨 Non-existent field
    }`)
	fmt.Println()
	fmt.Println("  🟢 New approach (type-safe):")
	fmt.Println(`    record := &AllTypes{
        TextRequired: "value",
        NumberRequired: 123,   // ✅ Correct type
        // WrongField: "value", // ✅ Compile error prevents this
    }`)
}

// Explain generated model structure
func explainGeneratedModels() {
	fmt.Println("  The pbc-gen tool generates the following from PocketBase schema:")
	fmt.Println()
	fmt.Println("  📋 Generated elements:")
	fmt.Println("    • Go structs (AllTypes, RelatedCollection)")
	fmt.Println("    • Type-safe service constructors (NewAllTypesService)")
	fmt.Println("    • JSON tags and field validation")
	fmt.Println("    • Pointer types for optional fields")
	fmt.Println()
	fmt.Println("  🔍 Example - AllTypes struct:")
	fmt.Println(`    type AllTypes struct {
        pocketbase.BaseModel
        TextRequired    string   // Required field
        TextOptional    *string  // Optional field (pointer)
        NumberRequired  float64  // Number type
        BoolRequired    bool     // Boolean type
        // ... all other PocketBase field types
    }`)
}

// Demonstrate actual CRUD operations
func demonstrateCRUD(ctx context.Context, client *pocketbase.Client) error {
	fmt.Println("  Basic CRUD operations using RelatedCollection:")
	fmt.Println()

	// Create type-safe service
	service := NewRelatedCollectionService(client)
	fmt.Println("  ✅ Type-safe service created:")
	fmt.Println("    service := NewRelatedCollectionService(client)")
	fmt.Println()

	// 1. Create
	fmt.Println("  📝 1. Record Creation:")
	newRecord := &RelatedCollection{
		Name: "Type-safe example record",
	}

	created, err := service.Create(ctx, newRecord)
	if err != nil {
		return fmt.Errorf("record creation failed: %w", err)
	}
	fmt.Printf("    ✅ Creation complete - ID: %s, Name: %s\n", created.ID, created.Name)
	fmt.Println()

	// 2. Read
	fmt.Println("  🔍 2. Record Retrieval:")
	retrieved, err := service.GetOne(ctx, created.ID, nil)
	if err != nil {
		return fmt.Errorf("record retrieval failed: %w", err)
	}
	fmt.Printf("    ✅ Retrieval complete - Name: %s\n", retrieved.Name)
	fmt.Println()

	// 3. Update
	fmt.Println("  ✏️ 3. Record Update:")
	retrieved.Name = "Updated name"
	updated, err := service.Update(ctx, retrieved.ID, retrieved)
	if err != nil {
		return fmt.Errorf("record update failed: %w", err)
	}
	fmt.Printf("    ✅ Update complete - New name: %s\n", updated.Name)
	fmt.Println()

	// 4. List
	fmt.Println("  📋 4. Record List Retrieval:")
	list, err := service.GetList(ctx, &pocketbase.ListOptions{
		PerPage: 5,
		Sort:    "-id", // Use id instead of created since related_collection doesn't have created field
	})
	if err != nil {
		return fmt.Errorf("list retrieval failed: %w", err)
	}
	fmt.Printf("    ✅ Retrieved %d out of %d total records\n", len(list.Items), list.TotalItems)
	for i, item := range list.Items {
		fmt.Printf("      %d. %s (ID: %s)\n", i+1, item.Name, item.ID)
	}

	return nil
}

// Demonstrate actual benefits of type safety
func demonstrateTypeSafetyBenefits() {
	fmt.Println("  Benefits verified at compile time:")
	fmt.Println()

	// 1. Field name validation
	fmt.Println("  ✅ 1. Automatic field name validation:")
	fmt.Println("    record.Name = \"value\"     // ✅ Correct field name")
	fmt.Println("    // record.Nmae = \"value\"  // ❌ Compile error: typo detected")
	fmt.Println()

	// 2. Type validation
	fmt.Println("  ✅ 2. Automatic data type validation:")
	fmt.Println("    record.NumberRequired = 42.5    // ✅ float64 type")
	fmt.Println("    // record.NumberRequired = \"42\" // ❌ Compile error: type mismatch")
	fmt.Println()

	// 3. IDE support
	fmt.Println("  ✅ 3. IDE Support:")
	fmt.Println("    • Auto-completion: Shows available fields when typing record.")
	fmt.Println("    • Type hints: Displays type information for each field")
	fmt.Println("    • Refactoring: Automatically updates all usages when field names change")
	fmt.Println("    • Error highlighting: Immediately shows incorrect usage with red underlines")
	fmt.Println()

	// 4. Runtime safety
	fmt.Println("  ✅ 4. Runtime Safety:")
	record := &RelatedCollection{Name: "example"}
	fmt.Printf("    record.Name type: %T\n", record.Name)
	fmt.Println("    • Prevents nil pointer dereference")
	fmt.Println("    • Prevents type casting errors")
	fmt.Println("    • Prevents access to non-existent fields")
}

// Explain development workflow
func explainDevelopmentWorkflow() {
	fmt.Println("  Typical workflow for type-safe development:")
	fmt.Println()
	fmt.Println("  1️⃣ Design collection schema in PocketBase")
	fmt.Println("    • Add/modify fields in admin UI")
	fmt.Println("    • Set field types and constraints")
	fmt.Println()
	fmt.Println("  2️⃣ Export schema")
	fmt.Println("    curl http://localhost:8090/api/collections > pb_schema.json")
	fmt.Println()
	fmt.Println("  3️⃣ Generate Go models")
	fmt.Println("    go run ./cmd/pbc-gen -schema ./pb_schema.json -path ./models.gen.go")
	fmt.Println()
	fmt.Println("  4️⃣ Write type-safe code")
	fmt.Println("    service := NewAllTypesService(client)")
	fmt.Println("    record := &AllTypes{...}")
	fmt.Println()
	fmt.Println("  5️⃣ Compile and test")
	fmt.Println("    go build .  // Automatically detects type errors")
	fmt.Println("    go test .   // Type-safe testing")
}

// Demonstrate real-world usage patterns
func demonstrateRealWorldPatterns(ctx context.Context, client *pocketbase.Client) error {
	service := NewRelatedCollectionService(client)

	fmt.Println("  Common patterns used in real applications:")
	fmt.Println()

	// 1. Conditional field setting
	fmt.Println("  🔀 1. Conditional Field Setting:")
	shouldCreateRecord := true
	if shouldCreateRecord {
		record := &RelatedCollection{
			Name: "Conditionally created record",
		}
		created, err := service.Create(ctx, record)
		if err == nil {
			fmt.Printf("    ✅ Conditional record created: %s\n", created.ID)
		}
	}
	fmt.Println()

	// 2. Error handling pattern
	fmt.Println("  🛡️ 2. Safe Error Handling:")
	_, err := service.GetOne(ctx, "nonexistent_id", nil)
	if err != nil {
		fmt.Println("    ✅ Non-existent record access safely handled with error")
	}
	fmt.Println()

	// 3. Batch processing
	fmt.Println("  📦 3. Batch Processing Pattern:")
	names := []string{"Batch1", "Batch2", "Batch3"}
	successCount := 0
	for _, name := range names {
		record := &RelatedCollection{Name: name}
		if _, err := service.Create(ctx, record); err == nil {
			successCount++
		}
	}
	fmt.Printf("    ✅ Successfully created %d out of %d records in batch\n", successCount, len(names))
	fmt.Println()

	// 4. Filtering and sorting
	fmt.Println("  🔍 4. Advanced Query Pattern:")
	list, err := service.GetList(ctx, &pocketbase.ListOptions{
		Filter:  "name ~ 'Batch%'", // Records with names starting with 'Batch'
		Sort:    "-id",             // Descending by ID (since created field doesn't exist)
		PerPage: 10,
	})
	if err == nil {
		fmt.Printf("    ✅ Filtered results: %d records\n", len(list.Items))
		for _, item := range list.Items {
			fmt.Printf("      - %s\n", item.Name)
		}
	} else {
		fmt.Println("    ⚠️ Advanced query example (may be limited in some PocketBase versions)")
		fmt.Println("    💡 Basic queries work normally")
	}

	return nil
}
