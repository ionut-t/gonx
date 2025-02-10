package program

import (
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/ui"
)

type descriptionInputMsg string

type descriptionInputModel struct {
	textInput textinput.Model
}

func newDescription(wWidth int) descriptionInputModel {
	input := textinput.New()
	input.Placeholder = "Description"
	input.CharLimit = 100
	input.Width = min(wWidth, 60)
	input.Cursor.SetMode(cursor.CursorBlink)
	input.Focus()

	return descriptionInputModel{
		textInput: input,
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
		ui.DimFg.Render("(press enter to continue)"),
	)
}
