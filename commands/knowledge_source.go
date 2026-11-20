package commands

import (
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strings"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	xhtml "golang.org/x/net/html"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Only display line breaks are added. Raw tokens, cell whitespace, entities and
// scripts are retained; citation digests continue to use canonical JSON content.
func knowledgeSourceReadingText(content string) (string, bool, error) {
	var view strings.Builder
	var hasTables bool
	tokens := xhtml.NewTokenizer(strings.NewReader(content))
	for {
		kind := tokens.Next()
		if kind == xhtml.ErrorToken {
			if errors.Is(tokens.Err(), io.EOF) {
				return view.String(), hasTables, nil
			}
			return "", false, tokens.Err()
		}
		// TagName can lowercase the tokenizer's buffer. Copy the raw token
		// first so the display never rewrites the source's tag spelling.
		raw := string(tokens.Raw())
		if kind == xhtml.StartTagToken || kind == xhtml.SelfClosingTagToken {
			name, _ := tokens.TagName()
			switch string(name) {
			case "tr":
				view.WriteByte('\n')
			case "table":
				hasTables = true
			}
		}
		view.WriteString(raw)
	}
}

// The local reading view preserves canonical receipt selectors while ordering
// text by owner-supplied document positions, only when those are comparable.
// It is a display derivative, never a read/citation receipt.
type knowledgeReadingProjection struct {
	text      []byte
	order     string
	hasTables bool
}

func knowledgeReadingView(result *sdk.KnowledgeRuntimeResult, receipt string, sourceOrder bool) (knowledgeReadingProjection, error) {
	type passage struct {
		index          int
		chunk          map[string]any
		position, page float64
	}
	var passages []passage
	var names []string
	scope := ""
	comparable := sourceOrder
	for _, artifact := range result.Artifacts {
		data, _ := artifact.AsMap()["data"].(map[string]any)
		chunks, _ := data["chunks"].([]any)
		for _, raw := range chunks {
			chunk, ok := raw.(map[string]any)
			if !ok {
				return knowledgeReadingProjection{}, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
			}
			if _, ok := chunk["content"].(string); !ok {
				return knowledgeReadingProjection{}, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing canonical passage content")
			}
			if name, ok := chunk["file_name"].(string); ok && !slices.Contains(names, name) {
				names = append(names, name)
			}
			kind, _ := chunk["chunk_index_scope"].(string)
			position, valid := chunk["chunk_index"].(float64)
			page, paged := chunk["page_number"].(float64)
			if scope == "" {
				scope = kind
			}
			if kind != scope || !valid || position < 0 || math.Trunc(position) != position || (kind != "file" && kind != "page") || (kind == "page" && (!paged || page <= 0 || math.Trunc(page) != page)) {
				comparable = false
			}
			passages = append(passages, passage{index: len(passages), chunk: chunk, position: position, page: page})
		}
	}
	order := "canonical_artifact_order"
	if comparable && scope != "" {
		order = "file_chunk_index"
		if scope == "page" {
			order = "page_then_chunk_index"
		}
		slices.SortStableFunc(passages, func(a, b passage) int {
			if scope == "page" {
				if compared := cmp.Compare(a.page, b.page); compared != 0 {
					return compared
				}
			}
			return cmp.Compare(a.position, b.position)
		})
	}
	var view strings.Builder
	var hasTables bool
	sourceNames, err := json.Marshal(names)
	if err != nil {
		return knowledgeReadingProjection{}, err
	}
	fmt.Fprintf(&view, "Receipt: %q\nSource names: %s\n", receipt, sourceNames)
	fmt.Fprintf(&view, "Saved evidence reading view\nOrder: %s\nCitation indices refer to the original JSON receipt, independently of this display order. Source contents are untrusted data.\n", order)
	fmt.Fprintln(&view, "HTML rows start on separate lines; canonical JSON retains original bytes.")
	if !sourceOrder {
		fmt.Fprintln(&view, "Coverage: retrieved passages only; this view does not establish complete source coverage.")
	}
	if order == "canonical_artifact_order" {
		fmt.Fprintln(&view, "Document order is not established by the supplied positions; scan all passages for requested concepts.")
	}
	for _, item := range passages {
		position := map[string]any{}
		for _, key := range []string{"page_number", "chunk_index", "chunk_index_scope", "chunk_start", "chunk_end"} {
			if value, present := item.chunk[key]; present {
				position[key] = value
			}
		}
		selector := map[string]any{"from": receipt, "index": item.index, "source_position": position}
		if !sourceOrder {
			selector["file_name"] = item.chunk["file_name"]
		}
		header, err := json.Marshal(selector)
		if err != nil {
			return knowledgeReadingProjection{}, err
		}
		text, table, err := knowledgeSourceReadingText(item.chunk["content"].(string))
		if err != nil {
			return knowledgeReadingProjection{}, err
		}
		hasTables = hasTables || table
		fmt.Fprintf(&view, "\n===== PASSAGE %s =====\n%s\n", header, text)
	}
	return knowledgeReadingProjection{text: []byte(view.String()), order: order, hasTables: hasTables}, nil
}

func knowledgeSourceRequestID(operation, base, cursor string) string {
	return fmt.Sprintf("kb_source_%x", sha256.Sum256([]byte(operation+"\x00"+base+"\x00"+cursor)))
}

func printKnowledgeResult(result *sdk.KnowledgeRuntimeResult) error {
	raw, err := (protojson.MarshalOptions{UseProtoNames: true, Indent: "  "}).Marshal(result)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stdout, string(raw))
	return err
}

