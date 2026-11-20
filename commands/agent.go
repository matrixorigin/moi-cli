package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteAgentCommand runs product SDK AgentService methods via moi-cli.
func ExecuteAgentCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowAgentHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "artifacts":
		return runAgentServiceArtifacts(ctx, args[1:])
	case "artifact":
		return runAgentServiceArtifact(ctx, args[1:])
	case "delete-artifacts":
		return runAgentServiceDeleteArtifacts(ctx, args[1:])
	case "resolve-artifacts":
		return runAgentServiceResolveArtifacts(ctx, args[1:])
	case "backfill-artifacts":
		return runAgentServiceBackfillArtifacts(ctx, args[1:])
	case "connect-browser":
		return runAgentBrowser(ctx, args[1:], sub)
	case "watch-browser":
		return runAgentBrowser(ctx, args[1:], sub)
	case "list-resource-meta":
		return runAgentResourceMetadata(ctx, args[1:])
	case "builder":
		return runAgentBuilder(ctx, args[1:])
	case "search-plugins":
		return runAgentServiceSearchPlugins(ctx, args[1:])
	case "describe-plugin-capability":
		return runAgentServiceDescribePluginCapability(ctx, args[1:])
	case "list-plugin-releases":
		return runAgentServiceListPluginReleases(ctx, args[1:])
	case "get-plugin-release":
		return runAgentServiceGetPluginRelease(ctx, args[1:])
	case "install-plugin":
		return runAgentServiceInstallPlugin(ctx, args[1:])
	case "list-plugin-installations":
		return runAgentServiceListPluginInstallations(ctx, args[1:])
	case "get-plugin-installation":
		return runAgentServiceGetPluginInstallation(ctx, args[1:])
	case "update-plugin-installation":
		return runAgentServiceUpdatePluginInstallation(ctx, args[1:])
	case "delete-plugin-installation":
		return runAgentServiceDeletePluginInstallation(ctx, args[1:])
	case "create-plugin-connection":
		return runAgentServiceCreatePluginConnection(ctx, args[1:])
	case "list-plugin-connections":
		return runAgentServiceListPluginConnections(ctx, args[1:])
	case "get-plugin-connection":
		return runAgentServiceGetPluginConnection(ctx, args[1:])
	case "update-plugin-connection":
		return runAgentServiceUpdatePluginConnection(ctx, args[1:])
	case "delete-plugin-connection":
		return runAgentServiceDeletePluginConnection(ctx, args[1:])
	case "get-plugin-authorization":
		return runAgentServiceGetPluginAuthorization(ctx, args[1:])
	case "store-plugin-secret":
		return runAgentServiceStorePluginSecret(ctx, args[1:])
	case "begin-plugin-oauth":
		return runAgentServiceBeginPluginOAuth(ctx, args[1:])
	case "poll-plugin-oauth":
		return runAgentServicePollPluginOAuth(ctx, args[1:])
	case "disconnect-plugin-connection":
		return runAgentServiceDisconnectPluginConnection(ctx, args[1:])
	case "list-agent-catalog-cards":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "create-team":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "list-teams":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "get-team":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "update-team":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "delete-team":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "list-team-members":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "add-team-member":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "remove-team-member":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "list-ingress-capabilities":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "resolve-we-com-qr-binding":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "list-ingress-connections":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "create-ingress-connection":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "delete-ingress-connection":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "get-ingress-connection":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "update-ingress-connection":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "verify-ingress-connection":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "generate-we-com-callback-secrets":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "verify-channel-callback", "post-channel-callback", "post-we-com-callback", "verify-ingress-callback", "post-ingress-callback":
		return runAgentPublicCallback(ctx, sub, args[1:])
	case "list-team-sessions":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "create-team-session":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "rename-team-session":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "submit-team-message":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "list-team-messages":
		return runAgentTeamCommand(ctx, sub, args[1:])
	case "workspace-mcp":
		return runAgentServiceWorkspaceMCP(ctx, args[1:])
	case "runtime-browser-token":
		return runAgentRuntimeBrowserToken(ctx, args[1:])
	case "connect-runtime-browser":
		return runAgentConnectRuntimeBrowser(ctx, args[1:])
	case "update-admin-feedback-ticket":
		return runAgentServiceUpdateAdminFeedbackTicket(ctx, args[1:])
	case "admin-feedback-tickets":
		return runAgentServiceAdminFeedbackTickets(ctx, args[1:])
	case "init":
		return runAgentInit(ctx, args[1:])
	case "build":
		return runAgentBuild(ctx, args[1:])
	case "verify":
		return runAgentVerify(ctx, args[1:])
	case "unpack":
		return runAgentUnpack(ctx, args[1:])
	case "list":
		return runAgentServiceList(ctx, args[1:])
	case "executor-health":
		return runAgentServiceExecutorHealth(ctx, args[1:])
	case "preview-query-visual":
		return runAgentServicePreviewQueryVisual(ctx, args[1:])
	case "admin-feedback-ticket":
		return runAgentServiceAdminFeedbackTicket(ctx, args[1:])
	case "build-feedback-github-issue-draft":
		return runAgentServiceBuildFeedbackGithubIssueDraft(ctx, args[1:])
	case "create-feedback-github-issue":
		return runAgentServiceCreateFeedbackGithubIssue(ctx, args[1:])
	case "approve-feedback-reward":
		return runAgentServiceApproveFeedbackReward(ctx, args[1:])
	case "my-feedback-tickets":
		return runAgentServiceMyFeedbackTickets(ctx, args[1:])
	case "create-feedback-ticket":
		return runAgentServiceCreateFeedbackTicket(ctx, args[1:])
	case "export-agent-package":
		return runAgentServiceExportAgentPackage(ctx, args[1:])
	case "send":
		return runAgentSendAlias(ctx, args[1:])
	case "get-task":
		return runAgentGetTaskAlias(ctx, args[1:])
	case "cancel-task":
		return runAgentCancelTaskAlias(ctx, args[1:])
	case "card":
		return runAgentCardAlias(ctx, args[1:])
	case "agent-card":
		return runAgentServiceAgentCard(ctx, args[1:])
	case "a2a":
		return runAgentServiceA2A(ctx, args[1:])
	case "a2a-stream":
		return runAgentServiceA2AStream(ctx, args[1:])
	case "load-agent-package":
		return runAgentServiceLoadAgentPackage(ctx, args[1:])
	case "upload-workspace-file":
		return runAgentServiceUploadWorkspaceFile(ctx, args[1:])
	case "inspect-skill-package":
		return runAgentServiceInspectSkillPackage(ctx, args[1:])
	case "import-skill-package":
		return runAgentServiceImportSkillPackage(ctx, args[1:])
	case "confirm-skill-candidate":
		return runAgentServiceConfirmSkillCandidate(ctx, args[1:])
	case "get-skill-candidate-confirmation":
		return runAgentServiceGetSkillCandidateConfirmation(ctx, args[1:])
	case "get-skill-file":
		return runAgentServiceGetSkillFile(ctx, args[1:])
	case "polish-skill-stream":
		return runAgentServicePolishSkillStream(ctx, args[1:])
	case "resource-get":
		return runAgentServiceResourceGet(ctx, args[1:])
	case "resource-post":
		return runAgentServiceResourcePost(ctx, args[1:])
	case "resource-put":
		return runAgentServiceResourcePut(ctx, args[1:])
	case "resource-patch":
		return runAgentServiceResourcePatch(ctx, args[1:])
	case "resource-delete":
		return runAgentServiceResourceDelete(ctx, args[1:])
	case "resource-any":
		return runAgentServiceResourceAny(ctx, args[1:])
	case "create-channel-instance":
		return runAgentServiceCreateChannelInstance(ctx, args[1:])
	case "list-channel-instances":
		return runAgentServiceListChannelInstances(ctx, args[1:])
	case "bind-channel":
		return runAgentBindChannel(ctx, args[1:])
	case "list-connector-catalog":
		return runAgentServiceListConnectorCatalog(ctx, args[1:])
	case "get-channel-instance":
		return runAgentServiceGetChannelInstance(ctx, args[1:])
	case "update-channel-instance":
		return runAgentServiceUpdateChannelInstance(ctx, args[1:])
	case "test-channel-configuration":
		return runAgentServiceTestChannelConfiguration(ctx, args[1:])
	case "disable-channel-instance":
		return runAgentServiceDisableChannelInstance(ctx, args[1:])
	case "test-channel-instance":
		return runAgentServiceTestChannelInstance(ctx, args[1:])
	case "create-callback-rule":
		return runAgentServiceCreateCallbackRule(ctx, args[1:])
	case "list-callback-rules":
		return runAgentServiceListCallbackRules(ctx, args[1:])
	case "get-callback-rule":
		return runAgentServiceGetCallbackRule(ctx, args[1:])
	case "update-callback-rule":
		return runAgentServiceUpdateCallbackRule(ctx, args[1:])
	case "delete-callback-rule":
		return runAgentServiceDeleteCallbackRule(ctx, args[1:])
	default:
		return fmt.Errorf("unknown agent subcommand: %s", sub)
	}
}

