# Command Reference

Complete reference for all `gh planning` commands. For a narrative guide to what gh-planning does and how to use it day-to-day, see [the guide](guide.md).

---

## Global Flags

These flags work on any command that produces output:

| Flag | Description |
|------|-------------|
| `--json` | Output raw JSON instead of formatted text |
| `--jq <expr>` | Apply a jq expression to filter/transform JSON output |

---

## Getting Started

### `gh planning`

Shows a quick summary of your current focus session and default project status counts. Run it with no arguments to see where you left off.

```bash
gh planning
```

### `gh planning setup`

Interactive walkthrough that explains what gh-planning does and configures your default project, team, 1-1 repo pattern, and agent rate limit step by step. Run this once when you first install the extension.

```bash
gh planning setup
```

### `gh planning init`

Initialize config and verify the project exists. Useful when scripting or when you want to skip the interactive setup.

```bash
gh planning init --project 25 --owner maxbeizer
```

### `gh planning tutorial`

Interactive tutorial that teaches gh-planning by doing. Runs real commands against your project so you learn hands-on. Progress is saved automatically — resume where you left off.

```bash
gh planning tutorial
gh planning tutorial --hands-on    # guided exercises with your real project
gh planning tutorial --explore     # free-form exploration mode
gh planning tutorial --list        # list available lessons
gh planning tutorial --reset       # start over from scratch
```

### `gh planning cheatsheet`

Browsable quick-reference organized by scenario. Search and filter interactively to find the command you need fast.

```bash
gh planning cheatsheet
gh planning cheatsheet --plain     # plain text output (no interactivity)
```

### `gh planning guide`

Step-by-step workflow walkthroughs for common scenarios. Run with no arguments to see all available guides.

```bash
gh planning guide              # list available guides
gh planning guide morning      # morning standup routine
gh planning guide new-task     # picking up new work
gh planning guide one-on-one   # preparing for 1-1s
gh planning guide agent        # working with AI agents
gh planning guide breakdown    # breaking down large issues
```

---

## Configuration & Profiles

Profiles let you maintain separate configurations for different contexts (e.g., work vs. personal). The `config` command is an alias for `profile`.

### `gh planning profile set <key> <value>`

Set a configuration value in the active profile.

**Supported keys:**

| Key | Description | Example |
|-----|-------------|---------|
| `default-project` | GitHub Project number | `25` |
| `default-owner` | Project owner (user or org) | `maxbeizer` |
| `team` | Comma-separated GitHub usernames | `maxbeizer,claudia-bot` |
| `1-1-repo-pattern` | Repo pattern for 1-1 notes | `maxbeizer/{handle}-1-1` |
| `agent.max-per-hour` | Agent rate limit | `10` |
| `repos` | Comma-separated repos (supports globs) | `github/github,github/gh-*` |
| `orgs` | Comma-separated GitHub orgs for auto-detection | `github` |

```bash
gh planning profile set team maxbeizer,claudia-bot
gh planning profile set repos github/github,github/gh-*
gh planning profile set orgs github
```

### `gh planning profile show`

Show the current profile configuration (YAML by default). If auto-detection matched a profile, that profile is shown.

```bash
gh planning profile show
```

### `gh planning profile use <profile>`

Switch to a named profile. Creates the profile if it doesn't exist. Your existing config is preserved as the "default" profile on first use.

```bash
gh planning profile use work
gh planning profile use personal
```

### `gh planning profile list`

List all profiles. Marks which profile is active and which would be auto-detected for the current repo.

```bash
gh planning profile list
```

### `gh planning profile detect`

Show which profile matches the current repo based on `repos` and `orgs` fields. Useful for debugging auto-detection.

```bash
gh planning profile detect
```

### `gh planning profile delete <profile>`

Delete a profile. Cannot delete the currently active profile.

```bash
gh planning profile delete old-project
```

---

## Project Views

### `gh planning status`

Display project items grouped by status. The default view for seeing everything on your plate at a glance.

```bash
gh planning status --project 25 --owner maxbeizer
gh planning status --assignee maxbeizer --stale 7d   # highlight stale items
gh planning status --board                            # kanban-style output
gh planning status --swimlanes                        # grouped by assignee
gh planning status --exclude Done,Closed              # hide specific statuses
gh planning status --open                             # shorthand: exclude Done/Closed
```

### `gh planning board`

Kanban board view of your project. Excludes Done/Completed/Closed columns by default so you see only active work.

```bash
gh planning board
gh planning board --swimlanes      # group rows by assignee
gh planning board --include-done   # show completed items too
gh planning board --open           # only open items
```

---

## Daily Workflow

### `gh planning track "<title>"`

