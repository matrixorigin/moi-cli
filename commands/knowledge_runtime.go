package commands

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

// Put usable local locators and inspection commands before potentially large
// evidence text. A model receiving a bounded tool view can then read the saved
// result without inferring its schema from a clipped preview.
type knowledgeResultSummary struct {
	Path            string              `json:"path"`
	TableCount      any                 `json:"table_count,omitempty"`
	Available       *bool               `json:"available,omitempty"`
	ResultPreview   map[string]any      `json:"result_preview,omitempty"`
	ReadingPath     string              `json:"reading_path,omitempty"`
	ReadCommands    map[string][]string `json:"read_commands,omitempty"`
	ReceiptIndices  []int               `json:"receipt_indices,omitempty"`
	NextAction      string              `json:"next_action,omitempty"`
	Guidance        string              `json:"guidance,omitempty"`
	LocalAnalysis   map[string]any      `json:"local_analysis,omitempty"`
	NextRead        [][]string          `json:"next_read,omitempty"`
	Accepted        *bool               `json:"accepted,omitempty"`
	Coverage        string              `json:"coverage,omitempty"`
	ReadingOrder    string              `json:"reading_order,omitempty"`
	SourcePage      any                 `json:"source_page,omitempty"`
	Rerank          any                 `json:"rerank,omitempty"`
	CallID          string              `json:"call_id"`
	Bytes           int                 `json:"bytes"`
	RowCount        any                 `json:"row_count,omitempty"`
	ChunkCount      any                 `json:"chunk_count,omitempty"`
	Routes          any                 `json:"routes,omitempty"`
	EvidencePreview string              `json:"evidence_preview,omitempty"`
}

type knowledgeSourceSelection struct {
	pathIndex int
	indices   []string
	quote     string
}

func addKnowledgeQuoteFlag(fs *flag.FlagSet, previous *[]string, selections *[]knowledgeSourceSelection) {
	fs.Func("quote", "Exact inspected text identifying one canonical passage; pair each --from with one --quote", func(value string) error {
		*selections = append(*selections, knowledgeSourceSelection{pathIndex: len(*previous) - 1, quote: value})
		return nil
	})
}

