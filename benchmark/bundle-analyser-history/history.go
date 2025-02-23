package bundle_analyser_history

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	data "github.com/ionut-t/gonx/benchmark/data"
	"github.com/ionut-t/gonx/internal/keymap"
	"github.com/ionut-t/gonx/internal/messages"
	"github.com/ionut-t/gonx/ui/help"
	"github.com/ionut-t/gonx/ui/input"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/ui/viewport"
	"strings"
)

const padding = 2

const title = "📊 Bundle Analyser History"

type view int

const (
	listView view = iota
	tableView
	jsonView
)

type Model struct {
	view     view
	metrics  []data.BundleBenchmark
	viewport viewport.Model
	table    tableModel
	search   input.Model
	error    error

	help help.Model

	width, height int
}

func New(width, height int) Model {
	metrics, err := readAllMetrics()

	helpMenu := help.New(width, height)
	helpMenu.CombineWithHistoryKeys(keymap.Model{
		BuildAnalyserHistory: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "build analyser history"),
		),
	})

	model := Model{
		view:    listView,
		metrics: metrics,
		error:   err,
		width:   width,
		height:  height,
		search: input.New(input.Options{
			Width:       60,
			Placeholder: "Search by app name or description",
			Mode:        input.Text,
			HideHelp:    true,
		}),
		help: helpMenu,
	}

	options := viewport.Options{
		Width:   model.width,
		Height:  model.height - lipgloss.Height(styles.Header(model.search.View(), title)) - lipgloss.Height(model.help.View()),
		Content: getListContent(model),
	}

	model.viewport = viewport.New(options)

	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	if m.error != nil {
		return fmt.Sprintf("Error reading metrics: %s", m.error)
	}

	switch m.view {
	case listView, jsonView:
		return lipgloss.JoinVertical(
			lipgloss.Top,
			styles.Header(m.search.View(), title),
			m.viewport.View(),
			m.help.View(),
		)

	case tableView:
		return lipgloss.JoinVertical(
			lipgloss.Top,
			styles.SimpleHeader(m.search.View(), title),
			m.table.View(),
			m.help.View(),
		)
	}

	return m.viewport.View()
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

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.help.Keys.Back):
			if m.search.Focused() {
				m.search.Blur()
				m.help.Searching = false
				return m, cmd
			}

			return m, messages.Dispatch(messages.NavigateToViewMsg(0))

		case key.Matches(msg, m.help.Keys.ListView):
			if !m.search.Focused() {
				m.view = listView
				m.viewport.SetContent(getListContent(m))
			}

		case key.Matches(msg, m.help.Keys.TableView):
			if !m.search.Focused() {
				m.view = tableView
				m.table = createTable(m.getFilteredMetrics(), m.width, m.height-lipgloss.Height(styles.SimpleHeader(m.search.View(), title))-lipgloss.Height(m.help.View()))
				m.viewport.SetContent(m.table.View())
			}

		case key.Matches(msg, m.help.Keys.JSONView):
			if !m.search.Focused() {
				m.view = jsonView
				m.viewport.SetContent(getJsonContent(m))
			}

		case key.Matches(msg, m.help.Keys.Search):
			if !m.search.Focused() && !m.help.FullViewOpened() {
				m.search.Focus()
				m.help.Searching = true
				return m, cmd
			}
		}
	}

	if m.search.Focused() {
		searchModel, _ := m.search.Update(msg)
		m.search = searchModel.(input.Model)

		switch m.view {
		case listView:
			m.viewport.SetContent(getListContent(m))
		case tableView:
			m.table = createTable(m.getFilteredMetrics(), m.width, m.height-lipgloss.Height(styles.SimpleHeader(m.search.View(), title))-lipgloss.Height(m.help.View()))
		case jsonView:
			m.viewport.SetContent(getJsonContent(m))
		}

		return m, nil
	}

	// Handle keyboard and mouse events in the viewport
	viewportModel, cmd := m.viewport.Update(msg)
	m.viewport = viewportModel.(viewport.Model)

	if m.view == tableView {
		tModel, cmd := m.table.Update(msg)
		m.table = tModel.(tableModel)
		cmds = append(cmds, cmd)
	}

	helpMenu, cmd := m.help.Update(msg)
	m.help = helpMenu.(help.Model)

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) getFilteredMetrics() []data.BundleBenchmark {
	if m.search.Value() == "" {
		return m.metrics
	}

	filtered := make([]data.BundleBenchmark, 0)
	for _, metric := range m.metrics {
		if strings.Contains(metric.AppName, m.search.Value()) || strings.Contains(metric.Description, m.search.Value()) {
			filtered = append(filtered, metric)
		}
	}

	return filtered
}

func (m Model) Searching() bool {
	return m.search.Focused()
}
