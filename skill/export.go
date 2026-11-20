// Package skill owns the product-distributed Skill, not repository rules.
package skill

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed moi-cli
var assets embed.FS

// Files exposes the immutable documentation paired with this product release.
// MOI uses it to prepare conversation references without consulting a sandbox
// image or maintaining a second copy of the CLI command documentation.
func Files() fs.FS {
	documents, err := fs.Sub(assets, "moi-cli")
	if err != nil {
		panic(err) // The embedded assets directory is a build-time invariant.
	}
	return documents
}

// Read returns a document from the paired, embedded Skill. It never reads the
// host filesystem, resolves installation paths, or accesses credentials.
func Read(name string) ([]byte, error) {
	if !fs.ValidPath(name) || !strings.HasSuffix(name, ".md") {
		return nil, fmt.Errorf("CLI_SKILL_DOCUMENT_INVALID: use a relative Markdown path such as references/files.md")
	}
	data, err := assets.ReadFile("moi-cli/" + name)
	if err != nil {
		return nil, fmt.Errorf("CLI_SKILL_DOCUMENT_NOT_FOUND: document is not included in this CLI; read references/index.md for available domains")
	}
	return data, nil
}

// Export creates a new directory; it never replaces caller-owned data.
func Export(dir string) (err error) {
	if dir == "" {
		return fmt.Errorf("CLI_SKILL_DIRECTORY_REQUIRED: choose a new export directory")
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return fmt.Errorf("CLI_SKILL_PARENT_UNAVAILABLE: export parent is not writable")
	}
	if err := os.Mkdir(dir, 0755); err != nil {
		return fmt.Errorf("CLI_SKILL_TARGET_UNAVAILABLE: target must be new and writable")
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(dir)
		}
	}()
	return fs.WalkDir(assets, "moi-cli", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel("moi-cli", path)
		if err != nil {
			return err
		}
		target := filepath.Join(dir, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := assets.ReadFile(path)
		if err != nil {
			return err
		}
		file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, writeErr := file.Write(data)
		closeErr := file.Close()
		if writeErr != nil {
			return writeErr
		}
		return closeErr
	})
}
