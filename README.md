# gh-planning

The developer's command center for GitHub-native project management.

## Installation

```bash
gh extension install maxbeizer/gh-planning
```

## Quick Start

```bash
# 1) Initialize default project
gh planning init --project 25 --owner maxbeizer

# 2) View status
gh planning status

# 3) Track a new issue in your project
gh planning track "Fix auth bug" --repo maxbeizer/app --status "In Progress"

# 4) Focus on a specific issue
gh planning focus maxbeizer/app#42
```

## Commands

### `gh planning`

Shows a quick summary of your current focus session and default project status counts.

### `gh planning init`

Initialize config and verify the project exists.

```bash
gh planning init --project 25 --owner maxbeizer
```

### `gh planning config set <key> <value>`

Set configuration values.

Supported keys:
- `default-project`
- `default-owner`
- `team` (comma-separated GitHub usernames)
- `1-1-repo-pattern` (example: `maxbeizer/{handle}-1-1`)
- `agent.max-per-hour`

Example:

```bash
gh planning config set team maxbeizer,claudia-bot
```

### `gh planning config show`

Show the current config (YAML by default).

```bash
gh planning config show
```

### `gh planning status`

Display project items grouped by status.

```bash
gh planning status --project 25 --owner maxbeizer
gh planning status --assignee maxbeizer --stale 7d
```

### `gh planning track "<title>"`

Create a new issue and add it to the project.

```bash
gh planning track "Fix auth bug" --repo maxbeizer/app --body "Details" --label bug --assignee maxbeizer --status "In Progress"
```

### `gh planning focus <issue>`

Set or show focus. The issue must be in `owner/repo#number` format.

```bash
gh planning focus maxbeizer/app#42
gh planning focus
```

### `gh planning unfocus`

Clear focus and optionally comment on the issue.

```bash
gh planning unfocus --comment "Wrapped this up"
```

## Global Flags

All commands accept:

- `--json` for JSON output
- `--jq <expr>` to filter JSON output (requires `--json`)

## Phase 2+ (Coming Soon)

Standup, catch-up, breakdown, team, and agent workflows are on the roadmap.

## Development

```bash
make build
```
