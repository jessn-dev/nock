module github.com/jessn-dev/nock

go 1.26

// Heavy dependencies are added per-milestone, deliberately:
//   - github.com/charmbracelet/bubbletea  (Milestone 2 — TUI)
//   - github.com/modelcontextprotocol/go-sdk (Milestone 3 — MCP server)
//   - github.com/anthropics/anthropic-sdk-go (Milestone 4 — AI ranker)
//   - github.com/goccy/go-yaml            (Milestone 0/1 — cheatsheet parsing)
// The scaffold is stdlib-only so it builds and tests offline with zero network.
