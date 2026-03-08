# Copilot Instructions for gh-planning

## Project Overview

`gh-planning` is a GitHub CLI extension (`gh planning`) written in Go. It provides a developer command center for GitHub-native project management — tracking issues, generating standups, managing focus sessions, breaking down issues, coordinating team activity, and integrating with Copilot via MCP.

## Repository Structure

- `main.go` — entry point; sets up signal handling and delegates to `cmd.ExecuteContext`
- `cmd/` — all CLI subcommands (cobra-based): `init`, `status`, `track`, `focus`, `standup`, `breakdown`, `team`, `prep`, `pulse`, `queue`, `review`, `copilot`, etc.
- `internal/` — shared packages (config, GitHub API helpers, formatting)
- `copilot-skills/` — markdown skill definitions for Copilot integration
- `mcp/` — MCP (Model Context Protocol) server for Copilot tool registration

## Build & Run

```bash
make build        # produces ./gh-planning binary
make install      # installs as a gh extension
go mod tidy       # resolve dependency issues
```

The Makefile builds with `go build -o gh-planning .` (single main package, not `./...`).

## Conventions

- **CLI framework**: [cobra](https://github.com/spf13/cobra) — each command is a file in `cmd/`.
- **Config**: YAML-based, managed via `internal/config`. Stored at `~/.config/gh-planning/config.yml`.
- **GitHub API calls**: Use the `gh` CLI under the hood (shelling out via `exec.Command`), not the Go API client.
- **Output**: Support `--json` and `--jq` global flags on all commands.
- **Error handling**: Return errors up to `main`; avoid `os.Exit` in commands.
- **No tests yet**: When adding tests, use standard `go test` and place `_test.go` files alongside the code they test.

## Adding a New Command

1. Create `cmd/<command>.go` with a `cobra.Command`.
2. Register it in `cmd/root.go` via `rootCmd.AddCommand(...)`.
3. If the command should be exposed as a Copilot skill, add a markdown file in `copilot-skills/` and register it in `mcp/tools.go`.

## Style

- Keep commands self-contained in their own file.
- Prefer helper functions in `cmd/helpers.go` or `cmd/agent_helpers.go` for reusable logic.
- Use `fmt.Fprintf(cmd.OutOrStdout(), ...)` for output so it can be captured in tests.
