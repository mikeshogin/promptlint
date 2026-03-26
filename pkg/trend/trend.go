// Package trend tracks prompt complexity over time, enabling detection of
// degradation patterns across projects.
package trend

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultLogPath is the default path for the trend log file.
const DefaultLogPath = "~/.promptlint-trend.jsonl"

// TrendEntry represents a single recorded analysis event.
type TrendEntry struct {
	Timestamp    time.Time `json:"timestamp"`
	PromptHash   string    `json:"prompt_hash"`
	Complexity   string    `json:"complexity"`
	Score        int       `json:"score"`
	ModelRouted  string    `json:"model_routed"`
}

// TrendSummary contains aggregate statistics derived from the trend log.
type TrendSummary struct {
	TotalEntries  int     `json:"total_entries"`
	AvgScore      float64 `json:"avg_score"`
	Trend         string  `json:"trend"` // "improving", "degrading", "stable"
	Last7Avg      float64 `json:"last_7_avg"`
	Previous7Avg  float64 `json:"previous_7_avg"`
}

// TrendLog manages reading and writing of trend entries to a JSONL file.
type TrendLog struct {
	path string
}

// New creates a TrendLog backed by the given file path.
// Tilde in path is expanded to the user home directory.
func New(path string) *TrendLog {
	if path == "" {
		path = DefaultLogPath
	}
	return &TrendLog{path: expandHome(path)}
}

// NewDefault creates a TrendLog using DefaultLogPath.
func NewDefault() *TrendLog {
	return New(DefaultLogPath)
}

// Record appends a new entry derived from an analysis result to the log.
// promptText is used only to compute a stable hash; it is never stored.
func (l *TrendLog) Record(promptText, complexity string, score int, modelRouted string) error {
	entry := TrendEntry{
		Timestamp:   time.Now().UTC(),
		PromptHash:  hashPrompt(promptText),
		Complexity:  complexity,
		Score:       score,
		ModelRouted: modelRouted,
	}
	return l.append(entry)
}

// append serialises entry as JSON and appends a newline to the log file.
func (l *TrendLog) append(entry TrendEntry) error {
	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return fmt.Errorf("trend: create dir: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("trend: open log: %w", err)
	}
	defer f.Close()

	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("trend: marshal entry: %w", err)
	}

	_, err = fmt.Fprintf(f, "%s\n", line)
	return err
}

// Entries loads all entries from the log file.
// Returns an empty slice when the file does not exist.
func (l *TrendLog) Entries() ([]TrendEntry, error) {
	f, err := os.Open(l.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("trend: open log: %w", err)
	}
	defer f.Close()

	var entries []TrendEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e TrendEntry
		if err := json.Unmarshal(line, &e); err != nil {
			// Skip malformed lines rather than aborting.
			continue
		}
		entries = append(entries, e)
	}
	return entries, scanner.Err()
}

// Summary computes aggregate statistics from the trend log.
func (l *TrendLog) Summary() (TrendSummary, error) {
	entries, err := l.Entries()
	if err != nil {
		return TrendSummary{}, err
	}
	return computeSummary(entries), nil
}

// computeSummary derives a TrendSummary from a slice of entries.
func computeSummary(entries []TrendEntry) TrendSummary {
	n := len(entries)
	if n == 0 {
		return TrendSummary{Trend: "stable"}
	}

	var total float64
	for _, e := range entries {
		total += float64(e.Score)
	}
	avg := total / float64(n)

	// Last 7 vs previous 7 entries for trend direction.
	last7Avg := windowAvg(entries, n-7, n)
	prev7Avg := windowAvg(entries, n-14, n-7)

	trend := "stable"
	const threshold = 2.0
	if last7Avg-prev7Avg > threshold {
		trend = "improving"
	} else if prev7Avg-last7Avg > threshold {
		trend = "degrading"
	}

	return TrendSummary{
		TotalEntries: n,
		AvgScore:     round2(avg),
		Trend:        trend,
		Last7Avg:     round2(last7Avg),
		Previous7Avg: round2(prev7Avg),
	}
}

// windowAvg returns the average Score for entries[lo:hi], clamping bounds.
func windowAvg(entries []TrendEntry, lo, hi int) float64 {
	if lo < 0 {
		lo = 0
	}
	if hi > len(entries) {
		hi = len(entries)
	}
	if lo >= hi {
		return 0
	}
	var sum float64
	for _, e := range entries[lo:hi] {
		sum += float64(e.Score)
	}
	return sum / float64(hi-lo)
}

// hashPrompt returns the first 16 hex characters of SHA-256(prompt).
func hashPrompt(prompt string) string {
	h := sha256.Sum256([]byte(prompt))
	return fmt.Sprintf("%x", h[:8])
}

// expandHome replaces a leading "~/" with the user home directory.
func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// round2 rounds f to 2 decimal places.
func round2(f float64) float64 {
	return float64(int(f*100+0.5)) / 100
}
