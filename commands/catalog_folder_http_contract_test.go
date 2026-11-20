package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type catalogFolderCLIContract struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Input     struct {
		CatalogID         int                    `json:"catalog_id"`
		VolumeID          int                    `json:"volume_id"`
		VolumeKey         string                 `json:"volume_key"`
		ParentID          *string                `json:"parent_id"`
		Extensions        []string               `json:"extensions"`
		FileType          *string                `json:"file_type"`
		Order             *string                `json:"order"`
		OrderBy           *string                `json:"order_by"`
		ParentVolumeID    int                    `json:"parent_volume_id"`
		DatabaseID        int                    `json:"database_id"`
		FolderID          int                    `json:"folder_id"`
		TableID           int                    `json:"table_id"`
		TableIDs          []int                  `json:"table_ids"`
		ResourceType      string                 `json:"resource_type"`
		ResourceIDs       []string               `json:"resource_ids"`
		ReferenceID       int64                  `json:"id"`
		SQL               string                 `json:"sql"`
		AssetID           string                 `json:"asset_id"`
		FileID            string                 `json:"file_id"`
		RawFileIDs        []string               `json:"raw_file_ids"`
		FileIDs           []string               `json:"file_ids"`
		Name              *string                `json:"name"`
		DatabaseName      *string                `json:"database_name"`
		Description       *string                `json:"description"`
		Comment           *string                `json:"comment"`
		DataFileURL       *string                `json:"data_file_url"`
		Columns           json.RawMessage        `json:"columns"`
		ColumnLoadOptions []sdk.ColumnLoadOption `json:"column_load_options"`
		IncludeStats      *bool                  `json:"include_stats"`
		IncludeTotal      *bool                  `json:"include_total"`
		Page              *int                   `json:"page"`
		PageSize          *int                   `json:"page_size"`
		PageToken         *string                `json:"page_token"`
		PreviewLines      *int                   `json:"preview_lines"`
	} `json:"input"`
	Request         json.RawMessage     `json:"request"`
	Query           map[string][]string `json:"query"`
	Response        json.RawMessage     `json:"response"`
	ResponseHeaders map[string]string   `json:"response_headers"`
	Status          int                 `json:"status"`
}

func catalogFolderCLIJSON(raw json.RawMessage) any {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		panic(err)
	}
	return value
}

func catalogFolderCLIError(t *testing.T, err error, scenario catalogFolderCLIContract) *sdk.Error {
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
		t.Fatalf("unexpected product error: %+v", productErr)
	}
	return productErr
}

func TestCatalogFolderSharedHTTPContracts(t *testing.T) {
	runCatalogFolderSharedHTTPContracts(t, "catalog_folder_http_contract.json")
}

func TestCatalogInfoSharedHTTPContracts(t *testing.T) {
	runCatalogFolderSharedHTTPContracts(t, "catalog_info_http_contract.json")
}

func TestCatalogDataAssetSharedHTTPContracts(t *testing.T) {
	runCatalogFolderSharedHTTPContracts(t, "catalog_data_asset_http_contract.json")
}

func TestCatalogTableSharedHTTPContracts(t *testing.T) {
	runCatalogFolderSharedHTTPContracts(t, "catalog_table_http_contract.json")
}

func TestCatalogFileSharedHTTPContracts(t *testing.T) {
	runCatalogFolderSharedHTTPContracts(t, "catalog_file_http_contract.json")
}

func TestCatalogCoreSharedHTTPContracts(t *testing.T) {
	runCatalogFolderSharedHTTPContracts(t, "catalog_core_http_contract.json")
}

