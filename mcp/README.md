# MCP Server (Scaffold)

This folder provides a minimal MCP-style JSON-RPC server that exposes `gh planning` commands as tools. The MCP format is still evolving; this is a scaffold to make Copilot integration easy when native registration is available.

## Run the Server

```bash
gh planning copilot serve
```

The server listens on stdin and writes JSON-RPC responses to stdout.

## Copilot Registration (Future)

When Copilot supports MCP server registration, point it at:

```bash
path/to/gh-planning copilot serve
```

## Supported Methods

- `initialize`
- `tools/list`
- `tools/call`

## Tool Schemas

`tools/list` returns the following tools (each expects `arguments` matching the input schema):

- `planning.status`
- `planning.standup`
- `planning.catchup`
- `planning.breakdown`
- `planning.track`
- `planning.team`
- `planning.prep`
- `planning.pulse`
- `planning.agentContext`
- `planning.claim`
- `planning.complete`
- `planning.queue`
- `planning.review`
- `planning.focus`
- `planning.handoff`

Each tool shells out to:

```bash
gh planning <command> --json
```

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
