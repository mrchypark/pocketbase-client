package pocketbase

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// 5.1 RecordService[T] 기본 기능 테스트

// TestRecordService_BasicFunctionality는 RecordService[T]의 기본 기능을 테스트합니다.
func TestRecordService_BasicFunctionality(t *testing.T) {
	// Admin 타입으로 제네릭 서비스 생성 테스트
	client := &Client{}
	adminService := NewRecordService[Admin](client, "admins")

	// 타입 검증 - 컴파일 타임에 확인됨
	if adminService == nil {
		t.Fatal("NewRecordService should return non-nil service")
	}

	// 서비스가 올바른 컬렉션 이름을 가지는지 확인
	if adminService.collectionName != "admins" {
		t.Errorf("Expected collection name 'admins', got '%s'", adminService.collectionName)
	}

	// 클라이언트가 올바르게 설정되었는지 확인
	if adminService.client != client {
		t.Error("Client should be properly set")
	}
}

// TestRecordService_MultipleTypes는 다양한 타입으로 제네릭 서비스를 테스트합니다.
func TestRecordService_MultipleTypes(t *testing.T) {
	client := &Client{}

	// 1. Admin 타입으로 서비스 생성
	adminService := NewRecordService[Admin](client, "admins")
	if adminService.collectionName != "admins" {
		t.Errorf("Expected collection name 'admins', got '%s'", adminService.collectionName)
	}

	// 2. Record 타입으로 서비스 생성
	recordService := NewRecordService[Record](client, "posts")
	if recordService.collectionName != "posts" {
		t.Errorf("Expected collection name 'posts', got '%s'", recordService.collectionName)
	}

	// 3. 사용자 정의 타입으로 서비스 생성
	type Post struct {
		BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}

	postService := NewRecordService[Post](client, "posts")
	if postService.collectionName != "posts" {
		t.Errorf("Expected collection name 'posts', got '%s'", postService.collectionName)
	}

	// 4. 복잡한 사용자 정의 타입으로 서비스 생성
	type Article struct {
		BaseModel
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		Tags        []string `json:"tags"`
		PublishedAt string   `json:"published_at"`
		AuthorID    string   `json:"author_id"`
		ViewCount   int      `json:"view_count"`
		IsPublished bool     `json:"is_published"`
	}

	articleService := NewRecordService[Article](client, "articles")
	if articleService.collectionName != "articles" {
		t.Errorf("Expected collection name 'articles', got '%s'", articleService.collectionName)
	}

	// 5. 타입 안전성 검증 - 컴파일 타임에 확인됨
	var _ RecordServiceAPI[Admin] = adminService
	var _ RecordServiceAPI[Record] = recordService
	var _ RecordServiceAPI[Post] = postService
	var _ RecordServiceAPI[Article] = articleService
}

// TestRecordService_CRUDOperations는 각 CRUD 메서드의 정상 동작을 검증합니다.
func TestRecordService_CRUDOperations(t *testing.T) {
	// 테스트용 HTTP 서버 설정
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/api/collections/posts/records" && r.Method == http.MethodGet:
			// GetList 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"perPage": 20,
				"totalItems": 2,
				"totalPages": 1,
				"items": [
					{
						"id": "post1",
						"collectionId": "posts_id",
						"collectionName": "posts",
						"title": "Test Post 1",
						"content": "Test content 1",
						"status": "published"
					},
					{
						"id": "post2",
						"collectionId": "posts_id",
						"collectionName": "posts",
						"title": "Test Post 2",
						"content": "Test content 2",
						"status": "draft"
					}
				]
			}`))
		case r.URL.Path == "/api/collections/posts/records/post1" && r.Method == http.MethodGet:
			// GetOne 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "post1",
				"collectionId": "posts_id",
				"collectionName": "posts",
				"title": "Test Post 1",
				"content": "Test content 1",
				"status": "published"
			}`))
		case r.URL.Path == "/api/collections/posts/records" && r.Method == http.MethodPost:
			// Create 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "new_post",
				"collectionId": "posts_id",
				"collectionName": "posts",
				"title": "New Post",
				"content": "New content",
				"status": "draft"
			}`))
		case r.URL.Path == "/api/collections/posts/records/post1" && r.Method == http.MethodPatch:
			// Update 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "post1",
				"collectionId": "posts_id",
				"collectionName": "posts",
				"title": "Updated Post",
				"content": "Updated content",
				"status": "published"
			}`))
		case r.URL.Path == "/api/collections/posts/records/post1" && r.Method == http.MethodDelete:
			// Delete 요청 처리
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// 클라이언트 생성
	client := NewClient(srv.URL)

	// 사용자 정의 타입 정의
	type Post struct {
		BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
		Status  string `json:"status"`
	}

	// 제네릭 서비스 생성
	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 1. GetList 메서드 테스트
	t.Run("GetList", func(t *testing.T) {
		result, err := postService.GetList(ctx, nil)
		if err != nil {
			t.Fatalf("GetList failed: %v", err)
		}

		if result.Page != 1 {
			t.Errorf("Expected page 1, got %d", result.Page)
		}
		if result.TotalItems != 2 {
			t.Errorf("Expected totalItems 2, got %d", result.TotalItems)
		}
		if len(result.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(result.Items))
		}

		// 첫 번째 아이템 검증
		firstItem := result.Items[0]
		if firstItem.ID != "post1" {
			t.Errorf("Expected first item ID 'post1', got '%s'", firstItem.ID)
		}
		if firstItem.Title != "Test Post 1" {
			t.Errorf("Expected first item title 'Test Post 1', got '%s'", firstItem.Title)
		}
		if firstItem.Status != "published" {
			t.Errorf("Expected first item status 'published', got '%s'", firstItem.Status)
		}
	})

	// 2. GetOne 메서드 테스트
	t.Run("GetOne", func(t *testing.T) {
		result, err := postService.GetOne(ctx, "post1", nil)
		if err != nil {
			t.Fatalf("GetOne failed: %v", err)
		}

		if result.ID != "post1" {
			t.Errorf("Expected ID 'post1', got '%s'", result.ID)
		}
		if result.Title != "Test Post 1" {
			t.Errorf("Expected title 'Test Post 1', got '%s'", result.Title)
		}
		if result.Content != "Test content 1" {
			t.Errorf("Expected content 'Test content 1', got '%s'", result.Content)
		}
		if result.Status != "published" {
			t.Errorf("Expected status 'published', got '%s'", result.Status)
		}
	})

	// 3. Create 메서드 테스트
	t.Run("Create", func(t *testing.T) {
		newPost := &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
			Title:   "New Post",
			Content: "New content",
			Status:  "draft",
		}

		result, err := postService.Create(ctx, newPost, nil)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if result.ID != "new_post" {
			t.Errorf("Expected ID 'new_post', got '%s'", result.ID)
		}
		if result.Title != "New Post" {
			t.Errorf("Expected title 'New Post', got '%s'", result.Title)
		}
		if result.Content != "New content" {
			t.Errorf("Expected content 'New content', got '%s'", result.Content)
		}
		if result.Status != "draft" {
			t.Errorf("Expected status 'draft', got '%s'", result.Status)
		}
	})

	// 4. Update 메서드 테스트
	t.Run("Update", func(t *testing.T) {
		updatePost := &Post{
			BaseModel: BaseModel{
				ID:             "post1",
				CollectionName: "posts",
			},
			Title:   "Updated Post",
			Content: "Updated content",
			Status:  "published",
		}

		result, err := postService.Update(ctx, "post1", updatePost, nil)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		if result.ID != "post1" {
			t.Errorf("Expected ID 'post1', got '%s'", result.ID)
		}
		if result.Title != "Updated Post" {
			t.Errorf("Expected title 'Updated Post', got '%s'", result.Title)
		}
		if result.Content != "Updated content" {
			t.Errorf("Expected content 'Updated content', got '%s'", result.Content)
		}
		if result.Status != "published" {
			t.Errorf("Expected status 'published', got '%s'", result.Status)
		}
	})

	// 5. Delete 메서드 테스트
	t.Run("Delete", func(t *testing.T) {
		err := postService.Delete(ctx, "post1")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		// Delete는 성공적으로 완료되면 에러가 없어야 함
	})
}

// TestRecordService_TypeSafetyCompileTime는 타입 안전성 컴파일 타임 검증을 테스트합니다.
func TestRecordService_TypeSafetyCompileTime(t *testing.T) {
	client := &Client{}

	// 1. Admin 타입 서비스
	adminService := NewRecordService[Admin](client, "admins")
	var _ RecordServiceAPI[Admin] = adminService

	// 2. 사용자 정의 타입 서비스
	type User struct {
		BaseModel
		Email    string `json:"email"`
		Username string `json:"username"`
		IsActive bool   `json:"is_active"`
	}

	userService := NewRecordService[User](client, "users")
	var _ RecordServiceAPI[User] = userService

	// 3. 복잡한 타입 서비스
	type Product struct {
		BaseModel
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Price       float64                `json:"price"`
		CategoryID  string                 `json:"category_id"`
		Tags        []string               `json:"tags"`
		InStock     bool                   `json:"in_stock"`
		Metadata    map[string]interface{} `json:"metadata"`
	}

	productService := NewRecordService[Product](client, "products")
	var _ RecordServiceAPI[Product] = productService

	// 4. 메서드 시그니처 검증 (컴파일 타임)
	ctx := context.Background()

	// 이 호출들은 컴파일되어야 함 (실제 실행은 하지 않음)
	if false {
		// Admin 서비스 메서드들
		_, _ = adminService.GetList(ctx, nil)
		_, _ = adminService.GetOne(ctx, "id", nil)
		_, _ = adminService.Create(ctx, &Admin{}, nil)
		_, _ = adminService.Update(ctx, "id", &Admin{}, nil)
		_ = adminService.Delete(ctx, "id")

		// User 서비스 메서드들
		_, _ = userService.GetList(ctx, &ListOptions{})
		_, _ = userService.GetOne(ctx, "id", &GetOneOptions{})
		_, _ = userService.Create(ctx, &User{}, &WriteOptions{})
		_, _ = userService.Update(ctx, "id", &User{}, &WriteOptions{})
		_ = userService.Delete(ctx, "id")

		// Product 서비스 메서드들
		_, _ = productService.GetList(ctx, &ListOptions{Page: 1})
		_, _ = productService.GetOne(ctx, "id", &GetOneOptions{Expand: "category"})
		_, _ = productService.Create(ctx, &Product{}, &WriteOptions{Fields: "id,name"})
		_, _ = productService.Update(ctx, "id", &Product{}, &WriteOptions{Expand: "category"})
		_ = productService.Delete(ctx, "id")
	}

	// 5. 반환 타입 검증 (컴파일 타임)
	if false {
		// GetList 반환 타입 검증
		adminList, _ := adminService.GetList(ctx, nil)
		var _ *ListResultAs[Admin] = adminList

		userList, _ := userService.GetList(ctx, nil)
		var _ *ListResultAs[User] = userList

		// GetOne 반환 타입 검증
		admin, _ := adminService.GetOne(ctx, "id", nil)
		var _ *Admin = admin

		user, _ := userService.GetOne(ctx, "id", nil)
		var _ *User = user

		// Create 반환 타입 검증
		createdAdmin, _ := adminService.Create(ctx, &Admin{}, nil)
		var _ *Admin = createdAdmin

		createdUser, _ := userService.Create(ctx, &User{}, nil)
		var _ *User = createdUser

		// Update 반환 타입 검증
		updatedAdmin, _ := adminService.Update(ctx, "id", &Admin{}, nil)
		var _ *Admin = updatedAdmin

		updatedUser, _ := userService.Update(ctx, "id", &User{}, nil)
		var _ *User = updatedUser
	}
}

