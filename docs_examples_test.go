package pocketbase

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// docPost is a test model matching the generated model structure.
type docPost struct {
	ID             string   `json:"id"`
	CollectionID   string   `json:"collectionId"`
	CollectionName string   `json:"collectionName"`
	Created        DateTime `json:"created"`
	Updated        DateTime `json:"updated"`
	Title          string   `json:"title"`
	Content        *string  `json:"content,omitempty"`
	Published      bool     `json:"published"`
}

func (p *docPost) GetID() string                { return p.ID }
func (p *docPost) GetCollectionName() string     { return p.CollectionName }
func (p *docPost) SetID(id string)               { p.ID = id }
func (p *docPost) SetCollectionID(id string)     { p.CollectionID = id }
func (p *docPost) SetCollectionName(name string) { p.CollectionName = name }
func (p *docPost) SetTitle(v string)             { p.Title = v }
func (p *docPost) SetContent(v string)           { p.Content = &v }
func (p *docPost) SetPublished(v bool)           { p.Published = v }

func (p *docPost) ToMap() map[string]any {
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
	return data
}

// mockDocServer creates a test HTTP server that handles PocketBase API endpoints.
func mockDocServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		// POST /api/collections/{collection}/records
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/collections/"):
			body, _ := io.ReadAll(r.Body)
			var record map[string]any
			_ = json.Unmarshal(body, &record)
			if record == nil {
				record = make(map[string]any)
			}
			record["id"] = "rec1"
			record["collectionId"] = "col1"
			collectionName := strings.TrimPrefix(r.URL.Path, "/api/collections/")
			collectionName = strings.TrimSuffix(collectionName, "/records")
			record["collectionName"] = collectionName
			record["created"] = "2024-01-01 00:00:00.000Z"
			record["updated"] = "2024-01-01 00:00:00.000Z"
			_ = json.NewEncoder(w).Encode(record)

		// GET /api/collections/{collection}/records/{id}
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/collections/") && strings.Count(r.URL.Path, "/") > 5:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":             "rec1",
				"collectionId":   "col1",
				"collectionName": "posts",
				"title":          "Test Post",
				"published":      true,
				"created":        "2024-01-01 00:00:00.000Z",
				"updated":        "2024-01-01 00:00:00.000Z",
			})

		// GET /api/collections/{collection}/records (list)
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/collections/"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"page":       1,
				"perPage":    20,
				"totalItems": 1,
				"totalPages": 1,
				"items": []map[string]any{
					{
						"id":             "rec1",
						"collectionId":   "col1",
						"collectionName": "posts",
						"title":          "Test Post",
						"published":      true,
						"created":        "2024-01-01 00:00:00.000Z",
						"updated":        "2024-01-01 00:00:00.000Z",
					},
				},
			})

		// PATCH /api/collections/{collection}/records/{id}
		case r.Method == http.MethodPatch && strings.HasPrefix(r.URL.Path, "/api/collections/"):
			body, _ := io.ReadAll(r.Body)
			var record map[string]any
			_ = json.Unmarshal(body, &record)
			if record == nil {
				record = make(map[string]any)
			}
			record["id"] = "rec1"
			record["collectionId"] = "col1"
			record["collectionName"] = "posts"
			record["created"] = "2024-01-01 00:00:00.000Z"
			record["updated"] = "2024-01-02 00:00:00.000Z"
			_ = json.NewEncoder(w).Encode(record)

		// DELETE /api/collections/{collection}/records/{id}
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/collections/"):
			w.WriteHeader(http.StatusNoContent)

		// GET /api/files/{collection}/{id}/{filename}
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/files/"):
			w.Header().Set("Content-Type", "application/octet-stream")
			_, _ = w.Write([]byte("fake file content"))

		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
}

