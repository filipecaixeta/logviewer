package logs

import (
	"context"
	"fmt"
	"strings"

	"github.com/filipecaixeta/logviewer/internal/common"
	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/logs/pipeline"
	"github.com/filipecaixeta/logviewer/internal/logs/viewlist"
	"github.com/filipecaixeta/logviewer/internal/state"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"
)

const (
	helpHeight      = 2
	textModelHeight = 3
)

type Model struct {
	// lChan is the channel that receives log messages from the streaming API
	lChan      chan string
	logEntries circularLogBuffer

	scrollOffset int
	maxScroll    int
	autoScroll   bool
	width        int

	textareaTitle string
	textModel     textarea.Model
	help          help.Model

	viewList *viewlist.Model

	pipeline *pipeline.LogPipeline

	common *common.Common
	ctx    context.Context
	cancel context.CancelFunc
}

type LogMsg string

func New(c *common.Common) *Model {
	m := &Model{
		lChan:      make(chan string),
		logEntries: circularLogBuffer{Buffer: make([]pipeline.LogEntry, 0, 20000)},
		help:       help.New(),
		common:     c,
		autoScroll: true,
		textModel:  textarea.New(),
		viewList:   viewlist.New(c),
	}
	c.AddWindowResizeEventListener(m)

	var err error
	m.pipeline, err = pipeline.New(nil, uint(c.Width))
	if err != nil {
		panic(err)
	}

	m.textModel.SetWidth(m.common.Width)
	m.textModel.SetHeight(textModelHeight)
	m.textModel.Blur()

	return m
}

func (m *Model) Init() tea.Cmd {
	ctx := context.Background()
	m.ctx, m.cancel = context.WithCancel(ctx)
	return m.common.Src.Logs(ctx, m.common.StateChan, m.lChan)
}

func (m *Model) Close() {
	m.cancel()
}

func (m *Model) updateFilterTextModel(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, textModelKeys.Save):
		m.textModel.Blur()
		if err := m.pipeline.SetFilter(m.textModel.Value()); err != nil {
			fmt.Printf("err: %v\n", err)
		}
		m.logEntries.RunPipeline(m.pipeline.RunFilterChanged)
		return nil
	case key.Matches(msg, textModelKeys.Run):
		if err := m.pipeline.SetFilter(m.textModel.Value()); err != nil {
			fmt.Printf("err: %v\n", err)
		}
		m.logEntries.RunPipeline(m.pipeline.RunFilterChanged)
		return nil
	case key.Matches(msg, textModelKeys.Cancel):
		m.textModel.Blur()
	default:
		var cmd tea.Cmd
		m.textModel, cmd = m.textModel.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) updateReturnedFieldsTextModel(msg tea.KeyMsg) tea.Cmd {
	returnedFields := strings.Split(m.textModel.Value(), ",")

	switch {
	case key.Matches(msg, textModelKeys.Save):
		m.textModel.Blur()
		if err := m.pipeline.SetReturnedFields(returnedFields); err != nil {
			fmt.Printf("err: %v\n", err)
		}
		m.logEntries.RunPipeline(m.pipeline.RunReturnedFieldsChanged)
		return nil
	case key.Matches(msg, textModelKeys.Run):
		if err := m.pipeline.SetReturnedFields(returnedFields); err != nil {
			fmt.Printf("err: %v\n", err)
		}
		m.logEntries.RunPipeline(m.pipeline.RunReturnedFieldsChanged)
		return nil
	case key.Matches(msg, textModelKeys.Cancel):
		m.textModel.Blur()
	default:
		var cmd tea.Cmd
		m.textModel, cmd = m.textModel.Update(msg)
		return cmd
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.textModel.Focused() && m.textareaTitle == "Filter" {
			return m, m.updateFilterTextModel(msg)
		} else if m.textModel.Focused() && m.textareaTitle == "Returned Fields" {
			return m, m.updateReturnedFieldsTextModel(msg)
		}
		if m.viewList.Visible {
			var cmd tea.Cmd
			_, cmd = m.viewList.Update(msg)
			return m, cmd
		}
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Esc):
			m.common.SetState(state.StateBrose)
			return m, nil
		case key.Matches(msg, keys.Up):
			m.scrollUp(1)
		case key.Matches(msg, keys.Down):
			m.scrollDown(1)
		case key.Matches(msg, keys.PageTop):
			m.scrollUp(1<<31 - 1)
		case key.Matches(msg, keys.PageEnd):
			m.scrollDown(1<<31 - 1)
		case key.Matches(msg, keys.Filter):
			m.textModel.Focus()
			m.textareaTitle = "Filter"
			m.textModel.SetValue(m.pipeline.Cfg.Filter)
			return m, textarea.Blink
		case key.Matches(msg, keys.ReturnedFields):
			m.textModel.Focus()
			m.textareaTitle = "Returned Fields"
			m.textModel.SetValue(strings.Join(m.pipeline.Cfg.ReturnedFields, ", "))
			return m, textarea.Blink
		case key.Matches(msg, keys.View):
			m.viewList.Visible = true
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			m.scrollUp(1)
		case tea.MouseButtonWheelDown:
			m.scrollDown(1)
		case tea.MouseButtonRight:
			m.copyToClipboard(msg.Y)
		}
	case tea.WindowSizeMsg:
		if msg.Width != m.width {
			m.width = msg.Width
			m.textModel.SetWidth(m.common.Width)
			_ = m.pipeline.SetWidth(uint(msg.Width))
			m.logEntries.RunPipeline(m.pipeline.RunWidthChanged)
		}
		return m, nil
	case state.State:
		switch msg {
		case state.StateLoadView:
			view := m.viewList.SelectedView()
			if view != nil {
				if err := m.pipeline.SetView(view); err != nil {
					fmt.Printf("err: %v\n", err)
				}
				m.logEntries.RunPipeline(m.pipeline.RunViewChanged)
				m.common.State = state.StateLogs
			}
			return m, m.common.HandleStateChange()
		}
	case LogMsg:
		if msg != "" {
			m.handleLogMsg(msg)
		}
		return m, m.handleLogEntry()
	}
	return m, nil
}