// TestRecordService_AdminTypeCRUD는 Admin 타입으로 실제 CRUD 작업을 테스트합니다.
func TestRecordService_AdminTypeCRUD(t *testing.T) {
	// Admin 타입 전용 테스트 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/api/collections/admins/records" && r.Method == http.MethodGet:
			// Admin GetList 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"perPage": 20,
				"totalItems": 2,
				"totalPages": 1,
				"items": [
					{
						"id": "admin1",
						"collectionId": "admins_id",
						"collectionName": "admins",
						"email": "admin1@example.com",
						"avatar": 1
					},
					{
						"id": "admin2",
						"collectionId": "admins_id",
						"collectionName": "admins",
						"email": "admin2@example.com",
						"avatar": 2
					}
				]
			}`))
		case r.URL.Path == "/api/collections/admins/records/admin1" && r.Method == http.MethodGet:
			// Admin GetOne 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "admin1",
				"collectionId": "admins_id",
				"collectionName": "admins",
				"email": "admin1@example.com",
				"avatar": 1
			}`))
		case r.URL.Path == "/api/collections/admins/records" && r.Method == http.MethodPost:
			// Admin Create 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "new_admin",
				"collectionId": "admins_id",
				"collectionName": "admins",
				"email": "newadmin@example.com",
				"avatar": 0
			}`))
		case r.URL.Path == "/api/collections/admins/records/admin1" && r.Method == http.MethodPatch:
			// Admin Update 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "admin1",
				"collectionId": "admins_id",
				"collectionName": "admins",
				"email": "updated@example.com",
				"avatar": 3
			}`))
		case r.URL.Path == "/api/collections/admins/records/admin1" && r.Method == http.MethodDelete:
			// Admin Delete 요청 처리
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// 클라이언트 생성
	client := NewClient(srv.URL)

	// Admin 타입 제네릭 서비스 생성
	adminService := NewRecordService[Admin](client, "admins")
	ctx := context.Background()

	// 1. Admin GetList 테스트
	t.Run("Admin_GetList", func(t *testing.T) {
		result, err := adminService.GetList(ctx, nil)
		if err != nil {
			t.Fatalf("Admin GetList failed: %v", err)
		}

		if len(result.Items) != 2 {
			t.Errorf("Expected 2 admin items, got %d", len(result.Items))
		}

		// 첫 번째 Admin 검증
		firstAdmin := result.Items[0]
		if firstAdmin.ID != "admin1" {
			t.Errorf("Expected first admin ID 'admin1', got '%s'", firstAdmin.ID)
		}
		if firstAdmin.Email != "admin1@example.com" {
			t.Errorf("Expected first admin email 'admin1@example.com', got '%s'", firstAdmin.Email)
		}
		if firstAdmin.Avatar != 1 {
			t.Errorf("Expected first admin avatar 1, got %d", firstAdmin.Avatar)
		}
	})

	// 2. Admin GetOne 테스트
	t.Run("Admin_GetOne", func(t *testing.T) {
		result, err := adminService.GetOne(ctx, "admin1", nil)
		if err != nil {
			t.Fatalf("Admin GetOne failed: %v", err)
		}

		if result.ID != "admin1" {
			t.Errorf("Expected admin ID 'admin1', got '%s'", result.ID)
		}
		if result.Email != "admin1@example.com" {
			t.Errorf("Expected admin email 'admin1@example.com', got '%s'", result.Email)
		}
		if result.Avatar != 1 {
			t.Errorf("Expected admin avatar 1, got %d", result.Avatar)
		}
	})

	// 3. Admin Create 테스트
	t.Run("Admin_Create", func(t *testing.T) {
		newAdmin := &Admin{
			BaseModel: BaseModel{
				CollectionName: "admins",
			},
			Email:  "newadmin@example.com",
			Avatar: 0,
		}

		result, err := adminService.Create(ctx, newAdmin, nil)
		if err != nil {
			t.Fatalf("Admin Create failed: %v", err)
		}

		if result.ID != "new_admin" {
			t.Errorf("Expected admin ID 'new_admin', got '%s'", result.ID)
		}
		if result.Email != "newadmin@example.com" {
			t.Errorf("Expected admin email 'newadmin@example.com', got '%s'", result.Email)
		}
		if result.Avatar != 0 {
			t.Errorf("Expected admin avatar 0, got %d", result.Avatar)
		}
	})

	// 4. Admin Update 테스트
	t.Run("Admin_Update", func(t *testing.T) {
		updateAdmin := &Admin{
			BaseModel: BaseModel{
				ID:             "admin1",
				CollectionName: "admins",
			},
			Email:  "updated@example.com",
			Avatar: 3,
		}

		result, err := adminService.Update(ctx, "admin1", updateAdmin, nil)
		if err != nil {
			t.Fatalf("Admin Update failed: %v", err)
		}

		if result.ID != "admin1" {
			t.Errorf("Expected admin ID 'admin1', got '%s'", result.ID)
		}
		if result.Email != "updated@example.com" {
			t.Errorf("Expected admin email 'updated@example.com', got '%s'", result.Email)
		}
		if result.Avatar != 3 {
			t.Errorf("Expected admin avatar 3, got %d", result.Avatar)
		}
	})

	// 5. Admin Delete 테스트
	t.Run("Admin_Delete", func(t *testing.T) {
		err := adminService.Delete(ctx, "admin1")
		if err != nil {
			t.Fatalf("Admin Delete failed: %v", err)
		}
		// Delete는 성공적으로 완료되면 에러가 없어야 함
	})
}

// TestRecordService_RecordTypeCRUD는 Record 타입으로 실제 CRUD 작업을 테스트합니다.
func TestRecordService_RecordTypeCRUD(t *testing.T) {
	// Record 타입 전용 테스트 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/api/collections/generic_records/records" && r.Method == http.MethodGet:
			// Record GetList 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"perPage": 20,
				"totalItems": 1,
				"totalPages": 1,
				"items": [
					{
						"id": "record1",
						"collectionId": "generic_records_id",
						"collectionName": "generic_records",
						"title": "Generic Record",
						"description": "This is a generic record",
						"custom_field": "custom_value"
					}
				]
			}`))
		case r.URL.Path == "/api/collections/generic_records/records/record1" && r.Method == http.MethodGet:
			// Record GetOne 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "record1",
				"collectionId": "generic_records_id",
				"collectionName": "generic_records",
				"title": "Generic Record",
				"description": "This is a generic record",
				"custom_field": "custom_value"
			}`))
		case r.URL.Path == "/api/collections/generic_records/records" && r.Method == http.MethodPost:
			// Record Create 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "new_record",
				"collectionId": "generic_records_id",
				"collectionName": "generic_records",
				"title": "New Generic Record",
				"description": "This is a new generic record"
			}`))
		case r.URL.Path == "/api/collections/generic_records/records/record1" && r.Method == http.MethodPatch:
			// Record Update 요청 처리
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "record1",
				"collectionId": "generic_records_id",
				"collectionName": "generic_records",
				"title": "Updated Generic Record",
				"description": "This is an updated generic record"
			}`))
		case r.URL.Path == "/api/collections/generic_records/records/record1" && r.Method == http.MethodDelete:
			// Record Delete 요청 처리
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// 클라이언트 생성
	client := NewClient(srv.URL)

	// Record 타입 제네릭 서비스 생성
	recordService := NewRecordService[Record](client, "generic_records")
	ctx := context.Background()

	// 1. Record GetList 테스트
	t.Run("Record_GetList", func(t *testing.T) {
		result, err := recordService.GetList(ctx, nil)
		if err != nil {
			t.Fatalf("Record GetList failed: %v", err)
		}

		if len(result.Items) != 1 {
			t.Errorf("Expected 1 record item, got %d", len(result.Items))
		}

		// 첫 번째 Record 검증
		firstRecord := result.Items[0]
		if firstRecord.ID != "record1" {
			t.Errorf("Expected first record ID 'record1', got '%s'", firstRecord.ID)
		}
		if firstRecord.CollectionName != "generic_records" {
			t.Errorf("Expected first record collection 'generic_records', got '%s'", firstRecord.CollectionName)
		}

		// Record의 동적 필드 검증
		if title := firstRecord.GetString("title"); title != "Generic Record" {
			t.Errorf("Expected record title 'Generic Record', got '%s'", title)
		}
		if description := firstRecord.GetString("description"); description != "This is a generic record" {
			t.Errorf("Expected record description 'This is a generic record', got '%s'", description)
		}
		if customField := firstRecord.GetString("custom_field"); customField != "custom_value" {
			t.Errorf("Expected record custom_field 'custom_value', got '%s'", customField)
		}
	})

	// 2. Record GetOne 테스트
	t.Run("Record_GetOne", func(t *testing.T) {
		result, err := recordService.GetOne(ctx, "record1", nil)
		if err != nil {
			t.Fatalf("Record GetOne failed: %v", err)
		}

		if result.ID != "record1" {
			t.Errorf("Expected record ID 'record1', got '%s'", result.ID)
		}
		if result.CollectionName != "generic_records" {
			t.Errorf("Expected record collection 'generic_records', got '%s'", result.CollectionName)
		}

		// 동적 필드 검증
		if title := result.GetString("title"); title != "Generic Record" {
			t.Errorf("Expected record title 'Generic Record', got '%s'", title)
		}
	})

	// 3. Record Create 테스트
	t.Run("Record_Create", func(t *testing.T) {
		newRecord := &Record{
			BaseModel: BaseModel{
				CollectionName: "generic_records",
			},
		}
		// 동적 필드 설정
		newRecord.Set("title", "New Generic Record")
		newRecord.Set("description", "This is a new generic record")

		result, err := recordService.Create(ctx, newRecord, nil)
		if err != nil {
			t.Fatalf("Record Create failed: %v", err)
		}

		if result.ID != "new_record" {
			t.Errorf("Expected record ID 'new_record', got '%s'", result.ID)
		}
		if result.CollectionName != "generic_records" {
			t.Errorf("Expected record collection 'generic_records', got '%s'", result.CollectionName)
		}

		// 생성된 레코드의 동적 필드 검증
		if title := result.GetString("title"); title != "New Generic Record" {
			t.Errorf("Expected record title 'New Generic Record', got '%s'", title)
		}
	})

	// 4. Record Update 테스트
	t.Run("Record_Update", func(t *testing.T) {
		updateRecord := &Record{
			BaseModel: BaseModel{
				ID:             "record1",
				CollectionName: "generic_records",
			},
		}
		// 동적 필드 설정
		updateRecord.Set("title", "Updated Generic Record")
		updateRecord.Set("description", "This is an updated generic record")

		result, err := recordService.Update(ctx, "record1", updateRecord, nil)
		if err != nil {
			t.Fatalf("Record Update failed: %v", err)
		}

		if result.ID != "record1" {
			t.Errorf("Expected record ID 'record1', got '%s'", result.ID)
		}

		// 업데이트된 레코드의 동적 필드 검증
		if title := result.GetString("title"); title != "Updated Generic Record" {
			t.Errorf("Expected record title 'Updated Generic Record', got '%s'", title)
		}
	})

	// 5. Record Delete 테스트
	t.Run("Record_Delete", func(t *testing.T) {
		err := recordService.Delete(ctx, "record1")
		if err != nil {
			t.Fatalf("Record Delete failed: %v", err)
		}
		// Delete는 성공적으로 완료되면 에러가 없어야 함
	})
}

