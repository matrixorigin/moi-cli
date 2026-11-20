package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"os"
)

func ExecuteConversationCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowConversationHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "append-modified-response":
		return runConversationServiceAppendModifiedResponse(ctx, args[1:])
	case "modify-response":
		return runConversationServiceModifyResponse(ctx, args[1:])
	case "remove-message-tag":
		return runConversationServiceRemoveMessageTag(ctx, args[1:])
	case "add-message-tag":
		return runConversationServiceAddMessageTag(ctx, args[1:])
	case "list-message-tags":
		return runConversationServiceListMessageTags(ctx, args[1:])
	case "remove-session-tag":
		return runConversationServiceRemoveSessionTag(ctx, args[1:])
	case "add-session-tag":
		return runConversationServiceAddSessionTag(ctx, args[1:])
	case "list-session-tags":
		return runConversationServiceListSessionTags(ctx, args[1:])
	case "delete-tag":
		return runConversationServiceDeleteTag(ctx, args[1:])
	case "list-tags":
		return runConversationServiceListTags(ctx, args[1:])
	case "create-tag":
		return runConversationServiceCreateTag(ctx, args[1:])
	case "latest-message-id":
		return runConversationServiceLatestMessageID(ctx, args[1:])
	case "get-message":
		return runConversationServiceGetMessage(ctx, args[1:])
	case "list-messages":
		return runConversationServiceListMessages(ctx, args[1:])
	case "create-message":
		return runConversationServiceCreateMessage(ctx, args[1:])
	case "delete-session":
		return runConversationServiceDeleteSession(ctx, args[1:])
	case "update-session":
		return runConversationServiceUpdateSession(ctx, args[1:])
	case "get-session":
		return runConversationServiceGetSession(ctx, args[1:])
	case "list-sessions":
		return runConversationServiceListSessions(ctx, args[1:])
	case "create-session":
		return runConversationServiceCreateSession(ctx, args[1:])
	case "latest-ui-message-id":
		return runConversationServiceLatestUIMessageID(ctx, args[1:])
	case "dislike-ui-message":
		return runConversationServiceDislikeUIMessage(ctx, args[1:])
	case "like-ui-message":
		return runConversationServiceLikeUIMessage(ctx, args[1:])
	case "get-ui-message":
		return runConversationServiceGetUIMessage(ctx, args[1:])
	case "list-ui-messages":
		return runConversationServiceListUIMessages(ctx, args[1:])
	case "list-unpinned-ui-sessions":
		return runConversationServiceListUnpinnedUISessions(ctx, args[1:])
	case "list-pinned-ui-sessions":
		return runConversationServiceListPinnedUISessions(ctx, args[1:])
	case "list-ui-sessions":
		return runConversationServiceListUISessions(ctx, args[1:])
	case "unpin-ui-session":
		return runConversationServiceUnpinUISession(ctx, args[1:])
	case "pin-ui-session":
		return runConversationServicePinUISession(ctx, args[1:])
	case "delete-ui-session":
		return runConversationServiceDeleteUISession(ctx, args[1:])
	case "get-ui-session":
		return runConversationServiceGetUISession(ctx, args[1:])
	case "update-ui-session":
		return runConversationServiceUpdateUISession(ctx, args[1:])
	case "create-ui-session":
		return runConversationServiceCreateUISession(ctx, args[1:])
	default:
		return fmt.Errorf("unknown conversation subcommand: %s", sub)
	}
}
func ShowConversationHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: append-modified-response, modify-response")

	fmt.Fprint(os.Stderr, `Usage: moi-cli conversation <subcommand> [options]
Subcommands:
  remove-message-tag               ConversationService.RemoveMessageTag
  add-message-tag                  ConversationService.AddMessageTag
  list-message-tags                ConversationService.ListMessageTags
  remove-session-tag               ConversationService.RemoveSessionTag
  add-session-tag                  ConversationService.AddSessionTag
  list-session-tags                ConversationService.ListSessionTags
  delete-tag                       ConversationService.DeleteTag
  list-tags                        ConversationService.ListTags
  create-tag                       ConversationService.CreateTag
  latest-message-id                ConversationService.LatestMessageID
  get-message                      ConversationService.GetMessage
  list-messages                    ConversationService.ListMessages
  create-message                   ConversationService.CreateMessage
  delete-session                   ConversationService.DeleteSession
  update-session                   ConversationService.UpdateSession
  get-session                      ConversationService.GetSession
  list-sessions                    ConversationService.ListSessions
  create-session                   ConversationService.CreateSession
  latest-ui-message-id             ConversationService.LatestUIMessageID
  dislike-ui-message               ConversationService.DislikeUIMessage
  like-ui-message                  ConversationService.LikeUIMessage
  get-ui-message                   ConversationService.GetUIMessage
  list-ui-messages                 ConversationService.ListUIMessages
  list-unpinned-ui-sessions        ConversationService.ListUnpinnedUISessions
  list-pinned-ui-sessions          ConversationService.ListPinnedUISessions
  list-ui-sessions                 ConversationService.ListUISessions
  unpin-ui-session                 ConversationService.UnpinUISession
  pin-ui-session                   ConversationService.PinUISession
  delete-ui-session                ConversationService.DeleteUISession
  get-ui-session                   ConversationService.GetUISession
  update-ui-session                ConversationService.UpdateUISession
  create-ui-session                ConversationService.CreateUISession
`)
}

