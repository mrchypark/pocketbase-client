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

func TestRecordGetters(t *testing.T) {
	r := &Record{}
	r.deserializedData = map[string]interface{}{
		"string_key":      "hello",
		"bool_key":        true,
		"float_key":       123.45,
		"int_key":         123,
		"json_num_key":    json.Number("123.45"),
		"datetime_key":    "2025-07-03T10:00:00.000Z",
		"string_slice_key": []interface{}{"a", "b", "c"},
		"raw_message_key": json.RawMessage(`{"a":1}`),
		"nil_key":         nil,
	}

	t.Run("GetString", func(t *testing.T) {
		if r.GetString("string_key") != "hello" {
			t.Errorf("Expected 'hello', got '%s'", r.GetString("string_key"))
		}
		if r.GetString("missing_key") != "" {
			t.Errorf("Expected empty string for missing key, got '%s'", r.GetString("missing_key"))
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		if !r.GetBool("bool_key") {
			t.Error("Expected true, got false")
		}
		if r.GetBool("missing_key") {
			t.Error("Expected false for missing key, got true")
		}
	})

	t.Run("GetFloat", func(t *testing.T) {
		if r.GetFloat("float_key") != 123.45 {
			t.Errorf("Expected 123.45, got %f", r.GetFloat("float_key"))
		}
		if r.GetFloat("int_key") != 123 {
			t.Errorf("Expected 123, got %f", r.GetFloat("int_key"))
		}
		if r.GetFloat("json_num_key") != 123.45 {
			t.Errorf("Expected 123.45 from json.Number, got %f", r.GetFloat("json_num_key"))
		}
		if r.GetFloat("missing_key") != 0 {
			t.Errorf("Expected 0 for missing key, got %f", r.GetFloat("missing_key"))
		}
	})

	t.Run("GetDateTime", func(t *testing.T) {
		dt := r.GetDateTime("datetime_key")
		if dt.Time().Year() != 2025 {
			t.Errorf("Expected year 2025, got %d", dt.Time().Year())
		}
		if !r.GetDateTime("missing_key").IsZero() {
			t.Error("Expected zero DateTime for missing key")
		}
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		slice := r.GetStringSlice("string_slice_key")
		if len(slice) != 3 || slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
			t.Errorf("Unexpected slice content: %v", slice)
		}
		if len(r.GetStringSlice("missing_key")) != 0 {
			t.Error("Expected empty slice for missing key")
		}
	})

	t.Run("GetRawMessage", func(t *testing.T) {
		raw := r.GetRawMessage("raw_message_key")
		if string(raw) != `{"a":1}` {
			t.Errorf("Unexpected raw message content: %s", string(raw))
		}
		if r.GetRawMessage("missing_key") != nil {
			t.Error("Expected nil for missing key")
		}
	})

	t.Run("GetPointerTypes", func(t *testing.T) {
		if *r.GetStringPointer("string_key") != "hello" {
			t.Error("GetStringPointer failed")
		}
		if r.GetStringPointer("missing_key") != nil {
			t.Error("GetStringPointer should return nil for missing key")
		}
		if *r.GetBoolPointer("bool_key") != true {
			t.Error("GetBoolPointer failed")
		}
		if r.GetBoolPointer("missing_key") != nil {
			t.Error("GetBoolPointer should return nil for missing key")
		}
		if *r.GetFloatPointer("float_key") != 123.45 {
			t.Error("GetFloatPointer failed for float")
		}
		if *r.GetFloatPointer("int_key") != 123 {
			t.Error("GetFloatPointer failed for int")
		}
		if r.GetFloatPointer("missing_key") != nil {
			t.Error("GetFloatPointer should return nil for missing key")
		}
		if r.GetDateTimePointer("datetime_key").Time().Year() != 2025 {
			t.Error("GetDateTimePointer failed")
		}
		if r.GetDateTimePointer("missing_key") != nil {
			t.Error("GetDateTimePointer should return nil for missing key")
		}
	})
}

func TestRecordSet(t *testing.T) {
	r := &Record{}
	r.Set("my_key", "my_value")
	if r.Get("my_key") != "my_value" {
		t.Error("Set/Get failed to store and retrieve a value")
	}
}