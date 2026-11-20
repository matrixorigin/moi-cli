package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestKnowledgeSourceGovernanceCLI(t *testing.T) {
	for _, scenario := range []struct {
		name string
		args []string
		body string
	}{
		{"disable", []string{"--enabled=false"}, `{"enabled":false}`},
		{"enable", []string{"--enabled=true"}, `{"enabled":true}`},
		{"tags and expiry", []string{"--tags", `["finance","policy"]`, "--expires-at", "1900000000"}, `{"tags":["finance","policy"],"expires_at":1900000000}`},
		{"clear tags", []string{"--tags", `[]`}, `{"tags":[]}`},
		{"clear expiry", []string{"--clear-expires-at"}, `{"expires_at":null}`},
		{"explicit zero", []string{"--expires-at", "0"}, `{"expires_at":0}`},
		{"force expired", []string{"--force-enabled-after-expiry=true"}, `{"force_enabled_after_expiry":true}`},
		{"remove force", []string{"--force-enabled-after-expiry=false"}, `{"force_enabled_after_expiry":false}`},
		{"omitted fields", nil, `{}`},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch || r.URL.Path != "/newmoi/semantic-models/17/sources/source-1/governance" || r.Header.Get("X-Workspace-ID") != "workspace" {
					t.Errorf("unexpected governance target: %s %s", r.Method, r.URL.Path)
				}
				var got, want map[string]any
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Error(err)
				}
				if err := json.Unmarshal([]byte(scenario.body), &want); err != nil {
					t.Error(err)
					return
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("governance patch = %#v, want %#v", got, want)
				}
				_, _ = w.Write([]byte(`{"code":"OK","data":{"source":{"row_id":"source-1"}}}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := append([]string{"update-source-governance", "--model-id", "17", "--source-id", "source-1"}, scenario.args...)
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteSemanticModelCommand(NewContext(client, "workspace", false, true, true), args)
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			var receipt struct {
				Source struct {
					RowID string `json:"row_id"`
				} `json:"source"`
			}
			if err := json.Unmarshal([]byte(output), &receipt); err != nil || receipt.Source.RowID != "source-1" {
				t.Fatalf("governance receipt = %s, error = %v", output, err)
			}
		})
	}
}

func TestKnowledgeSourceGovernanceCLIRejectsInvalidInputBeforeRequest(t *testing.T) {
	for _, args := range [][]string{
		{"--tags", `null`},
		{"--tags", `{}`},
		{"--tags", `[1]`},
		{"--tags", `[`},
		{"--expires-at", "1900000000", "--clear-expires-at"},
	} {
		t.Run(args[0]+"/"+args[1], func(t *testing.T) {
			contacted := false
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { contacted = true }))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			err = ExecuteSemanticModelCommand(NewContext(client, "workspace", false, true, true), append([]string{
				"update-source-governance", "--model-id", "17", "--source-id", "source-1",
			}, args...))
			if err == nil || contacted {
				t.Fatalf("invalid patch reached owner: err=%v contacted=%v", err, contacted)
			}
		})
	}
}

func TestKnowledgeSourceGovernanceCLIPreservesOwnerDenial(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"ErrForbidden","msg":"governance denied"}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	var callErr error
	output := captureStdout(t, func() {
		callErr = ExecuteSemanticModelCommand(NewContext(client, "workspace", false, true, true), []string{
			"update-source-governance", "--model-id", "17", "--source-id", "source-1", "--enabled=false",
		})
	})
	var denied *sdk.Error
	if !errors.As(callErr, &denied) || denied.Code != "ErrForbidden" || denied.Status != http.StatusForbidden || output != "" {
		t.Fatalf("owner denial was changed: err=%v output=%s", callErr, output)
	}
}
