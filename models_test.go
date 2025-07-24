package pocketbase

import (
	"bytes"
	"sync"
	"testing"

	"github.com/goccy/go-json"
	"github.com/pocketbase/pocketbase/tools/types"
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

// --- Lazy Parsing (existing approach) ---
// Define the existing lazy parsing implementation as a separate struct for comparison.

type RecordLazy struct {
	BaseModel
	BaseDateTime
	Expand           map[string][]*Record
	Data             json.RawMessage `json:"-"`
	once             sync.Once
	deserializedData map[string]any
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

// ✅ BenchmarkUnmarshalLazy (existing approach)
// Measure performance of the previous lazy parsing approach that includes re-encoding process.
func BenchmarkUnmarshalLazy(b *testing.B) {
	b.ReportAllocs() // Report memory allocations
	for i := 0; i < b.N; i++ {
		var r RecordLazy
		if err := json.Unmarshal(sampleRecordJSON, &r); err != nil {
			b.Fatal(err)
		}
	}
}

// ✅ BenchmarkUnmarshalEager (new approach)
// Measure performance of the proposed new approach (eager parsing).
func BenchmarkUnmarshalEager(b *testing.B) {
	b.ReportAllocs() // Report memory allocations
	for i := 0; i < b.N; i++ {
		var r Record // Use modified Record struct
		if err := json.Unmarshal(sampleRecordJSON, &r); err != nil {
			b.Fatal(err)
		}
	}
}

// TestRecordUnmarshalInvalidExpand is kept the same as before.
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
	r.deserializedData = map[string]any{
		"string_key":       "hello",
		"bool_key":         true,
		"float_key":        123.45,
		"int_key":          123,
		"json_num_key":     json.Number("123.45"),
		"datetime_key":     "2025-07-03T10:00:00.000Z",
		"string_slice_key": []any{"a", "b", "c"},
		"raw_message_key":  json.RawMessage(`{"a":1}`),
		"nil_key":          nil,
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

// TestBaseModel tests the BaseModel structure and its JSON serialization/deserialization.
func TestBaseModel(t *testing.T) {
	t.Run("BaseModel fields", func(t *testing.T) {
		base := BaseModel{
			ID:             "test_id_123",
			CollectionID:   "collection_456",
			CollectionName: "test_collection",
		}

		// Test field access
		if base.ID != "test_id_123" {
			t.Errorf("Expected ID 'test_id_123', got '%s'", base.ID)
		}
		if base.CollectionID != "collection_456" {
			t.Errorf("Expected CollectionID 'collection_456', got '%s'", base.CollectionID)
		}
		if base.CollectionName != "test_collection" {
			t.Errorf("Expected CollectionName 'test_collection', got '%s'", base.CollectionName)
		}
	})

	t.Run("BaseModel JSON serialization", func(t *testing.T) {
		base := BaseModel{
			ID:             "test_id_123",
			CollectionID:   "collection_456",
			CollectionName: "test_collection",
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(base)
		if err != nil {
			t.Fatalf("Failed to marshal BaseModel: %v", err)
		}

		expectedJSON := `{"id":"test_id_123","collectionId":"collection_456","collectionName":"test_collection"}`
		if string(jsonData) != expectedJSON {
			t.Errorf("Expected JSON %s, got %s", expectedJSON, string(jsonData))
		}

		// Test JSON unmarshaling
		var unmarshaled BaseModel
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal BaseModel: %v", err)
		}

		if unmarshaled.ID != base.ID {
			t.Errorf("Expected ID '%s', got '%s'", base.ID, unmarshaled.ID)
		}
		if unmarshaled.CollectionID != base.CollectionID {
			t.Errorf("Expected CollectionID '%s', got '%s'", base.CollectionID, unmarshaled.CollectionID)
		}
		if unmarshaled.CollectionName != base.CollectionName {
			t.Errorf("Expected CollectionName '%s', got '%s'", base.CollectionName, unmarshaled.CollectionName)
		}
	})
}

// TestBaseDateTime tests the BaseDateTime structure and its JSON serialization/deserialization.
func TestBaseDateTime(t *testing.T) {
	t.Run("BaseDateTime fields", func(t *testing.T) {
		// Create test datetime values
		createdTime := "2025-07-20T10:30:00.123Z"
		updatedTime := "2025-07-20T11:45:00.456Z"

		created, err := types.ParseDateTime(createdTime)
		if err != nil {
			t.Fatalf("Failed to parse created time: %v", err)
		}

		updated, err := types.ParseDateTime(updatedTime)
		if err != nil {
			t.Fatalf("Failed to parse updated time: %v", err)
		}

		baseTime := BaseDateTime{
			Created: created,
			Updated: updated,
		}

		// Test field access - types.DateTime uses space format, not T format
		expectedCreatedStr := "2025-07-20 10:30:00.123Z"
		expectedUpdatedStr := "2025-07-20 11:45:00.456Z"
		if baseTime.Created.String() != expectedCreatedStr {
			t.Errorf("Expected Created '%s', got '%s'", expectedCreatedStr, baseTime.Created.String())
		}
		if baseTime.Updated.String() != expectedUpdatedStr {
			t.Errorf("Expected Updated '%s', got '%s'", expectedUpdatedStr, baseTime.Updated.String())
		}
	})

	t.Run("BaseDateTime JSON serialization", func(t *testing.T) {
		createdTime := "2025-07-20T10:30:00.123Z"
		updatedTime := "2025-07-20T11:45:00.456Z"

		created, _ := types.ParseDateTime(createdTime)
		updated, _ := types.ParseDateTime(updatedTime)

		baseTime := BaseDateTime{
			Created: created,
			Updated: updated,
		}

		// Test JSON marshaling
		jsonData, err := json.Marshal(baseTime)
		if err != nil {
			t.Fatalf("Failed to marshal BaseDateTime: %v", err)
		}

		// types.DateTime uses space format in JSON, not T format
		expectedJSON := `{"created":"2025-07-20 10:30:00.123Z","updated":"2025-07-20 11:45:00.456Z"}`
		if string(jsonData) != expectedJSON {
			t.Errorf("Expected JSON %s, got %s", expectedJSON, string(jsonData))
		}

		// Test JSON unmarshaling
		var unmarshaled BaseDateTime
		err = json.Unmarshal(jsonData, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal BaseDateTime: %v", err)
		}

		if unmarshaled.Created.String() != baseTime.Created.String() {
			t.Errorf("Expected Created '%s', got '%s'", baseTime.Created.String(), unmarshaled.Created.String())
		}
		if unmarshaled.Updated.String() != baseTime.Updated.String() {
			t.Errorf("Expected Updated '%s', got '%s'", baseTime.Updated.String(), unmarshaled.Updated.String())
		}
	})
}

// TestStructEmbedding tests the embedding of BaseModel and BaseDateTime in generated structs.
func TestStructEmbedding(t *testing.T) {
	t.Run("Legacy schema struct (BaseModel + BaseDateTime)", func(t *testing.T) {
		// Simulate a legacy schema generated struct
		type LegacyPost struct {
			BaseModel
			BaseDateTime
			Title   string `json:"title"`
			Content string `json:"content"`
		}

		post := LegacyPost{
			BaseModel: BaseModel{
				ID:             "post_123",
				CollectionID:   "posts_collection",
				CollectionName: "posts",
			},
			BaseDateTime: BaseDateTime{
				Created: func() types.DateTime {
					dt, _ := types.ParseDateTime("2025-07-20T10:30:00.123Z")
					return dt
				}(),
				Updated: func() types.DateTime {
					dt, _ := types.ParseDateTime("2025-07-20T11:45:00.456Z")
					return dt
				}(),
			},
			Title:   "Test Post",
			Content: "This is a test post content",
		}

		// Test field access through embedding
		if post.ID != "post_123" {
			t.Errorf("Expected ID 'post_123', got '%s'", post.ID)
		}
		if post.CollectionName != "posts" {
			t.Errorf("Expected CollectionName 'posts', got '%s'", post.CollectionName)
		}
		if post.Title != "Test Post" {
			t.Errorf("Expected Title 'Test Post', got '%s'", post.Title)
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(post)
		if err != nil {
			t.Fatalf("Failed to marshal LegacyPost: %v", err)
		}

		// Verify JSON contains all expected fields
		var jsonMap map[string]any
		err = json.Unmarshal(jsonData, &jsonMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON to map: %v", err)
		}

		expectedFields := []string{"id", "collectionId", "collectionName", "created", "updated", "title", "content"}
		for _, field := range expectedFields {
			if _, exists := jsonMap[field]; !exists {
				t.Errorf("Expected field '%s' not found in JSON", field)
			}
		}
	})

	t.Run("Latest schema struct (BaseModel only)", func(t *testing.T) {
		// Simulate a latest schema generated struct
		type LatestPost struct {
			BaseModel
			Title   string          `json:"title"`
			Content string          `json:"content"`
			Created *types.DateTime `json:"created,omitempty"`
			Updated *types.DateTime `json:"updated,omitempty"`
		}

		createdTime, _ := types.ParseDateTime("2025-07-20T10:30:00.123Z")
		updatedTime, _ := types.ParseDateTime("2025-07-20T11:45:00.456Z")

		post := LatestPost{
			BaseModel: BaseModel{
				ID:             "post_456",
				CollectionID:   "posts_collection",
				CollectionName: "posts",
			},
			Title:   "Latest Post",
			Content: "This is a latest schema post",
			Created: &createdTime,
			Updated: &updatedTime,
		}

		// Test field access
		if post.ID != "post_456" {
			t.Errorf("Expected ID 'post_456', got '%s'", post.ID)
		}
		if post.Title != "Latest Post" {
			t.Errorf("Expected Title 'Latest Post', got '%s'", post.Title)
		}
		if post.Created == nil || post.Created.String() != "2025-07-20 10:30:00.123Z" {
			t.Errorf("Expected Created timestamp, got %v", post.Created)
		}

		// Test JSON serialization
		jsonData, err := json.Marshal(post)
		if err != nil {
			t.Fatalf("Failed to marshal LatestPost: %v", err)
		}

		// Verify JSON contains all expected fields
		var jsonMap map[string]any
		err = json.Unmarshal(jsonData, &jsonMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON to map: %v", err)
		}

		expectedFields := []string{"id", "collectionId", "collectionName", "title", "content", "created", "updated"}
		for _, field := range expectedFields {
			if _, exists := jsonMap[field]; !exists {
				t.Errorf("Expected field '%s' not found in JSON", field)
			}
		}
	})
}

// TestRecordWithNewStructure tests the Record struct with the new BaseModel structure.
func TestRecordWithNewStructure(t *testing.T) {
	t.Run("Record unmarshaling with new BaseModel", func(t *testing.T) {
		jsonData := []byte(`{
			"id": "record_789",
			"collectionId": "test_collection_id",
			"collectionName": "test_collection",
			"created": "2025-07-20T10:30:00.123Z",
			"updated": "2025-07-20T11:45:00.456Z",
			"title": "Test Record",
			"is_published": true,
			"view_count": 42
		}`)

		var record Record
		err := json.Unmarshal(jsonData, &record)
		if err != nil {
			t.Fatalf("Failed to unmarshal Record: %v", err)
		}

		// Test BaseModel fields
		if record.ID != "record_789" {
			t.Errorf("Expected ID 'record_789', got '%s'", record.ID)
		}
		if record.CollectionID != "test_collection_id" {
			t.Errorf("Expected CollectionID 'test_collection_id', got '%s'", record.CollectionID)
		}
		if record.CollectionName != "test_collection" {
			t.Errorf("Expected CollectionName 'test_collection', got '%s'", record.CollectionName)
		}

		// Test that created and updated are now in deserializedData
		if record.GetString("created") != "2025-07-20T10:30:00.123Z" {
			t.Errorf("Expected created in deserializedData, got '%s'", record.GetString("created"))
		}
		if record.GetString("updated") != "2025-07-20T11:45:00.456Z" {
			t.Errorf("Expected updated in deserializedData, got '%s'", record.GetString("updated"))
		}

		// Test other fields
		if record.GetString("title") != "Test Record" {
			t.Errorf("Expected title 'Test Record', got '%s'", record.GetString("title"))
		}
		if !record.GetBool("is_published") {
			t.Error("Expected is_published to be true")
		}
		if record.GetFloat("view_count") != 42 {
			t.Errorf("Expected view_count 42, got %f", record.GetFloat("view_count"))
		}
	})

	t.Run("Record marshaling with new BaseModel", func(t *testing.T) {
		record := &Record{
			BaseModel: BaseModel{
				ID:             "record_999",
				CollectionID:   "test_collection_id",
				CollectionName: "test_collection",
			},
		}

		// Set some data fields including timestamps
		record.Set("title", "Marshaled Record")
		record.Set("created", "2025-07-20T10:30:00.123Z")
		record.Set("updated", "2025-07-20T11:45:00.456Z")
		record.Set("is_active", true)

		jsonData, err := json.Marshal(record)
		if err != nil {
			t.Fatalf("Failed to marshal Record: %v", err)
		}

		// Verify JSON contains all expected fields
		var jsonMap map[string]any
		err = json.Unmarshal(jsonData, &jsonMap)
		if err != nil {
			t.Fatalf("Failed to unmarshal JSON to map: %v", err)
		}

		// Check BaseModel fields
		if jsonMap["id"] != "record_999" {
			t.Errorf("Expected id 'record_999', got %v", jsonMap["id"])
		}
		if jsonMap["collectionId"] != "test_collection_id" {
			t.Errorf("Expected collectionId 'test_collection_id', got %v", jsonMap["collectionId"])
		}
		if jsonMap["collectionName"] != "test_collection" {
			t.Errorf("Expected collectionName 'test_collection', got %v", jsonMap["collectionName"])
		}

		// Check data fields
		if jsonMap["title"] != "Marshaled Record" {
			t.Errorf("Expected title 'Marshaled Record', got %v", jsonMap["title"])
		}
		if jsonMap["created"] != "2025-07-20T10:30:00.123Z" {
			t.Errorf("Expected created timestamp, got %v", jsonMap["created"])
		}
		if jsonMap["updated"] != "2025-07-20T11:45:00.456Z" {
			t.Errorf("Expected updated timestamp, got %v", jsonMap["updated"])
		}
		if jsonMap["is_active"] != true {
			t.Errorf("Expected is_active true, got %v", jsonMap["is_active"])
		}
	})
}
