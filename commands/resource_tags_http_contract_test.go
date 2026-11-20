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
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestResourceTagsCLIMatchSharedHTTPContract(t *testing.T) {
	var cases []struct {
		Name, Operation, Method, Path string
		TagID                         string `json:"tag_id"`
		ValueID                       string `json:"value_id"`
		JSONFile                      bool   `json:"json_file"`
		Input, Request, Response      json.RawMessage
		Query                         map[string]string
		Status                        int
	}
	data, err := os.ReadFile("../testdata/product-contracts/resource_tags_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			decodeJSON := func(data []byte) any {
				t.Helper()
				var value any
				decoder := json.NewDecoder(bytes.NewReader(data))
				decoder.UseNumber()
				if err := decoder.Decode(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			query := url.Values{}
			for key, value := range scenario.Query {
				query.Set(key, value)
			}
			received := make(chan any, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.Path != scenario.Path || !reflect.DeepEqual(r.URL.Query(), query) || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected scoped request: %s %s", r.Method, r.URL.RequestURI())
				}
				raw, err := io.ReadAll(r.Body)
				var body any
				if err != nil {
					t.Error(err)
				} else if len(raw) > 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil {
						t.Error(err)
					}
				}
				received <- body
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
			args := []string{scenario.Operation}
			if scenario.TagID != "" {
				args = append(args, "--tag-id", scenario.TagID)
			}
			if scenario.ValueID != "" {
				args = append(args, "--value-id", scenario.ValueID)
			}
			if scenario.Method == http.MethodGet {
				var parameters map[string]json.RawMessage
				if err := json.Unmarshal(scenario.Input, &parameters); err != nil {
					t.Fatal(err)
				}
				for name, raw := range parameters {
					value := string(raw)
					if len(raw) > 0 && raw[0] == '"' {
						if err := json.Unmarshal(raw, &value); err != nil {
							t.Fatal(err)
						}
					}
					args = append(args, "--"+strings.ReplaceAll(name, "_", "-"), value)
				}
			} else if scenario.Method != http.MethodDelete {
				if scenario.JSONFile {
					path := filepath.Join(t.TempDir(), "request.json")
					if err := os.WriteFile(path, scenario.Input, 0600); err != nil {
						t.Fatal(err)
					}
					args = append(args, "--json-file", path)
				} else {
					args = append(args, "--json", string(scenario.Input))
				}
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteResourceTagCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			select {
			case body := <-received:
				if want := decodeJSON(scenario.Request); !reflect.DeepEqual(body, want) {
					t.Errorf("HTTP body=%v, want=%v", body, want)
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
					t.Fatalf("owner error changed: %v, output=%s", callErr, output)
				}
				return
			}
			if callErr != nil {
				t.Fatal(callErr)
			}
			var expected proto.Message
			switch scenario.Operation {
			case "create", "update":
				expected = &sdk.ResourceTag{}
			case "list":
				expected = &sdk.ResourceTagList{}
			case "create-value", "update-value":
				expected = &sdk.ResourceTagValue{}
			case "list-values":
				expected = &sdk.ResourceTagValueList{}
			case "groups":
				expected = &sdk.ResourceTagGroupList{}
			case "resources":
				expected = &sdk.ResourceTagResourceList{}
			case "assignments":
				expected = &sdk.ResourceTagAssignmentsList{}
			}
			if expected == nil {
				if output != "" {
					t.Fatalf("unexpected mutation output: %s", output)
				}
				return
			}
			if err := protojson.Unmarshal(response.Data, expected); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(expected)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := decodeJSON([]byte(output)), decodeJSON(encoded); !reflect.DeepEqual(got, want) {
				t.Errorf("CLI response=%s, want=%s", output, encoded)
			}
		})
	}
}
