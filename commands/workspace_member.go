package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

type workspaceMemberFilterFlag struct {
	name   string
	values []string
}

type workspaceMemberFilterFlags []workspaceMemberFilterFlag

func (flags *workspaceMemberFilterFlags) String() string { return "" }

func (flags *workspaceMemberFilterFlags) Set(raw string) error {
	name, rawValues, ok := strings.Cut(raw, "=")
	if !ok || name == "" {
		return fmt.Errorf("--filter requires name=value1,value2")
	}
	values := splitCSV(rawValues)
	if values == nil {
		values = []string{}
	}
	*flags = append(*flags, workspaceMemberFilterFlag{name: name, values: values})
	return nil
}

type workspaceMemberTagFlag struct {
	id    int
	value string
}

type workspaceMemberTagFlags []workspaceMemberTagFlag

func (flags *workspaceMemberTagFlags) String() string { return "" }

func (flags *workspaceMemberTagFlags) Set(raw string) error {
	rawID, value, ok := strings.Cut(raw, "=")
	if !ok {
		return fmt.Errorf("--tag requires id=value")
	}
	id, err := strconv.Atoi(rawID)
	if err != nil {
		return fmt.Errorf("--tag: %q is not an integer", rawID)
	}
	*flags = append(*flags, workspaceMemberTagFlag{id: id, value: value})
	return nil
}

// ExecuteWorkspaceMemberCommand runs product SDK WorkspaceMemberService methods via moi-cli.
func ExecuteWorkspaceMemberCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowWorkspaceMemberHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "accept-invitation":
		return runWorkspaceMemberServiceAcceptInvitation(ctx, args[1:])
	case "current-principal":
		return runWorkspaceMemberServiceCurrentPrincipal(ctx, args[1:])
	case "members":
		return runWorkspaceMemberServiceMembers(ctx, args[1:])
	case "invite":
		return runWorkspaceMemberServiceInvite(ctx, args[1:])
	case "member":
		return runWorkspaceMemberServiceMember(ctx, args[1:])
	case "update":
		return runWorkspaceMemberServiceUpdate(ctx, args[1:])
	case "delete":
		return runWorkspaceMemberServiceDelete(ctx, args[1:])
	case "update-status":
		return runWorkspaceMemberServiceUpdateStatus(ctx, args[1:])
	case "create-user-tag":
		return runWorkspaceMemberServiceCreateUserTag(ctx, args[1:])
	case "user-tags":
		return runWorkspaceMemberServiceUserTags(ctx, args[1:])
	case "update-user-tag":
		return runWorkspaceMemberServiceUpdateUserTag(ctx, args[1:])
	case "delete-user-tag":
		return runWorkspaceMemberServiceDeleteUserTag(ctx, args[1:])
	case "audit-logs":
		return runWorkspaceMemberServiceAuditLogs(ctx, args[1:])
	case "audit-log-overview":
		return runWorkspaceMemberServiceAuditLogOverview(ctx, args[1:])
	default:
		return fmt.Errorf("unknown workspace-member subcommand: %s", sub)
	}
}

// ShowWorkspaceMemberHelp prints help for workspace-member.
func ShowWorkspaceMemberHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: workspace-member <subcommand> [options]

Product SDK surface: WorkspaceMemberService

