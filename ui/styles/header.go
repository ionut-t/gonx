package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"os"
	"strings"
)

func SimpleHeader(elements ...string) string {
	if len(elements) == 0 {
		return ""
	}

	termWidth, _, _ := term.GetSize(uintptr(int(os.Stdout.Fd())))
	if termWidth <= 0 {
		termWidth = 80
	}

	// Style for individual elements
	elementStyle := Primary.Bold(true).
		PaddingRight(5).
		PaddingLeft(1)

	// Calculate total content width
	totalWidth := 0
	styledElements := make([]string, len(elements))
	for i, elem := range elements {
		styledElements[i] = elementStyle.Render(elem)
		totalWidth += lipgloss.Width(styledElements[i])
	}

	if len(elements) == 1 {
		return lipgloss.NewStyle().
			Width(termWidth).
			Align(lipgloss.Center).
			Render(styledElements[0])
	}

	// Calculate spacing between elements
	availableSpace := termWidth - totalWidth
	if availableSpace < 0 {
		availableSpace = 0
	}

	spacingCount := len(elements) - 1
	spacePerGap := availableSpace / spacingCount
	if spacePerGap < 0 {
		spacePerGap = 0
	}

	// Create the spacer
	spacer := ""
	if spacePerGap > 0 {
		spacer = strings.Repeat(" ", spacePerGap)
	}

	// Join elements with spacing
	return lipgloss.NewStyle().
		Width(termWidth).
		Render(strings.Join(styledElements, spacer))
}

func Header(elements ...string) string {
	header := lipgloss.NewStyle().
		Border(lipgloss.Border{Bottom: "_"}).
		BorderForeground(Primary.GetForeground()).Render(SimpleHeader(elements...))

	return lipgloss.NewStyle().PaddingBottom(1).Render(header)
}
