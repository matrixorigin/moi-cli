package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteIAMCommand runs product SDK IAMService methods via moi-cli.
func ExecuteIAMCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowIAMHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "ownership-impact":
		return runIAMServiceOwnershipImpact(ctx, args[1:])
	case "role-audit-events":
		return runIAMServiceRoleAuditEvents(ctx, args[1:])
	case "user-audit-events":
		return runIAMServiceUserAuditEvents(ctx, args[1:])
	case "set-default-role":
		return runIAMServiceSetDefaultRole(ctx, args[1:])
	case "role-members":
		return runIAMServiceRoleMembers(ctx, args[1:])
	case "role-inheritance":
		return runIAMServiceRoleInheritance(ctx, args[1:])
	case "role-inheritance-graph":
		return runIAMServiceRoleInheritanceGraph(ctx, args[1:])
	case "add-role-inheritance":
		return runIAMServiceAddRoleInheritance(ctx, args[1:])
	case "remove-role-inheritance":
		return runIAMServiceRemoveRoleInheritance(ctx, args[1:])
	case "validate-application-policy":
		return runIAMServiceValidateApplicationPolicy(ctx, args[1:])
	case "apply-application-policy":
		return runIAMServiceApplyApplicationPolicy(ctx, args[1:])
	case "current-principal-roles":
		return runIAMServiceCurrentPrincipalRoles(ctx, args[1:])
	case "current-principal-capabilities":
		return runIAMServiceCurrentPrincipalCapabilities(ctx, args[1:])
	case "permission-schema":
		return runIAMServicePermissionSchema(ctx, args[1:])
	case "roles":
		return runIAMServiceRoles(ctx, args[1:])
	case "user-role-bindings":
		return runIAMServiceUserRoleBindings(ctx, args[1:])
	case "put-user-role-bindings":
		return runIAMServicePutUserRoleBindings(ctx, args[1:])
	case "create-role":
		return runIAMServiceCreateRole(ctx, args[1:])
	case "role-permission-view":
		return runIAMServiceRolePermissionView(ctx, args[1:])
	case "update-role":
		return runIAMServiceUpdateRole(ctx, args[1:])
	case "update-role-lifecycle":
		return runIAMServiceUpdateRoleLifecycle(ctx, args[1:])
	case "role-delete-impact":
		return runIAMServiceRoleDeleteImpact(ctx, args[1:])
	case "role-delete-ownerships":
		return runIAMServiceRoleDeleteOwnerships(ctx, args[1:])
	case "delete-role":
		return runIAMServiceDeleteRole(ctx, args[1:])
	case "validate-object-grants":
		return runIAMServiceValidateObjectGrants(ctx, args[1:])
	case "apply-object-grants":
		return runIAMServiceApplyObjectGrants(ctx, args[1:])
	case "object-permission-view":
		return runIAMServiceObjectPermissionView(ctx, args[1:])
	case "transfer-object-ownership":
		return runIAMServiceTransferObjectOwnership(ctx, args[1:])
	case "list-permission-resources":
		return runIAMServiceListPermissionResources(ctx, args[1:])
	case "batch-describe-permission-resources":
		return runIAMServiceBatchDescribePermissionResources(ctx, args[1:])
	case "data-privileges":
		return runIAMServiceDataPrivileges(ctx, args[1:])
	case "validate-data-privileges":
		return runIAMServiceValidateDataPrivileges(ctx, args[1:])
	case "apply-data-privileges":
		return runIAMServiceApplyDataPrivileges(ctx, args[1:])
	case "data-privilege-operation":
		return runIAMServiceDataPrivilegeOperation(ctx, args[1:])
	case "repair-data-privileges":
		return runIAMServiceRepairDataPrivileges(ctx, args[1:])
	default:
		return fmt.Errorf("unknown iam subcommand: %s", sub)
	}
}

