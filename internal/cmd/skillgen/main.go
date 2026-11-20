// skillgen derives CLI reference data from the actual FlagSets and JSON decode
// targets. Run from sdk/cli; no executable, account or service is consulted.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/matrixorigin/matrixflow/sdk/cli/commands"
)

type option struct{ Name, Help string }
type command struct {
	Name   string
	Flags  []option
	Inputs map[string]string
}

func main() {
	check := flag.Bool("check", false, "Check generated files without modifying them")
	flag.Parse()
	if err := generate(*check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func literal(e ast.Expr) string {
	if value, ok := e.(*ast.BasicLit); ok && value.Kind == token.STRING {
		s, _ := strconv.Unquote(value.Value)
		return s
	}
	return ""
}

func selector(call *ast.CallExpr, receiver, name string) bool {
	s, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || s.Sel.Name != name {
		return false
	}
	i, ok := s.X.(*ast.Ident)
	return ok && i.Name == receiver
}

func customJSONDecoder(call *ast.CallExpr) bool {
	decoder, ok := call.Fun.(*ast.Ident)
	return ok && (decoder.Name == "unmarshalProductJSON" || decoder.Name == "unmarshalJSONUseNumber")
}

func flagOption(call *ast.CallExpr) (option, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return option{}, false
	}
	receiver, ok := sel.X.(*ast.Ident)
	if !ok || receiver.Name != "fs" {
		return option{}, false
	}
	index := 0
	if sel.Sel.Name == "Var" && len(call.Args) == 3 {
		name := literal(call.Args[1])
		return option{name, literal(call.Args[2])}, name != ""
	}
	if strings.HasSuffix(sel.Sel.Name, "Var") {
		index = 1
	}
	if sel.Sel.Name == "Func" && len(call.Args) == 3 {
		name := literal(call.Args[0])
		return option{name, literal(call.Args[1])}, name != ""
	}
	if len(call.Args) < index+3 {
		return option{}, false
	}
	name := literal(call.Args[index])
	return option{name, literal(call.Args[index+2])}, name != ""
}

func commandFlags(node, skip ast.Node) []option {
	var flags []option
	ast.Inspect(node, func(node ast.Node) bool {
		if node == skip {
			return false
		}
		if call, ok := node.(*ast.CallExpr); ok {
			if value, found := flagOption(call); found {
				flags = append(flags, value)
			}
		}
		return true
	})
	return flags
}

// helperJSONFileInputs follows the decoder's actual parameter types. Commands
// may share file-option parsing without losing the schemas exposed to users.
func helperJSONFileInputs(fn *ast.FuncDecl) map[int]string {
	parameters := map[string]int{}
	position := 0
	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			parameters[name.Name] = position
			position++
		}
	}
	inputs := map[int]string{}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		typed, ok := call.Fun.(*ast.IndexExpr)
		if !ok {
			return true
		}
		decoder, ok := typed.X.(*ast.Ident)
		if !ok || decoder.Name != "loadJSONFile" {
			return true
		}
		argument, ok := call.Args[0].(*ast.Ident)
		if !ok {
			return true
		}
		index, found := parameters[argument.Name]
		if !found {
			return true
		}
		var value bytes.Buffer
		_ = format.Node(&value, token.NewFileSet(), typed.Index)
		inputs[index] = value.String()
		return true
	})
	return inputs
}

