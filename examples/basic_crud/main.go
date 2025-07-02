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

	if _, err := client.AuthenticateAsAdmin(ctx, "admin@example.com", "1q2w3e4r5t"); err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		return
	}

	rec, err := client.Records.Create(ctx, "posts", map[string]any{"title": "hello"})
	if err != nil {
		panic(err)
	}
	fmt.Println("created", rec.ID)

	read, err := client.Records.GetOne(ctx, "posts", rec.ID, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("read", read.ID)

	l, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("read", l.TotalItems)

	rec, err = client.Records.Update(ctx, "posts", rec.ID, map[string]any{"title": "updated"})
	if err != nil {
		panic(err)
	}
	fmt.Println("updated", rec.ID)

	if err := client.Records.Delete(ctx, "posts", rec.ID); err != nil {
		panic(err)
	}
	fmt.Println("deleted", rec.ID)
}
