package program

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"io"
	"strings"
)

type actionType int

const (
	runBenchmark actionType = iota
	runBulkBenchmark
	viewMetrics
	viewBulkMetrics
)

type actionSelectedMsg struct {
	action actionType
}

type selectActionModel struct {
	list   list.Model
	action actionType
}

func (m selectActionModel) Init() tea.Cmd {
	return nil
}

func (m selectActionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "enter":
			_, ok := m.list.SelectedItem().(actionItem)
			if ok {
				m.action = actionType(m.list.Index())
			}

			return m, func() tea.Msg {
				return actionSelectedMsg{action: m.action}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectActionModel) View() string {
	return "\n" + m.list.View()
}

func newActionsList(width int) selectActionModel {
	items := []list.Item{
		actionItem("Run benchmark"),
		actionItem("Run bulk builds"),
		actionItem("View metrics"),
		actionItem("View bulk build metrics"),
	}

	const defaultWidth = 20

	actionList := list.New(items, actionItemDelegate{}, defaultWidth, listHeight)
	actionList.Title = "What would you like to do?"
	actionList.SetShowStatusBar(false)
	actionList.SetFilteringEnabled(false)
	actionList.Styles.Title = listTitleStyle
	actionList.Styles.HelpStyle = helpStyle
	actionList.InfiniteScrolling = true

	actionList.KeyMap = list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		// Selection.
		AcceptWhileFiltering: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select option"),
		),

		// Quitting.
		Quit: key.NewBinding(
			key.WithKeys("ctrl-q"),
			key.WithHelp("ctrl-q", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}

	actionList.SetWidth(width)

	m := selectActionModel{list: actionList}

	return m
}

type actionItem string

func (i actionItem) FilterValue() string { return "" }

type actionItemDelegate struct{}

func (d actionItemDelegate) Height() int                             { return 1 }
func (d actionItemDelegate) Spacing() int                            { return 0 }
func (d actionItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d actionItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(actionItem)
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