// TestDocExamples_CorrectSignatures verifies the documentation code examples
// match the actual API signatures. This test ensures all examples compile correctly.
func TestDocExamples_CorrectSignatures(t *testing.T) {
	srv := mockDocServer(t)
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	// === Create ===
	// From docs: postService.Create(ctx, newPost)
	postService := NewTypedRecordService[docPost](client, "posts")
	newPost := &docPost{Title: "New Post", Content: strPtr("Content here"), Published: true}

	created, err := postService.Create(ctx, newPost)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created == nil {
		t.Fatal("Create returned nil")
	}
	t.Logf("Created: ID=%s, Title=%s", created.ID, created.Title)

	// === GetOne ===
	// From docs: postService.GetOne(ctx, "RECORD_ID", nil)
	post, err := postService.GetOne(ctx, "rec1", nil)
	if err != nil {
		t.Fatalf("GetOne failed: %v", err)
	}
	if post == nil {
		t.Fatal("GetOne returned nil")
	}
	t.Logf("GetOne: Title=%s", post.Title)

	// === GetList ===
	// From docs: postService.GetList(ctx, &pocketbase.ListOptions{...})
	result, err := postService.GetList(ctx, &ListOptions{
		Page:    1,
		PerPage: 20,
		Sort:    "-created",
		Filter:  "published = true",
		Expand:  "author",
	})
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}
	if result == nil {
		t.Fatal("GetList returned nil")
	}
	t.Logf("GetList: TotalItems=%d", result.TotalItems)

	// === Update ===
	// From docs: postService.Update(ctx, post.GetID(), post)
	post.SetTitle("Updated Title")
	updated, err := postService.Update(ctx, post.GetID(), post)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated == nil {
		t.Fatal("Update returned nil")
	}
	t.Logf("Updated: Title=%s", updated.Title)

	// === Delete (via embedded RecordService) ===
	// From docs: postService.Delete(ctx, postService.Collection, "RECORD_ID")
	err = postService.Delete(ctx, "rec1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	t.Log("Delete: success")
}

// TestDocExamples_WriteOptions verifies the WriteOptions documentation example.
func TestDocExamples_WriteOptions(t *testing.T) {
	srv := mockDocServer(t)
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	postService := NewTypedRecordService[docPost](client, "posts")

	// From docs: WriteOptions usage via embedded RecordService
	opts := &WriteOptions{
		Expand: "author",
		Fields: "id,title,author.name",
	}

	newPost := &docPost{Title: "New Post", Content: strPtr("Content")}

	// From docs: postService.RecordService.CreateWithOptions(ctx, postService.Collection, newPost.ToMap(), opts)
	created, err := postService.RecordService.CreateWithOptions(ctx, postService.Collection, newPost.ToMap(), opts)
	if err != nil {
		t.Fatalf("CreateWithOptions failed: %v", err)
	}
	if created == nil {
		t.Fatal("CreateWithOptions returned nil")
	}
	t.Log("CreateWithOptions: success")

	// From docs: postService.RecordService.UpdateWithOptions(ctx, postService.Collection, recordID, post.ToMap(), opts)
	updated, err := postService.RecordService.UpdateWithOptions(ctx, postService.Collection, "rec1", newPost.ToMap(), opts)
	if err != nil {
		t.Fatalf("UpdateWithOptions failed: %v", err)
	}
	if updated == nil {
		t.Fatal("UpdateWithOptions returned nil")
	}
	t.Log("UpdateWithOptions: success")
}

// TestDocExamples_FileService verifies the FileService documentation examples.
func TestDocExamples_FileService(t *testing.T) {
	srv := mockDocServer(t)
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	// From docs: client.Files.Upload(ctx, "posts", recordID, "image", "image.jpg", file)
	file := io.NopCloser(strings.NewReader("fake image content"))
	_, err := client.Files.Upload(ctx, "posts", "rec1", "image", "image.jpg", file)
	if err != nil {
		t.Fatalf("Files.Upload failed: %v", err)
	}
	t.Log("Files.Upload: success")

	// From docs: client.Files.Download(ctx, "posts", recordID, "image.jpg", nil)
	reader, err := client.Files.Download(ctx, "posts", "rec1", "image.jpg", nil)
	if err != nil {
		t.Fatalf("Files.Download failed: %v", err)
	}
	if reader != nil {
		reader.Close()
	}
	t.Log("Files.Download: success")

	// From docs: client.Files.GetFileURL("posts", recordID, "cover.jpg", nil)
	fileURL := client.Files.GetFileURL("posts", "rec1", "cover.jpg", nil)
	if fileURL == "" {
		t.Fatal("GetFileURL returned empty string")
	}
	t.Logf("Files.GetFileURL: %s", fileURL)

	// From docs: client.Files.Delete(ctx, "posts", recordID, "cover", "cover.jpg")
	_, err = client.Files.Delete(ctx, "posts", "rec1", "cover", "cover.jpg")
	if err != nil {
		t.Fatalf("Files.Delete failed: %v", err)
	}
	t.Log("Files.Delete: success")
}

