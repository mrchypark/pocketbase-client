package main

import (
	"context"
	"fmt"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	ctx := context.Background()
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	// Check server health before attempting authentication
	if _, err := client.HealthCheck(ctx); err != nil {
		fmt.Printf("Error connecting to PocketBase server: %v\n", err)
		fmt.Println("Please ensure the PocketBase server is running at http://127.0.0.1:8090")
		return
	}

	if _, err := client.WithPassword(ctx, "users", "user@example.com", "1q2w3e4r"); err != nil {
		fmt.Printf("Authentication failed: %v\n", err)
		return
	}
	for i := 1; i <= 5; i++ {
		rec, err := client.Records.Create(ctx, "posts", map[string]any{"title": "hello"})
		if err != nil {
			panic(err)
		}
		fmt.Println("created", rec.ID)
	}

	l, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{
		SkipTotal: true,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("read", l.TotalItems)

}
