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
	"sort"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type dataShareCLIContract struct {
	Name      string                 `json:"name"`
	Operation string                 `json:"operation"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Query     map[string]string      `json:"query"`
	Input     dataShareContractInput `json:"input"`
	Request   json.RawMessage        `json:"request"`
	Response  json.RawMessage        `json:"response"`
	Status    int                    `json:"status"`
}

type dataShareContractInput struct {
	IDs                       []string            `json:"ids"`
	SourceDatabaseID          string              `json:"source_database_id"`
	PublishID                 string              `json:"publish_id"`
	SubscriptionID            string              `json:"subscription_id"`
	Name                      string              `json:"name"`
	SubName                   string              `json:"sub_name"`
	TableScope                *sdk.DataShareScope `json:"table_scope"`
	TableScopePresent         bool                `json:"table_scope_present"`
	TargetWorkspaceIDs        []string            `json:"target_workspace_ids"`
	TargetWorkspaceIDsPresent bool                `json:"target_workspace_ids_present"`
	Remark                    string              `json:"remark"`
	RemarkPresent             bool                `json:"remark_present"`
}

func dataShareCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func executeDataShareCLI(ctx *Context, args []string) ([]byte, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteDataShareCommand(ctx, args)
	os.Stdout = original
	closeErr := write.Close()
	output, readErr := io.ReadAll(read)
	readCloseErr := read.Close()
	return output, errors.Join(callErr, closeErr, readErr, readCloseErr)
}

func TestDataShareCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/data_share_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []dataShareCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var received any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, value := range scenario.Query {
					expectedQuery.Set(key, value)
				}
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "user-key" {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				if len(raw) > 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					_ = decoder.Decode(&received)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
			if err != nil {
				t.Fatal(err)
			}
			ctx := NewContext(client, "workspace-1", false, true, true)
			output, callErr := executeDataShareCLI(ctx, dataShareCLIArgs(scenario))
			if scenario.Status >= 400 {
				var envelope struct {
					Code string `json:"code"`
					Msg  string `json:"msg"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				var failure *sdk.Error
				if !errors.As(callErr, &failure) || failure.Code != envelope.Code || failure.Message != envelope.Msg || failure.Status != scenario.Status {
					t.Fatalf("error=%#v", callErr)
				}
				if len(output) != 0 {
					t.Fatalf("error output=%s", output)
				}
			} else if callErr != nil {
				t.Fatal(callErr)
			} else {
				assertDataShareCLIOutput(t, scenario, output)
			}
			if !reflect.DeepEqual(received, dataShareCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, dataShareCLIJSON(scenario.Request))
			}
		})
	}
}

func dataShareCLIArgs(scenario dataShareCLIContract) []string {
	input := scenario.Input
	args := []string{}
	appendQuery := func() {
		keys := make([]string, 0, len(scenario.Query))
		for key := range scenario.Query {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			name := key
			if key == "page_size" {
				name = "page-size"
			}
			args = append(args, "--"+name, scenario.Query[key])
		}
	}
	switch scenario.Operation {
	case "publishes", "subscriptions", "check-publish-name":
		args = []string{scenario.Operation}
		appendQuery()
	case "workspace-names":
		args = []string{"workspace-names-result", "--ids", strings.Join(input.IDs, ",")}
	case "resolve-workspace-names":
		args = []string{"resolve-workspace-names", "--ids", strings.Join(input.IDs, ",")}
	case "publish-summary":
		args = []string{"publish-summary"}
	case "source-tables":
		args = []string{"source-tables", "--source-database-id", input.SourceDatabaseID}
		appendQuery()
	case "create-publish":
		args = []string{"create-publish", "--name", input.Name, "--source-database-id", input.SourceDatabaseID, "--table-scope-mode", input.TableScope.GetMode(), "--table-ids", strings.Join(input.TableScope.GetObjectIds(), ","), "--target-workspace-ids", strings.Join(input.TargetWorkspaceIDs, ",")}
		if input.RemarkPresent {
			args = append(args, "--remark", input.Remark)
		}
	case "update-publish":
		args = []string{"update-publish", "--publish-id", input.PublishID}
		if input.TableScopePresent {
			args = append(args, "--table-scope-mode", input.TableScope.GetMode(), "--table-ids", strings.Join(input.TableScope.GetObjectIds(), ","))
		}
		if input.TargetWorkspaceIDsPresent {
			args = append(args, "--target-workspace-ids", strings.Join(input.TargetWorkspaceIDs, ","))
		}
		if input.RemarkPresent {
			args = append(args, "--remark", input.Remark)
		}
	case "delete-publish":
		args = []string{"delete-publish", "--publish-id", input.PublishID}
	case "subscribe":
		args = []string{"subscribe", "--subscription-id", input.SubscriptionID, "--sub-name", input.SubName}
	case "unsubscribe":
		args = []string{"unsubscribe", "--subscription-id", input.SubscriptionID}
	}
	return args
}

func assertDataShareCLIOutput(t *testing.T, scenario dataShareCLIContract, output []byte) {
	t.Helper()
	if scenario.Operation == "delete-publish" || scenario.Operation == "unsubscribe" {
		expected := "data-share " + scenario.Operation + ": ok\n"
		if string(output) != expected {
			t.Fatalf("output=%q want=%q", output, expected)
		}
		return
	}
	var got proto.Message
	switch scenario.Operation {
	case "publishes":
		got = &sdk.DataSharePublishListResult{}
	case "workspace-names", "resolve-workspace-names":
		got = &sdk.DataShareWorkspaceNameListResult{}
	case "publish-summary":
		got = &sdk.PublishSummaryResult{}
	case "source-tables":
		got = &sdk.DataShareSourceTableListResult{}
	case "check-publish-name":
		got = &sdk.CheckNameResult{}
	case "create-publish", "update-publish":
		got = &sdk.DataSharePublish{}
	case "subscriptions":
		got = &sdk.DataShareSubscriptionListResult{}
	case "subscribe":
		got = &sdk.DataShareSubscription{}
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	if err := json.Unmarshal(output, got); err != nil {
		t.Fatalf("output=%s error=%v", output, err)
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	expected := got.ProtoReflect().New().Interface()
	if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(got, expected) {
		t.Fatalf("result=%v want=%v", got, expected)
	}
}
