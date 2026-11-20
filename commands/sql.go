package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteSQLCommand runs product SDK SQLService methods via moi-cli.
func ExecuteSQLCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowSQLHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "download-query-result":
		return runSQLServiceDownloadQueryResult(ctx, args[1:])
	case "connect-query-web-socket":
		return runSQLQueryWebSocketConnect(ctx, args[1:])
	case "execute":
		return runSQLExecuteAlias(ctx, args[1:])
	case "run":
		return runSQLRunQuery(ctx, args[1:])
	case "describe":
		return runSQLDescribeAlias(ctx, args[1:])
	case "kill":
		return runSQLKillAlias(ctx, args[1:])
	case "result":
		return runSQLResultAlias(ctx, args[1:])
	case "history":
		return runSQLHistoryAlias(ctx, args[1:])
	case "meta-tables":
		return runSQLMetaTablesAlias(ctx, args[1:])
	case "create-workbook":
		return runSQLServiceCreateWorkbook(ctx, args[1:])
	case "workbooks":
		return runSQLServiceWorkbooks(ctx, args[1:])
	case "update-workbook":
		return runSQLServiceUpdateWorkbook(ctx, args[1:])
	case "delete-workbook":
		return runSQLServiceDeleteWorkbook(ctx, args[1:])
	case "workbook-detail":
		return runSQLServiceWorkbookDetail(ctx, args[1:])
	case "workbook-versions":
		return runSQLServiceWorkbookVersions(ctx, args[1:])
	case "create-workbook-version":
		return runSQLServiceCreateWorkbookVersion(ctx, args[1:])
	case "update-workbook-version":
		return runSQLServiceUpdateWorkbookVersion(ctx, args[1:])
	case "save-workbook-version":
		return runSQLServiceSaveWorkbookVersion(ctx, args[1:])
	case "meta-databases":
		return runSQLServiceMetaDatabases(ctx, args[1:])
	case "meta-db-tree":
		return runSQLServiceMetaDBTree(ctx, args[1:])
	case "meta-db-tables":
		return runSQLServiceMetaDBTables(ctx, args[1:])
	case "meta-table-info":
		return runSQLServiceMetaTableInfo(ctx, args[1:])
	case "meta-table-columns":
		return runSQLServiceMetaTableColumns(ctx, args[1:])
	case "meta-table-column-stats":
		return runSQLServiceMetaTableColumnStats(ctx, args[1:])
	case "meta-table-ddl":
		return runSQLServiceMetaTableDDL(ctx, args[1:])
	case "drop-meta-database":
		return runSQLServiceDropMetaDatabase(ctx, args[1:])
	case "drop-meta-table":
		return runSQLServiceDropMetaTable(ctx, args[1:])
	case "drop-meta-view":
		return runSQLServiceDropMetaView(ctx, args[1:])
	case "execute-query":
		return runSQLServiceExecuteQuery(ctx, args[1:])
	case "describe-query":
		return runSQLServiceDescribeQuery(ctx, args[1:])
	case "kill-query":
		return runSQLServiceKillQuery(ctx, args[1:])
	case "query-result":
		return runSQLServiceQueryResult(ctx, args[1:])
	case "query-history-overview":
		return runSQLServiceQueryHistoryOverview(ctx, args[1:])
	case "query-history":
		return runSQLServiceQueryHistory(ctx, args[1:])
	case "query-detail":
		return runSQLServiceQueryDetail(ctx, args[1:])
	case "query-profile":
		return runSQLServiceQueryProfile(ctx, args[1:])
	case "query-web-socket-url":
		return runSQLServiceQueryWebSocketURL(ctx, args[1:])
	default:
		return fmt.Errorf("unknown sql subcommand: %s", sub)
	}
}

// ShowSQLHelp prints help for sql.
func ShowSQLHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: sql <subcommand> [options]

Product SDK surface: SQLService