// 5.2 에러 처리 테스트

// TestRecordService_HTTPErrorHandling은 HTTP 에러 상황에서의 제네릭 서비스 동작을 검증합니다.
func TestRecordService_HTTPErrorHandling(t *testing.T) {
	// 다양한 HTTP 에러를 반환하는 테스트 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/collections/posts/records":
			if r.Method == http.MethodGet {
				// 404 Not Found 에러
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": 404, "message": "Collection not found", "data": {}}`))
			} else if r.Method == http.MethodPost {
				// 400 Bad Request 에러
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"code": 400, "message": "Invalid request data", "data": {"title": {"code": "validation_required", "message": "Missing required value."}}}`))
			}
		case "/api/collections/posts/records/nonexistent":
			if r.Method == http.MethodGet {
				// 404 Not Found 에러
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": 404, "message": "Record not found", "data": {}}`))
			} else if r.Method == http.MethodPatch {
				// 403 Forbidden 에러
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`{"code": 403, "message": "Insufficient permissions", "data": {}}`))
			} else if r.Method == http.MethodDelete {
				// 409 Conflict 에러
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`{"code": 409, "message": "Record cannot be deleted due to constraints", "data": {}}`))
			}
		case "/api/collections/posts/records/server_error":
			// 500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"code": 500, "message": "Internal server error", "data": {}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// 클라이언트 생성
	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 1. GetList 404 에러 테스트
	t.Run("GetList_404Error", func(t *testing.T) {
		_, err := postService.GetList(ctx, nil)
		if err == nil {
			t.Fatal("Expected error for 404 response, got nil")
		}

		// 에러 메시지에 "pocketbase: fetch typed records list:" 접두사가 있는지 확인
		if !strings.Contains(err.Error(), "pocketbase: fetch typed records list:") {
			t.Errorf("Expected error message to contain 'pocketbase: fetch typed records list:', got: %v", err)
		}
	})

	// 2. GetOne 404 에러 테스트
	t.Run("GetOne_404Error", func(t *testing.T) {
		_, err := postService.GetOne(ctx, "nonexistent", nil)
		if err == nil {
			t.Fatal("Expected error for 404 response, got nil")
		}

		// 에러 메시지에 "pocketbase: fetch typed record:" 접두사가 있는지 확인
		if !strings.Contains(err.Error(), "pocketbase: fetch typed record:") {
			t.Errorf("Expected error message to contain 'pocketbase: fetch typed record:', got: %v", err)
		}
	})

	// 3. Create 400 에러 테스트
	t.Run("Create_400Error", func(t *testing.T) {
		newPost := &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
			// Title 필드를 의도적으로 비워둠 (validation error 유발)
			Content: "Test content",
		}

		_, err := postService.Create(ctx, newPost, nil)
		if err == nil {
			t.Fatal("Expected error for 400 response, got nil")
		}

		// 에러 메시지에 "pocketbase: create typed record:" 접두사가 있는지 확인
		if !strings.Contains(err.Error(), "pocketbase: create typed record:") {
			t.Errorf("Expected error message to contain 'pocketbase: create typed record:', got: %v", err)
		}
	})

	// 4. Update 403 에러 테스트
	t.Run("Update_403Error", func(t *testing.T) {
		updatePost := &Post{
			BaseModel: BaseModel{
				ID:             "nonexistent",
				CollectionName: "posts",
			},
			Title:   "Updated Title",
			Content: "Updated content",
		}

		_, err := postService.Update(ctx, "nonexistent", updatePost, nil)
		if err == nil {
			t.Fatal("Expected error for 403 response, got nil")
		}

		// 에러 메시지에 "pocketbase: update typed record:" 접두사가 있는지 확인
		if !strings.Contains(err.Error(), "pocketbase: update typed record:") {
			t.Errorf("Expected error message to contain 'pocketbase: update typed record:', got: %v", err)
		}
	})

	// 5. Delete 409 에러 테스트
	t.Run("Delete_409Error", func(t *testing.T) {
		err := postService.Delete(ctx, "nonexistent")
		if err == nil {
			t.Fatal("Expected error for 409 response, got nil")
		}

		// 에러 메시지에 "pocketbase: delete typed record:" 접두사가 있는지 확인
		if !strings.Contains(err.Error(), "pocketbase: delete typed record:") {
			t.Errorf("Expected error message to contain 'pocketbase: delete typed record:', got: %v", err)
		}
	})

	// 6. 500 Internal Server Error 테스트
	t.Run("GetOne_500Error", func(t *testing.T) {
		_, err := postService.GetOne(ctx, "server_error", nil)
		if err == nil {
			t.Fatal("Expected error for 500 response, got nil")
		}

		// 에러 메시지에 "pocketbase: fetch typed record:" 접두사가 있는지 확인
		if !strings.Contains(err.Error(), "pocketbase: fetch typed record:") {
			t.Errorf("Expected error message to contain 'pocketbase: fetch typed record:', got: %v", err)
		}
	})
}

