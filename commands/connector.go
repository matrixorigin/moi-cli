package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteConnectorCommand runs product SDK ConnectorService methods via moi-cli.
func ExecuteConnectorCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowConnectorHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "list":
		return runConnectorServiceList(ctx, args[1:])
	case "get":
		return runConnectorServiceGet(ctx, args[1:])
	case "create":
		return runConnectorServiceCreate(ctx, args[1:])
	case "update":
		return runConnectorServiceUpdate(ctx, args[1:])
	case "delete":
		return runConnectorServiceDelete(ctx, args[1:])
	case "validate":
		return runConnectorServiceValidate(ctx, args[1:])
	case "probe-rest-draft":
		return runConnectorServiceProbeRESTDraft(ctx, args[1:])
	case "probe-rest":
		return runConnectorServiceProbeREST(ctx, args[1:])
	case "preview-rest":
		return runConnectorServicePreviewREST(ctx, args[1:])
	case "files":
		return runConnectorServiceFiles(ctx, args[1:])
	case "upload":
		return runConnectorServiceUpload(ctx, args[1:])
	case "local-upload":
		return runConnectorServiceLocalUpload(ctx, args[1:])
	case "preview-file":
		return runConnectorServicePreviewFile(ctx, args[1:])
	case "db-list-databases":
		return runConnectorServiceDBListDatabases(ctx, args[1:])
	case "db-list-schemas":
		return runConnectorServiceDBListSchemas(ctx, args[1:])
	case "db-list-tables":
		return runConnectorServiceDBListTables(ctx, args[1:])
	case "db-source-schema":
		return runConnectorServiceDBSourceSchema(ctx, args[1:])
	case "structured-sources":
		return runConnectorServiceStructuredSources(ctx, args[1:])
	case "structured-source-databases":
		return runConnectorServiceStructuredSourceDatabases(ctx, args[1:])
	case "structured-source-schemas":
		return runConnectorServiceStructuredSourceSchemas(ctx, args[1:])
	case "structured-source-objects":
		return runConnectorServiceStructuredSourceObjects(ctx, args[1:])
	case "structured-source-schema":
		return runConnectorServiceStructuredSourceSchema(ctx, args[1:])
	case "structured-source-capability":
		return runConnectorServiceStructuredSourceCapability(ctx, args[1:])
	case "structured-source-sample":
		return runConnectorServiceStructuredSourceSample(ctx, args[1:])
	case "structured-backfill-start-resolve":
		return runConnectorServiceStructuredBackfillStartResolve(ctx, args[1:])
	case "file-download":
		return runConnectorServiceFileDownload(ctx, args[1:])
	case "file-stream":
		return runConnectorServiceFileStream(ctx, args[1:])
	case "file-delete":
		return runConnectorServiceFileDelete(ctx, args[1:])
	case "mo-database-list":
		return runConnectorServiceMODatabaseList(ctx, args[1:])
	case "mo-database-table-list":
		return runConnectorServiceMODatabaseTableList(ctx, args[1:])
	case "mo-check-table-name":
		return runConnectorServiceMOCheckTableName(ctx, args[1:])
	case "mo-column-mapping":
		return runConnectorServiceMOColumnMapping(ctx, args[1:])
	case "task-stats":
		return runConnectorServiceTaskStats(ctx, args[1:])
	case "tasks":
		return runConnectorServiceTasks(ctx, args[1:])
	case "create-task":
		return runConnectorServiceCreateTask(ctx, args[1:])
	case "update-task":
		return runConnectorServiceUpdateTask(ctx, args[1:])
	case "update-task-from-spec":
		return runConnectorServiceUpdateTaskFromSpec(ctx, args[1:])
	case "delete-task":
		return runConnectorServiceDeleteTask(ctx, args[1:])
	case "pause-task":
		return runConnectorServicePauseTask(ctx, args[1:])
	case "resume-task":
		return runConnectorServiceResumeTask(ctx, args[1:])
	case "run-task-now":
		return runConnectorServiceRunTaskNow(ctx, args[1:])
	case "run-task-full-refresh":
		return runConnectorServiceRunTaskFullRefresh(ctx, args[1:])
	case "get-task":
		return runConnectorServiceGetTask(ctx, args[1:])
	case "task-files":
		return runConnectorServiceTaskFiles(ctx, args[1:])
	case "task-runs":
		return runConnectorServiceTaskRuns(ctx, args[1:])
	case "retry-task-files":
		return runConnectorServiceRetryTaskFiles(ctx, args[1:])
	case "export-task-stats":
		return runConnectorServiceExportTaskStats(ctx, args[1:])
	case "export-tasks":
		return runConnectorServiceExportTasks(ctx, args[1:])
	case "export-task-info":
		return runConnectorServiceExportTaskInfo(ctx, args[1:])
	case "export-task-files":
		return runConnectorServiceExportTaskFiles(ctx, args[1:])
	case "delete-export-task":
		return runConnectorServiceDeleteExportTask(ctx, args[1:])
	case "create-export-task":
		return runConnectorServiceCreateExportTask(ctx, args[1:])
	case "retry-export-task":
		return runConnectorServiceRetryExportTask(ctx, args[1:])
	case "export-task-state":
		return runConnectorServiceExportTaskState(ctx, args[1:])
	default:
		return fmt.Errorf("unknown connector subcommand: %s", sub)
	}
}

