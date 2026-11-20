package commands

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func ExecuteWebappCommand(ctx *Context, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli webapp <resolve|create-session|refresh-session|create-conversation|list-conversations|rename-conversation|delete-conversation|list-messages|list-task-events|submit-feedback|call-mcp-app|read-mcp-app-resource|a2a-stream|download-task-file|preview-task-file|upload-file> [options]")
		return nil
	}
	operation := args[0]
	fs := flag.NewFlagSet("webapp "+operation, flag.ContinueOnError)
	setSDKCommandUsage(fs)
	bodyJSON := fs.String("json", "", "Request JSON object")
	bodyFile := fs.String("json-file", "", "Request JSON object file")
	conversationID := fs.String("conversation-id", "", "Public conversation ID")
	taskID := fs.String("task-id", "", "Runtime task ID")
	fileID := fs.String("file-id", "", "Runtime artifact file ID")
	filePath := fs.String("file", "", "File to upload")
	contentType := fs.String("content-type", "application/octet-stream", "Uploaded file MIME type")
	output := fs.String("output", "", "Write downloaded bytes to this file instead of stdout")
	var query queryParameters
	fs.Var(&query, "query-param", "Public webapp query parameter name=value (repeatable; includes webapp, t, s and filters)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	decode := func(target *map[string]any) error {
		if len(rawBody) == 0 {
			return sdk.ErrSpecRequired
		}
		if err := unmarshalJSONUseNumber(rawBody, target); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
		if *target == nil {
			return sdk.ErrSpecRequired
		}
		return nil
	}
	service := ctx.Client.Webapp()
	var result *sdk.ProductData
	switch operation {
	case "resolve":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.Resolve(context.Background(), request)
	case "create-session":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.CreateSession(context.Background(), request)
	case "refresh-session":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.RefreshSession(context.Background(), request)
	case "create-conversation":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.CreateConversation(context.Background(), request, query...)
	case "list-conversations":
		result, err = service.ListConversations(context.Background(), query...)
	case "rename-conversation":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.RenameConversation(context.Background(), *conversationID, request, query...)
	case "delete-conversation":
		result, err = service.DeleteConversation(context.Background(), *conversationID, query...)
	case "list-messages":
		result, err = service.ListMessages(context.Background(), *conversationID, query...)
	case "list-task-events":
		result, err = service.ListTaskEvents(context.Background(), *conversationID, *taskID, query...)
	case "submit-feedback":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.SubmitFeedback(context.Background(), request, query...)
	case "call-mcp-app":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.CallMCPApp(context.Background(), request, query...)
	case "read-mcp-app-resource":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		result, err = service.ReadMCPAppResource(context.Background(), request, query...)
	case "upload-file":
		if *filePath == "" {
			return sdk.ErrPathRequired
		}
		raw, readErr := os.ReadFile(*filePath)
		if readErr != nil {
			return readErr
		}
		result, err = service.UploadFile(context.Background(), filepath.Base(*filePath), *contentType, bytes.NewReader(raw), query...)
	case "a2a-stream":
		var request map[string]any
		if err := decode(&request); err != nil {
			return err
		}
		stream, streamErr := service.A2AStream(context.Background(), request, query...)
		err = streamErr
		if err == nil {
			defer stream.Body.Close()
			return writeStream(stream.Body, *output)
		}
	case "download-task-file", "preview-task-file":
		var stream *sdk.StreamResponse
		switch operation {
		case "download-task-file":
			stream, err = service.DownloadTaskFile(context.Background(), *conversationID, *taskID, *fileID, query...)
		case "preview-task-file":
			stream, err = service.PreviewTaskFile(context.Background(), *conversationID, *taskID, *fileID, query...)
		}
		if err == nil {
			defer stream.Body.Close()
			return writeStream(stream.Body, *output)
		}
	default:
		return fmt.Errorf("unknown public webapp operation %q", operation)
	}
	if err != nil {
		return err
	}
	return printJSON(result)
}

// ShowWebappHelp prints the public Webapp command family without creating a client.
func ShowWebappHelp(_ []string) {
	_ = ExecuteWebappCommand(nil, nil)
}
