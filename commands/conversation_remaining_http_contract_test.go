package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type conversationRemainingContract struct {
	Name            string              `json:"name"`
	Operation       string              `json:"operation"`
	Method          string              `json:"method"`
	Path            string              `json:"path"`
	Query           map[string][]string `json:"query"`
	Request         json.RawMessage     `json:"request"`
	Response        json.RawMessage     `json:"response"`
	ResponseHeaders map[string]string   `json:"response_headers"`
	Status          int                 `json:"status"`
	Resource        string              `json:"resource"`
	TailPath        string              `json:"tail_path"`
	SessionID       int64               `json:"session_id"`
	MessageID       int64               `json:"message_id"`
	Source          string              `json:"source"`
	TagName         string              `json:"tag_name"`
	Content         string              `json:"content"`
}

func conversationRemainingJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func conversationRemainingQuery(values map[string][]string) []sdk.QueryOption {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var options []sdk.QueryOption
	for _, key := range keys {
		for _, value := range values[key] {
			options = append(options, sdk.WithQueryValues(key, value))
		}
	}
	return options
}

func conversationRemainingArgs(scenario conversationRemainingContract) []string {
	if scenario.Operation == "resource-any" {
		return []string{"resource-any", "--method", scenario.Method, "--resource", scenario.Resource, "--tail-path", scenario.TailPath}
	}
	args := []string{scenario.Operation}
	switch scenario.Operation {
	case "get-message", "list-message-tags":
		args = append(args, "--message-id", strconv.FormatInt(scenario.MessageID, 10))
	case "remove-message-tag":
		args = append(args, "--message-id", strconv.FormatInt(scenario.MessageID, 10), "--source", scenario.Source, "--name", scenario.TagName)
	case "delete-session", "get-session", "list-messages", "latest-message-id", "list-session-tags":
		args = append(args, "--session-id", strconv.FormatInt(scenario.SessionID, 10))
	case "remove-session-tag":
		args = append(args, "--session-id", strconv.FormatInt(scenario.SessionID, 10), "--source", scenario.Source, "--name", scenario.TagName)
	case "append-modified-response", "modify-response":
		args = append(args, "--session-id", strconv.FormatInt(scenario.SessionID, 10), "--message-id", strconv.FormatInt(scenario.MessageID, 10), "--content", scenario.Content)
	case "delete-tag":
		args = append(args, "--source", scenario.Source, "--name", scenario.TagName)
	}
	keys := make([]string, 0, len(scenario.Query))
	for key := range scenario.Query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range scenario.Query[key] {
			args = append(args, "--query-param", key+"="+value)
		}
	}
	return args
}

func invokeConversationRemainingSDK(t *testing.T, client *sdk.Client, scenario conversationRemainingContract) (proto.Message, *sdk.ProductData, error) {
	t.Helper()
	service := client.Conversations("workspace-1")
	options := conversationRemainingQuery(scenario.Query)
	var result proto.Message
	var err error
	switch scenario.Operation {
	case "resource-any":
		raw, resourceErr := client.Agents("workspace-1").ResourceAny(t.Context(), scenario.Method, scenario.Resource, scenario.TailPath, nil, options...)
		return nil, raw, resourceErr
	case "get-message":
		result, err = service.GetMessage(t.Context(), scenario.MessageID)
	case "list-message-tags":
		result, err = service.ListMessageTags(t.Context(), scenario.MessageID)
	case "remove-message-tag":
		err = service.RemoveMessageTag(t.Context(), scenario.MessageID, scenario.Source, scenario.TagName)
	case "list-sessions":
		result, err = service.ListSessions(t.Context(), options...)
	case "delete-session":
		err = service.DeleteSession(t.Context(), scenario.SessionID)
	case "get-session":
		result, err = service.GetSession(t.Context(), scenario.SessionID)
	case "list-messages":
		result, err = service.ListMessages(t.Context(), scenario.SessionID, options...)
	case "latest-message-id":
		result, err = service.LatestMessageID(t.Context(), scenario.SessionID)
	case "append-modified-response":
		err = service.AppendModifiedResponse(t.Context(), scenario.SessionID, scenario.MessageID, scenario.Content)
	case "modify-response":
		err = service.ModifyResponse(t.Context(), scenario.SessionID, scenario.MessageID, scenario.Content)
	case "list-session-tags":
		result, err = service.ListSessionTags(t.Context(), scenario.SessionID)
	case "remove-session-tag":
		err = service.RemoveSessionTag(t.Context(), scenario.SessionID, scenario.Source, scenario.TagName)
	case "list-tags":
		result, err = service.ListTags(t.Context(), options...)
	case "delete-tag":
		err = service.DeleteTag(t.Context(), scenario.Source, scenario.TagName)
	default:
		t.Fatalf("unknown operation %q", scenario.Operation)
	}
	return result, nil, err
}

