package analyzer

import (
	"strings"
	"unicode"

	"github.com/mikeshogin/promptlint/pkg/config"
	"github.com/mikeshogin/promptlint/pkg/metrics"
)

// Result contains all extracted metrics from a prompt.
type Result struct {
	// Basic metrics
	Length     int `json:"length"`
	Words      int `json:"words"`
	Sentences  int `json:"sentences"`
	Paragraphs int `json:"paragraphs"`

	// Content detection
	HasCodeBlock bool `json:"has_code_block"`
	HasCodeRef   bool `json:"has_code_ref"`
	HasURL       bool `json:"has_url"`
	HasFilePath  bool `json:"has_file_path"`
	Questions    int  `json:"questions"`

	// Classification
	Action     string             `json:"action"`
	Domain     map[string]float64 `json:"domain"`
	Complexity string             `json:"complexity"`

	// Routing suggestion
	SuggestedModel string `json:"suggested_model"`
}

// Analyze extracts metrics from a prompt string using the default config.
func Analyze(prompt string) Result {
	return AnalyzeWithConfig(prompt, config.LoadOrDefault())
}

// AnalyzeWithConfig extracts metrics from a prompt string using the provided config.
func AnalyzeWithConfig(prompt string, cfg *config.Config) Result {
	r := Result{
		Domain: make(map[string]float64),
	}

	// Basic metrics
	r.Length = len(prompt)
	r.Words = countWords(prompt)
	r.Sentences = metrics.CountSentences(prompt)
	r.Paragraphs = metrics.CountParagraphs(prompt)

	// Content detection
	r.HasCodeBlock = metrics.HasCodeBlock(prompt)
	r.HasCodeRef = metrics.HasCodeRef(prompt)
	r.HasURL = metrics.HasURL(prompt)
	r.HasFilePath = metrics.HasFilePath(prompt)
	r.Questions = metrics.CountQuestions(prompt)

	// Classification
	r.Action = metrics.DetectAction(prompt)
	r.Domain = metrics.ClassifyDomain(prompt)
	r.Complexity = classifyComplexity(r)

	// Routing
	r.SuggestedModel = suggestModel(r, cfg)

	return r
}

func countWords(s string) int {
	return len(strings.Fields(s))
}

func classifyComplexity(r Result) string {
	score := 0

	// Length-based signals
	if r.Words > 200 {
		score += 2
	} else if r.Words > 50 {
		score++
	}

	if r.Sentences > 5 {
		score++
	}

	if r.Questions > 2 {
		score++
	}

	if r.HasCodeBlock {
		score++
	}

	// Multiple domains = cross-cutting concern = complex
	activeDomains := 0
	for _, v := range r.Domain {
		if v > 0.3 {
			activeDomains++
		}
	}
	if activeDomains > 2 {
		score += 2
	} else if activeDomains == 2 {
		score++
	}

	// Architecture domain is inherently complex
	if archScore, ok := r.Domain["architecture"]; ok && archScore > 0.5 {
		score += 2
	}

	// Action type weight: design/refactor are harder than fix/explain
	switch r.Action {
	case "create":
		score++
	case "refactor":
		score += 2
	}

	// High domain keyword density boosts complexity even for short prompts
	maxDomainScore := 0.0
	for _, v := range r.Domain {
		if v > maxDomainScore {
			maxDomainScore = v
		}
	}
	if maxDomainScore >= 0.8 {
		score++
	}

	switch {
	case score >= 4:
		return "high"
	case score >= 2:
		return "medium"
	default:
		return "low"
	}
}

func suggestModel(r Result, cfg *config.Config) string {
	if cfg != nil {
		return cfg.RouteByComplexity(r.Complexity)
	}
	switch r.Complexity {
	case "high":
		return "opus"
	case "medium":
		return "sonnet"
	default:
		return "haiku"
	}
}

// isLetter checks if a rune is a letter (unused but kept for future use).
var _ = unicode.IsLetter
