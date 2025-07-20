package pocketbase

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pocketbase/pocketbase/tools/types"
)

// TestMigrationBackwardCompatibility는 기존 코드와의 하위 호환성을 테스트합니다
func TestMigrationBackwardCompatibility(t *testing.T) {
	t.Run("BaseModel 변경사항 호환성", func(t *testing.T) {
		// 기존 방식으로 BaseModel 사용
		baseModel := BaseModel{
			ID:             "test_id",
			CollectionID:   "test_collection_id",
			CollectionName: "test_collection",
		}

		// 기본 필드들이 여전히 접근 가능한지 확인
		if baseModel.ID != "test_id" {
			t.Errorf("BaseModel.ID 접근 실패: %s", baseModel.ID)
		}

		if baseModel.CollectionID != "test_collection_id" {
			t.Errorf("BaseModel.CollectionID 접근 실패: %s", baseModel.CollectionID)
		}

		if baseModel.CollectionName != "test_collection" {
			t.Errorf("BaseModel.CollectionName 접근 실패: %s", baseModel.CollectionName)
		}
	})

	t.Run("BaseDateTime 분리 후 호환성", func(t *testing.T) {
		// BaseDateTime 구조체가 독립적으로 사용 가능한지 확인
		now := types.DateTime{}
		baseDateTime := BaseDateTime{
			Created: now,
			Updated: now,
		}

		// 필드 접근 확인
		if baseDateTime.Created != now {
			t.Error("BaseDateTime.Created 접근 실패")
		}

		if baseDateTime.Updated != now {
			t.Error("BaseDateTime.Updated 접근 실패")
		}

		// JSON 직렬화/역직렬화 확인
		jsonData, err := json.Marshal(baseDateTime)
		if err != nil {
			t.Fatalf("BaseDateTime JSON 직렬화 실패: %v", err)
		}

		var deserializedDateTime BaseDateTime
		err = json.Unmarshal(jsonData, &deserializedDateTime)
		if err != nil {
			t.Fatalf("BaseDateTime JSON 역직렬화 실패: %v", err)
		}
	})

	t.Run("Record 구조체 호환성", func(t *testing.T) {
		// 기존 Record 사용 방식이 여전히 작동하는지 확인
		record := &Record{}
		record.CollectionName = "test_collection"

		// 기본 필드 설정 및 조회
		record.Set("name", "Test Record")
		record.Set("value", 42)
		record.Set("active", true)

		if record.GetString("name") != "Test Record" {
			t.Errorf("Record.GetString 호환성 실패: %s", record.GetString("name"))
		}

		if record.GetFloat("value") != 42 {
			t.Errorf("Record.GetInt 호환성 실패: %f", record.GetFloat("value"))
		}

		if !record.GetBool("active") {
			t.Error("Record.GetBool 호환성 실패")
		}

		// 필드 접근 호환성 확인
		if record.GetString("name") != "Test Record" {
			t.Errorf("Record 필드 접근 호환성 실패: %v", record.GetString("name"))
		}
	})

	t.Run("기존 API 메서드 호환성", func(t *testing.T) {
		// 테스트용 서버 설정
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"id": "test_record",
				"collectionId": "test_collection_id",
				"collectionName": "test_collection",
				"created": "2025-01-01T10:00:00.000Z",
				"updated": "2025-01-01T10:00:00.000Z",
				"name": "Test Record",
				"value": 42
			}`
			w.Write([]byte(response))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		// 기존 API 메서드들이 여전히 작동하는지 확인
		record, err := client.Records.GetOne(ctx, "test_collection", "test_record", nil)
		if err != nil {
			t.Fatalf("기존 GetOne API 호환성 실패: %v", err)
		}

		// 반환된 Record가 기존 방식으로 사용 가능한지 확인
		if record.GetString("name") != "Test Record" {
			t.Errorf("기존 API 반환값 호환성 실패: %s", record.GetString("name"))
		}

		if record.ID != "test_record" {
			t.Errorf("기존 API BaseModel 필드 호환성 실패: %s", record.ID)
		}
	})
}

// TestMigrationScenarios는 다양한 마이그레이션 시나리오를 테스트합니다
func TestMigrationScenarios(t *testing.T) {
	t.Run("Legacy 구조체에서 Latest 구조체로 마이그레이션", func(t *testing.T) {
		// Legacy 스타일 구조체 (BaseModel + BaseDateTime 임베딩)
		type LegacyPost struct {
			BaseModel
			BaseDateTime
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		// Latest 스타일 구조체 (BaseModel만 임베딩)
		type LatestPost struct {
			BaseModel
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		// Legacy 구조체 생성
		legacyPost := LegacyPost{
			BaseModel: BaseModel{
				ID:             "post1",
				CollectionID:   "posts_collection",
				CollectionName: "posts",
			},
			BaseDateTime: BaseDateTime{
				Created: types.DateTime{},
				Updated: types.DateTime{},
			},
			Title:   "Legacy Post",
			Content: "This is a legacy post",
		}

		// Legacy에서 Latest로 마이그레이션
		latestPost := LatestPost{
			BaseModel: legacyPost.BaseModel,
			Title:     legacyPost.Title,
			Content:   legacyPost.Content,
		}

		// 마이그레이션 결과 검증
		if latestPost.ID != legacyPost.ID {
			t.Errorf("마이그레이션 ID 불일치: %s != %s", latestPost.ID, legacyPost.ID)
		}

		if latestPost.Title != legacyPost.Title {
			t.Errorf("마이그레이션 Title 불일치: %s != %s", latestPost.Title, legacyPost.Title)
		}

		if latestPost.Content != legacyPost.Content {
			t.Errorf("마이그레이션 Content 불일치: %s != %s", latestPost.Content, legacyPost.Content)
		}

		// JSON 직렬화 호환성 확인
		legacyJson, _ := json.Marshal(legacyPost)
		latestJson, _ := json.Marshal(latestPost)

		var legacyMap, latestMap map[string]any
		json.Unmarshal(legacyJson, &legacyMap)
		json.Unmarshal(latestJson, &latestMap)

		// 공통 필드들이 동일한지 확인
		commonFields := []string{"id", "collectionId", "collectionName", "title", "content"}
		for _, field := range commonFields {
			if legacyMap[field] != latestMap[field] {
				t.Errorf("마이그레이션 후 %s 필드 불일치: %v != %v", field, legacyMap[field], latestMap[field])
			}
		}
	})

	t.Run("점진적 마이그레이션 시나리오", func(t *testing.T) {
		// 1단계: 기존 코드 (Record 사용)
		oldRecord := &Record{}
		oldRecord.CollectionName = "users"
		oldRecord.Set("name", "John Doe")
		oldRecord.Set("email", "john@example.com")
		oldRecord.ID = "user1"

		// 2단계: Legacy 구조체로 마이그레이션
		type LegacyUser struct {
			BaseModel
			BaseDateTime
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		var legacyUser LegacyUser
		recordData, _ := json.Marshal(oldRecord)
		json.Unmarshal(recordData, &legacyUser)
		legacyUser.BaseModel = oldRecord.BaseModel

		// 3단계: Latest 구조체로 마이그레이션
		type LatestUser struct {
			BaseModel
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		latestUser := LatestUser{
			BaseModel: legacyUser.BaseModel,
			Name:      legacyUser.Name,
			Email:     legacyUser.Email,
		}

		// 각 단계별 데이터 일관성 확인
		if oldRecord.GetString("name") != legacyUser.Name {
			t.Errorf("1->2단계 마이그레이션 Name 불일치: %s != %s", oldRecord.GetString("name"), legacyUser.Name)
		}

		if legacyUser.Name != latestUser.Name {
			t.Errorf("2->3단계 마이그레이션 Name 불일치: %s != %s", legacyUser.Name, latestUser.Name)
		}

		if oldRecord.GetString("email") != latestUser.Email {
			t.Errorf("전체 마이그레이션 Email 불일치: %s != %s", oldRecord.GetString("email"), latestUser.Email)
		}

		if oldRecord.ID != latestUser.ID {
			t.Errorf("전체 마이그레이션 ID 불일치: %s != %s", oldRecord.ID, latestUser.ID)
		}
	})

	t.Run("스키마 버전별 코드 생성 마이그레이션", func(t *testing.T) {
		// Legacy 스키마로 생성된 구조체 시뮬레이션
		type LegacyGeneratedPost struct {
			BaseModel
			BaseDateTime
			Title     string `json:"title"`
			Content   string `json:"content"`
			Published bool   `json:"published"`
		}

		// Latest 스키마로 생성된 구조체 시뮬레이션
		type LatestGeneratedPost struct {
			BaseModel
			Title     string `json:"title"`
			Content   string `json:"content"`
			Published bool   `json:"published"`
		}

		// Legacy 구조체 데이터
		legacyPost := LegacyGeneratedPost{
			BaseModel: BaseModel{
				ID:             "generated_post1",
				CollectionID:   "posts_collection",
				CollectionName: "posts",
			},
			BaseDateTime: BaseDateTime{
				Created: types.DateTime{},
				Updated: types.DateTime{},
			},
			Title:     "Generated Legacy Post",
			Content:   "Content from legacy generated struct",
			Published: true,
		}

		// 마이그레이션 함수 시뮬레이션
		migrateToLatest := func(legacy LegacyGeneratedPost) LatestGeneratedPost {
			return LatestGeneratedPost{
				BaseModel: legacy.BaseModel,
				Title:     legacy.Title,
				Content:   legacy.Content,
				Published: legacy.Published,
			}
		}

		// 마이그레이션 실행
		latestPost := migrateToLatest(legacyPost)

		// 마이그레이션 결과 검증
		if latestPost.BaseModel != legacyPost.BaseModel {
			t.Error("BaseModel 마이그레이션 실패")
		}

		if latestPost.Title != legacyPost.Title {
			t.Errorf("Title 마이그레이션 실패: %s != %s", latestPost.Title, legacyPost.Title)
		}

		if latestPost.Content != legacyPost.Content {
			t.Errorf("Content 마이그레이션 실패: %s != %s", latestPost.Content, legacyPost.Content)
		}

		if latestPost.Published != legacyPost.Published {
			t.Errorf("Published 마이그레이션 실패: %t != %t", latestPost.Published, legacyPost.Published)
		}
	})
}

// TestBaseModelChangesImpact는 BaseModel 변경사항이 기존 코드에 미치는 영향을 테스트합니다
func TestBaseModelChangesImpact(t *testing.T) {
	t.Run("BaseModel 필드 접근 영향도", func(t *testing.T) {
		// 기존 코드에서 BaseModel 필드에 직접 접근하는 패턴
		type ExistingStruct struct {
			BaseModel
			Name string `json:"name"`
		}

		existing := ExistingStruct{
			BaseModel: BaseModel{
				ID:             "existing_id",
				CollectionID:   "existing_collection_id",
				CollectionName: "existing_collection",
			},
			Name: "Existing Item",
		}

		// 기존 코드 패턴들이 여전히 작동하는지 확인
		if existing.ID != "existing_id" {
			t.Errorf("기존 ID 접근 패턴 실패: %s", existing.ID)
		}

		if existing.CollectionID != "existing_collection_id" {
			t.Errorf("기존 CollectionID 접근 패턴 실패: %s", existing.CollectionID)
		}

		if existing.CollectionName != "existing_collection" {
			t.Errorf("기존 CollectionName 접근 패턴 실패: %s", existing.CollectionName)
		}

		// JSON 직렬화가 여전히 작동하는지 확인
		jsonData, err := json.Marshal(existing)
		if err != nil {
			t.Fatalf("기존 구조체 JSON 직렬화 실패: %v", err)
		}

		var deserialized ExistingStruct
		err = json.Unmarshal(jsonData, &deserialized)
		if err != nil {
			t.Fatalf("기존 구조체 JSON 역직렬화 실패: %v", err)
		}

		if deserialized.ID != existing.ID {
			t.Errorf("JSON 직렬화 후 ID 불일치: %s != %s", deserialized.ID, existing.ID)
		}
	})

	t.Run("Created/Updated 필드 제거 영향도", func(t *testing.T) {
		// BaseModel에서 Created/Updated 필드가 제거된 후의 영향 확인
		baseModel := BaseModel{
			ID:             "test_id",
			CollectionID:   "test_collection_id",
			CollectionName: "test_collection",
		}

		// BaseModel에 Created/Updated 필드가 없는지 확인
		jsonData, _ := json.Marshal(baseModel)
		var baseModelMap map[string]any
		json.Unmarshal(jsonData, &baseModelMap)

		// Created/Updated 필드가 BaseModel에 없어야 함
		if _, exists := baseModelMap["created"]; exists {
			t.Error("BaseModel에 created 필드가 여전히 존재합니다")
		}

		if _, exists := baseModelMap["updated"]; exists {
			t.Error("BaseModel에 updated 필드가 여전히 존재합니다")
		}

		// BaseDateTime을 별도로 사용해야 함
		baseDateTime := BaseDateTime{
			Created: types.DateTime{},
			Updated: types.DateTime{},
		}

		dateTimeJson, _ := json.Marshal(baseDateTime)
		var dateTimeMap map[string]any
		json.Unmarshal(dateTimeJson, &dateTimeMap)

		// BaseDateTime에는 created/updated 필드가 있어야 함
		if _, exists := dateTimeMap["created"]; !exists {
			t.Error("BaseDateTime에 created 필드가 없습니다")
		}

		if _, exists := dateTimeMap["updated"]; !exists {
			t.Error("BaseDateTime에 updated 필드가 없습니다")
		}
	})

	t.Run("임베딩 패턴 변경 영향도", func(t *testing.T) {
		// Legacy 패턴: BaseModel + BaseDateTime 임베딩
		type LegacyPattern struct {
			BaseModel
			BaseDateTime
			Data string `json:"data"`
		}

		// Latest 패턴: BaseModel만 임베딩
		type LatestPattern struct {
			BaseModel
			Data string `json:"data"`
		}

		// Legacy 패턴 사용
		legacyItem := LegacyPattern{
			BaseModel: BaseModel{
				ID:             "pattern_test",
				CollectionID:   "pattern_collection",
				CollectionName: "patterns",
			},
			BaseDateTime: BaseDateTime{
				Created: types.DateTime{},
				Updated: types.DateTime{},
			},
			Data: "Legacy Pattern Data",
		}

		// Latest 패턴 사용
		latestItem := LatestPattern{
			BaseModel: BaseModel{
				ID:             "pattern_test",
				CollectionID:   "pattern_collection",
				CollectionName: "patterns",
			},
			Data: "Latest Pattern Data",
		}

		// 두 패턴 모두 BaseModel 필드에 접근 가능해야 함
		if legacyItem.ID != latestItem.ID {
			t.Errorf("패턴별 BaseModel 접근 불일치: %s != %s", legacyItem.ID, latestItem.ID)
		}

		// Legacy 패턴은 타임스탬프 필드에 접근 가능
		_ = legacyItem.Created
		_ = legacyItem.Updated

		// Latest 패턴은 타임스탬프 필드에 직접 접근 불가 (컴파일 에러 방지를 위해 주석)
		// _ = latestItem.Created // 이 라인은 컴파일 에러를 발생시킴

		// JSON 직렬화 결과 비교
		legacyJson, _ := json.Marshal(legacyItem)
		latestJson, _ := json.Marshal(latestItem)

		var legacyMap, latestMap map[string]any
		json.Unmarshal(legacyJson, &legacyMap)
		json.Unmarshal(latestJson, &latestMap)

		// Legacy는 created/updated 필드를 포함해야 함
		if _, exists := legacyMap["created"]; !exists {
			t.Error("Legacy 패턴에 created 필드가 없습니다")
		}

		// Latest는 created/updated 필드를 포함하지 않아야 함
		if _, exists := latestMap["created"]; exists {
			t.Error("Latest 패턴에 created 필드가 포함되었습니다")
		}
	})
}

// TestGradualMigrationStrategy는 점진적 마이그레이션 전략을 테스트합니다
func TestGradualMigrationStrategy(t *testing.T) {
	t.Run("단계별 마이그레이션 전략", func(t *testing.T) {
		// 1단계: 기존 Record 기반 코드
		type Phase1Service struct {
			client *Client
		}

		phase1GetUser := func(s *Phase1Service, ctx context.Context, id string) (*Record, error) {
			return s.client.Records.GetOne(ctx, "users", id, nil)
		}

		// 2단계: Legacy 구조체 도입 (BaseModel + BaseDateTime)
		type Phase2User struct {
			BaseModel
			BaseDateTime
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		type Phase2Service struct {
			client *Client
		}

		phase2GetUser := func(s *Phase2Service, ctx context.Context, id string) (*Phase2User, error) {
			record, err := s.client.Records.GetOne(ctx, "users", id, nil)
			if err != nil {
				return nil, err
			}

			var user Phase2User
			recordData, _ := json.Marshal(record)
			json.Unmarshal(recordData, &user)
			user.BaseModel = record.BaseModel
			return &user, nil
		}

		// 3단계: Latest 구조체로 전환 (BaseModel만)
		type Phase3User struct {
			BaseModel
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		type Phase3Service struct {
			client *Client
		}

		phase3GetUser := func(s *Phase3Service, ctx context.Context, id string) (*Phase3User, error) {
			record, err := s.client.Records.GetOne(ctx, "users", id, nil)
			if err != nil {
				return nil, err
			}

			var user Phase3User
			recordData, _ := json.Marshal(record)
			json.Unmarshal(recordData, &user)
			user.BaseModel = record.BaseModel
			return &user, nil
		}

		// 테스트용 서버 설정
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			response := `{
				"id": "user123",
				"collectionId": "users_collection",
				"collectionName": "users",
				"created": "2025-01-01T10:00:00.000Z",
				"updated": "2025-01-01T10:00:00.000Z",
				"name": "Test User",
				"email": "test@example.com"
			}`
			w.Write([]byte(response))
		}))
		defer server.Close()

		client := NewClient(server.URL)
		ctx := context.Background()

		// 각 단계별 서비스 테스트
		phase1Service := &Phase1Service{client: client}
		phase2Service := &Phase2Service{client: client}
		phase3Service := &Phase3Service{client: client}

		// Phase 1 테스트
		phase1User, err := phase1GetUser(phase1Service, ctx, "user123")
		if err != nil {
			t.Fatalf("Phase 1 사용자 조회 실패: %v", err)
		}

		// Phase 2 테스트
		phase2User, err := phase2GetUser(phase2Service, ctx, "user123")
		if err != nil {
			t.Fatalf("Phase 2 사용자 조회 실패: %v", err)
		}

		// Phase 3 테스트
		phase3User, err := phase3GetUser(phase3Service, ctx, "user123")
		if err != nil {
			t.Fatalf("Phase 3 사용자 조회 실패: %v", err)
		}

		// 각 단계별 결과 일관성 확인
		if phase1User.GetString("name") != phase2User.Name {
			t.Errorf("Phase 1->2 Name 불일치: %s != %s", phase1User.GetString("name"), phase2User.Name)
		}

		if phase2User.Name != phase3User.Name {
			t.Errorf("Phase 2->3 Name 불일치: %s != %s", phase2User.Name, phase3User.Name)
		}

		if phase1User.ID != phase3User.ID {
			t.Errorf("Phase 1->3 ID 불일치: %s != %s", phase1User.ID, phase3User.ID)
		}
	})

	t.Run("호환성 어댑터 패턴", func(t *testing.T) {
		// 기존 Legacy 구조체
		type LegacyProduct struct {
			BaseModel
			BaseDateTime
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}

		// 새로운 Latest 구조체
		type LatestProduct struct {
			BaseModel
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		}

		// 호환성 어댑터
		type ProductAdapter struct {
			legacy *LegacyProduct
			latest *LatestProduct
		}

		// 어댑터 생성자
		NewProductAdapter := func() *ProductAdapter {
			return &ProductAdapter{}
		}

		// Legacy 데이터 설정
		setLegacy := func(adapter *ProductAdapter, legacy *LegacyProduct) {
			adapter.legacy = legacy
		}

		// Latest 데이터 설정
		setLatest := func(adapter *ProductAdapter, latest *LatestProduct) {
			adapter.latest = latest
		}

		// 통합 인터페이스 메서드들
		getName := func(adapter *ProductAdapter) string {
			if adapter.latest != nil {
				return adapter.latest.Name
			}
			if adapter.legacy != nil {
				return adapter.legacy.Name
			}
			return ""
		}

		getPrice := func(adapter *ProductAdapter) float64 {
			if adapter.latest != nil {
				return adapter.latest.Price
			}
			if adapter.legacy != nil {
				return adapter.legacy.Price
			}
			return 0
		}

		getID := func(adapter *ProductAdapter) string {
			if adapter.latest != nil {
				return adapter.latest.ID
			}
			if adapter.legacy != nil {
				return adapter.legacy.ID
			}
			return ""
		}

		// 어댑터 테스트
		adapter := NewProductAdapter()

		// Legacy 데이터로 테스트
		legacyProduct := &LegacyProduct{
			BaseModel: BaseModel{
				ID:             "legacy_product",
				CollectionName: "products",
			},
			Name:  "Legacy Product",
			Price: 99.99,
		}

		setLegacy(adapter, legacyProduct)

		if getName(adapter) != "Legacy Product" {
			t.Errorf("어댑터 Legacy Name 실패: %s", getName(adapter))
		}

		if getPrice(adapter) != 99.99 {
			t.Errorf("어댑터 Legacy Price 실패: %f", getPrice(adapter))
		}

		if getID(adapter) != "legacy_product" {
			t.Errorf("어댑터 Legacy ID 실패: %s", getID(adapter))
		}

		// Latest 데이터로 테스트
		latestProduct := &LatestProduct{
			BaseModel: BaseModel{
				ID:             "latest_product",
				CollectionName: "products",
			},
			Name:  "Latest Product",
			Price: 149.99,
		}

		adapter = NewProductAdapter() // 새 어댑터 생성
		setLatest(adapter, latestProduct)

		if getName(adapter) != "Latest Product" {
			t.Errorf("어댑터 Latest Name 실패: %s", getName(adapter))
		}

		if getPrice(adapter) != 149.99 {
			t.Errorf("어댑터 Latest Price 실패: %f", getPrice(adapter))
		}

		if getID(adapter) != "latest_product" {
			t.Errorf("어댑터 Latest ID 실패: %s", getID(adapter))
		}
	})
}

// TestMigrationUtilities는 마이그레이션을 위한 유틸리티 함수들을 테스트합니다
func TestMigrationUtilities(t *testing.T) {
	t.Run("구조체 변환 유틸리티", func(t *testing.T) {
		// Legacy 구조체
		type LegacyItem struct {
			BaseModel
			BaseDateTime
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		// Latest 구조체
		type LatestItem struct {
			BaseModel
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		// 변환 유틸리티 함수
		convertLegacyToLatest := func(legacy LegacyItem) LatestItem {
			return LatestItem{
				BaseModel:   legacy.BaseModel,
				Title:       legacy.Title,
				Description: legacy.Description,
			}
		}

		convertLatestToLegacy := func(latest LatestItem) LegacyItem {
			return LegacyItem{
				BaseModel:    latest.BaseModel,
				BaseDateTime: BaseDateTime{}, // 빈 타임스탬프
				Title:        latest.Title,
				Description:  latest.Description,
			}
		}

		// 테스트 데이터
		legacyItem := LegacyItem{
			BaseModel: BaseModel{
				ID:             "item123",
				CollectionName: "items",
			},
			BaseDateTime: BaseDateTime{
				Created: types.DateTime{},
				Updated: types.DateTime{},
			},
			Title:       "Test Item",
			Description: "Test Description",
		}

		// Legacy -> Latest 변환 테스트
		latestItem := convertLegacyToLatest(legacyItem)

		if latestItem.ID != legacyItem.ID {
			t.Errorf("Legacy->Latest ID 변환 실패: %s != %s", latestItem.ID, legacyItem.ID)
		}

		if latestItem.Title != legacyItem.Title {
			t.Errorf("Legacy->Latest Title 변환 실패: %s != %s", latestItem.Title, legacyItem.Title)
		}

		// Latest -> Legacy 변환 테스트
		backToLegacy := convertLatestToLegacy(latestItem)

		if backToLegacy.ID != latestItem.ID {
			t.Errorf("Latest->Legacy ID 변환 실패: %s != %s", backToLegacy.ID, latestItem.ID)
		}

		if backToLegacy.Title != latestItem.Title {
			t.Errorf("Latest->Legacy Title 변환 실패: %s != %s", backToLegacy.Title, latestItem.Title)
		}
	})

	t.Run("JSON 기반 변환 유틸리티", func(t *testing.T) {
		// 범용 JSON 변환 함수
		convertViaJSON := func(source any, target any) error {
			jsonData, err := json.Marshal(source)
			if err != nil {
				return err
			}
			return json.Unmarshal(jsonData, target)
		}

		// Record -> 구조체 변환
		record := &Record{}
		record.CollectionName = "test"
		record.Set("name", "JSON Test")
		record.Set("value", 42)
		record.ID = "json_test"

		type TestStruct struct {
			BaseModel
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		var testStruct TestStruct
		err := convertViaJSON(record, &testStruct)
		if err != nil {
			t.Fatalf("JSON 변환 실패: %v", err)
		}

		testStruct.BaseModel = record.BaseModel

		if testStruct.Name != "JSON Test" {
			t.Errorf("JSON 변환 Name 실패: %s", testStruct.Name)
		}

		if testStruct.Value != 42 {
			t.Errorf("JSON 변환 Value 실패: %d", testStruct.Value)
		}

		if testStruct.ID != "json_test" {
			t.Errorf("JSON 변환 ID 실패: %s", testStruct.ID)
		}
	})

	t.Run("배치 마이그레이션 유틸리티", func(t *testing.T) {
		// Legacy 구조체 슬라이스
		type LegacyUser struct {
			BaseModel
			BaseDateTime
			Name string `json:"name"`
		}

		// Latest 구조체 슬라이스
		type LatestUser struct {
			BaseModel
			Name string `json:"name"`
		}

		// 배치 변환 함수
		batchConvert := func(legacyUsers []LegacyUser) []LatestUser {
			var latestUsers []LatestUser
			for _, legacy := range legacyUsers {
				latest := LatestUser{
					BaseModel: legacy.BaseModel,
					Name:      legacy.Name,
				}
				latestUsers = append(latestUsers, latest)
			}
			return latestUsers
		}

		// 테스트 데이터
		legacyUsers := []LegacyUser{
			{
				BaseModel: BaseModel{ID: "user1", CollectionName: "users"},
				Name:      "User 1",
			},
			{
				BaseModel: BaseModel{ID: "user2", CollectionName: "users"},
				Name:      "User 2",
			},
			{
				BaseModel: BaseModel{ID: "user3", CollectionName: "users"},
				Name:      "User 3",
			},
		}

		// 배치 변환 실행
		latestUsers := batchConvert(legacyUsers)

		// 결과 검증
		if len(latestUsers) != len(legacyUsers) {
			t.Errorf("배치 변환 개수 불일치: %d != %d", len(latestUsers), len(legacyUsers))
		}

		for i, latest := range latestUsers {
			legacy := legacyUsers[i]
			if latest.ID != legacy.ID {
				t.Errorf("배치 변환 ID[%d] 불일치: %s != %s", i, latest.ID, legacy.ID)
			}
			if latest.Name != legacy.Name {
				t.Errorf("배치 변환 Name[%d] 불일치: %s != %s", i, latest.Name, legacy.Name)
			}
		}
	})
}
