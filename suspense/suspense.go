package suspense

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Loading bool
	Message string
	spinner spinner.Model
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			return m, nil
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m Model) View() string {
	return fmt.Sprintf("\n\n   %s %s\n\n", m.spinner.View(), m.Message)
}

func New(message string, loading bool) Model {
	suspense := Model{Message: message, Loading: loading}

	suspense.spinner = spinner.New()
	suspense.spinner.Spinner = spinner.Points
	suspense.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return suspense
}