func scanFunction(fn *ast.FuncDecl, helperInputs map[string]map[int]string) (command, error) {
	c := command{Inputs: map[string]string{}}
	types := map[string]string{}
	flagVars := map[string]string{}
	jsonSources := map[string]string{}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.ValueSpec:
			if n.Type != nil {
				for _, name := range n.Names {
					var b bytes.Buffer
					_ = format.Node(&b, token.NewFileSet(), n.Type)
					types[name.Name] = b.String()
				}
			}
		case *ast.AssignStmt:
			if len(n.Lhs) > 0 && len(n.Rhs) == 1 {
				if id, ok := n.Lhs[0].(*ast.Ident); ok {
					if call, ok := n.Rhs[0].(*ast.CallExpr); ok && len(call.Args) == 2 {
						if fn, ok := call.Fun.(*ast.Ident); ok && fn.Name == "loadJSONBytes" {
							if inline, ok := call.Args[0].(*ast.BasicLit); ok && inline.Kind == token.STRING && literal(inline) == "" {
								if file, ok := call.Args[1].(*ast.StarExpr); ok {
									if variable, ok := file.X.(*ast.Ident); ok {
										jsonSources[id.Name] = flagVars[variable.Name]
									}
								}
							}
						}
					}
				}
			}
			if len(n.Lhs) == 1 && len(n.Rhs) == 1 {
				id, ok := n.Lhs[0].(*ast.Ident)
				if !ok {
					break
				}
				if call, ok := n.Rhs[0].(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						if recv, ok := sel.X.(*ast.Ident); ok && recv.Name == "fs" && len(call.Args) > 0 {
							flagVars[id.Name] = literal(call.Args[0])
						}
					}
				}
				if value, ok := n.Rhs[0].(*ast.CompositeLit); ok {
					var b bytes.Buffer
					_ = format.Node(&b, token.NewFileSet(), value.Type)
					types[id.Name] = b.String()
				}
			}
		case *ast.CallExpr:
			if callee, ok := n.Fun.(*ast.Ident); ok {
				for index, typ := range helperInputs[callee.Name] {
					if index >= len(n.Args) {
						continue
					}
					pointer, ok := n.Args[index].(*ast.StarExpr)
					if !ok {
						continue
					}
					variable, ok := pointer.X.(*ast.Ident)
					if !ok {
						continue
					}
					if argument := flagVars[variable.Name]; argument != "" {
						c.Inputs[argument] = typ
					}
				}
			}
			if typed, ok := n.Fun.(*ast.IndexExpr); ok {
				if fn, ok := typed.X.(*ast.Ident); ok && fn.Name == "loadJSONFile" {
					if pointer, ok := n.Args[0].(*ast.StarExpr); ok {
						if variable, ok := pointer.X.(*ast.Ident); ok && flagVars[variable.Name] != "" {
							var b bytes.Buffer
							_ = format.Node(&b, token.NewFileSet(), typed.Index)
							c.Inputs[flagVars[variable.Name]] = b.String()
						}
					}
				}
			}
			if selector(n, "flag", "NewFlagSet") && len(n.Args) > 0 {
				c.Name = literal(n.Args[0])
			}
			if value, found := flagOption(n); found {
				c.Flags = append(c.Flags, value)
			}
		}
		return true
	})
	var scanErr error
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			return true
		}
		if !selector(call, "json", "Unmarshal") && !customJSONDecoder(call) {
			return true
		}
		address, ok := call.Args[1].(*ast.UnaryExpr)
		if !ok {
			if strings.Contains(c.Name, " ") {
				scanErr = fmt.Errorf("%s: unsupported JSON decode target; update skillgen", fn.Name)
			}
			return true
		}
		id, ok := address.X.(*ast.Ident)
		if !ok {
			if strings.Contains(c.Name, " ") {
				scanErr = fmt.Errorf("%s: unsupported JSON decode target; update skillgen", fn.Name)
			}
			return true
		}
		argument := ""
		if raw, ok := call.Args[0].(*ast.Ident); ok {
			argument = jsonSources[raw.Name]
			if argument == "" && raw.Name == "rawBody" {
				argument = "json"
			}
		}
		if conv, ok := call.Args[0].(*ast.CallExpr); ok && len(conv.Args) == 1 {
			if star, ok := conv.Args[0].(*ast.StarExpr); ok {
				if id, ok := star.X.(*ast.Ident); ok {
					argument = flagVars[id.Name]
				}
			}
		}
		if argument == "" && strings.Contains(c.Name, " ") {
			scanErr = fmt.Errorf("%s: unsupported JSON argument source; update skillgen", fn.Name)
		} else if argument != "" {
			typ, ok := types[id.Name]
			if !ok {
				scanErr = fmt.Errorf("%s: cannot resolve JSON type of %s", fn.Name, id.Name)
			} else {
				c.Inputs[argument] = typ
			}
		}
		return true
	})
	return c, scanErr
}

