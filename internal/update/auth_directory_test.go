//go:build darwin || linux

package update

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallRejectsUnsafeAuthDirectoryBeforeActivation(t *testing.T) {
	base := t.TempDir()
	auth := filepath.Join(base, "auth")
	if err := os.WriteFile(auth, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOI_AUTH_DIR", auth)
	var out bytes.Buffer
	prefix := filepath.Join(base, "client")
	err := Run(t.Context(), []string{"install", "--dir", prefix, "--skill-dir", filepath.Join(base, "skills"), "--channel", "https://downloads.example.test/channels/stable.json", "--compatibility", "v1", "--astra-url", "https://astra.example.test"}, &out)
	if err == nil || !strings.Contains(PublicError(err), "CLI_UPDATE_AUTH_DIRECTORY:") || out.Len() != 0 {
		t.Fatalf("error=%v output=%s", err, out.String())
	}
	if _, err := os.Lstat(filepath.Join(prefix, "current")); !os.IsNotExist(err) {
		t.Fatalf("installation activated: %v", err)
	}
}

func TestPrepareAuthDirectory(t *testing.T) {
	for _, kind := range []string{"missing", "existing", "trailing", "symlink", "symlink-trailing", "dangling", "file", "relative", "empty", "root", "unclean"} {
		t.Run(kind, func(t *testing.T) {
			base := t.TempDir()
			root := filepath.Join(base, "auth with spaces")
			target := filepath.Join(base, "target")
			if err := os.Mkdir(target, 0755); err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "existing":
				if err := os.Mkdir(root, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, "auth.json"), []byte("preserve"), 0600); err != nil {
					t.Fatal(err)
				}
			case "trailing":
				root += "///"
			case "symlink", "symlink-trailing", "dangling":
				link := target
				if kind == "dangling" {
					link = filepath.Join(base, "absent")
				}
				if err := os.Symlink(link, root); err != nil {
					t.Fatal(err)
				}
				if kind == "symlink-trailing" {
					root += "/"
				}
			case "file":
				if err := os.WriteFile(root, []byte("preserve"), 0600); err != nil {
					t.Fatal(err)
				}
			case "relative":
				root = "relative"
			case "empty":
				root = ""
			case "root":
				root = "/"
			case "unclean":
				root = base + "/target/../auth"
			}
			t.Setenv("MOI_AUTH_DIR", root)
			err := prepareAuthDirectory()
			if kind != "missing" && kind != "existing" && kind != "trailing" {
				if err == nil || !strings.Contains(PublicError(err), "CLI_UPDATE_AUTH_DIRECTORY:") {
					t.Fatalf("error = %v", err)
				}
				info, _ := os.Stat(target)
				if info.Mode().Perm() != 0755 {
					t.Fatal("changed symlink target")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := prepareAuthDirectory(); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(root)
			if err != nil || info.Mode().Perm() != 0700 {
				t.Fatalf("mode: %v %v", info, err)
			}
			if kind == "existing" {
				b, err := os.ReadFile(filepath.Join(root, "auth.json"))
				if err != nil || string(b) != "preserve" {
					t.Fatal("credentials changed")
				}
			}
		})
	}
}

func TestPrepareDefaultAuthDirectory(t *testing.T) {
	old, set := os.LookupEnv("MOI_AUTH_DIR")
	os.Unsetenv("MOI_AUTH_DIR")
	t.Cleanup(func() {
		if set {
			os.Setenv("MOI_AUTH_DIR", old)
		}
	})
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := prepareAuthDirectory(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(home, ".moi"))
	if err != nil || info.Mode().Perm() != 0700 {
		t.Fatalf("mode: %v %v", info, err)
	}
}
