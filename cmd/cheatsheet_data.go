package cmd

import "github.com/maxbeizer/gh-planning/internal/tui"

func cheatsheetItems() []tui.ListItem {
	return []tui.ListItem{
		// Morning Routine
		{
			Category:    "🌅 Morning Routine",
			Title:       "Catch up on what you missed",
			Command:     "gh planning catch-up",
			Description: "Summarize project activity since your last session or a given time.",
			Example:     "gh planning catch-up --since friday",
			DocURL:      "https://docs.github.com/en/issues/planning-and-tracking-with-projects",
		},
		{
			Category:    "🌅 Morning Routine",
			Title:       "Generate a standup report",
			Command:     "gh planning standup",
			Description: "Auto-generate what you did, what's next, and blockers.",
			Example:     "gh planning standup --since 24h",
		},
		{
			Category:    "🌅 Morning Routine",
			Title:       "View your project board",
			Command:     "gh planning status",
			Description: "See all items grouped by status. Add --board for kanban view.",
			Example:     "gh planning status --board",
		},
		{
			Category:    "🌅 Morning Routine",
			Title:       "Open the kanban board",
			Command:     "gh planning board",
			Description: "Kanban view with columns per status. Excludes Done by default.",
			Example:     "gh planning board --swimlanes",
		},

		// Starting a Task
		{
			Category:    "🎯 Starting a Task",
			Title:       "Claim an issue",
			Command:     "gh planning claim",
			Description: "Assign yourself and move the issue to In Progress in one step.",
			Example:     "gh planning claim maxbeizer/app#42",
			DocURL:      "https://docs.github.com/en/issues/tracking-your-work-with-issues",
		},
		{
			Category:    "🎯 Starting a Task",
			Title:       "Focus on an issue",
			Command:     "gh planning focus",
			Description: "Set your active focus issue. Tracks elapsed time automatically.",
			Example:     "gh planning focus maxbeizer/app#42",
		},
		{
			Category:    "🎯 Starting a Task",
			Title:       "Create & track a new issue",
			Command:     "gh planning track",
			Description: "Create an issue and add it to your project board in one command.",
			Example:     `gh planning track "Fix auth bug" --repo maxbeizer/app --status "In Progress"`,
		},
		{
			Category:    "🎯 Starting a Task",
			Title:       "Find work to pick up",
			Command:     "gh planning queue",
			Description: "Show items ready for processing, filtered by label or status.",
			Example:     "gh planning queue --label agent-ready --status Backlog",
		},

		// During Work
		{
			Category:    "📝 During Work",
			Title:       "Log progress",
			Command:     "gh planning log",
			Description: "Record progress, decisions, blockers, or findings against your focus issue.",
			Example:     `gh planning log "OAuth callback working"`,
		},
		{
			Category:    "📝 During Work",
			Title:       "Log a decision",
			Command:     "gh planning log --decision",
			Description: "Record an architectural or design decision.",
			Example:     `gh planning log --decision "Using JWT for stateless auth"`,
		},
		{
			Category:    "📝 During Work",
			Title:       "Log a blocker",
			Command:     "gh planning log --blocker",
			Description: "Record what's blocking you so it shows up in standups.",
			Example:     `gh planning log --blocker "Need API key from team lead"`,
		},
		{
			Category:    "📝 During Work",
			Title:       "View your progress log",
			Command:     "gh planning logs",
			Description: "Show the timeline of logged progress entries.",
			Example:     "gh planning logs --all --since 7d",
		},
		{
			Category:    "📝 During Work",
			Title:       "Break down a large issue",
			Command:     "gh planning breakdown",
			Description: "Use AI to split a big issue into actionable sub-issues.",
			Example:     "gh planning breakdown 42 --repo maxbeizer/app --dry-run",
			DocURL:      "https://docs.github.com/en/issues/tracking-your-work-with-issues/using-issues/creating-sub-issues",
		},
		{
			Category:    "📝 During Work",
			Title:       "Review a pull request",
			Command:     "gh planning review",
			Description: "Quick summary of a PR — files changed, review status, CI checks.",
			Example:     "gh planning review 48 --repo maxbeizer/app",
			DocURL:      "https://docs.github.com/en/pull-requests",
		},

		// Collaboration
		{
			Category:    "🤝 Collaboration",
			Title:       "Team activity dashboard",
			Command:     "gh planning team",
			Description: "See what your teammates have been working on recently.",
			Example:     "gh planning team --since 7d",
		},
		{
			Category:    "🤝 Collaboration",
			Title:       "Prepare for a 1-1",
			Command:     "gh planning prep",
			Description: "Auto-generate a 1-1 preparation doc from project activity.",
			Example:     "gh planning prep maxbeizer --since 14d",
		},
		{
			Category:    "🤝 Collaboration",
			Title:       "Team health metrics",
			Command:     "gh planning pulse",
			Description: "Show velocity, throughput, and staleness metrics for your team.",
			Example:     "gh planning pulse --since 30d",
		},
		{
			Category:    "🤝 Collaboration",
			Title:       "Hand off to the next session",
			Command:     "gh planning handoff",
			Description: "Post a structured handoff comment to an issue for the next person.",
			Example:     `gh planning handoff maxbeizer/app#42 --done "OAuth flow" --remaining "Logout"`,
		},

		// AI & Agents
		{
			Category:    "🤖 AI & Agents",
			Title:       "Get agent context",
			Command:     "gh planning agent-context",
			Description: "Summarize everything an AI agent needs to start work on your project.",
			Example:     "gh planning agent-context --new-session",
		},
		{
			Category:    "🤖 AI & Agents",
			Title:       "View agent work queue",
			Command:     "gh planning queue",
			Description: "Show items labeled for agent processing.",
			Example:     "gh planning queue --label agent-ready",
		},
		{
			Category:    "🤖 AI & Agents",
			Title:       "Break down with AI",
			Command:     "gh planning breakdown",
			Description: "Use GitHub Models to split issues into sub-issues automatically.",
			Example:     "gh planning breakdown 42 --repo maxbeizer/app",
		},
		{
			Category:    "🤖 AI & Agents",
			Title:       "Start MCP server for Copilot",
			Command:     "gh planning copilot serve",
			Description: "Launch the MCP server so Copilot can use gh-planning as a tool.",
			Example:     "gh planning copilot serve",
		},

		// Wrapping Up
		{
			Category:    "🏁 Wrapping Up",
			Title:       "Complete an issue",
			Command:     "gh planning complete",
			Description: "Post a completion handoff and move the issue forward.",
			Example:     `gh planning complete maxbeizer/app#42 --done "OAuth flow" --pr 48`,
		},
		{
			Category:    "🏁 Wrapping Up",
			Title:       "Clear your focus",
			Command:     "gh planning unfocus",
			Description: "End your focus session, optionally with a wrap-up comment.",
			Example:     `gh planning unfocus --comment "Wrapped this up"`,
		},
		{
			Category:    "🏁 Wrapping Up",
			Title:       "Generate end-of-day standup",
			Command:     "gh planning standup",
			Description: "Summarize your day for async standups or journals.",
			Example:     "gh planning standup --since 8h",
		},

		// Configuration
		{
			Category:    "⚙️  Configuration",
			Title:       "Interactive setup wizard",
			Command:     "gh planning setup",
			Description: "Guided walkthrough to configure your project, team, and preferences.",
			Example:     "gh planning setup",
		},
		{
			Category:    "⚙️  Configuration",
			Title:       "Set a config value",
			Command:     "gh planning config set",
			Description: "Set configuration keys: default-project, team, 1-1-repo-pattern, etc.",
			Example:     "gh planning config set team maxbeizer,claudia-bot",
		},
		{
			Category:    "⚙️  Configuration",
			Title:       "Switch config profiles",
			Command:     "gh planning config use",
			Description: "Switch between named config profiles (e.g., work vs personal).",
			Example:     "gh planning config use work",
		},
		{
			Category:    "⚙️  Configuration",
			Title:       "Show current config",
			Command:     "gh planning config show",
			Description: "Display all configuration values.",
			Example:     "gh planning config show",
		},
	}
}
