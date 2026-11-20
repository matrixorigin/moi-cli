package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"golang.org/x/net/websocket"
)

func TestSandboxConnectTerminalCommand(t *testing.T) {
	handshake := make(chan *http.Request, 1)
	closed := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handshake <- r.Clone(r.Context())
		websocket.Handler(func(connection *websocket.Conn) {
			defer connection.Close()
			var data []byte
			_ = websocket.Message.Receive(connection, &data)
			closed <- struct{}{}
		}).ServeHTTP(w, r)
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
	if err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteSandboxCommand(NewContext(client, "workspace-1", false, true, true), []string{"connect-terminal", "--sandbox-id", "sandbox-1", "--token", "signed token", "--cols", "100", "--rows", "30"})
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
	if err := json.Unmarshal(output, &result); err != nil || !reflect.DeepEqual(result, map[string]any{"connected": true}) {
		t.Fatalf("output=%s decoded=%v error=%v", output, result, err)
	}
	select {
	case request := <-handshake:
		if request.URL.Path != "/newmoi/agent-sandboxes/sandbox-1/terminal/ws" {
			t.Fatalf("path=%s", request.URL.Path)
		}
		want := url.Values{"token": {"signed token"}, "cols": {"100"}, "rows": {"30"}}
		if !reflect.DeepEqual(request.URL.Query(), want) {
			t.Fatalf("query=%v", request.URL.Query())
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
