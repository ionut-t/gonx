package build_analyser

import (
	"encoding/json"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	data "github.com/ionut-t/gonx/benchmark/data"
	"github.com/ionut-t/gonx/internal/constants"
	"github.com/ionut-t/gonx/utils"
	"math"
	"os"
	"os/exec"
	"time"
)

type BuildBenchmark data.BuildBenchmark

func (b *BuildBenchmark) WriteStats() error {
	b.CreatedAt = time.Now()

	benchmark, err := utils.ToJsonString(b)

	if err != nil {
		return err
	}

	var results []json.RawMessage
	currentValue, err := os.ReadFile(constants.BuildAnalyserFilePath)

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

	return os.WriteFile(constants.BuildAnalyserFilePath, content, 0644)
}

func startBenchmark(apps []string, description string, count int) tea.Cmd {
	// Calculate total number of processes:
	// - Initial message: TotalProcessesMsg (1 total)
	// - For each app:
	//  - Reset messages: NxCacheResetStartMsg (1 per app)
	//  - For each count run:
	//    - BuildStartMsg, CalculateBundleSizeMsg, WriteStatsMsg, BuildCompleteMsg/BuildFailedMsg (4 per count)
	totalProcesses := 1 + count*(1+len(apps)*4)

	// Create channel for build results
	results := make(chan tea.Msg, totalProcesses) // Buffer for start and complete/fail messages, nx reset

	// Run builds sequentially in a separate goroutine
	go func() {
		benchmarkStartTime := time.Now()

		defer close(results)

		results <- TotalProcessesMsg(totalProcesses - 1) // -1 for this message

		for _, app := range apps {
			var currentBuildEndTime time.Time

			durations := make([]float64, count)

			benchmark := BuildBenchmark{
				ID:          uuid.New(),
				AppName:     app,
				Description: description,
			}

			for i := 0; i < count; i++ {
				results <- NxCacheResetStartMsg{
					StartTime: time.Now(),
				}

				// First, run nx reset for the whole workspace
				cmdReset := exec.Command("nx", "reset")
				cmdReset.Env = append(os.Environ(), "NX_DAEMON=false")

				if err := cmdReset.Run(); err != nil {
					// If reset fails, send failed messages for all apps
					for _, app := range apps {
						results <- BuildFailedMsg{
							App:   app,
							Error: fmt.Errorf("nx reset failed: %v", err),
						}
					}
					return
				}

				startTime := time.Now()

				// Send start message
				results <- BuildStartMsg{
					App:       app,
					StartTime: startTime,
				}

				// Run build
				cmdBuild := exec.Command("nx", "build", app)
				cmdBuild.Env = append(os.Environ(), "NX_DAEMON=false")
				if err := cmdBuild.Run(); err != nil {
					results <- BuildFailedMsg{
						App:      app,
						RunIndex: i,
						EndTime:  time.Now(),
						Error:    fmt.Errorf("build failed: %v", err),
					}
					continue // Continue with next run even if one fails
				}

				currentBuildEndTime = time.Now()
				duration := currentBuildEndTime.Sub(startTime).Seconds()

				durations[i] = duration

				results <- BuildCompleteMsg{
					App:      app,
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

			results <- WriteStatsStartMsg{App: app, StartTime: time.Now()}

			err := benchmark.WriteStats()
			if err != nil {
				results <- WriteStatsFailedMsg{
					App:   app,
					Time:  time.Now(),
					Error: fmt.Errorf("failed to write stats: %v", err),
				}
				continue
			}

			results <- WriteStatsCompleteMsg{
				App:       app,
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
