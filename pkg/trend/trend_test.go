package trend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRecordAndSummary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-trend.jsonl")
	log := New(path)

	// Empty log should give a stable trend.
	summary, err := log.Summary()
	if err != nil {
		t.Fatalf("Summary on empty log: %v", err)
	}
	if summary.TotalEntries != 0 {
		t.Errorf("expected 0 entries, got %d", summary.TotalEntries)
	}
	if summary.Trend != "stable" {
		t.Errorf("expected stable trend, got %s", summary.Trend)
	}

	// Record a few entries.
	for i := 0; i < 3; i++ {
		if err := log.Record("test prompt", "low", 70+i, "haiku"); err != nil {
			t.Fatalf("Record: %v", err)
		}
	}

	summary, err = log.Summary()
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if summary.TotalEntries != 3 {
		t.Errorf("expected 3 entries, got %d", summary.TotalEntries)
	}
	if summary.AvgScore != 71 {
		t.Errorf("expected avg_score 71.00, got %.2f", summary.AvgScore)
	}
}

func TestHashPrompt(t *testing.T) {
	h1 := hashPrompt("hello")
	h2 := hashPrompt("hello")
	h3 := hashPrompt("world")

	if h1 != h2 {
		t.Error("same prompt should produce same hash")
	}
	if h1 == h3 {
		t.Error("different prompts should produce different hashes")
	}
	if len(h1) != 16 {
		t.Errorf("expected 16 hex chars, got %d", len(h1))
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := expandHome("~/foo/bar")
	expected := filepath.Join(home, "foo/bar")
	if got != expected {
		t.Errorf("expandHome: got %s, want %s", got, expected)
	}

	// Non-tilde path should be returned unchanged.
	plain := "/absolute/path"
	if expandHome(plain) != plain {
		t.Error("plain path should not be modified")
	}
}

func TestTrendDirection(t *testing.T) {
	dir := t.TempDir()
	log := New(filepath.Join(dir, "trend.jsonl"))

	// 7 low-score entries followed by 7 high-score entries => improving.
	for i := 0; i < 7; i++ {
		if err := log.Record("p", "low", 40, "haiku"); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 7; i++ {
		if err := log.Record("p", "high", 80, "opus"); err != nil {
			t.Fatal(err)
		}
	}

	summary, err := log.Summary()
	if err != nil {
		t.Fatal(err)
	}
	if summary.Trend != "improving" {
		t.Errorf("expected improving, got %s", summary.Trend)
	}
}
