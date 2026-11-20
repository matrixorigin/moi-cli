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
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type catalogUploadCLIFile struct {
	Filename      string `json:"filename"`
	ContentType   string `json:"content_type"`
	ContentBase64 string `json:"content_base64"`
}

type catalogUploadCLIMultipart struct {
	Fields map[string]string    `json:"fields"`
	File   catalogUploadCLIFile `json:"file"`
}

type catalogUploadCLIContract struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Input     struct {
		Filename          string  `json:"filename"`
		ContentBase64     string  `json:"content_base64"`
		ContentType       *string `json:"content_type"`
		ExternalSourceURI *string `json:"external_source_uri"`
	} `json:"input"`
	Multipart       catalogUploadCLIMultipart `json:"multipart"`
	Response        json.RawMessage           `json:"response"`
	ResponseHeaders map[string]string         `json:"response_headers"`
	Status          int                       `json:"status"`
}

func readCatalogUploadCLIRequest(r *http.Request) (catalogUploadCLIMultipart, error) {
	result := catalogUploadCLIMultipart{Fields: map[string]string{}}
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		return result, err
	}
	defer r.MultipartForm.RemoveAll()
	for name, values := range r.MultipartForm.Value {
		if len(values) != 1 {
			return result, fmt.Errorf("multipart field %s has %d values", name, len(values))
		}
		result.Fields[name] = values[0]
	}
	files := r.MultipartForm.File["file"]
	if len(files) != 1 {
		return result, fmt.Errorf("multipart file count=%d", len(files))
	}
	file, err := files[0].Open()
	if err != nil {
		return result, err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return result, err
	}
	result.File = catalogUploadCLIFile{
		Filename: files[0].Filename, ContentType: files[0].Header.Get("Content-Type"),
		ContentBase64: base64.StdEncoding.EncodeToString(content),
	}
	return result, nil
}

func catalogUploadCLIError(t *testing.T, err error, scenario catalogUploadCLIContract) *sdk.Error {
	t.Helper()
	if scenario.Status < 400 {
		if err != nil {
			t.Fatal(err)
		}
		return nil
	}
	var productErr *sdk.Error
	if !errors.As(err, &productErr) {
		t.Fatalf("error=%T %v, want *sdk.Error", err, err)
	}
	var envelope struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if productErr.Code != envelope.Code || productErr.Message != envelope.Msg || productErr.Status != scenario.Status || productErr.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v, want code=%s message=%q status=%d request_id=%q", productErr, envelope.Code, envelope.Msg, scenario.Status, scenario.ResponseHeaders["X-Request-ID"])
	}
	return productErr
}

func catalogUploadCLIResult(t *testing.T, result *sdk.UploadFileResult, scenario catalogUploadCLIContract) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	expected := &sdk.UploadFileResult{}
	if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(result, expected) {
		t.Fatalf("result=%v, want=%v", result, expected)
	}
}

func TestCatalogUploadSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/catalog_upload_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []catalogUploadCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			observations := make(chan catalogUploadCLIMultipart, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
				}
				observation, err := readCatalogUploadCLIRequest(r)
				if err != nil {
					t.Errorf("read multipart: %v", err)
				}
				observations <- observation
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
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
			content, err := base64.StdEncoding.DecodeString(scenario.Input.ContentBase64)
			if err != nil {
				t.Fatal(err)
			}
			path := uploadFixture(t, "upload.bin", content)
			var options []sdk.CatalogUploadOption
			operation := scenario.Operation
			if operation == "" {
				operation = "upload"
			}
			args := []string{operation, "--file-path", path, "--filename", scenario.Input.Filename}
			if scenario.Input.ContentType != nil {
				options = append(options, sdk.WithCatalogUploadContentType(*scenario.Input.ContentType))
				args = append(args, "--content-type", *scenario.Input.ContentType)
			}
			if scenario.Input.ExternalSourceURI != nil {
				options = append(options, sdk.WithCatalogUploadExternalSourceURI(*scenario.Input.ExternalSourceURI))
				args = append(args, "--external-source-uri", *scenario.Input.ExternalSourceURI)
			}
			var goResult *sdk.UploadFileResult
			var goErr error
			if operation == "upload-private" {
				goResult, goErr = client.Catalog("workspace-1").UploadPrivate(t.Context(), scenario.Input.Filename, bytes.NewReader(content), options...)
			} else {
				goResult, goErr = client.Catalog("workspace-1").Upload(t.Context(), scenario.Input.Filename, bytes.NewReader(content), options...)
			}
			if got := <-observations; !reflect.DeepEqual(got, scenario.Multipart) {
				t.Fatalf("Go multipart=%+v, want=%+v", got, scenario.Multipart)
			}
			goProductErr := catalogUploadCLIError(t, goErr, scenario)
			if goProductErr == nil {
				catalogUploadCLIResult(t, goResult, scenario)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteCatalogCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if got := <-observations; !reflect.DeepEqual(got, scenario.Multipart) {
				t.Fatalf("CLI multipart=%+v, want=%+v", got, scenario.Multipart)
			}
			cliProductErr := catalogUploadCLIError(t, cliErr, scenario)
			if goProductErr == nil {
				var cliResult sdk.UploadFileResult
				if err := json.Unmarshal([]byte(output), &cliResult); err != nil {
					t.Fatalf("CLI output=%q: %v", output, err)
				}
				catalogUploadCLIResult(t, &cliResult, scenario)
			} else if cliProductErr.Code != goProductErr.Code || cliProductErr.Message != goProductErr.Message || cliProductErr.Status != goProductErr.Status || cliProductErr.RequestID != goProductErr.RequestID {
				t.Fatalf("Go error=%+v, CLI error=%+v", goProductErr, cliProductErr)
			}
		})
	}
}