func (m *Model) copyToClipboard(y int) {
	l := m.logEntries.GetLogEntryAtScrollOffset(m.scrollOffset + y)
	err := clipboard.Init()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	if err == nil && l != nil {
		lCopy := *l
		returnedFields := []string{}
		if view := m.viewList.SelectedView(); view != nil {
			returnedFields = view.ReturnedFields
		}
		t := pipeline.LogFormat{
			ReturnedFields: returnedFields,
		}
		_ = t.RunReturnedFieldsAndFormat(&lCopy)
		clipboard.Write(clipboard.FmtText, []byte(lCopy.Formatted))
	}
}

func (m *Model) scrollUp(n int) {
	m.autoScroll = false
	var minOffset int
	if f := m.logEntries.First(); f != nil {
		minOffset = f.CumHeight - f.Height
	}
	m.scrollOffset = max(minOffset, m.scrollOffset-n)
}

func (m *Model) scrollDown(n int) {
	maxScroll := m.logEntries.Height() - m.common.Height + helpHeight
	if m.scrollOffset+n >= maxScroll {
		m.autoScroll = true
	}
	m.scrollOffset = min(maxScroll, m.scrollOffset+n)
}

func (m *Model) View() string {
	var helpView string

	height := m.common.Height - helpHeight

	var footerView string
	if m.textModel.Focused() {
		footerView = config.TitleBorderStyle.Width(m.common.Width).Render(m.textareaTitle) + "\n" + m.textModel.View() + "\n"
		height -= textModelHeight + 3
		helpView = m.help.View(textModelKeys)
	} else if m.viewList.Visible {
		height -= m.viewList.Height + 1
		footerView = m.viewList.View()
		helpView = m.help.View(viewlist.Keys)
	} else {
		helpView = m.help.View(keys)
	}

	m.maxScroll = m.logEntries.Height() - height
	if m.autoScroll {
		m.scrollOffset = m.maxScroll
	}

	start := max(0, min(m.maxScroll, m.scrollOffset))

	return m.logEntries.View(start, height) + footerView + helpView
}

func (m *Model) handleLogEntry() tea.Cmd {
	return func() tea.Msg {
		l := <-m.lChan
		return LogMsg(l)
	}
}

func (m *Model) handleLogMsg(msg LogMsg) tea.Msg {
	l := pipeline.LogEntry{
		Raw: string(msg),
	}
	_ = m.pipeline.Run(&l)
	old := m.logEntries.Add(l)
	if old != nil {
		m.scrollOffset -= old.Height
	}
	if m.autoScroll {
		m.scrollOffset = m.maxScroll
	}
	return nil
}
