# 기본 CRUD 작업 예제 (타입 안전)

이 예제는 PocketBase Go 클라이언트에서 사용자 정의 구조체를 사용한 타입 안전한 CRUD 작업을 보여줍니다.

## 주요 기능

### 1. 타입 안전한 구조체 정의
```go
type Post struct {
    pocketbase.BaseModel
    Title   string `json:"title"`
    Content string `json:"content"`
}
```

### 2. 제네릭 레코드 서비스 사용
```go
postsService := pocketbase.NewRecordService[Post](client, "posts")
```

### 3. 타입 안전한 CRUD 작업
- **생성**: `postsService.Create(ctx, &post)`
- **조회**: `postsService.GetList(ctx, options)`
- **단일 조회**: `postsService.GetOne(ctx, id, options)`
- **업데이트**: `postsService.Update(ctx, id, &post)`
- **삭제**: `postsService.Delete(ctx, id)`

## 장점

### 타입 안전성
- 컴파일 타임에 타입 오류 검출
- IDE의 자동완성 지원
- 리팩토링 시 안전성 보장

### 코드 가독성
- 명확한 데이터 구조
- 직관적인 필드 접근 (`post.Title`)
- 비즈니스 로직과 데이터 모델의 분리

## 실행 방법

```bash
# 환경 변수 설정
export POCKETBASE_URL="http://127.0.0.1:8090"

# 예제 실행
go run examples/basic_crud/main.go
```

## 필요한 PocketBase 설정

1. `posts` 컬렉션 생성
2. 다음 필드 추가:
   - `title` (text)
   - `content` (text)
3. 관리자 계정 생성 (`admin@example.com` / `password123`)

## 관련 예제

- [Record 직접 사용](../record_direct/) - 동적 스키마를 위한 Record 객체 직접 사용
- [타입 안전 생성기](../type_safe_generator/) - 스키마에서 자동으로 타입 생성