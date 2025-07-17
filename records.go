package pocketbase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"time"
)

// RecordServiceAPI defines the API operations related to records.
type RecordServiceAPI interface {
	GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error)
	GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error)
	Create(ctx context.Context, collection string, body interface{}) (*Record, error)
	CreateWithOptions(ctx context.Context, collection string, body interface{}, opts *WriteOptions) (*Record, error)
	Update(ctx context.Context, collection, recordID string, body interface{}) (*Record, error)
	UpdateWithOptions(ctx context.Context, collection, recordID string, body interface{}, opts *WriteOptions) (*Record, error)
	Delete(ctx context.Context, collection, recordID string) error
	NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error)
	NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error)
	NewDeleteRequest(collection, recordID string) (*BatchRequest, error)
	NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error)

	// 페이지네이션 헬퍼 메서드들
	GetAll(ctx context.Context, collection string, opts *ListOptions) ([]*Record, error)
	GetAllWithBatchSize(ctx context.Context, collection string, opts *ListOptions, batchSize int) ([]*Record, error)
	Iterate(ctx context.Context, collection string, opts *ListOptions) *RecordIterator
	IterateWithBatchSize(ctx context.Context, collection string, opts *ListOptions, batchSize int) *RecordIterator
}

type Mappable interface {
	ToMap() map[string]any
}

// RecordService handles record-related API operations.
type RecordService struct {
	Client *Client
}

var _ RecordServiceAPI = (*RecordService)(nil)

// GetList retrieves a list of records from a collection.
func (s *RecordService) GetList(ctx context.Context, collection string, opts *ListOptions) (*ListResult, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	applyListOptions(q, opts)
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var result ListResult
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &result); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch records list: %w", err)
	}
	return &result, nil
}

// GetOne retrieves a single record.
func (s *RecordService) GetOne(ctx context.Context, collection, recordID string, opts *GetOneOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	var rec Record
	if err := s.Client.send(ctx, http.MethodGet, path, nil, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: fetch record: %w", err)
	}
	return &rec, nil
}
func (s *RecordService) Create(ctx context.Context, collection string, body interface{}) (*Record, error) {
	return s.CreateWithOptions(ctx, collection, body, nil)
}

func (s *RecordService) CreateWithOptions(ctx context.Context, collection string, body interface{}, opts *WriteOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	requestBody := body
	if mappable, ok := body.(Mappable); ok {
		// If implemented, call ToMap() to convert to map
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPost, path, requestBody, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: create record: %w", err)
	}
	return &rec, nil
}

func (s *RecordService) Update(ctx context.Context, collection, recordID string, body interface{}) (*Record, error) {
	return s.UpdateWithOptions(ctx, collection, recordID, body, nil)
}

func (s *RecordService) UpdateWithOptions(ctx context.Context, collection, recordID string, body interface{}, opts *WriteOptions) (*Record, error) {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	q := url.Values{}
	if opts != nil {
		if opts.Expand != "" {
			q.Set("expand", opts.Expand)
		}
		if opts.Fields != "" {
			q.Set("fields", opts.Fields)
		}
	}
	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}
	requestBody := body
	if mappable, ok := body.(Mappable); ok {
		// If implemented, call ToMap() to convert to map
		requestBody = mappable.ToMap()
	}

	var rec Record
	if err := s.Client.send(ctx, http.MethodPatch, path, requestBody, &rec); err != nil {
		return nil, fmt.Errorf("pocketbase: update record: %w", err)
	}
	return &rec, nil
}

func (s *RecordService) Delete(ctx context.Context, collection, recordID string) error {
	path := fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID))
	if err := s.Client.send(ctx, http.MethodDelete, path, nil, nil); err != nil {
		return fmt.Errorf("pocketbase: delete record: %w", err)
	}
	return nil
}

func (s *RecordService) NewCreateRequest(collection string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPost,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}

func (s *RecordService) NewUpdateRequest(collection, recordID string, body map[string]any) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodPatch,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
		Body:   body,
	}, nil
}

func (s *RecordService) NewDeleteRequest(collection, recordID string) (*BatchRequest, error) {
	return &BatchRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/collections/%s/records/%s", url.PathEscape(collection), url.PathEscape(recordID)),
	}, nil
}

