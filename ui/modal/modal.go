package modal

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/utils"
)

type Model struct {
	content string
	show    bool
	width   int
	height  int
}

type Options struct {
	Content string
	Show    bool
	Width   int
	Height  int
}

func (m *Model) SetContent(content string) {
	m.content = content
}

func (m *Model) Show() {
	m.show = true
}

func (m *Model) Hide() {
	m.show = false
}

func (m *Model) Toggle() {
	m.show = !m.show
}

func (m Model) IsVisible() bool {
	return m.show
}

func (m *Model) Set(options Options) {
	m.content = utils.Ternary(options.Content == "", m.content, options.Content)
	m.show = utils.Ternary(options.Show, options.Show, m.show)
	m.width = utils.Ternary(options.Width == 0, m.width, options.Width)
	m.height = utils.Ternary(options.Height == 0, m.height, options.Height)
}

func New(options Options) Model {
	return Model{
		content: options.Content,
		show:    options.Show,
		width:   options.Width,
		height:  options.Height,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	if m.show {
		modalStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Overlay1.GetForeground()).
			Padding(1, 2).
			Width(m.width - 2).
			Height(m.height - 2)

		modalView := modalStyle.Render(m.content)

		modalView = lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center, lipgloss.Center,
			modalView,
		)

		return lipgloss.JoinVertical(lipgloss.Center, modalView)
	}

	return ""
}