func runCatalogFolderSharedHTTPContracts(t *testing.T, fixture string) {
	data, err := os.ReadFile("../testdata/product-contracts/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	var cases []catalogFolderCLIContract
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
			service := client.Catalog("workspace-1")
			args := []string{scenario.Operation}
			var goResult any
			var goErr error
			switch scenario.Operation {
			case "create-folder":
				args = append(args, "--parent-volume-id", fmt.Sprint(scenario.Input.ParentVolumeID), "--database-id", fmt.Sprint(scenario.Input.DatabaseID), "--name", *scenario.Input.Name)
				var options []sdk.CreateFolderOption
				if scenario.Input.Description != nil {
					options = append(options, sdk.WithFolderDescription(*scenario.Input.Description))
					args = append(args, "--description", *scenario.Input.Description)
				}
				goResult, goErr = service.CreateFolder(t.Context(), scenario.Input.ParentVolumeID, scenario.Input.DatabaseID, *scenario.Input.Name, options...)
			case "update-folder":
				args = append(args, "--folder-id", fmt.Sprint(scenario.Input.FolderID))
				var options []sdk.UpdateFolderOption
				if scenario.Input.Name != nil {
					options = append(options, sdk.WithFolderName(*scenario.Input.Name))
					args = append(args, "--name", *scenario.Input.Name)
				}
				if scenario.Input.Description != nil {
					options = append(options, sdk.WithUpdatedFolderDescription(*scenario.Input.Description))
					args = append(args, "--description", *scenario.Input.Description)
				}
				goErr = service.UpdateFolder(t.Context(), scenario.Input.FolderID, options...)
			case "delete-folder":
				args = append(args, "--folder-id", fmt.Sprint(scenario.Input.FolderID))
				goErr = service.DeleteFolder(t.Context(), scenario.Input.FolderID)
			case "clean-folder":
				args = append(args, "--folder-id", fmt.Sprint(scenario.Input.FolderID))
				goErr = service.CleanFolder(t.Context(), scenario.Input.FolderID)
			case "create", "create-database", "create-volume":
				var options []sdk.CatalogInfoOption
				if scenario.Input.Description != nil {
					options = append(options, sdk.WithCatalogDescription(*scenario.Input.Description))
					args = append(args, "--description", *scenario.Input.Description)
				}
				args = append(args, "--name", *scenario.Input.Name)
				switch scenario.Operation {
				case "create":
					goResult, goErr = service.Create(t.Context(), *scenario.Input.Name, options...)
				case "create-database":
					args = append(args, "--catalog-id", fmt.Sprint(scenario.Input.CatalogID))
					goResult, goErr = service.CreateDatabase(t.Context(), scenario.Input.CatalogID, *scenario.Input.Name, options...)
				case "create-volume":
					args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
					goResult, goErr = service.CreateVolume(t.Context(), scenario.Input.DatabaseID, *scenario.Input.Name, options...)
				}
			case "update", "update-database-description", "update-volume":
				var options []sdk.CatalogInfoOption
				if scenario.Input.Description != nil {
					options = append(options, sdk.WithCatalogDescription(*scenario.Input.Description))
					args = append(args, "--description", *scenario.Input.Description)
				}
				switch scenario.Operation {
				case "update":
					args = append(args, "--catalog-id", fmt.Sprint(scenario.Input.CatalogID), "--name", *scenario.Input.Name)
					goErr = service.Update(t.Context(), scenario.Input.CatalogID, *scenario.Input.Name, options...)
				case "update-database-description":
					args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
					goResult, goErr = service.UpdateDatabaseDescription(t.Context(), scenario.Input.DatabaseID, options...)
				case "update-volume":
					args = append(args, "--volume-id", fmt.Sprint(scenario.Input.VolumeID), "--name", *scenario.Input.Name)
					goResult, goErr = service.UpdateVolume(t.Context(), scenario.Input.VolumeID, *scenario.Input.Name, options...)
				}
			case "data-asset":
				args = append(args, "--asset-id", scenario.Input.AssetID)
				goResult, goErr = service.DataAsset(t.Context(), scenario.Input.AssetID)
			case "batch-get-data-assets":
				args = append(args, "--raw-file-ids", strings.Join(scenario.Input.RawFileIDs, ","))
				goResult, goErr = service.BatchGetDataAssets(t.Context(), scenario.Input.RawFileIDs...)
			case "file-data-asset":
				args = append(args, "--file-id", scenario.Input.FileID)
				goResult, goErr = service.FileDataAsset(t.Context(), scenario.Input.FileID)
			case "file-artifact":
				args = append(args, "--file-id", scenario.Input.FileID)
				goResult, goErr = service.FileArtifact(t.Context(), scenario.Input.FileID)
			case "batch-resolve-file-artifacts":
				args = append(args, "--file-ids", strings.Join(scenario.Input.FileIDs, ","))
				goResult, goErr = service.BatchResolveFileArtifacts(t.Context(), scenario.Input.FileIDs...)
			case "create-table":
				var columns []sdk.TableColumn
				if err := json.Unmarshal(scenario.Input.Columns, &columns); err != nil {
					t.Fatal(err)
				}
				var options []sdk.CreateTableOption
				args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID), "--name", *scenario.Input.Name, "--json", string(scenario.Input.Columns))
				if scenario.Input.Comment != nil {
					options = append(options, sdk.WithTableComment(*scenario.Input.Comment))
					args = append(args, "--comment", *scenario.Input.Comment)
				}
				if scenario.Input.DataFileURL != nil {
					options = append(options, sdk.WithTableDataFileURL(*scenario.Input.DataFileURL))
					args = append(args, "--data-file-url", *scenario.Input.DataFileURL)
				}
				if scenario.Input.ColumnLoadOptions != nil {
					encoded, err := json.Marshal(scenario.Input.ColumnLoadOptions)
					if err != nil {
						t.Fatal(err)
					}
					args = append(args, "--column-load-options", string(encoded))
					for _, option := range scenario.Input.ColumnLoadOptions {
						options = append(options, sdk.WithColumnLoadOption(option.Column, option.DataFrom))
					}
				}
				goResult, goErr = service.CreateTable(t.Context(), scenario.Input.DatabaseID, *scenario.Input.Name, columns, options...)
			case "table-info":
				var options []sdk.TableLookupOption
				args = append(args, "--table-id", fmt.Sprint(scenario.Input.TableID))
				if scenario.Input.Name != nil {
					options = append(options, sdk.WithTableLookupName(*scenario.Input.Name))
					args = append(args, "--name", *scenario.Input.Name)
				}
				if scenario.Input.DatabaseID != 0 {
					options = append(options, sdk.WithTableLookupDatabaseID(scenario.Input.DatabaseID))
					args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
				}
				if scenario.Input.DatabaseName != nil {
					options = append(options, sdk.WithTableLookupDatabaseName(*scenario.Input.DatabaseName))
					args = append(args, "--database-name", *scenario.Input.DatabaseName)
				}
				if scenario.Input.IncludeStats != nil {
					options = append(options, sdk.WithTableLookupIncludeStats(*scenario.Input.IncludeStats))
					args = append(args, fmt.Sprintf("--include-stats=%t", *scenario.Input.IncludeStats))
				}
				goResult, goErr = service.TableInfo(t.Context(), scenario.Input.TableID, options...)
			case "table-data", "preview-table":
				var options []sdk.TableDataOption
				args = append(args, "--table-id", fmt.Sprint(scenario.Input.TableID))
				if scenario.Input.Page != nil {
					options = append(options, sdk.WithTableDataPage(*scenario.Input.Page))
					args = append(args, "--page", fmt.Sprint(*scenario.Input.Page))
				}
				if scenario.Input.PageSize != nil {
					options = append(options, sdk.WithTableDataPageSize(*scenario.Input.PageSize))
					args = append(args, "--page-size", fmt.Sprint(*scenario.Input.PageSize))
				}
				if scenario.Input.Name != nil {
					options = append(options, sdk.WithTableDataName(*scenario.Input.Name))
					args = append(args, "--name", *scenario.Input.Name)
				}
				if scenario.Input.DatabaseID != 0 {
					options = append(options, sdk.WithTableDataDatabaseID(scenario.Input.DatabaseID))
					args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
				}
				if scenario.Input.DatabaseName != nil {
					options = append(options, sdk.WithTableDataDatabaseName(*scenario.Input.DatabaseName))
					args = append(args, "--database-name", *scenario.Input.DatabaseName)
				}
				if scenario.Input.PreviewLines != nil {
					options = append(options, sdk.WithTablePreviewLines(*scenario.Input.PreviewLines))
					args = append(args, "--preview-lines", fmt.Sprint(*scenario.Input.PreviewLines))
				}
				if scenario.Input.IncludeTotal != nil {
					options = append(options, sdk.WithTableDataIncludeTotal(*scenario.Input.IncludeTotal))
					args = append(args, fmt.Sprintf("--include-total=%t", *scenario.Input.IncludeTotal))
				}
				if scenario.Operation == "table-data" {
					goResult, goErr = service.TableData(t.Context(), scenario.Input.TableID, options...)
				} else {
					goResult, goErr = service.PreviewTable(t.Context(), scenario.Input.TableID, options...)
				}
			case "files":
				options := []sdk.FileListOption{
					sdk.WithFileListFilter("file_ext", scenario.Input.Extensions...),
					sdk.WithFileListFilter("file_ids", scenario.Input.FileIDs...),
				}
				args = append(args, "--extensions", strings.Join(scenario.Input.Extensions, ","), "--file-ids", strings.Join(scenario.Input.FileIDs, ","))
				if scenario.Input.FileType != nil {
					options = append(options, sdk.WithFileListFilter("file_type", *scenario.Input.FileType))
					args = append(args, "--file-type", *scenario.Input.FileType)
				}
				if scenario.Input.Name != nil {
					options = append(options, sdk.WithFuzzyFileListFilter("file_name", true, *scenario.Input.Name))
					args = append(args, "--name", *scenario.Input.Name)
				}
				if scenario.Input.ParentID != nil {
					options = append(options, sdk.WithFileListFilter("parent_id", *scenario.Input.ParentID))
					args = append(args, "--parent-id", *scenario.Input.ParentID)
				}
				options = append(options, sdk.WithFileListFilter("volume_id", scenario.Input.VolumeKey))
				args = append(args, "--volume-id", scenario.Input.VolumeKey)
				if scenario.Input.Page != nil {
					options = append(options, sdk.WithFileListPage(*scenario.Input.Page))
					args = append(args, "--page", fmt.Sprint(*scenario.Input.Page))
				}
				if scenario.Input.PageSize != nil {
					options = append(options, sdk.WithFileListPageSize(*scenario.Input.PageSize))
					args = append(args, "--page-size", fmt.Sprint(*scenario.Input.PageSize))
				}
				if scenario.Input.Order != nil {
					options = append(options, sdk.WithFileListOrder(*scenario.Input.Order))
					args = append(args, "--order", *scenario.Input.Order)
				}
				if scenario.Input.OrderBy != nil {
					options = append(options, sdk.WithFileListOrderBy(*scenario.Input.OrderBy))
					args = append(args, "--order-by", *scenario.Input.OrderBy)
				}
				goResult, goErr = service.Files(t.Context(), options...)
			case "create-file":
				args = append(args, "--volume-id", fmt.Sprint(scenario.Input.VolumeID), "--file-ids", strings.Join(scenario.Input.FileIDs, ","))
				goErr = service.CreateFile(t.Context(), scenario.Input.VolumeID, scenario.Input.FileIDs...)
			case "file-info":
				args = append(args, "--file-id", scenario.Input.FileID)
				goResult, goErr = service.FileInfo(t.Context(), scenario.Input.FileID)
			case "delete-file":
				args = append(args, "--file-id", scenario.Input.FileID, "--volume-id", scenario.Input.VolumeKey)
				goErr = service.DeleteFile(t.Context(), scenario.Input.FileID, scenario.Input.VolumeKey)
			case "delete-file-ref":
				args = append(args, "--volume-id", fmt.Sprint(scenario.Input.VolumeID), "--file-ids", strings.Join(scenario.Input.FileIDs, ","))
				goErr = service.DeleteFileRef(t.Context(), scenario.Input.VolumeID, scenario.Input.FileIDs...)
			case "trigger-file-workflow":
				args = append(args, "--volume-id", fmt.Sprint(scenario.Input.VolumeID), "--file-ids", strings.Join(scenario.Input.FileIDs, ","))
				goResult, goErr = service.TriggerFileWorkflow(t.Context(), scenario.Input.VolumeID, scenario.Input.FileIDs...)
			case "multi-table-info":
				ids := make([]string, 0, len(scenario.Input.TableIDs))
				for _, id := range scenario.Input.TableIDs {
					ids = append(ids, fmt.Sprint(id))
				}
				args = append(args, "--table-ids", strings.Join(ids, ","))
				goResult, goErr = service.MultiTableInfo(t.Context(), scenario.Input.TableIDs...)
			case "delete-table":
				args = append(args, "--table-id", fmt.Sprint(scenario.Input.TableID))
				goErr = service.DeleteTable(t.Context(), scenario.Input.TableID)
			case "table-exists":
				args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID), "--name", *scenario.Input.Name)
				goResult, goErr = service.TableExists(t.Context(), scenario.Input.DatabaseID, *scenario.Input.Name)
			case "table-full-path":
				args = append(args, "--table-id", fmt.Sprint(scenario.Input.TableID))
				goResult, goErr = service.TableFullPath(t.Context(), scenario.Input.TableID)
			case "truncate-table":
				args = append(args, "--table-id", fmt.Sprint(scenario.Input.TableID))
				goErr = service.TruncateTable(t.Context(), scenario.Input.TableID)
			case "table-overview":
				args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
				goResult, goErr = service.TableOverview(t.Context(), scenario.Input.DatabaseID)
			case "health":
				goResult, goErr = service.Health(t.Context())
			case "allowed-actions":
				args = append(args, "--resource-type", scenario.Input.ResourceType, "--resource-ids", strings.Join(scenario.Input.ResourceIDs, ","))
				goResult, goErr = service.AllowedActions(t.Context(), scenario.Input.ResourceType, scenario.Input.ResourceIDs)
			case "list", "list-databases", "database-children":
				var options []sdk.CatalogListOption
				if scenario.Input.IncludeStats != nil {
					options = append(options, sdk.WithCatalogIncludeStats(*scenario.Input.IncludeStats))
					args = append(args, fmt.Sprintf("--include-stats=%t", *scenario.Input.IncludeStats))
				}
				if scenario.Input.PageSize != nil {
					options = append(options, sdk.WithCatalogPageSize(*scenario.Input.PageSize))
					args = append(args, "--page-size", fmt.Sprint(*scenario.Input.PageSize))
				}
				if scenario.Input.PageToken != nil {
					options = append(options, sdk.WithCatalogPageToken(*scenario.Input.PageToken))
					args = append(args, "--page-token", *scenario.Input.PageToken)
				}
				switch scenario.Operation {
				case "list":
					goResult, goErr = service.List(t.Context(), options...)
				case "list-databases":
					args = append(args, "--catalog-id", fmt.Sprint(scenario.Input.CatalogID))
					goResult, goErr = service.ListDatabases(t.Context(), scenario.Input.CatalogID, options...)
				case "database-children":
					args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
					goResult, goErr = service.DatabaseChildren(t.Context(), scenario.Input.DatabaseID, options...)
				}
			case "info":
				args = append(args, "--catalog-id", fmt.Sprint(scenario.Input.CatalogID))
				goResult, goErr = service.Info(t.Context(), scenario.Input.CatalogID)
			case "delete":
				args = append(args, "--catalog-id", fmt.Sprint(scenario.Input.CatalogID))
				goErr = service.Delete(t.Context(), scenario.Input.CatalogID)
			case "run-nl2sql":
				args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID), "--sql", scenario.Input.SQL)
				goResult, goErr = service.RunNL2SQL(t.Context(), scenario.Input.DatabaseID, scenario.Input.SQL)
			case "database-info":
				args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
				goResult, goErr = service.DatabaseInfo(t.Context(), scenario.Input.DatabaseID)
			case "delete-database":
				args = append(args, "--database-id", fmt.Sprint(scenario.Input.DatabaseID))
				goErr = service.DeleteDatabase(t.Context(), scenario.Input.DatabaseID)
			case "volume-info":
				args = append(args, "--volume-id", scenario.Input.VolumeKey)
				goResult, goErr = service.VolumeInfo(t.Context(), scenario.Input.VolumeKey)
			case "delete-volume":
				args = append(args, "--volume-id", fmt.Sprint(scenario.Input.VolumeID))
				goErr = service.DeleteVolume(t.Context(), scenario.Input.VolumeID)
			case "volume-full-path":
				args = append(args, "--volume-id", fmt.Sprint(scenario.Input.VolumeID))
				goResult, goErr = service.VolumeFullPath(t.Context(), scenario.Input.VolumeID)
			case "catalog-references", "database-references", "volume-references", "folder-references", "table-references":
				args = append(args, "--id", fmt.Sprint(scenario.Input.ReferenceID))
				switch scenario.Operation {
				case "catalog-references":
					goResult, goErr = service.CatalogReferences(t.Context(), scenario.Input.ReferenceID)
				case "database-references":
					goResult, goErr = service.DatabaseReferences(t.Context(), scenario.Input.ReferenceID)
				case "volume-references":
					goResult, goErr = service.VolumeReferences(t.Context(), scenario.Input.ReferenceID)
				case "folder-references":
					goResult, goErr = service.FolderReferences(t.Context(), scenario.Input.ReferenceID)
				case "table-references":
					goResult, goErr = service.TableReferences(t.Context(), scenario.Input.ReferenceID)
				}
			case "list-workspace-files":
				options := []sdk.QueryOption{sdk.WithPageSize(*scenario.Input.PageSize), sdk.WithPageToken(*scenario.Input.PageToken), sdk.WithQueryParam("filter_file_name_fuzzy", *scenario.Input.Name)}
				args = append(args, "--page-size", fmt.Sprint(*scenario.Input.PageSize), "--page-token", *scenario.Input.PageToken, "--name", *scenario.Input.Name)
				for _, id := range scenario.Input.FileIDs {
					options = append(options, sdk.WithQueryValues("filter_file_id", id))
				}
				for _, extension := range scenario.Input.Extensions {
					options = append(options, sdk.WithQueryValues("filter_file_ext", extension))
				}
				args = append(args, "--file-ids", strings.Join(scenario.Input.FileIDs, ","), "--extensions", strings.Join(scenario.Input.Extensions, ","))
				goResult, goErr = service.ListWorkspaceFiles(t.Context(), options...)
			case "workspace-file-model-url":
				args = append(args, "--file-id", scenario.Input.FileID)
				goResult, goErr = service.WorkspaceFileModelURL(t.Context(), scenario.Input.FileID)
			default:
				t.Fatalf("unknown operation %s", scenario.Operation)
			}
			if got := <-bodies; !reflect.DeepEqual(got, catalogFolderCLIJSON(scenario.Request)) {
				t.Fatalf("Go body=%#v, want=%#v", got, catalogFolderCLIJSON(scenario.Request))
			}
			goProductErr := catalogFolderCLIError(t, goErr, scenario)

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteCatalogCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if got := <-bodies; !reflect.DeepEqual(got, catalogFolderCLIJSON(scenario.Request)) {
				t.Fatalf("CLI body=%#v, want=%#v", got, catalogFolderCLIJSON(scenario.Request))
			}
			cliProductErr := catalogFolderCLIError(t, cliErr, scenario)
			if goProductErr != nil {
				if cliProductErr.Code != goProductErr.Code || cliProductErr.Message != goProductErr.Message || cliProductErr.Status != goProductErr.Status || cliProductErr.RequestID != goProductErr.RequestID {
					t.Fatalf("Go error=%+v, CLI error=%+v", goProductErr, cliProductErr)
				}
			} else if !catalogFolderCLIVoid(scenario.Operation) {
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
				expected := catalogFolderCLIResult(scenario.Operation)
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
			} else if expected := "catalog " + scenario.Operation + ": ok\n"; output != expected {
				t.Fatalf("CLI output=%q, want=%q", output, expected)
			}
		})
	}
}