// TestRecordService_ErrorConsistency는 기존 에러 처리 방식과의 일관성을 확인합니다.
func TestRecordService_ErrorConsistency(t *testing.T) {
	// 동일한 에러를 반환하는 테스트 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 모든 요청에 대해 동일한 404 에러 반환
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"code": 404, "message": "Not found", "data": {}}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	type Post struct {
		BaseModel
		Title string `json:"title"`
	}

	// 제네릭 서비스와 레거시 서비스 생성
	postService := NewRecordService[Post](client, "posts")
	legacyService := &RecordServiceLegacy{Client: client}

	// 1. GetList 에러 일관성 테스트
	t.Run("GetList_ErrorConsistency", func(t *testing.T) {
		// 제네릭 서비스 에러
		_, genericErr := postService.GetList(ctx, nil)
		if genericErr == nil {
			t.Fatal("Expected error from generic service")
		}

		// 레거시 서비스 에러
		_, legacyErr := legacyService.GetList(ctx, "posts", nil)
		if legacyErr == nil {
			t.Fatal("Expected error from legacy service")
		}

		// 에러 메시지 패턴 비교 (정확한 일치는 아니지만 유사한 패턴이어야 함)
		if !strings.Contains(genericErr.Error(), "pocketbase:") {
			t.Errorf("Generic error should contain 'pocketbase:' prefix, got: %v", genericErr)
		}
		if !strings.Contains(legacyErr.Error(), "pocketbase:") {
			t.Errorf("Legacy error should contain 'pocketbase:' prefix, got: %v", legacyErr)
		}
	})

	// 2. GetOne 에러 일관성 테스트
	t.Run("GetOne_ErrorConsistency", func(t *testing.T) {
		// 제네릭 서비스 에러
		_, genericErr := postService.GetOne(ctx, "test", nil)
		if genericErr == nil {
			t.Fatal("Expected error from generic service")
		}

		// 레거시 서비스 에러
		_, legacyErr := legacyService.GetOne(ctx, "posts", "test", nil)
		if legacyErr == nil {
			t.Fatal("Expected error from legacy service")
		}

		// 에러 메시지 패턴 비교
		if !strings.Contains(genericErr.Error(), "pocketbase:") {
			t.Errorf("Generic error should contain 'pocketbase:' prefix, got: %v", genericErr)
		}
		if !strings.Contains(legacyErr.Error(), "pocketbase:") {
			t.Errorf("Legacy error should contain 'pocketbase:' prefix, got: %v", legacyErr)
		}
	})

	// 3. Create 에러 일관성 테스트
	t.Run("Create_ErrorConsistency", func(t *testing.T) {
		testPost := &Post{Title: "Test"}

		// 제네릭 서비스 에러
		_, genericErr := postService.Create(ctx, testPost, nil)
		if genericErr == nil {
			t.Fatal("Expected error from generic service")
		}

		// 레거시 서비스 에러 (CreateAs 함수 사용)
		_, legacyErr := CreateAs[Post](ctx, legacyService, "posts", testPost, nil)
		if legacyErr == nil {
			t.Fatal("Expected error from legacy service")
		}

		// 에러 메시지 패턴 비교
		if !strings.Contains(genericErr.Error(), "pocketbase:") {
			t.Errorf("Generic error should contain 'pocketbase:' prefix, got: %v", genericErr)
		}
		if !strings.Contains(legacyErr.Error(), "pocketbase:") {
			t.Errorf("Legacy error should contain 'pocketbase:' prefix, got: %v", legacyErr)
		}
	})

	// 4. Update 에러 일관성 테스트
	t.Run("Update_ErrorConsistency", func(t *testing.T) {
		testPost := &Post{
			BaseModel: BaseModel{ID: "test"},
			Title:     "Test",
		}

		// 제네릭 서비스 에러
		_, genericErr := postService.Update(ctx, "test", testPost, nil)
		if genericErr == nil {
			t.Fatal("Expected error from generic service")
		}

		// 레거시 서비스 에러 (UpdateAs 함수 사용)
		_, legacyErr := UpdateAs[Post](ctx, legacyService, "posts", "test", testPost, nil)
		if legacyErr == nil {
			t.Fatal("Expected error from legacy service")
		}

		// 에러 메시지 패턴 비교
		if !strings.Contains(genericErr.Error(), "pocketbase:") {
			t.Errorf("Generic error should contain 'pocketbase:' prefix, got: %v", genericErr)
		}
		if !strings.Contains(legacyErr.Error(), "pocketbase:") {
			t.Errorf("Legacy error should contain 'pocketbase:' prefix, got: %v", legacyErr)
		}
	})

	// 5. Delete 에러 일관성 테스트
	t.Run("Delete_ErrorConsistency", func(t *testing.T) {
		// 제네릭 서비스 에러
		genericErr := postService.Delete(ctx, "test")
		if genericErr == nil {
			t.Fatal("Expected error from generic service")
		}

		// 레거시 서비스 에러
		legacyErr := legacyService.Delete(ctx, "posts", "test")
		if legacyErr == nil {
			t.Fatal("Expected error from legacy service")
		}

		// 에러 메시지 패턴 비교
		if !strings.Contains(genericErr.Error(), "pocketbase:") {
			t.Errorf("Generic error should contain 'pocketbase:' prefix, got: %v", genericErr)
		}
		if !strings.Contains(legacyErr.Error(), "pocketbase:") {
			t.Errorf("Legacy error should contain 'pocketbase:' prefix, got: %v", legacyErr)
		}
	})
}

// TestRecordService_TypeConversionFailures는 타입 변환 실패 시나리오를 테스트합니다.
func TestRecordService_TypeConversionFailures(t *testing.T) {
	// 잘못된 JSON 응답을 반환하는 테스트 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/collections/posts/records":
			if r.Method == http.MethodGet {
				// 잘못된 JSON 구조 (items가 배열이 아님)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"page": 1,
					"perPage": 20,
					"totalItems": 1,
					"totalPages": 1,
					"items": "not_an_array"
				}`))
			} else if r.Method == http.MethodPost {
				// 타입 불일치 응답 (숫자 필드에 문자열)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "test",
					"collectionId": "posts_id",
					"collectionName": "posts",
					"title": "Test",
					"view_count": "not_a_number"
				}`))
			}
		case "/api/collections/posts/records/malformed":
			if r.Method == http.MethodGet {
				// 완전히 잘못된 JSON
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"invalid": json syntax`))
			}
		case "/api/collections/posts/records/type_mismatch":
			if r.Method == http.MethodGet {
				// 타입 불일치 (예상된 구조와 다른 응답)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": 123,
					"collectionId": ["array", "instead", "of", "string"],
					"collectionName": null,
					"title": {"nested": "object"},
					"content": true
				}`))
			}
		case "/api/collections/posts/records/missing_fields":
			if r.Method == http.MethodGet {
				// 필수 필드 누락
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"title": "Test Title"
				}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	type Post struct {
		BaseModel
		Title     string `json:"title"`
		Content   string `json:"content"`
		ViewCount int    `json:"view_count"`
	}

	postService := NewRecordService[Post](client, "posts")

	// 1. GetList 타입 변환 실패 테스트
	t.Run("GetList_TypeConversionFailure", func(t *testing.T) {
		_, err := postService.GetList(ctx, nil)
		if err == nil {
			t.Fatal("Expected error for malformed JSON response")
		}

		// JSON 파싱 에러가 포함되어야 함
		if !strings.Contains(err.Error(), "pocketbase: fetch typed records list:") {
			t.Errorf("Expected error message to contain 'pocketbase: fetch typed records list:', got: %v", err)
		}
	})

	// 2. GetOne 완전히 잘못된 JSON 테스트
	t.Run("GetOne_MalformedJSON", func(t *testing.T) {
		_, err := postService.GetOne(ctx, "malformed", nil)
		if err == nil {
			t.Fatal("Expected error for malformed JSON response")
		}

		// JSON 파싱 에러가 포함되어야 함
		if !strings.Contains(err.Error(), "pocketbase: fetch typed record:") {
			t.Errorf("Expected error message to contain 'pocketbase: fetch typed record:', got: %v", err)
		}
	})

	// 3. GetOne 타입 불일치 테스트 (Go의 JSON 파싱은 관대하므로 이 경우는 성공할 수 있음)
	t.Run("GetOne_TypeMismatch", func(t *testing.T) {
		result, err := postService.GetOne(ctx, "type_mismatch", nil)
		// Go의 JSON 파싱은 타입 불일치를 어느 정도 허용하므로 에러가 발생하지 않을 수 있음
		// 하지만 결과가 예상과 다를 수 있음
		if err != nil {
			// 에러가 발생한 경우, 적절한 에러 메시지인지 확인
			if !strings.Contains(err.Error(), "pocketbase: fetch typed record:") {
				t.Errorf("Expected error message to contain 'pocketbase: fetch typed record:', got: %v", err)
			}
		} else {
			// 성공한 경우, 타입 변환이 어떻게 처리되었는지 확인
			if result == nil {
				t.Error("Expected non-nil result even with type mismatch")
			}
			// ID가 숫자에서 문자열로 변환되었는지 확인
			// (Go JSON 파싱은 숫자를 문자열로 변환할 수 있음)
		}
	})

	// 4. GetOne 필수 필드 누락 테스트
	t.Run("GetOne_MissingFields", func(t *testing.T) {
		result, err := postService.GetOne(ctx, "missing_fields", nil)
		if err != nil {
			// 에러가 발생한 경우
			if !strings.Contains(err.Error(), "pocketbase: fetch typed record:") {
				t.Errorf("Expected error message to contain 'pocketbase: fetch typed record:', got: %v", err)
			}
		} else {
			// 성공한 경우, 누락된 필드가 기본값으로 설정되었는지 확인
			if result == nil {
				t.Error("Expected non-nil result")
			}
			if result.ID != "" {
				t.Error("Expected empty ID for missing field")
			}
			if result.CollectionName != "" {
				t.Error("Expected empty CollectionName for missing field")
			}
			if result.Title != "Test Title" {
				t.Errorf("Expected title 'Test Title', got '%s'", result.Title)
			}
		}
	})

	// 5. Create 타입 변환 실패 테스트
	t.Run("Create_TypeConversionFailure", func(t *testing.T) {
		newPost := &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
			Title:     "Test Post",
			Content:   "Test content",
			ViewCount: 100,
		}

		result, err := postService.Create(ctx, newPost, nil)
		if err != nil {
			// 에러가 발생한 경우
			if !strings.Contains(err.Error(), "pocketbase: create typed record:") {
				t.Errorf("Expected error message to contain 'pocketbase: create typed record:', got: %v", err)
			}
		} else {
			// 성공한 경우, 타입 변환이 어떻게 처리되었는지 확인
			if result == nil {
				t.Error("Expected non-nil result")
			}
			// ViewCount가 문자열에서 숫자로 변환되지 않았을 것임
			// (Go JSON 파싱은 문자열을 숫자로 자동 변환하지 않음)
			if result.ViewCount != 0 {
				t.Errorf("Expected ViewCount 0 for string-to-int conversion failure, got %d", result.ViewCount)
			}
		}
	})
}

