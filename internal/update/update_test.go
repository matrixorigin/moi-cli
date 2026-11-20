package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type fixture struct {
	t       *testing.T
	server  *httptest.Server
	mu      sync.Mutex
	objects map[string][]byte
	hits    map[string]int
	scripts map[string]string
	m       *Manager
}

func archiveBytes(t *testing.T, name, version string) []byte {
	t.Helper()
	script := fmt.Sprintf("#!/bin/sh\ncase \"$1\" in\n--version) echo '%s %s';;\n--moi-update-protocol) echo '%s';;\nskill) mkdir -p \"$4\"; echo 'skill %s' > \"$4/SKILL.md\";;\n*) exit 1;;\nesac\n", name, version, Protocol, version)
	if name == "moi-cli" {
		script = strings.Replace(script, "echo 'moi-cli ", "echo 'moi-cli version ", 1)
	}
	return archiveBody(t, name, script)
}

func archiveBody(t *testing.T, name, script string) []byte {
	t.Helper()
	entries := map[string]string{name: script}
	if name != "moi-cli" {
		entries["LICENSE"] = "license"
	}
	var out bytes.Buffer
	gz := gzip.NewWriter(&out)
	tw := tar.NewWriter(gz)
	// Fixed order for deterministic test component reuse.
	for _, n := range []string{name, "LICENSE"} {
		s, ok := entries[n]
		if !ok {
			continue
		}
		if err := tw.WriteHeader(&tar.Header{Name: n, Typeflag: tar.TypeReg, Mode: 0755, Size: int64(len(s))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(s)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{t: t, objects: map[string][]byte{}, hits: map[string]int{}}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.hits[r.URL.Path]++
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("distribution sent credentials")
		}
		b, ok := f.objects[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(b)
	}))
	t.Cleanup(f.server.Close)
	root := t.TempDir()
	m, err := Init(filepath.Join(root, "install"), Installation{Channel: f.server.URL + "/channels/stable.json", SkillRoot: filepath.Join(root, "skills"), Compatibility: "contract-1", AstraURL: f.server.URL, Loopback: true})
	if err != nil {
		t.Fatal(err)
	}
	f.m = m
	return f
}
func (f *fixture) publish(bundle, moi string, seq int64) Channel {
	f.t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	v := Manifest{Schema: "moi-client-bundle-v1", Bundle: bundle, Compatibility: "contract-1", Platforms: map[string]Platform{}}
	v.Skill.Owner = "moi-cli"
	v.Skill.Delivery = "embedded-export"
	v.ClientConfig.AstraURL = f.server.URL
	items := map[string]Component{}
	for _, name := range names {
		version := "1.0.0"
		if name == "moi-cli" {
			version = moi
		}
		if _, custom := f.scripts[name]; custom && name != "moi-cli" {
			version = "1.0.1"
		}
		b := archiveBytes(f.t, name, version)
		if body, ok := f.scripts[name]; ok {
			b = archiveBody(f.t, name, body)
		}
		key := "components/" + name + "/" + version + "/" + platform() + ".tar.gz"
		items[name] = Component{Version: version, Key: key, URL: f.server.URL + "/" + key, Hash: digest(b), Size: int64(len(b))}
		f.objects["/"+key] = b
	}
	v.Platforms[platform()] = Platform{Components: items}
	b, _ := json.Marshal(v)
	c := Channel{Schema: "moi-client-channel-v1", Bundle: bundle, URL: f.server.URL + "/bundles/" + bundle + "/manifest.json", Hash: digest(b), Sequence: seq, Issued: time.Now().UTC().Truncate(time.Second), Expires: time.Now().UTC().Truncate(time.Second).Add(24 * time.Hour)}
	f.objects["/bundles/"+bundle+"/manifest.json"] = b
	cb, _ := json.Marshal(c)
	f.objects["/channels/stable.json"] = cb
	return c
}
func (f *fixture) channel(c Channel) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, _ := json.Marshal(c)
	f.objects["/channels/stable.json"] = b
}
func requireError(t *testing.T, err error, part string) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), part) {
		t.Fatalf("expected %s, got %v", part, err)
	}
}

