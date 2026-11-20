package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type knowledgeLifecycleContract struct {
	Name            string                     `json:"name"`
	Operation       string                     `json:"operation"`
	Method          string                     `json:"method"`
	Path            string                     `json:"path"`
	Query           map[string][]string        `json:"query"`
	Input           knowledgeLifecycleInput    `json:"input"`
	Request         json.RawMessage            `json:"request"`
	Multipart       *catalogUploadCLIMultipart `json:"multipart"`
	Response        json.RawMessage            `json:"response"`
	ResponseHeaders map[string]string          `json:"response_headers"`
	Status          int                        `json:"status"`
}

type knowledgeLifecycleInput struct {
	ModelID                        string                        `json:"model_id"`
	SourceID                       string                        `json:"source_id"`
	EntryID                        string                        `json:"entry_id"`
	SegmentID                      string                        `json:"segment_id"`
	SourceFileID                   string                        `json:"source_file_id"`
	KnowledgeBaseID                int64                         `json:"knowledge_base_id"`
	DrawingNumber                  string                        `json:"drawing_number"`
	QueryText                      string                        `json:"query_text"`
	TopK                           int                           `json:"top_k"`
	TopKPresent                    bool                          `json:"top_k_present"`
	PageNumber                     int                           `json:"page_number"`
	PageNumberPresent              bool                          `json:"page_number_present"`
	Name                           string                        `json:"name"`
	Description                    string                        `json:"description"`
	DescriptionPresent             bool                          `json:"description_present"`
	ImageIndexEnabled              bool                          `json:"image_index_enabled"`
	ImageIndexEnabledPresent       bool                          `json:"image_index_enabled_present"`
	Request                        json.RawMessage               `json:"request"`
	Requests                       []json.RawMessage             `json:"requests"`
	Filename                       string                        `json:"filename"`
	ContentBase64                  string                        `json:"content_base64"`
	Selections                     []knowledgeLifecycleSelection `json:"selections"`
	SelectionsPresent              bool                          `json:"selections_present"`
	Sources                        []knowledgeLifecycleSource    `json:"sources"`
	Limit                          int                           `json:"limit"`
	LimitPresent                   bool                          `json:"limit_present"`
	Cursor                         string                        `json:"cursor"`
	RetrySourceRowID               string                        `json:"retry_source_row_id"`
	RetryPresent                   bool                          `json:"retry_present"`
	Page                           int                           `json:"page"`
	PageSize                       int                           `json:"page_size"`
	IncludeDisplayStatus           bool                          `json:"include_display_status"`
	IncludeDisplayStatusPresent    bool                          `json:"include_display_status_present"`
	FileIDs                        []string                      `json:"file_ids"`
	TableIDs                       []int64                       `json:"table_ids"`
	Entry                          knowledgeLifecycleEntry       `json:"entry"`
	Entries                        []knowledgeLifecycleEntry     `json:"entries"`
	IfMatch                        string                        `json:"if_match"`
	Level                          string                        `json:"level"`
	Content                        *string                       `json:"content"`
	OCRText                        *string                       `json:"ocr_text"`
	ImageDescription               *string                       `json:"image_description"`
	ImageFileID                    string                        `json:"image_file_id"`
	PageImageFileID                string                        `json:"page_image_file_id"`
	BBox                           any                           `json:"bbox"`
	Metadata                       any                           `json:"metadata"`
	Enabled                        bool                          `json:"enabled"`
	EnabledPresent                 bool                          `json:"enabled_present"`
	Tags                           []string                      `json:"tags"`
	TagsPresent                    bool                          `json:"tags_present"`
	ExpiresAt                      int64                         `json:"expires_at"`
	ExpiresAtPresent               bool                          `json:"expires_at_present"`
	ClearExpiresAt                 bool                          `json:"clear_expires_at"`
	ForceEnabledAfterExpiry        bool                          `json:"force_enabled_after_expiry"`
	ForceEnabledAfterExpiryPresent bool                          `json:"force_enabled_after_expiry_present"`
}

type knowledgeLifecycleEntry struct {
	Kind   string   `json:"kind"`
	Key    string   `json:"key"`
	Tables []string `json:"tables"`
	Spec   any      `json:"spec"`
	Domain string   `json:"domain"`
}