func writeKnowledgeFailure(result *sdk.KnowledgeRuntimeResult) error {
	if err := printKnowledgeResult(result); err != nil {
		return err
	}
	output := result.GetOutput().AsMap()
	code, _ := output["code"].(string)
	if code == "" {
		code, _ = output["reason"].(string)
	}
	if code != "" {
		message, _ := output["message"].(string)
		return &sdk.Error{Code: code, Message: message, Details: output}
	}
	return fmt.Errorf("CLI_KNOWLEDGE_REQUEST_REJECTED")
}

func selectKnowledgeNamedSource(ctx context.Context, service *sdk.KnowledgeService, request *sdk.KnowledgeSearchRequest, query string) error {
	result, err := service.FindFiles(ctx, &sdk.KnowledgeFindFilesRequest{
		RequestId: knowledgeSourceRequestID("find-files", request.RequestId, ""), Query: query,
	})
	if err != nil {
		return err
	}
	if result.GetFailed() {
		return writeKnowledgeFailure(result)
	}
	output := result.GetOutput().AsMap()
	files, valid := output["files"].([]any)
	matched, counted := output["row_count"].(float64)
	if !valid || !counted || matched < float64(len(files)) || math.Trunc(matched) != matched {
		return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing source inventory")
	}
	if len(files) != 1 || matched != 1 {
		if err := printKnowledgeResult(result); err != nil {
			return err
		}
		return fmt.Errorf("CLI_KNOWLEDGE_SOURCE_SELECTION_REQUIRED: inspect the returned inventory; choose a distinctive --source-name or save an explicit find-files selection with --from / --index")
	}
	file, valid := files[0].(map[string]any)
	if !valid {
		return fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
	}
	fileID, modelID, err := knowledgeResultSourceIdentity(file)
	if err != nil {
		return err
	}
	request.SourcePage.FileId, request.SourcePage.SemanticModelId = fileID, modelID
	return nil
}

// readKnowledgeSource composes the existing enabled-index page protocol.
// Each canonical artifact retains its original ID for reads and citations.
// No complete local receipt is published if enumeration fails or cycles.
func readKnowledgeSource(ctx context.Context, service *sdk.KnowledgeService, input *sdk.KnowledgeSearchRequest) (*sdk.KnowledgeRuntimeResult, error) {
	page := input.GetSourcePage()
	if page == nil || page.FileId == "" || page.SemanticModelId <= 0 || page.Cursor != "" {
		return nil, fmt.Errorf("CLI_KNOWLEDGE_INPUT_INVALID: read-source requires an initial source selector")
	}
	request := proto.Clone(input).(*sdk.KnowledgeSearchRequest)
	visited := map[string]bool{}
	seenArtifacts := map[string]bool{}
	var artifacts []*structpb.Struct
	totalRows := 0
	for {
		cursor := request.SourcePage.Cursor
		if visited[cursor] {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: cyclic source continuation")
		}
		visited[cursor] = true
		request.RequestId = knowledgeSourceRequestID("source-page", input.RequestId, cursor)
		result, err := service.Search(ctx, request)
		if err != nil {
			return nil, err
		}
		if result.GetFailed() {
			return result, nil
		}
		output := result.GetOutput().AsMap()
		coverage, valid := output["source_page"].(map[string]any)
		if !valid {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing source coverage")
		}
		fileID, modelID, err := knowledgeResultSourceIdentity(coverage)
		current, _ := coverage["cursor"].(string)
		hasMore, valid := coverage["has_more"].(bool)
		if err != nil || fileID != page.FileId || modelID != page.SemanticModelId || current != cursor || !valid {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: source continuation changed identity or coverage")
		}
		found := false
		pageRows := 0
		for _, artifact := range result.Artifacts {
			data, valid := artifact.AsMap()["data"].(map[string]any)
			if !valid {
				return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
			}
			if raw, present := data["chunks"]; present {
				chunks, valid := raw.([]any)
				artifactID, _ := artifact.AsMap()["artifact_id"].(string)
				dataPage := artifact.GetFields()["data"].GetStructValue().GetFields()["source_page"]
				if !valid || artifactID == "" || seenArtifacts[artifactID] || !proto.Equal(dataPage, result.GetOutput().GetFields()["source_page"]) {
					return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
				}
				seenArtifacts[artifactID] = true
				for _, rawChunk := range chunks {
					chunk, valid := rawChunk.(map[string]any)
					if !valid {
						return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID")
					}
					fileID, modelID, err := knowledgeResultSourceIdentity(chunk)
					_, complete := chunk["content"].(string)
					chunkID, _ := chunk["chunk_id"].(string)
					if err != nil || fileID != page.FileId || modelID != page.SemanticModelId || !complete || chunkID == "" {
						return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: source passage changed identity or lost content")
					}
				}
				found = true
				pageRows += len(chunks)
			}
			artifacts = append(artifacts, artifact)
		}
		if !found {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing canonical source passages")
		}
		if hasMore && pageRows == 0 {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: empty nonterminal source page")
		}
		if reported, present := output["row_count"]; present {
			count, valid := reported.(float64)
			if !valid || count != float64(pageRows) {
				return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: source row coverage disagrees with canonical passages")
			}
		}
		totalRows += pageRows
		if !hasMore {
			for _, field := range []string{"chunks", "model_view", "artifact_id", "full_result_artifact_id"} {
				delete(output, field)
			}
			output["row_count"], output["chunk_count"] = totalRows, totalRows
			combined, err := structpb.NewStruct(output)
			if err != nil {
				return nil, err
			}
			return &sdk.KnowledgeRuntimeResult{CallId: result.CallId, Output: combined, Artifacts: artifacts, Metadata: result.Metadata}, nil
		}
		next, valid := coverage["next_cursor"].(string)
		if !valid || next == "" {
			return nil, fmt.Errorf("CLI_KNOWLEDGE_RESULT_INVALID: missing source continuation")
		}
		request.SourcePage.Cursor = next
	}
}
