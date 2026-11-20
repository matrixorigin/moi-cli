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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type agentPackageContract struct {
	Name            string            `json:"name"`
	Operation       string            `json:"operation"`
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	AgentID         string            `json:"agent_id"`
	Version         string            `json:"version"`
	FileBase64      string            `json:"file_base64"`
	Response        json.RawMessage   `json:"response"`
	ResponseBase64  string            `json:"response_base64"`
	ResponseHeaders map[string]string `json:"response_headers"`
	Status          int               `json:"status"`
}

type agentPackageObservation struct {
	method, path, accept, contentType                     string
	workspace, apiKey, authorization, cookie, runtimeAuth string
	body                                                  []byte
}

func TestAgentPackageSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_package_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentPackageContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			file := agentPackageBytes(t, scenario.FileBase64)
			response := agentPackageBytes(t, scenario.ResponseBase64)
			if len(response) == 0 {
				response = scenario.Response
			}
			observations := make(chan agentPackageObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				observations <- agentPackageObservation{
					method: r.Method, path: r.URL.EscapedPath(), accept: r.Header.Get("Accept"), contentType: r.Header.Get("Content-Type"),
					workspace: r.Header.Get("X-Workspace-ID"), apiKey: r.Header.Get("X-API-Key"), authorization: r.Header.Get("Authorization"),
					cookie: r.Header.Get("Cookie"), runtimeAuth: r.Header.Get("Runtime-Authorization"), body: body,
				}
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				if w.Header().Get("Content-Type") == "" {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			agents := client.Agents("workspace-1")
			var result *sdk.ProductData
			var stream *sdk.StreamResponse
			var args []string
			switch scenario.Operation {
			case "load-agent-package":
				result, err = agents.LoadAgentPackage(t.Context(), bytes.NewReader(file))
				args = []string{scenario.Operation, "--file-path", uploadFixture(t, "package.moiagent", file)}
			case "export-agent-package":
				stream, err = agents.ExportAgentPackage(t.Context(), scenario.AgentID, scenario.Version)
				args = []string{scenario.Operation, "--agent-id", scenario.AgentID, "--version", scenario.Version, "--output", filepath.Join(t.TempDir(), "package.moiagent")}
			default:
				t.Fatalf("unknown operation %s", scenario.Operation)
			}
			assertAgentPackageError(t, err, scenario)
			assertAgentPackageObservation(t, scenario, <-observations, file)
			if scenario.Status < 400 {
				if scenario.Operation == "load-agent-package" {
					assertAgentPackageLoadResult(t, result.Value, scenario)
				} else {
					content, readErr := io.ReadAll(stream.Body)
					if closeErr := stream.Body.Close(); readErr != nil || closeErr != nil {
						t.Fatalf("read=%v close=%v", readErr, closeErr)
					}
					assertAgentPackageStream(t, stream.StatusCode, stream.Header, content, scenario, response)
				}
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			assertAgentPackageError(t, cliErr, scenario)
			assertAgentPackageObservation(t, scenario, <-observations, file)
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			if scenario.Operation == "load-agent-package" {
				value := agentResourceJSON([]byte(output))
				assertAgentPackageLoadResult(t, value, scenario)
			} else {
				content, err := os.ReadFile(args[len(args)-1])
				if err != nil || !bytes.Equal(content, response) {
					t.Fatalf("CLI export bytes=%v err=%v", content, err)
				}
			}
		})
	}
}

func agentPackageBytes(t *testing.T, encoded string) []byte {
	t.Helper()
	if encoded == "" {
		return nil
	}
	value, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func assertAgentPackageObservation(t *testing.T, scenario agentPackageContract, got agentPackageObservation, file []byte) {
	t.Helper()
	if got.method != scenario.Method || got.path != scenario.Path || got.workspace != "workspace-1" || got.apiKey != "synthetic-key" || got.authorization != "" || got.cookie != "" || got.runtimeAuth != "" {
		t.Fatalf("request=%#v", got)
	}
	if scenario.Operation == "load-agent-package" {
		if got.contentType != "application/vnd.moi.agent+zip" || got.accept != "application/json" || !bytes.Equal(got.body, file) {
			t.Fatalf("upload=%#v body=%v", got, got.body)
		}
	} else if got.accept != "*/*" || len(got.body) != 0 {
		t.Fatalf("download=%#v", got)
	}
}

func assertAgentPackageLoadResult(t *testing.T, actual any, scenario agentPackageContract) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, agentResourceJSON(envelope.Data)) {
		t.Fatalf("result=%#v want=%#v", actual, agentResourceJSON(envelope.Data))
	}
}

func assertAgentPackageStream(t *testing.T, status int, header http.Header, content []byte, scenario agentPackageContract, expected []byte) {
	t.Helper()
	if status != scenario.Status || !bytes.Equal(content, expected) {
		t.Fatalf("stream status=%d bytes=%v", status, content)
	}
	for key, value := range scenario.ResponseHeaders {
		if header.Get(key) != value {
			t.Fatalf("stream header %s=%q want=%q", key, header.Get(key), value)
		}
	}
}

func assertAgentPackageError(t *testing.T, err error, scenario agentPackageContract) {
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
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	message := envelope.Msg
	if message == "" {
		message = envelope.Message
	}
	if failure.Code != fmt.Sprint(envelope.Code) || failure.Message != message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v want code=%v message=%q status=%d request=%q", failure, envelope.Code, message, scenario.Status, scenario.ResponseHeaders["X-Request-ID"])
	}
	if len(envelope.Details) > 0 && !reflect.DeepEqual(map[string]any(failure.Details), agentResourceJSON(envelope.Details)) {
		t.Fatalf("details=%#v want=%#v", failure.Details, agentResourceJSON(envelope.Details))
	}
}
