package bundle_analyser_history

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	data "github.com/ionut-t/gonx/benchmark/data"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/utils"
)

var tableStyles = styles.DefaultTableStyles()

type tableModel struct {
	table table.Model
}

func (m tableModel) Init() tea.Cmd { return nil }

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableModel) View() string {
	return tableStyles.Base.Render(m.table.View())
}

func createTable(metrics []data.BundleBenchmark, width, height int) tableModel {
	lipgloss.NewStyle().Padding(0, 1)
	colWidth := (width - 55) / 8

	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "App", Width: 20},
		{Title: "Created", Width: 15},
		{Title: "Build time", Width: colWidth},
		{Title: "Main", Width: colWidth},
		{Title: "Runtime", Width: colWidth},
		{Title: "Polyfills", Width: colWidth},
		{Title: "Initial", Width: colWidth},
		{Title: "Lazy", Width: colWidth},
		{Title: "Styles", Width: colWidth},
		{Title: "Assets", Width: colWidth},
	}

	var rows []table.Row

	for idx, bm := range metrics {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", idx+1),
			bm.AppName,
			bm.CreatedAt.Format("02/01/06 15:04"),
			fmt.Sprintf("%.2fs", bm.Duration),
			utils.FormatFileSizeInMB(bm.Stats.Initial.Main),
			utils.FormatFileSizeInMB(bm.Stats.Initial.Runtime),
			utils.FormatFileSizeInMB(bm.Stats.Initial.Polyfills),
			utils.FormatFileSizeInMB(bm.Stats.Initial.Total),
			utils.FormatFileSizeInMB(bm.Stats.Lazy),
			utils.FormatFileSizeInMB(bm.Stats.Styles),
			utils.FormatFileSizeInMB(bm.Stats.Assets),
		})
	}

	newTable := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height-4),
		table.WithWidth(width-2),
	)

	newTable.SetStyles(table.Styles{
		Header:   tableStyles.Header,
		Selected: tableStyles.Selected,
		Cell:     tableStyles.Cell,
	})

	return tableModel{newTable}
}
