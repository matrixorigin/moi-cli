package update

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func isMissing(err error) bool { return errors.Is(err, os.ErrNotExist) }
func regular(path string) error {
	s, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !s.Mode().IsRegular() {
		return invalid("UNSAFE_FILE")
	}
	return nil
}
func directory(path string) error {
	s, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !s.IsDir() || s.Mode()&os.ModeSymlink != 0 {
		return invalid("UNSAFE_DIRECTORY")
	}
	return nil
}
func readJSON(path string, dest any) error {
	if err := regular(path); err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, (2<<20)+1))
	if err != nil {
		return err
	}
	if len(data) > 2<<20 {
		return invalid("METADATA_SIZE")
	}
	if json.Unmarshal(data, dest) != nil {
		return invalid("METADATA_JSON")
	}
	return nil
}
func syncDir(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
func atomicJSON(path string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return atomicFile(path, data)
}
func atomicFile(path string, data []byte) error {
	if err := regular(path); err != nil && !isMissing(err) {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".metadata-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}

// A stable inode shared by Go flock and Rust fs2. Never unlink lock files.
func lockFile(root, name string, exclusive bool) (*os.File, error) {
	p := filepath.Join(root, name)
	if err := regular(p); err != nil && !isMissing(err) {
		return nil, err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	operation := syscall.LOCK_SH | syscall.LOCK_NB
	if exclusive {
		operation = syscall.LOCK_EX | syscall.LOCK_NB
	}
	if err = syscall.Flock(int(f.Fd()), operation); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, invalid("BUSY")
		}
		return nil, err
	}
	return f, nil
}

// Explicit commands briefly wait for the bounded startup refresh. Runtime
// occupancy is separate and remains fail-fast. Background refresh never waits.
func (m *Manager) checkLock(ctx context.Context, wait bool) (*os.File, error) {
	deadline, cancel := context.WithTimeout(ctx, 4*time.Second)
	defer cancel()
	for {
		if ctx.Err() != nil {
			return nil, invalid("CHECK_CANCELLED")
		}
		f, err := lockFile(m.Root, "check.lock", true)
		if !errors.Is(err, invalid("BUSY")) {
			return f, err
		}
		if !wait {
			return nil, invalid("CHECK_BUSY")
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-deadline.Done():
			timer.Stop()
			if ctx.Err() != nil {
				return nil, invalid("CHECK_CANCELLED")
			}
			return nil, invalid("CHECK_BUSY")
		case <-timer.C:
		}
	}
}

func pointer(path, target string) error {
	if s, err := os.Lstat(path); err == nil && s.Mode()&os.ModeSymlink == 0 {
		return invalid("UNOWNED_POINTER")
	} else if err != nil && !isMissing(err) {
		return err
	}
	if target == "" {
		if err := os.Remove(path); err != nil && !isMissing(err) {
			return err
		}
		return syncDir(filepath.Dir(path))
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".pointer-")
	if err != nil {
		return err
	}
	f.Close()
	os.Remove(f.Name())
	defer os.Remove(f.Name())
	if err = os.Symlink(target, f.Name()); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return err
	}
	return syncDir(filepath.Dir(path))
}

// Managed installation discovery is executable-relative, never PATH/env based.
func Discover(executable string) (*Manager, error) {
	exe, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return nil, err
	}
	release := filepath.Dir(exe)
	if filepath.Base(filepath.Dir(release)) != "releases" || !hashPattern.MatchString(filepath.Base(release)) {
		return nil, nil
	}
	root := filepath.Dir(filepath.Dir(release))
	m, err := Open(root)
	if isMissing(err) {
		return nil, nil
	}
	return m, err
}

func Acquire(executable string) (*os.File, error) {
	m, err := Discover(executable)
	if err != nil || m == nil {
		return nil, err
	}
	f, err := lockFile(m.Root, "runtime.lock", false)
	if err != nil {
		return nil, err
	}
	reject := func(err error) (*os.File, error) { f.Close(); return nil, err }
	if _, err := os.Lstat(m.path("transaction.json")); !isMissing(err) {
		return reject(invalid("RECOVERY_REQUIRED"))
	}
	exe, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return reject(err)
	}
	current, err := filepath.EvalSymlinks(m.path("current"))
	if err != nil || current != filepath.Dir(exe) {
		return reject(invalid("RESTART_REQUIRED"))
	}
	return f, nil
}
