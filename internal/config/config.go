package config

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Color      string   `json:"color,omitempty" toml:"color,omitempty"`
	Command    string   `json:"command,omitempty" toml:"-"`
	Filename   string   `json:"filename,omitempty" toml:"-"`
	Namespaces []string `json:"namespaces,omitempty" toml:"namespaces,omitempty"`
	K8sContext string   `json:"k8sContext,omitempty" toml:"k8sContext,omitempty"`
	Views      []View   `json:"views,omitempty" toml:"views,omitempty"`
}

type View struct {
	Name           string      `json:"name" toml:"name"`
	Selector       string      `json:"selector,omitempty" toml:"selector,omitempty"`
	ReturnedFields []string    `json:"returnedFields,omitempty" toml:"returnedFields,omitempty"`
	Filter         string      `json:"filter,omitempty" toml:"filter,omitempty"`
	FilterDefault  bool        `json:"filterDefault,omitempty" toml:"filterDefault,omitempty"`
	Transforms     []Transform `json:"transforms,omitempty" toml:"transforms,omitempty"`
}

type Transform struct {
	Field      string `json:"field" toml:"field"`
	Expression string `json:"expression" toml:"expression"`
}

var (
	TitleBorderStyle lipgloss.Style
	TitleStyle       lipgloss.Style
	SpinnerStyle     lipgloss.Style
	ListActiveStyle  lipgloss.Style
	ListStyle        lipgloss.Style
	Theme            string
)

func SetColor(color string) {
	if color == "light" {
		ListActiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)
		ListStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		TitleBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Bold(true).Foreground(lipgloss.Color("#6b1cff")).MarginBottom(1)
		TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#6b1cff")).MarginBottom(1)
		SpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b1cff"))
		Theme = color
	} else {
		ListActiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#EEEEEE")).Bold(true)
		ListStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		TitleBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, false, false, false).Bold(true).Foreground(lipgloss.Color("#ae81ff")).MarginBottom(1)
		TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ae81ff")).MarginBottom(1)
		SpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ae81ff"))
		Theme = color
	}
}

// save into a toml file
func (c *Config) Save() error {
	// Create or truncate the file
	f, err := os.Create(c.Filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Encode the struct into TOML and write to the file
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(c); err != nil {
		return err
	}

	return nil
}

func (v *View) FilterValue() string {
	return v.Name
}
