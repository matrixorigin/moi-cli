package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type workflowVersionContract struct {
	Name, Operation, Method, Path string
	Query                         map[string][]string `json:"query"`
	Input                         struct {
		WorkflowID             string          `json:"workflow_id"`
		ExecutionID            string          `json:"execution_id"`
		BlockID                string          `json:"block_id"`
		FileID                 string          `json:"file_id"`
		TargetOutputEmissionID string          `json:"target_output_emission_id"`
		NodeKey                string          `json:"node_key"`
		RerunID                string          `json:"rerun_id"`
		Mode                   string          `json:"mode"`
		SourceNodeExecutionID  string          `json:"source_node_execution_id"`
		SourceDispatchJobID    string          `json:"source_dispatch_job_id"`
		ExecutionIDs           []string        `json:"execution_ids"`
		AllFailed              bool            `json:"all_failed"`
		NodeID                 string          `json:"node_id"`
		OperatorID             string          `json:"operator_id"`
		Version                string          `json:"version"`
		VersionID              string          `json:"version_id"`
		TemplateID             string          `json:"template_id"`
		DispatchJobID          string          `json:"dispatch_job_id"`
		Number                 int             `json:"number"`
		Request                json.RawMessage `json:"request"`
		Name                   string          `json:"name"`
		NamePresent            bool            `json:"name_present"`
		Description            string          `json:"description"`
		DescriptionPresent     bool            `json:"description_present"`
		ExecutionMode          string          `json:"execution_mode"`
		ExecutionModePresent   bool            `json:"execution_mode_present"`
		CronExpression         string          `json:"cron_expression"`
		CronExpressionPresent  bool            `json:"cron_expression_present"`
		DSLYAML                string          `json:"dsl_yaml"`
		DSLYAMLPresent         bool            `json:"dsl_yaml_present"`
		RuntimeFields          json.RawMessage `json:"runtime_fields"`
		RuntimeLayout          json.RawMessage `json:"runtime_layout"`
		DefaultValues          json.RawMessage `json:"default_values"`
		VolumeTrigger          json.RawMessage `json:"volume_trigger"`
		DynamicService         json.RawMessage `json:"dynamic_service"`
		TaskID                 string          `json:"task_id"`
		ArtifactID             string          `json:"artifact_id"`
		ParentTaskID           string          `json:"parent_task_id"`
		ParentArtifactID       string          `json:"parent_artifact_id"`
		ProposalID             string          `json:"proposal_id"`
		IdempotencyKey         string          `json:"idempotency_key"`
		ContextID              string          `json:"context_id"`
		RequestID              string          `json:"request_id"`
		Limit                  int             `json:"limit"`
		Offset                 int             `json:"offset"`
		Acceptance             json.RawMessage `json:"acceptance"`
		Filename               string          `json:"filename"`
		ContentBase64          string          `json:"content_base64"`
	} `json:"input"`
	Multipart       *catalogUploadCLIMultipart `json:"multipart"`
	Request         json.RawMessage            `json:"request"`
	Response        json.RawMessage            `json:"response"`
	ResponseHeaders map[string]string          `json:"response_headers"`
	Status          int                        `json:"status"`
}

type workflowContractObservation struct {
	body      any
	multipart *catalogUploadCLIMultipart
}

func workflowContractJSONFile(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func workflowVersionJSON(raw json.RawMessage) any {
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

func workflowContractTextFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func workflowContractBinaryFile(t *testing.T, filename string, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), filename)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func workflowContractField[T any](t *testing.T, fields map[string]json.RawMessage, name string) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(fields[name], &value); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	return value
}

func workflowCustomOperatorOptions(t *testing.T, raw json.RawMessage, create bool) ([]sdk.CustomOperatorOption, []sdk.CustomOperatorUpdateOption, []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(fields))
	for name := range fields {
		if create && (name == "name" || name == "identifier" || name == "description" || name == "input_schema" || name == "output_schema") {
			continue
		}
		keys = append(keys, name)
	}
	sort.Strings(keys)
	var createOptions []sdk.CustomOperatorOption
	var updateOptions []sdk.CustomOperatorUpdateOption
	var args []string
	for _, name := range keys {
		flagName := "--" + strings.ReplaceAll(name, "_", "-")
		switch name {
		case "node_id":
			value := workflowContractField[string](t, fields, name)
			createOptions = append(createOptions, sdk.WithCustomOperatorNodeID(value))
			args = append(args, flagName, value)
		case "name":
			value := workflowContractField[string](t, fields, name)
			updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateName(value))
			args = append(args, flagName, value)
		case "description":
			value := workflowContractField[string](t, fields, name)
			updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateDescription(value))
			args = append(args, flagName, value)
		case "kind":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorKind(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateKind(value))
			}
			args = append(args, flagName, value)
		case "language":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorLanguage(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateLanguage(value))
			}
			args = append(args, flagName, value)
		case "handler":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorHandler(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateHandler(value))
			}
			args = append(args, flagName, value)
		case "version":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorVersion(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateVersion(value))
			}
			args = append(args, flagName, value)
		case "isolation_level":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorIsolationLevel(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateIsolationLevel(value))
			}
			args = append(args, flagName, value)
		case "enabled":
			value := workflowContractField[bool](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorEnabled(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateEnabled(value))
			}
			args = append(args, "--enabled="+fmt.Sprint(value))
		case "code":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorCode(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateCode(value))
			}
			args = append(args, "--code-file", workflowContractTextFile(t, value))
		case "source_file_id":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorSourceFileID(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateSourceFileID(value))
			}
			args = append(args, flagName, value)
		case "base_node_id":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorBaseNodeID(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateBaseNodeID(value))
			}
			args = append(args, flagName, value)
		case "base_node_version":
			value := workflowContractField[string](t, fields, name)
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorBaseNodeVersion(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateBaseNodeVersion(value))
			}
			args = append(args, flagName, value)
		case "input_schema", "output_schema", "binding_config":
			value := json.RawMessage(fields[name])
			if create {
				createOptions = append(createOptions, sdk.WithCustomOperatorBindingConfig(value))
			} else {
				switch name {
				case "input_schema":
					updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateInputSchema(value))
				case "output_schema":
					updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateOutputSchema(value))
				case "binding_config":
					updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateBindingConfig(value))
				}
			}
			args = append(args, flagName+"-file", workflowContractJSONFile(t, fields[name]))
		case "catalog_id", "database_id":
			value := workflowContractField[int64](t, fields, name)
			if create {
				if name == "catalog_id" {
					createOptions = append(createOptions, sdk.WithCustomOperatorCatalogID(value))
				} else {
					createOptions = append(createOptions, sdk.WithCustomOperatorDatabaseID(value))
				}
			} else if name == "catalog_id" {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateCatalogID(value))
			} else {
				updateOptions = append(updateOptions, sdk.WithCustomOperatorUpdateDatabaseID(value))
			}
			args = append(args, flagName, fmt.Sprint(value))
		default:
			t.Fatalf("unknown custom operator option %s", name)
		}
	}
	return createOptions, updateOptions, args
}

