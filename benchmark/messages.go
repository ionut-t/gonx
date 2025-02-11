package benchmark

import "time"

type StartMsg struct {
	Apps        []string
	Description string
	StartTime   time.Time
}

type TotalProcessesMsg struct {
	Total int
}

type NxCacheResetStartMsg struct {
	StartTime time.Time
}

type NxCacheResetCompleteMsg struct {
	EndTime time.Time
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
	Benchmark Benchmark
}

type BuildFailedMsg struct {
	App     string
	EndTime time.Time
	Error   error
}

type DoneMsg struct {
	Benchmarks []Benchmark
}
