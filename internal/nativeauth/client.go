// Package nativeauth consumes Astra's versioned credential-helper protocol.
// Only Astra reads or rotates the shared credential store.
package nativeauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const maxCredentialBytes = 64 << 10

type Credential struct {
	Version     int     `json:"version"`
	Environment string  `json:"environment"`
	Issuer      string  `json:"issuer"`
	Subject     string  `json:"subject"`
	Generation  string  `json:"generation"`
	Endpoint    string  `json:"endpoint"`
	ExpiresAt   int64   `json:"expires_at"`
	AccessToken string  `json:"access_token"`
	WorkspaceID *string `json:"workspace_id"`
	RoleID      *string `json:"role_id"`
}

// Selected delegates default authentication to Astra, even before first login.
// Only an explicit PAT selection may consume legacy credentials.
func Selected(mode string) (bool, error) {
	switch mode {
	case "auto", "moi":
		return true, nil
	case "pat":
		return false, nil
	default:
		return false, errors.New("auth must be auto, moi, or pat")
	}
}

type Helper struct{ Binary string }

func NewHelper() (*Helper, error) {
	name := os.Getenv("MOI_ASTRA_BINARY")
	if name == "" {
		name = "astra"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, errors.New("Astra is required for MOI authentication; install Astra and run astra login")
	}
	if !filepath.IsAbs(path) {
		return nil, errors.New("MOI_ASTRA_BINARY must resolve to an absolute executable path")
	}
	return &Helper{Binary: path}, nil
}

// Credential uses an inherited private pipe. Tokens never appear in argv,
// environment variables, stdout or stderr, including on helper failure.
func (h *Helper) Credential(ctx context.Context, generation string) (Credential, error) {
	// NativeStore can wait 20s for rotation, then spend up to 15s each on
	// token exchange, signing-key validation and orphan-token revocation.
	// Allow the owner to finish its bounded transaction before terminating it;
	// killing it after exchange would unnecessarily leave refresh_pending set.
	// A caller's earlier deadline/cancellation still takes precedence.
	ctx, cancel := context.WithTimeout(ctx, 75*time.Second)
	defer cancel()
	reader, writer, err := os.Pipe()
	if err != nil {
		return Credential{}, err
	}
	defer reader.Close()
	defer writer.Close()
	args := []string{"auth", "credential", "--target", "moi", "--output-fd", "3"}
	if generation != "" {
		args = append(args, "--generation", generation)
	}
	cmd := exec.CommandContext(ctx, h.Binary, args...)
	cmd.ExtraFiles = []*os.File{writer}
	// nil output streams go to the null device; do not include subprocess
	// diagnostics in errors because they are not a credential-safe protocol.
	if err := cmd.Start(); err != nil {
		return Credential{}, errors.New("cannot start Astra credential helper")
	}
	writer.Close()
	// Cancel also closes the reader: a broken helper's descendant could retain
	// fd 3 even after the helper itself has exited.
	stop := context.AfterFunc(ctx, func() { reader.Close() })
	defer stop()
	body, readErr := io.ReadAll(io.LimitReader(reader, maxCredentialBytes+1))
	if readErr != nil || len(body) > maxCredentialBytes {
		cancel()
	}
	waitErr := cmd.Wait()
	if readErr != nil || waitErr != nil || len(body) > maxCredentialBytes {
		return Credential{}, errors.New("MOI credential unavailable; run astra auth status or astra login")
	}
	var credential Credential
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&credential); err != nil {
		return Credential{}, errors.New("invalid Astra credential response")
	}
	if decoder.Decode(new(any)) != io.EOF {
		return Credential{}, errors.New("invalid Astra credential response")
	}
	if err := credential.validate(); err != nil {
		return Credential{}, err
	}
	if generation != "" && credential.Generation != generation {
		return Credential{}, errors.New("MOI account changed during command")
	}
	return credential, nil
}

func (c Credential) validate() error {
	if c.Version != 1 || c.Environment == "" || c.Issuer == "" || c.Subject == "" || c.Generation == "" ||
		c.AccessToken == "" || len(c.AccessToken) > 16<<10 || strings.ContainsAny(c.AccessToken, " \t\r\n") || c.ExpiresAt <= time.Now().Unix() {
		return errors.New("invalid or expired Astra credential response")
	}
	if _, err := secureURL(c.Issuer); err != nil {
		return errors.New("invalid credential issuer")
	}
	_, err := secureURL(c.Endpoint)
	return err
}

func secureURL(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("invalid MOI endpoint")
	}
	local := u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost" || u.Hostname() == "::1"
	if u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !(u.Scheme == "https" || u.Scheme == "http" && local) {
		return nil, errors.New("MOI endpoint requires HTTPS; loopback HTTP is only for local development")
	}
	if u.Path != "" && path.Clean(u.Path) != strings.TrimSuffix(u.Path, "/") && u.Path != "/" {
		return nil, errors.New("non-canonical credential endpoint path")
	}
	return u, nil
}

// Transport freezes the account/environment for the entire command but asks
// the credential owner for a fresh access token before every HTTP request.
// It never replays a request after an authentication or transport failure.
type Transport struct {
	Helper  *Helper
	Binding Credential
	Base    http.RoundTripper
	// RequestTimeout starts after credential acquisition, including body reads.
	RequestTimeout time.Duration
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	bound, err := secureURL(t.Binding.Endpoint)
	if err != nil {
		return nil, err
	}
	prefix := strings.TrimSuffix(bound.Path, "/")
	if req.URL.Scheme != bound.Scheme || req.URL.Host != bound.Host || req.URL.User != nil ||
		path.Clean(req.URL.Path) != strings.TrimSuffix(req.URL.Path, "/") ||
		!(req.URL.Path == prefix || strings.HasPrefix(req.URL.Path, prefix+"/")) {
		return nil, errors.New("request endpoint differs from the signed-in MOI environment")
	}
	credential, err := t.Helper.Credential(req.Context(), t.Binding.Generation)
	if err != nil {
		return nil, err
	}
	if credential.Environment != t.Binding.Environment || credential.Issuer != t.Binding.Issuer || credential.Subject != t.Binding.Subject || credential.Endpoint != t.Binding.Endpoint {
		return nil, errors.New("MOI credential binding changed during command")
	}
	if req.Header.Get("Cookie") != "" || req.Header.Get("X-API-Key") != "" {
		return nil, errors.New("cannot mix MOI login with cookie or PAT authentication")
	}
	ctx := req.Context()
	cancel := func() {}
	if t.RequestTimeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, t.RequestTimeout)
	}
	clone := req.Clone(ctx)
	clone.Header.Set("Authorization", "Bearer "+credential.AccessToken)
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	response, err := base.RoundTrip(clone)
	if err != nil {
		cancel()
		return nil, err
	}
	response.Body = &cancelBody{ReadCloser: response.Body, cancel: cancel}
	return response, nil
}

type cancelBody struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (b *cancelBody) Close() error {
	defer b.cancel()
	return b.ReadCloser.Close()
}

func (c Credential) String() string {
	return fmt.Sprintf("MOI account %s (%s)", c.Subject, c.Environment)
}
