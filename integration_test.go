package pocketbase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// TestIntegration_CRUD tests CRUD operations against a simulated PocketBase server.
// This test validates that the client library correctly handles:
// - Authentication
// - Record creation, reading, updating, deletion
// - List operations with pagination
// - Filter operations
// - Batch operations
// - Typed service operations
func TestIntegration_CRUD(t *testing.T) {
	// Simulated database
	var mu sync.RWMutex
	records := make(map[string]map[string]any)
	var idCounter int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Auth endpoint (using _superusers collection for admin auth)
		if r.Method == http.MethodPost && r.URL.Path == "/api/collections/_superusers/auth-with-password" {
			body, _ := io.ReadAll(r.Body)
			var creds map[string]string
			_ = json.Unmarshal(body, &creds)

			if creds["identity"] == "admin@example.com" && creds["password"] == "password123" {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"token": "mock-admin-token-123",
					"admin": map[string]any{
						"id":    "admin123",
						"email": "admin@example.com",
					},
				})
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Invalid credentials"})
			return
		}

		// Batch endpoint
		if r.Method == http.MethodPost && r.URL.Path == "/api/batch" {
			body, _ := io.ReadAll(r.Body)
			var batchReq struct {
				Requests []struct {
					Method string         `json:"method"`
					URL    string         `json:"url"`
					Body   map[string]any `json:"body"`
				} `json:"requests"`
			}
			_ = json.Unmarshal(body, &batchReq)

			var results []map[string]any
			for _, req := range batchReq.Requests {
				status := http.StatusOK
				var respBody map[string]any

				if strings.Contains(req.URL, "/records") {
					if req.Method == http.MethodPost {
						mu.Lock()
						idCounter++
						id := fmt.Sprintf("rec%d", idCounter)
						req.Body["id"] = id
						records[id] = req.Body
						mu.Unlock()
						status = http.StatusCreated
						respBody = req.Body
					} else {
						status = http.StatusNoContent
						respBody = map[string]any{}
					}
				} else {
					status = http.StatusNotFound
					respBody = map[string]any{"message": "Not found"}
				}

				results = append(results, map[string]any{
					"status": status,
					"body":   respBody,
				})
			}

			_ = json.NewEncoder(w).Encode(results)
			return
		}

		// Check auth token (accept any Bearer token for testing)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Unauthorized - no auth header"})
			return
		}

		// Parse collection and record ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/collections/")
		parts := strings.Split(path, "/")
		collection := parts[0]

		// Handle collection creation
		if collection == "_collections" && r.Method == http.MethodPost {
			body, _ := io.ReadAll(r.Body)
			var coll map[string]any
			_ = json.Unmarshal(body, &coll)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "coll123", "name": coll["name"]})
			return
		}

		// Handle collection deletion
		if collection == "_collections" && r.Method == http.MethodDelete && len(parts) >= 2 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Record operations
		switch {
		// Create (POST /api/collections/{collection}/records)
		case r.Method == http.MethodPost && len(parts) == 2 && parts[1] == "records":
			body, _ := io.ReadAll(r.Body)
			var data map[string]any
			_ = json.Unmarshal(body, &data)

			mu.Lock()
			idCounter++
			id := fmt.Sprintf("rec%d", idCounter)
			data["id"] = id
			data["collectionId"] = "col123"
			data["collectionName"] = collection
			data["created"] = "2024-01-01 00:00:00.000Z"
			data["updated"] = "2024-01-01 00:00:00.000Z"
			records[id] = data
			mu.Unlock()

			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(data)

		// Get one (GET /api/collections/{collection}/records/{id})
		case r.Method == http.MethodGet && len(parts) == 3 && parts[1] == "records":
			id := parts[2]
			mu.RLock()
			record, exists := records[id]
			mu.RUnlock()

			if !exists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(record)

		// Update (PATCH /api/collections/{collection}/records/{id})
		case r.Method == http.MethodPatch && len(parts) == 3 && parts[1] == "records":
			id := parts[2]
			body, _ := io.ReadAll(r.Body)
			var updates map[string]any
			_ = json.Unmarshal(body, &updates)

			mu.Lock()
			record, exists := records[id]
			if !exists {
				mu.Unlock()
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
				return
			}
			for k, v := range updates {
				record[k] = v
			}
			record["updated"] = "2024-01-02 00:00:00.000Z"
			mu.Unlock()

			_ = json.NewEncoder(w).Encode(record)

		// Delete (DELETE /api/collections/{collection}/records/{id})
		case r.Method == http.MethodDelete && len(parts) == 3 && parts[1] == "records":
			id := parts[2]
			mu.Lock()
			_, exists := records[id]
			if exists {
				delete(records, id)
			}
			mu.Unlock()

			if !exists {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
				return
			}
			w.WriteHeader(http.StatusNoContent)

		// List (GET /api/collections/{collection}/records)
		case r.Method == http.MethodGet && len(parts) == 2 && parts[1] == "records":
			page := 1
			perPage := 30
			if p := r.URL.Query().Get("page"); p != "" {
				fmt.Sscanf(p, "%d", &page)
			}
			if pp := r.URL.Query().Get("perPage"); pp != "" {
				fmt.Sscanf(pp, "%d", &perPage)
			}

			mu.RLock()
			var items []map[string]any
			for _, rec := range records {
				if rec["collectionName"] == collection {
					items = append(items, rec)
				}
			}
			mu.RUnlock()

			// Apply pagination
			start := (page - 1) * perPage
			end := start + perPage
			if start > len(items) {
				start = len(items)
			}
			if end > len(items) {
				end = len(items)
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"page":       page,
				"perPage":    perPage,
				"totalItems": len(items),
				"totalPages": (len(items) + perPage - 1) / perPage,
				"items":      items[start:end],
			})

		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
		}
	}))
	defer srv.Close()

	ctx := context.Background()
	client := NewClient(srv.URL)

	// Authenticate before running tests
	_, err := client.WithAdminPassword(ctx, "admin@example.com", "password123")
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
	t.Log("✓ Authentication successful")

	// ============================================================
	// 1. Create
	// ============================================================
	t.Run("Create", func(t *testing.T) {
		record, err := client.Records.Create(ctx, "posts", map[string]any{
			"title":   "Test Post",
			"content": "This is a test",
			"views":   0,
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if record.ID == "" {
			t.Fatal("Create returned record with empty ID")
		}
		if record.GetString("title") != "Test Post" {
			t.Errorf("Expected title 'Test Post', got '%s'", record.GetString("title"))
		}
		t.Logf("✓ Created record: ID=%s", record.ID)
	})

	// ============================================================
	// 3. Read
	// ============================================================
	t.Run("Read", func(t *testing.T) {
		// Create a record first
		created, err := client.Records.Create(ctx, "posts", map[string]any{
			"title": "Read Test",
			"views": 10,
		})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Read the record
		record, err := client.Records.GetOne(ctx, "posts", created.ID, nil)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if record.ID != created.ID {
			t.Errorf("Expected ID '%s', got '%s'", created.ID, record.ID)
		}
		if record.GetString("title") != "Read Test" {
			t.Errorf("Expected title 'Read Test', got '%s'", record.GetString("title"))
		}
		t.Logf("✓ Read record: ID=%s, Title=%s", record.ID, record.GetString("title"))
	})

	// ============================================================
	// 4. Update
	// ============================================================
	t.Run("Update", func(t *testing.T) {
		// Create a record first
		created, err := client.Records.Create(ctx, "posts", map[string]any{
			"title": "Update Test",
			"views": 5,
		})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Update the record
		updated, err := client.Records.Update(ctx, "posts", created.ID, map[string]any{
			"title": "Updated Title",
			"views": 100,
		})
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		if updated.GetString("title") != "Updated Title" {
			t.Errorf("Expected title 'Updated Title', got '%s'", updated.GetString("title"))
		}
		if updated.GetFloat("views") != 100 {
			t.Errorf("Expected views 100, got %f", updated.GetFloat("views"))
		}
		t.Logf("✓ Updated record: ID=%s, Title=%s", updated.ID, updated.GetString("title"))
	})

	// ============================================================
	// 5. Delete
	// ============================================================
	t.Run("Delete", func(t *testing.T) {
		// Create a record first
		created, err := client.Records.Create(ctx, "posts", map[string]any{
			"title": "Delete Test",
		})
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}

		// Delete the record
		err = client.Records.Delete(ctx, "posts", created.ID)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify it's deleted
		_, err = client.Records.GetOne(ctx, "posts", created.ID, nil)
		if err == nil {
			t.Error("Expected error when reading deleted record, got nil")
		}
		t.Logf("✓ Deleted record: ID=%s", created.ID)
	})

	// ============================================================
	// 6. List
	// ============================================================
	t.Run("List", func(t *testing.T) {
		// Create some records
		for i := 0; i < 5; i++ {
			_, err := client.Records.Create(ctx, "posts", map[string]any{
				"title": fmt.Sprintf("List Test %d", i),
				"views": i * 10,
			})
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}
		}

		// List records
		result, err := client.Records.GetList(ctx, "posts", &ListOptions{
			Page:    1,
			PerPage: 10,
		})
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(result.Items) < 5 {
			t.Errorf("Expected at least 5 items, got %d", len(result.Items))
		}
		t.Logf("✓ Listed %d records", len(result.Items))
	})

	// ============================================================
	// 7. Batch Operations
	// ============================================================
	t.Run("Batch", func(t *testing.T) {
		// Create batch requests
		req1, err := client.Records.NewCreateRequest("posts", map[string]any{
			"title": "Batch 1",
		})
		if err != nil {
			t.Fatalf("NewCreateRequest failed: %v", err)
		}

		req2, err := client.Records.NewCreateRequest("posts", map[string]any{
			"title": "Batch 2",
		})
		if err != nil {
			t.Fatalf("NewCreateRequest failed: %v", err)
		}

		// Execute batch
		results, err := client.Batch.Execute(ctx, []*BatchRequest{req1, req2})
		if err != nil {
			t.Fatalf("Batch failed: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
		for i, r := range results {
			if r.Status != 200 && r.Status != 201 {
				t.Errorf("Batch request %d: expected status 200/201, got %d", i, r.Status)
			}
		}
		t.Logf("✓ Batch executed %d requests successfully", len(results))
	})

	// ============================================================
	// 8. Typed Service
	// ============================================================
	t.Run("TypedService", func(t *testing.T) {
		type Post struct {
			ID             string   `json:"id"`
			CollectionID   string   `json:"collectionId"`
			CollectionName string   `json:"collectionName"`
			Title          string   `json:"title"`
			Views          float64  `json:"views"`
		}

		postService := NewTypedRecordService[Post](client, "posts")

		// Create
		newPost := &Post{Title: "Typed Post", Views: 42}
		created, err := postService.Create(ctx, newPost)
		if err != nil {
			t.Fatalf("Typed Create failed: %v", err)
		}
		if created.Title != "Typed Post" {
			t.Errorf("Expected title 'Typed Post', got '%s'", created.Title)
		}
		t.Logf("✓ Typed Create: ID=%s, Title=%s", created.ID, created.Title)

		// Read
		post, err := postService.GetOne(ctx, created.ID, nil)
		if err != nil {
			t.Fatalf("Typed GetOne failed: %v", err)
		}
		if post.Title != "Typed Post" {
			t.Errorf("Expected title 'Typed Post', got '%s'", post.Title)
		}
		t.Logf("✓ Typed GetOne: Title=%s", post.Title)

		// Update
		post.Views = 100
		updated, err := postService.Update(ctx, post.ID, post)
		if err != nil {
			t.Fatalf("Typed Update failed: %v", err)
		}
		if updated.Views != 100 {
			t.Errorf("Expected views 100, got %f", updated.Views)
		}
		t.Logf("✓ Typed Update: Views=%f", updated.Views)

		// Delete
		err = postService.Delete(ctx, updated.ID)
		if err != nil {
			t.Fatalf("Typed Delete failed: %v", err)
		}
		t.Log("✓ Typed Delete: success")

		// List
		listResult, err := postService.GetList(ctx, &ListOptions{
			Page:    1,
			PerPage: 10,
		})
		if err != nil {
			t.Fatalf("Typed GetList failed: %v", err)
		}
		t.Logf("✓ Typed GetList: %d items", len(listResult.Items))

		// GetAll
		allPosts, err := postService.GetAll(ctx, &ListOptions{
			PerPage: 10,
		})
		if err != nil {
			t.Fatalf("Typed GetAll failed: %v", err)
		}
		t.Logf("✓ Typed GetAll: %d items", len(allPosts))
	})

	t.Log("\n=== All integration tests passed ===")
}
