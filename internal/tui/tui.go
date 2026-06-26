// Package tui is nock's terminal frontend — the default mode and the product's
// face. It is a thin shell over the engine: it renders search results, prompts for
// missing variables, and shows the fully resolved command before the operator
// fires it. It contains no search or resolution logic of its own.
//
// Security — show-before-fire (ADR 009): the operator always sees the exact,
// fully resolved command before anything leaves nock. What is displayed is exactly
// what is emitted; there is no hidden expansion and nock never auto-runs a command.
// On confirmation the resolved string is printed to stdout for the operator's shell
// to pick up — running it remains the operator's explicit act.
package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jessn-dev/nock/internal/engine"
	"github.com/jessn-dev/nock/pkg/format"
)

// maxVisible bounds how many search results are drawn at once; the list scrolls
// around the cursor beyond that.
const maxVisible = 12

// stage is which screen of the search -> fill -> confirm flow is active.
type stage int

const (
	stageSearch  stage = iota // typing a query, navigating results
	stageVars                 // prompting for the selected command's missing variables
	stageConfirm              // showing the fully resolved command before firing
)

// Run starts the interactive TUI against the given engine. It blocks until the
// operator quits. If a command was confirmed, its fully resolved form is printed
// to stdout after the UI tears down, so a shell can capture it.
func Run(ctx context.Context, e *engine.Engine) error {
	m := newModel(e)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithContext(ctx))
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	if fm, ok := final.(model); ok && fm.emit != "" {
		// Printed to the normal screen after the alt-screen UI is gone: this is the
		// command the operator confirmed, ready to run. nock does not run it.
		if _, werr := fmt.Fprintln(os.Stdout, fm.emit); werr != nil {
			return fmt.Errorf("tui: emit command: %w", werr)
		}
	}
	return nil
}

// model is the Bubble Tea state for the whole flow.
type model struct {
	engine *engine.Engine

	stage   stage
	query   textinput.Model
	varIn   textinput.Model
	results []engine.Result
	cursor  int // index into results
	offset  int // first visible result, for scrolling

	selected format.Command // the command being filled / confirmed
	missing  []string       // its still-unbound variables, in prompt order
	varIdx   int            // which missing variable is being prompted

	resolved string // fully resolved command shown on the confirm screen
	emit     string // set when the operator confirms; printed after teardown
	status   string // transient hint / error line

	width, height int
}

func newModel(e *engine.Engine) model {
	q := textinput.New()
	q.Placeholder = "search commands…"
	q.Prompt = "› "
	q.Focus()

	v := textinput.New()
	v.Prompt = "› "

	m := model{engine: e, query: q, varIn: v}
	m.results = e.Search("") // empty query lists everything in corpus order
	return m
}

// Init satisfies tea.Model; it starts the text-input cursor blink.
func (m model) Init() tea.Cmd { return textinput.Blink }

// Update routes messages by stage. Engine calls are the only business logic; all
// search and resolution stays behind the engine API.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.stage {
		case stageSearch:
			return m.updateSearch(msg)
		case stageVars:
			return m.updateVars(msg)
		case stageConfirm:
			return m.updateConfirm(msg)
		}
	}
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		return m, tea.Quit
	case "up", "ctrl+k":
		m.moveCursor(-1)
		return m, nil
	case "down", "ctrl+j":
		m.moveCursor(1)
		return m, nil
	case "enter":
		if len(m.results) == 0 {
			return m, nil
		}
		m.selected = m.results[m.cursor].Command
		return m.beginFill()
	}
	// Any other key edits the query; re-search and reset the cursor.
	var cmd tea.Cmd
	m.query, cmd = m.query.Update(msg)
	m.results = m.engine.Search(m.query.Value())
	m.cursor, m.offset = 0, 0
	return m, cmd
}

func (m model) updateVars(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.backToSearch(), nil
	case "enter":
		val := strings.TrimSpace(m.varIn.Value())
		if val == "" {
			m.status = "value cannot be empty — an unset variable would leave a hole"
			return m, nil
		}
		// Set into the engine's own store so Resolve sees it (set once, fill everywhere).
		m.engine.Vars().Set(m.missing[m.varIdx], val)
		m.varIdx++
		if m.varIdx < len(m.missing) {
			m.primeVarInput()
			return m, nil
		}
		return m.resolveAndConfirm()
	}
	var cmd tea.Cmd
	m.varIn, cmd = m.varIn.Update(msg)
	return m, cmd
}

func (m model) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		return m.backToSearch(), nil
	case "enter":
		// Show-before-fire: only here, on an explicit keypress against the displayed
		// command, does anything leave nock — and only to stdout, never executed.
		m.emit = m.resolved
		return m, tea.Quit
	}
	return m, nil
}

