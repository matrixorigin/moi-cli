package commands

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestConnectorTasksPassesOwnerFiltersAndPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/newmoi/task/list" || r.URL.RawQuery != "connector_sources=0&connector_sources=17&is_desc=false&keyword=&load_interval_types=0&load_interval_types=4&order_by=created_at&page=0&page_size=0&source_connector_ids=&source_connector_ids=connector%2F1&status=0&status=2" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL)
		}
		fmt.Fprint(w, `{"code":"OK","data":{"total":0,"tasks":[]}}`)
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("test-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-test", false, true, true)
	if err := runConnectorServiceTasks(ctx, []string{"--connector-source", "0", "--connector-source", "17", "--is-desc=false", "--keyword=", "--load-interval-type", "0", "--load-interval-type", "4", "--order-by", "created_at", "--page", "0", "--page-size", "0", "--source-connector-id", "", "--source-connector-id", "connector/1", "--status", "0", "--status", "2"}); err != nil {
		t.Fatal(err)
	}
}

func TestConnectorTasksPreservesParameterError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept-Language") != "en-US" {
			t.Errorf("missing locale")
		}
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"code":"ErrParamInvalid","msg":"Invalid parameters. Please check and try again."}`)
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithCustomHeader("Accept-Language", "en-US"))
	if err != nil {
		t.Fatal(err)
	}
	err = runConnectorServiceTasks(NewContext(client, "ws", false, true, true), []string{"--status", "0"})
	var failure *sdk.Error
	if !errors.As(err, &failure) || failure.Code != "ErrParamInvalid" || failure.Status != 400 || failure.Message != "Invalid parameters. Please check and try again." {
		t.Fatalf("public parameter error changed: %#v", err)
	}
}
