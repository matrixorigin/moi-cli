package main

import (
	"errors"
	"net/http"

	"github.com/matrixorigin/matrixflow/sdk/cli/internal/trace"
)

var (
	errRuntimeIdentityInvalid = errors.New("CLI_RUNTIME_IDENTITY_INVALID: runtime connection values must be trimmed and free of spaces, tabs, newlines or backslashes")
)

func runtimeHTTPClient() *http.Client {
	return &http.Client{
		Transport:     runtimeTransport{},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// The SDK omits workspace scope for account operations. Those routes use the
// same user's Product identity without a task proof; MOI owns authorization.
type runtimeTransport struct{}

func (transport runtimeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Header.Get("X-Workspace-ID") == "" {
		request = request.Clone(request.Context())
		request.Header.Del("X-MOI-Runtime-Authorization")
	}
	trace.Inject(request)
	return http.DefaultTransport.RoundTrip(request)
}