func TestInstallUpdateReuseRollbackAndCredentialPreservation(t *testing.T) {
	f := newFixture(t)
	first := f.publish("bundle-1", "1.0.0", 1)
	sentinel := filepath.Join(filepath.Dir(f.m.Root), "auth.json")
	if err := os.WriteFile(sentinel, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	r, err := f.m.Apply(t.Context(), false)
	if err != nil || r.Status != "updated" {
		t.Fatalf("install: %+v %v", r, err)
	}
	if err := f.m.verifyRelease(first.Hash); err != nil {
		t.Fatal(err)
	}
	second := f.publish("bundle-2", "1.0.1", 2)
	r, err = f.m.Apply(t.Context(), false)
	if err != nil || strings.Join(r.Changed, ",") != "moi-cli" {
		t.Fatalf("update: %+v %v", r, err)
	}
	for _, n := range []string{"astra", "astra-edge"} {
		key := "/components/" + n + "/1.0.0/" + platform() + ".tar.gz"
		if f.hits[key] != 1 {
			t.Fatalf("%s downloaded %d times", n, f.hits[key])
		}
	}
	s, _ := f.m.state()
	if s.Current != second.Hash || s.Previous != first.Hash {
		t.Fatal(s)
	}
	r, err = f.m.Apply(t.Context(), true)
	if err != nil || r.Status != "rolled_back" {
		t.Fatalf("rollback: %+v %v", r, err)
	}
	s, _ = f.m.state()
	if s.Current != first.Hash {
		t.Fatal(s)
	}
	b, _ := os.ReadFile(sentinel)
	if string(b) != "untouched" {
		t.Fatal("credentials changed")
	}
	if r, err = f.m.Apply(t.Context(), false); err != nil || r.Status != "updated" {
		t.Fatalf("reapply: %+v %v", r, err)
	}
	if r, err = f.m.Apply(t.Context(), false); err != nil || r.Status != "current" {
		t.Fatalf("repeat: %+v %v", r, err)
	}
}

func TestVersionProbeFailureKeepsCurrentRelease(t *testing.T) {
	f := newFixture(t)
	first := f.publish("bundle-1", "1.0.0", 1)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	f.scripts = map[string]string{
		"astra":      "#!/bin/sh\necho 'dyld: Library not loaded: /opt/homebrew/opt/openssl@3/lib/libssl.3.dylib' >&2\nexit 1\n",
		"astra-edge": fmt.Sprintf("#!/bin/sh\ncase \"$1\" in\n--version) echo 'astra-edge 1.0.1';;\n--moi-update-protocol) echo '%s';;\n*) exit 1;;\nesac\n", Protocol),
	}
	second := f.publish("bundle-2", "1.0.1", 2)
	_, err := f.m.Apply(t.Context(), false)
	message := PublicError(err)
	if !strings.Contains(message, "CLI_UPDATE_VERSION_PROBE") || !strings.Contains(message, "astra") {
		t.Fatalf("public error: %s (%v)", message, err)
	}
	if strings.Contains(message, "/opt/homebrew") || strings.Contains(message, "libssl") {
		t.Fatalf("loader path leaked: %s", message)
	}
	logText, readErr := os.ReadFile(f.m.path("probe.log"))
	if readErr != nil || !strings.Contains(string(logText), "component=astra") || !strings.Contains(string(logText), "libssl.3.dylib") {
		t.Fatalf("probe log: %s (%v)", logText, readErr)
	}
	s, stateErr := f.m.state()
	if stateErr != nil || s.Current != first.Hash {
		t.Fatalf("current release changed: %+v %v", s, stateErr)
	}
	if _, statErr := os.Lstat(f.m.path("releases/" + second.Hash)); !os.IsNotExist(statErr) {
		t.Fatalf("failed probe left a release: %v", statErr)
	}
	if _, statErr := os.Lstat(f.m.path("transaction.json")); !os.IsNotExist(statErr) {
		t.Fatalf("failed probe left a transaction: %v", statErr)
	}
}

func TestVersionProbeLogRecordsExecErrorWithoutStderr(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "astra"), []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := &Manager{Root: t.TempDir(), Config: Installation{Platform: platform()}}
	v := Manifest{Platforms: map[string]Platform{platform(): {Components: map[string]Component{
		"astra": {Version: "1.0.0"},
	}}}}
	err := probe(t.Context(), m, dir, v)
	message := PublicError(err)
	if !strings.Contains(message, "CLI_UPDATE_VERSION_PROBE") || !strings.Contains(message, "astra") {
		t.Fatalf("public error: %s (%v)", message, err)
	}
	if strings.Contains(message, "exit status") || strings.Contains(message, dir) {
		t.Fatalf("exec diagnostic leaked: %s", message)
	}
	logText, readErr := os.ReadFile(m.path("probe.log"))
	if readErr != nil || !strings.Contains(string(logText), "component=astra") || !strings.Contains(string(logText), "error=exit status 1") {
		t.Fatalf("probe log: %s (%v)", logText, readErr)
	}
	if strings.Contains(string(logText), "stderr:") {
		t.Fatalf("empty stderr was recorded: %s", logText)
	}
}

