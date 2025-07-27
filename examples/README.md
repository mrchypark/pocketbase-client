# PocketBase Go 클라이언트 예제

이 디렉토리에는 PocketBase Go 클라이언트의 다양한 사용법을 보여주는 예제들이 포함되어 있습니다.

## 📚 예제 목록

### 기본 사용법
- **[quick_start](quick_start/)** - 빠른 시작 가이드
- **[basic_crud](basic_crud/)** - 타입 안전한 CRUD 작업 (권장)
- **[record_direct](record_direct/)** - Record 객체 직접 사용

### 고급 기능
- **[auth](auth/)** - 인증 및 사용자 관리
- **[batch](batch/)** - 배치 작업
- **[file_management](file_management/)** - 파일 업로드/다운로드
- **[list_options](list_options/)** - 고급 조회 옵션
- **[realtime_subscriptions](realtime_subscriptions/)** - 실시간 구독
- **[realtime_chat](realtime_chat/)** - 실시간 채팅
- **[streaming_api](streaming_api/)** - 스트리밍 API
- **[type_safe_generator](type_safe_generator/)** - 타입 안전 코드 생성

## 🔄 Record 사용 방식 비교

### 1. 타입 안전한 구조체 사용 (권장)
```go
// 구조체 정의
type Post struct {
    pocketbase.BaseModel
    Title   string `json:"title"`
    Content string `json:"content"`
}

// 서비스 생성
postsService := pocketbase.NewRecordService[Post](client, "posts")

// 사용
post := &Post{Title: "제목", Content: "내용"}
created, err := postsService.Create(ctx, post)
fmt.Println(created.Title) // 타입 안전한 접근
```

**장점:**
- ✅ 컴파일 타임 타입 검사
- ✅ IDE 자동완성 지원
- ✅ 리팩토링 안전성
- ✅ 명확한 데이터 구조

**단점:**
- ❌ 사전에 구조체 정의 필요
- ❌ 동적 스키마 처리 어려움

### 2. Record 객체 직접 사용
```go
// 서비스 생성
recordsService := client.Records("posts")

// 사용
record := &pocketbase.Record{}
record.Set("title", "제목")
record.Set("content", "내용")
created, err := recordsService.Create(ctx, record)
fmt.Println(created.GetString("title")) // 런타임 타입 변환
```

**장점:**
- ✅ 동적 스키마 지원
- ✅ 빠른 프로토타이핑
- ✅ 런타임 필드 결정 가능
- ✅ 구조체 정의 불필요

**단점:**
- ❌ 런타임 타입 오류 가능
- ❌ IDE 지원 제한적
- ❌ 오타로 인한 버그 위험

## 🎯 언제 어떤 방식을 사용할까요?

### 타입 안전한 구조체 사용 시기:
- 프로덕션 애플리케이션
- 복잡한 비즈니스 로직
- 팀 개발 프로젝트
- 장기 유지보수가 필요한 코드

### Record 직접 사용 시기:
- 프로토타이핑
- 동적 스키마 처리
- 간단한 스크립트
- 스키마가 자주 변경되는 개발 초기

## 🚀 시작하기

1. **초보자**: [quick_start](quick_start/) 예제부터 시작
2. **일반적인 사용**: [basic_crud](basic_crud/) 예제 참고
3. **동적 처리 필요**: [record_direct](record_direct/) 예제 참고

## 📋 공통 설정

모든 예제를 실행하기 전에:

```bash
# PocketBase 서버 실행
make pb_run

# 환경 변수 설정
export POCKETBASE_URL="http://127.0.0.1:8090"
```