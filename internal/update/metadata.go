// Package update owns MOI-managed client installations. It never reads login data.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"
)

const Protocol = "moi-client-update-v1"
const maxArchive int64 = 512 << 20

var names = []string{"astra", "astra-edge", "moi-cli"}
var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var versionPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
var idPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,79}$`)
var compatibilityPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/@+-]{0,159}$`)

type Installation struct {
	Schema        string `json:"schema"`
	Channel       string `json:"channel"`
	Platform      string `json:"platform"`
	SkillRoot     string `json:"skill_root"`
	Compatibility string `json:"compatibility_reference"`
	AstraURL      string `json:"astra_url"`
	Loopback      bool   `json:"allow_loopback_http"`
}
type Channel struct {
	Schema   string    `json:"schema"`
	Bundle   string    `json:"bundle_version"`
	URL      string    `json:"manifest_url"`
	Hash     string    `json:"manifest_sha256"`
	Sequence int64     `json:"sequence"`
	Issued   time.Time `json:"issued_at"`
	Expires  time.Time `json:"expires_at"`
}
type Component struct {
	Version string `json:"version"`
	Key     string `json:"key"`
	URL     string `json:"url"`
	Hash    string `json:"sha256"`
	Size    int64  `json:"size"`
}
type Platform struct {
	Components map[string]Component `json:"components"`
}
type Manifest struct {
	Schema        string              `json:"schema"`
	Bundle        string              `json:"bundle_version"`
	Compatibility string              `json:"compatibility_reference"`
	Platforms     map[string]Platform `json:"platforms"`
	Skill         struct {
		Owner    string `json:"owner"`
		Delivery string `json:"delivery"`
	} `json:"skill"`
	ClientConfig struct {
		AstraURL string `json:"astra_url"`
	} `json:"client_config"`
}
type Cache struct {
	Checked   time.Time `json:"checked_at"`
	Channel   Channel   `json:"channel"`
	Available bool      `json:"available"`
	Reason    string    `json:"reason,omitempty"`
}
type State struct {
	Current  string `json:"current"`
	Previous string `json:"previous"`
}
type Result struct {
	Status  string   `json:"status"`
	Bundle  string   `json:"bundle_version,omitempty"`
	Changed []string `json:"changed,omitempty"`
	Reason  string   `json:"reason,omitempty"`
	Message string   `json:"message,omitempty"`
}

func digest(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

type updateError string

func (e updateError) Error() string { return "CLI_UPDATE_" + string(e) }
func invalid(reason string) error   { return updateError(reason) }
func platform() string              { return runtime.GOOS + "-" + runtime.GOARCH }

// Explicit loopback is only for local acceptance. URLs cannot carry credentials,
// query tokens, fragments, encoded path separators, or cross-origin references.
func trustedURL(raw string, base string, loopback bool) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" {
		return invalid("URL_REJECTED")
	}
	if u.Scheme != "https" && !(loopback && u.Scheme == "http" && u.Hostname() == "127.0.0.1" && u.Port() != "") {
		return invalid("HTTPS_REQUIRED")
	}
	if base != "" {
		b, e := url.Parse(base)
		if e != nil || b.Scheme != u.Scheme || b.Host != u.Host {
			return invalid("CROSS_ORIGIN_REJECTED")
		}
	}
	return nil
}

func trustedAstraURL(raw string, loopback bool) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || (u.Path != "" && u.Path != "/") {
		return invalid("ASTRA_URL_REJECTED")
	}
	if u.Scheme != "https" && !(loopback && u.Scheme == "http" && u.Hostname() == "127.0.0.1" && u.Port() != "") {
		return invalid("ASTRA_HTTPS_REQUIRED")
	}
	return nil
}

func (m *Manager) download(ctx context.Context, raw string, size int64, out io.Writer) error {
	if err := trustedURL(raw, m.Config.Channel, m.Config.Loopback); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return invalid("URL_REJECTED")
	}
	// No default transport proxy, cookies, authorization, or redirects. This is
	// anonymous software distribution, not an authenticated Product API request.
	client := &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Minute, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		return invalid("DOWNLOAD_UNAVAILABLE")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return invalid("DOWNLOAD_STATUS")
	}
	n, err := io.Copy(out, io.LimitReader(resp.Body, size+1))
	if err != nil || n > size {
		return invalid("DOWNLOAD_SIZE")
	}
	return nil
}

