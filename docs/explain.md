# 🔍 pocketbase-client 프로젝트 분석 보고서
## 📊 프로젝트 개요
| 항목 | 내용 |
|------|------|
| **모듈** | `github.com/mrchypark/pocketbase-client` |
| **Go 버전** | 1.24 |
| **테스트 상태** | ✅ All Passing |
| **파일 수** | 79개 Go 파일 |
| **아키텍처** | Flat structure + internal packages |
이 프로젝트는 **PocketBase API를 위한 타입-세이프 Go 클라이언트**로, 동적 레코드 접근과 제네릭 기반 타입 안전성을 모두 지원합니다.
---
🏗️ 전체 아키텍처
┌─────────────────────────────────────────────────────────────────┐
│                        사용자 코드                               │
└─────────────────────────────┬───────────────────────────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          ▼                   ▼                   ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  Dynamic API    │ │  Generic API    │ │  Generated API  │
│ client.Records  │ │ TypedRecord     │ │ models.NewXXX   │
│ (map[string]any)│ │ Service[T]      │ │ Service()       │
└────────┬────────┘ └────────┬────────┘ └────────┬────────┘
         │                   │                   │
         └───────────────────┴───────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │     Client        │
                    │  ┌─────────────┐  │
                    │  │ authInjector│──┼──▶ Auto Token Injection
                    │  └─────────────┘  │
                    └─────────┬─────────┘
                              │
                    ┌─────────▼─────────┐
                    │   PocketBase API  │
                    │  (REST + SSE)     │
                    └───────────────────┘
---
🧩 핵심 컴포넌트
1. Client (client.go)
중앙 허브로서 모든 서브서비스를 초기화하고 HTTP 요청을 조정합니다.
type Client struct {
    BaseURL    string
    HTTPClient *http.Client
    AuthStore  AuthStrategy      // 인증 전략 (Password, Token, Custom)
    Records    RecordServiceAPI  // CRUD 작업
    Realtime   RealtimeServiceAPI // SSE 구독
    Files      FileServiceAPI    // 파일 업/다운로드
    // ... 기타 서비스들
}
핵심 특징:
- authInjector RoundTripper가 모든 요청에 자동으로 Authorization 헤더 주입
- Thread-safe 설계 (sync.RWMutex)
- goccy/go-json 사용으로 고성능 JSON 처리
2. Record & Models (models.go)
PocketBase 레코드의 Go 표현:
type Record struct {
    ID             string               `json:"id"`
    CollectionID   string               `json:"collectionId"`
    CollectionName string               `json:"collectionName"`
    Expand         map[string][]*Record `json:"expand,omitempty"`
    deserializedData map[string]any     // 동적 필드 저장
}
인터페이스 계층:
| 인터페이스 | 역할 |
|-----------|------|
| BaseModel | GetID(), GetCollectionName() |
| RecordModel | BaseModel + SetID(), SetCollectionID(), SetCollectionName() |
| Mappable | ToMap() map[string]any - PATCH 시맨틱스 지원 |
3. Record Services (records.go, generic_client.go)
3가지 레벨의 타입 안전성:
| 레벨 | 서비스 | 타입 안전성 | 사용 사례 |
|------|--------|------------|----------|
| Dynamic | RecordService | ❌ | 스키마 모를 때 |
| Generic | TypedRecordService[T] | ✅ | 수동 정의 struct |
| Generated | Service[T] + pbc-gen | ✅✅ | 완전 자동화 |
---
⚙️ 코드 생성 시스템 (Deep Dive)
생성 파이프라인
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ pb_schema    │────▶│   parser.go  │────▶│  mapper.go   │────▶│ template.tpl │
│   .json      │     │ LoadSchema() │     │ Type Mapping │     │  Go Output   │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                             │                    │
                     ┌───────▼────────┐   ┌───────▼────────┐
                     │  schema.go     │   │  enum.go       │
                     │ (Compatibility)│   │  relation.go   │
                     │                │   │  file.go       │
                     └────────────────┘   └────────────────┘