// ShowConnectorHelp prints help for connector.
func ShowConnectorHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: connector <subcommand> [options]

Product SDK surface: ConnectorService

Subcommands:
  list                             ConnectorService.List
  get                              ConnectorService.Get
  create                           ConnectorService.Create
  update                           ConnectorService.Update
  delete                           ConnectorService.Delete
  validate                         ConnectorService.Validate
  probe-rest-draft                 ConnectorService.ProbeRESTDraft
  probe-rest                       ConnectorService.ProbeREST
  preview-rest                     ConnectorService.PreviewREST
  files                            ConnectorService.Files
  upload                           ConnectorService.Upload
  local-upload                     ConnectorService.LocalUpload
  preview-file                     ConnectorService.PreviewFile
  db-list-databases                ConnectorService.DBListDatabases
  db-list-schemas                  ConnectorService.DBListSchemas
  db-list-tables                   ConnectorService.DBListTables
  db-source-schema                 ConnectorService.DBSourceSchema
  structured-sources               ConnectorService.StructuredSources
  structured-source-databases      ConnectorService.StructuredSourceDatabases
  structured-source-schemas        ConnectorService.StructuredSourceSchemas
  structured-source-objects        ConnectorService.StructuredSourceObjects
  structured-source-schema         ConnectorService.StructuredSourceSchema
  structured-source-capability     ConnectorService.StructuredSourceCapability
  structured-source-sample         ConnectorService.StructuredSourceSample
  structured-backfill-start-resolve ConnectorService.StructuredBackfillStartResolve
  file-download                    ConnectorService.FileDownload
  file-stream                      ConnectorService.FileStream
  file-delete                      ConnectorService.FileDelete
  mo-database-list                 ConnectorService.MODatabaseList
  mo-database-table-list           ConnectorService.MODatabaseTableList
  mo-check-table-name              ConnectorService.MOCheckTableName
  mo-column-mapping                ConnectorService.MOColumnMapping
  task-stats                       ConnectorService.TaskStats
  tasks                            ConnectorService.Tasks
  create-task                      ConnectorService.CreateTask
  update-task                      ConnectorService.UpdateTask
  update-task-from-spec            ConnectorService.UpdateTaskFromSpec
  delete-task                      ConnectorService.DeleteTask
  pause-task                       ConnectorService.PauseTask
  resume-task                      ConnectorService.ResumeTask
  run-task-now                     ConnectorService.RunTaskNow
  run-task-full-refresh            ConnectorService.RunTaskFullRefresh
  get-task                         ConnectorService.GetTask
  task-files                       ConnectorService.TaskFiles
  task-runs                        ConnectorService.TaskRuns
  retry-task-files                 ConnectorService.RetryTaskFiles
  export-task-stats                ConnectorService.ExportTaskStats
  export-tasks                     ConnectorService.ExportTasks
  export-task-info                 ConnectorService.ExportTaskInfo
  export-task-files                ConnectorService.ExportTaskFiles
  delete-export-task               ConnectorService.DeleteExportTask
  create-export-task               ConnectorService.CreateExportTask
  retry-export-task                ConnectorService.RetryExportTask
  export-task-state                ConnectorService.ExportTaskState

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runConnectorServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector list", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector list [options]\n\nProduct SDK: ConnectorService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).List(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector get", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector get [options]\n\nProduct SDK: ConnectorService.Get\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).Get(context.Background(), *connectorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector create", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	sourceType := fs.Int("source-type", 0, "sourceType")
	usageType := fs.Int("usage-type", 0, "usageType")
	configFile := fs.String("config-file", "", "Connector config JSON object; use a private file or - for stdin to keep credentials out of process arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector create [options]\n\nProduct SDK: ConnectorService.Create\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.ConnectorDefinitionOption
	if *configFile != "" {
		config, err := loadJSONFile[map[string]any](*configFile)
		if err != nil {
			return fmt.Errorf("--config-file: %w", err)
		}
		options = append(options, sdk.WithConnectorConfig(config))
	}
	result, err := ctx.Client.Connectors(wid).Create(context.Background(), *name, *sourceType, *usageType, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	name := fs.String("name", "", "New connector name (unchanged when omitted)")
	sourceType := fs.Int("source-type", 0, "New source type (unchanged when omitted)")
	usageType := fs.Int("usage-type", 0, "New usage type (unchanged when omitted)")
	configFile := fs.String("config-file", "", "Replacement connector config JSON object; private file or - for stdin (unchanged when omitted)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector update [options]\n\nProduct SDK: ConnectorService.Update\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.ConnectorUpdateOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "name":
			options = append(options, sdk.WithConnectorName(*name))
		case "source-type":
			options = append(options, sdk.WithConnectorSourceType(*sourceType))
		case "usage-type":
			options = append(options, sdk.WithConnectorUsageType(*usageType))
		}
	})
	if *configFile != "" {
		config, err := loadJSONFile[map[string]any](*configFile)
		if err != nil {
			return fmt.Errorf("--config-file: %w", err)
		}
		options = append(options, sdk.WithConnectorUpdateConfig(config))
	}
	if err := ctx.Client.Connectors(wid).Update(context.Background(), *connectorID, options...); err != nil {
		return err
	}
	fmt.Printf("connector update: ok\n")
	return nil
}

func runConnectorServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector delete [options]\n\nProduct SDK: ConnectorService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).Delete(context.Background(), *connectorID); err != nil {
		return err
	}
	fmt.Printf("connector delete: ok\n")
	return nil
}

func runConnectorServiceValidate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector validate", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	sourceType := fs.Int("source-type", 0, "sourceType")
	usageType := fs.Int("usage-type", 0, "usageType")
	configFile := fs.String("config-file", "", "Connector config JSON object; use a private file or - for stdin to keep credentials out of process arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector validate [options]\n\nProduct SDK: ConnectorService.Validate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.ConnectorDefinitionOption
	if *configFile != "" {
		config, err := loadJSONFile[map[string]any](*configFile)
		if err != nil {
			return fmt.Errorf("--config-file: %w", err)
		}
		options = append(options, sdk.WithConnectorConfig(config))
	}
	result, err := ctx.Client.Connectors(wid).Validate(context.Background(), *name, *sourceType, *usageType, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceProbeRESTDraft(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector probe-rest-draft", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	configFile := fs.String("config-file", "", "REST API connector config JSON object; use a private file or - for stdin to keep credentials out of process arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector probe-rest-draft [options]\n\nProduct SDK: ConnectorService.ProbeRESTDraft\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	config, err := loadJSONFile[map[string]any](*configFile)
	if err != nil {
		return fmt.Errorf("--config-file: %w", err)
	}
	result, err := ctx.Client.Connectors(wid).ProbeRESTDraft(context.Background(), config)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceProbeREST(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector probe-rest", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector probe-rest [options]\n\nProduct SDK: ConnectorService.ProbeREST\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).ProbeREST(context.Background(), *connectorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServicePreviewREST(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector preview-rest", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	bodyJSON := fs.String("json", "", "REST preview JSON")
	bodyFile := fs.String("json-file", "", "REST preview JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector preview-rest [options]\n\nProduct SDK: ConnectorService.PreviewREST\n\nOptions:\n")
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
	var preview sdk.RESTPreview
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &preview); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).PreviewREST(context.Background(), *connectorID, preview)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceFiles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector files", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector files [options]\n\nProduct SDK: ConnectorService.Files\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).Files(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceUpload(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector upload", flag.ExitOnError)
	upload := addUploadFlags(fs)
	name := fs.String("name", "", "Import task name (omitted: server-generated name)")
	volumeID := fs.String("volume-id", "", "Target VolumeID")
	meta := fs.String("meta", "", "Connector upload metadata JSON (schema: moi-cli schema connector upload --argument meta)")
	fileTypes := fs.String("file-types", "", "File types JSON (schema: moi-cli schema connector upload --argument file-types)")
	pathRegex := fs.String("path-regex", "", "Path selection regular expression")
	keepStructure := fs.Bool("unzip-keep-structure", false, "Preserve archive directory structure")
	dedup := fs.String("dedup", "", "Deduplication configuration JSON (schema: moi-cli schema connector upload --argument dedup)")
	tableConfig := fs.String("table-config", "", "Table import configuration JSON (schema: moi-cli schema connector upload --argument table-config)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector upload [options]\n\nProduct SDK: ConnectorService.Upload\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	sources, err := upload.openFiles()
	if err != nil {
		return err
	}
	defer closeUploadSources(sources)
	var opts []sdk.ConnectorUploadOption
	for _, source := range sources {
		opts = append(opts, sdk.WithConnectorUploadFile("file", source.Name, source.Reader))
	}
	fields := map[string]string{"name": *name, "VolumeID": *volumeID, "meta": *meta, "file_types": *fileTypes,
		"path_regex": *pathRegex, "dedup": *dedup, "table_config": *tableConfig}
	for name, value := range fields {
		if value != "" {
			opts = append(opts, sdk.WithConnectorUploadField(name, value))
		}
	}
	// Omit unprovided false to preserve the historical empty multipart request.
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "unzip-keep-structure" {
			opts = append(opts, sdk.WithConnectorUploadField("unzip_keep_structure", fmt.Sprint(*keepStructure)))
		}
	})
	result, err := ctx.Client.Connectors(wid).Upload(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceLocalUpload(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector local-upload", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector local-upload [options]\n\nProduct SDK: ConnectorService.LocalUpload\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	sources, err := upload.openFiles()
	if err != nil {
		return err
	}
	defer closeUploadSources(sources)
	files := make([]sdk.MultipartFile, 0, len(sources))
	for _, source := range sources {
		files = append(files, sdk.MultipartFile{FieldName: "file", Filename: source.Name, Reader: source.Reader})
	}
	result, err := ctx.Client.Connectors(wid).LocalUpload(context.Background(), files...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceDBListDatabases(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector db-list-databases", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector db-list-databases [options]\n\nProduct SDK: ConnectorService.DBListDatabases\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).DBListDatabases(context.Background(), *connectorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceDBListSchemas(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector db-list-schemas", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	database := fs.String("database", "", "database")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector db-list-schemas [options]\n\nProduct SDK: ConnectorService.DBListSchemas\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).DBListSchemas(context.Background(), *connectorID, *database)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func connectorDBSchemaCLIOptions(fs *flag.FlagSet, schema string) []sdk.ConnectorDBOption {
	var opts []sdk.ConnectorDBOption
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "schema" {
			opts = append(opts, sdk.WithConnectorSchema(schema))
		}
	})
	return opts
}

func runConnectorServiceDBListTables(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector db-list-tables", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	database := fs.String("database", "", "database")
	schema := fs.String("schema", "", "Source database schema")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector db-list-tables [options]\n\nProduct SDK: ConnectorService.DBListTables\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := connectorDBSchemaCLIOptions(fs, *schema)
	result, err := ctx.Client.Connectors(wid).DBListTables(context.Background(), *connectorID, *database, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceDBSourceSchema(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector db-source-schema", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	database := fs.String("database", "", "database")
	table := fs.String("table", "", "table")
	schema := fs.String("schema", "", "Source database schema")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector db-source-schema [options]\n\nProduct SDK: ConnectorService.DBSourceSchema\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := connectorDBSchemaCLIOptions(fs, *schema)
	result, err := ctx.Client.Connectors(wid).DBSourceSchema(context.Background(), *connectorID, *database, *table, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredSources(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-sources", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-sources [options]\n\nProduct SDK: ConnectorService.StructuredSources\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).StructuredSources(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredSourceDatabases(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-source-databases", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	sourceType := fs.String("source-type", "", "sourceType")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-source-databases [options]\n\nProduct SDK: ConnectorService.StructuredSourceDatabases\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).StructuredSourceDatabases(context.Background(), *connectorID, *sourceType)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredSourceSchemas(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-source-schemas", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	sourceType := fs.String("source-type", "", "sourceType")
	database := fs.String("database", "", "database")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-source-schemas [options]\n\nProduct SDK: ConnectorService.StructuredSourceSchemas\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).StructuredSourceSchemas(context.Background(), *connectorID, *sourceType, *database)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredSourceObjects(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-source-objects", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-source-objects [options]\n\nProduct SDK: ConnectorService.StructuredSourceObjects\n\nOptions:\n")
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
	var req sdk.StructuredSourceObjectsRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).StructuredSourceObjects(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredSourceSchema(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-source-schema", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-source-schema [options]\n\nProduct SDK: ConnectorService.StructuredSourceSchema\n\nOptions:\n")
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
	var req sdk.StructuredSourceSchemaRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).StructuredSourceSchema(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredSourceCapability(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-source-capability", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	sourceType := fs.String("source-type", "", "sourceType")
	database := fs.String("database", "", "database")
	table := fs.String("table", "", "table")
	schema := fs.String("schema", "", "schema")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-source-capability [options]\n\nProduct SDK: ConnectorService.StructuredSourceCapability\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).StructuredSourceCapability(context.Background(), *connectorID, *sourceType, *database, *table, *schema)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredSourceSample(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-source-sample", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	sourceType := fs.String("source-type", "", "sourceType")
	limit := fs.Int("limit", 0, "limit")
	topicMode := fs.String("topic-mode", "", "topicMode")
	topicsJSON := fs.String("topics-json", "", "JSON array of topics")
	topicPattern := fs.String("topic-pattern", "", "topicPattern")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-source-sample [options]\n\nProduct SDK: ConnectorService.StructuredSourceSample\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var topics []string
	if *topicsJSON != "" {
		if err := json.Unmarshal([]byte(*topicsJSON), &topics); err != nil {
			return fmt.Errorf("parse --topics-json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).StructuredSourceSample(context.Background(), *connectorID, *sourceType, int32(*limit), *topicMode, topics, *topicPattern)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceStructuredBackfillStartResolve(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector structured-backfill-start-resolve", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector structured-backfill-start-resolve [options]\n\nProduct SDK: ConnectorService.StructuredBackfillStartResolve\n\nOptions:\n")
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
	var req sdk.StructuredBackfillStartResolveRequest
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).StructuredBackfillStartResolve(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceFileDownload(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector file-download", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connFileID := fs.String("conn-file-id", "", "connFileID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector file-download [options]\n\nProduct SDK: ConnectorService.FileDownload\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).FileDownload(context.Background(), *connFileID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceFileStream(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector file-stream", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connFileID := fs.String("conn-file-id", "", "connFileID")
	outPath := fs.String("output", "", "Write stream to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector file-stream [options]\n\nProduct SDK: ConnectorService.FileStream\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	stream, err := ctx.Client.Connectors(wid).FileStream(context.Background(), *connFileID)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runConnectorServiceFileDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector file-delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connFileID := fs.String("conn-file-id", "", "connFileID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector file-delete [options]\n\nProduct SDK: ConnectorService.FileDelete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).FileDelete(context.Background(), *connFileID); err != nil {
		return err
	}
	fmt.Printf("connector file-delete: ok\n")
	return nil
}

func runConnectorServiceMODatabaseList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector mo-database-list", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector mo-database-list [options]\n\nProduct SDK: ConnectorService.MODatabaseList\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).MODatabaseList(context.Background(), *connectorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceMODatabaseTableList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector mo-database-table-list", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	database := fs.String("database", "", "database")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector mo-database-table-list [options]\n\nProduct SDK: ConnectorService.MODatabaseTableList\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).MODatabaseTableList(context.Background(), *connectorID, *database)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceMOCheckTableName(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector mo-check-table-name", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	database := fs.String("database", "", "database")
	table := fs.String("table", "", "table")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector mo-check-table-name [options]\n\nProduct SDK: ConnectorService.MOCheckTableName\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).MOCheckTableName(context.Background(), *connectorID, *database, *table)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceMOColumnMapping(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector mo-column-mapping", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "connectorID")
	database := fs.String("database", "", "database")
	table := fs.String("table", "", "table")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector mo-column-mapping [options]\n\nProduct SDK: ConnectorService.MOColumnMapping\n\nOptions:\n")
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
	var exportColumns map[string]string
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &exportColumns); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).MOColumnMapping(context.Background(), *connectorID, *database, *table, exportColumns)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceCreateTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector create-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector create-task [options]\n\nProduct SDK: ConnectorService.CreateTask\n\nOptions:\n")
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
	var spec sdk.ImportTaskCreateSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).CreateTask(context.Background(), spec)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceUpdateTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector update-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector update-task [options]\n\nProduct SDK: ConnectorService.UpdateTask\n\nOptions:\n")
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
	var spec sdk.ImportTaskUpdateSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	if err := ctx.Client.Connectors(wid).UpdateTask(context.Background(), *taskID, spec); err != nil {
		return err
	}
	fmt.Printf("connector update-task: ok\n")
	return nil
}

func runConnectorServiceUpdateTaskFromSpec(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector update-task-from-spec", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector update-task-from-spec [options]\n\nProduct SDK: ConnectorService.UpdateTaskFromSpec\n\nOptions:\n")
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
	var spec sdk.ImportTaskUpdateSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	if err := ctx.Client.Connectors(wid).UpdateTaskFromSpec(context.Background(), *taskID, spec); err != nil {
		return err
	}
	fmt.Printf("connector update-task-from-spec: ok\n")
	return nil
}

func runConnectorServiceDeleteTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector delete-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector delete-task [options]\n\nProduct SDK: ConnectorService.DeleteTask\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).DeleteTask(context.Background(), *taskID); err != nil {
		return err
	}
	fmt.Printf("connector delete-task: ok\n")
	return nil
}

func runConnectorServicePauseTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector pause-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector pause-task [options]\n\nProduct SDK: ConnectorService.PauseTask\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).PauseTask(context.Background(), *taskID); err != nil {
		return err
	}
	fmt.Printf("connector pause-task: ok\n")
	return nil
}

func runConnectorServiceResumeTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector resume-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector resume-task [options]\n\nProduct SDK: ConnectorService.ResumeTask\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).ResumeTask(context.Background(), *taskID); err != nil {
		return err
	}
	fmt.Printf("connector resume-task: ok\n")
	return nil
}

