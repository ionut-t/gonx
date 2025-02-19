package table

import (
	"fmt"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type Model struct {
	headers   []string
	rows      [][]string
	cursor    int
	viewport  viewport.Model
	width     int
	height    int
	xOffset   int
	selected  map[int]bool
	colWidths map[int]int
	styles    Styles
}

type Styles struct {
	Header   lipgloss.Style
	Row      lipgloss.Style
	Selected lipgloss.Style
	Cursor   lipgloss.Style
	Cell     lipgloss.Style
}

type Options struct {
	Headers []string
	Rows    [][]string
	Width   int
	Height  int
}

func DefaultStyles() Styles {
	return Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252")),
		Row:      lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Background(lipgloss.Color("238")),
		Cursor:   lipgloss.NewStyle().Background(lipgloss.Color("236")),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}

func New(options Options) Model {
	// Calculate exact column widths from the data
	colWidths := make(map[int]int)
	for i, h := range options.Headers {
		colWidths[i] = len(h) + 20
		for _, row := range options.Rows {
			if i < len(row) && len(row[i]) > colWidths[i] {
				colWidths[i] = len(row[i])
			}
		}
	}

	vp := viewport.New(options.Width, options.Height-2)

	model := Model{
		headers:   options.Headers,
		rows:      options.Rows,
		viewport:  vp,
		width:     options.Width,
		height:    options.Height,
		selected:  make(map[int]bool),
		colWidths: colWidths,
	}

	model.viewport.SetContent(model.renderRows())
	return model
}

func calculateColumnWidths(headers []string, rows [][]string) map[int]int {
	widths := make(map[int]int)

	for i, h := range headers {
		widths[i] = len(h)
	}

	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	return widths
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisibleVertical()
			}
		case "down":
			if m.cursor < len(m.rows)-1 {
				m.cursor++
				m.ensureVisibleVertical()
			}
		case "left":
			if m.xOffset > 0 {
				m.xOffset--
				m.viewport.SetContent(m.renderRows())
			}
		case "right":
			maxScroll := m.getMaxHorizontalScroll()
			if m.xOffset < maxScroll {
				m.xOffset++
				m.viewport.SetContent(m.renderRows())
			}
		case "space":
			m.selected[m.cursor] = !m.selected[m.cursor]
			m.viewport.SetContent(m.renderRows())
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2
		m.viewport.SetContent(m.renderRows())
		m.adjustColumnsToWidth()
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	cmd = tea.Batch(cmd, vpCmd)

	return m, cmd
}

func (m *Model) adjustColumnsToWidth() {
	availableWidth := m.width - 2
	totalWidth := 0
	visibleCols := 0

	for i := 0; i < len(m.headers); i++ {
		colWidth := m.colWidths[i] + 2
		if totalWidth+colWidth > availableWidth {
			break
		}
		totalWidth += colWidth
		visibleCols++
	}

	if m.xOffset > len(m.headers)-visibleCols {
		m.xOffset = len(m.headers) - visibleCols
		if m.xOffset < 0 {
			m.xOffset = 0
		}
	}
}

func (m Model) View() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		m.viewport.View(),
	)
}

func (m *Model) renderHeader() string {
	var sb strings.Builder

	for i, h := range m.headers {
		if i < m.xOffset {
			continue
		}
		// Right-pad header to column width
		paddedHeader := fmt.Sprintf("%-*s", m.colWidths[i], h)
		sb.WriteString(paddedHeader + " ")
	}

	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("252")).
		Render(sb.String())
}

func (m *Model) renderRows() string {
	var sb strings.Builder

	style := lipgloss.NewStyle()
	for i, row := range m.rows {
		if m.selected[i] {
			style = style.Background(lipgloss.Color("238"))
		}
		if i == m.cursor {
			style = style.Background(lipgloss.Color("236"))
		}

		for j, cell := range row {
			if j < m.xOffset {
				continue
			}
			// Right-pad each cell to column width
			paddedCell := fmt.Sprintf("%-*s", m.colWidths[j], cell)
			sb.WriteString(paddedCell + " ")
		}
		sb.WriteString("\n")
	}

	return style.Render(sb.String())
}

func (m *Model) ensureVisibleVertical() {
	minVisible := m.viewport.YOffset
	maxVisible := minVisible + m.viewport.Height

	if m.cursor < minVisible {
		m.viewport.YOffset = m.cursor
	} else if m.cursor >= maxVisible {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}

	m.viewport.SetContent(m.renderRows())
}

func (m *Model) getMaxHorizontalScroll() int {
	maxWidth := 0
	for _, row := range m.rows {
		rowWidth := 0
		for i, _ := range row {
			rowWidth += m.colWidths[i] + 2
		}
		if rowWidth > maxWidth {
			maxWidth = rowWidth
		}
	}
	return maxWidth - m.width
}
