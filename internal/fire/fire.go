// Package fire delivers a confirmed, fully resolved command out of nock to the
// operator's chosen target. It is the concrete edge of show-before-fire (ADR 009/
// 015): nothing here resolves or expands anything — the string passed in is
// exactly what the confirm screen displayed.
//
// nock never executes the command itself:
//   - stdout prints it for the operator's shell to capture (the default; ADR 015).
//   - tmux prefills it into the current pane's command line via `send-keys -l`
//     with NO trailing newline, so the operator still presses Enter to run it.
//
// Both are explicit, operator-driven acts. Displayed == emitted; no auto-run.
package fire

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Target is where a confirmed command is delivered.
type Target string

const (
	// Stdout prints the command to standard output (default).
	Stdout Target = "stdout"
	// Tmux prefills the command into the current tmux pane without running it.
	Tmux Target = "tmux"
)

// ErrUnknownTarget is returned by Parse for an unrecognised target name.
var ErrUnknownTarget = errors.New("fire: unknown target")

// Parse converts a flag value to a Target. Empty string defaults to Stdout.
func Parse(s string) (Target, error) {
	switch Target(s) {
	case "", Stdout:
		return Stdout, nil
	case Tmux:
		return Tmux, nil
	default:
		return "", fmt.Errorf("%w %q (want stdout|tmux)", ErrUnknownTarget, s)
	}
}

// Available reports whether the target can be used in the current environment.
// Tmux requires running inside a tmux session ($TMUX set); Stdout is always
// available. The TUI uses this to avoid offering a tmux fire that would fail.
func (t Target) Available() bool {
	switch t {
	case Tmux:
		return os.Getenv("TMUX") != ""
	default:
		return true
	}
}

// Emit delivers cmd to the target. Stdout writes to the package's Stdout writer
// (a trailing newline, so a shell reads a complete line). Tmux shells out to
// `tmux send-keys -l` — literal, no Enter — leaving cmd on the pane's prompt for
// the operator to fire.
func Emit(t Target, cmd string) error {
	switch t {
	case Tmux:
		if !t.Available() {
			return errors.New("fire: not inside a tmux session ($TMUX unset)")
		}
		// -l sends the argument literally; the absence of a following Enter is
		// what makes this a prefill, not an auto-run.
		c := exec.Command("tmux", "send-keys", "-l", cmd)
		if out, err := c.CombinedOutput(); err != nil {
			return fmt.Errorf("fire: tmux send-keys: %w: %s", err, out)
		}
		return nil
	default:
		if _, err := fmt.Fprintln(Out, cmd); err != nil {
			return fmt.Errorf("fire: stdout: %w", err)
		}
		return nil
	}
}

// Out is the writer the Stdout target prints to. It is a package variable so
// tests can capture emitted commands without a real terminal.
var Out io.Writer = os.Stdout
