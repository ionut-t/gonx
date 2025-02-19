package viewport

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/utils"
	"strings"
)

var (
	footerStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return styles.Primary.Bold(true).BorderStyle(b).BorderForeground(styles.Primary.Bold(true).GetForeground()).Padding(0, 1)
	}()
)

const padding = 2

func (m Model) footerView() string {
	info := footerStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := styles.Primary.Bold(true).Render(strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info))))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

type Model struct {
	viewport viewport.Model
	width    int
	height   int
	content  string
}

type Options struct {
	Width   int
	Height  int
	Content string
}

func New(options Options) Model {
	model := Model{
		width:   options.Width,
		height:  options.Height,
		content: options.Content,
	}

	footerHeight := lipgloss.Height(model.footerView())

	vWidth := options.Width - padding
	vHeight := options.Height - footerHeight - 1

	model.viewport = viewport.New(vWidth, vHeight)
	model.viewport.SetContent(options.Content)

	return model
}

func (m Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Top,
		m.viewport.View(),
		m.footerView(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *Model) Set(options Options) {
	m.viewport.Width = utils.Ternary(options.Width > 0, options.Width, m.width)
	m.viewport.Height = utils.Ternary(options.Height > 0, options.Height, m.height)
	m.viewport.SetContent(utils.Ternary(options.Content != "", options.Content, m.content))
}

func (m *Model) SetContent(content string) {
	m.viewport.GotoTop()
	m.viewport.SetContent(content)
}
