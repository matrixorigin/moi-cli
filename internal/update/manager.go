package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Manager struct {
	Root     string
	Config   Installation
	progress *progressDisplay
}

func (m *Manager) path(p string) string { return filepath.Join(m.Root, p) }
func Open(root string) (*Manager, error) {
	if err := directory(root); err != nil {
		return nil, err
	}
	m := &Manager{Root: root}
	if err := readJSON(m.path("installation.json"), &m.Config); err != nil {
		return nil, err
	}
	if m.Config.Schema != Protocol || m.Config.Platform != platform() || !filepath.IsAbs(m.Config.SkillRoot) {
		return nil, invalid("INSTALLATION_CONTRACT")
	}
	if err := trustedURL(m.Config.Channel, "", m.Config.Loopback); err != nil {
		return nil, err
	}
	// Installations created before these fields became part of the contract are
	// bound to their already verified current release. Never infer from the
	// mutable channel, and never accept a partially missing contract.
	if m.Config.Compatibility == "" && m.Config.AstraURL == "" {
		s, err := m.state()
		if err != nil || s.Current == "" {
			return nil, invalid("INSTALLATION_CONTRACT")
		}
		current, err := m.localManifest(s.Current)
		if err != nil {
			return nil, err
		}
		m.Config.Compatibility = current.Compatibility
		m.Config.AstraURL = current.ClientConfig.AstraURL
	} else if !compatibilityPattern.MatchString(m.Config.Compatibility) || m.Config.AstraURL == "" {
		return nil, invalid("INSTALLATION_CONTRACT")
	}
	if err := trustedAstraURL(m.Config.AstraURL, m.Config.Loopback); err != nil {
		return nil, err
	}
	return m, nil
}

// Init only claims an empty dedicated prefix. Legacy preview installs must be
// left intact; explicitly choose a new directory instead of silent migration.
func Init(root string, c Installation) (*Manager, error) {
	home, _ := os.UserHomeDir()
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || root == "/" || root == home || !filepath.IsAbs(c.SkillRoot) {
		return nil, invalid("DEDICATED_DIRECTORY_REQUIRED")
	}
	c.Schema = Protocol
	c.Platform = platform()
	if err := trustedURL(c.Channel, "", c.Loopback); err != nil {
		return nil, err
	}
	if !compatibilityPattern.MatchString(c.Compatibility) {
		return nil, invalid("INSTALLATION_CONTRACT")
	}
	if err := trustedAstraURL(c.AstraURL, c.Loopback); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		return nil, err
	}
	real, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, err
	}
	// macOS /var aliases are canonicalized once, before storing ownership.
	if err := directory(root); err != nil {
		return nil, err
	}
	root = real
	if m, err := Open(root); err == nil {
		if m.Config != c {
			return nil, invalid("INSTALLATION_CONFLICT")
		}
		return m, nil
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	if len(entries) != 0 {
		return nil, invalid("INSTALLATION_NOT_EMPTY")
	}
	m := &Manager{Root: root, Config: c}
	lock, err := lockFile(root, "runtime.lock", true)
	if err != nil {
		return nil, err
	}
	defer lock.Close()
	if err := os.Mkdir(m.path("releases"), 0700); err != nil {
		return nil, err
	}
	if err := atomicJSON(m.path("installation.json"), c); err != nil {
		return nil, err
	}
	return m, nil
}
func (m *Manager) state() (State, error) {
	var s State
	err := readJSON(m.path("state.json"), &s)
	if isMissing(err) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if (s.Current != "" && !hashPattern.MatchString(s.Current)) || (s.Previous != "" && !hashPattern.MatchString(s.Previous)) {
		return s, invalid("STATE_CONTRACT")
	}
	return s, nil
}
func (m *Manager) localManifest(hash string) (Manifest, error) {
	var v Manifest
	if !hashPattern.MatchString(hash) {
		return v, invalid("STATE_CONTRACT")
	}
	p := m.path("releases/" + hash + "/manifest.json")
	if err := regular(p); err != nil {
		return v, err
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return v, err
	}
	if digest(b) != hash || json.Unmarshal(b, &v) != nil {
		return v, invalid("LOCAL_MANIFEST_CHECKSUM")
	}
	return v, m.validateManifest(v, v.Bundle)
}
func newerOrEqual(a, b string) bool {
	ax, bx := strings.Split(a, "."), strings.Split(b, ".")
	for i := range 3 {
		x, e1 := strconv.ParseUint(ax[i], 10, 64)
		y, e2 := strconv.ParseUint(bx[i], 10, 64)
		if e1 != nil || e2 != nil {
			return false
		}
		if x != y {
			return x > y
		}
	}
	return true
}
func (m *Manager) changes(v Manifest, s State) ([]string, error) {
	if v.Compatibility != m.Config.Compatibility || v.ClientConfig.AstraURL != m.Config.AstraURL {
		return nil, invalid("COMPATIBILITY_CHANGE")
	}
	if s.Current == "" {
		return append([]string(nil), names...), nil
	}
	old, err := m.localManifest(s.Current)
	if err != nil {
		return nil, err
	}
	if old.Compatibility != v.Compatibility || old.ClientConfig.AstraURL != v.ClientConfig.AstraURL {
		return nil, invalid("COMPATIBILITY_CHANGE")
	}
	var changed []string
	for _, n := range names {
		a, b := v.Platforms[m.Config.Platform].Components[n], old.Platforms[m.Config.Platform].Components[n]
		if a.Version == b.Version && a.Hash != b.Hash {
			return nil, invalid("VERSION_CONFLICT")
		}
		if !newerOrEqual(a.Version, b.Version) {
			return nil, invalid("DOWNGRADE_REJECTED")
		}
		if a.Hash != b.Hash {
			changed = append(changed, n)
		}
	}
	return changed, nil
}
func (m *Manager) Check(ctx context.Context) (Result, error) {
	c, v, _, err := m.check(ctx, true)
	if errors.Is(err, invalid("COMPATIBILITY_CHANGE")) {
		return Result{Status: "reinstall_required", Bundle: c.Bundle, Reason: err.Error(), Message: updateMessage(err)}, nil
	}
	if err != nil {
		return Result{}, err
	}
	s, err := m.state()
	if err != nil {
		return Result{}, err
	}
	changed, err := m.changes(v, s)
	if err != nil {
		return Result{}, err
	}
	status := "available"
	if c.Hash == s.Current {
		status = "current"
	}
	return Result{Status: status, Bundle: c.Bundle, Changed: changed}, nil
}

