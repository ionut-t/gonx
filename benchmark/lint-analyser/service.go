package lint_analyser

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	data "github.com/ionut-t/gonx/benchmark/data"
	"github.com/ionut-t/gonx/internal/constants"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"math"
	"os"
	"os/exec"
	"time"
)

type LintBenchmark data.LintBenchmark

func (b *LintBenchmark) WriteStats() error {
	b.CreatedAt = time.Now()

	benchmark, err := utils.ToJsonString(b)

	if err != nil {
		return err
	}

	var results []json.RawMessage
	currentValue, err := os.ReadFile(constants.LintAnalyserFilePath)

	if err == nil && len(currentValue) > 0 {
		if err := json.Unmarshal(currentValue, &results); err != nil {
			return err
		}
	}

	results = append([]json.RawMessage{json.RawMessage(benchmark)}, results...)

	content, err := json.MarshalIndent(results, "", "  ")

	if err != nil {
		return err
	}

	err = os.MkdirAll(constants.BenchmarkFolderPath, 0755)

	if err != nil {
		return err
	}

	return os.WriteFile(constants.LintAnalyserFilePath, content, 0644)
}

func startBenchmark(projects []workspace.Project, description string, count int) tea.Cmd {
	/// Calculate total number of processes:
	// - Initial TotalProcessesMsg (1)
	// - For each app:
	//   - For each run (count times):
	//     - NxCacheResetStartMsg (1)
	//     - LintStartMsg + LintCompleteMsg/LintFailedMsg (2)
	//   - After all runs:
	//     - WriteStatsStartMsg + WriteStatsCompleteMsg/WriteStatsFailedMsg (2)
	totalProcesses := 1 + len(projects)*(3*count+2)

	// Create channel for build results
	results := make(chan tea.Msg, totalProcesses)

	go func() {
		benchmarkStartTime := time.Now()

		defer close(results)

		results <- TotalProcessesMsg(totalProcesses - 1) // -1 for this message

		for _, project := range projects {
			durations := make([]float64, count)

			benchmark := LintBenchmark{
				ID:          uuid.New(),
				Project:     project.GetName(),
				Type:        project.GetType(),
				Description: description,
			}

			for i := 0; i < count; i++ {
				results <- NxCacheResetStartMsg{
					StartTime: time.Now(),
				}

				// First, run nx reset for the whole workspace
				cmdReset := exec.Command("nx", "reset")

				if err := cmdReset.Run(); err != nil {
					// If reset fails, send failed messages for all projects
					for _, app := range projects {
						results <- LintFailedMsg{
							Project: app,
							Error:   fmt.Errorf("nx reset failed: %v", err),
						}
					}
					return
				}

				// Send start message
				results <- LintStartMsg{
					Project:   project,
					StartTime: time.Now(),
				}

				startTime := time.Now()

				// Run lint
				cmdLint := exec.Command("nx", "lint", project.GetName())

				if err := cmdLint.Run(); err != nil {
					results <- LintFailedMsg{
						Project:  project,
						RunIndex: i,
						EndTime:  time.Now(),
						Error:    fmt.Errorf("lint failed: %v", err),
					}
					continue // Continue with next run even if one fails
				}

				duration := time.Since(startTime).Seconds()

				durations[i] = duration

				results <- LintCompleteMsg{
					Project:  project,
					Duration: duration,
				}
			}

			var sum, minDuration, maxDuration float64
			minDuration = durations[0]
			for _, duration := range durations {
				sum += duration
				minDuration = math.Min(minDuration, duration)
				maxDuration = math.Max(maxDuration, duration)
			}

			benchmark.Duration = time.Since(benchmarkStartTime).Seconds()
			benchmark.Min = minDuration
			benchmark.Max = maxDuration
			benchmark.Average = sum / float64(len(durations))
			benchmark.TotalRuns = count

			results <- WriteStatsStartMsg{Project: project, StartTime: time.Now()}

			err := benchmark.WriteStats()
			if err != nil {
				results <- WriteStatsFailedMsg{
					Project: project,
					Time:    time.Now(),
					Error:   fmt.Errorf("failed to write stats: %v", err),
				}
				continue
			}

			results <- WriteStatsCompleteMsg{
				Project:   project,
				Time:      time.Now(),
				Benchmark: benchmark,
			}
		}
	}()

	// Create commands to read all expected messages
	var cmds []tea.Cmd
	for i := 0; i < totalProcesses; i++ {
		cmds = append(cmds, func() tea.Msg {
			return <-results
		})
	}

	return tea.Batch(cmds...)
}
