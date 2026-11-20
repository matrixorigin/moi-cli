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

func TestWorkflowEnableAndCustomOperatorSourceCommands(t *testing.T) {
	t.Parallel()

	requests := make(chan *http.Request, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Clone(r.Context())
		data := map[string]any{"workflow": map[string]any{"id": "wf-1"}}
		if r.URL.Path == "/newmoi/workflow/v2/custom-operators/source" {
			data = map[string]any{"source": "package workflow", "origin": "draft", "node_id": "node-1", "version": "v1"}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": data})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)

	if err := ExecuteWorkflowCommand(ctx, []string{"enable", "--workflow-id", "wf-1"}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"get-custom-operator-source", "--node-id", "node-1", "--version", "v1", "--task-id", "task-1"}); err != nil {
		t.Fatalf("get custom operator source: %v", err)
	}

	enableReq := <-requests
	if enableReq.Method != http.MethodPost || enableReq.URL.Path != "/newmoi/workflow/v2/workflow-apps/wf-1/enable" {
		t.Fatalf("enable request = %s %s", enableReq.Method, enableReq.URL.Path)
	}
	sourceReq := <-requests
	if sourceReq.Method != http.MethodGet || sourceReq.URL.Path != "/newmoi/workflow/v2/custom-operators/source" {
		t.Fatalf("source request = %s %s", sourceReq.Method, sourceReq.URL.Path)
	}
	if sourceReq.URL.Query().Get("node_id") != "node-1" || sourceReq.URL.Query().Get("version") != "v1" || sourceReq.URL.Query().Get("task_id") != "task-1" {
		t.Fatalf("source query = %s", sourceReq.URL.RawQuery)
	}
}

func TestWorkflowCustomOperatorCodeFileCommands(t *testing.T) {
	code := "def handle(workspace_id, sdk, input):\n    return input\n"
	file := filepath.Join(t.TempDir(), "operator.py")
	if err := os.WriteFile(file, []byte(code), 0600); err != nil {
		t.Fatal(err)
	}
	requests := make(chan map[string]any, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		body["method"] = r.Method
		requests <- body
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": map[string]any{"id": 1}})
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	spec, err := json.Marshal(sdk.CustomOperatorSpec{Name: "Identity", Identifier: "identity", Description: "Pass through input", InputSchema: map[string]any{"type": "object"}, OutputSchema: map[string]any{"type": "object"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"create-custom-operator", "--json", string(spec), "--code-file", file, "--language", "python", "--handler", "handle"}); err != nil {
		t.Fatal(err)
	}
	created := <-requests
	if created["code"] != code || created["language"] != "python" || created["handler"] != "handle" || created["method"] != "POST" {
		t.Fatalf("create source request: %#v", created)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"update-custom-operator", "--operator-id", "operator-1", "--code-file", file, "--name", "Updated", "--description", "", "--language", "python"}); err != nil {
		t.Fatal(err)
	}
	updated := <-requests
	if updated["code"] != code || updated["language"] != "python" || updated["name"] != "Updated" || updated["description"] != "" || updated["method"] != "PUT" {
		t.Fatalf("update source request: %#v", updated)
	}
	if _, present := updated["handler"]; present {
		t.Fatal("omitted patch fields must remain omitted")
	}
	err = ExecuteWorkflowCommand(ctx, []string{"update-custom-operator", "--operator-id", "operator-1", "--code-file", file + ".missing"})
	if err == nil {
		t.Fatal("missing source must be rejected")
	}
	invalid := filepath.Join(t.TempDir(), "invalid.py")
	if err := os.WriteFile(invalid, []byte{0xff, 0xfe}, 0600); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"update-custom-operator", "--operator-id", "operator-1", "--code-file", invalid}); err == nil {
		t.Fatal("invalid source encoding must be rejected")
	}
	select {
	case request := <-requests:
		t.Fatalf("source read failure mutated the operator: %#v", request)
	default:
	}
}