type transaction struct {
	Old    State  `json:"old"`
	Target string `json:"target"`
}

func (m *Manager) skillStore() string {
	return filepath.Join(m.Config.SkillRoot, ".moi-client-releases")
}
func (m *Manager) ownSkills() error {
	if err := os.MkdirAll(m.Config.SkillRoot, 0755); err != nil {
		return err
	}
	if err := directory(m.Config.SkillRoot); err != nil {
		return err
	}
	store := m.skillStore()
	if err := os.Mkdir(store, 0700); err == nil {
		if err := atomicJSON(filepath.Join(store, "owner.json"), m.Root); err != nil {
			return err
		}
	} else if !os.IsExist(err) {
		return err
	}
	if err := directory(store); err != nil {
		return err
	}
	var owner string
	if err := readJSON(filepath.Join(store, "owner.json"), &owner); err != nil || owner != m.Root {
		return invalid("SKILL_NOT_OWNED")
	}
	link := filepath.Join(m.Config.SkillRoot, "moi-cli")
	if _, err := os.Lstat(link); isMissing(err) {
		return nil
	} else if err != nil {
		return err
	}
	target, err := os.Readlink(link)
	if err != nil || filepath.Dir(target) != store || !hashPattern.MatchString(filepath.Base(target)) {
		return invalid("SKILL_NOT_OWNED")
	}
	return nil
}
func (m *Manager) activate(hash string) error {
	skillTarget, current := "", ""
	if hash != "" {
		if !hashPattern.MatchString(hash) {
			return invalid("STATE_CONTRACT")
		}
		skillTarget = filepath.Join(m.skillStore(), hash)
		current = "releases/" + hash
	}
	if err := pointer(filepath.Join(m.Config.SkillRoot, "moi-cli"), skillTarget); err != nil {
		return err
	}
	return pointer(m.path("current"), current)
}
func (m *Manager) recover() error {
	var tx transaction
	if err := readJSON(m.path("transaction.json"), &tx); isMissing(err) {
		return nil
	} else if err != nil {
		return err
	}
	if !hashPattern.MatchString(tx.Target) || (tx.Old.Current != "" && !hashPattern.MatchString(tx.Old.Current)) || (tx.Old.Previous != "" && !hashPattern.MatchString(tx.Old.Previous)) {
		return invalid("TRANSACTION_CONTRACT")
	}
	if err := m.ownSkills(); err != nil {
		return err
	}
	if tx.Old.Current != "" {
		if err := m.readyRelease(tx.Old.Current); err != nil {
			return err
		}
	}
	if err := m.activate(tx.Old.Current); err != nil {
		return err
	}
	if err := atomicJSON(m.path("state.json"), tx.Old); err != nil {
		return err
	}
	if err := os.Remove(m.path("transaction.json")); err != nil {
		return err
	}
	return syncDir(m.Root)
}
func (m *Manager) commit(old State, target string) error {
	if err := atomicJSON(m.path("transaction.json"), transaction{Old: old, Target: target}); err != nil {
		return err
	}
	// On any failure leave the durable journal. No ordinary client is admitted
	// until the explicit updater has recovered both pointers and state.
	if err := m.activate(target); err != nil {
		return err
	}
	if err := atomicJSON(m.path("state.json"), State{Current: target, Previous: old.Current}); err != nil {
		return err
	}
	if err := os.Remove(m.path("transaction.json")); err != nil {
		return err
	}
	return syncDir(m.Root)
}
func (m *Manager) Apply(ctx context.Context, rollback bool) (Result, error) {
	m.progress.stage("Checking installation")
	lock, err := lockFile(m.Root, "runtime.lock", true)
	if err != nil {
		return Result{}, err
	}
	defer lock.Close()
	if err := directory(m.path("releases")); err != nil {
		return Result{}, err
	}
	if err := m.recover(); err != nil {
		return Result{}, err
	}
	s, err := m.state()
	if err != nil {
		return Result{}, err
	}
	if err := m.ownSkills(); err != nil {
		return Result{}, err
	}
	if err := m.clearReinstallPointer(s); err != nil {
		return Result{}, err
	}
	if err := m.validatePointers(s); err != nil {
		return Result{}, err
	}
	if rollback {
		m.progress.stage("Verifying rollback release")
		if s.Previous == "" {
			return Result{}, invalid("NO_PREVIOUS_RELEASE")
		}
		old, err := m.localManifest(s.Current)
		if err != nil {
			return Result{}, err
		}
		prev, err := m.localManifest(s.Previous)
		if err != nil {
			return Result{}, err
		}
		if old.Compatibility != prev.Compatibility {
			return Result{}, invalid("COMPATIBILITY_CHANGE")
		}
		if err := m.readyRelease(s.Previous); err != nil {
			return Result{}, err
		}
		if err := m.commit(s, s.Previous); err != nil {
			return Result{}, err
		}
		return Result{Status: "rolled_back", Bundle: prev.Bundle}, nil
	}
	m.progress.stage("Checking release")
	c, v, data, err := m.check(ctx, true)
	if err != nil {
		return Result{}, err
	}
	changed, err := m.changes(v, s)
	if err != nil {
		return Result{}, err
	}
	if c.Hash == s.Current {
		m.progress.stage("Verifying installed files")
		if err := m.readyRelease(s.Current); err != nil {
			return Result{}, err
		}
		return Result{Status: "current", Bundle: v.Bundle}, nil
	}
	m.progress.stage("Preparing release")
	if err := m.prepare(ctx, c, v, data, s); err != nil {
		return Result{}, err
	}
	m.progress.stage("Activating release")
	if err := m.commit(s, c.Hash); err != nil {
		return Result{}, err
	}
	return Result{Status: "updated", Bundle: v.Bundle, Changed: changed}, nil
}

