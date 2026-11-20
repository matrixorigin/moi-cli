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
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type aiStudioCLIContract struct {
	Name      string              `json:"name"`
	Operation string              `json:"operation"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Query     map[string][]string `json:"query"`
	Input     json.RawMessage     `json:"input"`
	Request   json.RawMessage     `json:"request"`
	Response  json.RawMessage     `json:"response"`
	Status    int                 `json:"status"`
	SlotID    string              `json:"slot_id"`
}

func aiStudioCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func executeAIStudioHTTPContract(ctx *Context, args []string) ([]byte, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteAIStudioCommand(ctx, args)
	os.Stdout = original
	closeErr := write.Close()
	output, readErr := io.ReadAll(read)
	readCloseErr := read.Close()
	return output, errors.Join(callErr, closeErr, readErr, readCloseErr)
}

func TestAIStudioCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/ai_studio_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []aiStudioCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var received any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, values := range scenario.Query {
					for _, value := range values {
						expectedQuery.Add(key, value)
					}
				}
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("Authorization") != "Bearer admin-token" || r.Header.Get("X-API-Key") != "" || r.Header.Get("X-Workspace-ID") != "" {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				if len(raw) > 0 {
					received = aiStudioCLIJSON(raw)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("admin-token"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{scenario.Operation}
			if scenario.Operation == "service-slot-logs" {
				args = append(args, "--slot-id", "", "--page-size", "0")
			}
			if scenario.Operation == "put-service-slot" {
				args = append(args, "--slot-id", scenario.SlotID, "--json", string(scenario.Input))
			}
			output, err := executeAIStudioHTTPContract(NewContext(client, "workspace-must-not-leak", false, true, true), args)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(received, aiStudioCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, aiStudioCLIJSON(scenario.Request))
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			if got := aiStudioCLIJSON(output); !reflect.DeepEqual(got, aiStudioCLIJSON(envelope.Data)) {
				t.Fatalf("output=%#v want=%#v", got, aiStudioCLIJSON(envelope.Data))
			}
		})
	}
}
