package ui

import "github.com/charmbracelet/lipgloss"

var (
	Bright  = lipgloss.NewStyle().Bold(true)
	Dim     = lipgloss.NewStyle().Faint(true)
	Red     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Green   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	Blue    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	Magenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	Cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
)
