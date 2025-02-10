package program

import (
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/ui"
)

type descriptionInputMsg string

type descriptionInputCancelMsg struct{}

type descriptionInputModel struct {
	textInput textinput.Model
	help      help.Model
}

func newDescription(wWidth int, text string) descriptionInputModel {
	input := textinput.New()
	input.Placeholder = "Description"
	input.CharLimit = 100
	input.Width = min(wWidth, 60)
	input.Cursor.SetMode(cursor.CursorBlink)
	input.Focus()

	if text != "" {
		input.SetValue(text)
	}

	return descriptionInputModel{
		textInput: input,
		help:      help.New(),
	}
}

func (m descriptionInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m descriptionInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, func() tea.Msg {
				return descriptionInputMsg(m.textInput.Value())
			}
		}

		if key.Matches(msg, inputKeys.Back) {
			return m, func() tea.Msg {
				return descriptionInputCancelMsg{}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m descriptionInputModel) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		ui.CyanFg.Render("You can provide an optional description"),
		"\n",
		ui.MagentaFg.Render(m.textInput.View()),
		"\n",
		m.help.View(inputKeys),
	)
}

type inputKeyMap struct {
	Confirm key.Binding
	Back    key.Binding
	Quit    key.Binding
}

func (k inputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Back, k.Quit}
}

func (k inputKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

var inputKeys = inputKeyMap{
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit"),
	),
}
