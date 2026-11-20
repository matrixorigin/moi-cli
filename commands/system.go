package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"os"
	"strings"
)

func ExecuteSystemCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowSystemHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "service-account-assignable-roles":
		return runSystemServiceServiceAccountAssignableRoles(ctx, args[1:])
	case "service-account-workspaces":
		return runSystemServiceServiceAccountWorkspaces(ctx, args[1:])
	case "enable-compute-billing":
		return runSystemServiceEnableComputeBilling(ctx, args[1:])
	case "enable-ai-service-billing":
		return runSystemServiceEnableAIServiceBilling(ctx, args[1:])
	case "terminate-data-share-saga":
		return runSystemServiceTerminateDataShareSaga(ctx, args[1:])
	case "resume-data-share-ddl":
		return runSystemServiceResumeDataShareDDL(ctx, args[1:])
	case "health":
		return runSystemServiceHealth(ctx, args[1:])
	case "versions":
		return runSystemServiceVersions(ctx, args[1:])
	case "webapp-config":
		return runSystemServiceWebappConfig(ctx, args[1:])
	case "route-coverage":
		return runSystemServiceRouteCoverage(ctx, args[1:])
	case "metrics":
		return runSystemServiceMetrics(ctx, args[1:])
	default:
		return fmt.Errorf("unknown system subcommand: %s", sub)
	}
}
func ShowSystemHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: service-account-assignable-roles, service-account-workspaces")

	fmt.Fprintf(os.Stderr, `Usage: system <subcommand> [options]
Product SDK surface: SystemService
Subcommands:
  enable-compute-billing           SystemService.EnableComputeBilling
  enable-ai-service-billing        SystemService.EnableAIServiceBilling
  terminate-data-share-saga        SystemService.TerminateDataShareSaga
  resume-data-share-ddl            SystemService.ResumeDataShareDDL
  metrics                          SystemService.Metrics
  route-coverage                   SystemService.RouteCoverage
  webapp-config                    SystemService.WebappConfig
  versions                         SystemService.Versions
  health                           SystemService.Health
`)
}

func runSystemServiceEnableComputeBilling(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system enable-compute-billing", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system enable-compute-billing [options]\n\nProduct SDK: SystemService.EnableComputeBilling\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.System().EnableComputeBilling(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceEnableAIServiceBilling(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system enable-ai-service-billing", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system enable-ai-service-billing [options]\n\nProduct SDK: SystemService.EnableAIServiceBilling\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.System().EnableAIServiceBilling(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceHealth(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system health", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system health [options]\n\nProduct SDK: SystemService.Health\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.System().Health(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceVersions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system versions", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system versions [options]\n\nProduct SDK: SystemService.Versions\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.System().Versions(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceWebappConfig(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system webapp-config", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system webapp-config [options]\n\nProduct SDK: SystemService.WebappConfig\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.System().WebappConfig(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceRouteCoverage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system route-coverage", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system route-coverage [options]\n\nProduct SDK: SystemService.RouteCoverage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.System().RouteCoverage(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceMetrics(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system metrics", flag.ExitOnError)
	output := fs.String("output", "", "output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system metrics [options]\n\nProduct SDK: SystemService.Metrics\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.System().Metrics(context.Background())
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}

func runSystemServiceResumeDataShareDDL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system resume-data-share-ddl", flag.ExitOnError)
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system resume-data-share-ddl [options]\n\nProduct SDK: SystemService.ResumeDataShareDDL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.DataShareDDLRecoveryRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.System().ResumeDataShareDDL(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceTerminateDataShareSaga(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system terminate-data-share-saga", flag.ExitOnError)
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli system terminate-data-share-saga [options]\n\nProduct SDK: SystemService.TerminateDataShareSaga\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}

	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.DataShareSagaTerminationRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.System().TerminateDataShareSaga(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceServiceAccountWorkspaces(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system service-account-workspaces", flag.ExitOnError)
	setSDKCommandUsage(fs)
	provisionerFile := fs.String("provisioner-token-file", "", "UC provisioner bearer token file")
	actorFile := fs.String("actor-assertion-file", "", "UC actor assertion file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	client, err := serviceAccountCatalogCLIClient(ctx, *provisionerFile, *actorFile)
	if err != nil {
		return err
	}
	defer client.Close()
	result, err := client.System().ServiceAccountWorkspaces(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSystemServiceServiceAccountAssignableRoles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("system service-account-assignable-roles", flag.ExitOnError)
	setSDKCommandUsage(fs)
	targetWorkspaceID := fs.String("target-workspace-id", "", "target-workspace-id")
	provisionerFile := fs.String("provisioner-token-file", "", "UC provisioner bearer token file")
	actorFile := fs.String("actor-assertion-file", "", "UC actor assertion file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	client, err := serviceAccountCatalogCLIClient(ctx, *provisionerFile, *actorFile)
	if err != nil {
		return err
	}
	defer client.Close()
	result, err := client.System().ServiceAccountAssignableRoles(context.Background(), *targetWorkspaceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func serviceAccountCatalogCLIClient(ctx *Context, provisionerFile, actorFile string) (*sdk.Client, error) {
	provisioner, err := os.ReadFile(provisionerFile)
	if err != nil {
		return nil, err
	}
	actor, err := os.ReadFile(actorFile)
	if err != nil {
		return nil, err
	}
	return sdk.New(ctx.Endpoint, sdk.WithToken(strings.TrimSpace(string(provisioner))), sdk.WithCustomHeader("X-UC-Actor-Assertion", strings.TrimSpace(string(actor))), sdk.WithCustomHeader("X-Service-Account-Contract-Version", "1"))
}
