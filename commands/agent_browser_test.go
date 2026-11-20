package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"golang.org/x/net/websocket"
)

func TestAgentBrowserTokenCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.EscapedPath() != "/newmoi/workspaces/workspace%20%2F1/agent-runtime-tasks/task%2F1%20%3F/browser/token" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace /1" {
			t.Errorf("request=%s %s headers=%v", r.Method, r.URL.EscapedPath(), r.Header)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) != "{}" {
			t.Errorf("body=%q error=%v", body, err)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"token": "moi-browser-token-v1.signed", "expires_at": "2026-11-10T12:00:00Z"}})
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
	if err != nil {
		t.Fatal(err)
	}
	output, err := executeAgentBrowserCLI(NewContext(client, "workspace /1", false, true, true), []string{"runtime-browser-token", "--task-id", "task/1 ?"})
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil || result["token"] != "moi-browser-token-v1.signed" || result["expires_at"] != "2026-11-10T12:00:00Z" {
		t.Fatalf("output=%s result=%v error=%v", output, result, err)
	}
}

func TestAgentBrowserConnectCommand(t *testing.T) {
	const (
		token  = "moi-browser-token-v1.signed"
		origin = "https://console.example.com"
	)
	handshake := make(chan *http.Request, 1)
	closed := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handshake <- r.Clone(r.Context())
		websocket.Server{
			Handshake: func(config *websocket.Config, _ *http.Request) error {
				config.Protocol = []string{"moi-browser.v1"}
				return nil
			},
			Handler: func(connection *websocket.Conn) {
				defer connection.Close()
				var ignored []byte
				_ = websocket.Message.Receive(connection, &ignored)
				closed <- struct{}{}
			},
		}.ServeHTTP(w, r)
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
	if err != nil {
		t.Fatal(err)
	}
	output, err := executeAgentBrowserCLI(NewContext(client, "workspace /1", false, true, true), []string{"connect-runtime-browser", "--task-id", "task/1 ?", "--token", token, "--origin", origin})
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil || !reflect.DeepEqual(result, map[string]any{"connected": true}) {
		t.Fatalf("output=%s result=%v error=%v", output, result, err)
	}
	select {
	case request := <-handshake:
		if request.URL.EscapedPath() != "/newmoi/workspaces/workspace%20%2F1/agent-runtime-tasks/task%2F1%20%3F/browser/stream" || request.URL.RawQuery != "" {
			t.Fatalf("URL=%s", request.URL.String())
		}
		protocols := agentBrowserCLIRequestedProtocols(request)
		if request.Header.Get("Origin") != origin || !reflect.DeepEqual(protocols, []string{"moi-browser.v1", token}) {
			t.Fatalf("origin=%q protocols=%v", request.Header.Get("Origin"), protocols)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no browser WebSocket handshake")
	}
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("browser probe did not close its connection")
	}
}

func agentBrowserCLIRequestedProtocols(request *http.Request) []string {
	parts := strings.Split(request.Header.Get("Sec-WebSocket-Protocol"), ",")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	return parts
}

func executeAgentBrowserCLI(ctx *Context, args []string) ([]byte, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteAgentCommand(ctx, args)
	os.Stdout = original
	closeErr := write.Close()
	output, readErr := io.ReadAll(read)
	_ = read.Close()
	if callErr != nil {
		return output, callErr
	}
	if closeErr != nil {
		return output, closeErr
	}
	return output, readErr
}
func TestAgentConnectBrowserReadsOriginalImage(t *testing.T) {
	for _, command := range []string{"connect-browser", "watch-browser"} {
		t.Run(command, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/newmoi/workspaces/ws/agent-runtime-tasks/task/browser/token":
					if r.Method != http.MethodPost || r.Header.Get("X-API-Key") != "product-key" {
						http.Error(w, "missing product identity", http.StatusUnauthorized)
						return
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]string{"token": "moi-browser-token-v1.test"}})
				case "/newmoi/workspaces/ws/agent-runtime-tasks/task/browser/stream":
					wantMode := ""
					if command == "watch-browser" {
						wantMode = "view"
					}
					if r.URL.Query().Get("mode") != wantMode {
						t.Errorf("browser mode=%q, want %q", r.URL.Query().Get("mode"), wantMode)
					}
					upgrader := gorillawebsocket.Upgrader{EnableCompression: true, Subprotocols: []string{"moi-browser.v1"}}
					connection, err := upgrader.Upgrade(w, r, nil)
					if err != nil {
						return
					}
					defer connection.Close()
					if connection.WriteJSON(map[string]any{"type": "frame", "seq": 7}) != nil {
						return
					}
					_ = connection.WriteMessage(gorillawebsocket.BinaryMessage, []byte("image\x00\xff"))
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("product-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			output := captureStdout(t, func() {
				err = ExecuteAgentCommand(NewContext(client, "ws", false, true, true), []string{command, "--task-id", "task"})
			})
			if err != nil {
				t.Fatal(err)
			}
			var result struct {
				Event       map[string]any `json:"event"`
				Image       []byte         `json:"image"`
				Compression bool           `json:"compression"`
			}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatal(err)
			}
			if result.Event["type"] != "frame" || result.Event["seq"] != float64(7) || string(result.Image) != "image\x00\xff" || !result.Compression {
				t.Fatalf("browser result=%s", output)
			}
		})
	}
}
