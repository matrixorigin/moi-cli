package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestWorkflowCandidateDeployConfigurationFiles(t *testing.T) {
	requests := make(chan map[string]any, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/workflow/v2/workflow-deployments" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		requests <- body
		_, _ = w.Write([]byte(`{"code":"OK","data":{"workflow":{"id":"wf-1"}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := NewContext(client, "ws-1", false, true, true)
	files := map[string]string{
		"default-values-file":  `{"minimum":0,"enabled":false,"text":"configured"}`,
		"volume-trigger-file":  `{"volume_id":7,"enabled":false,"vars_json":"{\"destination\":\"orders\"}"}`,
		"dynamic-service-file": `{"service_name":"accepted-example","input_schema":"{}","output_schema":"{}","result_mode":"one_shot"}`,
	}
	args := []string{"deploy-candidate", "--name", "Accepted workflow", "--acceptance-file", writeAcceptanceFile(t, validWorkflowCandidateAcceptanceJSON)}
	for flag, content := range files {
		path := filepath.Join(t.TempDir(), flag+".json")
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--"+flag, path)
	}
	if err := ExecuteWorkflowCommand(ctx, args); err != nil {
		t.Fatal(err)
	}
	body := <-requests
	if body["source_type"] != "nl2dsl" || body["candidate_ref"] == nil {
		t.Fatalf("candidate identity was lost: %#v", body)
	}
	for flag, key := range map[string]string{"default-values-file": "default_values", "volume-trigger-file": "volume_trigger", "dynamic-service-file": "dynamic_service"} {
		var expected map[string]any
		if err := json.Unmarshal([]byte(files[flag]), &expected); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected, body[key]) {
			t.Errorf("%s = %#v, want %#v", key, body[key], expected)
		}
	}
	if _, present := body["workflow"]; present {
		t.Fatal("candidate configuration must not inject a raw graph")
	}
}

const validWorkflowCandidateAcceptanceJSON = `{
  "candidate_ref": {
    "task_id": "task-123",
    "artifact_id": "workflow_candidate_0123456789abcdef0123456789abcdef"
  },
  "accepted_at": "2026-08-07T12:00:00Z",
  "candidate": {
    "ok": true,
    "candidate_id": "workflow_candidate_0123456789abcdef0123456789abcdef",
    "workflow": {"name": "accepted"},
    "envelope_digest": "envelope-sha256"
  }
}`

func TestAcceptedWorkflowCandidateFromFile(t *testing.T) {
	path := writeAcceptanceFile(t, validWorkflowCandidateAcceptanceJSON)

	accepted, err := acceptedWorkflowCandidateFromFile(path)
	if err != nil {
		t.Fatalf("acceptedWorkflowCandidateFromFile() error = %v", err)
	}
	ref := accepted.Ref()
	if ref.TaskID != "task-123" || ref.ArtifactID != "workflow_candidate_0123456789abcdef0123456789abcdef" {
		t.Fatalf("accepted ref = %+v", ref)
	}
}

func TestAcceptedWorkflowCandidateFromFileRejectsUntrustedInput(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr string
	}{
		{
			name:    "unknown field",
			content: strings.Replace(validWorkflowCandidateAcceptanceJSON, `"accepted_at":`, `"forged": true, "accepted_at":`, 1),
			wantErr: "unknown field",
		},
		{
			name:    "trailing JSON value",
			content: validWorkflowCandidateAcceptanceJSON + ` {}`,
			wantErr: "multiple JSON values",
		},
		{
			name:    "missing accepted timestamp",
			content: strings.Replace(validWorkflowCandidateAcceptanceJSON, `"2026-08-07T12:00:00Z"`, `""`, 1),
			wantErr: "accepted_at is required",
		},
		{
			name:    "mismatched candidate identity",
			content: strings.Replace(validWorkflowCandidateAcceptanceJSON, `"candidate_id": "workflow_candidate_0123456789abcdef0123456789abcdef"`, `"candidate_id": "workflow_candidate_other"`, 1),
			wantErr: "candidate_id does not match artifact_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeAcceptanceFile(t, tt.content)
			_, err := acceptedWorkflowCandidateFromFile(path)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("acceptedWorkflowCandidateFromFile() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestAcceptedWorkflowCandidateFromFileRequiresPath(t *testing.T) {
	_, err := acceptedWorkflowCandidateFromFile("")
	if err == nil || !strings.Contains(err.Error(), "acceptance-file is required") {
		t.Fatalf("acceptedWorkflowCandidateFromFile() error = %v", err)
	}
}

func writeAcceptanceFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "acceptance.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write acceptance file: %v", err)
	}
	return path
}
