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

func runAgentServiceArtifacts(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent artifacts", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	var params queryParameters
	fs.Var(&params, "query-param", "筛选或分页参数 name=value，可重复")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli agent artifacts [options]\n\nProduct SDK: AgentService.Artifacts\n\nOptions:")
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
	result, err := ctx.Client.Agents(wid).Artifacts(context.Background(), params...)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runAgentServiceArtifact(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent artifact", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	artifactID := fs.String("artifact-id", "", "目标 ID")
	version := fs.Int64("version", 0, "成果版本")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli agent artifact [options]\n\nProduct SDK: AgentService.Artifact\n\nOptions:")
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
	result, err := ctx.Client.Agents(wid).Artifact(context.Background(), *artifactID, *version)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runAgentServiceDeleteArtifacts(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent delete-artifacts", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	bodyJSON := fs.String("json", "", "请求 JSON")
	bodyFile := fs.String("json-file", "", "请求 JSON 文件")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli agent delete-artifacts [options]\n\nProduct SDK: AgentService.DeleteArtifacts\n\nOptions:")
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
	var request sdk.AgentArtifactBatchRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).DeleteArtifacts(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runAgentServiceResolveArtifacts(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent resolve-artifacts", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	bodyJSON := fs.String("json", "", "请求 JSON")
	bodyFile := fs.String("json-file", "", "请求 JSON 文件")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli agent resolve-artifacts [options]\n\nProduct SDK: AgentService.ResolveArtifacts\n\nOptions:")
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
	var request sdk.AgentArtifactBatchRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).ResolveArtifacts(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSON(result)
}

func runAgentServiceBackfillArtifacts(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent backfill-artifacts", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "工作区 ID，省略时读取配置")
	bodyJSON := fs.String("json", "", "请求 JSON")
	bodyFile := fs.String("json-file", "", "请求 JSON 文件")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli agent backfill-artifacts [options]\n\nProduct SDK: AgentService.BackfillArtifacts\n\nOptions:")
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
	var request sdk.AgentArtifactBackfillRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).BackfillArtifacts(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSON(result)
}
