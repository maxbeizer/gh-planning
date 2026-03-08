package state

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type Handoff struct {
	Issue       string    `json:"issue"`
	IssueNumber int       `json:"issueNumber"`
	Repo        string    `json:"repo"`
	SessionID   string    `json:"sessionId"`
	Time        time.Time `json:"time"`
	Done        []string  `json:"done,omitempty"`
	Remaining   []string  `json:"remaining,omitempty"`
	Decisions   []string  `json:"decisions,omitempty"`
	Uncertain   []string  `json:"uncertain,omitempty"`
}

type LogEntry struct {
	Issue     string    `json:"issue"`
	SessionID string    `json:"sessionId"`
	Time      time.Time `json:"time"`
	Message   string    `json:"message"`
	Kind      string    `json:"kind"` // progress, decision, blocker, hypothesis, tried, result
}

type State struct {
	LastSeen time.Time  `json:"lastSeen,omitempty"`
	Handoffs []Handoff  `json:"handoffs,omitempty"`
	Logs     []LogEntry `json:"logs,omitempty"`
}

// dirOverride, when non-empty, replaces the default state directory.
// Used by tests to redirect state I/O to a temp directory.
var dirOverride string

func path() (string, error) {
	if dirOverride != "" {
		return filepath.Join(dirOverride, "state.json"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gh-planning", "state.json"), nil
}

func Load() (*State, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &State{}, nil
		}
		return nil, err
	}
	var st State
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	return &st, nil
}

func Save(st *State) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func UpdateLastSeen(t time.Time) error {
	st, err := Load()
	if err != nil {
		return err
	}
	st.LastSeen = t
	return Save(st)
}

func AddHandoff(h Handoff) error {
	st, err := Load()
	if err != nil {
		return err
	}
	st.Handoffs = append(st.Handoffs, h)
	return Save(st)
}

func AddLog(entry LogEntry) error {
	st, err := Load()
	if err != nil {
		return err
	}
	st.Logs = append(st.Logs, entry)
	return Save(st)
}

func GetLogs(issue string, since time.Time) ([]LogEntry, error) {
	st, err := Load()
	if err != nil {
		return nil, err
	}
	var result []LogEntry
	for _, entry := range st.Logs {
		if issue != "" && entry.Issue != issue {
			continue
		}
		if !since.IsZero() && entry.Time.Before(since) {
			continue
		}
		result = append(result, entry)
	}
	return result, nil
}
