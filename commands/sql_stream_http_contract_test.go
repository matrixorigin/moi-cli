package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type sqlStreamHTTPContract struct {
	Name                string          `json:"name"`
	Operation           string          `json:"operation"`
	Method              string          `json:"method"`
	Path                string          `json:"path"`
	Request             json.RawMessage `json:"request"`
	ResponseBase64      string          `json:"response_base64"`
	ResponseContentType string          `json:"response_content_type"`
	Status              int             `json:"status"`
	StatementID         string          `json:"statement_id"`
	AgentID             string          `json:"agent_id"`
	FileID              string          `json:"file_id"`
}

func sqlStreamContractJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func TestSQLStreamSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/sql_stream_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []sqlStreamHTTPContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			response, err := base64.StdEncoding.DecodeString(scenario.ResponseBase64)
			if err != nil {
				t.Fatal(err)
			}
			observations := make(chan struct {
				method, path, workspace, apiKey string
				body                            []byte
			}, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				observations <- struct {
					method, path, workspace, apiKey string
					body                            []byte
				}{r.Method, r.URL.EscapedPath(), r.Header.Get("X-Workspace-ID"), r.Header.Get("X-API-Key"), body}
				w.Header().Set("Content-Type", scenario.ResponseContentType)
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			var stream *sdk.StreamResponse
			if scenario.Operation == "download-query-result" {
				stream, err = client.SQL("workspace-1").DownloadQueryResult(t.Context(), scenario.StatementID)
			} else {
				stream, err = client.Agents("workspace-1").PreviewQueryVisual(t.Context(), scenario.AgentID, scenario.FileID)
			}
			if err != nil {
				t.Fatal(err)
			}
			gotSDK, readErr := io.ReadAll(stream.Body)
			closeErr := stream.Body.Close()
			if readErr != nil || closeErr != nil || !bytes.Equal(gotSDK, response) {
				t.Fatalf("SDK stream=%q read=%v close=%v", gotSDK, readErr, closeErr)
			}
			observation := <-observations
			if observation.method != scenario.Method || observation.path != scenario.Path || observation.workspace != "workspace-1" || observation.apiKey != "synthetic-key" || !reflect.DeepEqual(sqlStreamContractJSON(observation.body), sqlStreamContractJSON(scenario.Request)) {
				t.Fatalf("SDK observation=%+v body=%s", observation, observation.body)
			}

			output := filepath.Join(t.TempDir(), "result.bin")
			ctx := NewContext(client, "workspace-1", false, true, true)
			if scenario.Operation == "download-query-result" {
				err = ExecuteSQLCommand(ctx, []string{"download-query-result", "--statement-id", scenario.StatementID, "--output", output})
			} else {
				err = ExecuteAgentCommand(ctx, []string{"preview-query-visual", "--agent-id", scenario.AgentID, "--file-id", scenario.FileID, "--output", output})
			}
			if err != nil {
				t.Fatal(err)
			}
			gotCLI, err := os.ReadFile(output)
			if err != nil || !bytes.Equal(gotCLI, response) {
				t.Fatalf("CLI stream=%q want=%q error=%v", gotCLI, response, err)
			}
			observation = <-observations
			if observation.method != scenario.Method || observation.path != scenario.Path || observation.workspace != "workspace-1" || observation.apiKey != "synthetic-key" || !reflect.DeepEqual(sqlStreamContractJSON(observation.body), sqlStreamContractJSON(scenario.Request)) {
				t.Fatalf("CLI observation=%+v body=%s", observation, observation.body)
			}
		})
	}
}
