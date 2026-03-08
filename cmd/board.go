package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/maxbeizer/gh-planning/internal/github"
	"golang.org/x/term"
)

const maxCardsPerColumn = 15

// statusOrder defines the preferred column ordering for the board.
var statusOrder = []string{
	"backlog", "ready", "todo",
	"in progress", "in review", "needs review", "needs my attention",
	"done", "closed", "complete", "completed",
}

func statusRank(status string) int {
	lower := strings.ToLower(status)
	for i, s := range statusOrder {
		if s == lower {
			return i
		}
	}
	return len(statusOrder)
}

func sortedStatuses(groups map[string][]github.ProjectItem) []string {
	statuses := make([]string, 0, len(groups))
	for status := range groups {
		statuses = append(statuses, status)
	}
	sort.Slice(statuses, func(i, j int) bool {
		ri, rj := statusRank(statuses[i]), statusRank(statuses[j])
		if ri != rj {
			return ri < rj
		}
		return statuses[i] < statuses[j]
	})
	return statuses
}

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 120
	}
	return w
}

func statusEmoji(status string) string {
	switch strings.ToLower(status) {
	case "in progress":
		return "🔵"
	case "backlog", "ready", "todo":
		return "📋"
	case "done", "closed", "complete", "completed":
		return "✅"
	case "in review", "needs review", "needs my attention":
		return "🔍"
	case "blocked":
		return "🚫"
	default:
		return "•"
	}
}

// printBoardView renders a kanban board with columns per status.
func printBoardView(groups map[string][]github.ProjectItem) {
	statuses := sortedStatuses(groups)
	if len(statuses) == 0 {
		fmt.Println("  (no items)")
		return
	}

	width := termWidth()
	numCols := len(statuses)

	// Cap column width so single-column boards don't stretch edge-to-edge
	maxColWidth := 50
	colWidth := width / numCols
	if colWidth > maxColWidth {
		colWidth = maxColWidth
	}
	if colWidth < 24 {
		colWidth = 24
	}
	innerWidth := colWidth - 3

	// Build columns: header + compact cards
	columns := make([][]string, numCols)
	for i, status := range statuses {
		items := groups[status]
		header := fmt.Sprintf("%s %s (%d)", statusEmoji(status), status, len(items))
		columns[i] = append(columns[i], padOrTruncate(header, innerWidth))

		capped := items
		overflow := 0
		if len(capped) > maxCardsPerColumn {
			overflow = len(capped) - maxCardsPerColumn
			capped = capped[:maxCardsPerColumn]
		}

		for _, item := range capped {
			assignee := ""
			if len(item.Assignees) > 0 {
				assignee = "@" + item.Assignees[0]
			}
			card := fmt.Sprintf("#%d %s", item.Number, item.Title)
			columns[i] = append(columns[i], padOrTruncate(card, innerWidth))
			if assignee != "" {
				columns[i] = append(columns[i], padOrTruncate("   "+assignee, innerWidth))
			}
		}
		if overflow > 0 {
			columns[i] = append(columns[i], padOrTruncate(fmt.Sprintf("   ... +%d more", overflow), innerWidth))
		}
	}

	// Find max rows
	maxRows := 0
	for _, col := range columns {
		if len(col) > maxRows {
			maxRows = len(col)
		}
	}

	// Top border
	topParts := make([]string, numCols)
	for i := range topParts {
		topParts[i] = strings.Repeat("─", innerWidth+2)
	}
	fmt.Printf("┌%s┐\n", strings.Join(topParts, "┬"))

	// Header row
	headerParts := make([]string, numCols)
	for col := range columns {
		headerParts[col] = " " + columns[col][0] + " "
	}
	fmt.Printf("│%s│\n", strings.Join(headerParts, "│"))

	// Header separator
	sepParts := make([]string, numCols)
	for i := range sepParts {
		sepParts[i] = strings.Repeat("─", innerWidth+2)
	}
	fmt.Printf("├%s┤\n", strings.Join(sepParts, "┼"))

	// Card rows (skip row 0, that was the header)
	for row := 1; row < maxRows; row++ {
		parts := make([]string, numCols)
		for col := range columns {
			if row < len(columns[col]) {
				parts[col] = " " + columns[col][row] + " "
			} else {
				parts[col] = " " + strings.Repeat(" ", innerWidth) + " "
			}
		}
		fmt.Printf("│%s│\n", strings.Join(parts, "│"))
	}

	// Bottom border
	bottomParts := make([]string, numCols)
	for i := range bottomParts {
		bottomParts[i] = strings.Repeat("─", innerWidth+2)
	}
	fmt.Printf("└%s┘\n", strings.Join(bottomParts, "┴"))
}

