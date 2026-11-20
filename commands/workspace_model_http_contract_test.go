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
	"strconv"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type workspaceModelCLIContract struct {
	Name        string          `json:"name"`
	Kind        string          `json:"kind"`
	Operation   string          `json:"operation"`
	Method      string          `json:"method"`
	Path        string          `json:"path"`
	ID          int64           `json:"id"`
	EndpointID  int64           `json:"endpoint_id"`
	Address     string          `json:"address"`
	StatusValue int32           `json:"status_value"`
	Input       json.RawMessage `json:"input"`
	Request     json.RawMessage `json:"request"`
	Response    json.RawMessage `json:"response"`
	Status      int             `json:"status"`
}

func workspaceModelCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func workspaceModelCLIExpected(t *testing.T, raw json.RawMessage, message proto.Message) {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	if err := protojson.Unmarshal(envelope.Data, message); err != nil {
		t.Fatal(err)
	}
}

func executeWorkspaceModelCLI(ctx *Context, args []string) ([]byte, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	original := os.Stdout
	os.Stdout = write
	callErr := ExecuteModelCommand(ctx, args)
	os.Stdout = original
	closeErr := write.Close()
	output, readErr := io.ReadAll(read)
	readCloseErr := read.Close()
	return output, errors.Join(callErr, closeErr, readErr, readCloseErr)
}

func TestWorkspaceModelCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/workspace_model_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []workspaceModelCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var received any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.Path != scenario.Path || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "user-key" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.Path, r.Header)
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
			args := workspaceModelCLIArgs(scenario)
			output, callErr := executeWorkspaceModelCLI(ctx, args)
			if scenario.Status >= 400 {
				var failure *sdk.Error
				if !errors.As(callErr, &failure) || failure.Code != "ErrConflict" || failure.Status != 409 || failure.Message != "Synthetic backend conflict" {
					t.Fatalf("error=%#v", callErr)
				}
				if len(output) != 0 {
					t.Fatalf("error output=%s", output)
				}
			} else if callErr != nil {
				t.Fatal(callErr)
			} else {
				assertWorkspaceModelCLIOutput(t, scenario, output)
			}
			if !reflect.DeepEqual(received, workspaceModelCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, workspaceModelCLIJSON(scenario.Request))
			}
		})
	}
}

func workspaceModelCLIArgs(scenario workspaceModelCLIContract) []string {
	kind := scenario.Kind
	backendID := strconv.FormatInt(scenario.ID, 10)
	args := []string{}
	switch scenario.Operation {
	case "create-backend", "update-router":
		command := "create-workspace-" + kind + "-backend"
		if scenario.Operation == "update-router" {
			command = "update-workspace-" + kind + "-router"
		}
		args = []string{command, "--json", string(scenario.Input)}
	case "list-backends":
		args = []string{"list-workspace-" + kind + "-backends"}
	case "get-backend":
		args = []string{"get-workspace-" + kind + "-backend", "--backend-id", backendID}
	case "update-backend":
		args = []string{"update-workspace-" + kind + "-backend", "--backend-id", backendID, "--json", string(scenario.Input)}
	case "delete-backend":
		args = []string{"delete-workspace-" + kind + "-backend", "--backend-id", backendID}
	case "create-endpoint":
		args = []string{"create-workspace-" + kind + "-endpoint", "--backend-id", backendID, "--address", scenario.Address}
	case "list-endpoints":
		args = []string{"list-workspace-" + kind + "-endpoints", "--backend-id", backendID}
	case "router":
		args = []string{"workspace-" + kind + "-router"}
	case "set-endpoint-status":
		args = []string{"set-workspace-" + kind + "-endpoint-status", "--backend-id", backendID, "--endpoint-id", strconv.FormatInt(scenario.EndpointID, 10), "--status", strconv.FormatInt(int64(scenario.StatusValue), 10)}
	case "list-models":
		args = []string{"list-workspace-llm-models"}
	}
	return args
}

func assertWorkspaceModelCLIOutput(t *testing.T, scenario workspaceModelCLIContract, output []byte) {
	t.Helper()
	if scenario.Operation == "delete-backend" || scenario.Operation == "set-endpoint-status" && scenario.Kind == "llm" {
		if len(output) != 0 {
			t.Fatalf("empty-result output=%s", output)
		}
		return
	}
	if scenario.Operation == "list-endpoints" {
		var got []*sdk.WorkspaceLLMEndpoint
		if err := json.Unmarshal(output, &got); err != nil {
			t.Fatalf("output=%s error=%v", output, err)
		}
		var raw struct {
			Data []json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(scenario.Response, &raw); err != nil {
			t.Fatal(err)
		}
		if len(got) != len(raw.Data) {
			t.Fatalf("output=%s", output)
		}
		for index := range got {
			expected := &sdk.WorkspaceLLMEndpoint{}
			if err := protojson.Unmarshal(raw.Data[index], expected); err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(got[index], expected) {
				t.Fatalf("endpoint=%v want=%v", got[index], expected)
			}
		}
		return
	}
	var got proto.Message
	switch scenario.Operation {
	case "create-backend", "get-backend", "update-backend":
		got = &sdk.WorkspaceLLMBackend{}
	case "list-backends":
		got = &sdk.WorkspaceLLMBackendList{}
	case "create-endpoint":
		got = &sdk.WorkspaceLLMEndpoint{}
	case "router", "update-router":
		got = &sdk.WorkspaceLLMRouterConfigResult{}
	case "set-endpoint-status":
		got = &sdk.WorkspaceEndpointMutationResult{}
	case "list-models":
		got = &sdk.ProviderModelListResult{}
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	if err := json.Unmarshal(output, got); err != nil {
		t.Fatalf("output=%s error=%v", output, err)
	}
	expected := got.ProtoReflect().New().Interface()
	workspaceModelCLIExpected(t, scenario.Response, expected)
	if !proto.Equal(got, expected) {
		t.Fatalf("result=%v want=%v", got, expected)
	}
	if mutation, ok := got.(*sdk.WorkspaceEndpointMutationResult); ok && !mutation.GetUpdated() {
		t.Fatalf("mutation=%v", mutation)
	}
}
