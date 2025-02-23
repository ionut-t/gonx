package benchmark

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	buildAnalyser "github.com/ionut-t/gonx/benchmark/build-analyser"
	buildAnalyserHistory "github.com/ionut-t/gonx/benchmark/build-analyser-history"
	bundleAnalyser "github.com/ionut-t/gonx/benchmark/bundle-analyser"
	bundleAnalyserHistory "github.com/ionut-t/gonx/benchmark/bundle-analyser-history"
	lintAnalyser "github.com/ionut-t/gonx/benchmark/lint-analyser"
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
	lintAnalyserView
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

	workspace workspace.Model

	taskList     tasksModel
	projectsList selectProjectsModel

	bundleAnalyser            bundleAnalyser.Model
	bundleAnalyserHistoryView bundleAnalyserHistory.Model

	buildAnalyser            buildAnalyser.Model
	buildAnalyserHistoryView buildAnalyserHistory.Model

	lintAnalyser lintAnalyser.Model

	width  int
	height int
}

type Options struct {
	Workspace workspace.Model
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
		return m.projectsList.View()

	case bundleAnalyserView:
		return viewStyle(m.bundleAnalyser.View())

	case bundleAnalyserHistoryView:
		return m.bundleAnalyserHistoryView.View()

	case buildAnalyserView:
		return viewStyle(m.buildAnalyser.View())

	case buildAnalyserHistoryView:
		return m.buildAnalyserHistoryView.View()

	case lintAnalyserView:
		return viewStyle(m.lintAnalyser.View())
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
			if (m.view == selectTasksView || m.isHistoryView()) && !m.bundleAnalyserHistoryView.Searching() {
				m.view = bundleAnalyserHistoryView
				m.bundleAnalyserHistoryView = bundleAnalyserHistory.New(m.width, m.height)
			}

		case "x":
			if m.view == selectTasksView || m.isHistoryView() && !m.bundleAnalyserHistoryView.Searching() {
				m.view = buildAnalyserHistoryView
				m.buildAnalyserHistoryView = buildAnalyserHistory.New(m.width, m.height)
			}
		}

	case taskMsg:
		m.view = selectAppsView

		includedTypes := []workspace.ProjectType{workspace.ApplicationType}

		if m.taskList.selected == lintAnalyserTask {
			includedTypes = append(includedTypes, workspace.LibraryType)
		}

		options := projectsListOptions{
			width:       m.width,
			height:      m.height,
			projects:    m.workspace.GetProjects(includedTypes),
			displayType: m.taskList.selected == lintAnalyserTask,
		}

		m.projectsList = newSelectionList(options)

	case projectsSelectedMsg:
		switch m.taskList.selected {
		case bundleAnalyserTask:
			m.view = bundleAnalyserView
			var apps = make([]workspace.Application, 0, len(msg))
			for _, app := range msg {
				apps = append(apps, app.(workspace.Application))
			}

			m.bundleAnalyser = bundleAnalyser.New(apps, m.width, m.height)

		case buildAnalyserTask:
			m.view = buildAnalyserView
			var apps = make([]string, 0, len(msg))
			for _, app := range msg {
				apps = append(apps, app.GetName())
			}
			m.buildAnalyser = buildAnalyser.New(apps, m.width, m.height)

		case lintAnalyserTask:
			m.view = lintAnalyserView
			m.lintAnalyser = lintAnalyser.New(msg, m.width, m.height)
		}

	case messages.NavigateToViewMsg:
		if m.view != selectTasksView {
			m.view = view(msg)
			return m, nil
		}
	}

	if m.view == selectTasksView {
		bModel, cmd := m.bundleAnalyser.Update(msg)
		m.bundleAnalyser = bModel.(bundleAnalyser.Model)
		cmds = append(cmds, cmd)
	}

	switch m.view {
	case selectTasksView:
		tModel, cmd := m.taskList.Update(msg)
		m.taskList = tModel.(tasksModel)
		cmds = append(cmds, cmd)

	case selectAppsView:
		aModel, cmd := m.projectsList.Update(msg)
		m.projectsList = aModel.(selectProjectsModel)
		cmds = append(cmds, cmd)

	case bundleAnalyserView:
		bModel, cmd := m.bundleAnalyser.Update(msg)
		m.bundleAnalyser = bModel.(bundleAnalyser.Model)
		cmds = append(cmds, cmd)

	case buildAnalyserView:
		bModel, cmd := m.buildAnalyser.Update(msg)
		m.buildAnalyser = bModel.(buildAnalyser.Model)
		cmds = append(cmds, cmd)

	case bundleAnalyserHistoryView:
		bModel, cmd := m.bundleAnalyserHistoryView.Update(msg)
		m.bundleAnalyserHistoryView = bModel.(bundleAnalyserHistory.Model)
		cmds = append(cmds, cmd)

	case buildAnalyserHistoryView:
		bModel, cmd := m.buildAnalyserHistoryView.Update(msg)
		m.buildAnalyserHistoryView = bModel.(buildAnalyserHistory.Model)
		cmds = append(cmds, cmd)

	case lintAnalyserView:
		lModel, cmd := m.lintAnalyser.Update(msg)
		m.lintAnalyser = lModel.(lintAnalyser.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) isHistoryView() bool {
	return slices.Contains(historyViews, m.view)
}
