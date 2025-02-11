package program

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/benchmark"
	"github.com/ionut-t/gonx/suspense"
	"github.com/ionut-t/gonx/ui"
	"github.com/ionut-t/gonx/ui/help"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"os"
	"slices"
	"strings"
	"time"
)

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return ui.RedFg.BorderStyle(b).BorderForeground(lipgloss.Color(ui.Red)).Padding(0, 1)
	}()

	footerStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

const padding = 2

func (m Model) headerView() string {
	title := ui.RedFg.Render(titleStyle.Render(m.viewportTitle))
	line := ui.RedFg.Render(strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) footerView() string {
	info := footerStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := ui.RedFg.Render(strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info))))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

type view int

const (
	suspenseView view = iota
	selectActionView
	runCountInputView
	selectAppsView
	descriptionInputView
	benchmarkRunView
	benchmarkResultsView
	allMetricsView
)

type Model struct {
	view             view
	suspense         suspense.Model
	workspace        workspace.Workspace
	viewport         viewport.Model
	viewportTitle    string
	width            int
	height           int
	benchmarkData    benchmarkData
	progress         progress.Model
	selectAction     selectActionModel
	selectApps       appSelectionModel
	descriptionInput descriptionInputModel
	runCountInput    runCountInputModel
}

type benchmarkData struct {
	completed      int
	totalProcesses int
	benchmarks     []benchmark.Benchmark
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
		m.view = selectActionView

		m.selectAction = newActionsList(m.width)

		actionModel, cmd := m.selectAction.Update(msg)
		m.selectAction = actionModel.(selectActionModel)

		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)

	case actionSelectedMsg:
		switch msg.action {
		case runBenchmark:
			m.view = selectAppsView

			m.selectApps = newAppSelectionList(m.width, m.workspace.Applications)

			selectModel, cmd := m.selectApps.Update(msg)
			m.selectApps = selectModel.(appSelectionModel)

			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)

		case runBulkBenchmark:
			m.view = runCountInputView
			m.runCountInput = newRunCount()

			return m, tea.Batch(cmds...)

		case viewMetrics:
			m.view = allMetricsView
			metrics, err := benchmark.ReadAllMetrics()

			recordsCount := func() string {
				count := len(metrics)

				if count == 1 {
					return "1 record"
				}

				return fmt.Sprintf("%d records", count)
			}

			m.viewportTitle = fmt.Sprintf("📊 Metrics history (%s)", recordsCount())

			if err != nil {
				if errors.Is(err, os.ErrNotExist) {

					m.viewport.SetContent(
						ui.CyanFg.Render("You don't have any metrics recorded yet. Run a benchmark to record them."),
					)
					return m, nil
				}

				m.viewport.SetContent(err.Error())
				return m, nil
			}

			renderBenchmarkMetrics(&m, metrics)

			return m, nil

		case viewBulkMetrics:
			m.view = allMetricsView
			metrics, err := benchmark.ReadAllMetrics()

			if err != nil {
				if errors.Is(err, os.ErrNotExist) {

					m.viewport.SetContent(
						ui.CyanFg.Render("You don't have any metrics recorded yet. Run a benchmark to record them."),
					)
					m.viewportTitle = "📊 Bulk Metrics history"
					return m, nil
				}

				m.viewport.SetContent(err.Error())
				return m, nil
			}

			renderBulkMetrics(&m, metrics)
			return m, nil
		}

	case appsSelectedMsg:
		m.view = descriptionInputView
		m.descriptionInput = newDescription(m.width, m.descriptionInput.textInput.Value())

		return m, tea.Batch(cmds...)

	case descriptionInputMsg:
		m.view = benchmarkRunView
		m.benchmarkData.completed = 0
		m.descriptionInput.textInput.Reset()

		return m, tea.Batch(
			m.progress.SetPercent(0.0),
			func() tea.Msg {
				return benchmark.StartMsg{StartTime: time.Now(), Apps: m.selectApps.apps, Description: string(msg)}
			},
		)

	case runCountCancelMsg:
		m.view = selectActionView
		return m, nil

	case runCountMsg:
		m.view = selectAppsView
		m.selectApps = newAppSelectionList(m.width, m.workspace.Applications)

		return m, nil

	case descriptionInputCancelMsg:
		m.view = selectAppsView
		return m, nil

	case benchmark.StartMsg:
		m.benchmarkData.completed = 0
		m.benchmarkData.benchmarks = make([]benchmark.Benchmark, 0)

		return m, benchmark.New(msg.Apps, msg.Description, m.getRunCount())

	case workspace.ErrMsg:
		m.suspense.Loading = false
		m.suspense.Message = lipgloss.JoinVertical(
			lipgloss.Left,
			ui.RedFg.Render("No Nx workspace found. Please run this command in the root of your Nx workspace."),
			"\n",
			ui.DimFg.Render("Press ctrl+c/ctrl-q to exit."),
		)
		return m, nil

	case benchmark.TotalProcessesMsg:
		m.benchmarkData.totalProcesses = msg.Total

	case benchmark.NxCacheResetStartMsg:
		m.suspense.Message = m.getProcessPrefix() + "Resetting the Nx cache and stopping the daemon"
		m.suspense.Loading = true
		return m, m.suspense.Spinner.Tick

	case benchmark.NxCacheResetCompleteMsg:
		m.suspense.Message = m.getProcessPrefix() + "Successfully reset the Nx workspace."
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case benchmark.BuildStartMsg:
		m.suspense.Message = fmt.Sprintf("%sBuilding %s application...", m.getProcessPrefix(), ui.CyanFg.Bold(true).Render(msg.App))

	case benchmark.CalculateBundleSizeMsg:
		m.suspense.Message = fmt.Sprintf("%sCalculating bundle size for %s application...", m.getProcessPrefix(), ui.CyanFg.Bold(true).Render(msg.App))
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case benchmark.WriteStatsMsg:
		m.suspense.Message = fmt.Sprintf("%sWriting stats for %s application...", m.getProcessPrefix(), ui.CyanFg.Bold(true).Render(msg.App))
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case benchmark.BuildCompleteMsg:
		m.benchmarkData.completed++
		m.benchmarkData.benchmarks = append(m.benchmarkData.benchmarks, msg.Benchmark)

		return m.handleBenchmarkBuild()

	case benchmark.BuildFailedMsg:
		m.benchmarkData.completed++
		m.suspense.Message = msg.Error.Error()

		return m.handleBenchmarkBuild()

	case benchmark.DoneMsg:
		if msg.Type == benchmark.Single {
			renderBenchmarkResults(&m, msg.Benchmarks)
		} else {
			renderBulkMetrics(&m, msg.Benchmarks)
			m.viewportTitle = "📊 Bulk benchmark results"
		}

		m.viewport, cmd = m.viewport.Update(msg)
		m.suspense.Loading = false
		m.view = benchmarkResultsView

		return m, m.progress.SetPercent(0.0)

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
		case "backspace", "esc":
			views := []view{allMetricsView, selectAppsView, benchmarkResultsView}

			if slices.Contains(views, m.view) {
				m.view = selectActionView
			}

			cmds = append(cmds, cmd)

		case "ctrl+q", "ctrl+c":
			return m, tea.Quit
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	if m.view == selectActionView {
		actionModel, cmd := m.selectAction.Update(msg)
		m.selectAction = actionModel.(selectActionModel)
		cmds = append(cmds, cmd)
	}

	if m.view == selectAppsView {
		actionModel, cmd := m.selectApps.Update(msg)
		m.selectApps = actionModel.(appSelectionModel)
		cmds = append(cmds, cmd)
	}

	if m.view == descriptionInputView {
		inputModel, cmd := m.descriptionInput.Update(msg)
		m.descriptionInput = inputModel.(descriptionInputModel)
		cmds = append(cmds, cmd)
	}

	if m.view == runCountInputView {
		inputModel, cmd := m.runCountInput.Update(msg)
		m.runCountInput = inputModel.(runCountInputModel)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.view {
	case suspenseView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.suspense.View())

	case selectActionView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.selectAction.View())

	case selectAppsView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.selectApps.View())

	case benchmarkRunView:
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Padding(1, 1).Render(m.suspense.View()),
			lipgloss.NewStyle().Padding(0, 1).Render(m.progress.View()),
		)

	case descriptionInputView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.descriptionInput.View())

	case runCountInputView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.runCountInput.View())

	default:
		return lipgloss.NewStyle().
			Width(m.width).
			Height(m.height).
			Padding(1, 1).
			Render(lipgloss.JoinVertical(
				lipgloss.Left,
				m.headerView(),
				m.viewport.View(),
				m.footerView(),
				help.New().View(),
			))
	}
}