func (s *RecordService) NewUpsertRequest(collection string, body map[string]any) (*BatchRequest, error) {
	if _, ok := body["id"]; !ok {
		return nil, fmt.Errorf("upsert error: 'id' field is required in the body")
	}
	return &BatchRequest{
		Method: http.MethodPut,
		URL:    fmt.Sprintf("/api/collections/%s/records", url.PathEscape(collection)),
		Body:   body,
	}, nil
}

// GetAll 지정된 컬렉션의 모든 레코드를 자동으로 페이지네이션하여 가져옵니다.
// 기본 배치 크기(100)를 사용하여 GetAllWithBatchSize를 호출합니다.
func (s *RecordService) GetAll(ctx context.Context, collection string, opts *ListOptions) ([]*Record, error) {
	return s.GetAllWithBatchSize(ctx, collection, opts, DefaultBatchSize)
}

// GetAllWithBatchSize 지정된 배치 크기로 컬렉션의 모든 레코드를 페이지네이션하여 가져옵니다.
// 컨텍스트 취소, 에러 처리, 부분 데이터 보존 등을 지원합니다.
func (s *RecordService) GetAllWithBatchSize(ctx context.Context, collection string, opts *ListOptions, batchSize int) ([]*Record, error) {
	return s.GetAllWithBatchSizeAndMetrics(ctx, collection, opts, batchSize, nil)
}

// GetAllWithBatchSizeAndMetrics 성능 모니터링과 함께 모든 레코드를 가져옵니다.
func (s *RecordService) GetAllWithBatchSizeAndMetrics(ctx context.Context, collection string, opts *ListOptions, batchSize int, monitor *PerformanceMonitor) ([]*Record, error) {
	// 배치 크기 검증 및 자동 조정
	adjustedBatchSize, wasAdjusted := ValidateAndAdjustBatchSize(batchSize)
	if wasAdjusted {
		// 로그나 디버그 정보를 위해 조정 사실을 기록할 수 있음
		// 현재는 조용히 조정
	}
	batchSize = adjustedBatchSize

	// 성능 모니터가 없으면 기본 모니터 생성 (비활성화 상태)
	if monitor == nil {
		monitor = NewPerformanceMonitor(collection, "GetAll", batchSize)
		monitor.SetEnabled(false) // 명시적으로 요청하지 않은 경우 비활성화
	}

	var allRecords []*Record
	page := 1
	totalProcessed := 0
	estimatedTotal := 0
	currentMemoryUsage := 0
	peakMemoryUsage := 0

	// ListOptions 복사하여 수정
	paginatedOpts := s.copyListOptions(opts)
	paginatedOpts.PerPage = batchSize

	for {
		// 컨텍스트 취소 확인
		select {
		case <-ctx.Done():
			// 성능 모니터링 완료 (부분 결과)
			monitor.UpdateTotals(totalProcessed, page-1)
			monitor.UpdateMemoryUsage(currentMemoryUsage, peakMemoryUsage)
			monitor.Finish(true)

			return allRecords, &PaginationError{
				Operation:      "GetAll",
				Page:           page,
				PartialData:    allRecords,
				OriginalErr:    ctx.Err(),
				TotalProcessed: totalProcessed,
				Message:        "context cancelled during pagination",
			}
		default:
		}

		// 페이지 요청 시작 모니터링
		monitor.StartPage(page)

		// 현재 페이지 설정
		paginatedOpts.Page = page

		// 페이지 데이터 요청
		result, err := s.GetList(ctx, collection, paginatedOpts)

		// 페이지 요청 완료 모니터링
		recordCount := 0
		if result != nil {
			recordCount = len(result.Items)
		}
		monitor.EndPage(page, recordCount, err)

		if err != nil {
			// 성능 모니터링 완료 (부분 결과)
			monitor.UpdateTotals(totalProcessed, page-1)
			monitor.UpdateMemoryUsage(currentMemoryUsage, peakMemoryUsage)
			monitor.Finish(true)

			return allRecords, &PaginationError{
				Operation:      "GetAll",
				Page:           page,
				PartialData:    allRecords,
				OriginalErr:    err,
				TotalProcessed: totalProcessed,
				Message:        fmt.Sprintf("failed to fetch page %d", page),
			}
		}

		// 결과가 없으면 종료
		if len(result.Items) == 0 {
			break
		}

		// 첫 번째 페이지에서 총 예상 크기를 알 수 있으면 슬라이스 용량 미리 할당
		if page == 1 && result.TotalItems > 0 {
			estimatedTotal = result.TotalItems
			// 메모리 효율성을 위해 슬라이스 용량을 미리 할당
			if cap(allRecords) < estimatedTotal {
				// 새로운 슬라이스를 생성하고 기존 데이터 복사
				newRecords := make([]*Record, len(allRecords), estimatedTotal)
				copy(newRecords, allRecords)
				allRecords = newRecords
			}
		}

		// 레코드 추가 - 메모리 효율적인 방식으로
		allRecords = append(allRecords, result.Items...)
		totalProcessed += len(result.Items)

		// 메모리 사용량 추정 및 업데이트
		batchMemory := EstimateBatchMemoryUsage(result.Items)
		currentMemoryUsage += batchMemory
		if currentMemoryUsage > peakMemoryUsage {
			peakMemoryUsage = currentMemoryUsage
		}
		monitor.UpdateMemoryUsage(currentMemoryUsage, peakMemoryUsage)

		// 마지막 페이지 확인
		if page >= result.TotalPages {
			break
		}

		page++

		// 중간 가비지 컬렉션 힌트 (대용량 데이터 처리 시)
		if totalProcessed%1000 == 0 {
			// 1000개 레코드마다 가비지 컬렉션 힌트
			// 실제 GC 호출은 하지 않고 런타임에 맡김
		}
	}

	// 성능 모니터링 완료 (전체 결과)
	monitor.UpdateTotals(totalProcessed, page-1)
	monitor.UpdateMemoryUsage(currentMemoryUsage, peakMemoryUsage)
	monitor.Finish(false)

	return allRecords, nil
}