// TestRecordService_NetworkErrors는 네트워크 관련 에러를 테스트합니다.
func TestRecordService_NetworkErrors(t *testing.T) {
	// 존재하지 않는 서버 URL로 클라이언트 생성
	client := NewClient("http://nonexistent-server:9999")

	type Post struct {
		BaseModel
		Title string `json:"title"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 1. 네트워크 연결 실패 테스트
	t.Run("NetworkConnectionFailure", func(t *testing.T) {
		_, err := postService.GetList(ctx, nil)
		if err == nil {
			t.Fatal("Expected network error, got nil")
		}

		// 네트워크 에러가 적절히 래핑되었는지 확인
		if !strings.Contains(err.Error(), "pocketbase: fetch typed records list:") {
			t.Errorf("Expected error message to contain 'pocketbase: fetch typed records list:', got: %v", err)
		}
	})

	// 2. 컨텍스트 취소 테스트
	t.Run("ContextCancellation", func(t *testing.T) {
		// 즉시 취소되는 컨텍스트 생성
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel() // 즉시 취소

		_, err := postService.GetList(cancelCtx, nil)
		if err == nil {
			t.Fatal("Expected context cancellation error, got nil")
		}

		// 컨텍스트 취소 에러가 적절히 처리되었는지 확인
		if !strings.Contains(err.Error(), "pocketbase: fetch typed records list:") {
			t.Errorf("Expected error message to contain 'pocketbase: fetch typed records list:', got: %v", err)
		}
	})
}

// 5.3 QueryParam 호환성 테스트

// TestRecordService_QueryParamCompatibility는 QueryParam 시스템 호환성을 테스트합니다.
func TestRecordService_QueryParamCompatibility(t *testing.T) {
	// 테스트용 HTTP 서버 설정
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 요청 URL과 쿼리 파라미터를 검증
		switch r.URL.Path {
		case "/api/collections/posts/records":
			// GetList 요청 처리
			if r.Method == http.MethodGet {
				// 쿼리 파라미터 검증
				query := r.URL.Query()

				// ListOptions 파라미터들이 올바르게 적용되었는지 확인
				if page := query.Get("page"); page != "2" {
					t.Errorf("Expected page=2, got page=%s", page)
				}
				if perPage := query.Get("perPage"); perPage != "10" {
					t.Errorf("Expected perPage=10, got perPage=%s", perPage)
				}
				if sort := query.Get("sort"); sort != "-created" {
					t.Errorf("Expected sort=-created, got sort=%s", sort)
				}
				if filter := query.Get("filter"); filter != "status='published'" {
					t.Errorf("Expected filter=status='published', got filter=%s", filter)
				}
				if expand := query.Get("expand"); expand != "author" {
					t.Errorf("Expected expand=author, got expand=%s", expand)
				}
				if fields := query.Get("fields"); fields != "id,title,content" {
					t.Errorf("Expected fields=id,title,content, got fields=%s", fields)
				}
				if skipTotal := query.Get("skipTotal"); skipTotal != "1" {
					t.Errorf("Expected skipTotal=1, got skipTotal=%s", skipTotal)
				}

				// 응답 반환
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"page": 2,
					"perPage": 10,
					"totalItems": 100,
					"totalPages": 10,
					"items": [
						{
							"id": "test1",
							"collectionId": "posts",
							"collectionName": "posts",
							"title": "Test Post 1",
							"content": "Test content 1"
						}
					]
				}`))
			} else if r.Method == http.MethodPost {
				// Create 요청 처리
				query := r.URL.Query()
				if expand := query.Get("expand"); expand != "author" {
					t.Errorf("Expected expand=author in Create, got expand=%s", expand)
				}
				if fields := query.Get("fields"); fields != "id,title" {
					t.Errorf("Expected fields=id,title in Create, got fields=%s", fields)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "new_post",
					"collectionId": "posts",
					"collectionName": "posts",
					"title": "New Post",
					"content": "New content"
				}`))
			}
		case "/api/collections/posts/records/test1":
			if r.Method == http.MethodGet {
				// GetOne 요청 처리
				query := r.URL.Query()
				if expand := query.Get("expand"); expand != "author,comments" {
					t.Errorf("Expected expand=author,comments in GetOne, got expand=%s", expand)
				}
				if fields := query.Get("fields"); fields != "id,title" {
					t.Errorf("Expected fields=id,title in GetOne, got fields=%s", fields)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "test1",
					"collectionId": "posts",
					"collectionName": "posts",
					"title": "Test Post 1",
					"content": "Test content 1"
				}`))
			} else if r.Method == http.MethodPatch {
				// Update 요청 처리
				query := r.URL.Query()
				if expand := query.Get("expand"); expand != "author" {
					t.Errorf("Expected expand=author in Update, got expand=%s", expand)
				}
				if fields := query.Get("fields"); fields != "id,title,updated" {
					t.Errorf("Expected fields=id,title,updated in Update, got fields=%s", fields)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "test1",
					"collectionId": "posts",
					"collectionName": "posts",
					"title": "Updated Post",
					"content": "Updated content"
				}`))
			}
		}
	}))
	defer srv.Close()

	// 클라이언트 생성
	client := NewClient(srv.URL)

	// 사용자 정의 타입 정의
	type Post struct {
		BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	// 제네릭 서비스 생성
	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 1. GetList 메서드의 ListOptions 호환성 테스트
	t.Run("GetList_ListOptions", func(t *testing.T) {
		opts := &ListOptions{
			Page:      2,
			PerPage:   10,
			Sort:      "-created",
			Filter:    "status='published'",
			Expand:    "author",
			Fields:    "id,title,content",
			SkipTotal: true,
			QueryParams: map[string]string{
				"custom": "value",
			},
		}

		result, err := postService.GetList(ctx, opts)
		if err != nil {
			t.Fatalf("GetList failed: %v", err)
		}

		if result.Page != 2 {
			t.Errorf("Expected page 2, got %d", result.Page)
		}
		if result.PerPage != 10 {
			t.Errorf("Expected perPage 10, got %d", result.PerPage)
		}
		if len(result.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(result.Items))
		}
		if result.Items[0].Title != "Test Post 1" {
			t.Errorf("Expected title 'Test Post 1', got '%s'", result.Items[0].Title)
		}
	})

	// 2. GetOne 메서드의 GetOneOptions 호환성 테스트
	t.Run("GetOne_GetOneOptions", func(t *testing.T) {
		opts := &GetOneOptions{
			Expand: "author,comments",
			Fields: "id,title",
		}

		result, err := postService.GetOne(ctx, "test1", opts)
		if err != nil {
			t.Fatalf("GetOne failed: %v", err)
		}

		if result.ID != "test1" {
			t.Errorf("Expected ID 'test1', got '%s'", result.ID)
		}
		if result.Title != "Test Post 1" {
			t.Errorf("Expected title 'Test Post 1', got '%s'", result.Title)
		}
	})

	// 3. Create 메서드의 WriteOptions 호환성 테스트
	t.Run("Create_WriteOptions", func(t *testing.T) {
		newPost := &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
			Title:   "New Post",
			Content: "New content",
		}

		opts := &WriteOptions{
			Expand: "author",
			Fields: "id,title",
		}

		result, err := postService.Create(ctx, newPost, opts)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if result.ID != "new_post" {
			t.Errorf("Expected ID 'new_post', got '%s'", result.ID)
		}
		if result.Title != "New Post" {
			t.Errorf("Expected title 'New Post', got '%s'", result.Title)
		}
	})

	// 4. Update 메서드의 WriteOptions 호환성 테스트
	t.Run("Update_WriteOptions", func(t *testing.T) {
		updatePost := &Post{
			BaseModel: BaseModel{
				ID:             "test1",
				CollectionName: "posts",
			},
			Title:   "Updated Post",
			Content: "Updated content",
		}

		opts := &WriteOptions{
			Expand: "author",
			Fields: "id,title,updated",
		}

		result, err := postService.Update(ctx, "test1", updatePost, opts)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		if result.ID != "test1" {
			t.Errorf("Expected ID 'test1', got '%s'", result.ID)
		}
		if result.Title != "Updated Post" {
			t.Errorf("Expected title 'Updated Post', got '%s'", result.Title)
		}
	})
}

// TestRecordService_ComplexFilteringAndSorting은 복잡한 필터링과 정렬 기능을 테스트합니다.
func TestRecordService_ComplexFilteringAndSorting(t *testing.T) {
	// 복잡한 필터링과 정렬 조건을 테스트하는 HTTP 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/collections/posts/records" && r.Method == http.MethodGet {
			query := r.URL.Query()

			// 복잡한 필터 조건 검증
			expectedFilter := "status='published' && created >= '2023-01-01' && (category='tech' || category='news')"
			if filter := query.Get("filter"); filter != expectedFilter {
				t.Errorf("Expected complex filter, got: %s", filter)
			}

			// 다중 정렬 조건 검증
			expectedSort := "-created,+title"
			if sort := query.Get("sort"); sort != expectedSort {
				t.Errorf("Expected sort '%s', got: %s", expectedSort, sort)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"perPage": 20,
				"totalItems": 5,
				"totalPages": 1,
				"items": [
					{
						"id": "post1",
						"collectionId": "posts",
						"collectionName": "posts",
						"title": "Tech Post",
						"category": "tech",
						"status": "published"
					},
					{
						"id": "post2",
						"collectionId": "posts",
						"collectionName": "posts",
						"title": "News Post",
						"category": "news",
						"status": "published"
					}
				]
			}`))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title    string `json:"title"`
		Category string `json:"category"`
		Status   string `json:"status"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 복잡한 필터링과 정렬 옵션 테스트
	opts := &ListOptions{
		Filter: "status='published' && created >= '2023-01-01' && (category='tech' || category='news')",
		Sort:   "-created,+title",
		Expand: "author,tags",
		Fields: "id,title,category,status",
	}

	result, err := postService.GetList(ctx, opts)
	if err != nil {
		t.Fatalf("GetList with complex filtering failed: %v", err)
	}

	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}

	// 결과 검증
	if result.Items[0].Category != "tech" {
		t.Errorf("Expected first item category 'tech', got '%s'", result.Items[0].Category)
	}
	if result.Items[1].Category != "news" {
		t.Errorf("Expected second item category 'news', got '%s'", result.Items[1].Category)
	}
}

// TestRecordService_PaginationCompatibility는 페이지네이션 호환성을 테스트합니다.
func TestRecordService_PaginationCompatibility(t *testing.T) {
	// 페이지네이션 테스트용 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/collections/posts/records" && r.Method == http.MethodGet {
			query := r.URL.Query()

			page := query.Get("page")
			perPage := query.Get("perPage")
			skipTotal := query.Get("skipTotal")

			// 페이지네이션 파라미터 검증
			if page != "3" {
				t.Errorf("Expected page=3, got page=%s", page)
			}
			if perPage != "5" {
				t.Errorf("Expected perPage=5, got perPage=%s", perPage)
			}
			if skipTotal != "1" {
				t.Errorf("Expected skipTotal=1, got skipTotal=%s", skipTotal)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 3,
				"perPage": 5,
				"totalItems": -1,
				"totalPages": -1,
				"items": [
					{
						"id": "post11",
						"collectionId": "posts",
						"collectionName": "posts",
						"title": "Post 11"
					},
					{
						"id": "post12",
						"collectionId": "posts",
						"collectionName": "posts",
						"title": "Post 12"
					}
				]
			}`))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title string `json:"title"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 페이지네이션 옵션 테스트
	opts := &ListOptions{
		Page:      3,
		PerPage:   5,
		SkipTotal: true,
	}

	result, err := postService.GetList(ctx, opts)
	if err != nil {
		t.Fatalf("GetList with pagination failed: %v", err)
	}

	// 페이지네이션 결과 검증
	if result.Page != 3 {
		t.Errorf("Expected page 3, got %d", result.Page)
	}
	if result.PerPage != 5 {
		t.Errorf("Expected perPage 5, got %d", result.PerPage)
	}
	if result.TotalItems != -1 {
		t.Errorf("Expected totalItems -1 (skipped), got %d", result.TotalItems)
	}
	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}
}

// TestRecordService_CustomQueryParams는 사용자 정의 쿼리 파라미터를 테스트합니다.
func TestRecordService_CustomQueryParams(t *testing.T) {
	// 사용자 정의 쿼리 파라미터 테스트용 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/collections/posts/records" && r.Method == http.MethodGet {
			query := r.URL.Query()

			// 사용자 정의 파라미터 검증
			if customParam := query.Get("customParam"); customParam != "customValue" {
				t.Errorf("Expected customParam=customValue, got customParam=%s", customParam)
			}
			if apiKey := query.Get("apiKey"); apiKey != "secret123" {
				t.Errorf("Expected apiKey=secret123, got apiKey=%s", apiKey)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"perPage": 20,
				"totalItems": 1,
				"totalPages": 1,
				"items": [
					{
						"id": "custom_post",
						"collectionId": "posts",
						"collectionName": "posts",
						"title": "Custom Post"
					}
				]
			}`))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title string `json:"title"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 사용자 정의 쿼리 파라미터 테스트
	opts := &ListOptions{
		QueryParams: map[string]string{
			"customParam": "customValue",
			"apiKey":      "secret123",
		},
	}

	result, err := postService.GetList(ctx, opts)
	if err != nil {
		t.Fatalf("GetList with custom query params failed: %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Title != "Custom Post" {
		t.Errorf("Expected title 'Custom Post', got '%s'", result.Items[0].Title)
	}
}

// TestRecordService_LegacyFunctionCompatibility는 기존 독립 함수들과의 호환성을 테스트합니다.
func TestRecordService_LegacyFunctionCompatibility(t *testing.T) {
	// 기존 함수와 제네릭 메서드의 결과가 동일한지 확인하는 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/collections/posts/records" && r.Method == http.MethodGet:
			// GetList/GetListAs 호환성 테스트
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"page": 1,
				"perPage": 20,
				"totalItems": 2,
				"totalPages": 1,
				"items": [
					{
						"id": "post1",
						"collectionId": "posts",
						"collectionName": "posts",
						"title": "Post 1",
						"content": "Content 1"
					},
					{
						"id": "post2",
						"collectionId": "posts",
						"collectionName": "posts",
						"title": "Post 2",
						"content": "Content 2"
					}
				]
			}`))
		case r.URL.Path == "/api/collections/posts/records/post1" && r.Method == http.MethodGet:
			// GetOne/GetOneAs 호환성 테스트
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "post1",
				"collectionId": "posts",
				"collectionName": "posts",
				"title": "Post 1",
				"content": "Content 1"
			}`))
		case r.URL.Path == "/api/collections/posts/records" && r.Method == http.MethodPost:
			// Create/CreateAs 호환성 테스트
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "new_post",
				"collectionId": "posts",
				"collectionName": "posts",
				"title": "New Post",
				"content": "New Content"
			}`))
		case r.URL.Path == "/api/collections/posts/records/post1" && r.Method == http.MethodPatch:
			// Update/UpdateAs 호환성 테스트
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "post1",
				"collectionId": "posts",
				"collectionName": "posts",
				"title": "Updated Post",
				"content": "Updated Content"
			}`))
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	// 제네릭 서비스와 레거시 서비스 생성
	postService := NewRecordService[Post](client, "posts")
	legacyService := &RecordServiceLegacy{Client: client}
	ctx := context.Background()

	// 1. GetList vs GetListAs 호환성 테스트
	t.Run("GetList_vs_GetListAs", func(t *testing.T) {
		opts := &ListOptions{
			Page:    1,
			PerPage: 20,
			Sort:    "-created",
			Filter:  "status='published'",
		}

		// 제네릭 메서드 호출
		genericResult, err := postService.GetList(ctx, opts)
		if err != nil {
			t.Fatalf("Generic GetList failed: %v", err)
		}

		// 기존 독립 함수 호출
		legacyResult, err := GetListAs[Post](ctx, client, "posts", opts)
		if err != nil {
			t.Fatalf("Legacy GetListAs failed: %v", err)
		}

		// 결과 비교
		if genericResult.Page != legacyResult.Page {
			t.Errorf("Page mismatch: generic=%d, legacy=%d", genericResult.Page, legacyResult.Page)
		}
		if genericResult.TotalItems != legacyResult.TotalItems {
			t.Errorf("TotalItems mismatch: generic=%d, legacy=%d", genericResult.TotalItems, legacyResult.TotalItems)
		}
		if len(genericResult.Items) != len(legacyResult.Items) {
			t.Errorf("Items length mismatch: generic=%d, legacy=%d", len(genericResult.Items), len(legacyResult.Items))
		}

		// 첫 번째 아이템 비교
		if len(genericResult.Items) > 0 && len(legacyResult.Items) > 0 {
			if genericResult.Items[0].ID != legacyResult.Items[0].ID {
				t.Errorf("First item ID mismatch: generic=%s, legacy=%s",
					genericResult.Items[0].ID, legacyResult.Items[0].ID)
			}
			if genericResult.Items[0].Title != legacyResult.Items[0].Title {
				t.Errorf("First item Title mismatch: generic=%s, legacy=%s",
					genericResult.Items[0].Title, legacyResult.Items[0].Title)
			}
		}
	})

	// 2. GetOne vs GetOneAs 호환성 테스트
	t.Run("GetOne_vs_GetOneAs", func(t *testing.T) {
		opts := &GetOneOptions{
			Expand: "author",
			Fields: "id,title,content",
		}

		// 제네릭 메서드 호출
		genericResult, err := postService.GetOne(ctx, "post1", opts)
		if err != nil {
			t.Fatalf("Generic GetOne failed: %v", err)
		}

		// 기존 독립 함수 호출
		legacyResult, err := GetOneAs[Post](ctx, client, "posts", "post1", opts)
		if err != nil {
			t.Fatalf("Legacy GetOneAs failed: %v", err)
		}

		// 결과 비교
		if genericResult.ID != legacyResult.ID {
			t.Errorf("ID mismatch: generic=%s, legacy=%s", genericResult.ID, legacyResult.ID)
		}
		if genericResult.Title != legacyResult.Title {
			t.Errorf("Title mismatch: generic=%s, legacy=%s", genericResult.Title, legacyResult.Title)
		}
		if genericResult.Content != legacyResult.Content {
			t.Errorf("Content mismatch: generic=%s, legacy=%s", genericResult.Content, legacyResult.Content)
		}
	})

	// 3. Create vs CreateAs 호환성 테스트
	t.Run("Create_vs_CreateAs", func(t *testing.T) {
		newPost := &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
			Title:   "New Post",
			Content: "New Content",
		}

		opts := &WriteOptions{
			Expand: "author",
			Fields: "id,title,content",
		}

		// 제네릭 메서드 호출
		genericResult, err := postService.Create(ctx, newPost, opts)
		if err != nil {
			t.Fatalf("Generic Create failed: %v", err)
		}

		// 기존 독립 함수 호출 (새로운 인스턴스 생성)
		newPost2 := &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
			Title:   "New Post",
			Content: "New Content",
		}
		legacyResult, err := CreateAs[Post](ctx, legacyService, "posts", newPost2, opts)
		if err != nil {
			t.Fatalf("Legacy CreateAs failed: %v", err)
		}

		// 결과 비교
		if genericResult.ID != legacyResult.ID {
			t.Errorf("ID mismatch: generic=%s, legacy=%s", genericResult.ID, legacyResult.ID)
		}
		if genericResult.Title != legacyResult.Title {
			t.Errorf("Title mismatch: generic=%s, legacy=%s", genericResult.Title, legacyResult.Title)
		}
		if genericResult.Content != legacyResult.Content {
			t.Errorf("Content mismatch: generic=%s, legacy=%s", genericResult.Content, legacyResult.Content)
		}
	})

	// 4. Update vs UpdateAs 호환성 테스트
	t.Run("Update_vs_UpdateAs", func(t *testing.T) {
		updatePost := &Post{
			BaseModel: BaseModel{
				ID:             "post1",
				CollectionName: "posts",
			},
			Title:   "Updated Post",
			Content: "Updated Content",
		}

		opts := &WriteOptions{
			Expand: "author",
			Fields: "id,title,content",
		}

		// 제네릭 메서드 호출
		genericResult, err := postService.Update(ctx, "post1", updatePost, opts)
		if err != nil {
			t.Fatalf("Generic Update failed: %v", err)
		}

		// 기존 독립 함수 호출 (새로운 인스턴스 생성)
		updatePost2 := &Post{
			BaseModel: BaseModel{
				ID:             "post1",
				CollectionName: "posts",
			},
			Title:   "Updated Post",
			Content: "Updated Content",
		}
		legacyResult, err := UpdateAs[Post](ctx, legacyService, "posts", "post1", updatePost2, opts)
		if err != nil {
			t.Fatalf("Legacy UpdateAs failed: %v", err)
		}

		// 결과 비교
		if genericResult.ID != legacyResult.ID {
			t.Errorf("ID mismatch: generic=%s, legacy=%s", genericResult.ID, legacyResult.ID)
		}
		if genericResult.Title != legacyResult.Title {
			t.Errorf("Title mismatch: generic=%s, legacy=%s", genericResult.Title, legacyResult.Title)
		}
		if genericResult.Content != legacyResult.Content {
			t.Errorf("Content mismatch: generic=%s, legacy=%s", genericResult.Content, legacyResult.Content)
		}
	})
}

// TestRecordService_QueryParamEdgeCases는 QueryParam의 엣지 케이스를 테스트합니다.
func TestRecordService_QueryParamEdgeCases(t *testing.T) {
	// 엣지 케이스 테스트용 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		// URL 인코딩된 특수 문자 검증
		if filter := query.Get("filter"); filter != "" {
			// 특수 문자가 올바르게 인코딩되었는지 확인
			if !strings.Contains(filter, "title") {
				t.Errorf("Filter should contain 'title', got: %s", filter)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"page": 1,
			"perPage": 20,
			"totalItems": 0,
			"totalPages": 0,
			"items": []
		}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title string `json:"title"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 1. nil 옵션 테스트
	t.Run("NilOptions", func(t *testing.T) {
		_, err := postService.GetList(ctx, nil)
		if err != nil {
			t.Fatalf("GetList with nil options should work: %v", err)
		}

		_, err = postService.GetOne(ctx, "test", nil)
		if err != nil {
			t.Fatalf("GetOne with nil options should work: %v", err)
		}

		testPost := &Post{Title: "Test"}
		_, err = postService.Create(ctx, testPost, nil)
		if err != nil {
			t.Fatalf("Create with nil options should work: %v", err)
		}

		_, err = postService.Update(ctx, "test", testPost, nil)
		if err != nil {
			t.Fatalf("Update with nil options should work: %v", err)
		}
	})

	// 2. 빈 옵션 테스트
	t.Run("EmptyOptions", func(t *testing.T) {
		emptyListOpts := &ListOptions{}
		_, err := postService.GetList(ctx, emptyListOpts)
		if err != nil {
			t.Fatalf("GetList with empty options should work: %v", err)
		}

		emptyGetOneOpts := &GetOneOptions{}
		_, err = postService.GetOne(ctx, "test", emptyGetOneOpts)
		if err != nil {
			t.Fatalf("GetOne with empty options should work: %v", err)
		}

		emptyWriteOpts := &WriteOptions{}
		testPost := &Post{Title: "Test"}
		_, err = postService.Create(ctx, testPost, emptyWriteOpts)
		if err != nil {
			t.Fatalf("Create with empty options should work: %v", err)
		}
	})

	// 3. 특수 문자가 포함된 필터 테스트
	t.Run("SpecialCharactersInFilter", func(t *testing.T) {
		opts := &ListOptions{
			Filter: "title~'test & example' && content!='<script>alert(\"xss\")</script>'",
			Sort:   "-created,+title",
		}

		_, err := postService.GetList(ctx, opts)
		if err != nil {
			t.Fatalf("GetList with special characters in filter should work: %v", err)
		}
	})

	// 4. 매우 긴 쿼리 파라미터 테스트
	t.Run("LongQueryParameters", func(t *testing.T) {
		longFilter := strings.Repeat("title='test' || ", 100) + "title='final'"
		longFields := strings.Repeat("field", 50)

		opts := &ListOptions{
			Filter: longFilter,
			Fields: longFields,
		}

		_, err := postService.GetList(ctx, opts)
		if err != nil {
			t.Fatalf("GetList with long query parameters should work: %v", err)
		}
	})
}

