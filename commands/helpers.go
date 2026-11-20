package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// queryParameters exposes the Product SDK's existing QueryOption contract.
// Names, values and repeated keys pass through to the SDK and product owner.
type queryParameters []sdk.QueryOption

func (parameters *queryParameters) String() string { return "" }

func (parameters *queryParameters) Set(value string) error {
	name, argument, ok := strings.Cut(value, "=")
	if !ok || name == "" {
		return fmt.Errorf("--query-param requires name=value")
	}
	*parameters = append(*parameters, sdk.WithQueryValues(name, argument))
	return nil
}

type catalogListParameters struct {
	includeStats bool
	pageSize     int
	pageToken    string
}

func (parameters catalogListParameters) options(fs *flag.FlagSet) []sdk.CatalogListOption {
	var options []sdk.CatalogListOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "include-stats":
			options = append(options, sdk.WithCatalogIncludeStats(parameters.includeStats))
		case "page-size":
			options = append(options, sdk.WithCatalogPageSize(parameters.pageSize))
		case "page-token":
			options = append(options, sdk.WithCatalogPageToken(parameters.pageToken))
		}
	})
	return options
}

type catalogInfoParameters struct {
	description string
}

func addCatalogInfoFlags(fs *flag.FlagSet) *catalogInfoParameters {
	parameters := &catalogInfoParameters{}
	fs.StringVar(&parameters.description, "description", "", "Resource description")
	return parameters
}

func (parameters catalogInfoParameters) options(fs *flag.FlagSet) []sdk.CatalogInfoOption {
	var options []sdk.CatalogInfoOption
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "description" {
			options = append(options, sdk.WithCatalogDescription(parameters.description))
		}
	})
	return options
}

type catalogTableLookupParameters struct {
	name         string
	databaseID   int
	databaseName string
	includeStats bool
}

func addCatalogTableLookupFlags(fs *flag.FlagSet) *catalogTableLookupParameters {
	parameters := &catalogTableLookupParameters{}
	fs.StringVar(&parameters.name, "name", "", "Table name")
	fs.IntVar(&parameters.databaseID, "database-id", 0, "Database ID")
	fs.StringVar(&parameters.databaseName, "database-name", "", "System database name")
	fs.BoolVar(&parameters.includeStats, "include-stats", true, "Include table statistics")
	return parameters
}

func (parameters catalogTableLookupParameters) options(fs *flag.FlagSet) []sdk.TableLookupOption {
	var options []sdk.TableLookupOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "name":
			options = append(options, sdk.WithTableLookupName(parameters.name))
		case "database-id":
			options = append(options, sdk.WithTableLookupDatabaseID(parameters.databaseID))
		case "database-name":
			options = append(options, sdk.WithTableLookupDatabaseName(parameters.databaseName))
		case "include-stats":
			options = append(options, sdk.WithTableLookupIncludeStats(parameters.includeStats))
		}
	})
	return options
}

type catalogTableDataParameters struct {
	page         int
	pageSize     int
	name         string
	databaseID   int
	databaseName string
	previewLines int
	includeTotal bool
}

func addCatalogTableDataFlags(fs *flag.FlagSet) *catalogTableDataParameters {
	parameters := &catalogTableDataParameters{}
	fs.IntVar(&parameters.page, "page", 0, "Page number")
	fs.IntVar(&parameters.pageSize, "page-size", 0, "Page size")
	fs.StringVar(&parameters.name, "name", "", "Table name")
	fs.IntVar(&parameters.databaseID, "database-id", 0, "Database ID")
	fs.StringVar(&parameters.databaseName, "database-name", "", "System database name")
	fs.IntVar(&parameters.previewLines, "preview-lines", 0, "Preview row limit")
	fs.BoolVar(&parameters.includeTotal, "include-total", true, "Include total row count")
	return parameters
}

func (parameters catalogTableDataParameters) options(fs *flag.FlagSet) []sdk.TableDataOption {
	var options []sdk.TableDataOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "page":
			options = append(options, sdk.WithTableDataPage(parameters.page))
		case "page-size":
			options = append(options, sdk.WithTableDataPageSize(parameters.pageSize))
		case "name":
			options = append(options, sdk.WithTableDataName(parameters.name))
		case "database-id":
			options = append(options, sdk.WithTableDataDatabaseID(parameters.databaseID))
		case "database-name":
			options = append(options, sdk.WithTableDataDatabaseName(parameters.databaseName))
		case "preview-lines":
			options = append(options, sdk.WithTablePreviewLines(parameters.previewLines))
		case "include-total":
			options = append(options, sdk.WithTableDataIncludeTotal(parameters.includeTotal))
		}
	})
	return options
}

func loadJSONBytes(inline, file string) ([]byte, error) {
	if file == "-" {
		return io.ReadAll(os.Stdin)
	}
	if file != "" {
		return os.ReadFile(file)
	}
	if inline != "" {
		return []byte(inline), nil
	}
	return nil, nil
}

func unmarshalJSONUseNumber(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

// unmarshalProductJSON preserves protobuf JSON values, including explicit null.
// Unknown input fields remain ignored, as with the ordinary CLI JSON loader.
func unmarshalProductJSON(data []byte, target proto.Message) error {
	return (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(data, target)
}

func loadJSONFile[T any](path string) (T, error) {
	var value T
	data, err := loadJSONBytes("", path)
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(data, &value)
	return value, err
}

func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseIntCSV parses a comma-separated integer flag and rejects malformed entries
// instead of silently dropping them.
func parseIntCSV(flagName, s string) ([]int, error) {
	parts := splitCSV(s)
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("--%s: %q is not an integer", flagName, p)
		}
		out = append(out, n)
	}
	return out, nil
}

func parseInt64CSV(s string) ([]int64, error) {
	parts := splitCSV(s)
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return nil, errors.New("CLI_INVALID_INT64_CSV")
		}
		out = append(out, n)
	}
	return out, nil
}

func writeStream(r io.Reader, outPath string) error {
	var w io.Writer = os.Stdout
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	_, err := io.Copy(w, r)
	return err
}

func providerBackendService(ctx *Context, workspaceID, kind string) *sdk.ProviderBackendService {
	models := ctx.Client.Models(workspaceID)
	switch strings.ToLower(kind) {
	case "embedding":
		return models.EmbeddingBackends()
	case "parser":
		return models.ParserBackends()
	default:
		return models.LLMBackends()
	}
}

// silence unused in some build tags / empty domains
var (
	_ = bytes.NewReader
	_ = fmt.Sprintf
)

func setSDKCommandUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli %s [options]\n\nOptions:\n", fs.Name())
		fs.PrintDefaults()
	}
}

// runWorkspaceRequest resolves the command's scope after flag parsing and keeps
// SDK errors intact. Nil results represent writes with no printable payload.
func runWorkspaceRequest(ctx *Context, fs *flag.FlagSet, args []string, workspaceID *string, invoke func(string) (any, error)) error {
	setSDKCommandUsage(fs)
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := invoke(wid)
	if err != nil || result == nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
