package fire

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		in      string
		want    Target
		wantErr bool
	}{
		{"", Stdout, false},
		{"stdout", Stdout, false},
		{"tmux", Tmux, false},
		{"clipboard", "", true},
		{"STDOUT", "", true}, // case-sensitive on purpose
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := Parse(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) err = nil, want error", tt.in)
				}
				if !errors.Is(err, ErrUnknownTarget) {
					t.Fatalf("Parse(%q) err = %v, want ErrUnknownTarget", tt.in, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) unexpected err: %v", tt.in, err)
			}
			if got != tt.want {
				t.Fatalf("Parse(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestStdoutAlwaysAvailable(t *testing.T) {
	if !Stdout.Available() {
		t.Fatal("stdout target must always be available")
	}
}

func TestTmuxAvailabilityFollowsEnv(t *testing.T) {
	t.Setenv("TMUX", "")
	if Tmux.Available() {
		t.Fatal("tmux must be unavailable when $TMUX is empty")
	}
	t.Setenv("TMUX", "/tmp/tmux-1000/default,1234,0")
	if !Tmux.Available() {
		t.Fatal("tmux must be available when $TMUX is set")
	}
}

func TestEmitStdoutWritesLine(t *testing.T) {
	var buf bytes.Buffer
	old := Out
	Out = &buf
	defer func() { Out = old }()

	if err := Emit(Stdout, "nmap -sV 10.0.0.5"); err != nil {
		t.Fatalf("Emit: %v", err)
	}
	got := buf.String()
	if !strings.HasSuffix(got, "\n") {
		t.Fatal("stdout emit must end with a newline so a shell reads a full line")
	}
	if strings.TrimSpace(got) != "nmap -sV 10.0.0.5" {
		t.Fatalf("emitted %q, want the command verbatim", got)
	}
}

func TestEmitTmuxWithoutSessionErrors(t *testing.T) {
	t.Setenv("TMUX", "")
	if err := Emit(Tmux, "id"); err == nil {
		t.Fatal("emitting to tmux outside a session must error, not silently drop")
	}
}
