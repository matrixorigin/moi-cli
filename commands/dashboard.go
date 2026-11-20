package commands

import (
	sdk "github.com/matrixorigin/moi-go-sdk/v5"

	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// ExecuteDashboardCommand runs product SDK DashboardService methods via moi-cli.
func ExecuteDashboardCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowDashboardHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "connect-web-socket":
		return runDashboardServiceConnectWebSocket(ctx, args[1:])
	case "download-drill-through":
		return runDashboardServiceDownloadDrillThrough(ctx, args[1:])
	case "drill-through":
		return runDashboardServiceDrillThrough(ctx, args[1:])
	case "download-chart-result":
		return runDashboardServiceDownloadChartResult(ctx, args[1:])
	case "chart-result":
		return runDashboardServiceChartResult(ctx, args[1:])
	case "run-chart":
		return runDashboardServiceRunChart(ctx, args[1:])
	case "refresh":
		return runDashboardServiceRefresh(ctx, args[1:])
	case "generate-sql":
		return runDashboardServiceGenerateSQL(ctx, args[1:])
	case "preview-sql":
		return runDashboardServicePreviewSQL(ctx, args[1:])
	case "update-chart-order":
		return runDashboardServiceUpdateChartOrder(ctx, args[1:])
	case "delete-chart":
		return runDashboardServiceDeleteChart(ctx, args[1:])
	case "update-chart":
		return runDashboardServiceUpdateChart(ctx, args[1:])
	case "create-chart":
		return runDashboardServiceCreateChart(ctx, args[1:])
	case "delete":
		return runDashboardServiceDelete(ctx, args[1:])
	case "update":
		return runDashboardServiceUpdate(ctx, args[1:])
	case "get":
		return runDashboardServiceGet(ctx, args[1:])
	case "list":
		return runDashboardServiceList(ctx, args[1:])
	case "create":
		return runDashboardServiceCreate(ctx, args[1:])
	case "task-execution-overview":
		return runDashboardServiceTaskExecutionOverview(ctx, args[1:])
	default:
		return fmt.Errorf("unknown dashboard subcommand: %s", sub)
	}
}

// ShowDashboardHelp prints help for dashboard.
func ShowDashboardHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: dashboard <subcommand> [options]

Product SDK surface: DashboardService

