package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

const (
	runtimeImageAnalysisCLIFileLimit = 650 * 1024
	runtimeImageAnalysisCLIQuestion  = 4096
	runtimeImageAnalysisCLITimeout   = 180 * time.Second
)

// ExecuteFileCommand uses the Product SDK and preserves sandbox runtime proof.
func ExecuteFileCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowFileHelp(nil)
		return nil
	}
	switch args[0] {
	case "list":
		return runWorkspaceFileList(ctx, args[1:])
	case "upload":
		return runRuntimeArtifactUpload(ctx, args[1:], "file upload", true)
	case "download":
		return runRuntimeWorkspaceFileDownload(ctx, args[1:], "file download")
	case "analyze-image":
		return runRuntimeImageAnalysis(ctx, args[1:], "file analyze-image")
	default:
		return fmt.Errorf("unknown file subcommand: %s", args[0])
	}
}

// ShowFileHelp prints the workspace and child-runtime file contract.
func ShowFileHelp(_ []string) {
	fmt.Fprint(os.Stderr, `Usage: file <subcommand> [options]

Task-scoped Product SDK surface. The sandbox supplies the short-lived runtime
authorization and technical API key through its environment.

Subcommands:
  list        Discover IAM-visible file placements in the workspace
  upload      Upload and publish a file to the active runtime task (--publish-artifact required)
  download    Download a task-visible workspace file
  analyze-image Analyze a prepared JPEG using the active sandbox identity

Examples:
  moi-cli -o file list --name report --extensions pdf,csv --page-size 100
  moi-cli file upload --file ./report.pdf --publish-artifact
  moi-cli file download --file-id abc123 --output ./downloaded.pdf
  moi-cli file analyze-image --file ./prepared.jpg --question "What is shown?"
`)
}

func runWorkspaceFileList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("file list", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config or MOI_WORKSPACE_ID)")
	name := fs.String("name", "", "File name fragment")
	fileIDs := fs.String("file-ids", "", "Comma-separated File IDs")
	extensions := fs.String("extensions", "", "Comma-separated file extensions without dots")
	pageSize := fs.Int("page-size", 100, "Page size (1-100)")
	pageToken := fs.String("page-token", "", "Next page token")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli file list [options]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *pageSize <= 0 || *pageSize > 100 {
		return fmt.Errorf("--page-size must be between 1 and 100")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := []sdk.QueryOption{sdk.WithPageSize(*pageSize)}
	if *pageToken != "" {
		opts = append(opts, sdk.WithPageToken(*pageToken))
	}
	if *name != "" {
		opts = append(opts, sdk.WithQueryParam("filter_file_name_fuzzy", *name))
	}
	for _, filter := range []struct {
		key, raw  string
		extension bool
	}{
		{key: "filter_file_id", raw: *fileIDs}, {key: "filter_file_ext", raw: *extensions, extension: true},
	} {
		if filter.raw == "" {
			continue
		}
		for _, raw := range strings.Split(filter.raw, ",") {
			value := strings.TrimSpace(raw)
			if value == "" || (filter.extension && strings.HasPrefix(value, ".")) {
				return fmt.Errorf("invalid value %q for %s", raw, filter.key)
			}
			opts = append(opts, sdk.WithQueryValues(filter.key, value))
		}
	}
	result, err := ctx.Client.Catalog(wid).ListWorkspaceFiles(context.Background(), opts...)
	if err != nil {
		return err
	}
	// Keep items present even for an empty list so the skill's JSON consumer
	// can iterate it without inventing a separate empty-result protocol.
	if result.Items == nil {
		result.Items = []*sdk.WorkspaceFilePlacement{}
	}
	return printJSONOrTable(ctx, struct {
		Items         []*sdk.WorkspaceFilePlacement `json:"items"`
		Total         int64                         `json:"total"`
		NextPageToken string                        `json:"next_page_token,omitempty"`
	}{Items: result.Items, Total: result.Total, NextPageToken: result.NextPageToken}, func() error {
		for _, file := range result.Items {
			if _, err := fmt.Fprintf(os.Stdout, "%d\t%s\t%s\t%s\t%d\n", file.VolumeId, file.FileId, file.FileName, file.FilePath, file.Size); err != nil {
				return err
			}
		}
		if result.NextPageToken != "" {
			_, err = fmt.Fprintf(os.Stdout, "NextPageToken: %s\n", result.NextPageToken)
		}
		return err
	})
}

