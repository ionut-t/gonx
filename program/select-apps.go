package program

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ionut-t/gonx/workspace"
	"io"
	"slices"
	"strings"
)

type appsSelectedMsg struct {
	apps []string
}

type appSelectionModel struct {
	list list.Model
	apps []string
}

func (m appSelectionModel) Init() tea.Cmd {
	return nil
}

func (m appSelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case " ":
			item, ok := m.list.SelectedItem().(appItem)

			if ok {
				selectedIndex := m.list.Index()
				selected := slices.Contains(m.apps, item.name)

				if selected {
					var newApps []string

					for _, app := range m.apps {
						if app != item.name {
							newApps = append(newApps, app)
						}
					}

					item.selected = false
					m.apps = newApps
				} else {
					m.apps = append(m.apps, item.name)
					item.selected = true
				}

				items := m.list.Items()
				items[selectedIndex] = item
				m.list.SetItems(items)
			}

			return m, nil

		case "enter":
			if len(m.apps) == 0 {
				return m, nil
			}

			return m, func() tea.Msg {
				return appsSelectedMsg{apps: m.apps}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m appSelectionModel) View() string {
	return "\n" + m.list.View()
}

func newAppSelectionList(width int, apps []workspace.Application) appSelectionModel {
	var items []list.Item

	for _, app := range apps {
		items = append(items, appItem{name: app.Name, selected: false})
	}

	const defaultWidth = 20

	appsList := list.New(items, appItemDelegate{}, defaultWidth, listHeight)
	appsList.Title = "What applications would you like to benchmark?"
	appsList.SetShowStatusBar(false)
	appsList.SetFilteringEnabled(false)
	appsList.Styles.Title = listTitleStyle
	appsList.Styles.HelpStyle = helpStyle
	appsList.SetWidth(width)

	appsList.KeyMap = list.KeyMap{
		// Browsing.
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl-q", "ctrl-c"),
			key.WithHelp("ctrl-q/ctrl-c", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}

	appsList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys(" "),
				key.WithHelp("space", "select/deselect"),
			),
			key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "confirm"),
			),
			key.NewBinding(
				key.WithKeys("backspace"),
				key.WithHelp("backspace", "back"),
			),
		}
	}

	return appSelectionModel{list: appsList, apps: make([]string, 0)}
}

type appItem struct {
	name     string
	selected bool
}

func (i appItem) FilterValue() string { return "" }

type appItemDelegate struct {
	selected bool
}

func (d appItemDelegate) Height() int                             { return 1 }
func (d appItemDelegate) Spacing() int                            { return 0 }
func (d appItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d appItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(appItem)

	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i.name)

	fn := itemStyle.Render

	if i.selected {
		fn = func(s ...string) string {
			return selectedItemStyle.PaddingLeft(4).Render("(*) " + strings.Join(s, " "))
		}
	} else {
		fn = func(s ...string) string {
			return itemStyle.Render("( ) " + strings.Join(s, " "))
		}
	}

	if index == m.Index() {
		fn = func(s ...string) string {
			if i.selected {
				return selectedItemStyle.Render("> (*) " + strings.Join(s, " "))
			}

			return currentItemStyle.Render("> ( ) " + strings.Join(s, " "))
		}
	}

	_, err := fmt.Fprint(w, fn(str))

	if err != nil {
		return
	}
}
