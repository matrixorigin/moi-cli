// Package main is the entry point for moi-cli (product SDK CLI).
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/matrixorigin/matrixflow/sdk/cli/commands"
	"github.com/matrixorigin/matrixflow/sdk/cli/internal/nativeauth"
	"github.com/matrixorigin/matrixflow/sdk/cli/internal/trace"
	"github.com/matrixorigin/matrixflow/sdk/cli/internal/update"
	"github.com/matrixorigin/matrixflow/sdk/cli/skill"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

const (
	appName         = "moi-cli"
	defaultEndpoint = "http://127.0.0.1:8050/newmoi"
)

// version is set by the release build with -ldflags=-X=main.version=VERSION.
// Local builds retain their existing default version.
var version = "0.1.0"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--moi-update-protocol" {
		fmt.Println(update.Protocol)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "update" {
		if err := update.Run(context.Background(), os.Args[2:], os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, update.PublicError(err))
			os.Exit(1)
		}
		return
	}
	endpoint := flag.String("endpoint", defaultEndpoint, "Product API endpoint (include /newmoi base path)")
	apiKey := flag.String("api-key", "", "UC personal access token (PAT)")
	authMode := flag.String("auth", "auto", "Authentication: auto, moi (Astra login), or pat")
	workspaceID := flag.String("workspace-id", "", "Workspace ID for this command")
	requestTimeout := flag.Duration("request-timeout", 0, "Explicit positive HTTP request deadline, including response body (e.g. 2m); omitted uses command default")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	outputJSON := flag.Bool("o", false, "Output as JSON")
	noInteractive := flag.Bool("no-interactive", false, "Disable interactive prompts (e.g. pagination)")
	showVersion := flag.Bool("version", false, "Show version")
	showHelp := flag.Bool("help", false, "Show help")

	flag.Usage = func() {
		printGlobalUsage()
	}

	cmdName, cmdArgs, showCmdHelp := checkCommandHelp()
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s version %s\n", appName, version)
		os.Exit(0)
	}
	if showCmdHelp && cmdName != "" {
		if err := showCommandHelp(cmdName, cmdArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no command specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	command := args[0]
	commandArgs := args[1:]
	timeout := commands.RequestTimeout(command, commandArgs)
	var timeoutProvided bool
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "request-timeout" {
			timeoutProvided = true
		}
	})
	if timeoutProvided {
		if *requestTimeout <= 0 {
			writeCommandError(os.Stderr, &sdk.Error{Code: "CLI_REQUEST_TIMEOUT_INVALID", Message: "--request-timeout must be positive"}, *outputJSON)
			os.Exit(1)
		}
		timeout = *requestTimeout
	}
	// Offline commands must run before updater initialization. Restricted roots
	// such as the sandbox image may not expose /proc/self/exe.
	if command == "agent" && len(commandArgs) > 0 && slices.Contains([]string{"init", "build", "verify", "unpack"}, commandArgs[0]) {
		ctx := commands.NewContext(nil, "", *verbose, *outputJSON, *noInteractive)
		ctx.RunningMOIVersion = version
		if err := commands.ExecuteAgentCommand(ctx, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if command == "schema" || command == "skill" {
		var err error
		if command == "schema" {
			err = commands.ExecuteSchemaCommand(commandArgs)
		} else {
			err = executeSkillCommand(commandArgs)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if command == "config" {
		if err := commands.ExecuteConfigCommand(nil, commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}
	// Help and version output above must work in restricted roots such as the
	// sandbox image, where /proc/self/exe is intentionally unavailable. Online
	// commands still initialize the managed updater before contacting the API.
	// Runtime child commands run inside that same sandbox, so their task-bound
	// credentials are resolved before os.Executable and the managed updater,
	// and they skip the updater entirely.
	cfg, runtimeProof, err := runtimeChildConfiguration()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	if runtimeProof == "" {
		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, "CLI_UPDATE_EXECUTABLE_UNAVAILABLE")
			os.Exit(1)
		}
		lease, err := update.Acquire(exe)
		if err != nil {
			fmt.Fprintln(os.Stderr, update.PublicError(err))
			os.Exit(1)
		}
		if lease != nil {
			defer lease.Close()
		}
		update.Startup(exe, os.Args[1:])

		cfg, err = commands.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}
	}
	native := false
	if runtimeProof == "" {
		native, err = nativeauth.Selected(*authMode)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if native {
		explicitEndpoint := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "endpoint" {
				explicitEndpoint = true
			}
		})
		if err := executeNativeCommand(cfg, *endpoint, explicitEndpoint, *apiKey, *workspaceID, *verbose, *outputJSON, *noInteractive, command, commandArgs, timeout); err != nil {
			writeCommandError(os.Stderr, err, *outputJSON)
			os.Exit(1)
		}
		return
	}

	if cfg != nil {
		if *endpoint == defaultEndpoint && cfg.Endpoint != "" {
			*endpoint = cfg.Endpoint
		}
		if *apiKey == "" && cfg.APIKey != "" {
			*apiKey = cfg.APIKey
		}
		if *workspaceID == "" && cfg.WorkspaceID != "" {
			*workspaceID = cfg.WorkspaceID
		}
	}

	clientOptions, err := workspaceRoleOptions(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if runtimeProof != "" {
		runtimeClient := runtimeHTTPClient()
		clientOptions = append(clientOptions, sdk.WithRuntimeAuthorization(runtimeProof), sdk.WithHTTPClient(runtimeClient))
		*noInteractive = true
	}

	if *endpoint == "" {
		fmt.Fprintf(os.Stderr, "Error: endpoint is required (use -endpoint or set in ~/.moi-cli.json)\n")
		os.Exit(1)
	}
	machineCredentialCommand := command == "runtime-gateway" || (command == "system" && len(commandArgs) > 0 && (commandArgs[0] == "service-account-workspaces" || commandArgs[0] == "service-account-assignable-roles"))
	if *apiKey == "" && !machineCredentialCommand {
		fmt.Fprintf(os.Stderr, "Error: api-key is required (use -api-key or set in ~/.moi-cli.json)\n")
		os.Exit(1)
	}

	clientOptions = append(clientOptions, sdk.WithTimeout(timeout))

	var client *sdk.Client
	if machineCredentialCommand {
		client, err = sdk.New(*endpoint, append(clientOptions, sdk.WithUnauthenticated(), sdk.WithAPIKey(*apiKey))...)
	} else {
		client, err = sdk.NewWithPersonalAccessToken(*endpoint, *apiKey, clientOptions...)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := commands.NewContext(client, *workspaceID, *verbose, *outputJSON, *noInteractive)
	ctx.Endpoint = *endpoint
	ctx.APIKey = *apiKey
	var stopTrace func()
	if runtimeProof != "" {
		stopTrace = trace.BeginRuntimeChild()
	}

	if err := executeCommand(ctx, command, commandArgs); err != nil {
		trace.MarkError()
		if stopTrace != nil {
			stopTrace()
		}
		writeCommandError(os.Stderr, err, *outputJSON)
		os.Exit(1)
	}
	if stopTrace != nil {
		stopTrace()
	}
}

func writeCommandError(out io.Writer, err error, outputJSON bool) {
	var productError *sdk.Error
	if outputJSON && errors.As(err, &productError) {
		_ = json.NewEncoder(out).Encode(productError)
		return
	}
	fmt.Fprintf(out, "Error: %v\n", err)
	if errors.As(err, &productError) && len(productError.Details) != 0 {
		_ = json.NewEncoder(out).Encode(productError.Details)
	}
}

// Runtime environment supplies connection defaults without reading saved login
// or Role configuration. Explicit -endpoint/-api-key flags that differ from the
// runtime bundle are rejected, so the task proof only reaches the bound
// endpoint. MOI
// validates the request identity, task proof and workspace authorization.
func runtimeChildConfiguration() (*commands.Config, string, error) {
	proof, present := os.LookupEnv("MOI_RUNTIME_AUTHORIZATION")
	if !present {
		return nil, "", nil
	}
	token, bearer := strings.CutPrefix(proof, "Bearer ")
	if !bearer || token == "" || strings.ContainsAny(token, " \t\r\n") {
		return nil, "", fmt.Errorf("CLI_RUNTIME_AUTHORIZATION_INVALID: MOI_RUNTIME_AUTHORIZATION must be a single \"Bearer <token>\" line with no leading, trailing or embedded whitespace")
	}
	cfg, _ := commands.RuntimeConfig()
	// An explicit flag that repeats the bound value is not a conflict; only a
	// different endpoint or key could send the task proof elsewhere.
	explicitConnectionFlag := ""
	flag.Visit(func(f *flag.Flag) {
		bound := ""
		switch f.Name {
		case "endpoint":
			bound = cfg.Endpoint
		case "api-key":
			bound = cfg.APIKey
		default:
			return
		}
		if f.Value.String() != bound && explicitConnectionFlag == "" {
			explicitConnectionFlag = "-" + f.Name
		}
	})
	if explicitConnectionFlag != "" {
		return nil, "", fmt.Errorf("CLI_RUNTIME_CONNECTION_CONFLICT: %s must not be combined with MOI_RUNTIME_AUTHORIZATION; the runtime binds MOI_ENDPOINT and MOI_API_KEY as one bundle, so remove the explicit flag or unset MOI_RUNTIME_AUTHORIZATION", explicitConnectionFlag)
	}
	if cfg.Endpoint == "" || cfg.APIKey == "" {
		return nil, "", fmt.Errorf("CLI_RUNTIME_CONNECTION_INCOMPLETE: MOI_RUNTIME_AUTHORIZATION is bound but the connection bundle is incomplete; runtime child mode carries MOI_ENDPOINT and MOI_API_KEY as a whole, so export both together or unset MOI_RUNTIME_AUTHORIZATION to fall back to saved login")
	}
	for _, value := range []string{cfg.Endpoint, cfg.APIKey, cfg.WorkspaceID} {
		if value != strings.TrimSpace(value) || strings.ContainsAny(value, "\r\n\t") {
			return nil, "", errRuntimeIdentityInvalid
		}
	}
	if strings.ContainsAny(cfg.APIKey, " ") || strings.ContainsAny(cfg.WorkspaceID, " /\\") {
		return nil, "", errRuntimeIdentityInvalid
	}
	return cfg, proof, nil
}

// Role selection belongs to moi-cli configuration, independently of the
// credential provider. Validate every mapping before any authentication or
// business request and freeze the complete workspace map for this command.
func workspaceRoleOptions(cfg *commands.Config) ([]sdk.Option, error) {
	roles, err := commands.WorkspaceCurrentRoleSnapshot(cfg)
	if err != nil {
		return nil, err
	}
	opts := make([]sdk.Option, 0, len(roles))
	for _, workspaceID := range slices.Sorted(maps.Keys(roles)) {
		opts = append(opts, sdk.WithWorkspaceCurrentRoleID(workspaceID, roles[workspaceID]))
	}
	return opts, nil
}

func executeNativeCommand(cfg *commands.Config, endpoint string, explicitEndpoint bool, apiKey, workspaceID string, verbose, outputJSON, noInteractive bool, command string, args []string, timeout time.Duration) error {
	if apiKey != "" {
		return fmt.Errorf("cannot mix Astra login with -api-key; use -auth pat explicitly for legacy credentials")
	}
	opts, err := workspaceRoleOptions(cfg)
	if err != nil {
		return err
	}
	helper, err := nativeauth.NewHelper()
	if err != nil {
		return err
	}
	credential, err := helper.Credential(context.Background(), "")
	if err != nil {
		return err
	}
	if explicitEndpoint && endpoint != credential.Endpoint {
		return fmt.Errorf("-endpoint differs from the signed-in MOI environment; sign in to that environment first")
	}
	// Shared login owns the endpoint, token and default workspace. Saved PAT
	// connection settings must not override or block this environment binding.
	if workspaceID == "" && credential.WorkspaceID != nil {
		workspaceID = *credential.WorkspaceID
	}
	opts = append(opts, sdk.WithTimeout(0), sdk.WithToken(credential.AccessToken), sdk.WithHTTPClient(&http.Client{
		Transport:     &nativeauth.Transport{Helper: helper, Binding: credential, RequestTimeout: timeout},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}))
	client, err := sdk.New(credential.Endpoint, opts...)
	if err != nil {
		return err
	}
	defer client.Close()
	ctx := commands.NewContext(client, workspaceID, verbose, outputJSON, noInteractive)
	ctx.Endpoint = credential.Endpoint
	return executeCommand(ctx, command, args)
}

func executeCommand(ctx *commands.Context, command string, args []string) error {
	switch command {
	case "config":
		return commands.ExecuteConfigCommand(ctx, args)
	case "workspace":
		return commands.ExecuteWorkspaceCommand(ctx, args)
	case "workspace-member":
		return commands.ExecuteWorkspaceMemberCommand(ctx, args)
	case "file":
		return commands.ExecuteFileCommand(ctx, args)
	case "catalog":
		return commands.ExecuteCatalogCommand(ctx, args)
	case "workflow":
		return commands.ExecuteWorkflowCommand(ctx, args)
	case "sql":
		return commands.ExecuteSQLCommand(ctx, args)
	case "agent":
		return commands.ExecuteAgentCommand(ctx, args)
	case "webapp":
		return commands.ExecuteWebappCommand(ctx, args)
	case "conversation":
		return commands.ExecuteConversationCommand(ctx, args)
	case "runtime-gateway":
		return commands.ExecuteRuntimeGatewayCommand(ctx, args)
	case "system":
		return commands.ExecuteSystemCommand(ctx, args)
	case "account":
		return commands.ExecuteAccountCommand(ctx, args)
	case "semantic-model":
		return commands.ExecuteSemanticModelCommand(ctx, args)
	case "iam":
		return commands.ExecuteIAMCommand(ctx, args)
	case "model":
		return commands.ExecuteModelCommand(ctx, args)
	case "provider-backend":
		return commands.ExecuteProviderBackendCommand(ctx, args)
	case "ai-studio":
		return commands.ExecuteAIStudioCommand(ctx, args)
	case "connector":
		return commands.ExecuteConnectorCommand(ctx, args)
	case "compute":
		return commands.ExecuteComputeCommand(ctx, args)
	case "dashboard":
		return commands.ExecuteDashboardCommand(ctx, args)
	case "resource-tag":
		return commands.ExecuteResourceTagCommand(ctx, args)
	case "runner":
		return commands.ExecuteRunnerCommand(ctx, args)
	case "sandbox":
		return commands.ExecuteSandboxCommand(ctx, args)
	case "data-share":
		return commands.ExecuteDataShareCommand(ctx, args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func showCommandHelp(command string, args []string) error {
	if len(args) > 0 && command != "schema" && command != "skill" {
		// Reuse the actual command's FlagSet. Pass only --help: Parse exits
		// successfully before context, credentials, files or business calls
		// are consulted. Never forward user-supplied operation arguments here.
		err := executeCommand(nil, command, []string{args[0], "--help"})
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	switch command {
	case "schema":
		fmt.Println("Usage: moi-cli schema [<domain> <command>] [--argument json] [--validate-file PATH]\nOffline wire-shape inspection and structural validation; no request is sent.")
	case "skill":
		fmt.Println("Usage: moi-cli skill search --query 'channel create' [--limit 8]\n       moi-cli skill read --path references/files.md\n       moi-cli skill export --dir DIRECTORY\nSearch paired capabilities and commands, read their procedures, or export them offline.")
	case "config":
		commands.ShowConfigHelp(args)
	case "workspace":
		commands.ShowWorkspaceHelp(args)
	case "workspace-member":
		commands.ShowWorkspaceMemberHelp(args)
	case "file":
		commands.ShowFileHelp(args)
	case "catalog":
		commands.ShowCatalogHelp(args)
	case "workflow":
		commands.ShowWorkflowHelp(args)
	case "sql":
		commands.ShowSQLHelp(args)
	case "agent":
		commands.ShowAgentHelp(args)
	case "webapp":
		commands.ShowWebappHelp(args)
	case "conversation":
		commands.ShowConversationHelp(args)
	case "runtime-gateway":
		commands.ShowRuntimeGatewayHelp(args)
	case "system":
		commands.ShowSystemHelp(args)
	case "account":
		commands.ShowAccountHelp(args)
	case "semantic-model":
		commands.ShowSemanticModelHelp(args)
	case "iam":
		commands.ShowIAMHelp(args)
	case "model":
		commands.ShowModelHelp(args)
	case "provider-backend":
		commands.ShowProviderBackendHelp(args)
	case "ai-studio":
		commands.ShowAIStudioHelp(args)
	case "connector":
		commands.ShowConnectorHelp(args)
	case "compute":
		commands.ShowComputeHelp(args)
	case "dashboard":
		commands.ShowDashboardHelp(args)
	case "resource-tag":
		commands.ShowResourceTagHelp(args)
	case "runner":
		commands.ShowRunnerHelp(args)
	case "sandbox":
		commands.ShowSandboxHelp(args)
	case "data-share":
		commands.ShowDataShareHelp(args)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
	return nil
}

func checkCommandHelp() (cmdName string, cmdArgs []string, showHelp bool) {
	skipGlobalOptionValue := false
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if skipGlobalOptionValue {
			skipGlobalOptionValue = false
			continue
		}
		if arg == "--help" || arg == "-h" {
			if cmdName != "" {
				return cmdName, cmdArgs, true
			}
			return "", nil, false
		}
		if cmdName == "" && globalOptionTakesValue(arg) {
			skipGlobalOptionValue = true
			continue
		}
		if !strings.HasPrefix(arg, "-") && cmdName == "" {
			cmdName = arg
			cmdArgs = []string{}
		} else if cmdName != "" && !strings.HasPrefix(arg, "-") {
			cmdArgs = append(cmdArgs, arg)
		}
	}
	return "", nil, false
}

func executeSkillCommand(args []string) error {
	if len(args) > 0 && args[0] == "search" {
		return executeSkillSearch(args[1:])
	}
	if len(args) > 0 && args[0] == "read" {
		return executeSkillRead(args[1:])
	}
	if len(args) == 0 || args[0] != "export" {
		return fmt.Errorf("CLI_SKILL_COMMAND_REQUIRED: use skill search --query KEYWORDS, skill read --path DOCUMENT or skill export --dir DIRECTORY")
	}
	fs := flag.NewFlagSet("skill export", flag.ContinueOnError)
	dir := fs.String("dir", "", "New directory for the paired MOI Skill")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("CLI_SKILL_UNEXPECTED_ARGUMENT: unexpected positional argument")
	}
	return skill.Export(*dir)
}

func executeSkillSearch(args []string) error {
	fs := flag.NewFlagSet("skill search", flag.ContinueOnError)
	query := fs.String("query", "", "Capability or operation keywords, for example channel create, connector import, workitem or mcp")
	limit := fs.Int("limit", 8, "Maximum matching documentation sections")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("CLI_SKILL_UNEXPECTED_ARGUMENT: use --query for operation keywords")
	}
	result, err := skill.Search(*query, *limit)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(result)
}

func executeSkillRead(args []string) error {
	fs := flag.NewFlagSet("skill read", flag.ContinueOnError)
	path := fs.String("path", "SKILL.md", "Relative path to a paired Skill Markdown document")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("CLI_SKILL_UNEXPECTED_ARGUMENT: unexpected positional argument")
	}
	data, err := skill.Read(*path)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}

func globalOptionTakesValue(arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}
	switch strings.TrimLeft(arg, "-") {
	case "endpoint", "api-key", "workspace-id", "auth", "request-timeout":
		return true
	default:
		return false
	}
}

func printGlobalUsage() {
	fmt.Fprintf(os.Stderr, `%s - MOI Product CLI (product SDK)

Uses MOI Product APIs and the same backend permissions as the product UI.

Usage:
  %s [global options] <command> [command options]

Global options:
  -endpoint string       Product API endpoint (default %q)
  -api-key string        UC personal access token (PAT)
  -auth string           auto (default), moi (Astra login), or pat
  -workspace-id string   Workspace ID for this command
  -request-timeout duration  Explicit positive HTTP deadline, including response body (e.g. 2m)
  -verbose               Enable verbose output
  -o                     Output as JSON
  -no-interactive        Disable interactive prompts
  -version               Show version
  -help                  Show this help

Commands:
  schema             Offline JSON input structure and validation
  skill              Search capabilities/commands; read or export the paired Skill (offline)
  update             Check/update the managed client bundle; --rollback restores the previous bundle
  config             Manage config (set-endpoint, set-api-key, set-workspace, set-role, unset-role, show)
  workspace          WorkspaceService (+ ensure/get aliases)
  workspace-member   WorkspaceMemberService
  catalog            Catalogs, databases, tables, Volumes, files and data placement
  workflow           Workflows, system WorkItems, custom operators and API/MCP publication
  sql                SQLService
  agent              Agents, Skills, channels, connections and plugins
  webapp             Anonymous public Agent Webapp protocol
  account            AccountService
  semantic-model     KnowledgeService
  iam                IAMService
  model              ModelService
  provider-backend   ProviderBackendService (--kind llm|embedding|parser)
  ai-studio          Account model policy and Genesis service-slot administration
  connector          External data connectors, connection validation, imports and exports
  compute            ComputeService
  dashboard          DashboardService
  resource-tag       Workspace tags, values, assignments and resource grouping
  runner             Custom runners and one-time connection commands
  sandbox            Sandboxes, network settings, quotas and terminal connections
  data-share         DataShareService

Examples:
  moi-cli skill search --query 'channel create'
  moi-cli skill search --query 'operator publish'
  moi-cli agent create-channel-instance --help
  %s config set-endpoint http://127.0.0.1:8050/newmoi
  %s config set-api-key <pat>
  %s config set-workspace <workspace-id>
  %s config set-role <workspace-id> <role-id>
  %s workspace list
  %s catalog list
  %s sql execute-query --query "SELECT 1"
  %s workflow list

For command-specific help: %s <command> --help

`, appName, appName, defaultEndpoint, appName, appName, appName, appName, appName, appName, appName, appName, appName)
}