func (m Model) getProgressIncrement() float64 {
	return 1.0 / float64(m.benchmarkData.totalProcesses)
}

func (m Model) handleBenchmarkBuild() (Model, tea.Cmd) {
	if m.benchmarkData.completed == m.getRunCount()*len(m.selectApps.apps) {
		bmType := benchmark.GetType(m.getRunCount())

		m.runCountInput.Reset()

		return m, tea.Sequence(
			m.progress.SetPercent(1.0),
			// wait for the progress bar to finish animating
			tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return benchmark.DoneMsg{Benchmarks: m.benchmarkData.benchmarks, Type: bmType}
			}),
		)
	}

	return m, m.progress.IncrPercent(m.getProgressIncrement())
}

func New() {
	progressBar := progress.New(progress.WithSolidFill(ui.Cyan))
	progressBar.PercentageStyle = ui.CyanFg

	program := Model{
		view:     suspenseView,
		suspense: suspense.New("Scanning workspace...", true),
		progress: progressBar,
	}

	if _, err := tea.NewProgram(program, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func renderBenchmarkResults(m *Model, benchmarks []benchmark.Benchmark) {
	m.viewportTitle = "📊 Benchmark results"
	var contents []string

	border := ui.CyanFg.Render(strings.Repeat("─", min(50, m.width-padding)))

	for i, bm := range benchmarks {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			border,
			fmt.Sprintf("Stats for %s app:", ui.CyanFg.Bold(true).Render(bm.AppName)),
			border,
			benchmark.RenderStats(bm),
			border,
		)

		if i < len(benchmarks)-1 {
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

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	vWidth := m.width - padding
	vHeight := m.height - verticalMarginHeight - padding

	m.viewport = viewport.New(vWidth, vHeight)
	m.viewport.YPosition = headerHeight
	m.viewport.SetContent(output)
}

func renderBenchmarkMetrics(m *Model, metrics []benchmark.Benchmark) {
	border := ui.CyanFg.Render(strings.Repeat("─", min(50, m.width-padding)))

	var contents []string

	getDescription := func(description string) string {
		if description == "" {
			return "-"
		}

		return description
	}

	for i, bm := range metrics {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			ui.CyanFg.Render(fmt.Sprintf(" 🗓️ Recorded on %s at %s", bm.CreatedAt.Format("02/01/2006"), bm.CreatedAt.Format("15:04:05"))),
			ui.CyanFg.Render(fmt.Sprintf(" 📝 Description: %s", getDescription(bm.Description))),
			ui.CyanFg.Render(fmt.Sprintf(" 💻 App: %s", bm.AppName)),
			benchmark.RenderStats(bm),
		)

		if i < len(metrics)-1 {
			content += "\n\n" + border + "\n"
		}

		contents = append(contents, content)
	}

	output := lipgloss.NewStyle().
		Padding(0, 4).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			contents...,
		))

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	vWidth := m.width - padding
	vHeight := m.height - verticalMarginHeight - padding

	m.viewport = viewport.New(vWidth, vHeight)
	m.viewport.YPosition = headerHeight
	m.viewport.SetContent(output)
}

