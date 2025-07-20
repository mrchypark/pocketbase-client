package pocketbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestUsagePatterns는 두 가지 사용 방식을 검증합니다
func TestUsagePatterns(t *testing.T) {
	// 테스트용 서버 설정
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/collections/posts/records":
			if r.Method == "GET" {
				// 목록 조회 응답
				response := `{
					"page": 1,
					"perPage": 30,
					"totalItems": 2,
					"totalPages": 1,
					"items": [
						{
							"id": "post1",
							"collectionId": "posts_collection",
							"collectionName": "posts",
							"created": "2025-01-01T10:00:00.000Z",
							"updated": "2025-01-01T10:00:00.000Z",
							"title": "First Post",
							"content": "This is the first post",
							"published": true,
							"author": "user1",
							"tags": ["tech", "go"]
						},
						{
							"id": "post2",
							"collectionId": "posts_collection", 
							"collectionName": "posts",
							"created": "2025-01-01T11:00:00.000Z",
							"updated": "2025-01-01T11:00:00.000Z",
							"title": "Second Post",
							"content": "This is the second post",
							"published": false,
							"author": "user2",
							"tags": ["tutorial"]
						}
					]
				}`
				w.Write([]byte(response))
			} else if r.Method == "POST" {
				// 생성 응답
				response := `{
					"id": "post3",
					"collectionId": "posts_collection",
					"collectionName": "posts", 
					"created": "2025-01-01T12:00:00.000Z",
					"updated": "2025-01-01T12:00:00.000Z",
					"title": "New Post",
					"content": "This is a new post",
					"published": true,
					"author": "user3",
					"tags": ["new"]
				}`
				w.Write([]byte(response))
			}
		case "/api/collections/posts/records/post1":
			if r.Method == "GET" {
				// 단일 조회 응답
				response := `{
					"id": "post1",
					"collectionId": "posts_collection",
					"collectionName": "posts",
					"created": "2025-01-01T10:00:00.000Z",
					"updated": "2025-01-01T10:00:00.000Z",
					"title": "First Post",
					"content": "This is the first post",
					"published": true,
					"author": "user1",
					"tags": ["tech", "go"]
				}`
				w.Write([]byte(response))
			} else if r.Method == "PATCH" {
				// 업데이트 응답
				response := `{
					"id": "post1",
					"collectionId": "posts_collection",
					"collectionName": "posts",
					"created": "2025-01-01T10:00:00.000Z",
					"updated": "2025-01-01T12:30:00.000Z",
					"title": "Updated First Post",
					"content": "This is the updated first post",
					"published": true,
					"author": "user1",
					"tags": ["tech", "go", "updated"]
				}`
				w.Write([]byte(response))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx := context.Background()

	t.Run("동적 Record 사용 방식", func(t *testing.T) {
		testDynamicRecordUsage(t, client, ctx)
	})

	t.Run("타입 안전한 구조체 사용 방식", func(t *testing.T) {
		testTypeSafeStructUsage(t, client, ctx)
	})

	t.Run("두 방식 간 상호 변환", func(t *testing.T) {
		testInterconversion(t, client, ctx)
	})

	t.Run("성능 비교", func(t *testing.T) {
		testPerformanceComparison(t, client, ctx)
	})
}

// testDynamicRecordUsage는 동적 Record 사용 방식을 테스트합니다
func testDynamicRecordUsage(t *testing.T, client *Client, ctx context.Context) {
	t.Run("Record 생성 및 필드 설정", func(t *testing.T) {
		// 새 Record 생성
		record := &Record{}
		record.CollectionName = "posts"

		// 다양한 타입의 필드 설정
		record.Set("title", "Dynamic Record Post")
		record.Set("content", "Content created using dynamic record")
		record.Set("published", true)
		record.Set("author", "dynamic_user")
		record.Set("tags", []string{"dynamic", "test"})

		// 필드 값 검증
		if record.GetString("title") != "Dynamic Record Post" {
			t.Errorf("Title 필드 설정/조회 실패: %s", record.GetString("title"))
		}

		if record.GetString("content") != "Content created using dynamic record" {
			t.Errorf("Content 필드 설정/조회 실패: %s", record.GetString("content"))
		}

		if !record.GetBool("published") {
			t.Error("Published 필드 설정/조회 실패")
		}

		if record.GetString("author") != "dynamic_user" {
			t.Errorf("Author 필드 설정/조회 실패: %s", record.GetString("author"))
		}

		tags := record.GetStringSlice("tags")
		if len(tags) != 2 || tags[0] != "dynamic" || tags[1] != "test" {
			t.Errorf("Tags 필드 설정/조회 실패: %v", tags)
		}
	})

	t.Run("Record 목록 조회 및 동적 접근", func(t *testing.T) {
		result, err := client.Records.GetList(ctx, "posts", nil)
		if err != nil {
			t.Fatalf("Record 목록 조회 실패: %v", err)
		}

		if len(result.Items) != 2 {
			t.Fatalf("예상 Record 수: 2, 실제: %d", len(result.Items))
		}

		// 첫 번째 Record 검증
		firstRecord := result.Items[0]
		if firstRecord.GetString("title") != "First Post" {
			t.Errorf("첫 번째 Record title 불일치: %s", firstRecord.GetString("title"))
		}

		if firstRecord.GetString("content") != "This is the first post" {
			t.Errorf("첫 번째 Record content 불일치: %s", firstRecord.GetString("content"))
		}

		if !firstRecord.GetBool("published") {
			t.Error("첫 번째 Record published 상태 불일치")
		}

		// 두 번째 Record 검증
		secondRecord := result.Items[1]
		if secondRecord.GetString("title") != "Second Post" {
			t.Errorf("두 번째 Record title 불일치: %s", secondRecord.GetString("title"))
		}

		if secondRecord.GetBool("published") {
			t.Error("두 번째 Record published 상태가 잘못됨 (false여야 함)")
		}
	})

	t.Run("Record 단일 조회 및 업데이트", func(t *testing.T) {
		// 단일 Record 조회
		record, err := client.Records.GetOne(ctx, "posts", "post1", nil)
		if err != nil {
			t.Fatalf("Record 단일 조회 실패: %v", err)
		}

		// 원본 값 확인
		if record.GetString("title") != "First Post" {
			t.Errorf("조회된 Record title 불일치: %s", record.GetString("title"))
		}

		// 필드 업데이트
		record.Set("title", "Updated First Post")
		record.Set("content", "This is the updated first post")

		// 기존 tags에 새 태그 추가
		existingTags := record.GetStringSlice("tags")
		newTags := append(existingTags, "updated")
		record.Set("tags", newTags)

		// 업데이트 실행
		updatedRecord, err := client.Records.Update(ctx, "posts", record.ID, record)
		if err != nil {
			t.Fatalf("Record 업데이트 실패: %v", err)
		}

		// 업데이트 결과 검증
		if updatedRecord.GetString("title") != "Updated First Post" {
			t.Errorf("업데이트된 title 불일치: %s", updatedRecord.GetString("title"))
		}

		updatedTags := updatedRecord.GetStringSlice("tags")
		if len(updatedTags) != 3 || updatedTags[2] != "updated" {
			t.Errorf("업데이트된 tags 불일치: %v", updatedTags)
		}
	})

	t.Run("Record 생성", func(t *testing.T) {
		// 새 Record 데이터 준비
		newRecord := &Record{}
		newRecord.CollectionName = "posts"
		newRecord.Set("title", "New Post")
		newRecord.Set("content", "This is a new post")
		newRecord.Set("published", true)
		newRecord.Set("author", "user3")
		newRecord.Set("tags", []string{"new"})

		// Record 생성
		createdRecord, err := client.Records.Create(ctx, "posts", newRecord)
		if err != nil {
			t.Fatalf("Record 생성 실패: %v", err)
		}

		// 생성 결과 검증
		if createdRecord.GetString("title") != "New Post" {
			t.Errorf("생성된 Record title 불일치: %s", createdRecord.GetString("title"))
		}

		if createdRecord.ID == "" {
			t.Error("생성된 Record에 ID가 없습니다")
		}

		if createdRecord.GetString("created") == "" {
			t.Error("생성된 Record에 created 필드가 없습니다")
		}
	})
}

// testTypeSafeStructUsage는 타입 안전한 구조체 사용 방식을 테스트합니다
func testTypeSafeStructUsage(t *testing.T, client *Client, ctx context.Context) {
	// 타입 안전한 구조체 정의 (실제로는 코드 생성으로 만들어짐)
	type Post struct {
		BaseModel
		BaseDateTime
		Title     string   `json:"title"`
		Content   string   `json:"content"`
		Published bool     `json:"published"`
		Author    string   `json:"author"`
		Tags      []string `json:"tags"`
	}

	// 생성자 함수
	NewPost := func() *Post {
		return &Post{
			BaseModel: BaseModel{
				CollectionName: "posts",
			},
		}
	}

	// ToMap 메서드
	postToMap := func(p *Post) map[string]any {
		return map[string]any{
			"title":     p.Title,
			"content":   p.Content,
			"published": p.Published,
			"author":    p.Author,
			"tags":      p.Tags,
		}
	}

	t.Run("타입 안전한 구조체 생성 및 필드 설정", func(t *testing.T) {
		// 새 Post 생성
		post := NewPost()

		// 타입 안전한 필드 설정
		post.Title = "Type Safe Post"
		post.Content = "Content created using type-safe struct"
		post.Published = true
		post.Author = "typesafe_user"
		post.Tags = []string{"typesafe", "test"}

		// 필드 값 검증
		if post.Title != "Type Safe Post" {
			t.Errorf("Title 필드 설정 실패: %s", post.Title)
		}

		if post.Content != "Content created using type-safe struct" {
			t.Errorf("Content 필드 설정 실패: %s", post.Content)
		}

		if !post.Published {
			t.Error("Published 필드 설정 실패")
		}

		if post.Author != "typesafe_user" {
			t.Errorf("Author 필드 설정 실패: %s", post.Author)
		}

		if len(post.Tags) != 2 || post.Tags[0] != "typesafe" || post.Tags[1] != "test" {
			t.Errorf("Tags 필드 설정 실패: %v", post.Tags)
		}
	})

	t.Run("타입 안전한 구조체로 Record 변환", func(t *testing.T) {
		// Record 조회
		record, err := client.Records.GetOne(ctx, "posts", "post1", nil)
		if err != nil {
			t.Fatalf("Record 조회 실패: %v", err)
		}

		// Record를 타입 안전한 구조체로 변환
		var post Post
		recordData, _ := json.Marshal(record)
		err = json.Unmarshal(recordData, &post)
		if err != nil {
			t.Fatalf("Record to Post 변환 실패: %v", err)
		}

		// BaseModel 필드 수동 설정
		post.BaseModel = record.BaseModel

		// 변환 결과 검증
		if post.Title != "First Post" {
			t.Errorf("변환된 Post title 불일치: %s", post.Title)
		}

		if post.Content != "This is the first post" {
			t.Errorf("변환된 Post content 불일치: %s", post.Content)
		}

		if !post.Published {
			t.Error("변환된 Post published 상태 불일치")
		}

		if post.ID != "post1" {
			t.Errorf("변환된 Post ID 불일치: %s", post.ID)
		}
	})

	t.Run("타입 안전한 구조체 업데이트", func(t *testing.T) {
		// 기존 Record 조회 후 구조체로 변환
		record, err := client.Records.GetOne(ctx, "posts", "post1", nil)
		if err != nil {
			t.Fatalf("Record 조회 실패: %v", err)
		}

		var post Post
		recordData, _ := json.Marshal(record)
		json.Unmarshal(recordData, &post)
		post.BaseModel = record.BaseModel

		// 타입 안전한 방식으로 필드 업데이트
		post.Title = "Updated First Post"
		post.Content = "This is the updated first post"
		post.Tags = append(post.Tags, "updated")

		// 업데이트 실행
		updatedRecord, err := client.Records.Update(ctx, "posts", post.ID, postToMap(&post))
		if err != nil {
			t.Fatalf("Post 업데이트 실패: %v", err)
		}

		// 업데이트 결과 검증
		if updatedRecord.GetString("title") != "Updated First Post" {
			t.Errorf("업데이트된 title 불일치: %s", updatedRecord.GetString("title"))
		}

		updatedTags := updatedRecord.GetStringSlice("tags")
		if len(updatedTags) < 3 || updatedTags[len(updatedTags)-1] != "updated" {
			t.Errorf("업데이트된 tags 불일치: %v", updatedTags)
		}
	})

	t.Run("타입 안전한 구조체 생성", func(t *testing.T) {
		// 새 Post 생성
		post := NewPost()
		post.Title = "New Type Safe Post"
		post.Content = "This is a new post created with type-safe struct"
		post.Published = true
		post.Author = "user3"
		post.Tags = []string{"new", "typesafe"}

		// Record 생성
		createdRecord, err := client.Records.Create(ctx, "posts", postToMap(post))
		if err != nil {
			t.Fatalf("Post 생성 실패: %v", err)
		}

		// 생성 결과 검증
		if createdRecord.GetString("title") != "New Type Safe Post" {
			t.Errorf("생성된 Post title 불일치: %s", createdRecord.GetString("title"))
		}

		if createdRecord.ID == "" {
			t.Error("생성된 Post에 ID가 없습니다")
		}
	})
}

// testInterconversion는 두 방식 간의 상호 변환을 테스트합니다
func testInterconversion(t *testing.T, client *Client, ctx context.Context) {
	// 타입 안전한 구조체 정의
	type User struct {
		BaseModel
		BaseDateTime
		Name   string `json:"name"`
		Email  string `json:"email"`
		Age    *int   `json:"age,omitempty"`
		Active *bool  `json:"active,omitempty"`
	}

	t.Run("Record에서 구조체로 변환", func(t *testing.T) {
		// Record 생성
		record := &Record{}
		record.CollectionName = "users"
		record.Set("name", "John Doe")
		record.Set("email", "john@example.com")
		record.Set("age", 30)
		record.Set("active", true)
		record.ID = "user1"
		record.CollectionID = "users_collection"

		// Record를 구조체로 변환
		var user User
		recordData, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("Record 직렬화 실패: %v", err)
		}

		err = json.Unmarshal(recordData, &user)
		if err != nil {
			t.Fatalf("구조체 역직렬화 실패: %v", err)
		}

		// BaseModel 필드 수동 복사
		user.BaseModel = record.BaseModel

		// 변환 결과 검증
		if user.Name != "John Doe" {
			t.Errorf("Name 변환 실패: %s", user.Name)
		}

		if user.Email != "john@example.com" {
			t.Errorf("Email 변환 실패: %s", user.Email)
		}

		if user.Age == nil || *user.Age != 30 {
			t.Errorf("Age 변환 실패: %v", user.Age)
		}

		if user.Active == nil || !*user.Active {
			t.Errorf("Active 변환 실패: %v", user.Active)
		}

		if user.ID != "user1" {
			t.Errorf("ID 변환 실패: %s", user.ID)
		}
	})

	t.Run("구조체에서 Record로 변환", func(t *testing.T) {
		// 구조체 생성
		age := 25
		active := false
		user := User{
			BaseModel: BaseModel{
				ID:             "user2",
				CollectionID:   "users_collection",
				CollectionName: "users",
			},
			Name:   "Jane Smith",
			Email:  "jane@example.com",
			Age:    &age,
			Active: &active,
		}

		// 구조체를 Record로 변환
		record := &Record{}
		record.BaseModel = user.BaseModel

		// 구조체 데이터를 JSON으로 변환 후 Record에 설정
		userData, err := json.Marshal(user)
		if err != nil {
			t.Fatalf("구조체 직렬화 실패: %v", err)
		}

		var userMap map[string]any
		err = json.Unmarshal(userData, &userMap)
		if err != nil {
			t.Fatalf("맵 역직렬화 실패: %v", err)
		}

		for key, value := range userMap {
			if key != "id" && key != "collectionId" && key != "collectionName" {
				record.Set(key, value)
			}
		}

		// 변환 결과 검증
		if record.GetString("name") != "Jane Smith" {
			t.Errorf("Name 변환 실패: %s", record.GetString("name"))
		}

		if record.GetString("email") != "jane@example.com" {
			t.Errorf("Email 변환 실패: %s", record.GetString("email"))
		}

		if record.GetFloat("age") != 25 {
			t.Errorf("Age 변환 실패: %f", record.GetFloat("age"))
		}

		if record.GetBool("active") {
			t.Error("Active 변환 실패: true여야 하는데 false")
		}

		if record.ID != "user2" {
			t.Errorf("ID 변환 실패: %s", record.ID)
		}
	})

	t.Run("양방향 변환 일관성", func(t *testing.T) {
		// 원본 Record 생성
		originalRecord := &Record{}
		originalRecord.CollectionName = "users"
		originalRecord.Set("name", "Test User")
		originalRecord.Set("email", "test@example.com")
		originalRecord.Set("age", 35)
		originalRecord.Set("active", true)
		originalRecord.ID = "test_user"

		// Record -> 구조체 -> Record 변환
		var user User
		recordData, _ := json.Marshal(originalRecord)
		json.Unmarshal(recordData, &user)
		user.BaseModel = originalRecord.BaseModel

		// 구조체에서 다시 Record로 변환
		convertedRecord := &Record{}
		convertedRecord.BaseModel = user.BaseModel

		userData, _ := json.Marshal(user)
		var userMap map[string]any
		json.Unmarshal(userData, &userMap)

		for key, value := range userMap {
			if key != "id" && key != "collectionId" && key != "collectionName" {
				convertedRecord.Set(key, value)
			}
		}

		// 원본과 변환된 Record 비교
		if originalRecord.GetString("name") != convertedRecord.GetString("name") {
			t.Errorf("Name 일관성 실패: %s != %s",
				originalRecord.GetString("name"), convertedRecord.GetString("name"))
		}

		if originalRecord.GetString("email") != convertedRecord.GetString("email") {
			t.Errorf("Email 일관성 실패: %s != %s",
				originalRecord.GetString("email"), convertedRecord.GetString("email"))
		}

		if originalRecord.GetFloat("age") != convertedRecord.GetFloat("age") {
			t.Errorf("Age 일관성 실패: %f != %f",
				originalRecord.GetFloat("age"), convertedRecord.GetFloat("age"))
		}

		if originalRecord.GetBool("active") != convertedRecord.GetBool("active") {
			t.Errorf("Active 일관성 실패: %t != %t",
				originalRecord.GetBool("active"), convertedRecord.GetBool("active"))
		}
	})
}

