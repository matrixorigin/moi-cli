package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteComputeCommand runs product SDK ComputeService methods via moi-cli.
func ExecuteComputeCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowComputeHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "list":
		return runComputeServiceList(ctx, args[1:])
	case "specs":
		return runComputeServiceSpecs(ctx, args[1:])
	case "spec-management-permission":
		return runComputeServiceSpecManagementPermission(ctx, args[1:])
	case "create-spec":
		return runComputeServiceCreateSpec(ctx, args[1:])
	case "update-spec":
		return runComputeServiceUpdateSpec(ctx, args[1:])
	case "set-spec-enabled":
		return runComputeServiceSetSpecEnabled(ctx, args[1:])
	case "create":
		return runComputeServiceCreate(ctx, args[1:])
	case "worker-images":
		return runComputeServiceWorkerImages(ctx, args[1:])
	case "get":
		return runComputeServiceGet(ctx, args[1:])
	case "update":
		return runComputeServiceUpdate(ctx, args[1:])
	case "adopt-legacy-ownership":
		return runComputeServiceAdoptLegacyOwnership(ctx, args[1:])
	case "preflight-delete":
		return runComputeServicePreflightDelete(ctx, args[1:])
	case "delete":
		return runComputeServiceDelete(ctx, args[1:])
	case "suspend":
		return runComputeServiceSuspend(ctx, args[1:])
	case "resume":
		return runComputeServiceResume(ctx, args[1:])
	case "retry":
		return runComputeServiceRetry(ctx, args[1:])
	case "runtime":
		return runComputeServiceRuntime(ctx, args[1:])
	case "workers":
		return runComputeServiceWorkers(ctx, args[1:])
	case "py-deps-limits":
		return runComputeServicePyDepsLimits(ctx, args[1:])
	case "admin-list":
		return runComputeServiceAdminList(ctx, args[1:])
	case "admin-update":
		return runComputeServiceAdminUpdate(ctx, args[1:])
	case "admin-runtime":
		return runComputeServiceAdminRuntime(ctx, args[1:])
	case "admin-preflight-delete":
		return runComputeServiceAdminPreflightDelete(ctx, args[1:])
	case "admin-delete":
		return runComputeServiceAdminDelete(ctx, args[1:])
	case "admin-suspend":
		return runComputeServiceAdminSuspend(ctx, args[1:])
	case "admin-resume":
		return runComputeServiceAdminResume(ctx, args[1:])
	case "admin-retry":
		return runComputeServiceAdminRetry(ctx, args[1:])
	default:
		return fmt.Errorf("unknown compute subcommand: %s", sub)
	}
}

// ShowComputeHelp prints help for compute.
func ShowComputeHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: compute <subcommand> [options]

Product SDK surface: ComputeService

