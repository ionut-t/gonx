package shared_form

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/ionut-t/gonx/internal/messages"
	"strconv"
)

type FormMsg struct {
	Description string
	Count       int
}

type Model struct {
	form *huh.Form
	help help.Model
}

func New() Model {
	count := huh.NewInput().
		Key("count").
		Title("How many times do you want it to run?").
		CharLimit(3).
		Validate(func(str string) error {
			if str == "" {
				return fmt.Errorf("count is required")
			}

			num, err := strconv.Atoi(str)
			if err != nil {
				return fmt.Errorf("please enter a valid number")
			}

			if num <= 0 {
				return fmt.Errorf("number must be greater than 0")
			}

			if num > 100 {
				return fmt.Errorf("number must be less or equal to 100")
			}

			return nil
		})

	description := huh.NewInput().
		Key("description").
		Title("You can provide an optional description")

	form := Model{
		form: huh.NewForm(
			huh.NewGroup(
				count,
				description,
			),
		).WithTheme(huh.ThemeCatppuccin()),
		help: help.New(),
	}

	form.form.WithKeyMap(&formKeyMap)
	form.form.WithShowHelp(false)
	count.Focus()

	return form
}

func (m Model) Init() tea.Cmd {
	return m.form.Init()
}

func (m Model) View() string {
	return m.form.View() + "\n" + m.help.View(formHelp)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, formKeyMap.Quit):
			return m, tea.Quit

		case key.Matches(msg, formHelp.Back):
			return m, messages.Dispatch(messages.NavigateToViewMsg(1))
		}

		switch msg.Type {
		case tea.KeyEnter:
			if m.form.State == huh.StateCompleted {
				return m, m.dispatchFormMsgIfValid()
			}
		}
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		return m, m.dispatchFormMsgIfValid()
	}

	return m, tea.Batch(cmds...)
}

func (m Model) dispatchFormMsgIfValid() tea.Cmd {
	countStr := m.form.GetString("count")
	count, err := strconv.Atoi(countStr)

	if err != nil || count <= 0 {
		return nil
	}

	return messages.Dispatch(FormMsg{
		Count:       count,
		Description: m.form.GetString("description"),
	})
}

var formKeyMap = huh.KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "ctrl+q"),
		key.WithHelp("ctrl+c/ctrl-q", "quit"),
	),
	Confirm: huh.ConfirmKeyMap{
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
	},
	Input: huh.InputKeyMap{
		AcceptSuggestion: key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("ctrl+e", "complete")),
		Prev:             key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "previous")),
		Next:             key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", "next")),
		Submit:           key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	},
}

var formHelp = formKeyMapHelp{
	huh: formKeyMap,
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}

type formKeyMapHelp struct {
	huh  huh.KeyMap
	Back key.Binding
}

func (k formKeyMapHelp) ShortHelp() []key.Binding {
	return []key.Binding{
		k.huh.Confirm.Next,
		k.huh.Confirm.Prev,
		k.huh.Confirm.Submit,
		k.Back,
		k.huh.Quit,
	}
}

func (k formKeyMapHelp) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
