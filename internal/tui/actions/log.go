package actions

import (
	"fmt"
	"os/exec"

	tea "charm.land/bubbletea/v2"
)

// LogResultMsg carries the result of a progress log command.
type LogResultMsg struct {
	Err error
}

// LogProgress shells out to gh planning log to record progress on an issue.
func LogProgress(message string, number int) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("gh", "planning", "log",
			"--message", message,
			"--issue", fmt.Sprintf("%d", number),
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return LogResultMsg{Err: fmt.Errorf("%w: %s", err, string(out))}
		}
		return LogResultMsg{}
	}
}
