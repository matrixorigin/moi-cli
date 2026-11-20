package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type iamWriteCLIContract struct {
	Name      string `json:"name"`
	Operation string `json:"operation"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Input     struct {
		UserID          string          `json:"user_id"`
		RoleID          string          `json:"role_id"`
		ResourceType    string          `json:"resource_type"`
		ResourceID      string          `json:"resource_id"`
		RepairOperation string          `json:"repair_operation"`
		Request         json.RawMessage `json:"request"`
	} `json:"input"`
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
	Status   int             `json:"status"`
}

func iamWriteCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func TestIAMWriteCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/iam_write_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []iamWriteCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var received any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || r.URL.RawQuery != "" || r.Header.Get("X-API-Key") != "user-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
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
			args := []string{scenario.Operation, "--json", string(scenario.Input.Request)}
			input := scenario.Input
			if input.UserID != "" {
				args = append(args, "--user-id", input.UserID)
			}
			if input.RoleID != "" {
				args = append(args, "--role-id", input.RoleID)
			}
			if input.ResourceType != "" {
				args = append(args, "--resource-type", input.ResourceType, "--resource-id", input.ResourceID)
			}
			if input.RepairOperation != "" {
				args = append(args, "--repair-operation", input.RepairOperation)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteIAMCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			assertIAMWriteCLIOutput(t, scenario, output)
			if !reflect.DeepEqual(received, iamWriteCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, iamWriteCLIJSON(scenario.Request))
			}
		})
	}
}

func assertIAMWriteCLIOutput(t *testing.T, scenario iamWriteCLIContract, output string) {
	t.Helper()
	var got proto.Message
	switch scenario.Operation {
	case "put-user-role-bindings":
		got = &sdk.IAMUserRoleBindingsMutationResult{}
	case "create-role", "update-role":
		got = &sdk.IAMRoleCreateResult{}
	case "validate-object-grants", "apply-object-grants":
		got = &sdk.IAMObjectGrantsMutationResult{}
	case "validate-data-privileges":
		got = &sdk.IAMDataPrivilegeValidationResult{}
	case "apply-data-privileges", "repair-data-privileges":
		got = &sdk.IAMDataPrivilegeOperationResult{}
	case "add-role-inheritance", "remove-role-inheritance":
		got = &sdk.IAMRoleInheritanceMutationResult{}
	case "validate-application-policy":
		got = &sdk.IAMApplicationPolicyValidationResult{}
	case "apply-application-policy":
		got = &sdk.IAMApplicationPolicyApplyResult{}
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	if err := json.Unmarshal([]byte(output), got); err != nil {
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
