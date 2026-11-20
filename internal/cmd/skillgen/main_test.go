package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedReferencesAreCurrent(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	commands, err := collect(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(commands) < 200 {
		t.Fatalf("command discovery unexpectedly small: %d", len(commands))
	}
	code, docs, err := render(commands)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, "commands/schema_generated.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(code, got) {
		t.Fatal("schema registry drift; run make -C sdk generate-cli-skill")
	}
	for name, want := range docs {
		got, err := os.ReadFile(filepath.Join(root, "skill/moi-cli/references/commands", name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("reference drift: %s", name)
		}
	}
	if !bytes.Contains(docs["connector.md"], []byte("--file-path")) {
		t.Fatal("shared upload flags not discovered")
	}
	for _, name := range []string{"sql execute", "sql run"} {
		_, section, found := strings.Cut(string(docs["sql.md"]), "## "+name+"\n")
		section, _, _ = strings.Cut(section, "\n## ")
		if !found || !strings.Contains(section, "--query-file") || !strings.Contains(section, "--db") {
			t.Fatalf("shared SQL flags not discovered for %s", name)
		}
	}
}

func TestQuoteReferenceMatchesSupportedKnowledgeOperations(t *testing.T) {
	commands, err := collect(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"semantic-model read-evidence": true, "semantic-model cite": true, "semantic-model query": false, "semantic-model search": false}
	for _, command := range commands {
		expected, relevant := want[command.Name]
		if !relevant {
			continue
		}
		found := false
		for _, option := range command.Flags {
			found = found || option.Name == "quote"
		}
		if found != expected {
			t.Errorf("%s advertises quote=%t, supported=%t", command.Name, found, expected)
		}
		delete(want, command.Name)
	}
	if len(want) != 0 {
		t.Fatalf("missing command contracts: %v", want)
	}
}

func TestSavedResultReferenceMatchesKnowledgeOperations(t *testing.T) {
	commands, err := collect(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range commands {
		var accepted, rejected []string
		switch command.Name {
		case "semantic-model search", "semantic-model read-source", "semantic-model list-passages", "semantic-model read-document", "semantic-model search-document":
			accepted, rejected = []string{"from", "index"}, []string{"indices", "quote", "local-read"}
		case "semantic-model query", "semantic-model find-files":
			rejected = []string{"from", "index", "indices", "quote", "local-read"}
		case "semantic-model cite":
			accepted, rejected = []string{"from", "indices", "quote"}, []string{"index", "local-read"}
		case "semantic-model read-evidence":
			accepted = []string{"from", "index", "indices", "quote", "local-read"}
		default:
			continue
		}
		flags := map[string]bool{}
		for _, option := range command.Flags {
			flags[option.Name] = true
		}
		for _, flag := range accepted {
			if !flags[flag] {
				t.Errorf("%s omitted supported --%s", command.Name, flag)
			}
		}
		for _, flag := range rejected {
			if flags[flag] {
				t.Errorf("%s advertises unsupported --%s", command.Name, flag)
			}
		}
	}
}

func TestScannerTypedInputsAndFailure(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		failure      bool
	}{
		{"typed", `func run(){ fs:=flag.NewFlagSet("test create",flag.ExitOnError); body:=fs.String("json","","input"); var req sdk.SomeSpec; json.Unmarshal(rawBody,&req); _=body }`, false},
		{"typed UseNumber", `func run(){ fs:=flag.NewFlagSet("test create",flag.ExitOnError); body:=fs.String("json","","input"); var req sdk.SomeSpec; unmarshalJSONUseNumber(rawBody,&req); _=body }`, false},
		{"unresolved", `func run(){ fs:=flag.NewFlagSet("test create",flag.ExitOnError); json.Unmarshal(rawBody,&missing); _=fs }`, true},
		{"changed input pattern", `func run(){ fs:=flag.NewFlagSet("test create",flag.ExitOnError); var req sdk.SomeSpec; json.Unmarshal(renamedRaw,&req); _=fs }`, true},
		{"changed target pattern", `func run(){ fs:=flag.NewFlagSet("test create",flag.ExitOnError); json.Unmarshal(rawBody,&object.field); _=fs }`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", "package commands\n"+tc.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			c, err := scanFunction(file.Decls[0].(*ast.FuncDecl), nil)
			if (err != nil) != tc.failure {
				t.Fatalf("error=%v", err)
			}
			if !tc.failure && (c.Inputs["json"] != "sdk.SomeSpec" || c.Name != "test create") {
				t.Errorf("discovery=%+v", c)
			}
		})
	}
}

