package template

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// placeholderRe matches {{placeholder}} patterns.
var placeholderRe = regexp.MustCompile(`\{\{([^}]+)\}\}`)

// actionVerbs is a set of common instruction verbs used to detect clear instructions.
var actionVerbs = map[string]bool{
	"write": true, "create": true, "make": true, "build": true, "generate": true,
	"explain": true, "describe": true, "summarize": true, "analyze": true,
	"fix": true, "debug": true, "update": true, "change": true, "modify": true,
	"list": true, "show": true, "find": true, "get": true, "fetch": true,
	"convert": true, "translate": true, "transform": true, "format": true,
	"review": true, "check": true, "test": true, "validate": true,
	"help": true, "tell": true, "give": true, "provide": true, "suggest": true,
	"compare": true, "evaluate": true, "assess": true, "calculate": true,
	"implement": true, "add": true, "remove": true, "refactor": true,
	"produce": true, "extract": true, "parse": true, "return": true,
}

// Template holds a parsed prompt template with its metadata.
type Template struct {
	Name         string
	Text         string
	Placeholders []string
}

// TemplateScore holds scoring results for a template.
type TemplateScore struct {
	QualityScore       int      `json:"quality_score"`
	OptimalModel       string   `json:"optimal_model"`
	EstimatedCostRange string   `json:"estimated_cost_range"`
	Suggestions        []string `json:"suggestions"`
}

// ParseTemplate parses a template text, detecting {{placeholder}} patterns.
func ParseTemplate(text string) Template {
	matches := placeholderRe.FindAllStringSubmatch(text, -1)
	seen := make(map[string]bool)
	var placeholders []string
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if !seen[name] {
			seen[name] = true
			placeholders = append(placeholders, name)
		}
	}
	return Template{
		Text:         text,
		Placeholders: placeholders,
	}
}

// ScoreTemplate evaluates a template and returns a TemplateScore.
//
// Scoring breakdown (5 criteria, 20 pts each = 100 max):
//   - Has placeholders:      +20 (reusable template)
//   - Has clear instruction: +20 (starts with or contains an action verb)
//   - Reasonable length:     +20 (10-2000 chars is optimal for templates)
//   - Low complexity:        +20 (few code blocks, few questions, focused)
//   - Good readability:      +20 (short sentences, low punctuation density)
func ScoreTemplate(tmpl Template) TemplateScore {
	text := strings.TrimSpace(tmpl.Text)

	score := 0
	var suggestions []string

	// 1. Has placeholders (reusability)
	if len(tmpl.Placeholders) > 0 {
		score += 20
	} else {
		suggestions = append(suggestions,
			"add {{placeholder}} variables to make this template reusable across different inputs")
	}

	// 2. Has clear instruction
	if hasInstruction(text) {
		score += 20
	} else {
		suggestions = append(suggestions,
			"start with a clear action verb (e.g. 'Write', 'Analyze', 'Generate') for better results")
	}

	// 3. Reasonable length (10-2000 chars is ideal for a template skeleton)
	l := len(text)
	switch {
	case l >= 10 && l <= 2000:
		score += 20
	case l > 2000:
		suggestions = append(suggestions,
			"template is long; consider splitting into smaller focused templates for cheaper model routing")
	default:
		suggestions = append(suggestions,
			"template is too short; add more context so the model can produce consistent results")
	}

	// 4. Low complexity (good for cheap/fast models)
	complex := isComplex(text)
	if !complex {
		score += 20
	} else {
		suggestions = append(suggestions,
			"template has high complexity (code blocks, multiple questions); consider splitting for cheaper routing")
	}

	// 5. Good readability
	if isReadable(text) {
		score += 20
	} else {
		suggestions = append(suggestions,
			"improve readability: use shorter sentences and avoid excessive punctuation")
	}

	// Determine optimal model based on score and complexity
	optimalModel, costRange := recommendModel(score, complex, len(tmpl.Placeholders))

	// Add model-specific suggestions
	if optimalModel == "haiku" {
		suggestions = append(suggestions,
			"this template works best on haiku - low complexity and clear instructions allow the cheapest model")
	} else if optimalModel == "sonnet" {
		suggestions = append(suggestions,
			"consider adding more context and constraints to get better sonnet results")
	} else {
		suggestions = append(suggestions,
			"this template requires opus due to its complexity; simplify to reduce cost")
	}

	return TemplateScore{
		QualityScore:       score,
		OptimalModel:       optimalModel,
		EstimatedCostRange: costRange,
		Suggestions:        suggestions,
	}
}

