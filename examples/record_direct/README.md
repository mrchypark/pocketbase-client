# Record 객체 직접 사용 예제

이 예제는 PocketBase Go 클라이언트에서 `Record` 객체를 직접 사용하는 방법을 보여줍니다. 사용자 정의 구조체를 만들지 않고도 동적으로 레코드를 생성하고 조작할 수 있습니다.

## 주요 기능

### 1. Record 객체 생성 및 데이터 설정
```go
newRecord := &pocketbase.Record{}
newRecord.Set("title", "Direct Record Post")
newRecord.Set("content", "Record 객체를 직접 사용한 예제입니다!")
```

### 2. 다양한 데이터 타입 지원
```go
record.Set("title", "문자열")
record.Set("is_published", true)           // 불린
record.Set("view_count", 42)               // 숫자
record.Set("rating", 4.5)                  // 실수
record.Set("tags", []string{"tag1", "tag2"}) // 문자열 배열
```

### 3. 데이터 조회 메서드
- `GetString(key)` - 문자열 값 조회
- `GetBool(key)` - 불린 값 조회  
- `GetFloat(key)` - 실수 값 조회
- `GetStringSlice(key)` - 문자열 배열 조회
- `GetDateTime(key)` - 날짜/시간 값 조회
- `Get(key)` - 원시 값 조회

### 4. 고급 조회 옵션
- **필터링**: `Filter: "title ~ '검색어'"` 
- **정렬**: `Sort: "-created"` (최신순)
- **필드 선택**: `Fields: "id,title,created"`
- **관계 확장**: `Expand: "user"`

## 언제 사용하나요?

### Record 객체 직접 사용이 적합한 경우:
- 동적인 스키마를 다룰 때
- 프로토타이핑이나 빠른 개발 시
- 런타임에 필드가 결정되는 경우
- 간단한 CRUD 작업

### 사용자 정의 구조체가 적합한 경우:
- 타입 안전성이 중요한 경우
- 복잡한 비즈니스 로직이 있는 경우
- IDE의 자동완성과 타입 체크를 활용하고 싶은 경우
- 대규모 애플리케이션 개발

## 실행 방법

```bash
# 환경 변수 설정
export POCKETBASE_URL="http://127.0.0.1:8090"

# 예제 실행
go run examples/record_direct/main.go
```

## 필요한 PocketBase 설정

이 예제를 실행하기 전에 PocketBase에서 다음을 설정해야 합니다:

1. `posts` 컬렉션 생성
2. 다음 필드 추가:
   - `title` (text)
   - `content` (text)  
   - `is_published` (bool, optional)
   - `view_count` (number, optional)
   - `rating` (number, optional)
   - `tags` (json, optional)
3. 관리자 계정 생성 (`admin@example.com` / `password123`)