// runKnowledgeRuntime keeps JSON evidence local. --from resolves opaque IDs
// programmatically from a previous result, so the model never retypes them.
func runKnowledgeRuntime(ctx *Context, operation string, args []string) error {
	fs := flag.NewFlagSet("semantic-model "+operation, flag.ContinueOnError)
	var body proto.Message
	var localRead bool
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli semantic-model %s [options]\n\nTask-scoped knowledge operation.\n\nOptions:\n", operation)
		visible := flag.NewFlagSet(fs.Name(), flag.ContinueOnError)
		visible.SetOutput(os.Stderr)
		fs.VisitAll(func(value *flag.Flag) {
			if KnowledgeRuntimeFlagSupported(operation, value.Name) {
				visible.Var(value.Value, value.Name, value.Usage)
				visible.Lookup(value.Name).DefValue = value.DefValue
			}
		})
		visible.PrintDefaults()
		if operation == "search" {
			fmt.Fprintln(os.Stderr, "\nKeyword flags default to max_hits=8, before=2, after=1. JSON input follows SDK field defaults: include max_hits for keyword searches and explicit bounds when needed. Tables retain atomic-block retrieval.")
		}
		fmt.Fprintln(os.Stderr, "\nTyped JSON fields for --json / --json-file:")
		fields := body.ProtoReflect().Descriptor().Fields()
		for i := range fields.Len() {
			field := fields.Get(i)
			if !knowledgeRuntimeJSONFieldSupported(operation, string(field.Name())) {
				continue
			}
			kind := field.Kind().String()
			if field.IsList() {
				kind = "array<" + kind + ">"
			}
			fmt.Fprintf(os.Stderr, "  %s: %s\n", field.Name(), kind)
			if field.Name() == "source_page" {
				selectors := field.Message().Fields()
				for j := range selectors.Len() {
					selector := selectors.Get(j)
					fmt.Fprintf(os.Stderr, "    source_page.%s: %s\n", selector.Name(), selector.Kind())
				}
			}
		}
	}
	inline := fs.String("json", "", "Typed request JSON")
	input := fs.String("json-file", "", "Typed request JSON file")
	output := fs.String("output", "", "Write the complete result to a new local file")
	var previous []string
	fs.Func("from", "Saved inventory, evidence result or previous source page (repeatable for citations)", func(path string) error {
		previous = append(previous, path)
		return nil
	})
	var selections []knowledgeSourceSelection
	var quotedSelections []knowledgeSourceSelection
	fs.Func("indices", "Evidence indices for citation or batch reads; citations may pair each --from with its own selection", func(value string) error {
		selections = append(selections, knowledgeSourceSelection{pathIndex: len(previous) - 1, indices: splitCSV(value)})
		return nil
	})
	index := fs.Int("index", -1, "Zero-based evidence index for a read")
	requestID := fs.String("request-id", "", "Stable ID for replay of this exact request")
	var sourceName string
	var ws *sdk.KnowledgeService
	var call func(context.Context) (*sdk.KnowledgeRuntimeResult, error)
	switch operation {
	case "find-files":
		v := &sdk.KnowledgeFindFilesRequest{}
		fs.StringVar(&v.Query, "query", "", "File name or inventory query")
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.FindFiles(c, v) }
	case "search":
		v := &sdk.KnowledgeSearchRequest{MaxHits: 8, Before: 2, After: 1}
		fs.BoolVar(&v.FulltextOnly, "fulltext-only", false, "Use full-text index matching without embedding or vector candidates")
		fs.Func("keywords", "Repeatable comma-separated retrieval keywords", knowledgeKeywordsFlag(&v.Keywords))
		fs.Func("max-hits", "Candidate limit per query, retrieval route and index group (default 8); merged results and neighboring passages may exceed it", func(value string) error {
			n, err := strconv.ParseInt(value, 10, 32)
			if err != nil || n < 0 {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: max-hits out of range")
			}
			v.MaxHits = int32(n)
			return nil
		})
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.Search(c, v) }
	case "list-passages":
		v := &sdk.KnowledgeSearchRequest{SourcePage: new(sdk.KnowledgeSourcePage)}
		fs.StringVar(&v.SourcePage.Cursor, "cursor", "", "Exact continuation cursor from this source's previous page")
		fs.Func("page-size", "Enabled text passages per page (default 20, maximum 100)", knowledgePageSizeFlag(v.SourcePage))
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.Search(c, v) }
	case "read-source":
		v := &sdk.KnowledgeSearchRequest{SourcePage: &sdk.KnowledgeSourcePage{PageSize: 100}}
		fs.StringVar(&sourceName, "source-name", "", "Resolve exactly one selected source by filename query")
		fs.Func("page-size", "Enabled text passages per page (default 100, maximum 100)", knowledgePageSizeFlag(v.SourcePage))
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return readKnowledgeSource(c, ws, v) }
	case "read-document":
		v := &sdk.KnowledgeReadDocumentRequest{}
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.ReadDocument(c, v) }
	case "search-document":
		v := &sdk.KnowledgeSearchDocumentRequest{}
		fs.StringVar(&v.Query, "query", "", "Text to find within parsed Markdown")
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.SearchDocument(c, v) }
	case "search-images":
		v := &sdk.KnowledgeSearchImagesRequest{}
		fs.StringVar(&v.QueryText, "query", "", "Description of visual evidence")
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.SearchImages(c, v) }
	case "read-evidence":
		addKnowledgeQuoteFlag(fs, &previous, &quotedSelections)
		v := &sdk.KnowledgeReadEvidenceRequest{}
		fs.BoolVar(&localRead, "local-read", false, "Register a complete locally inspected passage; verify its digest against this task's evidence")
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.ReadEvidence(c, v) }
	case "describe-tables":
		v := &sdk.KnowledgeDescribeTablesRequest{}
		fs.Func("keywords", "Repeatable comma-separated schema keywords", knowledgeKeywordsFlag(&v.Keywords))
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.DescribeTables(c, v) }
	case "query":
		v := &sdk.KnowledgeQueryRequest{}
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.Query(c, v) }
	case "cite":
		addKnowledgeQuoteFlag(fs, &previous, &quotedSelections)
		v := &sdk.KnowledgeCiteRequest{CompleteReadRequirements: true}
		body = v
		call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) { return ws.Cite(c, v) }
	default:
		return fmt.Errorf("CLI_KNOWLEDGE_OPERATION_INVALID")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		if operation == "search" || operation == "describe-tables" {
			return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: unexpected positional queries; repeat --keywords for each query or use a comma-separated list")
		}
		return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: unexpected positional arguments; use --help for supported flags")
	}
	if *inline != "" && *input != "" {
		return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID")
	}
	if len(selections) > 0 && len(previous) == 0 {
		return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: --indices requires --from")
	}
	var flagErr error
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "source-name" && strings.TrimSpace(value.Value.String()) == "" && flagErr == nil {
			flagErr = fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: source-name must not be empty")
		}
		if !KnowledgeRuntimeFlagSupported(operation, value.Name) && flagErr == nil {
			flagErr = fmt.Errorf("CLI_KNOWLEDGE_FLAG_NOT_APPLICABLE: --%s for %s", value.Name, operation)
		}
	})
	if flagErr != nil {
		return flagErr
	}
	if len(quotedSelections) > 0 {
		invalid := len(previous) != len(quotedSelections) || len(selections) > 0 || *index != -1
		if _, read := body.(*sdk.KnowledgeReadEvidenceRequest); read && len(previous) != 1 {
			invalid = true
		}
		for i, selection := range quotedSelections {
			invalid = invalid || selection.pathIndex != i || strings.TrimSpace(selection.quote) == ""
		}
		if invalid {
			return &sdk.Error{Code: "CLI_KNOWLEDGE_INPUT_INVALID", Message: "Pair each --from with one non-empty --quote. Quotes cannot be mixed with index/indices; each read uses one result file."}
		}
	}
	wsID, err := resolveWorkspaceID(ctx, "")
	if err != nil {
		return err
	}
	ws = ctx.Client.Knowledge(wsID)
	raw, err := loadJSONBytes(*inline, *input)
	if err != nil {
		return err
	}
	if len(raw) > 0 {
		if search, ok := body.(*sdk.KnowledgeSearchRequest); ok {
			// Constructor defaults belong to flag-only search. JSON parameters
			// must not acquire a conflicting keyword window in source-page mode.
			search.Before, search.After = 0, 0
			maxHitsSet := false
			fs.Visit(func(value *flag.Flag) {
				if value.Name == "max-hits" {
					maxHitsSet = true
				}
			})
			if !maxHitsSet {
				search.MaxHits = 0
			}
		}
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(body); err != nil {
			return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: %w", err)
		}
		var extra any
		if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
			return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID")
		}
	}
	field := body.ProtoReflect().Descriptor().Fields().ByName("request_id")
	existing := body.ProtoReflect().Get(field).String()
	if *requestID != "" && existing != "" && *requestID != existing {
		return fmt.Errorf("CLI_KNOWLEDGE_REQUEST_ID_CONFLICT")
	}
	if existing == "" {
		id := *requestID
		if id == "" {
			id = "kb_" + rand.Text()
		}
		body.ProtoReflect().Set(field, protoreflect.ValueOfString(id))
	}
	if sourceName != "" {
		request := body.(*sdk.KnowledgeSearchRequest)
		page := request.GetSourcePage()
		if len(previous) != 0 || page == nil || page.FileId != "" || page.SemanticModelId != 0 || page.Cursor != "" {
			return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: source-name cannot be combined with another source selector")
		}
		if err := selectKnowledgeNamedSource(context.Background(), ws, request, sourceName); err != nil {
			return err
		}
	}
	if read, ok := body.(*sdk.KnowledgeReadEvidenceRequest); ok && read.LocalReads != nil && (len(previous) > 0 || len(selections) > 0 || localRead) {
		return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: local_reads cannot be combined with saved-result selectors")
	}
	if len(quotedSelections) > 0 {
		for i, selection := range quotedSelections {
			if err := knowledgeRequestFromResult(body, []string{previous[i]}, -1, nil, localRead, selection.quote); err != nil {
				return err
			}
		}
	} else if len(previous) > 0 {
		if _, cite := body.(*sdk.KnowledgeCiteRequest); cite && len(selections) == 0 {
			if err := knowledgeRequestFromResult(body, previous, *index, nil, localRead, ""); err != nil {
				return err
			}
		} else if read, ok := body.(*sdk.KnowledgeReadEvidenceRequest); ok && len(selections) > 0 {
			if len(previous) != 1 || len(selections) != 1 || len(selections[0].indices) == 0 || *index != -1 {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: batch reads require one --from, one non-empty --indices and no --index")
			}
			var reads []knowledgeRead
			for position, text := range selections[0].indices {
				n, err := strconv.Atoi(text)
				if err != nil {
					return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: evidence index must be an integer: %w", err)
				}
				request := proto.Clone(read).(*sdk.KnowledgeReadEvidenceRequest)
				if err := knowledgeRequestFromResult(request, previous, n, nil, localRead, ""); err != nil {
					return err
				}
				mode := "read-evidence-batch"
				if localRead {
					mode = "local-read-batch"
				}
				request.RequestId = knowledgeSourceRequestID(mode, read.RequestId, strconv.Itoa(position))
				reads = append(reads, knowledgeRead{index: n, request: request})
			}
			call = func(c context.Context) (*sdk.KnowledgeRuntimeResult, error) {
				return readKnowledgeBatch(c, ws, read.RequestId, reads, localRead)
			}
		} else if _, cite := body.(*sdk.KnowledgeCiteRequest); cite && len(selections) != 1 {
			if len(selections) != len(previous) {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: provide --indices after each --from, or one flattened --indices")
			}
			for i, selection := range selections {
				if selection.pathIndex != i || len(selection.indices) == 0 {
					return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: each repeated --indices must follow its own --from")
				}
				if err := knowledgeRequestFromResult(body, []string{previous[i]}, *index, selection.indices, localRead, ""); err != nil {
					return err
				}
			}
		} else {
			var selected []string
			if len(selections) == 1 {
				selected = selections[0].indices
			}
			if _, cite := body.(*sdk.KnowledgeCiteRequest); cite && len(selected) == 0 {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: cite --from requires non-empty --indices")
			}
			if err := knowledgeRequestFromResult(body, previous, *index, selected, localRead, ""); err != nil {
				return err
			}
		}
	} else if localRead {
		return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: --local-read requires --from and --index")
	}
	result, err := call(context.Background())
	if err != nil {
		return err
	}
	if result.GetFailed() {
		return writeKnowledgeFailure(result)
	}
	data, err := (protojson.MarshalOptions{UseProtoNames: true, Indent: "  "}).Marshal(result)
	if err != nil {
		return err
	}
	accepted, hasSelection := result.GetOutput().AsMap()["accepted"]
	incomplete := hasSelection && accepted == false
	destination := *output
	if destination == "" && (len(data) > 32<<10 || incomplete || operation == "read-source") {
		if result.GetCallId() == "" || strings.ContainsAny(result.GetCallId(), "/\\\x00") {
			return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
		}
		destination = filepath.Join("kb-results", result.GetCallId()+".json")
	}
	if destination != "" {
		if err := writeCommandResultFile(destination, data, false); err != nil {
			return err
		}
		summary := knowledgeResultSummary{Path: destination, Bytes: len(data), CallID: result.GetCallId()}
		if operation == "query" {
			out := result.GetOutput().AsMap()
			if available, ok := out["available"].(bool); ok && !available {
				summary.Available = &available
				summary.NextAction, _ = out["next_action"].(string)
				summary.ReadCommands = map[string][]string{
					"result": {"jq", "--", ".output | {available,next_action}", destination},
				}
			} else {
				summary.ResultPreview = knowledgeSQLResultPreview(out)
				summary.ReadCommands = map[string][]string{
					"result": {"jq", "--", ".output | {artifact_id,columns,rows,row_count,truncated,table_names,semantic_keys_used}", destination},
					"query":  {"jq", "--", ".output | {sql,table_names,semantic_keys_used}", destination},
				}
				if blocks, ok := out["query_blocks"].([]any); ok && len(blocks) > 0 {
					summary.ReadCommands["query_block_index"] = []string{"jq", "-c", "--", ".output.query_blocks[] | {id,cte,set_group,branch,relations}", destination}
					summary.ReadCommands["query_blocks"] = []string{"jq", "-c", "--", ".output.query_blocks[]", destination}
				}
				if repairs, ok := out["sql_repairs"].([]any); ok && len(repairs) > 0 {
					summary.ReadCommands["sql_repairs"] = []string{"jq", "-c", "--", ".output.sql_repairs[]", destination}
				}
				summary.ReadCommands["cite"] = []string{"moi-cli", "semantic-model", "cite", "--from", destination}
				summary.NextAction = "assess_requested_facts"
				summary.Guidance = "result_preview preserves complete cells; preview_complete describes only this preview, while truncated describes the SQL result. Use read_commands.result for returned rows. For long SQL, query_block_index locates CTEs and set branches; inspect only the needed output.query_blocks record by its id. Blocks describe SELECT clauses, not schema validity, inferred time units or business correctness; outer set ORDER BY/LIMIT remain in the executed sql. Use query for that exact executed SQL and provenance. If sql_repairs is available, the owner already applied those exact source edits; do not rerun a successful query to repair it again. Full data remains in path."
			}
		}
		if operation == "describe-tables" {
			if tables, ok := result.GetOutput().AsMap()["tables"].([]any); ok {
				summary.TableCount = len(tables)
			}
			entries := "[.output.tables[]?.semantic_entries[]?] | unique | .[]"
			summary.ReadCommands = map[string][]string{
				"tables":             {"jq", "--", ".output.tables | map(.name)", destination},
				"schema":             {"jq", "--", ".output.tables | map({name,columns,ddl})", destination},
				"schema_fields":      {"jq", "-c", "--", ".output.tables[] | {name,columns}", destination},
				"standard_questions": {"jq", "--", entries + " | select(.kind == \"standard_qa\") | {domain,kind,key_name,question,similar_questions,description,usage_mode}", destination},
				"business_rules":     {"jq", "--", entries + " | select(.kind == \"business_rule\")", destination},
			}
			summary.NextAction = "inspect_schema_metadata"
			summary.Guidance = "Use read_commands to inspect the exact table inventory and distinct semantic definitions without dumping every schema or long SQL. Identical definitions recur under related tables; distinct definitions with the same key remain separate. Full standard-question SQL stays in output.tables[].semantic_entries[].sql in the saved file: extract the applicable query directly. Schema metadata describes scope, not a business-value result."
		}
		if operation == "find-files" {
			summary.ReadCommands = map[string][]string{
				"files": {"jq", "--", ".output.files | to_entries[] | {index: .key, file_name: .value.file_name}", destination},
			}
			summary.NextAction = "select_source"
			summary.Guidance = "Inspect read_commands.files, then use --from with this saved local path and --index with the chosen inventory index for search or read-source. --from never accepts a document name or file ID. Inventory entries locate sources; they are not evidence."
		}
		if operation == "read-source" || ((operation == "search" || operation == "list-passages") && len(result.GetArtifacts()) > 0) {
			view, err := knowledgeReadingView(result, destination, operation == "read-source")
			if err != nil {
				return err
			}
			if err := writeCommandResultFile(destination+".txt", view.text, false); err != nil {
				return err
			}
			summary.ReadingPath, summary.ReadingOrder = destination+".txt", view.order
			summary.ReadCommands = map[string][]string{
				"content":  {"cat", "--", summary.ReadingPath},
				"evidence": {"jq", "--arg", "from", destination, "--", "[.artifacts[].data.chunks[]?] | to_entries[] | {from:$from,index:.key,file_name:.value.file_name,source_position:(.value | {page_number,chunk_index,chunk_index_scope}),content:.value.content}", destination},
			}
			summary.Guidance = "Inspect reading_path or execute read_commands.evidence for complete passages and exact locators. Prefer --from RESULT --quote 'EXACT INSPECTED SENTENCE' for text reads/citations: the CLI selects its unique canonical passage without positional assumptions. Preserve programmatic from/index locators when filtering or reordering; source_position describes document order, not citation indices. Register only inspected supporting evidence."
			if operation == "read-source" {
				summary.NextAction = "inspect_complete_source"
				if view.hasTables {
					summary.LocalAnalysis = map[string]any{
						"skill_id":           "moi.agent.momo.skill.moi-cli-semantic-model",
						"script":             "scripts/compare_passages.py",
						"base_arguments":     []string{"--source", destination, "--output", destination + ".inspection.json"},
						"required_arguments": []string{"--contains"},
					}
					summary.Guidance = "Inspect prose, prerequisites and executable examples in reading_path; filtered table matches do not replace a complete procedure. For table structure, use local_analysis with the already prepared Skill package; resolve skill_id through skills/index.json only if its path is unknown. Run its script with python3, base_arguments and the REQUIRED --contains followed by a literal source label; repeat --contains for distinct labels. Add --table-layout for table conditions, or --complete-tables for complete table statistics. Inspect matching rows AND narrative text through the returned report queries, instead of repeatedly dumping HTML or guessing JSON fields. Keep each action with its governing condition. Distinguish an incident or requested change from verified configured behavior; sequence alone does not prove success gating. The analysis retains this receipt's from/index pairs; register only inspected supporting evidence, then cite those pairs."
				}
			} else {
				summary.Coverage = "retrieved_passages_only"
			}
		}
		if entries, ok := result.GetOutput().AsMap()["semantic_entries"].([]any); ok && len(entries) > 0 {
			if summary.ReadCommands == nil {
				summary.ReadCommands = make(map[string][]string)
			}
			summary.ReadCommands["semantic_entries"] = []string{"jq", "--", ".output.semantic_entries[]", destination}
			summary.Guidance += " Read semantic_entries for applicable configured business rules and standard answers, including their source scope and usage_mode, before applying the retrieved passages."
		}
		if operation == "search" {
			out := result.GetOutput().AsMap()
			summary.RowCount, summary.ChunkCount, summary.Routes = out["row_count"], out["chunk_count"], out["routes"]
			// The Knowledge owner already bounds and focuses this preview. Keep
			// it visible without projecting or truncating the complete artifact.
			if preview, ok := result.GetOutput().AsMap()["model_view"].(string); ok && preview != "" {
				summary.EvidencePreview = preview
				if summary.ReadCommands == nil {
					summary.ReadCommands = make(map[string][]string)
				}
				summary.ReadCommands["preview"] = []string{"jq", "-r", "--", ".output.model_view", destination}
			}
		}
		if operation == "read-evidence" {
			out := result.GetOutput().AsMap()
			if out["kind"] == "knowledge_read_batch" || out["kind"] == "knowledge_local_read_batch" {
				reads, _ := out["reads"].([]any) // Created by readKnowledgeBatch.
				receiptIndices := make([]int, len(reads))
				for index := range reads {
					receiptIndices[index] = index
				}
				summary.ReceiptIndices = receiptIndices
				summary.Guidance = "These indices belong to this saved receipt. Choose supporting evidence with --from this path and --indices for cite; input_index refers to the original retrieval file. Do not copy artifact/chunk IDs or use --local-read on a read receipt."
				if out["kind"] == "knowledge_read_batch" {
					summary.ReadCommands = map[string][]string{
						"content": {"jq", "--", ".output.reads | to_entries[] | {index: .key, input_index: .value.input_index, evidence: .value.output}", destination},
					}
				}
			} else if _, ok := out["content"].(string); ok {
				summary.ReadCommands = map[string][]string{
					"content": {"jq", "--", ".output", destination},
				}
			}
		}
		if page, ok := result.GetOutput().AsMap()["source_page"]; ok {
			summary.SourcePage = page
		}
		summary.Rerank = result.GetOutput().AsMap()["rerank"]
		if incomplete {
			summary.Accepted = new(bool)
			summary.NextAction = "register_inspected_sources_or_read_unseen_evidence"
			summary.Guidance = "If you already inspected complete canonical passages locally, register the supporting original indices with read-evidence --from ORIGINAL_RESULT --indices ORIGINAL_INDICES --local-read, then repeat cite. This rejection file contains instructions, not source text; do not use --local-read on it. next_read fetches evidence that has not been inspected. Do not reduce a total's selected sources to avoid reading requirements."
			var missing []string
			sources, ok := result.GetOutput().AsMap()["sources"].([]any)
			if !ok {
				return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
			}
			for i := range sources {
				missing = append(missing, strconv.Itoa(i))
			}
			summary.NextRead = [][]string{{"moi-cli", "semantic-model", "read-evidence", "--from", destination, "--indices", strings.Join(missing, ",")}}
		}
		if err := json.NewEncoder(os.Stdout).Encode(summary); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintln(os.Stdout, string(data)); err != nil {
			return err
		}
	}
	if incomplete {
		return fmt.Errorf("CLI_KNOWLEDGE_EVIDENCE_INCOMPLETE")
	}
	return nil
}

