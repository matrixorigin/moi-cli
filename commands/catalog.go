package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteCatalogCommand runs product SDK CatalogService methods via moi-cli.
func ExecuteCatalogCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowCatalogHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "list-workspace-files":
		return runWorkspaceFileList(ctx, args[1:])
	case "catalog-file-content":
		return runCatalogServiceCatalogFileContent(ctx, args[1:])
	case "workspace-table-overview":
		return runCatalogServiceWorkspaceTableOverview(ctx, args[1:])
	case "workspace-database-tables":
		return runCatalogServiceWorkspaceDatabaseTables(ctx, args[1:])
	case "workspace-table-columns":
		return runCatalogServiceWorkspaceTableColumns(ctx, args[1:])
	case "garbage-collection":
		return runCatalogServiceGarbageCollection(ctx, args[1:])
	case "workspace-file-model-url":
		return runCatalogServiceWorkspaceFileModelURL(ctx, args[1:])
	case "table-references":
		return runCatalogServiceTableReferences(ctx, args[1:])
	case "folder-references":
		return runCatalogServiceFolderReferences(ctx, args[1:])
	case "volume-references":
		return runCatalogServiceVolumeReferences(ctx, args[1:])
	case "database-references":
		return runCatalogServiceDatabaseReferences(ctx, args[1:])
	case "catalog-references":
		return runCatalogServiceCatalogReferences(ctx, args[1:])
	case "preview-workspace-file":
		return runCatalogServicePreviewWorkspaceFile(ctx, args[1:])
	case "get":
		return runCatalogGetAlias(ctx, args[1:])
	case "prepare":
		return runCatalogPrepareAlias(ctx, args[1:])
	case "health":
		return runCatalogServiceHealth(ctx, args[1:])
	case "allowed-actions":
		return runCatalogAllowedActions(ctx, args[1:])
	case "tree":
		return runCatalogServiceTree(ctx, args[1:])
	case "system-tree":
		return runCatalogServiceSystemTree(ctx, args[1:])
	case "list":
		return runCatalogServiceList(ctx, args[1:])
	case "create":
		return runCatalogServiceCreate(ctx, args[1:])
	case "info":
		return runCatalogServiceInfo(ctx, args[1:])
	case "update":
		return runCatalogServiceUpdate(ctx, args[1:])
	case "delete":
		return runCatalogServiceDelete(ctx, args[1:])
	case "list-databases":
		return runCatalogServiceListDatabases(ctx, args[1:])
	case "create-database":
		return runCatalogServiceCreateDatabase(ctx, args[1:])
	case "database-info":
		return runCatalogServiceDatabaseInfo(ctx, args[1:])
	case "update-database":
		return runCatalogServiceUpdateDatabase(ctx, args[1:])
	case "update-database-description":
		return runCatalogServiceUpdateDatabaseDescription(ctx, args[1:])
	case "delete-database":
		return runCatalogServiceDeleteDatabase(ctx, args[1:])
	case "database-children":
		return runCatalogServiceDatabaseChildren(ctx, args[1:])
	case "database-table-leaves":
		return runCatalogServiceDatabaseTableLeaves(ctx, args[1:])
	case "search-data-dashboard-tables":
		return runCatalogServiceSearchDataDashboardTables(ctx, args[1:])
	case "create-volume":
		return runCatalogServiceCreateVolume(ctx, args[1:])
	case "volume-info":
		return runCatalogServiceVolumeInfo(ctx, args[1:])
	case "update-volume":
		return runCatalogServiceUpdateVolume(ctx, args[1:])
	case "delete-volume":
		return runCatalogServiceDeleteVolume(ctx, args[1:])
	case "volume-full-path":
		return runCatalogServiceVolumeFullPath(ctx, args[1:])
	case "create-table":
		return runCatalogServiceCreateTable(ctx, args[1:])
	case "table-info":
		return runCatalogServiceTableInfo(ctx, args[1:])
	case "multi-table-info":
		return runCatalogServiceMultiTableInfo(ctx, args[1:])
	case "delete-table":
		return runCatalogServiceDeleteTable(ctx, args[1:])
	case "table-data":
		return runCatalogServiceTableData(ctx, args[1:])
	case "preview-table":
		return runCatalogServicePreviewTable(ctx, args[1:])
	case "download-table":
		return runCatalogServiceDownloadTable(ctx, args[1:])
	case "table-exists":
		return runCatalogServiceTableExists(ctx, args[1:])
	case "table-full-path":
		return runCatalogServiceTableFullPath(ctx, args[1:])
	case "truncate-table":
		return runCatalogServiceTruncateTable(ctx, args[1:])
	case "table-overview":
		return runCatalogServiceTableOverview(ctx, args[1:])
	case "files":
		return runCatalogServiceFiles(ctx, args[1:])
	case "create-file":
		return runCatalogServiceCreateFile(ctx, args[1:])
	case "file-info":
		return runCatalogServiceFileInfo(ctx, args[1:])
	case "delete-file":
		return runCatalogServiceDeleteFile(ctx, args[1:])
	case "delete-file-ref":
		return runCatalogServiceDeleteFileRef(ctx, args[1:])
	case "trigger-file-workflow":
		return runCatalogServiceTriggerFileWorkflow(ctx, args[1:])
	case "download-file":
		return runCatalogServiceDownloadFile(ctx, args[1:])
	case "preview-file":
		return runCatalogServicePreviewFile(ctx, args[1:])
	case "upload-bytes":
		return runCatalogServiceUploadBytes(ctx, args[1:])
	case "upload-file":
		return runCatalogServiceUploadFile(ctx, args[1:])
	case "upload":
		return runCatalogServiceUpload(ctx, args[1:])
	case "upload-private-bytes":
		return runCatalogServiceUploadPrivateBytes(ctx, args[1:])
	case "upload-private-file":
		return runCatalogServiceUploadPrivateFile(ctx, args[1:])
	case "upload-private":
		return runCatalogServiceUploadPrivate(ctx, args[1:])
	case "upload-runtime-artifact":
		return runRuntimeArtifactUpload(ctx, args[1:], "catalog upload-runtime-artifact", false)
	case "download-workspace-file":
		return runRuntimeWorkspaceFileDownload(ctx, args[1:], "catalog download-workspace-file")
	case "analyze-runtime-image":
		return runRuntimeImageAnalysis(ctx, args[1:], "catalog analyze-runtime-image")
	case "data-asset":
		return runCatalogServiceDataAsset(ctx, args[1:])
	case "batch-get-data-assets":
		return runCatalogServiceBatchGetDataAssets(ctx, args[1:])
	case "file-data-asset":
		return runCatalogServiceFileDataAsset(ctx, args[1:])
	case "file-artifact":
		return runCatalogServiceFileArtifact(ctx, args[1:])
	case "batch-resolve-file-artifacts":
		return runCatalogServiceBatchResolveFileArtifacts(ctx, args[1:])
	case "create-folder":
		return runCatalogServiceCreateFolder(ctx, args[1:])
	case "update-folder":
		return runCatalogServiceUpdateFolder(ctx, args[1:])
	case "delete-folder":
		return runCatalogServiceDeleteFolder(ctx, args[1:])
	case "clean-folder":
		return runCatalogServiceCleanFolder(ctx, args[1:])
	case "run-nl2sql":
		return runCatalogServiceRunNL2SQL(ctx, args[1:])
	default:
		return fmt.Errorf("unknown catalog subcommand: %s", sub)
	}
}

