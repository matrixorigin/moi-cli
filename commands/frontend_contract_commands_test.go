package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestFrontendContractCLIForwardsRoutesAndPublicErrors(t *testing.T) {
	for _, test := range []struct {
		name, path, query, workspaceID string
		run                            func(*Context) error
	}{
		{"registration-invite", "/newmoi/user/me/registration-invite", "", "", func(ctx *Context) error { return ExecuteAccountCommand(ctx, []string{"registration-invite"}) }},
		{"reward-center-items", "/newmoi/user/me/reward-center/items", "", "", func(ctx *Context) error { return ExecuteAccountCommand(ctx, []string{"reward-center-items"}) }},
		{"reward-center-events", "/newmoi/user/me/reward-center/events", "event_type=feedback", "", func(ctx *Context) error {
			return ExecuteAccountCommand(ctx, []string{"reward-center-events", "--event-type", "feedback"})
		}},
		{"reward-center-invitations", "/newmoi/user/me/reward-center/invitations", "page=2&page_size=10", "", func(ctx *Context) error {
			return ExecuteAccountCommand(ctx, []string{"reward-center-invitations", "--page", "2", "--page-size", "10"})
		}},
		{"agent-list", "/newmoi/agents", "query=source", "workspace-1", func(ctx *Context) error { return ExecuteAgentCommand(ctx, []string{"list", "--query", "source"}) }},
		{"executor-health", "/newmoi/agent-executors/health", "", "workspace-1", func(ctx *Context) error { return ExecuteAgentCommand(ctx, []string{"executor-health"}) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != test.path || r.URL.RawQuery != test.query || r.Header.Get("X-Workspace-ID") != test.workspaceID || r.Header.Get("Authorization") != "Bearer user-token" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL, r.Header)
				}
				w.Header().Set("X-Request-ID", "public-request-id")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":"ErrForbidden","msg":"原始本地化错误"}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("user-token"))
			if err != nil {
				t.Fatal(err)
			}
			err = test.run(NewContext(client, "workspace-1", false, true, true))
			var failure *sdk.Error
			if !errors.As(err, &failure) || failure.Status != http.StatusForbidden || failure.Code != "ErrForbidden" || failure.Message != "原始本地化错误" || failure.RequestID != "public-request-id" {
				t.Fatalf("public error=%#v", err)
			}
		})
	}
}

func TestFrontendContractCLIStreamsExactImageAndCSV(t *testing.T) {
	for _, test := range []struct {
		name, method, path, contentType, content string
		run                                      func(*Context, string) error
	}{
		{"query-visual", "GET", "/newmoi/workspaces/workspace-1/agents/agent-1/query-visuals/image-1/preview", "image/png", "\x89PNG\x00\xff\r\n", func(ctx *Context, output string) error {
			return ExecuteAgentCommand(ctx, []string{"preview-query-visual", "--agent-id", "agent-1", "--file-id", "image-1", "--output", output})
		}},
		{"query-csv", "POST", "/newmoi/query/result/download", "text/csv", "name,count\nsource,2\n", func(ctx *Context, output string) error {
			return ExecuteSQLCommand(ctx, []string{"download-query-result", "--statement-id", "statement-1", "--output", output})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != test.method || r.URL.Path != test.path || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "user-key" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.Path, r.Header)
				}
				if r.Method == http.MethodPost {
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body) != 1 || body["statement_id"] != "statement-1" {
						t.Errorf("body=%v error=%v", body, err)
					}
				}
				w.Header().Set("Content-Type", test.contentType)
				_, _ = w.Write([]byte(test.content))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
			if err != nil {
				t.Fatal(err)
			}
			output := filepath.Join(t.TempDir(), test.name)
			if err := test.run(NewContext(client, "workspace-1", false, true, true), output); err != nil {
				t.Fatal(err)
			}
			content, err := os.ReadFile(output)
			if err != nil || string(content) != test.content {
				t.Fatalf("content=%q error=%v", content, err)
			}
		})
	}
}
