package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"golang.org/x/net/websocket"
)

func TestSQLConnectQueryWebSocketCommandDoesNotExposeCredentials(t *testing.T) {
	const secret = "synthetic-key"
	handshake := make(chan *http.Request, 1)
	closed := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handshake <- r.Clone(r.Context())
		websocket.Server{
			Handshake: func(config *websocket.Config, _ *http.Request) error {
				config.Protocol = []string{"moi-api-key"}
				return nil
			},
			Handler: func(connection *websocket.Conn) {
				defer connection.Close()
				var data []byte
				_ = websocket.Message.Receive(connection, &data)
				closed <- struct{}{}
			},
		}.ServeHTTP(w, r)
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey(secret))
	if err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteSQLCommand(NewContext(client, "workspace-1", false, true, true), []string{"connect-query-web-socket"})
	os.Stdout = original
	_ = write.Close()
	output, readErr := io.ReadAll(read)
	_ = read.Close()
	if callErr != nil {
		t.Fatal(callErr)
	}
	if readErr != nil {
		t.Fatal(readErr)
	}
	var result map[string]any
	want := map[string]any{"protocols": []any{"moi-api-key"}}
	if err := json.Unmarshal(output, &result); err != nil || !reflect.DeepEqual(result, want) {
		t.Fatalf("output=%s decoded=%v error=%v", output, result, err)
	}
	if strings.Contains(string(output), secret) {
		t.Fatalf("WebSocket command exposed its credential: %s", output)
	}
	select {
	case request := <-handshake:
		if request.URL.Path != "/newmoi/query/ws" || !reflect.DeepEqual(request.URL.Query(), url.Values{"workspace_id": {"workspace-1"}}) {
			t.Fatalf("handshake=%s", request.URL)
		}
		if request.Header.Get("X-API-Key") != secret || request.Header.Get("X-Workspace-ID") != "workspace-1" {
			t.Fatalf("headers=%v", request.Header)
		}
		if got := request.Header.Get("Sec-WebSocket-Protocol"); !strings.Contains(got, "moi-api-key."+secret) {
			t.Fatalf("handshake did not exercise credential subprotocol: %q", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no WebSocket handshake")
	}
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("connection was not closed")
	}
}
