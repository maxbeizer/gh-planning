package actions

import (
	"fmt"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

// LabelResultMsg carries the result of a label mutation.
type LabelResultMsg struct {
	Err   error
	Label string
}

// AddLabel shells out to gh api to add a label to an issue.
func AddLabel(repo string, number int, label string) tea.Cmd {
	return func() tea.Msg {
		endpoint := fmt.Sprintf("repos/%s/issues/%d/labels", repo, number)
		arg := fmt.Sprintf("labels[]=%s", label)
		cmd := exec.Command("gh", "api", endpoint, "--method", "POST", "-f", arg)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return LabelResultMsg{Err: fmt.Errorf("%w: %s", err, string(out)), Label: label}
		}
		return LabelResultMsg{Label: label}
	}
}
