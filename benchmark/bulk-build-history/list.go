package bulk_build_history

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/ionut-t/gonx/ui/styles"
	"github.com/ionut-t/gonx/utils"
	"strings"
)

func getListContent(model Model) string {
	metrics := model.getFilteredMetrics()

	border := styles.NormalText.Render(strings.Repeat("─", min(50, model.width-padding)))

	var contents []string

	for i, bm := range metrics {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			styles.NormalText.Render(fmt.Sprintf("%sRecorded on %s at %s", styles.IconStyle("🗓️"), bm.CreatedAt.Format("02/01/2006"), bm.CreatedAt.Format("15:04:05"))),
			styles.NormalText.Render(fmt.Sprintf("%sDescription: %s", styles.IconStyle("📝"), utils.Ternary(bm.Description == "", "-", bm.Description))),
			styles.NormalText.Render(fmt.Sprintf("%sApp: %s", styles.IconStyle("💻"), bm.AppName)),
			styles.Success.Render(fmt.Sprintf("%sTotal runs: %d", styles.IconStyle("🔄"), bm.TotalRuns)),
			styles.Success.Render(fmt.Sprintf("%sMin: %.2fs", styles.IconStyle("🕒"), bm.Min)),
			styles.Success.Render(fmt.Sprintf("%sMax: %.2fs", styles.IconStyle("🕒"), bm.Max)),
			styles.Success.Render(fmt.Sprintf("%sAverage: %.2fs", styles.IconStyle("🕒"), bm.Average)),
		)

		if i < len(metrics)-1 {
			content += "\n\n" + border + "\n"
		}

		contents = append(contents, content)
	}

	return lipgloss.NewStyle().
		Padding(0, 4).
		Render(lipgloss.JoinVertical(
			lipgloss.Left,
			contents...,
		))
}
