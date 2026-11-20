package commands

import (
	"bytes"
	"encoding/json"
	"errors"
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

type agentAccountResourceCLIContract struct {
	Name      string              `json:"name"`
	Operation string              `json:"operation"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Auth      string              `json:"auth"`
	Query     map[string][]string `json:"query"`
	Input     json.RawMessage     `json:"input"`
	Request   json.RawMessage     `json:"request"`
	Response  json.RawMessage     `json:"response"`
	Status    int                 `json:"status"`
	Resource  string              `json:"resource"`
	TailPath  string              `json:"tail_path"`
	TicketID  string              `json:"ticket_id"`
	Kind      string              `json:"kind"`
	Title     string              `json:"title"`
	BodyText  string              `json:"body_text"`
	Retry     bool                `json:"retry"`
}

func agentAccountResourceCLIJSON(raw json.RawMessage) any {
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

func agentAccountResourceCLIArgs(scenario agentAccountResourceCLIContract) []string {
	var args []string
	switch scenario.Operation {
	case "resource-any":
		args = []string{"resource-any", "--method", scenario.Method, "--resource", scenario.Resource, "--tail-path", scenario.TailPath}
	case "admin-feedback-ticket", "update-admin-feedback-ticket", "create-feedback-github-issue", "build-feedback-github-issue-draft", "approve-feedback-reward":
		args = []string{scenario.Operation, "--ticket-id", scenario.TicketID}
	default:
		args = []string{scenario.Operation}
	}
	switch scenario.Operation {
	case "update-admin-feedback-ticket", "approve-feedback-reward", "create-feedback-ticket":
		args = append(args, "--json", string(scenario.Input))
	case "create-feedback-github-issue":
		args = append(args, "--kind", scenario.Kind, "--title", scenario.Title, "--body", scenario.BodyText)
		if scenario.Retry {
			args = append(args, "--retry")
		}
	case "build-feedback-github-issue-draft":
		args = append(args, "--kind", scenario.Kind)
	case "resource-any":
		if len(scenario.Input) > 0 && string(scenario.Input) != "null" {
			args = append(args, "--json", string(scenario.Input))
		}
	}
	keys := make([]string, 0, len(scenario.Query))
	for key := range scenario.Query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range scenario.Query[key] {
			args = append(args, "--query-param", key+"="+value)
		}
	}
	return args
}

func executeAgentAccountResourceCLI(ctx *Context, args []string) ([]byte, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteAgentCommand(ctx, args)
	os.Stdout = original
	closeErr := write.Close()
	output, readErr := io.ReadAll(read)
	readCloseErr := read.Close()
	return output, errors.Join(callErr, closeErr, readErr, readCloseErr)
}

func TestAgentAccountResourceCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_account_resource_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentAccountResourceCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var received any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, values := range scenario.Query {
					for _, value := range values {
						expectedQuery.Add(key, value)
					}
				}
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				if scenario.Auth == "token" {
					if r.Header.Get("Authorization") != "Bearer admin-token" || r.Header.Get("X-API-Key") != "" {
						t.Errorf("token headers=%v", r.Header)
					}
				} else if r.Header.Get("X-API-Key") != "user-key" || r.Header.Get("Authorization") != "" {
					t.Errorf("API key headers=%v", r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				if len(raw) > 0 {
					received = agentAccountResourceCLIJSON(raw)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()

			var client *sdk.Client
			if scenario.Auth == "token" {
				client, err = sdk.New(server.URL+"/newmoi", sdk.WithToken("admin-token"))
			} else {
				client, err = sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
			}
			if err != nil {
				t.Fatal(err)
			}
			ctx := NewContext(client, "workspace-1", false, true, true)
			output, err := executeAgentAccountResourceCLI(ctx, agentAccountResourceCLIArgs(scenario))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(received, agentAccountResourceCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, agentAccountResourceCLIJSON(scenario.Request))
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			if got := agentAccountResourceCLIJSON(output); !reflect.DeepEqual(got, agentAccountResourceCLIJSON(envelope.Data)) {
				t.Fatalf("output=%#v want=%#v", got, agentAccountResourceCLIJSON(envelope.Data))
			}
		})
	}
}
