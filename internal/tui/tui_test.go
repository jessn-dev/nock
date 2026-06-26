package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jessn-dev/nock/internal/engine"
	"github.com/jessn-dev/nock/internal/vars"
	"github.com/jessn-dev/nock/pkg/format"
)

func testEngine() *engine.Engine {
	sheet := format.Cheatsheet{Name: "t", Commands: []format.Command{
		{ID: "needs-var", Command: "echo <host>", Risk: format.RiskLow},
		{ID: "novar", Command: "id -a", Risk: format.RiskInfo},
	}}
	return engine.New([]format.Cheatsheet{sheet}, vars.New())
}

// key builds a tea.KeyMsg the model's Update switch will recognise via String().
func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

// send applies one key and returns the updated model plus any command.
func send(t *testing.T, m model, k string) (model, tea.Cmd) {
	t.Helper()
	next, cmd := m.Update(key(k))
	nm, ok := next.(model)
	if !ok {
		t.Fatalf("Update returned %T, want model", next)
	}
	return nm, cmd
}

// typeRunes feeds a string one rune at a time (as the terminal would).
func typeRunes(t *testing.T, m model, s string) model {
	t.Helper()
	for _, r := range s {
		m, _ = send(t, m, string(r))
	}
	return m
}

// selectByID moves the search cursor onto the command with the given id. The
// empty-query result list is in corpus order, so this is deterministic.
func selectByID(t *testing.T, m model, id string) model {
	t.Helper()
	for i := 0; i < len(m.results); i++ {
		if m.results[m.cursor].Command.ID == id {
			return m
		}
		m, _ = send(t, m, "down")
	}
	t.Fatalf("command %q not found in results", id)
	return m
}

func TestFlowFillVarThenConfirmEmits(t *testing.T) {
	m := newModel(testEngine())
	m = selectByID(t, m, "needs-var")
	m, _ = send(t, m, "enter")
	if m.stage != stageVars {
		t.Fatalf("stage = %d, want stageVars", m.stage)
	}
	if len(m.missing) != 1 || m.missing[0] != "host" {
		t.Fatalf("missing = %v, want [host]", m.missing)
	}
	// Fill the variable and confirm.
	m = typeRunes(t, m, "10.0.0.5")
	m, _ = send(t, m, "enter")
	if m.stage != stageConfirm {
		t.Fatalf("stage = %d, want stageConfirm", m.stage)
	}
	if m.resolved != "echo 10.0.0.5" {
		t.Fatalf("resolved = %q, want %q", m.resolved, "echo 10.0.0.5")
	}
	// Confirm fires: emit set, Quit issued. nothing emitted before this point.
	if m.emit != "" {
		t.Fatal("emit set before confirmation — show-before-fire violated")
	}
	final, cmd := send(t, m, "enter")
	if final.emit != "echo 10.0.0.5" {
		t.Fatalf("emit = %q, want %q", final.emit, "echo 10.0.0.5")
	}
	if cmd == nil {
		t.Fatal("expected tea.Quit command on confirm")
	}
}

func TestFlowNoVarsGoesStraightToConfirm(t *testing.T) {
	m := newModel(testEngine())
	m = selectByID(t, m, "novar")
	m, _ = send(t, m, "enter")
	if m.stage != stageConfirm {
		t.Fatalf("stage = %d, want stageConfirm (no vars)", m.stage)
	}
	if m.resolved != "id -a" {
		t.Fatalf("resolved = %q, want %q", m.resolved, "id -a")
	}
}

func TestEmptyVarRejected(t *testing.T) {
	m := newModel(testEngine())
	m = selectByID(t, m, "needs-var")
	m, _ = send(t, m, "enter")
	// Press enter with no value: must stay on vars with a hint, not advance.
	m, _ = send(t, m, "enter")
	if m.stage != stageVars {
		t.Fatalf("stage = %d, want stageVars (empty value must not advance)", m.stage)
	}
	if m.status == "" {
		t.Fatal("expected a status hint for empty variable")
	}
}

func TestEscFromConfirmReturnsToSearch(t *testing.T) {
	m := newModel(testEngine())
	m = selectByID(t, m, "novar")
	m, _ = send(t, m, "enter") // -> confirm
	m, _ = send(t, m, "esc")   // -> back to search
	if m.stage != stageSearch {
		t.Fatalf("stage = %d, want stageSearch after esc", m.stage)
	}
	if m.emit != "" {
		t.Fatal("esc must not emit")
	}
}

func TestConfirmViewShowsResolvedCommand(t *testing.T) {
	m := newModel(testEngine())
	m = selectByID(t, m, "novar")
	m, _ = send(t, m, "enter")
	if !strings.Contains(m.View(), "id -a") {
		t.Fatal("confirm view must display the resolved command (show-before-fire)")
	}
}