Subcommands:
  list                             ComputeService.List
  specs                            ComputeService.Specs
  spec-management-permission       ComputeService.SpecManagementPermission
  create-spec                      ComputeService.CreateSpec
  update-spec                      ComputeService.UpdateSpec
  set-spec-enabled                 ComputeService.SetSpecEnabled
  create                           ComputeService.Create
  worker-images                    ComputeService.WorkerImages
  get                              ComputeService.Get
  update                           ComputeService.Update
  adopt-legacy-ownership           ComputeService.AdoptLegacyOwnership
  preflight-delete                 ComputeService.PreflightDelete
  delete                           ComputeService.Delete
  suspend                          ComputeService.Suspend
  resume                           ComputeService.Resume
  retry                            ComputeService.Retry
  runtime                          ComputeService.Runtime
  workers                          ComputeService.Workers
  py-deps-limits                   ComputeService.PyDepsLimits
  admin-list                       ComputeService.AdminList
  admin-update                     ComputeService.AdminUpdate
  admin-runtime                    ComputeService.AdminRuntime
  admin-preflight-delete           ComputeService.AdminPreflightDelete
  admin-delete                     ComputeService.AdminDelete
  admin-suspend                    ComputeService.AdminSuspend
  admin-resume                     ComputeService.AdminResume
  admin-retry                      ComputeService.AdminRetry

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runComputeServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute list", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute list [options]\n\nProduct SDK: ComputeService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).List(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceSpecs(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute specs", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute specs [options]\n\nProduct SDK: ComputeService.Specs\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Specs(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceSpecManagementPermission(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute spec-management-permission", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute spec-management-permission [options]\n\nProduct SDK: ComputeService.SpecManagementPermission\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).SpecManagementPermission(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceCreateSpec(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute create-spec", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	specID := fs.String("spec-id", "", "specID")
	kind := fs.String("kind", "", "kind")
	family := fs.String("family", "", "family")
	familyName := fs.String("family-name", "", "familyName")
	familyNameEN := fs.String("family-name-en", "", "familyNameEN")
	cpuMilli := fs.Int("cpu-milli", 0, "cpuMilli")
	memoryMiB := fs.Int("memory-mi-b", 0, "memoryMiB")
	addComputeSpecFlags(fs)
	nodePlacementFile := fs.String("node-placement-file", "", "JSON file containing node placement; null clears placement")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute create-spec [options]\n\nProduct SDK: ComputeService.CreateSpec\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := computeSpecCreateOptions(fs)
	if *nodePlacementFile != "" {
		placement, err := loadJSONFile[map[string]any](*nodePlacementFile)
		if err != nil {
			return fmt.Errorf("parse --node-placement-file: %w", err)
		}
		opts = append(opts, sdk.WithComputeSpecNodePlacement(placement))
	}
	result, err := ctx.Client.Compute(wid).CreateSpec(context.Background(), *specID, *kind, *family, *familyName, *familyNameEN, *cpuMilli, *memoryMiB, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceUpdateSpec(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute update-spec", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	specID := fs.String("spec-id", "", "specID")
	fs.String("kind", "", "Spec kind")
	fs.String("family", "", "Spec family")
	fs.String("family-name", "", "Spec family display name")
	fs.String("family-name-en", "", "English spec family display name")
	fs.Int("cpu-milli", 0, "CPU in millicores")
	fs.Int("memory-mib", 0, "Memory in MiB")
	addComputeSpecFlags(fs)
	nodePlacementFile := fs.String("node-placement-file", "", "JSON file containing node placement; null clears placement")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute update-spec [options]\n\nProduct SDK: ComputeService.UpdateSpec\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := computeSpecUpdateOptions(fs)
	if *nodePlacementFile != "" {
		placement, err := loadJSONFile[map[string]any](*nodePlacementFile)
		if err != nil {
			return fmt.Errorf("parse --node-placement-file: %w", err)
		}
		opts = append(opts, sdk.WithComputeSpecUpdateNodePlacement(placement))
	}
	result, err := ctx.Client.Compute(wid).UpdateSpec(context.Background(), *specID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceSetSpecEnabled(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute set-spec-enabled", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	specID := fs.String("spec-id", "", "specID")
	enabled := fs.Bool("enabled", false, "enabled")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute set-spec-enabled [options]\n\nProduct SDK: ComputeService.SetSpecEnabled\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).SetSpecEnabled(context.Background(), *specID, *enabled)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute create", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	specID := fs.String("spec-id", "", "specID")
	cpu := fs.Int("cpu", 0, "cpu")
	memoryGiB := fs.Int("memory-gi-b", 0, "memoryGiB")
	fs.String("kind", "", "Compute resource kind")
	fs.String("description", "", "Compute resource description")
	fs.Int("gpu", 0, "GPU count")
	fs.Int("min-replicas", 0, "Minimum replicas")
	fs.Int("max-replicas", 0, "Maximum replicas")
	fs.Int("auto-suspend-minutes", 0, "Idle timeout in minutes (omit for the SDK default)")
	fs.String("go-worker-image-id", "", "Go worker image ID")
	fs.String("python-worker-image-id", "", "Python worker image ID")
	fs.String("platform", "", "Worker platform")
	fs.String("py-requirements-txt", "", "Python worker requirements.txt contents")
	workerImagesFile := fs.String("worker-images-file", "", "JSON file containing worker image selections")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute create [options]\n\nProduct SDK: ComputeService.Create\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts, err := computeCreateOptions(fs, *workerImagesFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Create(context.Background(), *name, *specID, *cpu, *memoryGiB, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceWorkerImages(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute worker-images", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute worker-images [options]\n\nProduct SDK: ComputeService.WorkerImages\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).WorkerImages(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute get", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute get [options]\n\nProduct SDK: ComputeService.Get\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Get(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	addComputeResourceUpdateFlags(fs)
	workerImagesFile := fs.String("worker-images-file", "", "JSON file containing worker image selections; [] clears selection")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute update [options]\n\nProduct SDK: ComputeService.Update\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts, err := computeUpdateOptions(fs, *workerImagesFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Update(context.Background(), *id, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdoptLegacyOwnership(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute adopt-legacy-ownership", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "Compute resource ID (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute adopt-legacy-ownership [options]\\n\\nProduct SDK: ComputeService.AdoptLegacyOwnership\\n\\nOptions:\\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("--id is required")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdoptLegacyOwnership(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServicePreflightDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute preflight-delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute preflight-delete [options]\n\nProduct SDK: ComputeService.PreflightDelete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).PreflightDelete(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute delete", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute delete [options]\n\nProduct SDK: ComputeService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Compute(wid).Delete(context.Background(), *id, queryParams...); err != nil {
		return err
	}
	fmt.Printf("compute delete: ok\n")
	return nil
}

func runComputeServiceSuspend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute suspend", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute suspend [options]\n\nProduct SDK: ComputeService.Suspend\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Suspend(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceResume(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute resume", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute resume [options]\n\nProduct SDK: ComputeService.Resume\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Resume(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceRetry(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute retry", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute retry [options]\n\nProduct SDK: ComputeService.Retry\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Retry(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceRuntime(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute runtime", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	computeResourceID := fs.String("compute-resource-id", "", "Compute resource ID (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute runtime [options]\\n\\nProduct SDK: ComputeService.Runtime\\n\\nOptions:\\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *computeResourceID == "" {
		return fmt.Errorf("--compute-resource-id is required")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Runtime(context.Background(), *computeResourceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceWorkers(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute workers", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	computeResourceID := fs.String("compute-resource-id", "", "Compute resource ID (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute workers [options]\n\nProduct SDK: ComputeService.Workers\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *computeResourceID == "" {
		return fmt.Errorf("--compute-resource-id is required")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).Workers(context.Background(), *computeResourceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServicePyDepsLimits(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute py-deps-limits", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute py-deps-limits [options]\n\nProduct SDK: ComputeService.PyDepsLimits\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).PyDepsLimits(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdminList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-list", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-list [options]\n\nProduct SDK: ComputeService.AdminList\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdminList(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdminUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "Compute resource ID (required)")
	addComputeResourceUpdateFlags(fs)
	workerImagesFile := fs.String("worker-images-file", "", "JSON file containing worker image selections; [] clears selection")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-update [options]\n\nProduct SDK: ComputeService.AdminUpdate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *id == "" {
		return fmt.Errorf("--id is required")
	}
	opts, err := computeUpdateOptions(fs, *workerImagesFile)
	if err != nil {
		return err
	}
	if len(opts) == 0 {
		return fmt.Errorf("at least one compute update option is required")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdminUpdate(context.Background(), *id, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdminRuntime(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-runtime", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	computeResourceID := fs.String("compute-resource-id", "", "Compute resource ID (required)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-runtime [options]\\n\\nProduct SDK: ComputeService.AdminRuntime\\n\\nOptions:\\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *computeResourceID == "" {
		return fmt.Errorf("--compute-resource-id is required")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdminRuntime(context.Background(), *computeResourceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdminPreflightDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-preflight-delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-preflight-delete [options]\n\nProduct SDK: ComputeService.AdminPreflightDelete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdminPreflightDelete(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdminDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-delete", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-delete [options]\n\nProduct SDK: ComputeService.AdminDelete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Compute(wid).AdminDelete(context.Background(), *id, queryParams...); err != nil {
		return err
	}
	fmt.Printf("compute admin-delete: ok\n")
	return nil
}

func runComputeServiceAdminSuspend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-suspend", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-suspend [options]\n\nProduct SDK: ComputeService.AdminSuspend\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdminSuspend(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdminResume(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-resume", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-resume [options]\n\nProduct SDK: ComputeService.AdminResume\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdminResume(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runComputeServiceAdminRetry(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("compute admin-retry", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli compute admin-retry [options]\n\nProduct SDK: ComputeService.AdminRetry\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Compute(wid).AdminRetry(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

// computeUpdateOptions shares the native update contract across workspace and
// administrator callers while preserving explicit zero and empty values.
func computeUpdateOptions(fs *flag.FlagSet, workerImagesFile string) ([]sdk.ComputeUpdateOption, error) {
	var opts []sdk.ComputeUpdateOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "kind":
			opts = append(opts, sdk.WithComputeUpdateKind(f.Value.String()))
		case "py-requirements-txt":
			opts = append(opts, sdk.WithComputeUpdatePyRequirementsTxt(f.Value.String()))
		case "name":
			opts = append(opts, sdk.WithComputeUpdateName(f.Value.String()))
		case "description":
			opts = append(opts, sdk.WithComputeUpdateDescription(f.Value.String()))
		case "spec-id":
			opts = append(opts, sdk.WithComputeUpdateSpecID(f.Value.String()))
		case "cpu":
			opts = append(opts, sdk.WithComputeUpdateCPU(f.Value.(flag.Getter).Get().(int)))
		case "memory-gib":
			opts = append(opts, sdk.WithComputeUpdateMemoryGiB(f.Value.(flag.Getter).Get().(int)))
		case "gpu":
			opts = append(opts, sdk.WithComputeUpdateGPU(f.Value.(flag.Getter).Get().(int)))
		case "min-replicas":
			opts = append(opts, sdk.WithComputeUpdateMinReplicas(f.Value.(flag.Getter).Get().(int)))
		case "max-replicas":
			opts = append(opts, sdk.WithComputeUpdateMaxReplicas(f.Value.(flag.Getter).Get().(int)))
		case "go-worker-image-id":
			opts = append(opts, sdk.WithComputeUpdateGoWorkerImageID(f.Value.String()))
		case "python-worker-image-id":
			opts = append(opts, sdk.WithComputeUpdatePythonWorkerImageID(f.Value.String()))
		case "auto-suspend-minutes":
			opts = append(opts, sdk.WithComputeUpdateAutoSuspendMinutes(f.Value.(flag.Getter).Get().(int)))
		case "platform":
			opts = append(opts, sdk.WithComputeUpdatePlatform(f.Value.String()))
		}
	})
	if workerImagesFile != "" {
		images, err := loadJSONFile[[]*sdk.WorkerImageSelection](workerImagesFile)
		if err != nil {
			return nil, fmt.Errorf("parse --worker-images-file: %w", err)
		}
		opts = append(opts, sdk.WithComputeUpdateWorkerImages(images...))
	}
	return opts, nil
}

func addComputeResourceUpdateFlags(fs *flag.FlagSet) {
	fs.String("name", "", "Compute resource name")
	fs.String("kind", "", "Compute resource kind")
	fs.String("description", "", "Compute resource description")
	fs.String("spec-id", "", "Compute resource spec ID")
	fs.Int("cpu", 0, "CPU cores")
	fs.Int("memory-gib", 0, "Memory in GiB")
	fs.Int("gpu", 0, "GPU count")
	fs.Int("min-replicas", 0, "Minimum replicas")
	fs.Int("max-replicas", 0, "Maximum replicas")
	fs.String("go-worker-image-id", "", "Go worker image ID")
	fs.String("python-worker-image-id", "", "Python worker image ID")
	fs.String("platform", "", "Worker platform")
	fs.Int("auto-suspend-minutes", 0, "Idle timeout in minutes (only submitted when specified)")
	fs.String("py-requirements-txt", "", "Python worker requirements.txt contents; empty clears dependencies")
}

func addComputeSpecFlags(fs *flag.FlagSet) {
	fs.Int("gpu-count", 0, "GPU count")
	fs.Int("gpu-memory-mib", 0, "GPU memory in MiB")
	fs.Int("gpu-cores", 0, "GPU cores")
	fs.Float64("credit-per-hour", 0, "Credits per hour")
	fs.String("description", "", "Spec description")
	fs.Bool("enabled", false, "Enabled value; use set-spec-enabled to disable an existing spec")
}

func computeSpecCreateOptions(fs *flag.FlagSet) []sdk.ComputeSpecOption {
	var opts []sdk.ComputeSpecOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "gpu-count":
			opts = append(opts, sdk.WithComputeSpecGPUCount(f.Value.(flag.Getter).Get().(int)))
		case "gpu-memory-mib":
			opts = append(opts, sdk.WithComputeSpecGPUMemoryMiB(f.Value.(flag.Getter).Get().(int)))
		case "gpu-cores":
			opts = append(opts, sdk.WithComputeSpecGPUCores(f.Value.(flag.Getter).Get().(int)))
		case "credit-per-hour":
			opts = append(opts, sdk.WithComputeSpecCreditPerHour(f.Value.(flag.Getter).Get().(float64)))
		case "description":
			opts = append(opts, sdk.WithComputeSpecDescription(f.Value.String()))
		case "enabled":
			opts = append(opts, sdk.WithComputeSpecEnabled(f.Value.(flag.Getter).Get().(bool)))
		}
	})
	return opts
}

func computeSpecUpdateOptions(fs *flag.FlagSet) []sdk.ComputeSpecUpdateOption {
	var opts []sdk.ComputeSpecUpdateOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "kind":
			opts = append(opts, sdk.WithComputeSpecUpdateKind(f.Value.String()))
		case "family":
			opts = append(opts, sdk.WithComputeSpecUpdateFamily(f.Value.String()))
		case "family-name":
			opts = append(opts, sdk.WithComputeSpecUpdateFamilyName(f.Value.String()))
		case "family-name-en":
			opts = append(opts, sdk.WithComputeSpecUpdateFamilyNameEN(f.Value.String()))
		case "cpu-milli":
			opts = append(opts, sdk.WithComputeSpecUpdateCPUMilli(f.Value.(flag.Getter).Get().(int)))
		case "memory-mib":
			opts = append(opts, sdk.WithComputeSpecUpdateMemoryMiB(f.Value.(flag.Getter).Get().(int)))
		case "gpu-count":
			opts = append(opts, sdk.WithComputeSpecUpdateGPUCount(f.Value.(flag.Getter).Get().(int)))
		case "gpu-memory-mib":
			opts = append(opts, sdk.WithComputeSpecUpdateGPUMemoryMiB(f.Value.(flag.Getter).Get().(int)))
		case "gpu-cores":
			opts = append(opts, sdk.WithComputeSpecUpdateGPUCores(f.Value.(flag.Getter).Get().(int)))
		case "credit-per-hour":
			opts = append(opts, sdk.WithComputeSpecUpdateCreditPerHour(f.Value.(flag.Getter).Get().(float64)))
		case "description":
			opts = append(opts, sdk.WithComputeSpecUpdateDescription(f.Value.String()))
		case "enabled":
			opts = append(opts, sdk.WithComputeSpecUpdateEnabled(f.Value.(flag.Getter).Get().(bool)))
		}
	})
	return opts
}

func computeCreateOptions(fs *flag.FlagSet, workerImagesFile string) ([]sdk.ComputeCreateOption, error) {
	var opts []sdk.ComputeCreateOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "kind":
			opts = append(opts, sdk.WithComputeKind(f.Value.String()))
		case "description":
			opts = append(opts, sdk.WithComputeDescription(f.Value.String()))
		case "gpu":
			opts = append(opts, sdk.WithComputeGPU(f.Value.(flag.Getter).Get().(int)))
		case "min-replicas":
			opts = append(opts, sdk.WithComputeMinReplicas(f.Value.(flag.Getter).Get().(int)))
		case "max-replicas":
			opts = append(opts, sdk.WithComputeMaxReplicas(f.Value.(flag.Getter).Get().(int)))
		case "auto-suspend-minutes":
			opts = append(opts, sdk.WithComputeAutoSuspendMinutes(f.Value.(flag.Getter).Get().(int)))
		case "go-worker-image-id":
			opts = append(opts, sdk.WithComputeGoWorkerImageID(f.Value.String()))
		case "python-worker-image-id":
			opts = append(opts, sdk.WithComputePythonWorkerImageID(f.Value.String()))
		case "platform":
			opts = append(opts, sdk.WithComputePlatform(f.Value.String()))
		case "py-requirements-txt":
			opts = append(opts, sdk.WithComputePyRequirementsTxt(f.Value.String()))
		}
	})
	if workerImagesFile != "" {
		images, err := loadJSONFile[[]*sdk.WorkerImageSelection](workerImagesFile)
		if err != nil {
			return nil, fmt.Errorf("parse --worker-images-file: %w", err)
		}
		opts = append(opts, sdk.WithComputeWorkerImages(images...))
	}
	return opts, nil
}
