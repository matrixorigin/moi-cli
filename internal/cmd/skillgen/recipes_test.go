package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Recipe shell blocks deliberately contain individual CLI invocations rather
// than executable scripts. Tokenize quoted example values without evaluating
// variables, substitutions, file reads or any business operation.
var recipeWord = regexp.MustCompile(`(?:[^\s"'\\]|\\.|"(?:[^"\\]|\\.)*"|'[^']*')+`)
var recipeBlock = regexp.MustCompile("(?ms)^```(?:sh|bash)\n(.*?)^```")

func validateRecipe(line string, commands map[string]command, globals map[string]bool) error {
	words := recipeWord.FindAllString(line, -1)
	if len(words) == 0 || strings.HasPrefix(words[0], "#") {
		return nil
	}
	if words[0] != "moi-cli" {
		return fmt.Errorf("recipe must contain a standalone moi-cli invocation: %s", line)
	}
	i := 1
	for i < len(words) && strings.HasPrefix(words[i], "-") {
		name, _, hasValue := strings.Cut(strings.TrimLeft(words[i], "-"), "=")
		boolean, ok := globals[name]
		if !ok {
			return fmt.Errorf("unknown global option %q", name)
		}
		i++
		if !boolean && !hasValue {
			if i == len(words) {
				return fmt.Errorf("missing global option value for %q", name)
			}
			i++
		}
	}
	if i+1 >= len(words) {
		return fmt.Errorf("missing domain/subcommand in %s", line)
	}
	schema := words[i] == "schema"
	if schema {
		i++
		if i+1 >= len(words) {
			return fmt.Errorf("missing schema target")
		}
	}
	name := strings.Join(words[i:i+2], " ")
	c, ok := commands[name]
	if !ok {
		return fmt.Errorf("unknown command %q", name)
	}
	if schema {
		if len(c.Inputs) == 0 {
			return fmt.Errorf("%s has no typed JSON schema", name)
		}
		c = commands["schema"]
	}
	flags := map[string]bool{"help": true, "h": true}
	for _, f := range c.Flags {
		flags[f.Name] = true
	}
	for _, word := range words[i+2:] {
		if !strings.HasPrefix(word, "-") {
			continue
		}
		flag, _, _ := strings.Cut(strings.TrimLeft(word, "-"), "=")
		if !flags[flag] {
			return fmt.Errorf("%s: unknown option %q", name, flag)
		}
	}
	return nil
}

func TestSkillRecipeCommandsMatchCLI(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	collected, err := collect(root)
	if err != nil {
		t.Fatal(err)
	}
	commands := map[string]command{}
	for _, c := range collected {
		commands[c.Name] = c
	}
	// Global options are owned by main, not by individual command FlagSets.
	main, err := parser.ParseFile(token.NewFileSet(), filepath.Join(root, "cmd/moi-cli/main.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Offline Skill commands live in main. Collect their real FlagSets too,
	// rather than teaching recipes a separate hard-coded option list.
	for _, decl := range main.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
			c, err := scanFunction(fn, nil)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(c.Name, "skill ") {
				commands[c.Name] = c
			}
		}
	}
	globals := map[string]bool{}
	ast.Inspect(main, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if ok && len(call.Args) == 3 && (selector(call, "flag", "String") || selector(call, "flag", "Bool")) {
			globals[literal(call.Args[0])] = selector(call, "flag", "Bool")
		}
		return true
	})
	count := 0
	err = filepath.WalkDir(filepath.Join(root, "skill/moi-cli"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, block := range recipeBlock.FindAllStringSubmatch(string(data), -1) {
			// Parse only; never execute examples or expand the user's environment.
			check := exec.CommandContext(t.Context(), "sh", "-n")
			check.Stdin = strings.NewReader(block[1])
			if out, err := check.CombinedOutput(); err != nil {
				t.Errorf("%s: invalid shell syntax: %v: %s", path, err, out)
			}
			for line := range strings.SplitSeq(strings.ReplaceAll(block[1], "\\\n", " "), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				count++
				if err := validateRecipe(line, commands, globals); err != nil {
					t.Errorf("%s: %s: %v", path, line, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Fatal("no executable recipes checked")
	}
	t.Logf("checked %d recipe commands", count)
	for _, bad := range []string{
		"moi-cli -unknown sql execute",
		"moi-cli nonexistent list",
		"moi-cli sql missing",
		"moi-cli sql execute --missing value",
		"moi-cli schema sql execute",
		"moi-cli schema connector create-task --missing value",
	} {
		if err := validateRecipe(bad, commands, globals); err == nil {
			t.Errorf("undetected recipe drift: %s", bad)
		}
	}
}
