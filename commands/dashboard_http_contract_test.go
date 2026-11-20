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
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestDashboardCLIMatchesSharedHTTPContract(t *testing.T) {
	var cases []struct {
		Name, Operation, Method, Path, CSV string
		DashboardID                        string `json:"dashboard_id"`
		ChartID                            string `json:"chart_id"`
		JSONFile                           bool   `json:"json_file"`
		DownloadFile                       bool   `json:"download_file"`
		Input, Request, Response           json.RawMessage
		Query                              map[string]string
		Status                             int
	}
	data, err := os.ReadFile("../testdata/product-contracts/dashboard_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			decodeJSON := func(data []byte) any {
				t.Helper()
				var value any
				decoder := json.NewDecoder(bytes.NewReader(data))
				decoder.UseNumber()
				if err := decoder.Decode(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			query := url.Values{}
			for key, value := range scenario.Query {
				query.Set(key, value)
			}
			received := make(chan any, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.Path != scenario.Path || !reflect.DeepEqual(r.URL.Query(), query) || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected scoped request: %s %s", r.Method, r.URL.RequestURI())
				}
				raw, err := io.ReadAll(r.Body)
				var body any
				if err != nil {
					t.Error(err)
				} else if len(raw) > 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil {
						t.Error(err)
					}
				}
				received <- body
				if scenario.CSV != "" {
					w.Header().Set("Content-Type", "text/csv; charset=utf-8")
					w.WriteHeader(scenario.Status)
					_, _ = io.WriteString(w, scenario.CSV)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(scenario.Status)
					_, _ = w.Write(scenario.Response)
				}
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{scenario.Operation}
			if scenario.DashboardID != "" {
				args = append(args, "--dashboard-id", scenario.DashboardID)
			}
			if scenario.ChartID != "" {
				args = append(args, "--chart-id", scenario.ChartID)
			}
			for key, value := range scenario.Query {
				args = append(args, "--query-param", key+"="+value)
			}
			switch scenario.Operation {
			case "create", "list", "update", "create-chart", "update-chart", "update-chart-order", "drill-through", "download-drill-through":
				if scenario.JSONFile {
					path := filepath.Join(t.TempDir(), "request.json")
					if err := os.WriteFile(path, scenario.Input, 0600); err != nil {
						t.Fatal(err)
					}
					args = append(args, "--json-file", path)
				} else {
					args = append(args, "--json", string(scenario.Input))
				}
			case "preview-sql":
				var input map[string]string
				if err := json.Unmarshal(scenario.Input, &input); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--sql-text", input["sql_text"])
				if chart, exists := input["chart_id"]; exists {
					args = append(args, "--chart-id", chart)
				}
			case "generate-sql":
				var input map[string]string
				if err := json.Unmarshal(scenario.Input, &input); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--question", input["question"], "--title", input["title"])
			}
			var downloadPath string
			if scenario.DownloadFile {
				downloadPath = filepath.Join(t.TempDir(), "result.csv")
				args = append(args, "--output", downloadPath)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteDashboardCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			select {
			case body := <-received:
				if want := decodeJSON(scenario.Request); !reflect.DeepEqual(body, want) {
					t.Errorf("HTTP body=%v, want=%v", body, want)
				}
			default:
				t.Fatal("no HTTP request received")
			}
			var response struct {
				Code, Msg string
				Data      json.RawMessage
			}
			if err := json.Unmarshal(scenario.Response, &response); err != nil {
				t.Fatal(err)
			}
			if scenario.Status >= 400 {
				var failure *sdk.Error
				if !errors.As(callErr, &failure) || failure.Code != response.Code || failure.Message != response.Msg || failure.Status != scenario.Status || output != "" {
					t.Fatalf("owner error changed: %v, output=%s", callErr, output)
				}
				return
			}
			if callErr != nil {
				t.Fatal(callErr)
			}
			if scenario.CSV != "" {
				if downloadPath != "" {
					if output != "" {
						t.Errorf("download file also wrote stdout: %s", output)
					}
					body, err := os.ReadFile(downloadPath)
					if err != nil {
						t.Fatal(err)
					}
					output = string(body)
				}
				if output != scenario.CSV {
					t.Errorf("CSV=%q, want=%q", output, scenario.CSV)
				}
				return
			}
			var expected proto.Message
			switch scenario.Operation {
			case "task-execution-overview":
				expected = &sdk.TaskExecutionOverviewResult{}
			case "create", "get":
				expected = &sdk.DataDashboard{}
			case "list":
				expected = &sdk.DataDashboardListResult{}
			case "update", "delete", "delete-chart":
				expected = &sdk.DataDashboardMutationResult{}
			case "create-chart", "update-chart":
				expected = &sdk.DataDashboardChart{}
			case "update-chart-order":
				expected = &sdk.DataDashboardChartOrderRequest{}
			case "preview-sql":
				expected = &sdk.SQLQueryResult{}
			case "generate-sql":
				expected = &sdk.DataDashboardGeneratedSQLResult{}
			case "refresh":
				expected = &sdk.DataDashboardChartRuns{}
			case "run-chart":
				expected = &sdk.DataDashboardChartRun{}
			case "chart-result":
				expected = &sdk.DataDashboardChartExecution{}
			case "drill-through":
				expected = &sdk.DataDashboardDrillThroughResult{}
			}
			encoded := []byte(response.Data)
			if expected != nil {
				if err := protojson.Unmarshal(response.Data, expected); err != nil {
					t.Fatal(err)
				}
				encoded, err = json.Marshal(expected)
				if err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(decodeJSON([]byte(output)), decodeJSON(encoded)) {
				t.Errorf("CLI response=%s, want=%s", output, encoded)
			}
		})
	}
}
