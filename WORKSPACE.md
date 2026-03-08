# WORKSPACE.md — Agent Context File

> This file provides structured context for AI agents working in gh-planning.
> Read this at session start to understand the project, its architecture,
> and current priorities.

---

## Project

| Field | Value |
|-------|-------|
| Name | gh-planning |
| Type | GitHub CLI extension (Go) |
| Install | `gh extension install maxbeizer/gh-planning` |
| Owner | @maxbeizer |
| Repo | [maxbeizer/gh-planning](https://github.com/maxbeizer/gh-planning) |

**What it does:** Terminal-based command center for GitHub Projects (V2).
Tracks issues, generates standups, manages focus sessions, breaks down
issues, coordinates team activity, and integrates with Copilot via MCP.

---

## Architecture

```
main.go                  → entry point, signal handling
cmd/                     → cobra commands (one file per command)
  root.go                → root command, global flags
  setup.go               → interactive first-time setup
  status.go              → project status (list view)
  board.go               → kanban/swimlane rendering
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

---

## Key Conventions

- **CLI framework:** cobra — each command is a file in `cmd/`
- **GitHub API:** shells out to `gh` CLI, NOT the Go API client
- **Config:** `~/.config/gh-planning/config.yaml`, supports named profiles
- **State:** `~/.config/gh-planning/state.json` (handoffs, logs)
- **Cache:** `~/Library/Caches/gh-planning/` (project data, 2min TTL)
- **Output:** all commands support `--json` and `--jq` flags
- **Error handling:** return errors up, no `os.Exit` in commands
- **Build:** `make build` → `bin/gh-planning`, `make ci` for full check

---

## Current Priorities

1. Merge open PRs (#6–#10): tests, MCP tools, org support, board cmd, CI
2. Improve board rendering for narrow terminals
3. Consider TUI mode (bubbletea) for interactive board

---

## Open PRs

<!-- Dynamic — refresh with: gh pr list --repo maxbeizer/gh-planning --state open -->

| PR | Title | Branch |
|----|-------|--------|
| #6 | Add test coverage | add-tests |
| #7 | Register new commands in MCP tools | register-mcp-tools |
| #8 | Support org-owned projects | support-org-projects |
| #9 | Add standalone board command | standalone-board-cmd |
| #10 | Add CI and release workflows | add-ci-workflow |

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `gopkg.in/yaml.v3` | Config file parsing |
| `github.com/mattn/go-runewidth` | Emoji-aware column alignment |
| `golang.org/x/term` | Terminal width detection |

---

*Last manual update: 2026-03-08*
