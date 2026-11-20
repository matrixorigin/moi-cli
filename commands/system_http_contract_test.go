package commands

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type systemHTTPContract struct {
	Name, Operation, Method, Path string
	Response                      json.RawMessage   `json:"response"`
	ResponseHeaders               map[string]string `json:"response_headers"`
	Status                        int               `json:"status"`
}

func TestSystemSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/system_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []systemHTTPContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			requests := make(chan struct{}, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				if len(raw) != 0 {
					t.Errorf("body=%q", raw)
				}
				requests <- struct{}{}
				w.Header().Set("Content-Type", "application/json")
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			var got proto.Message
			if scenario.Operation == "versions" {
				got, err = client.System().Versions(t.Context())
			} else {
				got, err = client.System().WebappConfig(t.Context())
			}
			<-requests
			assertSystemHTTPError(t, err, scenario)
			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteSystemCommand(NewContext(client, "workspace-1", false, true, true), []string{scenario.Operation})
			})
			<-requests
			assertSystemHTTPError(t, cliErr, scenario)
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}
			expected := systemHTTPExpected(t, scenario)
			if !proto.Equal(got, expected) {
				t.Fatalf("Go result=%v want=%v", got, expected)
			}
			cliGot := systemHTTPExpected(t, scenario)
			if err := protojson.Unmarshal([]byte(output), cliGot); err != nil {
				t.Fatalf("CLI output=%q error=%v", output, err)
			}
			if !proto.Equal(cliGot, expected) {
				t.Fatalf("CLI result=%v want=%v", cliGot, expected)
			}
		})
	}
}

func systemHTTPExpected(t *testing.T, scenario systemHTTPContract) proto.Message {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	var result proto.Message
	if scenario.Operation == "versions" {
		result = &sdk.SystemVersionsResult{}
	} else {
		result = &sdk.WebappConfigResult{}
	}
	if err := protojson.Unmarshal(envelope.Data, result); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertSystemHTTPError(t *testing.T, err error, scenario systemHTTPContract) {
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
		Code    string `json:"code"`
		Message string `json:"msg"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if failure.Code != envelope.Code || failure.Message != envelope.Message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v", failure)
	}
}
