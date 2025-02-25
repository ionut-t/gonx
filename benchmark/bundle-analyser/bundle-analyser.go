package bundle_analyser

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/internal/keymap"
	"github.com/ionut-t/gonx/internal/messages"
	"github.com/ionut-t/gonx/ui/input"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/ui/suspense"
	"github.com/ionut-t/gonx/ui/viewport"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"strings"
	"time"
)

const resultTitle = "📊 Benchmark results"

const padding = 2

type view int

const (
	descriptionView view = iota
	buildView
	resultsView
)

type Model struct {
	view        view
	apps        []workspace.Application
	description input.Model
	viewport    viewport.Model
	suspense    suspense.Model
	progress    progress.Model

	width  int
	height int

	completed      int
	totalProcesses int
	results        []BundleBenchmark
}

func New(apps []workspace.Application, width, height int) Model {
	return Model{
		apps:        apps,
		width:       width,
		height:      height,
		description: createDescriptionInput(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.description.Init()
}

func (m Model) View() string {
	switch m.view {
	case descriptionView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.description.View())

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

	case input.DoneMsg:
		m.view = buildView
		m.completed = 0

		return m, messages.Dispatch(StartMsg{StartTime: time.Now(), Apps: m.apps, Description: string(msg)})

	case StartMsg:
		m.completed = 0
		m.results = make([]BundleBenchmark, 0)
		m.suspense = suspense.New("Starting bundle benchmark", true)
		m.progress = progress.New(progress.WithDefaultGradient())
		m.progress.Width = m.width - padding*2
		m.progress.PercentageStyle = styles.Primary

		return m, tea.Batch(
			startBenchmark(msg.Apps, msg.Description),
			m.suspense.Spinner.Tick,
			m.progress.SetPercent(0.0),
		)

	case TotalProcessesMsg:
		m.totalProcesses = int(msg)

	case NxCacheResetStartMsg:
		m.suspense.Message = "Resetting the Nx cache and stopping the daemon"
		m.suspense.Loading = true
		return m, m.suspense.Spinner.Tick

	case BuildStartMsg:
		m.suspense.Message = fmt.Sprintf("Building %s application", styles.Primary.Bold(true).Render(msg.App))
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case CalculateBundleSizeMsg:
		m.suspense.Message = fmt.Sprintf("Calculating bundle size for %s application", styles.Primary.Bold(true).Render(msg.App))
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case WriteStatsMsg:
		m.suspense.Message = fmt.Sprintf("Writing stats for %s application", styles.Primary.Bold(true).Render(msg.App))
		return m, m.progress.IncrPercent(m.getProgressIncrement())

	case BuildCompleteMsg:
		m.completed++
		m.results = append(m.results, msg.Benchmark)

		return m.handleBenchmarkBuild(nil)

	case BuildFailedMsg:
		m.completed++
		m.suspense.Message = msg.Error.Error()

		return m.handleBenchmarkBuild(msg.Error)

	case DoneMsg:
		if msg.Error != nil {
			options := viewport.Options{
				Width:   m.width,
				Height:  m.viewportHeight(),
				Content: msg.Error.Error(),
			}

			m.viewport = viewport.New(options)
		} else {
			renderBenchmarkResults(&m)
		}

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
		switch {
		case key.Matches(msg, keymap.Back):
			if m.view != buildView {
				return m, messages.Dispatch(messages.NavigateToViewMsg(
					utils.Ternary(m.view == descriptionView, 1, 0)),
				)
			}
		}
	}

	// Handle keyboard and mouse events in the viewport
	viewportModel, cmd := m.viewport.Update(msg)
	m.viewport = viewportModel.(viewport.Model)
	cmds = append(cmds, cmd)

	switch m.view {
	case descriptionView:
		inputModel, cmd := m.description.Update(msg)
		m.description = inputModel.(input.Model)
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

func createDescriptionInput() input.Model {
	options := input.Options{
		Label:       "You can provide an optional description for this benchmark",
		Placeholder: "Description",
		Width:       60,
		Mode:        input.Text,
	}

	descriptionInput := input.New(options)
	descriptionInput.Focus()
	return descriptionInput
}

func (m Model) handleBenchmarkBuild(err error) (Model, tea.Cmd) {
	if m.completed == len(m.apps) {
		return m, tea.Sequence(
			m.progress.SetPercent(1.0),
			// wait for the progress bar to finish animating
			tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
				return DoneMsg{
					Error: err,
				}
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

	for i, bm := range results {
		content := renderStats(bm, m.width)

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

func renderStats(bm BundleBenchmark, width int) string {
	border := styles.NormalText.Render(strings.Repeat("─", min(50, width-padding)))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		border,
		styles.NormalText.Render(fmt.Sprintf("Stats for %s app:", styles.Primary.Render(bm.AppName))),
		border,
		styles.NormalText.Render(fmt.Sprintf("%sRecorded on %s at %s", styles.IconStyle("🗓️"), bm.CreatedAt.Format("02/01/2006"), bm.CreatedAt.Format("15:04:05"))),
		styles.NormalText.Render(fmt.Sprintf("%sDescription: %s", styles.IconStyle("📝"), utils.Ternary(bm.Description == "", "-", bm.Description))),
		styles.NormalText.Render(fmt.Sprintf("%sApp: %s", styles.IconStyle("💻"), bm.AppName)),
		styles.Success.Render(fmt.Sprintf("%sBuild time: %.2fs", styles.IconStyle("🕒"), bm.Duration)),
		styles.Success.Render(fmt.Sprintf("%sMain bundle: %s", styles.IconStyle("🎯"), utils.FormatFileSize(bm.Stats.Initial.Main))),
		styles.Success.Render(fmt.Sprintf("%sRuntime bundle: %s", styles.IconStyle("⚙️"), utils.FormatFileSize(bm.Stats.Initial.Runtime))),
		styles.Success.Render(fmt.Sprintf("%sPolyfills bundle: %s", styles.IconStyle("🔧"), utils.FormatFileSize(bm.Stats.Initial.Polyfills))),
		styles.Warning.Render(fmt.Sprintf("%sInitial total: %s", styles.IconStyle("📦"), utils.FormatFileSize(bm.Stats.Initial.Total))),
		styles.Accent.Render(fmt.Sprintf("%sLazy chunks total: %s", styles.IconStyle("📦"), utils.FormatFileSize(bm.Stats.Lazy))),
		styles.Info.Render(fmt.Sprintf("%sBundle total: %s", styles.IconStyle("📦"), utils.FormatFileSize(bm.Stats.Total))),
		styles.Info.Render(fmt.Sprintf("%sStyles total: %s", styles.IconStyle("🎨"), utils.FormatFileSize(bm.Stats.Styles))),
		styles.Info.Render(fmt.Sprintf("%sAssets total: %s", styles.IconStyle("📂"), utils.FormatFileSize(bm.Stats.Assets))),
		styles.Info.Render(fmt.Sprintf("%sOverall total: %s", styles.IconStyle("📊"), utils.FormatFileSize(bm.Stats.OverallTotal))),
		border,
	)
}

func (m Model) viewportHeight() int {
	return m.height - lipgloss.Height(styles.Header(resultTitle))
}
