package benchmark

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/internal/keymap"
	"github.com/ionut-t/gonx/ui/help"
	"github.com/ionut-t/gonx/ui/styles"
	"io"
	"strings"
)

var (
	listTitleStyle    = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	currentItemStyle  = styles.Accent.PaddingLeft(2)
	selectedItemStyle = styles.Primary.PaddingLeft(2)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

var tasks = [...]string{
	"Bundle analyser",
	"Build analyser",
}

type taskType int

const (
	bundleAnalysisTask taskType = iota
	bulkBuildTask
)

type taskMsg taskType

type tasksModel struct {
	list     list.Model
	selected taskType
	help     string
}

func (m tasksModel) Init() tea.Cmd {
	return nil
}

func (m tasksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			_, ok := m.list.SelectedItem().(taskItem)
			if ok {
				m.selected = taskType(m.list.Index())
			}

			return m, func() tea.Msg {
				return taskMsg(m.selected)
			}
		}
	}

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m tasksModel) View() string {
	return "\n" + m.list.View() + "\n" + m.help
}

func newTasksList(width, height int) tasksModel {
	items := make([]list.Item, len(tasks))

	for i, t := range tasks {
		items[i] = taskItem(t)
	}

	const defaultWidth = 20

	tasksList := list.New(items, taskItemDelegate{}, defaultWidth, height)
	tasksList.Title = "Select a task"
	tasksList.SetShowStatusBar(false)
	tasksList.SetFilteringEnabled(false)
	tasksList.Styles.Title = listTitleStyle
	tasksList.InfiniteScrolling = true
	tasksList.SetShowHelp(false)

	tasksList.KeyMap = list.KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select option"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl-q"),
			key.WithHelp("ctrl-q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}

	tasksList.SetWidth(width)

	listHelp := help.New(width, 10)

	listHelp.SetKeyMap(keymap.ListKeyMap)

	historyHelp := help.New(width, 10)

	historyKeys := keymap.Model{
		BundleAnalyserHistory: key.NewBinding(
			key.WithKeys("z"),
			key.WithHelp("z", "bundle analyser history"),
		),
		BuildAnalyserHistory: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "build analyser history"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl-q", "ctrl+c"),
			key.WithHelp("ctrl-(q/c)", "quit"),
		),
	}

	historyHelp.SetKeyMap(historyKeys)

	helpView := helpStyle.Render(listHelp.View() + "\n\n" + historyHelp.View())

	tasksList.SetHeight(height - lipgloss.Height(helpView))

	m := tasksModel{list: tasksList, help: helpView}

	return m
}

type taskItem string

func (i taskItem) FilterValue() string { return "" }

type taskItemDelegate struct{}

func (d taskItemDelegate) Height() int                             { return 1 }
func (d taskItemDelegate) Spacing() int                            { return 0 }
func (d taskItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d taskItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(taskItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return currentItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	_, err := fmt.Fprint(w, fn(str))

	if err != nil {
		return
	}
}
