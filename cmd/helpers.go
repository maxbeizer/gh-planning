package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func parseDuration(value string) (time.Duration, error) {
	if value == "" {
		return 0, nil
	}
	if strings.HasSuffix(value, "d") {
		days := strings.TrimSuffix(value, "d")
		count, err := strconv.Atoi(days)
		if err != nil {
			return 0, err
		}
		return time.Duration(count) * 24 * time.Hour, nil
	}
	return time.ParseDuration(value)
}

func humanizeDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}
	minutes := int(d.Minutes())
	if minutes < 60 {
		return fmt.Sprintf("%dm ago", minutes)
	}
	hours := int(d.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(d.Hours() / 24)
	if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	weeks := days / 7
	return fmt.Sprintf("%dw ago", weeks)
}

func projectURL(owner string, project int) string {
return fmt.Sprintf("https://github.com/users/%s/projects/%d", owner, project)
}

func openURL(url string) error {
_, err := runCommand("open", url)
return err
}

func runCommand(name string, args ...string) ([]byte, error) {
cmd := exec.Command(name, args...)
return cmd.Output()
}

// hyperlink wraps text in an OSC 8 terminal hyperlink escape sequence.
// Terminals that support it (iTerm2, Windows Terminal, etc.) make the
// text clickable. Only emits escapes when stdout is a terminal.
func hyperlink(url, text string) string {
	if !isTerminal() {
		return text
	}
	return fmt.Sprintf("\x1b]8;;%s\x07%s\x1b]8;;\x07", url, text)
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// issueRef returns a possibly-hyperlinked "#N" string.
func issueRef(number int, url string) string {
ref := fmt.Sprintf("#%d", number)
if url != "" {
return hyperlink(url, ref)
}
return ref
}
