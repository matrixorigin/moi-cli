package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type miscManagementContract struct {
	Name                 string              `json:"name"`
	Operation            string              `json:"operation"`
	Method               string              `json:"method"`
	Path                 string              `json:"path"`
	Query                map[string][]string `json:"query"`
	Input                json.RawMessage     `json:"input"`
	Request              json.RawMessage     `json:"request"`
	Response             json.RawMessage     `json:"response"`
	ResponseBase64       string              `json:"response_base64"`
	ResponseContentType  string              `json:"response_content_type"`
	Status               int                 `json:"status"`
	Filename             string              `json:"filename"`
	ContentBase64        string              `json:"content_base64"`
	ContentType          string              `json:"content_type"`
	SkillID              string              `json:"skill_id"`
	FilePath             string              `json:"file_path"`
	Question             string              `json:"question"`
	RuntimeAuthorization string              `json:"runtime_authorization"`
}

func miscManagementJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func miscManagementData(t *testing.T, raw json.RawMessage) any {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	return miscManagementJSON(envelope.Data)
}

func TestMiscManagementSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/misc_management_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []miscManagementContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			content, err := base64.StdEncoding.DecodeString(scenario.ContentBase64)
			if err != nil {
				t.Fatal(err)
			}
			var response []byte
			if scenario.ResponseBase64 != "" {
				response, err = base64.StdEncoding.DecodeString(scenario.ResponseBase64)
			} else {
				response = scenario.Response
			}
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, values := range scenario.Query {
					for _, value := range values {
						expectedQuery.Add(key, value)
					}
				}
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-MOI-Runtime-Authorization") != scenario.RuntimeAuthorization {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				switch scenario.Operation {
				case "upload-workspace-file", "upload-runtime-artifact", "analyze-runtime-image":
					if err := r.ParseMultipartForm(1 << 20); err != nil {
						t.Errorf("multipart: %v", err)
						break
					}
					files := r.MultipartForm.File["file"]
					if len(files) != 1 || files[0].Filename != scenario.Filename || files[0].Header.Get("Content-Type") != scenario.ContentType {
						t.Errorf("multipart files=%v", files)
						break
					}
					file, err := files[0].Open()
					if err != nil {
						t.Error(err)
						break
					}
					raw, readErr := io.ReadAll(file)
					closeErr := file.Close()
					if readErr != nil || closeErr != nil || !bytes.Equal(raw, content) {
						t.Errorf("multipart content=%q read=%v close=%v", raw, readErr, closeErr)
					}
					if scenario.Operation == "analyze-runtime-image" && r.FormValue("question") != scenario.Question {
						t.Errorf("question=%q", r.FormValue("question"))
					}
				default:
					raw, _ := io.ReadAll(r.Body)
					if !reflect.DeepEqual(miscManagementJSON(raw), miscManagementJSON(scenario.Request)) {
						t.Errorf("body=%s want=%s", raw, scenario.Request)
					}
				}
				if scenario.ResponseContentType != "" {
					w.Header().Set("Content-Type", scenario.ResponseContentType)
				} else {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(response)
			}))
			defer server.Close()
			options := []sdk.Option{sdk.WithAPIKey("synthetic-key")}
			if scenario.RuntimeAuthorization != "" {
				options = append(options, sdk.WithRuntimeAuthorization(scenario.RuntimeAuthorization))
			}
			client, err := sdk.New(server.URL+"/newmoi", options...)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			var protoResult proto.Message
			var rawResult *sdk.ProductData
			var customResult any
			var streamResult []byte
			switch scenario.Operation {
			case "upload-workspace-file":
				rawResult, err = client.Agents("workspace-1").UploadWorkspaceFile(t.Context(), scenario.Filename, bytes.NewReader(content))
			case "get-skill-file":
				var stream *sdk.StreamResponse
				stream, err = client.Agents("workspace-1").GetSkillFile(t.Context(), scenario.SkillID, scenario.FilePath)
				if err == nil {
					streamResult, err = io.ReadAll(stream.Body)
					closeErr := stream.Body.Close()
					if err == nil {
						err = closeErr
					}
				}
			case "list-connector-catalog":
				customResult, err = client.Agents("workspace-1").ListConnectorCatalog(t.Context())
			case "invoke-dynamic-service":
				var spec sdk.DynamicServiceInvokeSpec
				if err = json.Unmarshal(scenario.Input, &spec); err == nil {
					customResult, err = client.Workflows("workspace-1").InvokeDynamicService(t.Context(), spec)
				}
			case "garbage-collection":
				request := &sdk.CatalogGarbageCollectionRequest{}
				if err = protojson.Unmarshal(scenario.Input, request); err == nil {
					protoResult, err = client.Catalog("workspace-1").GarbageCollection(t.Context(), request)
				}
			case "upload-runtime-artifact":
				protoResult, err = client.Catalog("workspace-1").UploadRuntimeArtifact(t.Context(), scenario.Filename, bytes.NewReader(content))
			case "analyze-runtime-image":
				protoResult, err = client.Catalog("workspace-1").AnalyzeRuntimeImage(t.Context(), scenario.Filename, bytes.NewReader(content), scenario.Question)
			default:
				t.Fatalf("unknown operation %q", scenario.Operation)
			}
			if err != nil {
				t.Fatal(err)
			}
			if streamResult != nil && !bytes.Equal(streamResult, response) {
				t.Fatalf("SDK stream=%q want=%q", streamResult, response)
			}

			outputPath := ""
			args := []string{}
			var domain string
			switch scenario.Operation {
			case "upload-workspace-file":
				domain = "agent"
				path := filepath.Join(t.TempDir(), scenario.Filename)
				if err := os.WriteFile(path, content, 0o600); err != nil {
					t.Fatal(err)
				}
				args = []string{scenario.Operation, "--file-path", path, "--filename", scenario.Filename}
			case "get-skill-file":
				domain = "agent"
				outputPath = filepath.Join(t.TempDir(), "skill.bin")
				args = []string{scenario.Operation, "--skill-id", scenario.SkillID, "--file-path", scenario.FilePath, "--output", outputPath}
			case "list-connector-catalog":
				domain = "agent"
				args = []string{scenario.Operation}
			case "invoke-dynamic-service":
				domain = "workflow"
				args = []string{scenario.Operation, "--json", string(scenario.Input)}
			case "garbage-collection":
				domain = "catalog"
				args = []string{scenario.Operation, "--json", string(scenario.Input)}
			case "upload-runtime-artifact":
				domain = "catalog"
				path := filepath.Join(t.TempDir(), scenario.Filename)
				if err := os.WriteFile(path, content, 0o600); err != nil {
					t.Fatal(err)
				}
				args = []string{scenario.Operation, "--file", path, "--name", scenario.Filename}
			case "analyze-runtime-image":
				domain = "catalog"
				path := filepath.Join(t.TempDir(), scenario.Filename)
				if err := os.WriteFile(path, content, 0o600); err != nil {
					t.Fatal(err)
				}
				args = []string{scenario.Operation, "--file", path, "--question", scenario.Question}
			}
			ctx := NewContext(client, "workspace-1", false, true, true)
			var cliErr error
			output := captureStdout(t, func() {
				switch domain {
				case "agent":
					cliErr = ExecuteAgentCommand(ctx, args)
				case "workflow":
					cliErr = ExecuteWorkflowCommand(ctx, args)
				case "catalog":
					cliErr = ExecuteCatalogCommand(ctx, args)
				}
			})
			if cliErr != nil {
				t.Fatal(cliErr)
			}
			if outputPath != "" {
				got, err := os.ReadFile(outputPath)
				if err != nil || !bytes.Equal(got, response) {
					t.Fatalf("CLI stream=%q want=%q error=%v", got, response, err)
				}
			}

			if protoResult != nil {
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				expected := protoResult.ProtoReflect().New().Interface()
				if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
					t.Fatal(err)
				}
				if !proto.Equal(protoResult, expected) {
					t.Fatalf("SDK result=%v want=%v", protoResult, expected)
				}
				actual := protoResult.ProtoReflect().New().Interface()
				if err := protojson.Unmarshal([]byte(output), actual); err != nil {
					t.Fatalf("CLI output=%q error=%v", output, err)
				}
				if !proto.Equal(actual, expected) {
					t.Fatalf("CLI result=%v want=%v", actual, expected)
				}
			}
			if rawResult != nil && (!reflect.DeepEqual(rawResult.Value, miscManagementData(t, scenario.Response)) || !reflect.DeepEqual(miscManagementJSON([]byte(output)), miscManagementData(t, scenario.Response))) {
				t.Fatalf("raw=%#v output=%q", rawResult.Value, output)
			}
			if customResult != nil {
				raw, err := json.Marshal(customResult)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(miscManagementJSON(raw), miscManagementData(t, scenario.Response)) || !reflect.DeepEqual(miscManagementJSON([]byte(output)), miscManagementData(t, scenario.Response)) {
					t.Fatalf("custom=%s output=%q", raw, output)
				}
			}
		})
	}
}
