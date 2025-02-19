package messages

import tea "github.com/charmbracelet/bubbletea"

type NavigateToViewMsg int

func Dispatch(msg tea.Msg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}
