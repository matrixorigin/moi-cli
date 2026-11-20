package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type computeRemainingContract struct {
	Name       string              `json:"name"`
	APIID      string              `json:"api_id"`
	Operation  string              `json:"operation"`
	CLICommand string              `json:"cli_command"`
	CLIArgs    []string            `json:"cli_args"`
	Method     string              `json:"method"`
	Path       string              `json:"path"`
	Query      map[string][]string `json:"query"`
	Input      json.RawMessage     `json:"input"`
	Request    json.RawMessage     `json:"request"`
	Response   json.RawMessage     `json:"response"`
	Status     int                 `json:"status"`
	Enabled    bool                `json:"enabled"`
	Bodyless   bool                `json:"bodyless"`
	Error      struct {
		Status   int             `json:"status"`
		Response json.RawMessage `json:"response"`
	} `json:"error"`
}

func computeRemainingJSON(raw []byte) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func computeRemainingBodyMatches(scenario computeRemainingContract, body []byte) bool {
	got := computeRemainingJSON(body)
	if scenario.Bodyless {
		if got == nil {
			return true
		}
		object, ok := got.(map[string]any)
		return ok && len(object) == 0
	}
	return reflect.DeepEqual(got, computeRemainingJSON(scenario.Request))
}

func computeRemainingData(t *testing.T, raw json.RawMessage) json.RawMessage {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data
}

func computeRemainingProtoList[M proto.Message](t *testing.T, raw json.RawMessage, factory func() M) []M {
	t.Helper()
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		t.Fatal(err)
	}
	result := make([]M, 0, len(items))
	for _, item := range items {
		message := factory()
		if err := protojson.Unmarshal(item, message); err != nil {
			t.Fatal(err)
		}
		result = append(result, message)
	}
	return result
}

func computeRemainingExpected(t *testing.T, scenario computeRemainingContract) any {
	t.Helper()
	raw := computeRemainingData(t, scenario.Response)
	parse := func(message proto.Message) {
		t.Helper()
		if err := protojson.Unmarshal(raw, message); err != nil {
			t.Fatal(err)
		}
	}
	switch scenario.Operation {
	case "admin-delete", "delete":
		return nil
	case "admin-preflight-delete", "preflight-delete":
		expected := &sdk.ComputeDeletePreflightResult{}
		parse(expected)
		return expected
	case "admin-resume", "admin-retry", "admin-suspend", "resume", "retry", "suspend":
		expected := &sdk.ComputeResource{}
		parse(expected)
		return expected
	case "admin-runtime", "runtime":
		expected := &sdk.ComputeWorkerRuntimeListResult{}
		parse(expected)
		if expected.Runtimes == nil {
			return []*sdk.ComputeWorkerRuntime{}
		}
		return expected.Runtimes
	case "specs":
		return computeRemainingProtoList(t, raw, func() *sdk.ComputeResourceSpec { return &sdk.ComputeResourceSpec{} })
	case "spec-management-permission":
		expected := &sdk.ComputeResourceSpecManagementPermission{}
		parse(expected)
		return expected
	case "set-spec-enabled":
		expected := &sdk.ComputeResourceSpec{}
		parse(expected)
		return expected
	case "list":
		return computeRemainingProtoList(t, raw, func() *sdk.ComputeResource { return &sdk.ComputeResource{} })
	case "py-deps-limits":
		expected := &sdk.ComputePyDepsLimits{}
		if err := json.Unmarshal(raw, expected); err != nil {
			t.Fatal(err)
		}
		return expected
	case "worker-images":
		return computeRemainingProtoList(t, raw, func() *sdk.WorkerImage { return &sdk.WorkerImage{} })
	case "adopt-legacy-ownership":
		expected := &sdk.ComputeResourceOwnership{}
		parse(expected)
		return expected
	case "workers":
		var expected []*sdk.ComputeResourceWorker
		if err := json.Unmarshal(raw, &expected); err != nil {
			t.Fatal(err)
		}
		return expected
	default:
		t.Fatalf("unknown operation %q", scenario.Operation)
		return nil
	}
}

func computeRemainingQuery(scenario computeRemainingContract) queryParameters {
	var params queryParameters
	keys := make([]string, 0, len(scenario.Query))
	for key := range scenario.Query {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range scenario.Query[key] {
			_ = params.Set(key + "=" + value)
		}
	}
	return params
}

