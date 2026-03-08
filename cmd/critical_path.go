package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/maxbeizer/gh-planning/internal/output"
	"github.com/maxbeizer/gh-planning/internal/state"
	"github.com/spf13/cobra"
)

var criticalPathCmd = &cobra.Command{
	Use:   "critical-path",
	Short: "Show the blocking chain and critical path",
	RunE:  runCriticalPath,
}

func runCriticalPath(cmd *cobra.Command, args []string) error {
	st, err := state.Load()
	if err != nil {
		return err
	}
	if len(st.Dependencies) == 0 {
		if OutputOptions().JSON || OutputOptions().JQ != "" {
			return output.PrintJSON(map[string]interface{}{
				"criticalPath":  []string{},
				"depth":         0,
				"blocked":       []interface{}{},
				"impactRanking": []interface{}{},
			}, OutputOptions())
		}
		fmt.Fprintln(cmd.OutOrStdout(), "No dependencies tracked.")
		return nil
	}

	// Build adjacency: blocker -> list of issues it blocks.
	blocks := map[string][]string{}    // blockerRef -> []blockedRef
	blockedBy := map[string]string{}   // blockedRef -> blockerRef
	allNodes := map[string]struct{}{}

	for _, dep := range st.Dependencies {
		blocks[dep.BlockedBy] = append(blocks[dep.BlockedBy], dep.Blocked)
		blockedBy[dep.Blocked] = dep.BlockedBy
		allNodes[dep.Blocked] = struct{}{}
		allNodes[dep.BlockedBy] = struct{}{}
	}

	// Find roots: nodes that block others but are not themselves blocked.
	roots := []string{}
	for node := range allNodes {
		if _, isBlocked := blockedBy[node]; !isBlocked {
			if _, doesBlock := blocks[node]; doesBlock {
				roots = append(roots, node)
			}
		}
	}
	sort.Strings(roots)

	// Build all chains from each root using DFS.
	type chain struct {
		path []string
	}
	var allChains []chain
	var dfs func(node string, current []string, visited map[string]bool)
	dfs = func(node string, current []string, visited map[string]bool) {
		current = append(current, node)
		children := blocks[node]
		if len(children) == 0 {
			ch := make([]string, len(current))
			copy(ch, current)
			allChains = append(allChains, chain{path: ch})
			return
		}
		sort.Strings(children)
		for _, child := range children {
			if visited[child] {
				continue
			}
			visited[child] = true
			dfs(child, current, visited)
			visited[child] = false
		}
	}

	for _, root := range roots {
		visited := map[string]bool{root: true}
		dfs(root, nil, visited)
	}

	// Find the longest chain (critical path).
	var critical []string
	for _, ch := range allChains {
		if len(ch.path) > len(critical) {
			critical = ch.path
		}
	}

	// Compute impact: how many issues each blocker transitively blocks.
	impact := map[string]int{}
	var countBlocked func(node string) int
	countBlocked = func(node string) int {
		total := 0
		for _, child := range blocks[node] {
			total += 1 + countBlocked(child)
		}
		impact[node] = total
		return total
	}
	for _, root := range roots {
		countBlocked(root)
	}

	// Build impact ranking (nodes that unblocking would help the most).
	type impactEntry struct {
		Issue string `json:"issue"`
		Count int    `json:"unblocks"`
	}
	var ranking []impactEntry
	for issue, count := range impact {
		if count > 0 {
			ranking = append(ranking, impactEntry{Issue: issue, Count: count})
		}
	}
	sort.Slice(ranking, func(i, j int) bool {
		if ranking[i].Count != ranking[j].Count {
			return ranking[i].Count > ranking[j].Count
		}
		return ranking[i].Issue < ranking[j].Issue
	})

	if OutputOptions().JSON || OutputOptions().JQ != "" {
		payload := map[string]interface{}{
			"criticalPath":  critical,
			"depth":         len(critical),
			"dependencies":  st.Dependencies,
			"impactRanking": ranking,
		}
		return output.PrintJSON(payload, OutputOptions())
	}

	// Render critical path.
	fmt.Fprintf(cmd.OutOrStdout(), "🔴 Critical Path (%d deep):\n", len(critical))
	for i, node := range critical {
		indent := strings.Repeat("    ", i)
		num := issueNumber(node)
		if i == 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "%s%s\n", indent, num)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s└── blocks %s\n", indent, num)
		}
	}

	// Render other blocked items (chains not on the critical path).
	criticalSet := map[string]bool{}
	for _, n := range critical {
		criticalSet[n] = true
	}

	otherRoots := []string{}
	for _, root := range roots {
		if !criticalSet[root] {
			otherRoots = append(otherRoots, root)
		}
	}

	// Also show non-critical chains from critical roots.
	hasOther := len(otherRoots) > 0
	for _, dep := range st.Dependencies {
		if !criticalSet[dep.Blocked] || !criticalSet[dep.BlockedBy] {
			hasOther = true
			break
		}
	}

	if hasOther {
		fmt.Fprintf(cmd.OutOrStdout(), "\n🟡 Other blocked items:\n")
		printed := map[string]bool{}
		var printTree func(node string, depth int)
		printTree = func(node string, depth int) {
			children := blocks[node]
			sort.Strings(children)
			for _, child := range children {
				if printed[child] {
					continue
				}
				printed[child] = true
				indent := strings.Repeat("    ", depth)
				num := issueNumber(child)
				blockerNum := issueNumber(node)
				if depth == 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", num)
					fmt.Fprintf(cmd.OutOrStdout(), "└── blocked by %s\n", blockerNum)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%s└── blocked by %s\n", indent, blockerNum)
				}
				printTree(child, depth+1)
			}
		}
		for _, root := range otherRoots {
			num := issueNumber(root)
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", num)
			printTree(root, 0)
		}
		// Print non-critical branches from critical roots.
		for _, root := range roots {
			if criticalSet[root] {
				children := blocks[root]
				for _, child := range children {
					if !criticalSet[child] && !printed[child] {
						printed[child] = true
						num := issueNumber(child)
						blockerNum := issueNumber(root)
						fmt.Fprintf(cmd.OutOrStdout(), "%s\n", num)
						fmt.Fprintf(cmd.OutOrStdout(), "└── blocked by %s\n", blockerNum)
						printTree(child, 1)
					}
				}
			}
		}
	}

	// Impact ranking.
	if len(ranking) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "\n⚡ Unblocking impact:\n")
		for _, entry := range ranking {
			num := issueNumber(entry.Issue)
			fmt.Fprintf(cmd.OutOrStdout(), "  %s → would unblock %d issue(s)\n", num, entry.Count)
		}
	}

	return nil
}

// issueNumber extracts "#N" from "owner/repo#N", or returns the ref as-is.
func issueNumber(ref string) string {
	if idx := strings.LastIndex(ref, "#"); idx >= 0 {
		return "#" + ref[idx+1:]
	}
	return ref
}
