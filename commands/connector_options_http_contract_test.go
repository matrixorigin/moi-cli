package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type connectorCLIContract struct {
	Name      string              `json:"name"`
	Operation string              `json:"operation"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Query     map[string][]string `json:"query"`
	Input     struct {
		ConnectorID   string              `json:"connector_id"`
		URI           string              `json:"uri"`
		ConnFileID    *string             `json:"conn_file_id"`
		SheetName     *string             `json:"sheet_name"`
		IsColumnName  *bool               `json:"is_column_name"`
		RowStart      *int                `json:"row_start"`
		ColumnNameRow *int                `json:"column_name_row"`
		FileType      *int                `json:"file_type"`
		CSV           json.RawMessage     `json:"csv"`
		SourceType    json.RawMessage     `json:"source_type"`
		Database      string              `json:"database"`
		Table         string              `json:"table"`
		Schema        string              `json:"schema"`
		QueryParams   map[string]string   `json:"query_params"`
		Name          *string             `json:"name"`
		UsageType     *int                `json:"usage_type"`
		Config        map[string]any      `json:"config"`
		ExportColumns map[string]string   `json:"export_columns"`
		Body          json.RawMessage     `json:"body"`
		TaskID        string              `json:"task_id"`
		IDs           []string            `json:"ids"`
		QueryValues   map[string][]string `json:"query_values"`
	} `json:"input"`
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
	Status   int             `json:"status"`
}

func TestConnectorOptionsSharedHTTPContracts(t *testing.T) {
	runConnectorSharedHTTPContracts(t, "connector_options_http_contract.json")
}

func TestConnectorCoreSharedHTTPContracts(t *testing.T) {
	runConnectorSharedHTTPContracts(t, "connector_core_http_contract.json")
}

func TestConnectorStructuredSharedHTTPContracts(t *testing.T) {
	runConnectorSharedHTTPContracts(t, "connector_structured_http_contract.json")
}

func TestConnectorRESTSharedHTTPContracts(t *testing.T) {
	runConnectorSharedHTTPContracts(t, "connector_rest_http_contract.json")
}

func TestImportExportLifecycleSharedHTTPContracts(t *testing.T) {
	runConnectorSharedHTTPContracts(t, "import_export_lifecycle_http_contract.json")
}

func runConnectorSharedHTTPContracts(t *testing.T, fixture string) {
	data, err := os.ReadFile("../testdata/product-contracts/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	var cases []connectorCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			bodies := make(chan any, 2)
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
				var body any
				if len(scenario.Request) != 0 && string(scenario.Request) != "null" {
					decoder := json.NewDecoder(r.Body)
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil {
						t.Errorf("decode request: %v", err)
					}
				}
				bodies <- body
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
			var goResult any
			switch scenario.Operation {
			case "preview-file":
				args = append(args, "--connector-id", scenario.Input.ConnectorID, "--uri", scenario.Input.URI)
				var options []sdk.ConnectorPreviewOption
				if scenario.Input.ConnFileID != nil {
					options = append(options, sdk.WithConnectorPreviewConnFileID(*scenario.Input.ConnFileID))
					args = append(args, "--conn-file-id", *scenario.Input.ConnFileID)
				}
				if scenario.Input.SheetName != nil {
					options = append(options, sdk.WithConnectorPreviewSheetName(*scenario.Input.SheetName))
					args = append(args, "--sheet-name", *scenario.Input.SheetName)
				}
				if scenario.Input.IsColumnName != nil {
					options = append(options, sdk.WithConnectorPreviewColumnName(*scenario.Input.IsColumnName))
					args = append(args, fmt.Sprintf("--has-header=%t", *scenario.Input.IsColumnName))
				}
				if scenario.Input.RowStart != nil {
					options = append(options, sdk.WithConnectorPreviewRowStart(*scenario.Input.RowStart))
					args = append(args, "--row-start", fmt.Sprint(*scenario.Input.RowStart))
				}
				if scenario.Input.ColumnNameRow != nil {
					options = append(options, sdk.WithConnectorPreviewColumnNameRow(*scenario.Input.ColumnNameRow))
					args = append(args, "--header-row", fmt.Sprint(*scenario.Input.ColumnNameRow))
				}
				if scenario.Input.FileType != nil {
					options = append(options, sdk.WithConnectorPreviewFileType(*scenario.Input.FileType))
					args = append(args, "--file-type", fmt.Sprint(*scenario.Input.FileType))
				}
				if len(scenario.Input.CSV) != 0 {
					var csv struct {
						Separator string `json:"separator"`
						Delimiter string `json:"delimiter"`
						IsEscape  bool   `json:"isEscape"`
					}
					if err := json.Unmarshal(scenario.Input.CSV, &csv); err != nil {
						t.Fatal(err)
					}
					options = append(options, sdk.WithConnectorPreviewCSV(csv.Separator, csv.Delimiter, csv.IsEscape))
					args = append(args, "--csv-separator", csv.Separator, "--csv-delimiter", csv.Delimiter, fmt.Sprintf("--csv-escape=%t", csv.IsEscape))
				}
				goResult, err = service.PreviewFile(t.Context(), scenario.Input.ConnectorID, scenario.Input.URI, options...)
			case "structured-source-capability":
				sourceType := connectorSourceTypeString(t, scenario.Input.SourceType)
				args = append(args, "--connector-id", scenario.Input.ConnectorID, "--source-type", sourceType, "--database", scenario.Input.Database, "--table", scenario.Input.Table, "--schema", scenario.Input.Schema)
				goResult, err = service.StructuredSourceCapability(t.Context(), scenario.Input.ConnectorID, sourceType, scenario.Input.Database, scenario.Input.Table, scenario.Input.Schema)
			case "probe-rest-draft":
				args = append(args, "--config-file", writeConnectorConfig(t, scenario.Input.Config))
				goResult, err = service.ProbeRESTDraft(t.Context(), scenario.Input.Config)
			case "probe-rest":
				args = append(args, "--connector-id", scenario.Input.ConnectorID)
				goResult, err = service.ProbeREST(t.Context(), scenario.Input.ConnectorID)
			case "preview-rest":
				var preview sdk.RESTPreview
				if err := json.Unmarshal(scenario.Input.Body, &preview); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--connector-id", scenario.Input.ConnectorID, "--json", string(scenario.Input.Body))
				goResult, err = service.PreviewREST(t.Context(), scenario.Input.ConnectorID, preview)
			case "task-stats":
				goResult, err = service.TaskStats(t.Context())
			case "tasks":
				var options []sdk.QueryOption
				for key, values := range scenario.Input.QueryValues {
					options = append(options, sdk.WithQueryValues(key, values...))
					for _, value := range values {
						flagName := map[string]string{"page": "page", "page_size": "page-size", "keyword": "keyword", "order_by": "order-by", "is_desc": "is-desc", "status": "status", "connector_sources": "connector-source", "load_interval_types": "load-interval-type", "source_connector_ids": "source-connector-id"}[key]
						args = append(args, "--"+flagName+"="+value)
					}
				}
				goResult, err = service.Tasks(t.Context(), options...)
			case "update-task":
				var spec sdk.ImportTaskUpdateSpec
				if err := json.Unmarshal(scenario.Input.Body, &spec); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--task-id", scenario.Input.TaskID, "--json", string(scenario.Input.Body))
				err = service.UpdateTask(t.Context(), scenario.Input.TaskID, spec)
			case "create-task":
				var spec sdk.ImportTaskCreateSpec
				if err := json.Unmarshal(scenario.Input.Body, &spec); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--json", string(scenario.Input.Body))
				goResult, err = service.CreateTask(t.Context(), spec)
			case "create-export-task":
				var spec sdk.ExportTaskSpec
				if err := json.Unmarshal(scenario.Input.Body, &spec); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--json", string(scenario.Input.Body))
				goResult, err = service.CreateExportTask(t.Context(), spec)
			case "export-tasks":
				options, optionArgs := exportTaskListContractOptions(t, scenario.Input.Body)
				args = append(args, optionArgs...)
				goResult, err = service.ExportTasks(t.Context(), options...)
			case "export-task-files":
				args = append(args, "--task-id", scenario.Input.TaskID)
				options, optionArgs := exportTaskFilesContractOptions(t, scenario.Input.Body)
				args = append(args, optionArgs...)
				goResult, err = service.ExportTaskFiles(t.Context(), scenario.Input.TaskID, options...)
			case "delete-task", "pause-task", "resume-task", "run-task-now", "run-task-full-refresh":
				args = append(args, "--task-id", scenario.Input.TaskID)
				switch scenario.Operation {
				case "delete-task":
					err = service.DeleteTask(t.Context(), scenario.Input.TaskID)
				case "pause-task":
					err = service.PauseTask(t.Context(), scenario.Input.TaskID)
				case "resume-task":
					err = service.ResumeTask(t.Context(), scenario.Input.TaskID)
				case "run-task-now":
					err = service.RunTaskNow(t.Context(), scenario.Input.TaskID)
				case "run-task-full-refresh":
					err = service.RunTaskFullRefresh(t.Context(), scenario.Input.TaskID)
				}
			case "get-task":
				args = append(args, "--task-id", scenario.Input.TaskID)
				goResult, err = service.GetTask(t.Context(), scenario.Input.TaskID)
			case "task-files":
				var options []sdk.QueryOption
				for key, values := range scenario.Input.QueryValues {
					options = append(options, sdk.WithQueryValues(key, values...))
					for _, value := range values {
						args = append(args, "--query-param", key+"="+value)
					}
				}
				goResult, err = service.TaskFiles(t.Context(), options...)
			case "task-runs":
				args = append(args, "--task-id", scenario.Input.TaskID)
				goResult, err = service.TaskRuns(t.Context(), scenario.Input.TaskID)
			case "retry-task-files":
				args = append(args, "--task-id", scenario.Input.TaskID, "--ids", strings.Join(scenario.Input.IDs, ","))
				err = service.RetryTaskFiles(t.Context(), scenario.Input.TaskID, scenario.Input.IDs...)
			case "export-task-stats":
				goResult, err = service.ExportTaskStats(t.Context())
			case "export-task-info":
				args = append(args, "--task-id", scenario.Input.TaskID)
				goResult, err = service.ExportTaskInfo(t.Context(), scenario.Input.TaskID)
			case "delete-export-task":
				args = append(args, "--task-id", scenario.Input.TaskID)
				err = service.DeleteExportTask(t.Context(), scenario.Input.TaskID)
			case "retry-export-task":
				args = append(args, "--task-id", scenario.Input.TaskID, "--file-ids", strings.Join(scenario.Input.IDs, ","))
				err = service.RetryExportTask(t.Context(), scenario.Input.TaskID, scenario.Input.IDs...)
			case "export-task-state":
				args = append(args, "--task-id", scenario.Input.TaskID)
				goResult, err = service.ExportTaskState(t.Context(), scenario.Input.TaskID)
			case "structured-sources":
				goResult, err = service.StructuredSources(t.Context())
			case "structured-source-databases":
				sourceType := connectorSourceTypeString(t, scenario.Input.SourceType)
				args = append(args, "--connector-id", scenario.Input.ConnectorID, "--source-type", sourceType)
				goResult, err = service.StructuredSourceDatabases(t.Context(), scenario.Input.ConnectorID, sourceType)
			case "structured-source-schemas":
				sourceType := connectorSourceTypeString(t, scenario.Input.SourceType)
				args = append(args, "--connector-id", scenario.Input.ConnectorID, "--source-type", sourceType, "--database", scenario.Input.Database)
				goResult, err = service.StructuredSourceSchemas(t.Context(), scenario.Input.ConnectorID, sourceType, scenario.Input.Database)
			case "structured-source-objects":
				var req sdk.StructuredSourceObjectsRequest
				if err := json.Unmarshal(scenario.Input.Body, &req); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--json", string(scenario.Input.Body))
				goResult, err = service.StructuredSourceObjects(t.Context(), req)
			case "structured-source-schema":
				var req sdk.StructuredSourceSchemaRequest
				if err := json.Unmarshal(scenario.Input.Body, &req); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--json", string(scenario.Input.Body))
				goResult, err = service.StructuredSourceSchema(t.Context(), req)
			case "structured-source-sample":
				var req struct {
					ConnectorID string                        `json:"connector_id"`
					Source      sdk.StructuredSourceRef       `json:"source"`
					Limit       int32                         `json:"limit"`
					KafkaSample sdk.StructuredKafkaSampleSpec `json:"kafka_sample"`
				}
				if err := json.Unmarshal(scenario.Input.Body, &req); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--connector-id", req.ConnectorID, "--source-type", req.Source.SourceType, "--limit", fmt.Sprint(req.Limit), "--topic-mode", req.KafkaSample.TopicMode)
				if req.KafkaSample.Topics != nil {
					topics, marshalErr := json.Marshal(req.KafkaSample.Topics)
					if marshalErr != nil {
						t.Fatal(marshalErr)
					}
					args = append(args, "--topics-json", string(topics))
				}
				if req.KafkaSample.TopicPattern != "" {
					args = append(args, "--topic-pattern", req.KafkaSample.TopicPattern)
				}
				goResult, err = service.StructuredSourceSample(t.Context(), req.ConnectorID, req.Source.SourceType, req.Limit, req.KafkaSample.TopicMode, req.KafkaSample.Topics, req.KafkaSample.TopicPattern)
			case "structured-backfill-start-resolve":
				var req sdk.StructuredBackfillStartResolveRequest
				if err := json.Unmarshal(scenario.Input.Body, &req); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--json", string(scenario.Input.Body))
				goResult, err = service.StructuredBackfillStartResolve(t.Context(), req)
			case "list":
				var options []sdk.QueryOption
				for key, value := range scenario.Input.QueryParams {
					options = append(options, sdk.WithQueryParam(key, value))
					args = append(args, "--query-param", key+"="+value)
				}
				goResult, err = service.List(t.Context(), options...)
			case "get":
				args = append(args, "--connector-id", scenario.Input.ConnectorID)
				goResult, err = service.Get(t.Context(), scenario.Input.ConnectorID)
			case "create", "validate":
				sourceType := connectorSourceTypeInt(t, scenario.Input.SourceType)
				args = append(args, "--name", *scenario.Input.Name, "--source-type", fmt.Sprint(sourceType), "--usage-type", fmt.Sprint(*scenario.Input.UsageType))
				var options []sdk.ConnectorDefinitionOption
				if scenario.Input.Config != nil {
					options = append(options, sdk.WithConnectorConfig(scenario.Input.Config))
					args = append(args, "--config-file", writeConnectorConfig(t, scenario.Input.Config))
				}
				if scenario.Operation == "create" {
					goResult, err = service.Create(t.Context(), *scenario.Input.Name, sourceType, *scenario.Input.UsageType, options...)
				} else {
					goResult, err = service.Validate(t.Context(), *scenario.Input.Name, sourceType, *scenario.Input.UsageType, options...)
				}
			case "update":
				args = append(args, "--connector-id", scenario.Input.ConnectorID)
				var options []sdk.ConnectorUpdateOption
				if scenario.Input.Name != nil {
					options = append(options, sdk.WithConnectorName(*scenario.Input.Name))
					args = append(args, "--name", *scenario.Input.Name)
				}
				if len(scenario.Input.SourceType) != 0 {
					sourceType := connectorSourceTypeInt(t, scenario.Input.SourceType)
					options = append(options, sdk.WithConnectorSourceType(sourceType))
					args = append(args, "--source-type", fmt.Sprint(sourceType))
				}
				if scenario.Input.UsageType != nil {
					options = append(options, sdk.WithConnectorUsageType(*scenario.Input.UsageType))
					args = append(args, "--usage-type", fmt.Sprint(*scenario.Input.UsageType))
				}
				if scenario.Input.Config != nil {
					options = append(options, sdk.WithConnectorUpdateConfig(scenario.Input.Config))
					args = append(args, "--config-file", writeConnectorConfig(t, scenario.Input.Config))
				}
				err = service.Update(t.Context(), scenario.Input.ConnectorID, options...)
			case "delete":
				args = append(args, "--connector-id", scenario.Input.ConnectorID)
				err = service.Delete(t.Context(), scenario.Input.ConnectorID)
			case "db-list-databases", "db-list-schemas", "db-list-tables", "db-source-schema":
				args = append(args, "--connector-id", scenario.Input.ConnectorID)
				var options []sdk.ConnectorDBOption
				switch scenario.Operation {
				case "db-list-databases":
					goResult, err = service.DBListDatabases(t.Context(), scenario.Input.ConnectorID)
				case "db-list-schemas":
					args = append(args, "--database", scenario.Input.Database)
					goResult, err = service.DBListSchemas(t.Context(), scenario.Input.ConnectorID, scenario.Input.Database)
				case "db-list-tables", "db-source-schema":
					args = append(args, "--database", scenario.Input.Database, "--schema", scenario.Input.Schema)
					options = append(options, sdk.WithConnectorSchema(scenario.Input.Schema))
					if scenario.Operation == "db-list-tables" {
						goResult, err = service.DBListTables(t.Context(), scenario.Input.ConnectorID, scenario.Input.Database, options...)
					} else {
						args = append(args, "--table", scenario.Input.Table)
						goResult, err = service.DBSourceSchema(t.Context(), scenario.Input.ConnectorID, scenario.Input.Database, scenario.Input.Table, options...)
					}
				}
			case "mo-database-list", "mo-database-table-list", "mo-check-table-name", "mo-column-mapping":
				args = append(args, "--connector-id", scenario.Input.ConnectorID)
				switch scenario.Operation {
				case "mo-database-list":
					goResult, err = service.MODatabaseList(t.Context(), scenario.Input.ConnectorID)
				case "mo-database-table-list":
					args = append(args, "--database", scenario.Input.Database)
					goResult, err = service.MODatabaseTableList(t.Context(), scenario.Input.ConnectorID, scenario.Input.Database)
				case "mo-check-table-name":
					args = append(args, "--database", scenario.Input.Database, "--table", scenario.Input.Table)
					goResult, err = service.MOCheckTableName(t.Context(), scenario.Input.ConnectorID, scenario.Input.Database, scenario.Input.Table)
				case "mo-column-mapping":
					encoded, marshalErr := json.Marshal(scenario.Input.ExportColumns)
					if marshalErr != nil {
						t.Fatal(marshalErr)
					}
					args = append(args, "--database", scenario.Input.Database, "--table", scenario.Input.Table, "--json", string(encoded))
					goResult, err = service.MOColumnMapping(t.Context(), scenario.Input.ConnectorID, scenario.Input.Database, scenario.Input.Table, scenario.Input.ExportColumns)
				}
			default:
				t.Fatalf("unknown operation %s", scenario.Operation)
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := <-bodies; !reflect.DeepEqual(got, catalogFolderCLIJSON(scenario.Request)) {
				t.Fatalf("Go body=%#v, want=%#v", got, catalogFolderCLIJSON(scenario.Request))
			}
			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteConnectorCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if cliErr != nil {
				t.Fatal(cliErr)
			}
			if got := <-bodies; !reflect.DeepEqual(got, catalogFolderCLIJSON(scenario.Request)) {
				t.Fatalf("CLI body=%#v, want=%#v", got, catalogFolderCLIJSON(scenario.Request))
			}
			if connectorCLIVoid(scenario.Operation) {
				if expected := "connector " + scenario.Operation + ": ok\n"; output != expected {
					t.Fatalf("CLI output=%q, want=%q", output, expected)
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
			if fixture == "connector_structured_http_contract.json" || scenario.Operation == "structured-source-capability" {
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(catalogFolderCLIJSON(json.RawMessage(output)), catalogFolderCLIJSON(envelope.Data)) {
					t.Fatalf("CLI result=%s, owner data=%s", output, envelope.Data)
				}
			}
			if fixture == "connector_core_http_contract.json" || fixture == "connector_rest_http_contract.json" || fixture == "import_export_lifecycle_http_contract.json" {
				var envelope struct {
					Data json.RawMessage `json:"data"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				expected := connectorCLIResult(scenario.Operation)
				if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
					t.Fatal(err)
				}
				actual := expected.ProtoReflect().New().Interface()
				if err := json.Unmarshal([]byte(output), actual); err != nil {
					t.Fatalf("CLI output=%q: %v", output, err)
				}
				if !proto.Equal(actual, expected) {
					t.Fatalf("CLI result=%v, owner data=%v", actual, expected)
				}
			}
		})
	}
}

