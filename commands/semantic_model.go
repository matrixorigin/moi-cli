package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
)

// Both import commands accept the same SDK input. Keep their help and offline
// schema guidance together; business validation remains owned by the server.
func semanticModelSourceGuidance() map[string]any {
	return map[string]any{
		"input": "A JSON array of source objects, using the SDK field names shown below. This is the --json value, not an object wrapping sources.",
		"source_types": map[string]any{
			"local_file": map[string]any{
				"required": []string{"SourceType", "FileID", "FileName"},
				"notes":    "Upload first with semantic-model upload-local-file (or upload-model-local-file). FileID is the returned file_id; FileName includes its extension. UploadKind may be omitted for unstructured documents such as .md, .txt and .pdf.",
			},
			"catalog_file": map[string]any{
				"required": []string{"SourceType", "FileID"},
				"notes":    "Use an existing Catalog file. Supply VolumeID for its authoritative volume, especially for shared files; omission uses the server's file lookup.",
			},
			"catalog_table": map[string]any{
				"required": []string{"SourceType", "TableID"},
				"notes":    "TableID must be a positive existing Catalog table ID, not a table name.",
			},
		},
		"upload_kinds": map[string]string{
			"unstructured": "local_file document parsing; the default when UploadKind is omitted. Leave TableConfig empty.",
			"structured":   "local_file structured import; requires TableConfig as a JSON-encoded string containing the structured-load table configuration, not a nested JSON object.",
		},
		"examples": map[string][]sdk.SemanticModelSourceInput{
			"local_file":    {{SourceType: "local_file", FileID: "uploaded-file-id", FileName: "manual.md"}},
			"catalog_file":  {{SourceType: "catalog_file", FileID: "catalog-file-id", VolumeID: 29}},
			"catalog_table": {{SourceType: "catalog_table", TableID: 42}},
		},
		"verification": "Inspect the returned sources and jobs, then use semantic-model list-sources and list-source-jobs to check ingestion. An accepted request does not mean indexing has completed.",
	}
}

func printSemanticModelSourceGuidance(output io.Writer) {
	_, _ = fmt.Fprintln(output, "\nSource input guide (also available via moi-cli schema semantic-model COMMAND):")
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(semanticModelSourceGuidance()); err != nil {
		fmt.Fprintln(output, err)
	}
}

func semanticModelEmptyCLIOptions(fs *flag.FlagSet, description string, imageIndexEnabled bool) []sdk.SemanticModelEmptyOption {
	var opts []sdk.SemanticModelEmptyOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "description":
			opts = append(opts, sdk.WithSemanticModelEmptyDescription(description))
		case "image-index-enabled":
			opts = append(opts, sdk.WithSemanticModelEmptyImageIndexEnabled(imageIndexEnabled))
		}
	})
	return opts
}

