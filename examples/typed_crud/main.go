package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// Post represents a record from the 'posts' collection
type Post struct {
	pocketbase.Record
}

func (p *Post) GetTitle() string      { return p.GetString("title") }
func (p *Post) GetContent() string    { return p.GetString("content") }
func (p *Post) GetPublished() bool    { return p.GetBool("published") }
func (p *Post) GetViewCount() float64 { return p.GetFloat("view_count") }
func (p *Post) GetAuthorID() string   { return p.GetString("author") }
func (p *Post) GetTags() []string     { return p.GetStringSlice("tags") }

func (p *Post) SetTitle(v string)      { p.Set("title", v) }
func (p *Post) SetContent(v string)    { p.Set("content", v) }
func (p *Post) SetPublished(v bool)    { p.Set("published", v) }
func (p *Post) SetViewCount(v float64) { p.Set("view_count", v) }
func (p *Post) SetAuthorID(v string)   { p.Set("author", v) }
func (p *Post) SetTags(v []string)     { p.Set("tags", v) }

func (p *Post) ToMap() map[string]any {
	return map[string]any{
		"title":      p.GetTitle(),
		"content":    p.GetContent(),
		"published":  p.GetPublished(),
		"view_count": p.GetViewCount(),
		"author":     p.GetAuthorID(),
		"tags":       p.GetTags(),
	}
}

func main() {
	ctx := context.Background()

	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))
	if client == nil {
		client = pocketbase.NewClient("http://127.0.0.1:8090")
	}

	// Authenticate
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password123")
	if err != nil {
		log.Fatalf("Auth failed: %v", err)
	}

	// Create type-safe service
	postService := pocketbase.NewTypedRecordService[Post](client, "posts")

	// ============================================================
	// 1. Create - Make a new post
	// ============================================================
	fmt.Println("=== Create ===")
	newPost := &Post{}
	newPost.SetTitle("My First Typed Post")
	newPost.SetContent("This post was created using TypedRecordService!")
	newPost.SetPublished(true)
	newPost.SetViewCount(0)
	newPost.SetTags([]string{"go", "pocketbase", "tutorial"})

	created, err := postService.Create(ctx, newPost)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("Created: %s (ID: %s)\n", created.GetTitle(), created.ID)
	fmt.Printf("  ViewCount: %.0f\n", created.GetViewCount())
	fmt.Printf("  Tags: %v\n", created.GetTags())

	// ============================================================
	// 2. Read One - Get the post we just created
	// ============================================================
	fmt.Println("\n=== Read One ===")
	post, err := postService.GetOne(ctx, created.ID, nil)
	if err != nil {
		log.Fatalf("GetOne failed: %v", err)
	}
	fmt.Printf("Retrieved: %s\n", post.GetTitle())
	fmt.Printf("  Content: %s\n", post.GetContent())
	fmt.Printf("  Published: %v\n", post.GetPublished())

	// ============================================================
	// 3. Update - Increment view count
	// ============================================================
	fmt.Println("\n=== Update ===")
	post.SetViewCount(post.GetViewCount() + 1)
	updated, err := postService.Update(ctx, post.ID, post)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	fmt.Printf("Updated view count: %.0f\n", updated.GetViewCount())

	// ============================================================
	// 4. Read List - Get published posts
	// ============================================================
	fmt.Println("\n=== Read List ===")
	posts, err := postService.GetList(ctx, &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
		Filter:  "published = true",
		Sort:    "-created",
	})
	if err != nil {
		log.Fatalf("GetList failed: %v", err)
	}
	fmt.Printf("Found %d published posts:\n", len(posts.Items))
	for i, p := range posts.Items {
		fmt.Printf("  %d. %s (%.0f views)\n", i+1, p.GetTitle(), p.GetViewCount())
	}

	// ============================================================
	// 5. Read All - Get ALL posts (auto-pagination)
	// ============================================================
	fmt.Println("\n=== Read All (Auto-Pagination) ===")
	allPosts, err := postService.GetAll(ctx, &pocketbase.ListOptions{
		Filter: "published = true",
	})
	if err != nil {
		log.Fatalf("GetAll failed: %v", err)
	}
	fmt.Printf("Total published posts in database: %d\n", len(allPosts))

	// ============================================================
	// 6. Delete - Clean up
	// ============================================================
	fmt.Println("\n=== Delete ===")
	err = postService.Delete(ctx, postService.Collection, post.ID)
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("Post deleted successfully")

	fmt.Println("\n=== Done ===")
}
