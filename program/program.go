package program

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/benchmark"
	"github.com/ionut-t/gonx/suspense"
	"github.com/ionut-t/gonx/ui"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"os"
	"strings"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

const padding = 2

func (m Model) headerView() string {
	title := ui.Red.Render(titleStyle.Render("Benchmark"))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

type Model struct {
	suspense  suspense.Model
	workspace workspace.Workspace
	output    string
	viewport  viewport.Model
	width     int
	height    int
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			ws, err := workspace.NewWorkspace()
			if err != nil {
				return workspace.ErrMsg{Err: err}
			}
			return workspace.DoneMsg{Workspace: *ws}
		},
		m.suspense.Init(),
		tea.SetWindowTitle("gonx"),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		vWidth := m.width - padding
		vHeight := m.height - verticalMarginHeight - padding

		m.viewport.Width = vWidth
		m.viewport.Height = vHeight

	case workspace.DoneMsg:
		m.workspace = msg.Workspace
		m.suspense.Message = "Building benchmarks..."

		bmMsg := func() tea.Msg {
			bm, err := benchmark.New(m.workspace, "New benchmark")

			if err != nil {
				return benchmark.ErrMsg{Err: err}
			}

			return benchmark.DoneMsg{Benchmarks: bm}
		}

		return m, bmMsg

	case workspace.ErrMsg:
		return m, tea.Quit

	case benchmark.DoneMsg:
		m.suspense.Loading = false

		setViewportContent(&m, msg.Benchmarks)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)

	case benchmark.ErrMsg:
		return m, tea.Quit

	case spinner.TickMsg:
		if m.suspense.Loading {
			var suspenseModel tea.Model
			suspenseModel, cmd = m.suspense.Update(msg)
			m.suspense = suspenseModel.(suspense.Model)
			return m, cmd
		}

	case tea.KeyMsg:
		keyMsg := msg.String()
		switch keyMsg {
		case "q":
			return m, tea.Quit
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.suspense.Loading {
		return m.suspense.View()
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(1, 1).
		Render(lipgloss.JoinVertical(
			lipgloss.Center,
			m.headerView(),
			m.viewport.View(),
			m.footerView(),
		))
}

func New() {
	loadingMessage := "Scanning workspace..."
	suspenseModel := suspense.New(loadingMessage, true)

	program := Model{
		suspense: suspenseModel,
	}

	if _, err := tea.NewProgram(program, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func setViewportContent(m *Model, benchmarks []benchmark.Benchmark) {
	var contents []string

	for i, bm := range benchmarks {
		content := getBenchmarkContent(bm, m.width)

		if i < len(benchmarks)-1 {
			content += "\n"
		}

		contents = append(contents, content)
	}

	m.output = lipgloss.JoinVertical(
		lipgloss.Left,
		contents...,
	)

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	vWidth := m.width - padding
	vHeight := m.height - verticalMarginHeight - padding

	m.viewport = viewport.New(vWidth, vHeight)
	m.viewport.YPosition = headerHeight
	m.viewport.SetContent(m.output)
}

func getBenchmarkContent(bm benchmark.Benchmark, windowWidth int) string {
	border := ui.Cyan.Render(strings.Repeat("─", min(40, windowWidth-padding)))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		border,
		fmt.Sprintf("Report for %s app:", ui.Cyan.Bold(true).Render(bm.AppName)),
		border,
		ui.Green.Render(fmt.Sprintf(" 🕒 Build time: %.2fs", bm.Duration)),
		ui.Green.Render(fmt.Sprintf(" 🎯 Main bundle: %s", utils.FormatFileSize(bm.Stats.Initial.Main))),
		ui.Green.Render(fmt.Sprintf(" ⚙️ Runtime bundle: %s", utils.FormatFileSize(bm.Stats.Initial.Runtime))),
		ui.Green.Render(fmt.Sprintf(" 🔧 Polyfills bundle: %s", utils.FormatFileSize(bm.Stats.Initial.Polyfills))),
		ui.Yellow.Render(fmt.Sprintf(" 📦 Initial total: %s", utils.FormatFileSize(bm.Stats.Initial.Total))),
		ui.Magenta.Render(fmt.Sprintf(" 📦 Lazy chunks total: %s", utils.FormatFileSize(bm.Stats.Lazy))),
		ui.Blue.Render(fmt.Sprintf(" 📦 Bundle total: %s", utils.FormatFileSize(bm.Stats.Total))),
		ui.Blue.Render(fmt.Sprintf(" 🎨 Styles total: %s", utils.FormatFileSize(bm.Stats.Styles))),
		ui.Cyan.Render(fmt.Sprintf(" 📂 Assets total: %s", utils.FormatFileSize(bm.Stats.Assets))),
		ui.Blue.Render(fmt.Sprintf(" 📊 Overall total: %s", utils.FormatFileSize(bm.Stats.OverallTotal))),
		border,
	)
}
