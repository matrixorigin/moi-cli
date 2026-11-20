package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestAIStudioCLIMatchesSDKHTTPContracts(t *testing.T) {
	type recorded struct {
		method, path, query, auth, workspace string
		body                                 map[string]any
	}
	var calls []recorded
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		calls = append(calls, recorded{r.Method, r.URL.EscapedPath(), r.URL.RawQuery, r.Header.Get("Authorization"), r.Header.Get("X-Workspace-ID"), body})
		switch {
		case r.URL.Path == "/newmoi/admin/aistudio/service-slots/forbidden":
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"code":"ErrForbidden","msg":"需要 taas:admin"}`))
		case r.URL.Path == "/newmoi/aistudio/model-policy":
			_, _ = w.Write([]byte(`{"code":"OK","data":{"default_model":"second","models":["second","first"]}}`))
		case r.Method == http.MethodPut:
			_, _ = w.Write([]byte(`{"slot_id":"chat","default_model_id":"m1","model_ids":[]}`))
		case r.URL.Path == "/newmoi/admin/aistudio/service-slots":
			_, _ = w.Write([]byte(`{"slots":[{"slot_id":"chat"}]}`))
		case r.URL.Path == "/newmoi/admin/aistudio/service-slots/candidates":
			_, _ = w.Write([]byte(`{"models":[{"id":"m1"}]}`))
		default:
			_, _ = w.Write([]byte(`{"logs":[]}`))
		}
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("admin-token"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace-must-not-leak", false, true, true)
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	commands := [][]string{
		{"chat-model-policy"}, {"service-slots"}, {"service-slot-candidates"},
		{"service-slot-logs", "--slot-id", "chat/1", "--page-size", "5"},
		{"service-slot-logs"},
		{"put-service-slot", "--slot-id", "chat/1", "--json", `{"default_model_id":"m1"}`},
		{"put-service-slot", "--slot-id", "chat", "--json", `{"slot_id":"chat","default_model_id":"m1","model_ids":["m1","m2"],"extra_json":"{\"a\":1}"}`},
	}
	for _, args := range commands {
		if err := ExecuteAIStudioCommand(ctx, args); err != nil {
			os.Stdout = original
			t.Fatalf("%v: %v", args, err)
		}
	}
	os.Stdout = original
	_ = write.Close()
	output, _ := io.ReadAll(read)
	_ = read.Close()
	want := []recorded{
		{"GET", "/newmoi/aistudio/model-policy", "", "Bearer admin-token", "", nil},
		{"GET", "/newmoi/admin/aistudio/service-slots", "", "Bearer admin-token", "", nil},
		{"GET", "/newmoi/admin/aistudio/service-slots/candidates", "", "Bearer admin-token", "", nil},
		{"GET", "/newmoi/admin/aistudio/service-slots/operation-logs", "page_size=5&slot_id=chat%2F1", "Bearer admin-token", "", nil},
		{"GET", "/newmoi/admin/aistudio/service-slots/operation-logs", "", "Bearer admin-token", "", nil},
		{"PUT", "/newmoi/admin/aistudio/service-slots/chat%2F1", "", "Bearer admin-token", "", map[string]any{"default_model_id": "m1", "model_ids": []any{}, "extra_json": ""}},
		{"PUT", "/newmoi/admin/aistudio/service-slots/chat", "", "Bearer admin-token", "", map[string]any{"slot_id": "chat", "default_model_id": "m1", "model_ids": []any{"m1", "m2"}, "extra_json": "{\"a\":1}"}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%#v\nwant=%#v", calls, want)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	for index := range commands {
		var result any
		if err := decoder.Decode(&result); err != nil {
			t.Fatalf("output %d=%s error=%v", index, output, err)
		}
	}
	err = ExecuteAIStudioCommand(ctx, []string{"put-service-slot", "--slot-id", "forbidden", "--json", `{"default_model_id":"m1"}`})
	var failure *sdk.Error
	if !errors.As(err, &failure) || failure.Code != "ErrForbidden" || failure.Message != "需要 taas:admin" {
		t.Fatalf("error=%#v", err)
	}
}
