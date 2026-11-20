package nativeauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The test executable implements the real inherited-FD subprocess protocol.
// These are synthetic credentials, never a local user's credential store.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		pipe := os.NewFile(3, "credential-pipe")
		if pipe == nil {
			os.Exit(20)
		}
		defer pipe.Close()
		value := fixtureCredential()
		switch os.Getenv("NATIVE_AUTH_TEST_HELPER") {
		case "failure":
			os.Stderr.WriteString("private-token-must-not-escape")
			os.Exit(1)
		case "oversized":
			io.WriteString(pipe, strings.Repeat("x", maxCredentialBytes+1))
			os.Exit(0)
		case "trailing":
			json.NewEncoder(pipe).Encode(value)
			io.WriteString(pipe, "{}")
			os.Exit(0)
		case "unknown":
			io.WriteString(pipe, `{"secret_extra":"must-not-escape"}`)
			os.Exit(0)
		case "switched":
			value.Subject = "account-B"
		case "generation":
			value.Generation = "generation-B"
		case "endpoint":
			value.Endpoint = "https://other.example.test/newmoi"
		case "expired":
			value.ExpiresAt = 1
		case "hang":
			time.Sleep(time.Minute)
		case "slow":
			time.Sleep(150 * time.Millisecond)
		case "rotation":
			// A valid rotation may spend time waiting for another process and
			// then perform token exchange plus signing-key verification.
			time.Sleep(26 * time.Second)
		}
		json.NewEncoder(pipe).Encode(value)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func fixtureCredential() Credential {
	return Credential{Version: 1, Environment: "test-environment", Issuer: "https://uc.example.test/realms/moi", Subject: "account-A", Generation: "generation-A", Endpoint: "https://moi.example.test/newmoi", ExpiresAt: time.Now().Add(time.Hour).Unix(), AccessToken: "synthetic-access-token"}
}

func testHelper(t *testing.T) *Helper {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return &Helper{Binary: binary}
}

func TestHelperPrivatePipe(t *testing.T) {
	t.Setenv("NATIVE_AUTH_TEST_HELPER", "ok")
	value, err := testHelper(t).Credential(t.Context(), "generation-A")
	if err != nil {
		t.Fatal(err)
	}
	if value.Subject != "account-A" || value.AccessToken != "synthetic-access-token" {
		t.Fatal("private-pipe contract mismatch")
	}
	if strings.Contains(value.String(), value.AccessToken) {
		t.Fatal("credential debug representation exposes token")
	}
}

func TestHelperFailsClosed(t *testing.T) {
	for _, mode := range []string{"failure", "oversized", "trailing", "unknown", "generation", "expired"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("NATIVE_AUTH_TEST_HELPER", mode)
			_, err := testHelper(t).Credential(t.Context(), "generation-A")
			if err == nil {
				t.Fatal("accepted invalid helper response")
			}
			if strings.Contains(err.Error(), "must-not-escape") || strings.Contains(err.Error(), "synthetic-access-token") {
				t.Fatal("helper output leaked into diagnostics")
			}
		})
	}
}

func TestHelperCancellation(t *testing.T) {
	t.Setenv("NATIVE_AUTH_TEST_HELPER", "hang")
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if _, err := testHelper(t).Credential(ctx, ""); err == nil {
		t.Fatal("cancelled helper succeeded")
	}
}

func TestHelperAllowsCredentialOwnerRotationBudget(t *testing.T) {
	t.Setenv("NATIVE_AUTH_TEST_HELPER", "rotation")
	value, err := testHelper(t).Credential(t.Context(), "generation-A")
	if err != nil {
		t.Fatal(err)
	}
	if value.Generation != "generation-A" {
		t.Fatal("rotation changed the login generation")
	}
}

func TestCredentialRejectsUnsafeIssuer(t *testing.T) {
	for _, issuer := range []string{"not-a-url", "http://uc.example.test", "https://user@uc.example.test", "https://uc.example.test/realms/../admin", "https://uc.example.test/?secret=value"} {
		credential := fixtureCredential()
		credential.Issuer = issuer
		if err := credential.validate(); err == nil {
			t.Fatalf("unsafe issuer accepted: %s", issuer)
		}
	}
}