func TestWorkflowCLIExposesExistingListAndRunArguments(t *testing.T) {
	valuesFile := filepath.Join(t.TempDir(), "values.json")
	if err := os.WriteFile(valuesFile, []byte(`{"customer":"acme","limit":3}`), 0600); err != nil {
		t.Fatal(err)
	}
	type capturedRequest struct {
		method string
		path   string
		query  map[string]string
		body   map[string]any
	}
	requests := make(chan capturedRequest, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured := capturedRequest{method: r.Method, path: r.URL.Path, query: map[string]string{}}
		for key, values := range r.URL.Query() {
			if len(values) == 1 {
				captured.query[key] = values[0]
			}
		}
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&captured.body); err != nil {
				t.Error(err)
			}
		}
		requests <- captured
		switch r.URL.Path {
		case "/newmoi/workflow/v2/workflow-apps":
			_, _ = w.Write([]byte(`{"code":"OK","data":{"total":0,"workflows":[]}}`))
		case "/newmoi/workflow/v2/workflow-apps/wf-1/executions":
			if r.Method == http.MethodPost {
				_, _ = w.Write([]byte(`{"code":"OK","data":{"workflow_run":{"execution_id":"exec-1"}}}`))
			} else {
				_, _ = w.Write([]byte(`{"code":"OK","data":{"total":0,"executions":[]}}`))
			}
		default:
			_, _ = w.Write([]byte(`{"code":"OK","data":{"total":0,"executions":[]}}`))
		}
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace", false, true, true)
	if err := ExecuteWorkflowCommand(ctx, []string{"list", "--limit", "7", "--offset", "2", "--source-type", "manual_dsl", "--status", "ready", "--execution-mode", "cron", "--name-search", "daily", "--workflow-definition-id", "definition-1", "--include-dynamic-service"}); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"run", "--workflow-id", "wf-1", "--values-file", valuesFile, "--replace-saved-defaults", "--trigger-now=true"}); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"list-runs", "--workflow-id", "wf-1", "--limit", "5", "--offset", "1", "--status", "running", "--execution-mode", "one_shot", "--workflow-name", "daily"}); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"list-all-runs", "--limit", "9", "--offset", "3", "--status", "failed", "--execution-mode", "cron", "--workflow-name", "batch"}); err != nil {
		t.Fatal(err)
	}

	list := <-requests
	if list.method != http.MethodGet || list.path != "/newmoi/workflow/v2/workflow-apps" {
		t.Fatalf("list request = %#v", list)
	}
	for key, want := range map[string]string{
		"limit": "7", "offset": "2", "source_type": "manual_dsl", "status": "ready", "execution_mode": "cron",
		"name_search": "daily", "workflow_definition_id": "definition-1", "include_dynamic_service": "true",
	} {
		if list.query[key] != want {
			t.Fatalf("list query %s = %q, want %q", key, list.query[key], want)
		}
	}
	run := <-requests
	if run.method != http.MethodPost || run.path != "/newmoi/workflow/v2/workflow-apps/wf-1/executions" ||
		!reflect.DeepEqual(run.body["values"], map[string]any{"customer": "acme", "limit": float64(3)}) ||
		run.body["replace_saved_defaults"] != true || run.body["trigger_now"] != true {
		t.Fatalf("run request = %#v", run)
	}
	for _, expected := range []struct {
		request capturedRequest
		path    string
		query   map[string]string
	}{
		{request: <-requests, path: "/newmoi/workflow/v2/workflow-apps/wf-1/executions", query: map[string]string{"limit": "5", "offset": "1", "status": "running", "execution_mode": "one_shot", "workflow_name": "daily"}},
		{request: <-requests, path: "/newmoi/workflow/v2/workflow-apps/executions", query: map[string]string{"limit": "9", "offset": "3", "status": "failed", "execution_mode": "cron", "workflow_name": "batch"}},
	} {
		if expected.request.method != http.MethodGet || expected.request.path != expected.path {
			t.Fatalf("execution list request = %#v", expected.request)
		}
		for key, want := range expected.query {
			if expected.request.query[key] != want {
				t.Fatalf("execution list query %s = %q, want %q", key, expected.request.query[key], want)
			}
		}
	}
}

func TestWorkflowCLIRetryExistingContract(t *testing.T) {
	valuesFile := filepath.Join(t.TempDir(), "values.json")
	if err := os.WriteFile(valuesFile, []byte(`{"customer":"acme","limit":3}`), 0600); err != nil {
		t.Fatal(err)
	}
	requests := make(chan map[string]any, 4)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/workflow/v2/workflow-apps/wf-1/executions/exec-1/retry" {
			t.Errorf("retry request = %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		requests <- body
		_, _ = w.Write([]byte(`{"code":"OK","data":{"workflow_run":{"execution_id":"exec-2"}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace", false, true, true)
	baseArgs := []string{"retry-run", "--workflow-id", "wf-1", "--execution-id", "exec-1"}
	for _, tc := range []struct {
		name string
		args []string
		body map[string]any
	}{
		{"values and replace defaults", []string{"--values-file", valuesFile, "--replace-saved-defaults", "--trigger-now=false"}, map[string]any{"values": map[string]any{"customer": "acme", "limit": float64(3)}, "replace_saved_defaults": true, "trigger_now": false}},
		{"immediate dispatch", []string{"--trigger-now=true"}, map[string]any{"trigger_now": true}},
		{"omitted overrides", nil, map[string]any{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := ExecuteWorkflowCommand(ctx, append(append([]string{}, baseArgs...), tc.args...)); err != nil {
				t.Fatal(err)
			}
			if body := <-requests; !reflect.DeepEqual(body, tc.body) {
				t.Fatalf("retry body = %#v, want %#v", body, tc.body)
			}
		})
	}
	invalidFile := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(invalidFile, []byte(`[]`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{valuesFile + ".missing", invalidFile} {
		if err := ExecuteWorkflowCommand(ctx, append(append([]string{}, baseArgs...), "--values-file", file)); err == nil {
			t.Fatalf("invalid values file %q must be rejected", file)
		}
	}
	select {
	case body := <-requests:
		t.Fatalf("invalid values must not submit a retry: %#v", body)
	default:
	}
}

func TestWorkflowGetDispatchJobCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/newmoi/workflow/v2/workflow-apps/wf-1/dispatch-jobs/dispatch-1" {
			t.Errorf("unexpected dispatch observation: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"dispatch_job": map[string]any{"dispatch_job_id": "dispatch-1", "workflow_id": "wf-1", "status": "completed", "total_deliveries": 0, "pending_deliveries": 0, "failed_deliveries": 0}}})
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("fixture-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteWorkflowCommand(ctx, []string{"get-dispatch-job", "--workflow-id", "wf-1", "--dispatch-job-id", "dispatch-1"}); err != nil {
		t.Fatal(err)
	}
}