// ShowAgentHelp prints help for agent.
func ShowAgentHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: workspace-mcp, update-admin-feedback-ticket, admin-feedback-tickets")

	fmt.Fprintf(os.Stderr, `Usage: agent <subcommand> [options]

Product SDK surface: AgentService

Subcommands:
  artifacts                       AgentService.Artifacts
  artifact                        AgentService.Artifact
  delete-artifacts                AgentService.DeleteArtifacts
  resolve-artifacts               AgentService.ResolveArtifacts
  backfill-artifacts              AgentService.BackfillArtifacts
  list-resource-meta              Export selectable resource metadata to one JSON file
  builder                         Read authoring context, prepare workspace MCP and propose candidates
  init                            Create a local Agent package source tree
  build                           Build a deterministic local .moiagent package
  verify                          Verify a local source tree or .moiagent package
  unpack                          Restore editable sources from a .moiagent package
  list                            AgentService.List
  executor-health                 AgentService.ExecutorHealth
  runtime-browser-token           AgentService.RuntimeBrowserToken
  connect-runtime-browser         AgentService.ConnectRuntimeBrowser (handshake probe)
  connect-browser                 Read the initial task browser event and image
  watch-browser                   Read the initial task browser event and image without taking control
  preview-query-visual             AgentService.PreviewQueryVisual
  admin-feedback-ticket            AgentService.AdminFeedbackTicket
  build-feedback-github-issue-draft AgentService.BuildFeedbackGithubIssueDraft
  create-feedback-github-issue      AgentService.CreateFeedbackGithubIssue
  approve-feedback-reward          AgentService.ApproveFeedbackReward
  my-feedback-tickets              AgentService.MyFeedbackTickets
  create-feedback-ticket           AgentService.CreateFeedbackTicket
  export-agent-package             AgentService.ExportAgentPackage
  send                             (alias)
  get-task                         (alias)
  cancel-task                      (alias)
  card                             (alias)
  agent-card                       AgentService.AgentCard
  search-plugins                   AgentService.SearchPlugins
  describe-plugin-capability       AgentService.DescribePluginCapability
  list-plugin-releases             AgentService.ListPluginReleases
  get-plugin-release               AgentService.GetPluginRelease
  install-plugin                   AgentService.InstallPlugin
  list-plugin-installations        AgentService.ListPluginInstallations
  get-plugin-installation          AgentService.GetPluginInstallation
  update-plugin-installation       AgentService.UpdatePluginInstallation
  delete-plugin-installation       AgentService.DeletePluginInstallation
  create-plugin-connection         AgentService.CreatePluginConnection
  list-plugin-connections          AgentService.ListPluginConnections
  get-plugin-connection            AgentService.GetPluginConnection
  update-plugin-connection         AgentService.UpdatePluginConnection
  delete-plugin-connection         AgentService.DeletePluginConnection
  get-plugin-authorization         AgentService.GetPluginAuthorization
  store-plugin-secret              AgentService.StorePluginSecret
  begin-plugin-oauth               AgentService.BeginPluginOAuth
  poll-plugin-oauth                AgentService.PollPluginOAuth
  disconnect-plugin-connection     AgentService.DisconnectPluginConnection
  a2a                              AgentService.A2A
  a2a-stream                       AgentService.A2AStream
  load-agent-package               AgentService.LoadAgentPackage
  upload-workspace-file            AgentService.UploadWorkspaceFile
  inspect-skill-package            AgentService.InspectSkillPackage
  import-skill-package             AgentService.ImportSkillPackage
  confirm-skill-candidate          AgentService.ConfirmSkillCandidate
  get-skill-candidate-confirmation AgentService.GetSkillCandidateConfirmation
  get-skill-file                   AgentService.GetSkillFile
  polish-skill-stream              AgentService.PolishSkillStream
  resource-get                     AgentService.ResourceGet
  resource-post                    AgentService.ResourcePost
  resource-put                     AgentService.ResourcePut
  resource-patch                   AgentService.ResourcePatch
  resource-delete                  AgentService.ResourceDelete
  resource-any                     AgentService.ResourceAny
  list-agent-catalog-cards         AgentService.ListAgentCatalogCards
  create-team    AgentService.CreateTeam
  list-teams    AgentService.ListTeams
  get-team    AgentService.GetTeam
  update-team    AgentService.UpdateTeam
  delete-team    AgentService.DeleteTeam
  list-team-members    AgentService.ListTeamMembers
  add-team-member    AgentService.AddTeamMember
  remove-team-member    AgentService.RemoveTeamMember
  list-ingress-capabilities    AgentService.ListIngressCapabilities
  resolve-we-com-qr-binding    AgentService.ResolveWeComQRBinding
  list-ingress-connections    AgentService.ListIngressConnections
  create-ingress-connection    AgentService.CreateIngressConnection
  delete-ingress-connection    AgentService.DeleteIngressConnection
  get-ingress-connection    AgentService.GetIngressConnection
  update-ingress-connection    AgentService.UpdateIngressConnection
  verify-ingress-connection    AgentService.VerifyIngressConnection
  generate-we-com-callback-secrets AgentService.GenerateWeComCallbackSecrets
  verify-channel-callback        Client.VerifyChannelCallback
  post-channel-callback          Client.PostChannelCallback
  post-we-com-callback           Client.PostWeComCallback
  verify-ingress-callback        Client.VerifyIngressCallback
  post-ingress-callback          Client.PostIngressCallback
  list-team-sessions    AgentService.ListTeamSessions
  create-team-session    AgentService.CreateTeamSession
  rename-team-session    AgentService.RenameTeamSession
  submit-team-message    AgentService.SubmitTeamMessage
  list-team-messages    AgentService.ListTeamMessages
  create-channel-instance          AgentService.CreateChannelInstance
  list-channel-instances           AgentService.ListChannelInstances
  bind-channel                     Append an existing channel to the current Agent
  list-connector-catalog           AgentService.ListConnectorCatalog
  get-channel-instance             AgentService.GetChannelInstance
  update-channel-instance          AgentService.UpdateChannelInstance
  test-channel-configuration       AgentService.TestChannelConfiguration
  disable-channel-instance         AgentService.DisableChannelInstance
  test-channel-instance            AgentService.TestChannelInstance
  create-callback-rule             AgentService.CreateCallbackRule
  list-callback-rules              AgentService.ListCallbackRules
  get-callback-rule                AgentService.GetCallbackRule
  update-callback-rule             AgentService.UpdateCallbackRule
  delete-callback-rule             AgentService.DeleteCallbackRule

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runAgentServiceAgentCard(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent agent-card", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	agentCode := fs.String("agent-code", "", "Agent code")
	agentID := fs.String("agent-id", "", "Agent ID")
	agentWorkspaceID := fs.String("agent-workspace-id", "", "Workspace that owns the agent (query agent_workspace_id)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent agent-card (--agent-code|--agent-id) [options]\n\nProduct SDK: AgentService.AgentCard\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if *agentCode == "" && *agentID == "" {
		return fmt.Errorf("--agent-code or --agent-id is required")
	}
	var opts []sdk.QueryOption
	if *agentCode != "" {
		opts = append(opts, sdk.WithQueryParam("agent_code", *agentCode))
	}
	if *agentID != "" {
		opts = append(opts, sdk.WithQueryParam("agent_id", *agentID))
	}
	if *agentWorkspaceID != "" {
		opts = append(opts, sdk.WithQueryParam("agent_workspace_id", *agentWorkspaceID))
	}
	result, err := ctx.Client.Agents(wid).AgentCard(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceA2A(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent a2a", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	noWait := fs.Bool("no-wait", false, "Return a non-terminal message/send task admission without waiting for its terminal state")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent a2a [options]\n\nProduct SDK: AgentService.A2A\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var req sdk.AgentA2ARequest
	if len(rawBody) > 0 {
		if err := unmarshalJSONUseNumber(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	req.NoWait = *noWait
	result, err := ctx.Client.Agents(wid).A2A(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceA2AStream(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent a2a-stream", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	outPath := fs.String("output", "", "Write stream to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent a2a-stream [options]\n\nProduct SDK: AgentService.A2AStream\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var req sdk.AgentA2ARequest
	if len(rawBody) > 0 {
		if err := unmarshalJSONUseNumber(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	stream, err := ctx.Client.Agents(wid).A2AStream(context.Background(), req)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runAgentServiceLoadAgentPackage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent load-agent-package", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent load-agent-package [options]\n\nProduct SDK: AgentService.LoadAgentPackage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	source, err := upload.open("", *bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	defer source.Close()
	if err := source.bufferLocal(); err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).LoadAgentPackage(context.Background(), source.Reader)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceUploadWorkspaceFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent upload-workspace-file", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent upload-workspace-file [options]\n\nProduct SDK: AgentService.UploadWorkspaceFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	source, err := upload.open(*filename, *bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	defer source.Close()
	result, err := ctx.Client.Agents(wid).UploadWorkspaceFile(context.Background(), source.Name, source.Reader)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceInspectSkillPackage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent inspect-skill-package", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent inspect-skill-package [options]\n\nProduct SDK: AgentService.InspectSkillPackage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	source, err := upload.open(*filename, *bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	defer source.Close()
	result, err := ctx.Client.Agents(wid).InspectSkillPackage(context.Background(), source.Name, source.Reader)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceImportSkillPackage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent import-skill-package", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	name := fs.String("name", "", "Imported Skill name")
	description := fs.String("description", "", "Imported Skill description")
	category := fs.String("category", "", "Imported Skill category")
	tags := fs.String("tags", "", "Imported Skill tags as a JSON array or null")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent import-skill-package [options]\n\nProduct SDK: AgentService.ImportSkillPackage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	var importOptions []sdk.SkillPackageImportOption
	tagsProvided := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "name":
			importOptions = append(importOptions, sdk.WithSkillPackageImportName(*name))
		case "description":
			importOptions = append(importOptions, sdk.WithSkillPackageImportDescription(*description))
		case "category":
			importOptions = append(importOptions, sdk.WithSkillPackageImportCategory(*category))
		case "tags":
			tagsProvided = true
		}
	})
	if tagsProvided {
		var parsedTags []string
		if err := unmarshalJSONUseNumber([]byte(*tags), &parsedTags); err != nil {
			return fmt.Errorf("parse --tags: %w", err)
		}
		importOptions = append(importOptions, sdk.WithSkillPackageImportTags(parsedTags))
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	source, err := upload.open(*filename, *bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	defer source.Close()
	result, err := ctx.Client.Agents(wid).ImportSkillPackage(context.Background(), source.Name, source.Reader, importOptions...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceGetSkillFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-skill-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	skillID := fs.String("skill-id", "", "skillID")
	filePath := fs.String("file-path", "", "filePath")
	outPath := fs.String("output", "", "Write stream to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-skill-file [options]\n\nProduct SDK: AgentService.GetSkillFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	stream, err := ctx.Client.Agents(wid).GetSkillFile(context.Background(), *skillID, *filePath)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runAgentServicePolishSkillStream(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent polish-skill-stream", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	outPath := fs.String("output", "", "Write stream to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent polish-skill-stream [options]\n\nProduct SDK: AgentService.PolishSkillStream\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var draft map[string]any
	if len(rawBody) > 0 {
		if err := unmarshalJSONUseNumber(rawBody, &draft); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	stream, err := ctx.Client.Agents(wid).PolishSkillStream(context.Background(), draft)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runAgentServiceResourceGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent resource-get", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resource := fs.String("resource", "", "resource")
	tailPath := fs.String("tail-path", "", "tailPath")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent resource-get [options]\n\nProduct SDK: AgentService.ResourceGet\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).ResourceGet(context.Background(), *resource, *tailPath, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceResourcePost(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent resource-post", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resource := fs.String("resource", "", "resource")
	tailPath := fs.String("tail-path", "", "tailPath")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent resource-post [options]\n\nProduct SDK: AgentService.ResourcePost\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var body any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).ResourcePost(context.Background(), *resource, *tailPath, body, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceResourcePut(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent resource-put", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resource := fs.String("resource", "", "resource")
	tailPath := fs.String("tail-path", "", "tailPath")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent resource-put [options]\n\nProduct SDK: AgentService.ResourcePut\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var body any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).ResourcePut(context.Background(), *resource, *tailPath, body, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceResourcePatch(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent resource-patch", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resource := fs.String("resource", "", "resource")
	tailPath := fs.String("tail-path", "", "tailPath")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent resource-patch [options]\n\nProduct SDK: AgentService.ResourcePatch\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var body any
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &body); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).ResourcePatch(context.Background(), *resource, *tailPath, body, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceResourceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent resource-delete", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resource := fs.String("resource", "", "resource")
	tailPath := fs.String("tail-path", "", "tailPath")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent resource-delete [options]\n\nProduct SDK: AgentService.ResourceDelete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).ResourceDelete(context.Background(), *resource, *tailPath, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceResourceAny(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent resource-any", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	method := fs.String("method", "", "method")
	resource := fs.String("resource", "", "resource")
	tailPath := fs.String("tail-path", "", "tailPath")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent resource-any [options]\n\nProduct SDK: AgentService.ResourceAny\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var body any
	if len(rawBody) > 0 {
		if err := unmarshalJSONUseNumber(rawBody, &body); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).ResourceAny(context.Background(), *method, *resource, *tailPath, body, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceCreateChannelInstance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent create-channel-instance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "provider")
	bodyJSON := fs.String("json", "", "Native channel JSON with config and secrets; prefer --json-file for credentials")
	bodyFile := fs.String("json-file", "", "Native channel JSON file; GitHub secrets.access_token accepts OAuth/PAT (no vault reference)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent create-channel-instance [options]\n\nProduct SDK: AgentService.CreateChannelInstance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var req sdk.ChannelInstanceCreateRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).CreateChannelInstance(context.Background(), *provider, req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceListChannelInstances(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list-channel-instances", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "Channel provider ID (required; see agent list-connector-catalog)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent list-channel-instances --provider <provider> [options]\n\nLists channel instances in the current workspace. No Agent ID is needed.\n\nProduct SDK: AgentService.ListChannelInstances\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *provider == "" {
		return fmt.Errorf("--provider is required; run moi-cli -o agent list-connector-catalog to find channel provider IDs")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).ListChannelInstances(context.Background(), *provider, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentBindChannel(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent bind-channel", flag.ExitOnError)
	provider := fs.String("provider", "", "Provider of the existing channel instance (required)")
	instanceID := fs.String("instance-id", "", "Channel instance ID returned by list-channel-instances (required)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent bind-channel --provider <provider> --instance-id <id>\n\nBind an existing channel to the current Agent. Run from the session workspace root.\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *provider == "" || *instanceID == "" || fs.NArg() != 0 {
		return fmt.Errorf("bind-channel requires --provider and --instance-id")
	}
	wid, err := resolveWorkspaceID(ctx, "")
	if err != nil {
		return err
	}
	agentID, agentWorkspaceID, err := currentAgentBindingTarget(wid)
	if err != nil {
		return err
	}
	requestContext := context.Background()
	agents := ctx.Client.Agents(wid)
	instance, err := agents.GetChannelInstance(requestContext, *provider, *instanceID)
	if err != nil {
		return err
	}
	if instance.ID != *instanceID || instance.Provider != *provider || instance.WorkspaceID != wid || instance.ChannelType == "" {
		return &sdk.Error{Code: "ErrResponseInvalid", Message: "channel instance response does not match the requested instance"}
	}
	result, err := agents.ResourcePatch(requestContext, "agents", "/"+url.PathEscape(agentID)+"/bindings",
		map[string]any{"binding_ops": []map[string]any{{
			"op": "add_channel_binding", "provider": instance.Provider, "channel_type": instance.ChannelType,
			"channel_instance_ids": []string{instance.ID},
		}}}, sdk.WithQueryParam("agent_workspace_id", agentWorkspaceID))
	if err != nil {
		return err
	}
	saved, _ := result.AsMap()
	bindings, _ := saved["channel_bindings"].([]any)
	if saved["agent_id"] != agentID || saved["agent_workspace_id"] != agentWorkspaceID ||
		!slices.ContainsFunc(bindings, func(value any) bool {
			binding, _ := value.(map[string]any)
			ids, _ := binding["channel_instance_ids"].([]any)
			return binding["provider"] == instance.Provider && binding["channel_type"] == instance.ChannelType && slices.Contains(ids, any(instance.ID))
		}) {
		return &sdk.Error{Code: "ErrResponseInvalid", Message: "Agent bindings response does not confirm the requested channel binding"}
	}
	return printJSON(map[string]any{
		"agent_id": agentID, "agent_workspace_id": agentWorkspaceID, "provider": instance.Provider,
		"channel_type": instance.ChannelType, "instance_id": instance.ID, "bound": true,
	})
}

func currentAgentBindingTarget(workspaceID string) (agentID, agentWorkspaceID string, err error) {
	data, err := os.ReadFile("agent.json")
	if err != nil {
		return "", "", fmt.Errorf("read current Agent (run from the session workspace root): %w", err)
	}
	var agent struct {
		ID          string `json:"id"`
		WorkspaceID string `json:"agent_workspace_id"`
	}
	if err := json.Unmarshal(data, &agent); err != nil {
		return "", "", fmt.Errorf("read current Agent identity: %w", err)
	}
	if agent.ID == "" || agent.ID == "." || agent.ID == ".." || agent.ID != strings.TrimSpace(agent.ID) ||
		strings.ContainsAny(agent.ID, "/\\\r\n") || (agent.WorkspaceID != workspaceID && agent.WorkspaceID != "system") {
		return "", "", fmt.Errorf("agent.json does not identify an Agent in the current workspace or system")
	}
	return agent.ID, agent.WorkspaceID, nil
}

func runAgentServiceListConnectorCatalog(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list-connector-catalog", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent list-connector-catalog [options]\n\nProduct SDK: AgentService.ListConnectorCatalog\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).ListConnectorCatalog(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceGetChannelInstance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-channel-instance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "provider")
	instanceID := fs.String("instance-id", "", "instanceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-channel-instance [options]\n\nProduct SDK: AgentService.GetChannelInstance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).GetChannelInstance(context.Background(), *provider, *instanceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceUpdateChannelInstance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent update-channel-instance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "provider")
	instanceID := fs.String("instance-id", "", "instanceID")
	bodyJSON := fs.String("json", "", "Native channel JSON with config and secrets; prefer --json-file for credentials")
	bodyFile := fs.String("json-file", "", "Native channel JSON file; GitHub secrets.access_token accepts OAuth/PAT (no vault reference)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent update-channel-instance [options]\n\nProduct SDK: AgentService.UpdateChannelInstance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var req sdk.ChannelInstanceUpdateRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).UpdateChannelInstance(context.Background(), *provider, *instanceID, req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceTestChannelConfiguration(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent test-channel-configuration", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "provider")
	bodyJSON := fs.String("json", "", "Native channel JSON with config and secrets; prefer --json-file for credentials")
	bodyFile := fs.String("json-file", "", "Native channel JSON file; GitHub secrets.access_token accepts OAuth/PAT (no vault reference)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent test-channel-configuration [options]\n\nProduct SDK: AgentService.TestChannelConfiguration\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var req sdk.ChannelInstanceCreateRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).TestChannelConfiguration(context.Background(), *provider, req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceDisableChannelInstance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent disable-channel-instance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "provider")
	instanceID := fs.String("instance-id", "", "instanceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent disable-channel-instance [options]\n\nProduct SDK: AgentService.DisableChannelInstance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).DisableChannelInstance(context.Background(), *provider, *instanceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceTestChannelInstance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent test-channel-instance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "provider")
	instanceID := fs.String("instance-id", "", "instanceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent test-channel-instance [options]\n\nProduct SDK: AgentService.TestChannelInstance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).TestChannelInstance(context.Background(), *provider, *instanceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceCreateCallbackRule(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent create-callback-rule", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent create-callback-rule [options]\n\nProduct SDK: AgentService.CreateCallbackRule\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var req sdk.CallbackRuleCreateRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).CreateCallbackRule(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceListCallbackRules(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list-callback-rules", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent list-callback-rules [options]\n\nProduct SDK: AgentService.ListCallbackRules\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).ListCallbackRules(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceGetCallbackRule(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-callback-rule", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ruleID := fs.String("rule-id", "", "ruleID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-callback-rule [options]\n\nProduct SDK: AgentService.GetCallbackRule\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).GetCallbackRule(context.Background(), *ruleID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceUpdateCallbackRule(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent update-callback-rule", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ruleID := fs.String("rule-id", "", "ruleID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent update-callback-rule [options]\n\nProduct SDK: AgentService.UpdateCallbackRule\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var req sdk.CallbackRuleUpdateRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Agents(wid).UpdateCallbackRule(context.Background(), *ruleID, req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceDeleteCallbackRule(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent delete-callback-rule", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ruleID := fs.String("rule-id", "", "ruleID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent delete-callback-rule [options]\n\nProduct SDK: AgentService.DeleteCallbackRule\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).DeleteCallbackRule(context.Background(), *ruleID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceExportAgentPackage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent export-agent-package", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	agentID := fs.String("agent-id", "", "agent-id")
	version := fs.String("version", "", "version")
	output := fs.String("output", "", "output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent export-agent-package [options]\n\nProduct SDK: AgentService.ExportAgentPackage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).ExportAgentPackage(context.Background(), *agentID, *version)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}

func runAgentServiceCreateFeedbackTicket(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent create-feedback-ticket", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent create-feedback-ticket [options]\n\nProduct SDK: AgentService.CreateFeedbackTicket\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.FeedbackTicketCreateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).CreateFeedbackTicket(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceMyFeedbackTickets(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent my-feedback-tickets", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent my-feedback-tickets [options]\n\nProduct SDK: AgentService.MyFeedbackTickets\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).MyFeedbackTickets(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceAdminFeedbackTickets(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent admin-feedback-tickets", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).AdminFeedbackTickets(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceUpdateAdminFeedbackTicket(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent update-admin-feedback-ticket", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	ticketID := fs.String("ticket-id", "", "ticket-id")
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.FeedbackTicketAdminUpdateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).UpdateAdminFeedbackTicket(context.Background(), *ticketID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceWorkspaceMCP(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent workspace-mcp", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	agentID := fs.String("agent-id", "", "agent-id")
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.RuntimeRPCRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	client := ctx.Client
	if *grantFile != "" {
		raw, err := os.ReadFile(*grantFile)
		if err != nil {
			return err
		}
		client, err = sdk.New(ctx.Endpoint, sdk.WithAPIKey(ctx.APIKey), sdk.WithRuntimeAuthorization("Bearer "+strings.TrimSpace(string(raw))))
		if err != nil {
			return err
		}
	}
	if client != ctx.Client {
		defer client.Close()
	}
	result, err := client.Agents(wid).WorkspaceMCP(context.Background(), *agentID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
