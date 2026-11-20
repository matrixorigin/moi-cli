package commands

import (
	"bytes"
	"encoding/base64"
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
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type webappPublicContract struct {
	Name, Operation, Method, Path string
	Query                         map[string][]string `json:"query"`
	Input                         struct {
		Body json.RawMessage `json:"body"`
	} `json:"input"`
	Request             json.RawMessage   `json:"request"`
	Response            json.RawMessage   `json:"response"`
	ResponseBase64      string            `json:"response_base64"`
	ResponseContentType string            `json:"response_content_type"`
	ResponseHeaders     map[string]string `json:"response_headers"`
	Status              int               `json:"status"`
}

type webappPublicObservation struct {
	method, path, contentType, accept                       string
	workspace, role, apiKey, authorization, cookie, runtime string
	probe                                                   string
	query                                                   url.Values
	body                                                    []byte
}

func TestWebappPublicSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/webapp_public_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []webappPublicContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			response := scenario.Response
			if scenario.ResponseBase64 != "" {
				response, err = base64.StdEncoding.DecodeString(scenario.ResponseBase64)
				if err != nil {
					t.Fatal(err)
				}
			}
			observations := make(chan webappPublicObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				observations <- webappPublicObservation{
					method: r.Method, path: r.URL.EscapedPath(), query: r.URL.Query(), body: raw,
					contentType: r.Header.Get("Content-Type"), accept: r.Header.Get("Accept"),
					workspace: r.Header.Get("X-Workspace-ID"), role: r.Header.Get("X-MOI-Role-ID"),
					apiKey: r.Header.Get("X-API-Key"), authorization: r.Header.Get("Authorization"), cookie: r.Header.Get("Cookie"),
					runtime: r.Header.Get("X-MOI-Runtime-Authorization"), probe: r.Header.Get("X-Webapp-Probe"),
				}
				contentType := scenario.ResponseContentType
				if contentType == "" {
					contentType = "application/json"
				}
				w.Header().Set("Content-Type", contentType)
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi",
				sdk.WithAPIKey("private-key"), sdk.WithToken("private-token"), sdk.WithSessionCookie("private-cookie"),
				sdk.WithRuntimeAuthorization("private-runtime-grant"), sdk.WithWorkspaceCurrentRoleID("workspace-1", "7"),
				sdk.WithCustomHeader("X-Webapp-Probe", "preserved"),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			result, stream, args, goErr := invokeWebappPublicSDK(t, client, scenario)
			assertWebappPublicObservation(t, scenario, <-observations)
			assertWebappPublicError(t, goErr, scenario)
			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteWebappCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			assertWebappPublicObservation(t, scenario, <-observations)
			assertWebappPublicError(t, cliErr, scenario)
			if webappPublicExpectsError(scenario) {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			if scenario.Operation == "a2a-stream" {
				if !bytes.Equal(stream, response) || output != string(response) {
					t.Fatalf("stream=%q output=%q want=%q", stream, output, response)
				}
				return
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			expected := agentResourceJSON(envelope.Data)
			if result == nil || !reflect.DeepEqual(result.Value, expected) {
				t.Fatalf("Go result=%#v want=%#v", result, expected)
			}
			if got := agentResourceJSON([]byte(output)); !reflect.DeepEqual(got, expected) {
				t.Fatalf("CLI result=%#v want=%#v", got, expected)
			}
		})
	}
}

