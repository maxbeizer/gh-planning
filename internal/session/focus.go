package session

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type FocusSession struct {
	Issue       string    `json:"issue"`
	IssueNumber int       `json:"issueNumber"`
	Repo        string    `json:"repo"`
	StartedAt   time.Time `json:"startedAt"`
	SessionID   string    `json:"sessionId"`
}

func path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gh-planning", "sessions", "current.json"), nil
}

func LoadCurrent() (*FocusSession, error) {
	p, err := path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var sess FocusSession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func SaveCurrent(sess *FocusSession) error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func ClearCurrent() error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s FocusSession) Elapsed() time.Duration {
	return time.Since(s.StartedAt)
}
