package pocketbase

import (
	"context"
	"fmt"
	"time"

	"github.com/goccy/go-json" // Modified to use goccy/go-json directly
	"github.com/pocketbase/pocketbase/tools/types"
)

// BaseModel provides common fields for all PocketBase models.
type BaseModel struct {
	ID      string         `json:"id"`
	Created types.DateTime `json:"created"`
	Updated types.DateTime `json:"updated"`
}

// Admin represents a PocketBase administrator.
type Admin struct {
	BaseModel
	Avatar int    `json:"avatar"`
	Email  string `json:"email"`
}

// Record represents a PocketBase record.
// ✅ Modified: Remove Data field and use deserializedData directly.
type Record struct {
	BaseModel
	CollectionID   string               `json:"collectionId"`
	CollectionName string               `json:"collectionName"`
	Expand         map[string][]*Record `json:"expand,omitempty"`

	// Map to store data. From now on, this field is the only data source.
	deserializedData map[string]interface{}
}

// ListResult is a struct containing a list of records along with pagination information.
type ListResult struct {
	Page       int       `json:"page"`
	PerPage    int       `json:"perPage"`
	TotalItems int       `json:"totalItems"`
	TotalPages int       `json:"totalPages"`
	Items      []*Record `json:"items"`
}

// ... (Other option structs like ListOptions, GetOneOptions remain the same)
type ListOptions struct {
	Page        int
	PerPage     int
	Sort        string
	Filter      string
	Expand      string
	Fields      string
	SkipTotal   bool
	QueryParams map[string]string
}

type GetOneOptions struct {
	Expand string
	Fields string
}

type WriteOptions struct {
	Expand string
	Fields string
}

type FileDownloadOptions struct {
	Thumb    string
	Download bool
}

// AuthResponse represents the data returned after successful authentication.
type AuthResponse struct {
	Token  string  `json:"token"`
	Record *Record `json:"record,omitempty"`
	Admin  *Admin  `json:"admin,omitempty"`
}

// RealtimeEvent is an event delivered via real-time subscription.
type RealtimeEvent struct {
	Action string  `json:"action"`
	Record *Record `json:"record"`
}

// ✅ Modified: Change UnmarshalJSON to be much simpler and more efficient.
func (r *Record) UnmarshalJSON(data []byte) error {
	// Decode all JSON data into temporary map at once.
	var allData map[string]interface{}
	if err := json.Unmarshal(data, &allData); err != nil {
		return err
	}

	// Assign common fields directly to Record struct.
	if id, ok := allData["id"].(string); ok {
		r.ID = id
	}
	if created, ok := allData["created"].(string); ok {
		r.Created, _ = types.ParseDateTime(created)
	}
	if updated, ok := allData["updated"].(string); ok {
		r.Updated, _ = types.ParseDateTime(updated)
	}
	if colID, ok := allData["collectionId"].(string); ok {
		r.CollectionID = colID
	}
	if colName, ok := allData["collectionName"].(string); ok {
		r.CollectionName = colName
	}
	// Also handle Expand field.
	if expandData, ok := allData["expand"]; ok {
		// Re-serialize expand data to JSON then decode to Record's Expand field.
		// This is the safest method because expand can have complex nested structures.
		expandBytes, err := json.Marshal(expandData)
		if err == nil {
			json.Unmarshal(expandBytes, &r.Expand)
		}
	}

	// Remove common fields and expand from map.
	delete(allData, "id")
	delete(allData, "created")
	delete(allData, "updated")
	delete(allData, "collectionId")
	delete(allData, "collectionName")
	delete(allData, "expand")

	// Store remaining data in deserializedData.
	r.deserializedData = allData

	return nil
}

// Get returns a raw interface{} value for a given key.
func (r *Record) Get(key string) interface{} {
	if r.deserializedData == nil {
		r.deserializedData = make(map[string]interface{})
	}
	return r.deserializedData[key]
}

