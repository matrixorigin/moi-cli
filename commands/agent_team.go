package commands

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runAgentTeamCommand(ctx *Context, command string, args []string) error {
	fs := flag.NewFlagSet("agent "+command, flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: moi-cli agent %s [options]\n\nOptions:\n", command)
		fs.PrintDefaults()
	}
	workspace := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	team := fs.String("team-id", "", "Team ID")
	contextID := fs.String("context-id", "", "Team Session context ID")
	memberWorkspace := fs.String("agent-workspace-id", "", "Member workspace ID")
	agent := fs.String("agent-id", "", "Member Agent ID")
	connection := fs.String("connection-id", "", "Ingress connection ID")
	qrSession := fs.String("qr-session-token", "", "Existing WeCom QR session token (empty starts a session)")
	configVersion := fs.Int64("expected-config-version", 0, "Expected ingress configuration version")
	targetKind := fs.String("target-kind", "", "Channel target kind: agent or team")
	targetWorkspace := fs.String("target-workspace-id", "", "Channel target workspace ID")
	targetID := fs.String("target-id", "", "Channel target ID")
	version := fs.Int64("expected-version", 0, "Expected Team, membership or association version for this operation")
	bodyJSON := fs.String("json", "", "Typed request JSON")
	bodyFile := fs.String("json-file", "", "Typed request JSON file")
	catalog := fs.String("catalog", "", "Agent catalog filter")
	query := fs.String("query", "", "Agent catalog search query")
	includeDevelopment := fs.Bool("include-development", false, "Include development Agents")
	limit := fs.Int("limit", 50, "Page size")
	offset := fs.Int("offset", 0, "Page offset")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	workspaceID, err := resolveWorkspaceID(ctx, *workspace)
	if err != nil {
		return err
	}
	service := ctx.Client.Agents(workspaceID)
	callCtx := context.Background()
	decodeRequest := func(request any) error {
		raw, err := loadJSONBytes(*bodyJSON, *bodyFile)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, request); err != nil {
			return fmt.Errorf("parse Team request: %w", err)
		}
		return nil
	}
	var result any
	switch command {
	case "list-agent-catalog-cards":
		options := []sdk.QueryOption{sdk.WithQueryParam("limit", fmt.Sprint(*limit)), sdk.WithQueryParam("offset", fmt.Sprint(*offset))}
		fs.Visit(func(current *flag.Flag) {
			switch current.Name {
			case "catalog":
				options = append(options, sdk.WithQueryParam("catalog", *catalog))
			case "query":
				options = append(options, sdk.WithQueryParam("query", *query))
			case "include-development":
				options = append(options, sdk.WithQueryParam("include_development", fmt.Sprint(*includeDevelopment)))
			}
		})
		result, err = service.ListAgentCatalogCards(callCtx, options...)

	case "create-team":
		var request sdk.AgentTeamCreateRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.CreateTeam(callCtx, request)

	case "list-teams":
		result, err = service.ListTeams(callCtx, sdk.WithQueryParam("limit", fmt.Sprint(*limit)), sdk.WithQueryParam("offset", fmt.Sprint(*offset)))

	case "get-team":
		result, err = service.GetTeam(callCtx, *team)

	case "update-team":
		var request sdk.AgentTeamUpdateRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.UpdateTeam(callCtx, *team, request)

	case "delete-team":
		result, err = service.DeleteTeam(callCtx, *team, *version)

	case "list-team-members":
		result, err = service.ListTeamMembers(callCtx, *team)

	case "add-team-member":
		result, err = service.AddTeamMember(callCtx, *team, *memberWorkspace, *agent, *version)

	case "remove-team-member":
		result, err = service.RemoveTeamMember(callCtx, *team, *memberWorkspace, *agent, *version)

	case "list-ingress-capabilities":
		result, err = service.ListIngressCapabilities(callCtx)
	case "resolve-we-com-qr-binding":
		result, err = service.ResolveWeComQRBinding(callCtx, &sdk.AgentTargetRef{Kind: *targetKind, WorkspaceId: *targetWorkspace, Id: *targetID}, *qrSession)

	case "list-ingress-connections":
		result, err = service.ListIngressConnections(callCtx, &sdk.AgentTargetRef{Kind: *targetKind, WorkspaceId: *targetWorkspace, Id: *targetID})

	case "create-ingress-connection":
		var request sdk.IngressConnectionCreateRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.CreateIngressConnection(callCtx, request)

	case "delete-ingress-connection":
		result, err = service.DeleteIngressConnection(callCtx, *connection, *version, *configVersion)

	case "get-ingress-connection":
		result, err = service.GetIngressConnection(callCtx, *connection)

	case "update-ingress-connection":
		var request sdk.IngressConnectionUpdateRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.UpdateIngressConnection(callCtx, *connection, request)

	case "verify-ingress-connection":
		var request sdk.IngressConnectionVerifyRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.VerifyIngressConnection(callCtx, *connection, request)

	case "generate-we-com-callback-secrets":
		result, err = service.GenerateWeComCallbackSecrets(callCtx)

	case "list-team-sessions":
		result, err = service.ListTeamSessions(callCtx, *team, sdk.WithQueryParam("limit", fmt.Sprint(*limit)), sdk.WithQueryParam("offset", fmt.Sprint(*offset)))

	case "create-team-session":
		var request sdk.AgentTeamSessionCreateRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.CreateTeamSession(callCtx, *team, request)

	case "rename-team-session":
		var request sdk.AgentTeamSessionRenameRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.RenameTeamSession(callCtx, *team, *contextID, request)

	case "submit-team-message":
		var request sdk.AgentTeamMessageRequest
		if err := decodeRequest(&request); err != nil {
			return err
		}
		result, err = service.SubmitTeamMessage(callCtx, *team, *contextID, request)

	case "list-team-messages":
		result, err = service.ListTeamMessages(callCtx, *team, *contextID, sdk.WithQueryParam("limit", fmt.Sprint(*limit)), sdk.WithQueryParam("offset", fmt.Sprint(*offset)))

	default:
		return fmt.Errorf("unknown Team operation %q", command)
	}
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
