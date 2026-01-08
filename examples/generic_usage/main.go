package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

// ============================================================
// Generated Model (from pbc-gen)
// ============================================================
//
// When you run pbc-gen, it generates models like this:
//
//   type Post struct {
//       pocketbase.Record
//   }
//
// Plus getter/setter methods for each field:
//   - Title() string
//   - SetTitle(value string)
//   - TitleValueOr(default string) string (for optional fields)
// ============================================================

// Post represents a record from the 'posts' collection
type Post struct {
	pocketbase.Record
}

// Generated getter methods (normally from pbc-gen):
func (p *Post) Title() string      { return p.GetString("title") }
func (p *Post) Content() string    { return p.GetString("content") }
func (p *Post) Published() bool    { return p.GetBool("published") }
func (p *Post) ViewCount() float64 { return p.GetFloat("view_count") }
func (p *Post) AuthorID() string   { return p.GetString("author") }
func (p *Post) Tags() []string     { return p.GetStringSlice("tags") }

// Generated setter methods:
func (p *Post) SetTitle(value string)      { p.Set("title", value) }
func (p *Post) SetContent(value string)    { p.Set("content", value) }
func (p *Post) SetPublished(value bool)    { p.Set("published", value) }
func (p *Post) SetViewCount(value float64) { p.Set("view_count", value) }
func (p *Post) SetAuthorID(value string)   { p.Set("author", value) }
func (p *Post) SetTags(value []string)     { p.Set("tags", value) }

// ToMap implements Mappable for Create/Update
func (p *Post) ToMap() map[string]any {
	return map[string]any{
		"title":      p.Title(),
		"content":    p.Content(),
		"published":  p.Published(),
		"view_count": p.ViewCount(),
		"author":     p.AuthorID(),
		"tags":       p.Tags(),
	}
}

// Author represents a record from the 'authors' collection
type Author struct {
	pocketbase.Record
}

func (a *Author) Name() string   { return a.GetString("name") }
func (a *Author) Email() string  { return a.GetString("email") }
func (a *Author) Bio() string    { return a.GetString("bio") }
func (a *Author) Avatar() string { return a.GetString("avatar") }

func (a *Author) ToMap() map[string]any {
	return map[string]any{
		"name":   a.Name(),
		"email":  a.Email(),
		"bio":    a.Bio(),
		"avatar": a.Avatar(),
	}
}

// ============================================================
// Main Example: Using TypedRecordService with Generated Models
// ============================================================

func main() {
	ctx := context.Background()

	client := pocketbase.NewClient(os.Getenv("POCKETBASE_URL"))
	if client == nil {
		client = pocketbase.NewClient("http://127.0.0.1:8090")
	}

	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password123")
	if err != nil {
		log.Fatalf("Auth failed: %v", err)
	}

	// ============================================================
	// Method 1: Using NewTypedRecordService (NEW!)
	// ============================================================
	// Pass the MODEL TYPE (not pointer), get typed service
	// T = Post (the struct), methods return *Post

	postService := pocketbase.NewTypedRecordService[Post](client, "posts")

	// Get single record - returns *Post
	post, err := postService.GetOne(ctx, "RECORD_ID", nil)
	if err != nil {
		log.Fatalf("GetOne failed: %v", err)
	}
	// Access fields via generated getter methods
	fmt.Printf("Post: %s (Views: %.0f)\n", post.Title(), post.ViewCount())

	// Get list with pagination
	posts, err := postService.GetList(ctx, &pocketbase.ListOptions{
		Page:    1,
		PerPage: 10,
		Filter:  "published = true",
		Sort:    "-created",
	})
	if err != nil {
		log.Fatalf("GetList failed: %v", err)
	}
	fmt.Printf("Found %d posts on page 1\n", len(posts.Items))
	for _, p := range posts.Items {
		fmt.Printf("  - %s\n", p.Title())
	}

	// Get ALL records (auto-pagination)
	allPosts, err := postService.GetAll(ctx, &pocketbase.ListOptions{
		Filter: "published = true",
	})
	if err != nil {
		log.Fatalf("GetAll failed: %v", err)
	}
	fmt.Printf("Total published posts: %d\n", len(allPosts))

	// Create new record
	newPost := &Post{}
	newPost.SetTitle("My New Post")
	newPost.SetContent("This is the content...")
	newPost.SetPublished(true)
	newPost.SetViewCount(0)
	newPost.SetTags([]string{"go", "pocketbase"})

	created, err := postService.Create(ctx, newPost)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Printf("Created: %s (ID: %s)\n", created.Title(), created.ID)

	// Update record
	created.SetViewCount(100)
	updated, err := postService.Update(ctx, created.ID, created)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	fmt.Printf("Updated: %s (Views: %.0f)\n", updated.Title(), updated.ViewCount())

	// Delete record
	err = postService.Delete(ctx, postService.Collection, created.ID)
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
	fmt.Printf("Post by %s: %s\n", loadedAuthor.Name(), loadedPost.Title())
}

// ============================================================
// Type-safe Relation Loading Example
// ============================================================

func loadPostWithAuthor(ctx context.Context, client *pocketbase.Client, postID string) (*Post, *Author, error) {
	postService := pocketbase.NewTypedRecordService[Post](client, "posts")
	authorService := pocketbase.NewTypedRecordService[Author](client, "authors")

	post, err := postService.GetOne(ctx, postID, nil)
	if err != nil {
		return nil, nil, err
	}

	author, err := authorService.GetOne(ctx, post.AuthorID(), nil)
	if err != nil {
		return nil, nil, err
	}

	return post, author, nil
}
