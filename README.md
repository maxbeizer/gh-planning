# gh-planning

The developer's command center for GitHub-native project management.

## Installation

```bash
gh extension install maxbeizer/gh-planning
```

## Quick Start

```bash
# Interactive guided setup (recommended for first-time users)
gh planning setup

# Or initialize directly with flags
gh planning init --project 25 --owner maxbeizer

# View status
gh planning status

# 3) Track a new issue in your project
gh planning track "Fix auth bug" --repo maxbeizer/app --status "In Progress"

# 4) Focus on a specific issue
gh planning focus maxbeizer/app#42
```

## Learning & Discovery

```bash
# Interactive tutorial — runs real commands, two modes
gh planning tutorial --hands-on   # full lifecycle: create → claim → focus → log → complete
gh planning tutorial --explore    # read-only tour of your existing project

# Browsable cheatsheet — search commands by scenario
gh planning cheatsheet

# Workflow guides — step-by-step walkthroughs
gh planning guide morning      # morning catch-up routine
gh planning guide new-task     # full task lifecycle
gh planning guide one-on-one   # 1-1 preparation
gh planning guide agent        # AI agent workflow
gh planning guide breakdown    # issue decomposition
```

## Commands

### `gh planning`

Shows a quick summary of your current focus session and default project status counts.

### `gh planning setup`

Interactive walkthrough that explains what gh-planning does and configures
your default project, team, 1-1 repo pattern, and agent rate limit step by step.

```bash
gh planning setup
```

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

### `gh planning config use <profile>`

Switch to a named config profile. Creates the profile if it doesn't exist.
Your existing config is preserved as the "default" profile on first use.

```bash
gh planning config use work
gh planning config use personal
```

### `gh planning config profiles`

List all config profiles.

```bash
gh planning config profiles
```

### `gh planning config delete <profile>`

Delete a config profile (cannot delete the active one).

```bash
gh planning config delete old-project
```

### `gh planning status`

Display project items grouped by status.

```bash
gh planning status --project 25 --owner maxbeizer
gh planning status --assignee maxbeizer --stale 7d
gh planning status --board               # kanban board view
gh planning status --swimlanes           # board with assignee rows
gh planning status --exclude Done,Closed  # hide statuses
gh planning status --open                 # open project in browser (no fetch)
```

### `gh planning board`

Kanban board view of your project. Excludes Done/Completed/Closed by default.

```bash
gh planning board
gh planning board --swimlanes         # rows per assignee
gh planning board --include-done      # show completed items too
gh planning board --open              # open in browser
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

Summarize everything an AI agent needs to start work. Use `--new-session`
at the start of each agent conversation.

```bash
gh planning agent-context --new-session
gh planning agent-context --issue 42 --repo maxbeizer/app
```

### `gh planning log`

Log progress, decisions, blockers, and findings against the current focus issue.

```bash
gh planning log "OAuth callback working"
gh planning log --decision "Using JWT for stateless auth"
gh planning log --blocker "Need API key"
gh planning log --tried "Session approach, too complex"
gh planning log --result "Latency down to 45ms"
```

### `gh planning logs`

Show the progress log timeline.

```bash
gh planning logs
gh planning logs --all --since 7d
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

### `gh planning tutorial`

Interactive tutorial that teaches gh-planning by doing. Runs real commands
against your project. Progress is saved automatically — resume where you left off.

```bash
gh planning tutorial              # choose mode interactively
gh planning tutorial --hands-on   # full lifecycle: create → claim → focus → log → complete
gh planning tutorial --explore    # read-only tour of your board and reports
gh planning tutorial --list       # see progress
gh planning tutorial --reset      # start over
```

### `gh planning cheatsheet`

Browsable quick-reference organized by scenario. Search and filter interactively.

```bash
gh planning cheatsheet            # interactive browser
gh planning cheatsheet --plain    # static text output
```

### `gh planning guide <workflow>`

Step-by-step workflow walkthroughs.

```bash
gh planning guide                 # list available workflows
gh planning guide morning         # morning routine
gh planning guide new-task        # full task lifecycle
gh planning guide one-on-one      # 1-1 preparation
gh planning guide agent           # AI agent workflow
gh planning guide breakdown       # issue decomposition
```

## Copilot Integration

Copilot skills live in `copilot-skills/` and map to `gh planning` commands.

```bash
gh planning copilot skills
```

To start the MCP server (JSON-RPC over stdio):

```bash
gh planning copilot serve
```

Test a natural language query to see which skill/command would be selected:

```bash
gh planning copilot test "Show me blocked items"
```

Native Copilot plugin registration is planned once the MCP format stabilizes.

## Global Flags

All commands accept:

- `--json` for JSON output
- `--jq <expr>` to filter JSON output (requires `--json`)

## Development

```bash
make build
```
