# 페이지네이션 헬퍼 사용법 예제

이 예제는 PocketBase Go 클라이언트의 페이지네이션 헬퍼 기능 사용법을 보여줍니다.

## 실행 방법

```bash
# PocketBase 서버 실행 (다른 터미널에서)
make pb_run

# 예제 실행
go run examples/pagination_helpers/main.go
```

## 주요 기능

### 1. GetAll 메서드
모든 페이지의 레코드를 자동으로 가져오는 가장 간단한 방법입니다.

```go
// 기본 사용법
records, err := client.Records.GetAll(ctx, "posts", nil)

// 옵션과 함께 사용
options := &pocketbase.ListOptions{
    Filter: "status = 'published'",
    Sort:   "-created",
}
records, err := client.Records.GetAll(ctx, "posts", options)
```

### 2. GetAllWithBatchSize 메서드
배치 크기를 지정하여 메모리 사용량과 네트워크 요청 횟수를 조절할 수 있습니다.

```go
// 배치 크기 50으로 설정
records, err := client.Records.GetAllWithBatchSize(ctx, "posts", nil, 50)
```

### 3. Iterator 패턴
대용량 데이터를 메모리 효율적으로 처리할 때 사용합니다.

```go
iterator := client.Records.Iterate(ctx, "posts", nil)

for iterator.Next() {
    record := iterator.Record()
    // 레코드 처리 로직
}

if err := iterator.Error(); err != nil {
    // 에러 처리
}
```

### 4. Iterator with BatchSize
Iterator에서도 배치 크기를 지정할 수 있습니다.

```go
iterator := client.Records.IterateWithBatchSize(ctx, "posts", nil, 25)
```

## 에러 처리

페이지네이션 중 에러가 발생하면 `PaginationError` 타입으로 반환됩니다.

```go
records, err := client.Records.GetAll(ctx, "posts", nil)
if err != nil {
    if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
        // 부분 데이터 확인
        partialData := paginationErr.GetPartialData()
        if len(partialData) > 0 {
            // 부분 데이터 사용
        }
    }
}
```

## 성능 고려사항

### GetAll vs Iterator 선택 기준

- **GetAll 사용 시기:**
  - 전체 데이터가 메모리에 들어갈 수 있는 크기
  - 데이터를 여러 번 순회해야 하는 경우
  - 간단한 일회성 처리

- **Iterator 사용 시기:**
  - 대용량 데이터 처리
  - 메모리 사용량을 최소화해야 하는 경우
  - 스트리밍 방식의 처리가 필요한 경우

### 배치 크기 설정 가이드

- **작은 배치 크기 (10-50):** 메모리 사용량 최소화, 네트워크 요청 증가
- **중간 배치 크기 (50-200):** 균형잡힌 성능 (권장)
- **큰 배치 크기 (200-1000):** 네트워크 요청 최소화, 메모리 사용량 증가

## 컨텍스트 사용

모든 페이지네이션 메서드는 컨텍스트를 지원합니다.

```go
// 타임아웃 설정
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

records, err := client.Records.GetAll(ctx, "posts", nil)

// 취소 가능한 Iterator
ctx, cancel := context.WithCancel(context.Background())
iterator := client.Records.Iterate(ctx, "posts", nil)

// 필요시 취소
go func() {
    time.Sleep(5 * time.Second)
    cancel() // Iterator 중단
}()
```

## 주의사항

1. **메모리 사용량:** GetAll은 모든 데이터를 메모리에 로드하므로 대용량 데이터 시 주의
2. **네트워크 비용:** 작은 배치 크기는 네트워크 요청 횟수를 증가시킴
3. **에러 복구:** 부분 데이터를 활용하여 에러 상황에서도 가능한 데이터 보존
4. **컨텍스트 취소:** 장시간 실행되는 작업은 적절한 타임아웃 설정 권장