// ShowCatalogHelp prints help for catalog.
func ShowCatalogHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: catalog-file-content")

	fmt.Fprintf(os.Stderr, `Usage: catalog <subcommand> [options]

Product SDK surface: CatalogService

Subcommands:
  workspace-table-overview         CatalogService.WorkspaceTableOverview
  workspace-database-tables        CatalogService.WorkspaceDatabaseTables
  workspace-table-columns          CatalogService.WorkspaceTableColumns
  garbage-collection               CatalogService.GarbageCollection
  workspace-file-model-url         CatalogService.WorkspaceFileModelURL
  table-references                 CatalogService.TableReferences
  folder-references                CatalogService.FolderReferences
  volume-references                CatalogService.VolumeReferences
  database-references              CatalogService.DatabaseReferences
  catalog-references               CatalogService.CatalogReferences
  preview-workspace-file           CatalogService.PreviewWorkspaceFile
  get                              (alias)
  prepare                          (alias)
  health                           CatalogService.Health
  allowed-actions                  CatalogService.AllowedActions
  tree                             CatalogService.Tree
  system-tree                      CatalogService.SystemTree
  list                             CatalogService.List
  create                           CatalogService.Create
  info                             CatalogService.Info
  update                           CatalogService.Update
  delete                           CatalogService.Delete
  list-databases                   CatalogService.ListDatabases
  create-database                  CatalogService.CreateDatabase
  database-info                    CatalogService.DatabaseInfo
  update-database                  CatalogService.UpdateDatabase
  update-database-description      CatalogService.UpdateDatabaseDescription
  delete-database                  CatalogService.DeleteDatabase
  database-children                CatalogService.DatabaseChildren
  database-table-leaves            CatalogService.DatabaseTableLeaves
  search-data-dashboard-tables     CatalogService.SearchDataDashboardTables
  create-volume                    CatalogService.CreateVolume
  volume-info                      CatalogService.VolumeInfo
  update-volume                    CatalogService.UpdateVolume
  delete-volume                    CatalogService.DeleteVolume
  volume-full-path                 CatalogService.VolumeFullPath
  create-table                     CatalogService.CreateTable
  table-info                       CatalogService.TableInfo
  multi-table-info                 CatalogService.MultiTableInfo
  delete-table                     CatalogService.DeleteTable
  table-data                       CatalogService.TableData
  preview-table                    CatalogService.PreviewTable
  download-table                   CatalogService.DownloadTable
  table-exists                     CatalogService.TableExists
  table-full-path                  CatalogService.TableFullPath
  truncate-table                   CatalogService.TruncateTable
  table-overview                   CatalogService.TableOverview
  files                            CatalogService.Files
  list-workspace-files             CatalogService.ListWorkspaceFiles (alias: file list)
  create-file                      CatalogService.CreateFile
  file-info                        CatalogService.FileInfo
  delete-file                      CatalogService.DeleteFile
  delete-file-ref                  CatalogService.DeleteFileRef
  trigger-file-workflow            CatalogService.TriggerFileWorkflow
  download-file                    CatalogService.DownloadFile
  preview-file                     CatalogService.PreviewFile
  upload-bytes                     CatalogService.UploadBytes
  upload-file                      CatalogService.UploadFile
  upload                           CatalogService.Upload
  upload-private-bytes             CatalogService.UploadPrivateBytes
  upload-private-file              CatalogService.UploadPrivateFile
  upload-private                   CatalogService.UploadPrivate
  upload-runtime-artifact          CatalogService.UploadRuntimeArtifact
  download-workspace-file          CatalogService.DownloadWorkspaceFile
  analyze-runtime-image            CatalogService.AnalyzeRuntimeImage
  data-asset                       CatalogService.DataAsset
  batch-get-data-assets            CatalogService.BatchGetDataAssets
  file-data-asset                  CatalogService.FileDataAsset
  file-artifact                    CatalogService.FileArtifact
  batch-resolve-file-artifacts     CatalogService.BatchResolveFileArtifacts
  create-folder                    CatalogService.CreateFolder
  update-folder                    CatalogService.UpdateFolder
  delete-folder                    CatalogService.DeleteFolder
  clean-folder                     CatalogService.CleanFolder
  run-nl2sql                       CatalogService.RunNL2SQL

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runCatalogServiceHealth(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog health", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog health [options]\n\nProduct SDK: CatalogService.Health\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).Health(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceTree(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog tree", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	pageSize := fs.Int("page-size", 0, "Page size (omit to use the MOI default)")
	pageToken := fs.String("page-token", "", "Cursor from the previous page")
	search := fs.String("search", "", "Search catalog resources")
	resourceTypes := fs.String("resource-types", "", "Comma-separated resource types")
	resourceID := fs.Int64("resource-id", 0, "Resolve a specific resource ID")
	includeTableLeaves := fs.Bool("include-table-leaves", true, "Include table nodes")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog tree [options]\n\nProduct SDK: CatalogService.Tree\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.CatalogTreeOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "page-size":
			opts = append(opts, sdk.WithCatalogTreePageSize(*pageSize))
		case "page-token":
			opts = append(opts, sdk.WithCatalogTreePageToken(*pageToken))
		case "search":
			opts = append(opts, sdk.WithCatalogTreeSearch(*search))
		case "resource-types":
			opts = append(opts, sdk.WithCatalogTreeResourceTypes(splitCSV(*resourceTypes)...))
		case "resource-id":
			opts = append(opts, sdk.WithCatalogTreeResourceID(*resourceID))
		case "include-table-leaves":
			opts = append(opts, sdk.WithCatalogTreeTableLeaves(*includeTableLeaves))
		}
	})
	result, err := ctx.Client.Catalog(wid).Tree(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceSystemTree(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog system-tree", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	pageSize := fs.Int("page-size", 0, "Page size (omit to use the MOI default)")
	pageToken := fs.String("page-token", "", "Cursor from the previous page")
	search := fs.String("search", "", "Search system catalog resources")
	includeTableLeaves := fs.Bool("include-table-leaves", true, "Include table nodes")
	databaseName := fs.String("database-name", "", "Select a system database")
	tableName := fs.String("table-name", "", "Resolve an exact system table name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog system-tree [options]\n\nProduct SDK: CatalogService.SystemTree\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.SystemCatalogTreeOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "page-size":
			opts = append(opts, sdk.WithSystemCatalogPageSize(*pageSize))
		case "page-token":
			opts = append(opts, sdk.WithSystemCatalogPageToken(*pageToken))
		case "search":
			opts = append(opts, sdk.WithSystemCatalogSearch(*search))
		case "include-table-leaves":
			opts = append(opts, sdk.WithSystemCatalogTableLeaves(*includeTableLeaves))
		case "database-name":
			opts = append(opts, sdk.WithSystemCatalogDatabase(*databaseName))
		case "table-name":
			opts = append(opts, sdk.WithSystemCatalogTableName(*tableName))
		}
	})
	result, err := ctx.Client.Catalog(wid).SystemTree(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog list", flag.ExitOnError)
	var params catalogListParameters
	fs.BoolVar(&params.includeStats, "include-stats", true, "Include child statistics (omit for server default)")
	fs.IntVar(&params.pageSize, "page-size", 0, "Page size (omit for full list)")
	fs.StringVar(&params.pageToken, "page-token", "", "Continuation token returned by the previous page")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog list [options]\n\nProduct SDK: CatalogService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).List(context.Background(), params.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog create", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	info := addCatalogInfoFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog create [options]\n\nProduct SDK: CatalogService.Create\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).Create(context.Background(), *name, info.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	catalogID := fs.Int("catalog-id", 0, "catalogID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog info [options]\n\nProduct SDK: CatalogService.Info\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).Info(context.Background(), *catalogID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	catalogID := fs.Int("catalog-id", 0, "catalogID")
	name := fs.String("name", "", "name")
	info := addCatalogInfoFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog update [options]\n\nProduct SDK: CatalogService.Update\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).Update(context.Background(), *catalogID, *name, info.options(fs)...); err != nil {
		return err
	}
	fmt.Printf("catalog update: ok\n")
	return nil
}

func runCatalogServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	catalogID := fs.Int("catalog-id", 0, "catalogID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog delete [options]\n\nProduct SDK: CatalogService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).Delete(context.Background(), *catalogID); err != nil {
		return err
	}
	fmt.Printf("catalog delete: ok\n")
	return nil
}

func runCatalogServiceListDatabases(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog list-databases", flag.ExitOnError)
	var params catalogListParameters
	fs.BoolVar(&params.includeStats, "include-stats", true, "Include child statistics (omit for server default)")
	fs.IntVar(&params.pageSize, "page-size", 0, "Page size (omit for full list)")
	fs.StringVar(&params.pageToken, "page-token", "", "Continuation token returned by the previous page")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	catalogID := fs.Int("catalog-id", 0, "catalogID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog list-databases [options]\n\nProduct SDK: CatalogService.ListDatabases\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).ListDatabases(context.Background(), *catalogID, params.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceCreateDatabase(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog create-database", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	catalogID := fs.Int("catalog-id", 0, "catalogID")
	name := fs.String("name", "", "name")
	info := addCatalogInfoFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog create-database [options]\n\nProduct SDK: CatalogService.CreateDatabase\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).CreateDatabase(context.Background(), *catalogID, *name, info.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDatabaseInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog database-info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog database-info [options]\n\nProduct SDK: CatalogService.DatabaseInfo\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).DatabaseInfo(context.Background(), *databaseID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUpdateDatabase(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog update-database", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog update-database [options]\n\nProduct SDK: CatalogService.UpdateDatabase\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).UpdateDatabase(context.Background(), *databaseID, *name)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUpdateDatabaseDescription(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog update-database-description", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	info := addCatalogInfoFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog update-database-description [options]\n\nProduct SDK: CatalogService.UpdateDatabaseDescription\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).UpdateDatabaseDescription(context.Background(), *databaseID, info.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDeleteDatabase(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog delete-database", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog delete-database [options]\n\nProduct SDK: CatalogService.DeleteDatabase\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).DeleteDatabase(context.Background(), *databaseID); err != nil {
		return err
	}
	fmt.Printf("catalog delete-database: ok\n")
	return nil
}

func runCatalogServiceDatabaseChildren(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog database-children", flag.ExitOnError)
	var params catalogListParameters
	fs.BoolVar(&params.includeStats, "include-stats", true, "Include child statistics (omit for server default)")
	fs.IntVar(&params.pageSize, "page-size", 0, "Page size (omit for full list)")
	fs.StringVar(&params.pageToken, "page-token", "", "Continuation token returned by the previous page")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog database-children [options]\n\nProduct SDK: CatalogService.DatabaseChildren\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).DatabaseChildren(context.Background(), *databaseID, params.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDatabaseTableLeaves(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog database-table-leaves", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	pageSize := fs.Int("page-size", 0, "Page size (omit to use the MOI default)")
	pageToken := fs.String("page-token", "", "Cursor from the previous page")
	search := fs.String("search", "", "Search table names")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog database-table-leaves [options]\n\nProduct SDK: CatalogService.DatabaseTableLeaves\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.DatabaseTableLeavesOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "page-size":
			opts = append(opts, sdk.WithDatabaseTableLeavesPageSize(*pageSize))
		case "page-token":
			opts = append(opts, sdk.WithDatabaseTableLeavesPageToken(*pageToken))
		case "search":
			opts = append(opts, sdk.WithDatabaseTableLeavesSearch(*search))
		}
	})
	result, err := ctx.Client.Catalog(wid).DatabaseTableLeaves(context.Background(), *databaseID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceSearchDataDashboardTables(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog search-data-dashboard-tables", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	pageSize := fs.Int("page-size", 0, "Page size (omit to use the MOI default)")
	pageToken := fs.String("page-token", "", "Cursor from the previous page")
	search := fs.String("search", "", "Search tables across databases")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli catalog search-data-dashboard-tables [options]\n\nProduct SDK: CatalogService.SearchDataDashboardTables\n\nOptions:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.DataDashboardTableSearchOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "page-size":
			opts = append(opts, sdk.WithDataDashboardTableSearchPageSize(*pageSize))
		case "page-token":
			opts = append(opts, sdk.WithDataDashboardTableSearchPageToken(*pageToken))
		case "search":
			opts = append(opts, sdk.WithDataDashboardTableSearchQuery(*search))
		}
	})
	result, err := ctx.Client.Catalog(wid).SearchDataDashboardTables(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceCreateVolume(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog create-volume", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	name := fs.String("name", "", "name")
	info := addCatalogInfoFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog create-volume [options]\n\nProduct SDK: CatalogService.CreateVolume\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).CreateVolume(context.Background(), *databaseID, *name, info.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceVolumeInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog volume-info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.String("volume-id", "", "volumeID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog volume-info [options]\n\nProduct SDK: CatalogService.VolumeInfo\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).VolumeInfo(context.Background(), *volumeID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUpdateVolume(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog update-volume", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.Int("volume-id", 0, "volumeID")
	name := fs.String("name", "", "name")
	info := addCatalogInfoFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog update-volume [options]\n\nProduct SDK: CatalogService.UpdateVolume\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).UpdateVolume(context.Background(), *volumeID, *name, info.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDeleteVolume(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog delete-volume", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.Int("volume-id", 0, "volumeID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog delete-volume [options]\n\nProduct SDK: CatalogService.DeleteVolume\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).DeleteVolume(context.Background(), *volumeID); err != nil {
		return err
	}
	fmt.Printf("catalog delete-volume: ok\n")
	return nil
}

func runCatalogServiceVolumeFullPath(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog volume-full-path", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.Int("volume-id", 0, "volumeID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog volume-full-path [options]\n\nProduct SDK: CatalogService.VolumeFullPath\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).VolumeFullPath(context.Background(), *volumeID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceCreateTable(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog create-table", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	name := fs.String("name", "", "name")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	comment := fs.String("comment", "", "Table comment")
	dataFileURL := fs.String("data-file-url", "", "Source data file URL")
	columnLoadOptionsJSON := fs.String("column-load-options", "", "JSON array of column load options")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog create-table [options]\n\nProduct SDK: CatalogService.CreateTable\n\nOptions:\n")
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
	var columns []sdk.TableColumn
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &columns); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	var options []sdk.CreateTableOption
	columnLoadOptionsProvided := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "comment":
			options = append(options, sdk.WithTableComment(*comment))
		case "data-file-url":
			options = append(options, sdk.WithTableDataFileURL(*dataFileURL))
		case "column-load-options":
			columnLoadOptionsProvided = true
		}
	})
	if columnLoadOptionsProvided {
		var columnLoadOptions []sdk.ColumnLoadOption
		if err := json.Unmarshal([]byte(*columnLoadOptionsJSON), &columnLoadOptions); err != nil {
			return fmt.Errorf("parse --column-load-options: %w", err)
		}
		for _, option := range columnLoadOptions {
			options = append(options, sdk.WithColumnLoadOption(option.Column, option.DataFrom))
		}
	}
	result, err := ctx.Client.Catalog(wid).CreateTable(context.Background(), *databaseID, *name, columns, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceTableInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog table-info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableID := fs.Int("table-id", 0, "tableID")
	lookup := addCatalogTableLookupFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog table-info [options]\n\nProduct SDK: CatalogService.TableInfo\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).TableInfo(context.Background(), *tableID, lookup.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceMultiTableInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog multi-table-info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableIDs := fs.String("table-ids", "", "tableIDs (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog multi-table-info [options]\n\nProduct SDK: CatalogService.MultiTableInfo\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	ids, err := parseIntCSV("table-ids", *tableIDs)
	if err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).MultiTableInfo(context.Background(), ids...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDeleteTable(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog delete-table", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableID := fs.Int("table-id", 0, "tableID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog delete-table [options]\n\nProduct SDK: CatalogService.DeleteTable\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).DeleteTable(context.Background(), *tableID); err != nil {
		return err
	}
	fmt.Printf("catalog delete-table: ok\n")
	return nil
}

func runCatalogServiceTableData(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog table-data", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableID := fs.Int("table-id", 0, "tableID")
	parameters := addCatalogTableDataFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog table-data [options]\n\nProduct SDK: CatalogService.TableData\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).TableData(context.Background(), *tableID, parameters.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServicePreviewTable(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog preview-table", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableID := fs.Int("table-id", 0, "tableID")
	parameters := addCatalogTableDataFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog preview-table [options]\n\nProduct SDK: CatalogService.PreviewTable\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).PreviewTable(context.Background(), *tableID, parameters.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDownloadTable(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog download-table", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableID := fs.Int("table-id", 0, "tableID")
	outPath := fs.String("output", "", "Write stream to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog download-table [options]\n\nProduct SDK: CatalogService.DownloadTable\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	stream, err := ctx.Client.Catalog(wid).DownloadTable(context.Background(), *tableID)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runCatalogServiceTableExists(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog table-exists", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog table-exists [options]\n\nProduct SDK: CatalogService.TableExists\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).TableExists(context.Background(), *databaseID, *name)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceTableFullPath(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog table-full-path", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableID := fs.Int("table-id", 0, "tableID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog table-full-path [options]\n\nProduct SDK: CatalogService.TableFullPath\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).TableFullPath(context.Background(), *tableID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceTruncateTable(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog truncate-table", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	tableID := fs.Int("table-id", 0, "tableID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog truncate-table [options]\n\nProduct SDK: CatalogService.TruncateTable\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).TruncateTable(context.Background(), *tableID); err != nil {
		return err
	}
	fmt.Printf("catalog truncate-table: ok\n")
	return nil
}

func runCatalogServiceTableOverview(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog table-overview", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog table-overview [options]\n\nProduct SDK: CatalogService.TableOverview\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).TableOverview(context.Background(), *databaseID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceFiles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog files", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.String("volume-id", "", "Volume ID to list")
	parentID := fs.String("parent-id", "", "Parent folder ID (empty selects the Volume root)")
	name := fs.String("name", "", "File name substring")
	extensions := fs.String("extensions", "", "File extensions (comma-separated)")
	fileIDs := fs.String("file-ids", "", "File IDs (comma-separated)")
	fileType := fs.String("file-type", "", "File entry type")
	page := fs.Int("page", 0, "Page number (omit for server default)")
	pageSize := fs.Int("page-size", 0, "Page size (omit for server default)")
	order := fs.String("order", "", "Sort direction")
	orderBy := fs.String("order-by", "", "Sort field")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog files [options]\n\nProduct SDK: CatalogService.Files\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.FileListOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "volume-id":
			opts = append(opts, sdk.WithFileListFilter("volume_id", *volumeID))
		case "parent-id":
			opts = append(opts, sdk.WithFileListFilter("parent_id", *parentID))
		case "name":
			opts = append(opts, sdk.WithFuzzyFileListFilter("file_name", true, *name))
		case "extensions":
			opts = append(opts, sdk.WithFileListFilter("file_ext", splitCSV(*extensions)...))
		case "file-ids":
			opts = append(opts, sdk.WithFileListFilter("file_ids", splitCSV(*fileIDs)...))
		case "file-type":
			opts = append(opts, sdk.WithFileListFilter("file_type", *fileType))
		case "page":
			opts = append(opts, sdk.WithFileListPage(*page))
		case "page-size":
			opts = append(opts, sdk.WithFileListPageSize(*pageSize))
		case "order":
			opts = append(opts, sdk.WithFileListOrder(*order))
		case "order-by":
			opts = append(opts, sdk.WithFileListOrderBy(*orderBy))
		}
	})
	result, err := ctx.Client.Catalog(wid).Files(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceCreateFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog create-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.Int("volume-id", 0, "volumeID")
	fileIDs := fs.String("file-ids", "", "fileIDs (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog create-file [options]\n\nProduct SDK: CatalogService.CreateFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).CreateFile(context.Background(), *volumeID, splitCSV(*fileIDs)...); err != nil {
		return err
	}
	fmt.Printf("catalog create-file: ok\n")
	return nil
}

func runCatalogServiceFileInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog file-info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fileID := fs.String("file-id", "", "fileID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog file-info [options]\n\nProduct SDK: CatalogService.FileInfo\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).FileInfo(context.Background(), *fileID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDeleteFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog delete-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fileID := fs.String("file-id", "", "fileID")
	volumeID := fs.String("volume-id", "", "volumeID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog delete-file [options]\n\nProduct SDK: CatalogService.DeleteFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).DeleteFile(context.Background(), *fileID, *volumeID); err != nil {
		return err
	}
	fmt.Printf("catalog delete-file: ok\n")
	return nil
}

func runCatalogServiceDeleteFileRef(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog delete-file-ref", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.Int("volume-id", 0, "volumeID")
	fileIDs := fs.String("file-ids", "", "fileIDs (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog delete-file-ref [options]\n\nProduct SDK: CatalogService.DeleteFileRef\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).DeleteFileRef(context.Background(), *volumeID, splitCSV(*fileIDs)...); err != nil {
		return err
	}
	fmt.Printf("catalog delete-file-ref: ok\n")
	return nil
}

func runCatalogServiceTriggerFileWorkflow(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog trigger-file-workflow", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.Int("volume-id", 0, "volumeID")
	fileIDs := fs.String("file-ids", "", "fileIDs (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog trigger-file-workflow [options]\n\nProduct SDK: CatalogService.TriggerFileWorkflow\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).TriggerFileWorkflow(context.Background(), *volumeID, splitCSV(*fileIDs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDownloadFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog download-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.String("volume-id", "", "volumeID")
	fileID := fs.String("file-id", "", "fileID")
	asZip := fs.Bool("as-zip", false, "Download the file or folder as a ZIP archive")
	outPath := fs.String("output", "", "Write stream to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog download-file [options]\n\nProduct SDK: CatalogService.DownloadFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.DownloadFileOption
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "as-zip" {
			options = append(options, sdk.WithFileAsZip(*asZip))
		}
	})
	stream, err := ctx.Client.Catalog(wid).DownloadFile(context.Background(), *volumeID, *fileID, options...)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runCatalogServicePreviewFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog preview-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	volumeID := fs.String("volume-id", "", "volumeID")
	fileID := fs.String("file-id", "", "fileID")
	outPath := fs.String("output", "", "Write stream to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog preview-file [options]\n\nProduct SDK: CatalogService.PreviewFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	stream, err := ctx.Client.Catalog(wid).PreviewFile(context.Background(), *volumeID, *fileID)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runCatalogServiceUploadBytes(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog upload-bytes", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog upload-bytes [options]\n\nProduct SDK: CatalogService.UploadBytes\n\nOptions:\n")
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
	var data []byte
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &data); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Catalog(wid).UploadBytes(context.Background(), *filename, data)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUploadFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog upload-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filePath := fs.String("file-path", "", "filePath")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog upload-file [options]\n\nProduct SDK: CatalogService.UploadFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).UploadFile(context.Background(), *filePath)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUpload(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog upload", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	contentType := fs.String("content-type", "application/octet-stream", "File MIME type")
	externalSourceURI := fs.String("external-source-uri", "", "Trusted external source URI")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog upload [options]\n\nProduct SDK: CatalogService.Upload\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	source, err := upload.open(*filename, *bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	defer source.Close()
	result, err := ctx.Client.Catalog(wid).Upload(context.Background(), source.Name, source.Reader,
		sdk.WithCatalogUploadContentType(*contentType),
		sdk.WithCatalogUploadExternalSourceURI(*externalSourceURI),
	)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUploadPrivateBytes(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog upload-private-bytes", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog upload-private-bytes [options]\n\nProduct SDK: CatalogService.UploadPrivateBytes\n\nOptions:\n")
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
	var data []byte
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &data); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Catalog(wid).UploadPrivateBytes(context.Background(), *filename, data)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUploadPrivateFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog upload-private-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filePath := fs.String("file-path", "", "filePath")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog upload-private-file [options]\n\nProduct SDK: CatalogService.UploadPrivateFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).UploadPrivateFile(context.Background(), *filePath)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUploadPrivate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog upload-private", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	contentType := fs.String("content-type", "application/octet-stream", "File MIME type")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog upload-private [options]\n\nProduct SDK: CatalogService.UploadPrivate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	source, err := upload.open(*filename, *bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	defer source.Close()
	result, err := ctx.Client.Catalog(wid).UploadPrivate(context.Background(), source.Name, source.Reader, sdk.WithCatalogUploadContentType(*contentType))
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDataAsset(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog data-asset", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	assetID := fs.String("asset-id", "", "assetID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog data-asset [options]\n\nProduct SDK: CatalogService.DataAsset\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).DataAsset(context.Background(), *assetID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceBatchGetDataAssets(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog batch-get-data-assets", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	rawFileIDs := fs.String("raw-file-ids", "", "rawFileIDs (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog batch-get-data-assets [options]\n\nProduct SDK: CatalogService.BatchGetDataAssets\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).BatchGetDataAssets(context.Background(), splitCSV(*rawFileIDs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceFileDataAsset(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog file-data-asset", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fileID := fs.String("file-id", "", "fileID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog file-data-asset [options]\n\nProduct SDK: CatalogService.FileDataAsset\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).FileDataAsset(context.Background(), *fileID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceFileArtifact(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog file-artifact", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fileID := fs.String("file-id", "", "fileID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog file-artifact [options]\n\nProduct SDK: CatalogService.FileArtifact\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).FileArtifact(context.Background(), *fileID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceBatchResolveFileArtifacts(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog batch-resolve-file-artifacts", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fileIDs := fs.String("file-ids", "", "fileIDs (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog batch-resolve-file-artifacts [options]\n\nProduct SDK: CatalogService.BatchResolveFileArtifacts\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).BatchResolveFileArtifacts(context.Background(), splitCSV(*fileIDs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceCreateFolder(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog create-folder", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	parentVolumeID := fs.Int("parent-volume-id", 0, "parentVolumeID")
	databaseID := fs.Int("database-id", 0, "databaseID")
	name := fs.String("name", "", "name")
	description := fs.String("description", "", "Folder description")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog create-folder [options]\n\nProduct SDK: CatalogService.CreateFolder\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.CreateFolderOption
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "description" {
			options = append(options, sdk.WithFolderDescription(*description))
		}
	})
	result, err := ctx.Client.Catalog(wid).CreateFolder(context.Background(), *parentVolumeID, *databaseID, *name, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceUpdateFolder(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog update-folder", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	folderID := fs.Int("folder-id", 0, "folderID")
	name := fs.String("name", "", "Folder name")
	description := fs.String("description", "", "Folder description")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog update-folder [options]\n\nProduct SDK: CatalogService.UpdateFolder\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.UpdateFolderOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "name":
			options = append(options, sdk.WithFolderName(*name))
		case "description":
			options = append(options, sdk.WithUpdatedFolderDescription(*description))
		}
	})
	if err := ctx.Client.Catalog(wid).UpdateFolder(context.Background(), *folderID, options...); err != nil {
		return err
	}
	fmt.Printf("catalog update-folder: ok\n")
	return nil
}

func runCatalogServiceDeleteFolder(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog delete-folder", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	folderID := fs.Int("folder-id", 0, "folderID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog delete-folder [options]\n\nProduct SDK: CatalogService.DeleteFolder\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).DeleteFolder(context.Background(), *folderID); err != nil {
		return err
	}
	fmt.Printf("catalog delete-folder: ok\n")
	return nil
}

func runCatalogServiceCleanFolder(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog clean-folder", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	folderID := fs.Int("folder-id", 0, "folderID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog clean-folder [options]\n\nProduct SDK: CatalogService.CleanFolder\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.Catalog(wid).CleanFolder(context.Background(), *folderID); err != nil {
		return err
	}
	fmt.Printf("catalog clean-folder: ok\n")
	return nil
}

func runCatalogServiceRunNL2SQL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog run-nl2sql", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	databaseID := fs.Int("database-id", 0, "databaseID")
	sql := fs.String("sql", "", "sql")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog run-nl2sql [options]\n\nProduct SDK: CatalogService.RunNL2SQL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).RunNL2SQL(context.Background(), *databaseID, *sql)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServicePreviewWorkspaceFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog preview-workspace-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fileID := fs.String("file-id", "", "file-id")
	output := fs.String("output", "", "output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog preview-workspace-file [options]\n\nProduct SDK: CatalogService.PreviewWorkspaceFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).PreviewWorkspaceFile(context.Background(), *fileID)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}

func runCatalogServiceCatalogReferences(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog catalog-references", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	id := fs.Int64("id", 0, "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog catalog-references [options]\n\nProduct SDK: CatalogService.CatalogReferences\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).CatalogReferences(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceDatabaseReferences(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog database-references", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	id := fs.Int64("id", 0, "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog database-references [options]\n\nProduct SDK: CatalogService.DatabaseReferences\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).DatabaseReferences(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceVolumeReferences(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog volume-references", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	id := fs.Int64("id", 0, "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog volume-references [options]\n\nProduct SDK: CatalogService.VolumeReferences\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).VolumeReferences(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceFolderReferences(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog folder-references", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	id := fs.Int64("id", 0, "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog folder-references [options]\n\nProduct SDK: CatalogService.FolderReferences\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).FolderReferences(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceTableReferences(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog table-references", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	id := fs.Int64("id", 0, "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog table-references [options]\n\nProduct SDK: CatalogService.TableReferences\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).TableReferences(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceWorkspaceFileModelURL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog workspace-file-model-url", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fileID := fs.String("file-id", "", "file-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog workspace-file-model-url [options]\n\nProduct SDK: CatalogService.WorkspaceFileModelURL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).WorkspaceFileModelURL(context.Background(), *fileID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceGarbageCollection(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog garbage-collection", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog garbage-collection [options]\n\nProduct SDK: CatalogService.GarbageCollection\n\nOptions:\n")
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
	var request sdk.CatalogGarbageCollectionRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Catalog(wid).GarbageCollection(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceWorkspaceTableColumns(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog workspace-table-columns", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog workspace-table-columns [options]\n\nProduct SDK: CatalogService.WorkspaceTableColumns\n\nOptions:\n")
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
	var request sdk.WorkspaceTableColumnsRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Catalog(wid).WorkspaceTableColumns(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceWorkspaceDatabaseTables(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog workspace-database-tables", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	databaseID := fs.Int64("database-id", 0, "database-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog workspace-database-tables [options]\n\nProduct SDK: CatalogService.WorkspaceDatabaseTables\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).WorkspaceDatabaseTables(context.Background(), *databaseID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceWorkspaceTableOverview(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog workspace-table-overview", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog workspace-table-overview [options]\n\nProduct SDK: CatalogService.WorkspaceTableOverview\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).WorkspaceTableOverview(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runCatalogServiceCatalogFileContent(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog catalog-file-content", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	fileID := fs.String("file-id", "", "file-id")
	capability := fs.String("capability", "", "capability")
	output := fs.String("output", "", "Output file, or stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).CatalogFileContent(context.Background(), *fileID, *capability)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}