func runConnectorServiceRunTaskNow(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector run-task-now", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector run-task-now [options]\n\nProduct SDK: ConnectorService.RunTaskNow\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).RunTaskNow(context.Background(), *taskID); err != nil {
		return err
	}
	fmt.Printf("connector run-task-now: ok\n")
	return nil
}

func runConnectorServiceRunTaskFullRefresh(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector run-task-full-refresh", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector run-task-full-refresh [options]\n\nProduct SDK: ConnectorService.RunTaskFullRefresh\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).RunTaskFullRefresh(context.Background(), *taskID); err != nil {
		return err
	}
	fmt.Printf("connector run-task-full-refresh: ok\n")
	return nil
}

func runConnectorServiceGetTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector get-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector get-task [options]\n\nProduct SDK: ConnectorService.GetTask\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).GetTask(context.Background(), *taskID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceTaskFiles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector task-files", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector task-files [options]\n\nProduct SDK: ConnectorService.TaskFiles\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).TaskFiles(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceTaskRuns(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector task-runs", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector task-runs [options]\n\nProduct SDK: ConnectorService.TaskRuns\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).TaskRuns(context.Background(), *taskID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceRetryTaskFiles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector retry-task-files", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	ids := fs.String("ids", "", "ids of task files to retry (comma-separated; empty retries all failed files)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector retry-task-files [options]\n\nProduct SDK: ConnectorService.RetryTaskFiles\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).RetryTaskFiles(context.Background(), *taskID, splitCSVFlag(fs, "ids", *ids)...); err != nil {
		return err
	}
	fmt.Printf("connector retry-task-files: ok\n")
	return nil
}

