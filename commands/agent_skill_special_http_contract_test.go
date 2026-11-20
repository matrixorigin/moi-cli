package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type agentSkillSpecialContract struct {
	Name, Operation, Method, Path string
	Filename                      string                     `json:"filename"`
	FileBase64                    string                     `json:"file_base64"`
	Input                         map[string]json.RawMessage `json:"input"`
	MultipartFields               map[string]string          `json:"multipart_fields"`
	Request                       json.RawMessage            `json:"request"`
	Response                      json.RawMessage            `json:"response"`
	ResponseText                  string                     `json:"response_text"`
	ResponseBase64                string                     `json:"response_base64"`
	ResponseHeaders               map[string]string          `json:"response_headers"`
	Status                        int                        `json:"status"`
}

type agentSkillSpecialObservation struct {
	method, path, accept, contentType                     string
	workspace, apiKey, authorization, cookie, runtimeAuth string
	body, file                                            []byte
	filename                                              string
	fields                                                map[string]string
}

func agentSkillSpecialResponse(t *testing.T, scenario agentSkillSpecialContract) []byte {
	t.Helper()
	if scenario.ResponseBase64 != "" {
		value, err := base64.StdEncoding.DecodeString(scenario.ResponseBase64)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	if scenario.ResponseText != "" {
		return []byte(scenario.ResponseText)
	}
	return scenario.Response
}

func agentSkillSpecialOptions(t *testing.T, scenario agentSkillSpecialContract) ([]sdk.SkillPackageImportOption, []string) {
	t.Helper()
	var options []sdk.SkillPackageImportOption
	var args []string
	for _, field := range []string{"name", "description", "category"} {
		raw, exists := scenario.Input[field]
		if !exists {
			continue
		}
		var value string
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--"+field, value)
		switch field {
		case "name":
			options = append(options, sdk.WithSkillPackageImportName(value))
		case "description":
			options = append(options, sdk.WithSkillPackageImportDescription(value))
		case "category":
			options = append(options, sdk.WithSkillPackageImportCategory(value))
		}
	}
	if raw, exists := scenario.Input["tags"]; exists {
		var tags []string
		if err := json.Unmarshal(raw, &tags); err != nil {
			t.Fatal(err)
		}
		options = append(options, sdk.WithSkillPackageImportTags(tags))
		args = append(args, "--tags", string(raw))
	}
	return options, args
}

func assertAgentSkillSpecialObservation(t *testing.T, scenario agentSkillSpecialContract, got agentSkillSpecialObservation, file []byte) {
	t.Helper()
	if got.method != scenario.Method || got.path != scenario.Path || got.workspace != "workspace-1" || got.apiKey != "synthetic-key" || got.authorization != "" || got.cookie != "" || got.runtimeAuth != "" {
		t.Fatalf("request=%s %s headers=%#v", got.method, got.path, got)
	}
	switch scenario.Operation {
	case "inspect_skill_package", "import_skill_package":
		if !strings.HasPrefix(got.contentType, "multipart/form-data; boundary=") || got.accept != "application/json" || got.filename != scenario.Filename || !bytes.Equal(got.file, file) || !reflect.DeepEqual(got.fields, scenario.MultipartFields) {
			t.Fatalf("multipart=%#v want fields=%#v filename=%q", got, scenario.MultipartFields, scenario.Filename)
		}
	case "polish_skill_stream":
		if got.contentType != "application/json" || got.accept != "text/event-stream" || !reflect.DeepEqual(agentResourceJSON(got.body), agentResourceJSON(scenario.Request)) {
			t.Fatalf("stream request=%#v body=%s", got, got.body)
		}
	}
}

func TestAgentSkillSpecialSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_skill_special_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentSkillSpecialContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			file, err := base64.StdEncoding.DecodeString(scenario.FileBase64)
			if err != nil && scenario.FileBase64 != "" {
				t.Fatal(err)
			}
			response := agentSkillSpecialResponse(t, scenario)
			observations := make(chan agentSkillSpecialObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				observed := agentSkillSpecialObservation{
					method: r.Method, path: r.URL.EscapedPath(), accept: r.Header.Get("Accept"), contentType: r.Header.Get("Content-Type"),
					workspace: r.Header.Get("X-Workspace-ID"), apiKey: r.Header.Get("X-API-Key"), authorization: r.Header.Get("Authorization"),
					cookie: r.Header.Get("Cookie"), runtimeAuth: r.Header.Get("X-MOI-Runtime-Authorization"), fields: map[string]string{},
				}
				mediaType, parameters, _ := mime.ParseMediaType(observed.contentType)
				if strings.HasPrefix(mediaType, "multipart/") {
					reader := multipart.NewReader(r.Body, parameters["boundary"])
					for {
						part, err := reader.NextPart()
						if err == io.EOF {
							break
						}
						if err != nil {
							t.Errorf("multipart: %v", err)
							break
						}
						content, err := io.ReadAll(part)
						if err != nil {
							t.Errorf("multipart content: %v", err)
						}
						if part.FormName() == "file" {
							observed.filename, observed.file = part.FileName(), content
						} else {
							observed.fields[part.FormName()] = string(content)
						}
					}
				} else {
					observed.body, _ = io.ReadAll(r.Body)
				}
				observations <- observed
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
			ctx := NewContext(client, "workspace-1", false, true, true)
			var result *sdk.ProductData
			var stream *sdk.StreamResponse
			var args []string
			switch scenario.Operation {
			case "inspect_skill_package":
				result, err = agents.InspectSkillPackage(t.Context(), scenario.Filename, bytes.NewReader(file))
				args = []string{"inspect-skill-package", "--file-path", uploadFixture(t, scenario.Filename, file)}
			case "import_skill_package":
				options, optionArgs := agentSkillSpecialOptions(t, scenario)
				result, err = agents.ImportSkillPackage(t.Context(), scenario.Filename, bytes.NewReader(file), options...)
				args = append([]string{"import-skill-package", "--file-path", uploadFixture(t, scenario.Filename, file)}, optionArgs...)
			case "polish_skill_stream":
				draft, ok := agentResourceJSON(scenario.Input["body"]).(map[string]any)
				if !ok {
					t.Fatal("polish draft must be an object")
				}
				stream, err = agents.PolishSkillStream(t.Context(), draft)
				args = []string{"polish-skill-stream", "--json", string(scenario.Input["body"]), "--output", filepath.Join(t.TempDir(), "stream.txt")}
			default:
				t.Fatalf("unknown operation %s", scenario.Operation)
			}
			if err != nil {
				t.Fatal(err)
			}
			assertAgentSkillSpecialObservation(t, scenario, <-observations, file)
			if stream != nil {
				content, readErr := io.ReadAll(stream.Body)
				if closeErr := stream.Body.Close(); readErr != nil || closeErr != nil {
					t.Fatalf("read=%v close=%v", readErr, closeErr)
				}
				if stream.StatusCode != scenario.Status || !bytes.Equal(content, response) {
					t.Fatalf("stream status=%d bytes=%v", stream.StatusCode, content)
				}
				for key, value := range scenario.ResponseHeaders {
					if stream.Header.Get(key) != value {
						t.Fatalf("stream header %s=%q want=%q", key, stream.Header.Get(key), value)
					}
				}
			}
			var cliErr error
			output := captureStdout(t, func() { cliErr = ExecuteAgentCommand(ctx, args) })
			if cliErr != nil {
				t.Fatal(cliErr)
			}
			assertAgentSkillSpecialObservation(t, scenario, <-observations, file)
			switch scenario.Operation {
			case "inspect_skill_package", "import_skill_package":
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				expected := agentResourceJSON(envelope.Data)
				if result == nil || !reflect.DeepEqual(result.Value, expected) || !reflect.DeepEqual(agentResourceJSON([]byte(output)), expected) {
					t.Fatalf("Go=%#v CLI=%s want=%#v", result, output, expected)
				}
			case "polish_skill_stream":
				content, err := os.ReadFile(args[len(args)-1])
				if err != nil || !bytes.Equal(content, response) {
					t.Fatalf("CLI stream=%q err=%v", content, err)
				}
			}
		})
	}
}
