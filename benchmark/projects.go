package benchmark

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ionut-t/gonx/internal/messages"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"io"
	"slices"
	"strings"
)

type projectsSelectedMsg []workspace.Project

type selectProjectsModel struct {
	list        list.Model
	projects    []workspace.Project
	displayType bool
}

func (m selectProjectsModel) Init() tea.Cmd {
	return nil
}

func (m selectProjectsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case " ":
			item, ok := m.list.SelectedItem().(selectItem)

			if ok {
				selectedIndex := m.list.Index()
				selected := slices.Contains(m.getProjectNames(), item.item.GetName())

				if selected {
					var newItems []workspace.Project

					for _, i := range m.projects {
						if i.GetName() != item.item.GetName() {
							newItems = append(newItems, i)
						}
					}

					item.selected = false
					m.projects = newItems
				} else {
					m.projects = append(m.projects, item.item)
					item.selected = true
				}

				items := m.list.Items()
				items[selectedIndex] = item
				m.list.SetItems(items)
			}

			return m, nil

		case "enter":
			if len(m.projects) == 0 {
				return m, nil
			}

			return m, func() tea.Msg {
				return projectsSelectedMsg(m.projects)
			}

		case "esc", "backspace":
			return m, messages.Dispatch(messages.NavigateToViewMsg(0))
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectProjectsModel) View() string {
	return "\n" + m.list.View()
}

type projectsListOptions struct {
	projects      []workspace.Project
	width, height int
	displayType   bool
}

func newSelectionList(options projectsListOptions) selectProjectsModel {
	var listItems []list.Item

	for _, item := range options.projects {
		listItems = append(listItems, selectItem{item: item, selected: false})
	}

	const defaultWidth = 20

	itemsList := list.New(listItems, selectItemDelegate{
		displayType: options.displayType,
	}, defaultWidth, options.height)
	itemsList.Title = "Select one or more projects"
	itemsList.SetShowStatusBar(false)
	itemsList.SetFilteringEnabled(false)
	itemsList.Styles.Title = listTitleStyle
	itemsList.Help.Styles.ShortKey = styles.Subtext0
	itemsList.Help.Styles.ShortDesc = styles.Overlay1
	itemsList.Help.Styles.ShortSeparator = styles.Subtext0
	itemsList.Styles.HelpStyle = helpStyle

	itemsList.SetWidth(options.width)
	itemsList.InfiniteScrolling = true

	itemsList.KeyMap = list.KeyMap{
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

	itemsList.AdditionalShortHelpKeys = func() []key.Binding {
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

	return selectProjectsModel{
		list:        itemsList,
		projects:    make([]workspace.Project, 0),
		displayType: options.displayType,
	}
}

type selectItem struct {
	item     workspace.Project
	selected bool
}

func (i selectItem) FilterValue() string { return "" }

type selectItemDelegate struct {
	displayType bool
}

func (d selectItemDelegate) Height() int                             { return 1 }
func (d selectItemDelegate) Spacing() int                            { return 0 }
func (d selectItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d selectItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(selectItem)

	if !ok {
		return
	}

	str := fmt.Sprintf("%s", i.item.GetName())

	fn := itemStyle.Render

	projectType := utils.Ternary(d.displayType, " ("+string(i.item.GetType())+")", "")

	if i.selected {
		fn = func(s ...string) string {
			return selectedItemStyle.PaddingLeft(4).Render("(*) " + strings.Join(s, " ") + projectType)
		}
	} else {
		fn = func(s ...string) string {
			return itemStyle.Render("( ) " + strings.Join(s, " ") + projectType)
		}
	}

	if index == m.Index() {
		fn = func(s ...string) string {
			if i.selected {
				return selectedItemStyle.Render("> (*) " + strings.Join(s, " ") + projectType)
			}

			return currentItemStyle.Render("> ( ) " + strings.Join(s, " ") + projectType)
		}
	}

	_, err := fmt.Fprint(w, fn(str))

	if err != nil {
		return
	}
}

func (m selectProjectsModel) getProjectNames() []string {
	var names []string

	for _, item := range m.projects {
		names = append(names, item.GetName())
	}

	return names
}