func workflowQueryParamArgs(query map[string][]string) []string {
	keys := make([]string, 0, len(query))
	for key := range query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var args []string
	for _, key := range keys {
		for _, value := range query[key] {
			args = append(args, "--query-param", key+"="+value)
		}
	}
	return args
}

func workflowRunListArgs(query map[string][]string) []string {
	var args []string
	for _, item := range []struct{ query, flag string }{
		{"limit", "--limit"}, {"offset", "--offset"}, {"status", "--status"},
		{"execution_mode", "--execution-mode"}, {"workflow_name", "--workflow-name"},
	} {
		if values := query[item.query]; len(values) > 0 {
			args = append(args, item.flag, values[0])
		}
	}
	return args
}

func workflowExecutionOptions(t *testing.T, raw json.RawMessage, run bool) ([]sdk.WorkflowRunOption, []sdk.WorkflowRerunOption, []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(fields))
	for name := range fields {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	var runOptions []sdk.WorkflowRunOption
	var rerunOptions []sdk.WorkflowRerunOption
	var args []string
	for _, name := range keys {
		switch name {
		case "values":
			value := workflowContractField[map[string]any](t, fields, name)
			if run {
				runOptions = append(runOptions, sdk.WithWorkflowRunValues(value))
			} else {
				rerunOptions = append(rerunOptions, sdk.WithWorkflowRerunValues(value))
			}
			args = append(args, "--values-file", workflowContractJSONFile(t, fields[name]))
		case "replace_saved_defaults":
			value := workflowContractField[bool](t, fields, name)
			if run {
				runOptions = append(runOptions, sdk.WithWorkflowRunReplaceSavedDefaults(value))
			} else {
				rerunOptions = append(rerunOptions, sdk.WithWorkflowRetryReplaceSavedDefaults(value))
			}
			args = append(args, "--replace-saved-defaults="+fmt.Sprint(value))
		case "trigger_now":
			value := workflowContractField[bool](t, fields, name)
			if run {
				runOptions = append(runOptions, sdk.WithWorkflowRunTriggerNow(value))
			} else {
				rerunOptions = append(rerunOptions, sdk.WithWorkflowRerunTriggerNow(value))
			}
			args = append(args, "--trigger-now="+fmt.Sprint(value))
		case "config_override":
			value := json.RawMessage(fields[name])
			rerunOptions = append(rerunOptions, sdk.WithWorkflowRerunConfigOverrideJSON(value))
			args = append(args, "--config-override-file", workflowContractJSONFile(t, fields[name]))
		case "expected_rerun_contract_hash":
			value := workflowContractField[string](t, fields, name)
			rerunOptions = append(rerunOptions, sdk.WithExpectedWorkflowRerunContractHash(value))
			args = append(args, "--expected-rerun-contract-hash", value)
		case "reason":
			value := workflowContractField[string](t, fields, name)
			rerunOptions = append(rerunOptions, sdk.WithWorkflowRerunReason(value))
			args = append(args, "--reason", value)
		default:
			t.Fatalf("unknown execution option %s", name)
		}
	}
	return runOptions, rerunOptions, args
}

func workflowCancelOptions(t *testing.T, raw json.RawMessage) ([]sdk.WorkflowCancelOption, []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["reason"]; !ok {
		return nil, nil
	}
	reason := workflowContractField[string](t, fields, "reason")
	return []sdk.WorkflowCancelOption{sdk.WithWorkflowCancelReason(reason)}, []string{"--reason", reason}
}

