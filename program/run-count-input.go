package program

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/ui"
	"strconv"
	"unicode"
)

type runCountMsg int

type runCountCancelMsg struct{}

type runCountInputModel struct {
	input textinput.Model
	help  help.Model
	min   int
	max   int
	error error
}

func (m runCountInputModel) Value() int {
	value, err := strconv.Atoi(m.input.Value())

	if err != nil {
		return 0
	}

	return value
}

func (m runCountInputModel) Reset() {
	m.input.Reset()
}

func (m runCountInputModel) IsValid() bool {
	value := m.Value()

	return value >= m.min && value <= m.max
}

func newRunCount() runCountInputModel {
	input := textinput.New()
	input.Placeholder = "Run count"
	input.Width = 20
	input.Cursor.SetMode(cursor.CursorBlink)
	input.Focus()

	// Only allow numeric input
	input.Validate = func(s string) error {
		for _, r := range s {
			if !unicode.IsDigit(r) {
				return fmt.Errorf("please enter numbers only")
			}
		}
		return nil
	}

	return runCountInputModel{
		input: input,
		help:  help.New(),
		min:   2,
		max:   100,
	}
}

func (m runCountInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m runCountInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, inputKeys.Back) {
			return m, func() tea.Msg {
				return runCountCancelMsg{}
			}
		}

		switch msg.Type {
		case tea.KeyEnter:
			if !m.IsValid() {
				m.error = fmt.Errorf("please enter a number between %d and %d", m.min, m.max)
				return m, nil
			}

			return m, func() tea.Msg {
				return runCountMsg(m.Value())
			}

		case tea.KeyBackspace:
			break

		default:
			k := msg.String()

			if m.Value() >= m.max {
				return m, nil
			}

			for _, r := range k {
				if !unicode.IsDigit(r) {
					return m, nil
				}
			}
		}
	}

	m.input, cmd = m.input.Update(msg)

	if m.IsValid() {
		m.error = nil
	} else {
		m.error = fmt.Errorf("please enter a number between %d and %d", m.min, m.max)
	}
	return m, cmd
}

func (m runCountInputModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		ui.CyanFg.Render("You can enter how many times you want to run the benchmark"),
		"\n",
		ui.MagentaFg.Render(m.input.View()),
		m.errorMessage(),
		"\n",
		m.help.View(inputKeys),
	)
}

func (m runCountInputModel) errorMessage() string {
	if m.error != nil {
		return ui.RedFg.Render(m.error.Error())
	}

	return ""
}
