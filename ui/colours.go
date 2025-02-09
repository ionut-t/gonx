package ui

import "github.com/charmbracelet/lipgloss"

var (
	Red     = "1"
	Green   = "2"
	Yellow  = "3"
	Blue    = "4"
	Magenta = "5"
	Cyan    = "6"

	BrightFg  = lipgloss.NewStyle().Bold(true)
	DimFg     = lipgloss.NewStyle().Faint(true)
	RedFg     = lipgloss.NewStyle().Foreground(lipgloss.Color(Red))
	GreenFg   = lipgloss.NewStyle().Foreground(lipgloss.Color(Green))
	YellowFg  = lipgloss.NewStyle().Foreground(lipgloss.Color(Yellow))
	BlueFg    = lipgloss.NewStyle().Foreground(lipgloss.Color(Blue))
	MagentaFg = lipgloss.NewStyle().Foreground(lipgloss.Color(Magenta))
	CyanFg    = lipgloss.NewStyle().Foreground(lipgloss.Color(Cyan))
)
