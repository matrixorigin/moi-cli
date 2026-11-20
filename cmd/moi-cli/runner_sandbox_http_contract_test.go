package main

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
	"testing"

	"github.com/matrixorigin/matrixflow/sdk/cli/commands"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestRunnerSandboxCLIMatchSharedHTTPContract(t *testing.T) {
	var cases []struct {
		Name, Domain, Operation, Method, Path string
		Args                                  []string `json:"cli_args"`
		JSONFile                              bool     `json:"json_file"`
		Input, Request, Response              json.RawMessage
		Query                                 map[string]string
		Status                                int
	}
	data, err := os.ReadFile("../../testdata/product-contracts/runner_sandbox_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			// CLI output uses process stdout, so these cases stay serial.
			decode := func(raw []byte) any {
				t.Helper()
				if len(raw) == 0 {
					return nil
				}
				var value any
				decoder := json.NewDecoder(bytes.NewReader(raw))
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
			received := make(chan []byte, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.Path != scenario.Path || !reflect.DeepEqual(r.URL.Query(), query) || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected scoped request: %s %s", r.Method, r.URL.RequestURI())
				}
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				received <- raw
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
			args := scenario.Args
			if scenario.Domain == "sandbox" && string(scenario.Input) != "null" {
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
			read, write, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			originalStdout := os.Stdout
			os.Stdout = write
			defer func() { os.Stdout = originalStdout }()
			callErr := executeCommand(commands.NewContext(client, "workspace-1", false, true, true), scenario.Domain, args)
			os.Stdout = originalStdout
			_ = write.Close()
			output, err := io.ReadAll(read)
			_ = read.Close()
			if err != nil {
				t.Fatal(err)
			}
			select {
			case raw := <-received:
				if !reflect.DeepEqual(decode(raw), decode(scenario.Request)) {
					t.Errorf("HTTP body=%s, want=%s", raw, scenario.Request)
				}
			default:
				t.Fatalf("no HTTP request received: %v", callErr)
			}
			var response struct {
				Code, Msg, Message string
				Data               json.RawMessage
			}
			if err := json.Unmarshal(scenario.Response, &response); err != nil {
				t.Fatal(err)
			}
			if scenario.Status >= 400 {
				var failure *sdk.Error
				message := response.Msg
				if message == "" {
					message = response.Message
				}
				if !errors.As(callErr, &failure) || failure.Code != response.Code || failure.Message != message || failure.Status != scenario.Status || len(output) != 0 {
					t.Fatalf("owner error changed: %v, output=%s", callErr, output)
				}
				return
			}
			if callErr != nil {
				t.Fatal(callErr)
			}
			if scenario.Operation == "delete" || scenario.Operation == "admin-delete" {
				if len(output) != 0 {
					t.Errorf("unexpected deletion output: %s", output)
				}
				return
			}
			message := func() proto.Message {
				if scenario.Domain == "runner" {
					return &sdk.AgentRunner{}
				}
				switch scenario.Operation {
				case "get-quota":
					return &sdk.SandboxQuotaStatus{}
				case "update-quota", "admin-update-quota":
					return &sdk.SandboxQuota{}
				case "admin-list":
					return &sdk.SandboxAdminListResult{}
				case "terminal-token":
					return &sdk.SandboxTerminalToken{}
				case "create-ssh-session":
					return &sdk.SandboxSSHSession{}
				default:
					return &sdk.AgentSandbox{}
				}
			}
			var expected any
			if scenario.Operation == "list" {
				var items []json.RawMessage
				if err := json.Unmarshal(response.Data, &items); err != nil {
					t.Fatal(err)
				}
				values := []proto.Message{}
				for _, item := range items {
					value := message()
					if err := protojson.Unmarshal(item, value); err != nil {
						t.Fatal(err)
					}
					values = append(values, value)
				}
				expected = values
			} else {
				value := message()
				if err := protojson.Unmarshal(response.Data, value); err != nil {
					t.Fatal(err)
				}
				expected = value
			}
			encoded, err := json.Marshal(expected)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decode(output), decode(encoded)) {
				t.Errorf("CLI result=%s, want=%s", output, encoded)
			}
		})
	}
}
