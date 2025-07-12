package main

import (
	"context"
	"fmt"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// main demonstrates basic CRUD operations using the PocketBase client.
func main() {
	ctx := context.Background()
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	// Check server health before attempting authentication
	if _, err := client.HealthCheck(ctx); err != nil {
		fmt.Printf("Error connecting to PocketBase server: %v\n", err)
		fmt.Println("Please ensure the PocketBase server is running at http://127.0.0.1:8090")
		return
	}

	if _, err := client.WithAdminPassword(ctx, "admin@example.com", "1q2w3e4r5t"); err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		return
	}

	t, _ := client.AuthStore.Token()
	fmt.Println("token", t)

	client.ClearAuthStore()

	ct, _ := client.AuthStore.Token()
	fmt.Println("token", ct)

	client.WithToken(t)

	wt, _ := client.AuthStore.Token()
	fmt.Println("token", wt)

}
