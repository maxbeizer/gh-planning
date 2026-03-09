package tutorial

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Progress tracks which tutorial lessons the user has completed.
type Progress struct {
	CompletedLessons []string `yaml:"completed_lessons"`

	// Hands-on tutorial state (persisted across sessions)
	HandsOnIssue    string `yaml:"handson_issue,omitempty"`
	HandsOnRepo     string `yaml:"handson_repo,omitempty"`
	HandsOnIssueNum int    `yaml:"handson_issue_num,omitempty"`
}

func progressPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "gh-planning", "tutorial-progress.yml"), nil
}

// Load reads tutorial progress from disk.
func Load() (*Progress, error) {
	path, err := progressPath()
	if err != nil {
		return &Progress{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Progress{}, nil
		}
		return nil, err
	}

	var p Progress
	if err := yaml.Unmarshal(data, &p); err != nil {
		return &Progress{}, nil
	}
	return &p, nil
}

// Save writes tutorial progress to disk.
func (p *Progress) Save() error {
	path, err := progressPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// IsCompleted checks if a lesson has been completed.
func (p *Progress) IsCompleted(lessonID string) bool {
	for _, id := range p.CompletedLessons {
		if id == lessonID {
			return true
		}
	}
	return false
}

// MarkCompleted marks a lesson as completed.
func (p *Progress) MarkCompleted(lessonID string) {
	if !p.IsCompleted(lessonID) {
		p.CompletedLessons = append(p.CompletedLessons, lessonID)
	}
}

// Reset clears all progress.
func (p *Progress) Reset() {
	p.CompletedLessons = nil
}

// NextIncomplete returns the index of the first incomplete lesson
// given a list of lesson IDs, or -1 if all are complete.
func (p *Progress) NextIncomplete(lessonIDs []string) int {
	for i, id := range lessonIDs {
		if !p.IsCompleted(id) {
			return i
		}
	}
	return -1
}