func TestTransportRejectsPathEscapeBeforeRequestingCredentials(t *testing.T) {
	for _, address := range []string{"https://moi.example.test/newmoi/../other", "https://moi.example.test/newmoi/%2e%2e/other", "https://moi.example.test/newmoi/./test"} {
		request, _ := http.NewRequestWithContext(t.Context(), "GET", address, nil)
		transport := &Transport{Binding: fixtureCredential()}
		if _, err := transport.RoundTrip(request); err == nil {
			t.Fatal("endpoint path escape accepted")
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func TestBusinessDeadlineStartsAfterHelperAndBodyCloseCancels(t *testing.T) {
	t.Setenv("NATIVE_AUTH_TEST_HELPER", "slow")
	var businessContext context.Context
	transport := &Transport{Helper: testHelper(t), Binding: fixtureCredential(), RequestTimeout: 100 * time.Millisecond,
		Base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			businessContext = req.Context()
			deadline, ok := req.Context().Deadline()
			if !ok || time.Until(deadline) < 75*time.Millisecond {
				t.Fatal("credential acquisition consumed the business deadline")
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("ok")), Header: make(http.Header)}, nil
		}),
	}
	req, _ := http.NewRequestWithContext(t.Context(), "GET", fixtureCredential().Endpoint+"/test", nil)
	response, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	if businessContext.Err() != nil {
		t.Fatal("response context cancelled before body consumed")
	}
	response.Body.Close()
	if businessContext.Err() != context.Canceled {
		t.Fatal("body close did not release deadline")
	}
}

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTransportFreezesIdentityAndDoesNotReplay(t *testing.T) {
	for _, mode := range []string{"ok", "switched", "generation", "endpoint"} {
		t.Run(mode, func(t *testing.T) {
			t.Setenv("NATIVE_AUTH_TEST_HELPER", mode)
			calls := 0
			transport := &Transport{Helper: testHelper(t), Binding: fixtureCredential(), Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Header.Get("Authorization") != "Bearer synthetic-access-token" {
					t.Fatal("missing fresh bearer")
				}
				return &http.Response{StatusCode: http.StatusUnauthorized, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
			})}
			request, err := http.NewRequestWithContext(t.Context(), "POST", "https://moi.example.test/newmoi/workspaces", nil)
			if err != nil {
				t.Fatal(err)
			}
			response, err := transport.RoundTrip(request)
			if mode == "ok" {
				if err != nil || calls != 1 || response.StatusCode != http.StatusUnauthorized {
					t.Fatalf("unexpected request/replay result: calls=%d, error=%v", calls, err)
				}
				response.Body.Close()
			} else if err == nil || calls != 0 {
				t.Fatal("changed binding reached product API")
			}
			if request.Header.Get("Authorization") != "" {
				t.Fatal("mutated original request")
			}
		})
	}
}

func TestTransportRejectsOtherDestinationsAndMixedCredentials(t *testing.T) {
	t.Setenv("NATIVE_AUTH_TEST_HELPER", "ok")
	for _, address := range []string{"https://evil.example.test/newmoi/test", "http://moi.example.test/newmoi/test", "https://moi.example.test/newmoi-evil/test", "https://moi.example.test/other"} {
		request, _ := http.NewRequestWithContext(t.Context(), "GET", address, nil)
		transport := &Transport{Helper: testHelper(t), Binding: fixtureCredential()}
		if _, err := transport.RoundTrip(request); err == nil {
			t.Fatalf("accepted other destination: %s", address)
		}
	}
	for _, header := range []string{"Cookie", "X-API-Key"} {
		request, _ := http.NewRequestWithContext(t.Context(), "GET", fixtureCredential().Endpoint+"/test", nil)
		request.Header.Set(header, "old-credential")
		transport := &Transport{Helper: testHelper(t), Binding: fixtureCredential()}
		if _, err := transport.RoundTrip(request); err == nil {
			t.Fatal("accepted mixed credentials")
		}
	}
}

func TestSelectedDoesNotReadStoreOrFallBackAfterLogout(t *testing.T) {
	root := t.TempDir()
	t.Setenv("MOI_AUTH_DIR", root)
	if selected, err := Selected("auto"); err != nil || !selected {
		t.Fatal("empty auth directory fell back to PAT")
	}
	for _, content := range []string{`{"version":1,"active":"env","sessions":{"env":null}}`, "corrupt"} {
		if err := os.WriteFile(filepath.Join(root, "auth.json"), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if selected, err := Selected("auto"); err != nil || !selected {
			t.Fatal("tombstone/corrupt store fell back to PAT")
		}
		if selected, err := Selected("pat"); err != nil || selected {
			t.Fatal("explicit legacy mode changed")
		}
	}
}
