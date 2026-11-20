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
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type agentSkillConfirmationContract struct {
	Name      string              `json:"name"`
	Operation string              `json:"operation"`
	Method    string              `json:"method"`
	Path      string              `json:"path"`
	Query     map[string][]string `json:"query"`
	Input     struct {
		TaskID       string          `json:"task_id"`
		CallID       string          `json:"call_id"`
		SourceDigest string          `json:"source_digest"`
		Body         json.RawMessage `json:"body"`
	} `json:"input"`
	Request         json.RawMessage   `json:"request"`
	Response        json.RawMessage   `json:"response"`
	ResponseHeaders map[string]string `json:"response_headers"`
	Status          int               `json:"status"`
}

func TestAgentSkillConfirmationSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/agent_skill_confirmation_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []agentSkillConfirmationContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			bodies := make(chan any, 2)
			expectedQuery := url.Values{}
			for key, values := range scenario.Query {
				for _, value := range values {
					expectedQuery.Add(key, value)
				}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace 1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				raw, _ := io.ReadAll(r.Body)
				var body any
				if len(raw) > 0 {
					decoder := json.NewDecoder(bytes.NewReader(raw))
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil {
						t.Errorf("decode body: %v", err)
					}
				}
				bodies <- body
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			direct, directErr := callAgentSkillConfirmationGo(t, client, scenario)
			if got := <-bodies; !reflect.DeepEqual(got, agentSkillConfirmationJSON(scenario.Request)) {
				t.Fatalf("Go body=%#v want=%#v", got, agentSkillConfirmationJSON(scenario.Request))
			}
			assertAgentSkillConfirmationError(t, directErr, scenario)

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteAgentCommand(NewContext(client, "workspace 1", false, true, true), agentSkillConfirmationCLIArgs(scenario))
			})
			if got := <-bodies; !reflect.DeepEqual(got, agentSkillConfirmationJSON(scenario.Request)) {
				t.Fatalf("CLI body=%#v want=%#v", got, agentSkillConfirmationJSON(scenario.Request))
			}
			assertAgentSkillConfirmationError(t, cliErr, scenario)
			if scenario.Status >= 400 {
				if output != "" {
					t.Fatalf("CLI error output=%q", output)
				}
				return
			}

			expected := agentSkillConfirmationExpected(t, scenario)
			if !proto.Equal(direct, expected) {
				t.Fatalf("Go result=%v want=%v", direct, expected)
			}
			actual := expected.ProtoReflect().New().Interface()
			if err := protojson.Unmarshal([]byte(output), actual); err != nil {
				t.Fatalf("CLI output=%s error=%v", output, err)
			}
			if !proto.Equal(actual, expected) {
				t.Fatalf("CLI result=%v want=%v", actual, expected)
			}
		})
	}
}

func callAgentSkillConfirmationGo(t *testing.T, client *sdk.Client, scenario agentSkillConfirmationContract) (proto.Message, error) {
	t.Helper()
	service := client.Agents("workspace 1")
	switch scenario.Operation {
	case "get-skill-candidate-confirmation":
		return service.GetSkillCandidateConfirmation(t.Context(), scenario.Input.TaskID, scenario.Input.CallID, scenario.Input.SourceDigest)
	case "confirm-skill-candidate":
		request := &sdk.ConfirmSkillCandidateRequest{}
		if err := protojson.Unmarshal(scenario.Input.Body, request); err != nil {
			return nil, err
		}
		return service.ConfirmSkillCandidate(t.Context(), scenario.Input.TaskID, request)
	default:
		return nil, fmt.Errorf("unknown operation %q", scenario.Operation)
	}
}

func agentSkillConfirmationCLIArgs(scenario agentSkillConfirmationContract) []string {
	args := []string{scenario.Operation, "--task-id", scenario.Input.TaskID}
	if scenario.Operation == "get-skill-candidate-confirmation" {
		return append(args, "--call-id", scenario.Input.CallID, "--source-digest", scenario.Input.SourceDigest)
	}
	return append(args, "--json", string(scenario.Input.Body))
}

func agentSkillConfirmationExpected(t *testing.T, scenario agentSkillConfirmationContract) proto.Message {
	t.Helper()
	var result proto.Message
	if scenario.Operation == "get-skill-candidate-confirmation" {
		result = &sdk.SkillCandidateConfirmationResult{}
	} else {
		result = &sdk.ConfirmSkillCandidateResult{}
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if err := protojson.Unmarshal(envelope.Data, result); err != nil {
		t.Fatal(err)
	}
	return result
}

func agentSkillConfirmationJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		panic(err)
	}
	return value
}

func assertAgentSkillConfirmationError(t *testing.T, err error, scenario agentSkillConfirmationContract) {
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
		Code    any             `json:"code"`
		Msg     string          `json:"msg"`
		Message string          `json:"message"`
		Details json.RawMessage `json:"details"`
	}
	decoder := json.NewDecoder(bytes.NewReader(scenario.Response))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		t.Fatal(err)
	}
	message := envelope.Msg
	if message == "" {
		message = envelope.Message
	}
	if failure.Code != fmt.Sprint(envelope.Code) || failure.Message != message || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v want code=%v message=%q status=%d request=%q", failure, envelope.Code, message, scenario.Status, scenario.ResponseHeaders["X-Request-ID"])
	}
	if len(envelope.Details) > 0 && !reflect.DeepEqual(map[string]any(failure.Details), agentSkillConfirmationJSON(envelope.Details)) {
		t.Fatalf("details=%#v want=%#v", failure.Details, agentSkillConfirmationJSON(envelope.Details))
	}
}
