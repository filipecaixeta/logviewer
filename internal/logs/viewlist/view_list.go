package viewlist

import (
	"fmt"
	"io"

	"github.com/filipecaixeta/logviewer/internal/common"
	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/state"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var footerBorder = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false)

var (
	// CurrentView is the view that is persisted in the config file
	CurrentView *config.View
	// DisplayedView is the view that is currently displayed and could be
	// a modified version of the currentView
	DisplayedView config.View
)

type Model struct {
	common   *common.Common
	Visible  bool
	viewList list.Model
	Height   int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewList.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Up), key.Matches(msg, Keys.Down):
			m.viewList, cmd = m.viewList.Update(msg)
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keys.Esc):
			m.Visible = false
			// case key.Matches(msg, keys.New):
			// 	m.common.SetState(state.StateCreateView)
		case key.Matches(msg, Keys.Select):
			if len(m.viewList.Items()) != 0 {
				CurrentView = m.viewList.SelectedItem().(*config.View)
				DisplayedView = *CurrentView
			}
			m.common.SetState(state.StateLoadView)
			m.Visible = false
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	if !m.Visible {
		return ""
	}
	return footerBorder.Width(m.common.Width).Render(m.viewList.View()) + "\n"
}

func toListItem(items []config.View) []list.Item {
	listItems := make([]list.Item, len(items))
	for i := range items {
		listItems[i] = &items[i]
	}
	return listItems
}

func New(c *common.Common) *Model {
	items := c.Cfg.Views
	l := list.New(toListItem(items), listDelegate{}, 0, 0)
	l.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	l.Styles.TitleBar = lipgloss.NewStyle().Background(lipgloss.Color(""))
	l.Title = config.TitleStyle.Render("Views")
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.SetShowPagination(false)
	m := &Model{
		common:   c,
		viewList: l,
		Height:   7,
	}
	m.viewList.SetHeight(m.Height)
	c.AddWindowResizeEventListener(m)
	return m
}

type listDelegate struct {
}

func (d listDelegate) Height() int { return 1 }

func (d listDelegate) Spacing() int { return 0 }

func (d listDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d listDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	isSelected := index == m.Index()
	var style lipgloss.Style

	// Cast the item to listItem
	listItem, ok := item.(*config.View)
	if !ok {
		// Handle the error if the item is not of type listItem
		fmt.Fprint(w, "")
		return
	}

	if isSelected {
		style = config.ListActiveStyle
	} else {
		style = config.ListStyle
	}

	if isSelected {
		fmt.Fprint(w, style.Render("● "+listItem.Name))
	} else {
		fmt.Fprint(w, style.Render("• "+listItem.Name))
	}
}
