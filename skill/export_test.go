package skill

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestExportIsDeterministicAndNeverOverwrites(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "Unicode 工作区 ' spaced", "moi-cli")
	if err := Export(dir); err != nil {
		t.Fatal(err)
	}
	if err := Export(dir); err == nil {
		t.Fatal("existing directory overwritten")
	}
	if err := fs.WalkDir(assets, "moi-cli", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		want, err := assets.ReadFile(name)
		if err != nil {
			return err
		}
		got, err := os.ReadFile(filepath.Join(dir, strings.TrimPrefix(name, "moi-cli/")))
		if err != nil {
			return err
		}
		if !bytes.Equal(got, want) {
			t.Errorf("export mismatch %s", name)
		}
		if strings.HasSuffix(name, ".md") {
			read, err := Read(strings.TrimPrefix(name, "moi-cli/"))
			if err != nil || !bytes.Equal(read, want) {
				t.Errorf("embedded read differs from export %s: %v", name, err)
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	if err := Export(link); err == nil {
		t.Fatal("symlink target accepted")
	}
	if err := Export(""); err == nil {
		t.Fatal("empty target accepted")
	}
	file := filepath.Join(t.TempDir(), "owned")
	if err := os.WriteFile(file, []byte("user"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := Export(file); err == nil {
		t.Fatal("file overwritten")
	}
	data, _ := os.ReadFile(file)
	if string(data) != "user" {
		t.Fatal("user data changed")
	}
}

func TestReadOnlyEmbeddedDocuments(t *testing.T) {
	for _, name := range []string{"", ".", "../SKILL.md", "/etc/passwd", "references/../SKILL.md", "references//files.md", "auth.json", "references", "references/"} {
		data, err := Read(name)
		if len(data) != 0 || err == nil || !strings.HasPrefix(err.Error(), "CLI_SKILL_DOCUMENT_INVALID:") {
			t.Errorf("invalid path %q: %q, %v", name, data, err)
		}
	}
	data, err := Read("references/missing.md")
	if len(data) != 0 || err == nil || !strings.HasPrefix(err.Error(), "CLI_SKILL_DOCUMENT_NOT_FOUND:") || strings.Contains(err.Error(), "assets/") {
		t.Fatalf("missing document must have a safe, distinct error: %q, %v", data, err)
	}
}

func TestSkillLinksAndDomainCoverage(t *testing.T) {
	linkRE := regexp.MustCompile(`\]\(([^)]+)\)`)
	err := fs.WalkDir(assets, "moi-cli", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := assets.ReadFile(name)
		if err != nil {
			return err
		}
		if name == "moi-cli/SKILL.md" {
			if !bytes.HasPrefix(data, []byte("---\nname: moi-cli\ndescription:")) {
				t.Fatal("frontmatter missing")
			}
			if len(data) > 8000 {
				t.Fatal("entry must remain a short routing document")
			}
		}
		for _, match := range linkRE.FindAllStringSubmatch(string(data), -1) {
			if strings.Contains(match[1], "://") {
				continue
			}
			target, _, _ := strings.Cut(match[1], "#")
			if target == "" {
				continue
			}
			target = path.Join(path.Dir(name), target)
			if _, err := assets.ReadFile(target); err != nil {
				t.Errorf("broken link %s -> %s", name, target)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, domain := range []string{"workspace", "workspace-member", "catalog", "connector", "sql", "workflow", "semantic-model", "agent", "iam", "account", "model", "provider-backend", "compute", "dashboard", "data-share", "config"} {
		if _, err := assets.ReadFile("moi-cli/references/commands/" + domain + ".md"); err != nil {
			t.Error(err)
		}
	}
}
