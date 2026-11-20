package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// printJSON marshals v as indented JSON to stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// printJSONOrTable prints JSON when -o is set, otherwise runs table printer.
func printJSONOrTable(ctx *Context, v any, table func() error) error {
	if ctx != nil && ctx.OutputJSON {
		return printJSON(v)
	}
	return table()
}

func printKV(key, value string) {
	fmt.Printf("%-16s %s\n", key+":", value)
}

// writeCommandResultFile publishes a complete private file in one filesystem step.
// Immutable knowledge receipts use link; refreshable authoring exports use rename.
func writeCommandResultFile(path string, data []byte, replace bool) (resultErr error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".moi-result-*")
	if err != nil {
		return err
	}
	closed, renamed := false, false
	defer func() {
		if !closed {
			if err := file.Close(); err != nil {
				resultErr = errors.Join(resultErr, err)
			}
		}
		if !renamed {
			if err := os.Remove(file.Name()); err != nil {
				resultErr = errors.Join(resultErr, err)
			}
		}
	}()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	err = file.Close()
	closed = true
	if err != nil {
		return err
	}
	if replace {
		err = os.Rename(file.Name(), path)
		renamed = err == nil
		return err
	}
	return os.Link(file.Name(), path)
}
