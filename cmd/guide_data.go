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
		agentWorkflow(),
		breakdownWorkflow(),
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
				Content:     "The queue shows items ready for work. Claim one and set your focus\nto start tracking time and logging progress against it.",
				Command:     "gh planning queue",
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
				Title:       "Claim the Issue",
				Description: "Assign yourself and move it to In Progress.",
				Content:     "Claiming is a shortcut that assigns you and updates the project\nstatus in one step. Great for picking up items from the backlog.",
				Command:     "gh planning claim maxbeizer/app#42",
			},
			{
				Title:       "Focus on It",
				Description: "Set your active focus to track time and context.",
				Content:     "Focus mode tracks elapsed time and makes other commands\ncontext-aware — log, handoff, and complete all use your focus issue.",
				Command:     "gh planning focus maxbeizer/app#42",
			},
			{
				Title:       "Log Your Progress",
				Description: "Record progress, decisions, and blockers as you work.",
				Content:     "Logging creates a trail of what happened. Use --decision for\ndesign choices, --blocker for impediments, --tried for experiments.",
				Command:     `gh planning log "OAuth callback working"`,
			},
			{
				Title:       "Complete the Task",
				Description: "Post a completion handoff and move the issue forward.",
				Content:     "This posts a structured comment summarizing what was done,\nlinks the PR, and moves the item to the next status.",
				Command:     `gh planning complete maxbeizer/app#42 --done "OAuth flow" --pr 48`,
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

func agentWorkflow() workflow {
	return workflow{
		Name:        "agent",
		Title:       "🤖 AI Agent Workflow",
		Description: "Set up and coordinate AI coding agents with your project.",
		Steps: []tui.Step{
			{
				Title:       "Provide Agent Context",
				Description: "Give an agent everything it needs to start.",
				Content:     "This summarizes your project, current focus, recent activity,\nand conventions so an AI agent can jump in without guessing.",
				Command:     "gh planning agent-context --new-session",
			},
			{
				Title:       "Show Agent Work Queue",
				Description: "See what's ready for agent processing.",
				Content:     "Use labels like 'agent-ready' to tag issues for AI agents.\nThe queue command filters and displays them.",
				Command:     "gh planning queue --label agent-ready",
			},
			{
				Title:       "Break Down Large Issues",
				Description: "Use AI to decompose big issues into sub-tasks.",
				Content:     "Breakdown uses GitHub Models to analyze an issue and suggest\nsub-issues. Use --dry-run to preview before creating them.",
				Command:     "gh planning breakdown 42 --repo maxbeizer/app --dry-run",
				DocURL:      "https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-sub-issues",
			},
			{
				Title:       "Start MCP Server",
				Description: "Let Copilot use gh-planning as a tool.",
				Content:     "The MCP server exposes gh-planning commands as Copilot tools\nover JSON-RPC stdio. Add it to your Copilot config.",
				Command:     "gh planning copilot serve",
			},
			{
				Title:       "Review Agent Output",
				Description: "Review PRs created by AI agents.",
				Content:     "Quick review summary showing changed files, review status,\nand CI checks. Helps you validate agent work efficiently.",
				Command:     "gh planning review 48 --repo maxbeizer/app",
				DocURL:      "https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/reviewing-changes-in-pull-requests",
			},
		},
	}
}

func breakdownWorkflow() workflow {
	return workflow{
		Name:        "breakdown",
		Title:       "🔨 Issue Breakdown",
		Description: "Split large issues into actionable sub-issues with estimates.",
		Steps: []tui.Step{
			{
				Title:       "Preview the Breakdown",
				Description: "See what sub-issues the AI would create without actually creating them.",
				Content:     "Use --dry-run to preview the breakdown. This lets you review\nthe suggested sub-issues before committing to them.",
				Command:     "gh planning breakdown 42 --repo maxbeizer/app --dry-run",
			},
			{
				Title:       "Create Sub-Issues",
				Description: "Run the breakdown for real and create the sub-issues.",
				Content:     "When you're happy with the preview, run without --dry-run.\nSub-issues are created and linked to the parent automatically.",
				Command:     "gh planning breakdown 42 --repo maxbeizer/app",
				DocURL:      "https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-sub-issues",
			},
			{
				Title:       "Estimate the Work",
				Description: "Add effort estimates to your issues.",
				Content:     "Estimates help with sprint planning and capacity forecasting.\nUse T-shirt sizes or story points, depending on your preference.",
				Command:     "gh planning estimate maxbeizer/app#42 --size M",
			},
			{
				Title:       "Check the Roadmap",
				Description: "See how your broken-down work fits into the bigger picture.",
				Content:     "The roadmap view shows your project items on a timeline,\nhelping you visualize dependencies and delivery dates.",
				Command:     "gh planning roadmap",
			},
		},
	}
}