func TestCommandVariantsRawInputAndJSONSchemaGuard(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		failure      bool
	}{
		{
			name: "raw shared command",
			source: `func runAgentPublicCallback(){
				fs:=flag.NewFlagSet("agent "+command,flag.ContinueOnError)
				bodyFile:=fs.String("body-file","","Raw payload")
				_=bodyFile
				switch command { case "post-channel-callback": _=fs }
			}`,
		},
		{
			name: "JSON flag without decoder",
			source: `func runAgentShared(){
				fs:=flag.NewFlagSet("agent "+command,flag.ContinueOnError)
				body:=fs.String("json","","JSON input")
				_=body
				switch command { case "create": _=fs }
			}`,
			failure: true,
		},
		{
			name: "JSON file flag without decoder",
			source: `func runAgentShared(){
				fs:=flag.NewFlagSet("agent "+command,flag.ContinueOnError)
				bodyFile:=fs.String("json-file","","JSON input file")
				_=bodyFile
				switch command { case "create": _=fs }
			}`,
			failure: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "fixture.go", "package commands\n"+tc.source, 0)
			if err != nil {
				t.Fatal(err)
			}
			commands, err := commandVariants(file.Decls[0].(*ast.FuncDecl), nil)
			if (err != nil) != tc.failure {
				t.Fatalf("error=%v", err)
			}
			if tc.failure {
				return
			}
			if len(commands) != 1 || commands[0].Name != "agent post-channel-callback" {
				t.Fatalf("commands=%+v", commands)
			}
			if len(commands[0].Inputs) != 0 {
				t.Fatalf("raw command acquired JSON schema: %+v", commands[0].Inputs)
			}
		})
	}
}

func TestGeneratedCommandsDoNotInventSchemaForRawUploads(t *testing.T) {
	commands, err := collect(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range commands {
		if c.Name == "catalog upload" || c.Name == "semantic-model upload-local-file" {
			if len(c.Inputs) != 0 {
				t.Fatal("raw bytes described as JSON")
			}
		}
	}
	_, docs, err := render(commands)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range docs {
		if !strings.Contains(string(data), "Generated from CLI source") {
			t.Errorf("missing generation marker %s", name)
		}
	}
}

func TestGeneratorDetectsChangesAndObsoleteReferences(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, "commands")
	if err := os.Mkdir(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(sourceDir, "command.go")
	write := func(data string) {
		t.Helper()
		if err := os.WriteFile(source, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`package commands; func run(){ fs:=flag.NewFlagSet("old create",flag.ExitOnError); fs.String("name","","Name") }`)
	if err := generateAt(root, false); err != nil {
		t.Fatal(err)
	}
	if err := generateAt(root, true); err != nil {
		t.Fatal(err)
	}
	write(`package commands; func run(){ fs:=flag.NewFlagSet("new create",flag.ExitOnError); fs.String("renamed","","Name") }`)
	if err := generateAt(root, true); err == nil {
		t.Fatal("changed command reference was not detected")
	}
	if err := generateAt(root, false); err != nil {
		t.Fatal(err)
	}
	obsolete := filepath.Join(root, "skill/moi-cli/references/commands/old.md")
	if _, err := os.Stat(obsolete); !os.IsNotExist(err) {
		t.Fatal("obsolete generated document still exists")
	}
	if err := generateAt(root, true); err != nil {
		t.Fatal(err)
	}
}