func (m *Manager) manifest(ctx context.Context, c Channel) (Manifest, []byte, error) {
	var buf strings.Builder
	var manifest Manifest
	if err := m.download(ctx, c.URL, 2<<20, &buf); err != nil {
		return manifest, nil, err
	}
	data := []byte(buf.String())
	if digest(data) != c.Hash {
		return manifest, nil, invalid("MANIFEST_CHECKSUM")
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, nil, invalid("MANIFEST_JSON")
	}
	if err := m.validateManifest(manifest, c.Bundle); err != nil {
		return manifest, nil, err
	}
	return manifest, data, nil
}

func (m *Manager) validateManifest(v Manifest, bundle string) error {
	if v.Schema != "moi-client-bundle-v1" || !idPattern.MatchString(v.Bundle) || v.Bundle != bundle || !compatibilityPattern.MatchString(v.Compatibility) || v.Skill.Owner != "moi-cli" || v.Skill.Delivery != "embedded-export" {
		return invalid("MANIFEST_CONTRACT")
	}
	items, ok := v.Platforms[m.Config.Platform]
	if !ok || len(items.Components) != 3 {
		return invalid("PLATFORM_UNSUPPORTED")
	}
	if items.Components["astra"].Version != items.Components["astra-edge"].Version {
		return invalid("COMPONENT_CONTRACT")
	}
	if err := trustedAstraURL(v.ClientConfig.AstraURL, m.Config.Loopback); err != nil {
		return err
	}
	for _, name := range names {
		c, ok := items.Components[name]
		key := "components/" + name + "/" + c.Version + "/" + m.Config.Platform + ".tar.gz"
		if !ok || !versionPattern.MatchString(c.Version) || !hashPattern.MatchString(c.Hash) || c.Size <= 0 || c.Size > maxArchive || c.Key != key {
			return invalid("COMPONENT_CONTRACT")
		}
		if err := trustedURL(c.URL, m.Config.Channel, m.Config.Loopback); err != nil {
			return err
		}
		u, _ := url.Parse(c.URL)
		if u.Path != "/"+key {
			return invalid("COMPONENT_URL")
		}
	}
	return nil
}

func (m *Manager) check(ctx context.Context, wait bool) (Channel, Manifest, []byte, error) {
	var c Channel
	var v Manifest
	lock, err := m.checkLock(ctx, wait)
	if err != nil {
		return c, v, nil, err
	}
	defer lock.Close()
	var body strings.Builder
	if err = m.download(ctx, m.Config.Channel, 64<<10, &body); err != nil {
		return c, v, nil, err
	}
	if json.Unmarshal([]byte(body.String()), &c) != nil || c.Schema != "moi-client-channel-v1" || !idPattern.MatchString(c.Bundle) || !hashPattern.MatchString(c.Hash) || c.Sequence < 1 {
		return c, v, nil, invalid("CHANNEL_CONTRACT")
	}
	now := time.Now()
	if !c.Expires.After(now) || c.Issued.After(now.Add(5*time.Minute)) || !c.Expires.After(c.Issued) || c.Expires.Sub(c.Issued) > 31*24*time.Hour {
		return c, v, nil, invalid("CHANNEL_EXPIRED")
	}
	u, e := url.Parse(c.URL)
	if e != nil || u.Path != "/bundles/"+c.Bundle+"/manifest.json" {
		return c, v, nil, invalid("MANIFEST_URL")
	}
	var old Cache
	if err := readJSON(m.path("cache.json"), &old); err == nil {
		if c.Sequence < old.Channel.Sequence || (c.Sequence == old.Channel.Sequence && !sameChannel(c, old.Channel)) {
			return c, v, nil, invalid("CHANNEL_REPLAY")
		}
	} else if !isMissing(err) {
		return c, v, nil, err
	}
	v, data, err := m.manifest(ctx, c)
	if err == nil {
		state, stateErr := m.state()
		var eligibilityErr error
		if stateErr == nil {
			_, eligibilityErr = m.changes(v, state)
		} else {
			eligibilityErr = stateErr
		}
		cache := Cache{Checked: now, Channel: c, Available: eligibilityErr == nil && state.Current != c.Hash}
		if errors.Is(eligibilityErr, invalid("COMPATIBILITY_CHANGE")) {
			cache.Reason = eligibilityErr.Error()
		}
		err = atomicJSON(m.path("cache.json"), cache)
		if err == nil {
			err = eligibilityErr
		}
	}
	return c, v, data, err
}

func sameChannel(a, b Channel) bool {
	return a.Schema == b.Schema && a.Bundle == b.Bundle && a.URL == b.URL && a.Hash == b.Hash && a.Sequence == b.Sequence && a.Issued.Equal(b.Issued) && a.Expires.Equal(b.Expires)
}
