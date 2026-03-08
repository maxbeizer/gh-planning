# gh-planning Agent Instructions

Add this to your `CLAUDE.md`, Copilot instructions, or agent system prompt.

---

## For the agent

```markdown
## Project Management — gh-planning

This project uses `gh planning` for issue tracking and work coordination.
Run this at the start of every conversation (or after /clear):

    gh planning agent-context --new-session

This gives you: current focus, project status, recent logs, pending
handoffs, blocked items, and what to work on next.

### Workflow

1. **Start**: Run `gh planning agent-context --new-session`
2. **Pick work**: Use the suggested next item, or run `gh planning queue`
3. **Claim**: `gh planning claim <owner/repo#number>`
4. **Log progress** as you work:
   - `gh planning log "Implemented retry logic"`
   - `gh planning log --decision "Using exponential backoff"`
   - `gh planning log --blocker "Need API key for external service"`
   - `gh planning log --tried "Connection pooling — too complex for now"`
   - `gh planning log --result "Latency reduced from 200ms to 45ms"`
5. **Hand off** when the session ends or you need a different session to continue:
   - `gh planning handoff <issue> --done "X" --remaining "Y" --decision "Z"`
6. **Complete** when the work is done:
   - `gh planning complete <issue> --done "Implemented retry" --pr 48`

### Rules

- Always run `agent-context --new-session` first. It has everything you need.
- Log decisions and blockers as you go — the next session depends on them.
- Never skip the handoff. Even if the work is done, record what you did.
- If you're blocked, log it and move on: `gh planning log --blocker "reason"`
- Check `gh planning logs` to see what previous sessions recorded.
```

---

## Variations

### Minimal (one line in CLAUDE.md)

```markdown
Run `gh planning agent-context --new-session` at conversation start.
```

### With Copilot CLI

If using GitHub Copilot CLI, add to your custom instructions or
`.github/copilot-instructions.md`:

```markdown
## Project Tracking

This repo uses `gh planning` for project management. At session start:
1. Run `gh planning agent-context --new-session` to get full context
2. Use `gh planning log` to record progress, decisions, and blockers
3. Use `gh planning handoff` or `gh planning complete` before ending
```

### With Cursor / Windsurf

Add to `.cursorrules` or project instructions:

```markdown
Run `gh planning agent-context --new-session` at the start of each task.
Log progress with `gh planning log`. Hand off with `gh planning handoff`.
```
