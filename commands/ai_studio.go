package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func ExecuteAIStudioCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowAIStudioHelp(nil)
		return nil
	}
	switch args[0] {
	case "chat-model-policy":
		return runAIStudioChatModelPolicy(ctx, args[1:])
	case "service-slots":
		return runAIStudioServiceSlots(ctx, args[1:])
	case "service-slot-candidates":
		return runAIStudioServiceSlotCandidates(ctx, args[1:])
	case "service-slot-logs":
		return runAIStudioServiceSlotLogs(ctx, args[1:])
	case "put-service-slot":
		return runAIStudioPutServiceSlot(ctx, args[1:])
	default:
		return fmt.Errorf("unknown ai-studio subcommand: %s", args[0])
	}
}

func ShowAIStudioHelp(_ []string) {
	fmt.Fprintln(os.Stderr, `Usage: moi-cli ai-studio <subcommand> [options]
chat-model-policy
service-slots, service-slot-candidates, service-slot-logs
put-service-slot --slot-id <id> --json-file <request.json>

These are account-level routes and never add a workspace header.
Service-slot administration requires the server's Genesis administrator admission.`)
}

func runAIStudioChatModelPolicy(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("ai-studio chat-model-policy", flag.ContinueOnError)
	setSDKCommandUsage(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.AIStudio().ChatModelPolicy(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAIStudioServiceSlots(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("ai-studio service-slots", flag.ContinueOnError)
	setSDKCommandUsage(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.AIStudio().ServiceSlots(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAIStudioServiceSlotCandidates(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("ai-studio service-slot-candidates", flag.ContinueOnError)
	setSDKCommandUsage(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.AIStudio().ServiceSlotCandidates(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAIStudioServiceSlotLogs(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("ai-studio service-slot-logs", flag.ContinueOnError)
	slotID := fs.String("slot-id", "", "Filter operation logs by service slot ID")
	pageSize := fs.Int("page-size", 0, "Operation log page size (omitted unless supplied)")
	setSDKCommandUsage(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	var options []sdk.QueryOption
	fs.Visit(func(current *flag.Flag) {
		switch current.Name {
		case "slot-id":
			options = append(options, sdk.WithServiceSlotLogSlotID(*slotID))
		case "page-size":
			options = append(options, sdk.WithServiceSlotLogPageSize(*pageSize))
		}
	})
	result, err := ctx.Client.AIStudio().ServiceSlotLogs(context.Background(), options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAIStudioPutServiceSlot(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("ai-studio put-service-slot", flag.ContinueOnError)
	slotID := fs.String("slot-id", "", "Service slot ID")
	inline := fs.String("json", "", "Service slot request JSON")
	input := fs.String("json-file", "", "Service slot request JSON file")
	setSDKCommandUsage(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*inline, *input)
	if err != nil {
		return err
	}
	var request sdk.PutAIStudioServiceSlotRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.AIStudio().PutServiceSlot(context.Background(), *slotID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
