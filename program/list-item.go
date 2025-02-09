package program

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/ui"
)

const listHeight = 10

var (
	listTitleStyle    = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	currentItemStyle  = ui.CyanFg.PaddingLeft(2)
	selectedItemStyle = ui.MagentaFg.PaddingLeft(2)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)