func renderBulkMetrics(m *Model, metrics []benchmark.Benchmark) {
	border := ui.CyanFg.Render(strings.Repeat("─", min(50, m.width-padding)))

	var contents []string

	getDescription := func(description string) string {
		if description == "" {
			return "-"
		}

		return description
	}

	bulkMetrics := utils.Filter(metrics, func(bm benchmark.Benchmark) bool {
		return bm.Type == benchmark.Bulk
	})

	groupedMetrics := make(map[string][]benchmark.Benchmark)
	for _, bm := range bulkMetrics {
		groupedMetrics[bm.RunId] = append(groupedMetrics[bm.RunId], bm)
	}

	type bmGroup struct {
		appName     string
		createdAt   time.Time
		description string
		min         float64
		max         float64
		average     float64
		count       int
	}

	// First group by app name within each runId group
	appGroups := make(map[string]map[string][]benchmark.Benchmark)
	for runId, metrics := range groupedMetrics {
		appGroups[runId] = make(map[string][]benchmark.Benchmark)
		for _, bm := range metrics {
			appGroups[runId][bm.AppName] = append(appGroups[runId][bm.AppName], bm)
		}
	}

	benchmarks := make([]bmGroup, 0)

	// Calculate statistics for each app in each run group
	for _, apps := range appGroups {
		for appName, bms := range apps {
			if len(bms) == 0 {
				continue
			}

			// Calculate min, max, average
			var sum float64
			minDuration := bms[0].Duration
			maxDuration := bms[0].Duration

			for _, bm := range bms {
				sum += bm.Duration
				if bm.Duration < minDuration {
					minDuration = bm.Duration
				}
				if bm.Duration > maxDuration {
					maxDuration = bm.Duration
				}
			}

			avg := sum / float64(len(bms))

			benchmarks = append(benchmarks, bmGroup{
				appName:     appName,
				createdAt:   bms[0].CreatedAt,
				description: bms[0].Description,
				min:         minDuration,
				max:         maxDuration,
				average:     avg,
				count:       bms[0].RunCount,
			})
		}
	}

	recordsCount := func() string {
		count := len(benchmarks)

		if count == 1 {
			return "1 record"
		}

		return fmt.Sprintf("%d records", count)
	}

	m.viewportTitle = fmt.Sprintf("📊 Bulk Metrics history (%s)", recordsCount())

	for i, bm := range benchmarks {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			ui.CyanFg.Render(fmt.Sprintf(" 🗓️ Recorded on %s at %s", bm.createdAt.Format("02/01/2006"), bm.createdAt.Format("15:04:05"))),
			ui.CyanFg.Render(fmt.Sprintf(" 📝 Description: %s", getDescription(bm.description))),
			ui.CyanFg.Render(fmt.Sprintf(" 💻 App: %s", bm.appName)),
			ui.CyanFg.Render(fmt.Sprintf(" 🔄 Run count: %d", bm.count)),
			ui.GreenFg.Render(fmt.Sprintf(" 🕒 Min: %.2fs", bm.min)),
			ui.GreenFg.Render(fmt.Sprintf(" 🕒 Max: %.2fs", bm.max)),
			ui.GreenFg.Render(fmt.Sprintf(" 🕒 Average: %.2fs", bm.average)),
		)

		if i < len(metrics)-1 {
			content += "\n\n" + border + "\n"
		}

		contents = append(contents, content)
	}

	output := lipgloss.NewStyle().
		Padding(0, 4).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			contents...,
		))

	headerHeight := lipgloss.Height(m.headerView())
	footerHeight := lipgloss.Height(m.footerView())
	verticalMarginHeight := headerHeight + footerHeight

	vWidth := m.width - padding
	vHeight := m.height - verticalMarginHeight - padding

	m.viewport = viewport.New(vWidth, vHeight)
	m.viewport.YPosition = headerHeight
	m.viewport.SetContent(output)
}

func (m Model) getProcessPrefix() string {
	count := m.runCountInput.Value()
	bmType := benchmark.GetType(count)

	if bmType == benchmark.Single {
		return ""
	}

	return "Bulk benchmark: "
}

func (m Model) getRunCount() int {
	return max(m.runCountInput.Value(), 1)
}
