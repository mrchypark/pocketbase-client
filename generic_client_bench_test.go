package pocketbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
)

// BenchmarkDecodeDirect measures a single-pass decode straight into the target
// struct — the approach used by TypedRecordService after the direct-decoding
// refactor.
func BenchmarkDecodeDirect(b *testing.B) {
	data := []byte(`{"id":"rec1","name":"one"}`)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var m testModel
		if err := json.Unmarshal(data, &m); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeViaRecord measures the legacy typed path: dynamic Record decode
// (map[string]any) followed by a JSON marshal/unmarshal round-trip into the
// target struct.
func BenchmarkDecodeViaRecord(b *testing.B) {
	data := []byte(`{"id":"rec1","name":"one"}`)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var rec Record
		if err := json.Unmarshal(data, &rec); err != nil {
			b.Fatal(err)
		}
		enc, err := json.Marshal(rec)
		if err != nil {
			b.Fatal(err)
		}
		var m testModel
		if err := json.Unmarshal(enc, &m); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkTypedServiceGetOne measures the full typed GetOne request with the
// direct-decode path.
func BenchmarkTypedServiceGetOne(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "rec1", "name": "one"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	svc := NewTypedRecordService[testModel](client, "tests")
	ctx := context.Background()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := svc.GetOne(ctx, "rec1", nil); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDynamicServiceGetOne measures the dynamic RecordService.GetOne path
// (returns *Record) for comparison.
func BenchmarkDynamicServiceGetOne(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "rec1", "name": "one"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL)
	ctx := context.Background()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := client.Records.GetOne(ctx, "tests", "rec1", nil); err != nil {
			b.Fatal(err)
		}
	}
}