// computeRemainingCall exercises the typed Go SDK method that backs each
// product route. Operations that only report deletion return no payload.
func computeRemainingCall(t *testing.T, service *sdk.ComputeService, scenario computeRemainingContract) (any, error) {
	t.Helper()
	ctx := t.Context()
	params := computeRemainingQuery(scenario)
	var rawID string
	var input struct {
		ID string `json:"id"`
	}
	if len(scenario.Input) > 0 {
		if err := json.Unmarshal(scenario.Input, &input); err != nil {
			t.Fatal(err)
		}
		rawID = input.ID
	}
	switch scenario.Operation {
	case "admin-delete":
		return nil, service.AdminDelete(ctx, rawID, params...)
	case "admin-preflight-delete":
		return service.AdminPreflightDelete(ctx, rawID)
	case "admin-resume":
		return service.AdminResume(ctx, rawID)
	case "admin-retry":
		return service.AdminRetry(ctx, rawID)
	case "admin-runtime":
		return service.AdminRuntime(ctx, rawID)
	case "admin-suspend":
		return service.AdminSuspend(ctx, rawID)
	case "specs":
		return service.Specs(ctx)
	case "spec-management-permission":
		return service.SpecManagementPermission(ctx)
	case "set-spec-enabled":
		return service.SetSpecEnabled(ctx, rawID, scenario.Enabled)
	case "list":
		return service.List(ctx, params...)
	case "py-deps-limits":
		return service.PyDepsLimits(ctx)
	case "worker-images":
		return service.WorkerImages(ctx, params...)
	case "delete":
		return nil, service.Delete(ctx, rawID, params...)
	case "adopt-legacy-ownership":
		return service.AdoptLegacyOwnership(ctx, rawID)
	case "preflight-delete":
		return service.PreflightDelete(ctx, rawID)
	case "resume":
		return service.Resume(ctx, rawID)
	case "retry":
		return service.Retry(ctx, rawID)
	case "runtime":
		return service.Runtime(ctx, rawID)
	case "suspend":
		return service.Suspend(ctx, rawID)
	case "workers":
		return service.Workers(ctx, rawID)
	default:
		t.Fatalf("unknown operation %q", scenario.Operation)
		return nil, nil
	}
}

func assertComputeRemainingError(t *testing.T, err error, scenario computeRemainingContract, code, message string) {
	t.Helper()
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("error type=%T value=%v", err, err)
	}
	if failure.Code != code || failure.Status != scenario.Error.Status || failure.Message != message {
		t.Fatalf("error=%+v want code=%q status=%d message=%q", failure, code, scenario.Error.Status, message)
	}
}

func TestComputeRemainingHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/compute_remaining_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []computeRemainingContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 20 {
		t.Fatalf("cases=%d want 20", len(cases))
	}
	for _, scenario := range cases {
		scenario := scenario
		t.Run(scenario.APIID+" "+scenario.Name, func(t *testing.T) {
			handler := func(status int, payload []byte) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					expectedQuery := url.Values{}
					for key, values := range scenario.Query {
						for _, value := range values {
							expectedQuery.Add(key, value)
						}
					}
					if r.Method != scenario.Method || r.URL.EscapedPath() != "/newmoi"+scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" {
						t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
					}
					raw, _ := io.ReadAll(r.Body)
					if !computeRemainingBodyMatches(scenario, raw) {
						t.Errorf("body=%s want=%s", raw, scenario.Request)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write(payload)
				}
			}

			server := httptest.NewServer(handler(scenario.Status, scenario.Response))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			service := client.Compute("workspace-1")

			result, err := computeRemainingCall(t, service, scenario)
			if err != nil {
				t.Fatal(err)
			}
			expected := computeRemainingExpected(t, scenario)
			if (result == nil) != (expected == nil) {
				t.Fatalf("SDK result=%T want=%T", result, expected)
			}
			expectedJSON, err := json.Marshal(expected)
			if err != nil {
				t.Fatal(err)
			}
			if result != nil {
				actualJSON, err := json.Marshal(result)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(computeRemainingJSON(actualJSON), computeRemainingJSON(expectedJSON)) {
					t.Fatalf("SDK result=%s want=%s", actualJSON, expectedJSON)
				}
			}

			cliCtx := NewContext(client, "workspace-1", false, true, true)
			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteComputeCommand(cliCtx, scenario.CLIArgs)
			})
			if cliErr != nil {
				t.Fatal(cliErr)
			}
			if result == nil {
				want := "compute " + scenario.Operation + ": ok"
				if strings.TrimSpace(output) != want {
					t.Fatalf("CLI output=%q want=%q", output, want)
				}
			} else if !reflect.DeepEqual(computeRemainingJSON([]byte(output)), computeRemainingJSON(expectedJSON)) {
				t.Fatalf("CLI output=%q want=%s", output, expectedJSON)
			}

			var envelope struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(scenario.Error.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			errorServer := httptest.NewServer(handler(scenario.Error.Status, scenario.Error.Response))
			defer errorServer.Close()
			errorClient, err := sdk.New(errorServer.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer errorClient.Close()
			_, err = computeRemainingCall(t, errorClient.Compute("workspace-1"), scenario)
			assertComputeRemainingError(t, err, scenario, envelope.Code, envelope.Message)

			errorCLICtx := NewContext(errorClient, "workspace-1", false, true, true)
			var errorCLIErr error
			errorOutput := captureStdout(t, func() {
				errorCLIErr = ExecuteComputeCommand(errorCLICtx, scenario.CLIArgs)
			})
			assertComputeRemainingError(t, errorCLIErr, scenario, envelope.Code, envelope.Message)
			if errorOutput != "" {
				t.Fatalf("CLI error output=%q", errorOutput)
			}
		})
	}
}