Subcommands:
  connect-web-socket              DashboardService.ConnectWebSocket
  download-drill-through           DashboardService.DownloadDrillThrough
  drill-through                    DashboardService.DrillThrough
  download-chart-result            DashboardService.DownloadChartResult
  chart-result                     DashboardService.ChartResult
  run-chart                        DashboardService.RunChart
  refresh                          DashboardService.Refresh
  generate-sql                     DashboardService.GenerateSQL
  preview-sql                      DashboardService.PreviewSQL
  update-chart-order               DashboardService.UpdateChartOrder
  delete-chart                     DashboardService.DeleteChart
  update-chart                     DashboardService.UpdateChart
  create-chart                     DashboardService.CreateChart
  delete                           DashboardService.Delete
  update                           DashboardService.Update
  get                              DashboardService.Get
  list                             DashboardService.List
  create                           DashboardService.Create
  task-execution-overview          DashboardService.TaskExecutionOverview

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runDashboardServiceTaskExecutionOverview(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard task-execution-overview", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard task-execution-overview [options]\n\nProduct SDK: DashboardService.TaskExecutionOverview\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).TaskExecutionOverview(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
func runDashboardServiceCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard create", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard create [options]\n\nProduct SDK: DashboardService.Create\n\nOptions:\n")
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
	var request sdk.DataDashboardMutationRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).Create(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard list", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard list [options]\n\nProduct SDK: DashboardService.List\n\nOptions:\n")
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
	var request sdk.DataDashboardListRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).List(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard get", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard get [options]\n\nProduct SDK: DashboardService.Get\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).Get(context.Background(), *dashboardID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard update [options]\n\nProduct SDK: DashboardService.Update\n\nOptions:\n")
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
	var request sdk.DataDashboardMutationRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).Update(context.Background(), *dashboardID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard delete [options]\n\nProduct SDK: DashboardService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).Delete(context.Background(), *dashboardID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceCreateChart(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard create-chart", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	bodyJSON := fs.String("json", "", "Request JSON; inspect chart_config with: moi-cli schema dashboard create-chart --argument json.chart_config; layout_config with --argument json.layout_config")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard create-chart [options]\n\nProduct SDK: DashboardService.CreateChart\n\nOptions:\n")
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
	var request sdk.DataDashboardChartMutationRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).CreateChart(context.Background(), *dashboardID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceUpdateChart(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard update-chart", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	chartID := fs.String("chart-id", "", "chart-id")
	bodyJSON := fs.String("json", "", "Request JSON; inspect chart_config with: moi-cli schema dashboard update-chart --argument json.chart_config; layout_config with --argument json.layout_config")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard update-chart [options]\n\nProduct SDK: DashboardService.UpdateChart\n\nOptions:\n")
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
	var request sdk.DataDashboardChartMutationRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).UpdateChart(context.Background(), *dashboardID, *chartID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceDeleteChart(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard delete-chart", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	chartID := fs.String("chart-id", "", "chart-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard delete-chart [options]\n\nProduct SDK: DashboardService.DeleteChart\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).DeleteChart(context.Background(), *dashboardID, *chartID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceUpdateChartOrder(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard update-chart-order", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard update-chart-order [options]\n\nProduct SDK: DashboardService.UpdateChartOrder\n\nOptions:\n")
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
	var request sdk.DataDashboardChartOrderRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).UpdateChartOrder(context.Background(), *dashboardID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServicePreviewSQL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard preview-sql", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	sqlText := fs.String("sql-text", "", "sql-text")
	chartID := fs.String("chart-id", "", "Optional chart ID for persisted SQL/database scope")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard preview-sql [options]\n\nProduct SDK: DashboardService.PreviewSQL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.DashboardPreviewOption
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "chart-id" {
			options = append(options, sdk.WithDashboardPreviewChartID(*chartID))
		}
	})
	result, err := ctx.Client.Dashboard(wid).PreviewSQL(context.Background(), *dashboardID, *sqlText, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceGenerateSQL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard generate-sql", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	question := fs.String("question", "", "question")
	title := fs.String("title", "", "title")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard generate-sql [options]\n\nProduct SDK: DashboardService.GenerateSQL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).GenerateSQL(context.Background(), *dashboardID, *question, *title)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceRefresh(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard refresh", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard refresh [options]\n\nProduct SDK: DashboardService.Refresh\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).Refresh(context.Background(), *dashboardID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceRunChart(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard run-chart", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	chartID := fs.String("chart-id", "", "chart-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard run-chart [options]\n\nProduct SDK: DashboardService.RunChart\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).RunChart(context.Background(), *dashboardID, *chartID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceChartResult(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard chart-result", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value, including at=<RFC3339Nano> (repeatable)")
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	chartID := fs.String("chart-id", "", "chart-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard chart-result [options]\n\nProduct SDK: DashboardService.ChartResult\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).ChartResult(context.Background(), *dashboardID, *chartID, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceDownloadChartResult(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard download-chart-result", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value, including at=<RFC3339Nano> (repeatable)")
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	chartID := fs.String("chart-id", "", "chart-id")
	output := fs.String("output", "", "output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard download-chart-result [options]\n\nProduct SDK: DashboardService.DownloadChartResult\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Dashboard(wid).DownloadChartResult(context.Background(), *dashboardID, *chartID, queryParams...)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}

func runDashboardServiceDrillThrough(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard drill-through", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	chartID := fs.String("chart-id", "", "chart-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard drill-through [options]\n\nProduct SDK: DashboardService.DrillThrough\n\nOptions:\n")
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
	var request sdk.DataDashboardDrillThroughRequest
	if err := unmarshalProductJSON(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).DrillThrough(context.Background(), *dashboardID, *chartID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDashboardServiceDownloadDrillThrough(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dashboard download-drill-through", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	dashboardID := fs.String("dashboard-id", "", "dashboard-id")
	chartID := fs.String("chart-id", "", "chart-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	output := fs.String("output", "", "output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli dashboard download-drill-through [options]\n\nProduct SDK: DashboardService.DownloadDrillThrough\n\nOptions:\n")
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
	var request sdk.DataDashboardDrillThroughRequest
	if err := unmarshalProductJSON(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Dashboard(wid).DownloadDrillThrough(context.Background(), *dashboardID, *chartID, &request)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}