// TestRecordService_AdvancedQueryParamCompatibility는 고급 QueryParam 호환성을 테스트합니다.
func TestRecordService_AdvancedQueryParamCompatibility(t *testing.T) {
	// 고급 쿼리 파라미터 테스트용 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		query := r.URL.Query()

		// 복잡한 쿼리 파라미터 검증
		switch r.URL.Path {
		case "/api/collections/posts/records":
			if r.Method == http.MethodGet {
				// 모든 쿼리 파라미터가 올바르게 전달되었는지 확인
				expectedParams := map[string]string{
					"page":      "2",
					"perPage":   "15",
					"sort":      "-created,+title,-updated",
					"filter":    "status='published' && (category='tech' || category='science') && created >= '2023-01-01'",
					"expand":    "author,category,tags",
					"fields":    "id,title,content,status,created,updated",
					"skipTotal": "1",
					"custom1":   "value1",
					"custom2":   "value2",
				}

				for key, expected := range expectedParams {
					if actual := query.Get(key); actual != expected {
						t.Errorf("Expected %s=%s, got %s=%s", key, expected, key, actual)
					}
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"page": 2,
					"perPage": 15,
					"totalItems": -1,
					"totalPages": -1,
					"items": [
						{
							"id": "post1",
							"collectionId": "posts_id",
							"collectionName": "posts",
							"title": "Advanced Post 1",
							"content": "Advanced content 1",
							"status": "published",
							"category": "tech"
						},
						{
							"id": "post2",
							"collectionId": "posts_id",
							"collectionName": "posts",
							"title": "Advanced Post 2",
							"content": "Advanced content 2",
							"status": "published",
							"category": "science"
						}
					]
				}`))
			} else if r.Method == http.MethodPost {
				// Create 옵션 검증
				if expand := query.Get("expand"); expand != "author,category" {
					t.Errorf("Expected expand=author,category, got expand=%s", expand)
				}
				if fields := query.Get("fields"); fields != "id,title,content,created" {
					t.Errorf("Expected fields=id,title,content,created, got fields=%s", fields)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "new_post",
					"collectionId": "posts_id",
					"collectionName": "posts",
					"title": "New Advanced Post",
					"content": "New advanced content",
					"status": "draft"
				}`))
			}
		case "/api/collections/posts/records/post1":
			if r.Method == http.MethodGet {
				// GetOne 옵션 검증
				if expand := query.Get("expand"); expand != "author,category,tags,comments" {
					t.Errorf("Expected expand=author,category,tags,comments, got expand=%s", expand)
				}
				if fields := query.Get("fields"); fields != "id,title,content,author,category" {
					t.Errorf("Expected fields=id,title,content,author,category, got fields=%s", fields)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "post1",
					"collectionId": "posts_id",
					"collectionName": "posts",
					"title": "Advanced Post 1",
					"content": "Advanced content 1",
					"status": "published"
				}`))
			} else if r.Method == http.MethodPatch {
				// Update 옵션 검증
				if expand := query.Get("expand"); expand != "author,category" {
					t.Errorf("Expected expand=author,category, got expand=%s", expand)
				}
				if fields := query.Get("fields"); fields != "id,title,content,updated" {
					t.Errorf("Expected fields=id,title,content,updated, got fields=%s", fields)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"id": "post1",
					"collectionId": "posts_id",
					"collectionName": "posts",
					"title": "Updated Advanced Post",
					"content": "Updated advanced content",
					"status": "published"
				}`))
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title    string `json:"title"`
		Content  string `json:"content"`
		Status   string `json:"status"`
		Category string `json:"category"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 1. 복잡한 ListOptions 테스트
	t.Run("ComplexListOptions", func(t *testing.T) {
		opts := &ListOptions{
			Page:      2,
			PerPage:   15,
			Sort:      "-created,+title,-updated",
			Filter:    "status='published' && (category='tech' || category='science') && created >= '2023-01-01'",
			Expand:    "author,category,tags",
			Fields:    "id,title,content,status,created,updated",
			SkipTotal: true,
			QueryParams: map[string]string{
				"custom1": "value1",
				"custom2": "value2",
			},
		}

		result, err := postService.GetList(ctx, opts)
		if err != nil {
			t.Fatalf("GetList with complex options failed: %v", err)
		}

		if result.Page != 2 {
			t.Errorf("Expected page 2, got %d", result.Page)
		}
		if result.PerPage != 15 {
			t.Errorf("Expected perPage 15, got %d", result.PerPage)
		}
		if result.TotalItems != -1 {
			t.Errorf("Expected totalItems -1 (skipped), got %d", result.TotalItems)
		}
		if len(result.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(result.Items))
		}

		// 첫 번째 아이템 검증
		if result.Items[0].Title != "Advanced Post 1" {
			t.Errorf("Expected first item title 'Advanced Post 1', got '%s'", result.Items[0].Title)
		}
		if result.Items[0].Category != "tech" {
			t.Errorf("Expected first item category 'tech', got '%s'", result.Items[0].Category)
		}
	})

	// 2. 복잡한 GetOneOptions 테스트
	t.Run("ComplexGetOneOptions", func(t *testing.T) {
		opts := &GetOneOptions{
			Expand: "author,category,tags,comments",
			Fields: "id,title,content,author,category",
		}

		result, err := postService.GetOne(ctx, "post1", opts)
		if err != nil {
			t.Fatalf("GetOne with complex options failed: %v", err)
		}

		if result.ID != "post1" {
			t.Errorf("Expected ID 'post1', got '%s'", result.ID)
		}
		if result.Title != "Advanced Post 1" {
			t.Errorf("Expected title 'Advanced Post 1', got '%s'", result.Title)
		}
	})

	// 3. 복잡한 Create WriteOptions 테스트
	t.Run("ComplexCreateOptions", func(t *testing.T) {
		newPost := &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
			Title:    "New Advanced Post",
			Content:  "New advanced content",
			Status:   "draft",
			Category: "tech",
		}

		opts := &WriteOptions{
			Expand: "author,category",
			Fields: "id,title,content,created",
		}

		result, err := postService.Create(ctx, newPost, opts)
		if err != nil {
			t.Fatalf("Create with complex options failed: %v", err)
		}

		if result.ID != "new_post" {
			t.Errorf("Expected ID 'new_post', got '%s'", result.ID)
		}
		if result.Title != "New Advanced Post" {
			t.Errorf("Expected title 'New Advanced Post', got '%s'", result.Title)
		}
	})

	// 4. 복잡한 Update WriteOptions 테스트
	t.Run("ComplexUpdateOptions", func(t *testing.T) {
		updatePost := &Post{
			BaseModel: BaseModel{
				ID:             "post1",
				CollectionName: "posts",
			},
			Title:    "Updated Advanced Post",
			Content:  "Updated advanced content",
			Status:   "published",
			Category: "tech",
		}

		opts := &WriteOptions{
			Expand: "author,category",
			Fields: "id,title,content,updated",
		}

		result, err := postService.Update(ctx, "post1", updatePost, opts)
		if err != nil {
			t.Fatalf("Update with complex options failed: %v", err)
		}

		if result.ID != "post1" {
			t.Errorf("Expected ID 'post1', got '%s'", result.ID)
		}
		if result.Title != "Updated Advanced Post" {
			t.Errorf("Expected title 'Updated Advanced Post', got '%s'", result.Title)
		}
	})
}

// TestRecordService_QueryParamSystemCompatibility는 기존 QueryParam 시스템과의 완전한 호환성을 확인합니다.
func TestRecordService_QueryParamSystemCompatibility(t *testing.T) {
	// 동일한 쿼리 파라미터로 제네릭 서비스와 레거시 서비스를 비교하는 서버
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// 요청된 쿼리 파라미터를 응답에 포함하여 검증 가능하게 함
		query := r.URL.Query()

		switch r.URL.Path {
		case "/api/collections/posts/records":
			if r.Method == http.MethodGet {
				// 쿼리 파라미터를 응답에 포함
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{
					"page": %s,
					"perPage": %s,
					"totalItems": 10,
					"totalPages": 2,
					"items": [
						{
							"id": "post1",
							"collectionId": "posts_id",
							"collectionName": "posts",
							"title": "Test Post",
							"content": "Test content",
							"query_page": "%s",
							"query_sort": "%s",
							"query_filter": "%s"
						}
					]
				}`,
					query.Get("page"),
					query.Get("perPage"),
					query.Get("page"),
					query.Get("sort"),
					query.Get("filter"))))
			}
		case "/api/collections/posts/records/post1":
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(fmt.Sprintf(`{
					"id": "post1",
					"collectionId": "posts_id",
					"collectionName": "posts",
					"title": "Test Post",
					"content": "Test content",
					"query_expand": "%s",
					"query_fields": "%s"
				}`, query.Get("expand"), query.Get("fields"))))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()

	type Post struct {
		BaseModel
		Title       string `json:"title"`
		Content     string `json:"content"`
		QueryPage   string `json:"query_page"`
		QuerySort   string `json:"query_sort"`
		QueryFilter string `json:"query_filter"`
		QueryExpand string `json:"query_expand"`
		QueryFields string `json:"query_fields"`
	}

	// 제네릭 서비스 생성
	postService := NewRecordService[Post](client, "posts")

	// 1. GetList 쿼리 파라미터 호환성 테스트
	t.Run("GetList_QueryParamCompatibility", func(t *testing.T) {
		opts := &ListOptions{
			Page:    2,
			PerPage: 5,
			Sort:    "-created,+title",
			Filter:  "status='published' && category='tech'",
			Expand:  "author,tags",
			Fields:  "id,title,content",
		}

		// 제네릭 서비스 호출
		genericResult, err := postService.GetList(ctx, opts)
		if err != nil {
			t.Fatalf("Generic GetList failed: %v", err)
		}

		// 레거시 서비스 호출 (GetListAs 함수 사용)
		legacyResult, err := GetListAs[Post](ctx, client, "posts", opts)
		if err != nil {
			t.Fatalf("Legacy GetListAs failed: %v", err)
		}

		// 결과 비교
		if genericResult.Page != legacyResult.Page {
			t.Errorf("Page mismatch: generic=%d, legacy=%d", genericResult.Page, legacyResult.Page)
		}
		if genericResult.PerPage != legacyResult.PerPage {
			t.Errorf("PerPage mismatch: generic=%d, legacy=%d", genericResult.PerPage, legacyResult.PerPage)
		}
		if len(genericResult.Items) != len(legacyResult.Items) {
			t.Errorf("Items length mismatch: generic=%d, legacy=%d", len(genericResult.Items), len(legacyResult.Items))
		}

		// 첫 번째 아이템의 쿼리 파라미터 검증
		if len(genericResult.Items) > 0 && len(legacyResult.Items) > 0 {
			genericItem := genericResult.Items[0]
			legacyItem := legacyResult.Items[0]

			if genericItem.QueryPage != legacyItem.QueryPage {
				t.Errorf("QueryPage mismatch: generic=%s, legacy=%s", genericItem.QueryPage, legacyItem.QueryPage)
			}
			if genericItem.QuerySort != legacyItem.QuerySort {
				t.Errorf("QuerySort mismatch: generic=%s, legacy=%s", genericItem.QuerySort, legacyItem.QuerySort)
			}
			if genericItem.QueryFilter != legacyItem.QueryFilter {
				t.Errorf("QueryFilter mismatch: generic=%s, legacy=%s", genericItem.QueryFilter, legacyItem.QueryFilter)
			}
		}
	})

	// 2. GetOne 쿼리 파라미터 호환성 테스트
	t.Run("GetOne_QueryParamCompatibility", func(t *testing.T) {
		opts := &GetOneOptions{
			Expand: "author,category,tags",
			Fields: "id,title,content,author",
		}

		// 제네릭 서비스 호출
		genericResult, err := postService.GetOne(ctx, "post1", opts)
		if err != nil {
			t.Fatalf("Generic GetOne failed: %v", err)
		}

		// 레거시 서비스 호출 (GetOneAs 함수 사용)
		legacyResult, err := GetOneAs[Post](ctx, client, "posts", "post1", opts)
		if err != nil {
			t.Fatalf("Legacy GetOneAs failed: %v", err)
		}

		// 결과 비교
		if genericResult.ID != legacyResult.ID {
			t.Errorf("ID mismatch: generic=%s, legacy=%s", genericResult.ID, legacyResult.ID)
		}
		if genericResult.Title != legacyResult.Title {
			t.Errorf("Title mismatch: generic=%s, legacy=%s", genericResult.Title, legacyResult.Title)
		}
		if genericResult.QueryExpand != legacyResult.QueryExpand {
			t.Errorf("QueryExpand mismatch: generic=%s, legacy=%s", genericResult.QueryExpand, legacyResult.QueryExpand)
		}
		if genericResult.QueryFields != legacyResult.QueryFields {
			t.Errorf("QueryFields mismatch: generic=%s, legacy=%s", genericResult.QueryFields, legacyResult.QueryFields)
		}
	})

	// 3. 빈 옵션 호환성 테스트
	t.Run("EmptyOptions_Compatibility", func(t *testing.T) {
		// nil 옵션으로 테스트
		genericResult1, err1 := postService.GetList(ctx, nil)
		legacyResult1, err2 := GetListAs[Post](ctx, client, "posts", nil)

		if (err1 == nil) != (err2 == nil) {
			t.Errorf("Error consistency mismatch with nil options: generic_err=%v, legacy_err=%v", err1, err2)
		}

		if err1 == nil && err2 == nil {
			if genericResult1.Page != legacyResult1.Page {
				t.Errorf("Page mismatch with nil options: generic=%d, legacy=%d", genericResult1.Page, legacyResult1.Page)
			}
		}

		// 빈 옵션으로 테스트
		emptyOpts := &ListOptions{}
		genericResult2, err3 := postService.GetList(ctx, emptyOpts)
		legacyResult2, err4 := GetListAs[Post](ctx, client, "posts", emptyOpts)

		if (err3 == nil) != (err4 == nil) {
			t.Errorf("Error consistency mismatch with empty options: generic_err=%v, legacy_err=%v", err3, err4)
		}

		if err3 == nil && err4 == nil {
			if genericResult2.Page != legacyResult2.Page {
				t.Errorf("Page mismatch with empty options: generic=%d, legacy=%d", genericResult2.Page, legacyResult2.Page)
			}
		}
	})
}