// Iterate 지정된 컬렉션의 레코드를 메모리 효율적으로 순회하기 위한 Iterator를 반환합니다.
// 기본 배치 크기를 사용하여 IterateWithBatchSize를 호출합니다.
func (s *RecordService) Iterate(ctx context.Context, collection string, opts *ListOptions) *RecordIterator {
	return s.IterateWithBatchSize(ctx, collection, opts, DefaultBatchSize)
}

// IterateWithBatchSize 지정된 배치 크기로 컬렉션의 레코드를 순회하기 위한 Iterator를 반환합니다.
// 메모리 효율적인 페이지별 로딩을 지원하는 Iterator를 생성합니다.
func (s *RecordService) IterateWithBatchSize(ctx context.Context, collection string, opts *ListOptions, batchSize int) *RecordIterator {
	// 배치 크기 검증 및 자동 조정
	adjustedBatchSize, _ := ValidateAndAdjustBatchSize(batchSize)
	batchSize = adjustedBatchSize

	// ListOptions 복사하여 수정
	iteratorOpts := s.copyListOptions(opts)
	iteratorOpts.PerPage = batchSize

	// Iterator 인스턴스 생성
	iterator := &RecordIterator{
		service:        s,
		ctx:            ctx,
		collection:     collection,
		opts:           iteratorOpts,
		batchSize:      batchSize,
		currentPage:    1,
		currentBatch:   nil,
		currentIndex:   0,
		totalPages:     0,
		totalItems:     0,
		finished:       false,
		err:            nil,
		autoCleanup:    true, // 기본적으로 자동 메모리 정리 활성화
		maxBatchMemory: 0,    // 무제한 (향후 확장 가능)
	}

	return iterator
}

// copyListOptions ListOptions를 깊은 복사하여 새로운 인스턴스를 반환합니다.
// QueryParams 맵도 함께 복사하여 원본 옵션에 영향을 주지 않습니다.
func (s *RecordService) copyListOptions(opts *ListOptions) *ListOptions {
	if opts == nil {
		return &ListOptions{}
	}

	// 구조체 필드들을 복사
	copied := &ListOptions{
		Page:      opts.Page,
		PerPage:   opts.PerPage,
		Sort:      opts.Sort,
		Filter:    opts.Filter,
		Expand:    opts.Expand,
		Fields:    opts.Fields,
		SkipTotal: opts.SkipTotal,
	}

	// QueryParams 맵을 깊은 복사
	if opts.QueryParams != nil {
		copied.QueryParams = make(map[string]string, len(opts.QueryParams))
		for key, value := range opts.QueryParams {
			copied.QueryParams[key] = value
		}
	}

	return copied
}

