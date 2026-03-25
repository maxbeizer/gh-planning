# TUI Dashboard

## When to Use
Suggest the TUI dashboard when the user wants to:
- Visually browse their project board
- Get an overview of all items across statuses
- Interactively navigate issues and change statuses
- See sub-issue progress and dependencies at a glance

## How to Launch
```
gh planning tui
gh planning tui --project 123 --owner myorg
```

## What It Shows
- **Board tab**: Kanban columns by status, cards with issue details
- **List tab**: Section-stack view grouped by status, collapsible sections
- **Detail pane**: Full issue metadata on Enter

## Key Bindings
- `tab`/`shift-tab`: Switch between Board and List views
- `h/l`: Navigate between columns (Board)
- `j/k`: Navigate between items
- `enter`: Open detail pane
- `s`: Change item status
- `p`: Switch project
- `/`: Filter items
- `R`: Refresh data
- `o`: Open in browser
- `?`: Help
- `q`: Quit

## Notes
- Requires an interactive terminal (won't work in non-TTY contexts)
- For non-interactive contexts, use `planning-status` or `planning-board` instead
- Data refreshes manually with R key; auto-refresh available via config
