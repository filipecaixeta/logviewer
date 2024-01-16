package source

import (
	"fmt"
	"io"

	"github.com/filipecaixeta/logviewer/internal/config"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type listDelegate struct {
	IsActive bool
}

func (d listDelegate) Height() int { return 1 }

func (d listDelegate) Spacing() int { return 0 }

func (d listDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d listDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	isSelected := index == m.Index()
	var style lipgloss.Style

	// Cast the item to listItem
	listItem, ok := item.(ListItem)
	if !ok {
		// Handle the error if the item is not of type listItem
		fmt.Fprint(w, "")
		return
	}

	if isSelected && d.IsActive {
		style = config.ListActiveStyle
	} else {
		style = config.ListStyle
	}

	if isSelected {
		fmt.Fprint(w, style.Render("● "+listItem.String()))
	} else {
		fmt.Fprint(w, style.Render("• "+listItem.String()))
	}
}
