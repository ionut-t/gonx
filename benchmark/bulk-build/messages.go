package bulk_build

import (
	"time"
)

type StartMsg struct {
	Apps        []string
	Description string
	Count       int
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

type WriteStatsStartMsg struct {
	App       string
	StartTime time.Time
}

type WriteStatsCompleteMsg struct {
	App       string
	Time      time.Time
	Benchmark BulkBuildBenchmark
}

type WriteStatsFailedMsg struct {
	App   string
	Time  time.Time
	Error error
}

type BuildCompleteMsg struct {
	App      string
	Duration float64
}

type BuildFailedMsg struct {
	App      string
	RunIndex int
	EndTime  time.Time
	Error    error
}

type DoneMsg struct{}