func (r Result) String() string { return fmt.Sprintf("%s: %s", r.Status, r.Bundle) }

// Only a fresh, empty prefix may reclaim its same-owner residual Skill link.
// A missing state file in an existing installation is not a reinstall permit.
// Caller holds runtime.lock and has validated ownership with ownSkills.
func (m *Manager) clearReinstallPointer(s State) error {
	if s.Current != "" {
		return nil
	}
	link := filepath.Join(m.Config.SkillRoot, "moi-cli")
	if _, err := os.Lstat(link); isMissing(err) {
		return nil
	} else if err != nil {
		return err
	}
	for _, name := range []string{"current", "state.json"} {
		if _, err := os.Lstat(m.path(name)); !isMissing(err) {
			return invalid("POINTER_STATE_MISMATCH")
		}
	}
	entries, err := os.ReadDir(m.path("releases"))
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return invalid("POINTER_STATE_MISMATCH")
	}
	return pointer(link, "")
}

func (m *Manager) validatePointers(s State) error {
	for p, want := range map[string]string{m.path("current"): "releases/" + s.Current, filepath.Join(m.Config.SkillRoot, "moi-cli"): filepath.Join(m.skillStore(), s.Current)} {
		target, err := os.Readlink(p)
		if s.Current == "" && isMissing(err) {
			continue
		}
		if err != nil || s.Current == "" || target != want {
			return invalid("POINTER_STATE_MISMATCH")
		}
	}
	return nil
}