// testPerformanceComparison는 두 방식의 성능을 비교합니다
func testPerformanceComparison(t *testing.T, client *Client, ctx context.Context) {
	// 타입 안전한 구조체 정의
	type Product struct {
		BaseModel
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		InStock     bool    `json:"in_stock"`
		Category    string  `json:"category"`
	}

	// 테스트 데이터 준비
	testData := map[string]any{
		"name":        "Test Product",
		"description": "This is a test product for performance comparison",
		"price":       99.99,
		"in_stock":    true,
		"category":    "electronics",
	}

	t.Run("동적 Record 성능", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			// Record 생성 및 필드 설정
			record := &Record{}
			record.CollectionName = "products"

			for key, value := range testData {
				record.Set(key, value)
			}

			// 필드 조회
			_ = record.GetString("name")
			_ = record.GetString("description")
			_ = record.GetFloat("price")
			_ = record.GetBool("in_stock")
			_ = record.GetString("category")

			// JSON 직렬화 테스트
			_, _ = json.Marshal(record)
		}

		dynamicDuration := time.Since(start)
		t.Logf("동적 Record 방식 (1000회): %v", dynamicDuration)
	})

	t.Run("타입 안전한 구조체 성능", func(t *testing.T) {
		start := time.Now()

		for i := 0; i < 1000; i++ {
			// 구조체 생성 및 필드 설정
			product := Product{
				BaseModel: BaseModel{
					CollectionName: "products",
				},
				Name:        testData["name"].(string),
				Description: testData["description"].(string),
				Price:       testData["price"].(float64),
				InStock:     testData["in_stock"].(bool),
				Category:    testData["category"].(string),
			}

			// 필드 접근
			_ = product.Name
			_ = product.Description
			_ = product.Price
			_ = product.InStock
			_ = product.Category

			// JSON 변환 (ToMap 대신)
			_, _ = json.Marshal(product)
		}

		typeSafeDuration := time.Since(start)
		t.Logf("타입 안전한 구조체 방식 (1000회): %v", typeSafeDuration)
	})

	t.Run("메모리 사용량 비교", func(t *testing.T) {
		// 동적 Record 메모리 사용량 측정
		var dynamicRecords []*Record
		for i := 0; i < 100; i++ {
			record := &Record{}
			record.CollectionName = "products"
			for key, value := range testData {
				record.Set(key, value)
			}
			dynamicRecords = append(dynamicRecords, record)
		}

		// 타입 안전한 구조체 메모리 사용량 측정
		var typeSafeProducts []Product
		for i := 0; i < 100; i++ {
			product := Product{
				BaseModel: BaseModel{
					CollectionName: "products",
				},
				Name:        testData["name"].(string),
				Description: testData["description"].(string),
				Price:       testData["price"].(float64),
				InStock:     testData["in_stock"].(bool),
				Category:    testData["category"].(string),
			}
			typeSafeProducts = append(typeSafeProducts, product)
		}

		// 메모리 사용량은 실제 측정이 어려우므로 로그만 출력
		t.Logf("동적 Record 개수: %d", len(dynamicRecords))
		t.Logf("타입 안전한 구조체 개수: %d", len(typeSafeProducts))
	})
}