// Keep small aggregate values ahead of the potentially long SQL and its local
// readers. A preview never clips a cell or alters the complete saved result.
func knowledgeSQLResultPreview(output map[string]any) map[string]any {
	preview := map[string]any{
		"artifact_id": output["artifact_id"], "row_count": output["row_count"],
		"truncated": output["truncated"] == true, "preview_complete": false,
	}
	const budget = 2048
	fits := func() bool {
		raw, err := json.Marshal(preview)
		return err == nil && len(raw) <= budget
	}
	preview["columns"] = output["columns"]
	if !fits() {
		delete(preview, "columns")
		return preview
	}
	rows, ok := output["rows"].([]any)
	if !ok {
		return preview
	}
	selected := make([]any, 0)
	preview["rows"] = selected
	for _, row := range rows {
		preview["rows"] = append(selected, row)
		if !fits() {
			preview["rows"] = selected
			return preview
		}
		selected = preview["rows"].([]any)
	}
	preview["preview_complete"] = true
	return preview
}

func knowledgeRequestFromResult(body proto.Message, paths []string, index int, indices []string, localRead bool, quote string) error {
	type candidate struct {
		ArtifactID  string
		Chunk       map[string]any
		ReadReceipt bool
		SQLResult   bool
	}
	var candidates []candidate
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return &sdk.Error{
					Code:    "CLI_KNOWLEDGE_LOCAL_RESULT_NOT_FOUND",
					Message: "--from requires an existing saved local result JSON, not a document name or file ID. Use the path returned by --output with --index to select the source.",
					Details: map[string]any{"path": path, "flag": "--from", "expected": "saved_local_result_path"},
				}
			}
			return err
		}
		result := &sdk.KnowledgeRuntimeResult{}
		if err = protojson.Unmarshal(raw, result); err != nil || result.GetFailed() {
			return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: %v", err)
		}
		if _, cite := body.(*sdk.KnowledgeCiteRequest); cite {
			output := result.GetOutput().AsMap()
			id, _ := output["artifact_id"].(string)
			sql, _ := output["sql"].(string)
			_, columns := output["columns"].([]any)
			_, rows := output["rows"].([]any)
			rows = rows || (output["rows"] == nil && output["row_count"] == float64(0))
			if id != "" && sql != "" && columns && rows {
				// SQL evidence is the whole owner-issued query result, not a
				// passage or row index. The gateway still validates this task's
				// durable artifact identity before accepting the citation.
				candidates = append(candidates, candidate{ArtifactID: id, SQLResult: true})
				continue
			}
		}
		if request, ok := body.(*sdk.KnowledgeSearchRequest); ok && request.SourcePage != nil {
			if rawPage, present := result.GetOutput().AsMap()["source_page"]; present {
				if len(paths) != 1 || index != -1 || request.SourcePage.Cursor != "" || request.SourcePage.FileId != "" || request.SourcePage.SemanticModelId != 0 {
					return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: continue with --from PAGE only; do not combine its source and cursor with other selectors")
				}
				page, ok := rawPage.(map[string]any)
				if !ok {
					return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing source page")
				}
				fileID, modelID, err := knowledgeResultSourceIdentity(page)
				if err != nil {
					return err
				}
				hasMore, valid := page["has_more"].(bool)
				if !valid {
					return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing continuation status")
				}
				if !hasMore {
					return fmt.Errorf("CLI_KNOWLEDGE_SOURCE_COMPLETE: this source page has no continuation")
				}
				cursor, ok := page["next_cursor"].(string)
				if !ok || cursor == "" {
					return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing continuation cursor")
				}
				request.SourcePage.FileId, request.SourcePage.SemanticModelId, request.SourcePage.Cursor = fileID, modelID, cursor
				return nil
			}
		}
		reads, err := knowledgeResultReads(result.GetOutput().AsMap())
		if err != nil {
			return err
		}
		if reads != nil {
			for _, read := range reads {
				id, _ := read["artifact_id"].(string)
				chunkID, _ := read["chunk_id"].(string)
				if _, _, err := knowledgeResultSourceIdentity(read); err != nil || id == "" || chunkID == "" {
					return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
				}
				candidates = append(candidates, candidate{ArtifactID: id, Chunk: read, ReadReceipt: true})
			}
			continue
		}
		for _, artifact := range result.GetArtifacts() {
			value := artifact.AsMap()
			data, _ := value["data"].(map[string]any)
			chunks, _ := data["chunks"].([]any)
			for _, raw := range chunks {
				chunk, ok := raw.(map[string]any)
				if !ok {
					return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
				}
				id, _ := value["artifact_id"].(string)
				candidates = append(candidates, candidate{ArtifactID: id, Chunk: chunk})
			}
		}
		if files, ok := result.GetOutput().AsMap()["files"].([]any); ok {
			for _, raw := range files {
				file, ok := raw.(map[string]any)
				if !ok {
					return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
				}
				candidates = append(candidates, candidate{Chunk: file})
			}
		}
		if _, read := body.(*sdk.KnowledgeReadEvidenceRequest); read {
			// Citation repair receipts carry the owner's exact next read, with
			// the same provenance and cursor as the original reader contract.
			if sources, ok := result.GetOutput().AsMap()["sources"].([]any); ok {
				for _, raw := range sources {
					source, _ := raw.(map[string]any)
					next, _ := source["next_call"].(map[string]any)
					args, ok := next["arguments"].(map[string]any)
					if !ok || (next["tool_id"] != "read_artifact" && next["tool_id"] != "read_artifact_page") {
						return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
					}
					id, _ := args["artifact_id"].(string)
					candidates = append(candidates, candidate{ArtifactID: id, Chunk: args})
				}
			}
		}
	}
	if quote != "" {
		var matches []int
		for i, item := range candidates {
			content, canonical := item.Chunk["content"].(string)
			if canonical && !item.ReadReceipt && item.ArtifactID != "" && strings.Contains(content, quote) {
				matches = append(matches, i)
			}
		}
		if len(matches) == 0 {
			return &sdk.Error{Code: "CLI_KNOWLEDGE_QUOTE_NOT_FOUND", Message: "The exact quote does not occur in a canonical passage in this saved result. Inspect read_commands.evidence and copy the supporting text without changing it.", Details: map[string]any{"quote": quote, "paths": paths}}
		}
		if len(matches) > 1 {
			return &sdk.Error{Code: "CLI_KNOWLEDGE_QUOTE_AMBIGUOUS", Message: "The quote occurs in multiple passages. Use a more distinctive exact quote or an explicit inspected from/index locator; no passage was selected.", Details: map[string]any{"quote": quote, "paths": paths, "candidate_indices": matches}}
		}
		index = matches[0]
		if _, cite := body.(*sdk.KnowledgeCiteRequest); cite {
			indices = []string{strconv.Itoa(index)}
		}
	}
	selectItem := func(i int) (candidate, error) {
		if i < 0 || i >= len(candidates) {
			return candidate{}, fmt.Errorf("CLI_KNOWLEDGE_INDEX_OUT_OF_RANGE: %d", i)
		}
		return candidates[i], nil
	}
	switch v := body.(type) {
	case *sdk.KnowledgeSearchRequest:
		item, err := selectItem(index)
		if err != nil {
			return err
		}
		fileID, _ := item.Chunk["file_id"].(string)
		if fileID == "" {
			return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
		}
		v.FileIds = []string{fileID}
		if v.SourcePage != nil {
			fileID, modelID, err := knowledgeResultSourceIdentity(item.Chunk)
			if err != nil {
				return err
			}
			v.SourcePage.FileId, v.SourcePage.SemanticModelId = fileID, modelID
			v.FileIds = nil
		}
	case *sdk.KnowledgeReadEvidenceRequest:
		item, err := selectItem(index)
		if err != nil {
			return err
		}
		if item.ArtifactID == "" {
			return fmt.Errorf("CLI_KNOWLEDGE_RESULT_MISSING_RETRIEVAL: search this file with --from and --index first")
		}
		v.ArtifactId = item.ArtifactID
		if item.ReadReceipt {
			if localRead {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: --local-read requires complete canonical passage content, not a read receipt")
			}
			if v.Cursor == nil {
				if item.Chunk["has_more"] != true {
					return fmt.Errorf("CLI_KNOWLEDGE_SOURCE_COMPLETE: this read receipt has no continuation")
				}
				cursor, ok := item.Chunk["next_cursor"].(float64)
				if !ok || cursor < 0 || math.Trunc(cursor) != cursor {
					return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
				}
				v.Cursor = structpb.NewNumberValue(cursor)
			}
		}
		if part, ok := item.Chunk["part_index"].(float64); ok {
			if part < 0 || part > math.MaxInt32 || math.Trunc(part) != part {
				return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
			}
			n := int32(part)
			v.PartIndex = &n
		}
		if v.PartIndex != nil {
			if localRead {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: local read cannot use part_index")
			}
			return nil
		}
		v.ChunkId, _ = item.Chunk["chunk_id"].(string)
		v.FileId, _ = item.Chunk["file_id"].(string)
		if id, ok := item.Chunk["semantic_model_id"].(float64); ok {
			n := int64(id)
			v.SemanticModelId = &n
		}
		if localRead {
			content, ok := item.Chunk["content"].(string)
			if !ok || v.Cursor != nil || v.LimitChars != 0 {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: local read requires complete saved passage content and no pagination")
			}
			digest := sha256.Sum256([]byte(content))
			expected := hex.EncodeToString(digest[:])
			if v.LocalContentSha256 != "" && v.LocalContentSha256 != expected {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: local content digest conflicts with saved passage")
			}
			v.LocalContentSha256 = expected
		}
		if cursor, ok := item.Chunk["cursor"]; ok && v.Cursor == nil && !localRead {
			v.Cursor, err = structpb.NewValue(cursor)
			if err != nil {
				return err
			}
		}
	case *sdk.KnowledgeReadDocumentRequest:
		item, err := selectItem(index)
		if err != nil {
			return err
		}
		v.MarkdownFileId, _ = item.Chunk["markdown_file_id"].(string)
		if v.MarkdownFileId == "" {
			return fmt.Errorf("CLI_KNOWLEDGE_RESULT_MISSING_MARKDOWN_LOCATOR: use search --from and --index for this file, then read-evidence")
		}
	case *sdk.KnowledgeSearchDocumentRequest:
		item, err := selectItem(index)
		if err != nil {
			return err
		}
		v.MarkdownFileId, _ = item.Chunk["markdown_file_id"].(string)
		if v.MarkdownFileId == "" {
			return fmt.Errorf("CLI_KNOWLEDGE_RESULT_MISSING_MARKDOWN_LOCATOR: use search --from and --index for this file, then read-evidence")
		}
	case *sdk.KnowledgeCiteRequest:
		if len(indices) == 0 {
			if len(candidates) == 0 || slices.ContainsFunc(candidates, func(item candidate) bool { return !item.SQLResult }) {
				return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: passage citations require --quote or non-empty --indices")
			}
			for i := range candidates {
				indices = append(indices, strconv.Itoa(i))
			}
		}
		for _, text := range indices {
			n, err := strconv.Atoi(text)
			if err != nil {
				return err
			}
			item, err := selectItem(n)
			if err != nil {
				return err
			}
			if item.ArtifactID == "" {
				return fmt.Errorf("CLI_KNOWLEDGE_RESULT_MISSING_RETRIEVAL: search this file with --from and --index first")
			}
			source := map[string]any{"type": "rag_chunk", "artifact_id": item.ArtifactID}
			if item.SQLResult {
				source["type"] = "sql_result"
			}
			for _, key := range []string{"chunk_id", "file_id", "semantic_model_id"} {
				if x, ok := item.Chunk[key]; ok {
					source[key] = x
				}
			}
			pb, err := structpb.NewStruct(source)
			if err != nil {
				return err
			}
			v.Sources = append(v.Sources, pb)
		}
	default:
		return fmt.Errorf("CLI_KNOWLEDGE_FROM_UNSUPPORTED")
	}
	return nil
}

