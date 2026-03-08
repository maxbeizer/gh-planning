# Copilot Instructions for gh-planning

## Project Overview

`gh-planning` is a GitHub CLI extension (`gh planning`) written in Go. It provides a developer command center for GitHub-native project management — tracking issues, generating standups, managing focus sessions, breaking down issues, coordinating team activity, and integrating with Copilot via MCP.

## Repository Structure

```
main.go                  → entry point, signal handling
cmd/                     → cobra commands (one file per command)
  root.go                → root command, global flags
  setup.go               → interactive first-time setup
  status.go              → project status (list view)
  board.go               → kanban/swimlane rendering functions
  board_cmd.go           → standalone board command
  standup.go             → standup report generation
  agent_context.go       → agent session-start context
  log.go                 → progress logging
  claim.go / complete.go → agent work lifecycle
  handoff.go             → structured handoffs
  queue.go               → agent work queue
  config.go              → config management + profiles
  helpers.go             → shared utilities
  agent_helpers.go       → agent-specific helpers
internal/
  config/                → YAML config with named profiles
  github/                → gh CLI wrapper (GraphQL, REST, search)
  session/               → focus session tracking (current.json)
  state/                 → persistent state (handoffs, logs)
  output/                → JSON output formatting
copilot-skills/          → markdown skill definitions
mcp/                     → MCP server for Copilot tool registration
docs/                    → guide and agent instructions
```

## Build & Run

```bash
make build        # produces bin/gh-planning
make ci           # build + vet + test-race
make relink-local # reinstall extension from local checkout
make help         # see all targets
go mod tidy       # resolve dependency issues
```

## Conventions

- **CLI framework**: [cobra](https://github.com/spf13/cobra) — each command is a file in `cmd/`.
- **Config**: YAML-based with named profiles, managed via `internal/config`. Stored at `~/.config/gh-planning/config.yaml`.
- **State**: JSON at `~/.config/gh-planning/state.json` (handoffs, progress logs).
- **Cache**: Project data cached at `~/Library/Caches/gh-planning/` with 2-minute TTL.
- **GitHub API calls**: Use the `gh` CLI under the hood (shelling out via `exec.Command`), not the Go API client.
- **Output**: Support `--json` and `--jq` global flags on all commands.
- **Error handling**: Return errors up to `main`; avoid `os.Exit` in commands.
- **Tests**: Standard `go test` with `_test.go` files alongside the code they test.
- **Emoji alignment**: Use `github.com/mattn/go-runewidth` for terminal column math, never `len()`.

## Adding a New Command

1. Create `cmd/<command>.go` with a `cobra.Command`.
2. Register it in `cmd/root.go` via `rootCmd.AddCommand(...)`.
3. If the command should be exposed as a Copilot skill, add a markdown file in `copilot-skills/` and register it in `mcp/tools.go`.

## Style

- Keep commands self-contained in their own file.
- Prefer helper functions in `cmd/helpers.go` or `cmd/agent_helpers.go` for reusable logic.
- Use `fmt.Fprintf(cmd.OutOrStdout(), ...)` for output so it can be captured in tests.
- Parallelize independent API calls using goroutines + channels (see `standup.go`, `catchup.go` for patterns).
