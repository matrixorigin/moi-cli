package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// a2aCLIMock records the JSON-RPC method sequence and serves the frozen
// message/send admission contract: a working task that completes through
// resubscribe plus tasks/get.
type a2aCLIMock struct {
	mu      sync.Mutex
	methods []string
}

func (m *a2aCLIMock) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.mu.Lock()
		m.methods = append(m.methods, body.Method)
		m.mu.Unlock()
		task := func(state string, cursor int64) map[string]any {
			return map[string]any{
				"kind":      "task",
				"id":        "task-1",
				"contextId": "ctx-1",
				"status":    map[string]any{"state": state, "timestamp": "2026-11-05T08:00:00Z"},
				"metadata":  map[string]any{"eventCursor": cursor},
			}
		}
		result := func(value map[string]any) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": "OK", "msg": "OK",
				"data": map[string]any{"jsonrpc": "2.0", "id": "1", "result": value},
			})
		}
		switch body.Method {
		case "message/send":
			result(task("working", 3))
		case "tasks/resubscribe":
			w.Header().Set("Content-Type", "text/event-stream")
			payload, _ := json.Marshal(map[string]any{
				"jsonrpc": "2.0", "id": "1",
				"result": map[string]any{
					"kind": "status-update", "taskId": "task-1", "contextId": "ctx-1",
					"status":   map[string]any{"state": "completed", "timestamp": "2026-11-05T08:00:01Z"},
					"final":    true,
					"metadata": map[string]any{"eventCursor": 4},
				},
			})
			_, _ = fmt.Fprintf(w, "id: 4\nevent: message\ndata: %s\n\n", payload)
		case "tasks/get":
			result(task("completed", 4))
		default:
			t.Errorf("unexpected method %q", body.Method)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (m *a2aCLIMock) callSequence() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return strings.Join(m.methods, ",")
}

func TestAgentA2AWaitsForTerminalStateByDefault(t *testing.T) {
	t.Parallel()

	mock := &a2aCLIMock{}
	server := httptest.NewServer(mock.handler(t))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteAgentCommand(ctx, []string{
		"a2a",
		"--json", `{"agent_id":"agent-1","jsonrpc":"2.0","id":"1","method":"message/send","params":{"message":{"parts":[{"kind":"text","text":"hi"}]}}}`,
	}); err != nil {
		t.Fatalf("agent a2a: %v", err)
	}
	if got, want := mock.callSequence(), "message/send,tasks/resubscribe,tasks/get"; got != want {
		t.Fatalf("call sequence = %q, want %q", got, want)
	}
}

func TestAgentA2ANoWaitReturnsAdmission(t *testing.T) {
	t.Parallel()

	mock := &a2aCLIMock{}
	server := httptest.NewServer(mock.handler(t))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteAgentCommand(ctx, []string{
		"a2a",
		"--no-wait",
		"--json", `{"agent_id":"agent-1","jsonrpc":"2.0","id":"1","method":"message/send","params":{"message":{"parts":[{"kind":"text","text":"hi"}]}}}`,
	}); err != nil {
		t.Fatalf("agent a2a --no-wait: %v", err)
	}
	if got, want := mock.callSequence(), "message/send"; got != want {
		t.Fatalf("call sequence = %q, want %q", got, want)
	}
}
