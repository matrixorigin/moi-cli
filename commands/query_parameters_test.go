package commands

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestChannelListPreservesProductQueryParameters(t *testing.T) {
	var got url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/newmoi/workspaces/workspace/channels/github/instances" {
			t.Errorf("unexpected channel route: %s", r.URL.Path)
		}
		got = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"items":[],"total":0}}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-user-key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx := NewContext(client, "workspace", false, true, true)
	captureStdout(t, func() {
		err = ExecuteAgentCommand(ctx, []string{"list-channel-instances", "--provider", "github", "--query-param", "limit=20", "--query-param", "offset=20", "--query-param", "status=active", "--query-param", "name=a=b", "--query-param", "tag=one", "--query-param", "tag=two"})
	})
	if err != nil {
		t.Fatal(err)
	}
	want := url.Values{"limit": {"20"}, "offset": {"20"}, "status": {"active"}, "name": {"a=b"}, "tag": {"one", "two"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("product query = %#v, want %#v", got, want)
	}
}
