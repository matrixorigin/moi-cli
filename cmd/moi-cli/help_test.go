package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

// Exercise the public binary, including flag.ExitOnError, for every dispatched
// product command (not just commands discovered through a literal FlagSet).
func TestSubcommandHelpIsOfflineAndUsesRealFlags(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "moi-cli")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	home := t.TempDir()
	config := filepath.Join(home, ".moi-cli.json")
	before := []byte("{broken-config")
	if err := os.WriteFile(config, before, 0600); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		http.Error(w, "help must not send a request", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	run := func(args ...string) (string, error) {
		cmd := exec.CommandContext(t.Context(), binary, append([]string{"-endpoint", server.URL, "-auth", "moi"}, args...)...)
		cmd.Env = append(os.Environ(), "HOME="+home, "MOI_AUTH_DIR="+home,
			"MOI_ASTRA_BINARY="+filepath.Join(home, "missing-helper"), "MOI_AGENT_CALL=1")
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	files, err := filepath.Glob("../../commands/*.go")
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_generated.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		domain := strings.ReplaceAll(strings.TrimSuffix(filepath.Base(path), ".go"), "_", "-")
		if domain == "schema" {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(fn.Name.Name, "Execute") || !strings.HasSuffix(fn.Name.Name, "Command") {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				clause, ok := n.(*ast.CaseClause)
				if !ok {
					return true
				}
				for _, expr := range clause.List {
					lit, ok := expr.(*ast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						continue
					}
					sub, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatal(err)
					}
					count++
					t.Run(domain+"/"+sub, func(t *testing.T) {
						out, err := run(domain, sub, "--help")
						if err != nil || !strings.Contains(out, "Usage: moi-cli ") || strings.Contains(out, "Subcommands:") {
							t.Fatalf("subcommand help: %v\n%s", err, out)
						}
					})
				}
				return false
			})
		}
	}
	if count < 300 {
		t.Fatalf("unexpectedly few dispatched commands: %d", count)
	}
	t.Logf("checked %d dispatched subcommands", count)
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"sql", "execute", "--query-file", "/nonexistent/query.sql", "--help"}, "-query-file"},
		{[]string{"workflow", "run", "--workflow-id", "must-not-run", "--help"}, "-workflow-id"},
		{[]string{"connector", "upload", "--file-path", "/nonexistent/input.csv", "--help"}, "-file-path"},
		{[]string{"sql", "describe", "--help"}, "-query-id"},
		{[]string{"sql", "result", "--help"}, "-statement-id"},
		{[]string{"workspace", "--help"}, "Subcommands:"},
	} {
		out, err := run(tc.args...)
		if err != nil || !strings.Contains(out, tc.want) {
			t.Errorf("%v: %v\n%s", tc.args, err, out)
		}
	}
	for _, args := range [][]string{{"sql", "nonexistent", "--help"}, {"nonexistent", "--help"}} {
		if out, err := run(args...); err == nil || !strings.Contains(out, "unknown") {
			t.Errorf("unknown command should fail: %v: %v\n%s", args, err, out)
		}
	}
	// The paired reference reader must not depend on login, config, CWD or
	// a discoverable installed Skill directory. Compare it with the export source.
	want, err := os.ReadFile("../../skill/moi-cli/references/files.md")
	if err != nil {
		t.Fatal(err)
	}
	if out, err := run("skill", "read", "--path", "references/files.md"); err != nil || out != string(want) {
		t.Fatalf("offline Skill read: %v\n%s", err, out)
	}
	for _, tc := range []struct {
		args []string
		code string
	}{
		{[]string{"skill", "read", "--path", "../credentials.md"}, "CLI_SKILL_DOCUMENT_INVALID"},
		{[]string{"skill", "read", "--path", "references/missing.md"}, "CLI_SKILL_DOCUMENT_NOT_FOUND"},
		{[]string{"skill", "read", "unexpected"}, "CLI_SKILL_UNEXPECTED_ARGUMENT"},
	} {
		if out, err := run(tc.args...); err == nil || !strings.Contains(out, tc.code) {
			t.Errorf("%v: %v\n%s", tc.args, err, out)
		}
	}
	if requests.Load() != 0 {
		t.Fatal("help contacted the product server")
	}
	if after, err := os.ReadFile(config); err != nil || !bytes.Equal(before, after) {
		t.Fatalf("help changed config: %v", err)
	}
}