func runConnectorServiceExportTasks(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector export-tasks", flag.ExitOnError)
	keyword := fs.String("keyword", "", "Search task, connector and source path")
	target := fs.String("target-category", "", "Target category: database, storage, model or other")
	limit := fs.Int("limit", 0, "Page size (omit to use the MOI default)")
	offset := fs.Int("offset", 0, "Page offset")
	connectorName := fs.String("connector-name", "", "Filter by connector name")
	taskID := fs.String("task-id", "", "Filter by export task ID")
	orderBy := fs.String("order-by", "", "Sort field")
	orderDirection := fs.String("order-direction", "", "Sort direction: asc or desc")
	statuses := fs.String("statuses", "", "Comma-separated task status values")
	creator := fs.String("creator", "", "Filter by creator label")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector export-tasks [options]\n\nProduct SDK: ConnectorService.ExportTasks\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.ExportTaskListOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "keyword":
			opts = append(opts, sdk.WithExportTaskKeyword(*keyword))
		case "target-category":
			opts = append(opts, sdk.WithExportTaskTargetCategory(*target))
		case "limit":
			opts = append(opts, sdk.WithExportTaskLimit(*limit))
		case "offset":
			opts = append(opts, sdk.WithExportTaskOffset(*offset))
		case "connector-name":
			opts = append(opts, sdk.WithExportTaskConnectorName(*connectorName))
		case "task-id":
			opts = append(opts, sdk.WithExportTaskID(*taskID))
		case "order-by":
			opts = append(opts, sdk.WithExportTaskOrderBy(*orderBy))
		case "order-direction":
			opts = append(opts, sdk.WithExportTaskOrderDirection(*orderDirection))
		case "creator":
			opts = append(opts, sdk.WithExportTaskCreator(*creator))
		}
	})
	if *statuses != "" {
		values, err := parseIntCSV("statuses", *statuses)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithExportTaskStatuses(values...))
	}
	result, err := ctx.Client.Connectors(wid).ExportTasks(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceExportTaskInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector export-task-info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector export-task-info [options]\n\nProduct SDK: ConnectorService.ExportTaskInfo\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).ExportTaskInfo(context.Background(), *taskID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceExportTaskFiles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector export-task-files", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	limit := fs.Int("limit", 0, "Page size (omit to use the MOI default)")
	offset := fs.Int("offset", 0, "Page offset")
	statuses := fs.String("statuses", "", "Comma-separated file status values")
	orderBy := fs.String("order-by", "", "Sort field")
	orderDirection := fs.String("order-direction", "", "Sort direction: asc or desc")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector export-task-files [options]\n\nProduct SDK: ConnectorService.ExportTaskFiles\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.ExportTaskFilesOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "limit":
			opts = append(opts, sdk.WithExportTaskFilesLimit(*limit))
		case "offset":
			opts = append(opts, sdk.WithExportTaskFilesOffset(*offset))
		case "order-by":
			opts = append(opts, sdk.WithExportTaskFilesOrderBy(*orderBy))
		case "order-direction":
			opts = append(opts, sdk.WithExportTaskFilesOrderDirection(*orderDirection))
		}
	})
	if *statuses != "" {
		values, err := parseIntCSV("statuses", *statuses)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithExportTaskFileStatuses(values...))
	}
	result, err := ctx.Client.Connectors(wid).ExportTaskFiles(context.Background(), *taskID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceDeleteExportTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector delete-export-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector delete-export-task [options]\n\nProduct SDK: ConnectorService.DeleteExportTask\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).DeleteExportTask(context.Background(), *taskID); err != nil {
		return err
	}
	fmt.Printf("connector delete-export-task: ok\n")
	return nil
}

func runConnectorServiceCreateExportTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector create-export-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector create-export-task [options]\n\nProduct SDK: ConnectorService.CreateExportTask\n\nOptions:\n")
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
	var spec sdk.ExportTaskSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Connectors(wid).CreateExportTask(context.Background(), spec)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runConnectorServiceRetryExportTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector retry-export-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fileIDs := fs.String("file-ids", "", "fileIDs to rerun (comma-separated; empty reruns the whole task)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector retry-export-task [options]\n\nProduct SDK: ConnectorService.RetryExportTask\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Connectors(wid).RetryExportTask(context.Background(), *taskID, splitCSVFlag(fs, "file-ids", *fileIDs)...); err != nil {
		return err
	}
	fmt.Printf("connector retry-export-task: ok\n")
	return nil
}

func splitCSVFlag(fs *flag.FlagSet, name, value string) []string {
	values := splitCSV(value)
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name && values == nil {
			values = []string{}
		}
	})
	return values
}

func runConnectorServiceExportTaskState(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector export-task-state", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "taskID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector export-task-state [options]\n\nProduct SDK: ConnectorService.ExportTaskState\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).ExportTaskState(context.Background(), *taskID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
