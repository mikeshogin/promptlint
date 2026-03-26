// Package perf provides latency benchmarking for promptlint operations.
package perf

import (
	"time"

	"github.com/mikeshogin/promptlint/pkg/analyzer"
	"github.com/mikeshogin/promptlint/pkg/metrics"
	"github.com/mikeshogin/promptlint/pkg/router"
	"github.com/mikeshogin/promptlint/pkg/validator"
)

// PerfResult holds benchmark results for a single operation.
type PerfResult struct {
	Operation    string  `json:"operation"`
	Iterations   int     `json:"iterations"`
	AvgMs        float64 `json:"avg_ms"`
	MinMs        float64 `json:"min_ms"`
	MaxMs        float64 `json:"max_ms"`
	OpsPerSecond float64 `json:"ops_per_second"`
}

// Benchmark runs the provided function n times and returns latency statistics.
func Benchmark(operation string, n int, fn func()) PerfResult {
	if n <= 0 {
		n = 1
	}

	minDur := time.Duration(1<<63 - 1)
	maxDur := time.Duration(0)
	total := time.Duration(0)

	for i := 0; i < n; i++ {
		start := time.Now()
		fn()
		elapsed := time.Since(start)

		total += elapsed
		if elapsed < minDur {
			minDur = elapsed
		}
		if elapsed > maxDur {
			maxDur = elapsed
		}
	}

	avgDur := total / time.Duration(n)

	opsPerSec := 0.0
	if avgDur > 0 {
		opsPerSec = float64(time.Second) / float64(avgDur)
	}

	return PerfResult{
		Operation:    operation,
		Iterations:   n,
		AvgMs:        float64(avgDur.Nanoseconds()) / 1e6,
		MinMs:        float64(minDur.Nanoseconds()) / 1e6,
		MaxMs:        float64(maxDur.Nanoseconds()) / 1e6,
		OpsPerSecond: opsPerSec,
	}
}

// samplePrompt is a medium-complexity prompt used for all benchmarks.
const samplePrompt = `Implement a caching layer for the user authentication service.
The cache should store JWT tokens with expiry, support invalidation by user ID,
and be thread-safe. Use Redis as the backing store and ensure proper error handling.
Consider rate limiting and write unit tests for the implementation.`

// RunAll benchmarks analyze, route, validate, and nlp operations and returns
// the results as a slice.
func RunAll(iterations int) []PerfResult {
	r := router.NewDefault()
	v := validator.New()

	results := []PerfResult{
		Benchmark("analyze", iterations, func() {
			analyzer.Analyze(samplePrompt)
		}),
		Benchmark("route", iterations, func() {
			r.Route(samplePrompt)
		}),
		Benchmark("validate", iterations, func() {
			v.Validate(samplePrompt)
		}),
		Benchmark("nlp", iterations, func() {
			metrics.AnalyzeNLP(samplePrompt)
		}),
	}

	return results
}