func assertConversationRemainingError(t *testing.T, err error, scenario conversationRemainingContract) {
	t.Helper()
	if scenario.Status < 400 {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("error=%T %v", err, err)
	}
	var envelope struct {
		Code    any    `json:"code"`
		Message string `json:"msg"`
	}
	decoder := json.NewDecoder(bytes.NewReader(scenario.Response))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	if failure.Code != fmt.Sprint(envelope.Code) || failure.Message != envelope.Message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v", failure)
	}
}

func assertConversationRemainingResult(t *testing.T, scenario conversationRemainingContract, result proto.Message, raw *sdk.ProductData, output []byte) {
	t.Helper()
	if scenario.Status >= 400 {
		if len(output) != 0 {
			t.Fatalf("error output=%q", output)
		}
		return
	}
	if result != nil {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
			t.Fatal(err)
		}
		expected := result.ProtoReflect().New().Interface()
		if err := protojson.Unmarshal(envelope.Data, expected); err != nil {
			t.Fatal(err)
		}
		if !proto.Equal(result, expected) {
			t.Fatalf("SDK result=%v want=%v", result, expected)
		}
		actual := result.ProtoReflect().New().Interface()
		if err := protojson.Unmarshal(output, actual); err != nil {
			t.Fatalf("CLI output=%q error=%v", output, err)
		}
		if !proto.Equal(actual, expected) {
			t.Fatalf("CLI result=%v want=%v", actual, expected)
		}
		return
	}
	if raw != nil {
		var envelope struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
			t.Fatal(err)
		}
		expected := conversationRemainingJSON(envelope.Data)
		if !reflect.DeepEqual(raw.Value, expected) || !reflect.DeepEqual(conversationRemainingJSON(output), expected) {
			t.Fatalf("raw=%#v output=%s want=%#v", raw.Value, output, expected)
		}
		return
	}
	if len(output) != 0 {
		t.Fatalf("void output=%q", output)
	}
}

func TestConversationRemainingSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/conversation_remaining_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []conversationRemainingContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, values := range scenario.Query {
					for _, value := range values {
						expectedQuery.Add(key, value)
					}
				}
				raw, _ := io.ReadAll(r.Body)
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || !reflect.DeepEqual(conversationRemainingJSON(raw), conversationRemainingJSON(scenario.Request)) {
					t.Errorf("request=%s %s query=%v body=%s headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), raw, r.Header)
				}
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				if scenario.Status != http.StatusNoContent {
					_, _ = w.Write(scenario.Response)
				}
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			result, raw, sdkErr := invokeConversationRemainingSDK(t, client, scenario)
			assertConversationRemainingError(t, sdkErr, scenario)
			args := conversationRemainingArgs(scenario)
			var cliErr error
			output := captureStdout(t, func() {
				if scenario.Operation == "resource-any" {
					cliErr = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), args)
				} else {
					cliErr = ExecuteConversationCommand(NewContext(client, "workspace-1", false, true, true), args)
				}
			})
			assertConversationRemainingError(t, cliErr, scenario)
			assertConversationRemainingResult(t, scenario, result, raw, []byte(output))
		})
	}
}