// commandVariants derives shared FlagSet families from their typed dispatch.
// The CLI's operation switch remains the owner of both flags and input types.
func commandVariants(fn *ast.FuncDecl, sharedCommands map[string]command) ([]command, error) {
	var prefix, selectorName string
	var decodeTarget *ast.Object
	decoders := map[*ast.Object]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if assignment, ok := n.(*ast.AssignStmt); ok && len(assignment.Lhs) == 1 && len(assignment.Rhs) == 1 {
			if name, ok := assignment.Lhs[0].(*ast.Ident); ok {
				if closure, ok := assignment.Rhs[0].(*ast.FuncLit); ok {
					parameters := map[*ast.Object]bool{}
					for _, field := range closure.Type.Params.List {
						for _, parameter := range field.Names {
							parameters[parameter.Obj] = true
						}
					}
					ast.Inspect(closure.Body, func(n ast.Node) bool {
						if call, ok := n.(*ast.CallExpr); ok && (selector(call, "json", "Unmarshal") || customJSONDecoder(call)) && len(call.Args) == 2 {
							if target, ok := call.Args[1].(*ast.Ident); ok && parameters[target.Obj] {
								decoders[name.Obj] = true
							}
						}
						return true
					})
				}
			}
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if method, ok := call.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == "Decode" && len(call.Args) == 1 {
			if target, ok := call.Args[0].(*ast.Ident); ok {
				decodeTarget = target.Obj
			}
		}
		if method, ok := call.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == "Unmarshal" && len(call.Args) == 2 {
			if target, ok := call.Args[1].(*ast.Ident); ok {
				decodeTarget = target.Obj
			}
		}
		if customJSONDecoder(call) && len(call.Args) == 2 {
			if target, ok := call.Args[1].(*ast.Ident); ok {
				decodeTarget = target.Obj
			}
		}
		if !selector(call, "flag", "NewFlagSet") || len(call.Args) == 0 {
			return true
		}
		expression, ok := call.Args[0].(*ast.BinaryExpr)
		if !ok {
			return true
		}
		name, ok := expression.Y.(*ast.Ident)
		if !ok {
			return true
		}
		prefix, selectorName = literal(expression.X), name.Name
		return true
	})
	if prefix == "" {
		return nil, nil
	}
	if decodeTarget == nil && len(decoders) == 0 {
		for _, flag := range commandFlags(fn.Body, nil) {
			if flag.Name == "json" || flag.Name == "json-file" {
				return nil, fmt.Errorf("%s: cannot resolve shared JSON decode target", fn.Name)
			}
		}
	}
	var result []command
	var scanErr error
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		dispatch, ok := n.(*ast.SwitchStmt)
		if !ok {
			return true
		}
		tag, ok := dispatch.Tag.(*ast.Ident)
		if !ok || tag.Name != selectorName {
			return true
		}
		commonFlags := commandFlags(fn.Body, dispatch)
		for _, statement := range dispatch.Body.List {
			branch := statement.(*ast.CaseClause)
			var inputType string
			var decoded bool
			types := map[*ast.Object]string{}
			ast.Inspect(branch, func(n ast.Node) bool {
				if declaration, ok := n.(*ast.ValueSpec); ok && declaration.Type != nil {
					var typ bytes.Buffer
					_ = format.Node(&typ, token.NewFileSet(), declaration.Type)
					for _, name := range declaration.Names {
						types[name.Obj] = typ.String()
					}
				}
				if call, ok := n.(*ast.CallExpr); ok && len(call.Args) == 1 {
					if name, ok := call.Fun.(*ast.Ident); ok && decoders[name.Obj] {
						decoded = true
						if address, ok := call.Args[0].(*ast.UnaryExpr); ok && address.Op == token.AND {
							if target, ok := address.X.(*ast.Ident); ok {
								inputType = types[target.Obj]
							}
						}
					}
				}
				assignment, ok := n.(*ast.AssignStmt)
				if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
					return true
				}
				target, ok := assignment.Lhs[0].(*ast.Ident)
				if !ok {
					return true
				}
				expression := assignment.Rhs[0]
				if address, ok := expression.(*ast.UnaryExpr); ok && address.Op == token.AND {
					expression = address.X
				}
				switch value := expression.(type) {
				case *ast.CompositeLit:
					var typ bytes.Buffer
					_ = format.Node(&typ, token.NewFileSet(), value.Type)
					types[target.Obj] = typ.String()
				case *ast.Ident:
					types[target.Obj] = types[value.Obj]
				}
				if decodeTarget != nil && target.Obj == decodeTarget {
					decoded = true
					inputType = types[target.Obj]
				}
				return true
			})
			if len(branch.List) == 0 {
				continue
			}
			if decoded && inputType == "" {
				scanErr = fmt.Errorf("%s: cannot resolve shared JSON input type for %s", fn.Name, literal(branch.List[0]))
				return false
			}
			for _, name := range branch.List {
				if op := literal(name); op != "" {
					flags := append(slices.Clone(commonFlags), commandFlags(branch, nil)...)
					inputs := map[string]string{}
					if decoded {
						inputs["json"] = inputType
					}
					variant := command{Name: prefix + op, Flags: flags, Inputs: inputs}
					variant = withSharedOptions(variant, fn.Body, dispatch, sharedCommands)
					variant = withSharedOptions(variant, branch, nil, sharedCommands)
					if fn.Name.Name == "runKnowledgeRuntime" {
						variant.Flags = slices.DeleteFunc(variant.Flags, func(flag option) bool {
							return !commands.KnowledgeRuntimeFlagSupported(op, flag.Name)
						})
					}
					result = append(result, variant)
				}
			}
		}
		return false
	})
	return result, scanErr
}

