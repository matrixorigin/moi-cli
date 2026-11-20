package commands

import (
	"bytes"
	"encoding/base64"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func uploadFixture(t *testing.T, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func uploadContext(t *testing.T, handler http.HandlerFunc) *Context {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-upload-pat"), sdk.WithWorkspaceCurrentRoleID("ws-test", "17"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	return NewContext(client, "ws-test", false, true, true)
}

var uploadCommands = []struct {
	name, path string
	run        func(*Context, []string) error
	extra      []string
	raw        bool
}{
	{"catalog", "/newmoi/catalog/file/upload", runCatalogServiceUpload, nil, false},
	{"private", "/newmoi/workspaces/ws-test/catalog-files/uploads", runCatalogServiceUploadPrivate, nil, false},
	{"knowledge", "/newmoi/semantic-models/local-files/upload", runKnowledgeServiceUploadLocalFile, nil, false},
	{"knowledge-model", "/newmoi/semantic-models/101/local-files/upload", runKnowledgeServiceUploadModelLocalFile, []string{"--model-id", "101"}, false},
	{"agent-file", "/newmoi/workspaces/ws-test/files", runAgentServiceUploadWorkspaceFile, nil, false},
	{"inspect-skill", "/newmoi/workspaces/ws-test/skills/import/inspect", runAgentServiceInspectSkillPackage, nil, false},
	{"import-skill", "/newmoi/workspaces/ws-test/skills/import", runAgentServiceImportSkillPackage, nil, false},
	{"agent-package", "/newmoi/workspaces/ws-test/agent-packages/load", runAgentServiceLoadAgentPackage, nil, true},
	{"connector", "/newmoi/connectors/upload", runConnectorServiceUpload, nil, false},
	{"connector-stage", "/newmoi/connectors/file/upload", runConnectorServiceLocalUpload, nil, false},
}

func TestUploadPathAllCommandsPreserveBytesAndIdentity(t *testing.T) {
	for _, command := range uploadCommands {
		t.Run(command.name, func(t *testing.T) {
			for _, payload := range [][]byte{nil, []byte("id,name\n1,测试\n"), {0, 255, 1, 128, 10}} {
				t.Run(string(rune(len(payload)+'A')), func(t *testing.T) {
					path := uploadFixture(t, "销售 ' $(not-a-command).bin", payload)
					var calls atomic.Int32
					ctx := uploadContext(t, func(w http.ResponseWriter, r *http.Request) {
						calls.Add(1)
						if r.URL.Path != command.path || r.Method != http.MethodPost {
							t.Errorf("route: %s %s", r.Method, r.URL.Path)
						}
						if r.Header.Get("X-Workspace-ID") != "ws-test" || r.Header.Get("X-MOI-Role-ID") != "17" || r.Header.Get("X-API-Key") != "synthetic-upload-pat" {
							t.Errorf("identity was not forwarded")
						}
						var got []byte
						if command.raw {
							if r.Header.Get("Content-Type") != "application/vnd.moi.agent+zip" {
								t.Error("archive media type changed")
							}
							got, _ = io.ReadAll(r.Body)
						} else {
							if err := r.ParseMultipartForm(1 << 20); err != nil {
								t.Error(err)
								return
							}
							defer r.MultipartForm.RemoveAll()
							files := r.MultipartForm.File["file"]
							if len(files) != 1 {
								t.Errorf("files=%v", files)
								return
							}
							if files[0].Filename != filepath.Base(path) {
								t.Errorf("filename=%q", files[0].Filename)
							}
							f, err := files[0].Open()
							if err != nil {
								t.Error(err)
								return
							}
							defer f.Close()
							got, _ = io.ReadAll(f)
						}
						if !bytes.Equal(got, payload) {
							t.Errorf("bytes=%v want=%v", got, payload)
						}
						_, _ = io.WriteString(w, `{"code":"OK","data":{}}`)
					})
					args := append([]string{"--file-path", path}, command.extra...)
					if err := command.run(ctx, args); err != nil {
						t.Fatal(err)
					}
					if calls.Load() != 1 {
						t.Fatalf("requests=%d", calls.Load())
					}
				})
			}
		})
	}
}

func TestUploadPathFailureBeforeAnyRequest(t *testing.T) {
	path := uploadFixture(t, "input.bin", []byte("abc"))
	for _, command := range uploadCommands {
		t.Run(command.name, func(t *testing.T) {
			for _, tc := range []struct {
				name string
				args []string
				code string
			}{
				{"missing", []string{"--file-path", filepath.Join(t.TempDir(), "missing")}, "CLI_FILE_UNREADABLE"},
				{"directory", []string{"--file-path", t.TempDir()}, "CLI_FILE_NOT_REGULAR"},
				{"limit", []string{"--file-path", path, "--max-upload-bytes", "2"}, "CLI_FILE_TOO_LARGE"},
				{"invalid-limit", []string{"--file-path", path, "--max-upload-bytes", "0"}, "CLI_FILE_LIMIT_INVALID"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var calls atomic.Int32
					ctx := uploadContext(t, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
					err := command.run(ctx, append(tc.args, command.extra...))
					var inputErr *LocalInputError
					if !errors.As(err, &inputErr) || inputErr.Code != tc.code {
						t.Fatalf("error=%v want=%s", err, tc.code)
					}
					if calls.Load() != 0 {
						t.Fatal("invalid file reached server")
					}
				})
			}
		})
	}
}

func TestSingleUploadConflicts(t *testing.T) {
	path := uploadFixture(t, "data", []byte("abc"))
	for _, command := range uploadCommands[:8] {
		t.Run(command.name, func(t *testing.T) {
			for _, extra := range [][]string{{"--json", ""}, {"--json", "abc"}, {"--json-file", path}, {"--file-path", path}} {
				ctx := &Context{WorkspaceID: "ws-test"} // A network call would panic.
				args := append([]string{"--file-path", path}, command.extra...)
				args = append(args, extra...)
				if err := command.run(ctx, args); err == nil {
					t.Fatal("conflicting input accepted")
				}
			}
		})
	}
}

func TestUploadLegacyByteModesAndFilename(t *testing.T) {
	payload := []byte{0, 255, 1, 'x'}
	path := uploadFixture(t, "original.bin", payload)
	for _, tc := range []struct {
		name     string
		run      func(*Context, []string) error
		args     []string
		want     []byte
		filename string
	}{
		{"legacy-file", runCatalogServiceUpload, []string{"--filename", "old.bin", "--json-file", path}, payload, "old.bin"},
		{"legacy-inline", runCatalogServiceUpload, []string{"--filename", "old.bin", "--json", "not-json"}, []byte("not-json"), "old.bin"},
		{"legacy-priority", runCatalogServiceUpload, []string{"--filename", "old.bin", "--json", "ignored", "--json-file", path}, payload, "old.bin"},
		{"legacy-bytes", runCatalogServiceUploadBytes, []string{"--filename", "old.bin", "--json", `"` + base64.StdEncoding.EncodeToString(payload) + `"`}, payload, "old.bin"},
		{"existing-file", runCatalogServiceUploadFile, []string{"--file-path", path}, payload, "original.bin"},
		{"new-name", runCatalogServiceUpload, []string{"--file-path", path, "--filename", "chosen.bin"}, payload, "chosen.bin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := uploadContext(t, func(w http.ResponseWriter, r *http.Request) {
				f, h, err := r.FormFile("file")
				if err != nil {
					t.Error(err)
					return
				}
				defer f.Close()
				defer r.MultipartForm.RemoveAll()
				got, _ := io.ReadAll(f)
				if !bytes.Equal(got, tc.want) || h.Filename != tc.filename {
					t.Errorf("file=%q %v", h.Filename, got)
				}
				_, _ = io.WriteString(w, `{"code":"OK","data":{}}`)
			})
			if err := tc.run(ctx, tc.args); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestConnectorMultipleFilesAndAllFields(t *testing.T) {
	a := uploadFixture(t, "same.csv", []byte("a"))
	b := uploadFixture(t, "same.csv", []byte("b"))
	fields := map[string]string{"name": "每月订单 / 月报", "VolumeID": "volume-1", "meta": `[]`, "file_types": `["csv"]`, "path_regex": ".*", "dedup": `{"mode":"skip"}`, "table_config": `{}`, "unzip_keep_structure": "false"}
	ctx := uploadContext(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Error(err)
			return
		}
		defer r.MultipartForm.RemoveAll()
		for name, want := range fields {
			if got := r.FormValue(name); got != want {
				t.Errorf("%s=%q want=%q", name, got, want)
			}
		}
		files := r.MultipartForm.File["file"]
		if len(files) != 2 {
			t.Error("expected two files")
			return
		}
		var got []string
		for _, h := range files {
			f, _ := h.Open()
			data, _ := io.ReadAll(f)
			f.Close()
			got = append(got, string(data))
		}
		if !reflect.DeepEqual(got, []string{"a", "b"}) {
			t.Errorf("order/content=%v", got)
		}
		_, _ = io.WriteString(w, `{"code":"OK","data":{"acceptance_status":"partially_accepted","task_id":"task-1"}}`)
	})
	if err := runConnectorServiceUpload(ctx, []string{"--file-path", a, "--file-path", b, "--max-upload-bytes", "2", "--name", "每月订单 / 月报", "--volume-id", "volume-1", "--meta", "[]", "--file-types", `["csv"]`, "--path-regex", ".*", "--dedup", `{"mode":"skip"}`, "--table-config", "{}", "--unzip-keep-structure=false"}); err != nil {
		t.Fatal(err)
	}
	if err := runConnectorServiceLocalUpload(&Context{WorkspaceID: "ws-test"}, []string{"--file-path", a, "--file-path", b, "--max-upload-bytes", "1"}); err == nil {
		t.Fatal("aggregate limit ignored")
	}
}

func TestConnectorEmptyLegacyRequestAndExplicitBooleans(t *testing.T) {
	for _, value := range []string{"", "false", "true"} {
		t.Run("keep-structure="+value, func(t *testing.T) {
			ctx := uploadContext(t, func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(1024); err != nil {
					t.Error(err)
					return
				}
				defer r.MultipartForm.RemoveAll()
				if len(r.MultipartForm.File) != 0 {
					t.Error("unexpected file in legacy empty request")
				}
				if value == "" {
					if len(r.MultipartForm.Value) != 0 {
						t.Error("unprovided options changed legacy empty request")
					}
				} else if r.FormValue("unzip_keep_structure") != value || len(r.MultipartForm.Value) != 1 {
					t.Errorf("unexpected fields: %v", r.MultipartForm.Value)
				}
				_, _ = io.WriteString(w, `{"code":"OK","data":{}}`)
			})
			var args []string
			if value != "" {
				args = []string{"--unzip-keep-structure=" + value}
			}
			if err := runConnectorServiceUpload(ctx, args); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUploadReaderDetectsFileChangesAndClosesFiles(t *testing.T) {
	for _, tc := range []struct{ name, content string }{{"grow", "abcdef"}, {"shrink", "a"}} {
		t.Run(tc.name, func(t *testing.T) {
			path := uploadFixture(t, "file", []byte("abc"))
			u := &uploadFlags{paths: []string{path}, maxBytes: 3}
			s, err := u.open("", "", "")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := io.ReadAll(s.Reader); err == nil {
				t.Fatal("file change was not detected")
			}
			s.Close()
			if _, err := s.file.Stat(); err == nil {
				t.Fatal("file still open")
			}
		})
	}
	path := uploadFixture(t, "data", []byte("abc"))
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	u := &uploadFlags{paths: []string{link}, maxBytes: 3}
	s, err := u.open("", "", "")
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	u.paths = []string{path, filepath.Join(t.TempDir(), "missing")}
	if _, err := u.openFiles(); err == nil {
		t.Fatal("invalid second file accepted")
	}
	u.paths = []string{filepath.Join(t.TempDir(), "dangling")}
	if err := os.Symlink("missing", u.paths[0]); err != nil {
		t.Fatal(err)
	}
	if _, err := u.openFiles(); err == nil {
		t.Fatal("dangling link accepted")
	}
}

func TestUploadPreservesServerErrorAndNeverRetries(t *testing.T) {
	path := uploadFixture(t, "input", []byte("abc"))
	for _, status := range []int{401, 403, 404, 413, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			ctx := uploadContext(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				w.WriteHeader(status)
				_, _ = io.WriteString(w, `{"code":"ErrSpecific","msg":"server owned message"}`)
			})
			err := runCatalogServiceUpload(ctx, []string{"--file-path", path})
			var sdkErr *sdk.Error
			if !errors.As(err, &sdkErr) || sdkErr.Code != "ErrSpecific" || sdkErr.Message != "server owned message" {
				t.Fatalf("error=%v", err)
			}
			if calls.Load() != 1 {
				t.Fatal("write was retried")
			}
		})
	}
}

func TestUploadLimitAndEmptyFileBoundaries(t *testing.T) {
	empty := uploadFixture(t, "empty", nil)
	a := uploadFixture(t, "data", []byte("abc"))
	u := &uploadFlags{paths: []string{a, empty}, maxBytes: 3}
	sources, err := u.openFiles()
	if err != nil {
		t.Fatal(err)
	}
	defer closeUploadSources(sources)
	for _, s := range sources {
		if _, err := io.ReadAll(s.Reader); err != nil {
			t.Fatal(err)
		}
	}
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	u = addUploadFlags(fs)
	if err := fs.Parse([]string{"--file-path", a}); err != nil {
		t.Fatal(err)
	}
	if u.maxBytes != 64<<20 {
		t.Fatal("unexpected default limit")
	}
	if strings.Contains((&LocalInputError{Code: "CLI_FILE_UNREADABLE"}).Error(), a) {
		t.Fatal("raw OS path exposed")
	}
}

func TestUploadBrokenResponseAndConnectionNeverSucceedOrRetry(t *testing.T) {
	path := uploadFixture(t, "input", []byte("abc"))
	for _, brokenConnection := range []bool{false, true} {
		var calls atomic.Int32
		ctx := uploadContext(t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			if brokenConnection {
				conn, _, err := w.(http.Hijacker).Hijack()
				if err != nil {
					t.Error(err)
					return
				}
				_ = conn.Close()
				return
			}
			_, _ = io.WriteString(w, "broken response")
		})
		if err := runCatalogServiceUpload(ctx, []string{"--file-path", path}); err == nil {
			t.Fatal("invalid response or dropped connection reported success")
		}
		if calls.Load() != 1 {
			t.Fatalf("write attempts=%d", calls.Load())
		}
	}
}