핵심 파일 분석
1. cmd/pbc-gen/main.go - CLI 엔트리포인트
// CLI 플래그
-schema    // 입력 스키마 JSON 경로
-path      // 출력 Go 파일 경로
-pkgname   // 생성될 패키지명
-jsonlib   // JSON 라이브러리 (기본: encoding/json)
-enums     // Enum 상수 생성 (기본: true)
-relations // Relation 타입 생성 (기본: true)
-files     // File 타입 생성 (기본: true)
처리 흐름:
1. generator.LoadSchema() → 스키마 파싱
2. 각 컬렉션/필드 순회 → TemplateData 구성
3. Enhanced 기능 활성화 시 → EnumGenerator, RelationGenerator, FileGenerator 실행
4. text/template 실행 → Go 코드 생성
5. golang.org/x/tools/imports → 자동 포맷팅 및 import 정리
2. internal/generator/schema.go - 스키마 호환성
type CollectionSchema struct {
    Name   string        `json:"name"`
    Type   string        `json:"type"`
    Schema []FieldSchema `json:"schema"` // Legacy (PB < 0.22)
    Fields []FieldSchema `json:"fields"` // Modern (PB >= 0.22)
}
핵심: 커스텀 UnmarshalJSON
- schema 또는 fields 배열 모두 처리 (하위 호환성)
- maxSelect가 필드 레벨 또는 options 내부에 있는 경우 모두 처리
3. internal/generator/mapper.go - 타입 매핑
| PocketBase Type | Go Type (Required) | Go Type (Optional) |
|-----------------|-------------------|-------------------|
| text, email, url, editor | string | *string |
| number | float64 | *float64 |
| bool | bool | *bool |
| date, autodate | pocketbase.DateTime | *pocketbase.DateTime |
| json | json.RawMessage | json.RawMessage |
| relation (single) | string | *string |
| relation (multi) | []string | []string |
| file (single) | string | *string |
| file (multi) | []string | []string |
| select (single) | string | *string |
| select (multi) | []string | []string |
Multi 판단 로직:
isMulti := false
if field.Options != nil && field.Options.MaxSelect != nil {
    if *field.Options.MaxSelect != 1 {
        isMulti = true
    }
} else if field.Type == "relation" || field.Type == "file" || field.Type == "select" {
    // MaxSelect가 nil이면 multi로 간주 (기본값)
    isMulti = true
}
4. template.go.tpl - 생성 템플릿
생성되는 코드 구조:
// 1. Enum 상수 (--enums=true)
const (
    DeviceTypeM2 = "m2"
    DeviceTypeM3 = "m3"
)
func DeviceTypeValues() []string { ... }
func IsValidDeviceType(value string) bool { ... }
// 2. Relation 타입 (--relations=true)
type PlantRelation struct { id string }
func (r PlantRelation) Load(ctx, client) (*Plant, error) { ... }
type PlantRelations []PlantRelation
func (r PlantRelations) LoadAll(ctx, client) ([]*Plant, error) { ... }
// 3. File 타입 (--files=true)
type FileReference struct { filename, recordID, collection, fieldName string }
func (f FileReference) URL(baseURL string) string { ... }
func (f FileReference) ThumbURL(baseURL, thumb string) string { ... }
// 4. 컬렉션 Struct
type Posts struct {
    ID             string             `json:"id"`
    CollectionID   string             `json:"collectionId"`
    CollectionName string             `json:"collectionName"`
    Created        pocketbase.DateTime `json:"created"`
    Updated        pocketbase.DateTime `json:"updated"`
    Title          string             `json:"title"`
    Content        *string        `json:"content,omitempty"`
}
// 5. RecordModel 인터페이스 구현
func (m *Posts) GetID() string { ... }
func (m *Posts) SetID(id string) { ... }
// ... 등등
// 6. Setter 메서드
func (m *Posts) SetTitle(v string) { m.Title = v }
func (m *Posts) SetContent(v string) { m.Content = &v }
// 7. ToMap (PATCH 시맨틱스)
func (m *Posts) ToMap() map[string]any {
    data := make(map[string]any)
    if m.Content != nil {
        data["content"] = m.Content
    }
    // ...
    return data
}
// 8. 타입 안전 서비스 생성자
func NewPostsService(client *pocketbase.Client) *pocketbase.TypedRecordService[Posts] {
    return pocketbase.NewTypedRecordService[Posts](client, "posts")
}
// 9. 헬퍼 함수
func GetPosts(client, id, opts) (*Posts, error) { ... }
func GetPostsList(client, opts) (*PostsCollection, error) { ... }
Enhanced Generators
enum.go - Select 필드 → Enum 상수
type EnumData struct {
    CollectionName string
    FieldName      string
    EnumTypeName   string         // 예: "DeviceTypeType"
    Constants      []ConstantData // [{Name: "DeviceTypeM2", Value: "m2"}, ...]
}
relation.go - Relation 필드 → 관계 타입
type RelationTypeData struct {
    TypeName         string // 예: "PlantRelation"
    TargetCollection string // 예: "plants"
    TargetTypeName   string // 예: "Plant"
    IsMulti          bool   // maxSelect > 1 여부
}
file.go - File 필드 → 파일 참조 타입
type FileTypeData struct {
    TypeName       string   // 예: "ImageFile"
    IsMulti        bool
    HasThumbnails  bool
    ThumbnailSizes []string // 예: ["100x100", "200x200"]
}
---
📈 에러 처리 시스템
internal/generator/errors.go에서 구조화된 에러 시스템 제공:
type GenerationError struct {
    Type    ErrorType        // schema_load, template_parse, etc.
    Message string
    Details map[string]any   // 추가 컨텍스트
    Cause   error
}
에러 타입:
- ErrorTypeSchemaLoad, ErrorTypeSchemaParse, ErrorTypeSchemaValidate
- ErrorTypeTemplateParse, ErrorTypeTemplateExecute
- ErrorTypeFileCreate, ErrorTypeFileWrite, ErrorTypeFileRead
- ErrorTypeCodeFormat, ErrorTypeTypeMapping, ErrorTypeNameConflict
---
🚀 사용 예시
1. 스키마에서 코드 생성
# 스키마 추출
curl http://localhost:8090/api/collections > schema.json
# 코드 생성
pbc-gen -schema schema.json -path models/generated.go -pkgname models
2. 생성된 코드 사용
client := pocketbase.NewClient("http://localhost:8090")
client.WithAdminPassword(ctx, "admin@example.com", "password")
// 타입 안전 서비스 사용
posts := models.NewPostsService(client)
// Create
post := models.NewPosts()
post.SetTitle("Hello World")
created, err := posts.Create(ctx, post, nil)
// Read
one, err := posts.GetOne(ctx, "RECORD_ID", nil)
fmt.Println(one.Title)
// List with filtering
list, err := posts.GetList(ctx, &pocketbase.ListOptions{
    Filter: "published = true",
    Sort:   "-created",
})
---
📋 프로젝트 건강 상태
| 메트릭 | 상태 |
|--------|------|
| 테스트 | ✅ All Passing |
| 코드 스타일 | 일관됨 (go fmt 준수) |
| 문서화 | README + 예제 디렉토리 완비 |
| 에러 핸들링 | 구조화된 에러 타입 |
| 하위 호환성 | Legacy/Modern PocketBase 스키마 모두 지원 |
---
🔑 핵심 설계 결정
1. Interceptor 패턴: authInjector RoundTripper로 인증 토큰 자동 주입
2. Marshal-Unmarshal 전략: 생성된 모델에서 json.Marshal → json.Unmarshal로 데이터 하이드레이션
3. ToMap() PATCH 시맨틱스: omitempty 필드는 nil일 때 전송 안 함 (부분 업데이트 지원)
4. Factory 함수: NewXXX() 패턴으로 CollectionName 자동 설정
5. 다중 API 레벨: Dynamic → Generic → Generated로 점진적 타입 안전성 제공
