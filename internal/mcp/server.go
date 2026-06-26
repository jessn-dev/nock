// Package mcp exposes the engine as a Model Context Protocol server so AI agents
// (Claude Code, Cursor, Continue, Cline, Windsurf, Zed, …) can query nock's command
// knowledge directly. MCP is provider-agnostic by protocol: one server, every host,
// no per-host code. In MCP mode the host agent pays its own model cost — nock makes
// no LLM calls of its own.
//
// Scaffold status: Serve returns errNotImplemented. The real implementation
// (Milestone 3) wires github.com/modelcontextprotocol/go-sdk over stdio and exposes
// the tools below, each a thin pass-through to the engine API.
package mcp

import (
	"context"
	"errors"

	"github.com/jessn-dev/nock/internal/engine"
)

var errNotImplemented = errors.New("mcp: server arrives in Milestone 3 (modelcontextprotocol/go-sdk)")

// Tool names exposed to MCP hosts. Kept as constants so docs, registration, and
// tests share one source of truth.
const (
	ToolSearchCommands  = "search_commands" // intent -> ranked commands
	ToolGetCommand      = "get_command"     // id (+ vars) -> resolved command
	ToolListCheatsheets = "list_cheatsheets"
	ToolResolveVars     = "resolve_vars" // target -> populated variable set
)

// Server adapts an engine to the MCP protocol.
type Server struct {
	engine *engine.Engine
}

// New returns an MCP server backed by the given engine.
func New(e *engine.Engine) *Server { return &Server{engine: e} }

// Serve runs the MCP server over stdio until ctx is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	return errNotImplemented
}
