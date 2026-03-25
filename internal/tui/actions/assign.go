package actions

import (
	"fmt"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

// AssignResultMsg carries the result of an assign mutation.
type AssignResultMsg struct {
	Err  error
	User string
}

// AssignUser shells out to gh api to add an assignee to an issue.
func AssignUser(repo string, number int, user string) tea.Cmd {
	return func() tea.Msg {
		endpoint := fmt.Sprintf("repos/%s/issues/%d/assignees", repo, number)
		arg := fmt.Sprintf("assignees[]=%s", user)
		cmd := exec.Command("gh", "api", endpoint, "--method", "POST", "-f", arg)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return AssignResultMsg{Err: fmt.Errorf("%w: %s", err, string(out)), User: user}
		}
		return AssignResultMsg{User: user}
	}
}
