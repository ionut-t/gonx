package program

import (
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
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
	"time"
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
	title := ui.RedFg.Render(titleStyle.Render("Benchmark"))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

type Model struct {
	suspense      suspense.Model
	workspace     workspace.Workspace
	output        string
	viewport      viewport.Model
	width         int
	height        int
	benchmarkData benchmarkData
	progress      progress.Model
}

type benchmarkData struct {
	completed  int
	benchmarks []benchmark.Benchmark
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
		m.progress.Width = m.width - padding*2

	case workspace.DoneMsg:
		m.workspace = msg.Workspace

		return m, tea.Batch(
			m.progress.SetPercent(m.getProgressIncrement()),
			func() tea.Msg {
				return benchmark.StartMsg{StartTime: time.Now()}
			},
		)

	case benchmark.StartMsg:
		m.benchmarkData.completed = 0
		m.benchmarkData.benchmarks = make([]benchmark.Benchmark, 0)
		return m, benchmark.New(m.workspace, "New benchmark")

	case workspace.ErrMsg:
		return m, tea.Quit

	case benchmark.BuildStartMsg:
		m.suspense.Message = fmt.Sprintf("Building %s application...", ui.CyanFg.Bold(true).Render(msg.App))
		m.suspense.Loading = true
		return m, m.suspense.Spinner.Tick

	case benchmark.BuildCompleteMsg:
		m.benchmarkData.completed++
		m.benchmarkData.benchmarks = append(m.benchmarkData.benchmarks, msg.Benchmark)

		return m.handleBenchmarkBuild()

	case benchmark.BuildFailedMsg:
		m.benchmarkData.completed++
		m.suspense.Message = msg.Error.Error()

		return m.handleBenchmarkBuild()

	case benchmark.DoneMsg:
		setViewportContent(&m, msg.Benchmarks)
		m.viewport, cmd = m.viewport.Update(msg)
		m.suspense.Loading = false
		return m, cmd

	case spinner.TickMsg:
		if m.suspense.Loading {
			var suspenseModel tea.Model
			suspenseModel, cmd = m.suspense.Update(msg)
			m.suspense = suspenseModel.(suspense.Model)
			return m, cmd
		}
		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

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
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Padding(1, 1).Render(m.suspense.View()),
			lipgloss.NewStyle().Padding(0, 1).Render(m.progress.View()),
		)
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

func (m Model) getProgressIncrement() float64 {
	return 1.0 / float64(1.0+len(m.workspace.Applications))
}

func (m Model) handleBenchmarkBuild() (Model, tea.Cmd) {
	if m.benchmarkData.completed == len(m.workspace.Applications) {
		return m, tea.Sequence(
			m.progress.SetPercent(1.0),
			// wait for the progress bar to finish animating
			tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return benchmark.DoneMsg{Benchmarks: m.benchmarkData.benchmarks}
			}),
		)
	}

	return m, m.progress.IncrPercent(m.getProgressIncrement())
}

func New() {
	program := Model{
		suspense: suspense.New("Scanning workspace...", true),
		progress: progress.New(progress.WithSolidFill(ui.Cyan)),
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
	border := ui.CyanFg.Render(strings.Repeat("─", min(50, windowWidth-padding)))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		border,
		fmt.Sprintf("Report for %s app:", ui.CyanFg.Bold(true).Render(bm.AppName)),
		border,
		ui.GreenFg.Render(fmt.Sprintf(" 🕒 Build time: %.2fs", bm.Duration)),
		ui.GreenFg.Render(fmt.Sprintf(" 🎯 Main bundle: %s", utils.FormatFileSize(bm.Stats.Initial.Main))),
		ui.GreenFg.Render(fmt.Sprintf(" ⚙️ Runtime bundle: %s", utils.FormatFileSize(bm.Stats.Initial.Runtime))),
		ui.GreenFg.Render(fmt.Sprintf(" 🔧 Polyfills bundle: %s", utils.FormatFileSize(bm.Stats.Initial.Polyfills))),
		ui.YellowFg.Render(fmt.Sprintf(" 📦 Initial total: %s", utils.FormatFileSize(bm.Stats.Initial.Total))),
		ui.MagentaFg.Render(fmt.Sprintf(" 📦 Lazy chunks total: %s", utils.FormatFileSize(bm.Stats.Lazy))),
		ui.BlueFg.Render(fmt.Sprintf(" 📦 Bundle total: %s", utils.FormatFileSize(bm.Stats.Total))),
		ui.BlueFg.Render(fmt.Sprintf(" 🎨 Styles total: %s", utils.FormatFileSize(bm.Stats.Styles))),
		ui.CyanFg.Render(fmt.Sprintf(" 📂 Assets total: %s", utils.FormatFileSize(bm.Stats.Assets))),
		ui.BlueFg.Render(fmt.Sprintf(" 📊 Overall total: %s", utils.FormatFileSize(bm.Stats.OverallTotal))),
		border,
	)
}
