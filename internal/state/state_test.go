package state

import (
	"testing"
	"time"
)

func setup(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	dirOverride = dir
	t.Cleanup(func() { dirOverride = "" })
}

func TestLoad_NoFile(t *testing.T) {
	setup(t)
	st, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(st.Handoffs) != 0 || len(st.Logs) != 0 {
		t.Errorf("expected empty state, got %+v", st)
	}
}

func TestAddLog_AppendsEntries(t *testing.T) {
	setup(t)
	now := time.Now()
	if err := AddLog(LogEntry{Issue: "org/repo#1", Time: now, Message: "first", Kind: "progress"}); err != nil {
		t.Fatal(err)
	}
	if err := AddLog(LogEntry{Issue: "org/repo#1", Time: now.Add(time.Minute), Message: "second", Kind: "decision"}); err != nil {
		t.Fatal(err)
	}
	st, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(st.Logs))
	}
	if st.Logs[0].Message != "first" || st.Logs[1].Message != "second" {
		t.Errorf("log messages = [%q, %q], want [first, second]", st.Logs[0].Message, st.Logs[1].Message)
	}
}

func TestGetLogs_FiltersByIssue(t *testing.T) {
	setup(t)
	now := time.Now()
	entries := []LogEntry{
		{Issue: "org/repo#1", Time: now, Message: "a", Kind: "progress"},
		{Issue: "org/repo#2", Time: now, Message: "b", Kind: "progress"},
		{Issue: "org/repo#1", Time: now, Message: "c", Kind: "decision"},
	}
	for _, e := range entries {
		if err := AddLog(e); err != nil {
			t.Fatal(err)
		}
	}
	logs, err := GetLogs("org/repo#1", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs for issue #1, got %d", len(logs))
	}
	for _, l := range logs {
		if l.Issue != "org/repo#1" {
			t.Errorf("unexpected issue %q", l.Issue)
		}
	}
}

func TestGetLogs_FiltersBySince(t *testing.T) {
	setup(t)
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	entries := []LogEntry{
		{Issue: "x", Time: t0, Message: "old", Kind: "progress"},
		{Issue: "x", Time: t1, Message: "mid", Kind: "progress"},
		{Issue: "x", Time: t2, Message: "new", Kind: "progress"},
	}
	for _, e := range entries {
		if err := AddLog(e); err != nil {
			t.Fatal(err)
		}
	}
	logs, err := GetLogs("", t1)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 2 {
		t.Fatalf("expected 2 logs since t1, got %d", len(logs))
	}
	if logs[0].Message != "mid" || logs[1].Message != "new" {
		t.Errorf("log messages = [%q, %q], want [mid, new]", logs[0].Message, logs[1].Message)
	}
}

func TestGetLogs_BothFilters(t *testing.T) {
	setup(t)
	t0 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	entries := []LogEntry{
		{Issue: "a", Time: t0, Message: "a-old", Kind: "progress"},
		{Issue: "a", Time: t1, Message: "a-new", Kind: "progress"},
		{Issue: "b", Time: t1, Message: "b-new", Kind: "progress"},
	}
	for _, e := range entries {
		if err := AddLog(e); err != nil {
			t.Fatal(err)
		}
	}
	logs, err := GetLogs("a", t1)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].Message != "a-new" {
		t.Errorf("expected [a-new], got %v", logs)
	}
}

func TestAddHandoff_AppendsHandoffs(t *testing.T) {
	setup(t)
	h1 := Handoff{Issue: "org/repo#1", Repo: "org/repo", Done: []string{"task1"}}
	h2 := Handoff{Issue: "org/repo#2", Repo: "org/repo", Done: []string{"task2"}}
	if err := AddHandoff(h1); err != nil {
		t.Fatal(err)
	}
	if err := AddHandoff(h2); err != nil {
		t.Fatal(err)
	}
	st, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Handoffs) != 2 {
		t.Fatalf("expected 2 handoffs, got %d", len(st.Handoffs))
	}
	if st.Handoffs[0].Issue != "org/repo#1" || st.Handoffs[1].Issue != "org/repo#2" {
		t.Errorf("handoff issues = [%q, %q]", st.Handoffs[0].Issue, st.Handoffs[1].Issue)
	}
}

func TestSave_Roundtrip(t *testing.T) {
	setup(t)
	now := time.Now().Truncate(time.Second)
	original := &State{
		LastSeen: now,
		Logs:     []LogEntry{{Issue: "x", Time: now, Message: "test", Kind: "progress"}},
		Handoffs: []Handoff{{Issue: "x", Repo: "r", Done: []string{"a"}}},
	}
	if err := Save(original); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.LastSeen.Equal(original.LastSeen) {
		t.Errorf("LastSeen = %v, want %v", loaded.LastSeen, original.LastSeen)
	}
	if len(loaded.Logs) != 1 || loaded.Logs[0].Message != "test" {
		t.Errorf("Logs = %+v", loaded.Logs)
	}
	if len(loaded.Handoffs) != 1 || loaded.Handoffs[0].Repo != "r" {
		t.Errorf("Handoffs = %+v", loaded.Handoffs)
	}
}