// knowledgeResultReads recognizes only the owner's chunk-read receipts and
// the CLI's explicit batch envelope. Receipt positions are local to this file;
// input_index describes the original selection and is never a new selector.
func knowledgeResultReads(output map[string]any) ([]map[string]any, error) {
	kind, _ := output["kind"].(string)
	if kind == "knowledge_read_batch" || kind == "knowledge_local_read_batch" {
		entries, ok := output["reads"].([]any)
		if !ok || len(entries) == 0 {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
		}
		reads := make([]map[string]any, 0, len(entries))
		for _, entry := range entries {
			entry, _ := entry.(map[string]any)
			read, _ := entry["output"].(map[string]any)
			if !knowledgeIsChunkRead(read) {
				return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
			}
			reads = append(reads, read)
		}
		return reads, nil
	}
	if knowledgeIsChunkRead(output) {
		return []map[string]any{output}, nil
	}
	return nil, nil
}

func knowledgeIsChunkRead(output map[string]any) bool {
	if output["kind"] == "rag_local_evidence_read" {
		return true
	}
	_, window := output["has_more"].(bool)
	return window && output["chunk_id"] != nil && output["artifact_id"] != nil
}

type knowledgeRead struct {
	index   int
	request *sdk.KnowledgeReadEvidenceRequest
}