// ExecuteSemanticModelCommand runs product SDK KnowledgeService methods via moi-cli.
func ExecuteSemanticModelCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowSemanticModelHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "find-files":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "search":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "list-passages":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "read-source":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "read-document":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "search-document":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "search-images":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "read-evidence":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "describe-tables":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "query":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "cite":
		return runKnowledgeRuntime(ctx, sub, args[1:])
	case "import-models":
		return runKnowledgeServiceImportModels(ctx, args[1:])
	case "import-model":
		return runKnowledgeServiceImportModel(ctx, args[1:])
	case "preview-artifact":
		return runKnowledgeServicePreviewArtifact(ctx, args[1:])
	case "preview-source-file":
		return runKnowledgeServicePreviewSourceFile(ctx, args[1:])
	case "list":
		return runKnowledgeServiceList(ctx, args[1:])
	case "list-tags":
		return runKnowledgeServiceListTags(ctx, args[1:])
	case "create":
		return runKnowledgeServiceCreate(ctx, args[1:])
	case "create-with-spec":
		return runKnowledgeServiceCreateWithSpec(ctx, args[1:])
	case "get":
		return runKnowledgeServiceGet(ctx, args[1:])
	case "update":
		return runKnowledgeServiceUpdate(ctx, args[1:])
	case "delete":
		return runKnowledgeServiceDelete(ctx, args[1:])
	case "create-with-sources":
		return runKnowledgeServiceCreateWithSources(ctx, args[1:])
	case "create-empty":
		return runKnowledgeServiceCreateEmpty(ctx, args[1:])
	case "upload-local-file":
		return runKnowledgeServiceUploadLocalFile(ctx, args[1:])
	case "upload-model-local-file":
		return runKnowledgeServiceUploadModelLocalFile(ctx, args[1:])
	case "preview-source-selection-counts":
		return runKnowledgeServicePreviewSourceSelectionCounts(ctx, args[1:])
	case "preview-model-source-selection-counts":
		return runKnowledgeServicePreviewModelSourceSelectionCounts(ctx, args[1:])
	case "list-sources":
		return runKnowledgeServiceListSources(ctx, args[1:])
	case "list-referencing-agents":
		return runKnowledgeServiceListReferencingAgents(ctx, args[1:])
	case "list-sources-page":
		return runKnowledgeServiceListSourcesPage(ctx, args[1:])
	case "check-source-existence":
		return runKnowledgeServiceCheckSourceExistence(ctx, args[1:])
	case "append-sources":
		return runKnowledgeServiceAppendSources(ctx, args[1:])
	case "backfill-legacy-sources":
		return runKnowledgeServiceBackfillLegacySources(ctx, args[1:])
	case "delete-source":
		return runKnowledgeServiceDeleteSource(ctx, args[1:])
	case "get-source-document":
		return runKnowledgeServiceGetSourceDocument(ctx, args[1:])
	case "update-source-governance":
		return runKnowledgeServiceUpdateSourceGovernance(ctx, args[1:])
	case "update-segment":
		return runKnowledgeServiceUpdateSegment(ctx, args[1:])
	case "delete-segment":
		return runKnowledgeServiceDeleteSegment(ctx, args[1:])
	case "create-segment":
		return runKnowledgeServiceCreateSegment(ctx, args[1:])
	case "update-segment-enabled":
		return runKnowledgeServiceUpdateSegmentEnabled(ctx, args[1:])
	case "reembed-segments":
		return runKnowledgeServiceReembedSegments(ctx, args[1:])
	case "list-source-jobs":
		return runKnowledgeServiceListSourceJobs(ctx, args[1:])
	case "reconcile-source-jobs":
		return runKnowledgeServiceReconcileSourceJobs(ctx, args[1:])
	case "list-entries":
		return runKnowledgeServiceListEntries(ctx, args[1:])
	case "create-entry":
		return runKnowledgeServiceCreateEntry(ctx, args[1:])
	case "update-entry":
		return runKnowledgeServiceUpdateEntry(ctx, args[1:])
	case "delete-entry":
		return runKnowledgeServiceDeleteEntry(ctx, args[1:])
	case "import":
		return runKnowledgeServiceImport(ctx, args[1:])
	case "export":
		return runKnowledgeServiceExport(ctx, args[1:])
	case "validate":
		return runKnowledgeServiceValidate(ctx, args[1:])
	case "search-drawings":
		return runKnowledgeServiceSearchDrawings(ctx, args[1:])
	case "get-drawing-provenance":
		return runKnowledgeServiceGetDrawingProvenance(ctx, args[1:])
	default:
		return fmt.Errorf("unknown semantic-model subcommand: %s", sub)
	}
}

// ShowSemanticModelHelp prints help for semantic-model.
func ShowSemanticModelHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: semantic-model <subcommand> [options]

Product SDK surface: KnowledgeService

