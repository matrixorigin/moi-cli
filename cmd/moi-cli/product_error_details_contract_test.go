package main

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

	"github.com/matrixorigin/matrixflow/sdk/cli/commands"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestCLIProductErrorsMatchSharedDetailsContract(t *testing.T) {
	var cases []struct {
		Name, Operation string
		ExpectedCode    string `json:"expected_code"`
		Status          int
		Input, Response json.RawMessage
	}
	data, err := os.ReadFile("../../testdata/product-contracts/product_error_details_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel()
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
			received := make(chan []byte, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method, path := http.MethodPatch, "/newmoi/data-dashboards/dashboard-1"
				if scenario.Operation == "download-chart-result" {
					method, path = http.MethodGet, path+"/charts/chart-1/result/download"
				}
				if r.Method != method || r.URL.RequestURI() != path || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				received <- body
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", "transport-request")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{scenario.Operation, "--dashboard-id", "dashboard-1"}
			if scenario.Operation == "download-chart-result" {
				args = append(args, "--chart-id", "chart-1")
			} else {
				args = append(args, "--json", string(scenario.Input))
			}
			callErr := executeCommand(commands.NewContext(client, "workspace-1", false, true, true), "dashboard", args)
			select {
			case body := <-received:
				if !reflect.DeepEqual(decode(body), decode(scenario.Input)) {
					t.Errorf("request=%s, want=%s", body, scenario.Input)
				}
			default:
				t.Fatal("no HTTP request received")
			}
			var failure *sdk.Error
			if !errors.As(callErr, &failure) {
				t.Fatalf("unstructured command failure: %v", callErr)
			}
			var output bytes.Buffer
			writeCommandError(&output, callErr, true)
			var result struct {
				Code, Message string
				RequestID     string `json:"request_id"`
				Status        int
				Details, Data json.RawMessage
			}
			if err := json.Unmarshal(output.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Code != scenario.ExpectedCode || result.Status != scenario.Status || result.RequestID != "transport-request" {
				t.Fatalf("CLI lost error contract: %s; want code=%s status=%d", output.Bytes(), scenario.ExpectedCode, scenario.Status)
			}
			var response struct {
				Msg, Message  string
				Details, Data json.RawMessage
			}
			if err := json.Unmarshal(scenario.Response, &response); err != nil {
				t.Fatal(err)
			}
			message := response.Message
			if response.Msg != "" {
				message = response.Msg
			}
			if scenario.ExpectedCode == "ErrResponseInvalid" {
				if result.Message == "" || result.Message == message || len(failure.Details) != 0 {
					t.Fatalf("invalid response exposed owner payload: %s", output.Bytes())
				}
				return
			}
			if result.Message != message || !reflect.DeepEqual(decode(result.Data), decode(response.Data)) {
				t.Errorf("CLI changed owner message or repair data: %s", output.Bytes())
			}
			if decode(response.Details) == nil {
				if len(failure.Details) != 0 {
					t.Errorf("unexpected details: %s", output.Bytes())
				}
				return
			}
			if !reflect.DeepEqual(decode(result.Details), decode(response.Details)) {
				t.Errorf("CLI changed structured details: %s, want=%s", result.Details, response.Details)
			}
			output.Reset()
			writeCommandError(&output, callErr, false)
			lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
			if !bytes.Contains(output.Bytes(), []byte(message)) || !reflect.DeepEqual(decode(lines[len(lines)-1]), decode(response.Details)) {
				t.Errorf("text output lost owner diagnostics: %s", output.Bytes())
			}
		})
	}
}
