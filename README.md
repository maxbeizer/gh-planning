# gh-planning

Your project board, in your terminal.

## Why gh-planning?

**Stay in your flow.** No more switching to the browser to drag a card or update a status. Every project action — tracking, triaging, focusing, logging — is a single command in the terminal where you already work.

**Works with your existing setup.** gh-planning reads and writes GitHub Projects V2 directly. Your teammates using the web UI see the same board, the same statuses, the same assignments. There's nothing to migrate and nothing to sync.

**Built for AI-assisted workflows.** Agents can focus on issues, log progress, and track blockers — all through the same commands humans use. Your board becomes the shared workspace for human and AI contributors alike.

**Team visibility without meetings.** Generate standup reports, catch-up summaries, 1-1 prep docs, and team health metrics from real GitHub activity. No more "what did you work on?" — the data is already there.

## Documentation

- **[Guide](docs/guide.md)** — narrative walkthrough of a typical day
- **[Command Reference](docs/commands.md)** — complete reference for all commands
- **[Agent Instructions](docs/agent-instructions.md)** — setup for AI coding agents

## Quick Start

```bash
# Install
gh extension install maxbeizer/gh-planning

# Interactive setup (recommended)
gh planning setup

# See your board
gh planning status

# Focus on an issue
gh planning focus owner/repo#42

# Generate a standup
gh planning standup
```

## What can it do?

| Category              | Commands                                                                       |                                                               |
| --------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------------- |
| **Personal workflow** | `status` `board` `track` `focus` `unfocus` `log` `logs` `blocked` `unblock`   | View your board, track issues, focus on work, manage blockers |
| **Reports**           | `standup` `catch-up` `review`                                                  | Generate standups, catch-up summaries, PR reviews             |
| **Team**              | `team` `prep` `pulse`                                                          | Team activity, 1-1 prep docs, health metrics                  |
| **Learning**          | `tutorial` `cheatsheet` `guide`                                                | Interactive tutorials, searchable cheatsheet, workflow guides |
| **Configuration**     | `setup` `init` `profile`                                                       | Interactive setup, initialization, named profiles             |

All commands support `--json` and `--jq` flags for scripting.

→ **[Full command reference](docs/commands.md)**

## Copilot Integration

gh-planning exposes tools via MCP (Model Context Protocol) so Copilot can use your board directly.

**VS Code** — add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "gh-planning": {
      "command": "gh",
      "args": ["planning", "copilot", "serve"]
    }
  }
}
```

**Copilot CLI** — add to `~/.config/github-copilot/config.yml`:

```yaml
mcp_servers:
  gh-planning:
    command: gh
    args: ["planning", "copilot", "serve"]
```

See the [Agent Instructions](docs/agent-instructions.md) for more on agent workflows.

## Development

```bash
make build        # build
make ci           # build + vet + test
make help         # all targets
```

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.
