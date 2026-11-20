package update

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func installFixture(t *testing.T) (*fixture, Channel) {
	t.Helper()
	f := newFixture(t)
	c := f.publish("b1", "1.0.0", 1)
	if _, err := f.m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	return f, c
}

func TestReinstallDeletedPrefixWithOwnedSkill(t *testing.T) {
	f, c := installFixture(t)
	// Only this test's disposable installation is removed, never a user prefix.
	if err := os.RemoveAll(f.m.Root); err != nil {
		t.Fatal(err)
	}
	m, err := Init(f.m.Root, f.m.Config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Apply(t.Context(), false); err != nil {
		t.Fatal(err)
	}
	if err := m.verifyRelease(c.Hash); err != nil {
		t.Fatal(err)
	}
	s, err := m.state()
	if err != nil || m.validatePointers(s) != nil {
		t.Fatalf("invalid reinstalled state: %+v %v", s, err)
	}
}

func TestReinstallDoesNotClaimDamagedStateOrForeignSkills(t *testing.T) {
	for _, kind := range []string{"state-only", "state-and-current", "foreign-owner", "foreign-pointer", "user-directory"} {
		t.Run(kind, func(t *testing.T) {
			f, _ := installFixture(t)
			link := filepath.Join(f.m.Config.SkillRoot, "moi-cli")
			original, err := os.Readlink(link)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(kind, "state") {
				if err := os.Remove(f.m.path("state.json")); err != nil {
					t.Fatal(err)
				}
				if kind == "state-and-current" {
					if err := os.Remove(f.m.path("current")); err != nil {
						t.Fatal(err)
					}
				}
			} else {
				if err := os.RemoveAll(f.m.Root); err != nil {
					t.Fatal(err)
				}
				m, err := Init(f.m.Root, f.m.Config)
				if err != nil {
					t.Fatal(err)
				}
				f.m = m
				switch kind {
				case "foreign-owner":
					if err := atomicJSON(filepath.Join(m.skillStore(), "owner.json"), "/another/install"); err != nil {
						t.Fatal(err)
					}
				case "foreign-pointer":
					original = filepath.Join(t.TempDir(), "user-skill")
					if err := pointer(link, original); err != nil {
						t.Fatal(err)
					}
				case "user-directory":
					if err := os.Remove(link); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(link, 0700); err != nil {
						t.Fatal(err)
					}
				}
			}
			_, err = f.m.Apply(t.Context(), false)
			code := "SKILL_NOT_OWNED"
			if strings.HasPrefix(kind, "state") {
				code = "POINTER_STATE_MISMATCH"
			}
			requireError(t, err, code)
			if kind == "user-directory" {
				if err := directory(link); err != nil {
					t.Fatal("user directory changed", err)
				}
			} else if target, err := os.Readlink(link); err != nil || target != original {
				t.Fatal("unowned or live pointer changed", target, err)
			}
		})
	}
}

func TestPointerStateMismatchExplainsSafeReinstall(t *testing.T) {
	f, _ := installFixture(t)
	before, err := f.m.state()
	if err != nil {
		t.Fatal(err)
	}
	wrong := "releases/" + strings.Repeat("0", 64)
	if err := pointer(f.m.path("current"), wrong); err != nil {
		t.Fatal(err)
	}
	_, err = f.m.Apply(t.Context(), false)
	requireError(t, err, "POINTER_STATE_MISMATCH")
	message := PublicError(err)
	if !strings.HasPrefix(message, "CLI_UPDATE_POINTER_STATE_MISMATCH: ") ||
		!strings.Contains(message, "move the client installation directory aside") ||
		!strings.Contains(message, "sign-in data does not need to be removed") {
		t.Fatalf("missing recovery guidance: %s", message)
	}
	after, err := f.m.state()
	if err != nil || after != before {
		t.Fatalf("failed update changed installation state: %+v %v", after, err)
	}
	if target, err := os.Readlink(f.m.path("current")); err != nil || target != wrong {
		t.Fatalf("failed update changed the conflicting pointer: %q %v", target, err)
	}
}

func TestMissingSkillCopyRecovery(t *testing.T) {
	for _, mode := range []string{"current", "update", "rollback", "journal", "reuse"} {
		t.Run(mode, func(t *testing.T) {
			f, first := installFixture(t)
			target := first.Hash
			if mode == "rollback" || mode == "reuse" {
				second := f.publish("b2", "1.0.1", 2)
				if _, err := f.m.Apply(t.Context(), false); err != nil {
					t.Fatal(err)
				}
				if mode == "reuse" {
					if _, err := f.m.Apply(t.Context(), true); err != nil {
						t.Fatal(err)
					}
					target = second.Hash
				}
			}
			if mode == "update" {
				f.publish("b2", "1.0.1", 2)
			}
			copy := filepath.Join(f.m.skillStore(), target)
			if err := os.RemoveAll(copy); err != nil {
				t.Fatal(err)
			}
			requireError(t, f.m.verifyRelease(target), "SKILL_COPY_MISSING")
			if _, err := os.Lstat(copy); !isMissing(err) {
				t.Fatal("read-only verification mutated Skill copy")
			}
			if mode == "journal" {
				s, err := f.m.state()
				if err != nil {
					t.Fatal(err)
				}
				if err := atomicJSON(f.m.path("transaction.json"), transaction{Old: s, Target: target}); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := f.m.Apply(t.Context(), mode == "rollback"); err != nil {
				t.Fatal(err)
			}
			if err := f.m.verifyRelease(target); err != nil {
				t.Fatal(err)
			}
			if _, err := f.m.Apply(t.Context(), false); err != nil {
				t.Fatal("repeat", err)
			}
		})
	}
}

func TestSkillRepairRejectsModifiedSourceAndExistingTargets(t *testing.T) {
	for _, kind := range []string{"source", "partial", "symlink", "file"} {
		t.Run(kind, func(t *testing.T) {
			f, c := installFixture(t)
			copy := filepath.Join(f.m.skillStore(), c.Hash)
			if err := os.RemoveAll(copy); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "source":
				if err := os.WriteFile(f.m.path("releases/"+c.Hash+"/skills/moi-cli/SKILL.md"), []byte("modified"), 0600); err != nil {
					t.Fatal(err)
				}
			case "partial":
				if err := os.Mkdir(copy, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink":
				if err := os.Symlink(t.TempDir(), copy); err != nil {
					t.Fatal(err)
				}
			case "file":
				if err := os.WriteFile(copy, []byte("keep"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := f.m.Apply(t.Context(), false)
			code := "UNSAFE_DIRECTORY"
			if kind == "source" {
				code = "LOCAL_RELEASE_MODIFIED"
			}
			if kind == "partial" {
				code = "LOCAL_SKILL_MODIFIED"
			}
			requireError(t, err, code)
			s, err := f.m.state()
			if err != nil || s.Current != c.Hash {
				t.Fatal("changed current")
			}
			if kind == "source" {
				if _, err := os.Lstat(copy); !isMissing(err) {
					t.Fatal("repaired from damaged source")
				}
			}
		})
	}
}

func TestExplicitCheckWaitsForBackgroundLock(t *testing.T) {
	for _, apply := range []bool{false, true} {
		t.Run(map[bool]string{false: "check", true: "apply"}[apply], func(t *testing.T) {
			f, _ := installFixture(t)
			lock, err := lockFile(f.m.Root, "check.lock", true)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Close()
			done := make(chan error, 1)
			go func() {
				if apply {
					_, err := f.m.Apply(t.Context(), false)
					done <- err
				} else {
					_, err := f.m.Check(t.Context())
					done <- err
				}
			}()
			select {
			case err := <-done:
				t.Fatalf("did not wait: %v", err)
			case <-time.After(100 * time.Millisecond):
			}
			if err := lock.Close(); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("check did not resume")
			}
		})
	}
}

func TestCheckLockTimeoutCancellationAndBackground(t *testing.T) {
	f := newFixture(t)
	lock, err := lockFile(f.m.Root, "check.lock", true)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	started := time.Now()
	_, err = f.m.checkLock(t.Context(), false)
	requireError(t, err, "CHECK_BUSY")
	f.m.refresh()
	if time.Since(started) > time.Second {
		t.Fatal("background waited for lock")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	_, err = f.m.Check(ctx)
	requireError(t, err, "CHECK_CANCELLED")
	_, err = f.m.checkLock(t.Context(), true)
	requireError(t, err, "CHECK_BUSY")
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(f.m.path("check.lock")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(f.m.path("check.lock"), 0700); err != nil {
		t.Fatal(err)
	}
	_, err = f.m.checkLock(t.Context(), true)
	requireError(t, err, "UNSAFE_FILE")
}

func TestCompatibilityChangeIsVisibleWithoutActivation(t *testing.T) {
	f, first := installFixture(t)
	c := f.publish("b2", "1.0.1", 2)
	f.mu.Lock()
	var manifest Manifest
	if err := json.Unmarshal(f.objects["/bundles/b2/manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Compatibility = "contract-2"
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	f.objects["/bundles/b2/manifest.json"] = data
	f.mu.Unlock()
	c.Hash = digest(data)
	f.channel(c)
	r, err := f.m.Check(t.Context())
	if err != nil || r.Status != "reinstall_required" || r.Reason != "CLI_UPDATE_COMPATIBILITY_CHANGE" || !strings.Contains(r.Message, "--skill-dir") {
		t.Fatalf("check: %+v %v", r, err)
	}
	_, err = f.m.Apply(t.Context(), false)
	requireError(t, err, "COMPATIBILITY_CHANGE")
	if !strings.Contains(PublicError(err), "separate installation") {
		t.Fatal("missing guidance")
	}
	s, err := f.m.state()
	if err != nil || s.Current != first.Hash {
		t.Fatal("incompatible release activated")
	}
	var cache Cache
	if err := readJSON(f.m.path("cache.json"), &cache); err != nil {
		t.Fatal(err)
	}
	if cache.Available || !strings.Contains(cachedNotice(cache, s, time.Now()), "separate installation") {
		t.Fatal("wrong compatibility notice")
	}
	cache.Reason = "UNKNOWN"
	if cachedNotice(cache, s, time.Now()) != "" {
		t.Fatal("unknown reason shown")
	}
	cache.Available = true
	if !strings.Contains(cachedNotice(cache, s, time.Now()), "Run astra update.") {
		t.Fatal("missing update notice")
	}
	cache.Channel.Bundle = "bad\x1b[31m"
	if cachedNotice(cache, s, time.Now()) != "" {
		t.Fatal("terminal injection")
	}
	cache.Channel.Bundle = "b2"
	cache.Channel.Expires = time.Now().Add(-time.Second)
	if cachedNotice(cache, s, time.Now()) != "" {
		t.Fatal("expired notice")
	}
}

func TestUnsupportedPlatformKeepsInstallationAndProvidesGuidance(t *testing.T) {
	f, first := installFixture(t)
	c := f.publish("arm-only", "1.0.1", 2)
	other := "darwin-arm64"
	if platform() == other {
		other = "linux-arm64"
	}
	f.mu.Lock()
	var manifest Manifest
	decodeErr := json.Unmarshal(f.objects["/bundles/arm-only/manifest.json"], &manifest)
	f.mu.Unlock()
	if decodeErr != nil {
		t.Fatal(decodeErr)
	}
	items := manifest.Platforms[platform()]
	for name, item := range items.Components {
		item.Key = strings.ReplaceAll(item.Key, platform(), other)
		item.URL = strings.ReplaceAll(item.URL, platform(), other)
		items.Components[name] = item
	}
	manifest.Platforms = map[string]Platform{other: items}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	f.mu.Lock()
	f.objects["/bundles/arm-only/manifest.json"] = data
	f.mu.Unlock()
	c.Hash = digest(data)
	f.channel(c)

	for _, run := range []func(context.Context) (Result, error){f.m.Check, func(ctx context.Context) (Result, error) {
		return f.m.Apply(ctx, false)
	}} {
		_, err := run(t.Context())
		requireError(t, err, "PLATFORM_UNSUPPORTED")
		message := PublicError(err)
		if !strings.HasPrefix(message, "CLI_UPDATE_PLATFORM_UNSUPPORTED: ") ||
			!strings.Contains(message, "operating system and architecture") ||
			!strings.Contains(message, "Keep your current installation") ||
			!strings.Contains(message, "retry when a compatible release is available") {
			t.Fatalf("missing platform guidance: %s", message)
		}
	}
	s, err := f.m.state()
	if err != nil || s.Current != first.Hash {
		t.Fatalf("unsupported release changed installation: %+v %v", s, err)
	}
	if err := f.m.validatePointers(s); err != nil {
		t.Fatal(err)
	}
	if err := f.m.verifyRelease(first.Hash); err != nil {
		t.Fatal(err)
	}
}

func TestLocalErrorGuidanceDoesNotLeakRawPaths(t *testing.T) {
	for _, cause := range []error{os.ErrPermission, os.ErrNotExist, syscall.ENOSPC, errors.New("other")} {
		message := PublicError(&os.PathError{Op: "open", Path: "/private/secret", Err: cause})
		if strings.Contains(message, "/private") || !strings.HasPrefix(message, "CLI_UPDATE_LOCAL_") {
			t.Fatal(message)
		}
	}
	for _, reason := range []string{"BUSY", "CHECK_BUSY", "CHECK_CANCELLED", "SKILL_COPY_MISSING", "LOCAL_SKILL_MODIFIED", "COMPATIBILITY_CHANGE", "PLATFORM_UNSUPPORTED"} {
		if PublicError(invalid(reason)) == invalid(reason).Error() {
			t.Fatal("missing guidance", reason)
		}
	}
	wrapped := errors.Join(invalid("POINTER_STATE_MISMATCH"), &os.PathError{Op: "open", Path: "/private/secret", Err: os.ErrPermission})
	if message := PublicError(wrapped); !strings.HasPrefix(message, "CLI_UPDATE_POINTER_STATE_MISMATCH: ") || strings.Contains(message, "/private/secret") {
		t.Fatalf("wrapped updater error leaked a private path: %s", message)
	}
	if message := PublicError(invalid("FUTURE_REASON")); message == "CLI_UPDATE_FUTURE_REASON" || !strings.Contains(message, "Contact support") {
		t.Fatalf("unmapped updater error lacks public guidance: %s", message)
	}
}
