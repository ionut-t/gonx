package benchmark

import (
	"encoding/json"
	"fmt"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/ionut-t/gonx/suspense"
)

const folder = ".gonx"
const benchmarkFile = "benchmarks.json"
const filePath = folder + "/" + benchmarkFile

type Model struct {
	suspense  suspense.Model
	benchmark Benchmark
}

type DoneMsg struct {
	Benchmarks []Benchmark
}

type ErrMsg struct {
	Err error
}

func (m Model) Init() tea.Cmd {
	if m.suspense.Loading {
		return m.suspense.Init()
	}

	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	suspenseModel, cmd := m.suspense.Update(msg)
	m.suspense = suspenseModel.(suspense.Model)

	return m, cmd
}

func (m Model) View() string {
	return m.suspense.View()
}

func CreateBenchmarkModel() {
	loadingMessage := "Scanning workspace..."
	bm := Model{
		suspense: suspense.New(loadingMessage, true),
	}

	if _, err := tea.NewProgram(bm).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type Benchmark struct {
	AppName     string     `json:"appName"`
	Version     string     `json:"version"`
	CreatedAt   time.Time  `json:"createdAt"`
	Duration    float64    `json:"duration"`
	Description string     `json:"description"`
	Stats       buildStats `json:"stats"`
}

func (b *Benchmark) String() string {
	return fmt.Sprintf("AppName: %s\nVersion: %s\nCreatedAt: %s\nDuration: %f\nDescription: %s\nStats: %v\n", b.AppName, b.Version, b.CreatedAt, b.Duration, b.Description, b.Stats)

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
	currentValue, err := os.ReadFile(filePath)

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

	err = os.MkdirAll(folder, 0755)

	if err != nil {
		return err
	}

	return os.WriteFile(filePath, content, 0644)
}

func New(ws workspace.Workspace, description string) ([]Benchmark, error) {
	apps := ws.Applications

	var benchmarks []Benchmark

	var err error

	for _, app := range apps {
		benchmark := Benchmark{Description: description}
		startTime := time.Now()

		err = benchmark.Build(app.Name)

		if err != nil {
			log.Printf("Build failed for %s: %v", app, err)
			continue
		}

		stats, err := benchmark.calculateBundleSize(app.Name)

		if err != nil {
			log.Printf("Failed to calculate bundle size for %s: %v", app, err)
		} else {
			benchmark.Stats = *stats

			err = benchmark.WriteStats(app.Name, startTime)

			if err != nil {
				panic(err)
			}
		}

		benchmarks = append(benchmarks, benchmark)
	}

	return benchmarks, err
}
