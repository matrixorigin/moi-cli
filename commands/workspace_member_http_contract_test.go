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
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type workspaceMemberCLIContract struct {
	Name      string                       `json:"name"`
	Operation string                       `json:"operation"`
	Method    string                       `json:"method"`
	Path      string                       `json:"path"`
	Query     map[string]string            `json:"query"`
	Input     workspaceMemberContractInput `json:"input"`
	Request   json.RawMessage              `json:"request"`
	Response  json.RawMessage              `json:"response"`
	Status    int                          `json:"status"`
}

type workspaceMemberContractFilter struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type workspaceMemberContractTag struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

type workspaceMemberContractInput struct {
	InvitationID       string                          `json:"invitation_id"`
	UserID             string                          `json:"user_id"`
	Name               string                          `json:"name"`
	Description        string                          `json:"description"`
	DescriptionPresent bool                            `json:"description_present"`
	Action             string                          `json:"action"`
	Email              string                          `json:"email"`
	EmailPresent       bool                            `json:"email_present"`
	DefaultRoleID      int                             `json:"default_role_id"`
	RoleIDs            []int                           `json:"role_ids"`
	AutoAccept         bool                            `json:"auto_accept"`
	TagID              int                             `json:"tag_id"`
	Tags               []workspaceMemberContractTag    `json:"tags"`
	Page               int                             `json:"page"`
	PagePresent        bool                            `json:"page_present"`
	PageSize           int                             `json:"page_size"`
	PageSizePresent    bool                            `json:"page_size_present"`
	Filters            []workspaceMemberContractFilter `json:"filters"`
}

func workspaceMemberCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func TestWorkspaceMemberCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/workspace_member_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []workspaceMemberCLIContract
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
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteWorkspaceMemberCommand(NewContext(client, "workspace-1", false, true, true), workspaceMemberCLIArgs(scenario))
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			assertWorkspaceMemberCLIOutput(t, scenario, output)
			if !reflect.DeepEqual(received, workspaceMemberCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, workspaceMemberCLIJSON(scenario.Request))
			}
		})
	}
}

func workspaceMemberCLIArgs(scenario workspaceMemberCLIContract) []string {
	input := scenario.Input
	args := []string{scenario.Operation}
	switch scenario.Operation {
	case "accept-invitation":
		args = append(args, "--invitation-id", input.InvitationID)
	case "members":
		if input.PagePresent {
			args = append(args, "--page", strconv.Itoa(input.Page))
		}
		if input.PageSizePresent {
			args = append(args, "--page-size", strconv.Itoa(input.PageSize))
		}
		for _, filter := range input.Filters {
			args = append(args, "--filter", filter.Name+"="+strings.Join(filter.Values, ","))
		}
	case "invite":
		args = append(args, "--name", input.Name, "--default-role-id", strconv.Itoa(input.DefaultRoleID))
		if input.EmailPresent {
			args = append(args, "--email", input.Email)
		}
		if input.DescriptionPresent {
			args = append(args, "--description", input.Description)
		}
		if input.RoleIDs != nil {
			values := make([]string, 0, len(input.RoleIDs))
			for _, id := range input.RoleIDs {
				values = append(values, strconv.Itoa(id))
			}
			args = append(args, "--role-ids", strings.Join(values, ","))
		}
		if input.AutoAccept {
			args = append(args, "--auto-accept")
		}
		for _, tag := range input.Tags {
			args = append(args, "--tag", fmt.Sprintf("%d=%s", tag.ID, tag.Value))
		}
	case "member":
		args = append(args, "--user-id", input.UserID)
	case "update":
		args = append(args, "--user-id", input.UserID, "--name", input.Name, "--description", input.Description)
		for _, tag := range input.Tags {
			args = append(args, "--tag", fmt.Sprintf("%d=%s", tag.ID, tag.Value))
		}
	case "delete":
		args = append(args, "--user-id", input.UserID)
	case "update-status":
		args = append(args, "--user-id", input.UserID, "--action", input.Action)
	case "create-user-tag":
		args = append(args, "--name", input.Name)
	case "update-user-tag":
		args = append(args, "--tag-id", strconv.Itoa(input.TagID), "--name", input.Name)
	case "delete-user-tag":
		args = append(args, "--tag-id", strconv.Itoa(input.TagID))
	case "audit-logs":
		keys := make([]string, 0, len(scenario.Query))
		for key := range scenario.Query {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			args = append(args, "--"+strings.ReplaceAll(key, "_", "-"), scenario.Query[key])
		}
	}
	return args
}

func assertWorkspaceMemberCLIOutput(t *testing.T, scenario workspaceMemberCLIContract, output string) {
	t.Helper()
	switch scenario.Operation {
	case "update", "delete", "update-status":
		expected := "workspace-member " + scenario.Operation + ": ok\n"
		if output != expected {
			t.Fatalf("output=%q want=%q", output, expected)
		}
		return
	}
	var got proto.Message
	switch scenario.Operation {
	case "accept-invitation":
		got = &sdk.WorkspaceInvitationAcceptResult{}
	case "current-principal":
		got = &sdk.CurrentWorkspacePrincipalResult{}
	case "members":
		got = &sdk.WorkspaceMemberListResult{}
	case "invite":
		got = &sdk.WorkspaceMemberInviteResult{}
	case "member":
		got = &sdk.WorkspaceMember{}
	case "create-user-tag", "update-user-tag", "delete-user-tag":
		got = &sdk.IDResult{}
	case "user-tags":
		got = &sdk.DataTagListResult{}
	case "audit-logs":
		got = &sdk.AuditLogListResult{}
	case "audit-log-overview":
		got = &sdk.AuditLogOverviewResult{}
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
