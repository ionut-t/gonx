package help

import (
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/internal/keymap"
	"github.com/ionut-t/gonx/ui/modal"
	"github.com/ionut-t/gonx/ui/styles"
	"strings"
)

func (m Model) getKeys() keymap.Model {
	if m.Searching {
		return keymap.SearchKeyMap
	}

	return m.Keys
}

func (m *Model) CombineWithDefaultKeys(keys keymap.Model) {
	m.Keys = keymap.CombineKeys(keymap.DefaultKeyMap, keys)
}

func (m *Model) CombineWithHistoryKeys(keys keymap.Model) {
	m.Keys = keymap.CombineKeys(keymap.HistoryKeyMap, keys)
}

func (m *Model) SetKeyMap(keys keymap.Model) {
	m.Keys = keys
}

type Model struct {
	Keys      keymap.Model
	Searching bool

	help  help.Model
	modal modal.Model

	width  int
	height int
}

func New(width, height int) Model {
	helpMenu := help.New()
	helpMenu.Styles = help.Styles{
		ShortKey:       styles.Subtext0,
		ShortDesc:      styles.Overlay1,
		ShortSeparator: styles.Subtext0,
		FullKey:        styles.Subtext0,
		FullDesc:       styles.Overlay1,
		FullSeparator:  styles.Subtext0,
	}

	return Model{
		Keys: keymap.DefaultKeyMap,
		help: helpMenu,
		modal: modal.New(modal.Options{
			Content: "Help",
			Show:    false,
			Width:   width,
			Height:  height,
		}),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	if m.modal.IsVisible() {
		return m.modal.View()
	}

	return lipgloss.NewStyle().Padding(0, 1).Render(m.help.View(m.getKeys()))
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		m.modal.Set(modal.Options{Width: msg.Width, Height: msg.Height})

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Help):
			m.modal.SetContent(m.renderFullHelpView())
			m.modal.Toggle()
		}
	}

	return m, nil
}

func (m Model) FullViewOpened() bool {
	return m.modal.IsVisible()
}

func (m Model) renderFullHelpView() string {
	var sb strings.Builder

	bindings := keymap.ReplaceBinding(m.Keys.AllBindings(),
		key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
	)

	enabledBindings := make([]key.Binding, 0)
	maxKeyWidth := 0

	for _, binding := range bindings {
		if !binding.Enabled() {
			continue
		}

		enabledBindings = append(enabledBindings, binding)
		renderedWidth := lipgloss.Width(styles.Teal.Render(binding.Help().Key))
		maxKeyWidth = max(maxKeyWidth, renderedWidth)
	}

	for _, binding := range enabledBindings {
		keyText := binding.Help().Key
		renderedKey := styles.Teal.Render(keyText)
		renderedDescription := styles.Subtext1.Render(binding.Help().Desc)
		currentWidth := lipgloss.Width(renderedKey)
		padding := strings.Repeat(" ", maxKeyWidth-currentWidth+2)

		sb.WriteString(fmt.Sprintf("• %s%s%s\n",
			renderedKey,
			padding,
			renderedDescription,
		))
	}

	return sb.String()
}
