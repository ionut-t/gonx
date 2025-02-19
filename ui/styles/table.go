package styles

import "github.com/charmbracelet/lipgloss"

type TableStyles struct {
	Base     lipgloss.Style
	Header   lipgloss.Style
	Selected lipgloss.Style
	Cell     lipgloss.Style
}

func DefaultTableStyles() TableStyles {
	return TableStyles{
		Base: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(NormalText.GetForeground()),
		Header: lipgloss.
			NewStyle().
			Bold(true).
			Padding(0, 1).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(NormalText.GetForeground()).
			BorderBottom(true).
			Foreground(Primary.GetForeground()),
		Selected: Highlight,
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	}
}
