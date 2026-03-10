package theme

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	New    key.Binding
	Launch key.Binding
	Delete key.Binding
	Merge  key.Binding
	Up     key.Binding
	Down   key.Binding
	Quit   key.Binding
	Back   key.Binding
	Enter  key.Binding
	ForceYes key.Binding
	ForceNo  key.Binding
}

var Keys = KeyMap{
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	Launch: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "launch"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "remove"),
	),
	Merge: key.NewBinding(
		key.WithKeys("m"),
		key.WithHelp("m", "merge"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j", "down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	ForceYes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	ForceNo: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
}
