package pocketbase

import (
	"testing"

	"github.com/goccy/go-json"
)

// 테스트에 사용할 샘플 JSON 데이터입니다.
// 실제 데이터처럼 기본 필드와 추가 데이터 필드를 포함합니다.
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

// BenchmarkUnmarshalCurrent는 현재 구현된 UnmarshalJSON의 성능을 측정합니다.
// (JSON -> Struct, JSON -> Map 두 번의 언마샬링 수행)
func BenchmarkUnmarshalCurrent(b *testing.B) {
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
	if r.Data["foo"] != "bar" {
		t.Fatalf("unexpected data: %#v", r.Data)
	}
}
