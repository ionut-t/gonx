package tests_analyser

import (
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	form "github.com/ionut-t/gonx/benchmark/shared-form"
	"github.com/ionut-t/gonx/internal/messages"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/ui/suspense"
	"github.com/ionut-t/gonx/ui/viewport"
	"github.com/ionut-t/gonx/workspace"
	"strings"
	"time"
)

const resultTitle = "📊 Benchmark results"

const padding = 2

type view int

const (
	formView view = iota
	buildView
	resultsView
)

type Model struct {
	view     view
	projects []workspace.Project
	form     form.Model
	viewport viewport.Model
	suspense suspense.Model
	progress progress.Model

	width  int
	height int

	count          int
	completed      int
	totalProcesses int
	results        []TestBenchmark
}

func New(projects []workspace.Project, width, height int) Model {
	return Model{
		projects: projects,
		width:    width,
		height:   height,
		form:     form.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
	//return m.form.Init()
}

func (m Model) View() string {
	switch m.view {
	case formView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.form.View())

	case buildView:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Padding(1, 1).Render(m.suspense.View()),
			lipgloss.NewStyle().Padding(0, 1).Render(m.progress.View()),
		)

	case resultsView:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			styles.Header("", resultTitle),
			m.viewport.View(),
		)
	}

	return ""
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
		m.progress.Width = m.width - padding*2

	case form.FormMsg:
		m.view = buildView
		m.count = msg.Count

		return m, messages.Dispatch(StartMsg{
			StartTime:   time.Now(),
			Projects:    m.projects,
			Count:       msg.Count,
			Description: msg.Description,
		})

	case StartMsg:
		m.completed = 0
		m.results = make([]TestBenchmark, 0)
		m.suspense = suspense.New("Starting benchmark...", true)
		m.progress = progress.New(progress.WithDefaultGradient())
		m.progress.Width = m.width - padding*2
		m.progress.PercentageStyle = styles.Primary

		return m, tea.Batch(
			startBenchmark(msg.Projects, msg.Description, msg.Count),
			m.suspense.Spinner.Tick,
			m.progress.SetPercent(0.0),
		)

	case TotalProcessesMsg:
		m.totalProcesses = int(msg)

	case NxCacheResetStartMsg:
		m.suspense.Message = "Resetting the Nx cache and stopping the daemon"
		m.suspense.Loading = true
		return m, m.suspense.Spinner.Tick

	case TestStartMsg:
		m.suspense.Message = fmt.Sprintf("Testing %s %s...",
			styles.Primary.Bold(true).Render(msg.Project.GetName()),
			msg.Project.GetType(),
		)
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case TestCompleteMsg:
		m.completed++
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case TestFailedMsg:
		m.completed++
		m.suspense.Message = msg.Error.Error()
		return m.handleBenchmarkBuild()

	case WriteStatsStartMsg:
		m.suspense.Message = fmt.Sprintf("Writing stats for %s %s...",
			styles.Primary.Bold(true).Render(msg.Project.GetName()),
			msg.Project.GetType(),
		)
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case WriteStatsCompleteMsg:
		m.results = append(m.results, msg.Benchmark)
		return m.handleBenchmarkBuild()

	case WriteStatsFailedMsg:
		m.suspense.Message = msg.Error.Error()
		return m.handleBenchmarkBuild()

	case DoneMsg:
		renderBenchmarkResults(&m)
		m.viewport.Update(msg)
		m.suspense.Loading = false
		m.view = resultsView

	case spinner.TickMsg:
		if m.suspense.Loading {
			var suspenseModel tea.Model
			suspenseModel, cmd = m.suspense.Update(msg)
			m.suspense = suspenseModel.(suspense.Model)
			return m, cmd
		}

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		keyMsg := msg.String()
		switch keyMsg {
		case "esc":
			if m.view != formView {
				return m, messages.Dispatch(messages.NavigateToViewMsg(1))
			}
		}
	}

	// Handle keyboard and mouse events in the viewport
	viewportModel, cmd := m.viewport.Update(msg)
	m.viewport = viewportModel.(viewport.Model)
	cmds = append(cmds, cmd)

	switch m.view {
	case formView:
		formModel, cmd := m.form.Update(msg)
		m.form = formModel.(form.Model)
		cmds = append(cmds, cmd)

	case buildView:
		suspenseModel, cmd := m.suspense.Update(msg)
		m.suspense = suspenseModel.(suspense.Model)
		cmds = append(cmds, cmd)

		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	default:
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleBenchmarkBuild() (Model, tea.Cmd) {
	if m.completed == (len(m.projects) * m.count) {
		return m, tea.Sequence(
			m.progress.SetPercent(1.0),
			// wait for the progress bar to finish animating
			tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return DoneMsg{}
			}),
		)
	}

	return m, m.progress.IncrPercent(m.getProgressIncrement())
}

func (m Model) getProgressIncrement() float64 {
	return 1.0 / float64(m.totalProcesses)
}

func renderBenchmarkResults(m *Model) {
	results := m.results

	var contents []string

	border := styles.NormalText.Render(strings.Repeat("─", min(50, m.width-padding)))

	for i, bm := range results {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			border,
			fmt.Sprintf("Stats for %s %s:", styles.Primary.Bold(true).Render(bm.Project), bm.Type),
			border,
			renderStats(bm),
			border,
		)

		if i < len(results)-1 {
			content += "\n"
		}

		contents = append(contents, content)
	}

	output := lipgloss.NewStyle().
		Padding(0, 4).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			contents...,
		))

	options := viewport.Options{
		Width:   m.width,
		Height:  m.viewportHeight(),
		Content: output,
	}

	m.viewport = viewport.New(options)
}

func renderStats(bm TestBenchmark) string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		styles.Success.Render(fmt.Sprintf("%sMin: %.2fs", styles.IconStyle("🕒"), bm.Min)),
		styles.Success.Render(fmt.Sprintf("%sMax: %.2fs", styles.IconStyle("🕒"), bm.Max)),
		styles.Success.Render(fmt.Sprintf("%sAverage: %.2fs", styles.IconStyle("🕒"), bm.Average)),
		styles.Success.Render(fmt.Sprintf("%sTotal runs: %d", styles.IconStyle("🔄"), bm.TotalRuns)),
	)
}

func (m Model) viewportHeight() int {
	return m.height - lipgloss.Height(styles.Header(resultTitle))
}
