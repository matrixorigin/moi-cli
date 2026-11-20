package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestConnectorCLIOptionsAlignWithGoSDK(t *testing.T) {
	for _, scenario := range []struct {
		name, path, response string
		args                 []string
		invokeGo             func(*sdk.ConnectorService) (any, error)
	}{
		{
			name: "task defaults", path: "/export/task/list", response: `{"tasks":[],"total":0}`,
			args: []string{"export-tasks"},
			invokeGo: func(service *sdk.ConnectorService) (any, error) {
				return service.ExportTasks(t.Context())
			},
		},
		{
			name: "task explicit zero and empty values", path: "/export/task/list", response: `{"tasks":[],"total":0}`,
			args: []string{"export-tasks", "--limit", "0", "--offset", "0", "--keyword=", "--creator="},
			invokeGo: func(service *sdk.ConnectorService) (any, error) {
				return service.ExportTasks(t.Context(), sdk.WithExportTaskLimit(0), sdk.WithExportTaskOffset(0), sdk.WithExportTaskKeyword(""), sdk.WithExportTaskCreator(""))
			},
		},
		{
			name: "database table schema selection", path: "/connectors/db/database/table/list", response: `{"tables":["orders"]}`,
			args: []string{"db-list-tables", "--connector-id", "conn-1", "--database", "sales", "--schema", "public"},
			invokeGo: func(service *sdk.ConnectorService) (any, error) {
				return service.DBListTables(t.Context(), "conn-1", "sales", sdk.WithConnectorSchema("public"))
			},
		},
		{
			name: "table columns schema selection", path: "/connectors/db/schema", response: `{"columns":[{"name":"id","data_type":"INT","mapped_type":"INT64","is_primary_key":true,"is_nullable":false}]}`,
			args: []string{"db-source-schema", "--connector-id", "conn-1", "--database", "sales", "--table", "orders", "--schema", "public"},
			invokeGo: func(service *sdk.ConnectorService) (any, error) {
				return service.DBSourceSchema(t.Context(), "conn-1", "sales", "orders", sdk.WithConnectorSchema("public"))
			},
		},
		{
			name: "task filters", path: "/export/task/list", response: `{"tasks":[{"id":"export-1","name":"导出"}],"total":1}`,
			args: []string{"export-tasks", "--keyword", "报告", "--target-category", "storage", "--limit", "20", "--offset", "40", "--connector-name", "oss", "--task-id", "export-1", "--order-by", "created_at", "--order-direction", "desc", "--statuses", "1,2", "--creator", "alice"},
			invokeGo: func(service *sdk.ConnectorService) (any, error) {
				return service.ExportTasks(t.Context(), sdk.WithExportTaskKeyword("报告"), sdk.WithExportTaskTargetCategory("storage"), sdk.WithExportTaskLimit(20), sdk.WithExportTaskOffset(40), sdk.WithExportTaskConnectorName("oss"), sdk.WithExportTaskID("export-1"), sdk.WithExportTaskOrderBy("created_at"), sdk.WithExportTaskOrderDirection("desc"), sdk.WithExportTaskStatuses(1, 2), sdk.WithExportTaskCreator("alice"))
			},
		},
		{
			name: "file filters", path: "/export/task/files", response: `{"files":[{"id":"file-1","status":3}],"total":1,"total_failure":1,"total_files":1}`,
			args: []string{"export-task-files", "--task-id", "export-1", "--limit", "10", "--offset", "5", "--statuses", "3", "--order-by", "created_at", "--order-direction", "asc"},
			invokeGo: func(service *sdk.ConnectorService) (any, error) {
				return service.ExportTaskFiles(t.Context(), "export-1", sdk.WithExportTaskFilesLimit(10), sdk.WithExportTaskFilesOffset(5), sdk.WithExportTaskFileStatuses(3), sdk.WithExportTaskFilesOrderBy("created_at"), sdk.WithExportTaskFilesOrderDirection("asc"))
			},
		},
		{
			name: "file options omitted", path: "/export/task/files", response: `{"files":[],"total":0}`,
			args: []string{"export-task-files", "--task-id", "export-1"},
			invokeGo: func(service *sdk.ConnectorService) (any, error) {
				return service.ExportTaskFiles(t.Context(), "export-1")
			},
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			requests := make(chan map[string]any, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/newmoi"+scenario.path || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected target: %s %s, workspace=%s", r.Method, r.URL.Path, r.Header.Get("X-Workspace-ID"))
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				requests <- body
				_, _ = w.Write([]byte(`{"code":"OK","data":` + scenario.response + `}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			want, err := scenario.invokeGo(client.Connectors("workspace-1"))
			if err != nil {
				t.Fatal(err)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteConnectorCommand(NewContext(client, "workspace-1", false, true, true), scenario.args)
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			goBody, cliBody := <-requests, <-requests
			if !reflect.DeepEqual(cliBody, goBody) {
				t.Fatalf("CLI request=%v, Go SDK request=%v", cliBody, goBody)
			}
			encoded, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			var goOutput, cliOutput any
			if err := json.Unmarshal(encoded, &goOutput); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(output), &cliOutput); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(cliOutput, goOutput) {
				t.Fatalf("CLI output=%v, Go SDK output=%v", cliOutput, goOutput)
			}
		})
	}
}
