package history

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestAppendRecentRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	s := New(path)

	for i, id := range []string{"first", "second", "third"} {
		err := s.Append(Entry{
			Time:     time.Unix(int64(i+1), 0),
			ID:       id,
			Template: "echo <x>",
			Vars:     map[string]string{"x": id},
		})
		if err != nil {
			t.Fatalf("Append %q: %v", id, err)
		}
	}

	got, err := s.Recent(10)
	if err != nil {
		t.Fatalf("Recent: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3", len(got))
	}
	// Newest first.
	if got[0].ID != "third" || got[2].ID != "first" {
		t.Fatalf("order = %s..%s, want third..first (newest first)", got[0].ID, got[2].ID)
	}
	// Bindings survive the round trip and are stored, not the resolved string.
	if got[0].Vars["x"] != "third" || got[0].Template != "echo <x>" {
		t.Fatalf("entry not preserved: %+v", got[0])
	}
}

func TestRecentCapsAtN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "h.jsonl")
	s := New(path)
	for i := 0; i < 5; i++ {
		if err := s.Append(Entry{ID: "c", Template: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Recent(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("Recent(2) returned %d, want 2", len(got))
	}
}

func TestDisabledStoreIsNoOp(t *testing.T) {
	s := New("") // empty path = disabled
	if err := s.Append(Entry{ID: "x"}); err != nil {
		t.Fatalf("Append on disabled store must be a no-op, got %v", err)
	}
	got, err := s.Recent(10)
	if err != nil || got != nil {
		t.Fatalf("Recent on disabled store = (%v, %v), want (nil, nil)", got, err)
	}
}

func TestRecentMissingFileIsEmpty(t *testing.T) {
	s := New(filepath.Join(t.TempDir(), "does-not-exist.jsonl"))
	got, err := s.Recent(10)
	if err != nil {
		t.Fatalf("missing file must not error, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("missing file should yield no entries, got %d", len(got))
	}
}

func TestRecentSkipsCorruptLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "h.jsonl")
	good := `{"id":"ok","template":"x"}`
	// A garbage line between two good ones must be skipped, not fatal.
	content := good + "\n" + "{not json}\n" + good + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := New(path).Recent(10)
	if err != nil {
		t.Fatalf("corrupt line must not fail recall, got %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d valid entries, want 2 (corrupt line skipped)", len(got))
	}
}

func TestDefaultPathHonoursEnv(t *testing.T) {
	t.Setenv("NOCK_HISTORY", "off")
	p, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if p != "" {
		t.Fatalf("NOCK_HISTORY=off must disable history (empty path), got %q", p)
	}

	t.Setenv("NOCK_HISTORY", "/custom/h.jsonl")
	p, err = DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if p != "/custom/h.jsonl" {
		t.Fatalf("explicit NOCK_HISTORY path not honoured, got %q", p)
	}
}

func TestDefaultPathOffIsCaseInsensitive(t *testing.T) {
	for _, v := range []string{"off", "OFF", "Off", "oFf", "None", "FALSE"} {
		t.Setenv("NOCK_HISTORY", v)
		p, err := DefaultPath()
		if err != nil {
			t.Fatalf("%q: %v", v, err)
		}
		if p != "" {
			t.Fatalf("NOCK_HISTORY=%q must disable history, got %q", v, p)
		}
	}
}

// TestFileIsOwnerOnly checks the secrets-at-rest guarantee on Unix, where mode
// bits are enforced. Windows uses an ACL instead (covered by manual review and
// the openAppend implementation); os.FileMode bits are not meaningful there.
func TestFileIsOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX mode bits not meaningful on Windows (owner-only DACL used instead)")
	}
	path := filepath.Join(t.TempDir(), "h.jsonl")
	if err := New(path).Append(Entry{ID: "x", Template: "y"}); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("history file mode = %o, want 0600 (secrets at rest)", perm)
	}
}
