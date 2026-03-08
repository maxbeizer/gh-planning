# How gh-planning Works

A plain-English guide to what `gh planning` does and why you'd use it.

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

# A big issue? Break it down into sub-issues with AI
gh planning breakdown maxbeizer/app#42

# Quick PR review summary
gh planning review 48 --repo maxbeizer/app
```

### Wrap up

```bash
# Clear focus and leave a note on the issue
gh planning unfocus --comment "OAuth flow done, logout still needs work"

# Or do a structured handoff
gh planning handoff maxbeizer/app#42 --done "OAuth flow" --remaining "Logout"
```

## Team Features

These are useful if you're a tech lead or just want visibility into your
team's work.

```bash
# Configure your team once
gh planning config set team alice,bob,carol

# See what everyone's been up to
gh planning team --since 7d

# Generate a 1-1 prep doc
gh planning prep alice --since 14d

# Team health metrics (PR cycle time, issue velocity, etc.)
gh planning pulse --since 30d
```

## Agent / AI Integration

`gh planning` doubles as an agent command center, inspired by
[td](https://td.haplab.com). AI coding agents can pick up work, log
progress, and hand off results — all through the same project board.

### Session Start

Add this to your `CLAUDE.md`, Copilot instructions, or system prompt:

```
Run `gh planning agent-context --new-session` at conversation start.
```

This gives the agent everything in one shot: current focus, project
status, recent logs, pending handoffs, blocked items, and what to
work on next. See [docs/agent-instructions.md](docs/agent-instructions.md)
for full setup instructions.

### The Agent Loop

```bash
# 1. Get context (run this first, every session)
gh planning agent-context --new-session

# 2. Pick work (use suggested item or browse the queue)
gh planning queue --label agent-ready

# 3. Claim it
gh planning claim maxbeizer/app#42

# 4. Log progress as you work
gh planning log "OAuth callback working"
gh planning log --decision "Using JWT for stateless auth"
gh planning log --blocker "Need clarification on token rotation"
gh planning log --tried "Session-based approach, too complex"
gh planning log --result "Benchmarks show 2ms token validation"

# 5. View the log trail
gh planning logs

# 6. Hand off (if passing to another session)
gh planning handoff maxbeizer/app#42 \
  --done "OAuth flow" \
  --remaining "Token refresh" \
  --decision "Using JWT" \
  --uncertain "Should tokens expire on password change?"

# 7. Or complete (if done)
gh planning complete maxbeizer/app#42 --done "Implemented OAuth" --pr 48
```

### Why This Works

- **No context loss**: `agent-context --new-session` gives the next
  session everything the previous one knew
- **Structured handoffs**: Done, remaining, decisions, and open
  questions are captured explicitly — not guessed from code
- **Progress logging**: Decisions and findings during work are
  preserved for future sessions
- **Same board**: Humans and agents share the same GitHub Project,
  so handoffs between them are seamless

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
| **Focus** | Your current working issue. One at a time. Tracked locally with elapsed time. |
| **Team** | A list of GitHub usernames. Used by `standup --team`, `team`, `pulse`, and `prep`. |
| **Queue** | Project items filtered by label/status — the "inbox" for agent work. |
| **Handoff** | A structured comment posted to an issue summarizing done/remaining work. |

## All Commands

| Command | Purpose |
|---------|---------|
| `setup` | Interactive first-time configuration walkthrough |
| `init` | Set default project (non-interactive) |
| `config set/show` | Read or write individual config values |
| `status` | Project board summary grouped by status |
| `track` | Create an issue and add it to the project |
| `focus` / `unfocus` | Set or clear your current working issue |
| `log` | Log progress, decisions, blockers during work |
| `logs` | View the progress log timeline |
| `standup` | Generate a standup report from GitHub activity |
| `catch-up` | Summarize updates since your last session |
| `breakdown` | Split an issue into sub-issues with AI |
| `handoff` | Post a structured session handoff to an issue |
| `claim` | Assign yourself and move an issue to In Progress |
| `complete` | Post a completion summary and advance the issue |
| `agent-context` | Dump context an AI agent needs to start work |
| `queue` | Show items ready for agent processing |
| `review` | Quick review summary for a PR |
| `team` | Recent activity across your team |
| `prep` | Generate a 1-1 preparation document |
| `pulse` | Team health metrics |
| `copilot` | Copilot skill listing, MCP server, and testing |
