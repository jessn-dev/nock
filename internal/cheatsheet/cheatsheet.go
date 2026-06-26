// Package cheatsheet loads and validates cheatsheet files from disk into the
// format types the engine consumes.
//
// Scaffold status: this loader handles JSON (.json) using the standard library so
// the project builds and tests with zero external dependencies. YAML (.yaml/.yml)
// — the primary authoring format — is wired in during Milestone 0/1 via
// github.com/goccy/go-yaml; LoadFile returns errUnsupportedYAML until then. Keeping
// the seam here means the engine never changes when YAML lands.
package cheatsheet

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessn-dev/nock/pkg/format"
)

var errUnsupportedYAML = errors.New("cheatsheet: YAML support arrives in Milestone 0/1 (goccy/go-yaml); use .json for now")

// LoadFile reads and validates a single cheatsheet file.
func LoadFile(path string) (format.Cheatsheet, error) {
	// Loading a user-specified cheatsheet file by path is the intended behaviour;
	// the operator chooses which of their own files to load.
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is operator-supplied by design
	if err != nil {
		return format.Cheatsheet{}, fmt.Errorf("cheatsheet: read %s: %w", path, err)
	}
	var cs format.Cheatsheet
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		if err := json.Unmarshal(data, &cs); err != nil {
			return format.Cheatsheet{}, fmt.Errorf("cheatsheet: parse %s: %w", path, err)
		}
	case ".yaml", ".yml":
		return format.Cheatsheet{}, fmt.Errorf("%s: %w", path, errUnsupportedYAML)
	default:
		return format.Cheatsheet{}, fmt.Errorf("cheatsheet: unknown extension for %s", path)
	}
	if err := cs.Validate(); err != nil {
		return format.Cheatsheet{}, fmt.Errorf("cheatsheet: %s: %w", path, err)
	}
	return cs, nil
}

// LoadDir walks root and loads every cheatsheet file it finds. It returns all
// successfully loaded sheets and a joined error for any that failed, so one bad
// file does not silently drop the rest.
func LoadDir(root string) ([]format.Cheatsheet, error) {
	var sheets []format.Cheatsheet
	var errs []error
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !isCheatsheet(path) {
			return nil
		}
		cs, lerr := LoadFile(path)
		if lerr != nil {
			errs = append(errs, lerr)
			return nil
		}
		sheets = append(sheets, cs)
		return nil
	})
	if err != nil {
		errs = append(errs, fmt.Errorf("cheatsheet: walk %s: %w", root, err))
	}
	return sheets, errors.Join(errs...)
}

func isCheatsheet(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}
