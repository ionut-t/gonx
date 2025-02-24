package keymap

import (
	"github.com/charmbracelet/bubbles/key"
	"reflect"
)

var Up = key.NewBinding(
	key.WithKeys("up", "k"),
	key.WithHelp("↑/k", "up"),
)

var Down = key.NewBinding(
	key.WithKeys("down", "j"),
	key.WithHelp("↓/j", "down"),
)

var Left = key.NewBinding(
	key.WithKeys("left", "h"),
	key.WithHelp("←/h", "left"),
)

var Right = key.NewBinding(
	key.WithKeys("right", "l"),
	key.WithHelp("→/l", "right"),
)

var Select = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "select option"),
)

var Back = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "back"),
)

var Quit = key.NewBinding(
	key.WithKeys("ctrl+q", "ctrl+c"),
	key.WithHelp("ctrl+(q/c)", "quit"),
)

var Help = key.NewBinding(
	key.WithKeys("?"),
	key.WithHelp("?", "help"),
)

var Search = key.NewBinding(
	key.WithKeys("/"),
	key.WithHelp("/", "search"),
)

var ExitSearch = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "exit search"),
)

var BundleAnalyserHistory = key.NewBinding(
	key.WithKeys("z"),
	key.WithHelp("z", "bundle analyser history"),
)

var BuildAnalyserHistory = key.NewBinding(
	key.WithKeys("x"),
	key.WithHelp("x", "build analyser history"),
)

var LintAnalyserHistory = key.NewBinding(
	key.WithKeys("c"),
	key.WithHelp("c", "lint analyser history"),
)

var TestsAnalyserHistory = key.NewBinding(
	key.WithKeys("v"),
	key.WithHelp("v", "tests analyser history"),
)

var ListView = key.NewBinding(
	key.WithKeys("1"),
	key.WithHelp("1", "list"),
)

var TableView = key.NewBinding(
	key.WithKeys("2"),
	key.WithHelp("2", "table"),
)

var JSONView = key.NewBinding(
	key.WithKeys("3"),
	key.WithHelp("3", "json"),
)

type Model struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Help       key.Binding
	Quit       key.Binding
	Back       key.Binding
	Select     key.Binding
	Search     key.Binding
	ExitSearch key.Binding

	BundleAnalyserHistory key.Binding
	BuildAnalyserHistory  key.Binding
	LintAnalyserHistory   key.Binding
	TestsAnalyserHistory  key.Binding

	ListView  key.Binding
	TableView key.Binding
	JSONView  key.Binding
}

func (k Model) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Select,
		k.Search,
		k.ExitSearch,
		k.ListView,
		k.TableView,
		k.JSONView,
		k.Back,
		k.Quit,
		k.Help,
	}
}

func (k Model) AllBindings() []key.Binding {
	return []key.Binding{
		k.Up,
		k.Down,
		k.Left,
		k.Right,
		k.Search,
		k.ExitSearch,
		k.BundleAnalyserHistory,
		k.BuildAnalyserHistory,
		k.LintAnalyserHistory,
		k.TestsAnalyserHistory,
		k.ListView,
		k.TableView,
		k.JSONView,
		k.Back,
		k.Quit,
		k.Help,
	}
}

func (k Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

func CombineKeys(a, b Model) Model {
	result := Model{}

	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)
	resultVal := reflect.ValueOf(&result).Elem()

	for i := 0; i < resultVal.NumField(); i++ {
		field := resultVal.Type().Field(i)

		// Try to get value from b first, if it's not zero value, use it
		bField := bVal.FieldByName(field.Name)
		if !bField.IsZero() {
			resultVal.Field(i).Set(bField)
			continue
		}

		// If b's field is zero value, use a's field
		aField := aVal.FieldByName(field.Name)
		if !aField.IsZero() {
			resultVal.Field(i).Set(aField)
		}
	}

	return result
}

func ReplaceBinding(bindings []key.Binding, newBinding key.Binding) []key.Binding {
	for i, binding := range bindings {
		if binding.Help().Key == newBinding.Help().Key {
			bindings[i] = newBinding
		}
	}

	return bindings
}

var DefaultKeyMap = Model{
	Up:     Up,
	Down:   Down,
	Left:   Left,
	Right:  Right,
	Search: Search,
	Back:   Back,
	Help:   Help,
	Quit:   Quit,
}

var historyKeyMap = Model{
	Search:    Search,
	ListView:  ListView,
	TableView: TableView,
	JSONView:  JSONView,
}

var HistoryKeyMap = CombineKeys(DefaultKeyMap, historyKeyMap)

var SearchKeyMap = Model{
	ExitSearch: ExitSearch,
	Quit:       Quit,
}

var ListKeyMap = Model{
	Up:     Up,
	Down:   Down,
	Select: Select,
}
