# Project Status Skill

Ask about project board status, what's in progress, blocked, or stale.

## Usage
- "What's on my plate?"
- "Show me blocked items"
- "What's stale in project 25?"
- "Show workstream Quality"
- "Run a board hygiene report for items I manage"
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

### Get items by project field
```bash
gh planning status --project {project_number} --field {Field}={Value} --json
gh planning board --project {project_number} --field Manager=me
```

### Run board hygiene
```bash
gh planning hygiene --project {project_number} --field Manager=me --stale 7d --format markdown
```
