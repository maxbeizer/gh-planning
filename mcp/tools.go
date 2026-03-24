package mcp

import (
	"fmt"
	"sort"
	"strings"
)

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Command     []string               `json:"-"`
	Build       func(map[string]interface{}) ([]string, error) `json:"-"`
}

var tools = []ToolDefinition{
	{
		Name:        "planning-status",
		Description: "Query project status and filters",
		InputSchema: objectSchema(map[string]interface{}{
			"project":   intSchema("Project number"),
			"owner":     stringSchema("Project owner"),
			"stale":     stringSchema("Stale duration (e.g. 7d)"),
			"assignee":  stringSchema("Assignee filter"),
			"board":     boolSchema("Show kanban board view"),
			"swimlanes": boolSchema("Show assignee swimlanes"),
			"exclude":   arraySchema("Statuses to exclude", "string"),
		}),
		Command: []string{"planning", "status"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return buildFlags([]string{"planning", "status"}, args, flagSpec{
				"project":   flagInt("--project"),
				"owner":     flagString("--owner"),
				"stale":     flagString("--stale"),
				"assignee":  flagString("--assignee"),
				"board":     flagBool("--board"),
				"swimlanes": flagBool("--swimlanes"),
				"exclude":   flagRepeat("--exclude"),
			})
		},
	},
	{
		Name:        "planning-standup",
		Description: "Generate a standup report",
		InputSchema: objectSchema(map[string]interface{}{
			"project": intSchema("Project number"),
			"owner":   stringSchema("Project owner"),
			"since":   stringSchema("Lookback duration (e.g. 24h)"),
			"team":    boolSchema("Include team members"),
		}),
		Command: []string{"planning", "standup"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return buildFlags([]string{"planning", "standup"}, args, flagSpec{
				"project": flagInt("--project"),
				"owner":   flagString("--owner"),
				"since":   flagString("--since"),
				"team":    flagBool("--team"),
			})
		},
	},
	{
		Name:        "planning-catchup",
		Description: "Summarize updates since your last session",
		InputSchema: objectSchema(map[string]interface{}{
			"project": intSchema("Project number"),
			"owner":   stringSchema("Project owner"),
			"since":   stringSchema("Duration or date"),
		}),
		Command: []string{"planning", "catch-up"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return buildFlags([]string{"planning", "catch-up"}, args, flagSpec{
				"project": flagInt("--project"),
				"owner":   flagString("--owner"),
				"since":   flagString("--since"),
			})
		},
	},
	{
		Name:        "planning-track",
		Description: "Create an issue and add it to the project",
		InputSchema: objectSchema(map[string]interface{}{
			"title":    stringSchema("Issue title"),
			"repo":     stringSchema("Repository (owner/repo)"),
			"project":  intSchema("Project number"),
			"body":     stringSchema("Issue body"),
			"label":    arraySchema("Label names", "string"),
			"assignee": stringSchema("Assignee"),
			"status":   stringSchema("Initial status"),
		}, "title", "repo"),
		Command: []string{"planning", "track"},
		Build: func(args map[string]interface{}) ([]string, error) {
			title := firstString(args, "title")
			if title == "" {
				return nil, fmt.Errorf("title is required")
			}
			cmdArgs := []string{"planning", "track", title}
			return buildFlags(cmdArgs, args, flagSpec{
				"repo":     flagString("--repo"),
				"project":  flagInt("--project"),
				"body":     flagString("--body"),
				"label":    flagRepeat("--label"),
				"assignee": flagString("--assignee"),
				"status":   flagString("--status"),
			})
		},
	},
	{
		Name:        "planning-team",
		Description: "Show team activity summary",
		InputSchema: objectSchema(map[string]interface{}{
			"team":  stringSchema("Comma-separated team members"),
			"since": stringSchema("Lookback duration"),
			"quiet": boolSchema("Only show inactive teammates"),
		}),
		Command: []string{"planning", "team"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return buildFlags([]string{"planning", "team"}, args, flagSpec{
				"team":  flagString("--team"),
				"since": flagString("--since"),
				"quiet": flagBool("--quiet"),
			})
		},
	},
	{
		Name:        "planning-prep",
		Description: "Generate a 1-1 prep report",
		InputSchema: objectSchema(map[string]interface{}{
			"handle": stringSchema("GitHub handle"),
			"since":  stringSchema("Lookback duration"),
			"notes":  boolSchema("Open or create notes"),
		}, "handle"),
		Command: []string{"planning", "prep"},
		Build: func(args map[string]interface{}) ([]string, error) {
			handle := firstString(args, "handle")
			if handle == "" {
				return nil, fmt.Errorf("handle is required")
			}
			cmdArgs := []string{"planning", "prep", handle}
			return buildFlags(cmdArgs, args, flagSpec{
				"since": flagString("--since"),
				"notes": flagBool("--notes"),
			})
		},
	},
	{
		Name:        "planning-pulse",
		Description: "Show team health metrics",
		InputSchema: objectSchema(map[string]interface{}{
			"team":  stringSchema("Comma-separated team members"),
			"since": stringSchema("Lookback duration"),
		}),
		Command: []string{"planning", "pulse"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return buildFlags([]string{"planning", "pulse"}, args, flagSpec{
				"team":  flagString("--team"),
				"since": flagString("--since"),
			})
		},
	},
	{
		Name:        "planning-review",
		Description: "Summarize review status for a pull request",
		InputSchema: objectSchema(map[string]interface{}{
			"pr":   stringSchema("Pull request number"),
			"repo": stringSchema("Repository (owner/repo)"),
		}, "pr"),
		Command: []string{"planning", "review"},
		Build: func(args map[string]interface{}) ([]string, error) {
			pr := firstString(args, "pr")
			if pr == "" {
				return nil, fmt.Errorf("pr is required")
			}
			cmdArgs := []string{"planning", "review", pr}
			return buildFlags(cmdArgs, args, flagSpec{
				"repo": flagString("--repo"),
			})
		},
	},
	{
		Name:        "planning-focus",
		Description: "Set or show current focus",
		InputSchema: objectSchema(map[string]interface{}{
			"issue": stringSchema("Issue reference (owner/repo#number)"),
		}),
		Command: []string{"planning", "focus"},
		Build: func(args map[string]interface{}) ([]string, error) {
			issue := firstString(args, "issue")
			cmdArgs := []string{"planning", "focus"}
			if issue != "" {
				cmdArgs = append(cmdArgs, issue)
			}
			return cmdArgs, nil
		},
	},
	{
		Name:        "planning-log",
		Description: "Log progress on current focus issue",
		InputSchema: objectSchema(map[string]interface{}{
			"message": stringSchema("Log message"),
			"kind": map[string]interface{}{
				"type":        "string",
				"description": "Log entry kind",
				"enum":        []string{"progress", "decision", "blocker", "hypothesis", "tried", "result"},
			},
		}, "message"),
		Command: []string{"planning", "log"},
		Build: func(args map[string]interface{}) ([]string, error) {
			message := firstString(args, "message")
			if message == "" {
				return nil, fmt.Errorf("message is required")
			}
			cmdArgs := []string{"planning", "log"}
			kind := firstString(args, "kind")
			if kind != "" && kind != "progress" {
				cmdArgs = append(cmdArgs, "--"+kind)
			}
			cmdArgs = append(cmdArgs, message)
			return cmdArgs, nil
		},
	},
	{
		Name:        "planning-logs",
		Description: "View progress log timeline",
		InputSchema: objectSchema(map[string]interface{}{
			"all":   boolSchema("Show all log entries"),
			"since": stringSchema("Show entries since duration (e.g. 1h, 30m)"),
		}),
		Command: []string{"planning", "logs"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return buildFlags([]string{"planning", "logs"}, args, flagSpec{
				"all":   flagBool("--all"),
				"since": flagString("--since"),
			})
		},
	},
	{
		Name:        "planning-board",
		Description: "Show kanban board view of your project",
		InputSchema: objectSchema(map[string]interface{}{
			"project":      intSchema("Project number"),
			"owner":        stringSchema("Project owner"),
			"swimlanes":    boolSchema("Show assignee swimlanes"),
			"include_done": boolSchema("Include completed items"),
		}),
		Command: []string{"planning", "board"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return buildFlags([]string{"planning", "board"}, args, flagSpec{
				"project":      flagInt("--project"),
				"owner":        flagString("--owner"),
				"swimlanes":    flagBool("--swimlanes"),
				"include_done": flagBool("--include-done"),
			})
		},
	},
	{
		Name:        "planning-blocked",
		Description: "Mark an issue as blocked or show blocked items",
		InputSchema: objectSchema(map[string]interface{}{
			"issue":   stringSchema("Issue to mark as blocked"),
			"by":      stringSchema("Blocking issue reference"),
			"repo":    stringSchema("Repository (owner/repo)"),
			"project": intSchema("Project number"),
			"owner":   stringSchema("Project owner"),
		}),
		Command: []string{"planning", "blocked"},
		Build: func(args map[string]interface{}) ([]string, error) {
			issue := firstString(args, "issue")
			cmdArgs := []string{"planning", "blocked"}
			if issue != "" {
				cmdArgs = append(cmdArgs, issue)
			}
			return buildFlags(cmdArgs, args, flagSpec{
				"by":      flagString("--by"),
				"repo":    flagString("--repo"),
				"project": flagInt("--project"),
				"owner":   flagString("--owner"),
			})
		},
	},
	{
		Name:        "planning-profile-show",
		Description: "Show current profile configuration",
		InputSchema: objectSchema(map[string]interface{}{}),
		Command:     []string{"planning", "profile", "show"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return []string{"planning", "profile", "show"}, nil
		},
	},
	{
		Name:        "planning-profile-list",
		Description: "List all configuration profiles",
		InputSchema: objectSchema(map[string]interface{}{}),
		Command:     []string{"planning", "profile", "list"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return []string{"planning", "profile", "list"}, nil
		},
	},
	{
		Name:        "planning-profile-detect",
		Description: "Show which profile matches the current repo",
		InputSchema: objectSchema(map[string]interface{}{}),
		Command:     []string{"planning", "profile", "detect"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return []string{"planning", "profile", "detect"}, nil
		},
	},
	{
		Name:        "planning-cheatsheet",
		Description: "Show a quick-reference of gh-planning commands organized by scenario",
		InputSchema: objectSchema(map[string]interface{}{}),
		Command:     []string{"planning", "cheatsheet"},
		Build: func(args map[string]interface{}) ([]string, error) {
			return []string{"planning", "cheatsheet", "--plain"}, nil
		},
	},
	{
		Name:        "planning-guide",
		Description: "Show a workflow guide for a specific scenario",
		InputSchema: objectSchema(map[string]interface{}{
			"workflow": map[string]interface{}{
				"type":        "string",
				"description": "Workflow name",
				"enum":        []string{"morning", "new-task", "one-on-one"},
			},
		}),
		Command: []string{"planning", "guide"},
		Build: func(args map[string]interface{}) ([]string, error) {
			workflow := firstString(args, "workflow")
			cmdArgs := []string{"planning", "guide"}
			if workflow != "" {
				cmdArgs = append(cmdArgs, workflow)
			}
			cmdArgs = append(cmdArgs, "--plain")
			return cmdArgs, nil
		},
	},
	{
		Name:        "planning-profile-create",
		Description: "Create a new gh-planning profile for a project. Use this when a user is in a repo that has no matching profile, or when they want to set up gh-planning for a new project. Detects the current repo automatically.",
		InputSchema: objectSchema(map[string]interface{}{
			"name":    stringSchema("Profile name (e.g., work, personal, my-project)"),
			"project": intSchema("GitHub Projects V2 project number"),
			"owner":   stringSchema("Project owner (GitHub user or org)"),
			"repos":   stringSchema("Comma-separated repos to auto-detect this profile (e.g., github/github,github/*)"),
			"orgs":    stringSchema("Comma-separated orgs for auto-detection"),
			"team":    stringSchema("Comma-separated team member GitHub usernames"),
			"use":     boolSchema("Switch to the new profile after creating it"),
			"force":   boolSchema("Overwrite existing profile with the same name"),
		}, "name"),
		Command: []string{"planning", "profile", "create"},
		Build: func(args map[string]interface{}) ([]string, error) {
			name := firstString(args, "name")
			if name == "" {
				return nil, fmt.Errorf("name is required")
			}
			cmdArgs := []string{"planning", "profile", "create", name}
			return buildFlags(cmdArgs, args, flagSpec{
				"project": flagInt("--project"),
				"owner":   flagString("--owner"),
				"repos":   flagString("--repos"),
				"orgs":    flagString("--orgs"),
				"team":    flagString("--team"),
				"use":     flagBool("--use"),
				"force":   flagBool("--force"),
			})
		},
	},
	{
		Name:        "planning-profile-update",
		Description: "Update an existing profile. Only the provided fields are changed; other fields are preserved.",
		InputSchema: objectSchema(map[string]interface{}{
			"name":    stringSchema("Profile name to update"),
			"project": intSchema("New project number"),
			"owner":   stringSchema("New project owner"),
			"repos":   stringSchema("New comma-separated repos"),
			"orgs":    stringSchema("New comma-separated orgs"),
			"team":    stringSchema("New comma-separated team members"),
		}, "name"),
		Command: []string{"planning", "profile", "update"},
		Build: func(args map[string]interface{}) ([]string, error) {
			name := firstString(args, "name")
			if name == "" {
				return nil, fmt.Errorf("name is required")
			}
			cmdArgs := []string{"planning", "profile", "update", name}
			return buildFlags(cmdArgs, args, flagSpec{
				"project": flagInt("--project"),
				"owner":   flagString("--owner"),
				"repos":   flagString("--repos"),
				"orgs":    flagString("--orgs"),
				"team":    flagString("--team"),
			})
		},
	},
}