Subcommands:
  find-files / search / list-passages / read-source
  read-document / search-document / search-images
  read-evidence / describe-tables / query / cite
                                   Task-scoped knowledge retrieval and citations
  import-models                    KnowledgeService.ImportModels
  import-model                     KnowledgeService.ImportModel
  preview-artifact                 KnowledgeService.PreviewArtifact
  preview-source-file              KnowledgeService.PreviewSourceFile
  list                             KnowledgeService.List
  list-tags                        KnowledgeService.ListTags
  create                           KnowledgeService.CreateEmpty
  create-with-spec                 KnowledgeService.CreateWithSpec
  get                              KnowledgeService.Get
  update                           KnowledgeService.Update
  delete                           KnowledgeService.Delete
  create-with-sources              KnowledgeService.CreateWithSources
  create-empty                     KnowledgeService.CreateEmpty
  upload-local-file                KnowledgeService.UploadLocalFile
  upload-model-local-file          KnowledgeService.UploadModelLocalFile
  preview-source-selection-counts  KnowledgeService.PreviewSourceSelectionCounts
  preview-model-source-selection-counts KnowledgeService.PreviewModelSourceSelectionCounts
  list-sources                     KnowledgeService.ListSources
  list-referencing-agents          KnowledgeService.ListReferencingAgents
  list-sources-page                KnowledgeService.ListSourcesPage
  check-source-existence           KnowledgeService.CheckSourceExistence
  append-sources                   KnowledgeService.AppendSources
  backfill-legacy-sources          KnowledgeService.BackfillLegacySources
  delete-source                    KnowledgeService.DeleteSource
  get-source-document              KnowledgeService.GetSourceDocument
  update-source-governance         KnowledgeService.UpdateSourceGovernance
  update-segment                   KnowledgeService.UpdateSegment
  delete-segment                   KnowledgeService.DeleteSegment
  create-segment                   KnowledgeService.CreateSegment
  update-segment-enabled           KnowledgeService.UpdateSegmentEnabled
  reembed-segments                 KnowledgeService.ReembedSegments
  list-source-jobs                 KnowledgeService.ListSourceJobs
  reconcile-source-jobs            KnowledgeService.ReconcileSourceJobs
  list-entries                     KnowledgeService.ListEntries
  create-entry                     KnowledgeService.CreateEntry
  update-entry                     KnowledgeService.UpdateEntry
  delete-entry                     KnowledgeService.DeleteEntry
  import                           KnowledgeService.Import
  export                           KnowledgeService.Export
  validate                         KnowledgeService.Validate
  search-drawings                  KnowledgeService.SearchDrawings
  get-drawing-provenance           KnowledgeService.GetDrawingProvenance

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runKnowledgeServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model list", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model list [options]\n\nProduct SDK: KnowledgeService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).List(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceListTags(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model list-tags", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model list-tags [options]\n\nProduct SDK: KnowledgeService.ListTags\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).ListTags(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model create", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "knowledge base name and immutable Catalog database name")
	description := fs.String("description", "", "Knowledge base description")
	imageIndexEnabled := fs.Bool("image-index-enabled", false, "Enable image vector index")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model create [options]\n\nProduct SDK: KnowledgeService.CreateEmpty\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := semanticModelEmptyCLIOptions(fs, *description, *imageIndexEnabled)
	result, err := ctx.Client.Knowledge(wid).CreateEmpty(context.Background(), *name, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result.GetModel(), func() error { return printJSON(result.GetModel()) })
}

