package pocketbase

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goccy/go-json"
)

// ClientOption configures a Client instance.
// ClientOption configures a Client instance.
type ClientOption func(*Client)

// WithHTTPClient sets the HTTP client used for requests.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = hc
	}
}

type requestOptions struct {
	writer io.Writer
}

// RequestOption configures the behavior of a single request.
type RequestOption func(*requestOptions)

// WithResponseWriter streams the response body to the provided writer.
// If the writer implements http.Flusher, Flush is called after each write.
// This option cannot be used together with the responseData argument
// of Client.send or SendWithOptions; doing so results in an error.
func WithResponseWriter(w io.Writer) RequestOption {
	return func(o *requestOptions) {
		o.writer = w
	}
}

// 페이지네이션 관련 상수들
const (
	// DefaultBatchSize는 기본 배치 크기입니다
	DefaultBatchSize = 100

	// MinBatchSize는 최소 배치 크기입니다
	MinBatchSize = 1

	// MaxBatchSize는 최대 배치 크기입니다
	MaxBatchSize = 1000

	// DefaultMaxRetries는 기본 최대 재시도 횟수입니다
	DefaultMaxRetries = 3

	// DefaultRetryDelaySeconds는 기본 재시도 지연 시간(초)입니다
	DefaultRetryDelaySeconds = 1
)

// DefaultPaginationOptions 기본 페이지네이션 옵션을 반환합니다.
func DefaultPaginationOptions() *PaginationOptions {
	return &PaginationOptions{
		BatchSize:   DefaultBatchSize,
		MaxRetries:  DefaultMaxRetries,
		RetryDelay:  time.Duration(DefaultRetryDelaySeconds) * time.Second,
		StopOnError: false,
		Context:     context.Background(),
	}
}

// ValidateBatchSize 배치 크기가 유효한 범위 내에 있는지 검증합니다.
func ValidateBatchSize(batchSize int) error {
	if batchSize < MinBatchSize {
		return fmt.Errorf("batch size must be at least %d, got %d", MinBatchSize, batchSize)
	}
	if batchSize > MaxBatchSize {
		return fmt.Errorf("batch size must be at most %d, got %d", MaxBatchSize, batchSize)
	}
	return nil
}

// ValidateAndAdjustBatchSize 배치 크기를 검증하고 필요시 유효한 값으로 조정합니다.
// 잘못된 값이 입력되면 기본값으로 조정하고 조정된 값을 반환합니다.
func ValidateAndAdjustBatchSize(batchSize int) (int, bool) {
	if batchSize <= 0 {
		return DefaultBatchSize, true
	}
	if batchSize < MinBatchSize {
		return MinBatchSize, true
	}
	if batchSize > MaxBatchSize {
		return MaxBatchSize, true
	}
	return batchSize, false
}

// GetOptimalBatchSize 주어진 총 레코드 수에 대해 최적의 배치 크기를 계산합니다.
// 네트워크 요청 횟수와 메모리 사용량의 균형을 고려합니다.
func GetOptimalBatchSize(totalRecords int) int {
	if totalRecords <= 0 {
		return DefaultBatchSize
	}

	// 작은 데이터셋의 경우 한 번에 가져오기
	if totalRecords <= 50 {
		return totalRecords
	}

	// 중간 크기 데이터셋의 경우 적절한 배치 크기 계산
	if totalRecords <= 1000 {
		// 대략 5-10번의 요청으로 나누어 처리
		batchSize := totalRecords / 7
		if batchSize < MinBatchSize {
			return MinBatchSize
		}
		if batchSize > DefaultBatchSize {
			return DefaultBatchSize
		}
		return batchSize
	}

	// 대용량 데이터의 경우 기본 배치 크기 사용
	return DefaultBatchSize
}

// IsBatchSizeRecommended 주어진 배치 크기가 권장 범위 내에 있는지 확인합니다.
// 성능 최적화를 위한 권장 범위를 벗어나면 false를 반환합니다.
func IsBatchSizeRecommended(batchSize int) bool {
	// 권장 범위: 10-500
	const (
		RecommendedMinBatchSize = 10
		RecommendedMaxBatchSize = 500
	)

	return batchSize >= RecommendedMinBatchSize && batchSize <= RecommendedMaxBatchSize
}

// MemoryOptimizationOptions 메모리 최적화 관련 옵션들을 정의합니다.
type MemoryOptimizationOptions struct {
	// AutoCleanup 사용 완료된 배치 데이터 자동 정리 여부
	AutoCleanup bool

	// PreallocateSlice 총 레코드 수를 알 때 슬라이스 용량 미리 할당 여부
	PreallocateSlice bool

	// GCHintInterval 가비지 컬렉션 힌트를 제공할 레코드 처리 간격
	GCHintInterval int

	// MaxBatchMemory 배치당 최대 메모리 사용량 (바이트, 0은 무제한)
	MaxBatchMemory int
}

// DefaultMemoryOptimizationOptions 기본 메모리 최적화 옵션을 반환합니다.
func DefaultMemoryOptimizationOptions() *MemoryOptimizationOptions {
	return &MemoryOptimizationOptions{
		AutoCleanup:      true,
		PreallocateSlice: true,
		GCHintInterval:   1000, // 1000개 레코드마다
		MaxBatchMemory:   0,    // 무제한
	}
}

// EstimateRecordMemoryUsage 레코드 하나당 대략적인 메모리 사용량을 추정합니다.
// 이는 대략적인 값이며 실제 메모리 사용량과 다를 수 있습니다.
func EstimateRecordMemoryUsage(record *Record) int {
	if record == nil {
		return 0
	}

	// 기본 구조체 크기 (대략적)
	baseSize := 64 // Record 구조체 기본 크기

	// 데이터 맵의 크기 추정 - deserializedData는 private이므로 JSON 마샬링으로 추정
	// Record의 모든 필드를 JSON으로 마샬링한 크기로 추정
	if jsonData, err := json.Marshal(record); err == nil {
		baseSize += len(jsonData)
	} else {
		// JSON 마샬링 실패 시 기본 크기 사용
		baseSize += 256 // 기본 추정 크기
	}

	return baseSize
}

// EstimateBatchMemoryUsage 배치 전체의 메모리 사용량을 추정합니다.
func EstimateBatchMemoryUsage(records []*Record) int {
	if len(records) == 0 {
		return 0
	}

	totalSize := 0

	// 슬라이스 자체의 오버헤드
	totalSize += len(records) * 8 // 포인터 크기

	// 각 레코드의 메모리 사용량 합계
	for _, record := range records {
		totalSize += EstimateRecordMemoryUsage(record)
	}

	return totalSize
}
