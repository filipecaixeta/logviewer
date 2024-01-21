package logs

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Esc            key.Binding
	Quit           key.Binding
	Up             key.Binding
	Down           key.Binding
	PageTop        key.Binding
	PageEnd        key.Binding
	Filter         key.Binding
	ReturnedFields key.Binding
	Copy           key.Binding
	View           key.Binding
}

var keys = keyMap{
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("ctrl+c/q", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	PageTop: key.NewBinding(
		key.WithKeys("home", "0"),
		key.WithHelp("home/0", "page top"),
	),
	PageEnd: key.NewBinding(
		key.WithKeys("end", "$"),
		key.WithHelp("end/$", "page end"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	ReturnedFields: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "returned fields"),
	),
	Copy: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("right click", "copy"),
	),
	View: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "list views"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ReturnedFields, k.Filter, k.View, k.Copy, k.Esc, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding { return nil }

type textModelKeyMap struct {
	Back key.Binding
	Quit key.Binding
	Save key.Binding
	Run  key.Binding
}

var textModelKeys = textModelKeyMap{
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	Run: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "run"),
	),
}

func (k textModelKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Run, k.Back, k.Quit}
}

func (k textModelKeyMap) FullHelp() [][]key.Binding { return nil }