func readKnowledgeBatch(ctx context.Context, service *sdk.KnowledgeService, requestID string, reads []knowledgeRead, localRead bool) (*sdk.KnowledgeRuntimeResult, error) {
	var receipts []any
	var last *sdk.KnowledgeRuntimeResult
	var artifacts []*structpb.Struct
	if localRead {
		request := &sdk.KnowledgeReadEvidenceRequest{RequestId: requestID}
		for _, read := range reads {
			v := read.request
			request.LocalReads = append(request.LocalReads, &sdk.KnowledgeLocalEvidenceRead{ArtifactId: v.ArtifactId, ChunkId: v.ChunkId, FileId: v.FileId, SemanticModelId: v.SemanticModelId, LocalContentSha256: v.LocalContentSha256})
		}
		result, err := service.ReadEvidence(ctx, request)
		if err != nil || result.GetFailed() {
			return result, err
		}
		ownerOutput := result.GetOutput().AsMap()
		ownerReads, ok := ownerOutput["reads"].([]any)
		if !ok || ownerOutput["kind"] != "rag_local_evidence_read_batch" || len(ownerReads) != len(reads) {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: incomplete batch receipt")
		}
		for position, output := range ownerReads {
			receipt, ok := output.(map[string]any)
			requested := reads[position].request
			if !ok || receipt["kind"] != "rag_local_evidence_read" || receipt["artifact_id"] != requested.ArtifactId || receipt["chunk_id"] != requested.ChunkId || receipt["content_sha256"] != requested.LocalContentSha256 {
				return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: batch receipt changed requested evidence")
			}
			fileID, modelID, err := knowledgeResultSourceIdentity(receipt)
			if err != nil || (requested.FileId != "" && requested.FileId != fileID) || (requested.SemanticModelId != nil && requested.GetSemanticModelId() != modelID) {
				return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: batch receipt changed source identity")
			}
			receipts = append(receipts, map[string]any{"input_index": reads[position].index, "call_id": result.CallId, "output": output})
		}
		body, err := structpb.NewStruct(map[string]any{"kind": "knowledge_local_read_batch", "reads": receipts})
		if err != nil {
			return nil, err
		}
		result.Output = body
		return result, nil
	}
	for _, read := range reads {
		result, err := service.ReadEvidence(ctx, read.request)
		if err != nil {
			return nil, err
		}
		if result.GetFailed() {
			return result, nil
		}
		receipts = append(receipts, map[string]any{"input_index": read.index, "call_id": result.CallId, "output": result.GetOutput().AsMap()})
		artifacts = append(artifacts, result.Artifacts...)
		last = result
	}
	output, err := structpb.NewStruct(map[string]any{"kind": "knowledge_read_batch", "reads": receipts})
	if err != nil {
		return nil, err
	}
	return &sdk.KnowledgeRuntimeResult{CallId: last.CallId, Output: output, Artifacts: artifacts, Metadata: last.Metadata}, nil
}

