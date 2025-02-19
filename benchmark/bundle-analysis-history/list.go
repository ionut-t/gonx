package bundle_analysis_history

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
			styles.Success.Render(fmt.Sprintf("%sBuild time: %.2fs", styles.IconStyle("🕒"), bm.Duration)),
			styles.Success.Render(fmt.Sprintf("%sMain bundle: %s", styles.IconStyle("🎯"), utils.FormatFileSize(bm.Stats.Initial.Main))),
			styles.Success.Render(fmt.Sprintf("%sRuntime bundle: %s", styles.IconStyle("⚙️"), utils.FormatFileSize(bm.Stats.Initial.Runtime))),
			styles.Success.Render(fmt.Sprintf("%sPolyfills bundle: %s", styles.IconStyle("🔧"), utils.FormatFileSize(bm.Stats.Initial.Polyfills))),
			styles.Warning.Render(fmt.Sprintf("%sInitial total: %s", styles.IconStyle("📦"), utils.FormatFileSize(bm.Stats.Initial.Total))),
			styles.Accent.Render(fmt.Sprintf("%sLazy chunks total: %s", styles.IconStyle("📦"), utils.FormatFileSize(bm.Stats.Lazy))),
			styles.Info.Render(fmt.Sprintf("%sBundle total: %s", styles.IconStyle("📦"), utils.FormatFileSize(bm.Stats.Total))),
			styles.Info.Render(fmt.Sprintf("%sStyles total: %s", styles.IconStyle("🎨"), utils.FormatFileSize(bm.Stats.Styles))),
			styles.Info.Render(fmt.Sprintf("%sAssets total: %s", styles.IconStyle("📂"), utils.FormatFileSize(bm.Stats.Assets))),
			styles.Info.Render(fmt.Sprintf("%sOverall total: %s", styles.IconStyle("📊"), utils.FormatFileSize(bm.Stats.OverallTotal))),
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
