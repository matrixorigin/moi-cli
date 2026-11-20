//go:build darwin || linux

package commands

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestUploadRejectsSpecialAndUnreadableFiles(t *testing.T) {
	fifo := filepath.Join(t.TempDir(), "pipe")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{fifo, "/dev/null"} {
		budget := defaultMaxUploadBytes
		_, err := openUploadFile(path, &budget)
		var inputErr *LocalInputError
		if !errors.As(err, &inputErr) || inputErr.Code != "CLI_FILE_NOT_REGULAR" {
			t.Fatalf("special file: %v", err)
		}
	}
	link := filepath.Join(t.TempDir(), "directory-link")
	if err := os.Symlink(t.TempDir(), link); err != nil {
		t.Fatal(err)
	}
	budget := defaultMaxUploadBytes
	if _, err := openUploadFile(link, &budget); err == nil {
		t.Fatal("symlink to directory accepted")
	}
	t.Run("unreadable", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("root can read mode-000 files; requires non-root test runner")
		}
		path := uploadFixture(t, "private", []byte("data"))
		if err := os.Chmod(path, 0); err != nil {
			t.Fatal(err)
		}
		_, err := openUploadFile(path, &budget)
		var inputErr *LocalInputError
		if !errors.As(err, &inputErr) || inputErr.Code != "CLI_FILE_UNREADABLE" {
			t.Fatalf("unreadable file: %v", err)
		}
	})
}
