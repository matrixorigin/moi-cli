package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
)

func runSQLServiceDownloadQueryResult(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql download-query-result", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	statementID := fs.String("statement-id", "", "Statement ID")
	output := fs.String("output", "", "Write CSV to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli sql download-query-result [options]\n\nProduct SDK: SQLService.DownloadQueryResult\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.SQL(wid).DownloadQueryResult(context.Background(), *statementID)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}