func knowledgeLifecycleSDKEntry(value knowledgeLifecycleEntry) sdk.SemanticEntryInput {
	return sdk.SemanticEntryInput{Kind: value.Kind, Key: value.Key, Tables: value.Tables, Spec: value.Spec, Domain: value.Domain}
}

func knowledgeLifecycleSDKEntries(values []knowledgeLifecycleEntry) []sdk.SemanticEntryInput {
	result := make([]sdk.SemanticEntryInput, 0, len(values))
	for _, value := range values {
		result = append(result, knowledgeLifecycleSDKEntry(value))
	}
	return result
}

type knowledgeLifecycleSource struct {
	SourceType  string `json:"source_type"`
	FileName    string `json:"file_name"`
	UploadKind  string `json:"upload_kind"`
	TableConfig string `json:"table_config"`
	FileID      string `json:"file_id"`
	VolumeID    int64  `json:"volume_id"`
	TableID     int64  `json:"table_id"`
}

type knowledgeLifecycleSelection struct {
	Kind             string   `json:"kind"`
	DatabaseID       int64    `json:"database_id"`
	VolumeID         int64    `json:"volume_id"`
	AllSelected      bool     `json:"all_selected"`
	SelectedTableIDs []int64  `json:"selected_table_ids"`
	ExcludedTableIDs []int64  `json:"excluded_table_ids"`
	SelectedFileIDs  []string `json:"selected_file_ids"`
	ExcludedFileIDs  []string `json:"excluded_file_ids"`
	Filters          struct {
		TableName string   `json:"table_name"`
		FileName  string   `json:"file_name"`
		FileExt   []string `json:"file_ext"`
	} `json:"filters"`
}

func knowledgeLifecycleSDKSources(values []knowledgeLifecycleSource) []sdk.SemanticModelSourceInput {
	result := make([]sdk.SemanticModelSourceInput, 0, len(values))
	for _, value := range values {
		result = append(result, sdk.SemanticModelSourceInput{
			SourceType: value.SourceType, FileName: value.FileName, UploadKind: value.UploadKind,
			TableConfig: value.TableConfig, FileID: value.FileID, VolumeID: value.VolumeID, TableID: value.TableID,
		})
	}
	return result
}

func knowledgeLifecycleSDKSelections(values []knowledgeLifecycleSelection) []sdk.SemanticModelSourceSelectionInput {
	result := make([]sdk.SemanticModelSourceSelectionInput, 0, len(values))
	for _, value := range values {
		result = append(result, sdk.SemanticModelSourceSelectionInput{
			Kind: value.Kind, DatabaseID: value.DatabaseID, VolumeID: value.VolumeID, AllSelected: value.AllSelected,
			SelectedTableIDs: value.SelectedTableIDs, ExcludedTableIDs: value.ExcludedTableIDs,
			SelectedFileIDs: value.SelectedFileIDs, ExcludedFileIDs: value.ExcludedFileIDs,
			Filters: sdk.SemanticModelSourceSelectionFilters{TableName: value.Filters.TableName, FileName: value.Filters.FileName, FileExt: value.Filters.FileExt},
		})
	}
	return result
}

func knowledgeLifecycleQuery(values map[string][]string) []sdk.QueryOption {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]sdk.QueryOption, 0, len(keys))
	for _, key := range keys {
		result = append(result, sdk.WithQueryValues(key, values[key]...))
	}
	return result
}

func knowledgeLifecycleQueryArgs(args []string, values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		for _, value := range values[key] {
			args = append(args, "--query-param", key+"="+value)
		}
	}
	return args
}

func knowledgeLifecycleJSON(raw json.RawMessage) any {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		panic(err)
	}
	return value
}

func knowledgeLifecycleInt64CSV(values []int64) string {
	items := make([]string, 0, len(values))
	for _, value := range values {
		items = append(items, strconv.FormatInt(value, 10))
	}
	return strings.Join(items, ",")
}

func knowledgeLifecycleUpload(t *testing.T, input knowledgeLifecycleInput) ([]byte, string) {
	t.Helper()
	content, err := base64.StdEncoding.DecodeString(input.ContentBase64)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "upload.bin")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return content, path
}

