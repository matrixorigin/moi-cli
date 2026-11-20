package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type providerBackendContract struct {
	Name        string              `json:"name"`
	Service     string              `json:"service"`
	Kind        string              `json:"kind"`
	Operation   string              `json:"operation"`
	CLICommand  string              `json:"cli_command"`
	CLIArgs     []string            `json:"cli_args"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	Query       map[string][]string `json:"query"`
	Input       json.RawMessage     `json:"input"`
	Request     json.RawMessage     `json:"request"`
	Response    json.RawMessage     `json:"response"`
	Result      json.RawMessage     `json:"result"`
	Status      int                 `json:"status"`
	BackendID   string              `json:"backend_id"`
	EndpointID  string              `json:"endpoint_id"`
	StatusValue string              `json:"status_value"`
	Error       struct {
		Status   int             `json:"status"`
		Response json.RawMessage `json:"response"`
	} `json:"error_response"`
}

type providerContractInput struct {
	Name                     *string  `json:"name"`
	Type                     *string  `json:"type"`
	ReasoningControlProtocol *string  `json:"reasoning_control_protocol"`
	APIKey                   *string  `json:"api_key"`
	Address                  *string  `json:"address"`
	TimeoutSeconds           *int32   `json:"timeout_seconds"`
	Models                   []string `json:"models"`
	SupportedMimeTypes       []string `json:"supported_mime_types"`
	Strategy                 *string  `json:"strategy"`
	HealthCheckInterval      *int32   `json:"health_check_interval"`
	MaxRetries               *int32   `json:"max_retries"`
	SessionAffinity          *bool    `json:"session_affinity"`
	Status                   *string  `json:"status"`
}

func providerContractJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func providerContractData(t *testing.T, raw json.RawMessage) json.RawMessage {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data
}

func decodeProviderInput(t *testing.T, raw json.RawMessage) providerContractInput {
	t.Helper()
	var input providerContractInput
	if len(raw) == 0 || string(raw) == "null" {
		return input
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatal(err)
	}
	return input
}

func providerCreateOptions(input providerContractInput) []sdk.ProviderBackendOption {
	var options []sdk.ProviderBackendOption
	if input.Type != nil {
		options = append(options, sdk.WithProviderBackendType(*input.Type))
	}
	if input.ReasoningControlProtocol != nil {
		options = append(options, sdk.WithProviderBackendReasoningControlProtocol(*input.ReasoningControlProtocol))
	}
	if input.APIKey != nil {
		options = append(options, sdk.WithProviderBackendAPIKey(*input.APIKey))
	}
	if input.Address != nil {
		options = append(options, sdk.WithProviderBackendAddress(*input.Address))
	}
	if input.TimeoutSeconds != nil {
		options = append(options, sdk.WithProviderBackendTimeoutSeconds(*input.TimeoutSeconds))
	}
	if input.Models != nil {
		options = append(options, sdk.WithProviderBackendModels(input.Models...))
	}
	if input.SupportedMimeTypes != nil {
		options = append(options, sdk.WithProviderBackendSupportedMimeTypes(input.SupportedMimeTypes...))
	}
	return options
}

func providerUpdateOptions(input providerContractInput) []sdk.ProviderBackendUpdateOption {
	var options []sdk.ProviderBackendUpdateOption
	if input.Name != nil {
		options = append(options, sdk.WithProviderBackendUpdateName(*input.Name))
	}
	if input.APIKey != nil {
		options = append(options, sdk.WithProviderBackendUpdateAPIKey(*input.APIKey))
	}
	if input.ReasoningControlProtocol != nil {
		options = append(options, sdk.WithProviderBackendUpdateReasoningControlProtocol(*input.ReasoningControlProtocol))
	}
	if input.TimeoutSeconds != nil {
		options = append(options, sdk.WithProviderBackendUpdateTimeoutSeconds(*input.TimeoutSeconds))
	}
	if input.Models != nil {
		options = append(options, sdk.WithProviderBackendUpdateModels(input.Models...))
	}
	if input.SupportedMimeTypes != nil {
		options = append(options, sdk.WithProviderBackendUpdateSupportedMimeTypes(input.SupportedMimeTypes...))
	}
	return options
}

func providerRouterOptions(input providerContractInput) []sdk.RouterOption {
	var options []sdk.RouterOption
	if input.Strategy != nil {
		options = append(options, sdk.WithRouterStrategy(*input.Strategy))
	}
	if input.HealthCheckInterval != nil {
		options = append(options, sdk.WithRouterHealthCheckInterval(*input.HealthCheckInterval))
	}
	if input.MaxRetries != nil {
		options = append(options, sdk.WithRouterMaxRetries(*input.MaxRetries))
	}
	if input.SessionAffinity != nil {
		options = append(options, sdk.WithRouterSessionAffinity(*input.SessionAffinity))
	}
	return options
}

func providerString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func providerQueryOptions(scenario providerBackendContract) []sdk.QueryOption {
	var options []sdk.QueryOption
	for key, values := range scenario.Query {
		for _, value := range values {
			options = append(options, sdk.WithQueryParam(key, value))
		}
	}
	return options
}

func providerBackendHandle(client *sdk.Client, kind string) *sdk.ProviderBackendService {
	models := client.Models("workspace-1")
	switch kind {
	case "embedding":
		return models.EmbeddingBackends()
	case "parser":
		return models.ParserBackends()
	default:
		return models.LLMBackends()
	}
}

func callProviderBackendSDK(t *testing.T, ctx context.Context, client *sdk.Client, scenario providerBackendContract) (proto.Message, error) {
	t.Helper()
	input := decodeProviderInput(t, scenario.Input)
	models := client.Models("workspace-1")
	backends := providerBackendHandle(client, scenario.Kind)
	switch scenario.Operation {
	case "list":
		if scenario.Service == "model" {
			return models.List(ctx, providerQueryOptions(scenario)...)
		}
		return backends.List(ctx)
	case "list-models":
		return models.List(ctx, providerQueryOptions(scenario)...)
	case "list-embedding-models":
		return models.ListEmbeddingModels(ctx)
	case "create":
		return backends.Create(ctx, providerString(input.Name), providerCreateOptions(input)...)
	case "get":
		return backends.Get(ctx, scenario.BackendID)
	case "update":
		return backends.Update(ctx, scenario.BackendID, providerUpdateOptions(input)...)
	case "delete":
		return backends.Delete(ctx, scenario.BackendID)
	case "list-endpoints":
		return backends.ListEndpoints(ctx, scenario.BackendID)
	case "create-endpoint":
		return backends.CreateEndpoint(ctx, scenario.BackendID, providerString(input.Address))
	case "set-endpoint-status":
		return backends.SetEndpointStatus(ctx, scenario.BackendID, scenario.EndpointID, providerString(input.Status))
	case "probe-backend-models":
		return backends.ProbeBackendModels(ctx, scenario.BackendID, providerCreateOptions(input)...)
	case "router":
		return backends.Router(ctx)
	case "update-router":
		return backends.UpdateRouter(ctx, providerRouterOptions(input)...)
	case "probe-llm-models":
		return models.ProbeLLMModels(ctx, providerString(input.Address), providerCreateOptions(input)...)
	case "probe-embedding-models":
		return models.ProbeEmbeddingModels(ctx, providerString(input.Address), providerCreateOptions(input)...)
	case "setup-llm-backend":
		return models.SetupLLMBackend(ctx, providerString(input.Name), providerString(input.Address), input.Models, providerCreateOptions(input)...)
	default:
		return nil, errors.New("unknown provider backend operation: " + scenario.Operation)
	}
}

func providerContractServer(t *testing.T, scenario providerBackendContract, errorMode bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := url.Values{}
		for key, values := range scenario.Query {
			for _, value := range values {
				expectedQuery.Add(key, value)
			}
		}
		if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" {
			t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
		}
		raw, _ := io.ReadAll(r.Body)
		if !reflect.DeepEqual(providerContractJSON(raw), providerContractJSON(scenario.Request)) {
			t.Errorf("body=%s want=%s", raw, scenario.Request)
		}
		status := scenario.Status
		response := scenario.Response
		if errorMode {
			status = scenario.Error.Status
			response = scenario.Error.Response
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(response)
	}))
}

func providerContractClient(t *testing.T, server *httptest.Server) *sdk.Client {
	t.Helper()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func providerContractCLICommand(t *testing.T, command []string) []string {
	t.Helper()
	resolved := make([]string, len(command))
	for index, argument := range command {
		if argument == "{api_key}" {
			path := filepath.Join(t.TempDir(), "api-key")
			if err := os.WriteFile(path, []byte("secret"), 0o600); err != nil {
				t.Fatal(err)
			}
			argument = path
		}
		resolved[index] = argument
	}
	return resolved
}

func runProviderContractCLI(t *testing.T, client *sdk.Client, command []string) (string, error) {
	t.Helper()
	command = providerContractCLICommand(t, command)
	ctx := NewContext(client, "workspace-1", false, true, true)
	var cliErr error
	output := captureStdout(t, func() {
		switch command[0] {
		case "provider-backend":
			cliErr = ExecuteProviderBackendCommand(ctx, command[1:])
		case "model":
			cliErr = ExecuteModelCommand(ctx, command[1:])
		default:
			t.Fatalf("unknown CLI command %q", command[0])
		}
	})
	return output, cliErr
}

func assertProviderContractError(t *testing.T, err error, scenario providerBackendContract) {
	t.Helper()
	if err == nil {
		t.Fatal("expected provider backend error")
	}
	var apiErr *sdk.Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type=%T value=%v", err, err)
	}
	var envelope struct {
		Code    string `json:"code"`
		Message string `json:"msg"`
	}
	if err := json.Unmarshal(scenario.Error.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if apiErr.Code != envelope.Code || apiErr.Status != scenario.Error.Status || apiErr.Message != envelope.Message {
		t.Fatalf("error=%+v want code=%q status=%d message=%q", apiErr, envelope.Code, scenario.Error.Status, envelope.Message)
	}
}

func assertProviderContractResult(t *testing.T, scenario providerBackendContract, result proto.Message) {
	t.Helper()
	expectedRaw := scenario.Result
	if len(expectedRaw) == 0 || string(expectedRaw) == "null" {
		expectedRaw = providerContractData(t, scenario.Response)
	}
	expected := result.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal(expectedRaw, expected); err != nil {
		t.Fatalf("expected %s: %v", expectedRaw, err)
	}
	if !proto.Equal(result, expected) {
		t.Fatalf("SDK result=%v want=%v", result, expected)
	}
}

func TestProviderBackendSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/provider_backend_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []providerBackendContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 37 {
		t.Fatalf("provider backend contract cases=%d want=37", len(cases))
	}
	for _, scenario := range cases {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			server := providerContractServer(t, scenario, false)
			defer server.Close()
			client := providerContractClient(t, server)

			result, err := callProviderBackendSDK(t, t.Context(), client, scenario)
			if err != nil {
				t.Fatal(err)
			}
			assertProviderContractResult(t, scenario, result)

			output, err := runProviderContractCLI(t, client, scenario.CLIArgs)
			if err != nil {
				t.Fatal(err)
			}
			actual := result.ProtoReflect().New().Interface()
			if err := protojson.Unmarshal([]byte(output), actual); err != nil {
				t.Fatalf("CLI output=%q error=%v", output, err)
			}
			assertProviderContractResult(t, scenario, actual)

			errorServer := providerContractServer(t, scenario, true)
			defer errorServer.Close()
			errorClient := providerContractClient(t, errorServer)
			if _, err := callProviderBackendSDK(t, t.Context(), errorClient, scenario); err == nil {
				t.Fatal("expected SDK error")
			} else {
				assertProviderContractError(t, err, scenario)
			}
			errorOutput, cliErr := runProviderContractCLI(t, errorClient, scenario.CLIArgs)
			assertProviderContractError(t, cliErr, scenario)
			if errorOutput != "" {
				t.Fatalf("CLI error output=%q", errorOutput)
			}
		})
	}
}