type exportTaskListContractInput struct {
	Keyword        *string `json:"keyword"`
	TargetCategory *string `json:"target_category"`
	Limit          *int    `json:"limit"`
	Offset         *int    `json:"offset"`
	ConnectorName  *string `json:"connector_name"`
	TaskID         *string `json:"task_id"`
	OrderBy        *string `json:"order_by"`
	OrderDirection *string `json:"order_direction"`
	Statuses       []int   `json:"statuses"`
	Creator        *string `json:"creator"`
}

func exportTaskListContractOptions(t *testing.T, raw json.RawMessage) ([]sdk.ExportTaskListOption, []string) {
	t.Helper()
	var input exportTaskListContractInput
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &input); err != nil {
			t.Fatal(err)
		}
	}
	var options []sdk.ExportTaskListOption
	var args []string
	if input.Keyword != nil {
		options = append(options, sdk.WithExportTaskKeyword(*input.Keyword))
		args = append(args, "--keyword", *input.Keyword)
	}
	if input.TargetCategory != nil {
		options = append(options, sdk.WithExportTaskTargetCategory(*input.TargetCategory))
		args = append(args, "--target-category", *input.TargetCategory)
	}
	if input.Limit != nil {
		options = append(options, sdk.WithExportTaskLimit(*input.Limit))
		args = append(args, "--limit", fmt.Sprint(*input.Limit))
	}
	if input.Offset != nil {
		options = append(options, sdk.WithExportTaskOffset(*input.Offset))
		args = append(args, "--offset", fmt.Sprint(*input.Offset))
	}
	if input.ConnectorName != nil {
		options = append(options, sdk.WithExportTaskConnectorName(*input.ConnectorName))
		args = append(args, "--connector-name", *input.ConnectorName)
	}
	if input.TaskID != nil {
		options = append(options, sdk.WithExportTaskID(*input.TaskID))
		args = append(args, "--task-id", *input.TaskID)
	}
	if input.OrderBy != nil {
		options = append(options, sdk.WithExportTaskOrderBy(*input.OrderBy))
		args = append(args, "--order-by", *input.OrderBy)
	}
	if input.OrderDirection != nil {
		options = append(options, sdk.WithExportTaskOrderDirection(*input.OrderDirection))
		args = append(args, "--order-direction", *input.OrderDirection)
	}
	if len(input.Statuses) > 0 {
		options = append(options, sdk.WithExportTaskStatuses(input.Statuses...))
		args = append(args, "--statuses", contractIntCSV(input.Statuses))
	}
	if input.Creator != nil {
		options = append(options, sdk.WithExportTaskCreator(*input.Creator))
		args = append(args, "--creator", *input.Creator)
	}
	return options, args
}

