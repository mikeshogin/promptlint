package analyzer

import (
	"testing"
)

func TestAnalyzeSimplePrompt(t *testing.T) {
	r := Analyze("Fix the bug in server.go:57")

	if r.Action != "fix" {
		t.Errorf("expected action 'fix', got '%s'", r.Action)
	}
	if !r.HasCodeRef {
		t.Error("expected HasCodeRef to be true")
	}
	if !r.HasFilePath {
		t.Error("expected HasFilePath to be true")
	}
	if r.Complexity != "low" {
		t.Errorf("expected complexity 'low', got '%s'", r.Complexity)
	}
	if r.SuggestedModel != "haiku" {
		t.Errorf("expected model 'haiku', got '%s'", r.SuggestedModel)
	}
	if r.ComplexityScore >= 25 {
		t.Errorf("expected score < 25 for simple prompt, got %d", r.ComplexityScore)
	}
}

func TestAnalyzeComplexPrompt(t *testing.T) {
	prompt := `Review the architecture of our microservice system.
We have coupling issues between the payment service and order service.
The dependency graph shows circular dependencies.
Can you analyze the SOLID violations?
What refactoring pattern would you suggest?
Also check the Docker deployment pipeline.

` + "```go\nfunc main() {\n  // code\n}\n```"

	r := Analyze(prompt)

	// Complex prompt should be at least medium
	if r.Complexity == "low" {
		t.Errorf("expected complexity >= 'medium', got '%s' (score=%d)", r.Complexity, r.ComplexityScore)
	}
	if r.SuggestedModel == "haiku" {
		t.Errorf("expected model != 'haiku', got '%s'", r.SuggestedModel)
	}
	if r.Questions < 2 {
		t.Errorf("expected at least 2 questions, got %d", r.Questions)
	}
	if !r.HasCodeBlock {
		t.Error("expected HasCodeBlock to be true")
	}
	if r.Domain["architecture"] == 0 {
		t.Error("expected architecture domain score > 0")
	}
	if r.ComplexityScore < 25 {
		t.Errorf("expected score >= 25, got %d", r.ComplexityScore)
	}
}

func TestAnalyzeEmptyPrompt(t *testing.T) {
	r := Analyze("")

	if r.Length != 0 {
		t.Errorf("expected length 0, got %d", r.Length)
	}
	if r.SuggestedModel != "haiku" {
		t.Errorf("expected model 'haiku', got '%s'", r.SuggestedModel)
	}
	if r.ComplexityScore != 0 {
		t.Errorf("expected score 0 for empty prompt, got %d", r.ComplexityScore)
	}
}

func TestDesignActionDetected(t *testing.T) {
	r := Analyze("Design a distributed caching layer with at least 3 replicas. It must handle 10k requests per second and should ensure consistency across regions.")

	if r.Action != "design" {
		t.Errorf("expected action 'design', got '%s'", r.Action)
	}
	if r.Complexity == "low" {
		t.Errorf("expected complexity > 'low', got '%s' (score=%d)", r.Complexity, r.ComplexityScore)
	}
}

func TestArchitecturePromptScoresHigh(t *testing.T) {
	prompt := `You are an expert software architect. Design a microservice architecture for an e-commerce platform.
Requirements:
1. Must support 100k concurrent users
2. Should use event-driven communication
3. Must ensure data consistency across services
4. No more than 5 services initially
5. Must include API gateway and circuit breaker patterns
6. Should be deployable on Kubernetes`

	r := Analyze(prompt)

	if r.Complexity != "high" {
		t.Errorf("expected complexity 'high', got '%s' (score=%d)", r.Complexity, r.ComplexityScore)
	}
	if r.SuggestedModel != "opus" {
		t.Errorf("expected model 'opus', got '%s'", r.SuggestedModel)
	}
	if r.ComplexityScore < 50 {
		t.Errorf("expected score >= 50, got %d", r.ComplexityScore)
	}
}