func TestVersionProbeLogRecordsTimeoutWithoutStderr(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "astra"), []byte("#!/bin/sh\nsleep 30\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := &Manager{Root: t.TempDir(), Config: Installation{Platform: platform()}}
	v := Manifest{Platforms: map[string]Platform{platform(): {Components: map[string]Component{
		"astra": {Version: "1.0.0"},
	}}}}
	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()
	err := probe(ctx, m, dir, v)
	message := PublicError(err)
	if !strings.Contains(message, "CLI_UPDATE_VERSION_PROBE") || strings.Contains(message, "timed_out") || strings.Contains(message, dir) {
		t.Fatalf("public error: %s (%v)", message, err)
	}
	logText, readErr := os.ReadFile(m.path("probe.log"))
	if readErr != nil || !strings.Contains(string(logText), "timed_out=true") || !strings.Contains(string(logText), "error=") {
		t.Fatalf("probe log: %s (%v)", logText, readErr)
	}
	if strings.Contains(string(logText), "stderr:") {
		t.Fatalf("empty stderr was recorded: %s", logText)
	}
}

func TestChannelRejection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Channel)
		code   string
	}{
		{"expired", func(c *Channel) { c.Expires = time.Now().Add(-time.Minute) }, "EXPIRED"},
		{"future", func(c *Channel) { c.Issued = time.Now().Add(time.Hour) }, "EXPIRED"},
		{"schema", func(c *Channel) { c.Schema = "other" }, "CONTRACT"},
		{"hash", func(c *Channel) { c.Hash = strings.Repeat("b", 64) }, "CHECKSUM"},
		{"url", func(c *Channel) { c.URL = "http://example.com/bundles/" + c.Bundle + "/manifest.json" }, "HTTPS"},
		{"path", func(c *Channel) { c.Bundle = "../../escape" }, "CONTRACT"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			c := f.publish("b1", "1.0.0", 1)
			tc.mutate(&c)
			f.channel(c)
			_, err := f.m.Check(t.Context())
			requireError(t, err, tc.code)
		})
	}
}
func TestChannelReplayAndEquivocation(t *testing.T) {
	f := newFixture(t)
	c := f.publish("b1", "1.0.0", 2)
	if _, err := f.m.Check(t.Context()); err != nil {
		t.Fatal(err)
	}
	c.Sequence = 1
	f.channel(c)
	_, err := f.m.Check(t.Context())
	requireError(t, err, "REPLAY")
	c.Sequence = 2
	c.Expires = c.Expires.Add(time.Minute)
	f.channel(c)
	_, err = f.m.Check(t.Context())
	requireError(t, err, "REPLAY")
}

