package pocketbase_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/mrchypark/pocketbase-client"
)

// TestPocketBaseContract is opt-in: it never uses an existing PocketBase database.
func TestPocketBaseContract(t *testing.T) {
	bin := os.Getenv("POCKETBASE_BIN")
	if bin == "" {
		t.Skip("set POCKETBASE_BIN to PocketBase 0.39.10")
	}
	bin, err := filepath.Abs(bin)
	if err != nil {
		t.Fatal(err)
	}
	version, err := exec.Command(bin, "--version").CombinedOutput()
	if err != nil || !strings.Contains(string(version), "0.39.10") {
		t.Fatalf("expected PocketBase 0.39.10: %s (%v)", version, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	password := "contract-password-123"
	setup := exec.CommandContext(ctx, bin, "superuser", "upsert", "contract@example.com", password, "--dir", data)
	setup.Dir = dir
	if out, err := setup.CombinedOutput(); err != nil {
		t.Fatalf("superuser: %v\n%s", err, out)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	var logs bytes.Buffer
	server := exec.CommandContext(ctx, bin, "serve", "--http", addr, "--dir", data)
	server.Dir = dir
	server.Stdout, server.Stderr = &logs, &logs
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = server.Process.Kill()
		_ = server.Wait()
		if t.Failed() {
			t.Log(logs.String())
		}
	})
	base := "http://" + addr
	health := &http.Client{Timeout: time.Second}
	ready := false
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(25 * time.Millisecond) {
		res, err := health.Get(base + "/api/health")
		if err == nil {
			res.Body.Close()
			if res.StatusCode == 200 {
				ready = true
				break
			}
		}
	}
	if !ready {
		t.Fatal("PocketBase did not become ready")
	}
	client := pb.NewClient(base)
	auth, err := client.WithAdminPassword(ctx, "contract@example.com", password)
	if err != nil {
		t.Fatal(err)
	}
	if auth.Record == nil || auth.Admin == nil || auth.Record.CollectionName != "_superusers" {
		t.Fatal("superuser auth must contain both Record and legacy Admin")
	}
	admin, err := client.Admins.GetOne(ctx, auth.Record.ID)
	if err != nil || admin == nil || admin.Email != "contract@example.com" {
		t.Fatalf("modern admin adapter: %+v %v", admin, err)
	}
	// Decode the modern wire fixture so this test also detects lost field metadata.
	var collection pb.Collection
	if err := json.Unmarshal([]byte(`{"name":"contract_users","type":"auth","fields":[{"name":"label","type":"text"},{"name":"active","type":"bool"},{"name":"score","type":"number"},{"name":"tags","type":"select","maxSelect":2,"values":["a","b"]}],"indexes":["CREATE INDEX idx_contract_label ON contract_users (label)"],"listRule":null,"viewRule":null,"createRule":null,"updateRule":null,"deleteRule":null}`), &collection); err != nil {
		t.Fatal(err)
	}
	created, err := client.Collections.Create(ctx, &collection)
	if err != nil {
		t.Fatalf("create collection: %#v", err)
	}
	wire, err := json.Marshal(created)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(wire, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["listRule"]) != "null" || len(raw["fields"]) == 0 || !bytes.Contains(raw["indexes"], []byte("idx_contract_label")) {
		t.Fatalf("modern collection wire: %s", wire)
	}
	// Empty rule means public; nil means locked. Verify both on the actual server.
	var public pb.Collection
	publicWire := bytes.Replace(wire, []byte(`"listRule":null`), []byte(`"listRule":""`), 1)
	if err := json.Unmarshal(publicWire, &public); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Collections.Update(ctx, created.Name, &public); err != nil {
		t.Fatal(err)
	}
	if _, err := pb.NewClient(base).Records.GetList(ctx, created.Name, nil); err != nil {
		t.Fatalf("empty list rule should be public: %v", err)
	}
	if _, err := client.Collections.Import(ctx, []*pb.Collection{created}, false); err != nil {
		t.Fatalf("collection import 204: %v", err)
	}
	if _, err := pb.NewClient(base).Records.GetList(ctx, created.Name, nil); err == nil {
		t.Fatal("nil list rule should be locked")
	}
	// Generate into a temporary module, then run the generated service against PB.
	module := filepath.Join(dir, "generated")
	if err := os.Mkdir(module, 0700); err != nil {
		t.Fatal(err)
	}
	schema := filepath.Join(module, "schema.json")
	if err := os.WriteFile(schema, wire, 0600); err != nil {
		t.Fatal(err)
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	gen := exec.CommandContext(ctx, "go", "run", "./cmd/pbc-gen", "-schema", schema, "-path", filepath.Join(module, "models.go"), "-pkgname", "contract")
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("generate: %v\n%s", err, out)
	}
	mod := fmt.Sprintf("module contract\n\ngo 1.26\nrequire github.com/mrchypark/pocketbase-client v0.0.0\nreplace github.com/mrchypark/pocketbase-client => %s\n", filepath.ToSlash(root))
	if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte(mod), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(module, "contract_test.go"), []byte(generatedContractTest), 0600); err != nil {
		t.Fatal(err)
	}
	check := exec.CommandContext(ctx, "go", "test", "-mod=mod", "-count=1", "-v", ".")
	check.Dir = module
	check.Env = append(os.Environ(), "PBC_CONTRACT_URL="+base, "PBC_CONTRACT_TOKEN="+auth.Token, "GOWORK=off")
	if out, err := check.CombinedOutput(); err != nil {
		t.Fatalf("generated contracts: %v\n%s", err, out)
	} else {
		t.Log(string(out))
	}
}