func withSharedOptions(c command, node, skip ast.Node, sharedCommands map[string]command) command {
	ast.Inspect(node, func(n ast.Node) bool {
		if n == skip {
			return false
		}
		if call, ok := n.(*ast.CallExpr); ok {
			if name, ok := call.Fun.(*ast.Ident); ok {
				shared := sharedCommands[name.Name]
				c.Flags = append(c.Flags, shared.Flags...)
				for argument, typ := range shared.Inputs {
					c.Inputs[argument] = typ
				}
			}
		}
		return true
	})
	return c
}

func collect(root string) ([]command, error) {
	paths, err := filepath.Glob(filepath.Join(root, "commands", "*.go"))
	if err != nil {
		return nil, err
	}
	var funcs []*ast.FuncDecl
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") || strings.HasSuffix(path, "_generated.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
				funcs = append(funcs, fn)
			}
		}
	}
	helperInputs := map[string]map[int]string{}
	for _, fn := range funcs {
		if inputs := helperJSONFileInputs(fn); len(inputs) > 0 {
			helperInputs[fn.Name.Name] = inputs
		}
	}
	sharedCommands := map[string]command{"addKnowledgeQuoteFlag": {}, "addUploadFlags": {}, "runSQLStatement": {}, "runWorkspaceChatCompletion": {}, "runWorkflowDynamicServiceCommand": {}, "addWorkflowExecutionInputFlags": {}, "addComputeSpecFlags": {}, "addComputeResourceUpdateFlags": {}, "addSQLQueryFlags": {}, "addSQLHistoryRangeFlags": {}, "addSQLHistoryFilterFlags": {}, "addWorkflowTemplateFlags": {}, "loadWorkflowMOWLGraph": {}, "addResourceTagPageFlags": {}, "addProviderBackendCreateFlags": {}, "addProviderBackendUpdateFlags": {}, "addProviderBackendProbeFlags": {}, "addProviderRouterFlags": {}}
	for _, fn := range funcs {
		if _, shared := sharedCommands[fn.Name.Name]; shared {
			c, err := scanFunction(fn, helperInputs)
			if err != nil {
				return nil, err
			}
			sharedCommands[fn.Name.Name] = c
		}
	}
	var commands []command
	seen := map[string]bool{}
	for _, fn := range funcs {
		c, err := scanFunction(fn, helperInputs)
		if err != nil {
			return nil, err
		}
		if c.Name == "" {
			variants, err := commandVariants(fn, sharedCommands)
			if err != nil {
				return nil, err
			}
			commands = append(commands, variants...)
			continue
		}
		c = withSharedOptions(c, fn.Body, nil, sharedCommands)
		if seen[c.Name] {
			return nil, fmt.Errorf("duplicate FlagSet %q", c.Name)
		}
		seen[c.Name] = true
		slices.SortFunc(c.Flags, func(a, b option) int { return strings.Compare(a.Name, b.Name) })
		commands = append(commands, c)
	}
	slices.SortFunc(commands, func(a, b command) int { return strings.Compare(a.Name, b.Name) })
	return commands, nil
}

