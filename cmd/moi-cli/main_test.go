package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"os"
	"testing"
)

func TestCheckCommandHelpSkipsGlobalOptionValues(t *testing.T) {
	originalArgs := os.Args
	t.Cleanup(func() {
		os.Args = originalArgs
	})
	os.Args = []string{
		"moi-cli",
		"-endpoint", "http://127.0.0.1:8050/newmoi",
		"-api-key", "opaque-pat",
		"-workspace-id", "ws-a",
		"workflow",
		"--help",
	}

	command, args, showHelp := checkCommandHelp()
	if command != "workflow" || len(args) != 0 || !showHelp {
		t.Fatalf("checkCommandHelp() = %q, %#v, %t", command, args, showHelp)
	}
}

func TestJSONCommandErrorsPreserveProductContract(t *testing.T) {
	for _, locale := range []string{"zh-CN", "en-US"} {
		message := "Refresh and try again."
		if locale == "zh-CN" {
			message = "请刷新后再试。"
		}
		for _, code := range []string{"4", "ErrResponseInvalid", "ErrTransportUnavailable"} {
			t.Run(locale+"/"+code, func(t *testing.T) {
				original := &sdk.Error{Code: code, Message: message, Status: 409, RequestID: "request-1", Details: map[string]any{"reason": "PLUGIN_STATE_CONFLICT", "domain": "moi-core.agent-plugin"}}
				var output bytes.Buffer
				writeCommandError(&output, fmt.Errorf("command: %w", original), true)
				var parsed sdk.Error
				if err := json.Unmarshal(output.Bytes(), &parsed); err != nil {
					t.Fatal(err)
				}
				if parsed.Code != original.Code || parsed.Message != message || parsed.Status != 409 || parsed.RequestID != original.RequestID || parsed.Details["reason"] != original.Details["reason"] {
					t.Fatalf("command discarded owner contract: %+v", parsed)
				}
			})
		}
	}
}

func TestGlobalOptionTakesValueDoesNotIncludeRemovedRoleOverride(t *testing.T) {
	if globalOptionTakesValue("-role-id") {
		t.Fatal("removed -role-id must not be a global option")
	}
}

func TestRuntimeChildConfigurationUsesInjectedEnvironment(t *testing.T) {
	t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer task-proof")
	t.Setenv("MOI_API_KEY", "task-api-key")
	t.Setenv("MOI_ENDPOINT", "http://moi.example/newmoi")
	t.Setenv("MOI_WORKSPACE_ID", "workspace-1")

	cfg, proof, err := runtimeChildConfiguration()
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || proof != "Bearer task-proof" || cfg.APIKey != "task-api-key" ||
		cfg.Endpoint != "http://moi.example/newmoi" || cfg.WorkspaceID != "workspace-1" {
		t.Fatalf("runtime child configuration = %#v, proof %q", cfg, proof)
	}
}

func TestRuntimeChildConfigurationRejectsIncompleteEnvironment(t *testing.T) {
	t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer task-proof")
	t.Setenv("MOI_API_KEY", "")
	t.Setenv("MOI_ENDPOINT", "http://moi.example/newmoi")
	t.Setenv("MOI_WORKSPACE_ID", "workspace-1")

	cfg, proof, err := runtimeChildConfiguration()
	if cfg != nil || proof != "" || err == nil {
		t.Fatalf("runtimeChildConfiguration() = %#v, %q, %v", cfg, proof, err)
	}
}

func TestTextCommandErrorsPreserveOwnerDetails(t *testing.T) {
	original := &sdk.Error{Code: "ErrParamInvalid", Message: "Invalid parameter", Status: 400, RequestID: "request-1", Details: map[string]any{"path": "structured_load_config.mapping[0].allow_duplicate_source"}}
	var output bytes.Buffer
	writeCommandError(&output, fmt.Errorf("command: %w", original), false)
	lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
	if !bytes.Contains(lines[0], []byte(original.Message)) || !bytes.Contains(lines[0], []byte(original.Code)) {
		t.Fatalf("missing owner error: %s", output.Bytes())
	}
	if len(lines) < 2 {
		t.Fatalf("missing owner details: %s", output.Bytes())
	}
	var details map[string]string
	if err := json.Unmarshal(lines[1], &details); err != nil {
		t.Fatal(err)
	}
	if details["path"] != original.Details["path"] {
		t.Fatalf("owner field path was lost: %v", details)
	}
}
