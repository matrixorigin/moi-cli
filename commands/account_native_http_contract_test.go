package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type accountNativeCLIContract struct {
	Name     string          `json:"name"`
	Method   string          `json:"method"`
	Path     string          `json:"path"`
	Response json.RawMessage `json:"response"`
	Status   int             `json:"status"`
}

func accountNativeCLIJSON(raw json.RawMessage) any {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	_ = decoder.Decode(&value)
	return value
}

func TestAccountNativeCLISharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/account_native_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []accountNativeCLIContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || r.Header.Get("Authorization") != "Bearer native-token" || r.Header.Get("X-API-Key") != "" || r.Header.Get("X-Workspace-ID") != "" || r.Header.Get("Cookie") != "" || len(raw) != 0 {
					t.Errorf("request=%s %s body=%q headers=%v", r.Method, r.URL.RequestURI(), raw, r.Header)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("native-token"))
			if err != nil {
				t.Fatal(err)
			}
			read, write, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			original := os.Stdout
			os.Stdout = write
			callErr := ExecuteAccountCommand(NewContext(client, "", false, true, true), []string{"bootstrap-native"})
			os.Stdout = original
			closeErr := write.Close()
			output, readErr := io.ReadAll(read)
			readCloseErr := read.Close()
			if err := errors.Join(callErr, closeErr, readErr, readCloseErr); err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				Data json.RawMessage `json:"data"`
			}
			if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
				t.Fatal(err)
			}
			if got := accountNativeCLIJSON(output); !reflect.DeepEqual(got, accountNativeCLIJSON(envelope.Data)) {
				t.Fatalf("output=%#v want=%#v", got, accountNativeCLIJSON(envelope.Data))
			}
		})
	}
}