// TestCompatibilityBetweenUsagePatterns는 두 사용 방식 간의 호환성을 테스트합니다
func TestCompatibilityBetweenUsagePatterns(t *testing.T) {
	t.Run("API 호환성", func(t *testing.T) {
		// 동적 Record로 생성한 데이터
		dynamicRecord := &Record{}
		dynamicRecord.CollectionName = "compatibility_test"
		dynamicRecord.Set("field1", "value1")
		dynamicRecord.Set("field2", 42)
		dynamicRecord.Set("field3", true)

		// 타입 안전한 구조체 정의
		type CompatibilityTest struct {
			BaseModel
			Field1 string `json:"field1"`
			Field2 int    `json:"field2"`
			Field3 bool   `json:"field3"`
		}

		// Record 데이터를 구조체로 변환
		var structData CompatibilityTest
		jsonData, _ := json.Marshal(dynamicRecord)
		json.Unmarshal(jsonData, &structData)
		structData.BaseModel = dynamicRecord.BaseModel

		// 변환 검증
		if structData.Field1 != "value1" {
			t.Errorf("Field1 호환성 실패: %s", structData.Field1)
		}
		if structData.Field2 != 42 {
			t.Errorf("Field2 호환성 실패: %d", structData.Field2)
		}
		if structData.Field3 != true {
			t.Errorf("Field3 호환성 실패: %t", structData.Field3)
		}

		// 구조체 데이터를 다시 Record로 변환
		backToRecord := &Record{}
		backToRecord.BaseModel = structData.BaseModel
		structJson, _ := json.Marshal(structData)
		var structMap map[string]any
		json.Unmarshal(structJson, &structMap)

		for key, value := range structMap {
			if key != "id" && key != "collectionId" && key != "collectionName" {
				backToRecord.Set(key, value)
			}
		}

		// 역변환 검증
		if backToRecord.GetString("field1") != "value1" {
			t.Errorf("역변환 Field1 실패: %s", backToRecord.GetString("field1"))
		}
		if backToRecord.GetFloat("field2") != 42 {
			t.Errorf("역변환 Field2 실패: %f", backToRecord.GetFloat("field2"))
		}
		if backToRecord.GetBool("field3") != true {
			t.Errorf("역변환 Field3 실패: %t", backToRecord.GetBool("field3"))
		}
	})

	t.Run("JSON 직렬화 호환성", func(t *testing.T) {
		// 동적 Record JSON 직렬화
		record := &Record{}
		record.Set("name", "JSON Test")
		record.Set("value", 123)
		record.ID = "json_test"

		recordJson, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("Record JSON 직렬화 실패: %v", err)
		}

		// 타입 안전한 구조체 정의 및 JSON 직렬화
		type JsonTest struct {
			BaseModel
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		structData := JsonTest{
			BaseModel: BaseModel{ID: "json_test"},
			Name:      "JSON Test",
			Value:     123,
		}

		structJson, err := json.Marshal(structData)
		if err != nil {
			t.Fatalf("구조체 JSON 직렬화 실패: %v", err)
		}

		// JSON 데이터 비교 (순서는 다를 수 있으므로 개별 필드 확인)
		var recordData, structDataMap map[string]any
		json.Unmarshal(recordJson, &recordData)
		json.Unmarshal(structJson, &structDataMap)

		if recordData["name"] != structDataMap["name"] {
			t.Errorf("JSON name 필드 불일치: %v != %v", recordData["name"], structDataMap["name"])
		}

		// value 필드는 타입이 다를 수 있으므로 문자열로 비교
		if fmt.Sprintf("%v", recordData["value"]) != fmt.Sprintf("%v", structDataMap["value"]) {
			t.Errorf("JSON value 필드 불일치: %v != %v", recordData["value"], structDataMap["value"])
		}
	})
}
