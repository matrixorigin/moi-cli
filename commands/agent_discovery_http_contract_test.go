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
	"sort"
	"strconv"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type agentDiscoveryContract struct {
	Name      string              `json:"name"`
	Operation string              `json:"operation"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Query     map[string][]string `json:"query"`
	Input     struct {
		Catalog            *string         `json:"catalog"`
		Status             *string         `json:"status"`
		Query              *string         `json:"query"`
		Surface            *string         `json:"surface"`
		IncludeDevelopment *bool           `json:"include_development"`
		Limit              *int            `json:"limit"`
		Offset             *int            `json:"offset"`
		AgentCode          string          `json:"agent_code"`
		AgentID            string          `json:"agent_id"`
		AgentWorkspaceID   string          `json:"agent_workspace_id"`
		Body               json.RawMessage `json:"body"`
	} `json:"input"`
	Request         json.RawMessage   `json:"request"`
	Response        json.RawMessage   `json:"response"`
	ResponseHeaders map[string]string `json:"response_headers"`
	Status          int               `json:"status"`
}

func TestAgentDiscoverySharedHTTPContracts(t *testing.T) {
	runAgentDiscoverySharedHTTPContracts(t, "agent_discovery_http_contract.json")
}

func TestAgentCatalogAndHomeGuideSharedHTTPContracts(t *testing.T) {
	runAgentDiscoverySharedHTTPContracts(t, "agent_catalog_home_http_contract.json")
}

func TestAgentWorkspaceProtocolSharedHTTPContracts(t *testing.T) {
	runAgentDiscoverySharedHTTPContracts(t, "agent_workspace_protocol_http_contract.json")
}

func runAgentDiscoverySharedHTTPContracts(t *testing.T, fixture string) {
	t.Helper()
	data, err := os.ReadFile("../testdata/product-contracts/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentDiscoveryContract
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
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				var body any
				if len(raw) > 0 {
					body = agentDiscoveryJSON(raw)
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

			direct, directErr := callAgentDiscoveryGo(t, client, scenario)
			assertAgentDiscoveryError(t, directErr, scenario)
			if got := <-bodies; !reflect.DeepEqual(got, agentDiscoveryJSON(scenario.Request)) {
				t.Fatalf("Go body=%#v want=%#v", got, agentDiscoveryJSON(scenario.Request))
			}
			if scenario.Status < 400 {
				assertAgentDiscoveryResult(t, direct, scenario, "Go")
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), agentDiscoveryCLIArgs(scenario))
			})
			assertAgentDiscoveryError(t, cliErr, scenario)
			if got := <-bodies; !reflect.DeepEqual(got, agentDiscoveryJSON(scenario.Request)) {
				t.Fatalf("CLI body=%#v want=%#v", got, agentDiscoveryJSON(scenario.Request))
			}
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			assertAgentDiscoveryResult(t, []byte(output), scenario, "CLI")
		})
	}
}

func callAgentDiscoveryGo(t *testing.T, client *sdk.Client, scenario agentDiscoveryContract) (any, error) {
	t.Helper()
	service := client.Agents("workspace-1")
	switch scenario.Operation {
	case "list":
		return service.List(t.Context(), agentDiscoveryQuery(scenario.Query)...)
	case "agent-card":
		return service.AgentCard(t.Context(), agentDiscoveryQuery(scenario.Query)...)
	case "list-agent-catalog-cards":
		return service.ListAgentCatalogCards(t.Context(), agentDiscoveryQuery(scenario.Query)...)
	case "agent-home-guide-sample-locale":
		return service.ResourceGet(t.Context(), "agent-home-guide", "/sample-locale")
	case "workspace-agent-card":
		return service.ResourceGet(t.Context(), "agents", "/agent-1/.well-known/agent-card.json", agentDiscoveryQuery(scenario.Query)...)
	case "workspace-a2a":
		body, ok := agentDiscoveryJSON(scenario.Input.Body).(map[string]any)
		if !ok {
			return nil, fmt.Errorf("workspace A2A body must be an object")
		}
		return service.ResourceAny(t.Context(), http.MethodPost, "agents", "/agent-1/a2a", body, agentDiscoveryQuery(scenario.Query)...)
	case "workspace-mcp":
		request := &sdk.RuntimeRPCRequest{}
		if err := unmarshalProductJSON(scenario.Input.Body, request); err != nil {
			return nil, err
		}
		return service.WorkspaceMCP(t.Context(), scenario.Input.AgentID, request)
	case "a2a":
		var request sdk.AgentA2ARequest
		if err := unmarshalJSONUseNumber(scenario.Input.Body, &request); err != nil {
			return nil, err
		}
		return service.A2A(t.Context(), request)
	default:
		return nil, fmt.Errorf("unknown operation %q", scenario.Operation)
	}
}

func agentDiscoveryQuery(values map[string][]string) []sdk.QueryOption {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]sdk.QueryOption, 0, len(keys))
	for _, key := range keys {
		result = append(result, sdk.WithQueryValues(key, values[key]...))
	}
	return result
}

func agentDiscoveryCLIArgs(scenario agentDiscoveryContract) []string {
	args := []string{scenario.Operation}
	input := scenario.Input
	switch scenario.Operation {
	case "list":
		for _, item := range []struct {
			name  string
			value *string
		}{{"catalog", input.Catalog}, {"status", input.Status}, {"query", input.Query}, {"surface", input.Surface}} {
			if item.value != nil {
				args = append(args, "--"+item.name, *item.value)
			}
		}
		if input.IncludeDevelopment != nil {
			args = append(args, "--include-development="+strconv.FormatBool(*input.IncludeDevelopment))
		}
		if input.Limit != nil {
			args = append(args, "--limit", strconv.Itoa(*input.Limit))
		}
		if input.Offset != nil {
			args = append(args, "--offset", strconv.Itoa(*input.Offset))
		}
	case "agent-card":
		if input.AgentCode != "" {
			args = append(args, "--agent-code", input.AgentCode)
		}
		args = append(args, "--agent-id", input.AgentID, "--agent-workspace-id", input.AgentWorkspaceID)
	case "list-agent-catalog-cards":
		if input.Catalog != nil {
			args = append(args, "--catalog", *input.Catalog)
		}
		if input.Query != nil {
			args = append(args, "--query", *input.Query)
		}
		if input.IncludeDevelopment != nil {
			args = append(args, "--include-development="+strconv.FormatBool(*input.IncludeDevelopment))
		}
		if input.Limit != nil {
			args = append(args, "--limit", strconv.Itoa(*input.Limit))
		}
		if input.Offset != nil {
			args = append(args, "--offset", strconv.Itoa(*input.Offset))
		}
	case "agent-home-guide-sample-locale":
		args = []string{"resource-any", "--method", http.MethodGet, "--resource", "agent-home-guide", "--tail-path", "/sample-locale"}
	case "workspace-agent-card":
		args = []string{"resource-any", "--method", http.MethodGet, "--resource", "agents", "--tail-path", "/agent-1/.well-known/agent-card.json"}
		args = append(args, agentResourceQueryArgs(scenario.Query)...)
	case "workspace-a2a":
		args = []string{"resource-any", "--method", http.MethodPost, "--resource", "agents", "--tail-path", "/agent-1/a2a", "--json", string(input.Body)}
		args = append(args, agentResourceQueryArgs(scenario.Query)...)
	case "workspace-mcp":
		args = append(args, "--agent-id", input.AgentID, "--json", string(input.Body))
	case "a2a":
		args = append(args, "--json", string(input.Body))
	}
	return args
}

func assertAgentDiscoveryResult(t *testing.T, actual any, scenario agentDiscoveryContract, client string) {
	t.Helper()
	if scenario.Operation == "workspace-mcp" {
		expected := &sdk.RuntimeRPCResponse{}
		if err := protojson.Unmarshal(scenario.Response, expected); err != nil {
			t.Fatal(err)
		}
		var result *sdk.RuntimeRPCResponse
		switch value := actual.(type) {
		case *sdk.RuntimeRPCResponse:
			result = value
		case []byte:
			result = &sdk.RuntimeRPCResponse{}
			if err := protojson.Unmarshal(value, result); err != nil {
				t.Fatalf("%s output=%s error=%v", client, value, err)
			}
		default:
			t.Fatalf("%s result type=%T", client, actual)
		}
		if !proto.Equal(result, expected) {
			t.Fatalf("%s result=%v want=%v", client, result, expected)
		}
		return
	}
	if scenario.Operation == "list-agent-catalog-cards" {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
			t.Fatal(err)
		}
		expected := &sdk.AgentCatalogCardListResult{}
		if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
			t.Fatal(err)
		}
		var result *sdk.AgentCatalogCardListResult
		switch value := actual.(type) {
		case *sdk.AgentCatalogCardListResult:
			result = value
		case []byte:
			result = &sdk.AgentCatalogCardListResult{}
			if err := protojson.Unmarshal(value, result); err != nil {
				t.Fatalf("%s output=%s error=%v", client, value, err)
			}
		default:
			t.Fatalf("%s result type=%T", client, actual)
		}
		if !proto.Equal(result, expected) {
			t.Fatalf("%s result=%v want=%v", client, result, expected)
		}
		return
	}
	var raw []byte
	switch value := actual.(type) {
	case *sdk.ProductData:
		var err error
		raw, err = json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
	case *sdk.AgentCardResult:
		raw = []byte(value.GetJson())
	case *sdk.AgentA2AResult:
		raw = []byte(value.GetJson())
	case []byte:
		if scenario.Operation != "a2a" && scenario.Operation != "agent-card" {
			raw = value
		} else {
			var output struct {
				JSON string `json:"json"`
			}
			if err := json.Unmarshal(value, &output); err != nil {
				t.Fatalf("%s output=%s error=%v", client, value, err)
			}
			raw = []byte(output.JSON)
		}
	default:
		t.Fatalf("%s result type=%T", client, actual)
	}
	expected := scenario.Response
	if scenario.Operation != "a2a" {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
			t.Fatal(err)
		}
		if len(envelope.Data) > 0 && string(envelope.Data) != "null" {
			expected = envelope.Data
		}
	}
	if !reflect.DeepEqual(agentDiscoveryJSON(raw), agentDiscoveryJSON(expected)) {
		t.Fatalf("%s result=%s want=%s", client, raw, expected)
	}
}

func agentDiscoveryJSON(raw json.RawMessage) any {
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

func assertAgentDiscoveryError(t *testing.T, err error, scenario agentDiscoveryContract) {
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
		Code    any             `json:"code"`
		Msg     string          `json:"msg"`
		Message string          `json:"message"`
		Details json.RawMessage `json:"details"`
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
		t.Fatalf("error=%+v want code=%v message=%q status=%d request=%q", failure, envelope.Code, message, scenario.Status, scenario.ResponseHeaders["X-Request-ID"])
	}
	if len(envelope.Details) > 0 && !reflect.DeepEqual(map[string]any(failure.Details), agentDiscoveryJSON(envelope.Details)) {
		t.Fatalf("details=%#v want=%#v", failure.Details, agentDiscoveryJSON(envelope.Details))
	}
}
