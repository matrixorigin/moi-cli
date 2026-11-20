package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestConfirmSkillCandidateCLIContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.EscapedPath() != "/newmoi/workspaces/workspace%201/agent-runtime-tasks/task%201/skills/confirm" {
			t.Errorf("request=%s %s", r.Method, r.URL.EscapedPath())
		}
		var request map[string]json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if string(request["bind_current"]) != "false" || string(request["call_id"]) != `"call-1"` || string(request["source_digest"]) != `"digest-1"` {
			t.Errorf("body=%s", request)
		}
		w.WriteHeader(http.StatusConflict)
		_, err := w.Write([]byte(`{"code":"ErrConflict","msg":"Binding was rejected","details":{"reason":"SKILL_CREATOR_BINDING_FAILED","saved_skill_id":"skill-1"}}`))
		if err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	err = ExecuteAgentCommand(NewContext(client, "workspace 1", false, true, true), []string{"confirm-skill-candidate", "--task-id", "task 1", "--json", `{"call_id":"call-1","source_digest":"digest-1","bind_current":false}`})
	var failure *sdk.Error
	if !errors.As(err, &failure) || failure.Code != "ErrConflict" || failure.Message != "Binding was rejected" || failure.Details["saved_skill_id"] != "skill-1" || failure.Details["reason"] != "SKILL_CREATOR_BINDING_FAILED" {
		t.Fatalf("error=%#v", err)
	}
}

func TestGetSkillCandidateConfirmationCLIContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.EscapedPath() != "/newmoi/workspaces/workspace%201/agent-runtime-tasks/task%201/skills/confirmation" || r.URL.Query().Get("call_id") != "call /1" || r.URL.Query().Get("source_digest") != "sha256:a+b" {
			t.Errorf("request=%s %s", r.Method, r.URL)
		}
		if _, err := w.Write([]byte(`{"code":"OK","data":{"skill_id":"skill-1","workspace_id":"workspace 1","agent_id":"agent-1","saved":false,"bound":false}}`)); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	err = ExecuteAgentCommand(NewContext(client, "workspace 1", false, true, true), []string{"get-skill-candidate-confirmation", "--task-id", "task 1", "--call-id", "call /1", "--source-digest", "sha256:a+b"})
	if err != nil {
		t.Fatal(err)
	}
}
