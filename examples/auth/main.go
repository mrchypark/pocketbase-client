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

	// Example 1: Authenticate as an admin
	adminAuth, err := client.WithAdminPassword(context.Background(), "admin@example.com", "password123")
	if err != nil {
		log.Fatalf("Failed to authenticate admin: %v", err)
	}
	fmt.Printf("Authenticated as admin: %s (ID: %s)\n", adminAuth.Admin.Email, adminAuth.Admin.ID)

	// After authenticating, the client automatically uses the token for subsequent requests.
	// Let's verify by fetching the admin list, which requires authentication.
	admins, err := client.Admins.GetList(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to get admin list: %v", err)
	}
	fmt.Printf("Successfully fetched %d admin(s).\n\n", admins.TotalItems)

	// Clear the auth store to act as an unauthenticated client again
	client.ClearAuthStore()

	// Example 2: Authenticate as a regular user
	userAuth, err := client.WithPassword(context.Background(), "users", "testuser", "1234567890")
	if err != nil {
		log.Fatalf("Failed to authenticate user: %v", err)
	}
	fmt.Printf("Authenticated as user: %s (ID: %s)\n", userAuth.Record.ID, userAuth.Record.ID)

	// Verify authentication by fetching the user's own record
	userRecord, err := client.Records.GetOne(context.Background(), "users", userAuth.Record.ID, nil)
	if err != nil {
		log.Fatalf("Failed to get user record: %v", err)
	}
	fmt.Printf("Successfully fetched user record. Username: %s\n\n", userRecord.ID)

	// Example 3: Using a pre-existing auth token
	fmt.Println("Demonstrating authentication with a static token...")
	existingToken := userAuth.Token
	client.ClearAuthStore() // Clear previous auth

	// Set the token directly
	client.WithToken(existingToken)

	// Verify it works by fetching the user record again
	userRecord, err = client.Records.GetOne(context.Background(), "users", userAuth.Record.ID, nil)
	if err != nil {
		log.Fatalf("Failed to get user record with static token: %v", err)
	}
	fmt.Printf("Successfully fetched user record using a static token. Username: %s\n", userRecord.ID)
}
