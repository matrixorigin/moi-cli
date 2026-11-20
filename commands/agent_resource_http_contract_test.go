package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type agentResourceContract struct {
	Name, Method, Path, Resource string
	TailPath                     string              `json:"tail_path"`
	Query                        map[string][]string `json:"query"`
	RawResponse                  bool                `json:"raw_response"`
	Input                        struct {
		Body json.RawMessage `json:"body"`
	} `json:"input"`
	Request         json.RawMessage   `json:"request"`
	Response        json.RawMessage   `json:"response"`
	ResponseHeaders map[string]string `json:"response_headers"`
	Status          int               `json:"status"`
}

func agentResourceJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		panic(err)
	}
	return value
}

func agentResourceQueryArgs(query map[string][]string) []string {
	var args []string
	for _, key := range []string{"limit", "offset", "status", "source_type", "visibility", "index_status", "owner_user_id", "query"} {
		for _, value := range query[key] {
			args = append(args, "--query-param", key+"="+value)
		}
	}
	for key, values := range query {
		switch key {
		case "limit", "offset", "status", "source_type", "visibility", "index_status", "owner_user_id", "query":
			continue
		}
		for _, value := range values {
			args = append(args, "--query-param", key+"="+value)
		}
	}
	return args
}

func assertAgentResourceError(t *testing.T, err error, scenario agentResourceContract) {
	t.Helper()
	if scenario.Status < 400 {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("error=%T %v", err, err)
	}
	var envelope struct {
		Code    any    `json:"code"`
		Msg     string `json:"msg"`
		Message string `json:"message"`
	}
	decoder := json.NewDecoder(bytes.NewReader(scenario.Response))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	message := envelope.Msg
	if message == "" {
		message = envelope.Message
	}
	if failure.Code != fmt.Sprint(envelope.Code) || failure.Message != message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v", failure)
	}
}

func runAgentResourceSharedHTTPContracts(t *testing.T, fixture string) {
	t.Helper()
	data, err := os.ReadFile("../testdata/product-contracts/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentResourceContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			bodies := make(chan any, 2)
			expectedQuery := url.Values{}
			for key, values := range scenario.Query {
				for _, value := range values {
					expectedQuery.Add(key, value)
				}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				var body any
				if len(raw) > 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil {
						t.Errorf("decode body: %v", err)
					}
				}
				bodies <- body
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
			args := []string{"resource-any", "--method", scenario.Method, "--resource", scenario.Resource, "--tail-path", scenario.TailPath}
			args = append(args, agentResourceQueryArgs(scenario.Query)...)
			if len(scenario.Input.Body) > 0 {
				args = append(args, "--json", string(scenario.Input.Body))
			}
			result, goErr := client.Agents("workspace-1").ResourceAny(t.Context(), scenario.Method, scenario.Resource, scenario.TailPath, agentResourceJSON(scenario.Input.Body), knowledgeLifecycleQuery(scenario.Query)...)
			if got := <-bodies; !reflect.DeepEqual(got, agentResourceJSON(scenario.Request)) {
				t.Fatalf("Go body=%#v want=%#v", got, agentResourceJSON(scenario.Request))
			}
			assertAgentResourceError(t, goErr, scenario)
			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if got := <-bodies; !reflect.DeepEqual(got, agentResourceJSON(scenario.Request)) {
				t.Fatalf("CLI body=%#v want=%#v", got, agentResourceJSON(scenario.Request))
			}
			assertAgentResourceError(t, cliErr, scenario)
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			expectedRaw := envelope.Data
			if scenario.RawResponse {
				expectedRaw = scenario.Response
			}
			expected := agentResourceJSON(expectedRaw)
			if result == nil || !reflect.DeepEqual(result.Value, expected) {
				t.Fatalf("Go result=%#v want=%#v", result, expected)
			}
			if got := agentResourceJSON([]byte(output)); !reflect.DeepEqual(got, expected) {
				t.Fatalf("CLI result=%#v want=%#v", got, expected)
			}
		})
	}
}

func TestAgentKnowledgeBaseResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_knowledge_bases_http_contract.json")
}

func TestAgentTaskTemplateWorkflowBindingResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_task_template_workflow_binding_http_contract.json")
}

func TestAgentDefinitionResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_definition_http_contract.json")
}

func TestAgentVersionResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_version_http_contract.json")
}

func TestAgentSkillResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_skill_http_contract.json")
}

func TestAgentToolResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_tool_http_contract.json")
}

func TestAgentToolConnectionResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_tool_connection_http_contract.json")
}

func TestAgentAutomationResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_automation_http_contract.json")
}

func TestAgentArtifactResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_artifact_http_contract.json")
}

func TestAgentRuntimeReadResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_runtime_read_http_contract.json")
}

func TestAgentRuntimeActionResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_runtime_action_http_contract.json")
}

func TestAgentBuilderResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_builder_http_contract.json")
}

func TestAgentConversationResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_conversation_http_contract.json")
}

func TestAgentWorkspaceExtensionsResourceSharedHTTPContracts(t *testing.T) {
	runAgentResourceSharedHTTPContracts(t, "agent_workspace_extensions_http_contract.json")
}