// TestDocExamples_GetAll verifies the GetAll documentation example.
func TestDocExamples_GetAll(t *testing.T) {
	srv := mockDocServer(t)
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	postService := NewTypedRecordService[docPost](client, "posts")

	// From docs: postService.GetAll(ctx, &pocketbase.ListOptions{...})
	allPosts, err := postService.GetAll(ctx, &ListOptions{
		Filter: "published = true",
	})
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}
	t.Logf("GetAll: Total posts=%d", len(allPosts))
}

// TestDocExamples_RecordServiceDirect verifies direct RecordService usage.
func TestDocExamples_RecordServiceDirect(t *testing.T) {
	srv := mockDocServer(t)
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	// From docs: client.Records.Create(ctx, "posts", map[string]any{...})
	record, err := client.Records.Create(ctx, "posts", map[string]any{
		"title":   "New Post",
		"content": "Content here",
	})
	if err != nil {
		t.Fatalf("Records.Create failed: %v", err)
	}
	if record == nil {
		t.Fatal("Records.Create returned nil")
	}
	t.Log("Records.Create: success")

	// From docs: client.Records.GetOne(ctx, "posts", "RECORD_ID", nil)
	got, err := client.Records.GetOne(ctx, "posts", "rec1", nil)
	if err != nil {
		t.Fatalf("Records.GetOne failed: %v", err)
	}
	if got == nil {
		t.Fatal("Records.GetOne returned nil")
	}
	t.Log("Records.GetOne: success")

	// From docs: client.Records.Delete(ctx, "posts", "RECORD_ID")
	err = client.Records.Delete(ctx, "posts", "rec1")
	if err != nil {
		t.Fatalf("Records.Delete failed: %v", err)
	}
	t.Log("Records.Delete: success")
}

// TestDocExamples_BatchOperations verifies batch request creation.
func TestDocExamples_BatchOperations(t *testing.T) {
	srv := mockDocServer(t)
	defer srv.Close()

	client := NewClient(srv.URL)

	// From docs: client.Records.NewCreateRequest("posts", map[string]any{...})
	createReq, err := client.Records.NewCreateRequest("posts", map[string]any{
		"title": "New Post",
	})
	if err != nil {
		t.Fatalf("NewCreateRequest failed: %v", err)
	}
	if createReq == nil {
		t.Fatal("NewCreateRequest returned nil")
	}
	t.Log("NewCreateRequest: success")

	// From docs: client.Records.NewUpdateRequest("posts", "ID", map[string]any{...})
	updateReq, err := client.Records.NewUpdateRequest("posts", "rec1", map[string]any{
		"title": "Updated Post",
	})
	if err != nil {
		t.Fatalf("NewUpdateRequest failed: %v", err)
	}
	if updateReq == nil {
		t.Fatal("NewUpdateRequest returned nil")
	}
	t.Log("NewUpdateRequest: success")

	// From docs: client.Records.NewDeleteRequest("posts", "RECORD_ID")
	deleteReq, err := client.Records.NewDeleteRequest("posts", "rec1")
	if err != nil {
		t.Fatalf("NewDeleteRequest failed: %v", err)
	}
	if deleteReq == nil {
		t.Fatal("NewDeleteRequest returned nil")
	}
	t.Log("NewDeleteRequest: success")
}

func strPtr(s string) *string {
	return &s
}
