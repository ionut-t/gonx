package benchmark

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bulkBuild "github.com/ionut-t/gonx/benchmark/bulk-build"
	bulkBuildHistory "github.com/ionut-t/gonx/benchmark/bulk-build-history"
	bundleAnalysis "github.com/ionut-t/gonx/benchmark/bundle-analysis"
	bundleAnalysisHistory "github.com/ionut-t/gonx/benchmark/bundle-analysis-history"
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
	bundleAnalysisView
	bulkBuildView
	bundleAnalysisHistoryView
	bulkBuildHistoryView
)

var historyViews = []view{
	bundleAnalysisHistoryView,
	bulkBuildHistoryView,
}

var (
	viewStyle = lipgloss.NewStyle().Padding(0, 1).Render
)

type Model struct {
	view view

	workspace workspace.Workspace

	taskList tasksModel
	appList  appsModel

	bundleAnalysis            bundleAnalysis.Model
	bundleAnalysisHistoryView bundleAnalysisHistory.Model

	bulkBuild            bulkBuild.Model
	bulkBuildHistoryView bulkBuildHistory.Model

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
		taskList:  newTasksList(options.Width),
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

	case bundleAnalysisView:
		return viewStyle(m.bundleAnalysis.View())

	case bundleAnalysisHistoryView:
		return m.bundleAnalysisHistoryView.View()

	case bulkBuildView:
		return viewStyle(m.bulkBuild.View())

	case bulkBuildHistoryView:
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
				m.view = bundleAnalysisHistoryView
				m.bundleAnalysisHistoryView = bundleAnalysisHistory.New(m.width, m.height)
			}

		case "x":
			if m.view == selectTasksView || m.isHistoryView() && !m.bundleAnalysisHistoryView.Searching() {
				m.view = bulkBuildHistoryView
				m.bulkBuildHistoryView = bulkBuildHistory.New(m.width, m.height)
			}
		}

	case taskMsg:
		m.view = selectAppsView
		m.appList = newAppSelectionList(m.width, m.workspace.Applications)

	case appsSelectedMsg:
		switch m.taskList.selected {
		case bundleAnalysisTask:
			m.view = bundleAnalysisView
			m.bundleAnalysis = bundleAnalysis.New(msg, m.width, m.height)

		case bulkBuildTask:
			m.view = bulkBuildView
			var apps = make([]string, 0, len(msg))
			for _, app := range msg {
				apps = append(apps, app.Name)
			}
			m.bulkBuild = bulkBuild.New(apps, m.width, m.height)
		}

	case messages.NavigateToViewMsg:
		if m.view != selectTasksView {
			m.view = view(msg)
			return m, nil
		}
	}

	if m.view == selectTasksView {
		bModel, cmd := m.bundleAnalysis.Update(msg)
		m.bundleAnalysis = bModel.(bundleAnalysis.Model)
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

	case bundleAnalysisView:
		bModel, cmd := m.bundleAnalysis.Update(msg)
		m.bundleAnalysis = bModel.(bundleAnalysis.Model)
		cmds = append(cmds, cmd)

	case bulkBuildView:
		bModel, cmd := m.bulkBuild.Update(msg)
		m.bulkBuild = bModel.(bulkBuild.Model)
		cmds = append(cmds, cmd)

	case bundleAnalysisHistoryView:
		bModel, cmd := m.bundleAnalysisHistoryView.Update(msg)
		m.bundleAnalysisHistoryView = bModel.(bundleAnalysisHistory.Model)
		cmds = append(cmds, cmd)

	case bulkBuildHistoryView:
		bModel, cmd := m.bulkBuildHistoryView.Update(msg)
		m.bulkBuildHistoryView = bModel.(bulkBuildHistory.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) isHistoryView() bool {
	return slices.Contains(historyViews, m.view)
}
