package depgraph

import (
	"strings"
	"testing"

	"github.com/maxbeizer/gh-planning/internal/github"
	"github.com/maxbeizer/gh-planning/internal/tui/context"
)

func testCtx() *context.ProgramContext {
	return context.New(nil, "testowner", 1)
}

func TestSetItems_NoDependencies(t *testing.T) {
	m := New(testCtx())
	items := map[string][]github.ProjectItem{
		"Todo": {
			{Number: 1, Title: "Task A", Status: "Todo"},
			{Number: 2, Title: "Task B", Status: "Todo"},
		},
	}
	m.SetItems(items)

	if !m.ready {
		t.Fatal("expected ready after SetItems")
	}
	if len(m.nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(m.nodes))
	}
	// All at level 0 (no blockers).
	for i, n := range m.nodes {
		if n.Level != 0 {
			t.Errorf("node %d: expected level 0, got %d", i, n.Level)
		}
	}
	if m.criticalLen != 0 {
		t.Errorf("expected criticalLen 0, got %d", m.criticalLen)
	}
}

func TestSetItems_LinearChain(t *testing.T) {
	m := New(testCtx())
	// #1 → #2 → #3 (linear dependency chain)
	items := map[string][]github.ProjectItem{
		"Todo": {
			{Number: 1, Title: "Root", Status: "Todo", State: "OPEN"},
			{Number: 2, Title: "Mid", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 1, State: "OPEN"}}},
			{Number: 3, Title: "Leaf", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 2, State: "OPEN"}}},
		},
	}
	m.SetItems(items)

	if len(m.nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(m.nodes))
	}

	// Should be sorted by level: 0, 1, 2
	levels := []int{m.nodes[0].Level, m.nodes[1].Level, m.nodes[2].Level}
	if levels[0] != 0 || levels[1] != 1 || levels[2] != 2 {
		t.Errorf("expected levels [0,1,2], got %v", levels)
	}

	// Critical path should be 3 items deep.
	if m.criticalLen != 3 {
		t.Errorf("expected criticalLen 3, got %d", m.criticalLen)
	}

	// All should be on critical path.
	for i, n := range m.nodes {
		if !n.Critical {
			t.Errorf("node %d (#%d) should be on critical path", i, n.Item.Number)
		}
	}
}

func TestSetItems_ResolvedDependencies(t *testing.T) {
	m := New(testCtx())
	items := map[string][]github.ProjectItem{
		"Done": {
			{Number: 10, Title: "Done task", Status: "Done", State: "CLOSED"},
		},
		"Todo": {
			{Number: 11, Title: "Depends on done", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 10, State: "CLOSED"}}},
		},
	}
	m.SetItems(items)

	// #11 blocked by #10 (closed) → level 1 but not on critical path (dep is resolved).
	if m.criticalLen != 0 {
		t.Errorf("expected criticalLen 0 (resolved dep), got %d", m.criticalLen)
	}
}

func TestSetItems_CycleDetection(t *testing.T) {
	m := New(testCtx())
	// #1 blocks #2, #2 blocks #1 → cycle
	items := map[string][]github.ProjectItem{
		"Todo": {
			{Number: 1, Title: "A", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 2, State: "OPEN"}}},
			{Number: 2, Title: "B", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 1, State: "OPEN"}}},
		},
	}
	m.SetItems(items)

	circular := 0
	for _, n := range m.nodes {
		if n.Circular {
			circular++
		}
	}
	if circular != 2 {
		t.Errorf("expected 2 circular nodes, got %d", circular)
	}
}

func TestSetItems_Empty(t *testing.T) {
	m := New(testCtx())
	m.SetItems(map[string][]github.ProjectItem{})

	if len(m.nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(m.nodes))
	}
}

func TestView_NoDependencyData(t *testing.T) {
	m := New(testCtx())
	// hasDeps is false, no items
	v := m.View()
	if !strings.Contains(v, "No dependency data") {
		t.Errorf("expected graceful degradation message, got: %s", v)
	}
}

func TestView_WithItems(t *testing.T) {
	m := New(testCtx())
	items := map[string][]github.ProjectItem{
		"Todo": {
			{Number: 1, Title: "Root", Status: "Todo", State: "OPEN"},
			{Number: 2, Title: "Blocked", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 1, State: "OPEN"}}},
		},
	}
	m.SetItems(items)
	v := m.View()

	if !strings.Contains(v, "Unblocked") {
		t.Error("expected 'Unblocked' section header")
	}
	if !strings.Contains(v, "Blocked (level 1)") {
		t.Error("expected 'Blocked (level 1)' section header")
	}
	if !strings.Contains(v, "#1") {
		t.Error("expected issue #1 in output")
	}
	if !strings.Contains(v, "#2") {
		t.Error("expected issue #2 in output")
	}
}

func TestVisibleRows(t *testing.T) {
	m := New(testCtx())
	items := map[string][]github.ProjectItem{
		"Todo": {
			{Number: 1, Title: "A", Status: "Todo"},
			{Number: 2, Title: "B", Status: "Todo",
				BlockedBy: []github.DependencyRef{{Number: 1, State: "OPEN"}}},
		},
	}
	m.SetItems(items)

	rows := m.visibleRows()
	// Should have: header(level0), item#1, header(level1), item#2
	headers := 0
	items_count := 0
	for _, r := range rows {
		if r.isHeader {
			headers++
		} else {
			items_count++
		}
	}
	if headers != 2 {
		t.Errorf("expected 2 headers, got %d", headers)
	}
	if items_count != 2 {
		t.Errorf("expected 2 item rows, got %d", items_count)
	}
}

func TestCriticalPathHighlighting(t *testing.T) {
	m := New(testCtx())
	items := map[string][]github.ProjectItem{
		"Todo": {
			{Number: 1, Title: "Root", Status: "Todo", State: "OPEN"},
			{Number: 2, Title: "Mid", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 1, State: "OPEN"}}},
			{Number: 3, Title: "Leaf", Status: "Todo", State: "OPEN",
				BlockedBy: []github.DependencyRef{{Number: 2, State: "OPEN"}}},
		},
	}
	m.SetItems(items)

	v := m.View()
	if !strings.Contains(v, "Critical path: 3 items deep") {
		t.Errorf("expected critical path status, got: %s", v)
	}
	if !strings.Contains(v, "⚡") {
		t.Error("expected ⚡ critical path indicator in output")
	}
}
