package tests_analyser_history

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	data "github.com/ionut-t/gonx/benchmark/data"
	"github.com/ionut-t/gonx/ui/styles"
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

func createTable(metrics []data.TestBenchmark, width, height int) tableModel {
	colWidth := (width - 65) / 4

	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "App", Width: 25},
		{Title: "Created", Width: 20},
		{Title: "Min", Width: colWidth},
		{Title: "Max", Width: colWidth},
		{Title: "Average", Width: colWidth},
		{Title: "Total runs", Width: colWidth},
	}

	var rows []table.Row

	for idx, bm := range metrics {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", idx+1),
			bm.Project,
			bm.CreatedAt.Format("02/01/06 15:04"),
			fmt.Sprintf("%.2fs", bm.Min),
			fmt.Sprintf("%.2fs", bm.Max),
			fmt.Sprintf("%.2fs", bm.Average),
			fmt.Sprintf("%d", bm.TotalRuns),
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
