# How gh-planning Works

A plain-English guide to what `gh planning` does and why you'd use it. For a full list of commands and flags, see the [Command Reference](commands.md).

## The Problem

GitHub Projects is great, but using it means switching to the browser,
clicking through the UI, and losing your terminal flow. If you manage
work across multiple repos — or coordinate with teammates and AI agents —
the context-switching adds up fast.

## The Solution

`gh planning` brings your project board into the terminal. Every action
that normally requires the GitHub UI becomes a single command. It reads
and writes the same GitHub Projects (V2) data, so nothing changes for
teammates who prefer the web.

## A Typical Day

### Morning: catch up and plan

```bash
# What happened since you last looked?
gh planning catch-up

# Auto-generate a standup from real GitHub activity
gh planning standup --since 24h

# See your board at a glance
gh planning status
```

### Pick something to work on

```bash
# Set your current focus (like a pomodoro target)
gh planning focus maxbeizer/app#42

# The root command always shows your focus
gh planning
# => 🎯 Focus: maxbeizer/app#42 (1h 23m)
```

### During the day

```bash
# Create an issue and add it to your project in one shot
gh planning track "Add retry logic" --repo maxbeizer/app --status "In Progress"

# Quick PR review summary
gh planning review 48 --repo maxbeizer/app
```

### Wrap up

```bash
# Clear focus and leave a note on the issue
gh planning unfocus --comment "OAuth flow done, logout still needs work"
```

## Team Features

These are useful if you're a tech lead or just want visibility into your
team's work.

```bash
# Configure your team once
gh planning profile set team alice,bob,carol

# See what everyone's been up to
gh planning team --since 7d

# Generate a 1-1 prep doc
gh planning prep alice --since 14d

# Team health metrics (PR cycle time, issue velocity, etc.)
gh planning pulse --since 30d
```

## Agent / AI Integration

`gh planning` doubles as an agent command center. AI coding agents can
focus on issues, log progress, and track blockers — all through the same
commands humans use. Your board becomes the shared workspace for human
and AI contributors alike.

### Session Start

Add this to your `CLAUDE.md`, Copilot instructions, or system prompt:

```
Run `gh planning focus <owner/repo#number>` at conversation start, then
check `gh planning logs` for context from previous sessions.
```

This gives the agent a focused issue to work on and access to the
progress trail from earlier sessions.

### The Agent Loop

```bash
# 1. Focus on the issue you're working on
gh planning focus maxbeizer/app#42

# 2. Check what previous sessions recorded
gh planning logs

# 3. Log progress as you work
gh planning log "OAuth callback working"
gh planning log --decision "Using JWT for stateless auth"
gh planning log --blocker "Need clarification on token rotation"
gh planning log --tried "Session-based approach, too complex"
gh planning log --result "Benchmarks show 2ms token validation"

# 4. Mark blockers if needed
gh planning blocked maxbeizer/app#42 --by maxbeizer/app#38

# 5. Clear focus when done
gh planning unfocus --comment "OAuth flow complete, PR #48 ready for review"
```

### Why This Works

- **Progress logging**: Decisions and findings during work are
  preserved for future sessions via `log` / `logs`
- **Focus tracking**: `focus` / `unfocus` give clear session boundaries
  with elapsed time
- **Same board**: Humans and agents share the same GitHub Project,
  so coordination is seamless

### Copilot / MCP

`gh planning` includes an MCP server so Copilot can call these commands
as tools in a chat session.

```bash
# List available Copilot skills
gh planning copilot skills

# Start the MCP server (JSON-RPC over stdio)
gh planning copilot serve

# Test natural-language routing
gh planning copilot test "Show me blocked items"
```

## Setup

```bash
# Install the extension
gh extension install maxbeizer/gh-planning

# Interactive guided setup (recommended)
gh planning setup

# Or configure directly
gh planning init --project 25 --owner maxbeizer
```

## Key Concepts

| Concept | What it means |
|---------|---------------|
| **Project** | A GitHub Projects (V2) board. You set a default with `init` or `setup`. |
| **Profile** | A named config set. Switch between work/personal with `profile use`. Profiles can auto-detect based on your repo. |
| **Focus** | Your current working issue. One at a time. Tracked locally with elapsed time. |
| **Team** | A list of GitHub usernames. Used by `standup --team`, `team`, `pulse`, and `prep`. |
| **Blocked** | An issue blocked by another issue. Managed with `blocked` and `unblock`. |

## All Commands

| Command | Purpose |
|---------|---------|
| `setup` | Interactive first-time configuration walkthrough |
| `init` | Set default project (non-interactive) |
| `profile set/show` | Read or write individual profile values |
| `profile use/list/delete` | Switch between named profiles |
| `profile create/update` | Create or update profiles with flags or interactively |
| `profile detect` | Show which profile matches the current repo |
| `status` | Project board summary (list, `--board`, or `--swimlanes`) |
| `board` | Kanban board view (excludes Done by default) |
| `track` | Create an issue and add it to the project |
| `focus` / `unfocus` | Set or clear your current working issue |
| `log` | Log progress, decisions, blockers during work |
| `logs` | View the progress log timeline |
| `blocked` / `unblock` | Mark or remove blockers between issues |
| `standup` | Generate a standup report from GitHub activity |
| `catch-up` | Summarize updates since your last session |
| `review` | Quick review summary for a PR |
| `team` | Recent activity across your team |
| `prep` | Generate a 1-1 preparation document |
| `pulse` | Team health metrics |
| `tutorial` | Interactive hands-on tutorial |
| `cheatsheet` | Browsable quick-reference by scenario |
| `guide` | Step-by-step workflow walkthroughs |
| `copilot` | Copilot skill listing, MCP server, and testing |
