package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type iamReadCLIContract struct {
	Name      string              `json:"name"`
	Operation string              `json:"operation"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Query     map[string][]string `json:"query"`
	Input     struct {
		UserID        string          `json:"user_id"`
		RoleID        string          `json:"role_id"`
		ResourceType  string          `json:"resource_type"`
		ResourceID    string          `json:"resource_id"`
		RequestID     string          `json:"request_id"`
		CapabilityIDs []string        `json:"capability_ids"`
		Request       json.RawMessage `json:"request"`
	} `json:"input"`
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
	Status   int             `json:"status"`
}

func iamReadCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func TestIAMReadCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/iam_read_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []iamReadCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var received any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, values := range scenario.Query {
					for _, value := range values {
						expectedQuery.Add(key, value)
					}
				}
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-API-Key") != "user-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
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
			args := []string{scenario.Operation}
			input := scenario.Input
			if len(input.Request) > 0 && string(input.Request) != "null" {
				args = append(args, "--json", string(input.Request))
			}
			if input.UserID != "" {
				args = append(args, "--user-id", input.UserID)
			}
			if input.RoleID != "" {
				args = append(args, "--role-id", input.RoleID)
			}
			if input.ResourceType != "" {
				args = append(args, "--resource-type", input.ResourceType, "--resource-id", input.ResourceID)
			}
			if input.RequestID != "" {
				args = append(args, "--request-id", input.RequestID)
			}
			for _, id := range input.CapabilityIDs {
				args = append(args, "--capability-id", id)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteIAMCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			assertIAMReadCLIOutput(t, scenario, output)
			if !reflect.DeepEqual(received, iamReadCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, iamReadCLIJSON(scenario.Request))
			}
		})
	}
}

func assertIAMReadCLIOutput(t *testing.T, scenario iamReadCLIContract, output string) {
	t.Helper()
	var got proto.Message
	switch scenario.Operation {
	case "permission-schema":
		got = &sdk.IAMPermissionSchemaResult{}
	case "roles":
		got = &sdk.IAMRoleListResult{}
	case "user-role-bindings":
		got = &sdk.IAMUserRoleBindingsResult{}
	case "role-permission-view":
		got = &sdk.IAMRolePermissionViewResult{}
	case "update-role-lifecycle", "delete-role":
		got = &sdk.IAMRoleMutationResult{}
	case "transfer-object-ownership":
		got = &sdk.IAMObjectOwnershipTransferResult{}
	case "object-permission-view":
		got = &sdk.IAMObjectPermissionViewResult{}
	case "data-privileges":
		got = &sdk.IAMDataPrivilegeViewResult{}
	case "data-privilege-operation":
		got = &sdk.IAMDataPrivilegeOperationResult{}
	case "current-principal-roles":
		got = &sdk.IAMCurrentPrincipalRolesResult{}
	case "current-principal-capabilities":
		got = &sdk.IAMCurrentPrincipalCapabilitiesResult{}
	case "role-inheritance", "role-inheritance-graph":
		got = &sdk.IAMRoleInheritanceResult{}
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