func workflowLineageRevisionOptions(t *testing.T, raw json.RawMessage) ([]sdk.LineageRevisionOption, []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	stringOptions := map[string]func(string) sdk.LineageRevisionOption{
		"base_revision_id":           sdk.WithLineageBaseRevisionID,
		"expected_base_content_hash": sdk.WithLineageExpectedBaseContentHash,
		"revision_content_ref":       sdk.WithLineageRevisionContentRef,
		"revision_content_hash":      sdk.WithLineageRevisionContentHash,
		"revision_content_payload":   sdk.WithLineageRevisionContentPayload,
		"patch_json":                 sdk.WithLineagePatchJSON,
	}
	keys := make([]string, 0, len(fields))
	for name := range fields {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	var options []sdk.LineageRevisionOption
	var args []string
	for _, name := range keys {
		flagName := "--" + strings.ReplaceAll(name, "_", "-")
		switch name {
		case "use_original", "use_current_effective":
			value := workflowContractField[bool](t, fields, name)
			if name == "use_original" {
				options = append(options, sdk.WithLineageUseOriginal(value))
			} else {
				options = append(options, sdk.WithLineageUseCurrentEffective(value))
			}
			args = append(args, flagName+"="+fmt.Sprint(value))
		case "expected_effective_set_version":
			value := workflowContractField[int64](t, fields, name)
			options = append(options, sdk.WithLineageExpectedEffectiveSetVersion(value))
			args = append(args, flagName, fmt.Sprint(value))
		default:
			option, ok := stringOptions[name]
			if !ok {
				t.Fatalf("unknown lineage revision option %s", name)
			}
			value := workflowContractField[string](t, fields, name)
			options = append(options, option(value))
			args = append(args, flagName, value)
		}
	}
	return options, args
}

func workflowLineageSwitchOptions(t *testing.T, raw json.RawMessage) ([]sdk.LineageRevisionSwitch, []sdk.LineageSwitchOption, []string) {
	t.Helper()
	var body struct {
		Switches []sdk.LineageRevisionSwitch `json:"switches"`
		Expected *int64                      `json:"expected_effective_set_version"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	var options []sdk.LineageSwitchOption
	args := []string{"--json", string(raw)}
	switchesJSON, err := json.Marshal(body.Switches)
	if err != nil {
		t.Fatal(err)
	}
	args[1] = string(switchesJSON)
	if body.Expected != nil {
		options = append(options, sdk.WithLineageSwitchExpectedEffectiveSetVersion(*body.Expected))
		args = append(args, "--expected-effective-set-version", fmt.Sprint(*body.Expected))
	}
	return body.Switches, options, args
}

func workflowLineageRerunOptions(t *testing.T, raw json.RawMessage) ([]sdk.LineageRerunOption, []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(fields))
	for name := range fields {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	var options []sdk.LineageRerunOption
	var args []string
	for _, name := range keys {
		if name == "mode" {
			value := workflowContractField[string](t, fields, name)
			options = append(options, sdk.WithLineageRerunMode(value))
			args = append(args, "--mode", value)
			continue
		}
		flagName := "--" + strings.ReplaceAll(name, "_", "-") + "-file"
		value := json.RawMessage(fields[name])
		switch name {
		case "config_override":
			options = append(options, sdk.WithLineageRerunConfigOverride(value))
		case "input_payload_override":
			options = append(options, sdk.WithLineageRerunInputPayloadOverride(value))
		case "vars_payload_override":
			options = append(options, sdk.WithLineageRerunVarsPayloadOverride(value))
		default:
			t.Fatalf("unknown lineage rerun option %s", name)
		}
		args = append(args, flagName, workflowContractJSONFile(t, fields[name]))
	}
	return options, args
}

func workflowLineageRevisionRerunOptions(t *testing.T, raw json.RawMessage) ([]sdk.LineageRevisionRerunOption, []string) {
	t.Helper()
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(fields))
	for name := range fields {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	var options []sdk.LineageRevisionRerunOption
	var args []string
	for _, name := range keys {
		switch name {
		case "mode":
			value := workflowContractField[string](t, fields, name)
			options = append(options, sdk.WithLineageRevisionRerunMode(value))
			args = append(args, "--mode", value)
		case "source_effective_set_version":
			value := workflowContractField[int64](t, fields, name)
			options = append(options, sdk.WithLineageRevisionRerunSourceEffectiveSetVersion(value))
			args = append(args, "--source-effective-set-version", fmt.Sprint(value))
		case "selected_revisions_json":
			value := json.RawMessage(fields[name])
			options = append(options, sdk.WithLineageSelectedRevisionsJSON(value))
			args = append(args, "--selected-revisions-json-file", workflowContractJSONFile(t, fields[name]))
		default:
			t.Fatalf("unknown lineage revision rerun option %s", name)
		}
	}
	return options, args
}

func callWorkflowVersionGo(t *testing.T, client *sdk.Client, scenario workflowVersionContract) (any, []string, error) {
	t.Helper()
	service := client.Workflows("workspace-1")
	input := scenario.Input
	args := []string{scenario.Operation}
	var result any
	var err error
	switch scenario.Operation {
	case "workflow-templates":
		result, err = service.WorkflowTemplates(t.Context())
	case "delete-workflow-template":
		args = append(args, "--template-id", input.TemplateID)
		result, err = service.DeleteWorkflowTemplate(t.Context(), input.TemplateID)
	case "versions":
		args = append(args, "--workflow-id", input.WorkflowID)
		result, err = service.Versions(t.Context(), input.WorkflowID)
	case "create-version":
		request := &sdk.WorkflowVersionCreateRequest{}
		if unmarshalErr := protojson.Unmarshal(input.Request, request); unmarshalErr != nil {
			t.Fatal(unmarshalErr)
		}
		args = append(args, "--workflow-id", input.WorkflowID, "--json", string(input.Request))
		result, err = service.CreateVersion(t.Context(), input.WorkflowID, request)
	case "version-by-number":
		args = append(args, "--workflow-id", input.WorkflowID, "--number", fmt.Sprint(input.Number))
		result, err = service.VersionByNumber(t.Context(), input.WorkflowID, input.Number)
	case "latest-published-version":
		args = append(args, "--workflow-id", input.WorkflowID)
		result, err = service.LatestPublishedVersion(t.Context(), input.WorkflowID)
	case "version":
		args = append(args, "--workflow-id", input.WorkflowID, "--version-id", input.VersionID)
		result, err = service.Version(t.Context(), input.WorkflowID, input.VersionID)
	case "publish-version":
		args = append(args, "--workflow-id", input.WorkflowID, "--version-id", input.VersionID)
		result, err = service.PublishVersion(t.Context(), input.WorkflowID, input.VersionID)
	case "deprecate-version":
		args = append(args, "--workflow-id", input.WorkflowID, "--version-id", input.VersionID)
		result, err = service.DeprecateVersion(t.Context(), input.WorkflowID, input.VersionID)
	case "delete-version":
		args = append(args, "--workflow-id", input.WorkflowID, "--version-id", input.VersionID)
		err = service.DeleteVersion(t.Context(), input.WorkflowID, input.VersionID)
	case "apps":
		args = knowledgeLifecycleQueryArgs(args, scenario.Query)
		result, err = service.Apps(t.Context(), knowledgeLifecycleQuery(scenario.Query)...)
	case "list":
		for _, name := range []string{"limit", "offset"} {
			args = append(args, "--"+name, scenario.Query[name][0])
		}
		args = append(args, "--name-search", scenario.Query["name_search"][0], "--include-dynamic-service="+scenario.Query["include_dynamic_service"][0])
		result, err = service.List(t.Context(), knowledgeLifecycleQuery(scenario.Query)...)
	case "get":
		args = append(args, "--workflow-id", input.WorkflowID)
		result, err = service.Get(t.Context(), input.WorkflowID)
	case "update":
		args = append(args, "--workflow-id", input.WorkflowID)
		var options []sdk.WorkflowUpdateOption
		if input.NamePresent {
			options = append(options, sdk.WithWorkflowUpdateName(input.Name))
			args = append(args, "--name", input.Name)
		}
		if input.DescriptionPresent {
			options = append(options, sdk.WithWorkflowUpdateDescription(input.Description))
			args = append(args, "--description", input.Description)
		}
		if input.ExecutionModePresent {
			options = append(options, sdk.WithWorkflowUpdateExecutionMode(input.ExecutionMode))
			args = append(args, "--execution-mode", input.ExecutionMode)
		}
		if input.CronExpressionPresent {
			options = append(options, sdk.WithWorkflowUpdateCronExpression(input.CronExpression))
			args = append(args, "--cron-expression", input.CronExpression)
		}
		if input.DSLYAMLPresent {
			options = append(options, sdk.WithWorkflowUpdateDSLYAML(input.DSLYAML))
			args = append(args, "--dsl-yaml", input.DSLYAML)
		}
		if len(input.RuntimeFields) > 0 {
			var value any
			if err := json.Unmarshal(input.RuntimeFields, &value); err != nil {
				t.Fatal(err)
			}
			options = append(options, sdk.WithWorkflowUpdateInputForm(value))
			args = append(args, "--runtime-fields-file", workflowContractJSONFile(t, input.RuntimeFields))
		}
		if len(input.RuntimeLayout) > 0 {
			var value any
			if err := json.Unmarshal(input.RuntimeLayout, &value); err != nil {
				t.Fatal(err)
			}
			options = append(options, sdk.WithWorkflowUpdateFormLayout(value))
			args = append(args, "--runtime-layout-file", workflowContractJSONFile(t, input.RuntimeLayout))
		}
		if len(input.DefaultValues) > 0 {
			var value map[string]any
			if err := json.Unmarshal(input.DefaultValues, &value); err != nil {
				t.Fatal(err)
			}
			options = append(options, sdk.WithWorkflowUpdateDefaultValues(value))
			args = append(args, "--default-values-file", workflowContractJSONFile(t, input.DefaultValues))
		}
		if len(input.VolumeTrigger) > 0 {
			var value sdk.WorkflowVolumeTriggerParam
			if err := json.Unmarshal(input.VolumeTrigger, &value); err != nil {
				t.Fatal(err)
			}
			options = append(options, sdk.WithWorkflowUpdateVolumeTrigger(value))
			args = append(args, "--volume-trigger-file", workflowContractJSONFile(t, input.VolumeTrigger))
		}
		if len(input.DynamicService) > 0 {
			var value sdk.WorkflowDynamicServiceParam
			if err := json.Unmarshal(input.DynamicService, &value); err != nil {
				t.Fatal(err)
			}
			options = append(options, sdk.WithWorkflowUpdateDynamicService(value))
			args = append(args, "--dynamic-service-file", workflowContractJSONFile(t, input.DynamicService))
		}
		result, err = service.Update(t.Context(), input.WorkflowID, options...)
	case "delete":
		args = append(args, "--workflow-id", input.WorkflowID)
		result, err = service.Delete(t.Context(), input.WorkflowID)
	case "get-dispatch-job":
		args = append(args, "--workflow-id", input.WorkflowID, "--dispatch-job-id", input.DispatchJobID)
		result, err = service.GetDispatchJob(t.Context(), input.WorkflowID, input.DispatchJobID)
	case "enable", "pause", "resume":
		args = append(args, "--workflow-id", input.WorkflowID)
		switch scenario.Operation {
		case "enable":
			result, err = service.Enable(t.Context(), input.WorkflowID)
		case "pause":
			result, err = service.Pause(t.Context(), input.WorkflowID)
		case "resume":
			result, err = service.Resume(t.Context(), input.WorkflowID)
		}
	case "submit-candidate":
		request := &sdk.WorkflowCandidateSubmitRequest{}
		if err := protojson.Unmarshal(input.Request, request); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--json", string(input.Request))
		result, err = service.SubmitCandidate(t.Context(), request)
	case "test-candidate":
		request := &sdk.WorkflowCandidateTestRequest{}
		if err := protojson.Unmarshal(input.Request, request); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--json", string(input.Request))
		result, err = service.TestCandidate(t.Context(), request)
	case "accept-candidate":
		args = append(args, "--task-id", input.TaskID, "--artifact-id", input.ArtifactID, "--proposal-id", input.ProposalID, "--idempotency-key", input.IdempotencyKey)
		var accepted *sdk.AcceptedWorkflowCandidate
		accepted, err = service.AcceptCandidate(t.Context(), sdk.WorkflowCandidateRef{TaskID: input.TaskID, ArtifactID: input.ArtifactID}, &sdk.WorkflowCandidateConfirmation{ProposalId: input.ProposalID, IdempotencyKey: input.IdempotencyKey})
		if err == nil {
			result = accepted.Acceptance()
		}
	case "history":
		args = append(args, "--context-id", input.ContextID, "--limit", fmt.Sprint(input.Limit), "--offset", fmt.Sprint(input.Offset))
		result, err = service.History(t.Context(), input.ContextID, input.Limit, input.Offset)
	case "restore-history":
		args = append(args, "--context-id", input.ContextID, "--request-id", input.RequestID, "--task-id", input.TaskID, "--artifact-id", input.ArtifactID)
		request := sdk.WorkflowHistoryRestoreRequest{ContextID: input.ContextID, RequestID: input.RequestID, CandidateRef: sdk.WorkflowCandidateRef{TaskID: input.TaskID, ArtifactID: input.ArtifactID}}
		if input.ParentTaskID != "" || input.ParentArtifactID != "" {
			request.ParentRef = &sdk.WorkflowCandidateRef{TaskID: input.ParentTaskID, ArtifactID: input.ParentArtifactID}
			args = append(args, "--parent-task-id", input.ParentTaskID, "--parent-artifact-id", input.ParentArtifactID)
		}
		result, err = service.RestoreHistory(t.Context(), request)
	case "reject-candidate-proposal":
		args = append(args, "--proposal-id", input.ProposalID)
		result, err = service.RejectCandidateProposal(t.Context(), input.ProposalID)
	case "deploy":
		args = append(args, "--name", input.Name, "--dsl-yaml", input.DSLYAML)
		var options []sdk.WorkflowDeployOption
		if input.DescriptionPresent {
			options = append(options, sdk.WithWorkflowDescription(input.Description))
			args = append(args, "--description", input.Description)
		}
		result, err = service.Deploy(t.Context(), input.Name, input.DSLYAML, options...)
	case "deploy-candidate":
		acceptance := &sdk.WorkflowCandidateAcceptance{}
		if err := protojson.Unmarshal(input.Acceptance, acceptance); err != nil {
			t.Fatal(err)
		}
		accepted, err := sdk.AcceptedWorkflowCandidateFromAcceptance(acceptance)
		if err != nil {
			t.Fatal(err)
		}
		args = append(args, "--name", input.Name, "--acceptance-file", workflowContractJSONFile(t, input.Acceptance))
		var options []sdk.WorkflowDeployOption
		if input.DescriptionPresent {
			options = append(options, sdk.WithWorkflowDescription(input.Description))
			args = append(args, "--description", input.Description)
		}
		result, err = service.DeployCandidate(t.Context(), input.Name, accepted, options...)
	case "custom-operators":
		args = append(args, workflowQueryParamArgs(scenario.Query)...)
		result, err = service.CustomOperators(t.Context(), knowledgeLifecycleQuery(scenario.Query)...)
	case "create-custom-operator":
		var spec sdk.CustomOperatorSpec
		if err := json.Unmarshal(input.Request, &spec); err != nil {
			t.Fatal(err)
		}
		createOptions, _, optionArgs := workflowCustomOperatorOptions(t, input.Request, true)
		args = append(args, "--json", string(input.Request))
		args = append(args, optionArgs...)
		result, err = service.CreateCustomOperator(t.Context(), spec, createOptions...)
	case "get-custom-operator-source":
		args = append(args, "--node-id", input.NodeID, "--version", input.Version)
		if input.TaskID != "" {
			args = append(args, "--task-id", input.TaskID)
		}
		result, err = service.GetCustomOperatorSource(t.Context(), input.NodeID, input.Version, input.TaskID)
	case "delete-custom-operator":
		args = append(args, "--operator-id", input.OperatorID)
		result, err = service.DeleteCustomOperator(t.Context(), input.OperatorID)
	case "get-custom-operator":
		args = append(args, "--operator-id", input.OperatorID)
		result, err = service.GetCustomOperator(t.Context(), input.OperatorID)
	case "update-custom-operator":
		_, updateOptions, optionArgs := workflowCustomOperatorOptions(t, input.Request, false)
		args = append(args, "--operator-id", input.OperatorID)
		args = append(args, optionArgs...)
		result, err = service.UpdateCustomOperator(t.Context(), input.OperatorID, updateOptions...)
	case "get-custom-operator-code":
		args = append(args, "--operator-id", input.OperatorID)
		result, err = service.GetCustomOperatorCode(t.Context(), input.OperatorID)
	case "disable-custom-operator", "enable-custom-operator":
		args = append(args, "--operator-id", input.OperatorID)
		if scenario.Operation == "disable-custom-operator" {
			result, err = service.DisableCustomOperator(t.Context(), input.OperatorID)
		} else {
			result, err = service.EnableCustomOperator(t.Context(), input.OperatorID)
		}
	case "test-custom-operator":
		var testInput sdk.CustomOperatorTestInput
		if err := json.Unmarshal(input.Request, &testInput); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--operator-id", input.OperatorID, "--json", string(input.Request))
		result, err = service.TestCustomOperator(t.Context(), input.OperatorID, testInput)
	case "invoke-workflow-dynamic-service":
		var spec sdk.DynamicServiceInvokeSpec
		if err := json.Unmarshal(input.Request, &spec); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--json", string(input.Request))
		if input.Filename != "" {
			content, decodeErr := base64.StdEncoding.DecodeString(input.ContentBase64)
			if decodeErr != nil {
				t.Fatal(decodeErr)
			}
			spec.Files = []sdk.MultipartFile{{FieldName: "file", Filename: input.Filename, Reader: bytes.NewReader(content)}}
			args = append(args, "--file", workflowContractBinaryFile(t, input.Filename, content), "--file-field", "file")
		}
		result, err = service.InvokeWorkflowDynamicService(t.Context(), spec)
	case "ensure-test-scratch", "observe-test-scratch":
		request := &sdk.WorkflowScratchScope{}
		if err := protojson.Unmarshal(input.Request, request); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--json", string(input.Request))
		if scenario.Operation == "ensure-test-scratch" {
			result, err = service.EnsureTestScratch(t.Context(), request)
		} else {
			result, err = service.ObserveTestScratch(t.Context(), request)
		}
	case "reset-test-scratch":
		args = append(args, "--task-id", input.TaskID)
		result, err = service.ResetTestScratch(t.Context(), input.TaskID)
	case "workitem-catalog":
		result, err = service.WorkitemCatalog(t.Context())
	case "workitem-catalog-item":
		args = append(args, "--node-id", input.NodeID)
		result, err = service.WorkitemCatalogItem(t.Context(), input.NodeID)
	case "get-workitem-api-service":
		args = append(args, "--node-id", input.NodeID, "--version", input.Version)
		result, err = service.GetWorkitemAPIService(t.Context(), input.NodeID, input.Version)
	case "publish-workitem-api-service":
		var spec sdk.WorkitemAPIServicePublishSpec
		if err := json.Unmarshal(input.Request, &spec); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--node-id", input.NodeID, "--json", string(input.Request))
		result, err = service.PublishWorkitemAPIService(t.Context(), input.NodeID, spec)
	case "disable-workitem-api-service", "enable-workitem-api-service":
		args = append(args, "--node-id", input.NodeID, "--version", input.Version)
		if scenario.Operation == "disable-workitem-api-service" {
			result, err = service.DisableWorkitemAPIService(t.Context(), input.NodeID, input.Version)
		} else {
			result, err = service.EnableWorkitemAPIService(t.Context(), input.NodeID, input.Version)
		}
	case "invoke-workitem-api-service":
		var spec sdk.WorkitemAPIServiceInvokeSpec
		if err := json.Unmarshal(input.Request, &spec); err != nil {
			t.Fatal(err)
		}
		args = append(args, "--node-id", input.NodeID, "--json", string(input.Request))
		result, err = service.InvokeWorkitemAPIService(t.Context(), input.NodeID, spec)
	case "list-all-runs":
		args = append(args, workflowRunListArgs(scenario.Query)...)
		result, err = service.ListAllRuns(t.Context(), knowledgeLifecycleQuery(scenario.Query)...)
	case "get-run-by-id":
		args = append(args, "--execution-id", input.ExecutionID)
		result, err = service.GetRunByID(t.Context(), input.ExecutionID)
	case "get-run-import-context-by-id":
		args = append(args, "--execution-id", input.ExecutionID)
		result, err = service.GetRunImportContextByID(t.Context(), input.ExecutionID)
	case "list-run-import-files-by-id":
		args = append(args, "--execution-id", input.ExecutionID)
		args = append(args, workflowQueryParamArgs(scenario.Query)...)
		result, err = service.ListRunImportFilesByID(t.Context(), input.ExecutionID, knowledgeLifecycleQuery(scenario.Query)...)
	case "get-run-node-by-id":
		args = append(args, "--execution-id", input.ExecutionID, "--node-key", input.NodeKey)
		result, err = service.GetRunNodeByID(t.Context(), input.ExecutionID, input.NodeKey)
	case "get-run-result-by-id":
		args = append(args, "--execution-id", input.ExecutionID)
		result, err = service.GetRunResultByID(t.Context(), input.ExecutionID)
	case "list-file-executions":
		args = append(args, "--file-id", scenario.Query["file_id"][0])
		result, err = service.ListFileExecutions(t.Context(), knowledgeLifecycleQuery(scenario.Query)...)
	case "list-runs":
		args = append(args, "--workflow-id", input.WorkflowID)
		args = append(args, workflowRunListArgs(scenario.Query)...)
		result, err = service.ListRuns(t.Context(), input.WorkflowID, knowledgeLifecycleQuery(scenario.Query)...)
	case "run":
		runOptions, _, optionArgs := workflowExecutionOptions(t, input.Request, true)
		args = append(args, "--workflow-id", input.WorkflowID)
		args = append(args, optionArgs...)
		result, err = service.Run(t.Context(), input.WorkflowID, runOptions...)
	case "batch-retry-failed-files":
		args = append(args, "--workflow-id", input.WorkflowID, "--source-dispatch-job-id", input.SourceDispatchJobID, "--execution-ids", strings.Join(input.ExecutionIDs, ","), "--all-failed="+fmt.Sprint(input.AllFailed), "--request-id", input.RequestID)
		result, err = service.BatchRetryFailedFiles(t.Context(), input.WorkflowID, input.SourceDispatchJobID, input.ExecutionIDs, input.AllFailed, input.RequestID)
	case "delete-run":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		result, err = service.DeleteRun(t.Context(), input.WorkflowID, input.ExecutionID)
	case "get-run":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		result, err = service.GetRun(t.Context(), input.WorkflowID, input.ExecutionID)
	case "cancel-run":
		options, optionArgs := workflowCancelOptions(t, input.Request)
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		args = append(args, optionArgs...)
		result, err = service.CancelRun(t.Context(), input.WorkflowID, input.ExecutionID, options...)
	case "get-run-import-context":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		result, err = service.GetRunImportContext(t.Context(), input.WorkflowID, input.ExecutionID)
	case "list-run-import-files":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		args = append(args, workflowQueryParamArgs(scenario.Query)...)
		result, err = service.ListRunImportFiles(t.Context(), input.WorkflowID, input.ExecutionID, knowledgeLifecycleQuery(scenario.Query)...)
	case "get-run-node":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID, "--node-key", input.NodeKey)
		result, err = service.GetRunNode(t.Context(), input.WorkflowID, input.ExecutionID, input.NodeKey)
	case "get-run-node-rerun-plan":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID, "--node-key", input.NodeKey)
		args = append(args, workflowQueryParamArgs(scenario.Query)...)
		result, err = service.GetRunNodeRerunPlan(t.Context(), input.WorkflowID, input.ExecutionID, input.NodeKey, knowledgeLifecycleQuery(scenario.Query)...)
	case "create-run-node-rerun":
		_, options, optionArgs := workflowExecutionOptions(t, input.Request, false)
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID, "--node-key", input.NodeKey, "--mode", input.Mode, "--source-node-execution-id", input.SourceNodeExecutionID)
		args = append(args, optionArgs...)
		result, err = service.CreateRunNodeRerun(t.Context(), input.WorkflowID, input.ExecutionID, input.NodeKey, input.Mode, input.SourceNodeExecutionID, options...)
	case "pause-run", "refresh-run", "resume-run":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		switch scenario.Operation {
		case "pause-run":
			result, err = service.PauseRun(t.Context(), input.WorkflowID, input.ExecutionID)
		case "refresh-run":
			result, err = service.RefreshRun(t.Context(), input.WorkflowID, input.ExecutionID)
		case "resume-run":
			result, err = service.ResumeRun(t.Context(), input.WorkflowID, input.ExecutionID)
		}
	case "list-run-rerun-branches":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		result, err = service.ListRunRerunBranches(t.Context(), input.WorkflowID, input.ExecutionID)
	case "get-run-rerun":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID, "--rerun-id", input.RerunID)
		result, err = service.GetRunRerun(t.Context(), input.WorkflowID, input.ExecutionID, input.RerunID)
	case "cancel-run-rerun":
		options, optionArgs := workflowCancelOptions(t, input.Request)
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID, "--rerun-id", input.RerunID)
		args = append(args, optionArgs...)
		result, err = service.CancelRunRerun(t.Context(), input.WorkflowID, input.ExecutionID, input.RerunID, options...)
	case "get-run-result":
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		result, err = service.GetRunResult(t.Context(), input.WorkflowID, input.ExecutionID)
	case "retry-run":
		_, options, optionArgs := workflowExecutionOptions(t, input.Request, false)
		args = append(args, "--workflow-id", input.WorkflowID, "--execution-id", input.ExecutionID)
		args = append(args, optionArgs...)
		result, err = service.RetryRun(t.Context(), input.WorkflowID, input.ExecutionID, options...)
	case "artifact-block-revisions":
		args = append(args, "--artifact-id", input.ArtifactID, "--block-id", input.BlockID)
		result, err = service.ArtifactBlockRevisions(t.Context(), input.ArtifactID, input.BlockID)
	case "create-artifact-block-revision":
		options, optionArgs := workflowLineageRevisionOptions(t, input.Request)
		args = append(args, "--artifact-id", input.ArtifactID, "--block-id", input.BlockID)
		args = append(args, optionArgs...)
		result, err = service.CreateArtifactBlockRevision(t.Context(), input.ArtifactID, input.BlockID, options...)
	case "switch-artifact-effective-revisions":
		switches, options, optionArgs := workflowLineageSwitchOptions(t, input.Request)
		args = append(args, "--artifact-id", input.ArtifactID)
		args = append(args, optionArgs...)
		result, err = service.SwitchArtifactEffectiveRevisions(t.Context(), input.ArtifactID, switches, options...)
	case "artifact-node":
		args = append(args, "--artifact-id", input.ArtifactID, "--node-id", input.NodeID)
		result, err = service.ArtifactNode(t.Context(), input.ArtifactID, input.NodeID)
	case "artifact-node-block-revisions":
		args = append(args, "--artifact-id", input.ArtifactID, "--node-id", input.NodeID, "--block-id", input.BlockID)
		result, err = service.ArtifactNodeBlockRevisions(t.Context(), input.ArtifactID, input.NodeID, input.BlockID)
	case "create-artifact-node-block-revision":
		options, optionArgs := workflowLineageRevisionOptions(t, input.Request)
		args = append(args, "--artifact-id", input.ArtifactID, "--node-id", input.NodeID, "--block-id", input.BlockID)
		args = append(args, optionArgs...)
		result, err = service.CreateArtifactNodeBlockRevision(t.Context(), input.ArtifactID, input.NodeID, input.BlockID, options...)
	case "switch-artifact-node-effective-revisions":
		switches, options, optionArgs := workflowLineageSwitchOptions(t, input.Request)
		args = append(args, "--artifact-id", input.ArtifactID, "--node-id", input.NodeID)
		args = append(args, optionArgs...)
		result, err = service.SwitchArtifactNodeEffectiveRevisions(t.Context(), input.ArtifactID, input.NodeID, switches, options...)
	case "artifact-node-output-blocks":
		args = append(args, "--artifact-id", input.ArtifactID, "--node-id", input.NodeID)
		result, err = service.ArtifactNodeOutputBlocks(t.Context(), input.ArtifactID, input.NodeID)
	case "create-target-artifact-node-revision-rerun":
		options, optionArgs := workflowLineageRevisionRerunOptions(t, input.Request)
		args = append(args, "--artifact-id", input.ArtifactID, "--node-id", input.NodeID, "--target-output-emission-id", input.TargetOutputEmissionID)
		args = append(args, optionArgs...)
		result, err = service.CreateTargetArtifactNodeRevisionRerun(t.Context(), input.ArtifactID, input.NodeID, input.TargetOutputEmissionID, options...)
	case "create-artifact-node-rerun":
		options, optionArgs := workflowLineageRerunOptions(t, input.Request)
		args = append(args, "--artifact-id", input.ArtifactID, "--node-id", input.NodeID)
		args = append(args, optionArgs...)
		result, err = service.CreateArtifactNodeRerun(t.Context(), input.ArtifactID, input.NodeID, options...)
	case "artifact-output-blocks":
		args = append(args, "--artifact-id", input.ArtifactID)
		result, err = service.ArtifactOutputBlocks(t.Context(), input.ArtifactID)
	case "artifact-lineage-overview":
		args = append(args, "--artifact-id", input.ArtifactID)
		result, err = service.ArtifactLineageOverview(t.Context(), input.ArtifactID)
	case "catalog-file-lineage-overview":
		args = append(args, "--file-id", input.FileID)
		args = append(args, workflowQueryParamArgs(scenario.Query)...)
		result, err = service.CatalogFileLineageOverview(t.Context(), input.FileID, knowledgeLifecycleQuery(scenario.Query)...)
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	return result, args, err
}

func assertWorkflowVersionError(t *testing.T, err error, scenario workflowVersionContract) {
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
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if failure.Code != envelope.Code || failure.Message != envelope.Msg || failure.Status != scenario.Status || failure.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v", failure)
	}
}

func assertWorkflowVersionResult(t *testing.T, output []byte, reference any, scenario workflowVersionContract) {
	t.Helper()
	if reference == nil {
		if len(output) != 0 {
			t.Fatalf("void output=%q", output)
		}
		return
	}
	var actual any
	if message, ok := reference.(proto.Message); ok {
		actual = message.ProtoReflect().New().Interface()
	} else {
		actual = reflect.New(reflect.TypeOf(reference).Elem()).Interface()
	}
	if _, ok := actual.(proto.Message); !ok {
		if err := json.Unmarshal(output, actual); err != nil {
			t.Fatalf("result=%q: %v", output, err)
		}
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	var expected any
	if message, ok := reference.(proto.Message); ok {
		expected = message.ProtoReflect().New().Interface()
	} else {
		expected = reflect.New(reflect.TypeOf(reference).Elem()).Interface()
	}
	expectedData := envelope.Data
	if _, ok := reference.(*sdk.WorkflowTemplateList); ok {
		expectedData = []byte(`{"templates":` + string(expectedData) + `}`)
	}
	if _, ok := reference.(*sdk.WorkitemCatalogDetail); ok {
		expectedData = []byte(`{"item":` + string(expectedData) + `}`)
	}
	if message, ok := expected.(proto.Message); ok {
		if err := protojson.Unmarshal(expectedData, message); err != nil {
			t.Fatal(err)
		}
	} else if err := json.Unmarshal(expectedData, expected); err != nil {
		t.Fatal(err)
	}
	if message, ok := expected.(proto.Message); ok {
		expectedJSON, err := json.Marshal(message)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(workflowVersionJSON(output), workflowVersionJSON(expectedJSON)) {
			t.Fatalf("result=%s, want=%s", output, expectedJSON)
		}
	} else if _, ok := reference.(*sdk.DynamicServiceInvokeResult); ok {
		actualJSON, actualErr := json.Marshal(actual)
		expectedJSON, expectedErr := json.Marshal(expected)
		if actualErr != nil || expectedErr != nil || !bytes.Equal(actualJSON, expectedJSON) {
			t.Fatalf("result=%s, want=%s", actualJSON, expectedJSON)
		}
	} else if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("result=%v, want=%v", actual, expected)
	}
}

func TestWorkflowVersionSharedHTTPContracts(t *testing.T) {
	runWorkflowSharedHTTPContracts(t, "workflow_versions_http_contract.json")
}

func TestWorkflowAppSharedHTTPContracts(t *testing.T) {
	runWorkflowSharedHTTPContracts(t, "workflow_apps_http_contract.json")
}

func TestWorkflowCandidateSharedHTTPContracts(t *testing.T) {
	runWorkflowSharedHTTPContracts(t, "workflow_candidates_http_contract.json")
}

func TestWorkflowResourceSharedHTTPContracts(t *testing.T) {
	runWorkflowSharedHTTPContracts(t, "workflow_resources_http_contract.json")
}

func TestWorkflowExecutionSharedHTTPContracts(t *testing.T) {
	runWorkflowSharedHTTPContracts(t, "workflow_executions_http_contract.json")
}

func TestWorkflowLineageSharedHTTPContracts(t *testing.T) {
	runWorkflowSharedHTTPContracts(t, "workflow_lineage_http_contract.json")
}

func runWorkflowSharedHTTPContracts(t *testing.T, fixture string) {
	data, err := os.ReadFile("../testdata/product-contracts/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	var cases []workflowVersionContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			observations := make(chan workflowContractObservation, 2)
			expectedQuery := url.Values{}
			for key, values := range scenario.Query {
				for _, value := range values {
					expectedQuery.Add(key, value)
				}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
				}
				observation := workflowContractObservation{}
				if scenario.Multipart != nil {
					multipart, readErr := readCatalogUploadCLIRequest(r)
					if readErr != nil {
						t.Errorf("read multipart: %v", readErr)
					}
					observation.multipart = &multipart
				} else {
					raw, _ := io.ReadAll(r.Body)
					if len(raw) > 0 {
						decoder := json.NewDecoder(bytes.NewReader(raw))
						decoder.UseNumber()
						if err := decoder.Decode(&observation.body); err != nil {
							t.Errorf("decode body: %v", err)
						}
					}
				}
				observations <- observation
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
			goResult, args, goErr := callWorkflowVersionGo(t, client, scenario)
			assertWorkflowContractObservation(t, <-observations, scenario, "Go")
			assertWorkflowVersionError(t, goErr, scenario)
			if scenario.Status < 400 {
				encoded, marshalErr := json.Marshal(goResult)
				if goResult == nil {
					encoded = nil
					marshalErr = nil
				}
				if marshalErr != nil {
					t.Fatal(marshalErr)
				}
				assertWorkflowVersionResult(t, encoded, goResult, scenario)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteWorkflowCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			assertWorkflowContractObservation(t, <-observations, scenario, "CLI")
			assertWorkflowVersionError(t, cliErr, scenario)
			if scenario.Status < 400 {
				assertWorkflowVersionResult(t, []byte(output), goResult, scenario)
			} else if output != "" {
				t.Fatalf("CLI error output=%q", output)
			}
		})
	}
}

func assertWorkflowContractObservation(t *testing.T, got workflowContractObservation, scenario workflowVersionContract, client string) {
	t.Helper()
	if !reflect.DeepEqual(got.body, workflowVersionJSON(scenario.Request)) {
		t.Fatalf("%s body=%#v, want=%#v", client, got.body, workflowVersionJSON(scenario.Request))
	}
	if scenario.Multipart == nil {
		if got.multipart != nil {
			t.Fatalf("%s unexpected multipart=%#v", client, got.multipart)
		}
		return
	}
	if got.multipart == nil || !reflect.DeepEqual(*got.multipart, *scenario.Multipart) {
		t.Fatalf("%s multipart=%#v, want=%#v", client, got.multipart, scenario.Multipart)
	}
}
