package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runConnectorServicePreviewFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector preview-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	connectorID := fs.String("connector-id", "", "Source connector ID")
	uri := fs.String("uri", "", "Source file URI")
	connFileID := fs.String("conn-file-id", "", "Temporary file ID returned by local-upload or preview-file")
	sheetName := fs.String("sheet-name", "", "Spreadsheet sheet name")
	hasHeader := fs.Bool("has-header", false, "Use a row as column names")
	headerRow := fs.Int("header-row", 0, "Column-name row (1-based)")
	rowStart := fs.Int("row-start", 0, "First data row (1-based)")
	fileType := fs.Int("file-type", 0, "Product file type code; omit to use the stored file type")
	separator := fs.String("csv-separator", "", "CSV field separator; provide all three csv-* options together")
	delimiter := fs.String("csv-delimiter", "", "CSV quote character (empty disables quoting)")
	escape := fs.Bool("csv-escape", false, "Use backslash escapes in CSV quoted fields")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector preview-file [options]\n\nProduct SDK: ConnectorService.PreviewFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	var opts []sdk.ConnectorPreviewOption
	provided := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
		switch f.Name {
		case "conn-file-id":
			opts = append(opts, sdk.WithConnectorPreviewConnFileID(*connFileID))
		case "sheet-name":
			opts = append(opts, sdk.WithConnectorPreviewSheetName(*sheetName))
		case "has-header":
			opts = append(opts, sdk.WithConnectorPreviewColumnName(*hasHeader))
		case "header-row":
			opts = append(opts, sdk.WithConnectorPreviewColumnNameRow(*headerRow))
		case "row-start":
			opts = append(opts, sdk.WithConnectorPreviewRowStart(*rowStart))
		case "file-type":
			opts = append(opts, sdk.WithConnectorPreviewFileType(*fileType))
		}
	})
	if provided["csv-separator"] || provided["csv-delimiter"] || provided["csv-escape"] {
		if !provided["csv-separator"] || !provided["csv-delimiter"] || !provided["csv-escape"] {
			return &LocalInputError{Code: "CLI_CONNECTOR_CSV_INCOMPLETE"}
		}
		opts = append(opts, sdk.WithConnectorPreviewCSV(*separator, *delimiter, *escape))
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).PreviewFile(context.Background(), *connectorID, *uri, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
