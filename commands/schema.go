package commands

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/matrixorigin/matrixflow/sdk/cli/inputshape"
)

//go:embed schema_encoded_generated.json
var encodedInputShapes []byte

//go:embed schema_frontend_generated.json
var frontendInputShapes []byte

// ExecuteSchemaCommand is offline. Descriptions retain their CLI/server/renderer
// owner scope; none describes complete business validation.
func ExecuteSchemaCommand(args []string) error {
	var encoded map[string]map[string]json.RawMessage
	if err := json.Unmarshal(encodedInputShapes, &encoded); err != nil {
		return &LocalInputError{Code: "CLI_SCHEMA_METADATA_INVALID"}
	}
	var descriptions map[string]map[string]map[string]any
	if err := json.Unmarshal(frontendInputShapes, &descriptions); err != nil {
		return &LocalInputError{Code: "CLI_SCHEMA_METADATA_INVALID"}
	}
	for command, shapes := range encoded {
		if descriptions[command] == nil {
			descriptions[command] = map[string]map[string]any{}
		}
		for argument, shape := range shapes {
			if descriptions[command][argument] != nil {
				return &LocalInputError{Code: "CLI_SCHEMA_METADATA_INVALID"}
			}
			descriptions[command][argument] = map[string]any{
				"scope":  "server_decoded_json_shape_only",
				"notes":  "Generated from the server's JSON decoder types. Pass one JSON-encoded string to this flag. This describes structure only, not required fields, allowed values or business validation. Server decoding accepts null and ignores unknown object fields.",
				"schema": shape,
			}
		}
	}
	if len(args) == 0 {
		names := make([]string, 0, len(commandJSONTypes))
		for name := range commandJSONTypes {
			names = append(names, name)
		}
		for name := range descriptions {
			if _, exists := commandJSONTypes[name]; !exists {
				names = append(names, name)
			}
		}
		slices.Sort(names)
		return printJSON(names)
	}
	if len(args) < 2 {
		return fmt.Errorf("CLI_SCHEMA_COMMAND_REQUIRED: use schema <domain> <command>")
	}
	name := strings.Join(args[:2], " ")
	inputs, ok := commandJSONTypes[name]
	if !ok && descriptions[name] == nil {
		return fmt.Errorf("CLI_SCHEMA_NOT_FOUND: no typed JSON input for this command")
	}
	fs := flag.NewFlagSet("schema", flag.ContinueOnError)
	argument := fs.String("argument", "json", "JSON argument to describe or validate")
	file := fs.String("validate-file", "", "Validate JSON structure without sending a request")
	if err := fs.Parse(args[2:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("CLI_SCHEMA_UNEXPECTED_ARGUMENT: unexpected positional argument")
	}
	typ, ok := inputs[*argument]
	if !ok {
		if description, exists := descriptions[name][*argument]; exists {
			if *file != "" {
				return &LocalInputError{Code: "CLI_SCHEMA_VALIDATION_UNAVAILABLE"}
			}
			description["command"], description["argument"] = name, *argument
			return printJSON(description)
		}
		arguments := make([]string, 0, len(inputs)+len(descriptions[name]))
		for argument := range inputs {
			arguments = append(arguments, argument)
		}
		for argument := range descriptions[name] {
			arguments = append(arguments, argument)
		}
		slices.Sort(arguments)
		return fmt.Errorf("CLI_SCHEMA_ARGUMENT_NOT_FOUND: use --argument with one of: %s", strings.Join(arguments, ", "))
	}
	if *file != "" {
		info, err := os.Stat(*file)
		if err != nil {
			return &LocalInputError{Code: "CLI_FILE_UNREADABLE"}
		}
		if !info.Mode().IsRegular() {
			return &LocalInputError{Code: "CLI_FILE_NOT_REGULAR"}
		}
		f, err := os.Open(*file)
		if err != nil {
			return &LocalInputError{Code: "CLI_FILE_UNREADABLE"}
		}
		defer f.Close()
		if err := validateJSONInput(f, typ); err != nil {
			return err
		}
		return printJSON(map[string]any{"command": name, "argument": *argument, "valid": true, "validation": "structure_only"})
	}
	shape := inputshape.Describe(typ)
	if operation, runtime := strings.CutPrefix(name, "semantic-model "); runtime {
		if properties, ok := shape["properties"].(map[string]any); ok {
			maps.DeleteFunc(properties, func(name string, _ any) bool {
				return !knowledgeRuntimeJSONFieldSupported(operation, name)
			})
		}
	}
	description := map[string]any{
		"command": name, "argument": *argument, "scope": "cli_input_shape_only",
		"notes":  "Input uses the command's typed JSON decoder. Open objects and custom JSON values need domain documentation. SDK HTTP body keys may differ. This is not a complete JSON Schema or business validation.",
		"schema": shape,
	}
	if *argument == "json" && (name == "semantic-model create-with-sources" || name == "semantic-model append-sources") {
		description["guidance"] = semanticModelSourceGuidance()
	}
	if *argument == "json" && (name == "workflow invoke-workitem-api-service" || name == "workflow invoke-dynamic-service" || name == "workflow invoke-workflow-dynamic-service") {
		description["guidance"] = workflowServiceInvokeGuidance(name)
	}
	return printJSON(description)
}

func validateJSONInput(r io.Reader, typ reflect.Type) error {
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(reflect.New(typ).Interface()); err != nil {
		return fmt.Errorf("CLI_JSON_INVALID: input does not match the command JSON type")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("CLI_JSON_TRAILING_DATA: expected one JSON value")
	}
	return nil
}
