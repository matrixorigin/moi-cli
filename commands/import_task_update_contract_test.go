package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestImportTaskUpdateCLIContract(t *testing.T) {
	for _, command := range []string{"update-task", "update-task-from-spec"} {
		for _, scenario := range []struct {
			name string
			body string
			want string
		}{
			{"supported update", `{"Name":"renamed","LoadModeConfig":{"load_interval_type":7,"interval":30},"FileFilterConfig":{"file_types":[7]},"URIs":["s3://bucket/资料.pdf"]}`, `{"task_id":"task-1","name":"renamed","load_mode_config":{"load_interval_type":7,"interval":30},"file_filter_config":{"file_types":[7]},"uris":["s3://bucket/资料.pdf"]}`},
			{"empty update", `{}`, ""},
		} {
			t.Run(command+"/"+scenario.name, func(t *testing.T) {
				var calls atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					if r.Method != http.MethodPost || r.URL.Path != "/newmoi/task/update" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
						t.Errorf("unexpected target: %s %s", r.Method, r.URL.Path)
					}
					var got, want map[string]any
					if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
						t.Error(err)
					}
					if scenario.want != "" {
						if err := json.Unmarshal([]byte(scenario.want), &want); err != nil {
							t.Error(err)
						}
						if !reflect.DeepEqual(got, want) {
							t.Errorf("update body = %#v, want %#v", got, want)
						}
					}
					_, _ = w.Write([]byte(`{"code":"OK","data":null}`))
				}))
				t.Cleanup(server.Close)
				client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = client.Close() })
				var callErr error
				output := captureStdout(t, func() {
					callErr = ExecuteConnectorCommand(NewContext(client, "workspace-1", false, true, true), []string{command, "--task-id", "task-1", "--json", scenario.body})
				})
				if scenario.want == "" {
					if callErr == nil || calls.Load() != 0 || output != "" {
						t.Fatalf("invalid update reached owner: err=%v calls=%d output=%s", callErr, calls.Load(), output)
					}
				} else if callErr != nil || calls.Load() != 1 {
					t.Fatalf("valid update failed: err=%v calls=%d", callErr, calls.Load())
				}
			})
		}
	}
}
