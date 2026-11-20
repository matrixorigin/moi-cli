package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func fileDigest(path string) (string, error) {
	if err := regular(path); err != nil {
		return "", err
	}
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func copyFile(source, target string) error {
	if err := regular(source); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if err != nil {
		out.Close()
		return err
	}
	if err = out.Sync(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
func extract(archive, dir, name string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return invalid("ARCHIVE_FORMAT")
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return invalid("ARCHIVE_FORMAT")
		}
		if (h.Name != name && !(name != "moi-cli" && h.Name == "LICENSE")) || seen[h.Name] || h.Typeflag != tar.TypeReg || h.Mode&07000 != 0 || h.Size < 0 || h.Size > maxArchive {
			return invalid("ARCHIVE_MEMBER")
		}
		seen[h.Name] = true
		filename := h.Name
		if filename == "LICENSE" {
			filename = name + "-LICENSE"
		}
		out, err := os.OpenFile(filepath.Join(dir, filename), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return err
		}
		n, err := io.Copy(out, tr)
		if err != nil || n != h.Size {
			out.Close()
			return invalid("ARCHIVE_TRUNCATED")
		}
		if err = out.Sync(); err != nil {
			out.Close()
			return err
		}
		if err = out.Close(); err != nil {
			return err
		}
	}
	// Consume gzip trailer so CRC errors cannot hide behind tar EOF.
	if _, err := io.Copy(io.Discard, io.LimitReader(gz, 1025)); err != nil {
		return invalid("ARCHIVE_FORMAT")
	}
	if !seen[name] || (name != "moi-cli" && !seen["LICENSE"]) {
		return invalid("ARCHIVE_MISSING_MEMBER")
	}
	return os.Chmod(filepath.Join(dir, name), 0755)
}

// versionProbeError keeps CLI_UPDATE_VERSION_PROBE stable while naming the
// component that failed. Loader text stays in the installation probe log.
type versionProbeError struct {
	component string
}

func (e *versionProbeError) Error() string { return "CLI_UPDATE_VERSION_PROBE" }

func probe(ctx context.Context, m *Manager, dir string, v Manifest) error {
	p := m.Config.Platform
	for _, name := range names {
		probeCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		cmd := exec.CommandContext(probeCtx, filepath.Join(dir, name), "--version")
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		output, err := cmd.Output()
		timedOut := errors.Is(probeCtx.Err(), context.DeadlineExceeded)
		cancel()
		if err != nil {
			return m.failVersionProbe(name, probeFailureDetail(err, timedOut, stderr.Bytes()))
		}
		expected := name + " " + v.Platforms[p].Components[name].Version
		if name == "moi-cli" {
			expected = "moi-cli version " + v.Platforms[p].Components[name].Version
		}
		if strings.TrimSpace(string(output)) != expected {
			return m.failVersionProbe(name, output)
		}
		probeCtx, cancel = context.WithTimeout(ctx, 15*time.Second)
		output, err = exec.CommandContext(probeCtx, filepath.Join(dir, name), "--moi-update-protocol").Output()
		cancel()
		if err != nil || strings.TrimSpace(string(output)) != Protocol {
			return invalid("OCCUPANCY_PROTOCOL_REQUIRED")
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, filepath.Join(dir, "moi-cli"), "skill", "export", "--dir", filepath.Join(dir, "skills", "moi-cli")).Run(); err != nil {
		return invalid("SKILL_EXPORT")
	}
	return regular(filepath.Join(dir, "skills", "moi-cli", "SKILL.md"))
}

func (m *Manager) failVersionProbe(component string, detail []byte) error {
	probeErr := &versionProbeError{component: component}
	if !slices.Contains(names, component) {
		return probeErr
	}
	if err := m.appendProbeLog(component, sanitizeProbeDetail(detail)); err != nil {
		return errors.Join(probeErr, err)
	}
	return probeErr
}

func probeFailureDetail(err error, timedOut bool, stderr []byte) []byte {
	var detail bytes.Buffer
	if timedOut {
		detail.WriteString("timed_out=true\n")
	}
	fmt.Fprintf(&detail, "error=%s\n", err)
	if len(stderr) > 0 {
		detail.WriteString("stderr:\n")
		detail.Write(stderr)
	}
	return detail.Bytes()
}

func sanitizeProbeDetail(detail []byte) string {
	text := strings.TrimSpace(strings.ReplaceAll(string(detail), "\x00", ""))
	const limit = 2048
	if len(text) > limit {
		text = text[:limit]
	}
	return text
}

func (m *Manager) appendProbeLog(component, detail string) error {
	f, err := os.OpenFile(m.path("probe.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s component=%s operation=version_probe\n%s\n",
		time.Now().UTC().Format(time.RFC3339), component, detail)
	return err
}

func treeHashes(root string) (map[string]string, error) {
	result := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return invalid("UNSAFE_FILE")
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "files.json" {
			return nil
		}
		h, err := fileDigest(path)
		if err != nil {
			return err
		}
		result[rel] = h
		return nil
	})
	return result, err
}
func equalHashes(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
func (m *Manager) verifyRelease(hash string) error {
	if err := m.verifyReleaseFiles(hash); err != nil {
		return err
	}
	return m.skillCopy(hash, m.path("releases/"+hash+"/skills/moi-cli"), false)
}

// Mutation is explicit and limited to missing derived Skill directories. All
// callers hold runtime.lock exclusively and validate the Skill store owner.
func (m *Manager) readyRelease(hash string) error {
	if err := m.verifyReleaseFiles(hash); err != nil {
		return err
	}
	return m.skillCopy(hash, m.path("releases/"+hash+"/skills/moi-cli"), true)
}

func (m *Manager) verifyReleaseFiles(hash string) error {
	if _, err := m.localManifest(hash); err != nil {
		return err
	}
	dir := m.path("releases/" + hash)
	if err := directory(dir); err != nil {
		return err
	}
	var expected map[string]string
	if err := readJSON(filepath.Join(dir, "files.json"), &expected); err != nil {
		return err
	}
	actual, err := treeHashes(dir)
	if err != nil {
		return err
	}
	if !equalHashes(expected, actual) {
		return invalid("LOCAL_RELEASE_MODIFIED")
	}
	for _, name := range names {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if info.Mode().Perm() != 0755 {
			return invalid("LOCAL_RELEASE_MODIFIED")
		}
	}
	return nil
}

// source belongs to a verified release or the validated preparation stage.
// An existing modified copy is never overwritten, including partial deletion.
func (m *Manager) skillCopy(hash, source string, createMissing bool) error {
	target := filepath.Join(m.skillStore(), hash)
	if err := directory(source); err != nil {
		return err
	}
	a, err := treeHashes(source)
	if err != nil {
		return err
	}
	if err := directory(target); err == nil {
		b, err := treeHashes(target)
		if err != nil {
			return err
		}
		if !equalHashes(a, b) {
			return invalid("LOCAL_SKILL_MODIFIED")
		}
		return nil
	} else if !isMissing(err) {
		return err
	}
	if !createMissing {
		return invalid("SKILL_COPY_MISSING")
	}
	tmp, err := os.MkdirTemp(m.skillStore(), ".stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	staged := filepath.Join(tmp, "skill")
	if err := copyTree(source, staged); err != nil {
		return err
	}
	b, err := treeHashes(staged)
	if err != nil {
		return err
	}
	if !equalHashes(a, b) {
		return invalid("LOCAL_SKILL_MODIFIED")
	}
	if err := syncTree(staged); err != nil {
		return err
	}
	if err := os.Rename(staged, target); err != nil {
		return err
	}
	return syncDir(m.skillStore())
}
func copyTree(from, to string) error {
	if err := os.Mkdir(to, 0755); err != nil {
		return err
	}
	return filepath.WalkDir(from, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(from, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		dest := filepath.Join(to, rel)
		if d.IsDir() {
			return os.Mkdir(dest, 0755)
		}
		return copyFile(p, dest)
	})
}
func (m *Manager) prepare(ctx context.Context, c Channel, v Manifest, data []byte, s State) error {
	final := m.path("releases/" + c.Hash)
	if _, err := os.Lstat(final); err == nil {
		return m.readyRelease(c.Hash)
	} else if !isMissing(err) {
		return err
	}
	stage, err := os.MkdirTemp(m.path("releases"), ".stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stage)
	var old Manifest
	if s.Current != "" {
		m.progress.stage("Verifying installed files")
		if err := m.readyRelease(s.Current); err != nil {
			return err
		}
		old, err = m.localManifest(s.Current)
		if err != nil {
			return err
		}
	}
	for _, name := range names {
		item := v.Platforms[m.Config.Platform].Components[name]
		archive := filepath.Join(stage, name+".tar.gz")
		previous := old.Platforms[m.Config.Platform].Components[name]
		if s.Current != "" && item.Hash == previous.Hash {
			m.progress.stage("Reusing " + name)
			err = copyFile(m.path("releases/"+s.Current+"/"+name+".tar.gz"), archive)
		} else {
			m.progress.stage("Downloading " + name)
			var out *os.File
			out, err = os.OpenFile(archive, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
			if err == nil {
				err = m.download(ctx, item.URL, item.Size, m.progress.writer(out, name, item.Size))
				if e := out.Sync(); err == nil {
					err = e
				}
				if e := out.Close(); err == nil {
					err = e
				}
			}
		}
		if err != nil {
			return err
		}
		m.progress.stage("Verifying " + name)
		info, err := os.Stat(archive)
		if err != nil {
			return err
		}
		hash, err := fileDigest(archive)
		if err != nil {
			return err
		}
		if info.Size() != item.Size || hash != item.Hash {
			return invalid("COMPONENT_CHECKSUM")
		}
		m.progress.stage("Extracting " + name)
		if err := extract(archive, stage, name); err != nil {
			return err
		}
	}
	m.progress.stage("Checking binaries and Skill")
	if err := probe(ctx, m, stage, v); err != nil {
		return err
	}
	if err := atomicFile(filepath.Join(stage, "manifest.json"), data); err != nil {
		return err
	}
	hashes, err := treeHashes(stage)
	if err != nil {
		return err
	}
	if err := atomicJSON(filepath.Join(stage, "files.json"), hashes); err != nil {
		return err
	}
	// Publish the confined Skill copy before the release; interrupted preparation
	// may leave an immutable orphan, validated and reused on the next attempt.
	if err := m.skillCopy(c.Hash, filepath.Join(stage, "skills", "moi-cli"), true); err != nil {
		return err
	}
	if err := syncTree(stage); err != nil {
		return err
	}
	if err := os.Rename(stage, final); err != nil {
		return err
	}
	return syncDir(m.path("releases"))
}

func syncTree(root string) error {
	var dirs []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, p)
			return nil
		}
		if err := regular(p); err != nil {
			return err
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		err = f.Sync()
		f.Close()
		return err
	})
	if err != nil {
		return err
	}
	slices.Reverse(dirs)
	for _, d := range dirs {
		if err := syncDir(d); err != nil {
			return err
		}
	}
	return nil
}
