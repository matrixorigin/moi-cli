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
	"strconv"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type agentIngressContract struct {
	Name            string              `json:"name"`
	Operation       string              `json:"operation"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	ConnectionID    string              `json:"connection_id"`
	Target          json.RawMessage     `json:"target"`
	Query           map[string][]string `json:"query"`
	Input           json.RawMessage     `json:"input"`
	Response        json.RawMessage     `json:"response"`
	ResponseHeaders map[string]string   `json:"response_headers"`
	Status          int                 `json:"status"`
}

func TestAgentIngressSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_ingress_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentIngressContract
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
			result, args, err := invokeAgentIngressSDK(t, client, scenario)
			assertAgentIngressError(t, err, scenario)
			assertAgentIngressObservation(t, scenario, <-observations)
			if scenario.Status < 400 {
				assertAgentIngressResult(t, result, scenario)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			assertAgentIngressError(t, cliErr, scenario)
			assertAgentIngressObservation(t, scenario, <-observations)
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			assertAgentIngressCLI(t, []byte(output), scenario)
		})
	}
}

func invokeAgentIngressSDK(t *testing.T, client *sdk.Client, scenario agentIngressContract) (any, []string, error) {
	t.Helper()
	service := client.Agents("workspace-1")
	args := []string{scenario.Operation}
	if scenario.ConnectionID != "" {
		args = append(args, "--connection-id", scenario.ConnectionID)
	}
	var target *sdk.AgentTargetRef
	if len(scenario.Target) != 0 {
		target = &sdk.AgentTargetRef{}
		decodeAgentChannelCallbackInput(t, scenario.Target, target)
		args = append(args, "--target-kind", target.GetKind(), "--target-workspace-id", target.GetWorkspaceId(), "--target-id", target.GetId())
	}
	if len(scenario.Input) != 0 && scenario.Operation != "resolve-we-com-qr-binding" {
		args = append(args, "--json", string(scenario.Input))
	}
	var result any
	var err error
	switch scenario.Operation {
	case "list-ingress-capabilities":
		result, err = service.ListIngressCapabilities(t.Context())
	case "list-ingress-connections":
		result, err = service.ListIngressConnections(t.Context(), target)
	case "create-ingress-connection":
		var request sdk.IngressConnectionCreateRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.CreateIngressConnection(t.Context(), request)
	case "resolve-we-com-qr-binding":
		var input struct {
			SessionToken string `json:"session_token"`
		}
		decodeAgentChannelCallbackInput(t, scenario.Input, &input)
		args = append(args, "--qr-session-token", input.SessionToken)
		result, err = service.ResolveWeComQRBinding(t.Context(), target, input.SessionToken)
	case "delete-ingress-connection":
		association := agentIngressVersion(t, scenario.Query, "expected_association_version")
		configuration := agentIngressVersion(t, scenario.Query, "expected_config_version")
		args = append(args, "--expected-version", strconv.FormatInt(association, 10), "--expected-config-version", strconv.FormatInt(configuration, 10))
		result, err = service.DeleteIngressConnection(t.Context(), scenario.ConnectionID, association, configuration)
	case "get-ingress-connection":
		result, err = service.GetIngressConnection(t.Context(), scenario.ConnectionID)
	case "update-ingress-connection":
		var request sdk.IngressConnectionUpdateRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.UpdateIngressConnection(t.Context(), scenario.ConnectionID, request)
	case "verify-ingress-connection":
		var request sdk.IngressConnectionVerifyRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.VerifyIngressConnection(t.Context(), scenario.ConnectionID, request)
	case "generate-we-com-callback-secrets":
		result, err = service.GenerateWeComCallbackSecrets(t.Context())
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	return result, args, err
}

func agentIngressVersion(t *testing.T, query map[string][]string, key string) int64 {
	t.Helper()
	value, err := strconv.ParseInt(query[key][0], 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func assertAgentIngressObservation(t *testing.T, scenario agentIngressContract, got agentChannelCallbackObservation) {
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

func assertAgentIngressJSON(t *testing.T, got any, scenario agentIngressContract) {
	t.Helper()
	response := agentResourceJSON(scenario.Response).(map[string]any)
	if !reflect.DeepEqual(got, response["data"]) {
		t.Fatalf("result=%#v want=%#v", got, response["data"])
	}
}

func assertAgentIngressResult(t *testing.T, result any, scenario agentIngressContract) {
	t.Helper()
	if scenario.Operation == "verify-ingress-connection" || scenario.Operation == "generate-we-com-callback-secrets" {
		raw, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		assertAgentIngressJSON(t, agentResourceJSON(raw), scenario)
		return
	}
	message, ok := result.(proto.Message)
	if !ok || !proto.Equal(message, agentIngressExpectedMessage(t, scenario)) {
		t.Fatalf("result=%v", result)
	}
}

func assertAgentIngressCLI(t *testing.T, output []byte, scenario agentIngressContract) {
	t.Helper()
	if scenario.Operation == "verify-ingress-connection" || scenario.Operation == "generate-we-com-callback-secrets" {
		assertAgentIngressJSON(t, agentResourceJSON(output), scenario)
		return
	}
	actual := proto.Clone(agentIngressExpectedMessage(t, scenario))
	proto.Reset(actual)
	if err := protojson.Unmarshal(output, actual); err != nil {
		t.Fatalf("output=%s error=%v", output, err)
	}
	if !proto.Equal(actual, agentIngressExpectedMessage(t, scenario)) {
		t.Fatalf("output=%s", output)
	}
}

func agentIngressExpectedMessage(t *testing.T, scenario agentIngressContract) proto.Message {
	t.Helper()
	var message proto.Message
	switch scenario.Operation {
	case "list-ingress-capabilities":
		message = &sdk.IngressCapabilityList{}
	case "list-ingress-connections":
		message = &sdk.IngressConnectionList{}
	case "resolve-we-com-qr-binding":
		message = &sdk.WeComQRBindingResult{}
	case "create-ingress-connection", "delete-ingress-connection", "get-ingress-connection", "update-ingress-connection":
		message = &sdk.IngressConnection{}
	default:
		t.Fatalf("operation %s has no protobuf response", scenario.Operation)
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if err := protojson.Unmarshal(envelope.Data, message); err != nil {
		t.Fatal(err)
	}
	return message
}

func assertAgentIngressError(t *testing.T, err error, scenario agentIngressContract) {
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
