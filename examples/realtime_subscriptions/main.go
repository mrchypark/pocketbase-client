// Package main demonstrates real-time subscriptions using PocketBase Go client.
// This example shows how to subscribe to collection events and handle Record-based updates.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	ctx := context.Background()

	// Create Record service for posts collection
	postsService := client.Records("posts")

	fmt.Println("=== PocketBase Real-time Subscriptions Example ===")

	// --- Example 1: Single collection subscription ---
	fmt.Println("Subscribing to all events on the 'posts' collection...")

	unsubscribePosts, err := client.Realtime.Subscribe(ctx, []string{"posts"}, func(e *pocketbase.RealtimeEvent, err error) {
		if err != nil {
			log.Printf("Subscription error: %v", err)
			return
		}

		fmt.Printf("\n📢 Event received: %s\n", e.Action)
		fmt.Printf("   Record ID: %s\n", e.Record.ID)
		fmt.Printf("   Collection: %s\n", e.Record.CollectionName)

		// Access Record data
		switch e.Action {
		case "create":
			fmt.Printf("   ✅ New post created\n")
			fmt.Printf("   Title: %s\n", e.Record.GetString("title"))
			fmt.Printf("   Content: %s\n", e.Record.GetString("content"))
			fmt.Printf("   Created: %s\n", e.Record.GetDateTime("created").String())

		case "update":
			fmt.Printf("   🔄 Post updated\n")
			fmt.Printf("   Title: %s\n", e.Record.GetString("title"))
			fmt.Printf("   Updated: %s\n", e.Record.GetDateTime("updated").String())

		case "delete":
			fmt.Printf("   🗑️ Post deleted\n")
			fmt.Printf("   Deleted ID: %s\n", e.Record.ID)
		}
		fmt.Println("---")
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to posts collection: %v", err)
	}
	defer unsubscribePosts()

	// --- Example 2: Multiple collection subscription ---
	fmt.Println("Also subscribing to 'users' and 'comments' collections...")

	unsubscribeMultiple, err := client.Realtime.Subscribe(ctx, []string{"users", "comments"}, func(e *pocketbase.RealtimeEvent, err error) {
		if err != nil {
			log.Printf("Multiple subscription error: %v", err)
			return
		}

		fmt.Printf("\n🔔 [%s] %s event\n", e.Record.CollectionName, e.Action)
		fmt.Printf("   Record ID: %s\n", e.Record.ID)

		// Handle by collection
		switch e.Record.CollectionName {
		case "users":
			if e.Action == "create" {
				fmt.Printf("   👤 New user: %s\n", e.Record.GetString("username"))
			}
		case "comments":
			if e.Action == "create" {
				fmt.Printf("   💬 New comment: %s\n", e.Record.GetString("content"))
			}
		}
		fmt.Println("---")
	})
	if err != nil {
		log.Printf("Multiple collection subscription failed (expected): %v", err)
	} else {
		defer unsubscribeMultiple()
	}

	// --- Create test data ---
	fmt.Println("\nCreating some test records...")

	// Create test records after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)

		fmt.Println("\n🧪 Creating test records...")

		// Create new post
		newPost := &pocketbase.Record{}
		newPost.Set("title", "Real-time Test Post")
		newPost.Set("content", "This is a post to test real-time subscriptions.")
		newPost.Set("is_published", true)

		created, err := postsService.Create(context.Background(), newPost)
		if err != nil {
			log.Printf("Failed to create test post: %v", err)
			return
		}

		fmt.Printf("Test post created (ID: %s)\n", created.ID)

		// Update after 5 seconds
		time.Sleep(5 * time.Second)

		updatePost := &pocketbase.Record{}
		updatePost.Set("title", "Updated Real-time Test Post")
		updatePost.Set("content", "This post has been updated in real-time!")

		_, err = postsService.Update(context.Background(), created.ID, updatePost)
		if err != nil {
			log.Printf("Failed to update test post: %v", err)
			return
		}

		fmt.Printf("Test post updated (ID: %s)\n", created.ID)

		// Delete after 5 seconds
		time.Sleep(5 * time.Second)

		err = postsService.Delete(context.Background(), created.ID)
		if err != nil {
			log.Printf("Failed to delete test post: %v", err)
			return
		}

		fmt.Printf("Test post deleted (ID: %s)\n", created.ID)
	}()

	// --- Graceful shutdown handling ---
	fmt.Println("\nWaiting for real-time events... (Ctrl+C to exit)")
	fmt.Println("Try modifying the posts collection through PocketBase Admin UI in another terminal!")

	// Create channel for signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or auto-exit after 30 minutes
	select {
	case <-sigChan:
		fmt.Println("\n\nReceived exit signal. Unsubscribing and exiting...")
	case <-time.After(30 * time.Minute):
		fmt.Println("\n\n30 minutes elapsed. Auto-exiting...")
	}

	fmt.Println("Real-time subscriptions example completed!")
}
