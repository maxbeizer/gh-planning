# Project Status Skill

Ask about project board status, what's in progress, blocked, or stale.

## Usage
- "What's on my plate?"
- "Show me blocked items"
- "What's stale in project 25?"
- "How many items are in backlog?"

## Tools
This skill uses `gh planning status` to query GitHub Projects v2.

### Get project status
```bash
gh planning status --project {project_number} --owner {owner} --json
```

### Get stale items
```bash
gh planning status --project {project_number} --stale {duration} --json
```

### Get items by assignee
```bash
gh planning status --project {project_number} --assignee {user} --json
```