Subcommands:
  accept-invitation                WorkspaceMemberService.AcceptInvitation
  current-principal                WorkspaceMemberService.CurrentPrincipal
  members                          WorkspaceMemberService.Members
  invite                           WorkspaceMemberService.Invite
  member                           WorkspaceMemberService.Member
  update                           WorkspaceMemberService.Update
  delete                           WorkspaceMemberService.Delete
  update-status                    WorkspaceMemberService.UpdateStatus
  create-user-tag                  WorkspaceMemberService.CreateUserTag
  user-tags                        WorkspaceMemberService.UserTags
  update-user-tag                  WorkspaceMemberService.UpdateUserTag
  delete-user-tag                  WorkspaceMemberService.DeleteUserTag
  audit-logs                       WorkspaceMemberService.AuditLogs
  audit-log-overview               WorkspaceMemberService.AuditLogOverview

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runWorkspaceMemberServiceCurrentPrincipal(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member current-principal", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member current-principal [options]\n\nProduct SDK: WorkspaceMemberService.CurrentPrincipal\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).CurrentPrincipal(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceMembers(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member members", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	page := fs.Int("page", 0, "Page number")
	pageSize := fs.Int("page-size", 0, "Page size")
	var filters workspaceMemberFilterFlags
	fs.Var(&filters, "filter", "Member filter name=value1,value2 (repeatable)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member members [options]\n\nProduct SDK: WorkspaceMemberService.Members\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.WorkspaceMemberListOption
	fs.Visit(func(current *flag.Flag) {
		switch current.Name {
		case "page":
			options = append(options, sdk.WithMemberPage(*page))
		case "page-size":
			options = append(options, sdk.WithMemberPageSize(*pageSize))
		}
	})
	for _, filter := range filters {
		options = append(options, sdk.WithMemberFilter(filter.name, filter.values...))
	}
	result, err := ctx.Client.WorkspaceMembers(wid).Members(context.Background(), options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceInvite(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member invite", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	email := fs.String("email", "", "Invitee email address")
	mobile := fs.String("mobile-number", "", "Invitee mainland China mobile number")
	github := fs.String("github-login", "", "Invitee GitHub username")
	roleIDs := fs.String("role-ids", "", "Comma-separated assigned role IDs")
	description := fs.String("description", "", "Member description")
	defaultRoleID := fs.Int("default-role-id", 0, "defaultRoleID")
	autoAccept := fs.Bool("auto-accept", false, "autoAccept")
	var tags workspaceMemberTagFlags
	fs.Var(&tags, "tag", "User tag id=value (repeatable)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member invite [options]\n\nProduct SDK: WorkspaceMemberService.Invite\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	assignedRoles, err := parseIntCSV("role-ids", *roleIDs)
	if err != nil {
		return err
	}
	options := []sdk.CreateUserOption{sdk.WithUserRoleIDs(assignedRoles...), sdk.WithUserAutoAccept(*autoAccept)}
	if *email != "" {
		options = append(options, sdk.WithUserEmail(*email))
	}
	if *mobile != "" {
		options = append(options, sdk.WithUserMobileNumber(*mobile))
	}
	if *github != "" {
		options = append(options, sdk.WithUserGithubLogin(*github))
	}
	fs.Visit(func(current *flag.Flag) {
		if current.Name == "description" {
			options = append(options, sdk.WithUserDescription(*description))
		}
	})
	for _, tag := range tags {
		options = append(options, sdk.WithCreateUserTag(tag.id, tag.value))
	}
	result, err := ctx.Client.WorkspaceMembers(wid).Invite(context.Background(), *name, *defaultRoleID, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceMember(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member member", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	userID := fs.String("user-id", "", "userID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member member [options]\n\nProduct SDK: WorkspaceMemberService.Member\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).Member(context.Background(), *userID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	userID := fs.String("user-id", "", "userID")
	name := fs.String("name", "", "name")
	description := fs.String("description", "", "description")
	var tags workspaceMemberTagFlags
	fs.Var(&tags, "tag", "User tag id=value (repeatable)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member update [options]\n\nProduct SDK: WorkspaceMemberService.Update\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	options := make([]sdk.UpdateUserOption, 0, len(tags))
	for _, tag := range tags {
		options = append(options, sdk.WithUserTag(tag.id, tag.value))
	}
	if err := ctx.Client.WorkspaceMembers(wid).Update(context.Background(), *userID, *name, *description, options...); err != nil {
		return err
	}
	fmt.Printf("workspace-member update: ok\n")
	return nil
}

func runWorkspaceMemberServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	userID := fs.String("user-id", "", "userID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member delete [options]\n\nProduct SDK: WorkspaceMemberService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.WorkspaceMembers(wid).Delete(context.Background(), *userID); err != nil {
		return err
	}
	fmt.Printf("workspace-member delete: ok\n")
	return nil
}

func runWorkspaceMemberServiceUpdateStatus(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member update-status", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	userID := fs.String("user-id", "", "userID")
	action := fs.String("action", "", "action")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member update-status [options]\n\nProduct SDK: WorkspaceMemberService.UpdateStatus\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.WorkspaceMembers(wid).UpdateStatus(context.Background(), *userID, *action); err != nil {
		return err
	}
	fmt.Printf("workspace-member update-status: ok\n")
	return nil
}

func runWorkspaceMemberServiceCreateUserTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member create-user-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member create-user-tag [options]\n\nProduct SDK: WorkspaceMemberService.CreateUserTag\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).CreateUserTag(context.Background(), *name)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceUserTags(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member user-tags", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member user-tags [options]\n\nProduct SDK: WorkspaceMemberService.UserTags\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).UserTags(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceUpdateUserTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member update-user-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tagID := fs.Int("tag-id", 0, "tagID")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member update-user-tag [options]\n\nProduct SDK: WorkspaceMemberService.UpdateUserTag\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).UpdateUserTag(context.Background(), *tagID, *name)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceDeleteUserTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member delete-user-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tagID := fs.Int("tag-id", 0, "tagID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member delete-user-tag [options]\n\nProduct SDK: WorkspaceMemberService.DeleteUserTag\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).DeleteUserTag(context.Background(), *tagID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceMemberServiceAuditLogs(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member audit-logs", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	method := fs.String("method", "", "HTTP method")
	result := fs.String("result", "", "Audit result")
	statusCode := fs.Int("status-code", 0, "HTTP status code")
	startTime := fs.String("start-time", "", "Start time (RFC3339)")
	endTime := fs.String("end-time", "", "End time (RFC3339)")
	page := fs.Int("page", 0, "Page number")
	pageSize := fs.Int("page-size", 0, "Page size")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member audit-logs [options]\n\nProduct SDK: WorkspaceMemberService.AuditLogs\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.AuditLogOption
	var optionErr error
	fs.Visit(func(current *flag.Flag) {
		if optionErr != nil {
			return
		}
		switch current.Name {
		case "method":
			options = append(options, sdk.WithAuditMethod(*method))
		case "result":
			options = append(options, sdk.WithAuditResult(*result))
		case "status-code":
			options = append(options, sdk.WithAuditStatusCode(*statusCode))
		case "start-time", "end-time":
			raw := *startTime
			if current.Name == "end-time" {
				raw = *endTime
			}
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				optionErr = fmt.Errorf("--%s: %w", current.Name, err)
				return
			}
			if current.Name == "start-time" {
				options = append(options, sdk.WithAuditStartTime(parsed))
			} else {
				options = append(options, sdk.WithAuditEndTime(parsed))
			}
		case "page":
			options = append(options, sdk.WithPage(*page))
		case "page-size":
			options = append(options, sdk.WithPageSize(*pageSize))
		}
	})
	if optionErr != nil {
		return optionErr
	}
	auditResult, err := ctx.Client.WorkspaceMembers(wid).AuditLogs(context.Background(), options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, auditResult, func() error { return printJSON(auditResult) })
}

func runWorkspaceMemberServiceAuditLogOverview(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member audit-log-overview", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace-member audit-log-overview [options]\n\nProduct SDK: WorkspaceMemberService.AuditLogOverview\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).AuditLogOverview(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
