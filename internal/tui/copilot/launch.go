package copilot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/maxbeizer/gh-planning/internal/github"
)

// DoneMsg is sent when the Copilot CLI session finishes.
type DoneMsg struct {
	Err error
}

// Launch starts a Copilot CLI session with context from the given ProjectItem.
// It writes a temporary markdown context file and passes it via the
// --context-file flag. The TUI is suspended while Copilot runs.
func Launch(item *github.ProjectItem) tea.Cmd {
	contextStr := BuildContext(item)

	// Write context to a temp file so Copilot can ingest it.
	tmpDir := os.TempDir()
	ctxFile := filepath.Join(tmpDir, fmt.Sprintf("gh-planning-copilot-%d.md", item.Number))
	if err := os.WriteFile(ctxFile, []byte(contextStr), 0o600); err != nil {
		return func() tea.Msg {
			return DoneMsg{Err: fmt.Errorf("writing context file: %w", err)}
		}
	}

	c := exec.Command("gh", "copilot", "--context-file", ctxFile)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.ExecProcess(c, func(err error) tea.Msg {
		// Clean up the temp file after Copilot exits.
		os.Remove(ctxFile)
		return DoneMsg{Err: err}
	})
}