Subcommands:
  download-query-result           SQLService.DownloadQueryResult
  connect-query-web-socket          SQLService.ConnectQueryWebSocket
  execute                          (alias)
  run                              SQLService.RunQuery (inline result)
  describe                         (alias)
  kill                             (alias)
  result                           (alias)
  history                          (alias)
  meta-tables                      (alias)
  create-workbook                  SQLService.CreateWorkbook
  workbooks                        SQLService.Workbooks
  update-workbook                  SQLService.UpdateWorkbook
  delete-workbook                  SQLService.DeleteWorkbook
  workbook-detail                  SQLService.WorkbookDetail
  workbook-versions                SQLService.WorkbookVersions
  create-workbook-version          SQLService.CreateWorkbookVersion
  update-workbook-version          SQLService.UpdateWorkbookVersion
  save-workbook-version            SQLService.SaveWorkbookVersion
  meta-databases                   SQLService.MetaDatabases
  meta-db-tree                     SQLService.MetaDBTree
  meta-db-tables                   SQLService.MetaDBTables
  meta-table-info                  SQLService.MetaTableInfo
  meta-table-columns               SQLService.MetaTableColumns
  meta-table-column-stats          SQLService.MetaTableColumnStats
  meta-table-ddl                   SQLService.MetaTableDDL
  drop-meta-database               SQLService.DropMetaDatabase
  drop-meta-table                  SQLService.DropMetaTable
  drop-meta-view                   SQLService.DropMetaView
  execute-query                    SQLService.ExecuteQuery
  describe-query                   SQLService.DescribeQuery
  kill-query                       SQLService.KillQuery
  query-result                     SQLService.QueryResult
  query-history-overview           SQLService.QueryHistoryOverview
  query-history                    SQLService.QueryHistory
  query-detail                     SQLService.QueryDetail
  query-profile                    SQLService.QueryProfile
  query-web-socket-url             SQLService.QueryWebSocketURL

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runSQLServiceCreateWorkbook(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql create-workbook", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql create-workbook [options]\n\nProduct SDK: SQLService.CreateWorkbook\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).CreateWorkbook(context.Background(), *name)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceWorkbooks(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql workbooks", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("search", "", "Workbook name search")
	fs.Int("page-size", 0, "Page size")
	fs.String("page-token", "", "Opaque cursor from the previous page")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql workbooks [options]\n\nProduct SDK: SQLService.Workbooks\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.WorkbookListOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "search":
			opts = append(opts, sdk.WithWorkbookSearch(f.Value.String()))
		case "page-size":
			opts = append(opts, sdk.WithWorkbookPageSize(f.Value.(flag.Getter).Get().(int)))
		case "page-token":
			opts = append(opts, sdk.WithWorkbookPageToken(f.Value.String()))
		}
	})
	result, err := ctx.Client.SQL(wid).Workbooks(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceUpdateWorkbook(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql update-workbook", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workbookID := fs.String("workbook-id", "", "workbookID")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql update-workbook [options]\n\nProduct SDK: SQLService.UpdateWorkbook\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.SQL(wid).UpdateWorkbook(context.Background(), *workbookID, *name); err != nil {
		return err
	}
	fmt.Printf("sql update-workbook: ok\n")
	return nil
}

func runSQLServiceDeleteWorkbook(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql delete-workbook", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workbookID := fs.String("workbook-id", "", "workbookID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql delete-workbook [options]\n\nProduct SDK: SQLService.DeleteWorkbook\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.SQL(wid).DeleteWorkbook(context.Background(), *workbookID); err != nil {
		return err
	}
	fmt.Printf("sql delete-workbook: ok\n")
	return nil
}

func runSQLServiceWorkbookDetail(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql workbook-detail", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workbookID := fs.String("workbook-id", "", "workbookID")
	versionID := fs.String("version-id", "", "versionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql workbook-detail [options]\n\nProduct SDK: SQLService.WorkbookDetail\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).WorkbookDetail(context.Background(), *workbookID, *versionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceWorkbookVersions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql workbook-versions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workbookID := fs.String("workbook-id", "", "workbookID")
	fs.Int("page-size", 0, "Page size")
	fs.String("page-token", "", "Opaque cursor from the previous page")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql workbook-versions [options]\n\nProduct SDK: SQLService.WorkbookVersions\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.WorkbookVersionListOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "page-size":
			opts = append(opts, sdk.WithWorkbookVersionPageSize(f.Value.(flag.Getter).Get().(int)))
		case "page-token":
			opts = append(opts, sdk.WithWorkbookVersionPageToken(f.Value.String()))
		}
	})
	result, err := ctx.Client.SQL(wid).WorkbookVersions(context.Background(), *workbookID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceCreateWorkbookVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql create-workbook-version", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workbookID := fs.String("workbook-id", "", "workbookID")
	sqlContent := fs.String("sql-content", "", "sqlContent")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql create-workbook-version [options]\n\nProduct SDK: SQLService.CreateWorkbookVersion\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).CreateWorkbookVersion(context.Background(), *workbookID, *sqlContent)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceUpdateWorkbookVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql update-workbook-version", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workbookID := fs.String("workbook-id", "", "workbookID")
	versionID := fs.String("version-id", "", "versionID")
	sqlContent := fs.String("sql-content", "", "sqlContent")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql update-workbook-version [options]\n\nProduct SDK: SQLService.UpdateWorkbookVersion\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.SQL(wid).UpdateWorkbookVersion(context.Background(), *workbookID, *versionID, *sqlContent); err != nil {
		return err
	}
	fmt.Printf("sql update-workbook-version: ok\n")
	return nil
}

func runSQLServiceSaveWorkbookVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql save-workbook-version", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workbookID := fs.String("workbook-id", "", "workbookID")
	versionID := fs.String("version-id", "", "versionID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql save-workbook-version [options]\n\nProduct SDK: SQLService.SaveWorkbookVersion\n\nOptions:\n")
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
	var sqlContent *string
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &sqlContent); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	if err := ctx.Client.SQL(wid).SaveWorkbookVersion(context.Background(), *workbookID, *versionID, sqlContent); err != nil {
		return err
	}
	fmt.Printf("sql save-workbook-version: ok\n")
	return nil
}

func runSQLServiceMetaDatabases(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql meta-databases", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql meta-databases [options]\n\nProduct SDK: SQLService.MetaDatabases\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).MetaDatabases(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceMetaDBTree(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql meta-db-tree", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql meta-db-tree [options]\n\nProduct SDK: SQLService.MetaDBTree\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).MetaDBTree(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceMetaDBTables(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql meta-db-tables", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql meta-db-tables [options]\n\nProduct SDK: SQLService.MetaDBTables\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).MetaDBTables(context.Background(), *dbName)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceMetaTableInfo(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql meta-table-info", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	tableName := fs.String("table-name", "", "tableName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql meta-table-info [options]\n\nProduct SDK: SQLService.MetaTableInfo\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).MetaTableInfo(context.Background(), *dbName, *tableName)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceMetaTableColumns(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql meta-table-columns", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	tableName := fs.String("table-name", "", "tableName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql meta-table-columns [options]\n\nProduct SDK: SQLService.MetaTableColumns\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).MetaTableColumns(context.Background(), *dbName, *tableName)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceMetaTableColumnStats(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql meta-table-column-stats", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	tableName := fs.String("table-name", "", "tableName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql meta-table-column-stats [options]\n\nProduct SDK: SQLService.MetaTableColumnStats\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).MetaTableColumnStats(context.Background(), *dbName, *tableName)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceMetaTableDDL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql meta-table-ddl", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	tableName := fs.String("table-name", "", "tableName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql meta-table-ddl [options]\n\nProduct SDK: SQLService.MetaTableDDL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).MetaTableDDL(context.Background(), *dbName, *tableName)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceDropMetaDatabase(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql drop-meta-database", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql drop-meta-database [options]\n\nProduct SDK: SQLService.DropMetaDatabase\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.SQL(wid).DropMetaDatabase(context.Background(), *dbName); err != nil {
		return err
	}
	fmt.Printf("sql drop-meta-database: ok\n")
	return nil
}

func runSQLServiceDropMetaTable(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql drop-meta-table", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	tableName := fs.String("table-name", "", "tableName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql drop-meta-table [options]\n\nProduct SDK: SQLService.DropMetaTable\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.SQL(wid).DropMetaTable(context.Background(), *dbName, *tableName); err != nil {
		return err
	}
	fmt.Printf("sql drop-meta-table: ok\n")
	return nil
}

func runSQLServiceDropMetaView(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql drop-meta-view", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	dbName := fs.String("db-name", "", "dbName")
	tableName := fs.String("table-name", "", "tableName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql drop-meta-view [options]\n\nProduct SDK: SQLService.DropMetaView\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.SQL(wid).DropMetaView(context.Background(), *dbName, *tableName); err != nil {
		return err
	}
	fmt.Printf("sql drop-meta-view: ok\n")
	return nil
}

func runSQLServiceExecuteQuery(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql execute-query", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	query := fs.String("query", "", "query")
	fs.String("db-name", "", "Database name")
	addSQLQueryFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql execute-query [options]\n\nProduct SDK: SQLService.ExecuteQuery\n\nStarts an asynchronous query and returns query_id, not statement_id.\nPoll sql describe-query --query-id <query_id> until status is success,\nthen pass its statement_id to sql query-result. A failed query has no result.\nUse sql run for a query that returns rows inline.\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).ExecuteQuery(context.Background(), *query, sqlQueryCLIOptions(fs, "db-name")...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceDescribeQuery(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql describe-query", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	queryID := fs.String("query-id", "", "query_id returned by sql execute-query")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql describe-query [options]\n\nProduct SDK: SQLService.DescribeQuery\n\nReturns query status and, after successful execution, statement_id for sql query-result.\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).DescribeQuery(context.Background(), *queryID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceKillQuery(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql kill-query", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	queryID := fs.String("query-id", "", "queryID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql kill-query [options]\n\nProduct SDK: SQLService.KillQuery\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.SQL(wid).KillQuery(context.Background(), *queryID); err != nil {
		return err
	}
	fmt.Printf("sql kill-query: ok\n")
	return nil
}

func runSQLServiceQueryResult(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql query-result", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	statementID := fs.String("statement-id", "", "statementID")
	fs.Int("offset", 0, "Result offset")
	fs.Int("limit", 0, "Result limit")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql query-result [options]\n\nProduct SDK: SQLService.QueryResult\n\nReads a successful query's saved rows. --statement-id is the statement_id\nfrom sql describe-query, not the query_id returned by sql execute-query.\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.SQLResultOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "offset":
			opts = append(opts, sdk.WithSQLResultOffset(f.Value.(flag.Getter).Get().(int)))
		case "limit":
			opts = append(opts, sdk.WithSQLResultLimit(f.Value.(flag.Getter).Get().(int)))
		}
	})
	result, err := ctx.Client.SQL(wid).QueryResult(context.Background(), *statementID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceQueryHistoryOverview(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql query-history-overview", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	addSQLHistoryRangeFlags(fs)
	addSQLHistoryFilterFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql query-history-overview [options]\n\nProduct SDK: SQLService.QueryHistoryOverview\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts, err := sqlHistoryOverviewCLIOptions(fs)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).QueryHistoryOverview(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceQueryHistory(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql query-history", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	addSQLHistoryRangeFlags(fs)
	addSQLHistoryFilterFlags(fs)
	fs.String("status", "", "Statement status")
	fs.String("users", "", "Comma-separated user names")
	fs.String("databases", "", "Comma-separated database names")
	fs.String("statement", "", "Statement search pattern")
	fs.String("query-types", "", "Comma-separated query types")
	fs.Uint64("min-duration", 0, "Minimum query duration")
	fs.String("statement-id", "", "Statement ID")
	fs.String("transaction-id", "", "Transaction ID")
	fs.String("session-id", "", "Session ID")
	fs.String("selected-fields", "", "Comma-separated result field names")
	fs.Uint("offset", 0, "Result offset")
	fs.Uint("limit", 0, "Result limit")
	fs.String("order-by", "", "Comma-separated order fields")
	fs.String("order", "", "Order direction")
	fs.Uint("cu", 0, "Compute usage filter (CU)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql query-history [options]\n\nProduct SDK: SQLService.QueryHistory\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts, err := sqlHistoryCLIOptions(fs)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).QueryHistory(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceQueryDetail(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql query-detail", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	statementID := fs.String("statement-id", "", "statementID")
	addSQLHistoryRangeFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql query-detail [options]\n\nProduct SDK: SQLService.QueryDetail\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).QueryDetail(context.Background(), *statementID, sqlRecordCLIOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceQueryProfile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql query-profile", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	statementID := fs.String("statement-id", "", "statementID")
	addSQLHistoryRangeFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql query-profile [options]\n\nProduct SDK: SQLService.QueryProfile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).QueryProfile(context.Background(), *statementID, sqlRecordCLIOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLServiceQueryWebSocketURL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql query-web-socket-url", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli sql query-web-socket-url [options]\n\nProduct SDK: SQLService.QueryWebSocketURL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).QueryWebSocketURL()
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLQueryWebSocketConnect(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql connect-query-web-socket", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli sql connect-query-web-socket [options]")
		fmt.Fprintln(os.Stderr, "Performs a handshake probe, prints negotiated protocols and closes; it is not an interactive SQL session.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	connection, err := ctx.Client.SQL(wid).ConnectQueryWebSocket(context.Background())
	if err != nil {
		return err
	}
	defer connection.Close()
	return printJSON(map[string]any{"protocols": connection.Protocols()})
}
