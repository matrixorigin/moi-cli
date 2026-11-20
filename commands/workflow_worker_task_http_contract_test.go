package commands

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type workflowWorkerTaskContract struct {
	Name            string            `json:"name"`
	Input           json.RawMessage   `json:"input"`
	Request         json.RawMessage   `json:"request"`
	Response        json.RawMessage   `json:"response"`
	ResponseHeaders map[string]string `json:"response_headers"`
	Status          int               `json:"status"`
	ErrorCode       string            `json:"error_code"`
}

func TestWorkflowWorkerTaskSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/workflow_worker_task_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []workflowWorkerTaskContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			observations := make(chan agentChannelCallbackObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				observations <- agentChannelCallbackObservation{
					method: r.Method, path: r.URL.EscapedPath(), body: body,
					workspace: r.Header.Get("X-Workspace-ID"), apiKey: r.Header.Get("X-API-Key"),
					authorization: r.Header.Get("Authorization"), cookie: r.Header.Get("Cookie"), runtimeAuth: r.Header.Get("X-MOI-Runtime-Authorization"),
				}
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			result, err := invokeWorkflowWorkerTask(t, client, scenario)
			assertWorkflowWorkerTaskObservation(t, scenario, <-observations)
			assertWorkflowWorkerTaskError(t, err, scenario)
			if scenario.Status < 400 && scenario.ErrorCode == "" {
				assertWorkflowWorkerTaskResult(t, result, scenario)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteWorkflowCommand(NewContext(client, "workspace-1", false, true, true), []string{"create-worker-task", "--json", string(scenario.Input)})
			})
			assertWorkflowWorkerTaskObservation(t, scenario, <-observations)
			assertWorkflowWorkerTaskError(t, cliErr, scenario)
			if scenario.Status >= 400 || scenario.ErrorCode != "" {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			actual := &sdk.WorkflowWorkerTask{}
			if err := protojson.Unmarshal([]byte(output), actual); err != nil {
				t.Fatalf("output=%s error=%v", output, err)
			}
			assertWorkflowWorkerTaskResult(t, actual, scenario)
		})
	}
}

func invokeWorkflowWorkerTask(t *testing.T, client *sdk.Client, scenario workflowWorkerTaskContract) (*sdk.WorkflowWorkerTask, error) {
	t.Helper()
	var request sdk.WorkflowWorkerTaskCreateRequest
	if err := json.Unmarshal(scenario.Input, &request); err != nil {
		t.Fatal(err)
	}
	return client.Workflows("workspace-1").CreateWorkerTask(t.Context(), &request)
}

func assertWorkflowWorkerTaskObservation(t *testing.T, scenario workflowWorkerTaskContract, got agentChannelCallbackObservation) {
	t.Helper()
	if got.method != http.MethodPost || got.path != "/newmoi/workspaces/workspace-1/tasks" || len(got.query) != 0 {
		t.Fatalf("request=%s %s query=%v", got.method, got.path, got.query)
	}
	if got.workspace != "workspace-1" || got.apiKey != "synthetic-key" || got.authorization != "" || got.cookie != "" || got.runtimeAuth != "" {
		t.Fatalf("headers=%#v", got)
	}
	if !reflect.DeepEqual(agentResourceJSON(got.body), agentResourceJSON(scenario.Request)) {
		t.Fatalf("body=%s want=%s", got.body, scenario.Request)
	}
}

func assertWorkflowWorkerTaskError(t *testing.T, err error, scenario workflowWorkerTaskContract) {
	t.Helper()
	if scenario.Status < 400 && scenario.ErrorCode == "" {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("error=%#v", err)
	}
	if scenario.ErrorCode != "" {
		// Go attaches the received HTTP status; Python reports this local
		// decode failure with status 0. The shared contract is the stable code.
		if failure.Code != scenario.ErrorCode {
			t.Fatalf("error=%#v", failure)
		}
		return
	}
	if failure.Status != scenario.Status {
		t.Fatalf("error=%#v", failure)
	}
	var response struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(scenario.Response, &response); err != nil {
		t.Fatal(err)
	}
	if failure.Code != response.Code || failure.Message != response.Msg || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%#v", failure)
	}
}

func assertWorkflowWorkerTaskResult(t *testing.T, result *sdk.WorkflowWorkerTask, scenario workflowWorkerTaskContract) {
	t.Helper()
	var envelope struct {
		Task json.RawMessage `json:"task"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	expected := &sdk.WorkflowWorkerTask{}
	if err := protojson.Unmarshal(envelope.Task, expected); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(result, expected) {
		t.Fatalf("result=%v want=%v", result, expected)
	}
}
