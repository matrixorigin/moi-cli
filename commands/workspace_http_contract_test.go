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
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type workspaceCLIContract struct {
	Name      string            `json:"name"`
	Operation string            `json:"operation"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     map[string]string `json:"query"`
	Input     struct {
		Name        string `json:"name"`
		WorkspaceID string `json:"workspace_id"`
	} `json:"input"`
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
	Status   int             `json:"status"`
}

func workspaceCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func TestWorkspaceCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/workspace_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []workspaceCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var received any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, value := range scenario.Query {
					expectedQuery.Set(key, value)
				}
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-API-Key") != "user-key" || r.Header.Get("X-Workspace-ID") != "" {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				if len(raw) > 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					_ = decoder.Decode(&received)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{scenario.Operation}
			switch scenario.Operation {
			case "list":
				if name, ok := scenario.Query["name"]; ok {
					args = append(args, "--query-param", "name="+name)
				}
			case "create", "start-creation":
				args = append(args, "--name", scenario.Input.Name)
			case "claim", "delete":
				args = append(args, "--workspace-id", scenario.Input.WorkspaceID)
			case "update":
				args = append(args, "--workspace-id", scenario.Input.WorkspaceID, "--name", scenario.Input.Name)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteWorkspaceCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if scenario.Status >= 400 {
				var envelope struct {
					Code string `json:"code"`
					Msg  string `json:"msg"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				var failure *sdk.Error
				if !errors.As(callErr, &failure) || failure.Code != envelope.Code || failure.Message != envelope.Msg || failure.Status != scenario.Status || output != "" {
					t.Fatalf("error=%#v output=%q", callErr, output)
				}
			} else {
				if callErr != nil {
					t.Fatal(callErr)
				}
				assertWorkspaceCLIOutput(t, scenario, output)
			}
			if !reflect.DeepEqual(received, workspaceCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, workspaceCLIJSON(scenario.Request))
			}
		})
	}
}

func assertWorkspaceCLIOutput(t *testing.T, scenario workspaceCLIContract, output string) {
	t.Helper()
	if scenario.Operation == "update" {
		if output != "" {
			t.Fatalf("output=%q", output)
		}
		return
	}
	if scenario.Operation == "delete" {
		if output != "workspace delete: ok\n" {
			t.Fatalf("output=%q", output)
		}
		return
	}
	var got proto.Message
	switch scenario.Operation {
	case "list":
		got = &sdk.WorkspaceListResult{}
	case "create":
		got = &sdk.WorkspaceCreateResult{}
	case "claim":
		got = &sdk.WorkspaceClaimResult{}
	case "creations":
		got = &sdk.WorkspaceCreationsResult{}
	case "start-creation":
		got = &sdk.WorkspaceCreationJob{}
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	if err := json.Unmarshal([]byte(output), got); err != nil {
		t.Fatalf("output=%s error=%v", output, err)
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	expected := got.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(got, expected) {
		t.Fatalf("result=%v want=%v", got, expected)
	}
}
