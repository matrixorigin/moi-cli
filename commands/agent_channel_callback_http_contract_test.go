package commands

import (
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

type agentChannelCallbackContract struct {
	Name            string              `json:"name"`
	Operation       string              `json:"operation"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	Provider        string              `json:"provider"`
	InstanceID      string              `json:"instance_id"`
	RuleID          string              `json:"rule_id"`
	Query           map[string][]string `json:"query"`
	Input           json.RawMessage     `json:"input"`
	Response        json.RawMessage     `json:"response"`
	ResponseHeaders map[string]string   `json:"response_headers"`
	Status          int                 `json:"status"`
}

type agentChannelCallbackObservation struct {
	method, path, workspace, apiKey, authorization, cookie, runtimeAuth string
	query                                                               url.Values
	body                                                                []byte
}

func TestAgentChannelCallbackSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_channel_callback_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentChannelCallbackContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			observations := make(chan agentChannelCallbackObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				observations <- agentChannelCallbackObservation{
					method: r.Method, path: r.URL.EscapedPath(), query: r.URL.Query(), body: body,
					workspace: r.Header.Get("X-Workspace-ID"), apiKey: r.Header.Get("X-API-Key"),
					authorization: r.Header.Get("Authorization"), cookie: r.Header.Get("Cookie"),
					runtimeAuth: r.Header.Get("X-MOI-Runtime-Authorization"),
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
			result, args, err := invokeAgentChannelCallbackSDK(t, client, scenario)
			assertAgentChannelCallbackError(t, err, scenario)
			assertAgentChannelCallbackObservation(t, scenario, <-observations)
			if scenario.Status < 400 {
				assertAgentChannelCallbackResult(t, result, scenario)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			assertAgentChannelCallbackError(t, cliErr, scenario)
			assertAgentChannelCallbackObservation(t, scenario, <-observations)
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			assertAgentChannelCallbackJSON(t, agentResourceJSON([]byte(output)), scenario)
		})
	}
}

func invokeAgentChannelCallbackSDK(t *testing.T, client *sdk.Client, scenario agentChannelCallbackContract) (any, []string, error) {
	t.Helper()
	service := client.Agents("workspace-1")
	args := []string{scenario.Operation}
	if scenario.Provider != "" {
		args = append(args, "--provider", scenario.Provider)
	}
	if scenario.InstanceID != "" {
		args = append(args, "--instance-id", scenario.InstanceID)
	}
	if scenario.RuleID != "" {
		args = append(args, "--rule-id", scenario.RuleID)
	}
	if len(scenario.Input) != 0 {
		args = append(args, "--json", string(scenario.Input))
	}
	args = append(args, agentResourceQueryArgs(scenario.Query)...)
	query := make([]sdk.QueryOption, 0)
	for key, values := range scenario.Query {
		for _, value := range values {
			query = append(query, sdk.WithQueryValues(key, value))
		}
	}
	var result any
	var err error
	switch scenario.Operation {
	case "create-channel-instance":
		var request sdk.ChannelInstanceCreateRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.CreateChannelInstance(t.Context(), scenario.Provider, request)
	case "list-channel-instances":
		result, err = service.ListChannelInstances(t.Context(), scenario.Provider, query...)
	case "get-channel-instance":
		result, err = service.GetChannelInstance(t.Context(), scenario.Provider, scenario.InstanceID)
	case "update-channel-instance":
		var request sdk.ChannelInstanceUpdateRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.UpdateChannelInstance(t.Context(), scenario.Provider, scenario.InstanceID, request)
	case "test-channel-configuration":
		var request sdk.ChannelInstanceCreateRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.TestChannelConfiguration(t.Context(), scenario.Provider, request)
	case "disable-channel-instance":
		result, err = service.DisableChannelInstance(t.Context(), scenario.Provider, scenario.InstanceID)
	case "test-channel-instance":
		result, err = service.TestChannelInstance(t.Context(), scenario.Provider, scenario.InstanceID)
	case "create-callback-rule":
		var request sdk.CallbackRuleCreateRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.CreateCallbackRule(t.Context(), request)
	case "list-callback-rules":
		result, err = service.ListCallbackRules(t.Context(), query...)
	case "get-callback-rule":
		result, err = service.GetCallbackRule(t.Context(), scenario.RuleID)
	case "update-callback-rule":
		var request sdk.CallbackRuleUpdateRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.UpdateCallbackRule(t.Context(), scenario.RuleID, request)
	case "delete-callback-rule":
		result, err = service.DeleteCallbackRule(t.Context(), scenario.RuleID)
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	return result, args, err
}

func decodeAgentChannelCallbackInput(t *testing.T, raw json.RawMessage, target any) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatal(err)
	}
}

func assertAgentChannelCallbackObservation(t *testing.T, scenario agentChannelCallbackContract, got agentChannelCallbackObservation) {
	t.Helper()
	wantQuery := url.Values(scenario.Query)
	if scenario.Query == nil {
		wantQuery = url.Values{}
	}
	if got.method != scenario.Method || got.path != scenario.Path || !reflect.DeepEqual(got.query, wantQuery) {
		t.Fatalf("request=%s %s query=%v", got.method, got.path, got.query)
	}
	if got.workspace != "workspace-1" || got.apiKey != "synthetic-key" || got.authorization != "" || got.cookie != "" || got.runtimeAuth != "" {
		t.Fatalf("headers=%#v", got)
	}
	if len(scenario.Input) == 0 {
		if len(got.body) != 0 {
			t.Fatalf("body=%s", got.body)
		}
		return
	}
	if !reflect.DeepEqual(agentResourceJSON(got.body), agentResourceJSON(scenario.Input)) {
		t.Fatalf("body=%s want=%s", got.body, scenario.Input)
	}
}

func assertAgentChannelCallbackResult(t *testing.T, result any, scenario agentChannelCallbackContract) {
	t.Helper()
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	assertAgentChannelCallbackJSON(t, agentResourceJSON(raw), scenario)
}

func assertAgentChannelCallbackJSON(t *testing.T, got any, scenario agentChannelCallbackContract) {
	t.Helper()
	response := agentResourceJSON(scenario.Response).(map[string]any)
	if !reflect.DeepEqual(got, response["data"]) {
		t.Fatalf("result=%#v want=%#v", got, response["data"])
	}
}

func assertAgentChannelCallbackError(t *testing.T, err error, scenario agentChannelCallbackContract) {
	t.Helper()
	if scenario.Status < 400 {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("error=%v", err)
	}
	response := agentResourceJSON(scenario.Response).(map[string]any)
	message, _ := response["message"].(string)
	if failure.Code != fmt.Sprint(response["code"]) || failure.Message != message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%#v", failure)
	}
}