func catalogFolderCLIVoid(operation string) bool {
	switch operation {
	case "update", "update-folder", "delete-folder", "clean-folder", "create-file", "delete-file", "delete-file-ref", "delete-table", "truncate-table", "delete", "delete-database", "delete-volume":
		return true
	default:
		return false
	}
}

func catalogFolderCLIResult(operation string) proto.Message {
	switch operation {
	case "create-folder":
		return &sdk.CreateFolderResult{}
	case "data-asset", "file-data-asset":
		return &sdk.CatalogDataAssetResolveResult{}
	case "batch-get-data-assets":
		return &sdk.CatalogBatchGetDataAssetsResult{}
	case "file-artifact":
		return &sdk.CatalogFileArtifactResult{}
	case "batch-resolve-file-artifacts":
		return &sdk.CatalogBatchResolveFileArtifactsResult{}
	case "create-table":
		return &sdk.CatalogCreateTableResult{}
	case "table-info":
		return &sdk.CatalogTableInfoResult{}
	case "table-data", "preview-table":
		return &sdk.CatalogTableDataResult{}
	case "files":
		return &sdk.CatalogFileListResult{}
	case "file-info":
		return &sdk.FileInfoResult{}
	case "trigger-file-workflow":
		return &sdk.TriggerFileWorkflowResult{}
	case "multi-table-info":
		return &sdk.CatalogMultiTableInfoResult{}
	case "table-exists":
		return &sdk.BoolDataResult{}
	case "table-full-path":
		return &sdk.PathResult{}
	case "table-overview":
		return &sdk.CatalogTableOverviewResult{}
	case "health":
		return &sdk.HealthResult{}
	case "allowed-actions":
		return &sdk.CatalogAllowedActionsResult{}
	case "list":
		return &sdk.CatalogListResult{}
	case "info":
		return &sdk.CatalogInfoResult{}
	case "run-nl2sql":
		return &sdk.CatalogNL2SQLRunResult{}
	case "list-databases":
		return &sdk.DatabaseListResult{}
	case "database-info":
		return &sdk.DatabaseInfoResult{}
	case "database-children":
		return &sdk.DatabaseChildrenResult{}
	case "volume-info":
		return &sdk.VolumeInfoResult{}
	case "volume-full-path":
		return &sdk.PathResult{}
	case "catalog-references", "database-references", "volume-references", "folder-references", "table-references":
		return &sdk.RefListResult{}
	case "list-workspace-files":
		return &sdk.WorkspaceFileListResult{}
	case "workspace-file-model-url":
		return &sdk.URLResult{}
	default:
		return &sdk.IDResult{}
	}
}
