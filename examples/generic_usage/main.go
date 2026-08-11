package main

import (
	"context"
	"fmt"
	"log"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// ============================================================
// Generated Model (from pbc-gen)
// ============================================================
//
// When you run pbc-gen, it generates models with direct struct fields:
//
//   type Post struct {
//       ID             string             `json:"id"`
//       CollectionID   string             `json:"collectionId"`
//       CollectionName string             `json:"collectionName"`
//       Created        pocketbase.DateTime `json:"created"`
//       Updated        pocketbase.DateTime `json:"updated"`
//       Title          string             `json:"title"`
//       Content        *string            `json:"content,omitempty"`
//       Published      bool               `json:"published"`
//       ViewCount      float64            `json:"view_count"`
//       Author         string             `json:"author"`
//       Tags           []string           `json:"tags"`
//   }
// ============================================================

// Post represents a record from the 'posts' collection
type Post struct {
	ID             string             `json:"id"`
	CollectionID   string             `json:"collectionId"`
	CollectionName string             `json:"collectionName"`
	Created        pocketbase.DateTime `json:"created"`
	Updated        pocketbase.DateTime `json:"updated"`
	Title          string             `json:"title"`
	Content        *string            `json:"content,omitempty"`
	Published      bool               `json:"published"`
	ViewCount      float64            `json:"view_count"`
	Author         string             `json:"author"`
	Tags           []string           `json:"tags"`
}

func (p *Post) GetID() string                { return p.ID }
func (p *Post) GetCollectionName() string     { return p.CollectionName }
func (p *Post) SetID(id string)               { p.ID = id }
func (p *Post) SetCollectionID(id string)     { p.CollectionID = id }
func (p *Post) SetCollectionName(name string) { p.CollectionName = name }

func (p *Post) ToMap() map[string]any {
	data := make(map[string]any)
	if p.Title != "" {
		data["title"] = p.Title
	}
	if p.Content != nil {
		data["content"] = p.Content
	}
	if p.Published {
		data["published"] = p.Published
	}
	if p.ViewCount != 0 {
		data["view_count"] = p.ViewCount
	}
	if p.Author != "" {
		data["author"] = p.Author
	}
	if p.Tags != nil {
		data["tags"] = p.Tags
	}
	return data
}

// Author represents a record from the 'authors' collection
type Author struct {
	ID             string             `json:"id"`
	CollectionID   string             `json:"collectionId"`
	CollectionName string             `json:"collectionName"`
	Created        pocketbase.DateTime `json:"created"`
	Updated        pocketbase.DateTime `json:"updated"`
	Name           string             `json:"name"`
	Email          string             `json:"email"`
	Bio            string             `json:"bio"`
	Avatar         string             `json:"avatar"`
}

func (a *Author) GetID() string                { return a.ID }
func (a *Author) GetCollectionName() string     { return a.CollectionName }
func (a *Author) SetID(id string)               { a.ID = id }
func (a *Author) SetCollectionID(id string)     { a.CollectionID = id }
func (a *Author) SetCollectionName(name string) { a.CollectionName = name }

func (a *Author) ToMap() map[string]any {
	data := make(map[string]any)
	if a.Name != "" {
		data["name"] = a.Name
	}
	if a.Email != "" {
		data["email"] = a.Email
	}
	if a.Bio != "" {
		data["bio"] = a.Bio
	}
	if a.Avatar != "" {
		data["avatar"] = a.Avatar
	}
	return data
}

// ============================================================

func main() {
	ctx := context.Background()
	client := pocketbase.NewClient("http://127.0.0.1:8090")

	// Authenticate as admin
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password")
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	// ============================================================
	// Method 1: Typed Service (Recommended)
	// ============================================================
	fmt.Println("=== Method 1: Typed Service ===")

	// Create a typed service for the 'posts' collection
	postService := pocketbase.NewTypedRecordService[Post](client, "posts")

	// Create a new post
	newPost := &Post{
		Title:     "Hello World",
		Published: true,
		ViewCount: 0,
		Tags:      []string{"golang", "pocketbase"},
	}

	created, err := postService.Create(ctx, newPost)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("Created: %s (ID: %s)\n", created.Title, created.ID)

	// Get a single post
	post, err := postService.GetOne(ctx, created.ID, nil)
	if err != nil {
		log.Fatalf("GetOne failed: %v", err)
	}
	fmt.Printf("Post: %s (Views: %.0f)\n", post.Title, post.ViewCount)

	// Get a list of posts
	result, err := postService.GetList(ctx, &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
		Filter:  "published = true",
	})
	if err != nil {
		log.Fatalf("GetList failed: %v", err)
	}
	fmt.Printf("Found %d posts:\n", result.TotalItems)
	for _, p := range result.Items {
		fmt.Printf("  - %s\n", p.Title)
	}

	// Update a post
	post.ViewCount++
	post.Title = "Updated Title"
	updated, err := postService.Update(ctx, post.ID, post)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	fmt.Printf("Updated: %s (Views: %.0f)\n", updated.Title, updated.ViewCount)

	// Delete record
	err = postService.Delete(ctx, created.ID)
	if err != nil {
		log.Fatalf("Delete failed: %v", err)
	}
	fmt.Println("Deleted post")

	// ============================================================
	// Method 2: Type-safe Relation Loading
	// ============================================================

	loadedPost, loadedAuthor, err := loadPostWithAuthor(ctx, client, "POST_ID")
	if err != nil {
		log.Fatalf("loadPostWithAuthor failed: %v", err)
	}
	fmt.Printf("Post by %s: %s\n", loadedAuthor.Name, loadedPost.Title)

	// ============================================================
	// Method 3: Dynamic RecordService (for untyped access)
	// ============================================================

	fmt.Println("\n=== Method 3: Dynamic RecordService ===")

	// Use the dynamic RecordService for flexible queries
	dynamicPost, err := client.Records.GetOne(ctx, "posts", "RECORD_ID", nil)
	if err != nil {
		log.Fatalf("Failed to get record: %v", err)
	}
	fmt.Printf("Dynamic record: ID=%s\n", dynamicPost.ID)
}

// loadPostWithAuthor demonstrates type-safe relation loading
func loadPostWithAuthor(ctx context.Context, client *pocketbase.Client, postID string) (*Post, *Author, error) {
	postService := pocketbase.NewTypedRecordService[Post](client, "posts")

	post, err := postService.GetOne(ctx, postID, nil)
	if err != nil {
		return nil, nil, err
	}

	authorService := pocketbase.NewTypedRecordService[Author](client, "authors")
	author, err := authorService.GetOne(ctx, post.Author, nil)
	if err != nil {
		return nil, nil, err
	}

	return post, author, nil
}
