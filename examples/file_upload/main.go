package main

import (
	"context"
	"fmt"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

func main() {
	ctx := context.Background()
	client := pocketbase.NewClient("http://127.0.0.1:8090")
	if _, err := client.AuthenticateAsAdmin(ctx, "admin@example.com", "1q2w3e4r5t"); err != nil {
		panic(err)
	}

	post, err := client.Records.Create(ctx, "posts", map[string]any{"title": "with file"})
	if err != nil {
		panic(err)
	}
	fmt.Println("created", post.ID)

	file, err := os.Open("image.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	rec, err := client.Files.Upload(ctx, "posts", post.ID, "image", "image.png", file)
	if err != nil {
		panic(err)
	}
	fmt.Println("uploaded", rec.ID)
}