func runRuntimeArtifactUpload(ctx *Context, args []string, commandName string, requirePublishArtifact bool) error {
	fs := flag.NewFlagSet(commandName, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config or MOI_WORKSPACE_ID)")
	filePath := fs.String("file", "", "Local file path to upload")
	name := fs.String("name", "", "Override filename (optional)")
	publishArtifact := fs.Bool("publish-artifact", false, "Publish the uploaded file to the active runtime task")
	_ = fs.Bool("yes", false, "Accepted for sandbox command compatibility")
	fs.Usage = func() {
		if requirePublishArtifact {
			fmt.Fprintf(os.Stderr, "Usage: moi-cli file upload --file <path> [--name <filename>] --publish-artifact [--workspace-id <id>]\n")
		} else {
			fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog upload-runtime-artifact --file <path> [--name <filename>] [--workspace-id <id>]\n")
		}
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if requirePublishArtifact && !*publishArtifact {
		return fmt.Errorf("--publish-artifact is required")
	}
	if *filePath == "" {
		return fmt.Errorf("--file is required")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	file, err := os.Open(*filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	filename := *name
	if filename == "" {
		filename = filepath.Base(*filePath)
	}
	result, err := ctx.Client.Catalog(wid).UploadRuntimeArtifact(context.Background(), filename, file)
	if err != nil {
		return err
	}
	if ctx.OutputJSON {
		return json.NewEncoder(os.Stdout).Encode(result)
	}
	_, err = fmt.Fprintf(os.Stdout, "FileID: %s\nOriginalName: %s\nSize: %d\nMD5: %s\nArtifactID: %s\nSHA256: %s\n",
		result.GetFileId(), result.GetOriginalName(), result.GetSize(), result.GetMd5(), result.GetArtifactId(), result.GetSha256())
	return err
}

func runRuntimeWorkspaceFileDownload(ctx *Context, args []string, commandName string) error {
	fs := flag.NewFlagSet(commandName, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config or MOI_WORKSPACE_ID)")
	fileID := fs.String("file-id", "", "File ID to download")
	outPath := fs.String("output", "", "Output file path (default stdout)")
	_ = fs.Bool("yes", false, "Accepted for sandbox command compatibility")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli %s --file-id <id> [--output <path>] [--workspace-id <id>]\n", commandName)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *fileID == "" {
		return fmt.Errorf("--file-id is required")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	stream, err := ctx.Client.Catalog(wid).DownloadWorkspaceFile(context.Background(), *fileID)
	if err != nil {
		return err
	}
	defer stream.Body.Close()
	return writeStream(stream.Body, *outPath)
}

func runRuntimeImageAnalysis(ctx *Context, args []string, commandName string) error {
	fs := flag.NewFlagSet(commandName, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config or MOI_WORKSPACE_ID)")
	filePath := fs.String("file", "", "Prepared local JPEG image")
	question := fs.String("question", "", "Image analysis question")
	_ = fs.Bool("yes", false, "Accepted for sandbox command compatibility")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli %s --file <prepared.jpg> --question <question> [--workspace-id <id>]\n", commandName)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *filePath == "" {
		return fmt.Errorf("--file is required")
	}
	preparedQuestion := strings.TrimSpace(*question)
	if !utf8.ValidString(preparedQuestion) || preparedQuestion == "" || utf8.RuneCountInString(preparedQuestion) > runtimeImageAnalysisCLIQuestion {
		return fmt.Errorf("question must be valid UTF-8, nonempty, and at most %d characters", runtimeImageAnalysisCLIQuestion)
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	file, err := os.Open(*filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, runtimeImageAnalysisCLIFileLimit+1))
	if err != nil {
		return fmt.Errorf("read prepared image: %w", err)
	}
	if len(data) == 0 || len(data) > runtimeImageAnalysisCLIFileLimit {
		return fmt.Errorf("prepared image must be nonempty and at most %d KiB", runtimeImageAnalysisCLIFileLimit/1024)
	}
	requestCtx, cancel := context.WithTimeout(context.Background(), runtimeImageAnalysisCLITimeout)
	defer cancel()
	result, err := ctx.Client.Catalog(wid).AnalyzeRuntimeImage(requestCtx, filepath.Base(*filePath), bytes.NewReader(data), preparedQuestion)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]string{"answer": result.GetAnswer()})
}
