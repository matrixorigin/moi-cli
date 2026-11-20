package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestWorkflowTemplateCLIOptionsMatchSharedSDKContract(t *testing.T) {
	var cases []struct {
		Name, Command, ID, Method, Path string
		Request, Response               json.RawMessage
		Status                          int
	}
	data, err := os.ReadFile("../testdata/product-contracts/workflow_template_options_contract.json")
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
				if r.Method != scenario.Method || r.URL.Path != "/newmoi"+scenario.Path || r.URL.RawQuery != "" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
				}
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				var got any
				if want == nil {
					if len(raw) != 0 {
						t.Errorf("unexpected body=%s", raw)
					}
				} else {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					if err := decoder.Decode(&got); err != nil {
						t.Error(err)
					}
				}
				received <- got
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{scenario.Command}
			if scenario.Method != "POST" {
				args = append(args, "--template-id", scenario.ID)
			}
			var input map[string]json.RawMessage
			if err := json.Unmarshal(scenario.Request, &input); err != nil {
				t.Fatal(err)
			}
			flags := map[string]string{"name": "name", "dsl_yaml": "dsl-yaml", "template_key": "template-key", "description": "description", "runtime_fields": "input-form", "is_builtin": "is-builtin", "workflow": "json"}
			for field, value := range input {
				name, ok := flags[field]
				if !ok {
					continue
				}
				var text string
				if value[0] == '"' {
					if err := json.Unmarshal(value, &text); err != nil {
						t.Fatal(err)
					}
				} else {
					text = string(value)
				}
				args = append(args, "--"+name+"="+text)
			}
			var callErr error
			output := captureStdout(t, func() { callErr = ExecuteWorkflowCommand(NewContext(client, "workspace-1", false, true, true), args) })
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
			} else {
				if callErr != nil {
					t.Fatal(callErr)
				}
				var expected sdk.WorkflowTemplate
				if err := protojson.Unmarshal(response.Data, &expected); err != nil {
					t.Fatal(err)
				}
				encoded, err := json.Marshal(&expected)
				if err != nil {
					t.Fatal(err)
				}
				var actual, wantResponse any
				decoder := json.NewDecoder(bytes.NewReader([]byte(output)))
				decoder.UseNumber()
				if err := decoder.Decode(&actual); err != nil {
					t.Fatal(err)
				}
				decoder = json.NewDecoder(bytes.NewReader(encoded))
				decoder.UseNumber()
				if err := decoder.Decode(&wantResponse); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, wantResponse) {
					t.Errorf("response=%s, want=%s", output, encoded)
				}
			}
			select {
			case got := <-received:
				if !reflect.DeepEqual(got, want) {
					t.Errorf("request=%v, want=%v", got, want)
				}
			default:
				t.Fatal("no HTTP request received")
			}
		})
	}
}
