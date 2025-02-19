package bundle_analysis_history

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	chromStyles "github.com/alecthomas/chroma/v2/styles"
	data "github.com/ionut-t/gonx/benchmark/data"
	"github.com/ionut-t/gonx/internal/constants"
	"os"
)

func getJsonContent(model Model) string {
	// Convert metrics to pretty JSON
	jsonData, err := json.MarshalIndent(model.getFilteredMetrics(), "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}

	// Get JSON lexer
	lexer := lexers.Get("json")
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Get formatter and style
	formatter := formatters.Get("terminal")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	style := chromStyles.Get("dracula")
	if style == nil {
		style = chromStyles.Fallback
	}

	// Create buffer for output
	var buf bytes.Buffer
	iterator, err := lexer.Tokenise(nil, string(jsonData))
	if err != nil {
		return fmt.Sprintf("Error tokenizing: %v", err)
	}

	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return fmt.Sprintf("Error formatting: %v", err)
	}

	return buf.String()
}

func readAllMetrics() ([]data.BundleAnalysisBenchmark, error) {
	var metrics []data.BundleAnalysisBenchmark

	_bytes, err := os.ReadFile(constants.BundleAnalysisFilePath)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(_bytes, &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}
