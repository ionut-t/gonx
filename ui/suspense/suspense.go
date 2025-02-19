package suspense

import (
	"fmt"
	"github.com/ionut-t/gonx/ui/styles"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Loading bool
	Message string
	Spinner spinner.Model
}

func (m Model) Init() tea.Cmd {
	return m.Spinner.Tick
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
		m.Spinner, cmd = m.Spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m Model) View() string {
	return fmt.Sprintf("%s %s", m.Spinner.View(), m.Message)
}

func New(message string, loading bool) Model {
	suspense := Model{Message: message, Loading: loading}

	suspense.Spinner = spinner.New()
	suspense.Spinner.Spinner = spinner.Points
	suspense.Spinner.Style = styles.Primary

	return suspense
}