// hasInstruction checks whether the text contains a recognisable action verb.
func hasInstruction(text string) bool {
	lower := strings.ToLower(text)
	words := tokenize(lower)
	for _, w := range words {
		if actionVerbs[w] {
			return true
		}
	}
	return false
}

// isComplex returns true when the template has characteristics that require
// a more capable (and expensive) model.
func isComplex(text string) bool {
	lower := strings.ToLower(text)

	// Code blocks signal technical complexity.
	if strings.Contains(lower, "```") || strings.Contains(lower, "~~~") {
		return true
	}

	// Many questions indicate open-ended / multi-step reasoning.
	if strings.Count(text, "?") > 2 {
		return true
	}

	// Very long plain text suggests multi-step reasoning requirement.
	if len(text) > 3000 {
		return true
	}

	return false
}

// isReadable performs a lightweight readability check:
// - average sentence length <= 25 words
// - punctuation density <= 15%
func isReadable(text string) bool {
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		return true
	}

	totalWords := 0
	for _, s := range sentences {
		totalWords += len(strings.Fields(s))
	}
	avgWords := totalWords / len(sentences)
	if avgWords > 25 {
		return false
	}

	// Punctuation density
	punctCount := 0
	for _, r := range text {
		if unicode.IsPunct(r) {
			punctCount++
		}
	}
	if len(text) > 0 && float64(punctCount)/float64(len(text)) > 0.15 {
		return false
	}

	return true
}

// recommendModel returns the optimal model name and estimated cost range
// given the quality score and complexity flag.
func recommendModel(score int, complex bool, placeholders int) (string, string) {
	switch {
	case !complex && score >= 80:
		return "haiku", "$0.00025-$0.00125 per 1K tokens"
	case !complex && score >= 50:
		return "haiku", "$0.00025-$0.00125 per 1K tokens"
	case complex && score >= 60:
		return "sonnet", "$0.003-$0.015 per 1K tokens"
	case complex:
		return "opus", "$0.015-$0.075 per 1K tokens"
	default:
		return "sonnet", "$0.003-$0.015 per 1K tokens"
	}
}

// tokenize splits text into lowercase words, stripping punctuation boundaries.
func tokenize(text string) []string {
	var words []string
	var current strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}

// splitSentences splits text into sentences on '.', '!', '?' boundaries.
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	for _, r := range text {
		current.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			s := strings.TrimSpace(current.String())
			if s != "" {
				sentences = append(sentences, s)
			}
			current.Reset()
		}
	}
	if current.Len() > 0 {
		s := strings.TrimSpace(current.String())
		if s != "" {
			sentences = append(sentences, s)
		}
	}
	return sentences
}

// FormatScore returns a human-readable summary of a TemplateScore.
func FormatScore(ts TemplateScore) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("quality_score: %d/100\n", ts.QualityScore))
	sb.WriteString(fmt.Sprintf("optimal_model: %s\n", ts.OptimalModel))
	sb.WriteString(fmt.Sprintf("estimated_cost_range: %s\n", ts.EstimatedCostRange))
	if len(ts.Suggestions) > 0 {
		sb.WriteString("suggestions:\n")
		for _, s := range ts.Suggestions {
			sb.WriteString(fmt.Sprintf("  - %s\n", s))
		}
	}
	return sb.String()
}
