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

type agentPublicCallbackContract struct {
	Name                string              `json:"name"`
	Operation           string              `json:"operation"`
	Method              string              `json:"method"`
	Path                string              `json:"path"`
	Provider            string              `json:"provider"`
	WorkspaceID         string              `json:"workspace_id"`
	CredentialRef       string              `json:"credential_ref"`
	ConnectionID        string              `json:"connection_id"`
	Query               map[string][]string `json:"query"`
	BodyBase64          string              `json:"body_base64"`
	ContentType         string              `json:"content_type"`
	Accept              string              `json:"accept"`
	ResponseBase64      string              `json:"response_base64"`
	ResponseContentType string              `json:"response_content_type"`
	ResponseHeaders     map[string]string   `json:"response_headers"`
	Status              int                 `json:"status"`
}

type agentPublicCallbackObservation struct {
	method, path, contentType, accept                            string
	workspace, apiKey, authorization, cookie, runtime, signature string
	query                                                        url.Values
	body                                                         []byte
}

func TestAgentPublicCallbackSharedContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_public_callback_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentPublicCallbackContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			body := agentPackageBytes(t, scenario.BodyBase64)
			response := agentPackageBytes(t, scenario.ResponseBase64)
			observations := make(chan agentPublicCallbackObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				observations <- agentPublicCallbackObservation{
					method: r.Method, path: r.URL.EscapedPath(), query: r.URL.Query(), body: raw,
					contentType: r.Header.Get("Content-Type"), accept: r.Header.Get("Accept"),
					workspace: r.Header.Get("X-Workspace-ID"), apiKey: r.Header.Get("X-API-Key"), authorization: r.Header.Get("Authorization"),
					cookie: r.Header.Get("Cookie"), runtime: r.Header.Get("X-MOI-Runtime-Authorization"), signature: r.Header.Get("X-Provider-Signature"),
				}
				w.Header().Set("Content-Type", scenario.ResponseContentType)
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("private-key"), sdk.WithCustomHeader("X-Provider-Signature", "provider-signature"))
			if err != nil {
				t.Fatal(err)
			}
			result, args, err := invokeAgentPublicCallbackSDK(t, client, scenario, body)
			assertAgentPublicCallbackError(t, err, scenario, response)
			assertAgentPublicCallbackObservation(t, scenario, <-observations, body)
			if scenario.Status < 400 && !reflect.DeepEqual(result, response) {
				t.Fatalf("SDK response=%q want=%q", result, response)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, scenario.WorkspaceID, false, true, true), args)
			})
			assertAgentPublicCallbackError(t, cliErr, scenario, response)
			assertAgentPublicCallbackObservation(t, scenario, <-observations, body)
			if scenario.Status < 400 && output != string(response) {
				t.Fatalf("CLI response=%q want=%q", output, response)
			}
			if scenario.Status >= 400 && output != "" {
				t.Fatalf("CLI error output=%q", output)
			}
		})
	}
}

func invokeAgentPublicCallbackSDK(t *testing.T, client *sdk.Client, scenario agentPublicCallbackContract, body []byte) ([]byte, []string, error) {
	t.Helper()
	args := []string{scenario.Operation, "--provider", scenario.Provider}
	if scenario.CredentialRef != "" {
		args = append(args, "--credential-ref", scenario.CredentialRef)
	}
	if scenario.ConnectionID != "" {
		args = append(args, "--connection-id", scenario.ConnectionID)
	}
	if len(body) != 0 {
		args = append(args, "--body-file", uploadFixture(t, "callback.bin", body))
	}
	query := make([]sdk.QueryOption, 0)
	for key, values := range scenario.Query {
		for _, value := range values {
			query = append(query, sdk.WithQueryValues(key, value))
			args = append(args, "--query-param", key+"="+value)
		}
	}
	var result []byte
	var err error
	switch scenario.Operation {
	case "verify-channel-callback":
		result, err = client.VerifyChannelCallback(t.Context(), scenario.Provider, scenario.WorkspaceID, scenario.CredentialRef, query...)
	case "post-channel-callback":
		result, err = client.PostChannelCallback(t.Context(), scenario.Provider, scenario.WorkspaceID, scenario.CredentialRef, body, query...)
	case "post-we-com-callback":
		result, err = client.PostWeComCallback(t.Context(), scenario.WorkspaceID, scenario.CredentialRef, body, query...)
	case "verify-ingress-callback":
		result, err = client.VerifyIngressCallback(t.Context(), scenario.Provider, scenario.WorkspaceID, scenario.ConnectionID, query...)
	case "post-ingress-callback":
		result, err = client.PostIngressCallback(t.Context(), scenario.Provider, scenario.WorkspaceID, scenario.ConnectionID, body, query...)
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	return result, args, err
}

func assertAgentPublicCallbackObservation(t *testing.T, scenario agentPublicCallbackContract, got agentPublicCallbackObservation, body []byte) {
	t.Helper()
	wantQuery := url.Values(scenario.Query)
	if scenario.Query == nil {
		wantQuery = url.Values{}
	}
	if got.method != scenario.Method || got.path != scenario.Path || !reflect.DeepEqual(got.query, wantQuery) || !bytes.Equal(got.body, body) {
		t.Fatalf("request=%s %s query=%v body=%q", got.method, got.path, got.query, got.body)
	}
	if got.contentType != scenario.ContentType || got.accept != scenario.Accept || got.signature != "provider-signature" {
		t.Fatalf("headers=%#v", got)
	}
	if got.workspace != "" || got.apiKey != "" || got.authorization != "" || got.cookie != "" || got.runtime != "" {
		t.Fatalf("product authentication leaked into public callback: %#v", got)
	}
}

func assertAgentPublicCallbackError(t *testing.T, err error, scenario agentPublicCallbackContract, raw []byte) {
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
	response := agentResourceJSON(raw).(map[string]any)
	message, _ := response["message"].(string)
	if failure.Code != fmt.Sprint(response["code"]) || failure.Message != message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%#v", failure)
	}
}
