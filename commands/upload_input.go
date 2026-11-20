package commands

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
)

const defaultMaxUploadBytes int64 = 64 << 20

// LocalInputError describes a client-side failure, never a server business error.
// Stable codes and safe display messages must not expose raw OS/provider causes.
type LocalInputError struct{ Code string }

var localInputMessages = map[string]string{
	"CLI_SCHEMA_METADATA_INVALID":       "the bundled input description is invalid",
	"CLI_SCHEMA_VALIDATION_UNAVAILABLE": "this JSON string is decoded by the server; offline validation is not available",
	"CLI_CONNECTOR_CSV_INCOMPLETE":      "provide csv-separator, csv-delimiter and csv-escape together",
	"CLI_FILE_INPUT_CONFLICT":           "choose file-path or legacy JSON input, not both",
	"CLI_FILE_LIMIT_INVALID":            "max-upload-bytes must be positive",
	"CLI_FILE_NOT_REGULAR":              "the selected path must refer to a regular file",
	"CLI_FILE_UNREADABLE":               "the selected file could not be opened or read",
	"CLI_FILE_TOO_LARGE":                "the selected files exceed max-upload-bytes",
	"CLI_FILE_PATH_REPEATED":            "this command accepts exactly one file-path",
}

func (e *LocalInputError) Error() string { return e.Code + ": " + localInputMessages[e.Code] }

type uploadFlags struct {
	paths    []string
	maxBytes int64
	flags    *flag.FlagSet
}

func addUploadFlags(fs *flag.FlagSet) *uploadFlags {
	u := &uploadFlags{flags: fs}
	fs.Func("file-path", "Local file path (repeat only for connector uploads)", func(path string) error {
		u.paths = append(u.paths, path)
		return nil
	})
	fs.Int64Var(&u.maxBytes, "max-upload-bytes", defaultMaxUploadBytes, "Total byte limit for file-path input")
	return u
}

type uploadSource struct {
	Name   string
	Reader io.Reader
	file   *os.File
}

func (s *uploadSource) Close() {
	if s.file != nil {
		_ = s.file.Close()
	}
}

// open keeps the legacy byte-input contract unchanged. The new path mode has
// explicit bounds and filename inference only for a user-selected local path.
func (u *uploadFlags) open(filename, inline, legacyFile string) (*uploadSource, error) {
	if len(u.paths) == 0 {
		raw, err := loadJSONBytes(inline, legacyFile)
		if err != nil {
			return nil, err
		}
		return &uploadSource{Name: filename, Reader: bytes.NewReader(raw)}, nil
	}
	if len(u.paths) != 1 {
		return nil, &LocalInputError{Code: "CLI_FILE_PATH_REPEATED"}
	}
	legacyProvided := inline != "" || legacyFile != ""
	if u.flags != nil {
		u.flags.Visit(func(f *flag.Flag) {
			if f.Name == "json" || f.Name == "json-file" {
				legacyProvided = true
			}
		})
	}
	if legacyProvided {
		return nil, &LocalInputError{Code: "CLI_FILE_INPUT_CONFLICT"}
	}
	if u.maxBytes <= 0 {
		return nil, &LocalInputError{Code: "CLI_FILE_LIMIT_INVALID"}
	}
	remaining := u.maxBytes
	source, err := openUploadFile(u.paths[0], &remaining)
	if err != nil {
		return nil, err
	}
	if filename != "" {
		source.Name = filename
	}
	return source, nil
}

func (u *uploadFlags) openFiles() ([]*uploadSource, error) {
	if u.maxBytes <= 0 {
		return nil, &LocalInputError{Code: "CLI_FILE_LIMIT_INVALID"}
	}
	remaining := u.maxBytes
	sources := make([]*uploadSource, 0, len(u.paths))
	for _, path := range u.paths {
		source, err := openUploadFile(path, &remaining)
		if err != nil {
			closeUploadSources(sources)
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

func closeUploadSources(sources []*uploadSource) {
	for _, source := range sources {
		source.Close()
	}
}

// Raw archive endpoints do not construct multipart before sending. Buffer only
// the new bounded path mode so read/growth failures still precede the request.
func (s *uploadSource) bufferLocal() error {
	if s.file == nil {
		return nil
	}
	data, err := io.ReadAll(s.Reader)
	if err != nil {
		return err
	}
	s.Reader = bytes.NewReader(data)
	return nil
}

func openUploadFile(path string, remaining *int64) (*uploadSource, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, &LocalInputError{Code: "CLI_FILE_UNREADABLE"}
	}
	if !info.Mode().IsRegular() {
		return nil, &LocalInputError{Code: "CLI_FILE_NOT_REGULAR"}
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, &LocalInputError{Code: "CLI_FILE_UNREADABLE"}
	}
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, &LocalInputError{Code: "CLI_FILE_NOT_REGULAR"}
	}
	if info.Size() > *remaining {
		_ = file.Close()
		return nil, &LocalInputError{Code: "CLI_FILE_TOO_LARGE"}
	}
	// Each file is bounded to its validated size, so growing it cannot defeat the
	// aggregate budget between validation and multipart construction.
	*remaining -= info.Size()
	return &uploadSource{Name: filepath.Base(path), file: file, Reader: &uploadReader{file: file, remaining: info.Size()}}, nil
}

type uploadReader struct {
	file      io.Reader
	remaining int64
}

func (r *uploadReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.remaining == 0 {
		var probe [1]byte
		n, err := r.file.Read(probe[:])
		if n != 0 {
			return 0, &LocalInputError{Code: "CLI_FILE_TOO_LARGE"}
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, &LocalInputError{Code: "CLI_FILE_UNREADABLE"}
		}
		return 0, err
	}
	n, err := r.file.Read(p[:min(int64(len(p)), r.remaining)])
	r.remaining -= int64(n)
	if err != nil && !errors.Is(err, io.EOF) {
		return n, &LocalInputError{Code: "CLI_FILE_UNREADABLE"}
	}
	if errors.Is(err, io.EOF) && r.remaining != 0 {
		return n, &LocalInputError{Code: "CLI_FILE_UNREADABLE"}
	}
	return n, err
}