// Set stores a key-value pair in the record's data.
func (r *Record) Set(key string, value interface{}) {
	if r.deserializedData == nil {
		r.deserializedData = make(map[string]interface{})
	}
	r.deserializedData[key] = value
}

// ✅ Modified: Change MarshalJSON to also use deserializedData directly.
func (r *Record) MarshalJSON() ([]byte, error) {
	combinedData := make(map[string]interface{}, len(r.deserializedData)+6)
	for k, v := range r.deserializedData {
		combinedData[k] = v
	}

	combinedData["id"] = r.ID
	combinedData["collectionId"] = r.CollectionID
	combinedData["collectionName"] = r.CollectionName
	combinedData["created"] = r.Created
	combinedData["updated"] = r.Updated
	if r.Expand != nil {
		combinedData["expand"] = r.Expand
	}

	return json.Marshal(combinedData)
}

// GetString returns a string value for a given key.
func (r *Record) GetString(key string) string {
	val := r.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// GetBool returns a boolean value for a given key.
func (r *Record) GetBool(key string) bool {
	val := r.Get(key)
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GetFloat returns a float64 value for a given key.
func (r *Record) GetFloat(key string) float64 {
	val := r.Get(key)
	if f, ok := val.(float64); ok {
		return f
	}
	if i, ok := val.(int); ok {
		return float64(i)
	}
	// For JSON numbers
	if num, ok := val.(json.Number); ok {
		f, _ := num.Float64()
		return f
	}
	return 0
}

// GetDateTime returns a types.DateTime value for a given key.
func (r *Record) GetDateTime(key string) types.DateTime {
	val := r.Get(key)
	if dt, ok := val.(types.DateTime); ok {
		return dt
	}
	if str, ok := val.(string); ok {
		dt, err := types.ParseDateTime(str)
		if err == nil {
			return dt
		}
	}
	return types.DateTime{}
}

// GetStringSlice returns a slice of strings for a given key.
func (r *Record) GetStringSlice(key string) []string {
	val := r.Get(key)
	if slice, ok := val.([]string); ok {
		return slice
	}
	if slice, ok := val.([]interface{}); ok {
		result := make([]string, len(slice))
		for i, v := range slice {
			if str, ok := v.(string); ok {
				result[i] = str
			}
		}
		return result
	}

	return []string{}
}

// GetRawMessage returns a json.RawMessage value for a given key.
func (r *Record) GetRawMessage(key string) json.RawMessage {
	val := r.Get(key)
	if raw, ok := val.(json.RawMessage); ok {
		return raw
	}
	// If it was parsed into a map/slice, re-marshal it.
	// This can happen if Set() was called before.
	if val != nil {
		bytes, err := json.Marshal(val)
		if err == nil {
			return bytes
		}
	}
	return nil
}

// GetStringPointer returns a pointer to a string value for a given key.
// Returns nil if the key is not present or the value is not a string.
func (r *Record) GetStringPointer(key string) *string {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*string); ok {
		return ptr
	}
	if str, ok := val.(string); ok {
		return &str
	}
	return nil
}

// GetBoolPointer returns a pointer to a boolean value for a given key.
// Returns nil if the key is not present or the value is not a bool.
func (r *Record) GetBoolPointer(key string) *bool {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*bool); ok {
		return ptr
	}
	if b, ok := val.(bool); ok {
		return &b
	}
	return nil
}

// GetFloatPointer returns a pointer to a float64 value for a given key.
// Returns nil if the key is not present or the value is not a number.
func (r *Record) GetFloatPointer(key string) *float64 {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*float64); ok {
		return ptr
	}

	var f float64
	var ok bool

	if num, isNum := val.(json.Number); isNum {
		f, err := num.Float64()
		if err == nil {
			return &f
		}
	} else if f, ok = val.(float64); ok {
		return &f
	} else if i, ok := val.(int); ok {
		f = float64(i)
		return &f
	} else if i64, ok := val.(int64); ok {
		f = float64(i64)
		return &f
	}

	return nil
}