func knowledgeResultSourceIdentity(source map[string]any) (string, int64, error) {
	fileID, _ := source["file_id"].(string)
	modelID, ok := source["semantic_model_id"].(float64)
	if fileID == "" || !ok || modelID <= 0 || modelID >= float64(math.MaxInt64) || math.Trunc(modelID) != modelID {
		return "", 0, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing source identity")
	}
	return fileID, int64(modelID), nil
}

func knowledgeKeywordsFlag(keywords *[]string) func(string) error {
	return func(value string) error {
		*keywords = append(*keywords, splitCSV(value)...)
		return nil
	}
}

func knowledgePageSizeFlag(page *sdk.KnowledgeSourcePage) func(string) error {
	return func(value string) error {
		n, err := strconv.ParseInt(value, 10, 32)
		if err != nil || n < 0 {
			return fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: invalid page-size")
		}
		page.PageSize = int32(n)
		return nil
	}
}

// KnowledgeRuntimeFlagSupported is shared by execution, help and the reference
// generator. Other flags are declared only in their operation's own FlagSet.
func KnowledgeRuntimeFlagSupported(operation, name string) bool {
	switch name {
	case "quote":
		return operation == "read-evidence" || operation == "cite"
	case "local-read":
		return operation == "read-evidence"
	case "from", "index", "indices":
		switch operation {
		case "read-evidence":
			return true
		case "search", "list-passages", "read-source", "read-document", "search-document":
			return name != "indices"
		case "cite":
			return name != "index"
		default:
			return false
		}
	}
	return true
}

func knowledgeRuntimeJSONFieldSupported(operation, name string) bool {
	if operation == "list-passages" || operation == "read-source" {
		return name == "request_id" || name == "source_page"
	}
	return true
}
