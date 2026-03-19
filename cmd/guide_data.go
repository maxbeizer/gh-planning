package cmd

import "github.com/maxbeizer/gh-planning/internal/tui"

type workflow struct {
	Name        string
	Title       string
	Description string
	Steps       []tui.Step
}

func availableWorkflows() []workflow {
	return []workflow{
		morningWorkflow(),
		newTaskWorkflow(),
		oneOnOneWorkflow(),
	}
}

func morningWorkflow() workflow {
	return workflow{
		Name:        "morning",
		Title:       "🌅 Morning Routine",
		Description: "Start your day by catching up, reviewing your board, and planning ahead.",
		Steps: []tui.Step{
			{
				Title:       "Catch Up",
				Description: "See what happened while you were away.",
				Content:     "This pulls recent activity from your project — new issues, status changes,\nclosed PRs, and comments since your last session or a given time.",
				Command:     "gh planning catch-up --since friday",
				DocURL:      "https://docs.github.com/en/issues/planning-and-tracking-with-projects",
			},
			{
				Title:       "Review Your Board",
				Description: "Get a quick snapshot of where everything stands.",
				Content:     "The board command shows a kanban-style view of your project.\nUse --swimlanes to see items grouped by assignee.",
				Command:     "gh planning board",
			},
			{
				Title:       "Check for Stale Items",
				Description: "Find items that haven't moved in a while.",
				Content:     "Items without updates become invisible bottlenecks.\nUse --stale to surface them so you can re-prioritize or unblock.",
				Command:     "gh planning status --stale 7d",
			},
			{
				Title:       "Generate Your Standup",
				Description: "Auto-generate a standup report from your actual activity.",
				Content:     "Pulls your commits, PR activity, and issue updates to build\na summary of done, doing, and blocked — no manual writing needed.",
				Command:     "gh planning standup --since 24h",
			},
			{
				Title:       "Pick Up Work",
				Description: "Find the next thing to focus on.",
				Content:     "Use the board or status commands to find items ready for work,\nthen set your focus to start tracking time and progress.",
				Command:     "gh planning board",
			},
		},
	}
}

func newTaskWorkflow() workflow {
	return workflow{
		Name:        "new-task",
		Title:       "🎯 New Task Lifecycle",
		Description: "From creating an issue to completing it — the full task workflow.",
		Steps: []tui.Step{
			{
				Title:       "Track a New Issue",
				Description: "Create an issue and add it to your project board.",
				Content:     "This creates a GitHub issue and automatically adds it to your\nconfigured project board with the specified status.",
				Command:     `gh planning track "Fix auth bug" --repo maxbeizer/app --status "In Progress"`,
				DocURL:      "https://docs.github.com/en/issues/tracking-your-work-with-issues",
			},
			{
				Title:       "Focus on It",
				Description: "Set your active focus to track time and context.",
				Content:     "Focus mode tracks elapsed time and makes other commands\ncontext-aware — log entries are tied to your focus issue.",
				Command:     "gh planning focus maxbeizer/app#42",
			},
			{
				Title:       "Log Your Progress",
				Description: "Record progress, decisions, and blockers as you work.",
				Content:     "Logging creates a trail of what happened. Use --decision for\ndesign choices, --blocker for impediments, --tried for experiments.",
				Command:     `gh planning log "OAuth callback working"`,
			},
			{
				Title:       "Clear Your Focus",
				Description: "End your focus session when done.",
				Content:     "Unfocus clears your active session so you can pick up\nthe next task.",
				Command:     "gh planning unfocus",
			},
		},
	}
}

func oneOnOneWorkflow() workflow {
	return workflow{
		Name:        "one-on-one",
		Title:       "🤝 1-1 Preparation",
		Description: "Prepare for 1-1 meetings with data-driven talking points.",
		Steps: []tui.Step{
			{
				Title:       "Check Team Activity",
				Description: "See what your report has been working on.",
				Content:     "The team command shows recent commits, PRs, and issue activity\nfor team members. Great context before a 1-1.",
				Command:     "gh planning team --since 14d",
			},
			{
				Title:       "Generate 1-1 Prep",
				Description: "Auto-generate a preparation document.",
				Content:     "Prep pulls activity data and structures it into talking points:\ncompleted work, open items, potential blockers, and velocity trends.",
				Command:     "gh planning prep maxbeizer --since 14d",
			},
			{
				Title:       "Review Team Health",
				Description: "Check velocity and throughput metrics.",
				Content:     "Pulse gives you team health metrics — item throughput, cycle time,\nand staleness. Helps identify systemic issues for the 1-1 agenda.",
				Command:     "gh planning pulse --since 30d",
			},
		},
	}
}

