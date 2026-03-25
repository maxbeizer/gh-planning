package keys

import "charm.land/bubbles/v2/key"

// Global key bindings shared across all TUI views.
type GlobalKeyMap struct {
	Quit       key.Binding
	Help       key.Binding
	Refresh    key.Binding
	NextTab    key.Binding
	PrevTab    key.Binding
	Filter     key.Binding
	ClearFilter key.Binding
	OpenBrowser   key.Binding
	LaunchCopilot key.Binding
}

// NewGlobalKeyMap returns the default global key bindings.
func NewGlobalKeyMap() GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear/back"),
		),
		OpenBrowser: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		LaunchCopilot: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "copilot"),
		),
	}
}

// ActionKeyMap holds item-level action keys.
type ActionKeyMap struct {
	Assign key.Binding
	Label  key.Binding
	Log    key.Binding
}

// NewActionKeyMap returns the default action key bindings.
func NewActionKeyMap() ActionKeyMap {
	return ActionKeyMap{
		Assign: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "assign"),
		),
		Label: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "label"),
		),
		Log: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "log progress"),
		),
	}
}

// NavigationKeyMap holds view-specific navigation keys (j/k/h/l/enter).
type NavigationKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Select key.Binding
}

// NewNavigationKeyMap returns vim-style navigation bindings.
func NewNavigationKeyMap() NavigationKeyMap {
	return NavigationKeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "right"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}
