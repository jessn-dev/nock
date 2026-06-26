package cheatsheet

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jessn-dev/nock/pkg/format"
)

const validJSON = `{
  "schema_version": "1",
  "name": "json-sheet",
  "commands": [{"id": "a", "name": "alpha", "command": "nmap <target>", "risk": "low"}]
}`

const validYAML = `schema_version: "1"
name: yaml-sheet
commands:
  - id: a
    name: alpha
    command: nmap <target>
    risk: low
`

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestLoadFileCodecs(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name string
		file string
		body string
	}{
		{"json", "sheet.json", validJSON},
		{"yaml", "sheet.yaml", validYAML},
		{"yml", "sheet.yml", validYAML},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := write(t, dir, tt.file, tt.body)
			cs, err := LoadFile(path)
			if err != nil {
				t.Fatalf("LoadFile(%s): %v", tt.file, err)
			}
			if len(cs.Commands) != 1 || cs.Commands[0].Command != "nmap <target>" {
				t.Fatalf("parsed wrong content: %+v", cs)
			}
		})
	}
}

func TestLoadFileErrors(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		file    string
		body    string
		wantVal bool // expect a format.ErrValidation
	}{
		{"unknown extension", "sheet.txt", validJSON, false},
		{"malformed json", "bad.json", "{not json", false},
		{"malformed yaml", "bad.yaml", "name: [unterminated", false},
		{"validation: duplicate id", "dup.json", `{"name":"d","commands":[{"id":"a","command":"x"},{"id":"a","command":"y"}]}`, true},
		{"validation: no commands", "empty.yaml", "name: empty\n", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := write(t, dir, tt.file, tt.body)
			_, err := LoadFile(path)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tt.wantVal && !errors.Is(err, format.ErrValidation) {
				t.Fatalf("error %v is not ErrValidation", err)
			}
		})
	}
}

func TestLoadFileMissing(t *testing.T) {
	if _, err := LoadFile(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadDirSkipsBadKeepsGood(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "good.json", validJSON)
	write(t, dir, "good.yaml", validYAML)
	write(t, dir, "bad.json", "{not json")
	write(t, dir, "ignored.txt", "not a cheatsheet") // non-cheatsheet ext: skipped silently

	sheets, err := LoadDir(dir)
	if err == nil {
		t.Fatal("expected joined error reporting the bad file")
	}
	if len(sheets) != 2 {
		t.Fatalf("loaded %d sheets, want 2 good ones despite the bad file", len(sheets))
	}
}

func TestLoadDirEmpty(t *testing.T) {
	sheets, err := LoadDir(t.TempDir())
	if err != nil {
		t.Fatalf("empty dir should not error: %v", err)
	}
	if len(sheets) != 0 {
		t.Fatalf("expected no sheets, got %d", len(sheets))
	}
}