// GetDateTimePointer returns a pointer to a types.DateTime value for a given key.
// Returns nil if the key is not present or the value cannot be parsed as a DateTime.
func (r *Record) GetDateTimePointer(key string) *types.DateTime {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if ptr, ok := val.(*types.DateTime); ok {
		return ptr
	}
	if dt, ok := val.(types.DateTime); ok {
		return &dt
	}
	if str, ok := val.(string); ok {
		dt, err := types.ParseDateTime(str)
		if err == nil {
			return &dt
		}
	}
	return nil
}

// PaginationOptions 페이지네이션 헬퍼 메서드에 대한 설정 옵션을 정의합니다.
type PaginationOptions struct {
	// BatchSize는 한 번에 가져올 레코드 수를 지정합니다 (기본값: 100)
	BatchSize int

	// MaxRetries는 실패 시 최대 재시도 횟수를 지정합니다 (기본값: 3)
	MaxRetries int

	// RetryDelay는 재시도 간격을 지정합니다 (기본값: 1초)
	RetryDelay time.Duration

	// StopOnError가 true이면 에러 발생 시 즉시 중단합니다 (기본값: false)
	StopOnError bool

	// Context는 작업 취소를 위한 컨텍스트입니다
	Context context.Context
}

// PaginationError 페이지네이션 작업 중 발생한 에러와 부분적으로 수집된 데이터를 포함합니다.
type PaginationError struct {
	// Operation은 에러가 발생한 작업 유형입니다 ("GetAll" 또는 "Iterate")
	Operation string

	// Page는 에러가 발생한 페이지 번호입니다
	Page int

	// PartialData는 에러 발생 전까지 수집된 레코드들입니다
	PartialData []*Record

	// OriginalErr는 발생한 원본 에러입니다
	OriginalErr error

	// TotalProcessed는 에러 발생 전까지 처리된 총 레코드 수입니다
	TotalProcessed int

	// Message는 에러에 대한 추가 설명입니다
	Message string
}

// Error PaginationError가 error 인터페이스를 구현하도록 합니다.
func (pe *PaginationError) Error() string {
	if pe.Message != "" {
		return fmt.Sprintf("pagination %s failed at page %d (processed %d records): %s - %v",
			pe.Operation, pe.Page, pe.TotalProcessed, pe.Message, pe.OriginalErr)
	}
	return fmt.Sprintf("pagination %s failed at page %d (processed %d records): %v",
		pe.Operation, pe.Page, pe.TotalProcessed, pe.OriginalErr)
}

// Unwrap 원본 에러를 반환합니다 (Go 1.13+ error unwrapping 지원).
func (pe *PaginationError) Unwrap() error {
	return pe.OriginalErr
}

// GetPartialData 에러 발생 전까지 수집된 부분 데이터를 반환합니다.
func (pe *PaginationError) GetPartialData() []*Record {
	return pe.PartialData
}

// Is 에러 타입 비교를 위한 메서드입니다.
func (pe *PaginationError) Is(target error) bool {
	_, ok := target.(*PaginationError)
	return ok
}

// IteratorFunc 레코드 반복 처리를 위한 함수 타입입니다.
// 반환값이 false이면 반복을 중단합니다.
type IteratorFunc func(record *Record, index int) bool

// RecordIterator 메모리 효율적인 레코드 순회를 위한 Iterator 구조체입니다.
// 페이지별로 데이터를 로드하여 메모리 사용량을 최적화합니다.
type RecordIterator struct {
	// 서비스 참조
	service *RecordService

	// 요청 파라미터
	ctx        context.Context
	collection string
	opts       *ListOptions
	batchSize  int

	// 상태 관리
	currentPage  int
	currentBatch []*Record
	currentIndex int
	totalPages   int
	totalItems   int
	finished     bool

	// 에러 상태
	err error

	// 메모리 최적화 옵션
	autoCleanup    bool // 사용 완료된 배치 자동 정리
	maxBatchMemory int  // 최대 배치 메모리 크기 (바이트)

	// 성능 모니터링 (선택적)
	monitor *PerformanceMonitor
}