// retryableRequest 네트워크 오류에 대한 자동 재시도 메커니즘을 제공하는 함수입니다.
// 지수적 백오프 전략을 적용하여 재시도 간격을 조정합니다.
func (s *RecordService) retryableRequest(ctx context.Context, requestFunc func() error, maxRetries int, baseDelay time.Duration) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 컨텍스트 취소 확인
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 요청 실행
		err := requestFunc()
		if err == nil {
			return nil // 성공
		}

		lastErr = err

		// 재시도 불가능한 에러인지 확인
		if isNonRetryableError(err) {
			return err
		}

		// 마지막 시도였다면 에러 반환
		if attempt == maxRetries {
			break
		}

		// 지수적 백오프 계산 (2^attempt * baseDelay)
		delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay

		// 최대 지연 시간 제한 (30초)
		maxDelay := 30 * time.Second
		if delay > maxDelay {
			delay = maxDelay
		}

		// 지연 시간만큼 대기 (컨텍스트 취소 가능)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// 다음 시도 계속
		}
	}

	return lastErr
}

// isNonRetryableError 재시도하면 안 되는 에러 타입을 판단합니다.
// 인증, 권한, 잘못된 요청 등의 에러는 재시도해도 성공할 가능성이 낮습니다.
func isNonRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// APIError 타입 확인
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// HTTP 상태 코드 기반 판단
		switch apiErr.Code {
		case 400: // Bad Request - 잘못된 요청
			return true
		case 401: // Unauthorized - 인증 실패
			return true
		case 403: // Forbidden - 권한 없음
			return true
		case 404: // Not Found - 리소스 없음
			return true
		case 409: // Conflict - 충돌
			return true
		case 422: // Unprocessable Entity - 유효성 검사 실패
			return true
		case 429: // Too Many Requests - 속도 제한 (재시도 가능)
			return false
		case 500, 502, 503, 504: // 서버 에러 (재시도 가능)
			return false
		default:
			// 기타 4xx 에러는 재시도 불가능
			if apiErr.Code >= 400 && apiErr.Code < 500 {
				return true
			}
			// 5xx 에러는 재시도 가능
			return false
		}
	}

	// ClientError 타입 확인
	var clientErr *ClientError
	if errors.As(err, &clientErr) {
		return isNonRetryableError(clientErr.OriginalErr)
	}

	// 기본적으로 네트워크 관련 에러는 재시도 가능
	return false
}

// GetAllWithMetrics 성능 모니터링을 활성화하여 모든 레코드를 가져옵니다.
// 성능 메트릭을 수집하고 반환합니다.
func (s *RecordService) GetAllWithMetrics(ctx context.Context, collection string, opts *ListOptions) ([]*Record, *PaginationMetrics, error) {
	return s.GetAllWithBatchSizeAndMetricsEnabled(ctx, collection, opts, DefaultBatchSize, true)
}

// GetAllWithBatchSizeAndMetricsEnabled 배치 크기와 성능 모니터링 활성화 여부를 지정하여 모든 레코드를 가져옵니다.
func (s *RecordService) GetAllWithBatchSizeAndMetricsEnabled(ctx context.Context, collection string, opts *ListOptions, batchSize int, enableMetrics bool) ([]*Record, *PaginationMetrics, error) {
	var monitor *PerformanceMonitor
	if enableMetrics {
		monitor = NewPerformanceMonitor(collection, "GetAll", batchSize)
		monitor.SetCollectPageData(true) // 페이지별 세부 데이터도 수집
	}

	records, err := s.GetAllWithBatchSizeAndMetrics(ctx, collection, opts, batchSize, monitor)

	var metrics *PaginationMetrics
	if monitor != nil {
		metrics = monitor.GetMetrics()
	}

	return records, metrics, err
}

