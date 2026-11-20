package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestSDK19808HTTPCommands(t *testing.T) {
	for _, test := range []struct {
		name, method, path, output string
		args                       []string
		body                       map[string]any
		response                   any
		query                      string
		command                    func(*Context, []string) error
	}{
		{"ownership", "GET", "/iam/roles/7/ownership-impact", "table-1", []string{"ownership-impact", "--role-id", "7", "--query-param", "limit=1"}, nil, map[string]any{"items": []any{map[string]any{"resource_id": "table-1"}}}, "limit=1", ExecuteIAMCommand},
		{"role audit", "GET", "/iam/roles/7/audit-events", "role-next", []string{"role-audit-events", "--role-id", "7"}, nil, map[string]any{"next_cursor": "role-next"}, "", ExecuteIAMCommand},
		{"user audit", "GET", "/iam/users/u1/audit-events", "user-next", []string{"user-audit-events", "--user-id", "u1"}, nil, map[string]any{"next_cursor": "user-next"}, "", ExecuteIAMCommand},
		{"default role", "PUT", "/iam/me/default-role", "changed", []string{"set-default-role", "--json", `{"target_role_id":"7","expected_default_role_version":0,"request_id":"set-1"}`}, map[string]any{"target_role_id": "7", "expected_default_role_version": float64(0), "request_id": "set-1"}, map[string]any{"audit_event_id": "changed"}, "", ExecuteIAMCommand},
		{"artifacts", "GET", "/agent-artifacts", "snapshot-1", []string{"artifacts", "--query-param", "page=2"}, nil, map[string]any{"snapshot_at": "snapshot-1"}, "page=2", ExecuteAgentCommand},
		{"artifact", "GET", "/agent-artifacts/a1/versions/2", "a1", []string{"artifact", "--artifact-id", "a1", "--version", "2"}, nil, map[string]any{"artifact_id": "a1", "version": 2}, "", ExecuteAgentCommand},
		{"delete", "POST", "/agent-artifacts/batch-delete", "missing", []string{"delete-artifacts", "--json", `{"items":[{"artifact_id":"a1","version":2}]}`}, map[string]any{"items": []any{map[string]any{"artifact_id": "a1", "version": float64(2)}}}, map[string]any{"failed": []any{map[string]any{"artifact_id": "missing", "reason": "not_found"}}}, "", ExecuteAgentCommand},
		{"resolve", "POST", "/agent-artifacts/resolve", "f1", []string{"resolve-artifacts", "--json", `{"items":[{"artifact_id":"a1","version":2}]}`}, map[string]any{"items": []any{map[string]any{"artifact_id": "a1", "version": float64(2)}}}, map[string]any{"items": []any{map[string]any{"file_id": "f1"}}}, "", ExecuteAgentCommand},
		{"backfill", "POST", "/agent-artifacts/backfill", "dry_run", []string{"backfill-artifacts", "--json", `{"dry_run":true,"limit":1}`}, map[string]any{"dry_run": true, "limit": float64(1)}, map[string]any{"dry_run": true}, "", ExecuteAgentCommand},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != test.method || r.URL.Path != "/newmoi/workspaces/ws-1"+test.path || r.URL.RawQuery != test.query {
					t.Errorf("request: %s %s", r.Method, r.URL)
				}
				if test.body != nil {
					var body map[string]any
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					actual, _ := json.Marshal(body)
					expected, _ := json.Marshal(test.body)
					if string(actual) != string(expected) {
						t.Errorf("body: %s, want %s", actual, expected)
					}
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": test.response})
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("fixture"))
			if err != nil {
				t.Fatal(err)
			}
			output := captureStdout(t, func() {
				if err := test.command(NewContext(client, "ws-1", false, true, true), test.args); err != nil {
					t.Fatal(err)
				}
			})
			if !strings.Contains(output, test.output) {
				t.Fatalf("output: %s", output)
			}
		})
	}
}

func TestDashboardWebSocketCommand(t *testing.T) {
	for _, action := range []string{"ping", "preview", "create-chart", "update-chart"} {
		t.Run(action, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/newmoi/data-dashboards/ws" || r.URL.Query().Get("workspace_id") != "ws-1" {
					t.Errorf("handshake: %s", r.URL)
				}
				conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
				if err != nil {
					t.Error(err)
					return
				}
				defer conn.Close()
				var frame map[string]any
				if err := conn.ReadJSON(&frame); err != nil {
					t.Error(err)
					return
				}
				kind := map[string]string{"ping": "ping", "preview": "dashboard_preview_execute", "create-chart": "dashboard_chart_create", "update-chart": "dashboard_chart_update"}[action]
				if frame["message_type"] != kind || frame["id"] != "request-1" {
					t.Errorf("sent frame: %#v", frame)
				}
				if action != "ping" {
					data := frame["data"].(map[string]any)
					if data["dashboard_id"] != "d1" || data["sql_text"] != "SELECT 1" {
						t.Errorf("message data: %#v", data)
					}
					if action == "update-chart" && data["chart_id"] != "c1" {
						t.Errorf("chart data: %#v", data)
					}
				}
				_ = conn.WriteJSON(map[string]any{"id": "request-1", "message_type": "pong"})
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("fixture"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"connect-web-socket", "--action", action, "--message-id", "request-1", "--dashboard-id", "d1", "--chart-id", "c1", "--sql-text", "SELECT 1", "--json", `{"name":"report","sql_text":"SELECT 1","chart_type":"table"}`}
			output := captureStdout(t, func() {
				if err := ExecuteDashboardCommand(NewContext(client, "ws-1", false, true, true), args); err != nil {
					t.Fatal(err)
				}
			})
			var result map[string]any
			if err := json.Unmarshal([]byte(output), &result); err != nil || result["id"] != "request-1" || result["message_type"] != "pong" {
				t.Fatalf("response: %s", output)
			}
		})
	}
}

func TestSDK19808CommandErrorsPreserveOwnerResponse(t *testing.T) {
	message := "localized fixture conflict"
	for _, test := range []struct{ name, body, code string }{
		{"business conflict", `{"code":"IAM_DEFAULT_ROLE_VERSION_CONFLICT","msg":"` + message + `"}`, "IAM_DEFAULT_ROLE_VERSION_CONFLICT"},
		{"malformed response", `not-json`, "ErrResponseInvalid"},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Request-ID", "default-role-request")
				w.WriteHeader(http.StatusConflict)
				_, _ = w.Write([]byte(test.body))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("fixture"))
			if err != nil {
				t.Fatal(err)
			}
			err = ExecuteIAMCommand(NewContext(client, "ws", false, true, true), []string{"set-default-role", "--json", `{"target_role_id":"7","expected_default_role_version":0,"request_id":"set"}`})
			var failure *sdk.Error
			if !errors.As(err, &failure) || failure.Code != test.code || failure.Status != http.StatusConflict || failure.RequestID != "default-role-request" {
				t.Fatalf("command error: %v", err)
			}
			if test.name == "business conflict" && failure.Message != message {
				t.Fatalf("owner message lost: %v", failure)
			}
		})
	}
}
