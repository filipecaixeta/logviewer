package source

import (
	"reflect"

	"github.com/filipecaixeta/logviewer/internal/config"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type List struct {
	list.Model
	delegate *listDelegate
	items    []ListItem
	Title    string
}

type ListItem interface {
	String() string
	Children() []ListItem
	FilterValue() string
}

func NewList(title string, items []ListItem) *List {
	l := &List{
		delegate: &listDelegate{},
		items:    []ListItem{},
		Title:    title,
	}
	l.Model = list.New(Convert2BubbleItems(items), l.delegate, 0, 0)
	l.Model.SetDelegate(l.delegate)
	l.items = items
	l.Model.Styles.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	l.Model.Styles.TitleBar = lipgloss.NewStyle().Background(lipgloss.Color(""))
	l.Model.Title = config.TitleStyle.Render(title)
	l.Model.SetShowHelp(false)
	l.Model.SetShowStatusBar(false)
	l.Model.SetFilteringEnabled(false)
	return l
}

func (l *List) SelectedItem() ListItem {
	item := l.Model.SelectedItem()
	listItem, ok := item.(ListItem)
	if !ok {
		return nil
	}
	return listItem
}

func (l *List) SetItems(items []ListItem) {
	l.items = items
	l.Model.SetItems(Convert2BubbleItems(items))
}

func (l *List) SetActive(active bool) {
	l.delegate.IsActive = active
}

func (l *List) IsActive() bool {
	return l.delegate.IsActive
}

func (l *List) Update(msg tea.Msg) (*List, tea.Cmd) {
	l.Model, _ = l.Model.Update(msg)
	return l, nil
}

// convert2BubbleItems converts a slice of type T to a slice of type list.Item
func Convert2BubbleItems[T ListItem](items []T) []list.Item {
	r := make([]list.Item, len(items))
	for i, item := range items {
		r[i] = item
	}
	return r
}

// convertInterface2ListItems converts a slice to a slice of type ListItem
func ConvertInterface2ListItems(items interface{}) []ListItem {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return nil
	}

	r := make([]ListItem, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		if listItem, ok := item.(ListItem); ok {
			r[i] = listItem
		}
	}
	return r
}

// convert2ListItems converts a slice of type T to a slice of type ListItem
func Convert2ListItems[T ListItem](items []T) []ListItem {
	r := make([]ListItem, len(items))
	for i, item := range items {
		r[i] = item
	}
	return r
}
