package benchmark_data

import (
	"github.com/google/uuid"
	"github.com/ionut-t/gonx/utils"
	"github.com/ionut-t/gonx/workspace"
	"time"
)

type BundleBenchmark struct {
	ID          string     `json:"id"`
	AppName     string     `json:"appName"`
	CreatedAt   time.Time  `json:"createdAt"`
	Duration    float64    `json:"duration"`
	Description string     `json:"description"`
	Stats       BuildStats `json:"stats"`
}

type InitialStats struct {
	Main      int64 `json:"main"`
	Runtime   int64 `json:"runtime"`
	Polyfills int64 `json:"polyfills"`
	Total     int64 `json:"total"`
}

type BuildStats struct {
	Initial      InitialStats `json:"initial"`
	Lazy         int64        `json:"lazy"`
	Assets       int64        `json:"assets"`
	Total        int64        `json:"total"`
	OverallTotal int64        `json:"overallTotal"` // includes assets
	Styles       int64        `json:"styles"`
}

func (stats *BuildStats) String() string {
	return utils.PrettyJSON(stats)
}

type BuildBenchmark struct {
	ID          uuid.UUID `json:"id"`
	AppName     string    `json:"appName"`
	CreatedAt   time.Time `json:"createdAt"`
	Duration    float64   `json:"duration"`
	Description string    `json:"description"`
	Min         float64   `json:"min"`
	Max         float64   `json:"max"`
	Average     float64   `json:"avg"`
	TotalRuns   int       `json:"totalRuns"`
}

type LintBenchmark struct {
	ID          uuid.UUID             `json:"id"`
	Project     string                `json:"project"`
	Type        workspace.ProjectType `json:"type"`
	CreatedAt   time.Time             `json:"createdAt"`
	Duration    float64               `json:"duration"`
	Description string                `json:"description"`
	Min         float64               `json:"min"`
	Max         float64               `json:"max"`
	Average     float64               `json:"avg"`
	TotalRuns   int                   `json:"totalRuns"`
}
