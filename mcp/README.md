# MCP Server

This folder provides an MCP (Model Context Protocol) JSON-RPC server that exposes `gh planning` commands as tools for Copilot and other AI assistants.

## Setup

### VS Code / Copilot

Add to your `.vscode/mcp.json` (or `~/.vscode/mcp.json` for global):

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

### Copilot CLI

Add to `~/.config/github-copilot/config.yml`:

```yaml
mcp_servers:
  gh-planning:
    command: gh
    args: ["planning", "copilot", "serve"]
```

## Run the Server

```bash
gh planning copilot serve
```

The server listens on stdin and writes JSON-RPC responses to stdout.

## Supported Methods

- `initialize`
- `tools/list`
- `tools/call`

## Available Tools (29)

### Query & Views
- `planning.status` — Project status and filters
- `planning.board` — Kanban board view
- `planning.sprint` — Sprint overview
- `planning.roadmap` — Project roadmap and timeline
- `planning.blocked` — Blocked items and dependencies
- `planning.criticalPath` — Critical path through dependencies
- `planning.prioritize` — Prioritize project items

### Reports
- `planning.standup` — Generate standup report
- `planning.catchup` — Summarize updates since last session
- `planning.team` — Team activity summary
- `planning.prep` — 1-1 prep report
- `planning.pulse` — Team health metrics

### Work Lifecycle
- `planning.track` — Create and track issues
- `planning.claim` — Claim an issue
- `planning.complete` — Complete an issue
- `planning.focus` — Set or show focus
- `planning.estimate` — Add effort estimates
- `planning.review` — PR review summary
- `planning.queue` — Agent work queue
- `planning.breakdown` — Break down issues with AI

### Logging & Handoff
- `planning.log` — Log progress
- `planning.logs` — View log timeline
- `planning.handoff` — Post session handoff
- `planning.agentContext` — Agent context summary

### Configuration
- `planning.profile.show` — Show current profile
- `planning.profile.list` — List profiles
- `planning.profile.detect` — Detect profile from repo

### Discovery
- `planning.cheatsheet` — Command quick-reference
- `planning.guide` — Workflow guides

## Example

Request:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "planning.status",
    "arguments": {
      "project": 25,
      "owner": "maxbeizer"
    }
  }
}
```

Response:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "status": "ok"
  }
}
```
