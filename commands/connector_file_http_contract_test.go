package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type connectorFileCLIFile struct {
	FieldName     string `json:"field_name,omitempty"`
	Filename      string `json:"filename"`
	ContentType   string `json:"content_type,omitempty"`
	ContentBase64 string `json:"content_base64"`
}

type connectorFileCLIMultipart struct {
	Fields map[string]string      `json:"fields"`
	Files  []connectorFileCLIFile `json:"files"`
}

type connectorFileCLIObservation struct {
	Body      any
	Multipart *connectorFileCLIMultipart
}

type connectorFileCLIContract struct {
	Name      string              `json:"name"`
	Operation string              `json:"operation"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Query     map[string][]string `json:"query"`
	Input     struct {
		QueryParams map[string][]string    `json:"query_params"`
		ConnFileID  string                 `json:"conn_file_id"`
		Fields      map[string]string      `json:"fields"`
		Files       []connectorFileCLIFile `json:"files"`
	} `json:"input"`
	Request       json.RawMessage            `json:"request"`
	Multipart     *connectorFileCLIMultipart `json:"multipart"`
	Response      json.RawMessage            `json:"response"`
	ContentBase64 string                     `json:"content_base64"`
	Status        int                        `json:"status"`
}

func readConnectorFileMultipart(r *http.Request) (*connectorFileCLIMultipart, error) {
	result := &connectorFileCLIMultipart{Fields: map[string]string{}}
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		return nil, err
	}
	defer r.MultipartForm.RemoveAll()
	for name, values := range r.MultipartForm.Value {
		if len(values) != 1 {
			return nil, fmt.Errorf("multipart field %s has %d values", name, len(values))
		}
		result.Fields[name] = values[0]
	}
	for fieldName, headers := range r.MultipartForm.File {
		for _, header := range headers {
			file, err := header.Open()
			if err != nil {
				return nil, err
			}
			content, readErr := io.ReadAll(file)
			closeErr := file.Close()
			if readErr != nil {
				return nil, readErr
			}
			if closeErr != nil {
				return nil, closeErr
			}
			result.Files = append(result.Files, connectorFileCLIFile{
				FieldName: fieldName, Filename: header.Filename, ContentType: header.Header.Get("Content-Type"),
				ContentBase64: base64.StdEncoding.EncodeToString(content),
			})
		}
	}
	return result, nil
}

func TestConnectorFileSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/connector_file_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []connectorFileCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			observations := make(chan connectorFileCLIObservation, 2)
			content, err := base64.StdEncoding.DecodeString(scenario.ContentBase64)
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
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
				}
				if scenario.Operation == "file-stream" && r.Header.Get("Accept") != "*/*" {
					t.Errorf("Accept=%q", r.Header.Get("Accept"))
				}
				observation := connectorFileCLIObservation{}
				if scenario.Multipart != nil {
					multipart, readErr := readConnectorFileMultipart(r)
					if readErr != nil {
						t.Errorf("read multipart: %v", readErr)
					}
					observation.Multipart = multipart
				} else if len(scenario.Request) != 0 && string(scenario.Request) != "null" {
					decoder := json.NewDecoder(r.Body)
					decoder.UseNumber()
					if decodeErr := decoder.Decode(&observation.Body); decodeErr != nil {
						t.Errorf("decode request: %v", decodeErr)
					}
				}
				observations <- observation
				if scenario.Operation == "file-stream" {
					w.Header().Set("Content-Type", "application/octet-stream")
					w.WriteHeader(scenario.Status)
					_, _ = w.Write(content)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			service := client.Connectors("workspace-1")
			args := []string{scenario.Operation}
			if scenario.Operation == "file-stream" {
				args = append(args, "--conn-file-id", scenario.Input.ConnFileID)
				stream, streamErr := service.FileStream(t.Context(), scenario.Input.ConnFileID)
				if streamErr != nil {
					t.Fatal(streamErr)
				}
				goContent, readErr := io.ReadAll(stream.Body)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if closeErr := stream.Body.Close(); closeErr != nil {
					t.Fatal(closeErr)
				}
				assertConnectorFileObservation(t, <-observations, scenario)
				if !bytes.Equal(goContent, content) {
					t.Fatalf("Go content=%v, want=%v", goContent, content)
				}
				output := filepath.Join(t.TempDir(), "connector-file.bin")
				args = append(args, "--output", output)
				if cliErr := ExecuteConnectorCommand(NewContext(client, "workspace-1", false, false, true), args); cliErr != nil {
					t.Fatal(cliErr)
				}
				cliContent, readErr := os.ReadFile(output)
				if readErr != nil {
					t.Fatal(readErr)
				}
				assertConnectorFileObservation(t, <-observations, scenario)
				if !bytes.Equal(cliContent, content) {
					t.Fatalf("CLI content=%v, want=%v", cliContent, content)
				}
				return
			}

			var goResult any
			switch scenario.Operation {
			case "files":
				var options []sdk.QueryOption
				keys := make([]string, 0, len(scenario.Input.QueryParams))
				for key := range scenario.Input.QueryParams {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					values := scenario.Input.QueryParams[key]
					options = append(options, sdk.WithQueryValues(key, values...))
					for _, value := range values {
						args = append(args, "--query-param", key+"="+value)
					}
				}
				goResult, err = service.Files(t.Context(), options...)
			case "upload", "local-upload":
				paths := make([]string, 0, len(scenario.Input.Files))
				for _, file := range scenario.Input.Files {
					fileContent, decodeErr := base64.StdEncoding.DecodeString(file.ContentBase64)
					if decodeErr != nil {
						t.Fatal(decodeErr)
					}
					paths = append(paths, uploadFixture(t, file.Filename, fileContent))
				}
				if scenario.Operation == "upload" {
					var options []sdk.ConnectorUploadOption
					for index, file := range scenario.Input.Files {
						content, _ := base64.StdEncoding.DecodeString(file.ContentBase64)
						options = append(options, sdk.WithConnectorUploadFile("file", file.Filename, bytes.NewReader(content)))
						args = append(args, "--file-path", paths[index])
					}
					keys := make([]string, 0, len(scenario.Input.Fields))
					for key := range scenario.Input.Fields {
						keys = append(keys, key)
					}
					sort.Strings(keys)
					for _, key := range keys {
						value := scenario.Input.Fields[key]
						options = append(options, sdk.WithConnectorUploadField(key, value))
						flagName := map[string]string{"name": "name", "VolumeID": "volume-id", "meta": "meta", "file_types": "file-types", "path_regex": "path-regex", "unzip_keep_structure": "unzip-keep-structure", "dedup": "dedup", "table_config": "table-config"}[key]
						args = append(args, "--"+flagName+"="+value)
					}
					goResult, err = service.Upload(t.Context(), options...)
				} else {
					files := make([]sdk.MultipartFile, 0, len(scenario.Input.Files))
					for index, file := range scenario.Input.Files {
						content, _ := base64.StdEncoding.DecodeString(file.ContentBase64)
						files = append(files, sdk.MultipartFile{FieldName: "file", Filename: file.Filename, Reader: bytes.NewReader(content)})
						args = append(args, "--file-path", paths[index])
					}
					goResult, err = service.LocalUpload(t.Context(), files...)
				}
			case "file-download":
				args = append(args, "--conn-file-id", scenario.Input.ConnFileID)
				goResult, err = service.FileDownload(t.Context(), scenario.Input.ConnFileID)
			case "file-delete":
				args = append(args, "--conn-file-id", scenario.Input.ConnFileID)
				err = service.FileDelete(t.Context(), scenario.Input.ConnFileID)
			default:
				t.Fatalf("unknown operation %s", scenario.Operation)
			}
			if err != nil {
				t.Fatal(err)
			}
			assertConnectorFileObservation(t, <-observations, scenario)
			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteConnectorCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if cliErr != nil {
				t.Fatal(cliErr)
			}
			assertConnectorFileObservation(t, <-observations, scenario)
			if scenario.Operation == "file-delete" {
				if output != "connector file-delete: ok\n" {
					t.Fatalf("CLI output=%q", output)
				}
				return
			}
			encoded, err := json.Marshal(goResult)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(catalogFolderCLIJSON(json.RawMessage(output)), catalogFolderCLIJSON(encoded)) {
				t.Fatalf("CLI output=%s, Go result=%s", output, encoded)
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			expected := connectorFileCLIResult(scenario.Operation)
			if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
				t.Fatal(err)
			}
			actual := expected.ProtoReflect().New().Interface()
			if err := json.Unmarshal([]byte(output), actual); err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(actual, expected) {
				t.Fatalf("CLI result=%v, owner data=%v", actual, expected)
			}
		})
	}
}

func assertConnectorFileObservation(t *testing.T, got connectorFileCLIObservation, scenario connectorFileCLIContract) {
	t.Helper()
	if scenario.Multipart != nil {
		if !reflect.DeepEqual(got.Multipart, scenario.Multipart) {
			t.Fatalf("multipart=%#v, want=%#v", got.Multipart, scenario.Multipart)
		}
		return
	}
	if !reflect.DeepEqual(got.Body, catalogFolderCLIJSON(scenario.Request)) {
		t.Fatalf("body=%#v, want=%#v", got.Body, catalogFolderCLIJSON(scenario.Request))
	}
}

func connectorFileCLIResult(operation string) proto.Message {
	switch operation {
	case "files":
		return &sdk.ConnectorFileListResult{}
	case "upload":
		return &sdk.UploadResponse{}
	case "local-upload":
		return &sdk.LocalUploadResult{}
	case "file-download":
		return &sdk.FileDownloadResult{}
	default:
		return nil
	}
}