// Next 다음 레코드가 있는지 확인하고 필요시 다음 배치를 로드합니다.
// 다음 레코드가 있으면 true를, 더 이상 레코드가 없거나 에러가 발생하면 false를 반환합니다.
func (ri *RecordIterator) Next() bool {
	// 에러가 있으면 false 반환
	if ri.err != nil {
		return false
	}

	// 컨텍스트 취소 확인
	select {
	case <-ri.ctx.Done():
		ri.err = ri.ctx.Err()
		return false
	default:
	}

	// 현재 배치에 더 처리할 레코드가 있는지 먼저 확인
	if ri.currentBatch != nil && ri.currentIndex < len(ri.currentBatch) {
		return true
	}

	// 현재 배치가 없거나 모든 레코드를 처리했으면 다음 배치 로드 시도
	// 단, 이미 모든 페이지를 로드했다면 더 이상 시도하지 않음
	if ri.totalPages > 0 && ri.currentPage > ri.totalPages {
		return false
	}

	// 다음 배치 로드 시도
	if !ri.loadNextBatch() {
		return false
	}

	// 새로 로드된 배치에 레코드가 있는지 확인
	if ri.currentBatch != nil && len(ri.currentBatch) > 0 {
		return true
	}

	return false
}

// Record 현재 레코드를 반환하고 인덱스를 다음으로 이동합니다.
// Next()가 true를 반환한 후에 호출해야 합니다.
func (ri *RecordIterator) Record() *Record {
	if ri.currentBatch == nil || ri.currentIndex >= len(ri.currentBatch) {
		return nil
	}

	record := ri.currentBatch[ri.currentIndex]
	ri.currentIndex++
	return record
}

// Error Iterator 작업 중 발생한 에러를 반환합니다.
// 에러가 없으면 nil을 반환합니다.
func (ri *RecordIterator) Error() error {
	return ri.err
}

// loadNextBatch 다음 배치의 레코드들을 로드합니다.
// 성공적으로 로드하면 true를, 더 이상 로드할 데이터가 없거나 에러가 발생하면 false를 반환합니다.
func (ri *RecordIterator) loadNextBatch() bool {
	// 에러가 있으면 false 반환
	if ri.err != nil {
		return false
	}

	// 컨텍스트 취소 확인
	select {
	case <-ri.ctx.Done():
		ri.err = ri.ctx.Err()
		return false
	default:
	}

	// 이전 배치 메모리 정리 (자동 정리가 활성화된 경우)
	if ri.autoCleanup && ri.currentBatch != nil {
		ri.cleanupCurrentBatch()
	}

	// 현재 페이지 설정
	ri.opts.Page = ri.currentPage

	// 성능 모니터링 시작
	if ri.monitor != nil {
		ri.monitor.StartPage(ri.currentPage)
	}

	// 페이지 데이터 요청
	result, err := ri.service.GetList(ri.ctx, ri.collection, ri.opts)

	// 성능 모니터링 완료
	recordCount := 0
	if result != nil {
		recordCount = len(result.Items)
	}
	if ri.monitor != nil {
		ri.monitor.EndPage(ri.currentPage, recordCount, err)
	}

	if err != nil {
		ri.err = err
		return false
	}

	// 첫 번째 배치에서 총 페이지 수와 총 아이템 수 설정
	if ri.currentPage == 1 {
		ri.totalPages = result.TotalPages
		ri.totalItems = result.TotalItems
	}

	// 결과가 없으면 완료 처리
	if len(result.Items) == 0 {
		ri.finished = true
		return false
	}

	// 새로운 배치 설정
	ri.currentBatch = result.Items
	ri.currentIndex = 0

	// 다음 페이지로 이동
	ri.currentPage++

	return true
}

// cleanupCurrentBatch 현재 배치의 메모리를 정리합니다.
// 사용 완료된 레코드들의 참조를 제거하여 가비지 컬렉션을 돕습니다.
func (ri *RecordIterator) cleanupCurrentBatch() {
	if ri.currentBatch == nil {
		return
	}

	// 슬라이스의 각 요소를 nil로 설정하여 참조 해제
	for i := range ri.currentBatch {
		ri.currentBatch[i] = nil
	}

	// 슬라이스 자체를 nil로 설정
	ri.currentBatch = nil
}

