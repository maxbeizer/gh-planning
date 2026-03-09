package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Step represents a single step in a multi-step wizard or walkthrough.
type Step struct {
	Title       string
	Description string
	Content     string // main body text
	Command     string // optional command to show/run
	DocURL      string // optional link to GitHub docs
	Done        bool
}

// StepModel is a bubbletea model for navigating multi-step wizards.
type StepModel struct {
	Steps   []Step
	Current int
	Width   int
	Height  int
	Quit    bool

	// OnStep is called when the user advances to a new step.
	// Return a tea.Cmd if you want side effects (e.g., running a command).
	OnStep func(step int) tea.Cmd
}

// NewStepModel creates a StepModel from a slice of steps.
func NewStepModel(steps []Step) StepModel {
	return StepModel{
		Steps:  steps,
		Width:  80,
		Height: 24,
	}
}

func (m StepModel) Init() tea.Cmd {
	return nil
}

func (m StepModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.Quit = true
			return m, tea.Quit
		case "right", "l", "n", "enter":
			if m.Current < len(m.Steps)-1 {
				m.Steps[m.Current].Done = true
				m.Current++
				if m.OnStep != nil {
					return m, m.OnStep(m.Current)
				}
			} else {
				m.Steps[m.Current].Done = true
				m.Quit = true
				return m, tea.Quit
			}
		case "left", "h", "p":
			if m.Current > 0 {
				m.Current--
			}
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m StepModel) View() string {
	if len(m.Steps) == 0 {
		return ""
	}

	step := m.Steps[m.Current]
	var b strings.Builder

	// Progress bar
	progress := fmt.Sprintf("  Step %d of %d", m.Current+1, len(m.Steps))
	b.WriteString(Muted.Render(progress) + "  " + ProgressDots(m.Current, len(m.Steps)))
	b.WriteString("\n\n")

	// Step title
	b.WriteString(Title.Render(step.Title))
	b.WriteString("\n")

	// Description
	if step.Description != "" {
		b.WriteString(Muted.Render(step.Description))
		b.WriteString("\n\n")
	}

	// Content
	if step.Content != "" {
		b.WriteString(step.Content)
		b.WriteString("\n\n")
	}

	// Command box
	if step.Command != "" {
		cmdBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 2).
			Render("$ " + Command.Render(step.Command))
		b.WriteString(cmdBox)
		b.WriteString("\n\n")
	}

	// Doc link
	if step.DocURL != "" {
		b.WriteString(Muted.Render("📖 Docs: ") + lipgloss.NewStyle().Foreground(ColorPrimary).Underline(true).Render(step.DocURL))
		b.WriteString("\n")
	}

	// Navigation help
	nav := "← prev • → next"
	if m.Current == 0 {
		nav = "→ next"
	}
	if m.Current == len(m.Steps)-1 {
		nav = "← prev • enter to finish"
	}
	b.WriteString("\n")
	b.WriteString(HelpBar.Render(nav + " • q quit"))

	return b.String()
}

// RenderStepPlain renders a step as plain text (for --plain / non-TTY).
func RenderStepPlain(step Step, index, total int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== Step %d of %d: %s ===\n\n", index+1, total, step.Title))
	if step.Description != "" {
		b.WriteString(step.Description + "\n\n")
	}
	if step.Content != "" {
		b.WriteString(step.Content + "\n\n")
	}
	if step.Command != "" {
		b.WriteString(fmt.Sprintf("  $ %s\n\n", step.Command))
	}
	if step.DocURL != "" {
		b.WriteString(fmt.Sprintf("Docs: %s\n\n", step.DocURL))
	}
	return b.String()
}
