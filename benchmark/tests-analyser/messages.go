package tests_analyser

import (
	"github.com/ionut-t/gonx/workspace"
	"time"
)

type StartMsg struct {
	Projects    []workspace.Project
	Description string
	Count       int
	StartTime   time.Time
}

type TotalProcessesMsg int

type NxCacheResetStartMsg struct {
	StartTime time.Time
}

type TestsStartMsg struct {
	Project    workspace.Project
	StartTime  time.Time
	CurrentRun int
	TotalRuns  int
}

type WriteStatsStartMsg struct {
	Project   workspace.Project
	StartTime time.Time
}

type WriteStatsCompleteMsg struct {
	Project   workspace.Project
	Time      time.Time
	Benchmark TestBenchmark
}

type WriteStatsFailedMsg struct {
	Project workspace.Project
	Time    time.Time
	Error   error
}

type TestsCompleteMsg struct {
	Project  workspace.Project
	Duration float64
}

type TestsFailedMsg struct {
	Project  workspace.Project
	RunIndex int
	EndTime  time.Time
	Error    error
}

type DoneMsg struct{}
