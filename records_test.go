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

// --- 테스트 코드 ---

func TestRecordService_AsFunctions(t *testing.T) {
	// 1. Mock 서버 설정
	// 이 서버는 실제 PocketBase API처럼 동작합니다.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 요청 URL 경로에 따라 다른 JSON 응답을 반환합니다.
		switch r.URL.Path {
		case "/api/collections/users/records/USER_ID_123":
			// GetOneAs 테스트를 위한 단일 레코드 응답
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "USER_ID_123",
				"collectionId": "users_collection_id",
				"collectionName": "users",
				"created": "2024-01-01 10:00:00.000Z",
				"updated": "2024-01-01 10:00:00.000Z",
				"name": "Gom Veteran",
				"avatar": "gom_avatar.png"
			}`))
		case "/api/collections/users/records":
			// GetListAs 테스트를 위한 레코드 목록 응답
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"perPage": 30,
				"totalPages": 1,
				"totalItems": 2,
				"items": [
					{
						"id": "USER_ID_1",
						"collectionId": "users_collection_id",
						"collectionName": "users",
						"name": "User One",
						"avatar": "one.png"
					},
					{
						"id": "USER_ID_2",
						"collectionId": "users_collection_id",
						"collectionName": "users",
						"name": "User Two",
						"avatar": "two.png"
					}
				]
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"not found"}`))
		}
	}))
	defer mockServer.Close() // 테스트가 끝나면 서버를 종료합니다.

	// 2. Mock 서버를 바라보는 클라이언트 생성
	client := NewClient(mockServer.URL)
	ctx := context.Background()

	// 3. 테스트 케이스 실행
	t.Run("GetOneAs", func(t *testing.T) {
		// 함수를 호출합니다.
		user, err := GetOneAs[TestUser](ctx, client.Records.(*RecordService), "users", "USER_ID_123", nil)

		// 에러가 없는지 확인합니다.
		if err != nil {
			t.Fatalf("GetOneAs returned an unexpected error: %v", err)
		}

		// 반환된 구조체의 값이 올바른지 확인합니다.
		if user.Name != "Gom Veteran" {
			t.Errorf("expected user name to be 'Gom Veteran', but got '%s'", user.Name)
		}
		if user.Avatar != "gom_avatar.png" {
			t.Errorf("expected avatar to be 'gom_avatar.png', but got '%s'", user.Avatar)
		}

		t.Logf("GetOneAs successful: %+v", user)
	})

	t.Run("GetListAs", func(t *testing.T) {
		// 함수를 호출합니다.
		users, err := GetListAs[TestUser](ctx, client.Records.(*RecordService), "users", nil)

		// 에러가 없는지 확인합니다.
		if err != nil {
			t.Fatalf("GetListAs returned an unexpected error: %v", err)
		}

		// 반환된 슬라이스의 길이가 올바른지 확인합니다.
		if len(users) != 2 {
			t.Fatalf("expected 2 users, but got %d", len(users))
		}

		// 각 항목의 값이 올바른지 확인합니다.
		if users[0].Name != "User One" {
			t.Errorf("expected first user's name to be 'User One', but got '%s'", users[0].Name)
		}
		if users[1].Name != "User Two" {
			t.Errorf("expected second user's name to be 'User Two', but got '%s'", users[1].Name)
		}

		t.Logf("GetListAs successful: found %d users", len(users))
	})
}