Create a new issue and add it to the project in one step. Saves you from creating an issue, then manually adding it to the board.

```bash
gh planning track "Fix auth bug" --repo maxbeizer/app --body "Details" --label bug --assignee maxbeizer --status "In Progress"
```

### `gh planning focus <issue>`

Set your current focus to a specific issue, or show your current focus. Focusing on an issue lets other commands (like `log` and `unfocus`) know what you're working on.

```bash
gh planning focus maxbeizer/app#42   # set focus
gh planning focus                    # show current focus
```

### `gh planning unfocus`

Clear your current focus session and optionally leave a comment on the issue summarizing what you did.

```bash
gh planning unfocus
gh planning unfocus --comment "Wrapped this up"
```

### `gh planning log`

Log progress, decisions, blockers, and findings against your current focus issue. Builds a structured timeline of your work.

```bash
gh planning log "OAuth callback working"
gh planning log --decision "Using JWT for stateless auth"
gh planning log --blocker "Need API key"
gh planning log --tried "Session approach, too complex"
gh planning log --result "Latency down to 45ms"
```

### `gh planning logs`

Show the progress log timeline. Review what you (or your agents) have been working on.

```bash
gh planning logs
gh planning logs --all --since 7d
```

### `gh planning standup`

Generate a standup report from your recent activity. Optionally include your whole team.

```bash
gh planning standup --since 24h
gh planning standup --team
```

### `gh planning catch-up`

Summarize updates since your last session. Great for Monday mornings or after time away.

```bash
gh planning catch-up
gh planning catch-up --since friday
```

### `gh planning breakdown <issue>`

Split a large issue into sub-issues using GitHub Models. Helps you plan work without doing the breakdown manually.

```bash
gh planning breakdown https://github.com/maxbeizer/app/issues/42 --dry-run
gh planning breakdown 42 --repo maxbeizer/app
```

### `gh planning review <pr>`

Quick review summary for a pull request. Get up to speed on what a PR does before diving into the diff.

```bash
gh planning review 48 --repo maxbeizer/app
```

---

## Handoffs & Completion

### `gh planning claim <issue>`

Claim an issue by assigning yourself and moving it to In Progress. One command instead of two manual steps.

```bash
gh planning claim maxbeizer/app#42
```

### `gh planning complete <issue>`

Post a completion handoff and move the issue forward. Links to the PR that implements the work.

```bash
gh planning complete maxbeizer/app#42 --done "OAuth flow" --pr 48
```

### `gh planning handoff <issue>`

Post a structured session handoff to an issue. Use this when you're stopping work mid-task and want the next person (or your future self) to pick up seamlessly.

```bash
gh planning handoff maxbeizer/app#42 --done "OAuth flow" --remaining "Logout flow"
```

---

## Team

### `gh planning team`

Show recent activity across your team. See who's been working on what.

```bash
gh planning team --since 7d
gh planning team --team maxbeizer,claudia-bot --quiet
```

### `gh planning prep <github-handle>`

Generate a 1-1 preparation document. Pulls together recent activity, open items, and discussion topics for a specific teammate.

```bash
gh planning prep maxbeizer --since 14d
gh planning prep maxbeizer --notes
```

### `gh planning pulse`

Show team health metrics — throughput, cycle time, and workload distribution.

```bash
gh planning pulse --since 30d
gh planning pulse --team maxbeizer,claudia-bot
```

---

## Agent & AI

### `gh planning agent-context`

Summarize everything an AI agent needs to start work: current focus, project state, recent handoffs, and relevant context. Use `--new-session` at the start of each agent conversation.

```bash
gh planning agent-context --new-session
gh planning agent-context --issue 42 --repo maxbeizer/app
```

### `gh planning queue`

Show items ready for agent processing. Filter by label and status to find work that's been triaged for automation.

```bash
gh planning queue --label agent-ready --status Backlog --status Ready
```

---

## Copilot / MCP

gh-planning ships with an MCP (Model Context Protocol) server so Copilot can call planning commands as tools.

### `gh planning copilot serve`

Start the MCP server. This is what you point your editor or CLI at.

```bash
gh planning copilot serve
```

### `gh planning copilot skills`

List all Copilot skills registered by the MCP server.

```bash
gh planning copilot skills
```

### `gh planning copilot test`

Test which skill matches a given natural-language query. Useful for debugging skill routing.

```bash
gh planning copilot test "what am I working on?"
```

### MCP Configuration

**VS Code** (`.vscode/mcp.json` or user settings):

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

**Copilot CLI** (`~/.config/github-copilot/config.yml`):

```yaml
mcp_servers:
  gh-planning:
    command: gh
    args: ["planning", "copilot", "serve"]
```
