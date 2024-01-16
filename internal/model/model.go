package model

import (
	"context"
	"strings"

	"github.com/filipecaixeta/logviewer/internal/browse"
	"github.com/filipecaixeta/logviewer/internal/common"
	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/logs"
	"github.com/filipecaixeta/logviewer/internal/source/docker"
	"github.com/filipecaixeta/logviewer/internal/source/fake"
	"github.com/filipecaixeta/logviewer/internal/source/k8s"
	"github.com/filipecaixeta/logviewer/internal/source/stdin"
	"github.com/filipecaixeta/logviewer/internal/state"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	browse         *browse.Model
	logs           *logs.Model
	loadingSpinner spinner.Model
	help           help.Model
	common         *common.Common
	Err            error
}

type keyMap map[string]key.Binding

var keys = keyMap{
	"quit": key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{keys["quit"]}
}

func (k keyMap) FullHelp() [][]key.Binding { return nil }

func New(ctx context.Context, cfg *config.Config) *Model {
	c := common.New(cfg)
	if cfg.Command == "k8s" {
		c.Src = k8s.New(cfg)
	} else if cfg.Command == "docker" {
		c.Src = docker.New()
	} else if cfg.Command == "test" {
		c.Src = fake.New(cfg)
	} else if cfg.Command == "stdin" {
		c.Src = stdin.New(cfg)
	}
	b := browse.New(c)
	c.AddWindowResizeEventListener(b)
	return &Model{
		common: c,
		browse: b,
		loadingSpinner: spinner.New(
			spinner.WithSpinner(spinner.Points),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#ae81ff"))),
		),
		help: help.New(),
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadingSpinner.Tick,
		m.common.Src.Init(m.common.StateChan),
		m.browse.Init(),
		m.common.HandleStateChange(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// fmt.Printf("State: %d, Update %T %v\n", m.common.State, msg, msg)
	switch msg := msg.(type) {
	case error:
		m.Err = msg
		return m, tea.Quit
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys["quit"]):
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		return m, m.common.SetSize(msg)
	case state.State:
		if msg == state.StateLoading && m.common.PrevState != state.StateLoading {
			return m, tea.Batch(m.common.HandleStateChange(), m.loadingSpinner.Tick)
		} else if msg == state.StateLogsLoading && m.common.PrevState != state.StateLogsLoading {
			if _, ok := m.common.Src.(*stdin.Stdin); ok && m.logs != nil {
				m.common.PrevState = m.common.State
				m.common.State = state.StateLogs
				return m, tea.Batch(m.common.HandleStateChange())
			}
			m.logs = logs.New(m.common)
			m.common.AddWindowResizeEventListener(m.logs)
			return m, tea.Batch(m.common.HandleStateChange(), m.logs.Init(), m.loadingSpinner.Tick)
		} else if msg == state.StateLogs {
			_, cmd := m.logs.Update(logs.LogMsg(""))
			return m, tea.Batch(m.common.HandleStateChange(), cmd)
		} else if msg == state.StateBrose && m.common.PrevState == state.StateLogs {
			m.logs.Close()
			return m, m.common.HandleStateChange()
		} else if msg != state.StateNewView && msg != state.StateLoadView {
			return m, m.common.HandleStateChange()
		}
	case spinner.TickMsg:
		if m.common.State != state.StateLoading && m.common.State != state.StateLogsLoading {
			return m, nil
		}
		var cmds tea.Cmd
		m.loadingSpinner, cmds = m.loadingSpinner.Update(msg)
		return m, cmds
	}

	switch m.common.State {
	case state.StateLoading, state.StateLogsLoading:
		return m, nil
	case state.StateBrose:
		_, cmd := m.browse.Update(msg)
		return m, cmd
	case state.StateLogs, state.StateNewView, state.StateLoadView:
		_, cmd := m.logs.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) View() string {
	switch m.common.State {
	case state.StateLoading, state.StateLogsLoading:
		h := m.help.View(keys)
		helpHeight := strings.Count(h, "\n")
		r := m.common.Height - 2 - helpHeight
		if r < 0 {
			r = 0
		}
		loadMsg := " Loading k8s data ..."
		if m.common.State == state.StateLogsLoading {
			loadMsg = " Loading logs ..."
		}
		return m.loadingSpinner.View() + loadMsg + strings.Repeat("\n", r) + h
	case state.StateBrose:
		return m.browse.View()
	case state.StateLogs:
		return m.logs.View()
	}
	return ""
}
