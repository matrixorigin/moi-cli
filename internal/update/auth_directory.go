//go:build darwin || linux

package update

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// prepareAuthDirectory prepares only the directory, never Astra's credentials.
// Use the opened directory's identity for both ownership checks and chmod: a
// path swap must not cause chmod to follow a newly installed symlink.
func prepareAuthDirectory() error {
	root, set := os.LookupEnv("MOI_AUTH_DIR")
	if !set {
		home, err := os.UserHomeDir()
		if err != nil {
			return invalid("AUTH_DIRECTORY")
		}
		root = filepath.Join(home, ".moi")
	}
	// A trailing separator is harmless for a directory, but must be stripped
	// before O_NOFOLLOW so that a final symlink cannot become an intermediate one.
	root = strings.TrimRight(root, "/")
	// Do not silently change path semantics: cleaning a path containing a
	// symlink followed by '..' can select a different directory than Astra.
	if !filepath.IsAbs(root) || filepath.Clean(root) != root || root == "/" {
		return invalid("AUTH_DIRECTORY")
	}
	if err := os.Mkdir(root, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return invalid("AUTH_DIRECTORY")
	}
	fd, err := syscall.Open(root, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return invalid("AUTH_DIRECTORY")
	}
	f := os.NewFile(uintptr(fd), root)
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.IsDir() {
		return invalid("AUTH_DIRECTORY")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) {
		return invalid("AUTH_DIRECTORY")
	}
	if info.Mode().Perm() != 0700 {
		if err := f.Chmod(0700); err != nil {
			return invalid("AUTH_DIRECTORY")
		}
	}
	current, err := os.Lstat(root)
	if err != nil || !os.SameFile(info, current) || current.Mode().Perm() != 0700 || !current.IsDir() {
		return invalid("AUTH_DIRECTORY")
	}
	return nil
}