// IterateWithMetrics 성능 모니터링을 활성화하여 Iterator를 생성합니다.
func (s *RecordService) IterateWithMetrics(ctx context.Context, collection string, opts *ListOptions) (*RecordIterator, *PerformanceMonitor) {
	return s.IterateWithBatchSizeAndMetrics(ctx, collection, opts, DefaultBatchSize)
}

// IterateWithBatchSizeAndMetrics 배치 크기와 성능 모니터링을 지정하여 Iterator를 생성합니다.
func (s *RecordService) IterateWithBatchSizeAndMetrics(ctx context.Context, collection string, opts *ListOptions, batchSize int) (*RecordIterator, *PerformanceMonitor) {
	// 배치 크기 검증 및 자동 조정
	adjustedBatchSize, _ := ValidateAndAdjustBatchSize(batchSize)
	batchSize = adjustedBatchSize

	// 성능 모니터 생성
	monitor := NewPerformanceMonitor(collection, "Iterate", batchSize)
	monitor.SetCollectPageData(true)

	// ListOptions 복사하여 수정
	iteratorOpts := s.copyListOptions(opts)
	iteratorOpts.PerPage = batchSize

	// Iterator 인스턴스 생성 (성능 모니터링 포함)
	iterator := &RecordIterator{
		service:        s,
		ctx:            ctx,
		collection:     collection,
		opts:           iteratorOpts,
		batchSize:      batchSize,
		currentPage:    1,
		currentBatch:   nil,
		currentIndex:   0,
		totalPages:     0,
		totalItems:     0,
		finished:       false,
		err:            nil,
		autoCleanup:    true,    // 기본적으로 자동 메모리 정리 활성화
		maxBatchMemory: 0,       // 무제한 (향후 확장 가능)
		monitor:        monitor, // 성능 모니터 추가
	}

	return iterator, monitor
}

// CompareBatchSizePerformance 다양한 배치 크기로 성능을 비교합니다.
// 테스트용 함수로, 최적의 배치 크기를 찾는 데 도움이 됩니다.
func (s *RecordService) CompareBatchSizePerformance(ctx context.Context, collection string, opts *ListOptions, batchSizes []int) ([]*PaginationMetrics, error) {
	if len(batchSizes) == 0 {
		return nil, fmt.Errorf("batch sizes cannot be empty")
	}

	results := make([]*PaginationMetrics, 0, len(batchSizes))

	for _, batchSize := range batchSizes {
		// 각 배치 크기별로 성능 테스트 실행
		_, metrics, err := s.GetAllWithBatchSizeAndMetricsEnabled(ctx, collection, opts, batchSize, true)
		if err != nil {
			// 에러가 발생해도 부분 결과가 있으면 포함
			if metrics != nil {
				results = append(results, metrics)
			}
			continue
		}

		if metrics != nil {
			results = append(results, metrics)
		}
	}

	return results, nil
}

// GetOptimalBatchSizeForCollection 특정 컬렉션에 대한 최적의 배치 크기를 찾습니다.
// 여러 배치 크기를 테스트하여 가장 효율적인 크기를 반환합니다.
func (s *RecordService) GetOptimalBatchSizeForCollection(ctx context.Context, collection string, opts *ListOptions) (int, *PaginationMetrics, error) {
	// 테스트할 배치 크기들
	testBatchSizes := []int{25, 50, 100, 200, 500}

	metrics, err := s.CompareBatchSizePerformance(ctx, collection, opts, testBatchSizes)
	if err != nil {
		return DefaultBatchSize, nil, err
	}

	if len(metrics) == 0 {
		return DefaultBatchSize, nil, fmt.Errorf("no performance data collected")
	}

	// 최적의 배치 크기 선택 (초당 레코드 수 기준)
	bestMetrics := metrics[0]
	bestBatchSize := bestMetrics.BatchSize

	for _, metric := range metrics[1:] {
		// 초당 레코드 수가 더 높고, 에러가 적은 배치 크기 선택
		if metric.RecordsPerSecond > bestMetrics.RecordsPerSecond && metric.ErrorCount <= bestMetrics.ErrorCount {
			bestMetrics = metric
			bestBatchSize = metric.BatchSize
		}
	}

	return bestBatchSize, bestMetrics, nil
}