// Close Iterator를 명시적으로 종료하고 메모리를 정리합니다.
// 사용이 완료된 후 호출하는 것을 권장합니다.
func (ri *RecordIterator) Close() error {
	// 현재 배치 정리
	ri.cleanupCurrentBatch()

	// 상태 초기화
	ri.finished = true
	ri.currentIndex = 0
	ri.totalPages = 0
	ri.totalItems = 0

	// 옵션 객체 정리 (깊은 복사된 객체이므로 안전)
	if ri.opts != nil && ri.opts.QueryParams != nil {
		// QueryParams 맵 정리
		for k := range ri.opts.QueryParams {
			delete(ri.opts.QueryParams, k)
		}
		ri.opts.QueryParams = nil
	}

	return nil
}

// HasMore 더 처리할 레코드가 있는지 확인합니다.
// Next()를 호출하지 않고도 상태를 확인할 수 있습니다.
func (ri *RecordIterator) HasMore() bool {
	return !ri.finished && ri.err == nil
}

// GetProgress 현재 진행 상황을 반환합니다.
// 총 아이템 수가 알려진 경우 진행률을 계산할 수 있습니다.
func (ri *RecordIterator) GetProgress() (current, total int) {
	processedInCurrentBatch := ri.currentIndex
	processedInPreviousPages := (ri.currentPage - 1) * ri.batchSize
	current = processedInPreviousPages + processedInCurrentBatch
	total = ri.totalItems
	return current, total
}

// SetAutoCleanup 자동 메모리 정리 기능을 활성화/비활성화합니다.
// true로 설정하면 각 배치 처리 후 자동으로 메모리를 정리합니다.
func (ri *RecordIterator) SetAutoCleanup(enabled bool) {
	ri.autoCleanup = enabled
}

// SetMaxBatchMemory 배치당 최대 메모리 사용량을 설정합니다 (바이트 단위).
// 현재는 정보성 목적으로만 사용되며, 향후 메모리 모니터링에 활용될 수 있습니다.
func (ri *RecordIterator) SetMaxBatchMemory(maxBytes int) {
	ri.maxBatchMemory = maxBytes
}

// PaginationMetrics 페이지네이션 성능 메트릭을 수집하는 구조체입니다.
type PaginationMetrics struct {
	// 기본 정보
	Collection string    `json:"collection"`
	Operation  string    `json:"operation"` // "GetAll" 또는 "Iterate"
	BatchSize  int       `json:"batch_size"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`

	// 성능 지표
	TotalDuration    time.Duration `json:"total_duration"`
	TotalRecords     int           `json:"total_records"`
	TotalPages       int           `json:"total_pages"`
	TotalRequests    int           `json:"total_requests"`
	AveragePageTime  time.Duration `json:"average_page_time"`
	RecordsPerSecond float64       `json:"records_per_second"`

	// 메모리 사용량 (추정치)
	EstimatedMemoryUsage int `json:"estimated_memory_usage"`
	PeakMemoryUsage      int `json:"peak_memory_usage"`

	// 에러 정보
	ErrorCount    int    `json:"error_count"`
	LastError     string `json:"last_error,omitempty"`
	PartialResult bool   `json:"partial_result"`

	// 페이지별 세부 정보 (선택적)
	PageMetrics []PageMetric `json:"page_metrics,omitempty"`
}

// PageMetric 개별 페이지 요청의 성능 메트릭입니다.
type PageMetric struct {
	Page         int           `json:"page"`
	Duration     time.Duration `json:"duration"`
	RecordCount  int           `json:"record_count"`
	RequestSize  int           `json:"request_size"`  // 요청 크기 (바이트)
	ResponseSize int           `json:"response_size"` // 응답 크기 (바이트)
	Error        string        `json:"error,omitempty"`
}