const generatedContractTest = `package contract
import (
 "context"
 "encoding/json"
 "fmt"
 "os"
 "testing"
 pb "github.com/mrchypark/pocketbase-client"
)
func TestGeneratedAuthCRUD(t *testing.T) {
 ctx := context.Background()
 client := pb.NewClient(os.Getenv("PBC_CONTRACT_URL"))
 client.WithToken(os.Getenv("PBC_CONTRACT_TOKEN"))
 service := NewContractUsersService(client)
 user := NewContractUsers()
 user.SetEmail("user@example.com")
 user.SetPassword("contract-password-123")
 user.SetPasswordConfirm("contract-password-123")
 user.SetLabel("keep")
 user.SetActive(true)
 user.SetScore(7)
 user.SetTags([]string{"a"})
 encoded, err := json.Marshal(user)
 if err != nil { t.Fatal(err) }
 var raw map[string]any
 if err := json.Unmarshal(encoded, &raw); err != nil { t.Fatal(err) }
 if _, ok := raw["password"]; ok { t.Fatal("password leaked into JSON") }
 created, err := service.Create(ctx, user)
 if err != nil { t.Fatal(err) }
 if created.ID == "" || created.Password != nil || created.PasswordConfirm != nil { t.Fatalf("auth read leaked write-only fields: %+v", created) }
 sparse := NewContractUsers()
 sparse.SetLabel("keep")
 preserved, err := service.Update(ctx,created.ID,sparse)
 if err != nil { t.Fatalf("display-only auth PATCH must preserve email: %v",err) }
 if preserved.Email == nil || *preserved.Email != "user@example.com" { t.Fatalf("email reset by sparse PATCH: %+v",preserved) }
 session := pb.NewClient(os.Getenv("PBC_CONTRACT_URL"))
 auth, err := session.WithPassword(ctx, "contract_users", "user@example.com", "contract-password-123")
 if err != nil { t.Fatal(err) }
 refreshed, err := session.Users.AuthRefresh(ctx, "contract_users")
 if err != nil { t.Fatal(err) }
 session.UseAuthResponse(refreshed)
 if auth.Record.ID != created.ID { t.Fatal("wrong authenticated user") }
 patch := NewContractUsers()
 patch.SetActive(false)
 patch.SetScore(0)
 patch.SetTags([]string{})
 updated, err := service.Update(ctx, created.ID, patch)
 if err != nil { t.Fatal(err) }
 if updated.Active == nil || *updated.Active || updated.Score == nil || *updated.Score != 0 || len(updated.Tags) != 0 || updated.Label == nil || *updated.Label != "keep" { t.Fatalf("zero PATCH/omission: %+v", updated) }
 read, err := service.GetOne(ctx, created.ID, nil)
 if err != nil || read.ID != created.ID { t.Fatalf("generated read: %+v %v", read, err) }
 // Base records keep this pagination fixture cheap (no 1000 password hashes).
 _, err = client.SendRaw(ctx, "POST", "/api/collections", map[string]any{"name":"contract_pages", "type":"base", "fields":[]any{map[string]any{"name":"label","type":"text"},map[string]any{"name":"owner","type":"relation","collectionId":created.CollectionID,"maxSelect":1}}})
 if err != nil { t.Fatal(err) }
 for i:=0; i<1001; i++ {
  if _, err := client.Records.Create(ctx,"contract_pages",map[string]any{"label":fmt.Sprint(i)}); err != nil { t.Fatal(err) }
 }
 fixture, err := client.Records.GetList(ctx,"contract_pages",&pb.ListOptions{PerPage:1})
 if err != nil { t.Fatal(err) }
 id := fixture.Items[0].ID
 if _, err := client.Records.Update(ctx,"contract_pages",id,map[string]any{"owner":created.ID}); err != nil { t.Fatal(err) }
 expanded, err := client.Records.GetOne(ctx,"contract_pages",id,&pb.GetOneOptions{Expand:"owner"})
 if err != nil || len(expanded.Expand["owner"])!=1 || expanded.Expand["owner"][0].ID!=created.ID { t.Fatalf("single relation expand: %+v %v",expanded,err) }
 pages := pb.NewTypedRecordService[struct{ID string ` + "`json:\"id\"`" + `}](client,"contract_pages")
 for _, skip := range []bool{false,true} {
  all, err := pages.GetAll(ctx,&pb.ListOptions{PerPage:2000,SkipTotal:skip,Sort:"id"})
  if err != nil || len(all)!=1001 { t.Fatalf("server-clamped pagination skipTotal=%v: %d %v",skip,len(all),err) }
 }
 if err := service.Delete(ctx,created.ID); err != nil { t.Fatal(err) }
 if _, err := service.GetOne(ctx,created.ID,nil); err == nil { t.Fatal("deleted record still readable") }
}
`
