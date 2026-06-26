package format

import (
	"errors"
	"reflect"
	"testing"
)

func TestCommandVars(t *testing.T) {
	tests := []struct {
		name string
		tmpl string
		want []string
	}{
		{"none", "id", nil},
		{"single", "nmap -sV <target>", []string{"target"}},
		{"multiple", "ssh <user>@<host> -p <port>", []string{"user", "host", "port"}},
		{"dedup", "curl <url> -o <url>.html", []string{"url"}},
		{"underscores", "set <wordlist_path>", []string{"wordlist_path"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Command{Command: tt.tmpl}.Vars()
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Vars(%q) = %v, want %v", tt.tmpl, got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cs      Cheatsheet
		wantErr bool
	}{
		{
			name: "valid",
			cs: Cheatsheet{Name: "recon", Commands: []Command{
				{ID: "nmap-sv", Command: "nmap -sV <target>", Risk: RiskLow},
			}},
		},
		{
			name:    "missing name",
			cs:      Cheatsheet{Commands: []Command{{ID: "a", Command: "x"}}},
			wantErr: true,
		},
		{
			name:    "no commands",
			cs:      Cheatsheet{Name: "empty"},
			wantErr: true,
		},
		{
			name: "duplicate id",
			cs: Cheatsheet{Name: "dup", Commands: []Command{
				{ID: "a", Command: "x"}, {ID: "a", Command: "y"},
			}},
			wantErr: true,
		},
		{
			name: "bad schema version",
			cs: Cheatsheet{SchemaVersion: "999", Name: "v", Commands: []Command{
				{ID: "a", Command: "x"},
			}},
			wantErr: true,
		},
		{
			name: "invalid risk",
			cs: Cheatsheet{Name: "r", Commands: []Command{
				{ID: "a", Command: "x", Risk: Risk("nuclear")},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cs.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && !errors.Is(err, ErrValidation) {
				t.Fatalf("error %v is not ErrValidation", err)
			}
		})
	}
}
