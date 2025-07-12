package pocketbase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// TestRecordServiceGetList tests the GetList method of RecordService.
func TestRecordServiceGetList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") != "2" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(ListResult{Items: []*Record{}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	res, err := c.Records.GetList(context.Background(), "posts", &ListOptions{Page: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Items) != 0 {
		t.Fatalf("unexpected items: %v", res.Items)
	}
}

// TestRecordServiceGetListFieldsSkipTotal tests the GetList method with fields and skipTotal options.
func TestRecordServiceGetListFieldsSkipTotal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fields") != "id,title" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		if r.URL.Query().Get("skipTotal") != "1" {
			t.Fatalf("unexpected skipTotal: %s", r.URL.Query().Get("skipTotal"))
		}
		_ = json.NewEncoder(w).Encode(ListResult{Items: []*Record{}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &ListOptions{Fields: "id,title", SkipTotal: true}
	if _, err := c.Records.GetList(context.Background(), "posts", opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRecordServiceGetOne tests the GetOne method of RecordService.
func TestRecordServiceGetOne(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/collections/posts/records/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rec, err := c.Records.GetOne(context.Background(), "posts", "1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "1" {
		t.Fatalf("unexpected id: %s", rec.ID)
	}
}

// TestRecordServiceGetOneFields tests the GetOne method with fields option.
func TestRecordServiceGetOneFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fields") != "id,title" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &GetOneOptions{Fields: "id,title"}
	if _, err := c.Records.GetOne(context.Background(), "posts", "1", opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRecordServiceCreate tests the Create method of RecordService.
func TestRecordServiceCreate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/posts/records" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rec, err := c.Records.Create(context.Background(), "posts", map[string]string{"title": "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "1" {
		t.Fatalf("unexpected id: %s", rec.ID)
	}
}

// TestRecordServiceCreateWithQuery tests the CreateWithOptions method with query parameters.
func TestRecordServiceCreateWithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("expand") != "rel" {
			t.Fatalf("unexpected expand: %s", r.URL.Query().Get("expand"))
		}
		if r.URL.Query().Get("fields") != "id" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &WriteOptions{Expand: "rel", Fields: "id"}
	if _, err := c.Records.CreateWithOptions(context.Background(), "posts", map[string]string{"title": "hi"}, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRecordServiceUpdate tests the Update method of RecordService.
func TestRecordServiceUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/posts/records/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	rec, err := c.Records.Update(context.Background(), "posts", "1", map[string]string{"title": "new"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.ID != "1" {
		t.Fatalf("unexpected id: %s", rec.ID)
	}
}

func TestRecordServiceUpdateWithQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("expand") != "rel" {
			t.Fatalf("unexpected expand: %s", r.URL.Query().Get("expand"))
		}
		if r.URL.Query().Get("fields") != "id" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
		}
		_ = json.NewEncoder(w).Encode(Record{BaseModel: BaseModel{ID: "1"}})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	opts := &WriteOptions{Expand: "rel", Fields: "id"}
	if _, err := c.Records.UpdateWithOptions(context.Background(), "posts", "1", map[string]string{"title": "new"}, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordServiceDelete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/api/collections/posts/records/1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	if err := c.Records.Delete(context.Background(), "posts", "1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordServiceAuthHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "tok" {
			t.Fatalf("missing auth header: %s", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(ListResult{})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	c.AuthStore.Set("tok", &Record{CollectionName: "posts"})
	if _, err := c.Records.GetList(context.Background(), "posts", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRecordServiceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(APIError{Code: 404, Message: "no"})
	}))
	defer srv.Close()

	c := NewClient(srv.URL)
	_, err := c.Records.GetOne(context.Background(), "posts", "missing", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Code != 404 {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- 테스트를 위한 모델 정의 ---
// 실제로는 "models" 패키지에 있을 구조체들입니다.

type TestUser struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
