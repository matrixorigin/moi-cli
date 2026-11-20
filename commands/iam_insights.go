package commands

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"os"
)

func runIAMServiceOwnershipImpact(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam ownership-impact", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	roleID := fs.String("role-id", "", "目标 ID")
	var params queryParameters
	fs.Var(&params, "query-param", "筛选或分页参数 name=value，可重复")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli iam ownership-impact [options]\n\nProduct SDK: IAMService.OwnershipImpact\n\nOptions:")
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
	result, err := ctx.Client.IAM(wid).OwnershipImpact(context.Background(), *roleID, params...)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runIAMServiceRoleAuditEvents(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam role-audit-events", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	roleID := fs.String("role-id", "", "目标 ID")
	var params queryParameters
	fs.Var(&params, "query-param", "筛选或分页参数 name=value，可重复")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli iam role-audit-events [options]\n\nProduct SDK: IAMService.RoleAuditEvents\n\nOptions:")
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
	result, err := ctx.Client.IAM(wid).RoleAuditEvents(context.Background(), *roleID, params...)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runIAMServiceUserAuditEvents(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam user-audit-events", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	userID := fs.String("user-id", "", "目标 ID")
	var params queryParameters
	fs.Var(&params, "query-param", "筛选或分页参数 name=value，可重复")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli iam user-audit-events [options]\n\nProduct SDK: IAMService.UserAuditEvents\n\nOptions:")
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
	result, err := ctx.Client.IAM(wid).UserAuditEvents(context.Background(), *userID, params...)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runIAMServiceSetDefaultRole(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam set-default-role", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	bodyJSON := fs.String("json", "", "请求 JSON")
	bodyFile := fs.String("json-file", "", "请求 JSON 文件")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli iam set-default-role [options]\n\nProduct SDK: IAMService.SetDefaultRole\n\nOptions:")
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
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.IAMSetDefaultRoleRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.IAM(wid).SetDefaultRole(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSON(result)
}
