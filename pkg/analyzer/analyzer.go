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
	Action          string             `json:"action"`
	Domain          map[string]float64 `json:"domain"`
	Complexity      string             `json:"complexity"`
	ComplexityScore int                `json:"complexity_score"`

	// Routing suggestion
	SuggestedModel string `json:"suggested_model"`
}

// Analyze extracts metrics from a prompt string.
func Analyze(prompt string) Result {
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
	r.ComplexityScore = computeComplexityScore(prompt, r)
	r.Complexity = classifyFromScore(r.ComplexityScore)

	// Routing
	r.SuggestedModel = suggestModel(r)

	return r
}

func countWords(s string) int {
	return len(strings.Fields(s))
}

// computeComplexityScore returns a numeric score 0-100 based on prompt signals.
func computeComplexityScore(prompt string, r Result) int {
	score := 0

	// Length-based signals (0-15)
	switch {
	case r.Words > 200:
		score += 15
	case r.Words > 100:
		score += 10
	case r.Words > 50:
		score += 5
	}

	// Sentence count (0-5)
	if r.Sentences > 5 {
		score += 5
	} else if r.Sentences > 3 {
		score += 3
	}

	// Questions (0-5)
	if r.Questions > 2 {
		score += 5
	} else if r.Questions > 0 {
		score += 2
	}

	// Code block presence (0-5)
	if r.HasCodeBlock {
		score += 5
	}

	// Technical terms density (0-15)
	techCount := metrics.CountTechnicalTerms(prompt)
	switch {
	case techCount >= 8:
		score += 15
	case techCount >= 4:
		score += 10
	case techCount >= 2:
		score += 5
	}

	// Role/persona detection (0-10)
	if metrics.HasRolePersona(prompt) {
		score += 10
	}

	// Multi-step indicators (0-10)
	steps := metrics.CountMultiStepIndicators(prompt)
	switch {
	case steps >= 4:
		score += 10
	case steps >= 2:
		score += 6
	case steps >= 1:
		score += 3
	}

	// Constraint count (0-10)
	constraints := metrics.CountConstraints(prompt)
	switch {
	case constraints >= 5:
		score += 10
	case constraints >= 3:
		score += 7
	case constraints >= 1:
		score += 3
	}

	// Domain complexity (0-15)
	activeDomains := 0
	for _, v := range r.Domain {
		if v > 0.3 {
			activeDomains++
		}
	}
	if activeDomains > 2 {
		score += 10
	} else if activeDomains == 2 {
		score += 5
	}

	if archScore, ok := r.Domain["architecture"]; ok && archScore > 0.5 {
		score += 5
	}

	// Action verb complexity (0-15)
	switch r.Action {
	case "design":
		score += 15
	case "refactor":
		score += 12
	case "review":
		score += 8
	case "create":
		score += 5
	case "fix", "explain":
		score += 2
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// classifyFromScore converts numeric score to label.
func classifyFromScore(score int) string {
	switch {
	case score >= 50:
		return "high"
	case score >= 25:
		return "medium"
	default:
		return "low"
	}
}

func suggestModel(r Result) string {
	cfg := config.DefaultConfig()
	return cfg.RouteByScore(r.ComplexityScore)
}

// isLetter checks if a rune is a letter (unused but kept for future use).
var _ = unicode.IsLetter
