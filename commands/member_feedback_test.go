package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestMemberFeedbackCLIUsesSDKContracts(t *testing.T) {
	for _, test := range []struct {
		name, method, path string
		body               map[string]string
		run                func(*Context) error
	}{
		{"accept-invitation", "POST", "/newmoi/workspaces/invitations/inv-1/accept", map[string]string{}, func(ctx *Context) error {
			return ExecuteWorkspaceMemberCommand(ctx, []string{"accept-invitation", "--invitation-id", "inv-1"})
		}},
		{"admin-feedback-ticket", "GET", "/newmoi/admin/feedback-tickets/ticket-1", nil, func(ctx *Context) error {
			return ExecuteAgentCommand(ctx, []string{"admin-feedback-ticket", "--ticket-id", "ticket-1"})
		}},
		{"approve-feedback-reward", "POST", "/newmoi/admin/feedback-tickets/ticket-1/reward-approval", map[string]string{"status": "closed", "admin_reply": ""}, func(ctx *Context) error {
			return ExecuteAgentCommand(ctx, []string{"approve-feedback-reward", "--ticket-id", "ticket-1", "--json", `{"status":"closed","admin_reply":""}`})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != test.method || r.URL.Path != test.path || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("Authorization") != "Bearer target-token" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.Path, r.Header)
				}
				if test.body != nil {
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if len(body) != len(test.body) {
						t.Errorf("body=%v want=%v", body, test.body)
					}
					for key, value := range test.body {
						if got, exists := body[key]; !exists || got != value {
							t.Errorf("body[%s]=%q exists=%v", key, got, exists)
						}
					}
				}
				w.Header().Set("X-Request-ID", "public-request-id")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":"ErrForbidden","msg":"原始本地化错误"}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("target-token"))
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

func TestFeedbackGitHubIssueCLIUsesSDKContracts(t *testing.T) {
	for _, test := range []struct {
		name, path string
		body       map[string]any
		args       []string
	}{
		{"draft", "/newmoi/admin/feedback-tickets/ticket-1/github-issue-draft", map[string]any{"kind": "bug"},
			[]string{"build-feedback-github-issue-draft", "--ticket-id", "ticket-1", "--kind", "bug"}},
		{"create", "/newmoi/admin/feedback-tickets/ticket-1/github-issue", map[string]any{"kind": "bug", "title": "draft", "body": "details", "retry": true},
			[]string{"create-feedback-github-issue", "--ticket-id", "ticket-1", "--kind", "bug", "--title", "draft", "--body", "details", "--retry"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != test.path || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("Authorization") != "Bearer target-token" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.Path, r.Header)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(body, test.body) {
					t.Errorf("body=%v want=%v", body, test.body)
				}
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":"ErrForbidden","msg":"原始本地化错误"}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("target-token"))
			if err != nil {
				t.Fatal(err)
			}
			err = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), test.args)
			var failure *sdk.Error
			if !errors.As(err, &failure) || failure.Status != http.StatusForbidden || failure.Code != "ErrForbidden" || failure.Message != "原始本地化错误" {
				t.Fatalf("public error=%#v", err)
			}
		})
	}
}
