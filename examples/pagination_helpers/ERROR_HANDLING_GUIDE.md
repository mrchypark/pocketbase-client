# 페이지네이션 헬퍼 에러 처리 가이드

페이지네이션 헬퍼의 에러 처리 방법과 부분 데이터 복구, 재시도 로직 커스터마이징에 대한 상세 가이드입니다.

## 목차

1. [PaginationError 이해하기](#paginationerror-이해하기)
2. [부분 데이터 복구](#부분-데이터-복구)
3. [재시도 로직 이해](#재시도-로직-이해)
4. [에러 타입별 처리](#에러-타입별-처리)
5. [커스텀 에러 처리](#커스텀-에러-처리)
6. [모니터링과 로깅](#모니터링과-로깅)

## PaginationError 이해하기

### PaginationError 구조

```go
type PaginationError struct {
    Operation    string        // "GetAll" 또는 "Iterate"
    Page         int          // 에러가 발생한 페이지 번호
    PartialData  []*Record    // 에러 발생 전까지 수집된 데이터
    OriginalErr  error        // 원본 에러
}

// 메서드들
func (e *PaginationError) Error() string
func (e *PaginationError) Unwrap() error
func (e *PaginationError) GetPartialData() []*Record
```

### 기본 에러 처리

```go
func handlePaginationError(ctx context.Context, client *pocketbase.Client) {
    records, err := client.Records.GetAll(ctx, "posts", nil)
    if err != nil {
        // PaginationError 타입 확인
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            fmt.Printf("페이지네이션 에러 발생:\n")
            fmt.Printf("  작업: %s\n", paginationErr.Operation)
            fmt.Printf("  실패 페이지: %d\n", paginationErr.Page)
            fmt.Printf("  수집된 데이터: %d개\n", len(paginationErr.GetPartialData()))
            fmt.Printf("  원본 에러: %v\n", paginationErr.OriginalErr)
        } else {
            fmt.Printf("일반 에러: %v\n", err)
        }
        return
    }
    
    fmt.Printf("성공: %d개 레코드 수집\n", len(records))
}
```

## 부분 데이터 복구

### 1. 기본 부분 데이터 활용

```go
func getRecordsWithPartialRecovery(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            partialData := paginationErr.GetPartialData()
            if len(partialData) > 0 {
                log.Printf("경고: 페이지네이션 불완전, %d개 부분 데이터 반환", len(partialData))
                return partialData, nil // 부분 데이터라도 반환
            }
        }
        return nil, err
    }
    return records, nil
}
```

### 2. 임계값 기반 부분 데이터 처리

```go
func getRecordsWithThreshold(ctx context.Context, client *pocketbase.Client, collection string, minRecords int) ([]*pocketbase.Record, error) {
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            partialData := paginationErr.GetPartialData()
            
            // 최소 임계값 확인
            if len(partialData) >= minRecords {
                log.Printf("임계값 충족: %d개 부분 데이터 사용 (최소: %d개)", 
                    len(partialData), minRecords)
                return partialData, nil
            } else {
                log.Printf("임계값 미달: %d개 < %d개, 에러 반환", 
                    len(partialData), minRecords)
                return nil, err
            }
        }
        return nil, err
    }
    return records, nil
}
```

### 3. 진행률 기반 처리

```go
func getRecordsWithProgressCheck(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            partialData := paginationErr.GetPartialData()
            
            // 전체 데이터 수 추정 (첫 페이지 정보 활용)
            if paginationErr.Page > 1 {
                estimatedTotal := len(partialData) * 100 / (paginationErr.Page - 1) // 대략적 추정
                progress := float64(len(partialData)) / float64(estimatedTotal) * 100
                
                if progress >= 50.0 { // 50% 이상 수집된 경우
                    log.Printf("진행률 %.1f%% - 부분 데이터 사용", progress)
                    return partialData, nil
                }
            }
        }
        return nil, err
    }
    return records, nil
}
```

## 재시도 로직 이해

### 기본 재시도 설정

```go
// 기본 재시도 설정 (내장)
var DefaultPaginationOptions = &PaginationOptions{
    BatchSize:  100,
    MaxRetries: 3,           // 최대 3회 재시도
    RetryDelay: time.Second, // 1초 지연
    Prefetch:   false,
}
```

### 재시도되는 에러 vs 재시도되지 않는 에러

```go
// 재시도되는 에러들
var retryableErrors = []error{
    ErrNetworkTimeout,
    ErrServerUnavailable,
    ErrTooManyRequests,
    ErrInternalServerError,
}

// 재시도되지 않는 에러들
var nonRetryableErrors = []error{
    ErrUnauthorized,    // 401 - 인증 실패
    ErrForbidden,       // 403 - 권한 없음
    ErrNotFound,        // 404 - 리소스 없음
    ErrBadRequest,      // 400 - 잘못된 요청
}
```

### 재시도 로직 동작 예제

```go
func demonstrateRetryLogic(ctx context.Context, client *pocketbase.Client) {
    // 네트워크 불안정한 환경에서 테스트
    records, err := client.Records.GetAll(ctx, "posts", nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            // 원본 에러 확인
            switch {
            case errors.Is(paginationErr.OriginalErr, ErrNetworkTimeout):
                fmt.Println("네트워크 타임아웃 - 재시도 후 실패")
            case errors.Is(paginationErr.OriginalErr, ErrUnauthorized):
                fmt.Println("인증 실패 - 재시도하지 않음")
            case errors.Is(paginationErr.OriginalErr, ErrTooManyRequests):
                fmt.Println("요청 한도 초과 - 재시도 후 실패")
            }
        }
    }
}
```

## 에러 타입별 처리

### 1. 네트워크 에러 처리

```go
func handleNetworkErrors(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            // 네트워크 관련 에러 확인
            if isNetworkError(paginationErr.OriginalErr) {
                log.Printf("네트워크 에러 발생, 부분 데이터 활용: %d개", 
                    len(paginationErr.GetPartialData()))
                
                // 부분 데이터가 있으면 사용
                if len(paginationErr.GetPartialData()) > 0 {
                    return paginationErr.GetPartialData(), nil
                }
                
                // 나중에 다시 시도하도록 특별한 에러 반환
                return nil, &NetworkRetryableError{
                    OriginalErr: paginationErr.OriginalErr,
                    RetryAfter:  time.Minute * 5,
                }
            }
        }
        return nil, err
    }
    return records, nil
}

func isNetworkError(err error) bool {
    // 네트워크 관련 에러 판단 로직
    return errors.Is(err, ErrNetworkTimeout) || 
           errors.Is(err, ErrConnectionRefused) ||
           errors.Is(err, ErrDNSFailure)
}
```

### 2. 인증 에러 처리

```go
func handleAuthErrors(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            if isAuthError(paginationErr.OriginalErr) {
                log.Printf("인증 에러 발생, 토큰 갱신 필요")
                
                // 토큰 갱신 시도
                if refreshErr := refreshAuthToken(ctx, client); refreshErr != nil {
                    return nil, fmt.Errorf("토큰 갱신 실패: %w", refreshErr)
                }
                
                // 토큰 갱신 후 재시도
                return client.Records.GetAll(ctx, collection, nil)
            }
        }
        return nil, err
    }
    return records, nil
}

func isAuthError(err error) bool {
    return errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrForbidden)
}

func refreshAuthToken(ctx context.Context, client *pocketbase.Client) error {
    // 토큰 갱신 로직 구현
    return client.Auth.RefreshToken(ctx)
}
```

### 3. 서버 에러 처리

```go
func handleServerErrors(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            if isServerError(paginationErr.OriginalErr) {
                log.Printf("서버 에러 발생 (페이지 %d), 부분 데이터 보존", paginationErr.Page)
                
                // 서버 에러 시에도 부분 데이터 활용
                partialData := paginationErr.GetPartialData()
                if len(partialData) > 0 {
                    // 에러 정보와 함께 부분 데이터 반환
                    return partialData, &PartialDataWarning{
                        Message:     "서버 에러로 인한 부분 데이터",
                        TotalPages:  paginationErr.Page,
                        DataCount:   len(partialData),
                        OriginalErr: paginationErr.OriginalErr,
                    }
                }
            }
        }
        return nil, err
    }
    return records, nil
}

func isServerError(err error) bool {
    return errors.Is(err, ErrInternalServerError) || 
           errors.Is(err, ErrBadGateway) ||
           errors.Is(err, ErrServiceUnavailable)
}
```

## 커스텀 에러 처리

### 1. 커스텀 에러 타입 정의

```go
// 네트워크 재시도 가능 에러
type NetworkRetryableError struct {
    OriginalErr error
    RetryAfter  time.Duration
}

func (e *NetworkRetryableError) Error() string {
    return fmt.Sprintf("네트워크 에러 (재시도 권장: %v 후): %v", e.RetryAfter, e.OriginalErr)
}

// 부분 데이터 경고
type PartialDataWarning struct {
    Message     string
    TotalPages  int
    DataCount   int
    OriginalErr error
}

func (e *PartialDataWarning) Error() string {
    return fmt.Sprintf("%s: %d페이지 중 일부, %d개 데이터", e.Message, e.TotalPages, e.DataCount)
}
```

### 2. 커스텀 재시도 로직

```go
func getRecordsWithCustomRetry(ctx context.Context, client *pocketbase.Client, collection string, maxRetries int, backoffFactor float64) ([]*pocketbase.Record, error) {
    var lastErr error
    var partialData []*pocketbase.Record
    
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            // 커스텀 백오프 계산
            delay := time.Duration(float64(time.Second) * math.Pow(backoffFactor, float64(attempt-1)))
            log.Printf("재시도 %d/%d, %v 대기 중...", attempt, maxRetries, delay)
            
            select {
            case <-ctx.Done():
                return partialData, ctx.Err()
            case <-time.After(delay):
            }
        }
        
        records, err := client.Records.GetAll(ctx, collection, nil)
        if err == nil {
            return records, nil
        }
        
        lastErr = err
        
        // PaginationError에서 부분 데이터 추출
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            if len(paginationErr.GetPartialData()) > len(partialData) {
                partialData = paginationErr.GetPartialData()
            }
            
            // 재시도하지 않을 에러인지 확인
            if !isRetryableError(paginationErr.OriginalErr) {
                break
            }
        }
    }
    
    // 모든 재시도 실패 시 부분 데이터라도 반환
    if len(partialData) > 0 {
        return partialData, &PartialDataWarning{
            Message:     "재시도 실패, 부분 데이터 반환",
            DataCount:   len(partialData),
            OriginalErr: lastErr,
        }
    }
    
    return nil, lastErr
}

func isRetryableError(err error) bool {
    // 재시도 가능한 에러 판단
    return !errors.Is(err, ErrUnauthorized) &&
           !errors.Is(err, ErrForbidden) &&
           !errors.Is(err, ErrNotFound) &&
           !errors.Is(err, ErrBadRequest)
}
```

### 3. Iterator 에러 처리

```go
func processWithIteratorErrorHandling(ctx context.Context, client *pocketbase.Client, collection string) error {
    iterator := client.Records.Iterate(ctx, collection, nil)
    
    processedCount := 0
    var lastError error
    
    for iterator.Next() {
        record := iterator.Record()
        
        // 레코드 처리
        if err := processRecord(record); err != nil {
            log.Printf("레코드 처리 실패 (ID: %s): %v", record.Id, err)
            continue
        }
        
        processedCount++
        
        // 주기적으로 진행 상황 로깅
        if processedCount%100 == 0 {
            log.Printf("진행 상황: %d개 레코드 처리 완료", processedCount)
        }
    }
    
    // Iterator 에러 확인
    if err := iterator.Error(); err != nil {
        lastError = err
        log.Printf("Iterator 에러 발생: %v", err)
        log.Printf("처리 완료된 레코드: %d개", processedCount)
        
        // 부분 처리 결과도 유의미한 경우
        if processedCount > 0 {
            return &PartialProcessingError{
                ProcessedCount: processedCount,
                OriginalErr:    err,
            }
        }
    }
    
    log.Printf("처리 완료: 총 %d개 레코드", processedCount)
    return lastError
}

type PartialProcessingError struct {
    ProcessedCount int
    OriginalErr    error
}

func (e *PartialProcessingError) Error() string {
    return fmt.Sprintf("부분 처리 완료: %d개 처리됨, 에러: %v", e.ProcessedCount, e.OriginalErr)
}
```

## 모니터링과 로깅

### 1. 구조화된 로깅

```go
import (
    "log/slog"
)

func getRecordsWithStructuredLogging(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    logger := slog.With(
        "operation", "pagination",
        "collection", collection,
    )
    
    logger.Info("페이지네이션 시작")
    
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            logger.Error("페이지네이션 실패",
                "page", paginationErr.Page,
                "partial_data_count", len(paginationErr.GetPartialData()),
                "error", paginationErr.OriginalErr,
            )
            
            // 메트릭 수집
            recordPaginationError(collection, paginationErr.Page, paginationErr.OriginalErr)
        } else {
            logger.Error("일반 에러", "error", err)
        }
        return nil, err
    }
    
    logger.Info("페이지네이션 완료", "record_count", len(records))
    recordPaginationSuccess(collection, len(records))
    
    return records, nil
}
```

### 2. 메트릭 수집

```go
// 메트릭 수집 함수들
func recordPaginationSuccess(collection string, recordCount int) {
    // Prometheus, StatsD 등으로 메트릭 전송
    metrics.Counter("pagination_success_total").
        WithTag("collection", collection).
        Increment()
    
    metrics.Histogram("pagination_record_count").
        WithTag("collection", collection).
        Observe(float64(recordCount))
}

func recordPaginationError(collection string, failedPage int, err error) {
    metrics.Counter("pagination_error_total").
        WithTag("collection", collection).
        WithTag("error_type", getErrorType(err)).
        Increment()
    
    metrics.Histogram("pagination_failed_page").
        WithTag("collection", collection).
        Observe(float64(failedPage))
}

func getErrorType(err error) string {
    switch {
    case errors.Is(err, ErrUnauthorized):
        return "auth"
    case errors.Is(err, ErrNetworkTimeout):
        return "network"
    case errors.Is(err, ErrInternalServerError):
        return "server"
    default:
        return "unknown"
    }
}
```

### 3. 헬스 체크 통합

```go
func paginationHealthCheck(ctx context.Context, client *pocketbase.Client) error {
    // 작은 샘플로 페이지네이션 기능 테스트
    testOptions := &pocketbase.ListOptions{
        PerPage: 1, // 최소한의 데이터만 요청
    }
    
    _, err := client.Records.GetAllWithBatchSize(ctx, "health_check", testOptions, 1)
    if err != nil {
        if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
            // 특정 에러는 정상으로 간주 (예: 빈 컬렉션)
            if errors.Is(paginationErr.OriginalErr, ErrNotFound) {
                return nil
            }
        }
        return fmt.Errorf("페이지네이션 헬스 체크 실패: %w", err)
    }
    
    return nil
}
```

## 실전 예제

### 종합적인 에러 처리 패턴

```go
func robustPaginationExample(ctx context.Context, client *pocketbase.Client, collection string) ([]*pocketbase.Record, error) {
    logger := slog.With("collection", collection)
    
    // 1단계: 기본 시도
    records, err := client.Records.GetAll(ctx, collection, nil)
    if err == nil {
        logger.Info("페이지네이션 성공", "count", len(records))
        return records, nil
    }
    
    // 2단계: 에러 분석 및 복구 시도
    if paginationErr, ok := err.(*pocketbase.PaginationError); ok {
        logger.Warn("페이지네이션 에러 발생",
            "page", paginationErr.Page,
            "partial_count", len(paginationErr.GetPartialData()),
        )
        
        // 인증 에러 시 토큰 갱신 후 재시도
        if isAuthError(paginationErr.OriginalErr) {
            if refreshErr := refreshAuthToken(ctx, client); refreshErr == nil {
                logger.Info("토큰 갱신 후 재시도")
                return client.Records.GetAll(ctx, collection, nil)
            }
        }
        
        // 부분 데이터 활용 가능성 검토
        partialData := paginationErr.GetPartialData()
        if len(partialData) > 0 {
            progress := float64(paginationErr.Page-1) / float64(paginationErr.Page) * 100
            
            if progress >= 70.0 { // 70% 이상 완료된 경우
                logger.Info("부분 데이터 사용", 
                    "progress", progress,
                    "count", len(partialData),
                )
                return partialData, nil
            }
        }
    }
    
    // 3단계: 최종 에러 반환
    logger.Error("페이지네이션 최종 실패", "error", err)
    return nil, err
}
```

이 가이드를 통해 페이지네이션 헬퍼의 다양한 에러 상황을 효과적으로 처리하고, 부분 데이터를 활용하여 시스템의 복원력을 높일 수 있습니다.