package input

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/internal/keymap"
	"github.com/ionut-t/gonx/ui/help"
	"github.com/ionut-t/gonx/ui/styles"
	"strconv"
)

type mode string

const (
	Text    mode = "text"
	Numeric mode = "numeric"
)

type DoneMsg string

type CancelMsg struct{}

type Model struct {
	input          *huh.Input
	label          string
	help           help.Model
	mode           mode
	min            int
	max            int
	charLimit      int
	error          error
	displayedError error
	showHelp       bool
	focused        bool
}

type Options struct {
	Label          string
	Placeholder    string
	Value          string
	Width          int
	CharLimit      int
	Mode           mode
	Min            int
	Max            int
	DisplayedError error
	HideHelp       bool
}

func New(options Options) Model {
	input := huh.NewInput()

	input.WithWidth(options.Width)
	input.Placeholder(options.Placeholder)
	input.Title(options.Label)
	input.WithTheme(huh.ThemeCatppuccin())

	if options.CharLimit > 0 {
		input.CharLimit(options.CharLimit)
	}

	helpMenu := help.New(options.Width, 20)
	helpMenu.SetKeyMap(keymap.ListKeyMap)

	return Model{
		input:          input,
		label:          options.Label,
		help:           helpMenu,
		mode:           options.Mode,
		min:            options.Min,
		max:            options.Max,
		charLimit:      options.CharLimit,
		displayedError: options.DisplayedError,
		showHelp:       !options.HideHelp,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, keymap.ListKeyMap.Back) {
			return m, func() tea.Msg {
				return CancelMsg{}
			}
		}

		if m.input != nil {
			form, inputCmd := m.input.Update(msg)
			if input, ok := form.(*huh.Input); ok {
				m.input = input
				cmd = inputCmd
			}
		}

		switch msg.Type {
		case tea.KeyEnter:
			if !m.IsValid() {
				m.error = m.displayedError
				return m, nil
			}

			return m, func() tea.Msg {
				return DoneMsg(m.Value())
			}
		}
	}

	if m.IsValid() {
		m.error = nil
	} else {
		m.error = m.displayedError
	}

	return m, cmd
}

func (m Model) View() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Primary.Render(m.input.View()),
	)

	if m.error != nil {
		content += "\n" + styles.Error.Render(m.error.Error())
	}

	if m.showHelp {
		content += "\n\n" + m.help.View()
	}

	return content
}

func (m Model) Value() string {
	if m.input == nil {
		return ""
	}

	value := m.input.GetValue()
	if value == nil {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}

	return str
}

func (m Model) ValueAsInt() int {
	value, err := strconv.Atoi(m.Value())

	if err != nil {
		return 0
	}

	return value
}

func (m *Model) Blur() {
	m.focused = false
	m.input.Blur()
}

func (m *Model) Focus() {
	m.focused = true
	m.input.Focus()
}

func (m Model) Focused() bool {
	return m.focused
}

func (m Model) IsValid() bool {
	if m.mode == Text {
		if m.charLimit > 0 {
			return len(m.Value()) <= m.charLimit
		}

		return true
	}

	if m.min == 0 && m.max == 0 {
		return true
	}

	value := m.ValueAsInt()

	// If only min is set, check if value is greater than min
	if m.min > m.max {
		return value <= m.max
	}

	return value >= m.min && value <= m.max
}

func (m Model) SetValue(value string) {
	m.input.Value(&value)
}