func TestChannelEquivalentTimezoneRoundTrip(t *testing.T) {
	f := newFixture(t)
	c := f.publish("b1", "1.0.0", 1)
	zone := time.FixedZone("publisher", 0)
	c.Issued = c.Issued.In(zone)
	c.Expires = c.Expires.In(zone)
	f.channel(c)
	for range 2 {
		if _, err := f.m.Check(t.Context()); err != nil {
			t.Fatal(err)
		}
	}
}
func TestBusyAndOldExecutableAdmission(t *testing.T) {
	f := newFixture(t)
	c := f.publish("b1", "1.0.0", 1)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	exe := f.m.path("releases/" + c.Hash + "/astra")
	lease, err := Acquire(exe)
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.m.Apply(t.Context(), false)
	requireError(t, err, "BUSY")
	lease.Close()
	f.publish("b2", "1.0.1", 2)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	_, err = Acquire(exe)
	requireError(t, err, "RESTART_REQUIRED")
	lock, err := lockFile(f.m.Root, "runtime.lock", true)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	_, err = f.m.Apply(t.Context(), false)
	requireError(t, err, "BUSY")
}
func TestInterruptedActivationRecovery(t *testing.T) {
	for _, phase := range []string{"journal", "skill", "current", "state"} {
		t.Run(phase, func(t *testing.T) {
			f := newFixture(t)
			c1 := f.publish("b1", "1.0.0", 1)
			if _, err := f.m.Apply(t.Context(), false); err != nil {
				t.Fatal(err)
			}
			c2 := f.publish("b2", "1.0.1", 2)
			if _, err := f.m.Apply(t.Context(), false); err != nil {
				t.Fatal(err)
			}
			if _, err := f.m.Apply(t.Context(), true); err != nil {
				t.Fatal(err)
			}
			old, _ := f.m.state()
			if err := atomicJSON(f.m.path("transaction.json"), transaction{Old: old, Target: c2.Hash}); err != nil {
				t.Fatal(err)
			}
			if phase != "journal" {
				if err := pointer(filepath.Join(f.m.Config.SkillRoot, "moi-cli"), filepath.Join(f.m.skillStore(), c2.Hash)); err != nil {
					t.Fatal(err)
				}
			}
			if phase == "current" || phase == "state" {
				if err := pointer(f.m.path("current"), "releases/"+c2.Hash); err != nil {
					t.Fatal(err)
				}
			}
			if phase == "state" {
				if err := atomicJSON(f.m.path("state.json"), State{Current: c2.Hash, Previous: c1.Hash}); err != nil {
					t.Fatal(err)
				}
			}
			_, err := Acquire(f.m.path("releases/" + c1.Hash + "/astra"))
			requireError(t, err, "RECOVERY_REQUIRED")
			if err := f.m.recover(); err != nil {
				t.Fatal(err)
			}
			if err := f.m.recover(); err != nil {
				t.Fatal(err)
			}
			s, _ := f.m.state()
			if s != old {
				t.Fatal(s)
			}
			p, _ := os.Readlink(f.m.path("current"))
			if p != "releases/"+c1.Hash {
				t.Fatal(p)
			}
			lease, err := Acquire(f.m.path("current/astra"))
			if err != nil {
				t.Fatal(err)
			}
			lease.Close()
		})
	}
}
func TestCorruptionAndDowngradeLeaveCurrentUnchanged(t *testing.T) {
	for _, kind := range []string{"archive", "release", "skill", "downgrade"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			c := f.publish("b1", "1.0.1", 1)
			if _, err := f.m.Apply(t.Context(), false); err != nil {
				t.Fatal(err)
			}
			version := "1.0.2"
			if kind == "downgrade" {
				version = "1.0.0"
			}
			f.publish("b2", version, 2)
			switch kind {
			case "archive":
				f.objects["/components/moi-cli/"+version+"/"+platform()+".tar.gz"] = []byte("corrupt")
			case "release":
				if err := os.WriteFile(f.m.path("releases/"+c.Hash+"/astra"), []byte("corrupt"), 0755); err != nil {
					t.Fatal(err)
				}
			case "skill":
				if err := os.WriteFile(filepath.Join(f.m.skillStore(), c.Hash, "SKILL.md"), []byte("corrupt"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := f.m.Apply(t.Context(), false)
			if err == nil {
				t.Fatal("accepted corruption/downgrade")
			}
			s, _ := f.m.state()
			if s.Current != c.Hash {
				t.Fatal("changed current")
			}
		})
	}
}
func TestUnsafeArchives(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind byte
		mode int64
	}{
		{"../astra", tar.TypeReg, 0755}, {"/astra", tar.TypeReg, 0755}, {"astra", tar.TypeSymlink, 0755}, {"astra", tar.TypeLink, 0755}, {"astra", tar.TypeDir, 0755}, {"astra", tar.TypeReg, 04755},
	} {
		t.Run(fmt.Sprintf("%s-%d-%d", tc.name, tc.kind, tc.mode), func(t *testing.T) {
			var out bytes.Buffer
			gz := gzip.NewWriter(&out)
			tw := tar.NewWriter(gz)
			if err := tw.WriteHeader(&tar.Header{Name: tc.name, Typeflag: tc.kind, Mode: tc.mode}); err != nil {
				t.Fatal(err)
			}
			tw.Close()
			gz.Close()
			dir := t.TempDir()
			file := filepath.Join(dir, "a.tgz")
			os.WriteFile(file, out.Bytes(), 0600)
			requireError(t, extract(file, dir, "astra"), "ARCHIVE_MEMBER")
		})
	}
}
func TestTransportAndURLPolicy(t *testing.T) {
	for _, raw := range []string{"http://localhost:80/x", "https://u:p@host/x", "https://host/x?token=secret", "https://host/x#fragment", "https://host/a%2fb"} {
		if trustedURL(raw, "", true) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if trustedURL("https://other/x", "https://host/channel", false) == nil {
		t.Fatal("cross origin")
	}
	f := newFixture(t)
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, f.server.URL+"/x", 302) }))
	defer redirect.Close()
	f.m.Config.Channel = redirect.URL + "/channels/stable.json"
	var out bytes.Buffer
	requireError(t, f.m.download(t.Context(), f.m.Config.Channel, 100, &out), "STATUS")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	requireError(t, f.m.download(ctx, f.m.Config.Channel, 100, &out), "UNAVAILABLE")
}
func TestUnownedSkillAndStandaloneProtection(t *testing.T) {
	f := newFixture(t)
	f.publish("b1", "1.0.0", 1)
	path := filepath.Join(f.m.Config.SkillRoot, "moi-cli")
	if err := os.MkdirAll(path, 0700); err != nil {
		t.Fatal(err)
	}
	_, err := f.m.Apply(t.Context(), false)
	requireError(t, err, "SKILL_NOT_OWNED")
	if s, err := os.Stat(path); err != nil || !s.IsDir() {
		t.Fatal("user Skill changed")
	}
	exe := filepath.Join(t.TempDir(), "astra")
	os.WriteFile(exe, []byte("standalone"), 0700)
	lease, err := Acquire(exe)
	if err != nil || lease != nil {
		t.Fatalf("standalone: %v %v", lease, err)
	}
	_, err = Init(f.m.Root, Installation{Channel: "https://other/channel", SkillRoot: f.m.Config.SkillRoot, Compatibility: f.m.Config.Compatibility, AstraURL: f.m.Config.AstraURL, Loopback: f.m.Config.Loopback})
	requireError(t, err, "CONFLICT")
}
func TestMachineCallsAndCheckThrottle(t *testing.T) {
	for _, args := range [][]string{{"-o", "workspace", "list"}, {"auth", "credential"}, {"--version"}, {"skill", "export"}, {"--no-interactive"}} {
		if !MachineCall(args) {
			t.Fatal(args)
		}
	}
	t.Setenv("MOI_AGENT_CALL", "1")
	if !MachineCall([]string{"workspace", "list"}) {
		t.Fatal("agent marker")
	}
	f := newFixture(t)
	f.publish("b1", "1.0.0", 1)
	f.m.refresh()
	f.m.refresh()
	if f.hits["/channels/stable.json"] != 1 {
		t.Fatal("not throttled")
	}
}

