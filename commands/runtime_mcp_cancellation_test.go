package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestRuntimeMCPCancellationCLIHasEmptyAcknowledgement(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/mcp/http" || r.Header.Get("Authorization") != "Bearer runtime-grant" {
			t.Errorf("request=%s headers=%v", r.URL.Path, r.Header)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if _, exists := body["id"]; exists {
			t.Error("notification contains id")
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace-1", false, true, true)
	ctx.Endpoint = server.URL + "/newmoi"
	grant := filepath.Join(t.TempDir(), "grant")
	if err := os.WriteFile(grant, []byte("runtime-grant"), 0600); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteRuntimeGatewayCommand(ctx, []string{"mcp", "--runtime-token-file", grant, "--json", `{"jsonrpc":"2.0","method":"notifications/cancelled","params":{"requestId":"exchange","reason":"test"}}`})
	os.Stdout = original
	_ = write.Close()
	output, _ := io.ReadAll(read)
	_ = read.Close()
	if callErr != nil {
		t.Fatal(callErr)
	}
	if len(output) != 0 {
		t.Fatalf("notification fabricated output: %q", output)
	}
}
