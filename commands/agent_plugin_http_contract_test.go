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
	"slices"
	"strconv"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type agentPluginHTTPContract struct {
	Name            string              `json:"name"`
	Operation       string              `json:"operation"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	ConnectionID    string              `json:"connection_id"`
	InstallationID  string              `json:"installation_id"`
	ReleaseID       string              `json:"release_id"`
	AttemptID       string              `json:"attempt_id"`
	Query           map[string][]string `json:"query"`
	Input           json.RawMessage     `json:"input"`
	Response        json.RawMessage     `json:"response"`
	ResponseHeaders map[string]string   `json:"response_headers"`
	Status          int                 `json:"status"`
}

func TestAgentPluginSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_plugin_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentPluginHTTPContract
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
			result, args, err := invokeAgentPluginSDK(t, client, scenario)
			assertAgentPluginError(t, err, scenario)
			assertAgentPluginObservation(t, scenario, <-observations)
			if scenario.Status < 400 {
				assertAgentPluginResult(t, result, scenario)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			assertAgentPluginError(t, cliErr, scenario)
			assertAgentPluginObservation(t, scenario, <-observations)
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			assertAgentPluginCLI(t, []byte(output), scenario)
		})
	}
}

func invokeAgentPluginSDK(t *testing.T, client *sdk.Client, scenario agentPluginHTTPContract) (proto.Message, []string, error) {
	t.Helper()
	service := client.Agents("workspace-1")
	args := []string{scenario.Operation}
	switch {
	case scenario.ConnectionID != "":
		args = append(args, "--id", scenario.ConnectionID)
	case scenario.InstallationID != "":
		args = append(args, "--id", scenario.InstallationID)
	case scenario.ReleaseID != "":
		args = append(args, "--release-id", scenario.ReleaseID)
	}
	if scenario.AttemptID != "" {
		args = append(args, "--attempt-id", scenario.AttemptID)
	}
	query := make([]sdk.QueryOption, 0)
	for _, key := range mapsKeys(scenario.Query) {
		for _, value := range scenario.Query[key] {
			query = append(query, sdk.WithQueryValues(key, value))
			args = append(args, "--"+key, value)
		}
	}
	var versions struct {
		ExpectedVersion    int64 `json:"expected_version"`
		ExpectedGeneration int64 `json:"expected_generation"`
	}
	if len(scenario.Input) != 0 {
		decodeAgentChannelCallbackInput(t, scenario.Input, &versions)
	}
	switch scenario.Operation {
	case "delete-plugin-installation", "delete-plugin-connection", "begin-plugin-oauth", "poll-plugin-oauth":
		args = append(args, "--expected-version", strconv.FormatInt(versions.ExpectedVersion, 10))
	case "disconnect-plugin-connection":
		args = append(args, "--expected-generation", strconv.FormatInt(versions.ExpectedGeneration, 10))
	case "install-plugin", "update-plugin-installation", "create-plugin-connection", "update-plugin-connection", "store-plugin-secret":
		args = append(args, "--json", string(scenario.Input))
	}

	var result proto.Message
	var err error
	switch scenario.Operation {
	case "search-plugins":
		result, err = service.SearchPlugins(t.Context(), query...)
	case "describe-plugin-capability":
		result, err = service.DescribePluginCapability(t.Context(), agentPluginQueryValue(scenario, "ref"))
	case "list-plugin-releases":
		result, err = service.ListPluginReleases(t.Context(), query...)
	case "get-plugin-release":
		result, err = service.GetPluginRelease(t.Context(), scenario.ReleaseID, query...)
	case "install-plugin":
		var request sdk.PluginInstallRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.InstallPlugin(t.Context(), &request)
	case "list-plugin-installations":
		result, err = service.ListPluginInstallations(t.Context(), query...)
	case "get-plugin-installation":
		result, err = service.GetPluginInstallation(t.Context(), scenario.InstallationID)
	case "update-plugin-installation":
		var request sdk.PluginUpdateInstallationRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.UpdatePluginInstallation(t.Context(), scenario.InstallationID, &request)
	case "delete-plugin-installation":
		result, err = service.DeletePluginInstallation(t.Context(), scenario.InstallationID, versions.ExpectedVersion)
	case "create-plugin-connection":
		var request sdk.PluginCreateConnectionRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.CreatePluginConnection(t.Context(), &request)
	case "list-plugin-connections":
		result, err = service.ListPluginConnections(t.Context(), query...)
	case "get-plugin-connection":
		result, err = service.GetPluginConnection(t.Context(), scenario.ConnectionID)
	case "update-plugin-connection":
		var request sdk.PluginUpdateConnectionRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.UpdatePluginConnection(t.Context(), scenario.ConnectionID, &request)
	case "delete-plugin-connection":
		result, err = service.DeletePluginConnection(t.Context(), scenario.ConnectionID, int(versions.ExpectedVersion))
	case "get-plugin-authorization":
		result, err = service.GetPluginAuthorization(t.Context(), scenario.ConnectionID)
	case "store-plugin-secret":
		var request sdk.PluginStoreSecretRequest
		decodeAgentChannelCallbackInput(t, scenario.Input, &request)
		result, err = service.StorePluginSecret(t.Context(), scenario.ConnectionID, &request)
	case "begin-plugin-oauth":
		result, err = service.BeginPluginOAuth(t.Context(), scenario.ConnectionID, int(versions.ExpectedVersion))
	case "poll-plugin-oauth":
		result, err = service.PollPluginOAuth(t.Context(), scenario.ConnectionID, scenario.AttemptID, int(versions.ExpectedVersion))
	case "disconnect-plugin-connection":
		result, err = service.DisconnectPluginConnection(t.Context(), scenario.ConnectionID, versions.ExpectedGeneration)
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	return result, args, err
}

func mapsKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

func agentPluginQueryValue(scenario agentPluginHTTPContract, name string) string {
	if values := scenario.Query[name]; len(values) != 0 {
		return values[0]
	}
	return ""
}

func assertAgentPluginObservation(t *testing.T, scenario agentPluginHTTPContract, got agentChannelCallbackObservation) {
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

func assertAgentPluginResult(t *testing.T, result proto.Message, scenario agentPluginHTTPContract) {
	t.Helper()
	if !proto.Equal(result, agentPluginExpectedMessage(t, scenario)) {
		t.Fatalf("result=%v want=%v", result, agentPluginExpectedMessage(t, scenario))
	}
}

func assertAgentPluginCLI(t *testing.T, output []byte, scenario agentPluginHTTPContract) {
	t.Helper()
	expected := agentPluginExpectedMessage(t, scenario)
	actual := proto.Clone(expected)
	proto.Reset(actual)
	if err := protojson.Unmarshal(output, actual); err != nil {
		t.Fatalf("output=%s error=%v", output, err)
	}
	if !proto.Equal(actual, expected) {
		t.Fatalf("output=%s want=%v", output, expected)
	}
}

func agentPluginExpectedMessage(t *testing.T, scenario agentPluginHTTPContract) proto.Message {
	t.Helper()
	var message proto.Message
	switch scenario.Operation {
	case "search-plugins":
		message = &sdk.PluginDiscoveryPage{}
	case "describe-plugin-capability":
		message = &sdk.PluginCapabilityDescriptor{}
	case "list-plugin-releases":
		message = &sdk.PluginPublicationList{}
	case "get-plugin-release":
		message = &sdk.PluginRelease{}
	case "install-plugin", "get-plugin-installation", "update-plugin-installation":
		message = &sdk.PluginInstallation{}
	case "list-plugin-installations":
		message = &sdk.PluginInstallationList{}
	case "delete-plugin-installation", "delete-plugin-connection", "disconnect-plugin-connection":
		message = &sdk.PluginMutationResult{}
	case "create-plugin-connection", "get-plugin-connection", "update-plugin-connection":
		message = &sdk.PluginConnection{}
	case "list-plugin-connections":
		message = &sdk.PluginConnectionList{}
	case "get-plugin-authorization", "store-plugin-secret":
		message = &sdk.PluginAuthorizationStatus{}
	case "begin-plugin-oauth":
		message = &sdk.PluginOAuthAction{}
	case "poll-plugin-oauth":
		message = &sdk.PluginOAuthProgress{}
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

func assertAgentPluginError(t *testing.T, err error, scenario agentPluginHTTPContract) {
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
	message, _ := response["msg"].(string)
	if message == "" {
		message, _ = response["message"].(string)
	}
	if failure.Code != fmt.Sprint(response["code"]) || failure.Message != message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%#v", failure)
	}
	details, _ := response["details"].(map[string]any)
	if fmt.Sprint(failure.Details["reason"]) != fmt.Sprint(details["reason"]) || fmt.Sprint(failure.Details["domain"]) != fmt.Sprint(details["domain"]) {
		t.Fatalf("error details=%#v want=%#v", failure.Details, details)
	}
}