func render(commands []command) ([]byte, map[string][]byte, error) {
	var code strings.Builder
	code.WriteString("// Code generated by go run ./internal/cmd/skillgen; DO NOT EDIT.\npackage commands\nimport (\"reflect\"; sdk \"github.com/matrixorigin/moi-go-sdk/v5\"")
	needsJSON := false
	for _, command := range commands {
		for _, typ := range command.Inputs {
			needsJSON = needsJSON || strings.Contains(typ, "json.")
		}
	}
	if needsJSON {
		code.WriteString("; \"encoding/json\"")
	}
	code.WriteString(")\nvar commandJSONTypes = map[string]map[string]reflect.Type{\n")
	domains := map[string]*strings.Builder{}
	for _, c := range commands {
		if len(c.Inputs) > 0 {
			fmt.Fprintf(&code, "%q: {", c.Name)
			keys := make([]string, 0, len(c.Inputs))
			for k := range c.Inputs {
				keys = append(keys, k)
			}
			slices.Sort(keys)
			for _, k := range keys {
				fmt.Fprintf(&code, "%q: reflect.TypeFor[%s](),", k, c.Inputs[k])
			}
			code.WriteString("},\n")
		}
		domain, _, _ := strings.Cut(c.Name, " ")
		if domain == "schema" || domain == "skill" {
			continue
		}
		if domains[domain] == nil {
			domains[domain] = &strings.Builder{}
			fmt.Fprintf(domains[domain], "# %s command reference\n\nGenerated from CLI source. Do not edit. Business requirements remain server-owned.\nUse `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.\n\n", domain)
		}
		b := domains[domain]
		fmt.Fprintf(b, "## %s\n\n", c.Name)
		for _, f := range c.Flags {
			fmt.Fprintf(b, "- `--%s`: %s\n", f.Name, f.Help)
		}
		b.WriteString("\n")
	}
	code.WriteString("}\n")
	formatted, err := format.Source([]byte(code.String()))
	if err != nil {
		return nil, nil, err
	}
	docs := map[string][]byte{}
	for domain, b := range domains {
		docs[domain+".md"] = []byte(b.String())
	}
	return formatted, docs, nil
}

func generate(check bool) error {
	args := []string{"-C", "../../moi", "run", "./cmd/cli-input-shapes"}
	if check {
		args = append(args, "-check")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("generate server-decoded input shapes: %w", err)
	}
	// Frontend-owned assets are produced with frontend dependencies installed.
	// Their drift check runs in the renderer's tests, so CLI-only release checks
	// do not acquire a Node/frontend dependency.
	if !check {
		frontend := exec.Command("node", "../../moi-frontend/modules/moi-data-dashboard/scripts/export-cli-input-shapes.mjs")
		frontend.Stdout, frontend.Stderr = os.Stdout, os.Stderr
		if err := frontend.Run(); err != nil {
			return fmt.Errorf("generate frontend config input shapes: %w", err)
		}
	}
	return generateAt(".", check)
}

func generateAt(root string, check bool) error {
	commands, err := collect(root)
	if err != nil {
		return err
	}
	code, docs, err := render(commands)
	if err != nil {
		return err
	}
	files := map[string][]byte{filepath.Join(root, "commands", "schema_generated.go"): code}
	docsDir := filepath.Join(root, "skill", "moi-cli", "references", "commands")
	for name, data := range docs {
		files[filepath.Join(docsDir, name)] = data
	}
	previous, err := filepath.Glob(filepath.Join(docsDir, "*.md"))
	if err != nil {
		return err
	}
	for _, path := range previous {
		if _, ok := files[path]; ok {
			continue
		}
		if check {
			return fmt.Errorf("obsolete generated reference: %s; run make -C sdk generate-cli-skill", path)
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.Contains(old, []byte("Generated from CLI source. Do not edit.")) {
			return fmt.Errorf("unowned reference in generated directory: %s", path)
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	for path, data := range files {
		old, err := os.ReadFile(path)
		if err == nil && bytes.Equal(old, data) {
			continue
		}
		if check {
			return fmt.Errorf("generated file drift: %s; run make -C sdk generate-cli-skill", path)
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}
