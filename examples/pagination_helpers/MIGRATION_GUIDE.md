# 페이지네이션 헬퍼 마이그레이션 가이드

기존 수동 페이지네이션 코드를 새로운 페이지네이션 헬퍼로 마이그레이션하는 방법을 안내합니다.

## 목차

1. [기본 마이그레이션](#기본-마이그레이션)
2. [성능 비교](#성능-비교)
3. [메모리 사용량 최적화](#메모리-사용량-최적화)
4. [에러 처리 개선](#에러-처리-개선)
5. [권장사항](#권장사항)

## 기본 마이그레이션

### Before: 수동 페이지네이션

```go
// 기존 방식 - 수동으로 모든 페이지 순회
func getAllRecordsManually(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    var allRecords []*pocketbase.Record
    page := 1
    perPage := 100

    for {
        result, err := client.Records.GetList(ctx, collection, &pocketbase.ListOptions{
            Page:    page,
            PerPage: perPage,
        })
        if err != nil {
            return allRecords, err // 에러 시 부분 데이터 손실
        }

        allRecords = append(allRecords, result.Items...)

        // 마지막 페이지 확인
        if page >= result.TotalPages || len(result.Items) == 0 {
            break
        }

        page++
    }

    return allRecords, nil
}
```

### After: GetAll 헬퍼 사용

```go
// 새로운 방식 - GetAll 헬퍼 사용
func getAllRecordsWithHelper(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    return client.Records.GetAll(ctx, collection, nil)
}

// 배치 크기 지정 (기존 perPage와 동일한 효과)
func getAllRecordsWithBatchSize(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    return client.Records.GetAllWithBatchSize(ctx, collection, nil, 100)
}
```

## 대용량 데이터 처리 마이그레이션

### Before: 수동 배치 처리

```go
// 기존 방식 - 메모리 부족 위험
func processLargeDatasetManually(ctx context.Context, client *pocketbase.Client) error {
    var allRecords []*pocketbase.Record
    page := 1
    perPage := 100

    for {
        result, err := client.Records.GetList(ctx, "large_collection", &pocketbase.ListOptions{
            Page:    page,
            PerPage: perPage,
        })
        if err != nil {
            return err
        }

        // 모든 데이터를 메모리에 누적 - 메모리 부족 위험
        allRecords = append(allRecords, result.Items...)

        if page >= result.TotalPages {
            break
        }
        page++
    }

    // 모든 데이터를 한 번에 처리
    return processRecords(allRecords)
}
```

### After: Iterator 패턴 사용

```go
// 새로운 방식 - 메모리 효율적 처리
func processLargeDatasetWithIterator(ctx context.Context, client *pocketbase.Client) error {
    iterator := client.Records.IterateWithBatchSize(ctx, "large_collection", nil, 100)

    for iterator.Next() {
        record := iterator.Record()
        
        // 레코드를 하나씩 처리 - 메모리 효율적
        if err := processRecord(record); err != nil {
            return err
        }
    }

    return iterator.Error()
}
```

## 필터링과 정렬 마이그레이션

### Before: 옵션이 있는 수동 페이지네이션

```go
// 기존 방식
func getFilteredRecordsManually(ctx context.Context, client *pocketbase.Client) ([]*pocketbase.Record, error) {
    var allRecords []*pocketbase.Record
    page := 1

    baseOptions := &pocketbase.ListOptions{
        Filter: "status = 'active'",
        Sort:   "-created",
        Expand: "author",
        PerPage: 50,
    }

    for {
        // 매번 새로운 옵션 객체 생성 필요
        options := &pocketbase.ListOptions{
            Page:        page,
            PerPage:     baseOptions.PerPage,
            Filter:      baseOptions.Filter,
            Sort:        baseOptions.Sort,
            Expand:      baseOptions.Expand,
            QueryParams: make(map[string]string),
        }
        
        // QueryParams 복사
        for k, v := range baseOptions.QueryParams {
            options.QueryParams[k] = v
        }

        result, err := client.Records.GetList(ctx, "posts", options)
        if err != nil {
            return allRecords, err
        }

        allRecords = append(allRecords, result.Items...)

        if page >= result.TotalPages {
            break
        }
        page++
    }

    return allRecords, nil
}
```

### After: 헬퍼 사용 (자동 옵션 복사)

```go
// 새로운 방식 - 옵션 자동 복사
func getFilteredRecordsWithHelper(ctx context.Context, client *pocketbase.Client) ([]*pocketbase.Record, error) {
    options := &pocketbase.ListOptions{
        Filter: "status = 'active'",
        Sort:   "-created",
        Expand: "author",
    }

    // 헬퍼가 자동으로 옵션을 복사하고 페이지네이션 처리
    return client.Records.GetAllWithBatchSize(ctx, "posts", options, 50)
}
```

## 에러 처리 개선

### Before: 에러 시 데이터 손실

```go
// 기존 방식 - 에러 발생 시 모든 데이터 손실
func getRecordsWithBasicErrorHandling(ctx context.Context, client *pocketbase.Client) ([]*pocketbase.Record, error) {
    var allRecords []*pocketbase.Record
    page := 1

    for {
        result, err := client.Records.GetList(ctx, "posts", &pocketbase.ListOptions{
            Page:    page,
            PerPage: 100,
        })
        if err != nil {
            // 에러 발생 시 이미 수집된 데이터 손실
            return nil, err
        }

        allRecords = append(allRecords, result.Items...)

        if page >= result.TotalPages {
            break
        }
        page++
    }

    return allRecords, nil
}
```

### After: 부분 데이터 보존

```go
// 새로운 방식 - 에러 시에도 부분 데이터 보존
func getRecordsWithImprovedErrorHandling(ctx context.Context, client *pocketbase.Client) ([]*pocketbase.Record, error) {
    records, err := client.Records.GetAll(ctx, "posts", nil)
    if err != nil {
        // PaginationError 확인
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            partialData := paginationErr.GetPartialData()
            if len(partialData) > 0 {
                log.Printf("경고: 페이지네이션 중 에러 발생, %d개의 부분 데이터 반환", len(partialData))
                return partialData, nil // 부분 데이터라도 반환
            }
        }
        return nil, err
    }
    return records, nil
}
```

## 성능 비교

### 코드 복잡성

| 항목 | 수동 페이지네이션 | 헬퍼 사용 |
|------|------------------|-----------|
| 코드 라인 수 | 20-30줄 | 1-3줄 |
| 에러 처리 복잡성 | 높음 | 낮음 |
| 옵션 복사 로직 | 수동 구현 필요 | 자동 처리 |
| 메모리 관리 | 수동 관리 | 자동 최적화 |

### 성능 특성

| 기능 | 수동 구현 | GetAll | Iterator |
|------|-----------|--------|----------|
| 메모리 사용량 | 높음 | 높음 | 낮음 |
| 네트워크 효율성 | 보통 | 높음 | 높음 |
| 에러 복구 | 없음 | 부분 데이터 보존 | 부분 데이터 보존 |
| 재시도 로직 | 없음 | 자동 재시도 | 자동 재시도 |
| 컨텍스트 취소 | 수동 구현 | 자동 지원 | 자동 지원 |

### 벤치마크 예제

```go
// 성능 비교 테스트
func BenchmarkManualPagination(b *testing.B) {
    client := setupTestClient()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := getAllRecordsManually(ctx, client, "test_collection")
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGetAllHelper(b *testing.B) {
    client := setupTestClient()
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := client.Records.GetAll(ctx, "test_collection", nil)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 메모리 사용량 최적화

### 시나리오별 권장사항

#### 1. 소규모 데이터 (< 1,000개)
```go
// GetAll 사용 권장
records, err := client.Records.GetAll(ctx, "small_collection", nil)
```

#### 2. 중간 규모 데이터 (1,000 - 10,000개)
```go
// 배치 크기 조절하여 GetAll 사용
records, err := client.Records.GetAllWithBatchSize(ctx, "medium_collection", nil, 200)
```

#### 3. 대규모 데이터 (> 10,000개)
```go
// Iterator 패턴 사용
iterator := client.Records.IterateWithBatchSize(ctx, "large_collection", nil, 100)
for iterator.Next() {
    record := iterator.Record()
    // 즉시 처리하여 메모리 사용량 최소화
    processRecord(record)
}
```

## 권장사항

### 1. 마이그레이션 우선순위

1. **높은 우선순위:** 에러 처리가 중요한 코드
2. **중간 우선순위:** 반복적으로 사용되는 페이지네이션 코드
3. **낮은 우선순위:** 일회성 스크립트

### 2. 단계별 마이그레이션

#### Phase 1: 기본 교체
```go
// 기존 코드를 GetAll로 단순 교체
records, err := client.Records.GetAll(ctx, collection, options)
```

#### Phase 2: 배치 크기 최적화
```go
// 성능 테스트 후 적절한 배치 크기 설정
records, err := client.Records.GetAllWithBatchSize(ctx, collection, options, optimalBatchSize)
```

#### Phase 3: 대용량 데이터 최적화
```go
// 필요시 Iterator 패턴으로 전환
iterator := client.Records.IterateWithBatchSize(ctx, collection, options, batchSize)
```

### 3. 테스트 전략

```go
// 마이그레이션 후 동작 검증
func TestMigrationCompatibility(t *testing.T) {
    // 기존 방식과 새로운 방식의 결과 비교
    manualResults := getAllRecordsManually(ctx, client, "test")
    helperResults := client.Records.GetAll(ctx, "test", nil)
    
    assert.Equal(t, len(manualResults), len(helperResults))
    // 추가 검증 로직...
}
```

### 4. 점진적 도입

```go
// 기능 플래그를 사용한 점진적 도입
func getRecords(ctx context.Context, client *pocketbase.Client, useHelper bool) ([]*pocketbase.Record, error) {
    if useHelper {
        return client.Records.GetAll(ctx, "collection", nil)
    }
    return getAllRecordsManually(ctx, client, "collection")
}
```

## 주의사항

1. **메모리 사용량:** 대용량 데이터는 Iterator 사용 권장
2. **배치 크기:** 네트워크 환경에 따라 최적값 조정 필요
3. **에러 처리:** 부분 데이터 활용 로직 추가 고려
4. **테스트:** 마이그레이션 후 충분한 테스트 수행
5. **모니터링:** 성능 변화 모니터링 권장

## 마이그레이션 체크리스트

- [ ] 기존 코드의 페이지네이션 로직 식별
- [ ] 적절한 헬퍼 메서드 선택 (GetAll vs Iterator)
- [ ] 배치 크기 최적화
- [ ] 에러 처리 로직 개선
- [ ] 단위 테스트 작성/수정
- [ ] 성능 테스트 수행
- [ ] 메모리 사용량 확인
- [ ] 프로덕션 배포 전 충분한 테스트