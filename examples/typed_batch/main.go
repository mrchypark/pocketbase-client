package main

import (
	"context"
	"fmt"
	"log"
	"os"

	pocketbase "github.com/mrchypark/pocketbase-client"
)

type Post struct {
	pocketbase.Record
}

func (p *Post) GetTitle() string   { return p.GetString("title") }
func (p *Post) GetContent() string { return p.GetString("content") }
func (p *Post) GetPublished() bool { return p.GetBool("published") }

func (p *Post) SetTitle(v string)   { p.Set("title", v) }
func (p *Post) SetContent(v string) { p.Set("content", v) }
func (p *Post) SetPublished(v bool) { p.Set("published", v) }

func (p *Post) ToMap() map[string]any {
	return map[string]any{
		"title":     p.GetTitle(),
		"content":   p.GetContent(),
		"published": p.GetPublished(),
	}
}

type User struct {
	pocketbase.Record
}

func (u *User) GetName() string  { return u.GetString("name") }
func (u *User) GetEmail() string { return u.GetString("email") }
func (u *User) GetAge() float64  { return u.GetFloat("age") }

func (u *User) SetName(v string)  { u.Set("name", v) }
func (u *User) SetEmail(v string) { u.Set("email", v) }
func (u *User) SetAge(v float64)  { u.Set("age", v) }

func (u *User) ToMap() map[string]any {
	return map[string]any{
		"name":  u.GetName(),
		"email": u.GetEmail(),
		"age":   u.GetAge(),
	}
}

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

	postService := pocketbase.NewTypedRecordService[Post](client, "posts")
	userService := pocketbase.NewTypedRecordService[User](client, "users")

	// Create test data
	post1 := &Post{}
	post1.SetTitle("Post 1")
	post1.SetContent("Content 1")
	post1.SetPublished(true)

	post2 := &Post{}
	post2.SetTitle("Post 2")
	post2.SetContent("Content 2")
	post2.SetPublished(false)

	user1 := &User{}
	user1.SetName("Alice")
	user1.SetEmail("alice@example.com")
	user1.SetAge(25)

	user2 := &User{}
	user2.SetName("Bob")
	user2.SetEmail("bob@example.com")
	user2.SetAge(30)

	p1, _ := postService.Create(ctx, post1)
	p2, _ := postService.Create(ctx, post2)
	u1, _ := userService.Create(ctx, user1)
	u2, _ := userService.Create(ctx, user2)

	fmt.Println("Created test data:")
	fmt.Printf("  Posts: %s, %s\n", p1.ID, p2.ID)
	fmt.Printf("  Users: %s, %s\n", u1.ID, u2.ID)

	// Typed Batch Operations
	fmt.Println("\n=== Typed Batch Operations ===")

	createPostReq, _ := postService.NewCreateRequest("posts", map[string]any{
		"title":     "Batch Post",
		"content":   "Created via batch",
		"published": true,
	})
	createUserReq, _ := userService.NewCreateRequest("users", map[string]any{
		"name":  "Charlie",
		"email": "charlie@example.com",
		"age":   35,
	})
	updatePostReq, _ := postService.NewUpdateRequest("posts", p1.ID, map[string]any{
		"title": "Updated via Batch",
	})
	updateUserReq, _ := userService.NewUpdateRequest("users", u1.ID, map[string]any{
		"age": 26,
	})

	results, err := client.Batch.Execute(ctx, []*pocketbase.BatchRequest{
		createPostReq,
		createUserReq,
		updatePostReq,
		updateUserReq,
	})
	if err != nil {
		log.Fatalf("Batch Execute failed: %v", err)
	}

	fmt.Printf("Batch executed %d operations:\n", len(results))
	for i, res := range results {
		if res.ParsedError != nil {
			fmt.Printf("  %d: ERROR - %s\n", i+1, res.ParsedError.Message)
		} else {
			fmt.Printf("  %d: Status %d - OK\n", i+1, res.Status)
		}
	}

	// Bulk Create
	fmt.Println("\n=== Bulk Create ===")

	var batchReqs []*pocketbase.BatchRequest
	for i := 1; i <= 3; i++ {
		post := &Post{}
		post.SetTitle(fmt.Sprintf("Bulk Post %d", i))
		post.SetContent("Content")
		post.SetPublished(true)
		req, _ := postService.NewCreateRequest("posts", post.ToMap())
		batchReqs = append(batchReqs, req)
	}

	bulkResults, err := client.Batch.Execute(ctx, batchReqs)
	if err != nil {
		log.Fatalf("Bulk create failed: %v", err)
	}
	fmt.Printf("Bulk created %d posts\n", len(bulkResults))

	// Upsert
	fmt.Println("\n=== Upsert ===")

	upsertReq, _ := postService.NewUpsertRequest("posts", map[string]any{
		"id":        "upsert-post-id",
		"title":     "Upserted Post",
		"content":   "This was created or updated",
		"published": true,
	})

	upsertResults, err := client.Batch.Execute(ctx, []*pocketbase.BatchRequest{upsertReq})
	if err != nil {
		log.Fatalf("Upsert failed: %v", err)
	}
	fmt.Printf("Upsert result: Status %d\n", upsertResults[0].Status)

	// Cleanup
	fmt.Println("\n=== Cleanup ===")
	postService.Delete(ctx, postService.Collection, p1.ID)
	postService.Delete(ctx, postService.Collection, p2.ID)
	userService.Delete(ctx, userService.Collection, u1.ID)
	userService.Delete(ctx, userService.Collection, u2.ID)
	fmt.Println("Cleanup done")
}