type exportTaskFilesContractInput struct {
	Limit          *int    `json:"limit"`
	Offset         *int    `json:"offset"`
	Statuses       []int   `json:"statuses"`
	OrderBy        *string `json:"order_by"`
	OrderDirection *string `json:"order_direction"`
}

func exportTaskFilesContractOptions(t *testing.T, raw json.RawMessage) ([]sdk.ExportTaskFilesOption, []string) {
	t.Helper()
	var input exportTaskFilesContractInput
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &input); err != nil {
			t.Fatal(err)
		}
	}
	var options []sdk.ExportTaskFilesOption
	var args []string
	if input.Limit != nil {
		options = append(options, sdk.WithExportTaskFilesLimit(*input.Limit))
		args = append(args, "--limit", fmt.Sprint(*input.Limit))
	}
	if input.Offset != nil {
		options = append(options, sdk.WithExportTaskFilesOffset(*input.Offset))
		args = append(args, "--offset", fmt.Sprint(*input.Offset))
	}
	if len(input.Statuses) > 0 {
		options = append(options, sdk.WithExportTaskFileStatuses(input.Statuses...))
		args = append(args, "--statuses", contractIntCSV(input.Statuses))
	}
	if input.OrderBy != nil {
		options = append(options, sdk.WithExportTaskFilesOrderBy(*input.OrderBy))
		args = append(args, "--order-by", *input.OrderBy)
	}
	if input.OrderDirection != nil {
		options = append(options, sdk.WithExportTaskFilesOrderDirection(*input.OrderDirection))
		args = append(args, "--order-direction", *input.OrderDirection)
	}
	return options, args
}

