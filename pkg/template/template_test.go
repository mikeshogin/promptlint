package template

import (
	"strings"
	"testing"
)

func TestParseTemplate_NoPlaceholders(t *testing.T) {
	tmpl := ParseTemplate("Write a summary of the document.")
	if len(tmpl.Placeholders) != 0 {
		t.Errorf("expected 0 placeholders, got %d", len(tmpl.Placeholders))
	}
}

func TestParseTemplate_WithPlaceholders(t *testing.T) {
	tmpl := ParseTemplate("Write a {{style}} summary of {{document}}.")
	if len(tmpl.Placeholders) != 2 {
		t.Errorf("expected 2 placeholders, got %d: %v", len(tmpl.Placeholders), tmpl.Placeholders)
	}
	if tmpl.Placeholders[0] != "style" {
		t.Errorf("expected placeholder[0]='style', got '%s'", tmpl.Placeholders[0])
	}
	if tmpl.Placeholders[1] != "document" {
		t.Errorf("expected placeholder[1]='document', got '%s'", tmpl.Placeholders[1])
	}
}

func TestParseTemplate_DuplicatePlaceholders(t *testing.T) {
	tmpl := ParseTemplate("Translate {{text}} to {{language}}. Return only {{text}}.")
	if len(tmpl.Placeholders) != 2 {
		t.Errorf("expected 2 unique placeholders, got %d: %v", len(tmpl.Placeholders), tmpl.Placeholders)
	}
}

func TestScoreTemplate_HighQuality(t *testing.T) {
	text := "Write a {{tone}} summary of the following {{document}}. Focus on key points and return a concise result."
	tmpl := ParseTemplate(text)
	ts := ScoreTemplate(tmpl)

	if ts.QualityScore < 80 {
		t.Errorf("expected quality_score >= 80, got %d", ts.QualityScore)
	}
	if ts.OptimalModel == "" {
		t.Error("expected non-empty optimal_model")
	}
	if ts.EstimatedCostRange == "" {
		t.Error("expected non-empty estimated_cost_range")
	}
}

func TestScoreTemplate_NoPlaceholders(t *testing.T) {
	text := "Write a summary of the document."
	tmpl := ParseTemplate(text)
	ts := ScoreTemplate(tmpl)

	// Penalty for no placeholders
	if ts.QualityScore >= 100 {
		t.Errorf("expected quality_score < 100 when no placeholders, got %d", ts.QualityScore)
	}

	// Should suggest adding placeholders
	found := false
	for _, s := range ts.Suggestions {
		if strings.Contains(s, "placeholder") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected suggestion about placeholders, got: %v", ts.Suggestions)
	}
}

func TestScoreTemplate_EmptyText(t *testing.T) {
	tmpl := ParseTemplate("")
	ts := ScoreTemplate(tmpl)

	if ts.QualityScore >= 60 {
		t.Errorf("expected low quality_score for empty template, got %d", ts.QualityScore)
	}
}

func TestScoreTemplate_ComplexTemplate(t *testing.T) {
	text := "```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```\nWhat does this code do? How can I improve it? What are the edge cases? {{code}}"
	tmpl := ParseTemplate(text)
	ts := ScoreTemplate(tmpl)

	// Complex templates should go to sonnet or opus
	if ts.OptimalModel == "haiku" {
		t.Errorf("expected sonnet or opus for complex template, got %s", ts.OptimalModel)
	}
}

func TestScoreTemplate_HaikuModel(t *testing.T) {
	text := "Summarize {{input}} in one sentence."
	tmpl := ParseTemplate(text)
	ts := ScoreTemplate(tmpl)

	if ts.OptimalModel != "haiku" {
		t.Errorf("expected haiku for simple template, got %s", ts.OptimalModel)
	}
}

func TestFormatScore(t *testing.T) {
	ts := TemplateScore{
		QualityScore:       80,
		OptimalModel:       "haiku",
		EstimatedCostRange: "$0.00025-$0.00125 per 1K tokens",
		Suggestions:        []string{"add more context"},
	}
	out := FormatScore(ts)
	if !strings.Contains(out, "quality_score: 80/100") {
		t.Errorf("expected quality_score line in output, got: %s", out)
	}
	if !strings.Contains(out, "optimal_model: haiku") {
		t.Errorf("expected optimal_model line in output, got: %s", out)
	}
	if !strings.Contains(out, "add more context") {
		t.Errorf("expected suggestion in output, got: %s", out)
	}
}
