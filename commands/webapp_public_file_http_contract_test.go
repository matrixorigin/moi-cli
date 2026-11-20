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
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type webappPublicFileContract struct {
	Name, Operation, Method, Path string
	Query                         map[string][]string `json:"query"`
	Input                         struct {
		Filename      string `json:"filename"`
		ContentBase64 string `json:"content_base64"`
		ContentType   string `json:"content_type"`
	} `json:"input"`
	Response            json.RawMessage   `json:"response"`
	ResponseBase64      string            `json:"response_base64"`
	ResponseContentType string            `json:"response_content_type"`
	ResponseHeaders     map[string]string `json:"response_headers"`
	Status              int               `json:"status"`
}

func TestWebappPublicFileSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/webapp_public_file_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []webappPublicFileContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			response := scenario.Response
			if scenario.ResponseBase64 != "" {
				response, err = base64.StdEncoding.DecodeString(scenario.ResponseBase64)
				if err != nil {
					t.Fatal(err)
				}
			}
			observations := make(chan webappPublicObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				observations <- webappPublicObservation{
					method: r.Method, path: r.URL.EscapedPath(), query: r.URL.Query(), body: raw,
					contentType: r.Header.Get("Content-Type"), accept: r.Header.Get("Accept"),
					workspace: r.Header.Get("X-Workspace-ID"), role: r.Header.Get("X-MOI-Role-ID"), apiKey: r.Header.Get("X-API-Key"),
					authorization: r.Header.Get("Authorization"), cookie: r.Header.Get("Cookie"), runtime: r.Header.Get("X-MOI-Runtime-Authorization"),
					probe: r.Header.Get("X-Webapp-Probe"),
				}
				w.Header().Set("Content-Type", scenario.ResponseContentType)
				if scenario.ResponseContentType == "" {
					w.Header().Set("Content-Type", "application/json")
				}
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("private-key"), sdk.WithToken("private-token"), sdk.WithSessionCookie("private-cookie"), sdk.WithRuntimeAuthorization("private-runtime-grant"), sdk.WithWorkspaceCurrentRoleID("workspace-1", "7"), sdk.WithCustomHeader("X-Webapp-Probe", "preserved"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			var content []byte
			if scenario.Operation == "upload-file" {
				content, err = base64.StdEncoding.DecodeString(scenario.Input.ContentBase64)
				if err != nil {
					t.Fatal(err)
				}
			}
			result, stream, args := invokeWebappPublicFileSDK(t, client, scenario, content)
			assertWebappPublicFileObservation(t, scenario, <-observations, content)
			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteWebappCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if cliErr != nil {
				t.Fatal(cliErr)
			}
			assertWebappPublicFileObservation(t, scenario, <-observations, content)
			if scenario.Operation == "upload-file" {
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				expected := agentResourceJSON(envelope.Data)
				if result == nil || !reflect.DeepEqual(result.Value, expected) || !reflect.DeepEqual(agentResourceJSON([]byte(output)), expected) {
					t.Fatalf("result=%#v output=%s want=%#v", result, output, expected)
				}
			} else if !bytes.Equal(stream, response) || !bytes.Equal([]byte(output), response) {
				t.Fatalf("stream=%q output=%q want=%q", stream, output, response)
			}
		})
	}
}

func invokeWebappPublicFileSDK(t *testing.T, client *sdk.Client, scenario webappPublicFileContract, content []byte) (*sdk.ProductData, []byte, []string) {
	t.Helper()
	args := []string{scenario.Operation, "--conversation-id", "conversation/1", "--task-id", "task/1", "--file-id", "file/1"}
	keys := make([]string, 0, len(scenario.Query))
	for key := range scenario.Query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	opts := make([]sdk.QueryOption, 0)
	for _, key := range keys {
		for _, value := range scenario.Query[key] {
			opts = append(opts, sdk.WithQueryValues(key, value))
			args = append(args, "--query-param", key+"="+value)
		}
	}
	service := client.Webapp()
	if scenario.Operation == "upload-file" {
		path := filepath.Join(t.TempDir(), scenario.Input.Filename)
		if err := os.WriteFile(path, content, 0o600); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--file", path, "--content-type", scenario.Input.ContentType)
		result, err := service.UploadFile(t.Context(), scenario.Input.Filename, scenario.Input.ContentType, bytes.NewReader(content), opts...)
		if err != nil {
			t.Fatal(err)
		}
		return result, nil, args
	}
	var stream *sdk.StreamResponse
	var err error
	if scenario.Operation == "download-task-file" {
		stream, err = service.DownloadTaskFile(t.Context(), "conversation/1", "task/1", "file/1", opts...)
	} else {
		stream, err = service.PreviewTaskFile(t.Context(), "conversation/1", "task/1", "file/1", opts...)
	}
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(stream.Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := stream.Body.Close(); err != nil {
		t.Fatal(err)
	}
	return nil, raw, args
}

func assertWebappPublicFileObservation(t *testing.T, scenario webappPublicFileContract, got webappPublicObservation, content []byte) {
	t.Helper()
	expectedQuery := url.Values{}
	for key, values := range scenario.Query {
		for _, value := range values {
			expectedQuery.Add(key, value)
		}
	}
	if got.method != scenario.Method || got.path != scenario.Path || !reflect.DeepEqual(got.query, expectedQuery) {
		t.Fatalf("request=%s %s query=%v", got.method, got.path, got.query)
	}
	if got.workspace != "" || got.role != "" || got.apiKey != "" || got.authorization != "" || got.cookie != "" || got.runtime != "" || got.probe != "preserved" {
		t.Fatalf("headers=%+v", got)
	}
	if scenario.Operation != "upload-file" {
		if got.accept != "*/*" || got.contentType != "" || len(got.body) != 0 {
			t.Fatalf("download headers/body=%+v", got)
		}
		return
	}
	if got.accept != "application/json" {
		t.Fatalf("Accept=%q", got.accept)
	}
	mediaType, params, err := mime.ParseMediaType(got.contentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("Content-Type=%q error=%v", got.contentType, err)
	}
	reader := multipart.NewReader(bytes.NewReader(got.body), params["boundary"])
	part, err := reader.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if part.FormName() != "file" || part.FileName() != scenario.Input.Filename || part.Header.Get("Content-Type") != scenario.Input.ContentType {
		t.Fatalf("part name=%q filename=%q type=%q headers=%v", part.FormName(), part.FileName(), part.Header.Get("Content-Type"), part.Header)
	}
	raw, err := io.ReadAll(part)
	if err != nil || !bytes.Equal(raw, content) {
		t.Fatalf("content=%q want=%q error=%v", raw, content, err)
	}
	if _, err := reader.NextPart(); err != io.EOF {
		t.Fatalf("extra multipart part: %v", err)
	}
}
