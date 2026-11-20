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

func TestWorkflowDeploySubmitsAuthoredMOWLAndServiceConfiguration(t *testing.T) {
	project := t.TempDir()
	definition := filepath.Join(project, "workflow.json")
	service := filepath.Join(project, "service.json")
	graph := `{"name":"authored","root":"root","nets":{"root":{"nodes":{}}}}`
	if err := os.WriteFile(definition, []byte(graph), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(service, []byte(`{"service_name":"authored-service","input_schema":"{}","output_schema":"{}","result_mode":"oneshot"}`), 0600); err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/newmoi/workflow/v2/workflow-deployments" || r.Header.Get("X-Workspace-ID") != "workspace" {
			t.Errorf("unexpected deployment target %s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		var expected map[string]any
		if err := json.Unmarshal([]byte(graph), &expected); err != nil {
			t.Error(err)
		}
		if body["source_type"] != "manual_mowl" || body["execution_mode"] != "dynamic_service" || !reflect.DeepEqual(expected, body["workflow"]) {
			t.Errorf("authored definition changed: %+v", body)
		}
		config, _ := body["dynamic_service"].(map[string]any)
		if config["service_name"] != "authored-service" || config["result_mode"] != "oneshot" {
			t.Errorf("service configuration changed: %+v", config)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"workflow":{"id":"saved-workflow","source_type":"manual_mowl"}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace", false, true, true)
	if err := ExecuteWorkflowCommand(ctx, []string{"deploy", "--name", "authored", "--workflow-file", definition,
		"--execution-mode", "dynamic_service", "--dynamic-service-file", service}); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"deploy", "deploy-candidate"} {
		if err := ExecuteSchemaCommand([]string{"workflow", command, "--argument", "dynamic-service-file", "--validate-file", service}); err != nil {
			t.Fatalf("Agent cannot inspect %s service configuration: %v", command, err)
		}
	}
	if requests != 1 {
		t.Fatalf("deployment requests = %d", requests)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"deploy", "--name", "missing", "--workflow-file", filepath.Join(project, "missing.json")}); err == nil {
		t.Fatal("missing source must fail before mutation")
	}
	if requests != 1 {
		t.Fatal("missing source reached the server")
	}
	for _, invalid := range []string{"{", "null", "[]"} {
		if err := os.WriteFile(definition, []byte(invalid), 0600); err != nil {
			t.Fatal(err)
		}
		if err := ExecuteWorkflowCommand(ctx, []string{"deploy", "--name", "authored", "--workflow-file", definition}); err == nil {
			t.Fatalf("invalid MOWL %q accepted", invalid)
		}
		if requests != 1 {
			t.Fatal("invalid MOWL reached publication owner")
		}
	}
	if err := os.WriteFile(definition, []byte(graph), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"deploy", "--name", "authored", "--workflow-file", definition, "--dsl-yaml", "workflow: {}"}); err == nil {
		t.Fatal("mixed DSL and MOWL accepted")
	}
	if err := os.WriteFile(service, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"deploy", "--name", "authored", "--workflow-file", definition, "--dynamic-service-file", service}); err == nil {
		t.Fatal("malformed service JSON accepted")
	}
	if requests != 1 {
		t.Fatal("invalid deployment mutated server state")
	}
	if err := os.WriteFile(service, []byte(`{"result_mode":32}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteSchemaCommand([]string{"workflow", "deploy", "--argument", "dynamic-service-file", "--validate-file", service}); err == nil {
		t.Fatal("schema accepted an incompatible configuration field type")
	}
}

func TestWorkflowAutomationConfigurationThroughCLI(t *testing.T) {
	volumeFile := filepath.Join(t.TempDir(), "volume.json")
	if err := os.WriteFile(volumeFile, []byte(`{"volume_id":29,"enabled":false,"vars_json":"{\"processing\":\"native\"}"}`), 0600); err != nil {
		t.Fatal(err)
	}
	var expectedBody map[string]any
	var expectedMethod, expectedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectedBody == nil {
			t.Error("invalid local input reached the publication owner")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if r.Method != expectedMethod || r.URL.Path != expectedPath || !reflect.DeepEqual(expectedBody, body) {
			t.Errorf("publication changed: %s %s %+v; want %s %s %+v", r.Method, r.URL.Path, body, expectedMethod, expectedPath, expectedBody)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"workflow":{"id":"saved"}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace", false, true, true)
	volume := map[string]any{"volume_id": float64(29), "enabled": false, "vars_json": `{"processing":"native"}`}
	for _, operation := range []struct {
		name, method, path string
		args               []string
		body               map[string]any
	}{
		{"deploy volume automation", http.MethodPost, "/newmoi/workflow/v2/workflow-deployments",
			[]string{"deploy", "--name", "automation", "--dsl-yaml", "workflow: {}", "--execution-mode", "volume_trigger", "--volume-trigger-file", volumeFile},
			map[string]any{"name": "automation", "dsl_yaml": "workflow: {}", "execution_mode": "volume_trigger", "volume_trigger": volume}},
		{"update volume automation", http.MethodPatch, "/newmoi/workflow/v2/workflow-apps/saved",
			[]string{"update", "--workflow-id", "saved", "--execution-mode", "volume_trigger", "--volume-trigger-file", volumeFile},
			map[string]any{"execution_mode": "volume_trigger", "volume_trigger": volume}},
		{"schedule cron", http.MethodPatch, "/newmoi/workflow/v2/workflow-apps/saved",
			[]string{"update", "--workflow-id", "saved", "--execution-mode", "cron", "--cron-expression", "0 */15 * * * *"},
			map[string]any{"execution_mode": "cron", "cron_expression": "0 */15 * * * *"}},
		{"clear description explicitly", http.MethodPatch, "/newmoi/workflow/v2/workflow-apps/saved",
			[]string{"update", "--workflow-id", "saved", "--description", ""},
			map[string]any{"description": ""}},
	} {
		t.Run(operation.name, func(t *testing.T) {
			expectedBody, expectedMethod, expectedPath = operation.body, operation.method, operation.path
			if err := ExecuteWorkflowCommand(ctx, operation.args); err != nil {
				t.Fatal(err)
			}
		})
	}
	expectedBody = nil
	if err := os.WriteFile(volumeFile, []byte(`{"volume_id":"bad"}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{"deploy", "update"} {
		t.Run(operation+" rejects invalid trigger before mutation", func(t *testing.T) {
			args := []string{operation, "--volume-trigger-file", volumeFile}
			if operation == "deploy" {
				args = append(args, "--name", "automation", "--dsl-yaml", "workflow: {}")
			} else {
				args = append(args, "--workflow-id", "saved")
			}
			if err := ExecuteWorkflowCommand(ctx, args); err == nil {
				t.Fatal("invalid trigger accepted")
			}
		})
	}
}
