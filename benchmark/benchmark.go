package benchmark

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
)

const folderName = ".gonx"
const benchmarkFile = "benchmarks.json"
const benchmarkFilePath = folderName + "/" + benchmarkFile

type Benchmark struct {
	AppName     string     `json:"appName"`
	Version     string     `json:"version"`
	CreatedAt   time.Time  `json:"createdAt"`
	Duration    float64    `json:"duration"`
	Description string     `json:"description"`
	Stats       buildStats `json:"stats"`
}

func (b *Benchmark) Build(app string) error {
	cmdReset := exec.Command("nx", "reset")
	if output, err := cmdReset.CombinedOutput(); err != nil {
		log.Printf("Reset failed: %v\nOutput: %s", err, string(output))
		return err
	}

	cmdBuild := exec.Command("nx", "build", app)
	if output, err := cmdBuild.CombinedOutput(); err != nil {
		log.Printf("Build failed: %v\nOutput: %s", err, string(output))
		return err
	}

	return nil
}

func (b *Benchmark) calculateBundleSize(appName string) (*buildStats, error) {
	stats := buildStats{}

	cwd, err := os.Getwd()

	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
		return nil, err
	}

	path := cwd + "/dist/apps/" + appName + "/browser"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("Build output directory not found: %s. You're might be using an unsuported version of NX", path)
		return nil, err
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

			if strings.HasPrefix(file.Name(), "main-") {
				stats.Initial.Main += size
			} else if strings.HasPrefix(file.Name(), "scripts-") {
				stats.Initial.Runtime += size
			} else if strings.HasPrefix(file.Name(), "polyfills-") {
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

func (b *Benchmark) WriteStats(appName string, startTime time.Time) error {
	b.AppName = appName
	b.CreatedAt = time.Now()
	b.Version = uuid.New().String()
	b.Duration = time.Since(startTime).Seconds()

	benchmark, err := utils.ToJsonString(b)

	if err != nil {
		return err
	}

	var previousBenchmarks []json.RawMessage
	currentValue, err := os.ReadFile(benchmarkFilePath)

	if err == nil && len(currentValue) > 0 {
		if err := json.Unmarshal(currentValue, &previousBenchmarks); err != nil {
			return err
		}
	}

	previousBenchmarks = append([]json.RawMessage{json.RawMessage(benchmark)}, previousBenchmarks...)

	content, err := json.MarshalIndent(previousBenchmarks, "", "  ")

	if err != nil {
		return err
	}

	err = os.MkdirAll(folderName, 0755)

	if err != nil {
		return err
	}

	return os.WriteFile(benchmarkFilePath, content, 0644)
}

func New(ws workspace.Workspace, description string) tea.Cmd {
	apps := ws.Applications

	// Create channel for build results
	results := make(chan tea.Msg, len(apps)*2) // Buffer for start and complete/fail messages

	// Run builds sequentially in a separate goroutine
	go func() {
		// First, run nx reset for the whole workspace
		cmdReset := exec.Command("nx", "reset")
		cmdReset.Env = append(os.Environ(), "NX_DAEMON=false")
		if err := cmdReset.Run(); err != nil {
			// If reset fails, send failed messages for all apps
			for _, app := range apps {
				results <- BuildFailedMsg{
					App:   app.Name,
					Error: fmt.Errorf("workspace reset failed: %v", err),
				}
			}
			return
		}

		// Run builds sequentially
		for _, app := range apps {
			startTime := time.Now()
			benchmark := Benchmark{Description: description}

			// Send start message
			results <- BuildStartMsg{
				App:       app.Name,
				StartTime: startTime,
			}

			// Run build
			cmdBuild := exec.Command("nx", "build", app.Name)
			cmdBuild.Env = append(os.Environ(), "NX_DAEMON=false")
			if err := cmdBuild.Run(); err != nil {
				results <- BuildFailedMsg{
					App:       app.Name,
					StartTime: startTime,
					Error:     fmt.Errorf("build failed: %v", err),
				}
				continue // Continue with next app even if one fails
			}

			stats, err := benchmark.calculateBundleSize(app.Name)
			if err != nil {
				results <- BuildFailedMsg{
					App:       app.Name,
					StartTime: startTime,
					Error:     fmt.Errorf("bundle size calculation failed: %v", err),
				}
				continue
			}
			benchmark.Stats = *stats

			err = benchmark.WriteStats(app.Name, startTime)
			if err != nil {
				results <- BuildFailedMsg{
					App:       app.Name,
					StartTime: startTime,
					Error:     fmt.Errorf("failed to write stats: %v", err),
				}
				continue
			}

			results <- BuildCompleteMsg{
				App:       app.Name,
				Error:     nil,
				StartTime: startTime,
				Benchmark: benchmark,
			}
		}
	}()

	// Create commands to read all expected messages
	var cmds []tea.Cmd
	for i := 0; i < len(apps)*2; i++ { // *2 for start and complete/fail messages
		cmds = append(cmds, func() tea.Msg {
			return <-results
		})
	}

	return tea.Batch(cmds...)
}
