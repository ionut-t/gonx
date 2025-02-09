package benchmark

import "time"

type StartMsg struct {
	StartTime time.Time
}

type BuildStartMsg struct {
	App       string
	StartTime time.Time
}

type BuildCompleteMsg struct {
	App       string
	Error     error
	StartTime time.Time
	Benchmark Benchmark
}

type BuildFailedMsg struct {
	App       string
	StartTime time.Time
	Error     error
}

type DoneMsg struct {
	Benchmarks []Benchmark
}