func TestStartupRefreshUsesHourlyCache(t *testing.T) {
	for _, tt := range []struct {
		name string
		age  time.Duration
		want int
	}{
		{"within_hour", 59 * time.Minute, 1},
		{"one_hour", time.Hour, 2},
		{"older_than_hour", 2 * time.Hour, 2},
		{"future_clock", -time.Hour, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			f.publish("b1", "1.0.0", 1)
			f.m.refresh()
			var cache Cache
			if err := readJSON(f.m.path("cache.json"), &cache); err != nil {
				t.Fatal(err)
			}
			cache.Checked = time.Now().Add(-tt.age)
			if err := atomicJSON(f.m.path("cache.json"), cache); err != nil {
				t.Fatal(err)
			}
			f.m.refresh()
			f.mu.Lock()
			defer f.mu.Unlock()
			if got := f.hits["/channels/stable.json"]; got != tt.want {
				t.Fatalf("channel requests = %d, want %d", got, tt.want)
			}
			if len(f.hits) != 2 {
				t.Fatalf("refresh must only request channel and manifest: %v", f.hits)
			}
		})
	}
}

func TestManifestValidationMatrix(t *testing.T) {
	for _, kind := range []string{"schema", "platform", "component", "version", "digest", "size", "origin", "key", "skill"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			c := f.publish("b1", "1.0.0", 1)
			var v Manifest
			json.Unmarshal(f.objects["/bundles/b1/manifest.json"], &v)
			item := v.Platforms[platform()].Components["moi-cli"]
			switch kind {
			case "schema":
				v.Schema = "unknown"
			case "platform":
				delete(v.Platforms, platform())
			case "component":
				delete(v.Platforms[platform()].Components, "astra-edge")
			case "version":
				item.Version = "../bad"
			case "digest":
				item.Hash = "bad"
			case "size":
				item.Size = maxArchive + 1
			case "origin":
				item.URL = "https://other.invalid/" + item.Key
			case "key":
				item.Key = "../../escape"
			case "skill":
				v.Skill.Owner = "unknown"
			}
			if kind != "platform" {
				v.Platforms[platform()].Components["moi-cli"] = item
			}
			b, _ := json.Marshal(v)
			c.Hash = digest(b)
			f.objects["/bundles/b1/manifest.json"] = b
			f.channel(c)
			if _, err := f.m.Check(t.Context()); err == nil {
				t.Fatal("accepted invalid manifest")
			}
		})
	}
}
func TestVersionConflictAndCompatibilityChange(t *testing.T) {
	for _, kind := range []string{"version", "compatibility", "astra-url"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			c1 := f.publish("b1", "1.0.0", 1)
			if _, err := f.m.Apply(t.Context(), false); err != nil {
				t.Fatal(err)
			}
			c2 := f.publish("b2", "1.0.0", 2)
			var v Manifest
			json.Unmarshal(f.objects["/bundles/b2/manifest.json"], &v)
			if kind == "version" {
				item := v.Platforms[platform()].Components["moi-cli"]
				item.Hash = strings.Repeat("a", 64)
				v.Platforms[platform()].Components["moi-cli"] = item
			} else if kind == "compatibility" {
				v.Compatibility = "another-contract"
			} else {
				v.ClientConfig.AstraURL = "https://astra-new.example.test"
			}
			b, _ := json.Marshal(v)
			f.objects["/bundles/b2/manifest.json"] = b
			c2.Hash = digest(b)
			f.channel(c2)
			if _, err := f.m.Apply(t.Context(), false); err == nil {
				t.Fatal("accepted incompatible release")
			}
			var cache Cache
			if err := readJSON(f.m.path("cache.json"), &cache); err != nil {
				t.Fatal(err)
			}
			if cache.Available {
				t.Fatal("incompatible release advertised as installable")
			}
			s, _ := f.m.state()
			if s.Current != c1.Hash {
				t.Fatal("changed current")
			}
		})
	}
}