// TestRecordService_QueryParamPerformance는 QueryParam 처리 성능을 테스트합니다.
func TestRecordService_QueryParamPerformance(t *testing.T) {
	// 성능 테스트용 서버 (빠른 응답)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"page": 1,
			"perPage": 20,
			"totalItems": 1000,
			"totalPages": 50,
			"items": [
				{
					"id": "post1",
					"collectionId": "posts_id",
					"collectionName": "posts",
					"title": "Performance Test Post",
					"content": "Performance test content"
				}
			]
		}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL)

	type Post struct {
		BaseModel
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	postService := NewRecordService[Post](client, "posts")
	ctx := context.Background()

	// 복잡한 쿼리 옵션 생성
	complexOpts := &ListOptions{
		Page:    1,
		PerPage: 20,
		Sort:    "-created,+title,-updated,+status",
		Filter:  "status='published' && (category='tech' || category='science' || category='news') && created >= '2023-01-01' && updated <= '2024-12-31'",
		Expand:  "author,category,tags,comments,likes,shares",
		Fields:  "id,title,content,status,created,updated,author,category,tags",
		QueryParams: map[string]string{
			"custom1": "value1",
			"custom2": "value2",
			"custom3": "value3",
			"custom4": "value4",
			"custom5": "value5",
		},
	}

	// 성능 테스트 (여러 번 호출하여 일관성 확인)
	t.Run("PerformanceConsistency", func(t *testing.T) {
		const iterations = 100

		for i := 0; i < iterations; i++ {
			result, err := postService.GetList(ctx, complexOpts)
			if err != nil {
				t.Fatalf("Iteration %d failed: %v", i, err)
			}

			if result.Page != 1 {
				t.Errorf("Iteration %d: Expected page 1, got %d", i, result.Page)
			}
			if len(result.Items) != 1 {
				t.Errorf("Iteration %d: Expected 1 item, got %d", i, len(result.Items))
			}
		}
	})

	// 메모리 사용량 테스트 (간단한 확인)
	t.Run("MemoryUsage", func(t *testing.T) {
		// 많은 수의 쿼리 파라미터로 테스트
		largeQueryParams := make(map[string]string)
		for i := 0; i < 100; i++ {
			largeQueryParams[fmt.Sprintf("param%d", i)] = fmt.Sprintf("value%d", i)
		}

		largeOpts := &ListOptions{
			Page:        1,
			PerPage:     50,
			Sort:        strings.Repeat("-field,", 20) + "-final",
			Filter:      strings.Repeat("field='value' && ", 50) + "final='value'",
			Expand:      strings.Repeat("relation,", 30) + "final",
			Fields:      strings.Repeat("field,", 50) + "final",
			QueryParams: largeQueryParams,
		}

		result, err := postService.GetList(ctx, largeOpts)
		if err != nil {
			t.Fatalf("Large options test failed: %v", err)
		}

		if result == nil {
			t.Error("Expected non-nil result for large options")
		}
	})
}