func runKnowledgeServiceCreateWithSpec(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model create-with-spec", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON creation spec with name and embedding/drawing controls; see moi-cli schema semantic-model create-with-spec")
	bodyFile := fs.String("json-file", "", "JSON file containing the creation spec")
	description := fs.String("description", "", "Knowledge base description")
	tablesFile := fs.String("tables-file", "", "JSON file containing semantic table definitions")
	filesFile := fs.String("files-file", "", "JSON file containing semantic file configuration")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli semantic-model create-with-spec [options]\n\nProduct SDK: KnowledgeService.CreateWithSpec\n\nOptions:")
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
	var req sdk.SemanticModelCreateSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	opts, err := semanticModelCLIOptions(fs, *description, *tablesFile, *filesFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).CreateWithSpec(context.Background(), req, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model get", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model get [options]\n\nProduct SDK: KnowledgeService.Get\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).Get(context.Background(), *modelID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	name := fs.String("name", "", "existing immutable knowledge base and Catalog database name (rename unsupported)")
	description := fs.String("description", "", "Knowledge base description")
	tablesFile := fs.String("tables-file", "", "JSON file containing semantic table definitions")
	filesFile := fs.String("files-file", "", "JSON file containing semantic file configuration")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model update [options]\n\nProduct SDK: KnowledgeService.Update\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts, err := semanticModelCLIOptions(fs, *description, *tablesFile, *filesFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).Update(context.Background(), *modelID, *name, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func semanticModelCLIOptions(fs *flag.FlagSet, description, tablesFile, filesFile string) ([]sdk.SemanticModelOption, error) {
	var opts []sdk.SemanticModelOption
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "description" {
			opts = append(opts, sdk.WithSemanticModelDescription(description))
		}
	})
	if tablesFile != "" {
		tables, err := loadJSONFile[json.RawMessage](tablesFile)
		if err != nil {
			return nil, fmt.Errorf("parse --tables-file: %w", err)
		}
		opts = append(opts, sdk.WithSemanticModelTables(tables))
	}
	if filesFile != "" {
		files, err := loadJSONFile[json.RawMessage](filesFile)
		if err != nil {
			return nil, fmt.Errorf("parse --files-file: %w", err)
		}
		opts = append(opts, sdk.WithSemanticModelFiles(files))
	}
	return opts, nil
}

func runKnowledgeServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model delete [options]\n\nProduct SDK: KnowledgeService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).Delete(context.Background(), *modelID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceCreateWithSources(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model create-with-sources", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "knowledge base name and immutable Catalog database name")
	description := fs.String("description", "", "Knowledge base description")
	filesFile := fs.String("files-file", "", "JSON file containing semantic file configuration")
	imageIndexEnabled := fs.Bool("image-index-enabled", false, "Enable image vector index")
	targetCatalogID := fs.Int64("target-catalog-id", 0, "Target Catalog ID (omit to let MOI choose)")
	sourceSelectionsFile := fs.String("source-selections-file", "", "JSON file containing Catalog source selections")
	bodyJSON := fs.String("json", "", "JSON array of source objects; use --help or schema for source types and required fields")
	bodyFile := fs.String("json-file", "", "File containing the same source array as --json")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model create-with-sources [options]\n\nProduct SDK: KnowledgeService.CreateWithSources\n\nOptions:\n")
		fs.PrintDefaults()
		printSemanticModelSourceGuidance(fs.Output())
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
	var sources []sdk.SemanticModelSourceInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &sources); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	var opts []sdk.SemanticModelWithSourcesOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "target-catalog-id":
			opts = append(opts, sdk.WithSemanticModelWithSourcesTargetCatalogID(*targetCatalogID))
		case "description":
			opts = append(opts, sdk.WithSemanticModelWithSourcesDescription(*description))
		case "image-index-enabled":
			opts = append(opts, sdk.WithSemanticModelWithSourcesImageIndexEnabled(*imageIndexEnabled))
		}
	})
	if *filesFile != "" {
		files, err := loadJSONFile[json.RawMessage](*filesFile)
		if err != nil {
			return fmt.Errorf("parse --files-file: %w", err)
		}
		opts = append(opts, sdk.WithSemanticModelWithSourcesFiles(files))
	}
	if *sourceSelectionsFile != "" {
		selections, err := loadJSONFile[[]sdk.SemanticModelSourceSelectionInput](*sourceSelectionsFile)
		if err != nil {
			return fmt.Errorf("parse --source-selections-file: %w", err)
		}
		opts = append(opts, sdk.WithSemanticModelWithSourcesSourceSelections(selections...))
	}
	result, err := ctx.Client.Knowledge(wid).CreateWithSources(context.Background(), *name, sources, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceCreateEmpty(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model create-empty", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "knowledge base name and immutable Catalog database name")
	description := fs.String("description", "", "knowledge base description")
	imageIndexEnabled := fs.Bool("image-index-enabled", false, "Enable image vector index")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model create-empty [options]\n\nProduct SDK: KnowledgeService.CreateEmpty\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := semanticModelEmptyCLIOptions(fs, *description, *imageIndexEnabled)
	result, err := ctx.Client.Knowledge(wid).CreateEmpty(context.Background(), *name, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceUploadLocalFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model upload-local-file", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	filename := fs.String("filename", "", "filename")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model upload-local-file [options]\n\nProduct SDK: KnowledgeService.UploadLocalFile\n\nOptions:\n")
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
	result, err := ctx.Client.Knowledge(wid).UploadLocalFile(context.Background(), source.Name, source.Reader)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceUploadModelLocalFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model upload-model-local-file", flag.ExitOnError)
	upload := addUploadFlags(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	filename := fs.String("filename", "", "filename")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model upload-model-local-file [options]\n\nProduct SDK: KnowledgeService.UploadModelLocalFile\n\nOptions:\n")
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
	result, err := ctx.Client.Knowledge(wid).UploadModelLocalFile(context.Background(), *modelID, source.Name, source.Reader)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServicePreviewSourceSelectionCounts(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model preview-source-selection-counts", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model preview-source-selection-counts [options]\n\nProduct SDK: KnowledgeService.PreviewSourceSelectionCounts\n\nOptions:\n")
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
	var selections []sdk.SemanticModelSourceSelectionInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &selections); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).PreviewSourceSelectionCounts(context.Background(), selections)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServicePreviewModelSourceSelectionCounts(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model preview-model-source-selection-counts", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model preview-model-source-selection-counts [options]\n\nProduct SDK: KnowledgeService.PreviewModelSourceSelectionCounts\n\nOptions:\n")
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
	var selections []sdk.SemanticModelSourceSelectionInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &selections); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).PreviewModelSourceSelectionCounts(context.Background(), *modelID, selections)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceListSources(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model list-sources", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	includeDisplayStatus := fs.Bool("include-display-status", false, "Include source display status")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model list-sources [options]\n\nProduct SDK: KnowledgeService.ListSources\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).ListSources(context.Background(), *modelID, semanticModelSourceListCLIOptions(fs, *includeDisplayStatus)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceListReferencingAgents(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model list-referencing-agents", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "Knowledge base ID")
	limit := fs.Int("limit", 0, "Page size (omit to use the product default)")
	cursor := fs.String("cursor", "", "Cursor from the previous page")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model list-referencing-agents [options]\n\nProduct SDK: KnowledgeService.ListReferencingAgents\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).ListReferencingAgents(context.Background(), *modelID, *limit, *cursor)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceListSourcesPage(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model list-sources-page", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	page := fs.Int("page", 0, "page")
	pageSize := fs.Int("page-size", 0, "pageSize")
	includeDisplayStatus := fs.Bool("include-display-status", false, "Include source display status")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model list-sources-page [options]\n\nProduct SDK: KnowledgeService.ListSourcesPage\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).ListSourcesPage(context.Background(), *modelID, *page, *pageSize, semanticModelSourceListCLIOptions(fs, *includeDisplayStatus)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func semanticModelSourceListCLIOptions(fs *flag.FlagSet, includeDisplayStatus bool) []sdk.QueryOption {
	var options []sdk.QueryOption
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "include-display-status" {
			options = append(options, sdk.WithQueryParam("include_display_status", strconv.FormatBool(includeDisplayStatus)))
		}
	})
	return options
}

