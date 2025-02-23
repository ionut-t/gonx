package bundle_analyser

import (
	"encoding/json"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	data "github.com/ionut-t/gonx/benchmark/data"
	"github.com/ionut-t/gonx/internal/constants"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

type BundleBenchmark data.BundleBenchmark

func (b *BundleBenchmark) calculateBundleSize(app workspace.Application) (*data.BuildStats, error) {
	stats := data.BuildStats{}

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
		return nil, err
	}

	path := cwd + "/" + app.OutputPath + "/browser"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = cwd + "/" + app.OutputPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, utils.Errorf("Build output directory not found: %s. You're might be using an unsupported version of NX", path)
	}

	if assetsSize, err := utils.FindAndCalculateAssetsSize(path); !os.IsNotExist(err) {
		stats.Assets = assetsSize
	}

	files, _ := os.ReadDir(path)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".js") &&
			!strings.HasSuffix(file.Name(), ".css") {
			continue
		}

		stat, _ := file.Info()

		if !stat.IsDir() {
			size := stat.Size()

			if strings.HasPrefix(file.Name(), "main") {
				stats.Initial.Main += size
			} else if strings.HasPrefix(file.Name(), "scripts") {
				stats.Initial.Runtime += size
			} else if strings.HasPrefix(file.Name(), "polyfills") {
				stats.Initial.Polyfills += size
			} else if strings.HasPrefix(file.Name(), "chunk-") || strings.Contains(file.Name(), "chunk") {
				stats.Lazy += size
			} else if strings.HasSuffix(file.Name(), ".css") {
				stats.Styles += size
			}
		}
	}

	stats.Initial.Total = stats.Initial.Main + stats.Initial.Runtime + stats.Initial.Polyfills
	stats.Total = stats.Initial.Total + stats.Lazy
	stats.OverallTotal = stats.Total + stats.Assets + stats.Styles

	return &stats, nil
}

func (b *BundleBenchmark) WriteStats(appName string, startTime time.Time) error {
	b.AppName = appName
	b.CreatedAt = time.Now()
	b.ID = uuid.New().String()
	b.Duration = time.Since(startTime).Seconds()

	benchmark, err := utils.ToJsonString(b)

	if err != nil {
		return err
	}

	var results []json.RawMessage
	currentValue, err := os.ReadFile(constants.BundleAnalyserFilePath)

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

	return os.WriteFile(constants.BundleAnalyserFilePath, content, 0644)
}

func startBenchmark(apps []workspace.Application, description string) tea.Cmd {
	// - Global messages: TotalProcessesMsg, NxCacheResetStartMsg, (2 total)
	// - For each app: BuildStartMsg, CalculateBundleSizeMsg, WriteStatsMsg, BuildCompleteMsg/BuildFailedMsg (4 per app)
	totalProcesses := 2 + len(apps)*4

	// Create channel for build results
	results := make(chan tea.Msg, totalProcesses)

	go func() {
		defer close(results)

		results <- TotalProcessesMsg(totalProcesses - 1) // -1 for this message

		results <- NxCacheResetStartMsg{}

		cmdReset := exec.Command("nx", "reset")
		cmdReset.Env = append(os.Environ(), "NX_DAEMON=false")
		if err := cmdReset.Run(); err != nil {
			// If reset fails, send failed messages for all apps
			for _, app := range apps {
				results <- BuildFailedMsg{
					App:   app.Name,
					Error: utils.Errorf("Workspace reset failed: %v", err),
				}
			}
			return
		}

		// Run builds sequentially
		for _, app := range apps {
			startTime := time.Now()
			benchmark := BundleBenchmark{Description: description}

			// Send startBenchmark message
			results <- BuildStartMsg{
				App:       app.Name,
				StartTime: startTime,
			}

			// Run build
			cmdBuild := exec.Command("nx", "build", app.Name)
			cmdBuild.Env = append(os.Environ(), "NX_DAEMON=false")
			if err := cmdBuild.Run(); err != nil {
				results <- BuildFailedMsg{
					App:     app.Name,
					EndTime: time.Now(),
					Error:   utils.Errorf("Build failed for %s with: %v", app, err),
				}
				continue // Continue with next app even if one fails
			}

			results <- CalculateBundleSizeMsg{App: app.Name, StartTime: time.Now()}

			stats, err := benchmark.calculateBundleSize(app)
			if err != nil {
				results <- BuildFailedMsg{
					App:     app.Name,
					EndTime: time.Now(),
					Error:   utils.Errorf("Bundle size calculation failed: %v", err),
				}
				continue
			}
			benchmark.Stats = *stats

			results <- WriteStatsMsg{App: app.Name, StartMsg: time.Now()}

			err = benchmark.WriteStats(app.Name, startTime)
			if err != nil {
				results <- BuildFailedMsg{
					App:     app.Name,
					EndTime: time.Now(),
					Error:   utils.Errorf("Failed to write stats: %v", err),
				}
				continue
			}

			results <- BuildCompleteMsg{
				App:       app.Name,
				Error:     nil,
				EndTime:   time.Now(),
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