// beginFill computes the selected command's missing variables and either prompts
// for them or, if none are missing, jumps straight to the confirm screen.
func (m model) beginFill() (tea.Model, tea.Cmd) {
	m.missing = m.engine.Vars().Missing(m.selected.Command)
	m.varIdx = 0
	m.status = ""
	if len(m.missing) == 0 {
		return m.resolveAndConfirm()
	}
	m.stage = stageVars
	m.primeVarInput()
	return m, textinput.Blink
}

// primeVarInput focuses a fresh input for the current missing variable.
func (m *model) primeVarInput() {
	m.varIn.Reset()
	m.varIn.Placeholder = "<" + m.missing[m.varIdx] + ">"
	m.varIn.Focus()
}

// resolveAndConfirm fills the command via the engine and moves to the confirm
// screen. A resolution error (should not happen once all vars are set) is surfaced
// rather than emitting a half-formed command.
func (m model) resolveAndConfirm() (tea.Model, tea.Cmd) {
	out, err := m.engine.Resolve(m.selected)
	if err != nil {
		m.status = err.Error()
		m.stage = stageSearch
		return m, nil
	}
	m.resolved = out
	m.stage = stageConfirm
	return m, nil
}

func (m model) backToSearch() model {
	m.stage = stageSearch
	m.status = ""
	m.query.Focus()
	return m
}

func (m *model) moveCursor(delta int) {
	if len(m.results) == 0 {
		return
	}
	m.cursor += delta
	switch {
	case m.cursor < 0:
		m.cursor = 0
	case m.cursor >= len(m.results):
		m.cursor = len(m.results) - 1
	}
	switch {
	case m.cursor < m.offset:
		m.offset = m.cursor
	case m.cursor >= m.offset+maxVisible:
		m.offset = m.cursor - maxVisible + 1
	}
}

// --- view ---

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	sheetStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	resolvedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// View satisfies tea.Model; it renders the screen for the active stage.
func (m model) View() string {
	switch m.stage {
	case stageVars:
		return m.viewVars()
	case stageConfirm:
		return m.viewConfirm()
	default:
		return m.viewSearch()
	}
}

func (m model) viewSearch() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("nock") + dimStyle.Render(" — search") + "\n")
	b.WriteString(m.query.View() + "\n\n")

	if len(m.results) == 0 {
		b.WriteString(dimStyle.Render("  no matches") + "\n")
	}
	end := m.offset + maxVisible
	if end > len(m.results) {
		end = len(m.results)
	}
	for i := m.offset; i < end; i++ {
		r := m.results[i]
		line := fmt.Sprintf("%s  %s", r.Command.ID, dimStyle.Render(r.Command.Command))
		prefix := "  "
		if i == m.cursor {
			prefix = cursorStyle.Render("› ")
			line = selectedStyle.Render(r.Command.ID) + "  " + dimStyle.Render(r.Command.Command)
		}
		b.WriteString(prefix + sheetStyle.Render("["+r.Sheet+"] ") + line + "\n")
	}
	if m.status != "" {
		b.WriteString("\n" + statusStyle.Render(m.status) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render("type to filter · ↑/↓ move · enter select · esc quit"))
	return b.String()
}

func (m model) viewVars() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("nock") + dimStyle.Render(" — fill variables") + "\n\n")
	b.WriteString("  " + selectedStyle.Render(m.selected.ID) + "\n")
	b.WriteString("  " + dimStyle.Render(m.selected.Command) + "\n\n")
	b.WriteString(fmt.Sprintf("  variable %d of %d: %s\n",
		m.varIdx+1, len(m.missing), titleStyle.Render("<"+m.missing[m.varIdx]+">")))
	b.WriteString("  " + m.varIn.View() + "\n")
	if m.status != "" {
		b.WriteString("\n" + statusStyle.Render("  "+m.status) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render("enter set · esc back · ctrl+c quit"))
	return b.String()
}

func (m model) viewConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("nock") + dimStyle.Render(" — confirm") + "\n\n")
	b.WriteString("  " + selectedStyle.Render(m.selected.ID))
	if r := riskLabel(m.selected.Risk); r != "" {
		b.WriteString("  " + r)
	}
	if m.selected.RequiresAuth {
		b.WriteString("  " + statusStyle.Render("requires-auth"))
	}
	b.WriteString("\n\n")
	b.WriteString("  " + resolvedStyle.Render(m.resolved) + "\n\n")
	b.WriteString(dimStyle.Render("  nock will print this command to stdout. It does not run it —\n"))
	b.WriteString(dimStyle.Render("  you do. Review it before you fire.") + "\n")
	b.WriteString("\n" + helpStyle.Render("enter fire (print) · esc back · ctrl+c quit"))
	return b.String()
}

// riskLabel renders a coloured badge for a command's risk level, or "" if unset.
func riskLabel(r format.Risk) string {
	var color string
	switch r {
	case format.RiskInfo:
		color = "8"
	case format.RiskLow:
		color = "10"
	case format.RiskMedium:
		color = "11"
	case format.RiskHigh:
		color = "9"
	case format.RiskUnspecified:
		return ""
	default:
		return ""
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("risk:" + string(r))
}