func runKnowledgeServiceCheckSourceExistence(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model check-source-existence", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fileIDs := fs.String("file-ids", "", "fileIDs (comma-separated)")
	tableIDs := fs.String("table-ids", "", "tableIDs (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model check-source-existence [options]\n\nProduct SDK: KnowledgeService.CheckSourceExistence\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	parsedTableIDs, err := parseInt64CSV(*tableIDs)
	if err != nil {
		return fmt.Errorf("--table-ids: %w", err)
	}
	result, err := ctx.Client.Knowledge(wid).CheckSourceExistence(context.Background(), *modelID, splitCSVFlag(fs, "file-ids", *fileIDs), parsedTableIDs)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceAppendSources(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model append-sources", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceSelectionsFile := fs.String("source-selections-file", "", "JSON file containing Catalog source selections")
	bodyJSON := fs.String("json", "", "JSON array of source objects; use --help or schema for source types and required fields")
	bodyFile := fs.String("json-file", "", "File containing the same source array as --json")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model append-sources [options]\n\nProduct SDK: KnowledgeService.AppendSources\n\nOptions:\n")
		fs.PrintDefaults()
		printSemanticModelSourceGuidance(fs.Output())
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
	var sources []sdk.SemanticModelSourceInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &sources); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	var opts []sdk.SemanticModelAppendSourcesOption
	if *sourceSelectionsFile != "" {
		selections, err := loadJSONFile[[]sdk.SemanticModelSourceSelectionInput](*sourceSelectionsFile)
		if err != nil {
			return fmt.Errorf("parse --source-selections-file: %w", err)
		}
		opts = append(opts, sdk.WithSemanticModelAppendSourceSelections(selections...))
	}
	result, err := ctx.Client.Knowledge(wid).AppendSources(context.Background(), *modelID, sources, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceBackfillLegacySources(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model backfill-legacy-sources", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model backfill-legacy-sources [options]\n\nProduct SDK: KnowledgeService.BackfillLegacySources\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).BackfillLegacySources(context.Background(), *modelID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceDeleteSource(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model delete-source", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model delete-source [options]\n\nProduct SDK: KnowledgeService.DeleteSource\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).DeleteSource(context.Background(), *modelID, *sourceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceGetSourceDocument(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model get-source-document", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model get-source-document [options]\n\nProduct SDK: KnowledgeService.GetSourceDocument\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).GetSourceDocument(context.Background(), *modelID, *sourceID, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceUpdateSourceGovernance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model update-source-governance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	tagsJSON := fs.String("tags", "", "JSON string array of replacement tags; [] clears tags")
	expiresAt := fs.Int64("expires-at", 0, "Expiry as Unix seconds; omission preserves the current expiry")
	clearExpiresAt := fs.Bool("clear-expires-at", false, "Clear expiry (exclusive with --expires-at)")
	enabled := fs.Bool("enabled", false, "Enable or disable the source; omission preserves its state")
	forceEnabledAfterExpiry := fs.Bool("force-enabled-after-expiry", false, "Allow or disallow use after expiry; omission preserves the current setting")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model update-source-governance [options]\n\nProduct SDK: KnowledgeService.UpdateSourceGovernance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	var options []sdk.SemanticModelSourceGovernanceOption
	var inputErr error
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "tags":
			var tags []string
			if err := json.Unmarshal([]byte(*tagsJSON), &tags); err != nil {
				inputErr = fmt.Errorf("parse --tags: %w", err)
				return
			}
			if tags == nil {
				inputErr = fmt.Errorf("--tags requires a JSON string array; use [] to clear tags")
				return
			}
			options = append(options, sdk.WithSemanticModelSourceTags(tags...))
		case "expires-at":
			if *clearExpiresAt {
				inputErr = fmt.Errorf("--expires-at and --clear-expires-at are mutually exclusive")
				return
			}
			options = append(options, sdk.WithSemanticModelSourceExpiresAt(*expiresAt))
		case "enabled":
			options = append(options, sdk.WithSemanticModelSourceEnabled(*enabled))
		case "force-enabled-after-expiry":
			options = append(options, sdk.WithSemanticModelSourceForceEnabledAfterExpiry(*forceEnabledAfterExpiry))
		}
	})
	if inputErr != nil {
		return inputErr
	}
	if *clearExpiresAt {
		options = append(options, sdk.WithSemanticModelSourceExpiresAtCleared())
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).UpdateSourceGovernance(context.Background(), *modelID, *sourceID, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceUpdateSegment(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model update-segment", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	segmentID := fs.String("segment-id", "", "segmentID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model update-segment [options]\n\nProduct SDK: KnowledgeService.UpdateSegment\n\nOptions:\n")
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
	var input sdk.SemanticModelSegmentUpdateInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &input); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).UpdateSegment(context.Background(), *modelID, *sourceID, *segmentID, input)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceDeleteSegment(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model delete-segment", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	segmentID := fs.String("segment-id", "", "segmentID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model delete-segment [options]\n\nProduct SDK: KnowledgeService.DeleteSegment\n\nOptions:\n")
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
	var base sdk.SemanticModelSegmentBase
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &base); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).DeleteSegment(context.Background(), *modelID, *sourceID, *segmentID, base)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceCreateSegment(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model create-segment", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model create-segment [options]\n\nProduct SDK: KnowledgeService.CreateSegment\n\nOptions:\n")
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
	var input sdk.SemanticModelSegmentCreateInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &input); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).CreateSegment(context.Background(), *modelID, *sourceID, input)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceUpdateSegmentEnabled(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model update-segment-enabled", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	segmentID := fs.String("segment-id", "", "segmentID")
	enabled := fs.Bool("enabled", false, "enabled")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model update-segment-enabled [options]\n\nProduct SDK: KnowledgeService.UpdateSegmentEnabled\n\nOptions:\n")
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
	var base sdk.SemanticModelSegmentBase
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &base); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).UpdateSegmentEnabled(context.Background(), *modelID, *sourceID, *segmentID, base, *enabled)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceReembedSegments(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model reembed-segments", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	sourceID := fs.String("source-id", "", "sourceID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model reembed-segments [options]\n\nProduct SDK: KnowledgeService.ReembedSegments\n\nOptions:\n")
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
	var base sdk.SemanticModelSegmentBase
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &base); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).ReembedSegments(context.Background(), *modelID, *sourceID, base)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceListSourceJobs(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model list-source-jobs", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model list-source-jobs [options]\n\nProduct SDK: KnowledgeService.ListSourceJobs\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).ListSourceJobs(context.Background(), *modelID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceReconcileSourceJobs(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model reconcile-source-jobs", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	retrySourceRowID := fs.String("retry-source-row-id", "", "Retry this failed knowledge source row ID (omit to reconcile without retry)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model reconcile-source-jobs [options]\n\nProduct SDK: KnowledgeService.ReconcileSourceJobs\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).ReconcileSourceJobs(context.Background(), *modelID, sdk.WithRetrySourceRowID(*retrySourceRowID))
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceListEntries(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model list-entries", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model list-entries [options]\n\nProduct SDK: KnowledgeService.ListEntries\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).ListEntries(context.Background(), *modelID, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceCreateEntry(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model create-entry", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model create-entry [options]\n\nProduct SDK: KnowledgeService.CreateEntry\n\nOptions:\n")
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
	var entry sdk.SemanticEntryInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &entry); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).CreateEntry(context.Background(), *modelID, entry)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceUpdateEntry(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model update-entry", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	entryID := fs.String("entry-id", "", "entryID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model update-entry [options]\n\nProduct SDK: KnowledgeService.UpdateEntry\n\nOptions:\n")
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
	var entry sdk.SemanticEntryInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &entry); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).UpdateEntry(context.Background(), *modelID, *entryID, entry)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceDeleteEntry(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model delete-entry", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	entryID := fs.String("entry-id", "", "entryID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model delete-entry [options]\n\nProduct SDK: KnowledgeService.DeleteEntry\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).DeleteEntry(context.Background(), *modelID, *entryID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceImport(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model import", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model import [options]\n\nProduct SDK: KnowledgeService.Import\n\nOptions:\n")
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
	var entries []sdk.SemanticEntryInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &entries); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Knowledge(wid).Import(context.Background(), *modelID, entries)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceExport(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model export", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model export [options]\n\nProduct SDK: KnowledgeService.Export\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).Export(context.Background(), *modelID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceValidate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model validate", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	modelID := fs.String("model-id", "", "modelID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model validate [options]\n\nProduct SDK: KnowledgeService.Validate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).Validate(context.Background(), *modelID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceSearchDrawings(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model search-drawings", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	knowledgeBaseID := fs.Int64("knowledge-base-id", 0, "knowledge_base_id")
	query := fs.String("query", "", "text query")
	queryImage := fs.String("query-image", "", "local query image path")
	queryImageFilename := fs.String("query-image-filename", "", "filename sent with the query image (default: local path base)")
	topK := fs.Int("top-k", 0, "top_k")
	bodyJSON := fs.String("json", "", "JSON DrawingSearchRequest (overrides individual flags when set)")
	bodyFile := fs.String("json-file", "", "JSON file for DrawingSearchRequest")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model search-drawings [options]\n\nProduct SDK: KnowledgeService.SearchDrawings\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	req := sdk.DrawingSearchRequest{
		KnowledgeBaseID:    *knowledgeBaseID,
		Query:              *query,
		QueryImagePath:     *queryImage,
		QueryImageFilename: *queryImageFilename,
		TopK:               *topK,
	}
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "top-k" {
			req.TopKPresent = true
		}
	})
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &req); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(rawBody, &fields); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
		if _, present := fields["top_k"]; present {
			req.TopKPresent = true
		}
	}
	result, err := ctx.Client.Knowledge(wid).SearchDrawings(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceGetDrawingProvenance(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model get-drawing-provenance", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	knowledgeBaseID := fs.Int64("knowledge-base-id", 0, "knowledge_base_id")
	sourceFileID := fs.String("source-file-id", "", "catalog source file id")
	pageNumber := fs.Int("page-number", 0, "optional page number; 0 returns every page")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model get-drawing-provenance [options]\n\nProduct SDK: KnowledgeService.GetDrawingProvenance\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).GetDrawingProvenance(context.Background(), *knowledgeBaseID, *sourceFileID, *pageNumber)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServicePreviewSourceFile(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model preview-source-file", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	modelID := fs.String("model-id", "", "model-id")
	fileID := fs.String("file-id", "", "file-id")
	output := fs.String("output", "", "output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model preview-source-file [options]\n\nProduct SDK: KnowledgeService.PreviewSourceFile\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).PreviewSourceFile(context.Background(), *modelID, *fileID)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}

func runKnowledgeServicePreviewArtifact(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model preview-artifact", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	modelID := fs.String("model-id", "", "model-id")
	fileID := fs.String("file-id", "", "file-id")
	output := fs.String("output", "", "output")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model preview-artifact [options]\n\nProduct SDK: KnowledgeService.PreviewArtifact\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Knowledge(wid).PreviewArtifact(context.Background(), *modelID, *fileID)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}

func runKnowledgeServiceImportModel(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model import-model", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model import-model [options]\n\nProduct SDK: KnowledgeService.ImportModel\n\nOptions:\n")
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
	var request sdk.SemanticModelImportRequest
	if err := protojson.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Knowledge(wid).ImportModel(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runKnowledgeServiceImportModels(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("semantic-model import-models", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model import-models [options]\n\nProduct SDK: KnowledgeService.ImportModels\n\nOptions:\n")
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
	var rawRequests []json.RawMessage
	if err := json.Unmarshal(rawBody, &rawRequests); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	request := make([]*sdk.SemanticModelImportRequest, 0, len(rawRequests))
	for _, rawRequest := range rawRequests {
		item := &sdk.SemanticModelImportRequest{}
		if err := protojson.Unmarshal(rawRequest, item); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
		request = append(request, item)
	}
	result, err := ctx.Client.Knowledge(wid).ImportModels(context.Background(), request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
