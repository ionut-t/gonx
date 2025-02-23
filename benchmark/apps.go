package benchmark

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ionut-t/gonx/internal/messages"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/workspace"
	"io"
	"slices"
	"strings"
)

type appsSelectedMsg []workspace.Application

type appsModel struct {
	list list.Model
	apps []workspace.Application
}

func (m appsModel) Init() tea.Cmd {
	return nil
}

func (m appsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				selected := slices.Contains(m.getAppNames(), item.app.Name)

				if selected {
					var newApps []workspace.Application

					for _, app := range m.apps {
						if app.Name != item.app.Name {
							newApps = append(newApps, app)
						}
					}

					item.selected = false
					m.apps = newApps
				} else {
					m.apps = append(m.apps, item.app)
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
				return appsSelectedMsg(m.apps)
			}

		case "esc", "backspace":
			return m, messages.Dispatch(messages.NavigateToViewMsg(0))
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m appsModel) View() string {
	return "\n" + m.list.View()
}

func newAppSelectionList(width, height int, apps []workspace.Application) appsModel {
	var items []list.Item

	for _, app := range apps {
		items = append(items, appItem{app: app, selected: false})
	}

	const defaultWidth = 20

	appsList := list.New(items, appItemDelegate{}, defaultWidth, height)
	appsList.Title = "Select one or more apps"
	appsList.SetShowStatusBar(false)
	appsList.SetFilteringEnabled(false)
	appsList.Styles.Title = listTitleStyle
	appsList.Help.Styles.ShortKey = styles.Subtext0
	appsList.Help.Styles.ShortDesc = styles.Overlay1
	appsList.Help.Styles.ShortSeparator = styles.Subtext0
	appsList.Styles.HelpStyle = helpStyle

	appsList.SetWidth(width)
	appsList.InfiniteScrolling = true

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
			key.WithHelp("ctrl-(q/c)", "quit"),
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
				key.WithKeys("esc"),
				key.WithHelp("esc", "back"),
			),
		}
	}

	return appsModel{list: appsList, apps: make([]workspace.Application, 0)}
}

type appItem struct {
	app      workspace.Application
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

	str := fmt.Sprintf("%s", i.app.Name)

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

func (m appsModel) getAppNames() []string {
	var names []string

	for _, app := range m.apps {
		names = append(names, app.Name)
	}

	return names
}
