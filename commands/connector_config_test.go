package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestConnectorConfigurationInput(t *testing.T) {
	const config = `{"mysql":{"host":"db.example.test","port":3306,"username":"reader","password":"synthetic-password"}}`
	for _, operation := range []string{"create", "validate", "update"} {
		for _, source := range []string{"file", "stdin"} {
			t.Run(operation+"/"+source, func(t *testing.T) {
				var got map[string]any
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					path, method := "/newmoi/connectors", http.MethodPost
					if operation == "validate" {
						path += "/validate"
					} else if operation == "update" {
						path, method = path+"/connector-id", http.MethodPut
					}
					if r.URL.Path != path || r.Method != method {
						t.Errorf("request %s %s, want %s %s", r.Method, r.URL.Path, method, path)
					}
					if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
						t.Error(err)
					}
					w.Header().Set("Content-Type", "application/json")
					if operation == "validate" {
						_, _ = w.Write([]byte(`{"code":0,"data":{"valid":true}}`))
					} else {
						_, _ = w.Write([]byte(`{"code":0,"data":{"id":"connector-id"}}`))
					}
				}))
				t.Cleanup(server.Close)
				client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = client.Close() })
				path := filepath.Join(t.TempDir(), "config.json")
				if err := os.WriteFile(path, []byte(config), 0600); err != nil {
					t.Fatal(err)
				}
				if source == "stdin" {
					file, err := os.Open(path)
					if err != nil {
						t.Fatal(err)
					}
					previous := os.Stdin
					os.Stdin = file
					t.Cleanup(func() { os.Stdin = previous; _ = file.Close() })
					path = "-"
				}
				args := []string{operation, "--name", "configured", "--config-file", path}
				if operation == "update" {
					args = append(args, "--connector-id", "connector-id")
				} else {
					args = append(args, "--source-type", "9", "--usage-type", "1")
				}
				output := captureStdout(t, func() { err = ExecuteConnectorCommand(NewContext(client, "workspace", false, true, true), args) })
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(output, "synthetic-password") {
					t.Fatal("CLI echoed the submitted credential")
				}
				var wantConfig map[string]any
				if err := json.Unmarshal([]byte(config), &wantConfig); err != nil {
					t.Fatal(err)
				}
				want := map[string]any{"name": "configured", "config": wantConfig}
				if operation != "update" {
					want["source_type"], want["usage_type"] = float64(9), float64(1)
				}
				if !reflect.DeepEqual(want, got) {
					t.Fatalf("request fields do not preserve configuration or omitted update fields: got %v, want %v", got, want)
				}
			})
		}
	}
}

func TestConnectorMalformedConfigurationDoesNotSendRequest(t *testing.T) {
	for _, body := range []string{"[]", `{"mysql":`, `"configuration"`, ""} {
		t.Run(body, func(t *testing.T) {
			contacted := false
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { contacted = true }))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			err = ExecuteConnectorCommand(NewContext(client, "workspace", false, true, true), []string{"update", "--connector-id", "connector-id", "--name", "invalid", "--config-file", path})
			if err == nil || contacted {
				t.Fatalf("malformed configuration must fail before sending an update: err=%v contacted=%v", err, contacted)
			}
		})
	}
}

func TestConnectorValidationPreservesExplicitOutcome(t *testing.T) {
	for _, body := range []string{`{"valid":false}`, `{"valid":true}`, `{}`} {
		t.Run(body, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/newmoi/connectors/validate" {
					t.Errorf("unexpected validation request: %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":"OK","data":` + body + `}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			output := captureStdout(t, func() {
				err = ExecuteConnectorCommand(NewContext(client, "workspace", false, true, true), []string{"validate", "--name", "probe", "--source-type", "3", "--usage-type", "2"})
			})
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]any
			if err := json.Unmarshal([]byte(output), &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(body), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("CLI changed the validation outcome: got %v, want %v", got, want)
			}
		})
	}
}
