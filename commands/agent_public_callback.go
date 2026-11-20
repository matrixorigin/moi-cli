package commands

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
)

func runAgentPublicCallback(ctx *Context, command string, args []string) error {
	fs := flag.NewFlagSet("agent "+command, flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	provider := fs.String("provider", "", "Callback provider")
	credentialRef := fs.String("credential-ref", "", "Channel credential reference")
	connectionID := fs.String("connection-id", "", "Ingress connection ID")
	bodyFile := fs.String("body-file", "", "Raw provider payload file; use - for stdin")
	var query queryParameters
	fs.Var(&query, "query-param", "Provider query parameter name=value (repeatable)")
	setSDKCommandUsage(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var body []byte
	if command == "post-channel-callback" || command == "post-we-com-callback" || command == "post-ingress-callback" {
		if *bodyFile == "" {
			return fmt.Errorf("--body-file is required")
		}
		body, err = readAgentPublicCallbackBody(*bodyFile)
		if err != nil {
			return err
		}
	}
	var result []byte
	switch command {
	case "verify-channel-callback":
		result, err = ctx.Client.VerifyChannelCallback(context.Background(), *provider, wid, *credentialRef, query...)
	case "post-channel-callback":
		result, err = ctx.Client.PostChannelCallback(context.Background(), *provider, wid, *credentialRef, body, query...)
	case "post-we-com-callback":
		result, err = ctx.Client.PostWeComCallback(context.Background(), wid, *credentialRef, body, query...)
	case "verify-ingress-callback":
		result, err = ctx.Client.VerifyIngressCallback(context.Background(), *provider, wid, *connectionID, query...)
	case "post-ingress-callback":
		result, err = ctx.Client.PostIngressCallback(context.Background(), *provider, wid, *connectionID, body, query...)
	default:
		return fmt.Errorf("unknown public callback operation %q", command)
	}
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(result)
	return err
}

func readAgentPublicCallbackBody(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}
