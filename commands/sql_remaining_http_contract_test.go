package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

// sqlRemainingHTTPContract is one SQL/Workbook JSON route shared by the Go SDK,
// the Python SDK, and the real moi-cli commands. Every case exercises one route
// through typed entries only; no case falls back to a raw request.
type sqlRemainingHTTPContract struct {
	Name       string                     `json:"name"`
	API        string                     `json:"api"`
	Command    string                     `json:"command"`
	CLIGroup   string                     `json:"cli_group"`
	CLICommand string                     `json:"cli_command"`
	CLIArgs    []string                   `json:"cli_args"`
	Method     string                     `json:"method"`
	Path       string                     `json:"path"`
	Query      string                     `json:"query"`
	Request    json.RawMessage            `json:"request"`
	SDKRequest json.RawMessage            `json:"sdk_request"`
	Response   json.RawMessage            `json:"response"`
	Envelope   json.RawMessage            `json:"envelope"`
	Status     int                        `json:"status"`
	ErrorCode  string                     `json:"error_code"`
	Input      map[string]json.RawMessage `json:"input"`
	Bodyless   bool                       `json:"bodyless"`
}

type sqlRemainingObservation struct {
	method, path, rawQuery, workspace, apiKey, contentType string
	body                                                   []byte
}

func sqlRemainingValue[T any](t *testing.T, input map[string]json.RawMessage, key string) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(input[key], &value); err != nil {
		t.Fatalf("%s: %v", key, err)
	}
	return value
}

func sqlRemainingJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

// sqlRemainingBodyMatches compares the wire body to the fixture request. Routes
// without a request DTO are sent without a body by the Go SDK, so an empty body
// is accepted there and still rejects a body carrying unrelated fields.
func sqlRemainingBodyMatches(scenario sqlRemainingHTTPContract, body []byte) bool {
	if len(bytes.TrimSpace(body)) == 0 {
		return scenario.Bodyless
	}
	got := sqlRemainingJSON(body)
	want := sqlRemainingJSON(scenario.Request)
	if scenario.Bodyless {
		object, ok := got.(map[string]any)
		return ok && len(object) == 0
	}
	return reflect.DeepEqual(got, want)
}

