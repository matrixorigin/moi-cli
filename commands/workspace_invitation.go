package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
)

func runWorkspaceMemberServiceAcceptInvitation(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace-member accept-invitation", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	invitationID := fs.String("invitation-id", "", "Pending invitation ID")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli workspace-member accept-invitation [options]\n\nProduct SDK: WorkspaceMemberService.AcceptInvitation\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.WorkspaceMembers(wid).AcceptInvitation(context.Background(), *invitationID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
