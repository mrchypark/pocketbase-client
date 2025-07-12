package pocketbase

import (
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

// BenchmarkUnmarshalCurrent (기존 방식)은 비교를 위해 남겨둡니다.
// 이 테스트를 실행하려면 이전 버전의 Record 구조체와 UnmarshalJSON이 필요합니다.
// func BenchmarkUnmarshalCurrent(b *testing.B) { ... }

// BenchmarkUnmarshalLazy measures the performance of the new lazy-parsing UnmarshalJSON.
// (Performs one unmarshaling operation for top-level fields, delaying the 'Data' field)
func BenchmarkUnmarshalLazy(b *testing.B) {
	b.ReportAllocs() // 메모리 할당량 보고
	for i := 0; i < b.N; i++ {
		var r Record
		if err := json.Unmarshal(sampleRecordJSON, &r); err != nil {
			b.Fatal(err)
		}
	}
}

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

	// Access data to trigger lazy parsing
	if r.GetString("foo") != "bar" {
		r.parseData() // Manually parse for debugging
		t.Fatalf("unexpected data: %#v", r.deserializedData)
	}
}
