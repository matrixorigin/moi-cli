package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/proto"
)

// ExecuteResourceTagCommand uses the same typed tag API as Go and Python.
func ExecuteResourceTagCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowResourceTagHelp(nil)
		return nil
	}
	switch args[0] {
	case "create", "update", "create-value", "update-value", "update-relations", "groups", "resources", "assignments":
		return runResourceTagBody(ctx, args[0], args[1:])
	case "list":
		return runResourceTagList(ctx, args[1:])
	case "list-values":
		return runResourceTagListValues(ctx, args[1:])
	case "delete":
		return runResourceTagDelete(ctx, args[1:])
	case "delete-value":
		return runResourceTagDeleteValue(ctx, args[1:])
	default:
		return fmt.Errorf("unknown resource-tag subcommand: %s", args[0])
	}
}

func ShowResourceTagHelp(_ []string) {
	fmt.Fprintln(os.Stderr, `Usage: moi-cli resource-tag <subcommand> [options]

Workspace ResourceTagService:
  create, list, update, delete
  create-value, list-values, update-value, delete-value
  update-relations, groups, resources, assignments

Body commands accept --json / --json-file; inspect moi-cli schema resource-tag COMMAND.
Use value_id: null in a query filter to select unassigned resources.`)
}

func resourceTagUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli %s [options]\n\n", fs.Name())
		fs.PrintDefaults()
	}
}

func runResourceTagBody(ctx *Context, operation string, args []string) error {
	fs := flag.NewFlagSet("resource-tag "+operation, flag.ContinueOnError)
	resourceTagUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	inline := fs.String("json", "", "Typed request JSON")
	input := fs.String("json-file", "", "Typed request JSON file")
	var body proto.Message
	var service *sdk.ResourceTagService
	var invoke func() (any, error)
	switch operation {
	case "create":
		request := &sdk.ResourceTagCreateRequest{}
		body = request
		invoke = func() (any, error) { return service.Create(context.Background(), request) }
	case "update":
		tagID := fs.String("tag-id", "", "Tag ID")
		request := &sdk.ResourceTagUpdateRequest{}
		body = request
		invoke = func() (any, error) { return service.Update(context.Background(), *tagID, request) }
	case "create-value":
		tagID := fs.String("tag-id", "", "Tag ID")
		request := &sdk.ResourceTagValueRequest{}
		body = request
		invoke = func() (any, error) { return service.CreateValue(context.Background(), *tagID, request) }
	case "update-value":
		tagID := fs.String("tag-id", "", "Tag ID")
		valueID := fs.String("value-id", "", "Value ID")
		request := &sdk.ResourceTagValueRequest{}
		body = request
		invoke = func() (any, error) { return service.UpdateValue(context.Background(), *tagID, *valueID, request) }
	case "update-relations":
		request := &sdk.ResourceTagRelationsRequest{}
		body = request
		invoke = func() (any, error) { return nil, service.UpdateRelations(context.Background(), request) }
	case "groups":
		request := &sdk.ResourceTagQueryRequest{}
		body = request
		invoke = func() (any, error) { return service.Groups(context.Background(), request) }
	case "resources":
		request := &sdk.ResourceTagQueryRequest{}
		body = request
		invoke = func() (any, error) { return service.Resources(context.Background(), request) }
	case "assignments":
		request := &sdk.ResourceTagAssignmentsRequest{}
		body = request
		invoke = func() (any, error) { return service.Assignments(context.Background(), request) }
	default:
		return fmt.Errorf("unknown resource-tag subcommand: %s", operation)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*inline, *input)
	if err != nil {
		return err
	}
	if err := unmarshalProductJSON(rawBody, body); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	service = ctx.Client.ResourceTags(wid)
	result, err := invoke()
	if err != nil || result == nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func addResourceTagPageFlags(fs *flag.FlagSet, request *sdk.ResourceTagListRequest) {
	fs.Func("offset", "Result offset (default 0)", func(value string) error {
		n, err := strconv.ParseInt(value, 10, 32)
		request.Offset = int32(n)
		return err
	})
	fs.Func("limit", "Maximum results (default: product API)", func(value string) error {
		n, err := strconv.ParseInt(value, 10, 32)
		request.Limit = int32(n)
		return err
	})
}

func runResourceTagList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("resource-tag list", flag.ContinueOnError)
	resourceTagUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	var request sdk.ResourceTagListRequest
	fs.StringVar(&request.Type, "type", "", "Resource type")
	fs.StringVar(&request.NameSearch, "name-search", "", "Tag name filter")
	fs.StringVar(&request.ResourceId, "resource-id", "", "Resource ID filter")
	addResourceTagPageFlags(fs, &request)
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.ResourceTags(wid).List(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runResourceTagListValues(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("resource-tag list-values", flag.ContinueOnError)
	resourceTagUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tagID := fs.String("tag-id", "", "Tag ID")
	var request sdk.ResourceTagListRequest
	fs.StringVar(&request.ValueSearch, "value-search", "", "Tag value filter")
	addResourceTagPageFlags(fs, &request)
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.ResourceTags(wid).ListValues(context.Background(), *tagID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runResourceTagDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("resource-tag delete", flag.ContinueOnError)
	resourceTagUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tagID := fs.String("tag-id", "", "Tag ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.ResourceTags(wid).Delete(context.Background(), *tagID)
}

func runResourceTagDeleteValue(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("resource-tag delete-value", flag.ContinueOnError)
	resourceTagUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tagID := fs.String("tag-id", "", "Tag ID")
	valueID := fs.String("value-id", "", "Value ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.ResourceTags(wid).DeleteValue(context.Background(), *tagID, *valueID)
}
