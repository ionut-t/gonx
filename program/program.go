package program

import (
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/benchmark"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/ui/suspense"
	"github.com/ionut-t/gonx/workspace"
	"os"
	"strings"
)

type view int

const (
	suspenseView view = iota
	benchmarkView
)

type Model struct {
	view      view
	suspense  suspense.Model
	workspace workspace.Model
	benchmark benchmark.Model

	error error

	width  int
	height int
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			ws, err := workspace.New()
			if err != nil {
				return workspace.ErrMsg{Err: err}
			}
			return workspace.DoneMsg{Workspace: *ws}
		},
		m.suspense.Init(),
		tea.SetWindowTitle("gonx"),
	)
}

func (m Model) View() string {
	if m.error != nil {
		err := m.error.Error()

		withPadding := lipgloss.NewStyle().Padding(1, 1)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			withPadding.Render(styles.Error.Render(strings.ToUpper(err[:1])+err[1:]+".")),
			withPadding.Render(styles.DimText.Render("Press ctrl+c to exit.")),
		)
	}

	switch m.view {
	case suspenseView:
		return lipgloss.NewStyle().Padding(1, 1).Render(m.suspense.View())

	case benchmarkView:
		return m.benchmark.View()

	default:
		return ""
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case workspace.DoneMsg:
		m.workspace = msg.Workspace
		m.view = benchmarkView

		m.benchmark = benchmark.New(benchmark.Options{
			Workspace: m.workspace,
			Width:     m.width,
			Height:    m.height,
		})

	case workspace.ErrMsg:
		m.suspense.Loading = false
		m.error = errors.New("no Nx workspace found. Please run this command in the root of your Nx workspace")
		return m, nil

	case spinner.TickMsg:
		if m.suspense.Loading {
			sModel, cmd := m.suspense.Update(msg)
			m.suspense = sModel.(suspense.Model)
			cmds = append(cmds, cmd)
		}
	}

	switch m.view {
	case suspenseView:
		sModel, cmd := m.suspense.Update(msg)
		m.suspense = sModel.(suspense.Model)
		cmds = append(cmds, cmd)

	case benchmarkView:
		bModel, cmd := m.benchmark.Update(msg)
		m.benchmark = bModel.(benchmark.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func New() {
	program := Model{
		view:     suspenseView,
		suspense: suspense.New("Scanning workspace...", true),
	}

	if _, err := tea.NewProgram(program, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