// PerformanceMonitor 성능 모니터링을 위한 헬퍼 구조체입니다.
type PerformanceMonitor struct {
	enabled         bool
	collectPageData bool
	metrics         *PaginationMetrics
	pageStartTime   time.Time
}

// NewPerformanceMonitor 새로운 성능 모니터를 생성합니다.
func NewPerformanceMonitor(collection, operation string, batchSize int) *PerformanceMonitor {
	return &PerformanceMonitor{
		enabled: true,
		metrics: &PaginationMetrics{
			Collection:  collection,
			Operation:   operation,
			BatchSize:   batchSize,
			StartTime:   time.Now(),
			PageMetrics: make([]PageMetric, 0),
		},
	}
}

// SetEnabled 성능 모니터링 활성화/비활성화를 설정합니다.
func (pm *PerformanceMonitor) SetEnabled(enabled bool) {
	pm.enabled = enabled
}

// SetCollectPageData 페이지별 세부 데이터 수집 여부를 설정합니다.
func (pm *PerformanceMonitor) SetCollectPageData(collect bool) {
	pm.collectPageData = collect
}

// StartPage 페이지 요청 시작을 기록합니다.
func (pm *PerformanceMonitor) StartPage(page int) {
	if !pm.enabled {
		return
	}
	pm.pageStartTime = time.Now()
}

// EndPage 페이지 요청 완료를 기록합니다.
func (pm *PerformanceMonitor) EndPage(page int, recordCount int, err error) {
	if !pm.enabled {
		return
	}

	duration := time.Since(pm.pageStartTime)
	pm.metrics.TotalRequests++

	if err != nil {
		pm.metrics.ErrorCount++
		pm.metrics.LastError = err.Error()
	}

	// 페이지별 세부 데이터 수집
	if pm.collectPageData {
		pageMetric := PageMetric{
			Page:        page,
			Duration:    duration,
			RecordCount: recordCount,
		}
		if err != nil {
			pageMetric.Error = err.Error()
		}
		pm.metrics.PageMetrics = append(pm.metrics.PageMetrics, pageMetric)
	}
}

// UpdateTotals 전체 통계를 업데이트합니다.
func (pm *PerformanceMonitor) UpdateTotals(totalRecords, totalPages int) {
	if !pm.enabled {
		return
	}

	pm.metrics.TotalRecords = totalRecords
	pm.metrics.TotalPages = totalPages
}

// UpdateMemoryUsage 메모리 사용량을 업데이트합니다.
func (pm *PerformanceMonitor) UpdateMemoryUsage(current, peak int) {
	if !pm.enabled {
		return
	}

	pm.metrics.EstimatedMemoryUsage = current
	if peak > pm.metrics.PeakMemoryUsage {
		pm.metrics.PeakMemoryUsage = peak
	}
}

// Finish 성능 모니터링을 완료하고 최종 메트릭을 계산합니다.
func (pm *PerformanceMonitor) Finish(partialResult bool) *PaginationMetrics {
	if !pm.enabled {
		return nil
	}

	pm.metrics.EndTime = time.Now()
	pm.metrics.TotalDuration = pm.metrics.EndTime.Sub(pm.metrics.StartTime)
	pm.metrics.PartialResult = partialResult

	// 평균 페이지 시간 계산
	if pm.metrics.TotalRequests > 0 {
		pm.metrics.AveragePageTime = pm.metrics.TotalDuration / time.Duration(pm.metrics.TotalRequests)
	}

	// 초당 레코드 수 계산
	if pm.metrics.TotalDuration.Seconds() > 0 {
		pm.metrics.RecordsPerSecond = float64(pm.metrics.TotalRecords) / pm.metrics.TotalDuration.Seconds()
	}

	return pm.metrics
}

// GetMetrics 현재까지의 메트릭을 반환합니다.
func (pm *PerformanceMonitor) GetMetrics() *PaginationMetrics {
	if !pm.enabled {
		return nil
	}
	return pm.metrics
}
