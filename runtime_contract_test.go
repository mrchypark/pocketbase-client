package pocketbase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Run explicitly with POCKETBASE_BIN pointing to PocketBase v0.39.10.
func TestRuntimePocketBaseFileModifiers(t *testing.T) {
	binary := os.Getenv("POCKETBASE_BIN")
	if binary == "" {
		t.Skip("set POCKETBASE_BIN for isolated server verification")
	}
	dir := t.TempDir()
	if out, err := exec.Command(binary, "superuser", "create", "runtime@example.com", "runtime-password-123", "--dir", dir).CombinedOutput(); err != nil {
		t.Fatalf("superuser: %v: %s", err, out)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	command := exec.Command(binary, "serve", "--http", address, "--dir", dir)
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = command.Process.Kill(); _ = command.Wait() }()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := NewClient("http://" + address)
	for {
		if _, err := client.SendRaw(ctx, http.MethodGet, "/api/health", nil); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(20 * time.Millisecond):
		}
	}
	data, err := client.SendRaw(ctx, http.MethodPost, "/api/collections/_superusers/auth-with-password", map[string]string{"identity": "runtime@example.com", "password": "runtime-password-123"})
	if err != nil {
		t.Fatal(err)
	}
	var auth AuthResponse
	if err := json.Unmarshal(data, &auth); err != nil {
		t.Fatal(err)
	}
	client.WithToken(auth.Token)
	_, err = client.SendRaw(ctx, http.MethodPost, "/api/collections", map[string]any{"name": "runtime_files", "type": "base", "fields": []map[string]any{{"name": "single", "type": "file", "maxSelect": 1}, {"name": "many", "type": "file", "maxSelect": 5}}})
	if err != nil {
		t.Fatal(err)
	}
	record, err := client.Records.Create(ctx, "runtime_files", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"single", "many"} {
		got, err := client.Files.Upload(ctx, "runtime_files", record.ID, field, "remove.txt", strings.NewReader("remove"))
		if err != nil {
			t.Fatal(err)
		}
		filename := got.GetString(field)
		if field == "many" {
			filename = got.GetStringSlice(field)[0]
			if _, err := client.Files.Upload(ctx, "runtime_files", record.ID, field+"+", "keep.txt", strings.NewReader("keep")); err != nil {
				t.Fatal(err)
			}
		}
		got, err = client.Files.Delete(ctx, "runtime_files", record.ID, field, filename)
		if err != nil {
			t.Fatal(err)
		}
		if field == "single" && got.GetString(field) != "" {
			t.Fatalf("single file remained: %v", got.Get(field))
		}
		if field == "many" && len(got.GetStringSlice(field)) != 1 {
			t.Fatalf("multiple files: %v", got.Get(field))
		}
	}
}

func TestRuntimeGetAllServerPageSize(t *testing.T) {
	for _, skip := range []bool{false, true} {
		t.Run(fmt.Sprint(skip), func(t *testing.T) {
			var requests atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.Query().Get("perPage") != "2000" {
					t.Error("requested page size changed")
				}
				page, _ := strconv.Atoi(r.URL.Query().Get("page"))
				items := make([]map[string]any, 0)
				for i := (page - 1) * 1000; i < page*1000 && i < 1001; i++ {
					items = append(items, map[string]any{"id": fmt.Sprint(i)})
				}
				totalPages := 2
				if skip {
					totalPages = -1
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"page": page, "perPage": 1000, "totalPages": totalPages, "items": items})
			}))
			defer srv.Close()
			got, err := NewTypedRecordService[Record](NewClient(srv.URL), "posts").GetAll(context.Background(), &ListOptions{PerPage: 2000, SkipTotal: skip})
			if err != nil || len(got) != 1001 || requests.Load() != 2 {
				t.Fatalf("records=%d requests=%d err=%v", len(got), requests.Load(), err)
			}
		})
	}
}

func TestRuntimeRecordExpandShapes(t *testing.T) {
	var record Record
	if err := json.Unmarshal([]byte(`{"expand":{"single":{"id":"one","expand":{"nested":{"id":"two"}}},"many":[{"id":"three"}],"empty":[]}}`), &record); err != nil {
		t.Fatal(err)
	}
	if len(record.Expand["single"]) != 1 || record.Expand["single"][0].ID != "one" || record.Expand["single"][0].Expand["nested"][0].ID != "two" || record.Expand["many"][0].ID != "three" {
		t.Fatalf("unexpected expand: %#v", record.Expand)
	}
	for _, shape := range []string{`"bad"`, `42`, `[]`, `{"rel":false}`, `{"rel":null}`, `{"rel":[null]}`, `{"rel":[42]}`, `{"rel":{"expand":{"nested":false}}}`} {
		if err := json.Unmarshal([]byte(`{"expand":`+shape+`}`), &record); err == nil {
			t.Errorf("accepted invalid expand %s", shape)
		}
	}
}

func TestRuntimeFileDeleteModifier(t *testing.T) {
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if r.Method != http.MethodPatch {
			t.Errorf("method=%s; want PATCH only", r.Method)
		}
		var body map[string][]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body) != 1 || len(body["image-"]) != 1 || body["image-"][0] != "remove.txt" {
			t.Errorf("body=%v err=%v", body, err)
		}
		_, _ = io.WriteString(w, `{"id":"record","image":["keep.txt","concurrent.txt"]}`)
	}))
	defer srv.Close()
	got, err := NewClient(srv.URL).Files.Delete(context.Background(), "posts", "record", "image", "remove.txt")
	if err != nil || got == nil || len(got.GetStringSlice("image")) != 2 || requests.Load() != 1 {
		t.Fatalf("record=%v requests=%d err=%v", got, requests.Load(), err)
	}
}

func TestRuntimeRealtimeReconnect(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(fmt.Sprint(fail), func(t *testing.T) {
			var connections, posts atomic.Int32
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					n := posts.Add(1)
					var body struct {
						ClientID string `json:"clientId"`
					}
					_ = json.NewDecoder(r.Body).Decode(&body)
					if body.ClientID != fmt.Sprint(n) {
						t.Errorf("clientId=%q post=%d", body.ClientID, n)
					}
					if fail && n == 3 {
						http.Error(w, "denied", http.StatusForbidden)
						return
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}
				n := connections.Add(1)
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = fmt.Fprintf(w, "retry: 1\nevent: PB_CONNECT\ndata: {\"clientId\":\"%d\"}\n\n", n)
				w.(http.Flusher).Flush()
				if n >= 3 {
					_, _ = io.WriteString(w, "data: {\"action\":\"update\"}\n\n")
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				}
			}))
			defer srv.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			done := make(chan error, 4)
			unsub, err := NewClient(srv.URL).Realtime.Subscribe(ctx, []string{"posts/*"}, func(event *RealtimeEvent, err error) {
				if err != nil || (event != nil && event.Action == "update") {
					done <- err
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unsub()
			select {
			case err := <-done:
				if (err != nil) != fail || posts.Load() != 3 {
					t.Fatalf("posts=%d err=%v", posts.Load(), err)
				}
			case <-ctx.Done():
				t.Fatalf("reconnect stalled: connections=%d posts=%d", connections.Load(), posts.Load())
			}
		})
	}
}