// printSwimlaneBoardView renders a kanban board with swimlanes per assignee.
func printSwimlaneBoardView(groups map[string][]github.ProjectItem) {
	statuses := sortedStatuses(groups)
	if len(statuses) == 0 {
		fmt.Println("  (no items)")
		return
	}

	// Collect all assignees
	assigneeSet := map[string]bool{}
	for _, items := range groups {
		for _, item := range items {
			if len(item.Assignees) > 0 {
				assigneeSet[item.Assignees[0]] = true
			} else {
				assigneeSet["unassigned"] = true
			}
		}
	}
	assignees := make([]string, 0, len(assigneeSet))
	for a := range assigneeSet {
		assignees = append(assignees, a)
	}
	sort.Strings(assignees)
	// Put "unassigned" last
	for i, a := range assignees {
		if a == "unassigned" {
			assignees = append(assignees[:i], assignees[i+1:]...)
			assignees = append(assignees, "unassigned")
			break
		}
	}

	width := termWidth()
	numCols := len(statuses)
	maxColWidth := 50
	colWidth := width / numCols
	if colWidth > maxColWidth {
		colWidth = maxColWidth
	}
	if colWidth < 24 {
		colWidth = 24
	}
	innerWidth := colWidth - 3

	// Top border + column headers
	topParts := make([]string, numCols)
	for i := range topParts {
		topParts[i] = strings.Repeat("─", innerWidth+2)
	}
	fmt.Printf("┌%s┐\n", strings.Join(topParts, "┬"))

	headerParts := make([]string, numCols)
	for i, status := range statuses {
		header := fmt.Sprintf("%s %s (%d)", statusEmoji(status), status, len(groups[status]))
		headerParts[i] = " " + padOrTruncate(header, innerWidth) + " "
	}
	fmt.Printf("│%s│\n", strings.Join(headerParts, "│"))

	sepParts := make([]string, numCols)
	for i := range sepParts {
		sepParts[i] = strings.Repeat("═", innerWidth+2)
	}
	fmt.Printf("╞%s╡\n", strings.Join(sepParts, "╪"))

	// Print each swimlane
	for laneIdx, assignee := range assignees {
		// Swimlane header
		label := "👤 @" + assignee
		if assignee == "unassigned" {
			label = "👤 unassigned"
		}
		laneLine := make([]string, numCols)
		for i := range laneLine {
			if i == 0 {
				laneLine[i] = " " + padOrTruncate(label, innerWidth) + " "
			} else {
				laneLine[i] = " " + strings.Repeat(" ", innerWidth) + " "
			}
		}
		fmt.Printf("│%s│\n", strings.Join(laneLine, "│"))

		// Items for this assignee in each column
		columns := make([][]string, numCols)
		for colIdx, status := range statuses {
			count := 0
			total := 0
			for _, item := range groups[status] {
				itemAssignee := "unassigned"
				if len(item.Assignees) > 0 {
					itemAssignee = item.Assignees[0]
				}
				if itemAssignee != assignee {
					continue
				}
				total++
				if count >= maxCardsPerColumn {
					continue
				}
				card := fmt.Sprintf("  #%d %s", item.Number, item.Title)
				columns[colIdx] = append(columns[colIdx], padOrTruncate(card, innerWidth))
				count++
			}
			if total > maxCardsPerColumn {
				columns[colIdx] = append(columns[colIdx], padOrTruncate(fmt.Sprintf("  ... +%d more", total-maxCardsPerColumn), innerWidth))
			}
		}

		maxRows := 0
		for _, col := range columns {
			if len(col) > maxRows {
				maxRows = len(col)
			}
		}
		if maxRows == 0 {
			maxRows = 1
			for i := range columns {
				columns[i] = append(columns[i], padOrTruncate("  (none)", innerWidth))
			}
		}

		for row := 0; row < maxRows; row++ {
			parts := make([]string, numCols)
			for col := range columns {
				if row < len(columns[col]) {
					parts[col] = " " + columns[col][row] + " "
				} else {
					parts[col] = " " + strings.Repeat(" ", innerWidth) + " "
				}
			}
			fmt.Printf("│%s│\n", strings.Join(parts, "│"))
		}

		// Separator between swimlanes (not after last)
		if laneIdx < len(assignees)-1 {
			divParts := make([]string, numCols)
			for i := range divParts {
				divParts[i] = strings.Repeat("─", innerWidth+2)
			}
			fmt.Printf("├%s┤\n", strings.Join(divParts, "┼"))
		}
	}

	// Bottom border
	bottomParts := make([]string, numCols)
	for i := range bottomParts {
		bottomParts[i] = strings.Repeat("─", innerWidth+2)
	}
	fmt.Printf("└%s┘\n", strings.Join(bottomParts, "┴"))
}

func padOrTruncate(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw > width {
		return runewidth.Truncate(s, width, "…")
	}
	return s + strings.Repeat(" ", width-sw)
}
