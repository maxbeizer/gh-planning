# gh-planning Agent Instructions

Add this to your `CLAUDE.md`, Copilot instructions, or agent system prompt.

---

## For the agent

```markdown
## Project Management — gh-planning

This project uses `gh planning` for issue tracking and work coordination.
At the start of every conversation (or after /clear):

1. Focus on the issue you're working on: `gh planning focus <owner/repo#number>`
2. Check previous progress: `gh planning logs`

### Workflow

1. **Start**: `gh planning focus <owner/repo#number>` — sets your current working issue
2. **Check context**: `gh planning logs` — see what previous sessions recorded
3. **Log progress** as you work:
   - `gh planning log "Implemented retry logic"`
   - `gh planning log --decision "Using exponential backoff"`
   - `gh planning log --blocker "Need API key for external service"`
   - `gh planning log --tried "Connection pooling — too complex for now"`
   - `gh planning log --result "Latency reduced from 200ms to 45ms"`
4. **Track blockers**: `gh planning blocked <issue> --by <blocking-issue>`
5. **Finish**: `gh planning unfocus --comment "Summary of what was done"`

### Rules

- Always run `focus` first to set your working issue.
- Log decisions and blockers as you go — the next session depends on them.
- Check `gh planning logs` to see what previous sessions recorded.
- If you're blocked, log it: `gh planning log --blocker "reason"`
- When done, run `unfocus --comment` to record what you accomplished.
```

---

## Variations

### Minimal (one line in CLAUDE.md)

```markdown
Run `gh planning focus <owner/repo#number>` at conversation start, then check `gh planning logs` for context.
```

### With Copilot CLI

If using GitHub Copilot CLI, add to your custom instructions or
`.github/copilot-instructions.md`:

```markdown
## Project Tracking

This repo uses `gh planning` for project management. At session start:
1. Run `gh planning focus <owner/repo#number>` to set your working issue
2. Run `gh planning logs` to see previous progress
3. Use `gh planning log` to record progress, decisions, and blockers
4. Use `gh planning unfocus --comment "summary"` when done
```

### With Cursor / Windsurf

Add to `.cursorrules` or project instructions:

```markdown
Run `gh planning focus <owner/repo#number>` at the start of each task.
Log progress with `gh planning log`. Finish with `gh planning unfocus --comment "summary"`.
```
