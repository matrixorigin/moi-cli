package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestAgentResourceMetadataExport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/workspaces/workspace-1/agent-builder/runtime" {
			t.Errorf("unexpected route: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"models":[{"name":"Example","model_name":"example-model"}],"skills":[{"id":"skill-1","workspace_id":"system","name":"Kubernetes","description":"Inspect clusters"}],"tools":[],"knowledge_bases":[]}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL, sdk.WithAPIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	target := filepath.Join(t.TempDir(), "resources.json")
	cli := NewContext(client, "workspace-1", false, true, true)
	if err := ExecuteAgentCommand(cli, []string{"list-resource-meta", "--output-file", target}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var got, expected map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{"schema_version":1,"models":[{"name":"Example","model_name":"example-model"}],"skills":[{"id":"skill-1","workspace_id":"system","name":"Kubernetes","description":"Inspect clusters"}],"tools":[],"knowledge_bases":[]}`), &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("export changed the resource contract: %s", data)
	}
	info, err := os.Stat(target)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("export must remain private: %v %v", info, err)
	}
}

func TestAgentBuilderContextAcceptsExactProductNumbers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"schema_version":1,"snapshot_version":9007199254740993,"operations":{"create":{"allowed":true},"revise_current":{"allowed":false}}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL, sdk.WithAPIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	target := filepath.Join(t.TempDir(), "context.json")
	captureStdout(t, func() {
		err = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), []string{"builder", "context", "--output-file", target})
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got["schema_version"] != json.Number("1") || got["snapshot_version"] != json.Number("9007199254740993") {
		t.Fatalf("context numbers changed: %#v", got)
	}
}

func TestAgentBuilderMalformedResponsePreservesExistingFile(t *testing.T) {
	for _, command := range []string{"context", "propose"} {
		t.Run(command, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
			}))
			defer server.Close()
			client, err := sdk.New(server.URL, sdk.WithAPIKey("test-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			directory := t.TempDir()
			target := filepath.Join(directory, "result.json")
			original := []byte(`{"previous":"complete result"}`)
			if err := os.WriteFile(target, original, 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"builder", command, "--output-file", target}
			if command == "propose" {
				input := filepath.Join(directory, "proposal.json")
				if err := os.WriteFile(input, []byte(`{"config":{}}`), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--json-file", input)
			}
			err = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), args)
			if !sdk.IsCode(err, "CLI_AGENT_BUILDER_RESPONSE_INVALID") {
				t.Fatalf("malformed response must be an explicit failure: %v", err)
			}
			got, err := os.ReadFile(target)
			if err != nil || string(got) != string(original) {
				t.Fatalf("malformed response replaced existing file: %s %v", got, err)
			}
		})
	}
}

func TestAgentResourceMetadataFailedExportPreservesExistingFile(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{"denied", http.StatusForbidden, `{"code":"PermissionDenied","message":"permission denied"}`},
		{"incomplete_contract", http.StatusOK, `{"code":0,"data":{"models":[],"skills":[],"tools":[]}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			client, err := sdk.New(server.URL, sdk.WithAPIKey("test-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			target := filepath.Join(t.TempDir(), "resources.json")
			original := []byte(`{"previous":"complete export"}`)
			if err := os.WriteFile(target, original, 0600); err != nil {
				t.Fatal(err)
			}
			err = ExecuteAgentCommand(NewContext(client, "workspace-1", false, true, true), []string{"list-resource-meta", "--output-file", target})
			if err == nil {
				t.Fatal("failed export reported success")
			}
			got, readErr := os.ReadFile(target)
			if readErr != nil || string(got) != string(original) {
				t.Fatalf("failed export replaced existing file: %s %v", got, readErr)
			}
		})
	}
}