// sqlRemainingCall invokes the typed Go SDK entry for one fixture case. A nil
// message marks a route whose success response carries no product payload.
func sqlRemainingCall(t *testing.T, client *sdk.Client, scenario sqlRemainingHTTPContract) (proto.Message, error) {
	t.Helper()
	ctx := t.Context()
	sql := client.SQL("workspace-1")
	catalog := client.Catalog("workspace-1")
	i := scenario.Input
	switch scenario.Command {
	case "drop-meta-database":
		return nil, sql.DropMetaDatabase(ctx, sqlRemainingValue[string](t, i, "db_name"))
	case "meta-databases":
		return sql.MetaDatabases(ctx)
	case "meta-db-tables":
		return sql.MetaDBTables(ctx, sqlRemainingValue[string](t, i, "db_name"))
	case "meta-db-tree":
		return sql.MetaDBTree(ctx)
	case "meta-table-info":
		return sql.MetaTableInfo(ctx, sqlRemainingValue[string](t, i, "db_name"), sqlRemainingValue[string](t, i, "table_name"))
	case "meta-table-columns":
		return sql.MetaTableColumns(ctx, sqlRemainingValue[string](t, i, "db_name"), sqlRemainingValue[string](t, i, "table_name"))
	case "meta-table-column-stats":
		return sql.MetaTableColumnStats(ctx, sqlRemainingValue[string](t, i, "db_name"), sqlRemainingValue[string](t, i, "table_name"))
	case "meta-table-ddl":
		return sql.MetaTableDDL(ctx, sqlRemainingValue[string](t, i, "db_name"), sqlRemainingValue[string](t, i, "table_name"))
	case "drop-meta-table":
		return nil, sql.DropMetaTable(ctx, sqlRemainingValue[string](t, i, "db_name"), sqlRemainingValue[string](t, i, "table_name"))
	case "drop-meta-view":
		return nil, sql.DropMetaView(ctx, sqlRemainingValue[string](t, i, "db_name"), sqlRemainingValue[string](t, i, "table_name"))
	case "describe-query":
		return sql.DescribeQuery(ctx, sqlRemainingValue[string](t, i, "query_id"))
	case "kill-query":
		return nil, sql.KillQuery(ctx, sqlRemainingValue[string](t, i, "query_id"))
	case "workspace-table-columns":
		request := &sdk.WorkspaceTableColumnsRequest{}
		if err := json.Unmarshal(scenario.SDKRequest, request); err != nil {
			t.Fatal(err)
		}
		return catalog.WorkspaceTableColumns(ctx, request)
	case "workspace-database-tables":
		return catalog.WorkspaceDatabaseTables(ctx, sqlRemainingValue[int64](t, i, "database_id"))
	case "workspace-table-overview":
		return catalog.WorkspaceTableOverview(ctx)
	case "create-workbook":
		return sql.CreateWorkbook(ctx, sqlRemainingValue[string](t, i, "name"))
	case "delete-workbook":
		return nil, sql.DeleteWorkbook(ctx, sqlRemainingValue[string](t, i, "workbook_id"))
	case "workbook-detail":
		return sql.WorkbookDetail(ctx, sqlRemainingValue[string](t, i, "workbook_id"), sqlRemainingValue[string](t, i, "version_id"))
	case "update-workbook":
		return nil, sql.UpdateWorkbook(ctx, sqlRemainingValue[string](t, i, "workbook_id"), sqlRemainingValue[string](t, i, "name"))
	case "create-workbook-version":
		return sql.CreateWorkbookVersion(ctx, sqlRemainingValue[string](t, i, "workbook_id"), sqlRemainingValue[string](t, i, "sql_content"))
	case "save-workbook-version":
		content := sqlRemainingValue[string](t, i, "sql_content")
		return nil, sql.SaveWorkbookVersion(ctx, sqlRemainingValue[string](t, i, "workbook_id"), sqlRemainingValue[string](t, i, "version_id"), &content)
	case "update-workbook-version":
		return nil, sql.UpdateWorkbookVersion(ctx, sqlRemainingValue[string](t, i, "workbook_id"), sqlRemainingValue[string](t, i, "version_id"), sqlRemainingValue[string](t, i, "sql_content"))
	default:
		t.Fatalf("unknown fixture command %s", scenario.Command)
		return nil, nil
	}
}

// sqlRemainingRunCLI runs the real moi-cli command handler for one case.
func sqlRemainingRunCLI(ctx *Context, scenario sqlRemainingHTTPContract) error {
	switch scenario.CLIGroup {
	case "sql":
		return ExecuteSQLCommand(ctx, scenario.CLIArgs)
	case "catalog":
		args := scenario.CLIArgs[1:]
		switch scenario.CLIArgs[0] {
		case "workspace-table-columns":
			return runCatalogServiceWorkspaceTableColumns(ctx, args)
		case "workspace-database-tables":
			return runCatalogServiceWorkspaceDatabaseTables(ctx, args)
		case "workspace-table-overview":
			return runCatalogServiceWorkspaceTableOverview(ctx, args)
		}
	}
	return errors.New("unknown cli command " + strings.Join(scenario.CLIArgs, " "))
}

