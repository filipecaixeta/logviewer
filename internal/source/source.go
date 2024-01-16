package source

import (
	"context"

	"github.com/filipecaixeta/logviewer/internal/state"

	tea "github.com/charmbracelet/bubbletea"
)

type Source interface {
	Init(stateChan chan state.State) tea.Cmd
	Columns() []*List
	Logs(ctx context.Context, stateChan chan state.State, logChan chan string) tea.Cmd
}
