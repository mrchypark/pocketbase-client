package main

import (
	"fmt"
	"log"

	"github.com/cypark/pocketbase-client/tools/list"
	"github.com/cypark/pocketbase-client/database" // Import the generated package
)

func main() {
	// Initialize the generated database client
	db, err := database.New("http://127.0.0.1:8090")
	if err != nil {
		log.Fatal(err)
	}

	// Authenticate as admin
	if _, err := db.AuthWithAdminPassword("admin@example.com", "your_admin_password"); err != nil {
		log.Fatal(err)
	}

	// Create a new post using the generated type
	newPost, err := db.Posts.Create(database.PostsUpsert{
		Title:   "Type-Safe Title",
		Content: "This is amazing!",
	})
	if err != nil {
		log.Fatalf("Failed to create type-safe record: %v", err)
	}
	fmt.Printf("Created type-safe record with ID: %s and title: %s\n", newPost.ID, newPost.Title)

    // List posts
    posts, err := db.Posts.List(list.NewOptions())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Found %d posts.\n", len(posts.Items))
}