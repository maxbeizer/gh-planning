package preps

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type PrepDoc struct {
	Path      string
	Date      time.Time
	Content   string
	FollowUps []string
}

func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "gh-planning", "preps"), nil
}

func Save(handle string, date time.Time, content string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, fmt.Sprintf("%s-%s.md", handle, date.Format("2006-01-02")))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func Latest(handle string) (*PrepDoc, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	prefix := fmt.Sprintf("%s-", handle)
	candidates := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".md") {
			candidates = append(candidates, name)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	sort.Strings(candidates)
	name := candidates[len(candidates)-1]
	path := filepath.Join(dir, name)
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dateStr := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".md")
	parsed, _ := time.Parse("2006-01-02", dateStr)
	followUps := ExtractFollowUps(string(content))
	return &PrepDoc{Path: path, Date: parsed, Content: string(content), FollowUps: followUps}, nil
}

func ExtractFollowUps(content string) []string {
	sections := []string{"🔄 Follow-ups from Last Time", "💬 Suggested Topics"}
	lines := strings.Split(content, "\n")
	for _, section := range sections {
		found := false
		items := []string{}
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, section) {
				found = true
				continue
			}
			if found {
				if line == "" {
					break
				}
				if strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") {
					items = append(items, strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "•"), "-")))
				}
			}
		}
		if len(items) > 0 {
			return items
		}
	}
	return nil
}
