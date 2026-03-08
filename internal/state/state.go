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

type State struct {
	LastSeen time.Time `json:"lastSeen,omitempty"`
	Handoffs []Handoff `json:"handoffs,omitempty"`
}

func path() (string, error) {
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