func callKnowledgeLifecycleGo(t *testing.T, client *sdk.Client, scenario knowledgeLifecycleContract) (any, []string, error) {
	t.Helper()
	service := client.Knowledge("workspace-1")
	input := scenario.Input
	args := []string{scenario.Operation}
	var result any
	var err error
	switch scenario.Operation {
	case "list":
		result, err = service.List(t.Context(), knowledgeLifecycleQuery(scenario.Query)...)
		args = knowledgeLifecycleQueryArgs(args, scenario.Query)
	case "list-tags":
		result, err = service.ListTags(t.Context(), knowledgeLifecycleQuery(scenario.Query)...)
		args = knowledgeLifecycleQueryArgs(args, scenario.Query)
	case "create-empty":
		var options []sdk.SemanticModelEmptyOption
		args = append(args, "--name", input.Name)
		if input.DescriptionPresent {
			options = append(options, sdk.WithSemanticModelEmptyDescription(input.Description))
			args = append(args, "--description", input.Description)
		}
		if input.ImageIndexEnabledPresent {
			options = append(options, sdk.WithSemanticModelEmptyImageIndexEnabled(input.ImageIndexEnabled))
			args = append(args, fmt.Sprintf("--image-index-enabled=%t", input.ImageIndexEnabled))
		}
		result, err = service.CreateEmpty(t.Context(), input.Name, options...)
	case "import-model":
		request := &sdk.SemanticModelImportRequest{}
		if unmarshalErr := protojson.Unmarshal(input.Request, request); unmarshalErr != nil {
			t.Fatal(unmarshalErr)
		}
		encoded, marshalErr := json.Marshal(request)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--json", string(encoded))
		result, err = service.ImportModel(t.Context(), request)
	case "import-models":
		requests := make([]*sdk.SemanticModelImportRequest, 0, len(input.Requests))
		for _, raw := range input.Requests {
			request := &sdk.SemanticModelImportRequest{}
			if unmarshalErr := protojson.Unmarshal(raw, request); unmarshalErr != nil {
				t.Fatal(unmarshalErr)
			}
			requests = append(requests, request)
		}
		encoded, marshalErr := json.Marshal(requests)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--json", string(encoded))
		result, err = service.ImportModels(t.Context(), requests)
	case "upload-local-file", "upload-model-local-file":
		content, path := knowledgeLifecycleUpload(t, input)
		args = append(args, "--file-path", path, "--filename", input.Filename)
		if scenario.Operation == "upload-local-file" {
			result, err = service.UploadLocalFile(t.Context(), input.Filename, bytes.NewReader(content))
		} else {
			args = append(args, "--model-id", input.ModelID)
			result, err = service.UploadModelLocalFile(t.Context(), input.ModelID, input.Filename, bytes.NewReader(content))
		}
	case "get":
		args = append(args, "--model-id", input.ModelID)
		result, err = service.Get(t.Context(), input.ModelID)
	case "update":
		var options []sdk.SemanticModelOption
		args = append(args, "--model-id", input.ModelID, "--name", input.Name)
		if input.DescriptionPresent {
			options = append(options, sdk.WithSemanticModelDescription(input.Description))
			args = append(args, "--description", input.Description)
		}
		result, err = service.Update(t.Context(), input.ModelID, input.Name, options...)
	case "delete":
		args = append(args, "--model-id", input.ModelID)
		result, err = service.Delete(t.Context(), input.ModelID)
	case "preview-source-selection-counts", "preview-model-source-selection-counts":
		selections := knowledgeLifecycleSDKSelections(input.Selections)
		encoded, marshalErr := json.Marshal(selections)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--json", string(encoded))
		if scenario.Operation == "preview-source-selection-counts" {
			result, err = service.PreviewSourceSelectionCounts(t.Context(), selections)
		} else {
			args = append(args, "--model-id", input.ModelID)
			result, err = service.PreviewModelSourceSelectionCounts(t.Context(), input.ModelID, selections)
		}
	case "list-referencing-agents":
		args = append(args, "--model-id", input.ModelID)
		if input.LimitPresent {
			args = append(args, "--limit", fmt.Sprint(input.Limit))
		}
		if input.Cursor != "" {
			args = append(args, "--cursor", input.Cursor)
		}
		result, err = service.ListReferencingAgents(t.Context(), input.ModelID, input.Limit, input.Cursor)
	case "list-source-jobs":
		args = append(args, "--model-id", input.ModelID)
		result, err = service.ListSourceJobs(t.Context(), input.ModelID)
	case "reconcile-source-jobs":
		args = append(args, "--model-id", input.ModelID)
		var options []sdk.SourceJobReconcileOption
		if input.RetryPresent {
			options = append(options, sdk.WithRetrySourceRowID(input.RetrySourceRowID))
			args = append(args, "--retry-source-row-id", input.RetrySourceRowID)
		}
		result, err = service.ReconcileSourceJobs(t.Context(), input.ModelID, options...)
	case "list-sources":
		args = append(args, "--model-id", input.ModelID)
		var options []sdk.QueryOption
		if input.IncludeDisplayStatusPresent {
			options = append(options, sdk.WithQueryParam("include_display_status", strconv.FormatBool(input.IncludeDisplayStatus)))
			args = append(args, fmt.Sprintf("--include-display-status=%t", input.IncludeDisplayStatus))
		}
		result, err = service.ListSources(t.Context(), input.ModelID, options...)
	case "list-sources-page":
		args = append(args, "--model-id", input.ModelID, "--page", fmt.Sprint(input.Page), "--page-size", fmt.Sprint(input.PageSize))
		var options []sdk.QueryOption
		if input.IncludeDisplayStatusPresent {
			options = append(options, sdk.WithQueryParam("include_display_status", strconv.FormatBool(input.IncludeDisplayStatus)))
			args = append(args, fmt.Sprintf("--include-display-status=%t", input.IncludeDisplayStatus))
		}
		result, err = service.ListSourcesPage(t.Context(), input.ModelID, input.Page, input.PageSize, options...)
	case "check-source-existence":
		args = append(args, "--model-id", input.ModelID, "--file-ids", strings.Join(input.FileIDs, ","), "--table-ids", knowledgeLifecycleInt64CSV(input.TableIDs))
		result, err = service.CheckSourceExistence(t.Context(), input.ModelID, input.FileIDs, input.TableIDs)
	case "append-sources":
		sources := knowledgeLifecycleSDKSources(input.Sources)
		encoded, marshalErr := json.Marshal(sources)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--model-id", input.ModelID, "--json", string(encoded))
		var options []sdk.SemanticModelAppendSourcesOption
		if input.SelectionsPresent {
			selections := knowledgeLifecycleSDKSelections(input.Selections)
			options = append(options, sdk.WithSemanticModelAppendSourceSelections(selections...))
			encoded, marshalErr = json.Marshal(selections)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			path := filepath.Join(t.TempDir(), "selections.json")
			if writeErr := os.WriteFile(path, encoded, 0o600); writeErr != nil {
				t.Fatal(writeErr)
			}
			args = append(args, "--source-selections-file", path)
		}
		result, err = service.AppendSources(t.Context(), input.ModelID, sources, options...)
	case "backfill-legacy-sources":
		args = append(args, "--model-id", input.ModelID)
		result, err = service.BackfillLegacySources(t.Context(), input.ModelID)
	case "search-drawings":
		request := sdk.DrawingSearchRequest{KnowledgeBaseID: input.KnowledgeBaseID, DrawingNumber: input.DrawingNumber, Query: input.QueryText, TopK: input.TopK, TopKPresent: input.TopKPresent}
		body := map[string]any{"knowledge_base_id": input.KnowledgeBaseID}
		if input.DrawingNumber != "" {
			body["drawing_number"] = input.DrawingNumber
		}
		if input.QueryText != "" {
			body["query"] = input.QueryText
		}
		if input.TopKPresent {
			body["top_k"] = input.TopK
		}
		if input.Filename != "" {
			content, path := knowledgeLifecycleUpload(t, input)
			request.QueryImageBytes = content
			request.QueryImageFilename = input.Filename
			args = append(args, "--knowledge-base-id", strconv.FormatInt(input.KnowledgeBaseID, 10), "--query", input.QueryText, "--query-image", path, "--query-image-filename", input.Filename)
			if input.TopKPresent {
				args = append(args, "--top-k", fmt.Sprint(input.TopK))
			}
		} else {
			encoded, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			args = append(args, "--json", string(encoded))
		}
		result, err = service.SearchDrawings(t.Context(), request)
	case "get-drawing-provenance":
		args = append(args, "--knowledge-base-id", strconv.FormatInt(input.KnowledgeBaseID, 10), "--source-file-id", input.SourceFileID)
		if input.PageNumberPresent {
			args = append(args, "--page-number", fmt.Sprint(input.PageNumber))
		}
		result, err = service.GetDrawingProvenance(t.Context(), input.KnowledgeBaseID, input.SourceFileID, input.PageNumber)
	case "list-entries":
		args = append(args, "--model-id", input.ModelID)
		args = knowledgeLifecycleQueryArgs(args, scenario.Query)
		result, err = service.ListEntries(t.Context(), input.ModelID, knowledgeLifecycleQuery(scenario.Query)...)
	case "create-entry", "update-entry":
		entry := knowledgeLifecycleSDKEntry(input.Entry)
		encoded, marshalErr := json.Marshal(entry)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--model-id", input.ModelID, "--json", string(encoded))
		if scenario.Operation == "create-entry" {
			result, err = service.CreateEntry(t.Context(), input.ModelID, entry)
		} else {
			args = append(args, "--entry-id", input.EntryID)
			result, err = service.UpdateEntry(t.Context(), input.ModelID, input.EntryID, entry)
		}
	case "delete-entry":
		args = append(args, "--model-id", input.ModelID, "--entry-id", input.EntryID)
		result, err = service.DeleteEntry(t.Context(), input.ModelID, input.EntryID)
	case "export":
		args = append(args, "--model-id", input.ModelID)
		result, err = service.Export(t.Context(), input.ModelID)
	case "import":
		entries := knowledgeLifecycleSDKEntries(input.Entries)
		encoded, marshalErr := json.Marshal(entries)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--model-id", input.ModelID, "--json", string(encoded))
		result, err = service.Import(t.Context(), input.ModelID, entries)
	case "delete-source":
		args = append(args, "--model-id", input.ModelID, "--source-id", input.SourceID)
		result, err = service.DeleteSource(t.Context(), input.ModelID, input.SourceID)
	case "get-source-document":
		args = append(args, "--model-id", input.ModelID, "--source-id", input.SourceID)
		args = knowledgeLifecycleQueryArgs(args, scenario.Query)
		result, err = service.GetSourceDocument(t.Context(), input.ModelID, input.SourceID, knowledgeLifecycleQuery(scenario.Query)...)
	case "update-source-governance":
		args = append(args, "--model-id", input.ModelID, "--source-id", input.SourceID)
		var options []sdk.SemanticModelSourceGovernanceOption
		if input.TagsPresent {
			options = append(options, sdk.WithSemanticModelSourceTags(input.Tags...))
			encoded, marshalErr := json.Marshal(input.Tags)
			if marshalErr != nil {
				t.Fatal(marshalErr)
			}
			args = append(args, "--tags", string(encoded))
		}
		if input.ExpiresAtPresent {
			options = append(options, sdk.WithSemanticModelSourceExpiresAt(input.ExpiresAt))
			args = append(args, "--expires-at", strconv.FormatInt(input.ExpiresAt, 10))
		}
		if input.ClearExpiresAt {
			options = append(options, sdk.WithSemanticModelSourceExpiresAtCleared())
			args = append(args, "--clear-expires-at")
		}
		if input.EnabledPresent {
			options = append(options, sdk.WithSemanticModelSourceEnabled(input.Enabled))
			args = append(args, fmt.Sprintf("--enabled=%t", input.Enabled))
		}
		if input.ForceEnabledAfterExpiryPresent {
			options = append(options, sdk.WithSemanticModelSourceForceEnabledAfterExpiry(input.ForceEnabledAfterExpiry))
			args = append(args, fmt.Sprintf("--force-enabled-after-expiry=%t", input.ForceEnabledAfterExpiry))
		}
		result, err = service.UpdateSourceGovernance(t.Context(), input.ModelID, input.SourceID, options...)
	case "create-segment":
		segment := sdk.SemanticModelSegmentCreateInput{SemanticModelSegmentBase: sdk.SemanticModelSegmentBase{IfMatch: input.IfMatch}, Level: input.Level, Content: input.Content, OCRText: input.OCRText, ImageDescription: input.ImageDescription, ImageFileID: input.ImageFileID, PageImageFileID: input.PageImageFileID, BBox: input.BBox, Metadata: input.Metadata}
		encoded, marshalErr := json.Marshal(segment)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--model-id", input.ModelID, "--source-id", input.SourceID, "--json", string(encoded))
		result, err = service.CreateSegment(t.Context(), input.ModelID, input.SourceID, segment)
	case "update-segment":
		segment := sdk.SemanticModelSegmentUpdateInput{SemanticModelSegmentBase: sdk.SemanticModelSegmentBase{IfMatch: input.IfMatch}, Content: input.Content, OCRText: input.OCRText, ImageDescription: input.ImageDescription}
		encoded, marshalErr := json.Marshal(segment)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--model-id", input.ModelID, "--source-id", input.SourceID, "--segment-id", input.SegmentID, "--json", string(encoded))
		result, err = service.UpdateSegment(t.Context(), input.ModelID, input.SourceID, input.SegmentID, segment)
	case "update-segment-enabled", "delete-segment", "reembed-segments":
		base := sdk.SemanticModelSegmentBase{IfMatch: input.IfMatch}
		encoded, marshalErr := json.Marshal(base)
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		args = append(args, "--model-id", input.ModelID, "--source-id", input.SourceID, "--json", string(encoded))
		if scenario.Operation == "reembed-segments" {
			result, err = service.ReembedSegments(t.Context(), input.ModelID, input.SourceID, base)
		} else {
			args = append(args, "--segment-id", input.SegmentID)
			if scenario.Operation == "delete-segment" {
				result, err = service.DeleteSegment(t.Context(), input.ModelID, input.SourceID, input.SegmentID, base)
			} else {
				args = append(args, fmt.Sprintf("--enabled=%t", input.Enabled))
				result, err = service.UpdateSegmentEnabled(t.Context(), input.ModelID, input.SourceID, input.SegmentID, base, input.Enabled)
			}
		}
	case "validate":
		args = append(args, "--model-id", input.ModelID)
		result, err = service.Validate(t.Context(), input.ModelID)
	default:
		t.Fatalf("unknown operation %s", scenario.Operation)
	}
	return result, args, err
}

