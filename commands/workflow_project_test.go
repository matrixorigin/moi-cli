package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestWorkflowCandidateTransportsCurrentProjectFiles(t *testing.T) {
	project := t.TempDir()
	want := map[string][]byte{
		"workflow/workflow.json":      []byte(`{"name":"当前草稿"}`),
		"workflow/runtime-guide.json": []byte(`{"fields":[]}`),
		"workflow/review.json":        []byte(`{"summary":"current review"}`),
		"operators/clean/main.py":     []byte("print('clean')\n"),
		"tests/cases/main.json":       []byte(`{"target":"workflow"}`),
		"tests/fixtures/input.bin":    {0, 255, 10},
	}
	for name, content := range want {
		path := filepath.Join(project, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, content, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(project, ".matrixflow"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".matrixflow", "test-evidence.json"), []byte(`{"status":"passed"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"test-candidate", "submit-candidate"} {
		t.Run(operation, func(t *testing.T) {
			transported := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/newmoi/workflow/v2/workflow-candidates/"+map[string]string{"test-candidate": "test", "submit-candidate": "submit"}[operation] || r.Header.Get("X-Workspace-ID") != "workspace" {
					t.Errorf("unexpected target: %s", r.URL.Path)
				}
				var body struct {
					Files map[string][]byte `json:"project_files"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(want, body.Files) {
					t.Errorf("project content differs: %#v", body.Files)
				}
				transported = true
				if operation == "test-candidate" {
					_, _ = w.Write([]byte(`{"code":"OK","data":{"status":"completed"}}`))
				} else {
					_, _ = w.Write([]byte(`{"code":"OK","data":{"ok":true}}`))
				}
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
			if err != nil {
				t.Fatal(err)
			}
			if err := ExecuteWorkflowCommand(NewContext(client, "workspace", false, true, true), []string{operation, "--project-dir", project}); err != nil {
				t.Fatal(err)
			}
			if !transported {
				t.Fatal("project did not reach existing candidate API")
			}
		})
	}
}

func TestWorkflowCandidateRejectsAmbiguousOrNonRegularProject(t *testing.T) {
	project := t.TempDir()
	for _, directory := range []string{"workflow", "operators", "tests"} {
		if err := os.Mkdir(filepath.Join(project, directory), 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, input := range [][3]string{{`{}`, "", project}, {"", "body.json", project}, {"", "", filepath.Join(project, "missing")}} {
		if _, err := workflowCandidateBody(input[0], input[1], input[2]); err == nil {
			t.Fatalf("invalid project accepted: %v", input)
		}
	}
	if err := os.Symlink("/etc/hosts", filepath.Join(project, "workflow", "linked.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := workflowCandidateBody("", "", project); err == nil {
		t.Fatal("symlink project file accepted")
	}
}

func TestWorkflowCandidateProjectCanOmitUnusedDirectories(t *testing.T) {
	project := t.TempDir()
	if err := os.Mkdir(filepath.Join(project, "workflow"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "workflow", "workflow.json"), []byte(`{"name":"empty-project"}`), 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := workflowCandidateBody("", "", project)
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Files map[string][]byte `json:"project_files"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(body.Files, map[string][]byte{"workflow/workflow.json": []byte(`{"name":"empty-project"}`)}) {
		t.Fatalf("unexpected project: %#v", body.Files)
	}
}
