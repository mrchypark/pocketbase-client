package pocketbase

import (
	"bytes"
	"sync"
	"testing"

	"github.com/goccy/go-json"
)

// sampleRecordJSON is sample JSON data for testing.
// It includes basic fields and additional data fields, similar to real data.
var sampleRecordJSON = []byte(`{
    "id": "RECORD_ID",
    "collectionId": "COLLECTION_ID",
    "collectionName": "posts",
    "created": "2025-07-02T10:30:00.123Z",
    "updated": "2025-07-02T10:30:00.456Z",
    "title": "Hello, World!",
    "is_published": true,
    "view_count": 1024,
    "user": "USER_ID_123",
    "tags": ["go", "pocketbase", "benchmark"]
}`)

// --- Lazy Parsing (기존 방식) ---
// 비교를 위해 기존의 지연 파싱 구현을 별도 구조체로 정의합니다.

type RecordLazy struct {
	BaseModel
	CollectionID     string
	CollectionName   string
	Expand           map[string][]*Record
	Data             json.RawMessage `json:"-"`
	once             sync.Once
	deserializedData map[string]interface{}
}

func (r *RecordLazy) UnmarshalJSON(data []byte) error {
	type RecordAlias RecordLazy
	alias := &struct {
		*RecordAlias
		RawData map[string]json.RawMessage `json:"-"`
	}{
		RecordAlias: (*RecordAlias)(r),
		RawData:     make(map[string]json.RawMessage),
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&alias.RawData); err != nil {
		return err
	}

	if rawID, ok := alias.RawData["id"]; ok {
		_ = json.Unmarshal(rawID, &r.ID)
		delete(alias.RawData, "id")
	}
	if rawCreated, ok := alias.RawData["created"]; ok {
		_ = json.Unmarshal(rawCreated, &r.Created)
		delete(alias.RawData, "created")
	}
	if rawUpdated, ok := alias.RawData["updated"]; ok {
		_ = json.Unmarshal(rawUpdated, &r.Updated)
		delete(alias.RawData, "updated")
	}
	if rawColID, ok := alias.RawData["collectionId"]; ok {
		_ = json.Unmarshal(rawColID, &r.CollectionID)
		delete(alias.RawData, "collectionId")
	}
	if rawColName, ok := alias.RawData["collectionName"]; ok {
		_ = json.Unmarshal(rawColName, &r.CollectionName)
		delete(alias.RawData, "collectionName")
	}
	if rawExpand, ok := alias.RawData["expand"]; ok {
		_ = json.Unmarshal(rawExpand, &r.Expand)
		delete(alias.RawData, "expand")
	}

	remainingData, err := json.Marshal(alias.RawData)
	if err != nil {
		return err
	}
	r.Data = remainingData
	return nil
}

// ✅ BenchmarkUnmarshalLazy (기존 방식)
// 재인코딩 과정이 포함된 이전의 지연 파싱 방식의 성능을 측정합니다.
func BenchmarkUnmarshalLazy(b *testing.B) {
	b.ReportAllocs() // 메모리 할당량 보고
	for i := 0; i < b.N; i++ {
		var r RecordLazy
		if err := json.Unmarshal(sampleRecordJSON, &r); err != nil {
			b.Fatal(err)
		}
	}
}

// ✅ BenchmarkUnmarshalEager (새로운 방식)
// 제안된 새로운 방식(즉시 파싱)의 성능을 측정합니다.
func BenchmarkUnmarshalEager(b *testing.B) {
	b.ReportAllocs() // 메모리 할당량 보고
	for i := 0; i < b.N; i++ {
		var r Record // 수정된 Record 구조체를 사용
		if err := json.Unmarshal(sampleRecordJSON, &r); err != nil {
			b.Fatal(err)
		}
	}
}

// TestRecordUnmarshalInvalidExpand는 이전과 동일하게 유지합니다.
func TestRecordUnmarshalInvalidExpand(t *testing.T) {
	data := []byte(`{"id":"1","collectionId":"col","collectionName":"names","expand":"bad","foo":"bar"}`)
	var r Record
	if err := json.Unmarshal(data, &r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID != "1" {
		t.Fatalf("id mismatch: %s", r.ID)
	}
	if r.Expand != nil {
		t.Fatalf("expected nil expand: %#v", r.Expand)
	}
	if r.GetString("foo") != "bar" {
		t.Fatalf("unexpected data: %#v", r.deserializedData)
	}
}
