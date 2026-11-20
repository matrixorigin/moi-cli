package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestCatalogCLIOptionsAlignWithGoSDK(t *testing.T) {
	var cases []struct {
		Name, Command, Path string
		Input               json.RawMessage
		Request             json.RawMessage
		Response            json.RawMessage
	}
	value, err := os.ReadFile("../testdata/product-contracts/catalog_navigation_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(value, &cases); err != nil {
		t.Fatal(err)
	}
	decode := func(value []byte) any {
		t.Helper()
		var result any
		decoder := json.NewDecoder(bytes.NewReader(value))
		decoder.UseNumber()
		if err := decoder.Decode(&result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var input struct {
				PageSize           *int      `json:"page_size"`
				PageToken          *string   `json:"page_token"`
				Search             *string   `json:"search"`
				ResourceTypes      *[]string `json:"resource_types"`
				ResourceID         *int64    `json:"resource_id"`
				IncludeTableLeaves *bool     `json:"include_table_leaves"`
				DatabaseName       *string   `json:"database_name"`
				TableName          *string   `json:"table_name"`
				DatabaseID         int       `json:"database_id"`
			}
			if err := json.Unmarshal(scenario.Input, &input); err != nil {
				t.Fatal(err)
			}
			requests := make(chan any, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.RequestURI() != "/newmoi"+scenario.Path || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" {
					t.Errorf("unexpected scoped request: %s %s", r.Method, r.URL.RequestURI())
				}
				var body any
				decoder := json.NewDecoder(r.Body)
				decoder.UseNumber()
				if err := decoder.Decode(&body); err != nil {
					t.Error(err)
				}
				requests <- body
				_, _ = w.Write(append(append([]byte(`{"code":"OK","data":`), scenario.Response...), '}'))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			catalog := client.Catalog("workspace-1")
			var result any
			switch scenario.Command {
			case "tree":
				var opts []sdk.CatalogTreeOption
				if input.PageSize != nil {
					opts = append(opts, sdk.WithCatalogTreePageSize(*input.PageSize))
				}
				if input.PageToken != nil {
					opts = append(opts, sdk.WithCatalogTreePageToken(*input.PageToken))
				}
				if input.Search != nil {
					opts = append(opts, sdk.WithCatalogTreeSearch(*input.Search))
				}
				if input.ResourceTypes != nil {
					opts = append(opts, sdk.WithCatalogTreeResourceTypes(*input.ResourceTypes...))
				}
				if input.ResourceID != nil {
					opts = append(opts, sdk.WithCatalogTreeResourceID(*input.ResourceID))
				}
				if input.IncludeTableLeaves != nil {
					opts = append(opts, sdk.WithCatalogTreeTableLeaves(*input.IncludeTableLeaves))
				}
				result, err = catalog.Tree(t.Context(), opts...)
			case "system-tree":
				var opts []sdk.SystemCatalogTreeOption
				if input.PageSize != nil {
					opts = append(opts, sdk.WithSystemCatalogPageSize(*input.PageSize))
				}
				if input.PageToken != nil {
					opts = append(opts, sdk.WithSystemCatalogPageToken(*input.PageToken))
				}
				if input.Search != nil {
					opts = append(opts, sdk.WithSystemCatalogSearch(*input.Search))
				}
				if input.IncludeTableLeaves != nil {
					opts = append(opts, sdk.WithSystemCatalogTableLeaves(*input.IncludeTableLeaves))
				}
				if input.DatabaseName != nil {
					opts = append(opts, sdk.WithSystemCatalogDatabase(*input.DatabaseName))
				}
				if input.TableName != nil {
					opts = append(opts, sdk.WithSystemCatalogTableName(*input.TableName))
				}
				result, err = catalog.SystemTree(t.Context(), opts...)
			case "database-table-leaves":
				var opts []sdk.DatabaseTableLeavesOption
				if input.PageSize != nil {
					opts = append(opts, sdk.WithDatabaseTableLeavesPageSize(*input.PageSize))
				}
				if input.PageToken != nil {
					opts = append(opts, sdk.WithDatabaseTableLeavesPageToken(*input.PageToken))
				}
				if input.Search != nil {
					opts = append(opts, sdk.WithDatabaseTableLeavesSearch(*input.Search))
				}
				result, err = catalog.DatabaseTableLeaves(t.Context(), input.DatabaseID, opts...)
			case "search-data-dashboard-tables":
				var opts []sdk.DataDashboardTableSearchOption
				if input.PageSize != nil {
					opts = append(opts, sdk.WithDataDashboardTableSearchPageSize(*input.PageSize))
				}
				if input.PageToken != nil {
					opts = append(opts, sdk.WithDataDashboardTableSearchPageToken(*input.PageToken))
				}
				if input.Search != nil {
					opts = append(opts, sdk.WithDataDashboardTableSearchQuery(*input.Search))
				}
				result, err = catalog.SearchDataDashboardTables(t.Context(), opts...)
			default:
				t.Fatalf("unknown fixture operation: %s", scenario.Command)
			}
			if err != nil {
				t.Fatal(err)
			}
			args := []string{scenario.Command}
			for name, value := range decode(scenario.Input).(map[string]any) {
				flag := "--" + strings.ReplaceAll(name, "_", "-")
				switch value := value.(type) {
				case bool:
					args = append(args, fmt.Sprintf("%s=%t", flag, value))
				case []any:
					values := make([]string, 0, len(value))
					for _, item := range value {
						values = append(values, item.(string))
					}
					args = append(args, flag, strings.Join(values, ","))
				default:
					args = append(args, flag, fmt.Sprint(value))
				}
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteCatalogCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			goBody, cliBody := <-requests, <-requests
			if !reflect.DeepEqual(goBody, decode(scenario.Request)) || !reflect.DeepEqual(cliBody, goBody) {
				t.Fatalf("Go=%v, CLI=%v, expected=%s", goBody, cliBody, scenario.Request)
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decode([]byte(output)), decode(encoded)) {
				t.Fatalf("CLI output=%s, Go SDK output=%s", output, encoded)
			}
		})
	}
}