func invokeWebappPublicSDK(t *testing.T, client *sdk.Client, scenario webappPublicContract) (*sdk.ProductData, []byte, []string, error) {
	t.Helper()
	args := []string{scenario.Operation}
	switch scenario.Operation {
	case "rename-conversation", "delete-conversation", "list-messages", "list-task-events":
		args = append(args, "--conversation-id", "conversation/1")
	}
	if scenario.Operation == "list-task-events" {
		args = append(args, "--task-id", "task/1")
	}
	var body map[string]any
	if len(scenario.Input.Body) != 0 {
		if err := unmarshalJSONUseNumber(scenario.Input.Body, &body); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--json", string(scenario.Input.Body))
	}
	keys := make([]string, 0, len(scenario.Query))
	for key := range scenario.Query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	opts := make([]sdk.QueryOption, 0)
	for _, key := range keys {
		for _, value := range scenario.Query[key] {
			opts = append(opts, sdk.WithQueryValues(key, value))
			args = append(args, "--query-param", key+"="+value)
		}
	}
	service := client.Webapp()
	var result *sdk.ProductData
	var err error
	switch scenario.Operation {
	case "resolve":
		result, err = service.Resolve(t.Context(), body)
	case "create-session":
		result, err = service.CreateSession(t.Context(), body)
	case "refresh-session":
		result, err = service.RefreshSession(t.Context(), body)
	case "create-conversation":
		result, err = service.CreateConversation(t.Context(), body, opts...)
	case "list-conversations":
		result, err = service.ListConversations(t.Context(), opts...)
	case "rename-conversation":
		result, err = service.RenameConversation(t.Context(), "conversation/1", body, opts...)
	case "delete-conversation":
		result, err = service.DeleteConversation(t.Context(), "conversation/1", opts...)
	case "list-messages":
		result, err = service.ListMessages(t.Context(), "conversation/1", opts...)
	case "list-task-events":
		result, err = service.ListTaskEvents(t.Context(), "conversation/1", "task/1", opts...)
	case "submit-feedback":
		result, err = service.SubmitFeedback(t.Context(), body, opts...)
	case "call-mcp-app":
		result, err = service.CallMCPApp(t.Context(), body, opts...)
	case "read-mcp-app-resource":
		result, err = service.ReadMCPAppResource(t.Context(), body, opts...)
	case "a2a-stream":
		var stream *sdk.StreamResponse
		stream, err = service.A2AStream(t.Context(), body, opts...)
		if err == nil {
			raw, readErr := io.ReadAll(stream.Body)
			closeErr := stream.Body.Close()
			if readErr != nil {
				err = readErr
			} else if closeErr != nil {
				err = closeErr
			}
			return nil, raw, args, err
		}
	default:
		t.Fatalf("unknown operation %q", scenario.Operation)
	}
	return result, nil, args, err
}

func webappPublicExpectsError(scenario webappPublicContract) bool {
	return scenario.Status >= 400 || scenario.Name == "malformed-envelope"
}

func assertWebappPublicError(t *testing.T, err error, scenario webappPublicContract) {
	t.Helper()
	if !webappPublicExpectsError(scenario) {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("error=%T %v", err, err)
	}
	if scenario.Name == "malformed-envelope" {
		if failure.Code != "ErrResponseInvalid" || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
			t.Fatalf("error=%+v", failure)
		}
		return
	}
	var envelope struct {
		Code    any    `json:"code"`
		Message string `json:"message"`
	}
	decoder := json.NewDecoder(bytes.NewReader(scenario.Response))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if failure.Code != fmt.Sprint(envelope.Code) || failure.Message != envelope.Message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v", failure)
	}
}

func assertWebappPublicObservation(t *testing.T, scenario webappPublicContract, got webappPublicObservation) {
	t.Helper()
	expectedQuery := url.Values{}
	for key, values := range scenario.Query {
		for _, value := range values {
			expectedQuery.Add(key, value)
		}
	}
	if got.method != scenario.Method || got.path != scenario.Path || !reflect.DeepEqual(got.query, expectedQuery) {
		t.Fatalf("request=%s %s query=%v", got.method, got.path, got.query)
	}
	if got.workspace != "" || got.role != "" || got.apiKey != "" || got.authorization != "" || got.cookie != "" || got.runtime != "" {
		t.Fatalf("public credentials leaked: %+v", got)
	}
	if got.probe != "preserved" {
		t.Fatalf("custom header=%q", got.probe)
	}
	expectedAccept := "application/json"
	if scenario.Operation == "a2a-stream" {
		expectedAccept = "text/event-stream"
	}
	if got.accept != expectedAccept {
		t.Fatalf("Accept=%q want=%q", got.accept, expectedAccept)
	}
	expectedContentType := ""
	if len(scenario.Input.Body) != 0 {
		expectedContentType = "application/json"
	}
	if got.contentType != expectedContentType {
		t.Fatalf("Content-Type=%q want=%q", got.contentType, expectedContentType)
	}
	var body any
	if len(got.body) != 0 {
		decoder := json.NewDecoder(bytes.NewReader(got.body))
		decoder.UseNumber()
		if err := decoder.Decode(&body); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(body, agentResourceJSON(scenario.Request)) {
		t.Fatalf("body=%#v want=%#v", body, agentResourceJSON(scenario.Request))
	}
}
