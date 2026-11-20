package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/matrixorigin/matrixflow/sdk/cli/inputshape"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGeneratedInputTypesMatchSDK(t *testing.T) {
	for _, tc := range []struct {
		command string
		typ     reflect.Type
	}{
		{"connector create-task", reflect.TypeFor[sdk.ImportTaskCreateSpec]()},
		{"connector create-export-task", reflect.TypeFor[sdk.ExportTaskSpec]()},
		{"semantic-model create-with-sources", reflect.TypeFor[[]sdk.SemanticModelSourceInput]()},
		{"catalog upload-bytes", reflect.TypeFor[[]byte]()},
		{"model workspace-chat-completion", reflect.TypeFor[sdk.WorkspaceChatCompletionRequest]()},
		{"model workspace-chat-completion-stream", reflect.TypeFor[sdk.WorkspaceChatCompletionRequest]()},
		{"workflow invoke-workitem-api-service", reflect.TypeFor[sdk.WorkitemAPIServiceInvokeSpec]()},
		{"workflow invoke-dynamic-service", reflect.TypeFor[sdk.DynamicServiceInvokeSpec]()},
		{"workflow invoke-workflow-dynamic-service", reflect.TypeFor[sdk.DynamicServiceInvokeSpec]()},
	} {
		if got := commandJSONTypes[tc.command]["json"]; got != tc.typ {
			t.Errorf("%s=%v want=%v", tc.command, got, tc.typ)
		}
	}
	if len(commandJSONTypes) < 40 {
		t.Fatalf("unexpected loss of typed commands: %d", len(commandJSONTypes))
	}
	for command, inputs := range commandJSONTypes {
		for argument, typ := range inputs {
			t.Run(command+"/"+argument, func(t *testing.T) {
				if _, err := json.Marshal(inputshape.Describe(typ)); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}

func TestWorkflowServiceInvokeGuidance(t *testing.T) {
	dynamic := workflowServiceInvokeGuidance("workflow invoke-dynamic-service")
	example := dynamic["example"].(map[string]any)
	if example["service_name"] != "<returned-service-name>" {
		t.Fatalf("dynamic service example lost published identity: %#v", example)
	}
	payload, ok := example["payload"].(map[string]any)
	if !ok || payload["text"] != "hello" {
		t.Fatalf("dynamic service example must contain the business payload: %#v", example)
	}
	if _, found := example["type"]; found {
		t.Fatalf("dynamic service example acquired legacy type: %#v", example)
	}
	legacy := workflowServiceInvokeGuidance("workflow invoke-workitem-api-service")
	legacyExample := legacy["example"].(map[string]any)
	if legacyExample["type"] != "operator" || legacyExample["version"] != "<returned-version>" {
		t.Fatalf("legacy invocation guidance is not actionable: %#v", legacyExample)
	}
}

func TestSchemaWireTypes(t *testing.T) {
	type recursive struct {
		Next   *recursive `json:"next,omitempty"`
		Hidden string     `json:"-"`
		Items  []string   `json:"items"`
	}
	for _, tc := range []struct {
		typ    reflect.Type
		wanted string
	}{
		{reflect.TypeFor[recursive](), "object"}, {reflect.TypeFor[[]int](), "array"},
		{reflect.TypeFor[map[string]bool](), "object"}, {reflect.TypeFor[[]byte](), "string"},
		{reflect.TypeFor[int64](), "integer"}, {reflect.TypeFor[float64](), "number"},
		{reflect.TypeFor[bool](), "boolean"}, {reflect.TypeFor[*string](), "string"},
	} {
		schema := inputshape.Describe(tc.typ)
		if schema["type"] != tc.wanted {
			t.Errorf("%v=%v", tc.typ, schema)
		}
	}
	schema := inputshape.Describe(reflect.TypeFor[recursive]())
	props := schema["properties"].(map[string]any)
	if _, ok := props["Hidden"]; ok {
		t.Fatal("hidden field exposed")
	}
	if _, ok := schema["required"]; ok {
		t.Fatal("business required fields were inferred")
	}
	custom := inputshape.Describe(reflect.TypeFor[structpb.Value]())
	if _, ok := custom["properties"]; ok {
		t.Fatal("protobuf implementation fields exposed")
	}
}

func TestSchemaA2AAnonymousFields(t *testing.T) {
	body := `{"agent_code":"assistant","agent_id":"agent_1","agent_workspace_id":"ws_1","jsonrpc":"2.0","id":"request_1","method":"message/send","params":{"message":{"messageId":"message_1"}},"model":"test-model"}`
	var request sdk.AgentA2ARequest
	if err := json.Unmarshal([]byte(body), &request); err != nil {
		t.Fatal(err)
	}
	if request.AgentCode != "assistant" || request.Method != "message/send" || request.Model != "test-model" {
		t.Fatalf("flat A2A input did not decode: %+v", request)
	}
	for _, command := range []string{"agent a2a", "agent a2a-stream"} {
		t.Run(command, func(t *testing.T) {
			typ := commandJSONTypes[command]["json"]
			properties := inputshape.Describe(typ)["properties"].(map[string]any)
			var fields map[string]json.RawMessage
			if err := json.Unmarshal([]byte(body), &fields); err != nil {
				t.Fatal(err)
			}
			if len(properties) != len(fields) {
				t.Fatalf("schema properties = %v, want flat A2A fields %v", properties, fields)
			}
			for name := range fields {
				if _, ok := properties[name]; !ok {
					t.Errorf("missing A2A field %q", name)
				}
			}
			if err := validateJSONInput(strings.NewReader(body), typ); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSchemaEmbeddedFieldSelection(t *testing.T) {
	type first struct{ Value string }
	type second struct{ Value int }
	type tagged struct {
		Number int `json:"Value"`
	}
	type left struct{ first }
	type right struct{ first }
	type Recursive struct {
		*Recursive
		Name string
	}
	for _, value := range []any{
		struct{ first }{},
		struct{ *first }{&first{}},
		struct {
			first `json:"nested"`
		}{},
		struct {
			first
			Value bool
		}{},
		struct {
			first
			second
		}{},
		struct {
			first
			tagged
		}{},
		struct {
			first
			second
			tagged
		}{},
		struct {
			left
			right
		}{},
		struct {
			left
			first
		}{},
		Recursive{},
	} {
		t.Run(reflect.TypeOf(value).String(), func(t *testing.T) {
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &fields); err != nil {
				t.Fatal(err)
			}
			properties := inputshape.Describe(reflect.TypeOf(value))["properties"].(map[string]any)
			if len(properties) != len(fields) {
				t.Fatalf("schema fields = %v, encoding/json fields = %s", properties, encoded)
			}
			for name, raw := range fields {
				property, ok := properties[name].(map[string]any)
				if !ok {
					t.Fatalf("missing JSON field %q", name)
				}
				var decoded any
				if err := json.Unmarshal(raw, &decoded); err != nil {
					t.Fatal(err)
				}
				var want string
				switch decoded.(type) {
				case string:
					want = "string"
				case float64:
					want = "integer"
				case bool:
					want = "boolean"
				case map[string]any:
					want = "object"
				}
				if property["type"] != want {
					t.Errorf("%s type=%v want=%s", name, property["type"], want)
				}
			}
		})
	}
}

func TestValidateJSONStructureWithoutBusinessSideEffects(t *testing.T) {
	type input struct {
		Name string `json:"name"`
		Rows []int  `json:"rows"`
	}
	for _, tc := range []struct {
		body  string
		valid bool
	}{
		{`{"name":"真实输入","rows":[1,2]}`, true}, {`{}`, true}, // no business-required inference
		{`{"name":3}`, false}, {`{"rows":["x"]}`, false}, {`{"unexpected":true}`, true},
		{`{`, false}, {``, false}, {`{} {}`, false}, {`{} trailing`, false}, {`{"name":""}  `, true},
		{`null`, true}, {`{"NAME":"case insensitive"}`, true}, // same Go decoding as existing commands
	} {
		t.Run(tc.body, func(t *testing.T) {
			err := validateJSONInput(strings.NewReader(tc.body), reflect.TypeFor[input]())
			var decoded input
			if (err == nil) != (json.Unmarshal([]byte(tc.body), &decoded) == nil) {
				t.Fatal("schema validation differs from command decoding")
			}
			if (err == nil) != tc.valid {
				t.Errorf("error=%v", err)
			}
		})
	}
	if err := ExecuteSchemaCommand([]string{"missing", "operation"}); err == nil {
		t.Fatal("unknown command accepted")
	}
	if err := ExecuteSchemaCommand([]string{"connector", "create-task", "--argument", "missing"}); err == nil {
		t.Fatal("unknown argument accepted")
	}
	if err := ExecuteSchemaCommand([]string{"connector", "create-task", "surprise"}); err == nil {
		t.Fatal("extra argument accepted")
	}
}

func TestSchemaValidationFileErrors(t *testing.T) {
	for _, path := range []string{t.TempDir(), filepath.Join(t.TempDir(), "missing")} {
		if err := ExecuteSchemaCommand([]string{"connector", "create-task", "--validate-file", path}); err == nil {
			t.Fatal("invalid path accepted")
		}
	}
	path := filepath.Join(t.TempDir(), "empty.json")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteSchemaCommand([]string{"connector", "create-task", "--validate-file", path}); err == nil {
		t.Fatal("empty JSON file accepted")
	}
}