// TestSQLRemainingSharedHTTPContracts runs each fixture case through the typed
// Go SDK entry and the real CLI command, then checks method, escaped path,
// query, body, response payload and owner error against the same expectation.
func TestSQLRemainingSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/sql_remaining_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []sqlRemainingHTTPContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 22 {
		t.Fatalf("expected 22 routes, got %d", len(cases))
	}
	for _, scenario := range cases {
		t.Run(scenario.API, func(t *testing.T) {
			observations := make(chan sqlRemainingObservation, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				observations <- sqlRemainingObservation{
					method: r.Method, path: r.URL.EscapedPath(), rawQuery: r.URL.RawQuery,
					workspace: r.Header.Get("X-Workspace-ID"), apiKey: r.Header.Get("X-API-Key"),
					contentType: r.Header.Get("Content-Type"), body: body,
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Envelope)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			result, sdkErr := sqlRemainingCall(t, client, scenario)
			sqlRemainingCheckObservation(t, "SDK", <-observations, scenario)
			if scenario.ErrorCode != "" {
				sqlRemainingCheckOwnerError(t, "SDK", sdkErr, scenario)
			} else {
				if sdkErr != nil {
					t.Fatalf("SDK error: %v", sdkErr)
				}
				sqlRemainingCheckResponse(t, "SDK", result, scenario)
			}

			ctx := NewContext(client, "workspace-1", false, true, true)
			var callErr error
			output := captureStdout(t, func() { callErr = sqlRemainingRunCLI(ctx, scenario) })
			sqlRemainingCheckObservation(t, "CLI", <-observations, scenario)
			if scenario.ErrorCode != "" {
				sqlRemainingCheckOwnerError(t, "CLI", callErr, scenario)
				if output != "" {
					t.Fatalf("CLI printed success on owner rejection: %s", output)
				}
				return
			}
			if callErr != nil {
				t.Fatalf("CLI error: %v", callErr)
			}
			if result == nil {
				want := "sql " + scenario.Command + ": ok"
				if strings.TrimSpace(output) != want {
					t.Fatalf("CLI output=%q want=%q", output, want)
				}
				return
			}
			if err := sqlRemainingCLIResponse(result, output, scenario.Response); err != nil {
				t.Fatalf("CLI response: %v", err)
			}
		})
	}
}

func sqlRemainingCheckObservation(t *testing.T, side string, observation sqlRemainingObservation, scenario sqlRemainingHTTPContract) {
	t.Helper()
	if observation.method != scenario.Method || observation.path != "/newmoi"+scenario.Path || observation.rawQuery != scenario.Query || observation.workspace != "workspace-1" || observation.apiKey != "synthetic-key" {
		t.Fatalf("%s request=%s %s?%s workspace=%s", side, observation.method, observation.path, observation.rawQuery, observation.workspace)
	}
	if !sqlRemainingBodyMatches(scenario, observation.body) {
		t.Fatalf("%s body=%s want=%s", side, observation.body, scenario.Request)
	}
	if !scenario.Bodyless && observation.contentType != "application/json" {
		t.Fatalf("%s content type=%q", side, observation.contentType)
	}
}

func sqlRemainingCheckOwnerError(t *testing.T, side string, err error, scenario sqlRemainingHTTPContract) {
	t.Helper()
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("%s expected owner rejection, got %v", side, err)
	}
	var envelope struct{ Msg, Message string }
	if decodeErr := json.Unmarshal(scenario.Envelope, &envelope); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	message := envelope.Msg
	if message == "" {
		message = envelope.Message
	}
	if failure.Code != scenario.ErrorCode || failure.Status != scenario.Status || failure.Message != message {
		t.Fatalf("%s owner failure=%+v", side, failure)
	}
}

func sqlRemainingCheckResponse(t *testing.T, side string, result proto.Message, scenario sqlRemainingHTTPContract) {
	t.Helper()
	if result == nil {
		if string(scenario.Response) != "null" {
			t.Fatalf("%s missing typed result for %s", side, scenario.Response)
		}
		return
	}
	expected := result.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal(scenario.Response, expected); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(result, expected) {
		t.Fatalf("%s response=%v want=%v", side, result, expected)
	}
}

// sqlRemainingCLIResponse compares the CLI JSON output with the same expected
// product message the SDK decoded, so both surfaces share one expectation.
func sqlRemainingCLIResponse(result proto.Message, output string, expected []byte) error {
	message := result.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal(expected, message); err != nil {
		return err
	}
	encoded, err := json.Marshal(message)
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
		return errors.New("response mismatch: " + output)
	}
	return nil
}
