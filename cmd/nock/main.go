// Command nock is the single binary for the nock command launcher. One binary,
// several modes — the engine is shared; each mode is a thin frontend over it:
//
//	nock                 interactive TUI (default)
//	nock --mcp           Model Context Protocol server over stdio (for AI agents)
//	nock search <query>  non-interactive search (scriptable; proves the engine)
//	nock resolve <id>    fill a command's variables and print it
//	nock serve           team HTTP/SSE server (Milestone 5)
//	nock import <src>    import arsenal cheatsheets as seed content (Milestone 0/1)
//	nock version         print build metadata
//
// Cheatsheets are loaded from $NOCK_CHEATSHEETS (default ./examples/cheatsheets).
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/jessn-dev/nock/internal/cheatsheet"
	"github.com/jessn-dev/nock/internal/engine"
	"github.com/jessn-dev/nock/internal/fire"
	"github.com/jessn-dev/nock/internal/history"
	"github.com/jessn-dev/nock/internal/mcp"
	"github.com/jessn-dev/nock/internal/tui"
	"github.com/jessn-dev/nock/internal/vars"
	"github.com/jessn-dev/nock/internal/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "nock:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Subcommand dispatch. Flags that toggle mode (--mcp, --version) are handled
	// before subcommands so `nock --mcp` works without a positional verb. The TUI
	// is the default mode, so a bare invocation or one that leads with a flag
	// (e.g. `nock --fire=tmux`) is routed to it.
	if len(args) == 0 || strings.HasPrefix(args[0], "-") && args[0] != "--mcp" &&
		args[0] != "--version" && args[0] != "-v" && args[0] != "--help" && args[0] != "-h" {
		return cmdTUI(ctx, args)
	}
	switch args[0] {
	case "--mcp":
		return cmdMCP(ctx)
	case "version", "--version", "-v":
		fmt.Println(version.String())
		return nil
	case "search":
		return cmdSearch(ctx, args[1:])
	case "resolve":
		return cmdResolve(ctx, args[1:])
	case "serve":
		return fmt.Errorf("serve: team server arrives in Milestone 5")
	case "import":
		return fmt.Errorf("import: arsenal importer arrives in Milestone 0/1")
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

// loadEngine builds an engine from the configured cheatsheet directory. Load
// errors are non-fatal — they are reported and whatever parsed is still served.
func loadEngine(store *vars.Store) *engine.Engine {
	dir := os.Getenv("NOCK_CHEATSHEETS")
	if dir == "" {
		dir = "examples/cheatsheets"
	}
	sheets, err := cheatsheet.LoadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "nock: warning loading cheatsheets: %v\n", err)
	}
	return engine.New(sheets, store)
}

func cmdTUI(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("nock", flag.ContinueOnError)
	target := fs.String("fire", string(fire.Stdout),
		"where a confirmed command is delivered: stdout|tmux")
	if err := fs.Parse(args); err != nil {
		return err
	}
	t, err := fire.Parse(*target)
	if err != nil {
		return err
	}

	// History is best-effort: a path resolution problem disables it rather than
	// blocking the TUI. NOCK_HISTORY=off yields an empty path (a no-op store).
	path, err := history.DefaultPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "nock: warning, history disabled: %v\n", err)
	}

	return tui.Run(ctx, loadEngine(vars.New()), tui.Options{
		History:       history.New(path),
		DefaultTarget: t,
	})
}

func cmdMCP(ctx context.Context) error {
	return mcp.New(loadEngine(vars.New())).Serve(ctx)
}

func cmdSearch(_ context.Context, args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	if err := fs.Parse(args); err != nil {
		return err
	}
	query := strings.Join(fs.Args(), " ")
	e := loadEngine(vars.New())
	results := e.Search(query)
	if len(results) == 0 {
		fmt.Println("no matches")
		return nil
	}
	for _, r := range results {
		fmt.Printf("%-24s [%s] %s\n", r.Command.ID, r.Sheet, r.Command.Command)
	}
	return nil
}

func cmdResolve(_ context.Context, args []string) error {
	// Hand-parsed so --var bindings work in any position relative to the command id
	// (the stdlib flag package stops at the first positional argument).
	store := vars.New()
	var id string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--var" || args[i] == "-var":
			if i+1 >= len(args) {
				return fmt.Errorf("resolve: --var needs a key=value argument")
			}
			i++
			k, v, ok := strings.Cut(args[i], "=")
			if !ok {
				return fmt.Errorf("resolve: bad --var %q, want key=value", args[i])
			}
			store.Set(k, v)
		case strings.HasPrefix(args[i], "--var="):
			k, v, ok := strings.Cut(strings.TrimPrefix(args[i], "--var="), "=")
			if !ok {
				return fmt.Errorf("resolve: bad --var %q, want key=value", args[i])
			}
			store.Set(k, v)
		case id == "":
			id = args[i]
		default:
			return fmt.Errorf("resolve: unexpected argument %q", args[i])
		}
	}
	if id == "" {
		return fmt.Errorf("resolve: command id required")
	}
	e := loadEngine(store)
	cmd, ok := e.Get(id)
	if !ok {
		return fmt.Errorf("resolve: no command with id %q", id)
	}
	out, err := e.Resolve(cmd)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func printUsage() {
	fmt.Fprint(os.Stderr, `nock — AI-aware command launcher

Usage:
  nock                  start the interactive TUI (default)
  nock --mcp            run as an MCP server over stdio (for AI agents)
  nock search <query>   search commands non-interactively
  nock resolve <id> --var k=v ...   resolve a command's variables
  nock serve            team server (Milestone 5)
  nock import <src>     import arsenal cheatsheets (Milestone 0/1)
  nock version          print build metadata

Environment:
  NOCK_CHEATSHEETS      cheatsheet directory (default: examples/cheatsheets)
`)
}
