// Package main demonstrates various authentication methods with PocketBase Go client.
// This example shows admin authentication, user authentication, token usage, and profile updates.
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
	ctx := context.Background()

	// --- Example 1: Admin authentication ---
	fmt.Println("=== Admin Authentication ===")
	adminAuth, err := client.WithAdminPassword(ctx, "admin@example.com", "password123")
	if err != nil {
		log.Fatalf("Admin authentication failed: %v", err)
	}
	fmt.Printf("Authenticated as admin: %s (ID: %s)\n", adminAuth.Admin.Email, adminAuth.Admin.ID)

	// After authentication, the client automatically uses the token for subsequent requests.
	// Let's verify by fetching the admin list, which requires authentication.
	admins, err := client.Admins.GetList(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to get admin list: %v", err)
	}
	fmt.Printf("Successfully fetched %d admin(s).\n\n", admins.TotalItems)

	// Clear the auth store to act as an unauthenticated client again
	client.ClearAuthStore()

	// --- Example 2: User authentication ---
	fmt.Println("=== User Authentication ===")
	userAuth, err := client.WithPassword(ctx, "users", "testuser", "1234567890")
	if err != nil {
		log.Fatalf("User authentication failed: %v", err)
	}
	fmt.Printf("Authenticated as user: %s (ID: %s)\n", userAuth.Record.ID, userAuth.Record.ID)

	// Display authenticated user information using Record methods
	fmt.Printf("User information:\n")
	fmt.Printf("  ID: %s\n", userAuth.Record.ID)
	fmt.Printf("  Email: %s\n", userAuth.Record.GetString("email"))
	fmt.Printf("  Username: %s\n", userAuth.Record.GetString("username"))
	fmt.Printf("  Created: %s\n", userAuth.Record.GetDateTime("created").String())
	fmt.Printf("  Email verified: %t\n", userAuth.Record.GetBool("verified"))

	// Create Record service for users collection
	usersService := client.Records("users")

	// Verify authentication by fetching the user's own record
	userRecord, err := usersService.GetOne(ctx, userAuth.Record.ID, nil)
	if err != nil {
		log.Fatalf("Failed to get user record: %v", err)
	}
	fmt.Printf("Successfully fetched user record. Username: %s\n\n", userRecord.GetString("username"))

	// --- Example 3: Using existing auth token ---
	fmt.Println("=== Static Token Authentication ===")
	existingToken := userAuth.Token
	client.ClearAuthStore() // Clear previous auth

	// Set the token directly
	client.WithToken(existingToken)

	// Verify it works by fetching the user record again
	userRecord, err = usersService.GetOne(ctx, userAuth.Record.ID, nil)
	if err != nil {
		log.Fatalf("Failed to get user record with static token: %v", err)
	}
	fmt.Printf("Successfully fetched user record using static token.\n")
	fmt.Printf("User information:\n")
	fmt.Printf("  Username: %s\n", userRecord.GetString("username"))
	fmt.Printf("  Email: %s\n", userRecord.GetString("email"))
	fmt.Printf("  Last updated: %s\n\n", userRecord.GetDateTime("updated").String())

	// --- Example 4: User profile update ---
	fmt.Println("=== User Profile Update ===")

	// Update profile using Record object
	updateRecord := &pocketbase.Record{}
	updateRecord.Set("name", "Updated Name")

	// Add optional fields if they exist
	if userRecord.GetString("bio") != "" {
		updateRecord.Set("bio", "Updated bio using Record approach")
	}

	updatedUser, err := usersService.Update(ctx, userAuth.Record.ID, updateRecord)
	if err != nil {
		log.Printf("User profile update failed (optional): %v", err)
	} else {
		fmt.Printf("User profile updated:\n")
		fmt.Printf("  Name: %s\n", updatedUser.GetString("name"))
		if updatedUser.GetString("bio") != "" {
			fmt.Printf("  Bio: %s\n", updatedUser.GetString("bio"))
		}
	}

	// --- Example 5: List users (admin permission required) ---
	fmt.Println("\n=== List Users with Admin Permission ===")

	// Re-authenticate as admin
	_, err = client.WithAdminPassword(ctx, "admin@example.com", "password123")
	if err != nil {
		log.Fatalf("Admin re-authentication failed: %v", err)
	}

	// List users
	usersList, err := usersService.GetList(ctx, &pocketbase.ListOptions{
		PerPage: 5,
		Sort:    "-created",                           // Latest first
		Fields:  "id,username,email,created,verified", // Specific fields only
	})
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}

	fmt.Printf("Total %d users registered:\n", usersList.TotalItems)
	for i, user := range usersList.Items {
		fmt.Printf("  %d. %s (%s) - Joined: %s, Verified: %t\n",
			i+1,
			user.GetString("username"),
			user.GetString("email"),
			user.GetDateTime("created").String()[:10], // Date only
			user.GetBool("verified"))
	}

	fmt.Println("\nAuthentication example completed!")
}
