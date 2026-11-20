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
	"strconv"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type iamQueryCLIContract struct {
	Name      string            `json:"name"`
	Operation string            `json:"operation"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Query     map[string]string `json:"query"`
	Input     struct {
		RoleID                   string   `json:"role_id"`
		ExpectedLifecycleVersion int64    `json:"expected_lifecycle_version"`
		IncludeOwnership         bool     `json:"include_ownership"`
		OwnershipCursor          string   `json:"ownership_cursor"`
		OwnershipLimit           int      `json:"ownership_limit"`
		Cursor                   string   `json:"cursor"`
		Limit                    int      `json:"limit"`
		TransferRequested        bool     `json:"transfer_requested"`
		ResourceType             string   `json:"resource_type"`
		ActionID                 string   `json:"action_id"`
		Purpose                  string   `json:"purpose"`
		ResourceIDs              []string `json:"resource_ids"`
	} `json:"input"`
	Request  json.RawMessage `json:"request"`
	Response json.RawMessage `json:"response"`
	Status   int             `json:"status"`
}

func iamQueryCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func iamQueryCLIEnvelopeData(raw json.RawMessage) any {
	var envelope struct {
		Data any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		panic(err)
	}
	return envelope.Data
}

func TestIAMQueryCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/iam_query_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []iamQueryCLIContract
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
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteIAMCommand(NewContext(client, "workspace-1", false, true, true), iamQueryCLIArgs(scenario))
			})
			if scenario.Status >= 400 {
				var envelope struct {
					Code string `json:"code"`
					Msg  string `json:"msg"`
				}
				if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
					t.Fatal(err)
				}
				var failure *sdk.Error
				if !errors.As(callErr, &failure) || failure.Code != envelope.Code || failure.Message != envelope.Msg || failure.Status != scenario.Status || output != "" {
					t.Fatalf("error=%#v output=%q", callErr, output)
				}
			} else {
				if callErr != nil {
					t.Fatal(callErr)
				}
				assertIAMQueryCLIOutput(t, scenario, output)
			}
			if !reflect.DeepEqual(received, iamQueryCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, iamQueryCLIJSON(scenario.Request))
			}
		})
	}
}

func iamQueryCLIArgs(scenario iamQueryCLIContract) []string {
	input := scenario.Input
	args := []string{scenario.Operation}
	switch scenario.Operation {
	case "role-delete-impact":
		args = append(args, "--role-id", input.RoleID, "--expected-lifecycle-version", strconv.FormatInt(input.ExpectedLifecycleVersion, 10))
		if input.IncludeOwnership {
			args = append(args, "--include-ownership", "--ownership-limit", strconv.Itoa(input.OwnershipLimit))
			if input.OwnershipCursor != "" {
				args = append(args, "--ownership-cursor", input.OwnershipCursor)
			}
		}
		if input.TransferRequested {
			args = append(args, "--transfer-requested")
		}
	case "role-delete-ownerships":
		args = append(args, "--role-id", input.RoleID)
		if value, ok := scenario.Query["cursor"]; ok {
			args = append(args, "--cursor", value)
		}
		if value, ok := scenario.Query["limit"]; ok {
			args = append(args, "--limit", value)
		}
		if value, ok := scenario.Query["transfer_requested"]; ok {
			args = append(args, "--transfer-requested="+value)
		}
	case "role-members":
		args = append(args, "--role-id", input.RoleID)
		if value, ok := scenario.Query["cursor"]; ok {
			args = append(args, "--cursor", value)
		}
		if value, ok := scenario.Query["limit"]; ok {
			args = append(args, "--limit", value)
		}
	case "list-permission-resources":
		args = append(args, "--resource-type", input.ResourceType, "--action-id", input.ActionID, "--purpose", input.Purpose)
		for _, item := range []struct{ key, flag string }{{"parent_id", "parent-id"}, {"cursor", "cursor"}, {"q", "q"}, {"limit", "limit"}} {
			if value, ok := scenario.Query[item.key]; ok {
				args = append(args, "--"+item.flag, value)
			}
		}
	case "batch-describe-permission-resources":
		args = append(args, "--resource-type", input.ResourceType, "--action-id", input.ActionID, "--purpose", input.Purpose)
		for _, id := range input.ResourceIDs {
			args = append(args, "--resource-id", id)
		}
	}
	return args
}

func assertIAMQueryCLIOutput(t *testing.T, scenario iamQueryCLIContract, output string) {
	t.Helper()
	var got proto.Message
	switch scenario.Operation {
	case "role-delete-impact":
		got = &sdk.IAMRoleDeleteImpactResult{}
	case "role-delete-ownerships":
		got = &sdk.IAMRoleDeleteOwnershipList{}
	case "role-members":
		got = &sdk.IAMRoleMembersResult{}
	}
	if got == nil {
		var raw any
		if err := json.Unmarshal([]byte(output), &raw); err != nil {
			t.Fatal(err)
		}
		if expected := iamQueryCLIEnvelopeData(scenario.Response); !reflect.DeepEqual(raw, expected) {
			t.Fatalf("raw output=%#v want=%#v", raw, expected)
		}
		return
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
