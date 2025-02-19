package bundle_analysis

import (
	"github.com/ionut-t/gonx/workspace"
	"time"
)

type StartMsg struct {
	Apps        []workspace.Application
	Description string
	StartTime   time.Time
}

type TotalProcessesMsg int

type NxCacheResetStartMsg struct {
	StartTime time.Time
}

type BuildStartMsg struct {
	App       string
	StartTime time.Time
}

type CalculateBundleSizeMsg struct {
	App       string
	StartTime time.Time
}

type WriteStatsMsg struct {
	App      string
	StartMsg time.Time
}

type BuildCompleteMsg struct {
	App       string
	Error     error
	EndTime   time.Time
	Benchmark BundleAnalysisBenchmark
}

type BuildFailedMsg struct {
	App     string
	EndTime time.Time
	Error   error
}

type DoneMsg struct {
	Error error
}
