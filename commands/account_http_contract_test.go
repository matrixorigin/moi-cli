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
	product "github.com/matrixorigin/moi-go-sdk/v5/product"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type accountCLIContract struct {
	Name           string            `json:"name"`
	Operation      string            `json:"operation"`
	Method         string            `json:"method"`
	Path           string            `json:"path"`
	Query          map[string]string `json:"query"`
	WorkspaceScope string            `json:"workspace_scope"`
	Input          struct {
		WorkspaceID   string          `json:"workspace_id"`
		ReturnTo      string          `json:"return_to"`
		State         string          `json:"state"`
		Source        string          `json:"source"`
		EventType     string          `json:"event_type"`
		Page          string          `json:"page"`
		PageSize      string          `json:"page_size"`
		BrowserLogin  json.RawMessage `json:"browser_login"`
		BrowserLogout json.RawMessage `json:"browser_logout"`
		Preferences   json.RawMessage `json:"preferences"`
	} `json:"input"`
	Request         json.RawMessage   `json:"request"`
	Response        json.RawMessage   `json:"response"`
	ResponseHeaders map[string]string `json:"response_headers"`
	Stream          bool              `json:"stream"`
	Status          int               `json:"status"`
}

func accountCLIJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func accountCLIEnvelopeData(raw json.RawMessage) any {
	var envelope struct {
		Data any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		panic(err)
	}
	return envelope.Data
}

func TestAccountCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/account_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []accountCLIContract
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
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-API-Key") != "user-key" || r.Header.Get("X-Workspace-ID") != scenario.WorkspaceScope || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				if scenario.Stream && r.Header.Get("Accept") != "*/*" {
					t.Errorf("stream Accept=%q", r.Header.Get("Accept"))
				}
				raw, _ := io.ReadAll(r.Body)
				if len(raw) > 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					_ = decoder.Decode(&received)
				}
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				if !scenario.Stream {
					_, _ = w.Write(scenario.Response)
				}
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
			if err != nil {
				t.Fatal(err)
			}
			args := []string{scenario.Operation}
			switch scenario.Operation {
			case "db-connection":
				args = append(args, "--workspace-id", scenario.Input.WorkspaceID)
			case "begin-browser-login":
				args = append(args, "--return-to", scenario.Input.ReturnTo)
			case "complete-browser-login":
				args = append(args, "--json", string(scenario.Input.BrowserLogin))
			case "complete-browser-logout":
				args = append(args, "--state", scenario.Input.State)
			case "logout-browser-session":
				args = append(args, "--json", string(scenario.Input.BrowserLogout))
			case "update-preferences":
				args = append(args, "--json", string(scenario.Input.Preferences))
			case "reward-center-events":
				args = append(args, "--event-type", scenario.Input.EventType)
			case "reward-center-invitations":
				args = append(args, "--page", scenario.Input.Page, "--page-size", scenario.Input.PageSize)
			case "checkin":
				args = append(args, "--source", scenario.Input.Source)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteAccountCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			assertAccountCLIOutput(t, scenario, output)
			if !reflect.DeepEqual(received, accountCLIJSON(scenario.Request)) {
				t.Fatalf("body=%#v want=%#v", received, accountCLIJSON(scenario.Request))
			}
		})
	}
}

func assertAccountCLIOutput(t *testing.T, scenario accountCLIContract, output string) {
	t.Helper()
	if scenario.Stream {
		var got struct {
			Status   int    `json:"status"`
			Location string `json:"location"`
		}
		if err := json.Unmarshal([]byte(output), &got); err != nil {
			t.Fatal(err)
		}
		if got.Status != scenario.Status || got.Location != scenario.ResponseHeaders["Location"] {
			t.Fatalf("stream output=%+v", got)
		}
		return
	}
	var got proto.Message
	switch scenario.Operation {
	case "current-credit":
		got = &product.CurrentCreditResult{}
	case "db-connection":
		got = &sdk.DBConnectionResult{}
	case "browser-session":
		got = &sdk.BrowserSessionResult{}
	case "logout-browser-session":
		got = &sdk.BrowserLogoutResult{}
	case "preferences", "update-preferences":
		got = &sdk.AuthPreferencesResult{}
	}
	if got == nil {
		var raw any
		if err := json.Unmarshal([]byte(output), &raw); err != nil {
			t.Fatal(err)
		}
		if expected := accountCLIEnvelopeData(scenario.Response); !reflect.DeepEqual(raw, expected) {
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