// ShowIAMHelp prints help for iam.
func ShowIAMHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: role-members")

	fmt.Fprintf(os.Stderr, `Usage: iam <subcommand> [options]

Product SDK surface: IAMService

Subcommands:
  ownership-impact                IAMService.OwnershipImpact
  role-audit-events               IAMService.RoleAuditEvents
  user-audit-events               IAMService.UserAuditEvents
  set-default-role                IAMService.SetDefaultRole
  current-principal-capabilities   IAMService.CurrentPrincipalCapabilities
  current-principal-roles          IAMService.CurrentPrincipalRoles
  role-inheritance                 IAMService.RoleInheritance
  role-inheritance-graph           IAMService.RoleInheritanceGraph
  add-role-inheritance             IAMService.AddRoleInheritance
  remove-role-inheritance          IAMService.RemoveRoleInheritance
  validate-application-policy      IAMService.ValidateApplicationPolicy
  apply-application-policy         IAMService.ApplyApplicationPolicy
  permission-schema                IAMService.PermissionSchema
  roles                            IAMService.Roles
  user-role-bindings               IAMService.UserRoleBindings
  put-user-role-bindings           IAMService.PutUserRoleBindings
  create-role                      IAMService.CreateRole
  role-permission-view             IAMService.RolePermissionView
  update-role                      IAMService.UpdateRole
  update-role-lifecycle            IAMService.UpdateRoleLifecycle
  role-delete-impact               IAMService.RoleDeleteImpact
  role-delete-ownerships           IAMService.RoleDeleteOwnerships
  delete-role                      IAMService.DeleteRole
  validate-object-grants           IAMService.ValidateObjectGrants
  apply-object-grants              IAMService.ApplyObjectGrants
  object-permission-view           IAMService.ObjectPermissionView
  transfer-object-ownership        IAMService.TransferObjectOwnership
  list-permission-resources        IAMService.ListPermissionResources
  batch-describe-permission-resources IAMService.BatchDescribePermissionResources
  data-privileges                  IAMService.DataPrivileges
  validate-data-privileges         IAMService.ValidateDataPrivileges
  apply-data-privileges            IAMService.ApplyDataPrivileges
  data-privilege-operation         IAMService.DataPrivilegeOperation
  repair-data-privileges           IAMService.RepairDataPrivileges

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runIAMServicePermissionSchema(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam permission-schema", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam permission-schema [options]\n\nProduct SDK: IAMService.PermissionSchema\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).PermissionSchema(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRoles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam roles", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam roles [options]\n\nProduct SDK: IAMService.Roles\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).Roles(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceUserRoleBindings(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam user-role-bindings", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	userID := fs.String("user-id", "", "Canonical IAM user ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam user-role-bindings [options]\n\nProduct SDK: IAMService.UserRoleBindings\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).UserRoleBindings(context.Background(), *userID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServicePutUserRoleBindings(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam put-user-role-bindings", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	userID := fs.String("user-id", "", "Canonical IAM user ID")
	bodyJSON := fs.String("json", "", "JSON for IAM role-binding mutation")
	bodyFile := fs.String("json-file", "", "JSON file for IAM role-binding mutation")
	useCurrentVersions := fs.Bool("use-current-versions", false, "Read current bindings and derive both expected CAS versions")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam put-user-role-bindings [options]\n\nProduct SDK: IAMService.PutUserRoleBindings\n\nOptions:\n")
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
	var request sdk.IAMUserRoleBindingsRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	if *useCurrentVersions {
		current, err := ctx.Client.IAM(wid).UserRoleBindings(context.Background(), *userID)
		if err != nil {
			return err
		}
		bindingVersion, err := sdk.IAMBindingVersionNumber(current)
		if err != nil {
			return err
		}
		request.ExpectedBindingVersion = bindingVersion
		request.ExpectedDefaultRoleVersion = current.GetDefaultRoleVersion()
	}
	result, err := ctx.Client.IAM(wid).PutUserRoleBindings(context.Background(), *userID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceCreateRole(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam create-role", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam create-role [options]\n\nProduct SDK: IAMService.CreateRole\n\nOptions:\n")
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
	var requestVal sdk.IAMRoleCreateRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &requestVal); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	request := &requestVal
	result, err := ctx.Client.IAM(wid).CreateRole(context.Background(), request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRolePermissionView(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam role-permission-view", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "roleID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam role-permission-view [options]\n\nProduct SDK: IAMService.RolePermissionView\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).RolePermissionView(context.Background(), *roleID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceUpdateRole(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam update-role", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "roleID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam update-role [options]\n\nProduct SDK: IAMService.UpdateRole\n\nOptions:\n")
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
	var requestVal sdk.IAMRoleUpdateRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &requestVal); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	request := &requestVal
	result, err := ctx.Client.IAM(wid).UpdateRole(context.Background(), *roleID, request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceUpdateRoleLifecycle(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam update-role-lifecycle", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "roleID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam update-role-lifecycle [options]\n\nProduct SDK: IAMService.UpdateRoleLifecycle\n\nOptions:\n")
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
	var requestVal sdk.IAMRoleLifecycleRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &requestVal); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	request := &requestVal
	result, err := ctx.Client.IAM(wid).UpdateRoleLifecycle(context.Background(), *roleID, request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRoleDeleteImpact(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam role-delete-impact", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "roleID")
	expectedLifecycleVersion := fs.Int64("expected-lifecycle-version", 0, "expectedLifecycleVersion")
	includeOwnership := fs.Bool("include-ownership", false, "Include the live ownership page")
	ownershipCursor := fs.String("ownership-cursor", "", "Ownership pagination cursor")
	ownershipLimit := fs.Int("ownership-limit", 50, "Ownership page size")
	transferRequested := fs.Bool("transfer-requested", false, "Validate ownership transfer before deletion")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam role-delete-impact [options]\n\nProduct SDK: IAMService.RoleDeleteImpact\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.QueryOption
	if *includeOwnership {
		opts = append(opts, sdk.WithQueryParam("include_ownership", "true"), sdk.WithQueryParam("ownership_limit", strconv.Itoa(*ownershipLimit)))
		if *ownershipCursor != "" {
			opts = append(opts, sdk.WithQueryParam("ownership_cursor", *ownershipCursor))
		}
	}
	if *transferRequested {
		opts = append(opts, sdk.WithQueryParam("transfer_requested", "true"))
	}
	result, err := ctx.Client.IAM(wid).RoleDeleteImpact(context.Background(), *roleID, *expectedLifecycleVersion, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceDeleteRole(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam delete-role", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "roleID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam delete-role [options]\n\nProduct SDK: IAMService.DeleteRole\n\nOptions:\n")
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
	var requestVal sdk.IAMRoleDeleteRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &requestVal); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	request := &requestVal
	result, err := ctx.Client.IAM(wid).DeleteRole(context.Background(), *roleID, request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRoleDeleteOwnerships(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam role-delete-ownerships", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	cursor := fs.String("cursor", "", "Pagination cursor")
	limit := fs.Int("limit", 10, "Page size")
	transferRequested := fs.Bool("transfer-requested", false, "Read objects for ownership transfer")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.QueryOption
	fs.Visit(func(current *flag.Flag) {
		switch current.Name {
		case "cursor":
			opts = append(opts, sdk.WithQueryParam("cursor", *cursor))
		case "limit":
			opts = append(opts, sdk.WithQueryParam("limit", strconv.Itoa(*limit)))
		case "transfer-requested":
			opts = append(opts, sdk.WithQueryParam("transfer_requested", strconv.FormatBool(*transferRequested)))
		}
	})
	result, err := ctx.Client.IAM(wid).RoleDeleteOwnerships(context.Background(), *roleID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceTransferObjectOwnership(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam transfer-object-ownership", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resourceType := fs.String("resource-type", "", "IAM resource type")
	resourceID := fs.String("resource-id", "", "IAM resource ID")
	bodyJSON := fs.String("json", "", "JSON for ownership transfer")
	bodyFile := fs.String("json-file", "", "JSON file for ownership transfer")
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
	var request sdk.IAMObjectOwnershipTransferRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.IAM(wid).TransferObjectOwnership(context.Background(), *resourceType, *resourceID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceValidateObjectGrants(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam validate-object-grants", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resourceType := fs.String("resource-type", "", "resourceType")
	resourceID := fs.String("resource-id", "", "resourceID")
	roleID := fs.String("role-id", "", "roleID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam validate-object-grants [options]\n\nProduct SDK: IAMService.ValidateObjectGrants\n\nOptions:\n")
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
	var requestVal sdk.IAMObjectGrantsMutationRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &requestVal); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	request := &requestVal
	result, err := ctx.Client.IAM(wid).ValidateObjectGrants(context.Background(), *resourceType, *resourceID, *roleID, request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceApplyObjectGrants(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam apply-object-grants", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resourceType := fs.String("resource-type", "", "resourceType")
	resourceID := fs.String("resource-id", "", "resourceID")
	roleID := fs.String("role-id", "", "roleID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam apply-object-grants [options]\n\nProduct SDK: IAMService.ApplyObjectGrants\n\nOptions:\n")
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
	var requestVal sdk.IAMObjectGrantsMutationRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &requestVal); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	request := &requestVal
	result, err := ctx.Client.IAM(wid).ApplyObjectGrants(context.Background(), *resourceType, *resourceID, *roleID, request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceObjectPermissionView(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam object-permission-view", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resourceType := fs.String("resource-type", "", "resourceType")
	resourceID := fs.String("resource-id", "", "resourceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam object-permission-view [options]\n\nProduct SDK: IAMService.ObjectPermissionView\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).ObjectPermissionView(context.Background(), *resourceType, *resourceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceListPermissionResources(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam list-permission-resources", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resourceType := fs.String("resource-type", "", "IAM resource type")
	actionID := fs.String("action-id", "", "IAM action ID")
	purpose := fs.String("purpose", "", "Resource picker purpose")
	parentID := fs.String("parent-id", "", "Parent resource ID")
	cursor := fs.String("cursor", "", "Pagination cursor")
	search := fs.String("q", "", "Search text")
	limit := fs.Int("limit", 0, "Page size")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam list-permission-resources [options]\n\nProduct SDK: IAMService.ListPermissionResources\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}

	var opts []sdk.QueryOption
	fs.Visit(func(current *flag.Flag) {
		switch current.Name {
		case "parent-id":
			opts = append(opts, sdk.WithQueryParam("parent_id", *parentID))
		case "cursor":
			opts = append(opts, sdk.WithQueryParam("cursor", *cursor))
		case "q":
			opts = append(opts, sdk.WithQueryParam("q", *search))
		case "limit":
			opts = append(opts, sdk.WithQueryParam("limit", strconv.Itoa(*limit)))
		}
	})
	result, err := ctx.Client.IAM(wid).ListPermissionResources(
		context.Background(), *resourceType, *actionID, *purpose, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceBatchDescribePermissionResources(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam batch-describe-permission-resources", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resourceType := fs.String("resource-type", "", "IAM resource type")
	actionID := fs.String("action-id", "", "IAM action ID")
	purpose := fs.String("purpose", "", "Resource picker purpose")
	resourceIDs := []string{}
	fs.Func("resource-id", "IAM resource ID (repeatable)", func(value string) error {
		resourceIDs = append(resourceIDs, value)
		return nil
	})
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam batch-describe-permission-resources [options]\n\nProduct SDK: IAMService.BatchDescribePermissionResources\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}

	result, err := ctx.Client.IAM(wid).BatchDescribePermissionResources(
		context.Background(), *resourceType, *actionID, *purpose, resourceIDs)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceDataPrivileges(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam data-privileges", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam data-privileges [options]\n\nProduct SDK: IAMService.DataPrivileges\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).DataPrivileges(context.Background(), *roleID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceValidateDataPrivileges(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam validate-data-privileges", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	bodyJSON := fs.String("json", "", "JSON for IAM data-privilege mutation")
	bodyFile := fs.String("json-file", "", "JSON file for IAM data-privilege mutation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam validate-data-privileges [options]\n\nProduct SDK: IAMService.ValidateDataPrivileges\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	request, err := iamDataPrivilegeMutationRequest(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).ValidateDataPrivileges(context.Background(), *roleID, request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceApplyDataPrivileges(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam apply-data-privileges", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	bodyJSON := fs.String("json", "", "JSON for IAM data-privilege mutation")
	bodyFile := fs.String("json-file", "", "JSON file for IAM data-privilege mutation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam apply-data-privileges [options]\n\nProduct SDK: IAMService.ApplyDataPrivileges\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	request, err := iamDataPrivilegeMutationRequest(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).ApplyDataPrivileges(context.Background(), *roleID, request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceDataPrivilegeOperation(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam data-privilege-operation", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	requestID := fs.String("request-id", "", "Data-privilege operation request ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam data-privilege-operation [options]\n\nProduct SDK: IAMService.DataPrivilegeOperation\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).DataPrivilegeOperation(context.Background(), *roleID, *requestID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRepairDataPrivileges(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam repair-data-privileges", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	repairOperation := fs.String("repair-operation", "", "Data-privilege repair operation ID")
	bodyJSON := fs.String("json", "", "JSON for IAM data-privilege repair")
	bodyFile := fs.String("json-file", "", "JSON file for IAM data-privilege repair")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam repair-data-privileges [options]\n\nProduct SDK: IAMService.RepairDataPrivileges\n\nOptions:\n")
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
	var request sdk.IAMDataPrivilegeRepairRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.IAM(wid).RepairDataPrivileges(context.Background(), *roleID, *repairOperation, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func iamDataPrivilegeMutationRequest(bodyJSON, bodyFile string) (*sdk.IAMDataPrivilegeMutationRequest, error) {
	rawBody, err := loadJSONBytes(bodyJSON, bodyFile)
	if err != nil {
		return nil, err
	}
	request := &sdk.IAMDataPrivilegeMutationRequest{}
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, request); err != nil {
			return nil, fmt.Errorf("parse --json: %w", err)
		}
	}
	return request, nil
}

func runIAMServiceRoleInheritance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam role-inheritance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam role-inheritance [options]\n\nProduct SDK: IAMService.RoleInheritance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).RoleInheritance(context.Background(), *roleID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRoleInheritanceGraph(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam role-inheritance-graph", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam role-inheritance-graph [options]\n\nProduct SDK: IAMService.RoleInheritanceGraph\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).RoleInheritanceGraph(context.Background(), *roleID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceAddRoleInheritance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam add-role-inheritance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	bodyJSON := fs.String("json", "", "JSON for IAM mutation")
	bodyFile := fs.String("json-file", "", "JSON file for IAM mutation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam add-role-inheritance [options]\n\nProduct SDK: IAMService.AddRoleInheritance\n\nOptions:\n")
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
	var request sdk.IAMRoleInheritanceMutationRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.IAM(wid).AddRoleInheritance(context.Background(), *roleID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRemoveRoleInheritance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam remove-role-inheritance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	bodyJSON := fs.String("json", "", "JSON for IAM mutation")
	bodyFile := fs.String("json-file", "", "JSON file for IAM mutation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam remove-role-inheritance [options]\n\nProduct SDK: IAMService.RemoveRoleInheritance\n\nOptions:\n")
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
	var request sdk.IAMRoleInheritanceMutationRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.IAM(wid).RemoveRoleInheritance(context.Background(), *roleID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceValidateApplicationPolicy(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam validate-application-policy", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	bodyJSON := fs.String("json", "", "JSON for IAM mutation")
	bodyFile := fs.String("json-file", "", "JSON file for IAM mutation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam validate-application-policy [options]\n\nProduct SDK: IAMService.ValidateApplicationPolicy\n\nOptions:\n")
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
	var request sdk.IAMApplicationPolicyRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.IAM(wid).ValidateApplicationPolicy(context.Background(), *roleID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceApplyApplicationPolicy(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam apply-application-policy", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	roleID := fs.String("role-id", "", "Canonical IAM role ID")
	bodyJSON := fs.String("json", "", "JSON for IAM mutation")
	bodyFile := fs.String("json-file", "", "JSON file for IAM mutation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam apply-application-policy [options]\n\nProduct SDK: IAMService.ApplyApplicationPolicy\n\nOptions:\n")
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
	var request sdk.IAMApplicationPolicyRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &request); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.IAM(wid).ApplyApplicationPolicy(context.Background(), *roleID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceCurrentPrincipalRoles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam current-principal-roles", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam current-principal-roles [options]\n\nProduct SDK: IAMService.CurrentPrincipalRoles\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).CurrentPrincipalRoles(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceCurrentPrincipalCapabilities(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam current-principal-capabilities", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	var capabilityIDs []string
	fs.Func("capability-id", "Capability ID to query (repeatable; omit for all; unsupported IDs are reported)", func(value string) error {
		capabilityIDs = append(capabilityIDs, value)
		return nil
	})
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli iam current-principal-capabilities [options]\n\nProduct SDK: IAMService.CurrentPrincipalCapabilities\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.IAM(wid).CurrentPrincipalCapabilities(context.Background(), capabilityIDs...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runIAMServiceRoleMembers(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("iam role-members", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	roleID := fs.String("role-id", "", "role-id")
	cursor := fs.String("cursor", "", "Pagination cursor")
	limit := fs.Int("limit", 0, "Page size")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.QueryOption
	fs.Visit(func(current *flag.Flag) {
		switch current.Name {
		case "cursor":
			opts = append(opts, sdk.WithQueryParam("cursor", *cursor))
		case "limit":
			opts = append(opts, sdk.WithQueryParam("limit", strconv.Itoa(*limit)))
		}
	})
	result, err := ctx.Client.IAM(wid).RoleMembers(context.Background(), *roleID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