func runConversationServiceCreateUISession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation create-ui-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation create-ui-session [options]\n\nProduct SDK: ConversationService.CreateUISession\n\nOptions:\n")
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
	var request sdk.ConversationUISessionRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).CreateUISession(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceUpdateUISession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation update-ui-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation update-ui-session [options]\n\nProduct SDK: ConversationService.UpdateUISession\n\nOptions:\n")
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
	var request sdk.ConversationUISessionRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).UpdateUISession(context.Background(), *sessionID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceGetUISession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation get-ui-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation get-ui-session [options]\n\nProduct SDK: ConversationService.GetUISession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).GetUISession(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceDeleteUISession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation delete-ui-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation delete-ui-session [options]\n\nProduct SDK: ConversationService.DeleteUISession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).DeleteUISession(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServicePinUISession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation pin-ui-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation pin-ui-session [options]\n\nProduct SDK: ConversationService.PinUISession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).PinUISession(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceUnpinUISession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation unpin-ui-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation unpin-ui-session [options]\n\nProduct SDK: ConversationService.UnpinUISession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).UnpinUISession(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceListUISessions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-ui-sessions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-ui-sessions [options]\n\nProduct SDK: ConversationService.ListUISessions\n\nOptions:\n")
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
	var request sdk.ConversationUISessionListRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).ListUISessions(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceListPinnedUISessions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-pinned-ui-sessions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-pinned-ui-sessions [options]\n\nProduct SDK: ConversationService.ListPinnedUISessions\n\nOptions:\n")
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
	var request sdk.ConversationUISessionListRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).ListPinnedUISessions(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceListUnpinnedUISessions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-unpinned-ui-sessions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-unpinned-ui-sessions [options]\n\nProduct SDK: ConversationService.ListUnpinnedUISessions\n\nOptions:\n")
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
	var request sdk.ConversationUISessionListRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).ListUnpinnedUISessions(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceListUIMessages(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-ui-messages", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-ui-messages [options]\n\nProduct SDK: ConversationService.ListUIMessages\n\nOptions:\n")
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
	var request sdk.ConversationUIMessageListRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).ListUIMessages(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceGetUIMessage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation get-ui-message", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation get-ui-message [options]\n\nProduct SDK: ConversationService.GetUIMessage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).GetUIMessage(context.Background(), *messageID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceLikeUIMessage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation like-ui-message", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation like-ui-message [options]\n\nProduct SDK: ConversationService.LikeUIMessage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).LikeUIMessage(context.Background(), *messageID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceDislikeUIMessage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation dislike-ui-message", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation dislike-ui-message [options]\n\nProduct SDK: ConversationService.DislikeUIMessage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).DislikeUIMessage(context.Background(), *messageID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceLatestUIMessageID(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation latest-ui-message-id", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation latest-ui-message-id [options]\n\nProduct SDK: ConversationService.LatestUIMessageID\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).LatestUIMessageID(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceCreateSession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation create-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation create-session [options]\n\nProduct SDK: ConversationService.CreateSession\n\nOptions:\n")
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
	var request sdk.ConversationWorkspaceSessionRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).CreateSession(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceListSessions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-sessions", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-sessions [options]\n\nProduct SDK: ConversationService.ListSessions\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).ListSessions(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceGetSession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation get-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation get-session [options]\n\nProduct SDK: ConversationService.GetSession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).GetSession(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceUpdateSession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation update-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation update-session [options]\n\nProduct SDK: ConversationService.UpdateSession\n\nOptions:\n")
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
	var request sdk.ConversationWorkspaceSessionRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).UpdateSession(context.Background(), *sessionID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceDeleteSession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation delete-session", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation delete-session [options]\n\nProduct SDK: ConversationService.DeleteSession\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	err = ctx.Client.Conversations(wid).DeleteSession(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return nil
}

