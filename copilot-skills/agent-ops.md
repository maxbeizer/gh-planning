# Agent Operations Skill

Agent workflow primitives — context, claiming, completing, queueing work.

## Usage
- "What does the agent need to know?"
- "Claim issue #42 for the agent"
- "What's in the agent queue?"
- "Mark #42 as complete"

## Tools
```bash
gh planning agent-context --json
gh planning claim {issue_number} --repo {owner/repo} --json
gh planning complete {issue_number} --repo {owner/repo} --json
gh planning queue --project {project_number} --json
gh planning review {pr_number} --repo {owner/repo} --json
```
