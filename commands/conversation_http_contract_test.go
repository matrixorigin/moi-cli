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

func TestConversationCLIMatchesSharedHTTPContract(t *testing.T) {
	var cases []struct {
		Name, Operation, Method, Path string
		SessionID                     int64 `json:"session_id"`
		MessageID                     int64 `json:"message_id"`
		Input, Request, Response      json.RawMessage
		Status                        int
	}
	data, err := os.ReadFile("../testdata/product-contracts/conversation_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			decodeJSON := func(data []byte) any {
				t.Helper()
				var value any
				decoder := json.NewDecoder(bytes.NewReader(data))
				decoder.UseNumber()
				if err := decoder.Decode(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			received := make(chan any, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wantContentType := ""
				if len(scenario.Input) != 0 {
					wantContentType = "application/json"
				}
				if r.Method != scenario.Method || r.URL.RequestURI() != scenario.Path || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("Content-Type") != wantContentType {
					t.Errorf("unexpected scoped request: %s %s", r.Method, r.URL.RequestURI())
				}
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Error(err)
				}
				var body any
				if len(raw) != 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil {
						t.Error(err)
					}
				}
				received <- body
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				if scenario.Status != http.StatusNoContent {
					_, _ = w.Write(scenario.Response)
				}
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{scenario.Operation}
			if len(scenario.Input) != 0 {
				args = append(args, "--json", string(scenario.Input))
			}
			if scenario.SessionID != 0 {
				args = append(args, "--session-id", strconv.FormatInt(scenario.SessionID, 10))
			}
			if scenario.MessageID != 0 {
				args = append(args, "--message-id", strconv.FormatInt(scenario.MessageID, 10))
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteConversationCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			select {
			case body := <-received:
				if want := decodeJSON(scenario.Request); !reflect.DeepEqual(body, want) {
					t.Errorf("HTTP body=%v, want=%v", body, want)
				}
			default:
				t.Fatal("no HTTP request received")
			}
			if scenario.Status >= 400 {
				var response struct{ Code, Msg string }
				if err := json.Unmarshal(scenario.Response, &response); err != nil {
					t.Fatal(err)
				}
				var failure *sdk.Error
				if !errors.As(callErr, &failure) || failure.Code != response.Code || failure.Status != scenario.Status || failure.Message != response.Msg || output != "" {
					t.Fatalf("owner failure changed: %v, output=%s", callErr, output)
				}
				return
			}
			if callErr != nil {
				t.Fatal(callErr)
			}
			if scenario.Status == http.StatusNoContent {
				if output != "" {
					t.Fatalf("unexpected output for empty acknowledgement: %s", output)
				}
				return
			}
			var envelope struct{ Data json.RawMessage }
			if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			var value proto.Message
			switch scenario.Operation {
			case "create-ui-session", "update-ui-session", "create-session", "update-session", "get-ui-session":
				value = &sdk.ConversationSession{}
			case "delete-ui-session", "pin-ui-session", "unpin-ui-session":
				value = &sdk.ConversationMutation{}
			case "get-ui-message", "like-ui-message", "dislike-ui-message":
				value = &sdk.ConversationUIMessage{}
			case "latest-ui-message-id":
				value = &sdk.ConversationLatestMessageID{}
			case "list-ui-sessions", "list-pinned-ui-sessions", "list-unpinned-ui-sessions":
				value = &sdk.ConversationSessionList{}
			case "create-message":
				value = &sdk.ConversationWorkspaceMessage{}
			case "create-tag":
				value = &sdk.ConversationTag{}
			case "list-ui-messages":
			default:
				t.Fatalf("unsupported fixture operation %s", scenario.Operation)
			}
			var expected any
			if value != nil {
				if err := protojson.Unmarshal(envelope.Data, value); err != nil {
					t.Fatal(err)
				}
				expected = value
			} else {
				var items []json.RawMessage
				if err := json.Unmarshal(envelope.Data, &items); err != nil {
					t.Fatal(err)
				}
				messages := make([]*sdk.ConversationUIMessage, 0, len(items))
				for _, item := range items {
					message := &sdk.ConversationUIMessage{}
					if err := protojson.Unmarshal(item, message); err != nil {
						t.Fatal(err)
					}
					messages = append(messages, message)
				}
				expected = messages
			}
			encoded, err := json.Marshal(expected)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := decodeJSON([]byte(output)), decodeJSON(encoded); !reflect.DeepEqual(got, want) {
				t.Errorf("CLI response=%s, want=%s", output, encoded)
			}
		})
	}
}
