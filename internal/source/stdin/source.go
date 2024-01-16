package stdin

import (
	"bufio"
	"context"
	"os"

	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/state"

	"github.com/filipecaixeta/logviewer/internal/source"

	tea "github.com/charmbracelet/bubbletea"
)

type StdItem string

// String returns a string representation of the container, which will be displayed in the list.
func (c StdItem) String() string {
	return string(c)
}

// Children returns an empty slice, as Docker containers don't have a hierarchical structure in this context.
func (c StdItem) Children() []source.ListItem {
	return []source.ListItem{}
}

// FilterValue returns the value used for filtering the list. It could be the container's name or ID.
func (c StdItem) FilterValue() string {
	return string(c)
}

type Stdin struct {
	columns []*source.List
	scanner *bufio.Scanner
}

func New(config *config.Config) source.Source {
	f := &Stdin{
		scanner: bufio.NewScanner(os.Stdin),
	}

	f.columns = []*source.List{
		source.NewList("stdin", []source.ListItem{StdItem("stdin")}),
	}

	return f
}

func (f *Stdin) Init(stateChan chan state.State) tea.Cmd {
	return func() tea.Msg {
		stateChan <- state.StateBrose
		return nil
	}
}

func (f *Stdin) Columns() []*source.List {
	return f.columns
}

func (f *Stdin) Logs(ctx context.Context, stateChan chan state.State, logChan chan string) tea.Cmd {
	return func() tea.Msg {
		stateChan <- state.StateLogs

		for f.scanner.Scan() {
			select {
			case <-ctx.Done():
				return nil
			case logChan <- f.scanner.Text():
			}
		}
		return nil
	}
}

func (f *Stdin) Close() {
}
