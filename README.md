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

### `gh planning standup`

Generate a standup report (optionally for your team).

```bash
gh planning standup --since 24h
gh planning standup --team
```

### `gh planning catch-up`

Summarize updates since your last session.

```bash
gh planning catch-up
gh planning catch-up --since friday
```

### `gh planning breakdown <issue>`

Split an issue into sub-issues with GitHub Models.

```bash
gh planning breakdown https://github.com/maxbeizer/app/issues/42 --dry-run
gh planning breakdown 42 --repo maxbeizer/app
```

### `gh planning handoff <issue>`

Post a structured session handoff to an issue.

```bash
gh planning handoff maxbeizer/app#42 --done "OAuth flow" --remaining "Logout flow"
```

### `gh planning agent-context`

Summarize everything an AI agent needs to start work.

```bash
gh planning agent-context
gh planning agent-context --issue 42 --repo maxbeizer/app
```

### `gh planning claim <issue>`

Claim an issue and move it to In Progress.

```bash
gh planning claim maxbeizer/app#42
```

### `gh planning complete <issue>`

Post a completion handoff and move the issue forward.

```bash
gh planning complete maxbeizer/app#42 --done "OAuth flow" --pr 48
```

### `gh planning queue`

Show items ready for agent processing.

```bash
gh planning queue --label agent-ready --status Backlog --status Ready
```

### `gh planning review <pr>`

Quick review summary for a pull request.

```bash
gh planning review 48 --repo maxbeizer/app
```

### `gh planning team`

Show recent activity across your team.

```bash
gh planning team --since 7d
gh planning team --team maxbeizer,claudia-bot --quiet
```

### `gh planning prep <github-handle>`

Generate a 1-1 preparation document.

```bash
gh planning prep maxbeizer --since 14d
gh planning prep maxbeizer --notes
```

### `gh planning pulse`

Show team health metrics.

```bash
gh planning pulse --since 30d
gh planning pulse --team maxbeizer,claudia-bot
```

## Global Flags

All commands accept:

- `--json` for JSON output
- `--jq <expr>` to filter JSON output (requires `--json`)

## Development

```bash
make build
```
