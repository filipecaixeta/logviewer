package common

import (
	"os"

	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/source"
	"github.com/filipecaixeta/logviewer/internal/state"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

type Common struct {
	Width     int
	Height    int
	State     state.State
	PrevState state.State
	StateChan chan state.State
	Models    []tea.Model
	Cfg       *config.Config
	Src       source.Source
}

func New(cfg *config.Config) *Common {
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))
	return &Common{
		Width:     width,
		Height:    height,
		PrevState: state.StateLoading,
		State:     state.StateLoading,
		StateChan: make(chan state.State),
		Cfg:       cfg,
	}
}

func (c *Common) SetState(state state.State) {
	c.StateChan <- state
}

func (m *Common) HandleStateChange() tea.Cmd {
	return func() tea.Msg {
		m.PrevState = m.State
		m.State = <-m.StateChan
		return m.State
	}
}

func (c *Common) SetSize(msg tea.WindowSizeMsg) tea.Cmd {
	c.Width = msg.Width
	c.Height = msg.Height
	var cmds []tea.Cmd
	for _, m := range c.Models {
		_, cmd := m.Update(msg)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (c *Common) AddWindowResizeEventListener(model tea.Model) {
	for _, m := range c.Models {
		if m == model {
			return
		}
	}
	c.Models = append(c.Models, model)
}

func (c *Common) RemoveWindowResizeEventListener(model tea.Model) {
	for i, m := range c.Models {
		if m == model {
			c.Models = append(c.Models[:i], c.Models[i+1:]...)
			return
		}
	}
}