func Tools() []ToolDefinition {
	result := make([]ToolDefinition, len(tools))
	copy(result, tools)
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func ToolByName(name string) (*ToolDefinition, bool) {
	for _, tool := range tools {
		if tool.Name == name {
			return &tool, true
		}
	}
	return nil, false
}

type flagSpec map[string]flagConfig

type flagConfig struct {
	Flag       string
	Repeatable bool
	Boolean    bool
	Type       string
}

func flagString(flag string) flagConfig {
	return flagConfig{Flag: flag, Type: "string"}
}

func flagInt(flag string) flagConfig {
	return flagConfig{Flag: flag, Type: "int"}
}

func flagBool(flag string) flagConfig {
	return flagConfig{Flag: flag, Boolean: true, Type: "bool"}
}

func flagRepeat(flag string) flagConfig {
	return flagConfig{Flag: flag, Repeatable: true, Type: "string"}
}

func buildFlags(base []string, args map[string]interface{}, specs flagSpec) ([]string, error) {
	cmdArgs := append([]string{}, base...)
	for key, spec := range specs {
		value, ok := args[key]
		if !ok || value == nil {
			continue
		}
		if spec.Boolean {
			if boolVal, ok := boolValue(value); ok && boolVal {
				cmdArgs = append(cmdArgs, spec.Flag)
			}
			continue
		}
		if spec.Repeatable {
			items := sliceValue(value)
			for _, item := range items {
				strVal := strings.TrimSpace(fmt.Sprintf("%v", item))
				if strVal == "" {
					continue
				}
				cmdArgs = append(cmdArgs, spec.Flag, strVal)
			}
			continue
		}
		if spec.Type == "int" {
			if intVal, ok := intValue(value); ok {
				cmdArgs = append(cmdArgs, spec.Flag, fmt.Sprintf("%d", intVal))
			}
			continue
		}
		strVal := strings.TrimSpace(fmt.Sprintf("%v", value))
		if strVal != "" {
			cmdArgs = append(cmdArgs, spec.Flag, strVal)
		}
	}
	return cmdArgs, nil
}

func boolValue(value interface{}) (bool, bool) {
	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		lower := strings.ToLower(strings.TrimSpace(v))
		if lower == "true" {
			return true, true
		}
		if lower == "false" {
			return false, true
		}
	}
	return false, false
}

func intValue(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case string:
		if strings.TrimSpace(v) == "" {
			return 0, false
		}
		var parsed int
		_, err := fmt.Sscanf(v, "%d", &parsed)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func sliceValue(value interface{}) []interface{} {
	switch v := value.(type) {
	case []interface{}:
		return v
	case []string:
		items := make([]interface{}, 0, len(v))
		for _, item := range v {
			items = append(items, item)
		}
		return items
	default:
		return []interface{}{v}
	}
}

func firstString(args map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := args[key]; ok {
			str := strings.TrimSpace(fmt.Sprintf("%v", val))
			if str != "" {
				return str
			}
		}
	}
	return ""
}

func objectSchema(properties map[string]interface{}, required ...string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringSchema(description string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": description}
}

func intSchema(description string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": description}
}

func boolSchema(description string) map[string]interface{} {
	return map[string]interface{}{"type": "boolean", "description": description}
}

func arraySchema(description string, itemType string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": description,
		"items": map[string]interface{}{
			"type": itemType,
		},
	}
}
