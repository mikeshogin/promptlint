package metrics

import (
	"strings"
)

var actionVerbs = map[string]string{
	"fix":        "fix",
	"repair":     "fix",
	"debug":      "fix",
	"create":     "create",
	"add":        "create",
	"implement":  "create",
	"build":      "create",
	"write":      "create",
	"review":     "review",
	"check":      "review",
	"analyze":    "review",
	"refactor":   "refactor",
	"rewrite":    "refactor",
	"restructure": "refactor",
	"delete":     "delete",
	"remove":     "delete",
	"design":     "create",
	"architect":  "create",
	"plan":       "create",
	"propose":    "create",
	"deploy":     "deploy",
	"migrate":    "refactor",
	"optimize":   "refactor",
	"improve":    "refactor",
	"explain":    "explain",
	"describe":   "explain",
}

var domainKeywords = map[string][]string{
	"code": {
		"function", "method", "variable", "class", "struct", "interface",
		"loop", "array", "string", "int", "bool", "error", "return",
		"import", "package", "module", "test", "unittest", "assert",
	},
	"architecture": {
		"architecture", "design", "pattern", "solid", "dip", "srp",
		"coupling", "cohesion", "dependency", "layer", "boundary",
		"component", "service", "microservice", "monolith", "graph",
		"cycle", "fan-out", "fan-in", "metric", "cqrs", "event sourcing",
		"saga", "domain driven", "hexagonal", "clean architecture",
		"event-driven", "distributed", "scalab", "resilien",
		"load balanc", "api gateway", "circuit breaker", "retry",
		"dead letter", "idempoten",
	},
	"infrastructure": {
		"docker", "kubernetes", "k8s", "nginx", "deploy", "ci", "cd",
		"pipeline", "server", "vps", "ssh", "container", "pod",
		"helm", "terraform", "ansible",
	},
	"content": {
		"article", "post", "blog", "linkedin", "twitter", "write",
		"publish", "draft", "headline", "summary", "translate",
	},
}

// DetectAction identifies the primary action requested in the prompt.
func DetectAction(text string) string {
	lower := strings.ToLower(text)
	words := strings.Fields(lower)

	for _, word := range words {
		clean := strings.Trim(word, ".,!?;:")
		if action, ok := actionVerbs[clean]; ok {
			return action
		}
	}

	return "unknown"
}

// ClassifyDomain returns a vector of domain scores.
func ClassifyDomain(text string) map[string]float64 {
	lower := strings.ToLower(text)
	result := make(map[string]float64)

	for domain, keywords := range domainKeywords {
		count := 0
		for _, kw := range keywords {
			count += strings.Count(lower, kw)
		}
		if count > 0 {
			// Normalize: more keywords = higher score, cap at 1.0
			score := float64(count) / 5.0
			if score > 1.0 {
				score = 1.0
			}
			result[domain] = score
		}
	}

	// Default if nothing detected
	if len(result) == 0 {
		result["general"] = 1.0
	}

	return result
}
