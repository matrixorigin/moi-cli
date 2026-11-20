package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func TestComputeCLIRequestsMatchSharedSDKContract(t *testing.T) {
	var cases []struct {
		Name, Command, ID, Method, Path string
		Input                           map[string]json.RawMessage
		Query                           url.Values
		Request, Response               json.RawMessage
	}
	value, err := os.ReadFile("../testdata/product-contracts/compute_options_contract.json")
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
			wantBody := decode(scenario.Request)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != scenario.Method || r.URL.Path != "/newmoi"+scenario.Path || r.URL.Query().Encode() != scenario.Query.Encode() || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" {
					t.Errorf("unexpected scoped request: %s %s", r.Method, r.URL.RequestURI())
				}
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				if wantBody == nil {
					if len(raw) != 0 {
						t.Errorf("unexpected body: %s", raw)
					}
				} else {
					var got any
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					if err := decoder.Decode(&got); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(got, wantBody) {
						t.Errorf("request=%s, expected=%s", raw, scenario.Request)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write(append(append([]byte(`{"code":"OK","data":`), scenario.Response...), '}'))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{scenario.Command}
			if scenario.ID != "" {
				flag := "--id"
				if strings.HasSuffix(scenario.Command, "spec") {
					flag = "--spec-id"
				}
				args = append(args, flag, scenario.ID)
			}
			for name, raw := range scenario.Input {
				flag := "--" + strings.ReplaceAll(name, "_", "-")
				if name == "memory_mib" && scenario.Command == "create-spec" {
					flag = "--memory-mi-b"
				}
				if name == "memory_gib" && scenario.Command == "create" {
					flag = "--memory-gi-b"
				}
				if name == "worker_images" || name == "node_placement" {
					path := filepath.Join(t.TempDir(), name+".json")
					if err := os.WriteFile(path, raw, 0o600); err != nil {
						t.Fatal(err)
					}
					args = append(args, flag+"-file", path)
				} else if value, ok := decode(raw).(bool); ok {
					args = append(args, fmt.Sprintf("%s=%t", flag, value))
				} else {
					args = append(args, flag, fmt.Sprint(decode(raw)))
				}
			}
			for name, values := range scenario.Query {
				for _, value := range values {
					args = append(args, "--query-param", name+"="+value)
				}
			}
			var callErr error
			output := captureStdout(t, func() { callErr = ExecuteComputeCommand(NewContext(client, "workspace-1", false, true, true), args) })
			if callErr != nil {
				t.Fatal(callErr)
			}
			var result proto.Message
			switch scenario.Command {
			case "create-spec", "update-spec":
				result = &sdk.ComputeResourceSpec{}
			case "admin-list":
				result = &sdk.AdminComputeResourceListResult{}
			default:
				result = &sdk.ComputeResource{}
			}
			if err := protojson.Unmarshal([]byte(output), result); err != nil {
				t.Fatalf("decode CLI output: %v; output=%s", err, output)
			}
			want := result.ProtoReflect().Type().New().Interface()
			if err := protojson.Unmarshal(scenario.Response, want); err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(result, want) {
				t.Errorf("CLI result=%v, expected=%v", result, want)
			}
			if calls.Load() != 1 {
				t.Errorf("requests=%d; expected one operation", calls.Load())
			}
		})
	}
}
