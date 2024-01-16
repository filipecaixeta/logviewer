package browse

import (
	"strings"

	"github.com/filipecaixeta/logviewer/internal/common"
	"github.com/filipecaixeta/logviewer/internal/source"
	"github.com/filipecaixeta/logviewer/internal/state"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const footerHeight = 3

var (
	columnSpace = lipgloss.NewStyle().PaddingLeft(4).Render("")
)

type Model struct {
	columns      []*source.List
	ActiveColumn int
	help         help.Model
	common       *common.Common
}

func New(c *common.Common) *Model {
	m := &Model{
		columns: c.Src.Columns(),
		help:    help.New(),
		common:  c,
	}
	m.UpdateChildren()
	m.columns[m.ActiveColumn].SetActive(true)

	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		for _, col := range m.columns {
			col.SetSize(msg.Width, msg.Height-footerHeight)
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up), key.Matches(msg, keys.Down):
			m.columns[m.ActiveColumn], cmd = m.columns[m.ActiveColumn].Update(msg)
			m.UpdateChildren()
		case key.Matches(msg, keys.Left):
			if m.ActiveColumn > 0 {
				m.setActiveColumn(m.ActiveColumn - 1)
			} else {
				m.setActiveColumn(m.ActiveColumn)
			}
		case key.Matches(msg, keys.Right):
			if m.ActiveColumn < len(m.columns)-1 {
				m.setActiveColumn(m.ActiveColumn + 1)
			} else {
				m.setActiveColumn(m.ActiveColumn)
			}
			m.columns[m.ActiveColumn], cmd = m.columns[m.ActiveColumn].Update(msg)
			m.UpdateChildren()
		case key.Matches(msg, keys.Log):
			m.common.SetState(state.StateLogsLoading)
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		}
	}

	return m, cmd
}

func (m *Model) View() string {
	h := m.help.View(keys)
	helpHeight := strings.Count(h, "\n")

	if len(m.columns) == 1 {
		listView := m.columns[0].View()
		r := m.common.Height - 2 - helpHeight - strings.Count(listView, "\n")
		if r < 0 {
			r = 0
		}
		return listView + strings.Repeat("\n", r) + h
	}

	col := m.ActiveColumn
	if col == len(m.columns)-1 {
		col--
	}
	list1View := m.columns[col].View()
	list2View := m.columns[col+1].View()

	columnsView := lipgloss.JoinHorizontal(lipgloss.Top, list1View, columnSpace, list2View)

	r := m.common.Height - 2 - helpHeight - strings.Count(columnsView, "\n")
	if r < 0 {
		r = 0
	}

	return columnsView + strings.Repeat("\n", r) + h
}

func (m *Model) UpdateChildren() {
	for i := m.ActiveColumn; i < len(m.columns)-1; i++ {
		if item := m.columns[i].SelectedItem(); item != nil {
			m.columns[i+1].SetItems(item.Children())
		} else {
			m.columns[i+1].SetItems([]source.ListItem{})
		}
		m.columns[i+1].ResetFilter()
		m.columns[i+1].ResetSelected()
	}
}

func (m *Model) setActiveColumn(i int) {
	m.ActiveColumn = i
	for j := 0; j < len(m.columns); j++ {
		m.columns[j].SetActive(false)
	}
	m.columns[m.ActiveColumn].SetActive(true)
}
