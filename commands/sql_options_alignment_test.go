package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestSQLCLIOptionsMatchSharedSDKContract(t *testing.T) {
	var cases []struct {
		Name, Command, Path string
		Input               map[string]json.RawMessage
		Request, Response   json.RawMessage
		Envelope            json.RawMessage
		Status              int
		ErrorCode           string `json:"error_code"`
	}
	data, err := os.ReadFile("../testdata/product-contracts/sql_options_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var want any
			decoder := json.NewDecoder(bytes.NewReader(scenario.Request))
			decoder.UseNumber()
			if err := decoder.Decode(&want); err != nil {
				t.Fatal(err)
			}
			received := make(chan any, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/newmoi"+scenario.Path || r.URL.RawQuery != "" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.RequestURI())
				}
				var got any
				decoder := json.NewDecoder(r.Body)
				decoder.UseNumber()
				if err := decoder.Decode(&got); err != nil {
					t.Error(err)
				}
				received <- got
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Envelope)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{scenario.Command}
			flags := map[string]string{"user": "users", "query_type": "query-types", "duration": "min-duration", "selected_field": "selected-fields", "sql_source_type": "sql-source-types"}
			for field, value := range scenario.Input {
				name := strings.ReplaceAll(field, "_", "-")
				if mapped, ok := flags[field]; ok {
					name = mapped
				}
				if field == "db_name" && (scenario.Command == "execute" || scenario.Command == "run") {
					name = "db"
				}
				var text string
				switch value[0] {
				case '"':
					if err := json.Unmarshal(value, &text); err != nil {
						t.Fatal(err)
					}
				case '[':
					var values []string
					if err := json.Unmarshal(value, &values); err != nil {
						t.Fatal(err)
					}
					text = strings.Join(values, ",")
				default:
					text = string(value)
				}
				args = append(args, "--"+name+"="+text)
			}
			var callErr error
			output := captureStdout(t, func() { callErr = ExecuteSQLCommand(NewContext(client, "workspace-1", false, true, true), args) })
			select {
			case got := <-received:
				if !reflect.DeepEqual(got, want) {
					t.Errorf("request=%v, want=%v", got, want)
				}
			default:
				t.Fatal("no HTTP request received")
			}
			if scenario.ErrorCode != "" {
				var failure *sdk.Error
				if !errors.As(callErr, &failure) {
					t.Fatalf("expected owner rejection, got %v", callErr)
				}
				var envelope struct{ Msg, Message string }
				if err := json.Unmarshal(scenario.Envelope, &envelope); err != nil {
					t.Fatal(err)
				}
				message := envelope.Msg
				if message == "" {
					message = envelope.Message
				}
				if failure.Code != scenario.ErrorCode || failure.Status != scenario.Status || failure.Message != message {
					t.Fatalf("owner failure=%+v", failure)
				}
				if output != "" {
					t.Fatalf("owner rejection printed success: %s", output)
				}
				return
			}
			if callErr != nil {
				t.Fatal(callErr)
			}
			// Compare the generated SDK response through CLI's public JSON encoding,
			// including its existing timestamp representation.
			if err := sqlCLIContractResponse(scenario.Command, output, scenario.Response); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func sqlCLIContractResponse(command, output string, expected []byte) error {
	var got proto.Message
	switch command {
	case "workbooks":
		got = &sdk.SQLWorkbookListResult{}
	case "workbook-versions":
		got = &sdk.SQLWorkbookVersionListResult{}
	case "query-history-overview":
		got = &sdk.SQLQueryHistoryOverviewResult{}
	case "query-history":
		got = &sdk.SQLQueryHistoryResult{}
	case "query-detail":
		got = &sdk.SQLStatementRecord{}
	case "query-profile":
		got = &sdk.SQLQueryProfileResult{}
	case "execute-query", "execute":
		got = &sdk.SQLQueryStartResult{}
	case "query-result", "run":
		got = &sdk.SQLQueryResult{}
	}
	if err := protojson.Unmarshal(expected, got); err != nil {
		return err
	}
	encoded, err := json.Marshal(got)
	if err != nil {
		return err
	}
	var actual, want any
	decoder := json.NewDecoder(strings.NewReader(output))
	decoder.UseNumber()
	if err := decoder.Decode(&actual); err != nil {
		return err
	}
	decoder = json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&want); err != nil {
		return err
	}
	if !reflect.DeepEqual(actual, want) {
		return fmt.Errorf("response=%v, want=%v", actual, want)
	}
	return nil
}