func contractIntCSV(values []int) string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, fmt.Sprint(value))
	}
	return strings.Join(items, ",")
}

func connectorSourceTypeString(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func connectorSourceTypeInt(t *testing.T, raw json.RawMessage) int {
	t.Helper()
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func writeConnectorConfig(t *testing.T, config map[string]any) string {
	t.Helper()
	encoded, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "connector-config.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func connectorCLIVoid(operation string) bool {
	switch operation {
	case "update", "delete", "update-task", "delete-task", "pause-task", "resume-task", "run-task-now", "run-task-full-refresh", "retry-task-files", "delete-export-task", "retry-export-task":
		return true
	default:
		return false
	}
}

func connectorCLIResult(operation string) proto.Message {
	switch operation {
	case "list":
		return &sdk.ConnectorListResult{}
	case "get", "create":
		return &sdk.Connector{}
	case "validate":
		return &sdk.ConnectorValidateResult{}
	case "probe-rest-draft", "probe-rest":
		return &sdk.ConnectorValidateResult{}
	case "preview-rest":
		return &sdk.RESTReadResult{}
	case "task-stats":
		return &sdk.ImportTaskStats{}
	case "create-task":
		return &sdk.ImportTaskCreateResult{}
	case "tasks":
		return &sdk.ImportTaskListResult{}
	case "get-task":
		return &sdk.ImportTaskInfoResult{}
	case "task-files":
		return &sdk.ImportTaskFileListResult{}
	case "task-runs":
		return &sdk.ImportTaskRunListResult{}
	case "export-task-stats":
		return &sdk.ExportTaskStats{}
	case "export-task-info":
		return &sdk.ExportTaskInfo{}
	case "export-task-state":
		return &sdk.ExportTaskStateResult{}
	case "create-export-task":
		return &sdk.ExportTask{}
	case "export-tasks":
		return &sdk.ExportTaskListResult{}
	case "export-task-files":
		return &sdk.ExportTaskFilesResult{}
	case "db-list-databases":
		return &sdk.DBListDatabasesResult{}
	case "db-list-schemas":
		return &sdk.DBListSchemasResult{}
	case "db-list-tables":
		return &sdk.DBListTablesResult{}
	case "db-source-schema":
		return &sdk.DBSourceSchemaResult{}
	case "mo-database-list":
		return &sdk.MODatabaseListResult{}
	case "mo-database-table-list":
		return &sdk.MODatabaseTableListResult{}
	case "mo-check-table-name":
		return &sdk.MOCheckTableNameResult{}
	case "mo-column-mapping":
		return &sdk.MOColumnMappingResult{}
	default:
		return nil
	}
}