func knowledgeLifecycleError(t *testing.T, err error, scenario knowledgeLifecycleContract) *sdk.Error {
	t.Helper()
	if scenario.Status < 400 {
		if err != nil {
			t.Fatal(err)
		}
		return nil
	}
	var productErr *sdk.Error
	if !errors.As(err, &productErr) {
		t.Fatalf("error=%T %v, want *sdk.Error", err, err)
	}
	var envelope struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	if productErr.Code != envelope.Code || productErr.Message != envelope.Msg || productErr.Status != scenario.Status || productErr.RequestID != scenario.ResponseHeaders["X-Request-ID"] {
		t.Fatalf("error=%+v, want code=%s message=%q status=%d request_id=%q", productErr, envelope.Code, envelope.Msg, scenario.Status, scenario.ResponseHeaders["X-Request-ID"])
	}
	return productErr
}

func knowledgeLifecycleResultData(t *testing.T, scenario knowledgeLifecycleContract) json.RawMessage {
	t.Helper()
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(scenario.Response, &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data
}

func knowledgeLifecycleDecodeResult(t *testing.T, raw []byte, reference any) any {
	t.Helper()
	var actual any
	switch reference.(type) {
	case proto.Message:
		actual = reference.(proto.Message).ProtoReflect().New().Interface()
	case *sdk.DrawingSearchResponse:
		actual = &sdk.DrawingSearchResponse{}
	case *sdk.DrawingProvenance:
		actual = &sdk.DrawingProvenance{}
	default:
		t.Fatalf("unsupported result type %T", reference)
	}
	if message, ok := actual.(proto.Message); ok {
		if err := protojson.Unmarshal(raw, message); err != nil {
			t.Fatal(err)
		}
	} else if err := json.Unmarshal(raw, actual); err != nil {
		t.Fatal(err)
	}
	return actual
}

func assertKnowledgeLifecycleResult(t *testing.T, scenario knowledgeLifecycleContract, actual any) {
	t.Helper()
	expected := knowledgeLifecycleDecodeResult(t, knowledgeLifecycleResultData(t, scenario), actual)
	if message, ok := actual.(proto.Message); ok {
		if !proto.Equal(message, expected.(proto.Message)) {
			t.Fatalf("result=%v, want=%v", actual, expected)
		}
		return
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("result=%#v, want=%#v", actual, expected)
	}
}

type knowledgeLifecycleObservation struct {
	body      any
	multipart *catalogUploadCLIMultipart
}

func assertKnowledgeLifecycleObservation(t *testing.T, got knowledgeLifecycleObservation, scenario knowledgeLifecycleContract) {
	t.Helper()
	if scenario.Multipart != nil {
		if got.multipart == nil || !reflect.DeepEqual(*got.multipart, *scenario.Multipart) {
			t.Fatalf("multipart=%+v, want=%+v", got.multipart, scenario.Multipart)
		}
		return
	}
	if !reflect.DeepEqual(got.body, knowledgeLifecycleJSON(scenario.Request)) {
		t.Fatalf("body=%#v, want=%#v", got.body, knowledgeLifecycleJSON(scenario.Request))
	}
}

func TestKnowledgeLifecycleSharedHTTPContracts(t *testing.T) {
	runKnowledgeSharedHTTPContracts(t, "knowledge_lifecycle_http_contract.json")
}

func TestKnowledgeGovernanceSharedHTTPContracts(t *testing.T) {
	runKnowledgeSharedHTTPContracts(t, "knowledge_governance_http_contract.json")
}

func runKnowledgeSharedHTTPContracts(t *testing.T, fixture string) {
	data, err := os.ReadFile("../testdata/product-contracts/" + fixture)
	if err != nil {
		t.Fatal(err)
	}
	var cases []knowledgeLifecycleContract
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			observations := make(chan knowledgeLifecycleObservation, 2)
			expectedQuery := url.Values{}
			for key, values := range scenario.Query {
				for _, value := range values {
					expectedQuery.Add(key, value)
				}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s query=%v headers=%v", r.Method, r.URL.RequestURI(), r.URL.Query(), r.Header)
				}
				observation := knowledgeLifecycleObservation{}
				if scenario.Multipart != nil {
					multipart, readErr := readCatalogUploadCLIRequest(r)
					if readErr != nil {
						t.Errorf("read multipart: %v", readErr)
					}
					observation.multipart = &multipart
				} else {
					raw, _ := io.ReadAll(r.Body)
					if len(raw) > 0 {
						decoder := json.NewDecoder(bytes.NewReader(raw))
						decoder.UseNumber()
						if decodeErr := decoder.Decode(&observation.body); decodeErr != nil {
							t.Errorf("decode request: %v", decodeErr)
						}
					}
				}
				observations <- observation
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(scenario.Response)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			goResult, args, goErr := callKnowledgeLifecycleGo(t, client, scenario)
			assertKnowledgeLifecycleObservation(t, <-observations, scenario)
			if knowledgeLifecycleError(t, goErr, scenario) == nil {
				assertKnowledgeLifecycleResult(t, scenario, goResult)
			}

			var cliErr error
			output := captureStdout(t, func() {
				cliErr = ExecuteSemanticModelCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			assertKnowledgeLifecycleObservation(t, <-observations, scenario)
			if knowledgeLifecycleError(t, cliErr, scenario) == nil {
				if output == "" {
					t.Fatal("CLI returned no result")
				}
				actual := knowledgeLifecycleDecodeResult(t, []byte(output), goResult)
				assertKnowledgeLifecycleResult(t, scenario, actual)
			} else if output != "" {
				t.Fatalf("CLI error output=%q", output)
			}
		})
	}
}