func runConversationServiceCreateMessage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation create-message", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation create-message [options]\n\nProduct SDK: ConversationService.CreateMessage\n\nOptions:\n")
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
	var request sdk.ConversationWorkspaceMessageRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).CreateMessage(context.Background(), *sessionID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceListMessages(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-messages", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-messages [options]\n\nProduct SDK: ConversationService.ListMessages\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).ListMessages(context.Background(), *sessionID, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceGetMessage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation get-message", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation get-message [options]\n\nProduct SDK: ConversationService.GetMessage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).GetMessage(context.Background(), *messageID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceLatestMessageID(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation latest-message-id", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation latest-message-id [options]\n\nProduct SDK: ConversationService.LatestMessageID\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).LatestMessageID(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceCreateTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation create-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation create-tag [options]\n\nProduct SDK: ConversationService.CreateTag\n\nOptions:\n")
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
	var request sdk.ConversationTagRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Conversations(wid).CreateTag(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceListTags(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-tags", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-tags [options]\n\nProduct SDK: ConversationService.ListTags\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).ListTags(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceDeleteTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation delete-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	source := fs.String("source", "", "source")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation delete-tag [options]\n\nProduct SDK: ConversationService.DeleteTag\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	err = ctx.Client.Conversations(wid).DeleteTag(context.Background(), *source, *name)
	if err != nil {
		return err
	}
	return nil
}

func runConversationServiceListSessionTags(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-session-tags", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-session-tags [options]\n\nProduct SDK: ConversationService.ListSessionTags\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).ListSessionTags(context.Background(), *sessionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceAddSessionTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation add-session-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation add-session-tag [options]\n\nProduct SDK: ConversationService.AddSessionTag\n\nOptions:\n")
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
	var request sdk.ConversationTagRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	err = ctx.Client.Conversations(wid).AddSessionTag(context.Background(), *sessionID, &request)
	if err != nil {
		return err
	}
	return nil
}

func runConversationServiceRemoveSessionTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation remove-session-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	sessionID := fs.Int64("session-id", 0, "session-id")
	source := fs.String("source", "", "source")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation remove-session-tag [options]\n\nProduct SDK: ConversationService.RemoveSessionTag\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	err = ctx.Client.Conversations(wid).RemoveSessionTag(context.Background(), *sessionID, *source, *name)
	if err != nil {
		return err
	}
	return nil
}

func runConversationServiceListMessageTags(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation list-message-tags", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation list-message-tags [options]\n\nProduct SDK: ConversationService.ListMessageTags\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Conversations(wid).ListMessageTags(context.Background(), *messageID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConversationServiceAddMessageTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation add-message-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation add-message-tag [options]\n\nProduct SDK: ConversationService.AddMessageTag\n\nOptions:\n")
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
	var request sdk.ConversationTagRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	err = ctx.Client.Conversations(wid).AddMessageTag(context.Background(), *messageID, &request)
	if err != nil {
		return err
	}
	return nil
}

func runConversationServiceRemoveMessageTag(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation remove-message-tag", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	source := fs.String("source", "", "source")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli conversation remove-message-tag [options]\n\nProduct SDK: ConversationService.RemoveMessageTag\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	err = ctx.Client.Conversations(wid).RemoveMessageTag(context.Background(), *messageID, *source, *name)
	if err != nil {
		return err
	}
	return nil
}

func runConversationServiceModifyResponse(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation modify-response", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	sessionID := fs.Int64("session-id", 0, "session-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	content := fs.String("content", "", "content")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.Conversations(wid).ModifyResponse(context.Background(), *sessionID, *messageID, *content)
}

func runConversationServiceAppendModifiedResponse(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("conversation append-modified-response", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	sessionID := fs.Int64("session-id", 0, "session-id")
	messageID := fs.Int64("message-id", 0, "message-id")
	content := fs.String("content", "", "content")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.Conversations(wid).AppendModifiedResponse(context.Background(), *sessionID, *messageID, *content)
}
