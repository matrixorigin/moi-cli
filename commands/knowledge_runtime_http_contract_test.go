package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestKnowledgeRuntimeCLIMatchesSharedHTTPContract(t *testing.T) {
	var cases []struct {
		Name, Operation          string
		Input, Request, Response json.RawMessage
		Status                   int
		CLIArgs                  []string `json:"cli_args"`
	}
	data, err := os.ReadFile("../testdata/product-contracts/knowledge_runtime_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var want any
			decoder := json.NewDecoder(bytes.NewReader(scenario.Request))
			decoder.UseNumber()
			if err := decoder.Decode(&want); err != nil {
				t.Fatal(err)
			}
			received := make(chan any, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" || r.URL.Path != "/newmoi/semantic-models/runtime/"+scenario.Operation || r.URL.RawQuery != "" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-MOI-Runtime-Authorization") != "synthetic-live-task" {
					t.Errorf("unexpected scoped request: %s %s", r.Method, r.URL.RequestURI())
				}
				var got any
				decoder := json.NewDecoder(r.Body)
				decoder.UseNumber()
				if err := decoder.Decode(&got); err != nil {
					t.Error(err)
				}
				received <- got
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"), sdk.WithRuntimeAuthorization("synthetic-live-task"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{scenario.Operation}
			if len(scenario.CLIArgs) > 0 {
				args = append(args, scenario.CLIArgs...)
			} else {
				args = append(args, "--json", string(scenario.Input))
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteSemanticModelCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			select {
			case got := <-received:
				if !reflect.DeepEqual(got, want) {
					t.Errorf("request=%v, want=%v", got, want)
				}
			default:
				t.Fatal("no HTTP request received")
			}
			var response struct {
				Code, Msg string
				Data      json.RawMessage
			}
			if err := json.Unmarshal(scenario.Response, &response); err != nil {
				t.Fatal(err)
			}
			if scenario.Status >= 400 {
				var failure *sdk.Error
				if !errors.As(callErr, &failure) || failure.Code != response.Code || failure.Message != response.Msg || failure.Status != scenario.Status || output != "" {
					t.Fatalf("owner failure changed: %v, output=%s", callErr, output)
				}
				return
			}
			if callErr != nil {
				t.Fatal(callErr)
			}
			var expected sdk.KnowledgeRuntimeResult
			if err := protojson.Unmarshal(response.Data, &expected); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(&expected)
			if err != nil {
				t.Fatal(err)
			}
			var actual, wantOutput any
			decoder = json.NewDecoder(bytes.NewReader([]byte(output)))
			decoder.UseNumber()
			if err := decoder.Decode(&actual); err != nil {
				t.Fatal(err)
			}
			decoder = json.NewDecoder(bytes.NewReader(encoded))
			decoder.UseNumber()
			if err := decoder.Decode(&wantOutput); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, wantOutput) {
				t.Errorf("result=%s, want=%s", output, encoded)
			}
		})
	}
}
