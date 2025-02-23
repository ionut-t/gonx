package benchmark

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	buildAnalyser "github.com/ionut-t/gonx/benchmark/build-analyser"
	buildAnalyserHistory "github.com/ionut-t/gonx/benchmark/build-analyser-history"
	bundleAnalyser "github.com/ionut-t/gonx/benchmark/bundle-analyser"
	bundleAnalyserHistory "github.com/ionut-t/gonx/benchmark/bundle-analyser-history"
	"github.com/ionut-t/gonx/internal/messages"
	"github.com/ionut-t/gonx/workspace"
	"slices"
)

/*
Note for the future:

type BenchmarkType string

const (
	BundleAnalysis BenchmarkType = "bundle_analysis" // For analyzing bundle sizes

	BulkBuild       BenchmarkType = "bulk_build"       // Multiple builds of same app
	BulkLint        BenchmarkType = "bulk_lint"        // Multiple lint runs
	BulkUnitTest    BenchmarkType = "bulk_unit_test"   // Multiple unit test runs
	BulkIntegration BenchmarkType = "bulk_integration" // Multiple integration test runs
	BulkE2E         BenchmarkType = "bulk_e2e"         // Multiple e2e test runs
)
*/

type view int

const (
	selectTasksView view = iota
	selectAppsView
	bundleAnalyserView
	buildAnalyserView
	bundleAnalyserHistoryView
	buildAnalyserHistoryView
)

var historyViews = []view{
	bundleAnalyserHistoryView,
	buildAnalyserHistoryView,
}

var (
	viewStyle = lipgloss.NewStyle().Padding(0, 1).Render
)

type Model struct {
	view view

	workspace workspace.Workspace

	taskList tasksModel
	appList  appsModel

	bundleAnalysis            bundleAnalyser.Model
	bundleAnalysisHistoryView bundleAnalyserHistory.Model

	bulkBuild            buildAnalyser.Model
	bulkBuildHistoryView buildAnalyserHistory.Model

	width  int
	height int
}

type Options struct {
	Workspace workspace.Workspace
	Width     int
	Height    int
}

func New(options Options) Model {
	return Model{
		width:     options.Width,
		height:    options.Height,
		workspace: options.Workspace,
		taskList:  newTasksList(options.Width, options.Height),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	switch m.view {
	case selectTasksView:
		return viewStyle(m.taskList.View())

	case selectAppsView:
		return m.appList.View()

	case bundleAnalyserView:
		return viewStyle(m.bundleAnalysis.View())

	case bundleAnalyserHistoryView:
		return m.bundleAnalysisHistoryView.View()

	case buildAnalyserView:
		return viewStyle(m.bulkBuild.View())

	case buildAnalyserHistoryView:
		return m.bulkBuildHistoryView.View()
	}

	return ""
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		keyMsg := msg.String()
		switch keyMsg {

		case "ctrl+q", "ctrl+c":
			return m, tea.Quit

		case "z":
			if (m.view == selectTasksView || m.isHistoryView()) && !m.bundleAnalysisHistoryView.Searching() {
				m.view = bundleAnalyserHistoryView
				m.bundleAnalysisHistoryView = bundleAnalyserHistory.New(m.width, m.height)
			}

		case "x":
			if m.view == selectTasksView || m.isHistoryView() && !m.bundleAnalysisHistoryView.Searching() {
				m.view = buildAnalyserHistoryView
				m.bulkBuildHistoryView = buildAnalyserHistory.New(m.width, m.height)
			}
		}

	case taskMsg:
		m.view = selectAppsView
		m.appList = newAppSelectionList(m.width, m.height, m.workspace.Applications)

	case appsSelectedMsg:
		switch m.taskList.selected {
		case bundleAnalysisTask:
			m.view = bundleAnalyserView
			m.bundleAnalysis = bundleAnalyser.New(msg, m.width, m.height)

		case bulkBuildTask:
			m.view = buildAnalyserView
			var apps = make([]string, 0, len(msg))
			for _, app := range msg {
				apps = append(apps, app.Name)
			}
			m.bulkBuild = buildAnalyser.New(apps, m.width, m.height)
		}

	case messages.NavigateToViewMsg:
		if m.view != selectTasksView {
			m.view = view(msg)
			return m, nil
		}
	}

	if m.view == selectTasksView {
		bModel, cmd := m.bundleAnalysis.Update(msg)
		m.bundleAnalysis = bModel.(bundleAnalyser.Model)
		cmds = append(cmds, cmd)
	}

	switch m.view {
	case selectTasksView:
		tModel, cmd := m.taskList.Update(msg)
		m.taskList = tModel.(tasksModel)
		cmds = append(cmds, cmd)

	case selectAppsView:
		aModel, cmd := m.appList.Update(msg)
		m.appList = aModel.(appsModel)
		cmds = append(cmds, cmd)

	case bundleAnalyserView:
		bModel, cmd := m.bundleAnalysis.Update(msg)
		m.bundleAnalysis = bModel.(bundleAnalyser.Model)
		cmds = append(cmds, cmd)

	case buildAnalyserView:
		bModel, cmd := m.bulkBuild.Update(msg)
		m.bulkBuild = bModel.(buildAnalyser.Model)
		cmds = append(cmds, cmd)

	case bundleAnalyserHistoryView:
		bModel, cmd := m.bundleAnalysisHistoryView.Update(msg)
		m.bundleAnalysisHistoryView = bModel.(bundleAnalyserHistory.Model)
		cmds = append(cmds, cmd)

	case buildAnalyserHistoryView:
		bModel, cmd := m.bulkBuildHistoryView.Update(msg)
		m.bulkBuildHistoryView = bModel.(buildAnalyserHistory.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) isHistoryView() bool {
	return slices.Contains(historyViews, m.view)
}