func TestFreshInstallerRejectsDifferentStableContract(t *testing.T) {
	for _, kind := range []string{"compatibility", "astra-url"} {
		t.Run(kind, func(t *testing.T) {
			f := newFixture(t)
			c := f.publish("b1", "1.0.0", 1)
			var v Manifest
			if err := json.Unmarshal(f.objects["/bundles/b1/manifest.json"], &v); err != nil {
				t.Fatal(err)
			}
			if kind == "compatibility" {
				v.Compatibility = "contract-2"
			} else {
				v.ClientConfig.AstraURL = "https://astra-new.example.test"
			}
			data, err := json.Marshal(v)
			if err != nil {
				t.Fatal(err)
			}
			f.objects["/bundles/b1/manifest.json"] = data
			c.Hash = digest(data)
			f.channel(c)
			_, err = f.m.Apply(t.Context(), false)
			requireError(t, err, "COMPATIBILITY_CHANGE")
			s, stateErr := f.m.state()
			if stateErr != nil || s.Current != "" {
				t.Fatalf("incompatible fresh install activated: %+v %v", s, stateErr)
			}
		})
	}
}

func TestLegacyInstallationDerivesContractFromVerifiedCurrentRelease(t *testing.T) {
	f := newFixture(t)
	c := f.publish("b1", "1.0.0", 1)
	f.channel(c)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	path := f.m.path("installation.json")
	var legacy map[string]any
	if err := readJSON(path, &legacy); err != nil {
		t.Fatal(err)
	}
	delete(legacy, "compatibility_reference")
	delete(legacy, "astra_url")
	if err := atomicJSON(path, legacy); err != nil {
		t.Fatal(err)
	}
	m, err := Open(f.m.Root)
	if err != nil {
		t.Fatal(err)
	}
	if m.Config.Compatibility != "contract-1" || m.Config.AstraURL != f.server.URL {
		t.Fatalf("legacy contract not derived from current release: %+v", m.Config)
	}

	delete(legacy, "astra_url")
	legacy["compatibility_reference"] = "contract-1"
	if err := atomicJSON(path, legacy); err != nil {
		t.Fatal(err)
	}
	_, err = Open(f.m.Root)
	requireError(t, err, "INSTALLATION_CONTRACT")
}
func TestArchiveDuplicateMissingTruncated(t *testing.T) {
	for _, kind := range []string{"duplicate", "missing", "truncated"} {
		t.Run(kind, func(t *testing.T) {
			var out bytes.Buffer
			gz := gzip.NewWriter(&out)
			tw := tar.NewWriter(gz)
			count := 1
			if kind == "duplicate" {
				count = 2
			}
			for range count {
				tw.WriteHeader(&tar.Header{Name: "astra", Mode: 0755, Typeflag: tar.TypeReg, Size: 1})
				tw.Write([]byte("x"))
			}
			tw.Close()
			gz.Close()
			b := out.Bytes()
			if kind == "truncated" {
				b = b[:len(b)/2]
			}
			dir := t.TempDir()
			file := filepath.Join(dir, "a.tgz")
			os.WriteFile(file, b, 0600)
			if err := extract(file, dir, "astra"); err == nil {
				t.Fatal("accepted malformed archive")
			}
		})
	}
}
func TestRollbackCorruptionAndUnownedPointer(t *testing.T) {
	f := newFixture(t)
	c := f.publish("b1", "1.0.0", 1)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	f.publish("b2", "1.0.1", 2)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	before, _ := f.m.state()
	os.WriteFile(f.m.path("releases/"+c.Hash+"/moi-cli"), []byte("corrupt"), 0755)
	_, err := f.m.Apply(t.Context(), true)
	requireError(t, err, "LOCAL_RELEASE_MODIFIED")
	after, _ := f.m.state()
	if after != before {
		t.Fatal("rollback changed state")
	}
	if err := pointer(f.m.path("current"), "releases/"+c.Hash); err != nil {
		t.Fatal(err)
	}
	_, err = f.m.Apply(t.Context(), false)
	requireError(t, err, "POINTER_STATE_MISMATCH")
}
func TestFirstInstallRecoveryAndSymlinkRefusal(t *testing.T) {
	f := newFixture(t)
	c := f.publish("b1", "1.0.0", 1)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	if err := atomicJSON(f.m.path("transaction.json"), transaction{Old: State{}, Target: c.Hash}); err != nil {
		t.Fatal(err)
	}
	if err := f.m.recover(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(f.m.path("current")); !isMissing(err) {
		t.Fatal("first install pointer remains")
	}
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(f.m.path("cache.json")); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(t.TempDir(), "sentinel")
	os.WriteFile(sentinel, []byte("unchanged"), 0600)
	os.Symlink(sentinel, f.m.path("cache.json"))
	if _, err := f.m.Check(t.Context()); err == nil {
		t.Fatal("accepted cache symlink")
	}
	b, _ := os.ReadFile(sentinel)
	if string(b) != "unchanged" {
		t.Fatal("overwrote external file")
	}
}

func TestDevNullIsNotTerminal(t *testing.T) {
	f, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if terminal(f.Fd()) {
		t.Fatal("character device is not a terminal")
	}
}
