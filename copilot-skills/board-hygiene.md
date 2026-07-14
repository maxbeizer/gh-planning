# Board Hygiene Skill

Find actionable GitHub Projects board cleanup work.

## Usage
- "Run board hygiene"
- "What active items are stale?"
- "Which items are missing owner or target date?"
- "Check hygiene for items where Manager is me"

## Tools
```bash
gh planning hygiene --json
gh planning hygiene --field Manager=me --stale 7d --json
gh planning hygiene --format markdown
```
