package middleware

import (
	"github.com/mikeshogin/promptlint/pkg/analyzer"
	"github.com/mikeshogin/promptlint/pkg/config"
)

// Router decides which model to use based on prompt metrics.
type Router struct {
	Config *config.Config
}

// NewRouter creates a router with default settings.
func NewRouter() *Router {
	return &Router{
		Config: config.DefaultConfig(),
	}
}

// NewRouterWithConfig creates a router with custom configuration.
func NewRouterWithConfig(cfg *config.Config) *Router {
	return &Router{Config: cfg}
}

// RouteResult contains the routing decision and analysis.
type RouteResult struct {
	Model    string          `json:"model"`
	Score    int             `json:"score"`
	Analysis analyzer.Result `json:"analysis"`
}

// Route analyzes a prompt and returns routing decision.
func (r *Router) Route(prompt string) RouteResult {
	analysis := analyzer.Analyze(prompt)
	model := r.Config.RouteByScore(analysis.ComplexityScore)

	return RouteResult{
		Model:    model,
		Score:    analysis.ComplexityScore,
		Analysis: analysis,
	}
}
